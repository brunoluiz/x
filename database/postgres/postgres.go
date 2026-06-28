package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pgx-contrib/pgxotel"
)

type config struct {
	maxOpenConns    int32
	minConns        int32
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	tracerName      string
}

type DB struct {
	pool  *pgxpool.Pool
	sqlDB *sql.DB
}

func New(ctx context.Context, dsn string, opts ...Option) (*DB, error) {
	c := &config{
		maxOpenConns:    100,
		connMaxLifetime: 30 * time.Minute,
		connMaxIdleTime: 5 * time.Minute,
		tracerName:      "pgx",
	}
	for _, opt := range opts {
		opt(c)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing postgres DSN: %w", err)
	}
	poolConfig.MaxConns = c.maxOpenConns
	poolConfig.MinConns = c.minConns
	poolConfig.MaxConnLifetime = c.connMaxLifetime
	poolConfig.MaxConnIdleTime = c.connMaxIdleTime
	poolConfig.ConnConfig.Tracer = &pgxotel.QueryTracer{
		Name: c.tracerName,
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	db := &DB{pool: pool, sqlDB: sqlDB}
	if pingErr := db.Health(ctx); pingErr != nil {
		db.pool.Close()
		db.sqlDB.Close()
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

// Get must only be used for migrations or libs that are not compatible with pgx... otherwise use Pgx
func (db *DB) Get() *sql.DB {
	return db.sqlDB
}

func (db *DB) Pgx() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Type() string {
	return "pgx"
}
