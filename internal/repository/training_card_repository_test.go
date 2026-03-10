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

	repo := NewTrainingCardRepository(db, logger)
	_ = repo // Verify repository is created
}

func TestTrainingCardRepository_CreateTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)

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

func TestTrainingCardRepository_CreateTrainingCard_ReturnsExistingIDWhenDuplicate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "dup", "def")
	card := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "dup",
		SenseIndex: 0,
		WordRU:     "дубль",
		MeaningEN:  "duplicate",
	}
	id1, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard() first error = %v", err)
	}
	id2, err := repo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard() second (duplicate) error = %v", err)
	}
	if id1 != id2 {
		t.Errorf("duplicate CreateTrainingCard should return same ID: %d vs %d", id1, id2)
	}
}

func TestTrainingCardRepository_CreateTrainingCard_WithDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)

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

func TestTrainingCardRepository_GetWordCardsWithoutTrainingCards_ExcludesProcessedAt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	// Word card with processed_at set (no training card) must not appear
	_, err := db.Exec("INSERT INTO word_cards (word, definition, processed_at) VALUES (?, ?, CURRENT_TIMESTAMP)", "processed_word", "def")
	if err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	cards, err := repo.GetWordCardsWithoutTrainingCards(10)
	if err != nil {
		t.Fatalf("GetWordCardsWithoutTrainingCards() error = %v", err)
	}
	for _, c := range cards {
		if c.Word == "processed_word" {
			t.Error("Expected processed_word to be excluded (processed_at is set)")
		}
	}
}

func TestTrainingCardRepository_GetWordCardsWithoutTrainingCards_OptionalFieldsAndLimit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	// Word card with optional fields set, no training card
	_, err := db.Exec(`INSERT INTO word_cards (word, definition, pos, transcription, definition_ru) VALUES (?, ?, ?, ?, ?)`,
		"optword", "def", "noun", "ˈɒpt", "определение")
	if err != nil {
		t.Fatalf("insert word card: %v", err)
	}
	cards, err := repo.GetWordCardsWithoutTrainingCards(10)
	if err != nil {
		t.Fatalf("GetWordCardsWithoutTrainingCards() error = %v", err)
	}
	var found *models.WordCard
	for _, c := range cards {
		if c.Word == "optword" {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("Expected to find optword in results")
	}
	if found.POS == nil || *found.POS != "noun" {
		t.Errorf("Expected POS noun, got %v", found.POS)
	}
	if found.Transcription == nil || *found.Transcription != "ˈɒpt" {
		t.Errorf("Expected Transcription, got %v", found.Transcription)
	}
	if found.DefinitionRU == nil || *found.DefinitionRU != "определение" {
		t.Errorf("Expected DefinitionRU, got %v", found.DefinitionRU)
	}

	// Limit 0 returns no rows
	empty, err := repo.GetWordCardsWithoutTrainingCards(0)
	if err != nil {
		t.Fatalf("GetWordCardsWithoutTrainingCards(0) error = %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("Expected 0 cards for limit 0, got %d", len(empty))
	}
}

func TestTrainingCardRepository_GetTrainingCardByWordCardIDAndSenseIndex(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupTrainingCardTestDB(t)
	repo := NewTrainingCardRepository(db, logger)

	t.Run("not found returns nil", func(t *testing.T) {
		_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "senseword", "def")
		if err != nil {
			t.Fatalf("insert word card: %v", err)
		}
		got, err := repo.GetTrainingCardByWordCardIDAndSenseIndex(1, 99)
		if err != nil {
			t.Fatalf("GetTrainingCardByWordCardIDAndSenseIndex() error = %v", err)
		}
		if got != nil {
			t.Errorf("expected nil when no row, got %+v", got)
		}
	})

	t.Run("found with pos and display_word", func(t *testing.T) {
		_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "posword", "def")
		pos := "verb"
		displayWord := "to run"
		_, err := db.Exec(`INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word)
			VALUES (2, 'posword', 1, 'бежать', 'to run', $1, $2)`, pos, displayWord)
		if err != nil {
			t.Fatalf("insert training card: %v", err)
		}
		got, err := repo.GetTrainingCardByWordCardIDAndSenseIndex(2, 1)
		if err != nil {
			t.Fatalf("GetTrainingCardByWordCardIDAndSenseIndex() error = %v", err)
		}
		if got == nil {
			t.Fatal("expected card")
		}
		if got.WordCardID != 2 || got.SenseIndex != 1 {
			t.Errorf("expected word_card_id=2 sense_index=1, got %d %d", got.WordCardID, got.SenseIndex)
		}
		if got.POS == nil || *got.POS != pos {
			t.Errorf("expected POS %q, got %v", pos, got.POS)
		}
		if got.DisplayWord == nil || *got.DisplayWord != displayWord {
			t.Errorf("expected DisplayWord %q, got %v", displayWord, got.DisplayWord)
		}
	})
}
