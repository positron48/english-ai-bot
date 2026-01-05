package repository

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestUserCardRepository_GetUserCardByTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, "bytraining", 0, "по обучению", "by training", "noun", "bytraining")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         3000,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get user card by training card
	found, err := repo.GetUserCardByTrainingCard(3000, 1, models.DirectionENtoRU)
	if err != nil {
		t.Fatalf("GetUserCardByTrainingCard() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetUserCardByTrainingCard() should not return nil")
	}
	if found.UserID != 3000 {
		t.Errorf("Expected UserID 3000, got %d", found.UserID)
	}
	if found.TrainingCardID != 1 {
		t.Errorf("Expected TrainingCardID 1, got %d", found.TrainingCardID)
	}
}

func TestUserCardRepository_DeleteOrphanedUserCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	repo := NewUserCardRepository(db, logger)

	// Create a training card
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, "orphan", 0, "сирота", "orphan", "noun", "orphan")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         4000,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Delete the training card (making user card orphaned)
	_, err = db.Exec("DELETE FROM training_cards WHERE id = 1")
	if err != nil {
		t.Fatalf("Failed to delete training card: %v", err)
	}

	// Delete orphaned user cards
	deleted, err := repo.DeleteOrphanedUserCards()
	if err != nil {
		t.Fatalf("DeleteOrphanedUserCards() error = %v", err)
	}
	if deleted == 0 {
		t.Error("DeleteOrphanedUserCards() should delete at least one card")
	}
}

func TestUserCardRepository_DeleteUserCardsByWordENForUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	defer db.Close()

	// Create word_cards table
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS word_cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		word TEXT UNIQUE NOT NULL,
		definition TEXT NOT NULL,
		pos TEXT,
		transcription TEXT,
		definition_ru TEXT,
		examples_json TEXT,
		verb_forms_json TEXT,
		display_en TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create word_cards table: %v", err)
	}

	repo := NewUserCardRepository(db, logger)

	// Create word cards and training cards
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "deleteword", "to delete")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (id, word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		1, 1, "deleteword", 0, "удалить", "to delete", "verb", "to deleteword")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user cards
	now := time.Now()
	card1 := &models.UserCard{
		UserID:         5000,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(card1)
	if err != nil {
		t.Fatalf("Failed to create user card 1: %v", err)
	}

	card2 := &models.UserCard{
		UserID:         5000,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(card2)
	if err != nil {
		t.Fatalf("Failed to create user card 2: %v", err)
	}

	// Delete user cards by word EN
	deleted, err := repo.DeleteUserCardsByWordENForUser(5000, "deleteword")
	if err != nil {
		t.Fatalf("DeleteUserCardsByWordENForUser() error = %v", err)
	}
	if deleted < 2 {
		t.Errorf("Expected at least 2 cards deleted, got %d", deleted)
	}
}
