package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupDashboardRouterFull(t *testing.T) (*Router, *database.DB, func()) {
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

func TestHandleDashboard_WithData(t *testing.T) {
	router, db, cleanup := setupDashboardRouterFull(t)
	defer cleanup()

	now := time.Now()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at) VALUES (?,?,?)`, 777, now, now)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	_, err = db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES (?,?,?,?)`, "apple", "", now, now)
	if err != nil {
		t.Fatalf("insert word_card error: %v", err)
	}

	_, err = db.GetConnection().Exec(`INSERT INTO training_cards (word_card_id, word_en, word_ru, meaning_en, sense_index, created_at)
		VALUES (1, 'apple', 'яблоко', 'apple', 0, ?)`, now)
	if err != nil {
		t.Fatalf("insert training_card error: %v", err)
	}

	_, err = db.GetConnection().Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, created_at, updated_at)
		VALUES (1, 1, 'en_ru', 'new', 2.5, ?, ?)`, now, now)
	if err != nil {
		t.Fatalf("insert user_card error: %v", err)
	}

	_, err = db.GetConnection().Exec(`INSERT INTO training_sessions (user_id, started_at, ended_at, source, planned_count, done_count)
		VALUES (1, ?, ?, 'manual', 1, 1)`, now, now)
	if err != nil {
		t.Fatalf("insert training_session error: %v", err)
	}

	_, err = db.GetConnection().Exec(`INSERT INTO review_events (session_id, user_id, user_card_id, direction, shown_at, answered_at, option_count, is_correct, quality)
		VALUES (1, 1, 1, 'en_ru', ?, ?, 4, 1, 5)`, now, now)
	if err != nil {
		t.Fatalf("insert review_event error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleDashboard(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDashboard_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupDashboardRouterFull(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	router.handleDashboard(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
