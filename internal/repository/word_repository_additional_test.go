package repository

import (
	"testing"

	"go.uber.org/zap"
)

func TestWordRepository_GetUserIDsByWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Add word request history for multiple users
	repo.AddWordRequestHistory(1100, "uniqueword")
	repo.AddWordRequestHistory(1200, "uniqueword")
	repo.AddWordRequestHistory(1300, "uniqueword")
	repo.AddWordRequestHistory(1100, "otherword") // Same user, different word

	// Get user IDs for the word
	userIDs, err := repo.GetUserIDsByWord("uniqueword")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) != 3 {
		t.Errorf("Expected 3 unique user IDs, got %d", len(userIDs))
	}
	// Verify all users are in the list
	userMap := make(map[int64]bool)
	for _, id := range userIDs {
		userMap[id] = true
	}
	if !userMap[1100] || !userMap[1200] || !userMap[1300] {
		t.Error("Expected all three user IDs in results")
	}
}

func TestWordRepository_GetUserIDsByWord_CaseInsensitive(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Add word request history with different cases
	repo.AddWordRequestHistory(1400, "CaseWord")
	repo.AddWordRequestHistory(1500, "caseword")
	repo.AddWordRequestHistory(1600, "CASEWORD")

	// Get user IDs (should be case-insensitive)
	userIDs, err := repo.GetUserIDsByWord("caseword")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) < 3 {
		t.Errorf("Expected at least 3 user IDs (case-insensitive), got %d", len(userIDs))
	}
}

func TestWordRepository_GetUserIDsByWord_NoUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	defer db.Close()

	repo := NewWordRepository(db, logger)

	// Get user IDs for word with no history
	userIDs, err := repo.GetUserIDsByWord("nonexistent")
	if err != nil {
		t.Fatalf("GetUserIDsByWord() error = %v", err)
	}
	if len(userIDs) != 0 {
		t.Errorf("Expected 0 user IDs, got %d", len(userIDs))
	}
}
