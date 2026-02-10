//go:build postgres

package postgres

import (
	"os"
	"testing"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func newPostgresDB(t *testing.T) *database.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required for postgres smoke tests")
	}

	logger := zap.NewNop()
	db, err := database.NewWithConfig("postgres", "", dsn, logger)
	if err != nil {
		t.Fatalf("failed to connect/migrate postgres: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestPostgresSmoke_OTPAndPendingCards(t *testing.T) {
	db := newPostgresDB(t)
	conn := db.GetConnection()
	logger := zap.NewNop()

	userRepo := repository.NewUserRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingRepo := repository.NewTrainingCardRepository(conn, logger)
	otpRepo := repository.NewWebOTPRepository(conn, logger)

	user, err := userRepo.GetOrCreateUser(900000001)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}
	if user == nil || user.ID == 0 {
		t.Fatalf("unexpected user: %+v", user)
	}

	_, otp, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP failed: %v", err)
	}
	if otp == nil || otp.ID == 0 {
		t.Fatalf("unexpected otp: %+v", otp)
	}

	wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:       "postgres-smoke-pending",
		Definition: "pending",
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma failed: %v", err)
	}
	if wordID == 0 {
		t.Fatalf("unexpected word ID: %d", wordID)
	}

	pending, err := trainingRepo.GetWordCardsWithoutTrainingCards(10)
	if err != nil {
		t.Fatalf("GetWordCardsWithoutTrainingCards failed: %v", err)
	}
	if len(pending) == 0 {
		t.Fatal("expected at least one pending word card")
	}
}

func TestPostgresSmoke_CreateFlowWithoutLastInsertID(t *testing.T) {
	db := newPostgresDB(t)
	conn := db.GetConnection()
	logger := zap.NewNop()

	userRepo := repository.NewUserRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	sessionRepo := repository.NewSessionRepository(conn, logger)
	webSessionRepo := repository.NewWebSessionRepository(conn, logger)
	nudgeRepo := repository.NewNudgeRepository(conn, logger)
	accessRepo := repository.NewUserAccessCategoryRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	grammarRepo := repository.NewGrammarAttemptRepository(conn, logger)

	user, err := userRepo.GetOrCreateUser(900000002)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:       "postgres-smoke-flow",
		Definition: "flow",
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma failed: %v", err)
	}

	tcID, err := trainingRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordID,
		WordEN:     "postgres-smoke-flow",
		SenseIndex: 0,
		WordRU:     "поток",
		MeaningEN:  "flow",
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard failed: %v", err)
	}

	ucID, err := userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: tcID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             2.5,
	})
	if err != nil {
		t.Fatalf("CreateUserCard failed: %v", err)
	}

	sessID, err := sessionRepo.CreateSession(&models.TrainingSession{
		UserID:       user.ID,
		Source:       models.SourceManual,
		PlannedCount: 1,
		DoneCount:    0,
	})
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	now := time.Now().UTC()
	reviewID, err := sessionRepo.CreateReviewEvent(&models.ReviewEvent{
		SessionID:   &sessID,
		UserID:      user.ID,
		UserCardID:  ucID,
		Direction:   models.DirectionENtoRU,
		ShownAt:     now,
		AnsweredAt:  &now,
		OptionCount: 4,
		OptionsJSON: "[]",
		IsCorrect:   true,
		Quality:     4,
	})
	if err != nil {
		t.Fatalf("CreateReviewEvent failed: %v", err)
	}
	if reviewID == 0 {
		t.Fatal("expected review event id")
	}

	err = webSessionRepo.CreateSession(&repository.WebSession{
		UserID:    user.ID,
		Token:     "postgres-smoke-token",
		ExpiresAt: now.Add(30 * time.Minute),
	})
	if err != nil {
		t.Fatalf("Create web session failed: %v", err)
	}

	_, err = nudgeRepo.CreateNudge(&models.TrainingNudge{
		UserID:         user.ID,
		LocalDate:      "2099-01-01",
		DueCountAtSend: 1,
	})
	if err != nil {
		t.Fatalf("CreateNudge failed: %v", err)
	}

	catID, err := accessRepo.CreateCategory(&models.UserAccessCategory{Name: "postgres-smoke-access"})
	if err != nil || catID == 0 {
		t.Fatalf("Create access category failed: %v, id=%d", err, catID)
	}

	wsCatID, err := wordSetCategoryRepo.CreateCategory(&models.WordSetCategory{Name: "postgres-smoke-wordsets"})
	if err != nil || wsCatID == 0 {
		t.Fatalf("Create word set category failed: %v, id=%d", err, wsCatID)
	}

	wsID, err := wordSetRepo.CreateWordSet(&models.WordSet{
		CategoryID: &wsCatID,
		Title:      "postgres-smoke-set",
		SortOrder:  1,
	})
	if err != nil || wsID == 0 {
		t.Fatalf("Create word set failed: %v, id=%d", err, wsID)
	}

	attemptID, err := grammarRepo.CreateAttempt(&repository.TestAttempt{
		UserID:         user.ID,
		ScopeType:      "chapter",
		ScopeID:        "smoke.chapter",
		StartedAt:      now,
		Score:          60,
		Passed:         true,
		TotalQuestions: 1,
		AnswersJSON:    "{}",
		ResultsJSON:    "{}",
	})
	if err != nil || attemptID == 0 {
		t.Fatalf("CreateAttempt failed: %v, id=%d", err, attemptID)
	}
}

