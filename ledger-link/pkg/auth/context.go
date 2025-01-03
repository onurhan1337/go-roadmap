package auth

import (
	"context"
	"ledger-link/internal/models"
)

type contextKey string

const (
	userIDKey      contextKey = "user_id"
	userContextKey contextKey = "user"
	userRoleKey    contextKey = "user_role"
)

// GetUserFromContext retrieves the user from the context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}

// GetUserIDFromContext retrieves the user ID from the context
func GetUserIDFromContext(ctx context.Context) uint {
	userID, ok := ctx.Value(userIDKey).(uint)
	if !ok {
		return 0
	}
	return userID
}

// GetUserRoleFromContext retrieves the user role from the context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleKey).(string)
	return role, ok
}

// SetUserInContext sets the user in the context
func SetUserInContext(ctx context.Context, user *models.User) context.Context {
	ctx = context.WithValue(ctx, userContextKey, user)
	ctx = context.WithValue(ctx, userIDKey, user.ID)
	ctx = context.WithValue(ctx, userRoleKey, user.Role)
	return ctx
}
