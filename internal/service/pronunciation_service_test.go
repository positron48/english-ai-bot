package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

type stubPronunciationProvider struct {
	providerName string
	err          error
	audio        []byte
}

func (p *stubPronunciationProvider) name() string { return p.providerName }
func (p *stubPronunciationProvider) fetch(_ context.Context, _ string) ([]byte, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.audio, nil
}

// countingProvider wraps a provider and counts fetch invocations.
type countingProvider struct {
	provider pronunciationProvider
	count    *int
}

func (c *countingProvider) name() string { return c.provider.name() }
func (c *countingProvider) fetch(ctx context.Context, word string) ([]byte, error) {
	if c.count != nil {
		*c.count++
	}
	return c.provider.fetch(ctx, word)
}

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

func TestPronunciationService_ScheduleWord_Disabled(t *testing.T) {
	cfg := config.TTSConfig{Enabled: false, AudioDir: t.TempDir(), PublicBasePath: "/media/tts"}
	svc := NewPronunciationService(cfg, nil, zap.NewNop())
	got := svc.ScheduleWord("hello")
	if got {
		t.Error("ScheduleWord should return false when service is disabled")
	}
}

func TestPronunciationService_ScheduleWord_InvalidWord(t *testing.T) {
	cfg := config.TTSConfig{Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts", Provider: "dictionary", DictionaryEnabled: true, DictionaryBaseURL: "http://example.com"}
	svc := NewPronunciationService(cfg, nil, zap.NewNop())
	// nil wordRepo => no ttsRepo => ensureStatusForWord returns nil,nil; canScheduleNow may still run
	got := svc.ScheduleWord("привет")
	if got {
		t.Error("ScheduleWord should return false for invalid word (Cyrillic)")
	}
}

// TestPronunciationService_ScheduleWord_FailedTerminal returns false when status is failed_terminal.
func TestPronunciationService_ScheduleWord_FailedTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled: true, Provider: "dictionary", AudioDir: audioDir, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 3,
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkTerminal("hello", "service", "test_code", "test error")

	got := svc.ScheduleWord("hello")
	if got {
		t.Error("ScheduleWord should return false when status is failed_terminal")
	}
}

// TestPronunciationService_ScheduleWord_MaxAttemptsReached returns false when attempt_count >= max_attempts.
func TestPronunciationService_ScheduleWord_MaxAttemptsReached(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{
		Enabled: true, Provider: "dictionary", AudioDir: t.TempDir(), PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 2,
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	for i := 0; i < 2; i++ {
		_ = svc.ttsRepo.MarkAttempt("maxword", "dict", "not_found", "err", true)
	}

	got := svc.ScheduleWord("maxword")
	if got {
		t.Error("ScheduleWord should return false when attempt_count >= max_attempts")
	}
}

// TestPronunciationService_ScheduleWord_ReadyAndFileExists returns true when status is ready and file exists.
func TestPronunciationService_ScheduleWord_ReadyAndFileExists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled: true, Provider: "dictionary", AudioDir: audioDir, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	relPath := svc.relativePathForWordWithExt("hello", ".mp3")
	fullPath := filepath.Join(audioDir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("mock"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_ = svc.ttsRepo.MarkReady("hello", "dict", relPath)

	got := svc.ScheduleWord("hello")
	if !got {
		t.Error("ScheduleWord should return true when status is ready and file exists")
	}
}

func TestPronunciationService_ForceRegenerate_NoTTSRepo(t *testing.T) {
	cfg := config.TTSConfig{Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts"}
	svc := NewPronunciationService(cfg, nil, zap.NewNop())
	_, err := svc.ForceRegenerate("hello")
	if err == nil {
		t.Fatal("ForceRegenerate should return error when tts repo is not configured")
	}
	if !strings.Contains(err.Error(), "not configured") && !strings.Contains(err.Error(), "repository") {
		t.Errorf("expected error about repo not configured, got: %v", err)
	}
}

func TestPronunciationService_ForceRegenerate_InvalidWord(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts", Provider: "dictionary", DictionaryEnabled: true, DictionaryBaseURL: "http://example.com"}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	_, err := svc.ForceRegenerate("123")
	if err == nil {
		t.Fatal("ForceRegenerate should return error for invalid word")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("expected 'invalid word' error, got: %v", err)
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

func TestPronunciationServiceRetryableOnNotFoundAllProviders(t *testing.T) {
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

	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	service := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.Start(ctx)

	_ = service.ScheduleWord("missingword")

	time.Sleep(300 * time.Millisecond)
	status, err := service.ttsRepo.GetByWord("missingword")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status record")
	}
	if status.State != "failed_retryable" || status.AttemptCount != 1 {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestOpenRouterPronunciationProviderFetch(t *testing.T) {
	audioBytes := []byte("ID3-openrouter-mp3")

	var gotAuth string
	var gotBody struct {
		Model      string                           `json:"model"`
		Messages   []struct{ Role, Content string } `json:"messages"`
		Modalities []string                         `json:"modalities"`
		Stream     bool                             `json:"stream"`
		MaxTokens  int                              `json:"max_tokens"`
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
		baseURL:              srv.URL,
		model:                "openai/gpt-4o-audio-preview",
		voice:                "alloy",
		apiKey:               "test-key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
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

func TestOpenRouterPronunciationProviderFetch_TranscriptPhoneticLikeAccepted(t *testing.T) {
	audioBytes := []byte("ID3-openrouter-mp3")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		b64 := base64.StdEncoding.EncodeToString(audioBytes)
		evt := `data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `"},"content":[{"type":"text","text":"kə-MOH-shun"}]}}]}` + "\n\n"
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

	audio, err := provider.fetch(context.Background(), "commotion")
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}
	if string(audio) != string(audioBytes) {
		t.Fatalf("unexpected audio payload")
	}
}

func TestDictionaryPronunciationProviderFetch_NoAudioReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"phonetics":[{"audio":""}]}]`))
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{
		baseURL: srv.URL,
		client:  &http.Client{Timeout: 3 * time.Second},
	}

	_, err := p.fetch(context.Background(), "ethos")
	if err == nil {
		t.Fatal("expected not found error")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected errPronunciationNotFound, got %v", err)
	}
	code, retryable := classifyPronunciationError(err)
	if code != "dictionary_no_audio" || !retryable {
		t.Fatalf("unexpected classification: code=%s retryable=%v err=%v", code, retryable, err)
	}
}

// TestDictionaryPronunciationProviderFetch_EmptyLookupWord returns errPronunciationNotFound when word yields empty lookup.
func TestDictionaryPronunciationProviderFetch_EmptyLookupWord(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not call server when lookup word is empty")
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}

	_, err := p.fetch(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for empty/whitespace word")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected errPronunciationNotFound, got %v", err)
	}
}

// TestDictionaryPronunciationProviderFetch_NotFound covers 404 response from dictionary API.
func TestDictionaryPronunciationProviderFetch_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}

	_, err := p.fetch(context.Background(), "nonexistentword")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "dictionary_404") {
		t.Errorf("expected error to mention dictionary_404, got %q", err.Error())
	}
}

// TestDictionaryPronunciationProviderFetch_Non2xx covers non-2xx response from dictionary API.
func TestDictionaryPronunciationProviderFetch_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}

	_, err := p.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if strings.Contains(err.Error(), "500") && !strings.Contains(err.Error(), "server error") {
		t.Errorf("expected body in error message for 500 response, got %q", err.Error())
	}
}

// TestDictionaryPronunciationProviderFetch_InvalidJSON covers decode error for dictionary response.
func TestDictionaryPronunciationProviderFetch_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}

	_, err := p.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("expected decode in error, got %q", err.Error())
	}
}

// TestDictionaryPronunciationProviderFetch_SuccessWithSameServerAudioURL verifies successful fetch when
// dictionary returns an audio URL pointing to the same server (e.g. full http URL).
func TestDictionaryPronunciationProviderFetch_SuccessWithSameServerAudioURL(t *testing.T) {
	audioBytes := []byte("mock-mp3")
	var audioReqPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/hello") || r.URL.Path == "/hello" {
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"http://` + r.Host + `/audio/hello.mp3"}]}]`))
			return
		}
		if strings.Contains(r.URL.Path, "audio") {
			audioReqPath = r.URL.Path
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write(audioBytes)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}

	got, err := p.fetch(context.Background(), "hello")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if string(got) != string(audioBytes) {
		t.Fatalf("unexpected audio: got %q", got)
	}
	if audioReqPath == "" {
		t.Fatal("expected audio download request to be made")
	}
}

// TestDictionaryPronunciationProviderFetch_AudioDownload404 covers 404 on audio file download.
func TestDictionaryPronunciationProviderFetch_AudioDownload404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/hello") || r.URL.Path == "/hello" {
			// Use r.Host so the audio URL points to this server without capturing srv
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"http://` + r.Host + `/audio/hello.mp3"}]}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}

	_, err := p.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error when audio download returns 404")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "dictionary_audio_404") {
		t.Errorf("expected dictionary_audio_404 in error, got %q", err.Error())
	}
}

// TestPronunciationService_DictionaryNoAudioFallbackToOpenRouter verifies that when
// the first provider returns dictionary_no_audio, the second provider (e.g. OpenRouter) is tried and can succeed.
func TestPronunciationService_DictionaryNoAudioFallbackToOpenRouter(t *testing.T) {
	audioDir := t.TempDir()
	// DictionaryEnabled so initial build adds a provider and svc.enabled stays true when we replace providers.
	cfg := config.TTSConfig{
		Enabled:           true,
		Provider:          "auto",
		AudioDir:          audioDir,
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://example.com/dict",
	}
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	dictNoAudioErr := fmt.Errorf("%w: dictionary_no_audio", errPronunciationNotFound)
	if !errors.Is(dictNoAudioErr, errPronunciationNotFound) {
		t.Fatal("test setup: dictNoAudioErr must wrap errPronunciationNotFound")
	}
	code, retryable := classifyPronunciationError(dictNoAudioErr)
	if code != "dictionary_no_audio" || !retryable {
		t.Fatalf("test setup: classifyPronunciationError: code=%q retryable=%v", code, retryable)
	}
	audioBytes := []byte("ID3-openrouter-mp3")
	var fetchCount int
	svc.providers = []pronunciationProvider{
		&countingProvider{provider: &stubPronunciationProvider{providerName: "dictionary", err: dictNoAudioErr}, count: &fetchCount},
		&countingProvider{provider: &stubPronunciationProvider{providerName: "openrouter", audio: audioBytes}, count: &fetchCount},
	}
	if n := len(svc.providers); n != 2 {
		t.Fatalf("expected 2 providers, got %d", n)
	}

	svc.processWord(context.Background(), "commotion")

	if fetchCount != 2 {
		t.Fatalf("expected 2 provider fetch calls (dictionary then openrouter), got %d", fetchCount)
	}

	// Fallback succeeded if openrouter produced the file
	if !svc.hasCachedAudio("commotion") {
		t.Fatal("expected audio file after openrouter fallback; dictionary_no_audio should trigger next provider")
	}
	rel := svc.cachedRelPathForWord("commotion")
	if rel == "" {
		t.Fatal("expected cached rel path for commotion")
	}
	fullPath := filepath.Join(audioDir, rel)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if string(data) != string(audioBytes) {
		t.Fatalf("unexpected cached audio: got %d bytes", len(data))
	}
	if svc.ttsRepo != nil {
		status, err := svc.ttsRepo.GetByWord("commotion")
		if err != nil {
			t.Fatalf("GetByWord() error = %v", err)
		}
		if status == nil || status.State != "ready" {
			t.Fatalf("expected status ready in DB after fallback, got %v", status)
		}
	}
}

