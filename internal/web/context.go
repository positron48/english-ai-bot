package web

import (
	"context"
)

type contextKey string

const userIDKey contextKey = "user_id"
const userRoleKey contextKey = "user_role"
const userCategoriesKey contextKey = "user_categories"
const userPermissionsKey contextKey = "user_permissions"

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(ctx context.Context) int64 {
	if userID, ok := ctx.Value(userIDKey).(int64); ok {
		return userID
	}
	return 0
}

// getUserCategoriesFromContext extracts user categories from request context
func getUserCategoriesFromContext(ctx context.Context) []int64 {
	if categories, ok := ctx.Value(userCategoriesKey).([]int64); ok {
		return categories
	}
	return []int64{}
}

// getUserPermissionsFromContext extracts user permissions from request context
func getUserPermissionsFromContext(ctx context.Context) []string {
	if perms, ok := ctx.Value(userPermissionsKey).([]string); ok {
		return perms
	}
	return []string{}
}

