package repository

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

const ucInvalidDSN = "postgres://x:x@invalid.invalid:1/db?connect_timeout=1"

// invalidDB returns a *sql.DB with an invalid DSN so any operation returns an error.
func invalidDB(t *testing.T) *sql.DB {
	t.Helper()
	// Ensure postgres_compat driver is registered by initializing the shared test DB first.
	testutil.SetupTestDB(t)
	db, err := sql.Open("postgres_compat", ucInvalidDSN)
	if err != nil {
		t.Skip("postgres_compat driver not registered or open failed:", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUserCardRepository_CreateUserCard_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	now := time.Now()
	card := &models.UserCard{
		UserID:         1,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	_, err := repo.CreateUserCard(card)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_GetDueCards_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.GetDueCards(1, time.Now(), 10)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_GetDueCount_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.GetDueCount(1, time.Now())
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_CountNewCardsSince_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.CountNewCardsSince(1, time.Now())
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_GetNewCards_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.GetNewCards(1, 10)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_GetWordMasteringStats_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.GetWordMasteringStats(1, 1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_GetWordsEligibleForSpellByMastery_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.GetWordsEligibleForSpellByMastery(1, 50, 10)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_UpdateUserCard_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	card := &models.UserCard{ID: 1, State: models.StateReview}
	err := repo.UpdateUserCard(card)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_DeleteOrphanedUserCards_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.DeleteOrphanedUserCards()
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_DeleteUserCardsByWordENForUser_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.DeleteUserCardsByWordENForUser(1, "word")
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_DeleteUserCardsByWordCardIDForUser_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.DeleteUserCardsByWordCardIDForUser(1, 1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_ListOrphanedUserCards_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.ListOrphanedUserCards(10, 0)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_CountOrphanedUserCards_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.CountOrphanedUserCards()
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_ListUserCardsWithOrphanedTrainingCards_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.ListUserCardsWithOrphanedTrainingCards(10, 0)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_CountUserCardsWithOrphanedTrainingCards_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.CountUserCardsWithOrphanedTrainingCards()
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_DeleteUserCard_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	err := repo.DeleteUserCard(1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_GetUserIDsByWordCardID_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.GetUserIDsByWordCardID(1)
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

func TestUserCardRepository_GetUpcomingCardsByDate_DBError(t *testing.T) {
	db := invalidDB(t)
	repo := NewUserCardRepository(db, zap.NewNop())

	_, err := repo.GetUpcomingCardsByDate(1, time.Now())
	if err == nil {
		t.Fatal("expected error when DB is invalid")
	}
}

// TestUserCardRepository_scanUserCards_ScanError covers the scan error path in scanUserCards.
func TestUserCardRepository_scanUserCards_ScanError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)

	// Use a query that returns wrong number of columns to trigger scan error
	rows, err := db.Query("SELECT 1, 2, 3")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	_, err = repo.scanUserCards(rows)
	if err == nil {
		t.Fatal("expected scan error with wrong columns")
	}
}

// TestUserCardRepository_scanUserCard_ScanError covers the scan error path in scanUserCard.
func TestUserCardRepository_scanUserCard_ScanError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)

	// Use a query that returns wrong number of columns to trigger scan error
	row := db.QueryRow("SELECT 1, 2, 3")
	_, err := repo.scanUserCard(row)
	if err == nil {
		t.Fatal("expected scan error with wrong columns")
	}
}

// TestUserCardRepository_GetUserIDsByWordCardID_Empty verifies empty result.
func TestUserCardRepository_GetUserIDsByWordCardID_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)

	userIDs, err := repo.GetUserIDsByWordCardID(99999)
	if err != nil {
		t.Fatalf("GetUserIDsByWordCardID() error = %v", err)
	}
	if len(userIDs) != 0 {
		t.Errorf("expected empty result, got %v", userIDs)
	}
}

// TestUserCardRepository_GetWordsEligibleForSpellByMastery_WithDisplayWord covers the scan path
// where displayWord and wordRU are valid (non-null).
func TestUserCardRepository_GetWordsEligibleForSpellByMastery_WithDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(99001)
	displayEN := "to display"
	wordCard := &models.WordCard{Word: "displayword", Definition: "def", DisplayEN: &displayEN}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "verb"
	displayWord := "display"
	wordRU := "отображать"
	tc := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "displayword",
		SenseIndex:  0,
		WordRU:      wordRU,
		MeaningEN:   "to display",
		POS:         &pos,
		DisplayWord: &displayWord,
	}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)
	now := time.Now()
	_, _ = repo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: tcID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.5,
		NextDueAt:      &now,
	})
	// Set mastering_score >= 50
	_, _ = db.Exec("INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score) VALUES ($1, $2, 75) ON CONFLICT (user_id, word_card_id) DO UPDATE SET mastering_score = 75", user.ID, wordCardID)

	words, err := repo.GetWordsEligibleForSpellByMastery(user.ID, 50, 10)
	if err != nil {
		t.Fatalf("GetWordsEligibleForSpellByMastery() error = %v", err)
	}
	if len(words) == 0 {
		t.Fatal("expected at least 1 word")
	}
	// Verify DisplayWord and WordRU are populated
	found := false
	for _, w := range words {
		if w.WordCardID == wordCardID {
			found = true
			if w.DisplayWord == "" {
				t.Error("expected non-empty DisplayWord")
			}
			if w.WordRU == "" {
				t.Error("expected non-empty WordRU")
			}
		}
	}
	if !found {
		t.Error("expected word with wordCardID in results")
	}
}

