package closer

import "context"

type Log interface {
	Error(msg string, keysAndValues ...any)
	ErrorContext(ctx context.Context, msg string, keysAndValues ...any)
}

func WithLogContext(ctx context.Context, log Log, msg string, cb func(ctx context.Context) error) {
	if err := cb(ctx); err != nil {
		log.ErrorContext(ctx, msg, "error", err)
	}
}

func WithLog(ctx context.Context, log Log, msg string, cb func() error) {
	if err := cb(); err != nil {
		log.ErrorContext(ctx, msg, "error", err)
	}
}
