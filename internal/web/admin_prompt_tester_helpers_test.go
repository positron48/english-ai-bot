package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setupAdminPromptTesterHelpersTest(t *testing.T) (*Router, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	cleanup := func() {} // shared db, do not close

	return router, db, cleanup
}

func TestSendNDJSONEvent(t *testing.T) {
	router, _, cleanup := setupAdminPromptTesterHelpersTest(t)
	defer cleanup()

	rr := httptest.NewRecorder()

	event := PromptTesterEvent{
		Word:       "test",
		Step:       "word",
		OK:         true,
		Raw:        `{"word": "test"}`,
		DurationMS: 100,
	}

	err := router.sendNDJSONEvent(rr, event)
	if err != nil {
		t.Fatalf("sendNDJSONEvent() error = %v", err)
	}

	// Check that response contains JSON
	body := rr.Body.String()
	if body == "" {
		t.Error("Expected non-empty response body")
	}

	// Verify it's valid JSON ending with newline
	var decoded PromptTesterEvent
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Errorf("Response is not valid JSON: %v, body: %s", err, body)
	}

	if decoded.Word != "test" {
		t.Errorf("Expected word 'test', got %s", decoded.Word)
	}
}

func TestSendNDJSONEvent_WithError(t *testing.T) {
	router, _, cleanup := setupAdminPromptTesterHelpersTest(t)
	defer cleanup()

	rr := httptest.NewRecorder()

	event := PromptTesterEvent{
		Word:       "test",
		Step:       "word",
		OK:         false,
		Error:      "test error",
		DurationMS: 50,
	}

	err := router.sendNDJSONEvent(rr, event)
	if err != nil {
		t.Fatalf("sendNDJSONEvent() error = %v", err)
	}

	// Verify error is in response
	body := rr.Body.String()
	var decoded PromptTesterEvent
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if decoded.Error != "test error" {
		t.Errorf("Expected error 'test error', got %s", decoded.Error)
	}
}

func TestSendNDJSONEvent_WithParsed(t *testing.T) {
	router, _, cleanup := setupAdminPromptTesterHelpersTest(t)
	defer cleanup()

	rr := httptest.NewRecorder()

	event := PromptTesterEvent{
		Word:   "test",
		Step:   "cards",
		OK:     true,
		Parsed: map[string]interface{}{"key": "value", "number": 42},
		DurationMS: 200,
	}

	err := router.sendNDJSONEvent(rr, event)
	if err != nil {
		t.Fatalf("sendNDJSONEvent() error = %v", err)
	}

	// Verify parsed data is in response
	body := rr.Body.String()
	var decoded PromptTesterEvent
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if decoded.Parsed == nil {
		t.Error("Expected parsed data in response")
	}
	if decoded.Parsed["key"] != "value" {
		t.Errorf("Expected parsed['key'] = 'value', got %v", decoded.Parsed["key"])
	}
}

func TestSendNDJSONEvent_Flusher(t *testing.T) {
	router, _, cleanup := setupAdminPromptTesterHelpersTest(t)
	defer cleanup()

	// httptest.ResponseRecorder implements http.Flusher
	rr := httptest.NewRecorder()

	event := PromptTesterEvent{
		Word:       "test",
		Step:       "word",
		OK:         true,
		DurationMS: 100,
	}

	err := router.sendNDJSONEvent(rr, event)
	if err != nil {
		t.Fatalf("sendNDJSONEvent() error = %v", err)
	}

	// Response should be written and flushed
	if rr.Body.Len() == 0 {
		t.Error("Expected response body to be written")
	}
}
