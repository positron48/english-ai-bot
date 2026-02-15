package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupUserCardTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewUserCardRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)

	repo := NewUserCardRepository(db, logger)
	_ = repo // Verify repository is created
}

func TestUserCardRepository_CreateUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(123)

	repo := NewUserCardRepository(db, logger)

	// Create word_card and training card first
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "hello", "greeting")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, ?, ?, ?, ?, ?, ?)",
		"hello", 0, "привет", "greeting", "noun", "hello")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		Reps:           0,
		IntervalDays:   0,
		LearningStep:   0,
		LapseCount:     0,
		NextDueAt:      &now,
	}

	id, err := repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("CreateUserCard() error = %v", err)
	}
	if id == 0 {
		t.Error("CreateUserCard() should return non-zero ID")
	}
}

func TestUserCardRepository_GetUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(456)

	repo := NewUserCardRepository(db, logger)

	// Create word_card and training card first
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "world", "earth")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, ?, ?, ?, ?, ?, ?)",
		"world", 0, "мир", "earth", "noun", "world")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	now := time.Now()
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateLearning,
		EF:             2.0,
		Reps:           1,
		IntervalDays:   3,
		LearningStep:   1,
		NextDueAt:      &now,
	}
	id, err := repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get the user card
	found, err := repo.GetUserCard(id)
	if err != nil {
		t.Fatalf("GetUserCard() error = %v", err)
	}
	if found == nil {
		t.Fatal("GetUserCard() should not return nil")
	}
	if found.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, found.UserID)
	}
	if found.State != models.StateLearning {
		t.Errorf("Expected State %v, got %v", models.StateLearning, found.State)
	}
}

func TestUserCardRepository_GetDueCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(789)

	repo := NewUserCardRepository(db, logger)

	// Create word_card and training card first
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "test", "test")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, ?, ?, ?, ?, ?, ?)",
		"test", 0, "тест", "test", "noun", "test")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user cards - one due, one not due
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	dueCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = repo.CreateUserCard(dueCard)
	if err != nil {
		t.Fatalf("Failed to create due card: %v", err)
	}

	notDueCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &future,
	}
	_, err = repo.CreateUserCard(notDueCard)
	if err != nil {
		t.Fatalf("Failed to create not due card: %v", err)
	}

	// Get due cards
	dueCards, err := repo.GetDueCards(user.ID, now, 10)
	if err != nil {
		t.Fatalf("GetDueCards() error = %v", err)
	}
	if len(dueCards) == 0 {
		t.Error("GetDueCards() should return at least one card")
	}
}

func TestUserCardRepository_GetNewCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(999)

	repo := NewUserCardRepository(db, logger)

	// Create word_card and training card first
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "new", "new")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, ?, ?, ?, ?, ?, ?)",
		"new", 0, "новый", "new", "adjective", "new")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a new card
	newCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = repo.CreateUserCard(newCard)
	if err != nil {
		t.Fatalf("Failed to create new card: %v", err)
	}

	// Get new cards
	newCards, err := repo.GetNewCards(user.ID, 10)
	if err != nil {
		t.Fatalf("GetNewCards() error = %v", err)
	}
	if len(newCards) == 0 {
		t.Error("GetNewCards() should return at least one card")
	}
	if newCards[0].State != models.StateNew {
		t.Errorf("Expected State %v, got %v", models.StateNew, newCards[0].State)
	}
}

func TestUserCardRepository_UpdateUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(111)

	repo := NewUserCardRepository(db, logger)

	// Create word_card and training card first
	_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", "update", "update")
	_, err := db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, ?, ?, ?, ?, ?, ?)",
		"update", 0, "обновить", "update", "verb", "to update")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create a user card
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	id, err := repo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Update the card
	card.ID = id
	card.State = models.StateReview
	card.EF = 2.2
	card.Reps = 5
	card.IntervalDays = 10

	err = repo.UpdateUserCard(card)
	if err != nil {
		t.Fatalf("UpdateUserCard() error = %v", err)
	}

	// Verify update
	updated, err := repo.GetUserCard(id)
	if err != nil {
		t.Fatalf("GetUserCard() error = %v", err)
	}
	if updated.State != models.StateReview {
		t.Errorf("Expected State %v, got %v", models.StateReview, updated.State)
	}
	if updated.EF != 2.2 {
		t.Errorf("Expected EF 2.2, got %f", updated.EF)
	}
}

func TestUserCardRepository_GetDueCount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupUserCardTestDB(t)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(222)

	repo := NewUserCardRepository(db, logger)

	// Create word_cards and training cards first
	var err error
	for i := 1; i <= 3; i++ {
		word := "count" + string(rune('0'+i))
		_, _ = db.Exec("INSERT INTO word_cards (word, definition) VALUES (?, ?)", word, "count")
		_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
			i, "count", 0, "считать", "count", "verb", "to count")
		if err != nil {
			t.Fatalf("Failed to create training card: %v", err)
		}
	}

	// Create due cards
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	for i := 1; i <= 3; i++ {
		card := &models.UserCard{
			UserID:         user.ID,
			TrainingCardID: int64(i),
			Direction:      models.DirectionENtoRU,
			State:          models.StateReview,
			EF:             2.0,
			NextDueAt:      &past,
		}
		_, err = repo.CreateUserCard(card)
		if err != nil {
			t.Fatalf("Failed to create user card: %v", err)
		}
	}

	// Get due count
	count, err := repo.GetDueCount(user.ID, now)
	if err != nil {
		t.Fatalf("GetDueCount() error = %v", err)
	}
	if count < 3 {
		t.Errorf("Expected at least 3 due cards, got %d", count)
	}
}
