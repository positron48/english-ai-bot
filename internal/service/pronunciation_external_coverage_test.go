package service

// Coverage tests for external TTS API paths in pronunciation_service.go:
// SetCircuitBreaker, ListPendingExternal, StoreExternalAudio, MarkExternalFailure,
// isLikelyMP3, resolveTTSEnabled, and circuit-breaker branches in processWord.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

const pronunciationExternalUserID = 900001

func pronunciationExternalTTSConfig(audioDir string) config.TTSConfig {
	return config.TTSConfig{
		Enabled:        true,
		Provider:       "external",
		ExternalOnly:   true,
		AudioDir:       audioDir,
		PublicBasePath: "/media/tts",
		MaxRetries:     3,
	}
}

func pronunciationExternalService(t *testing.T, db *sql.DB, audioDir string) *PronunciationService {
	t.Helper()
	wordRepo := repository.NewWordRepository(db, zap.NewNop())
	return NewPronunciationService(
		pronunciationExternalTTSConfig(audioDir),
		config.DefaultLearningConfig(),
		wordRepo,
		zap.NewNop(),
	)
}

func pronunciationExternalInsertWord(t *testing.T, db *sql.DB, word string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO word_cards (word, definition, course_code, created_at, updated_at)
		 VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (word, course_code) DO UPDATE SET updated_at = CURRENT_TIMESTAMP`,
		word, "definition", "en_ru",
	)
	if err != nil {
		t.Fatalf("insert word_cards %q: %v", word, err)
	}
}

func pronunciationExternalSampleMP3() []byte {
	// ID3 header satisfies isLikelyMP3.
	return []byte{'I', 'D', '3', 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
}

func pronunciationExternalMPEGSyncMP3() []byte {
	return []byte{0xFF, 0xFB, 0x90, 0x00, 0x00, 0x00}
}

type mockPronunciationCircuitBreaker struct {
	open             bool
	isOpenErr        error
	recordFailureErr error
	recordSuccessErr error
}

func (m *mockPronunciationCircuitBreaker) IsOpen() (bool, error) {
	return m.open, m.isOpenErr
}

func (m *mockPronunciationCircuitBreaker) RecordFailure(errorMessage string) error {
	return m.recordFailureErr
}

func (m *mockPronunciationCircuitBreaker) RecordSuccess() error {
	return m.recordSuccessErr
}

func TestPronunciationExternal_SetCircuitBreaker(t *testing.T) {
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(t.TempDir()),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	cb := &mockPronunciationCircuitBreaker{open: true}
	svc.SetCircuitBreaker(cb)
	if svc.cbService != cb {
		t.Fatal("SetCircuitBreaker did not wire circuit breaker")
	}
	svc.SetCircuitBreaker(nil)
	if svc.cbService != nil {
		t.Fatal("SetCircuitBreaker(nil) should clear circuit breaker")
	}
}

func TestPronunciationExternal_ResolveTTSEnabled(t *testing.T) {
	tests := []struct {
		name         string
		unsetBoth    bool
		unsetEnabled bool
		key          string
		val          string
		defaultVal   bool
		want         bool
	}{
		{name: "unset env uses default false", unsetBoth: true, defaultVal: false, want: false},
		{name: "unset env uses default true", unsetBoth: true, defaultVal: true, want: true},
		{name: "true", key: "TTS_ENABLED", val: "true", defaultVal: false, want: true},
		{name: "false", key: "TTS_ENABLED", val: "false", defaultVal: true, want: false},
		{name: "1", key: "TTS_ENABLED", val: "1", defaultVal: false, want: true},
		{name: "0", key: "TTS_ENABLED", val: "0", defaultVal: true, want: false},
		{name: "yes", key: "TTS_ENABLED", val: "yes", defaultVal: false, want: true},
		{name: "no", key: "TTS_ENABLED", val: "no", defaultVal: true, want: false},
		{name: "on", key: "TTS_ENABLED", val: "on", defaultVal: false, want: true},
		{name: "off", key: "TTS_ENABLED", val: "off", defaultVal: true, want: false},
		{name: "unknown keeps default true", key: "TTS_ENABLED", val: "maybe", defaultVal: true, want: true},
		{name: "unknown keeps default false", key: "TTS_ENABLED", val: "maybe", defaultVal: false, want: false},
		{name: "TTS_ENABLE key", key: "TTS_ENABLE", val: "false", defaultVal: true, want: false, unsetEnabled: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TTS_ENABLED", "")
			t.Setenv("TTS_ENABLE", "")
			if tt.unsetBoth || tt.unsetEnabled {
				_ = os.Unsetenv("TTS_ENABLED")
			}
			if tt.unsetBoth {
				_ = os.Unsetenv("TTS_ENABLE")
			}
			if tt.key != "" {
				t.Setenv(tt.key, tt.val)
			}
			got := resolveTTSEnabled(tt.defaultVal)
			if got != tt.want {
				t.Fatalf("resolveTTSEnabled(%v) = %v, want %v", tt.defaultVal, got, tt.want)
			}
		})
	}

	t.Run("NewPronunciationService respects env override", func(t *testing.T) {
		t.Setenv("TTS_ENABLED", "false")
		svc := NewPronunciationService(
			config.TTSConfig{Enabled: true, Provider: "external", ExternalOnly: true, AudioDir: t.TempDir()},
			config.DefaultLearningConfig(),
			nil,
			zap.NewNop(),
		)
		if svc.IsEnabled() {
			t.Fatal("expected TTS disabled when TTS_ENABLED=false")
		}
	})
}

func TestPronunciationExternal_IsLikelyMP3(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{name: "too short", data: []byte{0xFF}, want: false},
		{name: "id3", data: pronunciationExternalSampleMP3(), want: true},
		{name: "mpeg sync", data: pronunciationExternalMPEGSyncMP3(), want: true},
		{name: "not mp3", data: []byte("not-an-mp3-file"), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLikelyMP3(tt.data); got != tt.want {
				t.Fatalf("isLikelyMP3() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPronunciationExternal_ListPendingExternal_DependenciesMissing(t *testing.T) {
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(t.TempDir()),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	_, err := svc.ListPendingExternal(10)
	if err == nil || !strings.Contains(err.Error(), "dependencies are not configured") {
		t.Fatalf("expected dependencies error, got %v", err)
	}

	db := testutil.SetupTestDB(t)
	svc2 := NewPronunciationService(
		pronunciationExternalTTSConfig(t.TempDir()),
		config.DefaultLearningConfig(),
		repository.NewWordRepository(db, zap.NewNop()),
		zap.NewNop(),
	)
	svc2.ttsRepo = nil
	_, err = svc2.ListPendingExternal(10)
	if err == nil || !strings.Contains(err.Error(), "dependencies are not configured") {
		t.Fatalf("expected dependencies error when ttsRepo nil, got %v", err)
	}
}

func TestPronunciationExternal_ListPendingExternal_WordStates(t *testing.T) {
	db := testutil.SetupTestDB(t)
	audioDir := t.TempDir()
	svc := pronunciationExternalService(t, db, audioDir)

	// Reserve user ID range for isolation (word_cards do not require user, but keep convention).
	userRepo := repository.NewUserRepository(db, zap.NewNop())
	if _, err := userRepo.GetOrCreateUser(pronunciationExternalUserID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	pronunciationExternalInsertWord(t, db, "extnilstatus")
	pronunciationExternalInsertWord(t, db, "extpending")
	pronunciationExternalInsertWord(t, db, "extretryable")
	pronunciationExternalInsertWord(t, db, "extterminal")
	pronunciationExternalInsertWord(t, db, "extreadymissing")
	pronunciationExternalInsertWord(t, db, "extreadyok")
	pronunciationExternalInsertWord(t, db, "123") // invalid pronunciation word

	_ = svc.ttsRepo.UpsertPending("extpending")
	_ = svc.ttsRepo.MarkAttempt("extretryable", "external", "err", "retry", true)
	_ = svc.ttsRepo.MarkTerminal("extterminal", "external", "terminal", "done")
	_ = svc.ttsRepo.MarkReady("extreadymissing", "external", "missing/file.mp3")
	readyRel := svc.relativePathForWordWithExt("extreadyok", ".mp3")
	readyFull := filepath.Join(audioDir, readyRel)
	if err := os.MkdirAll(filepath.Dir(readyFull), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(readyFull, pronunciationExternalSampleMP3(), 0o644); err != nil {
		t.Fatalf("write ready file: %v", err)
	}
	_ = svc.ttsRepo.MarkReady("extreadyok", "external", readyRel)

	items, err := svc.ListPendingExternal(0)
	if err != nil {
		t.Fatalf("ListPendingExternal: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected pending words")
	}

	words := make(map[string]ExternalPendingWord, len(items))
	for _, item := range items {
		words[item.Word] = item
		if item.TargetLang != "en" {
			t.Errorf("target_lang = %q, want en", item.TargetLang)
		}
	}

	for _, want := range []string{"extnilstatus", "extpending", "extretryable", "extterminal", "extreadymissing"} {
		if _, ok := words[want]; !ok {
			t.Errorf("expected pending word %q in result, got %v", want, words)
		}
	}
	if _, ok := words["extreadyok"]; ok {
		t.Error("ready word with existing file should not be pending")
	}
	if _, ok := words["123"]; ok {
		t.Error("invalid word should be skipped")
	}
}

func TestPronunciationExternal_ListPendingExternal_GetByWordError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := pronunciationExternalService(t, db, t.TempDir())
	pronunciationExternalInsertWord(t, db, "extgeterr")

	getErr := errors.New("get by word failed")
	svc.ttsRepo = &mockTTSRepo{
		getByWordFn: func(word string) (*models.TTSGenerationStatus, error) {
			return nil, getErr
		},
	}

	items, err := svc.ListPendingExternal(5)
	if err != nil {
		t.Fatalf("ListPendingExternal: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty result when GetByWord fails, got %v", items)
	}
}

func TestPronunciationExternal_ListPendingExternal_ListWordsError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := pronunciationExternalService(t, db, t.TempDir())
	dsn := testutil.GetTestDSN(t)
	closedDB, err := sql.Open("postgres_compat", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	closedDB.Close()
	svc.wordRepo = repository.NewWordRepository(closedDB, zap.NewNop())

	_, err = svc.ListPendingExternal(5)
	if err == nil {
		t.Fatal("expected error when ListRecentWordsPage fails")
	}
}

func TestPronunciationExternal_StoreExternalAudio_Validation(t *testing.T) {
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(t.TempDir()),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	mp3 := pronunciationExternalSampleMP3()

	_, err := svc.StoreExternalAudio("hello", "ext", "mp3", mp3)
	if err == nil || !strings.Contains(err.Error(), "status repository is not configured") {
		t.Fatalf("expected ttsRepo error, got %v", err)
	}

	db := testutil.SetupTestDB(t)
	svc = pronunciationExternalService(t, db, t.TempDir())

	cases := []struct {
		name   string
		word   string
		format string
		audio  []byte
		want   string
	}{
		{name: "invalid word", word: "привет", format: "mp3", audio: mp3, want: "invalid word"},
		{name: "unsupported format", word: "hello", format: "wav", audio: mp3, want: "unsupported format"},
		{name: "empty audio", word: "hello", format: "mp3", audio: nil, want: "audio is empty"},
		{name: "not mp3", word: "hello", format: "mp3", audio: []byte("plain-text"), want: "does not look like mp3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.StoreExternalAudio(tc.word, "ext", tc.format, tc.audio)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q in error, got %v", tc.want, err)
			}
		})
	}
}

func TestPronunciationExternal_StoreExternalAudio_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	audioDir := t.TempDir()
	svc := pronunciationExternalService(t, db, audioDir)
	mp3 := pronunciationExternalSampleMP3()

	out, err := svc.StoreExternalAudio("ExternalWord", "", "MP3", mp3)
	if err != nil {
		t.Fatalf("StoreExternalAudio: %v", err)
	}
	if out.State != models.TTSStateReady {
		t.Fatalf("state = %q, want ready", out.State)
	}
	if out.LastProvider != "external" {
		t.Fatalf("provider = %q, want external", out.LastProvider)
	}
	if out.AudioURL == "" {
		t.Fatal("expected audio URL in status")
	}

	rel := svc.relativePathForWordWithExt("externalword", ".mp3")
	full := filepath.Join(audioDir, rel)
	data, err := os.ReadFile(full)
	if err != nil {
		t.Fatalf("read stored file: %v", err)
	}
	if string(data) != string(mp3) {
		t.Fatal("stored audio mismatch")
	}
}

func TestPronunciationExternal_StoreExternalAudio_WriteError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	audioDir := t.TempDir()
	svc := pronunciationExternalService(t, db, audioDir)
	mp3 := pronunciationExternalSampleMP3()

	writeErr := errors.New("write failed")
	orig := createTempFileFn
	createTempFileFn = func(dir, pattern string) (osFile, error) {
		return &mockOsFile{name: filepath.Join(dir, "tts-bad.tmp"), writeErr: writeErr}, nil
	}
	t.Cleanup(func() { createTempFileFn = orig })

	_, err := svc.StoreExternalAudio("writeword", "ext", "mp3", mp3)
	if err == nil || !strings.Contains(err.Error(), "write external audio") {
		t.Fatalf("expected write external audio error, got %v", err)
	}
}

func TestPronunciationExternal_StoreExternalAudio_MarkReadyError(t *testing.T) {
	audioDir := t.TempDir()
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(audioDir),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	markErr := errors.New("mark ready failed")
	svc.ttsRepo = &mockTTSRepo{
		markReadyFn: func(word, provider, relPath string) error {
			return markErr
		},
	}

	_, err := svc.StoreExternalAudio("markfail", "ext", "mp3", pronunciationExternalSampleMP3())
	if !errors.Is(err, markErr) {
		t.Fatalf("expected mark ready error, got %v", err)
	}
}

func TestPronunciationExternal_MarkExternalFailure_Validation(t *testing.T) {
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(t.TempDir()),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	_, err := svc.MarkExternalFailure("hello", "ext", models.TTSStateFailedRetryable, "c", "m")
	if err == nil || !strings.Contains(err.Error(), "status repository is not configured") {
		t.Fatalf("expected ttsRepo error, got %v", err)
	}

	db := testutil.SetupTestDB(t)
	svc = pronunciationExternalService(t, db, t.TempDir())
	_, err = svc.MarkExternalFailure("привет", "ext", models.TTSStateFailedRetryable, "c", "m")
	if err == nil || !strings.Contains(err.Error(), "invalid word") {
		t.Fatalf("expected invalid word error, got %v", err)
	}
}

func TestPronunciationExternal_MarkExternalFailure_RetryableAndTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := pronunciationExternalService(t, db, t.TempDir())

	retryOut, err := svc.MarkExternalFailure("retryext", "", models.TTSStateFailedRetryable, "rate_limited", "429")
	if err != nil {
		t.Fatalf("MarkExternalFailure retryable: %v", err)
	}
	if retryOut.State != models.TTSStateFailedRetryable {
		t.Fatalf("retryable state = %q", retryOut.State)
	}
	if retryOut.LastProvider != "external" {
		t.Fatalf("provider = %q, want external", retryOut.LastProvider)
	}

	termOut, err := svc.MarkExternalFailure("termext", "worker", models.TTSStateFailedTerminal, "bad_word", "unsupported")
	if err != nil {
		t.Fatalf("MarkExternalFailure terminal: %v", err)
	}
	if termOut.State != models.TTSStateFailedTerminal {
		t.Fatalf("terminal state = %q", termOut.State)
	}
	if termOut.LastProvider != "worker" {
		t.Fatalf("provider = %q, want worker", termOut.LastProvider)
	}
}

func TestPronunciationExternal_MarkExternalFailure_RepoErrors(t *testing.T) {
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(t.TempDir()),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	attemptErr := errors.New("mark attempt failed")
	terminalErr := errors.New("mark terminal failed")
	svc.ttsRepo = &mockTTSRepo{
		markAttemptFn: func(word, provider, errorCode, errorMessage string, retryable bool) error {
			return attemptErr
		},
		markTerminalFn: func(word, provider, errorCode, errorMessage string) error {
			return terminalErr
		},
		getByWordFn: func(word string) (*models.TTSGenerationStatus, error) {
			return &models.TTSGenerationStatus{Word: word, State: models.TTSStateFailedRetryable}, nil
		},
	}

	_, err := svc.MarkExternalFailure("a", "ext", models.TTSStateFailedRetryable, "c", "m")
	if !errors.Is(err, attemptErr) {
		t.Fatalf("expected attempt error, got %v", err)
	}
	_, err = svc.MarkExternalFailure("b", "ext", models.TTSStateFailedTerminal, "c", "m")
	if !errors.Is(err, terminalErr) {
		t.Fatalf("expected terminal error, got %v", err)
	}
}

func TestPronunciationExternal_ProcessWord_CircuitBreakerOpen(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := pronunciationExternalService(t, db, t.TempDir())
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "stub", audio: pronunciationExternalSampleMP3()},
	}
	cb := &mockPronunciationCircuitBreaker{open: true}
	svc.SetCircuitBreaker(cb)

	var fetchCalls int
	svc.providers = []pronunciationProvider{
		&countingProvider{
			provider: &stubPronunciationProvider{providerName: "stub", audio: pronunciationExternalSampleMP3()},
			count:    &fetchCalls,
		},
	}

	svc.processWord(context.Background(), "cbopenword")
	if fetchCalls != 0 {
		t.Fatalf("expected no fetch when circuit breaker open, got %d", fetchCalls)
	}
}

func TestPronunciationExternal_ProcessWord_CircuitBreakerIsOpenError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := pronunciationExternalService(t, db, t.TempDir())
	cb := &mockPronunciationCircuitBreaker{isOpenErr: fmt.Errorf("cb check failed")}
	svc.SetCircuitBreaker(cb)

	var fetchCalls int
	svc.providers = []pronunciationProvider{
		&countingProvider{
			provider: &stubPronunciationProvider{providerName: "stub", audio: pronunciationExternalSampleMP3()},
			count:    &fetchCalls,
		},
	}

	svc.processWord(context.Background(), "cbisopenerr")
	if fetchCalls != 1 {
		t.Fatalf("expected fetch to proceed when IsOpen errors, got %d", fetchCalls)
	}
}

func TestPronunciationExternal_ProcessWord_CircuitBreakerRecordFailureError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := pronunciationExternalService(t, db, t.TempDir())
	cb := &mockPronunciationCircuitBreaker{recordFailureErr: fmt.Errorf("record failure failed")}
	svc.SetCircuitBreaker(cb)
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "stub", err: fmt.Errorf("status 429: too many requests")},
	}

	svc.processWord(context.Background(), "cbrecfail")
	status, _ := svc.ttsRepo.GetByWord("cbrecfail")
	if status == nil || status.State != models.TTSStateFailedRetryable {
		t.Fatalf("expected retryable status after provider failure, got %v", status)
	}
}

func TestPronunciationExternal_ProcessWord_CircuitBreakerRecordSuccessError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	audioDir := t.TempDir()
	svc := pronunciationExternalService(t, db, audioDir)
	cb := &mockPronunciationCircuitBreaker{recordSuccessErr: fmt.Errorf("record success failed")}
	svc.SetCircuitBreaker(cb)
	svc.providers = []pronunciationProvider{
		&stubPronunciationProvider{providerName: "stub", audio: pronunciationExternalSampleMP3()},
	}

	svc.processWord(context.Background(), "cbsuccess")
	status, _ := svc.ttsRepo.GetByWord("cbsuccess")
	if status == nil || status.State != models.TTSStateReady {
		t.Fatalf("expected ready status after success, got %v", status)
	}
}

func TestPronunciationExternal_BuildPronunciationProviders_ExternalProvider(t *testing.T) {
	providers := buildPronunciationProviders(config.TTSConfig{
		Provider: "external",
	}, config.DefaultLearningConfig(), zap.NewNop())
	if providers != nil {
		t.Fatalf("expected nil providers for external provider mode, got %v", providers)
	}
}

func TestPronunciationExternal_ListPendingExternal_LimitAndDuplicates(t *testing.T) {
	db := testutil.SetupTestDB(t)
	svc := pronunciationExternalService(t, db, t.TempDir())

	for _, word := range []string{"extlimita", "extlimitb", "extlimitc"} {
		pronunciationExternalInsertWord(t, db, word)
	}
	// Different raw strings that normalize to the same word exercise the seen-map branch.
	for _, word := range []string{"ExtDupMix!", "extdupmix"} {
		_, err := db.Exec(
			`INSERT INTO word_cards (word, definition, course_code, created_at, updated_at)
			 VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			word, "definition", "en_ru",
		)
		if err != nil {
			t.Fatalf("insert duplicate-normalized word %q: %v", word, err)
		}
	}

	items, err := svc.ListPendingExternal(2)
	if err != nil {
		t.Fatalf("ListPendingExternal: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected limit=2 results, got %d: %v", len(items), items)
	}
}

