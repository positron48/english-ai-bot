package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

var errPronunciationNotFound = errors.New("pronunciation not found")

type dictPhonetic struct {
	Audio string `json:"audio"`
}

type dictEntry struct {
	Phonetics []dictPhonetic `json:"phonetics"`
}

type pronunciationProvider interface {
	name() string
	fetch(ctx context.Context, word string) ([]byte, error)
}

// pcmToMp3Func converts raw PCM (s16le 24kHz mono) to MP3. Used for OpenRouter chat/audio path; test can inject identity.
type pcmToMp3Func func(pcm []byte) ([]byte, error)

type openRouterPronunciationProvider struct {
	baseURL             string
	model               string
	voice               string
	apiKey              string
	client              *http.Client
	pcmToMp3            pcmToMp3Func // nil = use ffmpeg
	forceChatCompletions bool        // if true, use chat path even when baseURL is not openrouter (for tests)
}

func (p *openRouterPronunciationProvider) name() string {
	return "openrouter"
}

func isOpenRouterBaseURL(baseURL string) bool {
	u := strings.ToLower(strings.TrimSpace(baseURL))
	return strings.Contains(u, "openrouter.ai")
}

func (p *openRouterPronunciationProvider) fetch(ctx context.Context, word string) ([]byte, error) {
	if p.forceChatCompletions || isOpenRouterBaseURL(p.baseURL) {
		return p.fetchViaChatCompletions(ctx, word)
	}
	return p.fetchViaAudioSpeech(ctx, word)
}

// fetchViaAudioSpeech uses POST /audio/speech (OpenAI); returns MP3 body as-is.
func (p *openRouterPronunciationProvider) fetchViaAudioSpeech(ctx context.Context, word string) ([]byte, error) {
	body := map[string]interface{}{
		"input":           word,
		"model":           p.model,
		"voice":           p.voice,
		"response_format": "mp3",
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal openrouter tts payload: %w", err)
	}
	endpoint := strings.TrimRight(p.baseURL, "/") + "/audio/speech"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build openrouter tts request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter tts request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%w: openrouter endpoint/model rejected request (%d): %s", errPronunciationNotFound, resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("openrouter tts request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	audio, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("openrouter tts read body: %w", err)
	}
	if len(audio) == 0 {
		return nil, errPronunciationNotFound
	}
	return audio, nil
}

// fetchViaChatCompletions uses POST /chat/completions with modalities audio (OpenRouter); parses SSE, collects PCM, converts to MP3.
func (p *openRouterPronunciationProvider) fetchViaChatCompletions(ctx context.Context, word string) ([]byte, error) {
	quotedWord := "`" + word + "`"
	userPrompt := "You are a pronunciation machine. Say ONLY the exact word below as audio. One word, no greeting, no pause, no repetition. Word: " + quotedWord
	payload := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
		"modalities": []string{"text", "audio"},
		"audio":      map[string]string{"voice": p.voice, "format": "pcm16"},
		"stream":     true,
		"max_tokens": 150,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal openrouter chat payload: %w", err)
	}
	endpoint := strings.TrimRight(p.baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build openrouter chat request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter chat request failed: %w", err)
	}
	defer resp.Body.Close()
	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if !strings.Contains(contentType, "text/event-stream") {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
			return nil, fmt.Errorf("%w: openrouter chat rejected (%d): %s", errPronunciationNotFound, resp.StatusCode, strings.TrimSpace(string(payload)))
		}
		return nil, fmt.Errorf("openrouter chat unexpected content-type %q (%d): %s", contentType, resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
			return nil, fmt.Errorf("%w: openrouter chat rejected (%d): %s", errPronunciationNotFound, resp.StatusCode, strings.TrimSpace(string(payload)))
		}
		return nil, fmt.Errorf("openrouter chat rejected (%d): %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	pcm, transcript, err := parseSSEAudioStream(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openrouter chat stream: %w", err)
	}
	if len(pcm) == 0 {
		return nil, errPronunciationNotFound
	}
	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("%w: missing transcript for validation", errPronunciationNotFound)
	}
	if !isPronunciationTranscriptMatch(word, transcript) {
		return nil, fmt.Errorf("%w: transcript mismatch: expected %q got %q", errPronunciationNotFound, word, transcript)
	}
	convert := p.pcmToMp3
	if convert == nil {
		convert = defaultPcmToMp3
	}
	mp3, err := convert(pcm)
	if err != nil {
		return nil, fmt.Errorf("openrouter pcm to mp3: %w", err)
	}
	return mp3, nil
}

