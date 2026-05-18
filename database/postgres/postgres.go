package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgx-contrib/pgxotel"
)

type config struct {
	migration       fs.FS
	maxOpenConns    int32
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	logger          *slog.Logger
}

type DB struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func New(ctx context.Context, dsn string, logger *slog.Logger, opts ...option) (*DB, error) {
	c := &config{
		maxOpenConns:    25,
		connMaxLifetime: 5 * time.Minute,
		connMaxIdleTime: 5 * time.Minute,
	}
	for _, opt := range opts {
		opt(c)
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	config.MaxConns = c.maxOpenConns
	config.MaxConnLifetime = c.connMaxLifetime
	config.MaxConnIdleTime = c.connMaxIdleTime
	config.ConnConfig.Tracer = &pgxotel.QueryTracer{
		Name: "pgx",
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	db := &DB{pool: pool, logger: logger}
	if pingErr := db.Health(ctx); pingErr != nil {
		db.pool.Close()
		return nil, pingErr
	}

	return db, nil
}

func (db *DB) Health(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// Basic ping check
	if err := db.pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("database health check failed on ping: %w", err)
	}

	// Simple query to verify database is responsive
	if err := db.pool.QueryRow(pingCtx, "SELECT 1").Scan(new(int)); err != nil {
		return fmt.Errorf("database health check failed on query: %w", err)
	}

	return nil
}

func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}
