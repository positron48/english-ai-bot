package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
