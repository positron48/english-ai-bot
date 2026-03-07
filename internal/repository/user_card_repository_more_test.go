package repository

import (
	"context"
	"strings"
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

	// DeleteUserCard with non-existent ID returns error
	errNotFound := repo.DeleteUserCard(99999999)
	if errNotFound == nil {
		t.Error("DeleteUserCard(non-existent id) expected error")
	}
	if errNotFound != nil && !strings.Contains(errNotFound.Error(), "not found") {
		t.Errorf("DeleteUserCard(non-existent): expected 'not found' in error, got %v", errNotFound)
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

// TestUserCardRepository_GetUpcomingCardsByDate_EmptyResult verifies empty map when user has no upcoming cards in range.
func TestUserCardRepository_GetUpcomingCardsByDate_EmptyResult(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	user, _ := userRepo.GetOrCreateUser(123570)
	// User has no user_cards; upcoming should be 0 for all 7 days.
	startDate := time.Now()
	upcoming, err := repo.GetUpcomingCardsByDate(user.ID, startDate)
	if err != nil {
		t.Fatalf("GetUpcomingCardsByDate() error = %v", err)
	}
	totalCount := 0
	for _, count := range upcoming {
		totalCount += count
	}
	if totalCount != 0 {
		t.Errorf("expected 0 upcoming cards for user with no cards, got %d", totalCount)
	}
	// Result should still have 7 days initialized to 0
	if len(upcoming) != 7 {
		t.Errorf("expected 7 days in result map, got %d", len(upcoming))
	}
}

// Test_nullTimeScanner_parseString covers the parseString method used when scanning time from string/[]byte.
func Test_nullTimeScanner_parseString(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr bool
	}{
		{"RFC3339", "2006-01-02T15:04:05Z", false},
		{"standard", "2006-01-02 15:04:05", false},
		{"invalid", "not-a-date", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n nullTimeScanner
			err := n.parseString(tt.s)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseString() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !n.Valid {
				t.Fatal("expected Valid true")
			}
		})
	}
}

// Test_nullTimeScanner_Scan covers Scan with nil, time.Time, []byte, string, and invalid type.
func Test_nullTimeScanner_Scan(t *testing.T) {
	refTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	tests := []struct {
		name      string
		value     interface{}
		wantValid bool
		wantErr   bool
	}{
		{"nil", nil, false, false},
		{"time.Time", refTime, true, false},
		{"[]byte", []byte("2006-01-02 15:04:05"), true, false},
		{"string RFC3339", "2006-01-02T15:04:05Z", true, false},
		{"string with TZ", "2006-01-02 15:04:05-07:00", true, false},
		{"invalid type", 42, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n nullTimeScanner
			err := n.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && n.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", n.Valid, tt.wantValid)
			}
			if tt.wantValid && n.Valid && n.Time.Year() != 2006 {
				t.Errorf("expected year 2006, got %d", n.Time.Year())
			}
		})
	}
}

