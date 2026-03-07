package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func setUserIDInContextPromptTester(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), userIDKey, userID)
	return req.WithContext(ctx)
}

func setupPromptTesterTest(t *testing.T) (*Router, *database.DB) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"
	tempFile, err := os.CreateTemp("", "training-prompt-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp prompt file: %v", err)
	}
	if _, err := tempFile.WriteString("test training prompt"); err != nil {
		_ = tempFile.Close()
		t.Fatalf("Failed to write temp prompt file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp prompt file: %v", err)
	}
	cfg.Training.PromptFile = tempFile.Name()
	t.Cleanup(func() { _ = os.Remove(tempFile.Name()) })

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)

	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	return router, db
}

func TestHandleAdminPromptTesterDefaultPrompts_Get(t *testing.T) {
	router, _ := setupPromptTesterTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/prompt-tester/default-prompts", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterDefaultPrompts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminPromptTesterDefaultPrompts_MethodNotAllowed(t *testing.T) {
	router, _ := setupPromptTesterTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/default-prompts", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterDefaultPrompts(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminPromptTesterDefaultPrompts_WordPromptFromFile(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"

	wordPromptFile, err := os.CreateTemp("", "word-prompt-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp word prompt file: %v", err)
	}
	wordPromptContent := "custom word prompt from file"
	if _, err := wordPromptFile.WriteString(wordPromptContent); err != nil {
		_ = wordPromptFile.Close()
		t.Fatalf("Failed to write word prompt file: %v", err)
	}
	if err := wordPromptFile.Close(); err != nil {
		t.Fatalf("Failed to close word prompt file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(wordPromptFile.Name()) })
	cfg.AI.PromptFile = wordPromptFile.Name()
	cfg.AI.Prompt = "fallback env prompt"

	trainingPromptFile, err := os.CreateTemp("", "training-prompt-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp training prompt file: %v", err)
	}
	if _, err := trainingPromptFile.WriteString("test training prompt"); err != nil {
		_ = trainingPromptFile.Close()
		t.Fatalf("Failed to write training prompt file: %v", err)
	}
	if err := trainingPromptFile.Close(); err != nil {
		t.Fatalf("Failed to close training prompt file: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(trainingPromptFile.Name()) })
	cfg.Training.PromptFile = trainingPromptFile.Name()

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/prompt-tester/default-prompts", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterDefaultPrompts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["word_prompt"] != wordPromptContent {
		t.Errorf("Expected word_prompt from file %q, got %q", wordPromptContent, resp["word_prompt"])
	}
	if resp["word_prompt_source"] != cfg.AI.PromptFile {
		t.Errorf("Expected word_prompt_source %q, got %q", cfg.AI.PromptFile, resp["word_prompt_source"])
	}
}

func TestHandleAdminPromptTesterDefaultPrompts_TrainingFileNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	cfg := &config.Config{}
	cfg.Admin.TelegramID = 12345
	cfg.WebApp.JWTSecret = "test-secret"
	cfg.AI.Prompt = "env word prompt"
	cfg.Training.PromptFile = "/nonexistent/training-prompt.txt"

	cbRepo := repository.NewCircuitBreakerRepository(db.GetConnection(), logger)
	cbService := service.NewCircuitBreakerService(cbRepo, 5, logger)
	router := NewRouter(logger, cfg, db.GetConnection(), nil, nil, nil, cbService)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/prompt-tester/default-prompts", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterDefaultPrompts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 even when training file missing, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if src, ok := resp["training_prompt_source"].(string); !ok || src == "" {
		t.Errorf("Expected training_prompt_source to indicate not found, got %v", resp["training_prompt_source"])
	}
	if !strings.Contains(resp["training_prompt_source"].(string), "not found") {
		t.Errorf("Expected training_prompt_source to contain 'not found', got %v", resp["training_prompt_source"])
	}
}

