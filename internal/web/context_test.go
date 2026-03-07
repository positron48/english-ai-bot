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

func TestGetUserCategoriesFromContext(t *testing.T) {
	categories := []int64{1, 2, 3}
	ctx := context.WithValue(context.Background(), userCategoriesKey, categories)

	result := getUserCategoriesFromContext(ctx)
	if len(result) != 3 || result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Expected categories [1,2,3], got %v", result)
	}
}

func TestGetUserCategoriesFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	result := getUserCategoriesFromContext(ctx)
	if len(result) != 0 {
		t.Errorf("Expected empty categories, got %v", result)
	}
}

func TestGetUserCategoriesFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), userCategoriesKey, "not-a-slice")

	result := getUserCategoriesFromContext(ctx)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for wrong type, got %v", result)
	}
}

func TestGetUserPermissionsFromContext(t *testing.T) {
	perms := []string{"words.read_all", "word_sets.read"}
	ctx := context.WithValue(context.Background(), userPermissionsKey, perms)

	result := getUserPermissionsFromContext(ctx)
	if len(result) != 2 || result[0] != "words.read_all" || result[1] != "word_sets.read" {
		t.Errorf("Expected permissions [words.read_all, word_sets.read], got %v", result)
	}
}

func TestGetUserPermissionsFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	result := getUserPermissionsFromContext(ctx)
	if len(result) != 0 {
		t.Errorf("Expected empty permissions, got %v", result)
	}
}

func TestGetUserPermissionsFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), userPermissionsKey, 123)

	result := getUserPermissionsFromContext(ctx)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for wrong type, got %v", result)
	}
}