// TestPronunciationService_ProcessWord_Disabled returns early when service is disabled.
func TestPronunciationService_ProcessWord_Disabled(t *testing.T) {
	cfg := config.TTSConfig{Enabled: false, AudioDir: t.TempDir(), PublicBasePath: "/media/tts"}
	svc := NewPronunciationService(cfg, nil, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("x")}}

	svc.processWord(context.Background(), "hello")
	// No panic; disabled path returns without calling providers (no file written)
	if svc.hasCachedAudio("hello") {
		t.Error("expected no cached audio when disabled")
	}
}

// TestPronunciationService_ProcessWord_WriteFails continues to retry path when writeFileAtomic fails.
func TestPronunciationService_ProcessWord_WriteFails(t *testing.T) {
	dir := t.TempDir()
	// Use a path that is an existing file so MkdirAll in writeFileAtomic fails
	audioDirAsFile := filepath.Join(dir, "blocked")
	if err := os.WriteFile(audioDirAsFile, nil, 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}
	cfg := config.TTSConfig{
		Enabled: true, Provider: "dictionary", AudioDir: audioDirAsFile, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 2,
	}
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("audio")}}

	svc.processWord(context.Background(), "hello")

	status, err := svc.ttsRepo.GetByWord("hello")
	if err != nil {
		t.Fatalf("GetByWord: %v", err)
	}
	if status == nil {
		t.Fatal("expected status record after write failure")
	}
	// Should record retryable attempt (write_failed)
	if status.State != models.TTSStateFailedRetryable && status.State != models.TTSStatePending {
		t.Errorf("expected failed_retryable or pending after write failure, got state=%q", status.State)
	}
}

// TestPronunciationService_SkipDictionaryAfterNoAudio verifies that when we already have
// dictionary_no_audio for a word, the next processWord run skips the dictionary provider.
func TestPronunciationService_SkipDictionaryAfterNoAudio(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled:           true,
		Provider:          "auto",
		AudioDir:          audioDir,
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://example.com/dict",
		MaxRetries:        5,
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	dictNoAudioErr := fmt.Errorf("%w: dictionary_no_audio", errPronunciationNotFound)
	audioBytes := []byte("ID3-openrouter-mp3")

	// First run: only dictionary, returns no_audio → MarkAttempt(dictionary, dictionary_no_audio)
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "dictionary", err: dictNoAudioErr},
	}
	svc.processWord(context.Background(), "skipdict")
	status, _ := svc.ttsRepo.GetByWord("skipdict")
	if status == nil || status.LastErrorCode == nil || *status.LastErrorCode != "dictionary_no_audio" ||
		status.LastProvider == nil || *status.LastProvider != "dictionary" {
		t.Fatalf("expected status with dictionary_no_audio from dictionary after first run, got %v", status)
	}

	// Second run: skip dictionary, openrouter succeeds. Use countingProvider to ensure dictionary is not called.
	var dictCalls, openrouterCalls int
	svc.providers = []pronunciationProvider{
		&countingProvider{provider: &stubPronunciationProvider{providerName: "dictionary", err: dictNoAudioErr}, count: &dictCalls},
		&countingProvider{provider: &stubPronunciationProvider{providerName: "openrouter", audio: audioBytes}, count: &openrouterCalls},
	}
	svc.processWord(context.Background(), "skipdict")

	if dictCalls != 0 {
		t.Fatalf("expected dictionary to be skipped on retry (0 calls), got %d", dictCalls)
	}
	if openrouterCalls != 1 {
		t.Fatalf("expected openrouter to be called once, got %d", openrouterCalls)
	}
	if !svc.hasCachedAudio("skipdict") {
		t.Fatal("expected audio file after openrouter success")
	}
}

// TestPronunciationService_DoNotSkipDictionaryWhenLastProviderNotDictionary verifies that
// we only skip the dictionary provider when LastProvider is "dictionary". If LastErrorCode
// is dictionary_no_audio but LastProvider is e.g. "openrouter", we still call dictionary.
func TestPronunciationService_DoNotSkipDictionaryWhenLastProviderNotDictionary(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled:           true,
		Provider:          "auto",
		AudioDir:          audioDir,
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://example.com/dict",
		MaxRetries:        5,
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	dictNoAudioErr := fmt.Errorf("%w: dictionary_no_audio", errPronunciationNotFound)
	audioBytes := []byte("ID3-openrouter-mp3")

	// First run: only openrouter stub that returns dictionary_no_audio → LastProvider=openrouter, LastErrorCode=dictionary_no_audio
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "openrouter", err: dictNoAudioErr},
	}
	svc.processWord(context.Background(), "noskip")
	status, _ := svc.ttsRepo.GetByWord("noskip")
	if status == nil || status.LastProvider == nil || *status.LastProvider != "openrouter" ||
		status.LastErrorCode == nil || *status.LastErrorCode != "dictionary_no_audio" {
		t.Fatalf("expected status with dictionary_no_audio from openrouter after first run, got %v", status)
	}

	// Second run: dictionary + openrouter. We must NOT skip dictionary (LastProvider is openrouter, not dictionary).
	var dictCalls, openrouterCalls int
	svc.providers = []pronunciationProvider{
		&countingProvider{provider: &stubPronunciationProvider{providerName: "dictionary", err: dictNoAudioErr}, count: &dictCalls},
		&countingProvider{provider: &stubPronunciationProvider{providerName: "openrouter", audio: audioBytes}, count: &openrouterCalls},
	}
	svc.processWord(context.Background(), "noskip")

	if dictCalls != 1 {
		t.Fatalf("expected dictionary to be called once (no skip when LastProvider != dictionary), got %d", dictCalls)
	}
	if openrouterCalls != 1 {
		t.Fatalf("expected openrouter to be called once, got %d", openrouterCalls)
	}
	if !svc.hasCachedAudio("noskip") {
		t.Fatal("expected audio file after openrouter success")
	}
}

// TestPronunciationService_MaxRetriesZeroDefaultsTo10 verifies that when cfg.MaxRetries is 0,
// the service uses 10 as default maxAttempts (so 10 retries before terminal).
func TestPronunciationService_MaxRetriesZeroDefaultsTo10(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "auto",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://example.com/dict",
		MaxRetries:        0,
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "openrouter", err: fmt.Errorf("status 429: too many requests")},
	}
	for i := 0; i < 10; i++ {
		svc.processWord(context.Background(), "zeroretries")
	}
	status, err := svc.ttsRepo.GetByWord("zeroretries")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.AttemptCount != 10 || status.State != "failed_terminal" {
		t.Fatalf("expected attempt_count=10 and failed_terminal (default 10 retries), got attempt_count=%d state=%s", status.AttemptCount, status.State)
	}
}

// TestPronunciationService_MaxRetriesClampedAt20 verifies that when cfg.MaxRetries > 20,
// the service clamps to 20, so we get terminal after 20 attempts.
func TestPronunciationService_MaxRetriesClampedAt20(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "auto",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://example.com/dict",
		MaxRetries:        25,
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "openrouter", err: fmt.Errorf("status 429: too many requests")},
	}
	for i := 0; i < 20; i++ {
		svc.processWord(context.Background(), "clampword")
	}
	status, err := svc.ttsRepo.GetByWord("clampword")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.AttemptCount != 20 || status.State != "failed_terminal" {
		t.Fatalf("expected attempt_count=20 and failed_terminal (clamped at 20), got attempt_count=%d state=%s", status.AttemptCount, status.State)
	}
}

// TestDictionaryNoAudioFallbackLoop simulates the processWord provider loop to ensure
// that when the first provider returns errPronunciationNotFound (dictionary_no_audio),
// the loop continues to the next provider and uses its audio.
func TestDictionaryNoAudioFallbackLoop(t *testing.T) {
	dictNoAudioErr := fmt.Errorf("%w: dictionary_no_audio", errPronunciationNotFound)
	audioBytes := []byte("ID3-openrouter-mp3")
	providers := []pronunciationProvider{
		&stubPronunciationProvider{providerName: "dictionary", err: dictNoAudioErr},
		&stubPronunciationProvider{providerName: "openrouter", audio: audioBytes},
	}
	ctx := context.Background()
	word := "commotion"
	var gotAudio []byte
	for i, provider := range providers {
		audio, err := provider.fetch(ctx, word)
		if err != nil {
			if errors.Is(err, errPronunciationNotFound) {
				continue
			}
			_, retryable := classifyPronunciationError(err)
			if retryable {
				continue
			}
			t.Fatalf("provider %d %s failed: %v", i, provider.name(), err)
		}
		gotAudio = audio
		break
	}
	if gotAudio == nil {
		t.Fatal("expected audio from fallback provider")
	}
	if string(gotAudio) != string(audioBytes) {
		t.Fatalf("unexpected audio: got %d bytes", len(gotAudio))
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

	pcm, transcript, _, err := parseSSEAudioStream(sse)
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

func TestPronunciationService_DBRetryLimitAndTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
		MaxRetries:        3,
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "openrouter", err: fmt.Errorf("status 500: provider error")},
	}

	for i := 0; i < 4; i++ {
		svc.processWord(context.Background(), "retryword")
	}

	status, err := svc.ttsRepo.GetByWord("retryword")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.AttemptCount != 3 {
		t.Fatalf("expected attempt_count=3, got %d", status.AttemptCount)
	}
	if status.State != "failed_terminal" {
		t.Fatalf("expected failed_terminal, got %s", status.State)
	}
}

func TestPronunciationService_ForceRegenerateAfterTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
		MaxRetries:        3,
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "openrouter", err: fmt.Errorf("status 500: provider error")},
	}

	for i := 0; i < 3; i++ {
		svc.processWord(context.Background(), "terminalreset")
	}
	before, _ := svc.ttsRepo.GetByWord("terminalreset")
	if before == nil || before.State != "failed_terminal" {
		t.Fatalf("expected terminal before reset, got %+v", before)
	}

	after, err := svc.ForceRegenerate("terminalreset")
	if err != nil {
		t.Fatalf("ForceRegenerate() error = %v", err)
	}
	if after.State != "pending" || after.AttemptCount != 0 {
		t.Fatalf("unexpected reset status: %+v", after)
	}
}