func TestPronunciationExternal_AudioRelPathExists_Guards(t *testing.T) {
	audioDir := t.TempDir()

	tests := []struct {
		name     string
		rel      string
		want     bool
		audioDir string
	}{
		{name: "empty", rel: "", want: false},
		{name: "whitespace", rel: "   ", want: false},
		{name: "dot", rel: ".", want: false},
		{name: "parent traversal", rel: "../secret.mp3", want: false},
		{name: "absolute outside root", rel: "/etc/passwd", want: false},
		{name: "outside audio dir", rel: "../../../tmp/outside.mp3", want: false},
		{name: "sibling path under dot root", rel: "outside/file.mp3", want: false, audioDir: "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := audioDir
			if tt.audioDir != "" {
				dir = tt.audioDir
			}
			svc := NewPronunciationService(
				pronunciationExternalTTSConfig(dir),
				config.DefaultLearningConfig(),
				nil,
				zap.NewNop(),
			)
			if got := svc.audioRelPathExists(tt.rel); got != tt.want {
				t.Fatalf("audioRelPathExists(%q) = %v, want %v", tt.rel, got, tt.want)
			}
		})
	}

	validRel := filepath.Join("en_ru", "ab", "cd", "valid.mp3")
	validFull := filepath.Join(audioDir, validRel)
	if err := os.MkdirAll(filepath.Dir(validFull), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(validFull, pronunciationExternalSampleMP3(), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(audioDir),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	if !svc.audioRelPathExists(validRel) {
		t.Fatal("expected existing relative path to be accepted")
	}
}

func TestPronunciationExternal_ClassifyPronunciationError_InsufficientBalance(t *testing.T) {
	code, retryable := classifyPronunciationError(errors.New("openrouter rejected (402): balance requires at least $0.10"))
	if code != "insufficient_balance" || !retryable {
		t.Fatalf("classifyPronunciationError() = (%q, %v), want (insufficient_balance, true)", code, retryable)
	}
}

func TestPronunciationExternal_Start_NoProvidersWaitsForCancel(t *testing.T) {
	svc := NewPronunciationService(
		pronunciationExternalTTSConfig(t.TempDir()),
		config.DefaultLearningConfig(),
		nil,
		zap.NewNop(),
	)
	if len(svc.providers) != 0 {
		t.Fatalf("expected no providers in external-only mode, got %d", len(svc.providers))
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		svc.Start(ctx)
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after cancel with zero providers")
	}
}
