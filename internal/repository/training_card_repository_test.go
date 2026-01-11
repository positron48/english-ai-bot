package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupTrainingCardTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewTrainingCardRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)
	_ = repo // Verify repository is created
}

func TestTrainingCardRepository_CreateTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// Create a word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "hello", "a greeting")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	card := &models.TrainingCard{
		WordCardID:    1,
		WordEN:        "hello",
		Transcription: "həˈloʊ",
		SenseIndex:    0,
		WordRU:        "привет",
		MeaningEN:     "a greeting",
		ExampleEN:     "Hello, world!",
		ExampleRU:     "Привет, мир!",
		DistractorsRU: `["пока", "да"]`,
		DistractorsEN: `["hi", "hey"]`,
		Hint:          "greeting",
	}

	id, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard() error = %v", err)
	}
	if id == 0 {
		t.Error("CreateTrainingCard() should return non-zero ID")
	}
}

func TestTrainingCardRepository_CreateTrainingCard_WithDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// Create a word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "spy", "a secret agent")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	displayWord := "to spy"
	pos := "verb"
	card := &models.TrainingCard{
		WordCardID:    1,
		WordEN:        "spy",
		Transcription: "spaɪ",
		SenseIndex:    0,
		WordRU:        "шпионить",
		MeaningEN:     "to watch secretly",
		DisplayWord:   &displayWord,
		POS:           &pos,
	}

	id, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard() error = %v", err)
	}
	if id == 0 {
		t.Error("CreateTrainingCard() should return non-zero ID")
	}

	// Verify display_word was saved
	found, err := repo.GetTrainingCard(id)
	if err != nil {
		t.Fatalf("GetTrainingCard() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetTrainingCard() should not return nil")
	}
	if found.DisplayWord == nil || *found.DisplayWord != displayWord {
		t.Errorf("Expected DisplayWord %q, got %v", displayWord, found.DisplayWord)
	}
}

func TestTrainingCardRepository_GetTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// Create a word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "world", "the earth")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create a training card
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "world",
		SenseIndex: 0,
		WordRU:     "мир",
		MeaningEN:  "the earth",
	}
	id, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Get the training card
	found, err := repo.GetTrainingCard(id)
	if err != nil {
		t.Fatalf("GetTrainingCard() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetTrainingCard() should not return nil")
	}
	if found.WordEN != "world" {
		t.Errorf("Expected WordEN 'world', got %q", found.WordEN)
	}
}

func TestTrainingCardRepository_GetTrainingCardsByWordCardID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// Create a word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "test", "a test")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create multiple training cards for the same word
	card1 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "test",
		SenseIndex: 0,
		WordRU:     "тест1",
		MeaningEN:  "meaning 1",
	}
	card2 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "test",
		SenseIndex: 1,
		WordRU:     "тест2",
		MeaningEN:  "meaning 2",
	}

	_, err = repo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card 1: %v", err)
	}
	_, err = repo.CreateTrainingCard(card2)
	if err != nil {
		t.Fatalf("Failed to create training card 2: %v", err)
	}

	// Get all training cards for this word
	cards, err := repo.GetTrainingCardsByWordCardID(1)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID() error = %v", err)
	}
	if len(cards) != 2 {
		t.Errorf("Expected 2 cards, got %d", len(cards))
	}
}

func TestTrainingCardRepository_HasTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// Create a word card first
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "check", "to verify")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Check before creating training card
	has, err := repo.HasTrainingCards(1)
	if err != nil {
		t.Fatalf("HasTrainingCards() error = %v", err)
	}
	if has {
		t.Error("HasTrainingCards() should return false before creating training card")
	}

	// Create a training card
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "check",
		SenseIndex: 0,
		WordRU:     "проверить",
		MeaningEN:  "to verify",
	}
	_, err = repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Check after creating training card
	has, err = repo.HasTrainingCards(1)
	if err != nil {
		t.Fatalf("HasTrainingCards() error = %v", err)
	}
	if !has {
		t.Error("HasTrainingCards() should return true after creating training card")
	}
}

func TestTrainingCardRepository_GetWordCardsWithoutTrainingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	defer db.Close()

	repo := NewTrainingCardRepository(db, logger)

	// Create word cards
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "word1", "def1")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	_, err = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "word2", "def2")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card only for word1
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "word1",
		SenseIndex: 0,
		WordRU:     "слово1",
		MeaningEN:  "def1",
	}
	_, err = repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Get word cards without training cards
	cards, err := repo.GetWordCardsWithoutTrainingCards(10)
	if err != nil {
		t.Fatalf("GetWordCardsWithoutTrainingCards() error = %v", err)
	}
	// Should return word2 (word1 has training card)
	found := false
	for _, c := range cards {
		if c.Word == "word2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find word2 in results")
	}
}
