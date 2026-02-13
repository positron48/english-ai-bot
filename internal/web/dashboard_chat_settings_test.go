package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

type mockWordService struct {
	isSingle bool
	resp     string
}

func (m *mockWordService) IsSingleWord(text string) bool {
	return m.isSingle
}

func (m *mockWordService) GetWordDefinition(ctx context.Context, userID int64, word string) (string, error) {
	return m.resp, nil
}

type mockAIService struct {
	resp string
}

func (m *mockAIService) GenerateResponse(ctx context.Context, text string) (string, error) {
	return m.resp, nil
}

func setupDashboardRouterDeps(t *testing.T, ws *mockWordService, ai *mockAIService) (*Router, *database.DB, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	cfg := &config.Config{}
	cfg.WebApp.JWTSecret = "test-secret"
	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)
	router.SetDependencies(repository.NewUserRepository(db.GetConnection(), logger), ws, ai, nil, "")

	cleanup := func() {
		_ = db.Close()
	}

	return router, db, cleanup
}

func TestHandleChatSettings_Unauthorized(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString("message=hi"))
	w := httptest.NewRecorder()
	router.handleChat(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleChatSettings_SingleWordEnglish(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{isSingle: true, resp: "def"}, &mockAIService{resp: "ai"})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString("message=apple"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleChat(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleChatSettings_SingleWordCyrillic(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{isSingle: true, resp: "def"}, &mockAIService{resp: "ai"})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString("message=дом"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleChat(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleChatSettings_MessageMissing(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleChat(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSettingsChatSettings_Get(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 999, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	settings, ok := payload["settings"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected settings object")
	}
	if settings["notification_frequency"] == nil {
		t.Fatalf("expected notification_frequency")
	}
}

func TestHandleSettingsChatSettings_Unauthorized(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()
	router.handleSettings(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleSettingsChatSettings_UserNotFound(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req = setUserIDInContext(req, 999)
	w := httptest.NewRecorder()
	router.handleSettings(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleChatSettings_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/chat", nil)
	w := httptest.NewRecorder()
	router.handleChat(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleSettingsChatSettings_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings", nil)
	w := httptest.NewRecorder()
	router.handleSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleChatSettings_NonSingleWord(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{isSingle: false, resp: "def"}, &mockAIService{resp: "ai"})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString("message=hello world"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleChat(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleSettingsChatSettings_ParsesSettings(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	settings := models.UserSettings{NotificationFrequency: "never"}
	payload, _ := json.Marshal(settings)
	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 1001, "2026-01-01 00:00:00", "2026-01-01 00:00:00", string(payload))
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleNotificationSettings_Success(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 2001, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications", bytes.NewBufferString(`{"frequency":"3"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleNotificationSettings_InvalidFrequency(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 2002, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications", bytes.NewBufferString(`{"frequency":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleLanguageSettings_Success(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 2003, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/language", bytes.NewBufferString(`{"language":"ru"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLanguageSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleNotificationSettings_Unauthorized(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications", bytes.NewBufferString(`{"frequency":"daily"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleNotificationSettings_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/settings/notifications", nil)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleNotificationSettings_UserNotFound(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications", bytes.NewBufferString(`{"frequency":"daily"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 999)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleNotificationSettings_InvalidJSON(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleNotificationSettings_EmptyFrequency(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 2005, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/notifications", bytes.NewBufferString(`{"frequency":""}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleNotificationSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleLanguageSettings_InvalidJSON(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/language", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLanguageSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleLanguageSettings_UserNotFound(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/language", bytes.NewBufferString(`{"language":"en"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 999)
	w := httptest.NewRecorder()
	router.handleLanguageSettings(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleLanguageSettings_InvalidLanguage(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 2004, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/language", bytes.NewBufferString(`{"language":"es"}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleLanguageSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleLanguageSettings_Unauthorized(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/language", bytes.NewBufferString(`{"language":"en"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleLanguageSettings(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleLanguageSettings_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/settings/language", nil)
	w := httptest.NewRecorder()
	router.handleLanguageSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleTrainingSettings_Success(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 3001, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/training", bytes.NewBufferString(`{"options_delay_seconds":3,"wrong_answer_delay_seconds":2}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleTrainingSettings(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	settings, _ := resp["settings"].(map[string]interface{})
	if settings == nil {
		t.Fatal("expected settings in response")
	}
	if v, _ := settings["options_delay_seconds"].(float64); v != 3 {
		t.Errorf("expected options_delay_seconds 3, got %v", v)
	}
	if v, _ := settings["wrong_answer_delay_seconds"].(float64); v != 2 {
		t.Errorf("expected wrong_answer_delay_seconds 2, got %v", v)
	}

	// Verify GET /api/settings returns saved values
	getReq := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	getReq = setUserIDInContext(getReq, 1)
	getW := httptest.NewRecorder()
	router.handleSettings(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("GET settings expected 200, got %d", getW.Code)
	}
	var getResp struct {
		Settings models.UserSettings `json:"settings"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if getResp.Settings.OptionsDelaySeconds == nil || *getResp.Settings.OptionsDelaySeconds != 3 {
		t.Errorf("GET settings: expected options_delay_seconds 3, got %v", getResp.Settings.OptionsDelaySeconds)
	}
	if getResp.Settings.WrongAnswerDelaySeconds == nil || *getResp.Settings.WrongAnswerDelaySeconds != 2 {
		t.Errorf("GET settings: expected wrong_answer_delay_seconds 2, got %v", getResp.Settings.WrongAnswerDelaySeconds)
	}

	// Save spell settings and verify
	req2 := httptest.NewRequest(http.MethodPost, "/api/settings/training", bytes.NewBufferString(`{"spell_mode_enabled":false,"spell_mastering_threshold":75}`))
	req2.Header.Set("Content-Type", "application/json")
	req2 = setUserIDInContext(req2, 1)
	w2 := httptest.NewRecorder()
	router.handleTrainingSettings(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("POST spell settings expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	getReq2 := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	getReq2 = setUserIDInContext(getReq2, 1)
	getW2 := httptest.NewRecorder()
	router.handleSettings(getW2, getReq2)
	if getW2.Code != http.StatusOK {
		t.Fatalf("GET settings (spell) expected 200, got %d", getW2.Code)
	}
	var getResp2 struct {
		Settings models.UserSettings `json:"settings"`
	}
	if err := json.NewDecoder(getW2.Body).Decode(&getResp2); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if getResp2.Settings.SpellModeEnabled == nil || *getResp2.Settings.SpellModeEnabled != false {
		t.Errorf("GET settings: expected spell_mode_enabled false, got %v", getResp2.Settings.SpellModeEnabled)
	}
	if getResp2.Settings.SpellMasteringThreshold == nil || *getResp2.Settings.SpellMasteringThreshold != 75 {
		t.Errorf("GET settings: expected spell_mastering_threshold 75, got %v", getResp2.Settings.SpellMasteringThreshold)
	}

	// Save type settings and verify
	req3 := httptest.NewRequest(http.MethodPost, "/api/settings/training", bytes.NewBufferString(`{"type_mode_enabled":false,"type_mastering_threshold":80}`))
	req3.Header.Set("Content-Type", "application/json")
	req3 = setUserIDInContext(req3, 1)
	w3 := httptest.NewRecorder()
	router.handleTrainingSettings(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("POST type settings expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
	getReq3 := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	getReq3 = setUserIDInContext(getReq3, 1)
	getW3 := httptest.NewRecorder()
	router.handleSettings(getW3, getReq3)
	if getW3.Code != http.StatusOK {
		t.Fatalf("GET settings (type) expected 200, got %d", getW3.Code)
	}
	var getResp3 struct {
		Settings models.UserSettings `json:"settings"`
	}
	if err := json.NewDecoder(getW3.Body).Decode(&getResp3); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	if getResp3.Settings.TypeModeEnabled == nil || *getResp3.Settings.TypeModeEnabled != false {
		t.Errorf("GET settings: expected type_mode_enabled false, got %v", getResp3.Settings.TypeModeEnabled)
	}
	if getResp3.Settings.TypeMasteringThreshold == nil || *getResp3.Settings.TypeMasteringThreshold != 80 {
		t.Errorf("GET settings: expected type_mastering_threshold 80, got %v", getResp3.Settings.TypeMasteringThreshold)
	}
}

func TestHandleTrainingSettings_InvalidValue(t *testing.T) {
	router, db, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	_, err := db.GetConnection().Exec(`INSERT INTO users (telegram_id, created_at, updated_at, settings_json) VALUES (?,?,?,?)`, 3002, "2026-01-01 00:00:00", "2026-01-01 00:00:00", `{}`)
	if err != nil {
		t.Fatalf("insert user error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/settings/training", bytes.NewBufferString(`{"options_delay_seconds":11}`))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContext(req, 1)
	w := httptest.NewRecorder()
	router.handleTrainingSettings(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleTrainingSettings_Unauthorized(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/settings/training", bytes.NewBufferString(`{"options_delay_seconds":5}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.handleTrainingSettings(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleTrainingSettings_MethodNotAllowed(t *testing.T) {
	router, _, cleanup := setupDashboardRouterDeps(t, &mockWordService{}, &mockAIService{})
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/settings/training", nil)
	w := httptest.NewRecorder()
	router.handleTrainingSettings(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