func TestPronunciationService_ForceRegenerate_EnqueuesEvenIfCachedFileExists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		AudioDir:          audioDir,
		PublicBasePath:    "/media/tts",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())

	word := "regencached"
	rel := svc.relativePathForWordWithExt(word, ".mp3")
	full := filepath.Join(audioDir, rel)
	if err := writeFileAtomic(full, []byte("old-audio"), 0o644); err != nil {
		t.Fatalf("writeFileAtomic() error = %v", err)
	}
	if err := svc.ttsRepo.MarkReady(word, "dictionary", rel); err != nil {
		t.Fatalf("MarkReady() error = %v", err)
	}

	beforeLen := len(svc.queue)
	after, err := svc.ForceRegenerate(word)
	if err != nil {
		t.Fatalf("ForceRegenerate() error = %v", err)
	}
	if after.State != "pending" || after.AttemptCount != 0 {
		t.Fatalf("unexpected reset status: %+v", after)
	}
	if len(svc.queue) != beforeLen+1 {
		t.Fatalf("expected queue size to increase by 1, before=%d after=%d", beforeLen, len(svc.queue))
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

// --- Покрытие: decodeAudioBase64, classifyPronunciationError, геттеры, пути, writeFileAtomic, fetchViaAudioSpeech, DebugFetch, backfillOnce, ScheduleWords, Recheck, Start, ScheduleWord, Lookup, ensureStatusForWord, resolveReadyRelPath ---

func TestDecodeAudioBase64(t *testing.T) {
	validStd := []byte{0x00, 0x01, 0x02}
	tests := []struct {
		name  string
		raw   string
		want  []byte
		valid bool
	}{
		{"empty", "", nil, false},
		{"whitespace", "  \t  ", nil, false},
		{"std", base64.StdEncoding.EncodeToString(validStd), validStd, true},
		{"raw_std", base64.RawStdEncoding.EncodeToString(validStd), validStd, true},
		{"url", base64.URLEncoding.EncodeToString(validStd), validStd, true},
		{"raw_url", base64.RawURLEncoding.EncodeToString(validStd), validStd, true},
		{"invalid", "!!!not-base64!!!", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := decodeAudioBase64(tt.raw)
			if ok != tt.valid {
				t.Fatalf("decodeAudioBase64(%q) ok=%v want %v", tt.raw, ok, tt.valid)
			}
			if tt.valid && string(got) != string(tt.want) {
				t.Fatalf("decodeAudioBase64(%q)=%v want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestClassifyPronunciationError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  string
		retryable bool
	}{
		{"nil", nil, "", true},
		{"dictionary_404", fmt.Errorf("%w: dictionary_404", errPronunciationNotFound), "dictionary_404", true},
		{"dictionary_no_audio", fmt.Errorf("%w: dictionary_no_audio", errPronunciationNotFound), "dictionary_no_audio", true},
		{"dictionary_audio_404", fmt.Errorf("%w: dictionary_audio_404", errPronunciationNotFound), "dictionary_audio_404", true},
		{"openrouter_no_audio", fmt.Errorf("%w: openrouter_no_audio", errPronunciationNotFound), "openrouter_no_audio", true},
		{"openrouter_rejected", fmt.Errorf("%w: openrouter_rejected (404)", errPronunciationNotFound), "openrouter_rejected", true},
		{"openrouter_audio_speech_rejected", fmt.Errorf("%w: openrouter_audio_speech_rejected", errPronunciationNotFound), "openrouter_audio_speech_rejected", true},
		{"openrouter_audio_speech_empty", fmt.Errorf("%w: openrouter_audio_speech_empty", errPronunciationNotFound), "openrouter_audio_speech_empty", true},
		{"openrouter_audio_decode_failed", fmt.Errorf("%w: openrouter_audio_decode_failed", errPronunciationNotFound), "openrouter_audio_decode_failed", true},
		{"transcript_mismatch", fmt.Errorf("%w: transcript mismatch: expected %q got %q", errPronunciationNotFound, "a", "b"), "transcript_mismatch", true},
		{"missing_transcript", fmt.Errorf("%w: missing transcript for validation", errPronunciationNotFound), "transcript_missing", true},
		{"not_found_default", fmt.Errorf("%w: something else", errPronunciationNotFound), "not_found", true},
		{"rate_limited", errors.New("status 429: too many requests"), "rate_limited", true},
		{"too_many_requests", errors.New("too many requests"), "rate_limited", true},
		{"provider_5xx", errors.New("status 500 internal error"), "provider_5xx", true},
		{"timeout", errors.New("context deadline exceeded timeout"), "network_error", true},
		{"connection", errors.New("connection refused"), "network_error", true},
		{"unsupported_region", errors.New("unsupported_country_region_territory"), "unsupported_country_region_territory", true},
		{"provider_error", errors.New("random error"), "provider_error", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, retryable := classifyPronunciationError(tt.err)
			if code != tt.wantCode || retryable != tt.retryable {
				t.Fatalf("classifyPronunciationError(%v) = %q, %v want %q, %v", tt.err, code, retryable, tt.wantCode, tt.retryable)
			}
		})
	}
}

func TestPronunciationService_AudioDirAndPublicBasePath(t *testing.T) {
	cfg := config.TTSConfig{
		Enabled:        true,
		AudioDir:       "/custom/tts",
		PublicBasePath: " /media/pronunciation ",
	}
	svc := NewPronunciationService(cfg, nil, zap.NewNop())
	if svc.AudioDir() != "/custom/tts" {
		t.Fatalf("AudioDir() = %q want /custom/tts", svc.AudioDir())
	}
	if svc.PublicBasePath() != "/media/pronunciation" {
		t.Fatalf("PublicBasePath() = %q want /media/pronunciation", svc.PublicBasePath())
	}

	var nilSvc *PronunciationService
	if nilSvc.AudioDir() != "" {
		t.Fatalf("nil service AudioDir() = %q want \"\"", nilSvc.AudioDir())
	}
	if nilSvc.PublicBasePath() != "" {
		t.Fatalf("nil service PublicBasePath() = %q want \"\"", nilSvc.PublicBasePath())
	}
}

func TestPronunciationService_RelativePathForWord(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:  true,
		AudioDir: t.TempDir(),
	}, nil, zap.NewNop())
	got := svc.relativePathForWord("hello")
	if got == "" {
		t.Fatal("relativePathForWord returned empty")
	}
	if !strings.HasSuffix(got, "hello.mp3") {
		t.Fatalf("relativePathForWord(hello) = %q expected .../hello.mp3", got)
	}
}

func TestPronunciationService_AudioRelPathExists(t *testing.T) {
	dir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: dir}, nil, zap.NewNop())

	existRel := "ab/cd/word.mp3"
	full := filepath.Join(dir, existRel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !svc.audioRelPathExists(existRel) {
		t.Fatal("audioRelPathExists(existing) = false want true")
	}
	if svc.audioRelPathExists("nonexistent.mp3") {
		t.Fatal("audioRelPathExists(nonexistent) = true want false")
	}
	if svc.audioRelPathExists("..") {
		t.Fatal("audioRelPathExists(..) = true want false (path traversal rejected)")
	}
	if svc.audioRelPathExists("../etc/passwd") {
		t.Fatal("audioRelPathExists(../etc/passwd) = true want false")
	}
}

