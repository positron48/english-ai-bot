package repository

import (
	"testing"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestWordRepository_UpdateWordCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Create a word card first
	wordCard := &models.WordCard{
		Word:       "test",
		Definition: "test definition",
	}
	id, err := repo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Update the word card with all new fields
	pos := "noun"
	transcription := "/test/"
	definitionRU := "тест"
	examplesJSON := `[{"en": "This is a test", "ru": "Это тест"}]`
	verbFormsJSON := ""
	displayEN := "test"

	updatedCard := &models.WordCard{
		ID:            id,
		Word:          "test",
		Definition:    "updated definition",
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU: &definitionRU,
		ExamplesJSON: &examplesJSON,
		VerbFormsJSON: &verbFormsJSON,
		DisplayEN:     &displayEN,
	}

	err = repo.UpdateWordCard(updatedCard)
	if err != nil {
		t.Fatalf("UpdateWordCard() error = %v", err)
	}

	// Verify the update
	retrieved, err := repo.GetWordCardByID(id)
	if err != nil {
		t.Fatalf("GetWordCardByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetWordCardByID() returned nil")
	}

	if retrieved.Word != "test" {
		t.Errorf("Expected Word 'test', got %q", retrieved.Word)
	}
	if retrieved.Definition != "updated definition" {
		t.Errorf("Expected Definition 'updated definition', got %q", retrieved.Definition)
	}
	if retrieved.POS == nil || *retrieved.POS != "noun" {
		t.Errorf("Expected POS 'noun', got %v", retrieved.POS)
	}
	if retrieved.Transcription == nil || *retrieved.Transcription != "/test/" {
		t.Errorf("Expected Transcription '/test/', got %v", retrieved.Transcription)
	}
	if retrieved.DefinitionRU == nil || *retrieved.DefinitionRU != "тест" {
		t.Errorf("Expected DefinitionRU 'тест', got %v", retrieved.DefinitionRU)
	}
	if retrieved.DisplayEN == nil || *retrieved.DisplayEN != "test" {
		t.Errorf("Expected DisplayEN 'test', got %v", retrieved.DisplayEN)
	}
}

func TestWordRepository_UpdateWordCard_PartialUpdate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Create a word card first
	wordCard := &models.WordCard{
		Word:       "partial",
		Definition: "original definition",
	}
	id, err := repo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Update only some fields
	pos := "verb"
	displayEN := "to partial"

	updatedCard := &models.WordCard{
		ID:         id,
		Word:       "partial",
		Definition: "original definition", // Keep original
		POS:        &pos,
		DisplayEN:  &displayEN,
	}

	err = repo.UpdateWordCard(updatedCard)
	if err != nil {
		t.Fatalf("UpdateWordCard() error = %v", err)
	}

	// Verify the update
	retrieved, err := repo.GetWordCardByID(id)
	if err != nil {
		t.Fatalf("GetWordCardByID() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetWordCardByID() returned nil")
	}

	if retrieved.POS == nil || *retrieved.POS != "verb" {
		t.Errorf("Expected POS 'verb', got %v", retrieved.POS)
	}
	if retrieved.DisplayEN == nil || *retrieved.DisplayEN != "to partial" {
		t.Errorf("Expected DisplayEN 'to partial', got %v", retrieved.DisplayEN)
	}
	if retrieved.Definition != "original definition" {
		t.Errorf("Expected Definition to remain 'original definition', got %q", retrieved.Definition)
	}
}