const pcmSampleRate = 24000 // gpt-4o-audio pcm16 24kHz mono

func defaultPcmToMp3(pcm []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	trimFilter := "silenceremove=start_periods=1:start_threshold=-45dB,areverse,silenceremove=start_periods=1:start_threshold=-45dB,areverse"
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-loglevel", "error",
		"-f", "s16le", "-ar", fmt.Sprintf("%d", pcmSampleRate), "-ac", "1",
		"-i", "pipe:0", "-af", trimFilter, "-f", "mp3", "pipe:1")
	cmd.Stdin = bytes.NewReader(pcm)
	var out bytes.Buffer
	cmd.Stdout = &out
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w (stderr: %s)", err, errOut.Bytes())
	}
	return out.Bytes(), nil
}

// parseSSEAudioStream reads SSE from r and collects all base64 delta.audio.data into raw PCM and text delta as transcript.
func parseSSEAudioStream(r io.Reader) ([]byte, string, error) {
	var pcm bytes.Buffer
	var text strings.Builder
	scanner := bufio.NewScanner(r)
	const maxLine = 1024 * 1024
	scanner.Buffer(make([]byte, 0, 64*1024), maxLine)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}
		jsonPart := line[6:]
		if bytes.Equal(jsonPart, []byte("[DONE]")) {
			continue
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Audio struct {
						Data string `json:"data"`
						Transcript string `json:"transcript"`
					} `json:"audio"`
					Content []struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(jsonPart, &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		if delta.Audio.Data != "" {
			decoded, err := base64.StdEncoding.DecodeString(delta.Audio.Data)
			if err == nil {
				pcm.Write(decoded)
			}
		}
		if strings.TrimSpace(delta.Audio.Transcript) != "" {
			text.WriteString(delta.Audio.Transcript)
		}
		for _, content := range delta.Content {
			if strings.TrimSpace(content.Text) == "" {
				continue
			}
			if content.Type == "" || strings.EqualFold(content.Type, "text") {
				text.WriteString(content.Text)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return pcm.Bytes(), text.String(), nil
}

func isPronunciationTranscriptMatch(word, transcript string) bool {
	return normalizePronunciationTranscript(word) == normalizePronunciationTranscript(transcript)
}

func normalizePronunciationTranscript(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.Trim(s, "'\"`")
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			continue
		}
		if r == '\'' || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

type dictionaryPronunciationProvider struct {
	baseURL string
	client  *http.Client
}

func (p *dictionaryPronunciationProvider) name() string {
	return "dictionary"
}

func (p *dictionaryPronunciationProvider) fetch(ctx context.Context, word string) ([]byte, error) {
	lookupWord := dictionaryLookupWord(word)
	if lookupWord == "" {
		return nil, errPronunciationNotFound
	}

	reqURL := strings.TrimRight(p.baseURL, "/") + "/" + url.PathEscape(lookupWord)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build dictionary request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dictionary lookup failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errPronunciationNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("dictionary lookup failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var entries []dictEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decode dictionary response: %w", err)
	}

	audioURL := pickDictionaryAudioURL(entries)
	if audioURL == "" {
		return nil, errPronunciationNotFound
	}
	if strings.HasPrefix(audioURL, "//") {
		audioURL = "https:" + audioURL
	}

	audioReq, err := http.NewRequestWithContext(ctx, http.MethodGet, audioURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build audio download request: %w", err)
	}
	audioResp, err := p.client.Do(audioReq)
	if err != nil {
		return nil, fmt.Errorf("dictionary audio download failed: %w", err)
	}
	defer audioResp.Body.Close()

	if audioResp.StatusCode == http.StatusNotFound {
		return nil, errPronunciationNotFound
	}
	if audioResp.StatusCode < 200 || audioResp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(audioResp.Body, 1024))
		return nil, fmt.Errorf("dictionary audio download failed with status %d: %s", audioResp.StatusCode, strings.TrimSpace(string(payload)))
	}

	audio, err := io.ReadAll(io.LimitReader(audioResp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("read dictionary audio: %w", err)
	}
	if len(audio) == 0 {
		return nil, fmt.Errorf("dictionary audio is empty")
	}

	return audio, nil
}

func dictionaryLookupWord(word string) string {
	w := strings.TrimSpace(strings.ToLower(word))
	if strings.HasPrefix(w, "to ") && len(w) > 3 {
		w = strings.TrimSpace(strings.TrimPrefix(w, "to "))
	}
	return w
}

func pickDictionaryAudioURL(entries []dictEntry) string {
	var fallback string
	for _, entry := range entries {
		for _, ph := range entry.Phonetics {
			if strings.TrimSpace(ph.Audio) == "" {
				continue
			}
			audioURL := strings.TrimSpace(ph.Audio)
			if strings.Contains(strings.ToLower(audioURL), ".mp3") {
				return audioURL
			}
			if fallback == "" {
				fallback = audioURL
			}
		}
	}
	return fallback
}

type retryState struct {
	attempt int
	nextTry time.Time
}

// PronunciationLookupResult describes availability of cached pronunciation audio.
type PronunciationLookupResult struct {
	NormalizedWord string
	Available      bool
	URL            string
}

// PronunciationService manages word pronunciation audio generation, caching and background prefetch.
type PronunciationService struct {
	enabled         bool
	prefetchEnabled bool
	audioDir        string
	publicBasePath  string
	retryBase       time.Duration
	retryMax        time.Duration
	maxRetries      int
	backfillEvery   time.Duration
	backfillBatch   int
	workers         int

	wordRepo   *repository.WordRepository
	logger     *zap.Logger
	providers  []pronunciationProvider
	queue      chan string
	queueState map[string]struct{}
	retries    map[string]retryState
	mu         sync.Mutex
}

// NewPronunciationService creates a pronunciation service from config.
func NewPronunciationService(cfg config.TTSConfig, wordRepo *repository.WordRepository, logger *zap.Logger) *PronunciationService {
	if logger == nil {
		logger = zap.NewNop()
	}

	retryBase := parseDurationWithDefault(cfg.RetryBaseDelay, time.Minute)
	retryMax := parseDurationWithDefault(cfg.RetryMaxDelay, 24*time.Hour)
	backfillEvery := parseDurationWithDefault(cfg.BackfillInterval, 10*time.Minute)
	if retryBase <= 0 {
		retryBase = time.Minute
	}
	if retryMax < retryBase {
		retryMax = retryBase
	}
	if backfillEvery <= 0 {
		backfillEvery = 10 * time.Minute
	}

	workers := cfg.PrefetchWorkers
	if workers <= 0 {
		workers = 1
	}
	if workers > 8 {
		workers = 8
	}

	backfillBatch := cfg.BackfillBatchSize
	if backfillBatch <= 0 {
		backfillBatch = 200
	}
	if backfillBatch > 2000 {
		backfillBatch = 2000
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}
	if maxRetries > 20 {
		maxRetries = 20
	}

	service := &PronunciationService{
		enabled:         cfg.Enabled,
		prefetchEnabled: cfg.PrefetchEnabled,
		audioDir:        strings.TrimSpace(cfg.AudioDir),
		publicBasePath:  strings.TrimSpace(cfg.PublicBasePath),
		retryBase:       retryBase,
		retryMax:        retryMax,
		maxRetries:      maxRetries,
		backfillEvery:   backfillEvery,
		backfillBatch:   backfillBatch,
		workers:         workers,
		wordRepo:        wordRepo,
		logger:          logger,
		queue:           make(chan string, 4096),
		queueState:      make(map[string]struct{}),
		retries:         make(map[string]retryState),
	}

	if service.audioDir == "" {
		service.audioDir = "/app/data/tts"
	}
	if service.publicBasePath == "" {
		service.publicBasePath = "/media/tts"
	}
	service.publicBasePath = "/" + strings.Trim(service.publicBasePath, "/")

	service.providers = buildPronunciationProviders(cfg, logger)
	if len(service.providers) == 0 {
		service.enabled = false
		logger.Warn("tts disabled: no pronunciation providers available")
	}

	return service
}

func buildPronunciationProviders(cfg config.TTSConfig, logger *zap.Logger) []pronunciationProvider {
	timeout := parseDurationWithDefault(cfg.RequestTimeout, 45*time.Second)
	client := &http.Client{Timeout: timeout}
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = "auto"
	}

	var providers []pronunciationProvider
	dictEnabled := cfg.DictionaryEnabled
	openRouterEnabled := strings.TrimSpace(cfg.APIKey) != "" && strings.TrimSpace(cfg.Model) != ""

	addDictionary := func() {
		if !dictEnabled {
			return
		}
		baseURL := strings.TrimSpace(cfg.DictionaryBaseURL)
		if baseURL == "" {
			baseURL = "https://api.dictionaryapi.dev/api/v2/entries/en"
		}
		providers = append(providers, &dictionaryPronunciationProvider{
			baseURL: baseURL,
			client:  client,
		})
	}
	addOpenRouter := func() {
		if !openRouterEnabled {
			return
		}
		baseURL := strings.TrimSpace(cfg.BaseURL)
		if baseURL == "" {
			baseURL = "https://openrouter.ai/api/v1"
		}
		voice := strings.TrimSpace(cfg.Voice)
		if voice == "" {
			voice = "alloy"
		}
		providers = append(providers, &openRouterPronunciationProvider{
			baseURL:             baseURL,
			model:               strings.TrimSpace(cfg.Model),
			voice:               voice,
			apiKey:              strings.TrimSpace(cfg.APIKey),
			client:              client,
			forceChatCompletions: isOpenRouterBaseURL(baseURL),
		})
	}

	switch provider {
	case "dictionary":
		addDictionary()
	case "openrouter":
		addOpenRouter()
	case "auto":
		addDictionary()
		addOpenRouter()
	default:
		logger.Warn("unknown tts provider, falling back to auto", zap.String("provider", provider))
		addDictionary()
		addOpenRouter()
	}

	return providers
}

