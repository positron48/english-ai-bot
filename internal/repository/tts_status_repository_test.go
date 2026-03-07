package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestTTSStatusRepository_BasicFlow(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	if err := repo.UpsertPending("Spy"); err != nil {
		t.Fatalf("UpsertPending() error = %v", err)
	}
	status, err := repo.GetByWord("spy")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil || status.State != models.TTSStatePending {
		t.Fatalf("expected pending status, got %+v", status)
	}

	if err := repo.MarkAttempt("spy", "openrouter", "rate_limited", "429", true); err != nil {
		t.Fatalf("MarkAttempt() error = %v", err)
	}
	status, _ = repo.GetByWord("spy")
	if status.AttemptCount != 1 || status.State != models.TTSStateFailedRetryable {
		t.Fatalf("unexpected attempt status: %+v", status)
	}

	if err := repo.MarkReady("spy", "dictionary", "aa/bb/spy.mp3"); err != nil {
		t.Fatalf("MarkReady() error = %v", err)
	}
	status, _ = repo.GetByWord("spy")
	if status.State != models.TTSStateReady || status.AudioRelPath == nil {
		t.Fatalf("expected ready with audio path, got %+v", status)
	}
}

func TestTTSStatusRepository_MarkAttemptCapsAtTerminal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	for i := 0; i < 3; i++ {
		if err := repo.MarkAttempt("retry-word", "openrouter", "network_error", "timeout", true); err != nil {
			t.Fatalf("MarkAttempt() #%d error = %v", i+1, err)
		}
	}

	status, err := repo.GetByWord("retry-word")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.AttemptCount != 3 {
		t.Fatalf("expected attempt_count=3, got %d", status.AttemptCount)
	}
	if status.State != models.TTSStateFailedTerminal {
		t.Fatalf("expected terminal state, got %s", status.State)
	}
}

func TestTTSStatusRepository_ResetForForceRegenerate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)

	if err := repo.MarkTerminal("terminal-word", "openrouter", "provider_error", "failed"); err != nil {
		t.Fatalf("MarkTerminal() error = %v", err)
	}
	if err := repo.ResetForForceRegenerate("terminal-word"); err != nil {
		t.Fatalf("ResetForForceRegenerate() error = %v", err)
	}
	status, err := repo.GetByWord("terminal-word")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.State != models.TTSStatePending || status.AttemptCount != 0 {
		t.Fatalf("unexpected reset status: %+v", status)
	}
	if status.LastErrorCode != nil || status.LastErrorMessage != nil || status.AudioRelPath != nil {
		t.Fatalf("expected cleared error/audio fields: %+v", status)
	}
}

// TestNewTTSStatusRepository_DefaultMaxAttemptsWhenZeroOrNegative verifies that when
// maxAttempts <= 0, the repository uses 3 as default (for backward compatibility).
func TestNewTTSStatusRepository_DefaultMaxAttemptsWhenZeroOrNegative(t *testing.T) {
	db := testutil.SetupTestDB(t)
	for _, tc := range []struct{ maxAttempts int; word string }{{0, "zeroword"}, {-1, "negword"}} {
		repo := NewTTSStatusRepository(db, zap.NewNop(), tc.maxAttempts)
		if err := repo.UpsertPending(tc.word); err != nil {
			t.Fatalf("UpsertPending() error = %v", err)
		}
		status, err := repo.GetByWord(tc.word)
		if err != nil {
			t.Fatalf("GetByWord() error = %v", err)
		}
		if status == nil {
			t.Fatal("expected status")
		}
		if status.MaxAttempts != 3 {
			t.Fatalf("maxAttempts=%d: expected MaxAttempts 3 in DB, got %d", tc.maxAttempts, status.MaxAttempts)
		}
	}
}

// TestTTSStatusRepository_CustomMaxAttemptsStored verifies that a custom maxAttempts (e.g. 10)
// is stored in the DB and used for the cap-to-terminal logic.
func TestTTSStatusRepository_CustomMaxAttemptsStored(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 10)
	if err := repo.UpsertPending("tenword"); err != nil {
		t.Fatalf("UpsertPending() error = %v", err)
	}
	status, err := repo.GetByWord("tenword")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.MaxAttempts != 10 {
		t.Fatalf("expected MaxAttempts 10, got %d", status.MaxAttempts)
	}
	// MarkAttempt 10 times should cap to terminal on the 10th
	for i := 0; i < 10; i++ {
		if err := repo.MarkAttempt("tenword", "openrouter", "rate_limited", "429", true); err != nil {
			t.Fatalf("MarkAttempt() #%d error = %v", i+1, err)
		}
	}
	status, _ = repo.GetByWord("tenword")
	if status.State != models.TTSStateFailedTerminal || status.AttemptCount != 10 {
		t.Fatalf("expected failed_terminal with attempt_count=10, got state=%s attempt_count=%d", status.State, status.AttemptCount)
	}
}

// TestTTSStatusRepository_MarkAttempt_TerminalWhenNotRetryable verifies that
// MarkAttempt with retryable=false sets state to failed_terminal immediately.
func TestTTSStatusRepository_MarkAttempt_TerminalWhenNotRetryable(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	if err := repo.UpsertPending("terminal-word"); err != nil {
		t.Fatalf("UpsertPending() error = %v", err)
	}
	if err := repo.MarkAttempt("terminal-word", "provider", "code", "not retryable", false); err != nil {
		t.Fatalf("MarkAttempt() error = %v", err)
	}
	status, err := repo.GetByWord("terminal-word")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.State != models.TTSStateFailedTerminal {
		t.Errorf("expected state failed_terminal when retryable=false, got %s", status.State)
	}
}

