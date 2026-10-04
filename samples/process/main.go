package main

import (
	"context"
	"time"

	"github.com/brunoluiz/x/app/v2/process"
	httpserver "github.com/brunoluiz/x/http/v2"
	"github.com/brunoluiz/x/o11y/v2"
	"github.com/brunoluiz/x/otel/v2"
	"github.com/hellofresh/health-go/v5"
)

type application struct{}

func (application) Run(ctx context.Context, logger process.Logger) error {
	shutdown, err := otel.Setup(ctx, otel.WithErrorHandler(&otel.DiscardErrorHandler{}))
	if err != nil {
		return err
	}
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	defer func(shutdownCtx context.Context) {
		if shutdownErr := shutdown(shutdownCtx); shutdownErr != nil {
			logger.Error("shutting down telemetry", "error", shutdownErr)
		}
	}(shutdownCtx)

	healthz, err := health.New()
	if err != nil {
		return err
	}
	server := httpserver.New(
		httpserver.WithAddr("0.0.0.0", 9090),
		httpserver.WithName("o11y"),
		httpserver.WithLogger(logger),
	)
	server.Handle("/", o11y.Handler(o11y.WithHealthz(healthz.Handler())))
	logger.Info("sample application started")
	return server.Run(ctx)
}

func main() {
	process.New(application{})
}
