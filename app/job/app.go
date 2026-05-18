package job

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/brunoluiz/x/logger"
)

type config struct {
	LogLevel string `enum:"debug,info,warn,error" kong:"default=info,env=LOG_LEVEL,name=log-level"`
}

type Exec interface {
	Run(ctx context.Context, logger *slog.Logger) error
}

func New[T Exec](exec T) {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg := &config{}
	kong.Parse(exec, kong.Embed(cfg))

	logger := logger.New(logger.WithLevel(cfg.LogLevel))
	if err := run(ctx, logger, exec); err != nil {
		logger.ErrorContext(ctx, "application error", "error", err)
		os.Exit(1)
	}
}

func run[T Exec](ctx context.Context, logger *slog.Logger, exec T) error {
	return exec.Run(ctx, logger)
}
