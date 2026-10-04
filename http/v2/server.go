package httpserver

import (
	"context"
	"errors"
	"fmt"
	stdhttp "net/http"
	"time"
)

type Server struct {
	*stdhttp.Server

	mux             *stdhttp.ServeMux
	name            string
	shutdownTimeout time.Duration
	logger          Logger
}

type ServerOption func(*Server)

func WithShutdownTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.shutdownTimeout = d
	}
}

type Logger interface {
	Info(string, ...any)
}

func WithLogger(logger Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

func WithAddr(host string, port int) ServerOption {
	return func(s *Server) {
		s.Addr = fmt.Sprintf("%s:%d", host, port)
	}
}

func WithName(name string) ServerOption {
	return func(s *Server) {
		s.name = name
	}
}

func New(opts ...ServerOption) *Server {
	mux := stdhttp.NewServeMux()
	p := new(stdhttp.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	s := &Server{
		Server: &stdhttp.Server{
			Addr:              "0.0.0.0:4000",
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20,
			Protocols:         p,
		},
		name:            "app",
		shutdownTimeout: 5 * time.Second,
		logger:          nil,
		mux:             mux,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Handle(pattern string, handler stdhttp.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) HandleFunc(pattern string, handler func(stdhttp.ResponseWriter, *stdhttp.Request)) {
	s.mux.HandleFunc(pattern, handler)
}

func (s *Server) ServeHTTP(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.mux.ServeHTTP(w, r)
}

func WithReadTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.ReadTimeout = d
	}
}

func WithWriteTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.WriteTimeout = d
	}
}

func WithIdleTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.IdleTimeout = d
	}
}

func WithMaxHeaderBytes(n int) ServerOption {
	return func(s *Server) {
		s.MaxHeaderBytes = n
	}
}

func (s *Server) Run(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		if s.logger != nil {
			s.logger.Info("starting server", "address", s.Addr, "name", s.name)
		}
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.shutdownTimeout)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			shutdownErr := fmt.Errorf("failure to shutdown http:%s: %w", s.name, err)
			return errors.Join(shutdownErr, s.Server.Close())
		}
		return <-errChan
	}
}

func (s *Server) Close(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()
	return s.Shutdown(shutdownCtx)
}
