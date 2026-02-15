package repository

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestUserCardRepository_DeleteUserCardsByWordCardIDForUser(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12351)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card
	wordCard := &models.WordCard{
		Word:       "deleteword",
		Definition: "definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	pos := "verb"
	displayWord := "delete"
	trainingCard := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "deleteword",
		SenseIndex:  0,
		WordRU:      "удалить",
		MeaningEN:   "to delete",
		POS:         &pos,
		DisplayWord: &displayWord,
	}
	trainingCardID, err := trainingRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card
	now := time.Now()
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	userCardID, err := repo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Delete user cards by word card ID
	rowsAffected, err := repo.DeleteUserCardsByWordCardIDForUser(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("DeleteUserCardsByWordCardIDForUser() error = %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected, got %d", rowsAffected)
	}

	// Verify deletion - try to get the card, should not exist
	card, err := repo.GetUserCard(userCardID)
	if err != nil {
		t.Fatalf("GetUserCard() error = %v", err)
	}
	if card != nil {
		t.Error("User card should be deleted")
	}
}

func TestUserCardRepository_ListOrphanedUserCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12352)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card, training_card, user_card, then delete training_card to orphan user_card
	wordCard := &models.WordCard{Word: "orphlist", Definition: "def"}
	wcID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "orphlist"
	tc := &models.TrainingCard{WordCardID: wcID, WordEN: "orphlist", SenseIndex: 0, WordRU: "сирота", MeaningEN: "orphan", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)
	now := time.Now()
	uc := &models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateNew, EF: 2.5, NextDueAt: &now}
	_, _ = repo.CreateUserCard(uc)
	ctx := context.Background()
	c, _ := db.Conn(ctx)
	_, _ = c.ExecContext(ctx, "SET session_replication_role = replica")
	_, _ = c.ExecContext(ctx, "DELETE FROM training_cards WHERE id = $1", tcID)
	_, _ = c.ExecContext(ctx, "SET session_replication_role = DEFAULT")
	c.Close()

	// List orphaned user cards
	orphaned, err := repo.ListOrphanedUserCards(10, 0)
	if err != nil {
		t.Fatalf("ListOrphanedUserCards() error = %v", err)
	}
	if len(orphaned) == 0 {
		t.Error("Expected at least one orphaned user card")
	}

	// Verify structure
	for _, item := range orphaned {
		if item.UserCardID == 0 {
			t.Error("UserCardID should not be zero")
		}
		if item.UserID == 0 {
			t.Error("UserID should not be zero")
		}
		if item.TrainingCardID == 0 {
			t.Error("TrainingCardID should not be zero")
		}
	}
}

func TestUserCardRepository_CountOrphanedUserCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12353)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word_card, training_card, user_card, then delete training_card to orphan user_card
	wordCard := &models.WordCard{Word: "orphcount", Definition: "def"}
	wcID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "orphcount"
	tc := &models.TrainingCard{WordCardID: wcID, WordEN: "orphcount", SenseIndex: 0, WordRU: "сирота", MeaningEN: "orphan", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)
	now := time.Now()
	uc := &models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateNew, EF: 2.5, NextDueAt: &now}
	_, _ = repo.CreateUserCard(uc)
	ctx := context.Background()
	c, _ := db.Conn(ctx)
	_, _ = c.ExecContext(ctx, "SET session_replication_role = replica")
	_, _ = c.ExecContext(ctx, "DELETE FROM training_cards WHERE id = $1", tcID)
	_, _ = c.ExecContext(ctx, "SET session_replication_role = DEFAULT")
	c.Close()

	// Count orphaned user cards
	count, err := repo.CountOrphanedUserCards()
	if err != nil {
		t.Fatalf("CountOrphanedUserCards() error = %v", err)
	}
	if count < 1 {
		t.Errorf("Expected at least 1 orphaned user card, got %d", count)
	}
}