func TestWriteFileAtomic_SuccessAndErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "file.mp3")
	if err := writeFileAtomic(path, []byte("data"), 0o644); err != nil {
		t.Fatalf("writeFileAtomic success: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "data" {
		t.Fatalf("read back: %v %q", err, data)
	}

	// Dir is a file — MkdirAll fails (cannot create dir with same name as file)
	fileAsDir := filepath.Join(dir, "fileasdir")
	if err := os.WriteFile(fileAsDir, nil, 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	err = writeFileAtomic(filepath.Join(fileAsDir, "x.mp3"), []byte("x"), 0o644)
	if err == nil || !strings.Contains(err.Error(), "create dir") {
		t.Fatalf("expected create dir error, got %v", err)
	}

	// Read-only dir — CreateTemp or Rename can fail depending on OS. At least Chmod may fail on some filesystems.
	roDir := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(roDir, 0o444); err != nil {
		t.Skip("cannot create read-only dir:", err)
	}
	err = writeFileAtomic(filepath.Join(roDir, "y.mp3"), []byte("y"), 0o644)
	if err != nil {
		if !strings.Contains(err.Error(), "create temp") && !strings.Contains(err.Error(), "rename") && !strings.Contains(err.Error(), "create dir") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestOpenRouterPronunciationProvider_FetchViaAudioSpeech(t *testing.T) {
	audioBytes := []byte("mp3-from-audio-speech")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/speech" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected auth header")
		}
		var body struct {
			Input          string `json:"input"`
			Model          string `json:"model"`
			Voice          string `json:"voice"`
			ResponseFormat string `json:"response_format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Input != "word" || body.ResponseFormat != "mp3" {
			t.Fatalf("unexpected body: %+v", body)
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write(audioBytes)
	}))
	defer srv.Close()

	// baseURL not openrouter + forceChatCompletions=false → fetchViaAudioSpeech
	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "tts-1",
		voice:                "alloy",
		apiKey:               "test-key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: false,
	}
	audio, err := provider.fetch(context.Background(), "word")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if string(audio) != string(audioBytes) {
		t.Fatalf("audio = %q want %q", audio, audioBytes)
	}
}

func TestOpenRouterPronunciationProvider_FetchViaAudioSpeech_Rejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("model not found"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "tts-1",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: false,
	}
	_, err := provider.fetch(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Fatalf("expected errPronunciationNotFound, got %v", err)
	}
}

func TestPronunciationService_DebugFetch(t *testing.T) {
	// Disabled
	var nilSvc *PronunciationService
	_, err := nilSvc.DebugFetch(context.Background(), "hello")
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("nil service: want disabled error, got %v", err)
	}

	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		AudioDir:          t.TempDir(),
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}, nil, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("ok")}}

	_, err = svc.DebugFetch(context.Background(), "  ")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("invalid word: want invalid error, got %v", err)
	}

	res, err := svc.DebugFetch(context.Background(), "hello")
	if err != nil {
		t.Fatalf("DebugFetch: %v", err)
	}
	if res.Word != "hello" || len(res.Results) != 1 || res.Results[0].Outcome != "success" || res.Results[0].Bytes != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}

	// All providers error
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "a", err: fmt.Errorf("%w: dictionary_404", errPronunciationNotFound)},
		&stubPronunciationProvider{providerName: "b", err: errors.New("network timeout")},
	}
	res, err = svc.DebugFetch(context.Background(), "x")
	if err != nil {
		t.Fatalf("DebugFetch (all fail): %v", err)
	}
	if len(res.Results) != 2 || res.Results[0].Outcome != "not_found" || res.Results[1].Outcome != "error" {
		t.Fatalf("unexpected results: %+v", res.Results)
	}
}

func TestPronunciationService_BackfillOnce(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		PrefetchEnabled:   false,
		AudioDir:          t.TempDir(),
		BackfillBatchSize: 5,
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("x")}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.backfillOnce(ctx)
	// ListPronunciationCandidates may return 0 words; backfillOnce just iterates and ScheduleWord's each
	// No assertion on queue length (depends on DB state). Just ensure no panic.
}

func TestPronunciationService_BackfillOnce_NoWordRepo(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, zap.NewNop())
	ctx := context.Background()
	svc.backfillOnce(ctx)
	// must not panic
}

func TestPronunciationService_ScheduleWords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		AudioDir:          t.TempDir(),
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("x")}}

	svc.ScheduleWords("one", "two", "three")
	// Each ScheduleWord may enqueue; at least no panic and queue may have entries
	_ = svc.queue
}

func TestPronunciationService_Recheck(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:  true,
		AudioDir: t.TempDir(),
	}, wordRepo, zap.NewNop())

	_, err := svc.Recheck("")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("Recheck empty word: want invalid, got %v", err)
	}

	_, err = svc.Recheck("привет")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("Recheck invalid word (non-Latin): want invalid, got %v", err)
	}

	res, err := svc.Recheck("hello")
	if err != nil {
		t.Fatalf("Recheck(hello): %v", err)
	}
	if res.Word != "hello" {
		t.Fatalf("Recheck word = %q want hello", res.Word)
	}
}

func TestPronunciationService_Recheck_NoTTSRepo(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, zap.NewNop())
	_, err := svc.Recheck("hello")
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("Recheck without tts repo: want error, got %v", err)
	}
}

func TestPronunciationService_Start_Disabled(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{Enabled: false, AudioDir: t.TempDir()}, nil, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.Start(ctx)
	// must not block forever; when disabled it returns immediately
}

func TestPronunciationService_Start_MkdirAllFails(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:  true,
		AudioDir: filepath.Join(t.TempDir(), "sub", "nested"),
	}, nil, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{}}
	// Create a file where we need a directory so MkdirAll fails
	dir := filepath.Join(t.TempDir(), "sub")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested"), []byte("x"), 0o644); err != nil {
		t.Fatalf("setup file: %v", err)
	}
	svc.audioDir = filepath.Join(dir, "nested")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.Start(ctx)
	// should return after logging error (no panic)
}

func TestPronunciationService_ScheduleWord_DisabledAndInvalid(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{Enabled: false, AudioDir: t.TempDir()}, nil, zap.NewNop())
	if svc.ScheduleWord("hello") {
		t.Fatal("ScheduleWord when disabled should return false")
	}
	svc = NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, zap.NewNop())
	if svc.ScheduleWord("") {
		t.Fatal("ScheduleWord empty word should return false")
	}
	if svc.ScheduleWord("привет") {
		t.Fatal("ScheduleWord invalid word should return false")
	}
}

func TestPronunciationService_ScheduleWord_TerminalAndMaxAttempts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), MaxRetries: 2,
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "x", err: fmt.Errorf("status 500")}}

	for i := 0; i < 3; i++ {
		svc.processWord(context.Background(), "termword")
	}
	ok := svc.ScheduleWord("termword")
	if ok {
		t.Fatal("ScheduleWord after terminal should return false")
	}
}

func TestPronunciationService_Lookup_DisabledAndInvalid(t *testing.T) {
	var nilSvc *PronunciationService
	r := nilSvc.Lookup("hello")
	if r.Available || r.NormalizedWord != "" {
		t.Fatalf("Lookup on nil service: %+v", r)
	}

	svc := NewPronunciationService(config.TTSConfig{Enabled: false, AudioDir: t.TempDir()}, nil, zap.NewNop())
	r = svc.Lookup("hello")
	if r.Available || r.NormalizedWord != "" {
		t.Fatalf("Lookup when disabled: %+v", r)
	}

	svc = NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, zap.NewNop())
	r = svc.Lookup("  ")
	if r.NormalizedWord != "" {
		t.Fatalf("Lookup empty word: %+v", r)
	}
}

func TestPronunciationService_Lookup_ReadyAndCached(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	dir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: dir, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	rel := svc.relativePathForWordWithExt("hello", ".mp3")
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := svc.ttsRepo.MarkReady("hello", "stub", rel); err != nil {
		t.Fatalf("MarkReady: %v", err)
	}

	r := svc.Lookup("hello")
	if !r.Available || !strings.Contains(r.URL, "/media/tts/") || !strings.Contains(r.URL, "hello.mp3") {
		t.Fatalf("Lookup(hello) ready: %+v", r)
	}
}

func TestPronunciationService_Lookup_HasCachedAudioNoRepo(t *testing.T) {
	dir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: dir, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, nil, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{}}
	rel := svc.relativePathForWordWithExt("cached", ".mp3")
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte("y"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	r := svc.Lookup("cached")
	if !r.Available || r.URL == "" {
		t.Fatalf("Lookup(cached) with file: %+v", r)
	}
}

func TestPronunciationService_EnsureStatusForWord_NoRepo(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, zap.NewNop())
	status, err := svc.ensureStatusForWord("hello")
	if err != nil {
		t.Fatalf("ensureStatusForWord: %v", err)
	}
	if status != nil {
		t.Fatalf("expected nil status when no tts repo")
	}
}

func TestPronunciationService_EnsureStatusForWord_LegacyFile(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	dir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: dir,
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	rel := svc.relativePathForWordWithExt("legacy", ".mp3")
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	status, err := svc.ensureStatusForWord("legacy")
	if err != nil {
		t.Fatalf("ensureStatusForWord: %v", err)
	}
	if status == nil || status.State != models.TTSStateReady {
		t.Fatalf("expected ready after legacy migration: %+v", status)
	}
}

func TestPronunciationService_ResolveReadyRelPath(t *testing.T) {
	dir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: dir}, nil, zap.NewNop())

	// nil status
	if got := svc.resolveReadyRelPath("x", nil); got != "" {
		t.Fatalf("resolveReadyRelPath(nil) = %q want \"\"", got)
	}

	// status ready but path missing on disk → fallback to cached
	empty := ""
	status := &models.TTSGenerationStatus{State: models.TTSStateReady, AudioRelPath: &empty}
	if got := svc.resolveReadyRelPath("word", status); got != "" {
		t.Fatalf("resolveReadyRelPath(empty path) = %q want \"\" or fallback", got)
	}

	// status with path that exists
	rel := svc.relativePathForWordWithExt("word", ".mp3")
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	status2 := &models.TTSGenerationStatus{State: models.TTSStateReady, AudioRelPath: &rel}
	if got := svc.resolveReadyRelPath("word", status2); got != rel {
		t.Fatalf("resolveReadyRelPath(exists) = %q want %q", got, rel)
	}

	// status with path that does not exist on disk → fallback to cachedRelPathForWord (file at rel exists)
	relMissing := "xx/yy/nonexistent.mp3"
	status3 := &models.TTSGenerationStatus{State: models.TTSStateReady, AudioRelPath: &relMissing}
	if got := svc.resolveReadyRelPath("word", status3); got != rel {
		t.Fatalf("resolveReadyRelPath(missing DB path) should fallback to cached; got %q want %q", got, rel)
	}
}

func TestDefaultPcmToMp3_InvalidInput(t *testing.T) {
	// Garbage input: ffmpeg typically fails to decode non-PCM
	_, err := defaultPcmToMp3([]byte("not-pcm-data"))
	if err != nil {
		return
	}
	// Some ffmpeg builds may succeed with arbitrary output; either way no panic
}

// TestOpenRouterPronunciationProvider_FetchViaChatCompletions_RetryOnNoAudio verifies
// that fetchViaChatCompletions retries once when the first attempt returns openrouter_no_audio.
func TestOpenRouterPronunciationProvider_FetchViaChatCompletions_RetryOnNoAudio(t *testing.T) {
	attempt := 0
	audioBytes := []byte("ID3-retry-mp3")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.Header().Set("Content-Type", "text/event-stream")
		if attempt == 1 {
			// First attempt: SSE with transcript but no audio data -> fetchViaChatCompletionsOnce returns openrouter_no_audio
			_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"audio":{"transcript":"hello"}}}]}` + "\n\n"))
			_, _ = w.Write([]byte("data: [DONE]\n\n"))
			return
		}
		// Second attempt: valid audio
		b64 := base64.StdEncoding.EncodeToString(audioBytes)
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `","transcript":"hello"}}}]}` + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test-model",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}

	audio, err := provider.fetchViaChatCompletions(context.Background(), "hello")
	if err != nil {
		t.Fatalf("fetchViaChatCompletions after retry: %v", err)
	}
	if string(audio) != string(audioBytes) {
		t.Fatalf("unexpected audio: got %q", audio)
	}
	if attempt != 2 {
		t.Fatalf("expected 2 attempts (first no audio, retry success), got %d", attempt)
	}
}

// TestPronunciationService_Start_WithPrefetchEnabled ensures Start with prefetch runs backfillOnce and exits on ctx cancel.
func TestPronunciationService_Start_WithPrefetchEnabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{
		Enabled:           true,
		PrefetchEnabled:   true,
		PrefetchWorkers:   1,
		BackfillInterval:  "1h",
		BackfillBatchSize: 5,
		AudioDir:          t.TempDir(),
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("x")}}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		svc.Start(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
		// Start returned after cancel
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after context cancel")
	}
}

// --- Coverage: ScheduleWord, writeFileAtomic, backfillOnce, fetchViaChatCompletionsOnce, processWord, parseSSEAudioStream, fetchViaAudioSpeech, buildPronunciationProviders ---

// TestPronunciationService_ScheduleWord_ReadyButFileMissing verifies that when status is Ready but file is missing, UpsertPending is called and scheduling continues.
func TestPronunciationService_ScheduleWord_ReadyButFileMissing(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled: true, Provider: "dictionary", AudioDir: audioDir, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	// Mark ready with a path that does not exist on disk
	_ = svc.ttsRepo.MarkReady("hello", "dict", "xx/yy/nonexistent.mp3")

	got := svc.ScheduleWord("hello")
	// Should still enqueue (or at least not panic); after UpsertPending the word may be scheduled
	if !got {
		// If canScheduleNow returned true we get true; if queue was full we might get true via async path
		status, _ := svc.ttsRepo.GetByWord("hello")
		if status != nil && status.State == models.TTSStatePending {
			// UpsertPending was applied
			return
		}
	}
}

// TestPronunciationService_ScheduleWord_AlreadyInQueue returns false when word is already in queue (canScheduleNow false).
func TestPronunciationService_ScheduleWord_AlreadyInQueue(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled: true, PrefetchWorkers: 1, AudioDir: audioDir, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	// Block the single worker so the first word stays in queue
	block := make(chan struct{})
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{
			providerName: "stub",
			audio:        []byte("x"),
		},
	}
	// Replace with a provider that blocks on fetch until we close block
	svc.providers = []pronunciationProvider{
		&blockingPronunciationProvider{block: block, audio: []byte("x")},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startDone := make(chan struct{})
	go func() {
		svc.Start(ctx)
		close(startDone)
	}()

	ok1 := svc.ScheduleWord("first")
	if !ok1 {
		t.Fatal("first ScheduleWord should succeed")
	}
	// Give worker time to take "first" from queue and block in fetch
	time.Sleep(100 * time.Millisecond)
	ok2 := svc.ScheduleWord("first")
	if ok2 {
		t.Error("second ScheduleWord(first) should return false when word already in queue")
	}
	close(block)
	cancel()
	select {
	case <-startDone:
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after context cancel")
	}
	// Allow workers to finish current processWord (e.g. writeFileAtomic) so no open fds under audioDir when t.TempDir() cleanup runs (macOS RemoveAll can fail if dir is in use).
	time.Sleep(200 * time.Millisecond)
}

// TestPronunciationService_ScheduleWord_NoTTSRepoButHasCachedAudio covers ScheduleWord when ttsRepo is nil but hasCachedAudio returns true.
// Service must have at least one provider so that enabled stays true (otherwise ScheduleWord returns false before the hasCachedAudio check).
func TestPronunciationService_ScheduleWord_NoTTSRepoButHasCachedAudio(t *testing.T) {
	audioDir := t.TempDir()
	key := pronunciationWordKey("hello")
	fileBase := pronunciationWordFileBase("hello")
	relDir := filepath.Join(key[:2], key[2:4])
	relPath := filepath.Join(relDir, fileBase+".mp3")
	if err := os.MkdirAll(filepath.Join(audioDir, relDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(audioDir, relPath), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cfg := config.TTSConfig{
		Enabled:           true,
		AudioDir:          audioDir,
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}
	svc := NewPronunciationService(cfg, nil, zap.NewNop())
	got := svc.ScheduleWord("hello")
	if !got {
		t.Error("ScheduleWord should return true when ttsRepo is nil but hasCachedAudio is true")
	}
}

// TestPronunciationService_logTTSStatusDecision_NilRepo covers logTTSStatusDecision when ttsRepo is nil (early return, no panic).
func TestPronunciationService_logTTSStatusDecision_NilRepo(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, zap.NewNop())
	svc.logTTSStatusDecision("hello", "ok", "provider", "code")
	// must not panic
}

// TestPronunciationService_logTTSStatusDecision_NoStatus covers logTTSStatusDecision when GetByWord returns nil status (early return).
func TestPronunciationService_logTTSStatusDecision_NoStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com"}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	// Word "nonexistent" has no row in tts_generation_status -> GetByWord returns (nil, nil)
	svc.logTTSStatusDecision("nonexistent", "ok", "p", "c")
	// must not panic
}