func parseDurationWithDefault(raw string, fallback time.Duration) time.Duration {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}

// IsEnabled reports whether pronunciation service is active.
func (s *PronunciationService) IsEnabled() bool {
	return s != nil && s.enabled
}

// AudioDir returns absolute/relative filesystem directory where mp3 cache is stored.
func (s *PronunciationService) AudioDir() string {
	if s == nil {
		return ""
	}
	return s.audioDir
}

// PublicBasePath returns public URL prefix for pronunciation files.
func (s *PronunciationService) PublicBasePath() string {
	if s == nil {
		return ""
	}
	return s.publicBasePath
}

// Start launches pronunciation workers and background backfill loop.
func (s *PronunciationService) Start(ctx context.Context) {
	if !s.IsEnabled() {
		return
	}
	if err := os.MkdirAll(s.audioDir, 0o755); err != nil {
		s.logger.Error("failed to create tts audio directory", zap.String("dir", s.audioDir), zap.Error(err))
		return
	}

	for i := 0; i < s.workers; i++ {
		go s.worker(ctx)
	}

	if s.prefetchEnabled {
		s.backfillOnce(ctx)
		ticker := time.NewTicker(s.backfillEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.backfillOnce(ctx)
			}
		}
	}

	<-ctx.Done()
}

func (s *PronunciationService) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case word := <-s.queue:
			s.processWord(ctx, word)
		}
	}
}

