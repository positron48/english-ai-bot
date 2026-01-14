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

func setUserIDInContextWordSets(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

func setupAdminWordSetsTest(t *testing.T) (*Router, *database.DB, func()) {
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

func TestHandleAdminWordSetCategories_Get(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/app/admin/word-set-categories", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSetCategories_PostInvalidBody(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/app/admin/word-set-categories", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetCategories(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminWordSets_Get(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/app/admin/word-sets", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSets_GetWithCategory(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/app/admin/word-sets?category_id=1", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSets_PostInvalidBody(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/app/admin/word-sets", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSets(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleAdminWordSetDetailOrSets_ListWordSets(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/app/admin/word-sets/", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetailOrSets(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSetDetailOrSets_NotFound(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/app/admin/word-sets/invalid", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetailOrSets(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminWordSetDetail_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/app/admin/word-sets/1", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetail(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminWordSetDetail_NotFound(t *testing.T) {
	router, _, cleanup := setupAdminWordSetsTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/app/admin/word-sets/99999", nil)
	req = setUserIDInContextWordSets(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminWordSetDetail(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d: %s", rr.Code, rr.Body.String())
	}
}
