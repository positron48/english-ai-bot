package service

import (
	"context"
	"fmt"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWordSetServiceMoreTest(t *testing.T) (*WordSetService, *database.DB, *repository.UserRepository, func()) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)

	service := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, config.DefaultLearningConfig(), "gpt-4", logger)

	cleanup := func() {} // shared db, do not close

	return service, db, userRepo, cleanup
}

func TestMarkKnown(t *testing.T) {
	service, db, userRepo, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	user, _ := userRepo.GetOrCreateUser(12345)
	userID := user.ID

	// Create a word card
	wordRepo := repository.NewWordRepository(db.GetConnection(), service.logger)
	wordRepo.SaveWordCard("testword", "test definition")
	wordCard, _ := wordRepo.GetWordCard("testword")

	// Mark as known
	err := service.MarkKnown(userID, wordCard.ID)
	if err != nil {
		t.Fatalf("Failed to mark as known: %v", err)
	}

	// Verify it's marked as known
	uwkRepo := repository.NewUserWordKnowledgeRepository(db.GetConnection(), service.logger)
	isKnown, err := uwkRepo.IsKnown(userID, wordCard.ID)
	if err != nil {
		t.Fatalf("Failed to check if known: %v", err)
	}
	if !isKnown {
		t.Error("Expected word to be marked as known")
	}
}

func TestMarkKnown_WithUserCards(t *testing.T) {
	service, db, userRepo, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	user, _ := userRepo.GetOrCreateUser(12345)
	userID := user.ID

	// Create a word card
	wordRepo := repository.NewWordRepository(db.GetConnection(), service.logger)
	wordRepo.SaveWordCard("testword", "test definition")
	wordCard, _ := wordRepo.GetWordCard("testword")

	// Create a user card first
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), service.logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), service.logger)
	tc := &models.TrainingCard{
		WordCardID:    wordCard.ID,
		WordEN:        "testword",
		SenseIndex:    0,
		WordRU:        "тест",
		MeaningEN:     "test meaning",
		DistractorsRU: `[]`,
		DistractorsEN: `[]`,
	}
	tcID, _ := tcRepo.CreateTrainingCard(tc)

	uc := &models.UserCard{
		UserID:         userID,
		TrainingCardID: tcID,
		State:          models.StateNew,
	}
	ucRepo.CreateUserCard(uc)

	// Mark as known (should delete user cards)
	err := service.MarkKnown(userID, wordCard.ID)
	if err != nil {
		t.Fatalf("Failed to mark as known: %v", err)
	}

	// Verify user cards are deleted by checking count via SQL
	var count int
	err = db.GetConnection().QueryRow(
		`SELECT COUNT(*) FROM user_cards uc 
		 INNER JOIN training_cards tc ON uc.training_card_id = tc.id 
		 WHERE uc.user_id = ? AND tc.word_card_id = ?`,
		userID, wordCard.ID,
	).Scan(&count)
	if err == nil && count > 0 {
		t.Error("Expected user cards to be deleted after marking as known")
	}
}

// mockUserCardRepoForWordSet returns error from DeleteUserCardsByWordCardIDForUser to trigger MarkKnown Warn path.
type mockUserCardRepoForWordSet struct {
	deleteFunc func(userID, wordCardID int64) (int64, error)
	createFunc func(card *models.UserCard) (int64, error)
}

func (m *mockUserCardRepoForWordSet) DeleteUserCardsByWordCardIDForUser(userID, wordCardID int64) (int64, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(userID, wordCardID)
	}
	return 0, nil
}

func (m *mockUserCardRepoForWordSet) CreateUserCard(card *models.UserCard) (int64, error) {
	if m.createFunc != nil {
		return m.createFunc(card)
	}
	return 0, nil
}

// TestMarkKnown_DeleteUserCardsFails verifies that when DeleteUserCardsByWordCardIDForUser fails, MarkKnown still returns nil (Warn path).
func TestMarkKnown_DeleteUserCardsFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(db.GetConnection(), logger)
	userRepo := repository.NewUserRepository(db.GetConnection(), logger)
	user, _ := userRepo.GetOrCreateUser(99999)
	wordRepo.SaveWordCard("markknownword", "def")
	wordCard, _ := wordRepo.GetWordCard("markknownword")

	mockUC := &mockUserCardRepoForWordSet{
		deleteFunc: func(_, _ int64) (int64, error) {
			return 0, fmt.Errorf("mock delete error")
		},
	}
	svc := NewWordSetServiceWithMastering(
		wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, mockUC, uwkRepo, nil, nil, config.DefaultLearningConfig(), "gpt-4", logger,
	)
	err := svc.MarkKnown(user.ID, wordCard.ID)
	if err != nil {
		t.Errorf("MarkKnown should return nil when only delete fails (Warn path), got: %v", err)
	}
}

// TestEnsureUserCardsForWord_NoTrainingCards covers len(trainingCards)==0 -> nil.
func TestEnsureUserCardsForWord_NoTrainingCards(t *testing.T) {
	service, db, userRepo, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	user, _ := userRepo.GetOrCreateUser(11111)
	wordRepo := repository.NewWordRepository(db.GetConnection(), service.logger)
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "nocard", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}

	err = service.EnsureUserCardsForWord(user.ID, cardID)
	if err != nil {
		t.Fatalf("EnsureUserCardsForWord: %v", err)
	}
}