func TestHandleAdminPromptTesterRun_EmptyWordSkipped(t *testing.T) {
	mockAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(openAIChoiceResponse(`{"word":"test","translation":"тест"}`))
	}))
	defer mockAI.Close()

	router, _ := setupPromptTesterTest(t)
	router.config.AI.URL = strings.TrimSuffix(mockAI.URL, "/")
	router.config.AI.Model = "test-model"
	router.config.AI.APIKey = "test-key"

	// One valid word and one empty/whitespace - only one word should be processed
	body := `{"words": ["hello", "  ", ""], "word_prompt": "word", "training_prompt": "training"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	lines := strings.Split(strings.TrimSpace(rr.Body.String()), "\n")
	// One word "hello" -> 2 events (word + cards). Empty strings are skipped.
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 NDJSON lines for one word, got %d", len(lines))
	}
}

func TestHandleAdminPromptTesterRun_MethodNotAllowed(t *testing.T) {
	router, _ := setupPromptTesterTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/prompt-tester/run", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleAdminPromptTesterRun_InvalidBody(t *testing.T) {
	router, _ := setupPromptTesterTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", nil)
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminPromptTesterRun_EmptyWords(t *testing.T) {
	router, _ := setupPromptTesterTest(t)

	body := `{"words": [], "word_prompt": "p", "training_prompt": "tp"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "words") {
		t.Errorf("Expected error to mention words, got %s", rr.Body.String())
	}
}

func TestHandleAdminPromptTesterRun_NoWordPrompt(t *testing.T) {
	router, _ := setupPromptTesterTest(t)

	body := `{"words": ["hello"], "training_prompt": "tp"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandleAdminPromptTesterRun_NoTrainingPrompt(t *testing.T) {
	router, _ := setupPromptTesterTest(t)

	body := `{"words": ["hello"], "word_prompt": "wp"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d: %s", rr.Code, rr.Body.String())
	}
}

// openAIChoiceResponse returns a minimal OpenAI-compatible chat completion response.
func openAIChoiceResponse(content string) []byte {
	out := map[string]interface{}{
		"choices": []map[string]interface{}{
			{"message": map[string]interface{}{"content": content}},
		},
	}
	b, _ := json.Marshal(out)
	return b
}

func TestHandleAdminPromptTesterRun_Success(t *testing.T) {
	mockAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" && !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("Unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(openAIChoiceResponse(`{"word":"test","translation":"тест"}`))
	}))
	defer mockAI.Close()

	router, _ := setupPromptTesterTest(t)
	router.config.AI.URL = strings.TrimSuffix(mockAI.URL, "/")
	router.config.AI.Model = "test-model"
	router.config.AI.APIKey = "test-key"

	body := `{"words": ["hello"], "word_prompt": "word", "training_prompt": "training"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Errorf("Expected Content-Type application/x-ndjson, got %s", ct)
	}
	lines := strings.Split(strings.TrimSpace(rr.Body.String()), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 NDJSON lines (word + cards), got %d", len(lines))
	}
}

// TestHandleAdminPromptTesterRun_NonJSONResponse covers branches when LLM returns non-JSON (Raw fallback).
func TestHandleAdminPromptTesterRun_NonJSONResponse(t *testing.T) {
	mockAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("plain text response, not JSON"))
	}))
	defer mockAI.Close()

	router, _ := setupPromptTesterTest(t)
	router.config.AI.URL = strings.TrimSuffix(mockAI.URL, "/")
	router.config.AI.Model = "test-model"
	router.config.AI.APIKey = "test-key"

	body := `{"words": ["x"], "word_prompt": "wp", "training_prompt": "tp"}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/prompt-tester/run", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = setUserIDInContextPromptTester(req, 12345)
	rr := httptest.NewRecorder()

	router.handleAdminPromptTesterRun(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 (events still streamed), got %d: %s", rr.Code, rr.Body.String())
	}
	lines := strings.Split(strings.TrimSpace(rr.Body.String()), "\n")
	if len(lines) < 2 {
		t.Errorf("Expected at least 2 NDJSON lines, got %d", len(lines))
	}
	// First event (word step) should have Raw set to raw response when parse fails
	var wordEvent struct {
		Word  string `json:"word"`
		Step  string `json:"step"`
		OK    bool   `json:"ok"`
		Raw   string `json:"raw"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &wordEvent); err != nil {
		t.Fatalf("unmarshal word event: %v", err)
	}
	if wordEvent.Step != "word" {
		t.Errorf("expected first event step word, got %s", wordEvent.Step)
	}
	if wordEvent.OK && wordEvent.Raw == "" {
		t.Error("when response is non-JSON, raw should be set to response body")
	}
}
