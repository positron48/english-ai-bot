package web

import (
	"context"
)

type contextKey string

const userIDKey contextKey = "user_id"
const userRoleKey contextKey = "user_role"

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(ctx context.Context) int64 {
	if userID, ok := ctx.Value(userIDKey).(int64); ok {
		return userID
	}
	return 0
}

// getUserRoleFromContext extracts user role from request context
func getUserRoleFromContext(ctx context.Context) string {
	if role, ok := ctx.Value(userRoleKey).(string); ok {
		return role
	}
	return "user" // Default to "user" if not set
}

