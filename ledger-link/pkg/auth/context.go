package auth

import (
	"context"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
)

func GetUserIDFromContext(ctx context.Context) uint {
	userID, ok := ctx.Value(userIDKey).(uint)
	if !ok {
		return 0
	}
	return userID
}

func SetUserIDInContext(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}