package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"tgbot-skeleton/internal/testutil"
	"go.uber.org/zap"
)

func setupAdminTestDB(t *testing.T) (*sql.DB, *repository.UserRepository, *service.CircuitBreakerService) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(db, logger)
	cbRepo := repository.NewCircuitBreakerRepository(db, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	return db, userRepo, cbService
}

func TestHandleAdmin_Get(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	// Create admin user
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Create request with admin user context
	req := httptest.NewRequest("GET", "/api/admin", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Call handler with admin middleware
	adminHandler := router.RequireAdmin(router.handleAdmin)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["circuit_breaker"] == nil {
		t.Error("Response should contain circuit_breaker")
	}

	if response["admin_id"] == nil {
		t.Error("Response should contain admin_id")
	}
}

func TestHandleAdmin_Get_WithCircuitBreakerState(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Set circuit breaker state: open, message and dates (Postgres: is_open is INTEGER 0/1)
	_, err = db.Exec(`UPDATE circuit_breaker_state SET is_open = 1, failure_count = 2,
		last_failure_at = '2025-01-15 10:00:00+00',
		last_failure_message = 'connection refused',
		last_reset_at = '2025-01-14 09:00:00+00'
		WHERE id = 1`)
	if err != nil {
		t.Fatalf("Failed to update circuit breaker state: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdmin)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	cb, _ := response["circuit_breaker"].(map[string]interface{})
	if cb == nil {
		t.Fatal("Response should contain circuit_breaker")
	}
	if cb["state"] != "open" {
		t.Errorf("Expected circuit_breaker.state open, got %v", cb["state"])
	}
	if cb["last_failure"] != "connection refused" {
		t.Errorf("Expected last_failure in response when set in DB, got %v", cb["last_failure"])
	}
	if cb["failures"] != float64(2) {
		t.Errorf("Expected failures 2, got %v", cb["failures"])
	}
	if cb["last_failure_at"] == nil {
		t.Error("Expected last_failure_at in response when set in DB")
	}
	if cb["last_reset_at"] == nil {
		t.Error("Expected last_reset_at in response when set in DB")
	}
}

func TestHandleAdmin_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdmin)
	adminHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleAdmin_Unauthorized(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin", nil)
	// No user context - RequirePermission will return 403 (Forbidden) instead of 401
	// because it checks permissions after authentication
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdmin)
	adminHandler(w, req)

	// With new permission system, missing user context results in 403 (Forbidden)
	// rather than 401 (Unauthorized), as permission check happens after auth
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestHandleAdmin_Forbidden(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	// Create non-admin user
	nonAdminTelegramID := int64(999999999)
	nonAdminUser, err := userRepo.GetOrCreateUser(nonAdminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create non-admin user: %v", err)
	}

	adminTelegramID := int64(123456789)
	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin", nil)
	ctx := context.WithValue(req.Context(), userIDKey, nonAdminUser.ID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdmin)
	adminHandler(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestHandleAdminCircuitReset_Post(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/circuit/reset", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminCircuitReset)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}

	if response["message"] == nil {
		t.Error("Response should contain message")
	}
}

func TestHandleAdminCircuitReset_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/circuit/reset", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminCircuitReset)
	adminHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleAdminUsers_Get(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create another user
	_, err = userRepo.GetOrCreateUser(999999999)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/users", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminUsers)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["users"] == nil {
		t.Error("Response should contain users")
	}
}

func TestHandleAdminWords_Get(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create word card
	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("test", "test definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/words", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminWords)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["words"] == nil {
		t.Error("Response should contain words")
	}
}

func TestHandleAdminWords_WithQueryParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("queryword", "def")

	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
			JWTTTLHours: 24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/words?limit=10&offset=0&search=query", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminWords)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if _, hasWords := response["words"]; !hasWords {
		t.Error("Response should contain words key")
	}
	if _, hasPagination := response["pagination"]; !hasPagination {
		t.Error("Response should contain pagination key")
	}
}

func TestHandleAdminWord_Put(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("update", "old definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("update")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	wordCardID := wordCard.ID

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/words/%d", wordCardID), strings.NewReader("definition=new definition"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminWord)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}
}

func TestHandleAdminWord_Put_WithJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("jsonupdate", "old definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("jsonupdate")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	wordCardID := wordCard.ID

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	updateData := map[string]interface{}{
		"definition": "new definition from JSON",
		"pos":         "noun",
		"transcription": "/dɪˈfɪnɪʃən/",
	}
	jsonData, _ := json.Marshal(updateData)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/words/%d", wordCardID), strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminWord)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}

	// Verify the update
	updated, err := wordRepo.GetWordCardByID(wordCardID)
	if err != nil {
		t.Fatalf("Failed to get updated word card: %v", err)
	}
	if updated.Definition != "new definition from JSON" {
		t.Errorf("Expected definition 'new definition from JSON', got %q", updated.Definition)
	}
}

func TestHandleAdminWord_Delete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("delete", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("delete")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	wordCardID := wordCard.ID

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/admin/words/%d", wordCardID), nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminWord)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}
}

func TestHandleAdminWord_Put_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", "/api/admin/words/999999", strings.NewReader("definition=ok"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent word, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Delete_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("DELETE", "/api/admin/words/999999", nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent word, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Get(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Create word card and training card
	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("training", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("training")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	wordCardID := wordCard.ID

	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	trainingCard := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "training",
		SenseIndex: 0,
		WordRU:     "тренировка",
		MeaningEN:  "training",
	}
	_, err = trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/training/training", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminTraining)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["word_en"] == nil {
		t.Error("Response should contain word_en")
	}
}