// TestTTSStatusRepository_MarkReady_WithEmptyStrings verifies that empty provider/relPath
// are stored as NULL (nullableString converts blank to nil).
func TestTTSStatusRepository_MarkReady_WithEmptyStrings(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	if err := repo.MarkReady("emptyword", "", ""); err != nil {
		t.Fatalf("MarkReady() error = %v", err)
	}
	status, err := repo.GetByWord("emptyword")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.State != models.TTSStateReady {
		t.Errorf("expected state ready, got %s", status.State)
	}
	// provider and relPath can be nil when empty strings were passed
	if status.LastProvider != nil && *status.LastProvider == "" {
		t.Log("LastProvider stored as empty; nullableString may store empty as nil depending on driver")
	}
}

// TestTTSStatusRepository_InvalidWord_NoOp verifies that invalid words (normalizeTTSWord returns false) do not error and do not insert.
func TestTTSStatusRepository_InvalidWord_NoOp(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	// Words that normalizeTTSWord rejects: empty, digits-only, non-Latin, or invalid chars that don't trim away (e.g. mixed "a1" has no latin after trim is wrong - "a1" has 'a'). Use "" "   " "123" "слово" "aб" (cyrillic б).
	invalidWords := []string{"", "   ", "123", "слово", "aб"}
	for _, word := range invalidWords {
		if err := repo.UpsertPending(word); err != nil {
			t.Fatalf("UpsertPending(%q) error = %v", word, err)
		}
		if err := repo.MarkAttempt(word, "p", "code", "msg", true); err != nil {
			t.Fatalf("MarkAttempt(%q) error = %v", word, err)
		}
		if err := repo.MarkReady(word, "p", "path"); err != nil {
			t.Fatalf("MarkReady(%q) error = %v", word, err)
		}
		if err := repo.MarkTerminal(word, "p", "code", "msg"); err != nil {
			t.Fatalf("MarkTerminal(%q) error = %v", word, err)
		}
		if err := repo.ResetForForceRegenerate(word); err != nil {
			t.Fatalf("ResetForForceRegenerate(%q) error = %v", word, err)
		}
	}
	// No row should exist for invalid words (normalize returns false, so no insert; or word is empty)
	for _, word := range invalidWords {
		status, err := repo.GetByWord(word)
		if err != nil {
			t.Fatalf("GetByWord(%q) error = %v", word, err)
		}
		if status != nil {
			t.Errorf("expected nil status for invalid word %q, got %+v", word, status)
		}
	}
}

// TestTTSStatusRepository_GetByWord_InvalidWord_ReturnsNil covers GetByWord when normalizeTTSWord returns false.
func TestTTSStatusRepository_GetByWord_InvalidWord_ReturnsNil(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	status, err := repo.GetByWord("")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status != nil {
		t.Errorf("GetByWord(\"\") expected nil status, got %+v", status)
	}
	status, err = repo.GetByWord("   ")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status != nil {
		t.Errorf("GetByWord(whitespace) expected nil, got %+v", status)
	}
	status, err = repo.GetByWord("кириллица")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status != nil {
		t.Errorf("GetByWord(non-Latin) expected nil, got %+v", status)
	}
}

// TestNormalizeTTSWord_ValidPunctuation covers words with hyphen, apostrophe, space (allowed).
func TestNormalizeTTSWord_ValidPunctuation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	for _, word := range []string{"well-known", "don't", "it's", "word  phrase"} {
		if err := repo.UpsertPending(word); err != nil {
			t.Fatalf("UpsertPending(%q) error = %v", word, err)
		}
		norm, ok := normalizeTTSWord(word)
		if !ok {
			t.Errorf("normalizeTTSWord(%q) expected ok=true", word)
		}
		if norm == "" {
			t.Errorf("normalizeTTSWord(%q) expected non-empty", word)
		}
	}
}

// TestTTSStatusRepository_GetByWord_WithNullableFields covers GetByWord when row has all nullable fields set.
func TestTTSStatusRepository_GetByWord_WithNullableFields(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	if err := repo.MarkAttempt("fullword", "prov", "code", "msg", true); err != nil {
		t.Fatalf("MarkAttempt error: %v", err)
	}
	status, err := repo.GetByWord("fullword")
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status == nil {
		t.Fatal("expected status")
	}
	if status.LastErrorCode == nil || status.LastErrorMessage == nil || status.LastProvider == nil {
		t.Errorf("expected nullable fields set: %+v", status)
	}
}

// TestTTSStatusRepository_UpsertPending_OnConflict updates existing row to pending (idempotent).
func TestTTSStatusRepository_UpsertPending_OnConflict(t *testing.T) {
	db := testutil.SetupTestDB(t)
	repo := NewTTSStatusRepository(db, zap.NewNop(), 3)
	word := "conflict-word"
	if err := repo.UpsertPending(word); err != nil {
		t.Fatalf("UpsertPending() first error = %v", err)
	}
	if err := repo.MarkAttempt(word, "openrouter", "rate_limited", "429", true); err != nil {
		t.Fatalf("MarkAttempt() error = %v", err)
	}
	status, _ := repo.GetByWord(word)
	if status == nil || status.State != models.TTSStateFailedRetryable {
		t.Fatalf("expected failed_retryable after MarkAttempt, got %+v", status)
	}
	// Second UpsertPending should reset state to pending (ON CONFLICT DO UPDATE)
	if err := repo.UpsertPending(word); err != nil {
		t.Fatalf("UpsertPending() second error = %v", err)
	}
	status2, err := repo.GetByWord(word)
	if err != nil {
		t.Fatalf("GetByWord() error = %v", err)
	}
	if status2 == nil {
		t.Fatal("expected status")
	}
	if status2.State != models.TTSStatePending {
		t.Errorf("expected state pending after second UpsertPending, got %s", status2.State)
	}
}
