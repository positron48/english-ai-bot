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