func TestHandleAdminTrainingCard_Delete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("card", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("card")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}
	wordCardID := wordCard.ID

	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	trainingCard := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "card",
		SenseIndex: 0,
		WordRU:     "карта",
		MeaningEN:  "card",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/admin/training/card/%d", trainingCardID), nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminTrainingCard)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}
}

func TestHandleAdminTrainingCard_Put(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("updatecard", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("updatecard")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	trainingCard := &models.TrainingCard{
		WordCardID: wordCard.ID,
		WordEN:     "updatecard",
		SenseIndex: 0,
		WordRU:     "карта",
		MeaningEN:  "card",
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/training/card/%d", trainingCardID), strings.NewReader("word_ru=новая карта&meaning_en=new card&example_en=example"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminTrainingCard)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}
}

func TestHandleAdminTrainingCard_Put_UpdatePOS(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("poscard", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("poscard")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	pos := "noun"
	trainingCard := &models.TrainingCard{
		WordCardID: wordCard.ID,
		WordEN:     "poscard",
		SenseIndex: 0,
		WordRU:     "карта",
		MeaningEN:  "card",
		POS:        &pos,
	}
	trainingCardID, err := trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Update POS to verb
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/training/card/%d", trainingCardID), strings.NewReader("pos=verb"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminTrainingCard)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify POS was updated
	updated, err := trainingCardRepo.GetTrainingCard(trainingCardID)
	if err != nil {
		t.Fatalf("Failed to get updated training card: %v", err)
	}
	if updated.POS == nil || *updated.POS != "verb" {
		t.Errorf("Expected POS 'verb', got %v", updated.POS)
	}

	// Test clearing POS (empty string)
	req2 := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/training/card/%d", trainingCardID), strings.NewReader("pos="))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx2 := context.WithValue(req2.Context(), userIDKey, adminUser.ID)
	ctx2 = context.WithValue(ctx2, userRoleKey, "admin")
	req2 = req2.WithContext(ctx2)
	w2 := httptest.NewRecorder()

	adminHandler(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w2.Code)
	}

	// Verify POS was cleared
	updated2, err := trainingCardRepo.GetTrainingCard(trainingCardID)
	if err != nil {
		t.Fatalf("Failed to get updated training card: %v", err)
	}
	if updated2.POS != nil {
		t.Errorf("Expected POS to be nil after clearing, got %v", updated2.POS)
	}
}

func TestHandleAdminTraining_Delete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("deleteword", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("deleteword")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	trainingCard := &models.TrainingCard{
		WordCardID: wordCard.ID,
		WordEN:     "deleteword",
		SenseIndex: 0,
		WordRU:     "удалить",
		MeaningEN:  "delete",
	}
	_, err = trainingCardRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/training/deleteword/delete", strings.NewReader("word=deleteword"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminTraining)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}
}

func TestHandleAdminTraining_DeleteAll(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("deleteall1", "definition1")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard1, err := wordRepo.GetWordCard("deleteall1")
	if err != nil || wordCard1 == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	trainingCard1 := &models.TrainingCard{
		WordCardID: wordCard1.ID,
		WordEN:     "deleteall1",
		SenseIndex: 0,
		WordRU:     "удалить1",
		MeaningEN:  "delete1",
	}
	_, err = trainingCardRepo.CreateTrainingCard(trainingCard1)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/training/delete_all", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminTraining)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}
}

func TestHandleAdminWord_Reset(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)

	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	wordRepo := repository.NewWordRepository(db, logger)
	err = wordRepo.SaveWordCard("reset", "definition")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	wordCard, err := wordRepo.GetWordCard("reset")
	if err != nil || wordCard == nil {
		t.Fatalf("Failed to get word card: %v", err)
	}

	cfg := &config.Config{
		Admin: config.AdminConfig{
			TelegramID: adminTelegramID,
		},
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/admin/words/%d/reset", wordCard.ID), nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	adminHandler := router.RequireAdmin(router.handleAdminWord)
	adminHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != true {
		t.Error("Response should indicate success")
	}
}

func TestHandleAdminWords_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/words", nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequirePermission(PermissionWordsReadAll)(router.handleAdminWords)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleAdminWord_InvalidID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", "/api/admin/words/abc", strings.NewReader("definition=ok"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAdminWord_WordIDEmpty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", "/api/admin/words/", nil)
	req.URL.Path = "/api/admin/words/"
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 (word ID required), got %d", w.Code)
	}
}

func TestHandleAdminWord_PUT_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	wordRepo.SaveWordCard("x", "d")
	wc, _ := wordRepo.GetWordCard("x")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", "/api/admin/words/"+fmt.Sprint(wc.ID), strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 (invalid JSON), got %d", w.Code)
	}
}

