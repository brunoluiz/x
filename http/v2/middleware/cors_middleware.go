package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

type corsConfig struct {
	allowedOrigins        []string
	allowMethods          []string
	allowHeaders          []string
	allowedExposedHeaders []string
	allowedMaxAge         int
	allowedCredentials    bool
}

type CORSOption func(*corsConfig)

func WithCORSAllowedOrigins(origins ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedOrigins = origins
	}
}

func WithCORSAllowedMethods(methods ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowMethods = methods
	}
}

func WithCORSAllowedHeaders(headers ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowHeaders = headers
	}
}

func WithCORSAllowedExposedHeaders(headers ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedExposedHeaders = headers
	}
}

func WithCORSAllowedMaxAge(maxAge int) CORSOption {
	return func(c *corsConfig) {
		c.allowedMaxAge = maxAge
	}
}

func WithCORSAllowedCredentials(enabled bool) CORSOption {
	return func(c *corsConfig) {
		c.allowedCredentials = enabled
	}
}

func CORS(opts ...CORSOption) Middleware {
	cfg := &corsConfig{
		allowedOrigins:        []string{},
		allowMethods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		allowHeaders:          []string{"Authorization", "Content-Type", "X-Project-ID"},
		allowedExposedHeaders: []string{},
		allowedMaxAge:         3600,
		allowedCredentials:    false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := w.Header()
			origin := r.Header.Get("Origin")
			hasOrigin := false

			// Check if the origin is allowed
			if len(cfg.allowedOrigins) > 0 {
				for _, allowedOrigin := range cfg.allowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						headers.Set("Access-Control-Allow-Origin", allowedOrigin)
						hasOrigin = true
						break
					}
				}
			}

			if !hasOrigin && origin != "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Handle preflight request
			// nolint
			if r.Method == http.MethodOptions {
				if origin != "" {
					headers.Set("Access-Control-Allow-Methods", strings.Join(cfg.allowMethods, ", "))
					headers.Set("Access-Control-Allow-Headers", strings.Join(cfg.allowHeaders, ", "))
					if len(cfg.allowedExposedHeaders) > 0 {
						headers.Set("Access-Control-Expose-Headers", strings.Join(cfg.allowedExposedHeaders, ", "))
					}
					if cfg.allowedMaxAge > 0 {
						headers.Set("Access-Control-Max-Age", strconv.Itoa(cfg.allowedMaxAge))
					}
					if cfg.allowedCredentials {
						headers.Set("Access-Control-Allow-Credentials", "true")
					}
				}

				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Set response headers for actual requests
			if hasOrigin {
				if len(cfg.allowedExposedHeaders) > 0 {
					headers.Set("Access-Control-Expose-Headers", strings.Join(cfg.allowedExposedHeaders, ", "))
				}
				if cfg.allowedCredentials {
					headers.Set("Access-Control-Allow-Credentials", "true")
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
