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
	"tgbot-skeleton/internal/models"
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
	baseURL              string
	model                string
	voice                string
	apiKey               string
	client               *http.Client
	pcmToMp3             pcmToMp3Func // nil = use ffmpeg
	forceChatCompletions bool         // if true, use chat path even when baseURL is not openrouter (for tests)
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
	bodyBytes, _ := json.Marshal(body)
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
		return nil, fmt.Errorf("%w: openrouter_audio_speech_rejected (%d): %s", errPronunciationNotFound, resp.StatusCode, strings.TrimSpace(string(payload)))
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
		return nil, fmt.Errorf("%w: openrouter_audio_speech_empty", errPronunciationNotFound)
	}
	return audio, nil
}

// fetchViaChatCompletions uses POST /chat/completions with modalities audio (OpenRouter); parses SSE, collects PCM, converts to MP3.
// OpenRouter occasionally returns transcript without audio chunks; retries once on openrouter_no_audio.
func (p *openRouterPronunciationProvider) fetchViaChatCompletions(ctx context.Context, word string) ([]byte, error) {
	audio, err := p.fetchViaChatCompletionsOnce(ctx, word)
	if err == nil {
		return audio, nil
	}
	if !strings.Contains(strings.ToLower(err.Error()), "openrouter_no_audio") {
		return nil, err
	}
	time.Sleep(200 * time.Millisecond)
	return p.fetchViaChatCompletionsOnce(ctx, word)
}

