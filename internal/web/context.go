package web

import (
	"context"
)

type contextKey string

const userIDKey contextKey = "user_id"

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(ctx context.Context) int64 {
	if userID, ok := ctx.Value(userIDKey).(int64); ok {
		return userID
	}
	return 0
}