func (s *PronunciationService) processWord(ctx context.Context, word string) {
	defer s.dequeue(word)

	if !s.IsEnabled() {
		return
	}

	if s.cachedRelPathForWord(word) != "" {
		s.clearRetry(word)
		return
	}

	var retriableErr error
	notFoundSeen := false
	for _, provider := range s.providers {
		audio, err := provider.fetch(ctx, word)
		if err != nil {
			if errors.Is(err, errPronunciationNotFound) {
				notFoundSeen = true
				s.logger.Debug("pronunciation not found in provider", zap.String("provider", provider.name()), zap.String("word", word), zap.Error(err))
				continue
			}
			if retriableErr == nil {
				retriableErr = err
			}
			s.logger.Warn("pronunciation provider failed", zap.String("provider", provider.name()), zap.String("word", word), zap.Error(err))
			continue
		}
		ext := ".mp3"
		relPath := s.relativePathForWordWithExt(word, ext)
		fullPath := filepath.Join(s.audioDir, relPath)
		if err := writeFileAtomic(fullPath, audio, 0o644); err != nil {
			if retriableErr == nil {
				retriableErr = err
			}
			s.logger.Warn("failed to write pronunciation audio", zap.String("path", fullPath), zap.Error(err))
			continue
		}
		s.clearRetry(word)
		s.logger.Info("pronunciation cached", zap.String("word", word), zap.String("path", relPath), zap.String("provider", provider.name()), zap.Int("bytes", len(audio)))
		return
	}

	if retriableErr != nil {
		s.setRetry(word, retriableErr)
		return
	}
	if notFoundSeen {
		s.clearRetry(word)
		s.logger.Info("pronunciation not found in all providers, skip retry", zap.String("word", word))
		return
	}
	s.setRetry(word, errPronunciationNotFound)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "tts-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Chmod(mode); err != nil {
		tmpFile.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func (s *PronunciationService) setRetry(word string, lastErr error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.retries[word]
	state.attempt++
	if state.attempt > s.maxRetries {
		state.attempt = s.maxRetries
	}

	backoff := s.retryBase
	for i := 1; i < state.attempt; i++ {
		backoff *= 2
		if backoff >= s.retryMax {
			backoff = s.retryMax
			break
		}
	}
	if backoff > s.retryMax {
		backoff = s.retryMax
	}

	state.nextTry = time.Now().Add(backoff)
	s.retries[word] = state

	if lastErr != nil {
		s.logger.Warn("pronunciation generation failed, scheduled retry",
			zap.String("word", word),
			zap.Int("attempt", state.attempt),
			zap.Duration("retry_after", backoff),
			zap.Error(lastErr),
		)
	}
}

func (s *PronunciationService) clearRetry(word string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.retries, word)
}

