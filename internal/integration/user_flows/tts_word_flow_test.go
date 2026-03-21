//go:build integration

package user_flows

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
	"tgbot-skeleton/internal/integration/testkit"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func TestTTSWordFlow_CacheAndServeMedia(t *testing.T) {
	audioBytes := []byte("ID3-integration-audio")

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

	h := testkit.NewHarness(t)
	wordRepo := repository.NewWordRepository(h.GetConnection().GetConnection(), zap.NewNop())
	pronCfg := config.TTSConfig{
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

	pronService := service.NewPronunciationService(pronCfg, config.DefaultLearningConfig(), wordRepo, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go pronService.Start(ctx)
	h.Router.SetPronunciationService(pronService)

	token := h.AuthAsUser(77770001)

	type ttsResponse struct {
		Available bool   `json:"available"`
		URL       string `json:"url"`
	}

	fetch := func() (ttsResponse, int) {
		req := httptest.NewRequest(http.MethodGet, "/api/tts/word?word=spy", nil)
		req.Header.Set("Authorization", token)
		w := httptest.NewRecorder()
		h.Router.ServeHTTP(w, req)

		var resp ttsResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return resp, w.Code
	}

	resp, status := fetch()
	if status != http.StatusOK {
		t.Fatalf("expected 200 for /api/tts/word, got %d", status)
	}

	deadline := time.Now().Add(4 * time.Second)
	for !resp.Available && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
		resp, status = fetch()
		if status != http.StatusOK {
			t.Fatalf("expected 200 for /api/tts/word, got %d", status)
		}
	}
	if !resp.Available || resp.URL == "" {
		t.Fatalf("pronunciation did not become available in time: %+v", resp)
	}

	mediaReq := httptest.NewRequest(http.MethodGet, resp.URL, nil)
	mediaW := httptest.NewRecorder()
	h.Router.ServeHTTP(mediaW, mediaReq)
	if mediaW.Code != http.StatusOK {
		t.Fatalf("expected 200 for media url, got %d", mediaW.Code)
	}
	if cache := mediaW.Header().Get("Cache-Control"); !strings.Contains(cache, "immutable") {
		t.Fatalf("expected immutable cache header, got %q", cache)
	}
	if !bytes.Equal(mediaW.Body.Bytes(), audioBytes) {
		t.Fatalf("unexpected media body")
	}
}