// TestUserCardRepository_ListOrphanedUserCards_WithTelegramUsername covers the branch
// where telegramUsername is valid (non-null).
func TestUserCardRepository_ListOrphanedUserCards_WithTelegramUsername(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create user with a telegram username
	user, _ := userRepo.GetOrCreateUser(99002)
	// Update user to have a telegram username
	_, _ = db.Exec("UPDATE users SET telegram_username = 'testuser99002' WHERE id = $1", user.ID)

	wordCard := &models.WordCard{Word: "orphusername", Definition: "def"}
	wcID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "orphusername"
	tc := &models.TrainingCard{WordCardID: wcID, WordEN: "orphusername", SenseIndex: 0, WordRU: "сирота", MeaningEN: "orphan", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)
	now := time.Now()
	_, _ = repo.CreateUserCard(&models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateNew, EF: 2.5, NextDueAt: &now})

	// Orphan the user card by deleting training card (bypass FK)
	_, _ = db.Exec("SET session_replication_role = replica")
	_, _ = db.Exec("DELETE FROM training_cards WHERE id = $1", tcID)
	_, _ = db.Exec("SET session_replication_role = DEFAULT")

	orphaned, err := repo.ListOrphanedUserCards(10, 0)
	if err != nil {
		t.Fatalf("ListOrphanedUserCards() error = %v", err)
	}
	// Check if any item has a non-nil TelegramUsername (covers the telegramUsername.Valid branch)
	for _, item := range orphaned {
		if item.TelegramUsername != nil {
			return // covered the branch
		}
	}
}

// TestUserCardRepository_scanUserCard_WithLastReviewAtAndQuality covers the lastReviewAt.Valid
// and lastQuality.Valid branches in scanUserCard.
func TestUserCardRepository_scanUserCard_WithLastReviewAtAndQuality(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(99100)
	wordCard := &models.WordCard{Word: "reviewword", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "verb"
	displayWord := "review"
	tc := &models.TrainingCard{WordCardID: wordCardID, WordEN: "reviewword", SenseIndex: 0, WordRU: "обзор", MeaningEN: "review", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)

	now := time.Now()
	quality := 4
	uc := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: tcID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.5,
		Reps:           3,
		LastReviewAt:   &now,
		LastQuality:    &quality,
		NextDueAt:      &now,
	}
	ucID, err := repo.CreateUserCard(uc)
	if err != nil {
		t.Fatalf("CreateUserCard: %v", err)
	}

	// Update to set last_review_at and last_quality
	uc.ID = ucID
	if err := repo.UpdateUserCard(uc); err != nil {
		t.Fatalf("UpdateUserCard: %v", err)
	}

	// GetUserCard triggers scanUserCard with lastReviewAt.Valid and lastQuality.Valid
	card, err := repo.GetUserCard(ucID)
	if err != nil {
		t.Fatalf("GetUserCard() error = %v", err)
	}
	if card == nil {
		t.Fatal("expected non-nil card")
	}
	if card.LastReviewAt == nil {
		t.Error("expected LastReviewAt to be set")
	}
	if card.LastQuality == nil {
		t.Error("expected LastQuality to be set")
	}
}

