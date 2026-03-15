package bot

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func testBotConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = "127.0.0.1:0"
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = testutil.GetTestDSN(t)
	cfg.WebApp.JWTSecret = "test-jwt-secret-for-bot-tests"
	return cfg
}

// startMockTelegramServer starts an HTTP server that responds to Telegram API methods (getMe, getUpdates, setWebhook, etc.).
// Returns the base URL (e.g. "http://127.0.0.1:port") and a cleanup function.
func startMockTelegramServer(t *testing.T, statusCode int) (baseURL string, cleanup func()) {
	return startMockTelegramServerWithOpts(t, nil, statusCode)
}

// mockTelegramOpts allows customizing mock server behaviour per method.
type mockTelegramOpts struct {
	setWebhookStatusCode    int // 0 = use default statusCode
	deleteWebhookStatusCode int
	getUpdatesFirstResponse string // empty = use default result:[]
}

func startMockTelegramServerWithOpts(t *testing.T, opts *mockTelegramOpts, defaultStatusCode int) (baseURL string, cleanup func()) {
	t.Helper()
	getMeResp := `{"ok":true,"result":{"id":1,"is_bot":true,"first_name":"Test","username":"testbot"}}`
	getUpdatesResp := `{"ok":true,"result":[]}`
	okResp := `{"ok":true,"result":{}}`

	if opts == nil {
		opts = &mockTelegramOpts{}
	}
	statusFor := func(path string) int {
		if strings.Contains(path, "setWebhook") && opts.setWebhookStatusCode != 0 {
			return opts.setWebhookStatusCode
		}
		if strings.Contains(path, "deleteWebhook") && opts.deleteWebhookStatusCode != 0 {
			return opts.deleteWebhookStatusCode
		}
		return defaultStatusCode
	}

	var mu sync.Mutex
	firstGetUpdatesDone := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		code := statusFor(path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		if code != http.StatusOK {
			return
		}
		switch {
		case strings.HasSuffix(path, "getMe"):
			_, _ = w.Write([]byte(getMeResp))
		case strings.HasSuffix(path, "getUpdates"):
			mu.Lock()
			useFirst := opts.getUpdatesFirstResponse != "" && !firstGetUpdatesDone
			if useFirst {
				firstGetUpdatesDone = true
			}
			mu.Unlock()
			if useFirst {
				_, _ = w.Write([]byte(opts.getUpdatesFirstResponse))
			} else {
				_, _ = w.Write([]byte(getUpdatesResp))
			}
		default:
			_, _ = w.Write([]byte(okResp))
		}
	}))
	return srv.URL, srv.Close
}

func TestNew_NoToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	if bot.api != nil {
		t.Error("expected api nil when token empty")
	}
	if bot.db == nil {
		t.Error("expected db initialized")
	}
}

func TestNew_InvalidDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	cfg.Telegram.Token = ""
	cfg.Database.Driver = "postgres"
	cfg.Database.URL = "invalid-dsn"

	bot, err := New(cfg, logger)
	if err == nil {
		t.Fatal("New() expected error for invalid database")
	}
	if bot != nil {
		t.Error("New() expected nil bot on error")
	}
	if !strings.Contains(err.Error(), "database") {
		t.Errorf("error should mention database, got %q", err.Error())
	}
}

func TestNew_WithAPIBaseURL_Success(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	if bot.api == nil {
		t.Error("expected api initialized when token and APIBaseURL are set and server responds OK")
	}
}

func TestNew_TokenNoAPIBaseURL_InitFails(t *testing.T) {
	// Token set but no APIBaseURL: code uses NewBotAPI(token). With invalid token the real API fails -> bot.api == nil.
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "invalid-token-no-api-base"
	cfg.Telegram.APIBaseURL = ""

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	if bot.api != nil {
		t.Error("expected api nil when token invalid and no custom API base")
	}
}

func TestNew_WithAPIBaseURL_InitFails(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusInternalServerError)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	if bot.api != nil {
		t.Error("expected api nil when Telegram API returns error")
	}
}

