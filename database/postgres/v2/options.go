package postgres

import (
	"time"

	"github.com/jackc/pgx/v5"
)

type Option func(*config)

func WithMaxOpenConns(n int32) Option {
	return func(c *config) {
		c.maxOpenConns = n
	}
}

func WithMinConns(n int32) Option {
	return func(c *config) {
		c.minConns = n
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

func WithTracer(tracer pgx.QueryTracer) Option {
	return func(c *config) {
		c.tracer = tracer
	}
}
