package middleware

import (
	"net/http"
	"time"

	"ledger-link/pkg/logger"
)

type MetricsMiddleware struct {
	logger *logger.Logger
}

func NewMetricsMiddleware(logger *logger.Logger) *MetricsMiddleware {
	return &MetricsMiddleware{
		logger: logger,
	}
}

type metricsResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{ResponseWriter: w}
}

func (rw *metricsResponseWriter) Status() int {
	return rw.status
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (m *MetricsMiddleware) TrackPerformance(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := wrapResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log request metrics
		m.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.Status(),
			"duration", duration,
			"user_agent", r.UserAgent(),
		)
	})
}