// TestUserCardRepository_scanUserCards_WithLastReviewAtAndQuality covers the lastReviewAt.Valid
// and lastQuality.Valid branches in scanUserCards.
func TestUserCardRepository_scanUserCards_WithLastReviewAtAndQuality(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(99101)
	wordCard := &models.WordCard{Word: "reviewword2", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "verb"
	displayWord := "review2"
	tc := &models.TrainingCard{WordCardID: wordCardID, WordEN: "reviewword2", SenseIndex: 0, WordRU: "обзор2", MeaningEN: "review2", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	quality := 3
	uc := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: tcID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.5,
		Reps:           2,
		LastReviewAt:   &past,
		LastQuality:    &quality,
		NextDueAt:      &past,
	}
	ucID, err := repo.CreateUserCard(uc)
	if err != nil {
		t.Fatalf("CreateUserCard: %v", err)
	}
	uc.ID = ucID
	if err := repo.UpdateUserCard(uc); err != nil {
		t.Fatalf("UpdateUserCard: %v", err)
	}

	// GetDueCards triggers scanUserCards with lastReviewAt.Valid and lastQuality.Valid
	cards, err := repo.GetDueCards(user.ID, now, 10)
	if err != nil {
		t.Fatalf("GetDueCards() error = %v", err)
	}
	found := false
	for _, c := range cards {
		if c.ID == ucID {
			found = true
			if c.LastReviewAt == nil {
				t.Error("expected LastReviewAt to be set")
			}
			if c.LastQuality == nil {
				t.Error("expected LastQuality to be set")
			}
		}
	}
	if !found {
		t.Error("expected card in due cards")
	}
}

// TestUserCardRepository_ListUserCardsWithOrphanedTrainingCards_WithTelegramUsername covers
// the telegramUsername.Valid branch in ListUserCardsWithOrphanedTrainingCards.
func TestUserCardRepository_ListUserCardsWithOrphanedTrainingCards_WithTelegramUsername(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(99003)
	_, _ = db.Exec("UPDATE users SET telegram_username = 'testuser99003' WHERE id = $1", user.ID)

	wordCard := &models.WordCard{Word: "orphtcusername", Definition: "def"}
	wcID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "orphtcusername"
	tc := &models.TrainingCard{WordCardID: wcID, WordEN: "orphtcusername", SenseIndex: 0, WordRU: "сирота", MeaningEN: "orphan", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)
	now := time.Now()
	_, _ = repo.CreateUserCard(&models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateNew, EF: 2.5, NextDueAt: &now})

	// Delete word_card to orphan training_card (bypass FK)
	_, _ = db.Exec("SET session_replication_role = replica")
	_, _ = db.Exec("DELETE FROM word_cards WHERE id = $1", wcID)
	_, _ = db.Exec("SET session_replication_role = DEFAULT")

	orphaned, err := repo.ListUserCardsWithOrphanedTrainingCards(10, 0)
	if err != nil {
		t.Fatalf("ListUserCardsWithOrphanedTrainingCards() error = %v", err)
	}
	for _, item := range orphaned {
		if item.TelegramUsername != nil {
			return // covered the branch
		}
	}
}

