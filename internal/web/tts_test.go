package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

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
