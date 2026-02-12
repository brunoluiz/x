package o11y

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/brunoluiz/x/closer"
	"github.com/brunoluiz/x/httpx"
	"github.com/hellofresh/health-go/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Option func(*options)

type options struct {
	addr       string
	pprof      bool
	prometheus bool
}

func WithAddr(host string, port int) Option {
	return func(o *options) {
		o.addr = fmt.Sprintf("%s:%d", host, port)
	}
}

func WithPProf(enabled bool) Option {
	return func(o *options) {
		o.pprof = enabled
	}
}

func WithPrometheus(enabled bool) Option {
	return func(o *options) {
		o.prometheus = enabled
	}
}

func Run(ctx context.Context, logger *slog.Logger, healthz *health.Health, opts ...Option) error {
	o := &options{
		addr:       "0.0.0.0:9090",
		pprof:      true,
		prometheus: true,
	}
	for _, opt := range opts {
		opt(o)
	}

	mux := http.NewServeMux()

	mux.Handle("/healthz", healthz.Handler())
	if o.prometheus {
		mux.Handle("/metrics", promhttp.Handler())
	}

	if o.pprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	srv := httpx.New(o.addr, mux, httpx.WithName("o11y"), httpx.WithLogger(logger))
	defer closer.WithLogContext(ctx, logger, "failed to shutdown o11y server", srv.Close)

	return srv.Run(ctx)
}
