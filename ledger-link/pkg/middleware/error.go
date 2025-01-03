package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"ledger-link/pkg/logger"
)

type ErrorMiddleware struct {
	logger *logger.Logger
}

func NewErrorMiddleware(logger *logger.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (m *ErrorMiddleware) HandleError(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the stack trace
				m.logger.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", r.URL.Path,
					"method", r.Method,
				)

				// Return 500 error to client
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error:   "InternalServerError",
					Code:    http.StatusInternalServerError,
					Message: "An unexpected error occurred",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
