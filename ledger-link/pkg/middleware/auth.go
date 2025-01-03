package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

type AuthMiddleware struct {
	authService models.AuthService
	logger      *logger.Logger
}

var publicPaths = map[string]bool{
	"/api/v1/auth/register": true,
	"/api/v1/auth/login":    true,
	"/api/v1/auth/refresh":  true,
	"/health":               true,
}

func NewAuthMiddleware(authService models.AuthService, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.logger.Info("Processing request", "path", r.URL.Path)

		authHeader := r.Header.Get("Authorization")
		m.logger.Info("Auth header", "header", authHeader)

		if authHeader == "" {
			m.logger.Error("No auth header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Error("Invalid auth header format", "header", authHeader)
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		user, err := m.authService.ValidateToken(r.Context(), token)
		if err != nil {
			m.logger.Error("Token validation failed", "error", err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Log user details
		userJSON, _ := json.Marshal(user)
		m.logger.Info("User from token", "user", string(userJSON))

		// Validate user role
		if user.Role == "" {
			m.logger.Info("Setting default role for user", "user_id", user.ID)
			user.Role = models.RoleUser // Default to user role if none specified
		}
		m.logger.Info("User role", "role", user.Role)

		// Use the new context helper
		ctx := auth.SetUserInContext(r.Context(), user)

		// Verify the user was set correctly
		if verifyUser, ok := auth.GetUserFromContext(ctx); ok {
			m.logger.Info("User set in context", "user_id", verifyUser.ID, "role", verifyUser.Role)
		} else {
			m.logger.Error("Failed to verify user in context")
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