func (s *PronunciationService) dequeue(word string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.queueState, word)
}

func (s *PronunciationService) canScheduleNow(word string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.queueState[word]; exists {
		return false
	}
	if retry, exists := s.retries[word]; exists && retry.nextTry.After(time.Now()) {
		return false
	}
	s.queueState[word] = struct{}{}
	return true
}

func (s *PronunciationService) backfillOnce(ctx context.Context) {
	if s.wordRepo == nil {
		return
	}

	candidates, err := s.wordRepo.ListPronunciationCandidates(s.backfillBatch)
	if err != nil {
		s.logger.Warn("failed to list pronunciation candidates", zap.Error(err))
		return
	}
	for _, candidate := range candidates {
		_ = s.ScheduleWord(candidate)
	}

	select {
	case <-ctx.Done():
		return
	default:
	}
}

// ScheduleWord queues word pronunciation generation in background.
func (s *PronunciationService) ScheduleWord(word string) bool {
	if !s.IsEnabled() {
		return false
	}
	normalized, ok := normalizePronunciationWord(word)
	if !ok {
		return false
	}

	if s.hasCachedAudio(normalized) {
		return true
	}
	if !s.canScheduleNow(normalized) {
		return false
	}

	select {
	case s.queue <- normalized:
		return true
	default:
		go func() {
			select {
			case s.queue <- normalized:
			case <-time.After(2 * time.Second):
				s.dequeue(normalized)
			}
		}()
		return false
	}
}

