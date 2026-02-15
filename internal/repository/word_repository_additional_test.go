package repository

import (
	"testing"

	"go.uber.org/zap"
)

func TestWordRepository_GetUserIDsByWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	userRepo := NewUserRepository(db, logger)
	u1, _ := userRepo.GetOrCreateUser(1100)
	u2, _ := userRepo.GetOrCreateUser(1200)
	u3, _ := userRepo.GetOrCreateUser(1300)

	repo := NewWordRepository(db, logger)

	// Add word request history for multiple users
	repo.AddWordRequestHistory(u1.ID, "uniqueword")
	repo.AddWordRequestHistory(u2.ID, "uniqueword")
	repo.AddWordRequestHistory(u3.ID, "uniqueword")
	repo.AddWordRequestHistory(u1.ID, "otherword") // Same user, different word

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
	if !userMap[u1.ID] || !userMap[u2.ID] || !userMap[u3.ID] {
		t.Error("Expected all three user IDs in results")
	}
}

func TestWordRepository_GetUserIDsByWord_CaseInsensitive(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordTestDB(t)
	userRepo := NewUserRepository(db, logger)
	u1, _ := userRepo.GetOrCreateUser(1400)
	u2, _ := userRepo.GetOrCreateUser(1500)
	u3, _ := userRepo.GetOrCreateUser(1600)

	repo := NewWordRepository(db, logger)

	// Add word request history with different cases
	repo.AddWordRequestHistory(u1.ID, "CaseWord")
	repo.AddWordRequestHistory(u2.ID, "caseword")
	repo.AddWordRequestHistory(u3.ID, "CASEWORD")

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
