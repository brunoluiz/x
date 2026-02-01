package postgres

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"log/slog"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/brunoluiz/x/errx"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/hellofresh/health-go/v5"
	_ "github.com/jackc/pgx/stdlib" // registers pgx driver for database/sql
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type config struct {
	migration       fs.FS
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	logger          *slog.Logger
	connTimeout     time.Duration
	maxRetries      int
	health          *health.Health
}

type DB struct {
	Conn   *sql.DB
	logger *slog.Logger
}

func New(ctx context.Context, dsn string, logger *slog.Logger, opts ...option) (*DB, error) {
	c := &config{
		maxOpenConns:    25,
		maxIdleConns:    5,
		connMaxLifetime: 5 * time.Minute,
		connMaxIdleTime: 5 * time.Minute,
		connTimeout:     30 * time.Second,
		maxRetries:      3,
	}
	for _, opt := range opts {
		opt(c)
	}

	conn, err := otelsql.Open("pgx", dsn, otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
	))
	if err != nil {
		return nil, err
	}

	db := &DB{Conn: conn, logger: logger}
	db.Conn.SetMaxOpenConns(c.maxOpenConns)
	db.Conn.SetMaxIdleConns(c.maxIdleConns)
	db.Conn.SetConnMaxLifetime(c.connMaxLifetime)
	db.Conn.SetConnMaxIdleTime(c.connMaxIdleTime)

	if _, err = otelsql.RegisterDBStatsMetrics(db.Conn, otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
	)); err != nil {
		db.Conn.Close()
		return nil, err
	}

	if pingErr := db.ping(ctx, c.connTimeout, c.maxRetries); pingErr != nil {
		db.Conn.Close()
		return nil, pingErr
	}

	if c.migration != nil {
		if migrateErr := up(db.Conn, c.migration, c.logger); migrateErr != nil {
			db.Conn.Close()
			return nil, migrateErr
		}
	}

	if c.health != nil {
		if registerErr := c.health.Register(health.Config{
			Name:    "postgres",
			Timeout: 2 * time.Second,
			Check:   db.Health,
		}); registerErr != nil {
			db.Conn.Close()
			return nil, errx.ErrInternal.Wrap(registerErr)
		}
	}

	return db, nil
}

func up(db *sql.DB, fs fs.FS, _ *slog.Logger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errx.ErrInternal.Wrap(err)
	}

	src, err := iofs.New(fs, "migrations")
	if err != nil {
		return errx.ErrInternal.New("migrations directory not found in embedded filesystem")
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return errx.ErrInternal.Wrap(err)
	}

	if upErr := m.Up(); !errors.Is(upErr, migrate.ErrNoChange) {
		return errx.ErrInternal.Wrapf(err, "migration failed")
	}

	return nil
}

func (db *DB) ping(ctx context.Context, timeout time.Duration, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, timeout)
		err := db.Conn.PingContext(pingCtx)
		cancel()
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				db.logger.WarnContext(ctx, "database ping failed, retrying",
					"attempt", attempt+1,
					"max_retries", maxRetries,
					"error", err)
				time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
				continue
			}
		} else {
			return nil
		}
	}
	return lastErr
}

func (db *DB) Health(ctx context.Context) error {
	// Basic ping check
	if err := db.Conn.PingContext(ctx); err != nil {
		return errx.ErrInternal.Wrapf(err, "database health check failed: ping")
	}

	// Check connection pool stats
	if db.logger.Enabled(ctx, slog.LevelDebug) {
		stats := db.Conn.Stats()
		db.logger.DebugContext(ctx, "database connection pool stats",
			"open_connections", stats.OpenConnections,
			"in_use", stats.InUse,
			"idle", stats.Idle,
			"wait_count", stats.WaitCount,
			"wait_duration", stats.WaitDuration,
			"max_idle_closed", stats.MaxIdleClosed,
			"max_lifetime_closed", stats.MaxLifetimeClosed,
		)
	}

	// Simple query to verify database is responsive
	if err := db.Conn.QueryRowContext(ctx, "SELECT 1").Scan(new(int)); err != nil {
		return errx.ErrInternal.Wrapf(err, "database health check failed: query")
	}

	return nil
}
