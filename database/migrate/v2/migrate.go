package migrate

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

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

func Run(driver database.Driver, migrationsFS fs.FS, opts ...Option) (err error) {
	var c config
	for _, opt := range opts {
		opt(&c)
	}

	src, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("opening migrations source: %w", err)
	}
	defer func() {
		if closeErr := src.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("closing migrations source: %w", closeErr)
				return
			}
			err = errors.Join(err, fmt.Errorf("closing migrations source: %w", closeErr))
		}
	}()

	m, err := migrate.NewWithInstance("iofs", src, "database", driver)
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

	return nil
}
