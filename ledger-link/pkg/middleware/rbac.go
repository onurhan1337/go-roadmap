package middleware

import (
	"encoding/json"
	"net/http"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

type RBACMiddleware struct {
	logger *logger.Logger
}

func NewRBACMiddleware(logger *logger.Logger) *RBACMiddleware {
	return &RBACMiddleware{
		logger: logger,
	}
}

// RequireAdmin ensures the user has admin role
func (m *RBACMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			m.logger.Error("No user in context - RequireAdmin")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userJSON, _ := json.Marshal(user)
		m.logger.Info("RequireAdmin check", "user", string(userJSON))

		if user.Role != models.RoleAdmin {
			m.logger.Error("User is not admin", "role", user.Role)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireUser ensures the user has at least user role
func (m *RBACMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			m.logger.Error("No user in context - RequireUser")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userJSON, _ := json.Marshal(user)
		m.logger.Info("RequireUser check", "user", string(userJSON))

		// Allow both regular users and admins
		if user.Role == models.RoleUser || user.Role == models.RoleAdmin {
			m.logger.Info("User role check passed", "role", user.Role)
			next.ServeHTTP(w, r)
			return
		}

		m.logger.Error("Invalid user role", "role", user.Role)
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

// RequireOwnerOrAdmin ensures the user is either an admin or the owner of the resource
func (m *RBACMiddleware) RequireOwnerOrAdmin(getResourceID func(*http.Request) uint) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				m.logger.Error("No user in context - RequireOwnerOrAdmin")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userJSON, _ := json.Marshal(user)
			m.logger.Info("RequireOwnerOrAdmin check", "user", string(userJSON))

			// Admin can access everything
			if user.Role == models.RoleAdmin {
				m.logger.Info("Admin access granted")
				next.ServeHTTP(w, r)
				return
			}

			// Regular user can only access their own resources
			resourceID := getResourceID(r)
			m.logger.Info("Checking resource ownership", "user_id", user.ID, "resource_id", resourceID)

			if resourceID == user.ID {
				m.logger.Info("Owner access granted")
				next.ServeHTTP(w, r)
				return
			}

			m.logger.Error("Access denied", "user_id", user.ID, "resource_id", resourceID)
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}
