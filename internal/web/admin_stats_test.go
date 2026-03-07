package web

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupAdminStatsRouter(t *testing.T) (*Router, *database.DB, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	cleanup := func() {} // shared db, do not close

	return router, db, cleanup
}

func TestHandleAdminStats_OK(t *testing.T) {
	router, db, cleanup := setupAdminStatsRouter(t)
	defer cleanup()

	now := time.Now()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at) VALUES (?,?,?)`, 123, now, now)
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

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats?days=2", nil)
	w := httptest.NewRecorder()
	router.handleAdminStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if payload["users_total"] == nil {
		t.Fatalf("expected users_total")
	}
	if payload["windows"] == nil {
		t.Fatalf("expected windows")
	}
	if payload["cards_state"] == nil {
		t.Fatalf("expected cards_state")
	}
	if payload["daily"] == nil {
		t.Fatalf("expected daily stats")
	}
}

func TestHandleAdminStats_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupAdminStatsRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/stats", nil)
	w := httptest.NewRecorder()
	router.handleAdminStats(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleAdminStats_DaysParam(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantDays   int
		desc       string
	}{
		{"default no param", "", 30, "default 30 days"},
		{"invalid non-numeric", "?days=abc", 30, "invalid keeps default"},
		{"zero", "?days=0", 30, "zero keeps default"},
		{"negative", "?days=-5", 30, "negative keeps default"},
		{"over max", "?days=400", 30, "over 365 keeps default"},
		{"min valid", "?days=1", 1, "1 day"},
		{"max valid", "?days=365", 365, "365 days"},
		{"valid 7", "?days=7", 7, "7 days"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, _, cleanup := setupAdminStatsRouter(t)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/admin/stats"+tt.query, nil)
			w := httptest.NewRecorder()
			router.handleAdminStats(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", w.Code)
			}
			var payload map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			daily, _ := payload["daily"].([]interface{})
			if daily == nil {
				t.Fatal("expected daily array")
			}
			if got := len(daily); got != tt.wantDays {
				t.Errorf("daily length: got %d, want %d (%s)", got, tt.wantDays, tt.desc)
			}
		})
	}
}

func TestHandleAdminStats_EmptyDB(t *testing.T) {
	router, _, cleanup := setupAdminStatsRouter(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats?days=3", nil)
	w := httptest.NewRecorder()
	router.handleAdminStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["users_total"] != float64(0) {
		t.Errorf("users_total: got %v, want 0", payload["users_total"])
	}
	windows, _ := payload["windows"].(map[string]interface{})
	if windows != nil {
		for _, wname := range []string{"24h", "7d", "30d"} {
			if ww, ok := windows[wname].(map[string]interface{}); ok {
				if acc, ok := ww["accuracy_percent"].(float64); ok && acc != 0 {
					t.Errorf("windows[%s].accuracy_percent: got %v, want 0 (empty DB)", wname, acc)
				}
			}
		}
	}
}

// TestHandleAdminStats_DailyDataInWindow covers the branch where daily query rows
// have day in dayMap and we update dayData (active_users, sessions_started, etc.).
func TestHandleAdminStats_DailyDataInWindow(t *testing.T) {
	router, db, cleanup := setupAdminStatsRouter(t)
	defer cleanup()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at) VALUES (?,?,?)`, 111, now, now)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES (?,?,?,?)`, "w", "", now, now)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	_, err = db.GetConnection().Exec(`INSERT INTO training_cards (word_card_id, word_en, word_ru, meaning_en, sense_index, created_at) VALUES (1, 'w', 'в', 'w', 0, ?)`, now)
	if err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	_, err = db.GetConnection().Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, created_at, updated_at) VALUES (1, 1, 'en_ru', 'new', 2.5, ?, ?)`, now, now)
	if err != nil {
		t.Fatalf("insert user_card: %v", err)
	}
	// Second user_card with created_at yesterday so daily "cards_added" query has a row in dayMap
	_, err = db.GetConnection().Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, created_at, updated_at) VALUES (1, 1, 'ru_en', 'new', 2.5, ?, ?)`, yesterday, yesterday)
	if err != nil {
		t.Fatalf("insert user_card yesterday: %v", err)
	}
	// Session yesterday so DATE(started_at) is in the 2-day window
	_, err = db.GetConnection().Exec(`INSERT INTO training_sessions (user_id, started_at, ended_at, source, planned_count, done_count) VALUES (1, ?, ?, 'manual', 1, 1)`, yesterday, yesterday)
	if err != nil {
		t.Fatalf("insert training_session: %v", err)
	}
	_, err = db.GetConnection().Exec(`INSERT INTO review_events (session_id, user_id, user_card_id, direction, shown_at, answered_at, option_count, is_correct, quality) VALUES (1, 1, 1, 'en_ru', ?, ?, 4, 1, 5)`, yesterday, yesterday)
	if err != nil {
		t.Fatalf("insert review_event: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats?days=2", nil)
	w := httptest.NewRecorder()
	router.handleAdminStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	daily, _ := payload["daily"].([]interface{})
	if len(daily) != 2 {
		t.Fatalf("daily length: got %d, want 2", len(daily))
	}
	// At least one day should have non-zero activity
	gotNonZero := false
	for _, d := range daily {
		m, _ := d.(map[string]interface{})
		if m == nil {
			continue
		}
		if au, _ := m["active_users"].(float64); au > 0 {
			gotNonZero = true
			break
		}
		if ss, _ := m["sessions_started"].(float64); ss > 0 {
			gotNonZero = true
			break
		}
	}
	if !gotNonZero {
		t.Error("expected at least one day with active_users or sessions_started > 0")
	}
}

// TestHandleAdminStats_DailyRowOutsideWindow covers the branch where a daily query
// returns a date not in dayMap (e.g. activity on "today" when days=2 so the map only has yesterday and day before).
func TestHandleAdminStats_DailyRowOutsideWindow(t *testing.T) {
	router, db, cleanup := setupAdminStatsRouter(t)
	defer cleanup()

	now := time.Now()
	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at) VALUES (?,?,?)`, 999, now, now)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.GetConnection().Exec(`INSERT INTO word_cards (word, definition, created_at, updated_at) VALUES (?,?,?,?)`, "word", "", now, now)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	_, err = db.GetConnection().Exec(`INSERT INTO training_cards (word_card_id, word_en, word_ru, meaning_en, sense_index, created_at) VALUES (1, 'w', 'в', 'w', 0, ?)`, now)
	if err != nil {
		t.Fatalf("insert training_card: %v", err)
	}
	_, err = db.GetConnection().Exec(`INSERT INTO user_cards (user_id, training_card_id, direction, state, ef, created_at, updated_at) VALUES (1, 1, 'en_ru', 'new', 2.5, ?, ?)`, now, now)
	if err != nil {
		t.Fatalf("insert user_card: %v", err)
	}
	// Session with started_at = now (today); with days=2 the dayMap only has (today-2, today-1), so "today" is not in map
	_, err = db.GetConnection().Exec(`INSERT INTO training_sessions (user_id, started_at, ended_at, source, planned_count, done_count) VALUES (1, ?, ?, 'manual', 1, 1)`, now, now)
	if err != nil {
		t.Fatalf("insert training_session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats?days=2", nil)
	w := httptest.NewRecorder()
	router.handleAdminStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["daily"] == nil {
		t.Fatal("expected daily")
	}
}

func TestHandleAdminStats_DBError(t *testing.T) {
	testutil.SetupTestDatabase(t) // ensure postgres_compat driver is registered
	closedConn, err := sql.Open("postgres_compat", testutil.GetTestDSN(t))
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	if err := closedConn.Ping(); err != nil {
		_ = closedConn.Close()
		t.Skip("ping failed:", err)
	}
	closedConn.Close()

	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cbRepo := repository.NewCircuitBreakerRepository(closedConn, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, closedConn, nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/stats?days=2", nil)
	w := httptest.NewRecorder()
	router.handleAdminStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on DB error (handler degrades), got %d", w.Code)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload["users_total"] != float64(0) {
		t.Errorf("users_total with failed query should be 0, got %v", payload["users_total"])
	}
	if payload["daily"] == nil {
		t.Error("daily should be present")
	}
}