func TestUserCardRepository_DeleteUserCard(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12354)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card and training card
	wordCard := &models.WordCard{
		Word:       "deletecard",
		Definition: "definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	pos5 := "verb"
	displayWord5 := "delete"
	trainingCard := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "deletecard",
		SenseIndex:  0,
		WordRU:      "удалить",
		MeaningEN:   "to delete",
		POS:         &pos5,
		DisplayWord: &displayWord5,
	}
	trainingCardID, err := trainingRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card
	now := time.Now()
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	userCardID, err := repo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Delete user card
	err = repo.DeleteUserCard(userCardID)
	if err != nil {
		t.Fatalf("DeleteUserCard() error = %v", err)
	}

	// Verify deletion
	card, err := repo.GetUserCard(userCardID)
	if err != nil {
		t.Fatalf("GetUserCard() error = %v", err)
	}
	if card != nil {
		t.Error("GetUserCard() should return nil for deleted card")
	}
}

func TestUserCardRepository_GetUserIDsByWordCardID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create users
	user1, err := userRepo.GetOrCreateUser(12355)
	if err != nil {
		t.Fatalf("Failed to create user 1: %v", err)
	}
	user2, err := userRepo.GetOrCreateUser(12356)
	if err != nil {
		t.Fatalf("Failed to create user 2: %v", err)
	}

	// Create word card
	wordCard := &models.WordCard{
		Word:       "sharedword",
		Definition: "definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card
	pos2 := "adjective"
	displayWord2 := "shared"
	trainingCard := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "sharedword",
		SenseIndex:  0,
		WordRU:      "общее",
		MeaningEN:   "shared",
		POS:         &pos2,
		DisplayWord: &displayWord2,
	}
	trainingCardID, err := trainingRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user cards for both users
	now := time.Now()
	userCard1 := &models.UserCard{
		UserID:         user1.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(userCard1)
	if err != nil {
		t.Fatalf("Failed to create user card 1: %v", err)
	}

	userCard2 := &models.UserCard{
		UserID:         user2.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      &now,
	}
	_, err = repo.CreateUserCard(userCard2)
	if err != nil {
		t.Fatalf("Failed to create user card 2: %v", err)
	}

	// Get user IDs by word card ID
	userIDs, err := repo.GetUserIDsByWordCardID(wordCardID)
	if err != nil {
		t.Fatalf("GetUserIDsByWordCardID() error = %v", err)
	}
	if len(userIDs) != 2 {
		t.Errorf("Expected 2 user IDs, got %d", len(userIDs))
	}

	// Verify both users are in the list
	userMap := make(map[int64]bool)
	for _, id := range userIDs {
		userMap[id] = true
	}
	if !userMap[user1.ID] || !userMap[user2.ID] {
		t.Error("Expected both user IDs in results")
	}
}

func TestUserCardRepository_GetUpcomingCardsByDate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12357)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word card and training card
	wordCard := &models.WordCard{
		Word:       "upcoming",
		Definition: "definition",
	}
	wordCardID, err := wordRepo.UpsertWordCardLemma(wordCard)
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	pos3 := "adjective"
	displayWord3 := "upcoming"
	trainingCard := &models.TrainingCard{
		WordCardID:  wordCardID,
		WordEN:      "upcoming",
		SenseIndex:  0,
		WordRU:      "предстоящий",
		MeaningEN:   "upcoming",
		POS:         &pos3,
		DisplayWord: &displayWord3,
	}
	trainingCardID, err := trainingRepo.CreateTrainingCard(trainingCard)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create user card with future due date
	futureDate := time.Now().AddDate(0, 0, 3)
	userCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: trainingCardID,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.5,
		NextDueAt:      &futureDate,
	}
	_, err = repo.CreateUserCard(userCard)
	if err != nil {
		t.Fatalf("Failed to create user card: %v", err)
	}

	// Get upcoming cards by date
	startDate := time.Now()
	upcoming, err := repo.GetUpcomingCardsByDate(user.ID, startDate)
	if err != nil {
		t.Fatalf("GetUpcomingCardsByDate() error = %v", err)
	}

	// Should have at least one card in the next 7 days
	totalCount := 0
	for _, count := range upcoming {
		totalCount += count
	}
	if totalCount < 1 {
		t.Errorf("Expected at least 1 upcoming card, got %d", totalCount)
	}
}