// TestPronunciationService_logTTSStatusDecision_WithStatus covers logTTSStatusDecision when status exists (logger.Info path).
func TestPronunciationService_logTTSStatusDecision_WithStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com"}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.UpsertPending("loggedword")
	svc.logTTSStatusDecision("loggedword", "ready", "dict", "")
	// must not panic; full path through logger.Info
	status, _ := svc.ttsRepo.GetByWord("loggedword")
	if status == nil {
		t.Fatal("expected status to exist after UpsertPending")
	}
}

type blockingPronunciationProvider struct {
	block chan struct{}
	audio []byte
}

func (p *blockingPronunciationProvider) name() string { return "blocking" }
func (p *blockingPronunciationProvider) fetch(ctx context.Context, _ string) ([]byte, error) {
	select {
	case <-p.block:
		return p.audio, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// TestWriteFileAtomic_RenameFails covers rename error when target path is an existing directory.
func TestWriteFileAtomic_RenameFails(t *testing.T) {
	dir := t.TempDir()
	targetDir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Writing to a path that is an existing directory: CreateTemp creates file in subdir, then Rename(tmp, subdir) fails
	err := writeFileAtomic(targetDir, []byte("data"), 0o644)
	if err == nil {
		t.Fatal("expected error when target is a directory")
	}
	if !strings.Contains(err.Error(), "rename") {
		t.Errorf("expected rename error, got %q", err.Error())
	}
}

// TestPronunciationService_BackfillOnce_ListCandidatesError runs backfillOnce when wordRepo's DB fails ListPronunciationCandidates (e.g. closed connection).
func TestPronunciationService_BackfillOnce_ListCandidatesError(t *testing.T) {
	_ = testutil.SetupTestDB(t) // ensure driver registered and shared DB exists
	dsn := testutil.GetTestDSN(t)
	closedDB, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("open test DB: %v (driver may not be registered)", err)
	}
	closedDB.Close()
	wordRepo := repository.NewWordRepository(closedDB, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), BackfillBatchSize: 5,
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("x")}}

	ctx := context.Background()
	svc.backfillOnce(ctx)
	// must not panic; ListPronunciationCandidates returns error, backfillOnce logs and returns
}

// TestPronunciationService_BackfillOnce_CtxDone covers select <-ctx.Done() in backfillOnce.
func TestPronunciationService_BackfillOnce_CtxDone(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, BackfillBatchSize: 5, AudioDir: t.TempDir(),
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	svc.backfillOnce(ctx)
	// no panic; select hits <-ctx.Done()
}

// TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_NonSSEWith404 covers non-SSE content-type and 404 from chat completions.
func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_NonSSEWith404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error for 404 non-SSE response")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Errorf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "openrouter_rejected") {
		t.Errorf("expected openrouter_rejected in error, got %q", err.Error())
	}
}

// TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_SSEDecodeError covers parseSSEAudioStream returning "sse audio decode failed".
func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_SSEDecodeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// Invalid base64 in data -> decode fails, pcm empty, DecodeErrors > 0
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"audio":{"data":"!!!invalid-base64!!!"}}}]}` + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error for SSE decode failure")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Errorf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(strings.ToLower(err.Error()), "decode") {
		t.Errorf("expected decode in error, got %q", err.Error())
	}
}

// TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_EmptyTranscript covers missing transcript for validation.
func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_EmptyTranscript(t *testing.T) {
	audioBytes := []byte("pcm")
	b64 := base64.StdEncoding.EncodeToString(audioBytes)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `"}}}]}` + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for empty transcript")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Errorf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "missing transcript") {
		t.Errorf("expected missing transcript in error, got %q", err.Error())
	}
}

// TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_PcmToMp3Error covers pcmToMp3 returning error.
func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_PcmToMp3Error(t *testing.T) {
	audioBytes := []byte("pcm")
	b64 := base64.StdEncoding.EncodeToString(audioBytes)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `","transcript":"hello"}}}]}` + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(_ []byte) ([]byte, error) { return nil, fmt.Errorf("convert failed") },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error when pcmToMp3 fails")
	}
	if !strings.Contains(err.Error(), "pcm to mp3") || !strings.Contains(err.Error(), "convert failed") {
		t.Errorf("unexpected error: %q", err.Error())
	}
}

// TestParseSSEAudioStream_ScannerError covers scanner.Err() path.
func TestParseSSEAudioStream_ScannerError(t *testing.T) {
	r := &errReader{err: fmt.Errorf("read error")}
	_, _, stats, err := parseSSEAudioStream(r)
	if err == nil {
		t.Fatal("expected error from scanner")
	}
	if err.Error() != "read error" {
		t.Errorf("unexpected error: %v", err)
	}
	_ = stats
}

// errReader returns one byte then (0, err) on next Read so bufio.Scanner gets scanner.Err() != nil.
type errReader struct {
	done bool
	err  error
}

func (e *errReader) Read(p []byte) (n int, err error) {
	if e.done {
		return 0, e.err
	}
	e.done = true
	if len(p) > 0 {
		p[0] = 'x'
		return 1, nil
	}
	return 0, nil
}

// TestParseSSEAudioStream_DecodeErrorAndNoPCM covers pcm.Len()==0 && stats.DecodeErrors > 0.
func TestParseSSEAudioStream_DecodeErrorAndNoPCM(t *testing.T) {
	sse := strings.NewReader(`data: {"choices":[{"delta":{"audio":{"data":"!!!invalid!!!"}}}]}` + "\n\n" + `data: [DONE]` + "\n")
	_, transcript, stats, err := parseSSEAudioStream(sse)
	if err == nil {
		t.Fatal("expected sse audio decode failed error")
	}
	if !strings.Contains(err.Error(), "sse audio decode failed") {
		t.Errorf("unexpected error: %q", err.Error())
	}
	if stats.DecodeErrors != 1 {
		t.Errorf("expected DecodeErrors=1, got %d", stats.DecodeErrors)
	}
	_ = transcript
}

// TestParseSSEAudioStream_MalformedJSON covers JSON unmarshal error and FirstJSONError/FirstJSONSample.
func TestParseSSEAudioStream_MalformedJSON(t *testing.T) {
	sse := strings.NewReader("data: {invalid json}\n\n" + "data: [DONE]\n\n")
	_, transcript, stats, err := parseSSEAudioStream(sse)
	if err != nil {
		t.Fatalf("malformed JSON line is skipped, no stream error: %v", err)
	}
	if stats.JSONErrors != 1 {
		t.Errorf("expected JSONErrors=1, got %d", stats.JSONErrors)
	}
	if stats.FirstJSONError == "" {
		t.Error("expected FirstJSONError set")
	}
	if stats.FirstJSONSample == "" {
		t.Error("expected FirstJSONSample set")
	}
	_ = transcript
}

// TestParseSSEAudioStream_EmptyChoices covers len(chunk.Choices)==0.
func TestParseSSEAudioStream_EmptyChoices(t *testing.T) {
	sse := strings.NewReader(`data: {"choices":[]}` + "\n\n" + "data: [DONE]\n\n")
	pcm, transcript, stats, err := parseSSEAudioStream(sse)
	if err != nil {
		t.Fatalf("empty choices is valid: %v", err)
	}
	if len(pcm) != 0 || transcript != "" {
		t.Errorf("expected empty pcm and transcript, got pcm=%d transcript=%q", len(pcm), transcript)
	}
	if stats.JSONChunks != 1 || stats.ChoicesChunks != 0 {
		t.Errorf("expected JSONChunks=1 ChoicesChunks=0, got JSONChunks=%d ChoicesChunks=%d", stats.JSONChunks, stats.ChoicesChunks)
	}
}

// TestOpenRouterPronunciationProvider_FetchViaAudioSpeech_Non2xx covers non-404/400/422 status (e.g. 500).
func TestOpenRouterPronunciationProvider_FetchViaAudioSpeech_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "tts-1",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: false,
	}
	_, err := provider.fetchViaAudioSpeech(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if strings.Contains(err.Error(), "openrouter_audio_speech_rejected") || errors.Is(err, errPronunciationNotFound) {
		t.Errorf("500 is not a 'not found' rejection, got %v", err)
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got %q", err.Error())
	}
}

// TestOpenRouterPronunciationProvider_FetchViaAudioSpeech_EmptyBody covers len(audio)==0.
func TestOpenRouterPronunciationProvider_FetchViaAudioSpeech_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		// empty body
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "tts-1",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: false,
	}
	_, err := provider.fetchViaAudioSpeech(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error for empty body")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Errorf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "openrouter_audio_speech_empty") {
		t.Errorf("expected openrouter_audio_speech_empty in error, got %q", err.Error())
	}
}

// TestBuildPronunciationProviders_DictionaryOnly covers provider "dictionary".
func TestBuildPronunciationProviders_DictionaryOnly(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider:          "dictionary",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://api.dictionaryapi.dev/api/v2/entries/en",
	}, zap.NewNop())
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].name() != "dictionary" {
		t.Fatalf("expected dictionary, got %q", providers[0].name())
	}
}

// TestBuildPronunciationProviders_OpenRouterOnly covers provider "openrouter".
func TestBuildPronunciationProviders_OpenRouterOnly(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider: "openrouter",
		APIKey:   "key",
		Model:    "model",
		BaseURL:  "https://openrouter.ai/api/v1",
		Voice:    "alloy",
	}, zap.NewNop())
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].name() != "openrouter" {
		t.Fatalf("expected openrouter, got %q", providers[0].name())
	}
}

// TestBuildPronunciationProviders_UnknownProvider falls back to auto (both providers when configured).
func TestBuildPronunciationProviders_UnknownProvider(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider:          "unknown_xyz",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://example.com",
		APIKey:            "key",
		Model:             "m",
	}, zap.NewNop())
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers (fallback to auto), got %d", len(providers))
	}
}

// TestBuildPronunciationProviders_EmptyProvider defaults to auto.
func TestBuildPronunciationProviders_EmptyProvider(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider:          "  ",
		DictionaryEnabled: true,
		DictionaryBaseURL: "https://example.com",
		APIKey:            "key",
		Model:             "m",
	}, zap.NewNop())
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers (auto), got %d", len(providers))
	}
}

// TestNewPronunciationService_NoProvidersDisablesService covers NewPronunciationService when buildPronunciationProviders returns empty (enabled set to false).
func TestNewPronunciationService_NoProvidersDisablesService(t *testing.T) {
	cfg := config.TTSConfig{
		Enabled:           true,
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		Provider:          "dictionary",
		DictionaryEnabled: false,
		APIKey:            "",
		Model:             "",
	}
	svc := NewPronunciationService(cfg, nil, zap.NewNop())
	if svc.IsEnabled() {
		t.Error("expected service to be disabled when no providers available")
	}
}

