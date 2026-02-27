package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestNormalizePronunciationWord(t *testing.T) {
	tests := []struct {
		in   string
		want string
		ok   bool
	}{
		{in: "Spy", want: "spy", ok: true},
		{in: "to  Spy ", want: "to spy", ok: true},
		{in: "can't", want: "can't", ok: true},
		{in: "re-entry", want: "re-entry", ok: true},
		{in: "привет", want: "", ok: false},
		{in: "123", want: "", ok: false},
		{in: "spy!", want: "spy", ok: true},
	}

	for _, tt := range tests {
		got, ok := normalizePronunciationWord(tt.in)
		if ok != tt.ok {
			t.Fatalf("normalizePronunciationWord(%q) ok=%v want %v", tt.in, ok, tt.ok)
		}
		if got != tt.want {
			t.Fatalf("normalizePronunciationWord(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestPronunciationServiceLookupAndCache(t *testing.T) {
	audioBytes := []byte("ID3mock-audio")
	var lookupCalls int
	var audioCalls int

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v2/entries/en/"):
			lookupCalls++
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"` + srv.URL + `/audio/spy.mp3"}]}]`))
		case r.URL.Path == "/audio/spy.mp3":
			audioCalls++
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write(audioBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          tmpDir,
		PublicBasePath:    "/media/tts",
		RequestTimeout:    "3s",
		PrefetchEnabled:   false,
		PrefetchWorkers:   1,
		RetryBaseDelay:    "100ms",
		RetryMaxDelay:     "1s",
		MaxRetries:        2,
		DictionaryEnabled: true,
		DictionaryBaseURL: srv.URL + "/api/v2/entries/en",
	}

	logger := zap.NewNop()
	service := NewPronunciationService(cfg, nil, logger)
	if !service.IsEnabled() {
		t.Fatalf("service should be enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.Start(ctx)

	initial := service.Lookup("spy")
	if initial.Available {
		t.Fatalf("expected pronunciation to be initially unavailable")
	}

	waitFor := time.After(4 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var final PronunciationLookupResult
	found := false
	for !found {
		select {
		case <-waitFor:
			t.Fatalf("timed out waiting for pronunciation file")
		case <-ticker.C:
			final = service.Lookup("spy")
			found = final.Available
		}
	}

	if !strings.HasPrefix(final.URL, "/media/tts/") {
		t.Fatalf("unexpected pronunciation URL: %q", final.URL)
	}

	fullPath := filepath.Join(tmpDir, strings.TrimPrefix(final.URL, "/media/tts/"))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read cached pronunciation: %v", err)
	}
	if string(data) != string(audioBytes) {
		t.Fatalf("unexpected cached audio bytes")
	}
	if lookupCalls == 0 || audioCalls == 0 {
		t.Fatalf("expected dictionary and audio endpoints to be called at least once")
	}
}

func TestPronunciationServiceNoRetryOnNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

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
		DictionaryBaseURL: srv.URL + "/api/v2/entries/en",
	}

	service := NewPronunciationService(cfg, nil, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.Start(ctx)

	_ = service.ScheduleWord("missingword")

	time.Sleep(300 * time.Millisecond)
	service.mu.Lock()
	_, ok := service.retries["missingword"]
	service.mu.Unlock()
	if ok {
		t.Fatalf("expected no retry state for not found pronunciation")
	}
}

func TestOpenRouterPronunciationProviderFetch(t *testing.T) {
	audioBytes := []byte("ID3-openrouter-mp3")

	var gotAuth string
	var gotBody struct {
		Model      string          `json:"model"`
		Messages   []struct{ Role, Content string } `json:"messages"`
		Modalities []string        `json:"modalities"`
		Stream     bool            `json:"stream"`
		MaxTokens  int             `json:"max_tokens"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s (OpenRouter uses /chat/completions for TTS)", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		// One SSE chunk with audio data (base64 of mock MP3); test uses identity pcmToMp3 so no ffmpeg
		b64 := base64.StdEncoding.EncodeToString(audioBytes)
		evt := `data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `"},"content":[{"type":"text","text":"spy"}]}}]}` + "\n\n"
		_, _ = w.Write([]byte(evt))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:             srv.URL,
		model:               "openai/gpt-4o-audio-preview",
		voice:               "alloy",
		apiKey:              "test-key",
		client:              &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:            func(pcm []byte) ([]byte, error) { return pcm, nil },
	}

	audio, err := provider.fetch(context.Background(), "spy")
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}
	if string(audio) != string(audioBytes) {
		t.Fatalf("unexpected audio payload")
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("expected bearer auth header, got %q", gotAuth)
	}
	if gotBody.Model != "openai/gpt-4o-audio-preview" {
		t.Fatalf("expected model openai/gpt-4o-audio-preview, got %q", gotBody.Model)
	}
	if len(gotBody.Messages) < 1 || !strings.Contains(gotBody.Messages[0].Content, "Word: `spy`") {
		t.Fatalf("expected single user message with quoted word, got %+v", gotBody.Messages)
	}
	if !gotBody.Stream {
		t.Fatalf("expected stream=true")
	}
}

