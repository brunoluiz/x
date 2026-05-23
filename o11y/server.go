package o11y

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/brunoluiz/x/httpx"
)

type Option func(*options)

type options struct {
	addr       string
	pprof      http.Handler
	prometheus http.Handler
	healthz    http.Handler
}

func WithAddr(host string, port int) Option {
	return func(o *options) {
		o.addr = fmt.Sprintf("%s:%d", host, port)
	}
}

func WithPProf(h http.Handler) Option {
	return func(o *options) {
		o.pprof = h
	}
}

func WithPrometheus(h http.Handler) Option {
	return func(o *options) {
		o.prometheus = h
	}
}

func WithHealthz(h http.Handler) Option {
	return func(o *options) {
		o.healthz = h
	}
}

func Run(ctx context.Context, logger *slog.Logger, opts ...Option) error {
	o := &options{
		// defaults to 0.0.0.0 for dockerised workloads
		addr: "0.0.0.0:9090",
	}
	for _, opt := range opts {
		opt(o)
	}

	mux := http.NewServeMux()
	if o.healthz != nil {
		mux.Handle("/healthz", o.healthz)
	}
	if o.prometheus != nil {
		mux.Handle("/metrics", o.prometheus)
	}
	if o.pprof != nil {
		mux.Handle("/debug", o.pprof)
	}

	srv := httpx.New(o.addr, mux, httpx.WithName("o11y"), httpx.WithLogger(logger))
	defer srv.Close(ctx)

	return srv.Run(ctx)
}
