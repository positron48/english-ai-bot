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
