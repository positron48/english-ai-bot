package service

import (
	"testing"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func setupWordSetServiceTest(t *testing.T) (*WordSetService, *database.DB, func()) {
	logger, _ := zap.NewDevelopment()
	db, err := database.New(":memory:", logger)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	conn := db.GetConnection()
	wordSetRepo := repository.NewWordSetRepository(conn, logger)
	wordSetCategoryRepo := repository.NewWordSetCategoryRepository(conn, logger)
	wordRepo := repository.NewWordRepository(conn, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(conn, logger)
	userCardRepo := repository.NewUserCardRepository(conn, logger)
	userWordKnowledgeRepo := repository.NewUserWordKnowledgeRepository(conn, logger)

	service := NewWordSetService(
		wordSetRepo,
		wordSetCategoryRepo,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userWordKnowledgeRepo,
		nil, // No AI service for basic tests
		"",
		logger,
	)

	cleanup := func() {
		db.Close()
	}

	return service, db, cleanup
}

func TestNewWordSetService(t *testing.T) {
	service, _, cleanup := setupWordSetServiceTest(t)
	defer cleanup()

	if service == nil {
		t.Error("Expected non-nil service")
	}
}

func TestEnsureWordCardExistsMinimal_NewWord(t *testing.T) {
	service, _, cleanup := setupWordSetServiceTest(t)
	defer cleanup()

	wordCardID, err := service.EnsureWordCardExistsMinimal("test")
	if err != nil {
		t.Fatalf("EnsureWordCardExistsMinimal() error = %v", err)
	}

	if wordCardID <= 0 {
		t.Error("Expected positive word card ID")
	}
}

func TestEnsureWordCardExistsMinimal_ExistingWord(t *testing.T) {
	service, _, cleanup := setupWordSetServiceTest(t)
	defer cleanup()

	// Create first
	id1, err := service.EnsureWordCardExistsMinimal("test")
	if err != nil {
		t.Fatalf("First EnsureWordCardExistsMinimal() error = %v", err)
	}

	// Create again - should return same ID
	id2, err := service.EnsureWordCardExistsMinimal("test")
	if err != nil {
		t.Fatalf("Second EnsureWordCardExistsMinimal() error = %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same ID, got %d and %d", id1, id2)
	}
}

func TestEnsureWordCardExistsMinimal_Normalization(t *testing.T) {
	service, _, cleanup := setupWordSetServiceTest(t)
	defer cleanup()

	// Create with spaces and caps
	id1, err := service.EnsureWordCardExistsMinimal("  TEST  ")
	if err != nil {
		t.Fatalf("First EnsureWordCardExistsMinimal() error = %v", err)
	}

	// Create with lowercase
	id2, err := service.EnsureWordCardExistsMinimal("test")
	if err != nil {
		t.Fatalf("Second EnsureWordCardExistsMinimal() error = %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same ID for normalized words, got %d and %d", id1, id2)
	}
}

func TestEnsureUserCardsForWord_NoTrainingCards(t *testing.T) {
	service, _, cleanup := setupWordSetServiceTest(t)
	defer cleanup()

	// Create word card
	wordCardID, err := service.EnsureWordCardExistsMinimal("test")
	if err != nil {
		t.Fatalf("EnsureWordCardExistsMinimal() error = %v", err)
	}

	// Should not error even without training cards
	err = service.EnsureUserCardsForWord(1, wordCardID)
	if err != nil {
		t.Fatalf("EnsureUserCardsForWord() error = %v", err)
	}
}
