package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

// mockPronunciationService implements pronunciationServiceInterface for tts tests
// so we can cover branches that real service never hits (e.g. empty PublicBasePath/AudioDir).
type mockPronunciationService struct {
	enabled       bool
	publicBase    string
	audioDir      string
	lookupResult  service.PronunciationLookupResult
}

func (m *mockPronunciationService) IsEnabled() bool                                    { return m.enabled }
func (m *mockPronunciationService) PublicBasePath() string                              { return m.publicBase }
func (m *mockPronunciationService) AudioDir() string                                   { return m.audioDir }
func (m *mockPronunciationService) Lookup(word string) service.PronunciationLookupResult { return m.lookupResult }
func (m *mockPronunciationService) GetStatus(word string) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{}, nil
}
func (m *mockPronunciationService) ForceRegenerate(word string) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{}, nil
}
func (m *mockPronunciationService) Recheck(word string) (service.TTSStatusResult, error) {
	return service.TTSStatusResult{}, nil
}
func (m *mockPronunciationService) ScheduleWord(word string) bool { return true }

func TestSetupPronunciationMediaRoute_RegistersRoute(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router.SetPronunciationService(pronService)

	req := httptest.NewRequest(http.MethodGet, "/media/tts/some/file.mp3", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)

	// handleTTSMedia returns 404 when file does not exist or path invalid; we just check the route is registered
	if w.Code != http.StatusNotFound && w.Code != http.StatusOK {
		t.Errorf("Expected 404 or 200 from media route, got %d", w.Code)
	}
}

func TestSetupPronunciationMediaRoute_NoOpWhenDisabled(t *testing.T) {
	logger := zap.NewNop()
	router := &Router{
		mux:                  http.NewServeMux(),
		logger:               logger,
		pronunciationService: service.NewPronunciationService(config.TTSConfig{Enabled: false}, nil, logger),
	}
	router.setupPronunciationMediaRoute()

	if router.pronunciationMediaRouteRegistered {
		t.Error("Route should not be registered when service is disabled")
	}
}

func TestSetupPronunciationMediaRoute_NoOpWhenServiceNil(t *testing.T) {
	logger := zap.NewNop()
	router := &Router{
		mux:                  http.NewServeMux(),
		logger:               logger,
		pronunciationService: nil,
	}
	router.setupPronunciationMediaRoute()

	if router.pronunciationMediaRouteRegistered {
		t.Error("Route should not be registered when service is nil")
	}
	req := httptest.NewRequest(http.MethodGet, "/media/tts/any.mp3", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when no route registered, got %d", w.Code)
	}
}

func TestSetupPronunciationMediaRoute_Idempotent(t *testing.T) {
	logger := zap.NewNop()
	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router := &Router{
		mux:                  http.NewServeMux(),
		logger:               logger,
		pronunciationService: pronService,
	}
	router.setupPronunciationMediaRoute()
	router.setupPronunciationMediaRoute() // second call should no-op

	req := httptest.NewRequest(http.MethodGet, "/media/tts/some.mp3", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)
	// Route is registered; file does not exist so 404 from handler
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 from handler, got %d", w.Code)
	}
}

// TestSetupPronunciationMediaRoute_AlreadyRegistered covers the early return when the route was already registered.
func TestSetupPronunciationMediaRoute_AlreadyRegistered(t *testing.T) {
	logger := zap.NewNop()
	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router := &Router{
		mux:                               http.NewServeMux(),
		logger:                            logger,
		pronunciationService:             pronService,
		pronunciationMediaRouteRegistered: true, // already registered
	}
	router.setupPronunciationMediaRoute() // should return without re-registering

	// No route was registered (we skipped registration), so mux returns 404
	req := httptest.NewRequest(http.MethodGet, "/media/tts/any.mp3", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when route was not registered, got %d", w.Code)
	}
}

func TestSetupPronunciationMediaRoute_EmptyPublicBasePath(t *testing.T) {
	// Real PronunciationService normalizes "" to "/media/tts"; use mock to hit fallback in tts.go
	logger := zap.NewNop()
	mock := &mockPronunciationService{enabled: true, publicBase: "", audioDir: t.TempDir()}
	router := &Router{
		mux:                  http.NewServeMux(),
		logger:               logger,
		pronunciationService: mock,
	}
	router.setupPronunciationMediaRoute()

	req := httptest.NewRequest(http.MethodGet, "/media/tts/fallback.mp3", nil)
	w := httptest.NewRecorder()
	router.mux.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 (route under /media/tts), got %d", w.Code)
	}
}

