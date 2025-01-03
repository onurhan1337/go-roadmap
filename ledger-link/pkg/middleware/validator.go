package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type ValidationMiddleware struct {
	validator *validator.Validate
}

func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validator.New(),
	}
}

func (m *ValidationMiddleware) ValidateRequest(schema interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read the body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Error reading request body", http.StatusBadRequest)
				return
			}

			// Restore the body for subsequent reads
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Parse the request body into the schema
			if err := json.Unmarshal(body, schema); err != nil {
				http.Error(w, "Invalid request format", http.StatusBadRequest)
				return
			}

			// Validate the schema
			if err := m.validator.Struct(schema); err != nil {
				validationErrors := err.(validator.ValidationErrors)
				errorResponse := make(map[string]string)
				for _, e := range validationErrors {
					errorResponse[e.Field()] = e.Tag()
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
