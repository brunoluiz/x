package process

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/brunoluiz/x/logger/v2"
)

type Env string

const (
	EnvProduction Env = "production"
	EnvLocal      Env = "local"
)

type config struct {
	LogLevel string `enum:"debug,info,warn,error" kong:"default=info,env=LOG_LEVEL,name=log-level"`
}

type Exec interface {
	Run(ctx context.Context, logger Logger) error
}

type Logger interface {
	Info(string, ...any)
	Error(string, ...any)
}

func New[T Exec](exec T) {
	cfg := &config{}
	kong.Parse(exec, kong.Embed(cfg))

	logger := logger.New(logger.WithLevel(cfg.LogLevel))
	if err := runWithSignals(logger, exec); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}

func runWithSignals[T Exec](logger Logger, exec T) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	return run(ctx, logger, exec)
}

func run[T Exec](ctx context.Context, logger Logger, exec T) error {
	return exec.Run(ctx, logger)
}