func (p *openRouterPronunciationProvider) fetchViaChatCompletionsOnce(ctx context.Context, word string) ([]byte, error) {
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
	bodyBytes, _ := json.Marshal(payload)
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
			return nil, fmt.Errorf("%w: openrouter_rejected (%d): %s", errPronunciationNotFound, resp.StatusCode, strings.TrimSpace(string(payload)))
		}
		return nil, fmt.Errorf("openrouter chat unexpected content-type %q (%d): %s", contentType, resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
			return nil, fmt.Errorf("%w: openrouter_rejected (%d): %s", errPronunciationNotFound, resp.StatusCode, strings.TrimSpace(string(payload)))
		}
		return nil, fmt.Errorf("openrouter chat rejected (%d): %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	pcm, transcript, stats, err := parseSSEAudioStream(resp.Body)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "sse audio decode failed") {
			return nil, fmt.Errorf("%w: openrouter_audio_decode_failed (%v)", errPronunciationNotFound, err)
		}
		return nil, fmt.Errorf("openrouter chat stream: %w", err)
	}
	if len(pcm) == 0 {
		trimmed := strings.TrimSpace(transcript)
		if len(trimmed) > 120 {
			trimmed = trimmed[:120]
		}
		return nil, fmt.Errorf("%w: openrouter_no_audio (transcript=%q lines=%d data_lines=%d json_chunks=%d json_errors=%d choices=%d audio_chunks=%d decode_errors=%d transcript_parts=%d first_json_error=%q first_json_sample=%q)",
			errPronunciationNotFound, trimmed, stats.LinesTotal, stats.DataLines, stats.JSONChunks, stats.JSONErrors, stats.ChoicesChunks, stats.AudioDataChunks, stats.DecodeErrors, stats.TranscriptParts, stats.FirstJSONError, stats.FirstJSONSample)
	}
	if strings.TrimSpace(transcript) == "" {
		return nil, fmt.Errorf("%w: missing transcript for validation", errPronunciationNotFound)
	}
	if !isPronunciationTranscriptMatch(word, transcript) {
		if !isLikelySingleWordPronunciationTranscript(transcript) {
			return nil, fmt.Errorf("%w: transcript mismatch: expected %q got %q", errPronunciationNotFound, word, transcript)
		}
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
type sseAudioStats struct {
	LinesTotal      int
	DataLines       int
	JSONChunks      int
	JSONErrors      int
	FirstJSONError  string
	FirstJSONSample string
	ChoicesChunks   int
	AudioDataChunks int
	TranscriptParts int
	DecodeErrors    int
}

func parseSSEAudioStream(r io.Reader) ([]byte, string, sseAudioStats, error) {
	var pcm bytes.Buffer
	var text strings.Builder
	stats := sseAudioStats{}
	scanner := bufio.NewScanner(r)
	const maxLine = 1024 * 1024
	scanner.Buffer(make([]byte, 0, 64*1024), maxLine)
	for scanner.Scan() {
		stats.LinesTotal++
		line := bytes.TrimSpace(scanner.Bytes())
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}
		stats.DataLines++
		jsonPart := bytes.TrimSpace(line[5:])
		if bytes.Equal(jsonPart, []byte("[DONE]")) {
			continue
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Audio struct {
						Data       string `json:"data"`
						Transcript string `json:"transcript"`
					} `json:"audio"`
					Content json.RawMessage `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(jsonPart, &chunk); err != nil {
			stats.JSONErrors++
			if stats.FirstJSONError == "" {
				stats.FirstJSONError = err.Error()
				sample := string(jsonPart)
				if len(sample) > 180 {
					sample = sample[:180]
				}
				stats.FirstJSONSample = sample
			}
			continue
		}
		stats.JSONChunks++
		if len(chunk.Choices) == 0 {
			continue
		}
		stats.ChoicesChunks++
		delta := chunk.Choices[0].Delta
		if delta.Audio.Data != "" {
			stats.AudioDataChunks++
			decoded, ok := decodeAudioBase64(delta.Audio.Data)
			if ok {
				pcm.Write(decoded)
			} else {
				stats.DecodeErrors++
			}
		}
		if strings.TrimSpace(delta.Audio.Transcript) != "" {
			stats.TranscriptParts++
			text.WriteString(delta.Audio.Transcript)
		}
		for _, contentText := range extractDeltaContentText(delta.Content) {
			if strings.TrimSpace(contentText) == "" {
				continue
			}
			text.WriteString(contentText)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", stats, err
	}
	if pcm.Len() == 0 && stats.DecodeErrors > 0 {
		return nil, text.String(), stats, fmt.Errorf("sse audio decode failed: chunks=%d", stats.DecodeErrors)
	}
	return pcm.Bytes(), text.String(), stats, nil
}

func decodeAudioBase64(raw string) ([]byte, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, false
	}
	if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
		return decoded, true
	}
	if decoded, err := base64.RawStdEncoding.DecodeString(s); err == nil {
		return decoded, true
	}
	if decoded, err := base64.URLEncoding.DecodeString(s); err == nil {
		return decoded, true
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return decoded, true
	}
	return nil, false
}

func extractDeltaContentText(raw json.RawMessage) []string {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return []string{asString}
	}
	var asItems []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &asItems); err == nil {
		out := make([]string, 0, len(asItems))
		for _, item := range asItems {
			if item.Type == "" || strings.EqualFold(item.Type, "text") {
				out = append(out, item.Text)
			}
		}
		return out
	}
	return nil
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

func isLikelySingleWordPronunciationTranscript(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, ".!?") {
		return false
	}
	fields := strings.Fields(s)
	// allow one lexical token, or two tiny parts produced by models; reject long phrases
	if len(fields) > 2 {
		return false
	}
	joined := strings.Join(fields, "")
	letters := 0
	for _, r := range joined {
		if unicode.IsLetter(r) {
			letters++
		}
	}
	return letters >= 3
}

type dictionaryPronunciationProvider struct {
	baseURL       string
	client        *http.Client
	throttleEvery time.Duration
	throttleMu    sync.Mutex
	nextAllowedAt time.Time
}

func (p *dictionaryPronunciationProvider) name() string {
	return "dictionary"
}

func (p *dictionaryPronunciationProvider) fetch(ctx context.Context, word string) ([]byte, error) {
	lookupWord := dictionaryLookupWord(word)
	if lookupWord == "" {
		return nil, errPronunciationNotFound
	}
	if err := p.waitThrottle(ctx); err != nil {
		return nil, fmt.Errorf("dictionary throttle wait failed: %w", err)
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
		return nil, fmt.Errorf("%w: dictionary_404", errPronunciationNotFound)
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
		return nil, fmt.Errorf("%w: dictionary_no_audio", errPronunciationNotFound)
	}
	if strings.HasPrefix(audioURL, "//") {
		audioURL = "https:" + audioURL
	}

	audioReq, err := http.NewRequestWithContext(ctx, http.MethodGet, audioURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build audio download request: %w", err)
	}
	if err := p.waitThrottle(ctx); err != nil {
		return nil, fmt.Errorf("dictionary throttle wait failed: %w", err)
	}
	audioResp, err := p.client.Do(audioReq)
	if err != nil {
		return nil, fmt.Errorf("dictionary audio download failed: %w", err)
	}
	defer audioResp.Body.Close()

	if audioResp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: dictionary_audio_404", errPronunciationNotFound)
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

func (p *dictionaryPronunciationProvider) waitThrottle(ctx context.Context) error {
	if p.throttleEvery <= 0 {
		return nil
	}
	p.throttleMu.Lock()
	wait := time.Until(p.nextAllowedAt)
	if wait < 0 {
		wait = 0
	}
	p.nextAllowedAt = time.Now().Add(p.throttleEvery)
	p.throttleMu.Unlock()

	if wait == 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
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

// PronunciationLookupResult describes availability of cached pronunciation audio.
type PronunciationLookupResult struct {
	NormalizedWord string
	Available      bool
	URL            string
}

// ttsStatusRepo is the interface used by PronunciationService to interact with TTS status storage.
// The concrete implementation is *repository.TTSStatusRepository; the interface enables testing.
type ttsStatusRepo interface {
	GetByWord(word string) (*models.TTSGenerationStatus, error)
	UpsertPending(word string) error
	MarkAttempt(word, provider, errorCode, errorMessage string, retryable bool) error
	MarkReady(word, provider, relPath string) error
	MarkTerminal(word, provider, errorCode, errorMessage string) error
	ResetForForceRegenerate(word string) error
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

	learning   config.LearningConfig
	wordRepo   *repository.WordRepository
	ttsRepo    ttsStatusRepo
	logger     *zap.Logger
	providers  []pronunciationProvider
	queue      chan string
	queueState map[string]struct{}
	mu         sync.Mutex
}

// NewPronunciationService creates a pronunciation service from config.
func NewPronunciationService(cfg config.TTSConfig, learning config.LearningConfig, wordRepo *repository.WordRepository, logger *zap.Logger) *PronunciationService {
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
		learning:        learning,
		wordRepo:        wordRepo,
		logger:          logger,
		queue:           make(chan string, 4096),
		queueState:      make(map[string]struct{}),
	}
	if wordRepo != nil && wordRepo.DB() != nil {
		maxAttempts := cfg.MaxRetries
		if maxAttempts <= 0 {
			maxAttempts = 10
		}
		if maxAttempts > 20 {
			maxAttempts = 20
		}
		service.ttsRepo = repository.NewTTSStatusRepository(wordRepo.DB(), logger, maxAttempts)
	}

	if service.audioDir == "" {
		service.audioDir = "/app/data/tts"
	}
	if service.publicBasePath == "" {
		service.publicBasePath = "/media/tts"
	}
	service.publicBasePath = "/" + strings.Trim(service.publicBasePath, "/")

	service.providers = buildPronunciationProviders(cfg, learning, logger)
	if len(service.providers) == 0 {
		service.enabled = false
		logger.Warn("tts disabled: no pronunciation providers available")
	}

	return service
}

func buildPronunciationProviders(cfg config.TTSConfig, learning config.LearningConfig, logger *zap.Logger) []pronunciationProvider {
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
			tl := strings.TrimSpace(learning.TargetLang)
			if tl == "" {
				tl = "en"
			}
			baseURL = fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/%s", tl)
		}
		minDelay := parseDurationWithDefault(cfg.DictionaryMinDelay, 100*time.Millisecond)
		if minDelay < 0 {
			minDelay = 0
		}
		providers = append(providers, &dictionaryPronunciationProvider{
			baseURL:       baseURL,
			client:        client,
			throttleEvery: minDelay,
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
			baseURL:              baseURL,
			model:                strings.TrimSpace(cfg.Model),
			voice:                voice,
			apiKey:               strings.TrimSpace(cfg.APIKey),
			client:               client,
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

type TTSDebugProviderResult struct {
	Provider string `json:"provider"`
	Outcome  string `json:"outcome"`
	Reason   string `json:"reason,omitempty"`
	Error    string `json:"error,omitempty"`
	Bytes    int    `json:"bytes,omitempty"`
}

type TTSDebugResult struct {
	Word    string                   `json:"word"`
	Results []TTSDebugProviderResult `json:"results"`
}

// DebugFetch runs TTS providers in the same order as runtime generation and returns step-by-step outcomes.
func (s *PronunciationService) DebugFetch(ctx context.Context, rawWord string) (*TTSDebugResult, error) {
	if s == nil || !s.IsEnabled() {
		return nil, fmt.Errorf("tts is disabled")
	}
	word, ok := normalizePronunciationWord(rawWord)
	if !ok {
		return nil, fmt.Errorf("invalid word: %q", rawWord)
	}
	res := &TTSDebugResult{Word: word, Results: make([]TTSDebugProviderResult, 0, len(s.providers))}
	for _, provider := range s.providers {
		audio, err := provider.fetch(ctx, word)
		if err != nil {
			code, _ := classifyPronunciationError(err)
			outcome := "error"
			if errors.Is(err, errPronunciationNotFound) {
				outcome = "not_found"
			}
			res.Results = append(res.Results, TTSDebugProviderResult{
				Provider: provider.name(),
				Outcome:  outcome,
				Reason:   code,
				Error:    err.Error(),
			})
			continue
		}
		res.Results = append(res.Results, TTSDebugProviderResult{
			Provider: provider.name(),
			Outcome:  "success",
			Bytes:    len(audio),
		})
		return res, nil
	}
	return res, nil
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

	status, err := s.ensureStatusForWord(word)
	if err != nil {
		s.logger.Warn("failed to ensure tts status before generation", zap.String("word", word), zap.Error(err))
		return
	}
	if status != nil && status.State == models.TTSStateFailedTerminal {
		return
	}
	if status != nil && status.AttemptCount >= status.MaxAttempts {
		if s.ttsRepo != nil {
			_ = s.ttsRepo.MarkTerminal(word, "service", "max_attempts_reached", "attempt limit reached")
		}
		return
	}
	if status != nil && status.State == models.TTSStateReady {
		if relPath := s.resolveReadyRelPath(word, status); relPath != "" {
			return
		}
		if s.ttsRepo != nil {
			_ = s.ttsRepo.UpsertPending(word)
		}
	}

	var retryableErr error
	var retryableCode string
	var retryableProvider string
	notFoundSeen := false
	notFoundReasons := make([]string, 0, len(s.providers))
	for i, provider := range s.providers {
		// Do not retry dictionary if we already got dictionary_no_audio for this word.
		if provider.name() == "dictionary" && status != nil &&
			status.LastErrorCode != nil && *status.LastErrorCode == "dictionary_no_audio" &&
			status.LastProvider != nil && *status.LastProvider == "dictionary" {
			s.logger.Debug("skipping dictionary provider, already had dictionary_no_audio", zap.String("word", word))
			continue
		}
		audio, err := provider.fetch(ctx, word)
		if err != nil {
			hasFallback := i < len(s.providers)-1
			if errors.Is(err, errPronunciationNotFound) {
				notFoundSeen = true
				code, retryable := classifyPronunciationError(err)
				notFoundReasons = append(notFoundReasons, provider.name()+":"+code)
				s.logger.Debug("pronunciation not found in provider", zap.String("provider", provider.name()), zap.String("word", word), zap.String("reason", code), zap.Error(err))
				if retryable {
					retryableErr = err
					retryableCode = code
					retryableProvider = provider.name()
				}
				s.logger.Info("tts provider result",
					zap.String("word", word),
					zap.String("provider", provider.name()),
					zap.String("reason", code),
					zap.String("error", err.Error()),
					zap.String("outcome", "not_found"),
					zap.String("decision", map[bool]string{true: "fallback_next", false: "retry_or_terminal_after_chain"}[hasFallback]),
				)
				continue
			}
			code, retryable := classifyPronunciationError(err)
			if retryable && retryableErr == nil {
				retryableErr = err
				retryableCode = code
				retryableProvider = provider.name()
			}
			s.logger.Warn("pronunciation provider failed", zap.String("provider", provider.name()), zap.String("word", word), zap.Error(err))
			if !retryable {
				if s.ttsRepo != nil {
					_ = s.ttsRepo.MarkTerminal(word, provider.name(), code, err.Error())
				}
				s.logTTSStatusDecision(word, "terminal_reached", provider.name(), code)
				return
			}
			s.logger.Info("tts provider result",
				zap.String("word", word),
				zap.String("provider", provider.name()),
				zap.String("error_code", code),
				zap.String("error", err.Error()),
				zap.String("outcome", "retryable_error"),
				zap.String("decision", map[bool]string{true: "fallback_next", false: "retry_or_terminal_after_chain"}[hasFallback]),
			)
			continue
		}
		ext := ".mp3"
		relPath := s.relativePathForWordWithExt(word, ext)
		fullPath := filepath.Join(s.audioDir, relPath)
		if err := writeFileAtomic(fullPath, audio, 0o644); err != nil {
			retryableErr = err
			retryableCode = "write_failed"
			retryableProvider = provider.name()
			s.logger.Warn("failed to write pronunciation audio", zap.String("path", fullPath), zap.Error(err))
			continue
		}
		if s.ttsRepo != nil {
			_ = s.ttsRepo.MarkReady(word, provider.name(), relPath)
		}
		s.logger.Info("pronunciation cached", zap.String("word", word), zap.String("path", relPath), zap.String("provider", provider.name()), zap.Int("bytes", len(audio)))
		return
	}

	if retryableErr != nil {
		if s.ttsRepo != nil {
			_ = s.ttsRepo.MarkAttempt(word, retryableProvider, retryableCode, retryableErr.Error(), true)
		}
		s.logTTSStatusDecision(word, "retry_or_terminal_after_attempt", retryableProvider, retryableCode)
		return
	}
	if notFoundSeen {
		if s.ttsRepo != nil {
			_ = s.ttsRepo.MarkAttempt(word, "all", "not_found_all_providers", "not found in all providers", true)
		}
		s.logger.Info("tts chain result",
			zap.String("word", word),
			zap.String("outcome", "not_found_all_providers"),
			zap.Strings("reasons", notFoundReasons),
		)
		s.logTTSStatusDecision(word, "retry_or_terminal_after_attempt", "all", "not_found_all_providers")
		return
	}
	if s.ttsRepo != nil {
		_ = s.ttsRepo.MarkAttempt(word, "service", "unknown_generation_failure", "unknown generation failure", true)
	}
	s.logTTSStatusDecision(word, "retry_or_terminal_after_attempt", "service", "unknown_generation_failure")
}

// osFile is the subset of *os.File operations used by writeFileAtomic; injectable for testing.
type osFile interface {
	Name() string
	Write(b []byte) (int, error)
	Chmod(mode os.FileMode) error
	Close() error
}

// createTempFileFn is the function used to create temp files; can be overridden in tests.
var createTempFileFn = func(dir, pattern string) (osFile, error) {
	return os.CreateTemp(dir, pattern)
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	tmpFile, err := createTempFileFn(dir, "tts-*.tmp")
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

	status, err := s.ensureStatusForWord(normalized)
	if err != nil {
		s.logger.Warn("failed to check tts status before schedule", zap.String("word", normalized), zap.Error(err))
		return false
	}
	if status != nil && status.State == models.TTSStateFailedTerminal {
		return false
	}
	if status != nil && status.AttemptCount >= status.MaxAttempts {
		if s.ttsRepo != nil {
			_ = s.ttsRepo.MarkTerminal(normalized, "service", "max_attempts_reached", "attempt limit reached")
		}
		return false
	}
	if status != nil && status.State == models.TTSStateReady {
		if relPath := s.resolveReadyRelPath(normalized, status); relPath != "" {
			return true
		}
		if s.ttsRepo != nil {
			_ = s.ttsRepo.UpsertPending(normalized)
		}
	}
	if s.ttsRepo == nil && s.hasCachedAudio(normalized) {
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
		return true
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
	status, err := s.ensureStatusForWord(normalized)
	if err != nil {
		s.logger.Warn("failed to lookup tts status", zap.String("word", normalized), zap.Error(err))
		return result
	}
	if status != nil && status.State == models.TTSStateReady {
		if rel := s.resolveReadyRelPath(normalized, status); rel != "" {
			result.Available = true
			result.URL = s.publicBasePath + "/" + filepath.ToSlash(rel)
			return result
		}
	}
	if s.hasCachedAudio(normalized) {
		result.Available = true
		result.URL = s.publicURLForWord(normalized)
		return result
	}
	if status != nil && status.State == models.TTSStateFailedTerminal {
		return result
	}
	if status != nil && status.AttemptCount >= status.MaxAttempts {
		if s.ttsRepo != nil {
			_ = s.ttsRepo.MarkTerminal(normalized, "service", "max_attempts_reached", "attempt limit reached")
		}
		return result
	}

	// Missing file: schedule background generation and keep UI non-blocking.
	_ = s.ScheduleWord(normalized)
	return result
}

type TTSStatusResult struct {
	Word             string    `json:"word"`
	State            string    `json:"state"`
	AttemptCount     int       `json:"attempt_count"`
	MaxAttempts      int       `json:"max_attempts"`
	LastErrorCode    string    `json:"last_error_code"`
	LastErrorMessage string    `json:"last_error_message"`
	LastProvider     string    `json:"last_provider"`
	AudioURL         string    `json:"audio_url"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (s *PronunciationService) GetStatus(word string) (TTSStatusResult, error) {
	var out TTSStatusResult
	if s.ttsRepo == nil {
		return out, fmt.Errorf("tts status repository is not configured")
	}
	normalized, ok := normalizePronunciationWord(word)
	if !ok {
		return out, fmt.Errorf("invalid word")
	}
	out.Word = normalized
	status, err := s.ensureStatusForWord(normalized)
	if err != nil {
		return out, err
	}
	if status == nil {
		_ = s.ttsRepo.UpsertPending(normalized)
		status, _ = s.ttsRepo.GetByWord(normalized)
	}
	if status != nil {
		out.State = status.State
		out.AttemptCount = status.AttemptCount
		out.MaxAttempts = status.MaxAttempts
		if status.LastErrorCode != nil {
			out.LastErrorCode = *status.LastErrorCode
		}
		if status.LastErrorMessage != nil {
			out.LastErrorMessage = *status.LastErrorMessage
		}
		if status.LastProvider != nil {
			out.LastProvider = *status.LastProvider
		}
		out.UpdatedAt = status.UpdatedAt
		if rel := s.resolveReadyRelPath(normalized, status); rel != "" {
			out.AudioURL = s.publicBasePath + "/" + filepath.ToSlash(rel)
		}
	}
	return out, nil
}

func (s *PronunciationService) ForceRegenerate(word string) (TTSStatusResult, error) {
	var out TTSStatusResult
	if s.ttsRepo == nil {
		return out, fmt.Errorf("tts status repository is not configured")
	}
	normalized, ok := normalizePronunciationWord(word)
	if !ok {
		return out, fmt.Errorf("invalid word")
	}
	if err := s.ttsRepo.ResetForForceRegenerate(normalized); err != nil {
		return out, err
	}
	scheduled := s.ScheduleWord(normalized)
	s.logger.Info("tts regenerate requested",
		zap.String("word", normalized),
		zap.Bool("scheduled", scheduled),
	)
	return s.GetStatus(normalized)
}

func (s *PronunciationService) Recheck(word string) (TTSStatusResult, error) {
	var out TTSStatusResult
	if s.ttsRepo == nil {
		return out, fmt.Errorf("tts status repository is not configured")
	}
	normalized, ok := normalizePronunciationWord(word)
	if !ok {
		return out, fmt.Errorf("invalid word")
	}
	_, _ = s.ensureStatusForWord(normalized)
	return s.GetStatus(normalized)
}

func (s *PronunciationService) ensureStatusForWord(word string) (*models.TTSGenerationStatus, error) {
	if s.ttsRepo == nil {
		return nil, nil
	}
	status, err := s.ttsRepo.GetByWord(word)
	if err != nil {
		return nil, err
	}
	legacyRel := s.cachedRelPathForWord(word)
	if status == nil {
		if legacyRel != "" {
			if err := s.ttsRepo.MarkReady(word, "legacy_file", legacyRel); err != nil {
				return nil, err
			}
			return s.ttsRepo.GetByWord(word)
		}
		return nil, nil
	}
	if status.State == models.TTSStateReady {
		if rel := s.resolveReadyRelPath(word, status); rel == "" {
			if err := s.ttsRepo.UpsertPending(word); err != nil {
				return nil, err
			}
			return s.ttsRepo.GetByWord(word)
		}
	}
	return status, nil
}

func (s *PronunciationService) resolveReadyRelPath(word string, status *models.TTSGenerationStatus) string {
	if status != nil && status.AudioRelPath != nil && strings.TrimSpace(*status.AudioRelPath) != "" {
		if s.audioRelPathExists(*status.AudioRelPath) {
			return *status.AudioRelPath
		}
	}
	return s.cachedRelPathForWord(word)
}

func (s *PronunciationService) audioRelPathExists(relPath string) bool {
	clean := filepath.Clean(relPath)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return false
	}
	_, err := os.Stat(filepath.Join(s.audioDir, clean))
	return err == nil
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

func classifyPronunciationError(err error) (string, bool) {
	if err == nil {
		return "", true
	}
	msg := strings.ToLower(err.Error())
	if errors.Is(err, errPronunciationNotFound) {
		switch {
		case strings.Contains(msg, "dictionary_404"):
			return "dictionary_404", true
		case strings.Contains(msg, "dictionary_no_audio"):
			return "dictionary_no_audio", true
		case strings.Contains(msg, "dictionary_audio_404"):
			return "dictionary_audio_404", true
		case strings.Contains(msg, "openrouter_no_audio"):
			return "openrouter_no_audio", true
		case strings.Contains(msg, "openrouter_rejected"):
			return "openrouter_rejected", true
		case strings.Contains(msg, "openrouter_audio_speech_rejected"):
			return "openrouter_audio_speech_rejected", true
		case strings.Contains(msg, "openrouter_audio_speech_empty"):
			return "openrouter_audio_speech_empty", true
		case strings.Contains(msg, "openrouter_audio_decode_failed"):
			return "openrouter_audio_decode_failed", true
		case strings.Contains(msg, "transcript mismatch"):
			return "transcript_mismatch", true
		case strings.Contains(msg, "missing transcript"):
			return "transcript_missing", true
		case strings.Contains(msg, "not_found_nonretryable"):
			return "not_found_nonretryable", false
		default:
			return "not_found", true
		}
	}
	switch {
	case strings.Contains(msg, "transcript mismatch"):
		return "transcript_mismatch", true
	case strings.Contains(msg, "missing transcript"):
		return "transcript_missing", true
	case strings.Contains(msg, "unsupported_country_region_territory"):
		return "unsupported_country_region_territory", true
	case strings.Contains(msg, "status 429") || strings.Contains(msg, "too many requests"):
		return "rate_limited", true
	case strings.Contains(msg, "status 5"):
		return "provider_5xx", true
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "connection") || strings.Contains(msg, "network"):
		return "network_error", true
	default:
		return "provider_error", false
	}
}

func (s *PronunciationService) logTTSStatusDecision(word, outcome, provider, code string) {
	if s.ttsRepo == nil {
		return
	}
	status, err := s.ttsRepo.GetByWord(word)
	if err != nil || status == nil {
		return
	}
	s.logger.Info("tts decision",
		zap.String("word", word),
		zap.String("provider", provider),
		zap.String("error_code", code),
		zap.String("outcome", outcome),
		zap.String("state", status.State),
		zap.Int("attempt_count", status.AttemptCount),
		zap.Int("max_attempts", status.MaxAttempts),
	)
}