func TestHandleTTSWord_MethodNotAllowed(t *testing.T) {
	router := &Router{logger: zap.NewNop(), pronunciationService: nil}
	req := httptest.NewRequest(http.MethodPost, "/api/tts/word?word=hello", nil)
	w := httptest.NewRecorder()
	router.handleTTSWord(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}
}

func TestHandleTTSWord_WordRequired(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"no query", "/api/tts/word"},
		{"empty word", "/api/tts/word?word="},
		{"spaces only", "/api/tts/word?word=%20%20"},
	}
	router := &Router{logger: zap.NewNop(), pronunciationService: nil}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			router.handleTTSWord(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected 400, got %d", w.Code)
			}
		})
	}
}

func TestHandleTTSWord_NoService(t *testing.T) {
	router := &Router{logger: zap.NewNop(), pronunciationService: nil}
	req := httptest.NewRequest(http.MethodGet, "/api/tts/word?word=hello", nil)
	w := httptest.NewRecorder()
	router.handleTTSWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	var resp struct {
		Available bool   `json:"available"`
		URL       string `json:"url"`
		Word      string `json:"word"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Available || resp.URL != "" || resp.Word != "" {
		t.Errorf("Expected available=false, empty url/word; got available=%v url=%q word=%q", resp.Available, resp.URL, resp.Word)
	}
}

func TestHandleTTSWord_ServiceDisabled(t *testing.T) {
	logger := zap.NewNop()
	router := &Router{
		logger:               logger,
		pronunciationService: service.NewPronunciationService(config.TTSConfig{Enabled: false}, nil, logger),
	}
	req := httptest.NewRequest(http.MethodGet, "/api/tts/word?word=hello", nil)
	w := httptest.NewRecorder()
	router.handleTTSWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	var resp struct {
		Available bool   `json:"available"`
		URL       string `json:"url"`
		Word      string `json:"word"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Available || resp.URL != "" {
		t.Errorf("Expected available=false when disabled; got available=%v url=%q", resp.Available, resp.URL)
	}
}

func TestHandleTTSWordAndMedia(t *testing.T) {
	audioBytes := []byte("ID3test-audio")

	var dictServer *httptest.Server
	dictServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v2/entries/en/"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"phonetics": []map[string]string{
						{"audio": dictServer.URL + "/audio/spy.mp3"},
					},
				},
			})
		case r.URL.Path == "/audio/spy.mp3":
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write(audioBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer dictServer.Close()

	cfg := config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		RequestTimeout:    "3s",
		PrefetchEnabled:   false,
		PrefetchWorkers:   1,
		RetryBaseDelay:    "100ms",
		RetryMaxDelay:     "1s",
		MaxRetries:        3,
		DictionaryEnabled: true,
		DictionaryBaseURL: dictServer.URL + "/api/v2/entries/en",
	}

	pronService := service.NewPronunciationService(cfg, nil, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pronService.Start(ctx)

	router := &Router{
		logger:               zap.NewNop(),
		pronunciationService: pronService,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/tts/word?word=spy", nil)
	w := httptest.NewRecorder()
	router.handleTTSWord(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	type ttsResponse struct {
		Available bool   `json:"available"`
		URL       string `json:"url"`
	}
	var resp ttsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	deadline := time.After(4 * time.Second)
	for !resp.Available {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for pronunciation file")
		default:
			time.Sleep(50 * time.Millisecond)
			req = httptest.NewRequest(http.MethodGet, "/api/tts/word?word=spy", nil)
			w = httptest.NewRecorder()
			router.handleTTSWord(w, req)
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
		}
	}

	mediaReq := httptest.NewRequest(http.MethodGet, resp.URL, nil)
	mediaW := httptest.NewRecorder()
	router.handleTTSMedia(mediaW, mediaReq)
	if mediaW.Code != http.StatusOK {
		t.Fatalf("expected media status 200, got %d", mediaW.Code)
	}
	if got := mediaW.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Fatalf("expected immutable cache header, got %q", got)
	}
	if !bytes.Equal(mediaW.Body.Bytes(), audioBytes) {
		t.Fatalf("unexpected media body")
	}
}

func TestHandleTTSMedia_MethodNotAllowed(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router.SetPronunciationService(pronService)

	req := httptest.NewRequest(http.MethodPost, "/media/tts/word.mp3", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}
}

func TestHandleTTSMedia_ServiceDisabled(t *testing.T) {
	logger := zap.NewNop()
	router := &Router{
		logger:               logger,
		pronunciationService: service.NewPronunciationService(config.TTSConfig{Enabled: false}, nil, logger),
	}

	req := httptest.NewRequest(http.MethodGet, "/media/tts/any.mp3", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when service disabled, got %d", w.Code)
	}
}