func TestOpenRouterPronunciationProviderFetch_TranscriptMismatch(t *testing.T) {
	audioBytes := []byte("ID3-openrouter-mp3")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		b64 := base64.StdEncoding.EncodeToString(audioBytes)
		evt := `data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `"},"content":[{"type":"text","text":"Hi there! How is it going?"}]}}]}` + "\n\n"
		_, _ = w.Write([]byte(evt))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "openai/gpt-4o-audio-preview",
		voice:                "alloy",
		apiKey:               "test-key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}

	_, err := provider.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected transcript mismatch error")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected errPronunciationNotFound, got %v", err)
	}
}

func TestOpenRouterPronunciationProviderFetch_MissingTranscript(t *testing.T) {
	audioBytes := []byte("ID3-openrouter-mp3")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		b64 := base64.StdEncoding.EncodeToString(audioBytes)
		evt := `data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `"}}}]}` + "\n\n"
		_, _ = w.Write([]byte(evt))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "openai/gpt-4o-audio-preview",
		voice:                "alloy",
		apiKey:               "test-key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}

	_, err := provider.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected missing transcript error")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected errPronunciationNotFound, got %v", err)
	}
}

func TestParseSSEAudioStream_CollectsAudioAndTranscript(t *testing.T) {
	rawAudio := []byte{0x00, 0x01, 0x02}
	b64 := base64.StdEncoding.EncodeToString(rawAudio)
	sse := strings.NewReader(
		`data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `","transcript":"he"}}}]}` + "\n\n" +
			`data: {"choices":[{"delta":{"content":[{"type":"text","text":"llo"}]}}]}` + "\n\n" +
			`data: [DONE]` + "\n\n",
	)

	pcm, transcript, err := parseSSEAudioStream(sse)
	if err != nil {
		t.Fatalf("parseSSEAudioStream error: %v", err)
	}
	if string(pcm) != string(rawAudio) {
		t.Fatalf("unexpected pcm payload: %v", pcm)
	}
	if transcript != "hello" {
		t.Fatalf("unexpected transcript: %q", transcript)
	}
}

func TestOpenRouterPronunciationProviderFetch_NonSSEErrorBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"Provider returned error","code":403,"metadata":{"raw":"{\"error\":{\"code\":\"unsupported_country_region_territory\"}}"}}}`))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "openai/gpt-audio-mini",
		voice:                "alloy",
		apiKey:               "test-key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}

	_, err := provider.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for non-SSE error body")
	}
	if errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected provider error (not errPronunciationNotFound), got %v", err)
	}
}

func TestPronunciationWordFileBase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "hello", want: "hello"},
		{in: "to spy", want: "to_spy"},
		{in: "can't", want: "can't"},
		{in: "re-entry", want: "re-entry"},
		{in: " hello   world ", want: "hello_world"},
	}

	for _, tt := range tests {
		got := pronunciationWordFileBase(tt.in)
		if got != tt.want {
			t.Fatalf("pronunciationWordFileBase(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}


func TestBuildPronunciationProvidersAuto_WithOpenRouter(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider:          "auto",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://api.dictionaryapi.dev/api/v2/entries/en",
		APIKey:            "test-key",
		Model:             "openai/gpt-audio-mini",
		BaseURL:           "https://openrouter.ai/api/v1",
		Voice:             "alloy",
	}, zap.NewNop())

	if len(providers) != 2 {
		t.Fatalf("expected 2 providers (dictionary + openrouter), got %d", len(providers))
	}
	if providers[0].name() != "dictionary" {
		t.Fatalf("expected first provider dictionary, got %q", providers[0].name())
	}
	if providers[1].name() != "openrouter" {
		t.Fatalf("expected second provider openrouter, got %q", providers[1].name())
	}
}
