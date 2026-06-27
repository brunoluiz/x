package postgres

import (
	"time"
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

func WithTracerName(name string) Option {
	return func(c *config) {
		c.tracerName = name
	}
}
