package repository

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/database"

	"go.uber.org/zap"
)

func setupGrammarAttemptRepo(t *testing.T) (*GrammarAttemptRepository, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	repo := NewGrammarAttemptRepository(db.GetConnection(), logger)

	cleanup := func() {
		_ = db.Close()
	}

	return repo, cleanup
}

func TestGrammarAttemptRepository_CreateAndProgress(t *testing.T) {
	repo, cleanup := setupGrammarAttemptRepo(t)
	defer cleanup()

	finished := time.Now()
	attempt := &TestAttempt{
		UserID:         1,
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

	if err := repo.UpdateProgress(1, "chapter-1", 80, true); err != nil {
		t.Fatalf("UpdateProgress error: %v", err)
	}

	progress, err := repo.GetChapterProgress(1, "chapter-1")
	if err != nil {
		t.Fatalf("GetChapterProgress error: %v", err)
	}
	if progress.BestScore != 80 {
		t.Fatalf("expected best score 80, got %d", progress.BestScore)
	}
}

func TestGrammarAttemptRepository_CategoryProgress(t *testing.T) {
	repo, cleanup := setupGrammarAttemptRepo(t)
	defer cleanup()

	finished := time.Now()
	attempt := &TestAttempt{
		UserID:         1,
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

	hasAttempt, err := repo.HasCategoryTestAttempt(1, "section-1")
	if err != nil {
		t.Fatalf("HasCategoryTestAttempt error: %v", err)
	}
	if !hasAttempt {
		t.Fatalf("expected category attempt")
	}

	passed, err := repo.GetCategoryTestProgress(1, "section-1")
	if err != nil {
		t.Fatalf("GetCategoryTestProgress error: %v", err)
	}
	if !passed {
		t.Fatalf("expected category test to be passed")
	}

	bestScore, err := repo.GetCategoryTestBestScore(1, "section-1")
	if err != nil {
		t.Fatalf("GetCategoryTestBestScore error: %v", err)
	}
	if bestScore != 55 {
		t.Fatalf("expected best score 55, got %d", bestScore)
	}
}