func TestUserCardRepository_GetWordMasteringStats(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(12358)
	wordCard := &models.WordCard{Word: "masterword", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "masterword"
	tc := &models.TrainingCard{WordCardID: wordCardID, WordEN: "masterword", SenseIndex: 0, WordRU: "мастер", MeaningEN: "master", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)

	now := time.Now()
	uc := &models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.5, Reps: 3, NextDueAt: &now}
	_, _ = repo.CreateUserCard(uc)

	stats, err := repo.GetWordMasteringStats(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("GetWordMasteringStats() error = %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.TotalCards != 1 || stats.TotalReps != 3 {
		t.Errorf("expected TotalCards=1 TotalReps=3, got TotalCards=%d TotalReps=%d", stats.TotalCards, stats.TotalReps)
	}

	// no cards for word -> nil (or empty stats when query returns a row with zeros)
	statsNil, err := repo.GetWordMasteringStats(user.ID, 99999)
	if err != nil {
		t.Fatalf("GetWordMasteringStats(no cards) error = %v", err)
	}
	if statsNil != nil && statsNil.TotalCards != 0 {
		t.Errorf("expected nil or empty stats for unknown word_card_id, got TotalCards=%d", statsNil.TotalCards)
	}
}

// TestUserCardRepository_GetWordMasteringStats_NoCardsForWord verifies stats with zeros when user has no user_cards for the word (aggregate returns one row with zeros in Postgres).
func TestUserCardRepository_GetWordMasteringStats_NoCardsForWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(123580)
	wordCard := &models.WordCard{Word: "norowword", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "norowword"
	tc := &models.TrainingCard{WordCardID: wordCardID, WordEN: "norowword", SenseIndex: 0, WordRU: "слово", MeaningEN: "word", POS: &pos, DisplayWord: &displayWord}
	_, _ = trainingRepo.CreateTrainingCard(tc)
	// Do not create any user_card for this user+word. Aggregate still returns one row with zeros.
	stats, err := repo.GetWordMasteringStats(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("GetWordMasteringStats() error = %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats (aggregate returns one row with zeros)")
	}
	if stats.TotalCards != 0 || stats.TotalReps != 0 || stats.IsKnown {
		t.Errorf("expected zero stats when no user_cards for word, got %+v", stats)
	}
}

// TestUserCardRepository_GetWordMasteringStats_IsKnown verifies IsKnown is true when user has word in user_word_knowledge as 'known'.
func TestUserCardRepository_GetWordMasteringStats_IsKnown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(123581)
	wordCard := &models.WordCard{Word: "knownword", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "knownword"
	tc := &models.TrainingCard{WordCardID: wordCardID, WordEN: "knownword", SenseIndex: 0, WordRU: "известный", MeaningEN: "known", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)
	now := time.Now()
	_, _ = repo.CreateUserCard(&models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.5, NextDueAt: &now})
	_, err := db.Exec("INSERT INTO user_word_knowledge (user_id, word_card_id, status) VALUES ($1, $2, 'known') ON CONFLICT (user_id, word_card_id) DO UPDATE SET status = 'known'", user.ID, wordCardID)
	if err != nil {
		t.Fatalf("insert user_word_knowledge: %v", err)
	}
	stats, err := repo.GetWordMasteringStats(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("GetWordMasteringStats() error = %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if !stats.IsKnown {
		t.Errorf("expected IsKnown true when word in user_word_knowledge, got false")
	}
}

// TestUserCardRepository_GetWordMasteringStats_NoTrainingCardForWord verifies that when word_card has no training_cards
// the query returns no rows and we get nil,nil; or if the DB returns one aggregate row with zeros, stats are all zero.
func TestUserCardRepository_GetWordMasteringStats_NoTrainingCardForWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(123582)
	wordCard := &models.WordCard{Word: "notraining", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	// Do not create any training_card for this word_card (no user_cards either)
	stats, err := repo.GetWordMasteringStats(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("GetWordMasteringStats() error = %v", err)
	}
	// Query has JOIN so with no training_cards we get 0 rows -> nil; some DBs may return one row of zeros
	if stats != nil {
		if stats.TotalCards != 0 || stats.TotalReps != 0 || stats.IsKnown {
			t.Errorf("when word has no training cards expected nil or zero stats, got %+v", stats)
		}
	}
}

// TestUserCardRepository_GetWordMasteringStats_MultipleCardsSameWord verifies aggregates when user has multiple user_cards for same word (e.g. two senses/directions).
func TestUserCardRepository_GetWordMasteringStats_MultipleCardsSameWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(123583)
	wordCard := &models.WordCard{Word: "multicard", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "noun"
	displayWord := "multicard"
	tc1 := &models.TrainingCard{WordCardID: wordCardID, WordEN: "multicard", SenseIndex: 0, WordRU: "мульти", MeaningEN: "multi", POS: &pos, DisplayWord: &displayWord}
	tc2 := &models.TrainingCard{WordCardID: wordCardID, WordEN: "multicard", SenseIndex: 1, WordRU: "ещё", MeaningEN: "more", POS: &pos, DisplayWord: &displayWord}
	tcID1, _ := trainingRepo.CreateTrainingCard(tc1)
	tcID2, _ := trainingRepo.CreateTrainingCard(tc2)

	now := time.Now()
	_, _ = repo.CreateUserCard(&models.UserCard{UserID: user.ID, TrainingCardID: tcID1, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.0, Reps: 5, NextDueAt: &now})
	_, _ = repo.CreateUserCard(&models.UserCard{UserID: user.ID, TrainingCardID: tcID2, Direction: models.DirectionRUtoEN, State: models.StateLearning, EF: 2.5, Reps: 2, NextDueAt: &now})

	stats, err := repo.GetWordMasteringStats(user.ID, wordCardID)
	if err != nil {
		t.Fatalf("GetWordMasteringStats() error = %v", err)
	}
	if stats == nil {
		t.Fatal("expected non-nil stats")
	}
	if stats.TotalCards != 2 {
		t.Errorf("TotalCards = %d, want 2", stats.TotalCards)
	}
	if stats.TotalReps != 7 {
		t.Errorf("TotalReps = %d, want 7", stats.TotalReps)
	}
	if stats.ReviewStateCount != 1 || stats.LearningStateCount != 1 {
		t.Errorf("ReviewStateCount=%d LearningStateCount=%d, want 1 and 1", stats.ReviewStateCount, stats.LearningStateCount)
	}
}

func TestUserCardRepository_GetWordsEligibleForSpell_GetWordsEligibleForSpellByMastery(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	repo := NewUserCardRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	trainingRepo := NewTrainingCardRepository(db, logger)

	user, _ := userRepo.GetOrCreateUser(12359)
	wordCard := &models.WordCard{Word: "spellword", Definition: "def"}
	wordCardID, _ := wordRepo.UpsertWordCardLemma(wordCard)
	pos := "verb"
	displayWord := "spell"
	tc := &models.TrainingCard{WordCardID: wordCardID, WordEN: "spellword", SenseIndex: 0, WordRU: "писать", MeaningEN: "to spell", POS: &pos, DisplayWord: &displayWord}
	tcID, _ := trainingRepo.CreateTrainingCard(tc)

	now := time.Now()
	uc := &models.UserCard{UserID: user.ID, TrainingCardID: tcID, Direction: models.DirectionENtoRU, State: models.StateReview, EF: 2.5, NextDueAt: &now}
	_, _ = repo.CreateUserCard(uc)

	// Insert mastering_score so word is eligible for minScore 50
	_, _ = db.Exec("INSERT INTO user_word_mastering (user_id, word_card_id, mastering_score) VALUES ($1, $2, $3) ON CONFLICT (user_id, word_card_id) DO UPDATE SET mastering_score = $3", user.ID, wordCardID, 60)

	words, err := repo.GetWordsEligibleForSpell(user.ID, 10)
	if err != nil {
		t.Fatalf("GetWordsEligibleForSpell() error = %v", err)
	}
	if len(words) < 1 {
		t.Errorf("expected at least 1 word eligible for spell, got %d", len(words))
	}

	wordsByMastery, err := repo.GetWordsEligibleForSpellByMastery(user.ID, 50, 10)
	if err != nil {
		t.Fatalf("GetWordsEligibleForSpellByMastery() error = %v", err)
	}
	if len(wordsByMastery) < 1 {
		t.Errorf("expected at least 1 word with mastery >= 50, got %d", len(wordsByMastery))
	}
	// minScore clamping
	_, err = repo.GetWordsEligibleForSpellByMastery(user.ID, -1, 10)
	if err != nil {
		t.Fatalf("GetWordsEligibleForSpellByMastery(minScore -1) error = %v", err)
	}
	_, err = repo.GetWordsEligibleForSpellByMastery(user.ID, 101, 10)
	if err != nil {
		t.Fatalf("GetWordsEligibleForSpellByMastery(minScore 101) error = %v", err)
	}
}
