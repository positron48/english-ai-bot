package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupWordSetsLearningTest(t *testing.T) (*Router, *database.DB, func()) {
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

	cleanup := func() {
		db.Close()
	}

	return router, db, cleanup
}

func setWordSetsLearningUserContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

func TestHandleLearningWordsCategories_Get(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/categories", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleLearningWordsCategories_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/categories", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsCategories(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleLearningWordsSets_Get(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleLearningWordsSets_GetWithCategory(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets?category_id=1", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleLearningWordsSets_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/sets", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSets(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleLearningWordsSetDetailOrStudy_InvalidID(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/invalid", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSetDetailOrStudy(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleLearningWordsSetDetail_NotFound(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/99999", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSetDetail(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleLearningWordsSetDetail_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/learning/words/sets/1", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSetDetail(rr, req)

	if rr.Code != http.StatusMethodNotAllowed && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 405 or 404, got %d", rr.Code)
	}
}

func TestHandleLearningWordsSetStudy_BadRequest(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/learning/words/sets/99999/study", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSetStudy(rr, req)

	// Returns 400 because word_card_id is required
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 400 or 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleLearningWordsSetStudyLearn_BadRequest(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/sets/99999/study/learn", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSetStudyLearn(rr, req)

	// Returns 400 because word_card_id is required
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 400 or 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleLearningWordsSetStudyKnow_BadRequest(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/learning/words/sets/99999/study/know", nil)
	req = setWordSetsLearningUserContext(req, 12345)
	rr := httptest.NewRecorder()

	router.handleLearningWordsSetStudyKnow(rr, req)

	// Returns 400 because word_card_id is required
	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 400 or 404, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetWordSetService(t *testing.T) {
	router, _, cleanup := setupWordSetsLearningTest(t)
	defer cleanup()

	service := router.getWordSetService()
	if service == nil {
		t.Error("Expected non-nil word set service")
	}
}