func TestPostgresSmoke_GrammarProgressUpsert(t *testing.T) {
	db := newPostgresDB(t)
	conn := db.GetConnection()
	logger := zap.NewNop()

	userRepo := repository.NewUserRepository(conn, logger)
	grammarRepo := repository.NewGrammarAttemptRepository(conn, logger)

	user, err := userRepo.GetOrCreateUser(900000003)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	chapterID := "smoke.chapter.progress"
	if err := grammarRepo.UpdateProgress(user.ID, chapterID, 40, false); err != nil {
		t.Fatalf("UpdateProgress first attempt failed: %v", err)
	}
	if err := grammarRepo.UpdateProgress(user.ID, chapterID, 80, true); err != nil {
		t.Fatalf("UpdateProgress second attempt failed: %v", err)
	}

	progress, err := grammarRepo.GetChapterProgress(user.ID, chapterID)
	if err != nil {
		t.Fatalf("GetChapterProgress failed: %v", err)
	}
	if progress.BestScore != 80 {
		t.Fatalf("expected best score 80, got %d", progress.BestScore)
	}
	if !progress.Passed {
		t.Fatal("expected chapter to be marked as passed")
	}
}

func TestPostgresSmoke_UserWordKnowledgeMarkKnownUpsert(t *testing.T) {
	db := newPostgresDB(t)
	conn := db.GetConnection()
	logger := zap.NewNop()

	userRepo := repository.NewUserRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	user, err := userRepo.GetOrCreateUser(900000004)
	if err != nil {
		t.Fatalf("GetOrCreateUser failed: %v", err)
	}

	wordID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:       "postgres-smoke-known",
		Definition: "known",
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma failed: %v", err)
	}

	if err := uwkRepo.MarkKnown(user.ID, wordID); err != nil {
		t.Fatalf("MarkKnown first call failed: %v", err)
	}
	if err := uwkRepo.MarkKnown(user.ID, wordID); err != nil {
		t.Fatalf("MarkKnown second call failed: %v", err)
	}

	isKnown, err := uwkRepo.IsKnown(user.ID, wordID)
	if err != nil {
		t.Fatalf("IsKnown failed: %v", err)
	}
	if !isKnown {
		t.Fatal("expected word to remain known after second MarkKnown")
	}
}
