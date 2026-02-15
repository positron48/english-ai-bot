package repository

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupGrammarAttemptRepo(t *testing.T) *GrammarAttemptRepository {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	return NewGrammarAttemptRepository(conn, logger)
}

func TestGrammarAttemptRepository_CreateAndProgress(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(1)

	finished := time.Now()
	attempt := &TestAttempt{
		UserID:         user.ID,
		ScopeType:      "chapter",
		ScopeID:        "chapter-1",
		StartedAt:      time.Now(),
		FinishedAt:     &finished,
		Score:          80,
		Passed:         true,
		TotalQuestions: 10,
		AnswersJSON:    "[]",
		ResultsJSON:    "[]",
	}

	if _, err := repo.CreateAttempt(attempt); err != nil {
		t.Fatalf("CreateAttempt error: %v", err)
	}

	if err := repo.UpdateProgress(user.ID, "chapter-1", 80, true); err != nil {
		t.Fatalf("UpdateProgress error: %v", err)
	}

	progress, err := repo.GetChapterProgress(user.ID, "chapter-1")
	if err != nil {
		t.Fatalf("GetChapterProgress error: %v", err)
	}
	if progress.BestScore != 80 {
		t.Fatalf("expected best score 80, got %d", progress.BestScore)
	}
}

func TestGrammarAttemptRepository_CategoryProgress(t *testing.T) {
	repo := setupGrammarAttemptRepo(t)
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(2)

	finished := time.Now()
	attempt := &TestAttempt{
		UserID:         user.ID,
		ScopeType:      "category",
		ScopeID:        "section-1",
		StartedAt:      time.Now(),
		FinishedAt:     &finished,
		Score:          55,
		Passed:         true,
		TotalQuestions: 5,
		AnswersJSON:    "[]",
		ResultsJSON:    "[]",
	}

	if _, err := repo.CreateAttempt(attempt); err != nil {
		t.Fatalf("CreateAttempt error: %v", err)
	}

	hasAttempt, err := repo.HasCategoryTestAttempt(user.ID, "section-1")
	if err != nil {
		t.Fatalf("HasCategoryTestAttempt error: %v", err)
	}
	if !hasAttempt {
		t.Fatalf("expected category attempt")
	}

	passed, err := repo.GetCategoryTestProgress(user.ID, "section-1")
	if err != nil {
		t.Fatalf("GetCategoryTestProgress error: %v", err)
	}
	if !passed {
		t.Fatalf("expected category test to be passed")
	}

	bestScore, err := repo.GetCategoryTestBestScore(user.ID, "section-1")
	if err != nil {
		t.Fatalf("GetCategoryTestBestScore error: %v", err)
	}
	if bestScore != 55 {
		t.Fatalf("expected best score 55, got %d", bestScore)
	}
}