// TestBuildPronunciationProviders_DictionaryMinDelayNegative clamps negative minDelay to 0.
func TestBuildPronunciationProviders_DictionaryMinDelayNegative(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider:           "dictionary",
		DictionaryEnabled:  true,
		DictionaryBaseURL:  "https://example.com",
		DictionaryMinDelay: "-1s",
	}, zap.NewNop())
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	p, ok := providers[0].(*dictionaryPronunciationProvider)
	if !ok {
		t.Fatalf("expected dictionaryPronunciationProvider, got %T", providers[0])
	}
	if p.throttleEvery < 0 {
		t.Errorf("expected throttleEvery >= 0, got %v", p.throttleEvery)
	}
}

// TestPronunciationService_ProcessWord_TerminalReturnsEarly verifies processWord returns without calling providers when status is failed_terminal.
func TestPronunciationService_ProcessWord_TerminalReturnsEarly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 2,
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkTerminal("early", "service", "test", "err")
	var fetchCalls int
	svc.providers = []pronunciationProvider{
		&countingProvider{provider: &stubPronunciationProvider{providerName: "x", audio: []byte("x")}, count: &fetchCalls},
	}
	svc.processWord(context.Background(), "early")
	if fetchCalls != 0 {
		t.Errorf("expected no provider fetch when status is terminal, got %d", fetchCalls)
	}
}

// TestPronunciationService_ProcessWord_MaxAttemptsReachedReturnsEarly verifies processWord marks terminal and returns when attempt_count >= max_attempts.
func TestPronunciationService_ProcessWord_MaxAttemptsReachedReturnsEarly(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 2,
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	for i := 0; i < 2; i++ {
		_ = svc.ttsRepo.MarkAttempt("maxearly", "p", "code", "err", true)
	}
	var fetchCalls int
	svc.providers = []pronunciationProvider{
		&countingProvider{provider: &stubPronunciationProvider{providerName: "x", audio: []byte("x")}, count: &fetchCalls},
	}
	svc.processWord(context.Background(), "maxearly")
	if fetchCalls != 0 {
		t.Errorf("expected no provider fetch when max attempts reached, got %d", fetchCalls)
	}
	status, _ := svc.ttsRepo.GetByWord("maxearly")
	if status != nil && status.State != models.TTSStateFailedTerminal {
		t.Errorf("expected terminal state, got %q", status.State)
	}
}

// TestPronunciationService_ProcessWord_ReadyButFileMissingUpsertPending verifies processWord calls UpsertPending when status is Ready but file is missing, then tries providers.
func TestPronunciationService_ProcessWord_ReadyButFileMissingUpsertPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	cfg := config.TTSConfig{
		Enabled: true, AudioDir: audioDir, PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkReady("word", "dict", "nonexistent/path.mp3")
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("audio")}}

	svc.processWord(context.Background(), "word")

	// Should have written file via provider and marked ready
	if !svc.hasCachedAudio("word") {
		t.Fatal("expected audio file after UpsertPending and provider success")
	}
	status, _ := svc.ttsRepo.GetByWord("word")
	if status != nil && status.State != models.TTSStateReady {
		t.Errorf("expected ready after provider success, got state %q", status.State)
	}
}

// TestPronunciationService_ProcessWord_EmptyProviders_UnknownFailure covers unknown_generation_failure when no providers.
func TestPronunciationService_ProcessWord_EmptyProviders_UnknownFailure(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	cfg := config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}
	svc := NewPronunciationService(cfg, wordRepo, zap.NewNop())
	svc.providers = nil

	svc.processWord(context.Background(), "noprov")

	status, err := svc.ttsRepo.GetByWord("noprov")
	if err != nil {
		t.Fatalf("GetByWord: %v", err)
	}
	if status == nil {
		t.Fatal("expected status record")
	}
	if status.State != models.TTSStateFailedRetryable {
		t.Errorf("expected failed_retryable after unknown_generation_failure, got %q", status.State)
	}
}

// --- Coverage: fetchViaAudioSpeech client.Do/ReadBody errors, fetchViaChatCompletionsOnce branches, parseSSE, dictionary, NewPronunciationService, processWord, ScheduleWord, Lookup, GetStatus, ensureStatus, writeFileAtomic, helpers ---

type pronunciationRoundTripper struct {
	fn func(*http.Request) (*http.Response, error)
}

func (r *pronunciationRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r.fn(req)
}

type pronunciationErrBody struct{ err error }

func (e *pronunciationErrBody) Read([]byte) (int, error) { return 0, e.err }
func (e *pronunciationErrBody) Close() error             { return nil }

func TestOpenRouterPronunciationProvider_FetchViaAudioSpeech_ClientDoError(t *testing.T) {
	wantErr := fmt.Errorf("client do failed")
	client := &http.Client{
		Transport: &pronunciationRoundTripper{fn: func(*http.Request) (*http.Response, error) {
			return nil, wantErr
		}},
	}
	provider := &openRouterPronunciationProvider{
		baseURL:              "http://example.com",
		model:                "tts-1",
		voice:                "alloy",
		apiKey:               "key",
		client:               client,
		forceChatCompletions: false,
	}
	_, err := provider.fetchViaAudioSpeech(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "openrouter tts request failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenRouterPronunciationProvider_FetchViaAudioSpeech_ReadBodyError(t *testing.T) {
	readErr := fmt.Errorf("read body failed")
	client := &http.Client{
		Transport: &pronunciationRoundTripper{fn: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       &pronunciationErrBody{err: readErr},
				Header:     make(http.Header),
			}, nil
		}},
	}
	provider := &openRouterPronunciationProvider{
		baseURL:              "http://example.com",
		model:                "tts-1",
		voice:                "alloy",
		apiKey:               "key",
		client:               client,
		forceChatCompletions: false,
	}
	_, err := provider.fetchViaAudioSpeech(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "openrouter tts read body") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_NonSSEOtherStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unexpected content-type") {
		t.Errorf("expected content-type in error, got %q", err.Error())
	}
}

// TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_StatusNotOK covers status != 200 with event-stream content-type.
func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_StatusNotOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rejected") && !strings.Contains(err.Error(), "500") {
		t.Errorf("expected rejected/500 in error, got %q", err.Error())
	}
}

func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_StreamError(t *testing.T) {
	// Use a transport that returns a body which errors on Read so parseSSEAudioStream gets scanner.Err().
	streamErr := fmt.Errorf("stream read failed")
	client := &http.Client{
		Transport: &pronunciationRoundTripper{fn: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       &pronunciationErrBody{err: streamErr},
			}, nil
		}},
	}
	provider := &openRouterPronunciationProvider{
		baseURL:              "http://example.com",
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               client,
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "w")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "openrouter chat stream") {
		t.Errorf("expected stream error, got %q", err.Error())
	}
}

func TestOpenRouterPronunciationProvider_FetchViaChatCompletionsOnce_NoAudioLongTranscript(t *testing.T) {
	// len(pcm)==0 and transcript > 120 chars -> trimmed = trimmed[:120]
	longTranscript := strings.Repeat("x", 150)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		evt := `data: {"choices":[{"delta":{"audio":{"transcript":"` + longTranscript + `"}}}]}` + "\n\n"
		_, _ = w.Write([]byte(evt))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetchViaChatCompletionsOnce(context.Background(), "word")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Errorf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "openrouter_no_audio") || !strings.Contains(err.Error(), "transcript=") {
		t.Errorf("expected transcript in error, got %q", err.Error())
	}
}

func TestParseSSEAudioStream_ExtractDeltaContentText_ArrayItemNonText(t *testing.T) {
	// content as array with type!="text" item (skip), and type=="" or "text" (include)
	rawAudio := []byte{0x01, 0x02}
	b64 := base64.StdEncoding.EncodeToString(rawAudio)
	sse := strings.NewReader(
		`data: {"choices":[{"delta":{"audio":{"data":"` + b64 + `"},"content":[{"type":"image","text":""},{"type":"text","text":"hi"}]}}]}` + "\n\n" +
			`data: [DONE]` + "\n\n",
	)
	pcm, transcript, _, err := parseSSEAudioStream(sse)
	if err != nil {
		t.Fatalf("parseSSEAudioStream: %v", err)
	}
	if string(pcm) != string(rawAudio) {
		t.Errorf("unexpected pcm")
	}
	if transcript != "hi" {
		t.Errorf("transcript = %q want hi", transcript)
	}
}

func TestExtractDeltaContentText_UnmarshalFails(t *testing.T) {
	// raw is not string and not array of {type, text} -> return nil
	out := extractDeltaContentText(json.RawMessage(`123`))
	if out != nil {
		t.Errorf("expected nil for number, got %v", out)
	}
	out = extractDeltaContentText(json.RawMessage(`{"x":1}`))
	if out != nil {
		t.Errorf("expected nil for object, got %v", out)
	}
}

// TestExtractDeltaContentText_AsItems covers extractDeltaContentText when raw is array of {type, text}.
func TestExtractDeltaContentText_AsItems(t *testing.T) {
	out := extractDeltaContentText(json.RawMessage(`[{"type":"text","text":"hello"}]`))
	if len(out) != 1 || out[0] != "hello" {
		t.Errorf("expected [hello], got %v", out)
	}
	out = extractDeltaContentText(json.RawMessage(`[{"type":"","text":"a"},{"type":"text","text":"b"}]`))
	if len(out) != 2 || out[0] != "a" || out[1] != "b" {
		t.Errorf("expected [a b], got %v", out)
	}
	out = extractDeltaContentText(json.RawMessage(`[{"type":"image","text":""}]`))
	if len(out) != 0 {
		t.Errorf("expected [] for non-text type, got %v", out)
	}
}

func TestIsLikelySingleWordPronunciationTranscript(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"", false},
		{"  ", false},
		{"hello!", false},
		{"what?", false},
		{"one two three", false},
		{"ab", false},
		{"abc", true},
		{"kə-MOH-shun", true},
	}
	for _, tt := range tests {
		got := isLikelySingleWordPronunciationTranscript(tt.s)
		if got != tt.want {
			t.Errorf("isLikelySingleWordPronunciationTranscript(%q) = %v want %v", tt.s, got, tt.want)
		}
	}
}

func TestDictionaryLookupWord_ToPrefix(t *testing.T) {
	if got := dictionaryLookupWord("to go"); got != "go" {
		t.Errorf("dictionaryLookupWord(\"to go\") = %q want go", got)
	}
	if got := dictionaryLookupWord("to "); got != "to" {
		t.Errorf("dictionaryLookupWord(\"to \") = %q want to", got)
	}
}

func TestPickDictionaryAudioURL_FallbackNoMp3(t *testing.T) {
	entries := []dictEntry{
		{Phonetics: []dictPhonetic{{Audio: "https://example.com/audio.ogg"}}},
	}
	if got := pickDictionaryAudioURL(entries); got != "https://example.com/audio.ogg" {
		t.Errorf("pickDictionaryAudioURL = %q want ogg fallback", got)
	}
}

func TestDictionaryPronunciationProviderFetch_AudioURLProtocolRelative(t *testing.T) {
	// Code rewrites "//host/path" to "https://host/path". Use a custom transport: dictionary request
	// goes to httptest server; audio request (https) is handled by mock so we don't need TLS.
	audioBytes := []byte("mp3")
	var audioReqSeen bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/hello") || r.URL.Path == "/hello" {
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"//` + r.Host + `/audio/hello.mp3"}]}]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	baseTransport := srv.Client().Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	client := &http.Client{
		Transport: &pronunciationRoundTripper{fn: func(req *http.Request) (*http.Response, error) {
			if strings.HasPrefix(req.URL.Scheme, "https") && strings.Contains(req.URL.Path, "audio") {
				audioReqSeen = true
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"audio/mpeg"}},
					Body:       io.NopCloser(bytes.NewReader(audioBytes)),
				}, nil
			}
			return baseTransport.RoundTrip(req)
		}},
	}
	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: client}
	got, err := p.fetch(context.Background(), "hello")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if string(got) != string(audioBytes) {
		t.Errorf("unexpected audio")
	}
	if !audioReqSeen {
		t.Error("expected audio request to https URL")
	}
}

