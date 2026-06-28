package repository

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupGrammarSRSRepo(t *testing.T) (*GrammarSRSRepository, int64) {
	t.Helper()
	conn := testutil.SetupTestDB(t)
	for _, q := range []string{
		`DELETE FROM grammar_attempts`,
		`DELETE FROM grammar_theory_memory`,
	} {
		if _, err := conn.Exec(q); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
	}
	user, err := NewUserRepository(conn, zap.NewNop()).GetOrCreateUser(89001)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return NewGrammarSRSRepository(conn, zap.NewNop()), user.ID
}

func grammarSRSNow() time.Time {
	return time.Now().UTC().Add(time.Hour)
}

func TestGrammarSRSRepository_EnsureAndListDue(t *testing.T) {
	repo, userID := setupGrammarSRSRepo(t)
	now := grammarSRSNow()

	if err := repo.EnsureTheoryMemory(userID, "es", "es", "ch1", "block1", "concept1"); err != nil {
		t.Fatalf("EnsureTheoryMemory: %v", err)
	}
	if err := repo.EnsureTheoryMemory(userID, "es", "es", "ch1", "block1", "concept1"); err != nil {
		t.Fatalf("EnsureTheoryMemory idempotent: %v", err)
	}

	due, err := repo.ListDueMemories(userID, "es", "es", now, 10)
	if err != nil {
		t.Fatalf("ListDueMemories: %v", err)
	}
	if len(due) != 1 || due[0].TheoryBlockID != "block1" {
		t.Fatalf("due = %+v", due)
	}

	if _, err := repo.db.Exec(`
		UPDATE grammar_theory_memory SET next_review_at = ?
		WHERE user_id = ? AND theory_block_id = 'block1'`, now.Add(48*time.Hour), userID); err != nil {
		t.Fatalf("push review date: %v", err)
	}
	due, err = repo.ListDueMemories(userID, "es", "es", now, 10)
	if err != nil {
		t.Fatalf("ListDueMemories future: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("expected no due items, got %+v", due)
	}
}

func TestGrammarSRSRepository_CountDueOrNewTheoryBlocks(t *testing.T) {
	repo, userID := setupGrammarSRSRepo(t)
	now := grammarSRSNow()

	if n, err := repo.CountDueOrNewTheoryBlocks(userID, "es", "es", nil, now); err != nil || n != 0 {
		t.Fatalf("empty ids: n=%d err=%v", n, err)
	}
	if n, err := repo.CountDueOrNewTheoryBlocks(userID, "es", "es", []string{"", " "}, now); err != nil || n != 0 {
		t.Fatalf("blank ids: n=%d err=%v", n, err)
	}

	if err := repo.EnsureTheoryMemory(userID, "es", "es", "ch1", "block-due", "c1"); err != nil {
		t.Fatalf("EnsureTheoryMemory due: %v", err)
	}
	if err := repo.EnsureTheoryMemory(userID, "es", "es", "ch1", "block-future", "c2"); err != nil {
		t.Fatalf("EnsureTheoryMemory future: %v", err)
	}
	if _, err := repo.db.Exec(`
		UPDATE grammar_theory_memory SET next_review_at = ?
		WHERE user_id = ? AND theory_block_id = 'block-future'`, now.Add(24*time.Hour), userID); err != nil {
		t.Fatalf("set future review: %v", err)
	}

	n, err := repo.CountDueOrNewTheoryBlocks(userID, "es", "es", []string{"block-due", "block-future", "block-new"}, now)
	if err != nil {
		t.Fatalf("CountDueOrNewTheoryBlocks: %v", err)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2 (due + new)", n)
	}
}

func TestGrammarSRSRepository_UpdateAfterAnswer(t *testing.T) {
	repo, userID := setupGrammarSRSRepo(t)
	at := time.Now().UTC()

	if err := repo.EnsureTheoryMemory(userID, "es", "es", "ch1", "block1", "concept1"); err != nil {
		t.Fatalf("EnsureTheoryMemory: %v", err)
	}
	memories, err := repo.ListDueMemories(userID, "es", "es", at.Add(time.Hour), 1)
	if err != nil || len(memories) != 1 {
		t.Fatalf("ListDueMemories: %+v err=%v", memories, err)
	}
	memory := memories[0]

	if err := repo.UpdateAfterAnswerAt(memory, true, at); err != nil {
		t.Fatalf("UpdateAfterAnswerAt correct: %v", err)
	}
	afterCorrect, err := repo.ListDueMemories(userID, "es", "es", at.Add(23*time.Hour), 1)
	if err != nil {
		t.Fatalf("ListDueMemories after correct: %v", err)
	}
	if len(afterCorrect) != 0 {
		t.Fatalf("expected no due after correct answer, got %+v", afterCorrect)
	}

	if err := repo.UpdateAfterAnswerAt(memory, false, at.Add(time.Hour)); err != nil {
		t.Fatalf("UpdateAfterAnswerAt wrong: %v", err)
	}
	afterWrong, err := repo.ListDueMemories(userID, "es", "es", at.Add(2*time.Hour), 1)
	if err != nil {
		t.Fatalf("ListDueMemories after wrong: %v", err)
	}
	if len(afterWrong) != 1 || afterWrong[0].State != "relearning" {
		t.Fatalf("after wrong = %+v", afterWrong)
	}
}

func TestGrammarSRSRepository_SaveAttemptAndHasClientAttempt(t *testing.T) {
	repo, userID := setupGrammarSRSRepo(t)
	answeredAt := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)

	id, err := repo.SaveAttemptWithClientID(
		userID, "es", "es", "ch1", "block1", "concept1", "q1",
		map[string]string{"answer": "x"},
		map[string]string{"answer": "y"},
		true,
		"grammar-client-1",
		&answeredAt,
	)
	if err != nil || id == 0 {
		t.Fatalf("SaveAttemptWithClientID id=%d err=%v", id, err)
	}

	exists, err := repo.HasClientAttempt(userID, "grammar-client-1")
	if err != nil || !exists {
		t.Fatalf("HasClientAttempt existing: exists=%v err=%v", exists, err)
	}
	exists, err = repo.HasClientAttempt(userID, "")
	if err != nil || exists {
		t.Fatalf("HasClientAttempt empty id: exists=%v err=%v", exists, err)
	}
	exists, err = repo.HasClientAttempt(userID, "missing")
	if err != nil || exists {
		t.Fatalf("HasClientAttempt missing: exists=%v err=%v", exists, err)
	}

	if err := repo.SaveAttempt(userID, "es", "es", "ch1", "block1", "concept1", "q2", "a", "b", false); err != nil {
		t.Fatalf("SaveAttempt: %v", err)
	}
}

func TestGrammarSRSRepository_ListDueMemories_DefaultLimit(t *testing.T) {
	repo, userID := setupGrammarSRSRepo(t)
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		blockID := "block-" + string(rune('a'+i))
		if err := repo.EnsureTheoryMemory(userID, "es", "es", "ch1", blockID, "c"); err != nil {
			t.Fatalf("EnsureTheoryMemory %s: %v", blockID, err)
		}
	}
	due, err := repo.ListDueMemories(userID, "es", "es", now.Add(time.Hour), 0)
	if err != nil {
		t.Fatalf("ListDueMemories: %v", err)
	}
	if len(due) != 3 {
		t.Fatalf("len = %d, want 3 with default limit", len(due))
	}
}