// TestUserCardRepository_CreateUserCard_InsertError covers the InsertAndReturnID error path
// (when the check passes but insert fails).
func TestUserCardRepository_CreateUserCard_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// GetUserCardByTrainingCard: no existing card (ErrNoRows)
	mock.ExpectQuery("SELECT .+ FROM user_cards WHERE user_id").
		WillReturnError(sql.ErrNoRows)
	// InsertAndReturnID: INSERT fails
	mock.ExpectQuery("INSERT INTO user_cards").
		WillReturnError(fmt.Errorf("insert error"))

	repo := NewUserCardRepository(db, zap.NewNop())
	now := time.Now()
	card := &models.UserCard{
		UserID:         1,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(card)
	if err == nil {
		t.Fatal("expected error when insert fails")
	}
}

// TestUserCardRepository_GetWordMasteringStats_ErrNoRows covers the sql.ErrNoRows branch.
func TestUserCardRepository_GetWordMasteringStats_ErrNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	repo := NewUserCardRepository(db, zap.NewNop())
	result, err := repo.GetWordMasteringStats(1, 1)
	if err != nil {
		t.Fatalf("GetWordMasteringStats() unexpected error = %v", err)
	}
	if result != nil {
		t.Error("expected nil result for ErrNoRows")
	}
}

// TestUserCardRepository_GetWordsEligibleForSpellByMastery_ScanError covers the scan error path.
func TestUserCardRepository_GetWordsEligibleForSpellByMastery_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row that will fail to scan (wrong types)
	rows := sqlmock.NewRows([]string{"word_card_id", "display_word", "word_ru"}).
		AddRow("not-an-int", nil, nil)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.GetWordsEligibleForSpellByMastery(1, 50, 10)
	if err == nil {
		t.Fatal("expected scan error")
	}
}

// TestUserCardRepository_ListOrphanedUserCards_ScanError covers the scan error path.
func TestUserCardRepository_ListOrphanedUserCards_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row that will fail to scan
	rows := sqlmock.NewRows([]string{"user_card_id", "user_id", "telegram_id", "telegram_username",
		"training_card_id", "direction", "state", "reps", "created_at", "review_events_count"}).
		AddRow("not-an-int", 1, 1, nil, 1, "en_ru", "new", 0, "2024-01-01 00:00:00", 0)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.ListOrphanedUserCards(10, 0)
	if err == nil {
		t.Fatal("expected scan error")
	}
}

// TestUserCardRepository_ListUserCardsWithOrphanedTrainingCards_ScanError covers the scan error path.
func TestUserCardRepository_ListUserCardsWithOrphanedTrainingCards_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_card_id", "user_id", "telegram_id", "telegram_username",
		"training_card_id", "direction", "state", "reps", "created_at", "review_events_count"}).
		AddRow("not-an-int", 1, 1, nil, 1, "en_ru", "new", 0, "2024-01-01 00:00:00", 0)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.ListUserCardsWithOrphanedTrainingCards(10, 0)
	if err == nil {
		t.Fatal("expected scan error")
	}
}

// TestUserCardRepository_GetUserIDsByWordCardID_ScanError covers the scan error path.
func TestUserCardRepository_GetUserIDsByWordCardID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id"}).AddRow("not-an-int")
	mock.ExpectQuery("SELECT DISTINCT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.GetUserIDsByWordCardID(1)
	if err == nil {
		t.Fatal("expected scan error")
	}
}

// TestUserCardRepository_DeleteOrphanedUserCards_RowsAffectedError covers the RowsAffected error paths.
func TestUserCardRepository_DeleteOrphanedUserCards_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// First DELETE succeeds but RowsAffected fails
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.DeleteOrphanedUserCards()
	if err == nil {
		t.Fatal("expected error from RowsAffected")
	}
}

// TestUserCardRepository_DeleteOrphanedUserCards_SecondQueryError covers the second DELETE error path.
func TestUserCardRepository_DeleteOrphanedUserCards_SecondQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// First DELETE succeeds with RowsAffected
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	// Second DELETE fails
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnError(fmt.Errorf("second delete error"))

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.DeleteOrphanedUserCards()
	if err == nil {
		t.Fatal("expected error from second DELETE")
	}
}