func TestNew_TrainingPromptFile_Success(t *testing.T) {
	dir := t.TempDir()
	promptPath := filepath.Join(dir, "prompt.txt")
	if err := os.WriteFile(promptPath, []byte("You are a helpful trainer."), 0600); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Training.PromptFile = promptPath

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
}

func TestNew_TrainingPromptFile_Missing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Training.PromptFile = filepath.Join(t.TempDir(), "nonexistent.txt")

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot when prompt file missing (should continue without it)")
	}
}

func TestNew_WorkerEnabled_InvalidInterval(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Training.WorkerEnabled = true
	cfg.Training.WorkerInterval = "not-a-duration"
	cfg.Training.WorkerBatchSize = 5
	cfg.Training.LLMWorkers = 1
	cfg.Admin.TelegramID = 0

	bot, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if bot == nil {
		t.Fatal("New() returned nil bot")
	}
	if bot.trainingWorker == nil {
		t.Error("expected trainingWorker created with default interval when WorkerInterval invalid")
	}
}

func TestRegisterCommands_ApiNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	// Should not panic; logs "skipping bot commands registration"
	b.registerCommands()
}

func TestRegisterCommands_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	b.api = newTestBot(&mockTelegramClient{})
	b.registerCommands()
}

func TestRegisterCommands_RequestFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	api := &tgbotapi.BotAPI{Token: "test", Client: &failingTelegramClient{}, Buffer: 1}
	api.SetAPIEndpoint("http://example.com/bot%s/%s")
	b.api = api
	b.registerCommands()
	// Should not panic; logs "failed to set bot commands"
}

func TestNormalizeAPIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"two placeholders kept", "https://api.telegram.org/bot%s/%s", "https://api.telegram.org/bot%s/%s"},
		{"encoded placeholder fixed", "https://x/bot%25s/%25s", "https://x/bot%s/%s"},
		{"url with path", "https://host.example/path", "https://host.example/path/bot%s/%s"},
		{"url with path trailing slash", "https://host.example/path/", "https://host.example/path/bot%s/%s"},
		{"no scheme fallback append", "invalid", "invalid/bot%s/%s"},
		{"whitespace trimmed", "  https://host/  ", "https://host/bot%s/%s"},
		{"one placeholder rebuild from URL", "https://host.example/bot%s", "https://host.example/bot%s/bot%s/%s"},
		{"trailing slash fallback append", "something/", "something/bot%s/%s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAPIEndpoint(tt.in)
			if got != tt.want {
				t.Errorf("normalizeAPIEndpoint(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		in      string
		wantErr bool
	}{
		{"30s", false},
		{"5m", false},
		{"1h", false},
		{"0s", false},
		{"invalid", true},
		{"", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			d, err := parseDuration(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDuration(%q) expected error", tt.in)
				}
				return
			}
			if err != nil {
				t.Errorf("parseDuration(%q) error = %v", tt.in, err)
				return
			}
			if tt.in == "30s" && d != 30*time.Second {
				t.Errorf("parseDuration(30s) = %v", d)
			}
		})
	}
}

func TestStart_WebhookMode_ApiNil_CanceledContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.WebhookEnable = true
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_LongPollingMode_ApiNil_CanceledContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.WebhookEnable = false
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_Webhook_NoURLConfigured(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = ""
	cfg.Telegram.WebhookDomain = ""

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = b.Start(ctx)
	if err == nil {
		t.Fatal("Start() expected error when webhook enabled but neither webhook_url nor webhook_domain set")
	}
	if !strings.Contains(err.Error(), "webhook") {
		t.Errorf("error should mention webhook, got %q", err.Error())
	}
}

