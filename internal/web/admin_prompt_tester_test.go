package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setUserIDInContextPromptTester(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

func setupPromptTesterTest(t *testing.T) (*Router, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"
	tempFile, err := os.CreateTemp("", "training-prompt-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp prompt file: %v", err)
	}
	if _, err := tempFile.WriteString("test training prompt"); err != nil {
		_ = tempFile.Close()
		t.Fatalf("Failed to write temp prompt file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp prompt file: %v", err)
	}
	cfg.Training.PromptFile = tempFile.Name()

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	cleanup := func() {
		_ = db.Close()
		_ = os.Remove(tempFile.Name())
	}

	return router, db, cleanup
}

func TestHandleAdminPromptTesterDefaultPrompts_Get(t *testing.T) {
	router, _, cleanup := setupPromptTesterTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/prompt-tester/default-prompts", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterDefaultPrompts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminPromptTesterDefaultPrompts_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupPromptTesterTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/default-prompts", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterDefaultPrompts(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminPromptTesterRun_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupPromptTesterTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/prompt-tester/run", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminPromptTesterRun_InvalidBody(t *testing.T) {
	router, _, cleanup := setupPromptTesterTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}
