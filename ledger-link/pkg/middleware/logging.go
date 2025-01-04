package middleware

import (
	"net/http"
	"time"

	"ledger-link/pkg/logger"
)

func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := NewResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			log.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration", duration,
			)
		})
	}
}
