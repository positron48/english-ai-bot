package web

import (
	"context"
	"testing"
)

func TestGetUserIDFromContext(t *testing.T) {
	userID := int64(12345)
	ctx := context.WithValue(context.Background(), userIDKey, userID)

	result := getUserIDFromContext(ctx)
	if result != userID {
		t.Errorf("Expected UserID %d, got %d", userID, result)
	}
}

func TestGetUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	result := getUserIDFromContext(ctx)
	if result != 0 {
		t.Errorf("Expected UserID 0, got %d", result)
	}
}
