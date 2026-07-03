package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// TestHandleDashboardOverview_AggregatesParts verifies the aggregate endpoint returns a single
// JSON object folding together the successful sub-handler results (here: the always-available
// dashboard part), and that a wrong method is rejected.
func TestHandleDashboardOverview_AggregatesParts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupDashboardTestDB(t)

	user, err := userRepo.GetOrCreateUser(778899)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/overview/dashboard", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()

	router.handleDashboardOverview(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode aggregate response: %v", err)
	}
	// The dashboard part has no external dependencies and must always be present.
	dash, ok := response["dashboard"]
	if !ok {
		t.Fatalf("aggregate response missing 'dashboard' part: %v", response)
	}
	var dashObj map[string]interface{}
	if err := json.Unmarshal(dash, &dashObj); err != nil {
		t.Fatalf("dashboard part is not a JSON object: %v", err)
	}
	if _, ok := dashObj["due_count"]; !ok {
		t.Errorf("dashboard part should contain due_count, got %v", dashObj)
	}
}

// TestHandleProgressOverview_DashboardIncludesWordsAdded guards the progress screen's
// words-added chart: the dashboard part is fetched with sections=totals, which must still
// include the words_added_stats series (it cannot be rebuilt from canonical history).
func TestHandleProgressOverview_DashboardIncludesWordsAdded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupDashboardTestDB(t)

	user, err := userRepo.GetOrCreateUser(778900)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("GET", "/api/overview/progress", nil)
	req = req.WithContext(context.WithValue(req.Context(), userIDKey, user.ID))
	w := httptest.NewRecorder()

	router.handleProgressOverview(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode aggregate response: %v", err)
	}
	dash, ok := response["dashboard"]
	if !ok {
		t.Fatalf("aggregate response missing 'dashboard' part: %v", response)
	}
	var dashObj map[string]interface{}
	if err := json.Unmarshal(dash, &dashObj); err != nil {
		t.Fatalf("dashboard part is not a JSON object: %v", err)
	}
	if _, ok := dashObj["total_cards"]; !ok {
		t.Errorf("dashboard part should contain total_cards, got %v", dashObj)
	}
	if _, ok := dashObj["words_added_stats"]; !ok {
		t.Errorf("dashboard part (sections=totals) should contain words_added_stats, got %v", dashObj)
	}
	if _, ok := dashObj["weekly_stats"]; ok {
		t.Errorf("dashboard part (sections=totals) should not compute weekly_stats, got %v", dashObj)
	}
}

func TestHandleDashboardOverview_WrongMethod(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo := setupDashboardTestDB(t)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{JWTSecret: "test-secret", JWTTTLHours: 24, RefreshTTLHours: 720},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	router := NewRouter(logger, cfg, db, nil, nil, nil, nil)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")
	router.authMiddleware = authMiddleware

	req := httptest.NewRequest("POST", "/api/overview/dashboard", nil)
	w := httptest.NewRecorder()
	router.handleDashboardOverview(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}
