package migrate

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
)

type DB interface {
	Get() *sql.DB
	Type() string
}

type config struct {
	steps   *int
	version *uint
}

type Option func(*config)

func WithSteps(n int) Option {
	return func(c *config) {
		c.steps = &n
	}
}

func WithVersion(v uint) Option {
	return func(c *config) {
		c.version = &v
	}
}

func Run(db DB, migrationsDir string, opts ...Option) error {
	var c config
	for _, opt := range opts {
		opt(&c)
	}

	drv, err := driverFor(db)
	if err != nil {
		return err
	}

	src, err := (&file.File{}).Open(migrationsDir)
	if err != nil {
		return fmt.Errorf("opening migrations source: %w", err)
	}

	m, err := migrate.NewWithInstance("file", src, db.Type(), drv)
	if err != nil {
		return fmt.Errorf("creating migrate instance: %w", err)
	}

	switch {
	case c.steps != nil:
		err = m.Steps(*c.steps)
	case c.version != nil:
		err = m.Migrate(*c.version)
	default:
		err = m.Up()
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migration: %w", err)
	}

	srcErr, dbErr := m.Close()
	if srcErr != nil {
		return fmt.Errorf("closing migrations source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("closing migrations database: %w", dbErr)
	}

	return nil
}

func driverFor(db DB) (database.Driver, error) {
	switch db.Type() {
	case "pgx":
		d, err := postgres.WithInstance(db.Get(), &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("creating postgres migrate driver: %w", err)
		}
		return d, nil
	case "sqlite":
		d, err := sqlite3.WithInstance(db.Get(), &sqlite3.Config{})
		if err != nil {
			return nil, fmt.Errorf("creating sqlite migrate driver: %w", err)
		}
		return d, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", db.Type())
	}
}
