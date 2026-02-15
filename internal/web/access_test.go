package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestHandleAccessMe(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: 12345,
		},
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Create admin user
	adminUser, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create regular user
	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create regular user: %v", err)
	}

	// Create category with permissions
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{
		Name: "Test Category",
	})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	err = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all", "word_sets.read"})
	if err != nil {
		t.Fatalf("Failed to set category permissions: %v", err)
	}

	// Assign category to regular user
	err = accessCategoryRepo.SetUserCategories(regularUser.ID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to assign category to user: %v", err)
	}

	tests := []struct {
		name           string
		userID         int64
		categories     []int64
		expectedStatus int
		checkPerms     bool
	}{
		{
			name:           "super admin",
			userID:         adminUser.ID,
			categories:     []int64{},
			expectedStatus: http.StatusOK,
			checkPerms:     true, // Should have all permissions
		},
		{
			name:           "user with categories",
			userID:         regularUser.ID,
			categories:     []int64{categoryID},
			expectedStatus: http.StatusOK,
			checkPerms:     true,
		},
		{
			name:           "user without categories",
			userID:         regularUser.ID,
			categories:     []int64{},
			expectedStatus: http.StatusOK,
			checkPerms:     false,
		},
		{
			name:           "no user context",
			userID:         0,
			categories:     []int64{},
			expectedStatus: http.StatusUnauthorized,
			checkPerms:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/access/me", nil)
			if tt.userID > 0 {
				ctx := context.WithValue(req.Context(), userIDKey, tt.userID)
				ctx = context.WithValue(ctx, userCategoriesKey, tt.categories)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			router.handleAccessMe(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handleAccessMe() status = %d, want %d", rr.Code, tt.expectedStatus)
				return
			}

			if tt.expectedStatus == http.StatusOK && tt.checkPerms {
				var response struct {
					Categories  []int64  `json:"categories"`
					Permissions []string `json:"permissions"`
				}
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if len(response.Categories) != len(tt.categories) {
					t.Errorf("Categories count = %d, want %d", len(response.Categories), len(tt.categories))
				}

				if tt.userID == adminUser.ID {
					// Super admin should have all permissions
					allPerms := AllPermissionStrings()
					if len(response.Permissions) < len(allPerms) {
						t.Errorf("Super admin should have all permissions, got %d, want at least %d", len(response.Permissions), len(allPerms))
					}
				} else if len(tt.categories) > 0 {
					// User with categories should have permissions from categories
					if len(response.Permissions) == 0 {
						t.Error("User with categories should have permissions")
					}
				}
			}
		})
	}
}
