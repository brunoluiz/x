package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Option func(*slog.HandlerOptions)

func WithLevel(level string) Option {
	return func(o *slog.HandlerOptions) {
		switch strings.ToLower(level) {
		case "debug":
			o.Level = slog.LevelDebug
		case "info":
			o.Level = slog.LevelInfo
		case "warn":
			o.Level = slog.LevelWarn
		case "error":
			o.Level = slog.LevelError
		default:
			o.Level = slog.LevelInfo
		}
	}
}

func New(opts ...Option) *slog.Logger {
	handlerOpts := &slog.HandlerOptions{}
	for _, opt := range opts {
		opt(handlerOpts)
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, handlerOpts))
}