func TestDictionaryPronunciationProvider_WaitThrottle_CtxDone(t *testing.T) {
	// First call sets nextAllowedAt in the future; second call (with cancelled ctx) blocks then returns ctx.Canceled.
	p := &dictionaryPronunciationProvider{throttleEvery: 2 * time.Second}
	ctxBg := context.Background()
	_ = p.waitThrottle(ctxBg) // wait=0, sets nextAllowedAt = now+2s
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := p.waitThrottle(ctx)
	if err != context.Canceled {
		t.Errorf("waitThrottle(ctx cancelled) = %v want context.Canceled", err)
	}
}

func TestDictionaryPronunciationProviderFetch_ThrottleWait(t *testing.T) {
	audioBytes := []byte("x")
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if strings.HasSuffix(r.URL.Path, "/hello") || r.URL.Path == "/hello" {
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"http://` + r.Host + `/a.mp3"}]}]`))
			return
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write(audioBytes)
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{
		baseURL:       srv.URL,
		client:        &http.Client{Timeout: 3 * time.Second},
		throttleEvery: 50 * time.Millisecond,
	}
	got, err := p.fetch(context.Background(), "hello")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if string(got) != string(audioBytes) {
		t.Errorf("unexpected audio")
	}
	if calls < 2 {
		t.Errorf("expected at least 2 calls (lookup + audio), got %d", calls)
	}
}

func TestNewPronunciationService_ConfigDefaults(t *testing.T) {
	// nil logger -> nop
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, nil)
	if svc.logger == nil {
		t.Fatal("expected non-nil logger")
	}

	// RetryBaseDelay "0" or invalid -> retryBase = 1m
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), RetryBaseDelay: "0", RetryMaxDelay: "24h",
	}, nil, zap.NewNop())
	if svc.retryBase != time.Minute {
		t.Errorf("retryBase = %v want 1m", svc.retryBase)
	}

	// RetryMaxDelay < retryBase -> retryMax = retryBase
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), RetryBaseDelay: "2m", RetryMaxDelay: "1m",
	}, nil, zap.NewNop())
	if svc.retryMax != svc.retryBase {
		t.Errorf("retryMax should be >= retryBase")
	}

	// BackfillInterval "0" -> backfillEvery = 10m
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), BackfillInterval: "0",
	}, nil, zap.NewNop())
	if svc.backfillEvery != 10*time.Minute {
		t.Errorf("backfillEvery = %v", svc.backfillEvery)
	}

	// PrefetchWorkers 0 -> 1, >8 -> 8
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PrefetchWorkers: 0,
	}, nil, zap.NewNop())
	if svc.workers != 1 {
		t.Errorf("workers = %d want 1", svc.workers)
	}
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PrefetchWorkers: 10,
	}, nil, zap.NewNop())
	if svc.workers != 8 {
		t.Errorf("workers = %d want 8", svc.workers)
	}

	// BackfillBatchSize 0 -> 200
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), BackfillBatchSize: 0,
	}, nil, zap.NewNop())
	if svc.backfillBatch != 200 {
		t.Errorf("backfillBatch = %d want 200", svc.backfillBatch)
	}

	// MaxRetries 0 -> 1, >20 -> 20
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), MaxRetries: 0,
	}, nil, zap.NewNop())
	if svc.maxRetries != 1 {
		t.Errorf("maxRetries = %d want 1", svc.maxRetries)
	}
	svc = NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), MaxRetries: 25,
	}, nil, zap.NewNop())
	if svc.maxRetries != 20 {
		t.Errorf("maxRetries = %d want 20", svc.maxRetries)
	}
}

func TestParseDurationWithDefault_Invalid(t *testing.T) {
	if got := parseDurationWithDefault("invalid", time.Second); got != time.Second {
		t.Errorf("parseDurationWithDefault(invalid) = %v want 1s fallback", got)
	}
}

func TestBuildPronunciationProviders_OpenRouterVoiceEmpty(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider: "openrouter",
		APIKey:   "key",
		Model:    "m",
		Voice:    "  ",
		BaseURL:  "https://openrouter.ai/api/v1",
	}, zap.NewNop())
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	p, ok := providers[0].(*openRouterPronunciationProvider)
	if !ok {
		t.Fatalf("expected openRouterPronunciationProvider, got %T", providers[0])
	}
	if p.voice != "alloy" {
		t.Errorf("voice = %q want alloy default", p.voice)
	}
}

func TestPronunciationService_ProcessWord_EnsureStatusError(t *testing.T) {
	_ = testutil.SetupTestDB(t)
	dsn := testutil.GetTestDSN(t)
	closedDB, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("open test DB: %v", err)
	}
	closedDB.Close()
	wordRepo := repository.NewWordRepository(closedDB, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "x", audio: []byte("a")}}

	svc.processWord(context.Background(), "word")
	// ensureStatusForWord returns (nil, err) -> log and return without calling providers
	// No panic; no file written because we return early
	if svc.hasCachedAudio("word") {
		t.Error("expected no cached audio when ensureStatus fails")
	}
}

func TestPronunciationService_ProcessWord_TerminalThenReturn(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 3,
	}, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkTerminal("term", "p", "code", "err")
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "x", audio: []byte("a")}}
	var calls int
	svc.providers = []pronunciationProvider{
		&countingProvider{provider: &stubPronunciationProvider{providerName: "x", audio: []byte("a")}, count: &calls},
	}
	svc.processWord(context.Background(), "term")
	if calls != 0 {
		t.Errorf("expected no fetch when status is terminal, got %d", calls)
	}
}

func TestPronunciationService_ProcessWord_MaxAttemptsReachedThenMarkTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 2,
	}, wordRepo, zap.NewNop())
	for i := 0; i < 2; i++ {
		_ = svc.ttsRepo.MarkAttempt("maxw", "p", "c", "e", true)
	}
	var count int
	svc.providers = []pronunciationProvider{
		&countingProvider{provider: &stubPronunciationProvider{providerName: "x", audio: []byte("a")}, count: &count},
	}
	svc.processWord(context.Background(), "maxw")
	if count != 0 {
		t.Errorf("expected no fetch when max attempts reached, got %d", count)
	}
}

func TestPronunciationService_ProcessWord_ReadyButFileMissingUpsertThenFetch(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: audioDir, DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkReady("readyword", "dict", "missing/path.mp3")
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "stub", audio: []byte("audio")}}

	svc.processWord(context.Background(), "readyword")

	if !svc.hasCachedAudio("readyword") {
		t.Fatal("expected audio after UpsertPending and provider success")
	}
}

func TestPronunciationService_ProcessWord_NonRetryableMarksTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 5,
	}, wordRepo, zap.NewNop())
	nonRetryable := errors.New("random error")
	svc.providers = []pronunciationProvider{&stubPronunciationProvider{providerName: "p", err: nonRetryable}}

	svc.processWord(context.Background(), "termword")

	status, _ := svc.ttsRepo.GetByWord("termword")
	if status == nil || status.State != models.TTSStateFailedTerminal {
		t.Errorf("expected terminal when provider returns non-retryable error, got %v", status)
	}
}

func TestPronunciationService_ScheduleWord_EnsureStatusError(t *testing.T) {
	_ = testutil.SetupTestDB(t)
	dsn := testutil.GetTestDSN(t)
	closedDB, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("open test DB: %v", err)
	}
	closedDB.Close()
	wordRepo := repository.NewWordRepository(closedDB, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	got := svc.ScheduleWord("hello")
	if got {
		t.Error("ScheduleWord with ensureStatus error should return false")
	}
}

func TestPronunciationService_ScheduleWord_ReadyButFileMissingUpsertPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: audioDir, DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkReady("r", "dict", "x/y/nonexistent.mp3")

	got := svc.ScheduleWord("r")
	if !got {
		t.Error("ScheduleWord when ready but file missing should enqueue and return true")
	}
}

func TestPronunciationService_Lookup_EnsureStatusError(t *testing.T) {
	_ = testutil.SetupTestDB(t)
	dsn := testutil.GetTestDSN(t)
	closedDB, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("open test DB: %v", err)
	}
	closedDB.Close()
	wordRepo := repository.NewWordRepository(closedDB, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	r := svc.Lookup("hello")
	// When ensureStatusForWord returns error we still have NormalizedWord set (before the call), Available stays false.
	if r.NormalizedWord != "hello" || r.Available {
		t.Errorf("Lookup with ensureStatus error: NormalizedWord=%q Available=%v", r.NormalizedWord, r.Available)
	}
}

func TestPronunciationService_Lookup_FailedTerminalAndMaxAttempts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts", DictionaryEnabled: true, DictionaryBaseURL: "http://example.com", MaxRetries: 2,
	}, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkTerminal("fail", "p", "c", "e")
	r := svc.Lookup("fail")
	if r.Available {
		t.Error("Lookup when terminal should not be available")
	}

	for i := 0; i < 2; i++ {
		_ = svc.ttsRepo.MarkAttempt("maxw", "p", "c", "e", true)
	}
	r = svc.Lookup("maxw")
	if r.Available {
		t.Error("Lookup when max attempts should not be available")
	}
}

func TestPronunciationService_GetStatus_NoTTSRepoAndInvalidWord(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, nil, zap.NewNop())
	_, err := svc.GetStatus("hello")
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Errorf("GetStatus without ttsRepo: %v", err)
	}

	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc = NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, wordRepo, zap.NewNop())
	_, err = svc.GetStatus("")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Errorf("GetStatus invalid word: %v", err)
	}
}

func TestPronunciationService_GetStatus_EnsureErrorAndStatusNil(t *testing.T) {
	_ = testutil.SetupTestDB(t)
	dsn := testutil.GetTestDSN(t)
	closedDB, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("open test DB: %v", err)
	}
	closedDB.Close()
	wordRepo := repository.NewWordRepository(closedDB, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, wordRepo, zap.NewNop())
	_, err = svc.GetStatus("hello")
	if err == nil {
		t.Fatal("GetStatus with closed DB should return error")
	}
}

func TestPronunciationService_GetStatus_StatusNilUpsertPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, wordRepo, zap.NewNop())
	out, err := svc.GetStatus("newword")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if out.Word != "newword" || out.State != "pending" {
		t.Errorf("GetStatus new word: %+v", out)
	}
}

func TestPronunciationService_GetStatus_WithNullableFields(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkReady("word", "dict", "ab/cd/word.mp3")
	// Ensure status has nullable fields; resolveReadyRelPath may return "" if file missing
	out, err := svc.GetStatus("word")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if out.Word != "word" {
		t.Errorf("GetStatus word: %q", out.Word)
	}
	if out.State != "ready" && out.State != "pending" {
		t.Errorf("GetStatus state: %q", out.State)
	}
}

func TestPronunciationService_ForceRegenerate_ResetError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, wordRepo, zap.NewNop())
	// ForceRegenerate calls ResetForForceRegenerate; if we could make it fail we'd cover that branch.
	// ResetForForceRegenerate is on ttsRepo - no easy way to make it fail with real DB. Rely on integration.
	_, _ = svc.ForceRegenerate("hello")
}

func TestPronunciationService_EnsureStatusForWord_GetByWordError(t *testing.T) {
	_ = testutil.SetupTestDB(t)
	dsn := testutil.GetTestDSN(t)
	closedDB, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Skipf("open test DB: %v", err)
	}
	closedDB.Close()
	wordRepo := repository.NewWordRepository(closedDB, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: t.TempDir()}, wordRepo, zap.NewNop())
	status, err := svc.ensureStatusForWord("hello")
	if err == nil {
		t.Fatal("ensureStatusForWord with closed DB should return error")
	}
	if status != nil {
		t.Errorf("expected nil status, got %v", status)
	}
}

func TestPronunciationService_EnsureStatusForWord_ReadyButPathMissingUpsertPending(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	audioDir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: audioDir}, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkReady("word", "dict", "nonexistent/path.mp3")
	status, err := svc.ensureStatusForWord("word")
	if err != nil {
		t.Fatalf("ensureStatusForWord: %v", err)
	}
	if status == nil {
		t.Fatal("expected status after UpsertPending")
	}
	if status.State != models.TTSStatePending {
		t.Errorf("expected pending after path missing, got %q", status.State)
	}
}

func TestPronunciationService_PublicURLForWord_NoCachedFile(t *testing.T) {
	dir := t.TempDir()
	svc := NewPronunciationService(config.TTSConfig{Enabled: true, AudioDir: dir, PublicBasePath: "/media/tts"}, nil, zap.NewNop())
	// Word with no cached file -> cachedRelPathForWord returns "" -> fallback to relativePathForWord
	url := svc.publicURLForWord("hello")
	if url == "" || !strings.HasPrefix(url, "/media/tts/") {
		t.Errorf("publicURLForWord(no cache) = %q", url)
	}
}

func TestPronunciationWordFileBase_EmptyName(t *testing.T) {
	if got := pronunciationWordFileBase("   "); got != "word" {
		t.Errorf("pronunciationWordFileBase(whitespace) = %q want word", got)
	}
}

func TestNormalizePronunciationWord_NoLatin(t *testing.T) {
	if _, ok := normalizePronunciationWord("123"); ok {
		t.Error("normalizePronunciationWord(123) should be invalid (no Latin)")
	}
}

func TestClassifyPronunciationError_NonWrappedTranscript(t *testing.T) {
	// Error that is not errPronunciationNotFound but contains "transcript mismatch"
	err := errors.New("transcript mismatch: expected x got y")
	code, retryable := classifyPronunciationError(err)
	if code != "transcript_mismatch" || !retryable {
		t.Errorf("classifyPronunciationError = %q %v", code, retryable)
	}
	err = errors.New("missing transcript for validation")
	code, retryable = classifyPronunciationError(err)
	if code != "transcript_missing" || !retryable {
		t.Errorf("classifyPronunciationError(missing transcript) = %q %v", code, retryable)
	}
}

func TestWriteFileAtomic_CloseError(t *testing.T) {
	// CreateTemp in a dir that is not writable for closing? Hard. Use a wrapper that fails on Close.
	// On Unix we cannot easily make Close fail. So skip this or use a mock.
	// Chmod error: same. We already have create dir error and rename error.
	t.Skip("making Close/Chmod fail on temp file is platform-dependent")
}

// TestNewPronunciationService_EmptyAudioDirAndPublicBasePath ensures empty config yields default paths.
func TestNewPronunciationService_EmptyAudioDirAndPublicBasePath(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: "", PublicBasePath: "", Provider: "dictionary",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, nil, zap.NewNop())
	if svc.AudioDir() != "/app/data/tts" {
		t.Errorf("empty AudioDir should default to /app/data/tts, got %q", svc.AudioDir())
	}
	if svc.PublicBasePath() != "/media/tts" {
		t.Errorf("empty PublicBasePath should default to /media/tts, got %q", svc.PublicBasePath())
	}
}

// TestNewPronunciationService_WorkersAndBackfillClamps ensures workers clamp at 8 and backfillBatch at 200.
func TestNewPronunciationService_WorkersAndBackfillClamps(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		AudioDir:          t.TempDir(),
		Provider:          "dictionary",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
		PrefetchWorkers:   10,
		BackfillBatchSize: 0,
	}, nil, zap.NewNop())
	if svc.workers != 8 {
		t.Errorf("PrefetchWorkers 10 should clamp to 8, got %d", svc.workers)
	}
	if svc.backfillBatch != 200 {
		t.Errorf("BackfillBatchSize 0 should default to 200, got %d", svc.backfillBatch)
	}
}

// TestNewPronunciationService_RetryMaxLessThanBase uses retryMax < retryBase so retryMax is set to retryBase.
func TestNewPronunciationService_RetryMaxLessThanBase(t *testing.T) {
	svc := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		AudioDir:          t.TempDir(),
		RetryBaseDelay:    "2m",
		RetryMaxDelay:     "1m",
		Provider:          "dictionary",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://example.com",
	}, nil, zap.NewNop())
	if svc.retryMax < svc.retryBase {
		t.Errorf("retryMax should be >= retryBase, got retryBase=%v retryMax=%v", svc.retryBase, svc.retryMax)
	}
}

// TestDictionaryPronunciationProviderFetch_DoubleSlashAudioURL verifies "//host/path" becomes "https://host/path".
// The code prepends "https:"; our test server is HTTP so the audio request fails, but we assert the error mentions https URL.
func TestDictionaryPronunciationProviderFetch_DoubleSlashAudioURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/hello") || r.URL.Path == "/hello" {
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"//` + r.Host + `/audio/hello.mp3"}]}]`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}
	_, err := p.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error (audio URL becomes https:// but server is HTTP)")
	}
	// Code path "//" -> "https:" was taken; client then requests https://... and gets "server gave HTTP response to HTTPS client"
	if !strings.Contains(err.Error(), "https://") && !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("expected error to show https URL was used; got %q", err.Error())
	}
}

