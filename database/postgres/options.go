package postgres

import (
	"time"
)

type option func(*config)

func WithMaxOpenConns(n int32) func(*config) {
	return func(c *config) {
		c.maxOpenConns = n
	}
}

func WithConnMaxLifetime(d time.Duration) func(*config) {
	return func(c *config) {
		c.connMaxLifetime = d
	}
}

func WithConnMaxIdleTime(d time.Duration) func(*config) {
	return func(c *config) {
		c.connMaxIdleTime = d
	}
}
