package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestUserCardRepository_GetUserCardByTrainingCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)

	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(3000)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	repo := NewUserCardRepository(db, logger)

	// Create word card first, then training card
	var wordCardID int64
	err = db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "bytraining", "by training").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	var trainingCardID int64
	err = db.QueryRow("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		wordCardID, "bytraining", 0, "по обучению", "by training", "noun", "bytraining").Scan(&trainingCardID)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
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
	found, err := repo.GetUserCardByTrainingCard(user.ID, trainingCardID, models.DirectionENtoRU)
	if err != nil {
		t.Fatalf("GetUserCardByTrainingCard() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetUserCardByTrainingCard() should not return nil")
	}
	if found.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, found.UserID)
	}
	if found.TrainingCardID != trainingCardID {
		t.Errorf("Expected TrainingCardID %d, got %d", trainingCardID, found.TrainingCardID)
	}
}

func TestUserCardRepository_DeleteOrphanedUserCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)

	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(4000)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	repo := NewUserCardRepository(db, logger)

	// Create word card first, then training card
	var wordCardID int64
	err = db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "orphan", "orphan").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	var trainingCardID int64
	err = db.QueryRow("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		wordCardID, "orphan", 0, "сирота", "orphan", "noun", "orphan").Scan(&trainingCardID)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Disable triggers so CASCADE doesn't delete user_cards (Postgres)
	ctx := context.Background()
	conn, _ := db.Conn(ctx)
	defer conn.Close()
	_, _ = conn.ExecContext(ctx, "SET session_replication_role = replica")
	_, err = conn.ExecContext(ctx, "DELETE FROM training_cards WHERE id = $1", trainingCardID)
	_, _ = conn.ExecContext(ctx, "SET session_replication_role = DEFAULT")
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

	userRepo := NewUserRepository(db, logger)
	user, err := userRepo.GetOrCreateUser(5000)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	repo := NewUserCardRepository(db, logger)

	// Create word card and training card (schema already exists from migration)
	var wordCardID int64
	err = db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "deleteword", "to delete").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	var trainingCardID int64
	err = db.QueryRow("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		wordCardID, "deleteword", 0, "удалить", "to delete", "verb", "to deleteword").Scan(&trainingCardID)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user cards
	now := time.Now()
	card1 := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
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
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
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
	deleted, err := repo.DeleteUserCardsByWordENForUser(user.ID, "deleteword")
	if err != nil {
		t.Fatalf("DeleteUserCardsByWordENForUser() error = %v", err)
	}
	if deleted < 2 {
		t.Errorf("Expected at least 2 cards deleted, got %d", deleted)
	}
}
