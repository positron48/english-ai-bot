package service

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/testutil"
)

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func setupWordServiceTestDB(t *testing.T) (*sql.DB, *repository.WordRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	wordRepo := repository.NewWordRepository(db, logger)

	return db, wordRepo
}

func setupWordServiceTestDBWithTraining(t *testing.T) (*sql.DB, *repository.WordRepository, *repository.TrainingCardRepository, *repository.UserCardRepository, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	wordRepo := repository.NewWordRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)

	return db, wordRepo, trainingCardRepo, userCardRepo, userRepo
}

func TestWordService_GetWordDefinition_FromDB(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, wordRepo := setupWordServiceTestDB(t)

	// Save a word to database
	err := wordRepo.SaveWordCard("testword", "a test word definition")
	if err != nil {
		t.Fatalf("Failed to save word card: %v", err)
	}

	service := NewWordService(wordRepo, nil, nil, (*ai.Service)(nil), logger)

	ctx := context.Background()
	definition, err := service.GetWordDefinition(ctx, 123, "testword")
	if err != nil {
		t.Fatalf("GetWordDefinition() error = %v", err)
	}
	// After refactoring, GetWordDefinition returns markdown rendered from structured data
	// The definition should contain the word and some markdown structure
	if definition == "" {
		t.Error("Expected non-empty definition (markdown)")
	}
	if !contains(definition, "testword") {
		t.Errorf("Expected definition to contain 'testword', got %q", definition)
	}
}

func TestWordService_GetWordDefinition_FromAI(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, wordRepo := setupWordServiceTestDB(t)

	service := NewWordService(wordRepo, nil, nil, (*ai.Service)(nil), logger)
	_ = service // Verify service is created
	// Note: Full test would require a real AI service
}

func TestWordService_GetWordCard_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, wordRepo := setupWordServiceTestDB(t)

	// Save a word to database
	err := wordRepo.SaveWordCard("getcard", "card definition")
	if err != nil {
		t.Fatalf("Failed to save word card: %v", err)
	}

	service := NewWordService(wordRepo, nil, nil, nil, logger)
	card, err := service.GetWordCard("getcard")
	if err != nil {
		t.Fatalf("GetWordCard() error = %v", err)
	}
	if card == nil {
		t.Fatal("GetWordCard() should not return nil")
	}
	if card.Word != "getcard" {
		t.Errorf("Expected word 'getcard', got %q", card.Word)
	}
	if card.Definition != "card definition" {
		t.Errorf("Expected definition 'card definition', got %q", card.Definition)
	}
}

