package service

import (
	"context"
	"testing"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func setupWordSetServiceMoreTest(t *testing.T) (*WordSetService, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	wordSetRepo := repository.NewWordSetRepository(db.GetConnection(), logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(db.GetConnection(), logger)
	wordRepo := repository.NewWordRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	uwkRepo := repository.NewUserWordKnowledgeRepository(db.GetConnection(), logger)

	service := NewWordSetService(wordSetRepo, wordSetCategoryRepo, wordRepo, tcRepo, ucRepo, uwkRepo, nil, "gpt-4", logger)

	cleanup := func() {
		db.Close()
	}

	return service, db, cleanup
}

func TestMarkKnown(t *testing.T) {
	service, db, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	// Create a word card
	wordRepo := repository.NewWordRepository(db.GetConnection(), service.logger)
	wordRepo.SaveWordCard("testword", "test definition")
	wordCard, _ := wordRepo.GetWordCard("testword")

	userID := int64(12345)

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
	service, db, cleanup := setupWordSetServiceMoreTest(t)
	defer cleanup()

	// Create a word card
	wordRepo := repository.NewWordRepository(db.GetConnection(), service.logger)
	wordRepo.SaveWordCard("testword", "test definition")
	wordCard, _ := wordRepo.GetWordCard("testword")

	userID := int64(12345)

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
		UserID:        userID,
		TrainingCardID: tcID,
		State:         models.StateNew,
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

func TestProcessWordSetItems(t *testing.T) {
	service, db, cleanup := setupWordSetServiceMoreTest(t)
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
	service, db, cleanup := setupWordSetServiceMoreTest(t)
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
	service, db, cleanup := setupWordSetServiceMoreTest(t)
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
	service, db, cleanup := setupWordSetServiceMoreTest(t)
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