// TestUserCardRepository_DeleteOrphanedUserCards_SecondRowsAffectedError covers the second RowsAffected error.
func TestUserCardRepository_DeleteOrphanedUserCards_SecondRowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// First DELETE succeeds
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewResult(1, 0))
	// Second DELETE succeeds but RowsAffected fails
	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error 2")))

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.DeleteOrphanedUserCards()
	if err == nil {
		t.Fatal("expected error from second RowsAffected")
	}
}

// TestUserCardRepository_DeleteUserCardsByWordENForUser_RowsAffectedError covers RowsAffected error.
func TestUserCardRepository_DeleteUserCardsByWordENForUser_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.DeleteUserCardsByWordENForUser(1, "word")
	if err == nil {
		t.Fatal("expected error from RowsAffected")
	}
}

// TestUserCardRepository_DeleteUserCardsByWordCardIDForUser_RowsAffectedError covers RowsAffected error.
func TestUserCardRepository_DeleteUserCardsByWordCardIDForUser_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.DeleteUserCardsByWordCardIDForUser(1, 1)
	if err == nil {
		t.Fatal("expected error from RowsAffected")
	}
}

// TestUserCardRepository_DeleteUserCard_RowsAffectedError covers the RowsAffected error path.
func TestUserCardRepository_DeleteUserCard_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM user_cards").
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	repo := NewUserCardRepository(db, zap.NewNop())
	err = repo.DeleteUserCard(1)
	if err == nil {
		t.Fatal("expected error from RowsAffected")
	}
}

// TestUserCardRepository_GetUpcomingCardsByDate_RowsErr covers the rows.Err() path.
func TestUserCardRepository_GetUpcomingCardsByDate_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"next_due_at"}).CloseError(fmt.Errorf("rows error"))
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.GetUpcomingCardsByDate(1, time.Now())
	if err == nil {
		t.Fatal("expected rows.Err() error")
	}
}

// TestUserCardRepository_GetUserIDsByWordCardID_RowsErr covers the rows.Err() path.
func TestUserCardRepository_GetUserIDsByWordCardID_RowsErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id"}).CloseError(fmt.Errorf("rows error"))
	mock.ExpectQuery("SELECT DISTINCT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.GetUserIDsByWordCardID(1)
	if err == nil {
		t.Fatal("expected rows.Err() error")
	}
}

// TestUserCardRepository_GetUpcomingCardsByDate_ScanError covers the scan error path.
func TestUserCardRepository_GetUpcomingCardsByDate_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Return a row with a value that can't be scanned into sql.NullString
	// Actually NullString can scan anything, so let's use wrong column count
	rows := sqlmock.NewRows([]string{"next_due_at", "extra_col"}).
		AddRow("2024-01-01T00:00:00Z", "extra")
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	_, err = repo.GetUpcomingCardsByDate(1, time.Now())
	if err == nil {
		t.Fatal("expected scan error with extra column")
	}
}

// TestUserCardRepository_GetUpcomingCardsByDate_DateParseFallbacks covers the date parse
// fallback branches (RFC3339, standard format, format with TZ offset).
func TestUserCardRepository_GetUpcomingCardsByDate_DateParseFallbacks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Return rows with different date formats to cover fallback branches
	rows := sqlmock.NewRows([]string{"next_due_at"}).
		AddRow("2024-01-02T00:00:00Z").          // RFC3339 (no nanoseconds) - triggers second parse
		AddRow("2024-01-03 00:00:00").            // standard format - triggers third parse
		AddRow("2024-01-04 00:00:00+00:00").      // format with TZ offset - triggers fourth parse
		AddRow("invalid-date-format")             // unparseable - triggers warn+continue
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	repo := NewUserCardRepository(db, zap.NewNop())
	result, err := repo.GetUpcomingCardsByDate(1, startDate)
	if err != nil {
		t.Fatalf("GetUpcomingCardsByDate() error = %v", err)
	}
	_ = result
}
