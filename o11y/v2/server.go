package o11y

import (
	"net/http"
)

type Option func(*options)

type options struct {
	pprof      http.Handler
	prometheus http.Handler
	healthz    http.Handler
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

func Handler(opts ...Option) http.Handler {
	o := &options{}
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

	return mux
}