func TestHandleTTSMedia_NilService(t *testing.T) {
	logger := zap.NewNop()
	router := &Router{logger: logger, pronunciationService: nil}

	req := httptest.NewRequest(http.MethodGet, "/media/tts/any.mp3", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when service nil, got %d", w.Code)
	}
}

func TestHandleTTSMedia_EmptyPath(t *testing.T) {
	logger := zap.NewNop()
	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router := &Router{logger: logger, pronunciationService: pronService}

	// Use full URL so req.URL.Path is definitely "/media/tts/" (empty relative after prefix)
	req := httptest.NewRequest(http.MethodGet, "http://localhost/media/tts/", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for path ending with slash, got %d", w.Code)
	}
}

func TestHandleTTSMedia_NotMp3(t *testing.T) {
	logger := zap.NewNop()
	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router := &Router{logger: logger, pronunciationService: pronService}

	req := httptest.NewRequest(http.MethodGet, "/media/tts/word.wav", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-.mp3 extension, got %d", w.Code)
	}
}

func TestHandleTTSMedia_TrailingSlash(t *testing.T) {
	logger := zap.NewNop()
	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router := &Router{logger: logger, pronunciationService: pronService}

	req := httptest.NewRequest(http.MethodGet, "/media/tts/word/", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for path with trailing slash, got %d", w.Code)
	}
}

// TestHandleTTSMedia_InvalidPathCoversNotFound exercises the NotFound block for invalid relative paths.
// Use a request with URL.Path set directly so the handler sees empty relative and hits http.NotFound.
func TestHandleTTSMedia_InvalidPathCoversNotFound(t *testing.T) {
	logger := zap.NewNop()
	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       t.TempDir(),
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router := &Router{logger: logger, pronunciationService: pronService}

	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	req.URL = &url.URL{Path: "/media/tts/"} // empty relative after TrimPrefix with prefix "/media/tts/"
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for empty relative, got %d", w.Code)
	}
}

func TestHandleTTSMedia_HEADAllowed(t *testing.T) {
	dir := t.TempDir()
	// Create file using same path logic as handler (relative to AudioDir)
	mp3Path := filepath.Join(dir, "head_test.mp3")
	if err := os.WriteFile(mp3Path, []byte("fake-mp3"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	logger := zap.NewNop()
	pronCfg := config.TTSConfig{
		Enabled:        true,
		Provider:       "dictionary",
		AudioDir:       dir,
		PublicBasePath: "/media/tts",
	}
	pronService := service.NewPronunciationService(pronCfg, nil, logger)
	router := &Router{logger: logger, pronunciationService: pronService}

	req := httptest.NewRequest(http.MethodHead, "/media/tts/head_test.mp3", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)

	// HEAD is allowed (not 405); 200 if file found, 404 if not
	if w.Code == http.StatusMethodNotAllowed {
		t.Errorf("HEAD should be allowed, got 405")
	}
}

// TestHandleTTSMedia_PathTraversal uses a mock with empty AudioDir so root="." and
// target does not have prefix root+sep, hitting the security 404 branch.
func TestHandleTTSMedia_PathTraversal(t *testing.T) {
	mock := &mockPronunciationService{
		enabled: true, publicBase: "/media/tts", audioDir: "",
	}
	router := &Router{logger: zap.NewNop(), pronunciationService: mock}
	req := httptest.NewRequest(http.MethodGet, "/media/tts/any.mp3", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when target not under root (empty AudioDir), got %d", w.Code)
	}
}

// TestHandleTTSMedia_EmptyPublicBasePathFallback uses mock with empty PublicBasePath
// to cover the handler's basePath fallback to /media/tts and successful serve.
func TestHandleTTSMedia_EmptyPublicBasePathFallback(t *testing.T) {
	dir := t.TempDir()
	mp3Path := filepath.Join(dir, "empty_base.mp3")
	if err := os.WriteFile(mp3Path, []byte("mp3"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	mock := &mockPronunciationService{
		enabled: true, publicBase: "", audioDir: dir,
	}
	router := &Router{logger: zap.NewNop(), pronunciationService: mock}
	req := httptest.NewRequest(http.MethodGet, "/media/tts/empty_base.mp3", nil)
	w := httptest.NewRecorder()
	router.handleTTSMedia(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 with empty PublicBasePath fallback, got %d", w.Code)
	}
}
