package sqlite

import "time"

type Option func(*config)

func defaultConfig() config {
	return config{
		maxOpenConns:    4,
		maxIdleConns:    4,
		connMaxLifetime: 30 * time.Minute,
		connMaxIdleTime: 5 * time.Minute,
		pragmas: []string{
			"PRAGMA journal_mode=WAL",
			"PRAGMA synchronous=NORMAL",
			"PRAGMA busy_timeout=5000",
			"PRAGMA foreign_keys=ON",
			"PRAGMA cache_size=-64000",
			"PRAGMA temp_store=MEMORY",
		},
	}
}

func WithMaxOpenConns(n int) Option {
	return func(c *config) {
		c.maxOpenConns = n
	}
}

func WithMaxIdleConns(n int) Option {
	return func(c *config) {
		c.maxIdleConns = n
	}
}

func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *config) {
		c.connMaxLifetime = d
	}
}

func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *config) {
		c.connMaxIdleTime = d
	}
}

func WithPragmas(pragmas ...string) Option {
	return func(c *config) {
		c.pragmas = append(c.pragmas, pragmas...)
	}
}
