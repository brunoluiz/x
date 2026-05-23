package httpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	*http.Server

	name            string
	shutdownTimeout time.Duration
	logger          interface {
		Info(string, ...any)
	}
}

type ServerOption func(*Server)

func WithShutdownTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.shutdownTimeout = d
	}
}

func WithLogger(logger interface {
	Info(string, ...any)
},
) ServerOption {
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

func New(handler http.Handler, opts ...ServerOption) *Server {
	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	s := &Server{
		Server: &http.Server{
			Addr:              "0.0.0.0:4000",
			Handler:           handler,
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
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
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
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failure to shutdown http:%s: %w", s.name, err)
		}
		return nil
	}
}

func (s *Server) Close(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()
	return s.Shutdown(shutdownCtx)
}
