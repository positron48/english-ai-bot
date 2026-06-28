package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupSpeakingSessionRepo(t *testing.T) (*SpeakingSessionRepository, *sql.DB, int64) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	for _, q := range []string{
		`DELETE FROM speaking_attempts`,
		`DELETE FROM speaking_sessions`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	user, err := NewUserRepository(db, zap.NewNop()).GetOrCreateUser(88001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return NewSpeakingSessionRepository(db), db, user.ID
}

func TestNewSpeakingSessionRepository_NilDB(t *testing.T) {
	if got := NewSpeakingSessionRepository(nil); got != nil {
		t.Fatal("expected nil repo for nil db")
	}
}

func TestSpeakingSessionRepository_CreateGetAdvance(t *testing.T) {
	repo, _, userID := setupSpeakingSessionRepo(t)
	taskIDs := []string{"task-1", "task-2"}

	session, err := repo.CreateSession(userID, "speak.a0", taskIDs)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if session == nil || session.ID == 0 {
		t.Fatalf("session = %#v", session)
	}
	if session.Status != "active" || session.CurrentTaskIndex != 0 || len(session.TaskIDs) != 2 {
		t.Fatalf("session state = %#v", session)
	}

	got, err := repo.GetSession(session.ID, userID)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got == nil || got.ID != session.ID || got.CategoryID != "speak.a0" {
		t.Fatalf("got = %#v", got)
	}

	missing, err := repo.GetSession(session.ID, userID+999)
	if err != nil {
		t.Fatalf("GetSession wrong user: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for wrong user, got %#v", missing)
	}

	if err := repo.AdvanceSession(session.ID, userID, 1, false); err != nil {
		t.Fatalf("AdvanceSession partial: %v", err)
	}
	advanced, err := repo.GetSession(session.ID, userID)
	if err != nil {
		t.Fatalf("GetSession after advance: %v", err)
	}
	if advanced.CurrentTaskIndex != 1 || advanced.Status != "active" || advanced.CompletedAt != nil {
		t.Fatalf("partial advance = %#v", advanced)
	}

	if err := repo.AdvanceSession(session.ID, userID, 2, true); err != nil {
		t.Fatalf("AdvanceSession complete: %v", err)
	}
	completed, err := repo.GetSession(session.ID, userID)
	if err != nil {
		t.Fatalf("GetSession completed: %v", err)
	}
	if completed.Status != "completed" || completed.CompletedAt == nil {
		t.Fatalf("completed session = %#v", completed)
	}
}

func TestSpeakingSessionRepository_Attempts(t *testing.T) {
	repo, _, userID := setupSpeakingSessionRepo(t)
	session, err := repo.CreateSession(userID, "speak.a0", []string{"task-1"})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	count, err := repo.CountAttempts(session.ID, "task-1")
	if err != nil {
		t.Fatalf("CountAttempts empty: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}

	meaning := 4
	grammar := 3
	pron := 5
	fluency := 4
	ok := true
	id, err := repo.SaveAttempt(&SpeakingAttemptRecord{
		UserID:             userID,
		SessionID:          session.ID,
		TaskID:             "task-1",
		AttemptNo:          1,
		Mode:               "speech",
		UnderstoodAnswer:   "hola",
		MeaningScore:       &meaning,
		GrammarScore:       &grammar,
		PronunciationScore: &pron,
		FluencyScore:       &fluency,
		IsAcceptable:       &ok,
		AudioQuality:       "good",
		FeedbackRU:         "Хорошо",
		BetterVersion:      "Hola",
		RepeatTask:         "no",
	})
	if err != nil || id == 0 {
		t.Fatalf("SaveAttempt id=%d err=%v", id, err)
	}

	count, err = repo.CountAttempts(session.ID, "task-1")
	if err != nil {
		t.Fatalf("CountAttempts after save: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	last, err := repo.LastAttemptForTask(session.ID, "task-1")
	if err != nil {
		t.Fatalf("LastAttemptForTask: %v", err)
	}
	if last == nil || last.ID != id || last.UnderstoodAnswer != "hola" {
		t.Fatalf("last = %#v", last)
	}

	none, err := repo.LastAttemptForTask(session.ID, "missing")
	if err != nil {
		t.Fatalf("LastAttemptForTask missing: %v", err)
	}
	if none != nil {
		t.Fatalf("expected nil for missing task, got %#v", none)
	}
}

func TestSpeakingSessionRepository_NilReceiver(t *testing.T) {
	var repo *SpeakingSessionRepository

	if _, err := repo.CreateSession(1, "cat", nil); err == nil {
		t.Fatal("CreateSession expected error for nil db")
	}
	if _, err := repo.GetSession(1, 1); err == nil {
		t.Fatal("GetSession expected error for nil db")
	}
}