func TestStart_Webhook_WithAPI_Success(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/webhook"
	cfg.Telegram.WebhookPath = "/webhook"

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_LongPolling_WithAPI(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = false

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_Webhook_WithWebhookDomain(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = ""
	cfg.Telegram.WebhookDomain = "https://example.com"
	cfg.Telegram.WebhookPath = "/webhook"

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_Webhook_SetWebhookFails(t *testing.T) {
	baseURL, cleanup := startMockTelegramServerWithOpts(t, &mockTelegramOpts{setWebhookStatusCode: http.StatusInternalServerError}, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/webhook"

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = b.Start(ctx)
	if err == nil {
		t.Fatal("Start() expected error when setWebhook fails")
	}
	if !strings.Contains(err.Error(), "webhook") {
		t.Errorf("error should mention webhook, got %q", err.Error())
	}
}

func TestStart_Webhook_HandlerInvoked(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "http://" + addr + "/webhook"
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = addr

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = b.Start(ctx)
		close(done)
	}()
	defer func() { cancel(); <-done }()

	time.Sleep(200 * time.Millisecond)

	// Bad body triggers webhook handle error path (HandleUpdate returns err, 400)
	resp, err := http.Post("http://"+addr+"/webhook", "application/json", strings.NewReader("invalid"))
	if err != nil {
		t.Fatalf("post webhook: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid body, got %d", resp.StatusCode)
	}

	// Valid update triggers update != nil path (handler handles it)
	validUpdate := `{"update_id":1,"message":{"message_id":1,"date":123,"chat":{"id":1,"type":"private"},"from":{"id":1,"is_bot":false,"first_name":"Test"},"text":"/start"}}`
	resp2, err := http.Post("http://"+addr+"/webhook", "application/json", strings.NewReader(validUpdate))
	if err != nil {
		t.Fatalf("post webhook valid: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for valid update, got %d", resp2.StatusCode)
	}
}

func TestStart_Webhook_DeleteWebhookFails(t *testing.T) {
	baseURL, cleanup := startMockTelegramServerWithOpts(t, &mockTelegramOpts{deleteWebhookStatusCode: http.StatusInternalServerError}, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/webhook"
	cfg.Telegram.WebhookPath = "/webhook"

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_LongPolling_ReceivesUpdate(t *testing.T) {
	getUpdatesWithOne := `{"ok":true,"result":[{"update_id":1,"message":{"message_id":1,"date":123,"chat":{"id":1,"type":"private"},"from":{"id":1,"is_bot":false,"first_name":"Test"},"text":"/start"}}]}`
	baseURL, cleanup := startMockTelegramServerWithOpts(t, &mockTelegramOpts{getUpdatesFirstResponse: getUpdatesWithOne}, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = false

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestStart_WithTrainingWorker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.WebhookEnable = false
	cfg.Training.WorkerEnabled = true
	cfg.Training.WorkerInterval = "60s"
	cfg.Training.WorkerBatchSize = 2
	cfg.Training.LLMWorkers = 1
	cfg.Admin.TelegramID = 0

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if b.trainingWorker == nil {
		t.Fatal("expected trainingWorker when WorkerEnabled and valid interval")
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err = b.Start(ctx)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

// newTestBotWithEndpoint creates a *tgbotapi.BotAPI using the mock client, pointing at the given endpoint.
func newTestBotWithEndpoint(endpoint string) *tgbotapi.BotAPI {
	api := &tgbotapi.BotAPI{Token: "test-token", Client: &mockTelegramClient{}, Buffer: 100}
	api.SetAPIEndpoint(endpoint)
	return api
}

// failingResponseWriter is an http.ResponseWriter whose Write always returns an error.
type failingResponseWriter struct {
	header http.Header
	code   int
}

func (f *failingResponseWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failingResponseWriter) WriteHeader(code int) { f.code = code }

func (f *failingResponseWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

// TestHealthHandler_WriteError covers the b.logger.Error path inside healthHandler.
func TestHealthHandler_WriteError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	w := &failingResponseWriter{}
	b.healthHandler(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.code)
	}
}

// testBotWithFailingDBClose creates a Bot whose dbCloseFn returns an error,
// to exercise the "failed to close database" warn path.
func testBotWithFailingDBClose(t *testing.T, logger *zap.Logger) *Bot {
	t.Helper()
	cfg := testBotConfig(t)
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	b.dbCloseFn = func() error { return fmt.Errorf("simulated db close error") }
	return b
}

// TestStartWebhook_DbCloseError covers the db.Close() error warn path in startWebhook.
func TestStartWebhook_DbCloseError(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	b := testBotWithFailingDBClose(t, logger)

	b.api = newTestBotWithEndpoint(baseURL + "/bot%s/%s")
	b.config.Telegram.WebhookURL = "https://example.com/webhook"
	b.config.Telegram.WebhookPath = "/webhook"

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_ = b.startWebhook(ctx)
}

// TestStartLongPolling_DbCloseError covers the db.Close() error warn path in startLongPolling.
func TestStartLongPolling_DbCloseError(t *testing.T) {
	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	b := testBotWithFailingDBClose(t, logger)

	b.api = newTestBotWithEndpoint(baseURL + "/bot%s/%s")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_ = b.startLongPolling(ctx)
}

// TestStartWebServerOnly_DbCloseError covers the db.Close() error warn path in startWebServerOnly.
func TestStartWebServerOnly_DbCloseError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	b := testBotWithFailingDBClose(t, logger)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_ = b.startWebServerOnly(ctx)
}

// TestStartWebServerOnly_ShutdownError covers the server.Shutdown() error warn path.
// It uses a pre-canceled shutdown context so Shutdown returns context.Canceled
// when there are active connections.
func TestStartWebServerOnly_ShutdownError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Server.Address = addr
	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Use a pre-canceled context for Shutdown so it returns context.Canceled
	// when there are active (in-flight) connections.
	canceledCtx, cancelFn := context.WithCancel(context.Background())
	cancelFn()
	b.shutdownCtxFn = func() context.Context { return canceledCtx }

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = b.startWebServerOnly(ctx)
	}()

	// Wait for server to start.
	time.Sleep(100 * time.Millisecond)

	// Send a POST with Content-Length but no body — the server will be waiting
	// to read the body, keeping the connection in an active (non-idle) state.
	conn, dialErr := net.Dial("tcp", addr)
	if dialErr == nil {
		_, _ = conn.Write([]byte("POST /health HTTP/1.1\r\nHost: localhost\r\nContent-Length: 1000\r\nConnection: keep-alive\r\n\r\n"))
		defer conn.Close()
	}

	// Give the server a moment to accept the connection, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done
}

// TestStartWebhook_ListenAndServeError covers the http.ListenAndServe goroutine error path
// in startWebhook by using an address that is already bound.
func TestStartWebhook_ListenAndServeError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = true
	cfg.Telegram.WebhookURL = "https://example.com/webhook"
	cfg.Telegram.WebhookPath = "/webhook"
	cfg.Server.Address = addr

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_ = b.startWebhook(ctx)
}

// TestStartLongPolling_ListenAndServeError covers the http.ListenAndServe goroutine error path
// in startLongPolling by using an address that is already bound.
func TestStartLongPolling_ListenAndServeError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	baseURL, cleanup := startMockTelegramServer(t, http.StatusOK)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Telegram.Token = "test-token"
	cfg.Telegram.APIBaseURL = baseURL + "/bot%s/%s"
	cfg.Telegram.WebhookEnable = false
	cfg.Server.Address = addr

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_ = b.startLongPolling(ctx)
}

// TestStartWebServerOnly_ListenAndServeError covers the server.ListenAndServe goroutine error path
// (err != nil && err != http.ErrServerClosed) in startWebServerOnly.
func TestStartWebServerOnly_ListenAndServeError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	logger, _ := zap.NewDevelopment()
	cfg := testBotConfig(t)
	cfg.Server.Address = addr

	b, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(200 * time.Millisecond)
		cancel()
	}()

	_ = b.startWebServerOnly(ctx)
}