func TestWordService_GetWordDefinition_CreatesUserCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, wordRepo, trainingCardRepo, userCardRepo, userRepo := setupWordServiceTestDBWithTraining(t)

	ctx := context.Background()
	// Create user (required for FK in user_cards, word_request_history)
	user, err := userRepo.GetOrCreateUser(123)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userID := user.ID

	// Create a word card
	wordCard := &models.WordCard{
		Word:         "testword",
		Definition:   "test definition",
		DefinitionRU: stringPtr("тестовое определение"),
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards for this word
	trainingCard1 := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "testword",
		SenseIndex: 0,
		WordRU:     "тестовое слово",
		MeaningEN:  "test meaning",
	}
	trainingCardID1, err := trainingCardRepo.CreateTrainingCard(trainingCard1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}

	trainingCard2 := &models.TrainingCard{
		WordCardID: wordCardID,
		WordEN:     "testword",
		SenseIndex: 1,
		WordRU:     "другое значение",
		MeaningEN:  "another meaning",
	}
	trainingCardID2, err := trainingCardRepo.CreateTrainingCard(trainingCard2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Create service with repositories
	service := NewWordService(wordRepo, trainingCardRepo, userCardRepo, nil, logger)

	// Request the word - this should create user_cards
	definition, err := service.GetWordDefinition(ctx, userID, "testword")
	if err != nil {
		t.Fatalf("GetWordDefinition() error = %v", err)
	}
	if definition == "" {
		t.Error("Expected non-empty definition")
	}

	// Verify that user_cards were created for both training cards and both directions
	ruEnCard1, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID1, models.DirectionRUtoEN)
	if err != nil {
		t.Fatalf("Failed to get ru_en card 1: %v", err)
	}
	if ruEnCard1 == nil {
		t.Error("Expected ru_en user card 1 to be created")
	}

	enRuCard1, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID1, models.DirectionENtoRU)
	if err != nil {
		t.Fatalf("Failed to get en_ru card 1: %v", err)
	}
	if enRuCard1 == nil {
		t.Error("Expected en_ru user card 1 to be created")
	}

	ruEnCard2, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID2, models.DirectionRUtoEN)
	if err != nil {
		t.Fatalf("Failed to get ru_en card 2: %v", err)
	}
	if ruEnCard2 == nil {
		t.Error("Expected ru_en user card 2 to be created")
	}

	enRuCard2, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID2, models.DirectionENtoRU)
	if err != nil {
		t.Fatalf("Failed to get en_ru card 2: %v", err)
	}
	if enRuCard2 == nil {
		t.Error("Expected en_ru user card 2 to be created")
	}

	// Request the word again - should not create duplicates
	definition2, err := service.GetWordDefinition(ctx, userID, "testword")
	if err != nil {
		t.Fatalf("GetWordDefinition() error on second call = %v", err)
	}

	// Verify no new cards were created (should return existing IDs)
	ruEnCard1Again, err := userCardRepo.GetUserCardByTrainingCard(userID, trainingCardID1, models.DirectionRUtoEN)
	if err != nil {
		t.Fatalf("Failed to get ru_en card 1 again: %v", err)
	}
	if ruEnCard1Again == nil || ruEnCard1 == nil {
		t.Fatal("Expected user cards to exist for duplicate check")
	}
	if ruEnCard1Again.ID != ruEnCard1.ID {
		t.Errorf("Expected same card ID, got %d != %d", ruEnCard1Again.ID, ruEnCard1.ID)
	}

	_ = definition2 // Use the variable
}

func TestWordService_GetWordDefinition_SchedulesPronunciationByCanonicalWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, wordRepo := setupWordServiceTestDB(t)

	displayEN := "to spy"
	definitionRU := "шпионить"
	wordCardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{
		Word:         "spy",
		Definition:   "to secretly collect information",
		DisplayEN:    &displayEN,
		DefinitionRU: &definitionRU,
	})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma error: %v", err)
	}
	if wordCardID == 0 {
		t.Fatalf("expected non-zero wordCardID")
	}

	pronService := NewPronunciationService(config.TTSConfig{
		Enabled:           true,
		Provider:          "dictionary",
		DictionaryEnabled: true,
		DictionaryBaseURL: "http://127.0.0.1:1/api/v2/entries/en",
		AudioDir:          t.TempDir(),
		PublicBasePath:    "/media/tts",
		PrefetchEnabled:   true,
		PrefetchWorkers:   1,
	}, nil, zap.NewNop())

	service := NewWordService(wordRepo, nil, nil, nil, logger)
	service.SetPronunciationService(pronService)

	_, err = service.GetWordDefinition(context.Background(), 123, "spy")
	if err != nil {
		t.Fatalf("GetWordDefinition() error = %v", err)
	}

	select {
	case scheduled := <-pronService.queue:
		if scheduled != "spy" {
			t.Fatalf("expected pronunciation scheduled for canonical word 'spy', got %q", scheduled)
		}
	default:
		t.Fatalf("expected pronunciation scheduling for canonical word")
	}

	select {
	case extra := <-pronService.queue:
		t.Fatalf("expected no extra pronunciation scheduling (e.g., display form), got %q", extra)
	default:
	}
}

func stringPtr(s string) *string {
	return &s
}
