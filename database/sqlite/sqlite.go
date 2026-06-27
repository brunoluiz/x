package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	mcsqlite "modernc.org/sqlite"
)

type config struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	pragmas         []string
}

type DB struct {
	db *sql.DB
}

func New(ctx context.Context, path string, opts ...Option) (*DB, error) {
	c := defaultConfig()
	for _, opt := range opts {
		opt(&c)
	}

	d := &mcsqlite.Driver{}
	if _, err := d.Open(path); err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	db := sql.OpenDB(&connector{
		d:       d,
		dsn:     path,
		pragmas: c.pragmas,
	})
	db.SetMaxOpenConns(c.maxOpenConns)
	db.SetMaxIdleConns(c.maxIdleConns)
	db.SetConnMaxLifetime(c.connMaxLifetime)
	db.SetConnMaxIdleTime(c.connMaxIdleTime)

	sdb := &DB{db: db}

	if err := sdb.Health(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return sdb, nil
}

func (db *DB) Health(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("sqlite health check failed on ping: %w", err)
	}

	if err := db.db.QueryRowContext(pingCtx, "SELECT 1").Scan(new(int)); err != nil {
		return fmt.Errorf("sqlite health check failed on query: %w", err)
	}

	return nil
}

func (db *DB) Get() *sql.DB {
	return db.db
}

type connector struct {
	d       *mcsqlite.Driver
	dsn     string
	pragmas []string
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.d.Open(c.dsn)
	if err != nil {
		return nil, err
	}

	for _, p := range c.pragmas {
		execer, ok := conn.(driver.ExecerContext)
		if !ok {
			continue
		}
		if _, err := execer.ExecContext(ctx, p, nil); err != nil {
			conn.Close()
			return nil, fmt.Errorf("executing pragma %q: %w", p, err)
		}
	}

	return conn, nil
}

func (c *connector) Driver() driver.Driver {
	return c.d
}
