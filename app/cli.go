package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/brunoluiz/x/logger"
	"golang.org/x/sync/errgroup"
)

type cliConfig struct {
	LogLevel string `enum:"debug,info,warn,error" kong:"default=info,env=LOG_LEVEL,name=log-level"`
}

type CLIExec interface {
	Run(ctx context.Context, logger *slog.Logger) error
}

func CLI[T CLIExec](exec T) {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	cfg := &cliConfig{}
	kong.Parse(exec, kong.Embed(cfg))

	logger := logger.New(logger.WithLevel(cfg.LogLevel))
	if err := runCLI(ctx, logger, exec); err != nil {
		logger.ErrorContext(ctx, "application error", "error", err)
		os.Exit(1)
	}
}

func runCLI[T CLIExec](ctx context.Context, logger *slog.Logger, exec T) error {
	appCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	eg, ctx := errgroup.WithContext(appCtx)
	eg.Go(func() error {
		return exec.Run(ctx, logger)
	})

	if egErr := eg.Wait(); egErr != nil {
		return fmt.Errorf("application error: %w", egErr)
	}

	return nil
}
