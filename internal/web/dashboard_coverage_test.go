package web

// Additional tests to achieve 100% coverage for internal/web/dashboard.go

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// ── handleSettings ────────────────────────────────────────────────────────────

// TestHandleSettings_LanguageNotSet covers the path where settings.Language == ""
// (detect from Accept-Language header).
func TestHandleSettings_LanguageNotSet(t *testing.T) {
	router, db := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})

	// Insert user with empty settings_json (no language set)
	_, err := db.GetConnection().Exec(
		`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`,
		8001, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleSettings_InvalidSettingsJSON covers the Warn path when settings_json is invalid.
func TestHandleSettings_InvalidSettingsJSON(t *testing.T) {
	router, db := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})

	_, err := db.GetConnection().Exec(
		`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`,
		8002, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{invalid json}`,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleSettings(w, req)

	// Handler warns on invalid settings_json and continues with defaults
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// setupDashboardRouterWithUpdateBlocked creates a router with a second isolated DB that has a user,
// but blocks UPDATE on users table via a trigger to force UpdateUserSettings to fail.
func setupDashboardRouterWithUpdateBlocked(t *testing.T) (*Router, int64) {
	t.Helper()
	logger, _ := zap.NewDevelopment()

	dsn := testutil.SecondPostgresDSN(t)

	var dbWrap *database.DB
	var err error
	dbWrap, err = database.NewWithConfig("postgres", "", dsn, logger)
	if dbWrap == nil {
		t.Skipf("second DB not available: %v", err)
	}
	t.Cleanup(func() { _ = dbWrap.GetConnection().Close() })

	conn := dbWrap.GetConnection()

	// Insert a user
	var userID int64
	err = conn.QueryRow(`INSERT INTO users (telegram_id, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id`,
		99001).Scan(&userID)
	if err != nil {
		t.Skipf("insert user: %v", err)
	}

	// Add trigger to block updates on users table
	_, err = conn.Exec(`
		CREATE OR REPLACE FUNCTION block_users_update_fn() RETURNS TRIGGER AS $$
		BEGIN
			RAISE EXCEPTION 'Update blocked for testing';
		END;
		$$ LANGUAGE plpgsql;
		CREATE TRIGGER block_users_update
		BEFORE UPDATE ON users
		FOR EACH ROW EXECUTE FUNCTION block_users_update_fn();
	`)
	if err != nil {
		t.Skipf("create trigger: %v", err)
	}

	cfg := &config.Config{}
	cfg.WebApp.JWTSecret = "test-secret"
	cbRepo := repository.NewCircuitBreakerRepository(conn, logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	userRepo := repository.NewUserRepository(conn, logger)
	router := NewRouter(logger, cfg, conn, nil, nil, nil, cbService)
	router.SetDependencies(userRepo, &mockWordService{}, &mockAIService{}, nil, "")
	return router, userID
}

// ── handleNotificationSettings ────────────────────────────────────────────────

// TestHandleNotificationSettings_DailyFrequency covers the "daily" frequency path.
func TestHandleNotificationSettings_DailyFrequency(t *testing.T) {
	router, db := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})

	_, err := db.GetConnection().Exec(
		`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`,
		8010, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications",
		bytes.NewBufferString(`{"frequency":"daily"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleNotificationSettings_NeverFrequency covers the "never" frequency path.
func TestHandleNotificationSettings_NeverFrequency(t *testing.T) {
	router, db := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})

	_, err := db.GetConnection().Exec(
		`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`,
		8011, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications",
		bytes.NewBufferString(`{"frequency":"never"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleNotificationSettings_UpdateUserSettingsError covers the UpdateUserSettings error path.
func TestHandleNotificationSettings_UpdateUserSettingsError(t *testing.T) {
	router, userID := setupDashboardRouterWithUpdateBlocked(t)

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications",
		bytes.NewBufferString(`{"frequency":"daily"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when UpdateUserSettings fails, got %d: %s", w.Code, w.Body.String())
	}
}

// ── handleLanguageSettings ────────────────────────────────────────────────────

// TestHandleLanguageSettings_UpdateUserSettingsError covers the UpdateUserSettings error path.
func TestHandleLanguageSettings_UpdateUserSettingsError(t *testing.T) {
	router, userID := setupDashboardRouterWithUpdateBlocked(t)

	req := httptest.NewRequest(http.MethodPost, "/api/settings/language",
		bytes.NewBufferString(`{"language":"en"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleLanguageSettings(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when UpdateUserSettings fails, got %d: %s", w.Code, w.Body.String())
	}
}

// ── handleTrainingSettings ────────────────────────────────────────────────────

// TestHandleTrainingSettings_UpdateUserSettingsError covers the UpdateUserSettings error path.
func TestHandleTrainingSettings_UpdateUserSettingsError(t *testing.T) {
	router, userID := setupDashboardRouterWithUpdateBlocked(t)

	req := httptest.NewRequest(http.MethodPost, "/api/settings/training",
		bytes.NewBufferString(`{"options_delay_seconds":3}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, userID)
	w := httptest.NewRecorder()
	router.handleTrainingSettings(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when UpdateUserSettings fails, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTrainingSettings_NilPointers covers the path where all pointer fields are nil
// (defaultIntPtr returns defaults).
func TestHandleTrainingSettings_NilPointers(t *testing.T) {
	router, db := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})

	_, err := db.GetConnection().Exec(
		`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`,
		8020, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	// Send empty body (no fields) - all pointer fields remain nil in settings
	req := httptest.NewRequest(http.MethodPost, "/api/settings/training",
		bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleTrainingSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestHandleTrainingSettings_SpellModeEnabledTrue covers the spell_mode_enabled=true path
// in the response (settings.SpellModeEnabled != nil && *settings.SpellModeEnabled).
func TestHandleTrainingSettings_SpellModeEnabledTrue(t *testing.T) {
	router, db := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})

	_, err := db.GetConnection().Exec(
		`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`,
		8021, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/training",
		bytes.NewBufferString(`{"spell_mode_enabled":true,"type_mode_enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleTrainingSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
