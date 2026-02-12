package logger

import (
	"log/slog"
	"os"
)

const (
	HandlerTypeUnknown HandlerType = iota
	HandlerTypeJSON
	HandlerTypeText
	HandlerTypeNoop
)

type Option func(*config)

type HandlerType int

type config struct {
	level   slog.Level
	handler HandlerType
}

func WithLevel(level string) Option {
	return func(o *config) {
		switch level {
		case "debug":
			o.level = slog.LevelDebug
		case "info":
			o.level = slog.LevelInfo
		case "warn":
			o.level = slog.LevelWarn
		case "error":
			o.level = slog.LevelError
		default:
			o.level = slog.LevelInfo
		}
	}
}

func WithType(t HandlerType) Option {
	return func(o *config) {
		switch t {
		case HandlerTypeText:
			o.handler = HandlerTypeText
		case HandlerTypeNoop:
			o.handler = HandlerTypeNoop
		case HandlerTypeJSON:
			fallthrough
		default:
			o.handler = HandlerTypeJSON
		}
	}
}

func New(opts ...Option) *slog.Logger {
	cfg := config{
		level:   slog.LevelInfo,
		handler: HandlerTypeJSON,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	handlerOpts := &slog.HandlerOptions{
		Level: cfg.level,
	}
	switch cfg.handler {
	case HandlerTypeNoop:
		return slog.New(slog.DiscardHandler)
	case HandlerTypeText:
		return slog.New(slog.NewTextHandler(os.Stdout, handlerOpts))
	case HandlerTypeJSON:
		fallthrough
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, handlerOpts))
	}
}