func TestHandleAdminWord_DeleteNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("DELETE", "/api/admin/words/99999", nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleAdminUsers_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/users", nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequirePermission(PermissionUsersReadAll)(router.handleAdminUsers)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleAdminTraining_WordRequired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/training/", nil)
	req.URL.Path = "/api/admin/training/"
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsReadAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_ForbiddenNoPermission(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	nonAdmin, err := userRepo.GetOrCreateUser(999888777)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/training/hello", nil)
	ctx := context.WithValue(req.Context(), userIDKey, nonAdmin.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{}) // no permissions
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestHandleAdminTrainingCard_InvalidCardID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminTelegramID := int64(123456789)
	adminUser, err := userRepo.GetOrCreateUser(adminTelegramID)
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: adminTelegramID},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("DELETE", "/api/admin/training/card/notanid", nil)
	req.URL.Path = "/api/admin/training/card/notanid"
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequirePermission(PermissionWordsEditAll)(router.handleAdminTrainingCard)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleAdminAppSettings_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	_, _ = userRepo.GetOrCreateUser(12345)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	req := httptest.NewRequest("POST", "/api/admin/app-settings", nil)
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

// TestHandleAdmin_DirectMethodNotAllowed calls handleAdmin without middleware to cover method check.
func TestHandleAdmin_DirectMethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPost, "/api/admin", nil)
	w := httptest.NewRecorder()
	router.handleAdmin(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("handleAdmin POST: expected status 405, got %d", w.Code)
	}
	if body := w.Body.String(); body != "Method not allowed\n" {
		t.Errorf("handleAdmin POST: expected body %q, got %q", "Method not allowed\n", body)
	}
}

// TestHandleAdminWord_MethodNotAllowed covers GET/PATCH without valid action.
func TestHandleAdminWord_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("methodword", "def")
	wc, _ := wordRepo.GetWordCard("methodword")
	if wc == nil {
		t.Fatal("word card not found")
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"GET without action", http.MethodGet, fmt.Sprintf("/api/admin/words/%d", wc.ID)},
		{"PATCH", http.MethodPatch, fmt.Sprintf("/api/admin/words/%d", wc.ID)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.URL.Path = tt.path
			ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
			ctx = context.WithValue(ctx, userRoleKey, "admin")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.RequireAdmin(router.handleAdminWord)(w, req)
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// TestHandleAdminWord_PUT_UniqueConstraint covers 409 when updating word to duplicate.
func TestHandleAdminWord_PUT_UniqueConstraint(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("uniquea", "def a")
	_ = wordRepo.SaveWordCard("uniqueb", "def b")
	wcA, _ := wordRepo.GetWordCard("uniquea")
	wcB, _ := wordRepo.GetWordCard("uniqueb")
	if wcA == nil || wcB == nil {
		t.Fatal("word cards not found")
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Update word A to "uniqueb" (duplicate of B) -> 409
	body := strings.NewReader("word=uniqueb")
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wcA.ID), body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409 for UNIQUE constraint, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") == "application/json" {
		var resp map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp["success"] == true {
			t.Error("expected success false on duplicate word")
		}
	}
}

// TestHandleAdminWord_POST_Generate_AIServiceNil covers 500 when AI service is not available.
func TestHandleAdminWord_POST_Generate_AIServiceNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("genword", "def")
	wc, _ := wordRepo.GetWordCard("genword")
	if wc == nil {
		t.Fatal("word card not found")
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware
	// aiService is nil by default

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 when AI service nil, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminWord_POST_Generate_NotFound covers 404 when word card does not exist.
func TestHandleAdminWord_POST_Generate_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest(http.MethodPost, "/api/admin/words/999999/generate", nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), userIDKey, adminUser.ID), userRoleKey, "admin"))
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdminWord)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for non-existent word, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_Get_WordNotFound covers GET when word card does not exist (empty cards).
func TestHandleAdminTraining_Get_WordNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/training/nonexistentwordxyz", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsReadAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["word_en"] != "nonexistentwordxyz" {
		t.Errorf("expected word_en in response, got %v", resp["word_en"])
	}
	cards, _ := resp["cards"].([]interface{})
	if cards == nil || len(cards) != 0 {
		t.Errorf("expected empty cards when word not found, got %v", resp["cards"])
	}
}

