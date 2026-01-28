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

func setupAdminAppSettingsRouter(t *testing.T) (*Router, *database.DB, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	cleanup := func() {
		_ = db.Close()
	}

	return router, db, cleanup
}

func TestHandleAdminAppSettings_Get(t *testing.T) {
	router, _, cleanup := setupAdminAppSettingsRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/app-settings", nil)
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAdminAppSettings_PutUnauthorized(t *testing.T) {
	router, _, cleanup := setupAdminAppSettingsRouter(t)
	defer cleanup()

	payload := []byte(`{"hide_placement_test_button": true}`)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/app-settings", bytes.NewReader(payload))
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleAdminAppSettings_PutInvalidBody(t *testing.T) {
	router, _, cleanup := setupAdminAppSettingsRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/api/admin/app-settings", bytes.NewBufferString("invalid"))
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAdminAppSettings_PutSuccess(t *testing.T) {
	router, _, cleanup := setupAdminAppSettingsRouter(t)
	defer cleanup()

	payload, _ := json.Marshal(map[string]interface{}{"hide_placement_test_button": true})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/app-settings", bytes.NewReader(payload))
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleAdminAppSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/admin/app-settings", nil)
	getW := httptest.NewRecorder()
	router.handleAdminAppSettings(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getW.Code)
	}
}
