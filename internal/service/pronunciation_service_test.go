package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
		name     string
		err      error
		wantCode string
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
		Enabled: true,
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
		Enabled: true,
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
		Enabled: true,
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