// TestEnsureUserCardsForWord_SecondCallIdempotent covers CreateUserCard duplicate (warn path); second call still succeeds.
func TestEnsureUserCardsForWord_SecondCallIdempotent(t *testing.T) {
	service, db, userRepo, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	user, _ := userRepo.GetOrCreateUser(22222)
	wordRepo := repository.NewWordRepository(db.GetConnection(), service.logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), service.logger)
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "idem", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_, err = tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: cardID,
		WordEN:     "idem",
		WordRU:     "идем",
		MeaningEN:  "go",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	err = service.EnsureUserCardsForWord(user.ID, cardID)
	if err != nil {
		t.Fatalf("EnsureUserCardsForWord first call: %v", err)
	}
	err = service.EnsureUserCardsForWord(user.ID, cardID)
	if err != nil {
		t.Fatalf("EnsureUserCardsForWord second call (idempotent): %v", err)
	}
}

// TestEnsureUserCardsForWord_WithMasteringRepo covers createdCount > 0 and userWordMasteringRepo.Upsert.
func TestEnsureUserCardsForWord_WithMasteringRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	conn := db.GetConnection()

	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	tcRepo := repository.NewTrainingCardRepository(conn, logger)
	ucRepo := repository.NewUserCardRepository(conn, logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(conn, logger)
	masteringRepo := repository.NewUserWordMasteringRepository(conn, logger)
	userRepo := repository.NewUserRepository(conn, logger)

	service := NewWordSetServiceWithMastering(
		wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, masteringRepo, nil, config.DefaultLearningConfig(), "", logger,
	)
	user, _ := userRepo.GetOrCreateUser(33333)
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "masterw", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_, err = tcRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: cardID,
		WordEN:     "masterw",
		WordRU:     "мастер",
		MeaningEN:  "master",
		SenseIndex: 0,
	})
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	err = service.EnsureUserCardsForWord(user.ID, cardID)
	if err != nil {
		t.Fatalf("EnsureUserCardsForWord: %v", err)
	}
	score, err := masteringRepo.GetScore(user.ID, cardID)
	if err != nil {
		t.Fatalf("GetScore: %v", err)
	}
	if score != 0 {
		t.Errorf("expected initial mastering score 0, got %d", score)
	}
}

func TestProcessWordSetItems(t *testing.T) {
	service, db, _, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	// Create a word set
	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), service.logger)
	desc := "Test Description"
	wordSet := &models.WordSet{
		Title:       "Test Set",
		Description: &desc,
	}
	wordSetID, _ := wordSetRepo.CreateWordSet(wordSet)

	ctx := context.Background()
	wordsStr := "apple, banana, cherry"

	err := service.ProcessWordSetItems(ctx, wordSetID, wordsStr)
	if err != nil {
		t.Fatalf("Failed to process word set items: %v", err)
	}

	// Verify word set items were created by checking via GetWordSetWords
	// (requires a user, so create one)
	userRepo := repository.NewUserRepository(db.GetConnection(), service.logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	words, err := wordSetRepo.GetWordSetWords(wordSetID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get word set words: %v", err)
	}
	if len(words) != 3 {
		t.Errorf("Expected 3 word set words, got %d", len(words))
	}
}

func TestProcessWordSetItems_EmptyWords(t *testing.T) {
	service, db, _, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), service.logger)
	desc := "Test Description"
	wordSet := &models.WordSet{
		Title:       "Test Set",
		Description: &desc,
	}
	wordSetID, _ := wordSetRepo.CreateWordSet(wordSet)

	ctx := context.Background()
	wordsStr := ""

	err := service.ProcessWordSetItems(ctx, wordSetID, wordsStr)
	if err != nil {
		t.Fatalf("Failed to process empty word set items: %v", err)
	}

	// Should have no items
	userRepo := repository.NewUserRepository(db.GetConnection(), service.logger)
	user, _ := userRepo.GetOrCreateUser(12345)
	words, _ := wordSetRepo.GetWordSetWords(wordSetID, user.ID)
	if len(words) != 0 {
		t.Errorf("Expected 0 word set words, got %d", len(words))
	}
}

func TestProcessWordSetItems_WithDuplicates(t *testing.T) {
	service, db, _, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), service.logger)
	desc := "Test Description"
	wordSet := &models.WordSet{
		Title:       "Test Set",
		Description: &desc,
	}
	wordSetID, _ := wordSetRepo.CreateWordSet(wordSet)

	ctx := context.Background()
	wordsStr := "apple, apple, banana, banana, cherry"

	err := service.ProcessWordSetItems(ctx, wordSetID, wordsStr)
	if err != nil {
		t.Fatalf("Failed to process word set items with duplicates: %v", err)
	}

	// Should have only 3 unique items
	userRepo := repository.NewUserRepository(db.GetConnection(), service.logger)
	user, _ := userRepo.GetOrCreateUser(12345)
	words, err := wordSetRepo.GetWordSetWords(wordSetID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get word set words: %v", err)
	}
	if len(words) != 3 {
		t.Errorf("Expected 3 unique word set words, got %d", len(words))
	}
}

func TestProcessWordSetItems_WithWhitespace(t *testing.T) {
	service, db, _, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), service.logger)
	desc := "Test Description"
	wordSet := &models.WordSet{
		Title:       "Test Set",
		Description: &desc,
	}
	wordSetID, _ := wordSetRepo.CreateWordSet(wordSet)

	ctx := context.Background()
	wordsStr := " apple ,  banana , cherry "

	err := service.ProcessWordSetItems(ctx, wordSetID, wordsStr)
	if err != nil {
		t.Fatalf("Failed to process word set items with whitespace: %v", err)
	}

	// Should have 3 items (whitespace should be trimmed)
	userRepo := repository.NewUserRepository(db.GetConnection(), service.logger)
	user, _ := userRepo.GetOrCreateUser(12345)
	words, err := wordSetRepo.GetWordSetWords(wordSetID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get word set words: %v", err)
	}
	if len(words) != 3 {
		t.Errorf("Expected 3 word set words, got %d", len(words))
	}
}