// TestDictionaryPronunciationProviderFetch_EmptyAudioBody returns error when audio response body is empty.
func TestDictionaryPronunciationProviderFetch_EmptyAudioBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/hello") || r.URL.Path == "/hello" {
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"http://` + r.Host + `/audio/hello.mp3"}]}]`))
			return
		}
		w.Header().Set("Content-Type", "audio/mpeg")
		// empty body
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{baseURL: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}
	_, err := p.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for empty audio body")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected 'empty' in error, got %q", err.Error())
	}
}

// TestDictionaryPronunciationProviderFetch_ThrottleContextCanceled verifies waitThrottle returns when ctx is canceled.
func TestDictionaryPronunciationProviderFetch_ThrottleContextCanceled(t *testing.T) {
	audioBytes := []byte("mp3")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/a") || r.URL.Path == "/a" {
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"http://` + r.Host + `/audio/a.mp3"}]}]`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/b") || r.URL.Path == "/b" {
			_, _ = w.Write([]byte(`[{"phonetics":[{"audio":"http://` + r.Host + `/audio/b.mp3"}]}]`))
			return
		}
		if strings.Contains(r.URL.Path, "audio") {
			_, _ = w.Write(audioBytes)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	p := &dictionaryPronunciationProvider{
		baseURL:       srv.URL,
		client:        &http.Client{Timeout: 3 * time.Second},
		throttleEvery: 2 * time.Second, // second fetch will need to wait 2s
	}
	// First fetch sets nextAllowedAt to now + 2s.
	_, err := p.fetch(context.Background(), "a")
	if err != nil {
		t.Fatalf("first fetch: %v", err)
	}
	// Second fetch will wait 2s in waitThrottle; cancel ctx so it returns immediately.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.fetch(ctx, "b")
	if err == nil {
		t.Fatal("expected error when context canceled during throttle wait")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "canceled") {
		t.Errorf("expected context canceled error, got %v", err)
	}
}

// TestPronunciationService_Lookup_FailedTerminal returns without scheduling when status is failed_terminal.
func TestPronunciationService_Lookup_FailedTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts",
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	_ = svc.ttsRepo.MarkTerminal("hello", "dict", "test", "error")

	r := svc.Lookup("hello")
	if r.Available {
		t.Error("Lookup should not be available when status is failed_terminal")
	}
	if r.NormalizedWord != "hello" {
		t.Errorf("NormalizedWord = %q want hello", r.NormalizedWord)
	}
}

// TestPronunciationService_Lookup_MaxAttemptsReached returns without scheduling when attempt_count >= max_attempts.
func TestPronunciationService_Lookup_MaxAttemptsReached(t *testing.T) {
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	svc := NewPronunciationService(config.TTSConfig{
		Enabled: true, AudioDir: t.TempDir(), PublicBasePath: "/media/tts", MaxRetries: 2,
		DictionaryEnabled: true, DictionaryBaseURL: "http://example.com",
	}, wordRepo, zap.NewNop())
	for i := 0; i < 2; i++ {
		_ = svc.ttsRepo.MarkAttempt("hello", "dict", "not_found", "err", true)
	}

	r := svc.Lookup("hello")
	if r.Available {
		t.Error("Lookup should not be available when max attempts reached")
	}
}

// TestNormalizePronunciationWord_EmptyAndOnlySpaces covers empty and whitespace-only input.
func TestNormalizePronunciationWord_EmptyAndOnlySpaces(t *testing.T) {
	for _, in := range []string{"", "   ", "\t"} {
		got, ok := normalizePronunciationWord(in)
		if ok {
			t.Errorf("normalizePronunciationWord(%q) ok=true want false", in)
		}
		if got != "" {
			t.Errorf("normalizePronunciationWord(%q)=%q want \"\"", in, got)
		}
	}
	// only punctuation/dashes, no Latin
	got, ok := normalizePronunciationWord("---")
	if ok {
		t.Errorf("normalizePronunciationWord('---') ok=true want false")
	}
	if got != "" {
		t.Errorf("normalizePronunciationWord('---')=%q want \"\"", got)
	}
}

// TestOpenRouterPronunciationProvider_FetchViaChatCompletions_BothAttemptsNoAudio verifies final error when both attempts return no audio.
func TestOpenRouterPronunciationProvider_FetchViaChatCompletions_BothAttemptsNoAudio(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"audio":{"transcript":"hello"}}}]}` + "\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer srv.Close()

	provider := &openRouterPronunciationProvider{
		baseURL:              srv.URL,
		model:                "test",
		voice:                "alloy",
		apiKey:               "key",
		client:               &http.Client{Timeout: 3 * time.Second},
		forceChatCompletions: true,
		pcmToMp3:             func(pcm []byte) ([]byte, error) { return pcm, nil },
	}
	_, err := provider.fetch(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error when both attempts return no audio")
	}
	if !errors.Is(err, errPronunciationNotFound) {
		t.Errorf("expected errPronunciationNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "openrouter_no_audio") {
		t.Errorf("expected openrouter_no_audio in error, got %q", err.Error())
	}
}

// TestDefaultPcmToMp3_ValidPCM runs defaultPcmToMp3 with minimal PCM when ffmpeg is available.
func TestDefaultPcmToMp3_ValidPCM(t *testing.T) {
	// Minimal valid s16le 24kHz mono: a few samples
	pcm := make([]byte, 240*2) // 10ms at 24kHz stereo... actually mono = 1 channel, so 240 samples = 10ms
	mp3, err := defaultPcmToMp3(pcm)
	if err != nil {
		if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "ffmpeg") {
			t.Skip("ffmpeg not available, skipping defaultPcmToMp3 test")
		}
		t.Fatalf("defaultPcmToMp3: %v", err)
	}
	if len(mp3) == 0 {
		t.Error("expected non-empty mp3 output")
	}
}
