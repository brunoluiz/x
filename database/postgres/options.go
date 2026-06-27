package postgres

import (
	"time"
)

type option func(*config)

func WithMaxOpenConns(n int32) option {
	return func(c *config) {
		c.maxOpenConns = n
	}
}

func WithMinConns(n int32) option {
	return func(c *config) {
		c.minConns = n
	}
}

func WithConnMaxLifetime(d time.Duration) option {
	return func(c *config) {
		c.connMaxLifetime = d
	}
}

func WithConnMaxIdleTime(d time.Duration) option {
	return func(c *config) {
		c.connMaxIdleTime = d
	}
}

func WithTracerName(name string) option {
	return func(c *config) {
		c.tracerName = name
	}
}