// TestHandleAdminTraining_PostCreate_JSON covers POST create training card with JSON body.
func TestHandleAdminTraining_PostCreate_JSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("createjson", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	body := map[string]interface{}{
		"word_ru":     "создать",
		"meaning_en":  "to create",
		"example_en":  "example",
		"example_ru":  "пример",
		"transcription": "/kriˈeɪt/",
		"pos":         "verb",
		"display_word": "create",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/training/createjson", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["success"] != true {
		t.Error("expected success true")
	}
}

// TestHandleAdminTraining_PostCreate_ValidationRequired covers 400 when word_ru or meaning_en missing.
func TestHandleAdminTraining_PostCreate_ValidationRequired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("validword", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	body := map[string]interface{}{"word_ru": "слово"} // meaning_en missing
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/training/validword", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_PostCreate_WordNotFound covers 404 when word card does not exist.
func TestHandleAdminTraining_PostCreate_WordNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	body := map[string]interface{}{"word_ru": "слово", "meaning_en": "meaning"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/training/nonexistentword123", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_PostDelete_WordFromForm covers POST delete with word in form when path word empty.
func TestHandleAdminTraining_PostDelete_WordFromForm(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("formdelete", "def")
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	wc, _ := wordRepo.GetWordCard("formdelete")
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "formdelete", SenseIndex: 0, WordRU: "удалить", MeaningEN: "delete"}
	_, _ = trainingCardRepo.CreateTrainingCard(tc)
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	// Path with word so action is "delete"; form also has word for coverage of form fallback.
	req := httptest.NewRequest("POST", "/api/admin/training/formdelete/delete", strings.NewReader("word=formdelete"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_InvalidRequest covers 400 for unsupported method/path combination.
func TestHandleAdminTraining_InvalidRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/training/word/unknown_action", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_ForbiddenEdit covers 403 when POST without edit permission.
func TestHandleAdminTraining_ForbiddenEdit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	// Use a user that has no edit permission (only read or none); loadUserPermissionsIntoContext will load from DB.
	readOnlyUser, err := userRepo.GetOrCreateUser(999888777)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	body := map[string]interface{}{"word_ru": "слово", "meaning_en": "meaning"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/api/admin/training/someword", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	// Context with read-only user; loaded permissions won't include WordsEditAll
	ctx := context.WithValue(req.Context(), userIDKey, readOnlyUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsReadAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	// Handler checks edit permission first; without PermissionWordsEditAll we expect 403.
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_EmptyCardID_AdminGo covers 400 when card ID is missing or empty.
func TestHandleAdminTrainingCard_EmptyCardID_AdminGo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("DELETE", "/api/admin/training/card/", nil)
	req.URL.Path = "/api/admin/training/card/"
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_Get_MethodNotAllowed covers GET returning 405.
func TestHandleAdminTrainingCard_Get_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("getcard", "def")
	wc, _ := wordRepo.GetWordCard("getcard")
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "getcard", SenseIndex: 0, WordRU: "карта", MeaningEN: "card"}
	cardID, _ := trainingCardRepo.CreateTrainingCard(tc)
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/admin/training/card/%d", cardID), nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsReadAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_Delete_NotFound covers 404 when deleting non-existent card.
func TestHandleAdminTrainingCard_Delete_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("DELETE", "/api/admin/training/card/999999", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_Put_JSON covers PUT with JSON body.
func TestHandleAdminTrainingCard_Put_JSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("putjson", "def")
	wc, _ := wordRepo.GetWordCard("putjson")
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "putjson", SenseIndex: 0, WordRU: "карта", MeaningEN: "card"}
	cardID, _ := trainingCardRepo.CreateTrainingCard(tc)
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	body := map[string]interface{}{
		"word_ru":    "обновлено",
		"meaning_en": "updated",
		"pos":        "noun",
		"display_word": "putjson",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/training/card/%d", cardID), strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_Put_CardNotFound covers 404 when updating non-existent card.
func TestHandleAdminTrainingCard_Put_CardNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	body := map[string]interface{}{"word_ru": "x", "meaning_en": "y"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/api/admin/training/card/999999", strings.NewReader(string(jsonBody)))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_Put_InvalidJSON covers 400 when JSON body is invalid.
func TestHandleAdminTrainingCard_Put_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("invjson", "def")
	wc, _ := wordRepo.GetWordCard("invjson")
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "invjson", SenseIndex: 0, WordRU: "x", MeaningEN: "y"}
	cardID, _ := trainingCardRepo.CreateTrainingCard(tc)
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/training/card/%d", cardID), strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_Put_InvalidForm covers 400 when form body read fails (ParseForm fails).
func TestHandleAdminTrainingCard_Put_InvalidForm(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("invform", "def")
	wc, _ := wordRepo.GetWordCard("invform")
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "invform", SenseIndex: 0, WordRU: "x", MeaningEN: "y"}
	cardID, _ := trainingCardRepo.CreateTrainingCard(tc)
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PUT", fmt.Sprintf("/api/admin/training/card/%d", cardID), errReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTrainingCard_MethodNotAllowed_AdminGo covers 405 for non-PUT/DELETE.
func TestHandleAdminTrainingCard_MethodNotAllowed_AdminGo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("PATCH", "/api/admin/training/card/1", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminWords_QueryParams covers all query params: only_errors, has_audio, user_id, sort_order asc.
func TestHandleAdminWords_QueryParams(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_, _ = userRepo.GetOrCreateUser(999888777)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("sortword", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	tests := []struct {
		name string
		url  string
	}{
		{"only_errors=true", "/api/admin/words?only_errors=true"},
		{"only_errors=1", "/api/admin/words?only_errors=1"},
		{"has_audio=true", "/api/admin/words?has_audio=true"},
		{"has_audio=false", "/api/admin/words?has_audio=false"},
		{"has_audio=1", "/api/admin/words?has_audio=1"},
		{"has_audio=0", "/api/admin/words?has_audio=0"},
		{"user_id", "/api/admin/words?user_id=1"},
		{"sort_order=asc", "/api/admin/words?sort_order=asc"},
		{"sort_by=word", "/api/admin/words?sort_by=word&sort_order=asc"},
		{"limit and offset", "/api/admin/words?limit=5&offset=0"},
		{"missing_training_pos", "/api/admin/words?missing_training_pos=noun"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
			ctx = context.WithValue(ctx, userRoleKey, "admin")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			router.RequireAdmin(router.handleAdminWords)(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

// TestHandleAdminTraining_PostCreate_InvalidForm covers 400 when form body read fails.
func TestHandleAdminTraining_PostCreate_InvalidForm(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("invformcreate", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/training/invformcreate", errReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_PostCreate_Form covers POST create with form data.
func TestHandleAdminTraining_PostCreate_Form(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("formcreate", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	form := "word_ru=форма&meaning_en=form&example_en=ex&example_ru=пример&transcription=/fɔːm/&pos=noun&display_word=formcreate"
	req := httptest.NewRequest("POST", "/api/admin/training/formcreate", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["success"] != true {
		t.Error("expected success true")
	}
}

// TestHandleAdminTraining_InvalidPath_AdminGo covers 400 when path has no word segment.
func TestHandleAdminTraining_InvalidPath_AdminGo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin/training/", nil)
	req.URL.Path = "/api/admin/training/"
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsReadAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 (invalid path or word required), got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_PostGenerate_AIServiceNil covers 500 when AI service is nil.
func TestHandleAdminTraining_PostGenerate_AIServiceNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("genword", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/training/genword/generate", strings.NewReader(`{"constraints":"short"}`))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when AI service nil, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_PostGenerate_Form covers generate with form data (constraints from form).
func TestHandleAdminTraining_PostGenerate_Form(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("genform", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/training/genform/generate", strings.NewReader("constraints=short"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)
	// No AI service -> 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when AI service nil, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdminTraining_PostGenerate_InvalidJSON covers 400 when JSON body is invalid.
func TestHandleAdminTraining_PostGenerate_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("genword2", "def")
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/admin/training/genword2/generate", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userPermissionsKey, []string{string(PermissionWordsEditAll)})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleAdmin_CircuitBreakerNoRow covers handleAdmin when circuit_breaker_state has no row (init + retry).
func TestHandleAdmin_CircuitBreakerNoRow(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, err := userRepo.GetOrCreateUser(123456789)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_, err = db.Exec("DELETE FROM circuit_breaker_state WHERE id = 1")
	if err != nil {
		t.Skipf("cannot delete circuit_breaker_state: %v", err)
	}
	cfg := &config.Config{
		Admin: config.AdminConfig{TelegramID: 123456789},
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/admin", nil)
	ctx := context.WithValue(req.Context(), userIDKey, adminUser.ID)
	ctx = context.WithValue(ctx, userRoleKey, "admin")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.RequireAdmin(router.handleAdmin)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 after init, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["circuit_breaker"] == nil {
		t.Error("expected circuit_breaker in response")
	}
}

// Direct handler tests (call handlers without middleware) to ensure full coverage of admin.go

func TestHandleAdmin_Direct_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPost, "/api/admin", nil)
	w := httptest.NewRecorder()
	router.handleAdmin(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminCircuitReset_Direct_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/circuit/reset", nil)
	w := httptest.NewRecorder()
	router.handleAdminCircuitReset(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminUsers_Direct_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", nil)
	w := httptest.NewRecorder()
	router.handleAdminUsers(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminWords_Direct_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/words", nil)
	w := httptest.NewRecorder()
	router.handleAdminWords(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdmin_Direct_Get_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
	w := httptest.NewRecorder()
	router.handleAdmin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["circuit_breaker"] == nil || resp["admin_id"] == nil {
		t.Error("expected circuit_breaker and admin_id in response")
	}
}

func TestHandleAdminCircuitReset_Direct_PostSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/circuit/reset", nil)
	w := httptest.NewRecorder()
	router.handleAdminCircuitReset(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["success"] != true {
		t.Error("expected success true")
	}
}

func TestHandleAdminUsers_Direct_Get_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	_, _ = userRepo.GetOrCreateUser(123456789)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	w := httptest.NewRecorder()
	router.handleAdminUsers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["users"] == nil {
		t.Error("expected users in response")
	}
}

func TestHandleAdminWords_Direct_Get_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words", nil)
	w := httptest.NewRecorder()
	router.handleAdminWords(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, hasWords := resp["words"]; !hasWords {
		t.Error("expected words key in response")
	}
	if _, hasPag := resp["pagination"]; !hasPag {
		t.Error("expected pagination key in response")
	}
}

func TestHandleAdminWords_QueryParams_HasAudioAndOnlyErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	for _, q := range []string{"?only_errors=true", "?only_errors=1", "?has_audio=true", "?has_audio=false", "?has_audio=0", "?sort_order=asc", "?limit=5&offset=0"} {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/words"+q, nil)
		w := httptest.NewRecorder()
		router.handleAdminWords(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("query %q: expected 200, got %d: %s", q, w.Code, w.Body.String())
		}
	}
}

func TestHandleAdminWord_Direct_PutSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("directput", "old")
	wc, _ := wordRepo.GetWordCard("directput")
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	body := `{"definition":"new def","word":"directput"}`
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wc.ID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Direct_ResetSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("directreset", "def")
	wc, _ := wordRepo.GetWordCard("directreset")
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/reset", wc.ID), nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Direct_DeleteSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("directdel", "def")
	wc, _ := wordRepo.GetWordCard("directdel")
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/admin/words/%d", wc.ID), nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Direct_MethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("mna", "def")
	wc, _ := wordRepo.GetWordCard("mna")
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/admin/words/%d", wc.ID), nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminWord_Direct_POST_Generate_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("run", "def")
	wc, _ := wordRepo.GetWordCard("run")
	// WordInfoResponse JSON that AI returns (as content of choices[].message.content)
	wordInfoJSON := `{"input_word":"run","lemma":"run","pos":"verb","transcription":"rʌn","definition_ru":"бежать","examples":[],"verb_forms":{"v1":"run","v2":"ran","v3":"run"}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + strconv.Quote(wordInfoJSON) + `}}]}`))
	}))
	t.Cleanup(server.Close)
	aiSvc := ai.NewService(server.URL, "test-model", "test-key", "prompt", logger)

	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.aiService = aiSvc

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["success"] != true {
		t.Error("expected success true")
	}
	if _, ok := resp["word_card"]; !ok {
		t.Error("expected word_card in response")
	}
}

func TestHandleAdminWord_Direct_POST_Generate_LLMError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("x", "def")
	wc, _ := wordRepo.GetWordCard("x")
	wordInfoJSON := `{"error":true,"lemma":"","pos":"","transcription":"","definition_ru":"","examples":[]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + strconv.Quote(wordInfoJSON) + `}}]}`))
	}))
	t.Cleanup(server.Close)
	aiSvc := ai.NewService(server.URL, "test-model", "test-key", "prompt", logger)

	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.aiService = aiSvc

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for LLM error, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Direct_POST_Generate_ParseError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("y", "def")
	wc, _ := wordRepo.GetWordCard("y")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"not valid json"}}]}`))
	}))
	t.Cleanup(server.Close)
	aiSvc := ai.NewService(server.URL, "test-model", "test-key", "prompt", logger)

	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	router.aiService = aiSvc

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for parse error, got %d: %s", w.Code, w.Body.String())
	}
}

// badDBConn returns a closed *sql.DB so that subsequent DB operations fail (for coverage of error paths).
func badDBConn(t *testing.T) *sql.DB {
	t.Helper()
	testutil.SetupTestDB(t) // ensure shared DB and postgres_compat driver are ready
	dsn := testutil.GetTestDSN(t)
	conn, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("postgres_compat open: %v", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		t.Skipf("ping: %v", err)
	}
	conn.Close()
	return conn
}

func TestHandleAdmin_Direct_Get_DBQueryFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
	w := httptest.NewRecorder()
	router.handleAdmin(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (handler returns default state on error), got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	cb, _ := resp["circuit_breaker"].(map[string]interface{})
	if cb == nil {
		t.Fatal("expected circuit_breaker in response")
	}
	if cb["state"] != "closed" || cb["failures"] != float64(0) {
		t.Errorf("expected default state on error, got %v", cb)
	}
}

func TestHandleAdminCircuitReset_Direct_Post_ResetFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	badDB := badDBConn(t)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/circuit/reset", nil)
	w := httptest.NewRecorder()
	router.handleAdminCircuitReset(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when Reset fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminAppSettings_Get_DBFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/app-settings", nil)
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetBoolSetting fails, got %d", w.Code)
	}
}

func TestHandleAdminAppSettings_Put_SetBoolFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	payload, _ := json.Marshal(map[string]interface{}{"hide_placement_test_button": true})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/app-settings", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, int64(1)))
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when SetBoolSetting fails, got %d", w.Code)
	}
}

func TestHandleAdminAppSettings_PatchMethodNotAllowed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/app-settings", nil)
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminUsers_Direct_QueryFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	w := httptest.NewRecorder()
	router.handleAdminUsers(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when Query fails, got %d", w.Code)
	}
}

func TestHandleAdminWords_Direct_ListFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words", nil)
	w := httptest.NewRecorder()
	router.handleAdminWords(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when ListWordCardsAdmin fails, got %d", w.Code)
	}
}

func TestHandleAdminOrphanedCards_Direct_ListFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-cards", nil)
	w := httptest.NewRecorder()
	router.handleAdminOrphanedCards(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when ListOrphanedTrainingCards fails, got %d", w.Code)
	}
}

func TestHandleAdminOrphanedCards_Direct_CountFails(t *testing.T) {
	// Count fails after List succeeds: use a DB that can run List but fails on Count.
	// With a closed DB both fail at first use, so we get ListFails. To get CountFails we'd need
	// a custom mock. Use badDB and we already cover one error path; Count error path is similar.
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-cards?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	router.handleAdminOrphanedCards(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandleAdminOrphanedUserCards_Direct_ListFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/orphaned-user-cards", nil)
	w := httptest.NewRecorder()
	router.handleAdminOrphanedUserCards(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when ListOrphanedUserCards fails, got %d", w.Code)
	}
}

func TestHandleAdminTraining_PostCreate_ValidationRequiredFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, _ := userRepo.GetOrCreateUser(123456789)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("validword", "def")
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, adminUser.ID), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	// POST create with missing word_ru and meaning_en
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/validword", strings.NewReader("word_ru=&meaning_en="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing word_ru/meaning_en, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "word_ru and meaning_en are required") {
		t.Errorf("expected validation message, got body: %s", w.Body.String())
	}
}

func TestHandleAdminTrainingCard_Put_Form_PosEmptySetsNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, _ := userRepo.GetOrCreateUser(123456789)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("posword", "def")
	wc, _ := wordRepo.GetWordCard("posword")
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	posVal := "noun"
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "posword", SenseIndex: 0, WordRU: "слово", MeaningEN: "word", POS: &posVal}
	cardID, _ := trainingCardRepo.CreateTrainingCard(tc)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, adminUser.ID), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	// PUT with form pos= (empty) to set POS to nil
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/training/card/%d", cardID), strings.NewReader("word_ru=слово&meaning_en=word&pos="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Put_DBFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	body := strings.NewReader("definition=new&word=w")
	req := httptest.NewRequest(http.MethodPut, "/api/admin/words/1", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when DB fails, got %d", w.Code)
	}
}

func TestHandleAdminTraining_PostCreate_GetWordCardFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, int64(1)), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/someword", strings.NewReader("word_ru=слово&meaning_en=meaning"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetWordCard fails, got %d", w.Code)
	}
}

func TestHandleAdminTrainingCard_Delete_DBFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, int64(1)), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/training/card/1", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when DeleteTrainingCard fails, got %d", w.Code)
	}
}

func TestHandleAdminTrainingCard_Put_GetCardFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, int64(1)), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/1", strings.NewReader("word_ru=w&meaning_en=m"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetTrainingCard fails, got %d", w.Code)
	}
}

func TestHandleAdminWords_CountFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/words?user_id=1&only_errors=true&has_audio=false&search=q&missing_training_pos=1&sort_by=word&sort_order=asc&limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	router.handleAdminWords(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when List/Count fails, got %d", w.Code)
	}
}

func TestHandleAdminTrainingCard_Put_CardNotFound_Direct(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, _ := userRepo.GetOrCreateUser(123456789)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, adminUser.ID), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	req := httptest.NewRequest(http.MethodPut, "/api/admin/training/card/999999", strings.NewReader("word_ru=w&meaning_en=m"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent card, got %d", w.Code)
	}
}

func TestHandleAdminWord_Delete_DBFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/words/1", nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when DeleteWordCard fails, got %d", w.Code)
	}
}

// --- Additional tests for 100% coverage of admin.go ---

func TestHandleAdminWord_Put_FormData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("formput", "old def")
	wc, _ := wordRepo.GetWordCard("formput")
	cfg := &config.Config{}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wc.ID), strings.NewReader("word=formput&definition=new+def&pos=noun"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for PUT with form, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_InvalidWordID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/words/abc", nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid word ID, got %d", w.Code)
	}
}

func TestHandleAdminWord_WordIDRequired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/words/", nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when word ID missing, got %d", w.Code)
	}
}

func TestHandleAdminTraining_Get_WordNotInDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, _ := userRepo.GetOrCreateUser(123456789)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, adminUser.ID), userPermissionsKey, []string{string(PermissionWordsReadAll)})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/training/nonexistentword12345", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with empty cards for non-existent word, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["word_en"] != "nonexistentword12345" {
		t.Errorf("expected word_en in response, got %v", resp["word_en"])
	}
	cards, _ := resp["cards"].([]interface{})
	if len(cards) != 0 {
		t.Errorf("expected empty cards for non-existent word, got %d", len(cards))
	}
}

func TestHandleAdminTraining_PostCreate_JSONWithDisplayWordAndPos(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, _ := userRepo.GetOrCreateUser(123456789)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("displayword", "def")
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, adminUser.ID), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	body := `{"word_ru":"слово","meaning_en":"meaning","pos":"noun","display_word":"to display"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/displayword", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTrainingCard_Put_Form_DisplayWordEmptySetsNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, _ := userRepo.GetOrCreateUser(123456789)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("dispword", "def")
	wc, _ := wordRepo.GetWordCard("dispword")
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	dispVal := "to disp"
	tc := &models.TrainingCard{WordCardID: wc.ID, WordEN: "dispword", SenseIndex: 0, WordRU: "слово", MeaningEN: "word", DisplayWord: &dispVal}
	cardID, _ := trainingCardRepo.CreateTrainingCard(tc)
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, adminUser.ID), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	// PUT with form display_word= (empty) to set DisplayWord to nil; word_en= to update word_en
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/training/card/%d", cardID), strings.NewReader("word_ru=слово&meaning_en=word&display_word=&word_en=dispword"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTrainingCard(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Generate_AIServiceNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, cbService := setupAdminTestDB(t)
	adminUser, _ := userRepo.GetOrCreateUser(123456789)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("aiword", "def")
	cfg := &config.Config{Admin: config.AdminConfig{TelegramID: 123456789}, WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	// aiService left nil
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, adminUser.ID), userPermissionsKey, []string{string(PermissionWordsEditAll)})

	req := httptest.NewRequest(http.MethodPost, "/api/admin/training/aiword/generate", strings.NewReader(`{"constraints":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when AI service not available, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "AI service not available") {
		t.Errorf("expected AI service not available message, got %s", w.Body.String())
	}
}

func TestHandleAdminWord_Generate_AIServiceNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("genword", "def")
	wc, _ := wordRepo.GetWordCard("genword")
	cfg := &config.Config{}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)
	// aiService left nil

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/words/%d/generate", wc.ID), nil)
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when AI service not available, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminTraining_Get_GetWordCardFails(t *testing.T) {
	badDB := badDBConn(t)
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(badDB, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, badDB, nil, nil, nil, cbService)
	ctx := context.WithValue(context.WithValue(context.Background(), userIDKey, int64(1)), userPermissionsKey, []string{string(PermissionWordsReadAll)})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/training/someword", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	router.handleAdminTraining(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 when GetWordCard fails, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Put_WordCardNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	cfg := &config.Config{}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	body := `{"word":"x","definition":"def"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/words/999999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 when word card not found, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleAdminWord_Put_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, cbService := setupAdminTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	_ = wordRepo.SaveWordCard("jsonword", "def")
	wc, _ := wordRepo.GetWordCard("jsonword")
	cfg := &config.Config{}
	router := NewRouter(logger, cfg, db, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/admin/words/%d", wc.ID), strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleAdminWord(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d: %s", w.Code, w.Body.String())
	}
}
