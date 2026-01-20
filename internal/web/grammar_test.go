package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupGrammarTest(t *testing.T) (*Router, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	// Initialize grammar repositories and service
	contentRepo := repository.NewGrammarContentRepository(logger)
	publishRepo := repository.NewGrammarPublishRepository(db.GetConnection(), logger)
	attemptRepo := repository.NewGrammarAttemptRepository(db.GetConnection(), logger)
	grammarService := service.NewGrammarService(contentRepo, publishRepo, attemptRepo, logger)
	router.SetGrammarService(grammarService)

	cleanup := func() {
		db.Close()
	}

	return router, db, cleanup
}

func TestHandleLearningGrammarSubmitTest_BadRequest(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Invalid JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/tests/submit", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarSubmitTest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarSubmitTest_MissingFields(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	// Missing required fields
	body := map[string]interface{}{}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/tests/submit", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarSubmitTest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
}

func TestHandleLearningGrammarChapter_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/chapters/test-chapter", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarChapter(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleLearningGrammarChapterTest_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupGrammarTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/grammar/chapters/test-chapter/test", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()

	router.handleLearningGrammarChapterTest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
