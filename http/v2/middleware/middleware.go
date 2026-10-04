package middleware

import "net/http"

// Middleware defines a standard HTTP middleware signature.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in the order provided.
// Example: Chain(h, A, B) results in A(B(h)).
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	if len(middlewares) == 0 {
		return handler
	}
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}
