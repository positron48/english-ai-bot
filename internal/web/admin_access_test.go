package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupAdminAccessTest(t *testing.T) (*Router, *database.DB, *zap.Logger, int64, func()) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: 12345,
		},
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	// Create super admin user
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	adminUser, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	cleanup := func() {
		db.Close()
	}

	return router, db, logger, adminUser.ID, cleanup
}

func setAdminContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	ctx = context.WithValue(ctx, userCategoriesKey, []int64{})
	return req.WithContext(ctx)
}

func TestHandleAdminAccessAvailablePermissions(t *testing.T) {
	router, _, _, adminUserID, cleanup := setupAdminAccessTest(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/admin/access/available-permissions", nil)
	req = setAdminContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminAccessAvailablePermissions(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response struct {
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Permissions) == 0 {
		t.Error("Expected at least one permission")
	}

	// Check that all expected permissions are present
	expectedPerms := AllPermissionStrings()
	for _, expected := range expectedPerms {
		found := false
		for _, perm := range response.Permissions {
			if perm == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Permission %s not found in response", expected)
		}
	}
}

func TestHandleAdminAccessCategories_Get(t *testing.T) {
	router, db, logger, adminUserID, cleanup := setupAdminAccessTest(t)
	defer cleanup()

	// Create a test category
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{
		Name: "Test Category",
	})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	err = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all"})
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/admin/access/categories", nil)
	req = setAdminContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminAccessCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response struct {
		Categories []struct {
			ID          int64    `json:"id"`
			Name        string   `json:"name"`
			Permissions []string `json:"permissions"`
		} `json:"categories"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Categories) == 0 {
		t.Error("Expected at least one category")
	}

	found := false
	for _, cat := range response.Categories {
		if cat.ID == categoryID {
			found = true
			if len(cat.Permissions) == 0 {
				t.Error("Category should have permissions")
			}
			break
		}
	}
	if !found {
		t.Error("Created category not found in response")
	}
}

func TestHandleAdminAccessCategories_Post(t *testing.T) {
	router, _, _, adminUserID, cleanup := setupAdminAccessTest(t)
	defer cleanup()

	categoryData := map[string]interface{}{
		"name":        "New Category",
		"description": "Test description",
	}
	body, _ := json.Marshal(categoryData)

	req := httptest.NewRequest("POST", "/api/admin/access/categories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminAccessCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response struct {
		Success bool  `json:"success"`
		ID      int64 `json:"id"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success=true")
	}
	if response.ID == 0 {
		t.Error("Expected non-zero category ID")
	}
}

func TestHandleAdminAccessCategory_Get(t *testing.T) {
	router, db, _, adminUserID, cleanup := setupAdminAccessTest(t)
	defer cleanup()

	// Create a test category
	logger, _ := zap.NewDevelopment()
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{
		Name: "Test Category",
	})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	err = accessCategoryRepo.SetCategoryPermissions(categoryID, []string{"words.read_all"})
	if err != nil {
		t.Fatalf("Failed to set permissions: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/admin/access/categories/1", nil)
	req = setAdminContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminAccessCategoryByID(rr, req, categoryID)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response struct {
		Category struct {
			ID          int64    `json:"id"`
			Name        string   `json:"name"`
			Permissions []string `json:"permissions"`
		} `json:"category"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Category.ID != categoryID {
		t.Errorf("Expected category ID %d, got %d", categoryID, response.Category.ID)
	}
	if len(response.Category.Permissions) == 0 {
		t.Error("Category should have permissions")
	}
}

func TestHandleAdminAccessCategoryPermissions_Put(t *testing.T) {
	router, db, logger, adminUserID, cleanup := setupAdminAccessTest(t)
	defer cleanup()

	// Create a test category
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{
		Name: "Test Category",
	})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	permissionsData := map[string]interface{}{
		"permissions": []string{"words.read_all", "words.edit_all"},
	}
	body, _ := json.Marshal(permissionsData)

	req := httptest.NewRequest("PUT", "/api/admin/access/categories/1/permissions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminAccessCategoryPermissions(rr, req, categoryID)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify permissions were set
	perms, err := accessCategoryRepo.GetCategoryPermissions(categoryID)
	if err != nil {
		t.Fatalf("Failed to get category permissions: %v", err)
	}
	if len(perms) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(perms))
	}
}

func TestHandleAdminAccessUsers_Put(t *testing.T) {
	router, db, logger, adminUserID, cleanup := setupAdminAccessTest(t)
	defer cleanup()

	// Create a test category
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{
		Name: "Test Category",
	})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Create a regular user
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	categoriesData := map[string]interface{}{
		"category_ids": []int64{categoryID},
	}
	body, _ := json.Marshal(categoriesData)

	// Use regularUser.ID in the URL path
	userIDStr := fmt.Sprintf("%d", regularUser.ID)
	req := httptest.NewRequest("PUT", "/api/admin/access/users/"+userIDStr, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setAdminContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminAccessUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify categories were assigned (use regularUser.ID from DB)
	userCats, err := accessCategoryRepo.GetUserCategories(regularUser.ID)
	if err != nil {
		t.Fatalf("Failed to get user categories: %v", err)
	}
	if len(userCats) != 1 || userCats[0] != categoryID {
		t.Errorf("Expected user to have category %d, got %v", categoryID, userCats)
	}
}

func TestHandleAdminAccessUsers_Get(t *testing.T) {
	router, db, logger, adminUserID, cleanup := setupAdminAccessTest(t)
	defer cleanup()

	// Create a test category
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db.GetConnection(), logger)
	categoryID, err := accessCategoryRepo.CreateCategory(&models.UserAccessCategory{
		Name: "Test Category",
	})
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Create a regular user
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	regularUser, err := userRepo.GetOrCreateUser(99999)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Assign category to user
	err = accessCategoryRepo.SetUserCategories(regularUser.ID, []int64{categoryID})
	if err != nil {
		t.Fatalf("Failed to assign category: %v", err)
	}

	// Use regularUser.ID in the URL path
	userIDStr := fmt.Sprintf("%d", regularUser.ID)
	req := httptest.NewRequest("GET", "/api/admin/access/users/"+userIDStr, nil)
	req = setAdminContext(req, adminUserID)
	rr := httptest.NewRecorder()

	router.handleAdminAccessUsers(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response struct {
		UserID    int64   `json:"user_id"`
		Categories []int64 `json:"categories"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Categories) != 1 || response.Categories[0] != categoryID {
		t.Errorf("Expected categories [%d], got %v", categoryID, response.Categories)
	}
}
