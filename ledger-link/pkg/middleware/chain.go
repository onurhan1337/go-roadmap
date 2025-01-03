package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain creates a new middleware chain
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		return ChainHandler(final, middlewares...)
	}
}

// ChainHandler chains middleware handlers in reverse order
func ChainHandler(handler http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
