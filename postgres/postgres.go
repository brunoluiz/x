package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/hellofresh/health-go/v5"
	_ "github.com/jackc/pgx/stdlib" // registers pgx driver for database/sql
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type config struct {
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

	if c.health != nil {
		if registerErr := c.health.Register(health.Config{
			Name:    "postgres",
			Timeout: 2 * time.Second,
			Check:   db.Health,
		}); registerErr != nil {
			db.Conn.Close()
			return nil, registerErr
		}
	}

	return db, nil
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
		return fmt.Errorf("database health check failed (ping): %w", err)
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
		return fmt.Errorf("database health check failed (query): %w", err)
	}

	return nil
}