// ScheduleWords queues pronunciation generation for multiple words.
func (s *PronunciationService) ScheduleWords(words ...string) {
	for _, word := range words {
		_ = s.ScheduleWord(word)
	}
}

// Lookup returns cached pronunciation status for a word and schedules background generation when missing.
func (s *PronunciationService) Lookup(word string) PronunciationLookupResult {
	result := PronunciationLookupResult{}
	if !s.IsEnabled() {
		return result
	}

	normalized, ok := normalizePronunciationWord(word)
	if !ok {
		return result
	}

	result.NormalizedWord = normalized
	if s.hasCachedAudio(normalized) {
		result.Available = true
		result.URL = s.publicURLForWord(normalized)
		return result
	}

	// Missing file: schedule background generation and keep UI non-blocking.
	_ = s.ScheduleWord(normalized)
	return result
}

func (s *PronunciationService) hasCachedAudio(word string) bool {
	return s.cachedRelPathForWord(word) != ""
}

func (s *PronunciationService) publicURLForWord(word string) string {
	relPath := s.cachedRelPathForWord(word)
	if relPath == "" {
		relPath = filepath.ToSlash(s.relativePathForWord(word)) // fallback for URL shape before file exists
	} else {
		relPath = filepath.ToSlash(relPath)
	}
	return s.publicBasePath + "/" + relPath
}

// cachedRelPathForWord returns the relative path of the cached file (.mp3 or .wav) if it exists.
func (s *PronunciationService) cachedRelPathForWord(word string) string {
	key := pronunciationWordKey(word)
	fileBase := pronunciationWordFileBase(word)
	for _, ext := range []string{".mp3", ".wav"} {
		relByWord := filepath.Join(key[:2], key[2:4], fileBase+ext)
		if _, err := os.Stat(filepath.Join(s.audioDir, relByWord)); err == nil {
			return relByWord
		}
	}
	return ""
}

func (s *PronunciationService) relativePathForWord(word string) string {
	key := pronunciationWordKey(word)
	return filepath.Join(key[:2], key[2:4], pronunciationWordFileBase(word)+".mp3")
}

func (s *PronunciationService) relativePathForWordWithExt(word, ext string) string {
	key := pronunciationWordKey(word)
	return filepath.Join(key[:2], key[2:4], pronunciationWordFileBase(word)+ext)
}

func pronunciationWordKey(word string) string {
	sum := sha256.Sum256([]byte(word))
	return hex.EncodeToString(sum[:])
}

func pronunciationWordFileBase(word string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(word)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}
		if r == '-' || r == '\'' {
			b.WriteRune(r)
			continue
		}
		if unicode.IsSpace(r) {
			b.WriteByte('_')
		}
	}

	name := strings.Trim(b.String(), "_")
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	if name == "" {
		return "word"
	}
	return name
}

func normalizePronunciationWord(raw string) (string, bool) {
	word := strings.TrimSpace(strings.ToLower(raw))
	word = strings.Trim(word, ".,!?;:()[]{}\"`")
	word = strings.Join(strings.Fields(word), " ")
	if word == "" {
		return "", false
	}

	hasLatin := false
	for _, r := range word {
		if unicode.IsLetter(r) {
			if !unicode.Is(unicode.Latin, r) {
				return "", false
			}
			hasLatin = true
			continue
		}
		switch r {
		case ' ', '-', '\'', '’':
			continue
		default:
			return "", false
		}
	}

	if !hasLatin {
		return "", false
	}
	return word, true
}
