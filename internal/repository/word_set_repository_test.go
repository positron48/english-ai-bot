package repository

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWordSetTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewWordSetRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)
	if repo == nil {
		t.Error("NewWordSetRepository() should not return nil")
	}
}

func TestWordSetRepository_CreateWordSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)

	t.Run("Create word set without category", func(t *testing.T) {
		wordSet := &models.WordSet{
			Title:       "Test Set",
			IsPublished: true,
			SortOrder:   1,
		}

		id, err := repo.CreateWordSet(wordSet)
		if err != nil {
			t.Fatalf("CreateWordSet() error = %v", err)
		}
		if id == 0 {
			t.Error("CreateWordSet() should return non-zero ID")
		}
	})

	t.Run("Create word set with category", func(t *testing.T) {
		categoryRepo := NewWordSetCategoryRepository(db, logger)
		category := &models.WordSetCategory{
			Name:      "Test Category",
			SortOrder: 1,
		}
		categoryID, err := categoryRepo.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		wordSet := &models.WordSet{
			CategoryID:  &categoryID,
			Title:       "Test Set with Category",
			IsPublished: true,
			SortOrder:   1,
		}

		id, err := repo.CreateWordSet(wordSet)
		if err != nil {
			t.Fatalf("CreateWordSet() error = %v", err)
		}
		if id == 0 {
			t.Error("CreateWordSet() should return non-zero ID")
		}
	})

	t.Run("Create unpublished word set", func(t *testing.T) {
		wordSet := &models.WordSet{
			Title:       "Unpublished Set",
			IsPublished: false,
			SortOrder:   1,
		}

		id, err := repo.CreateWordSet(wordSet)
		if err != nil {
			t.Fatalf("CreateWordSet() error = %v", err)
		}
		if id == 0 {
			t.Error("CreateWordSet() should return non-zero ID")
		}
	})

	t.Run("Create word set with description and preferred_pos", func(t *testing.T) {
		desc := "Test description"
		pos := "noun"
		wordSet := &models.WordSet{
			Title:        "Set with Details",
			Description:  &desc,
			PreferredPOS: &pos,
			IsPublished:  true,
			SortOrder:    1,
		}

		id, err := repo.CreateWordSet(wordSet)
		if err != nil {
			t.Fatalf("CreateWordSet() error = %v", err)
		}
		if id == 0 {
			t.Error("CreateWordSet() should return non-zero ID")
		}
	})
}

func TestWordSetRepository_GetWordSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)

	t.Run("Get non-existent word set", func(t *testing.T) {
		wordSet, err := repo.GetWordSet(99999)
		if err != nil {
			t.Fatalf("GetWordSet() error = %v", err)
		}
		if wordSet != nil {
			t.Error("GetWordSet() should return nil for non-existent word set")
		}
	})

	t.Run("Get existing word set", func(t *testing.T) {
		createdSet := &models.WordSet{
			Title:       "Get Test Set",
			IsPublished: true,
			SortOrder:   1,
		}
		id, err := repo.CreateWordSet(createdSet)
		if err != nil {
			t.Fatalf("Failed to create word set: %v", err)
		}

		wordSet, err := repo.GetWordSet(id)
		if err != nil {
			t.Fatalf("GetWordSet() error = %v", err)
		}
		if wordSet == nil {
			t.Fatal("GetWordSet() should not return nil")
		}
		if wordSet.ID != id {
			t.Errorf("Expected ID %d, got %d", id, wordSet.ID)
		}
		if wordSet.Title != "Get Test Set" {
			t.Errorf("Expected title 'Get Test Set', got %q", wordSet.Title)
		}
		if !wordSet.IsPublished {
			t.Error("Expected IsPublished to be true")
		}
	})
}

func TestWordSetRepository_ListWordSets(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)

	t.Run("List word sets when empty", func(t *testing.T) {
		sets, err := repo.ListWordSets(nil, 10, 0)
		if err != nil {
			t.Fatalf("ListWordSets() error = %v", err)
		}
		if len(sets) != 0 {
			t.Errorf("Expected 0 word sets, got %d", len(sets))
		}
	})

	t.Run("List all published word sets", func(t *testing.T) {
		// Create published sets
		set1 := &models.WordSet{
			Title:       "Published Set 1",
			IsPublished: true,
			SortOrder:   1,
		}
		_, err := repo.CreateWordSet(set1)
		if err != nil {
			t.Fatalf("Failed to create set 1: %v", err)
		}

		set2 := &models.WordSet{
			Title:       "Published Set 2",
			IsPublished: true,
			SortOrder:   2,
		}
		_, err = repo.CreateWordSet(set2)
		if err != nil {
			t.Fatalf("Failed to create set 2: %v", err)
		}

		// Create unpublished set
		set3 := &models.WordSet{
			Title:       "Unpublished Set",
			IsPublished: false,
			SortOrder:   3,
		}
		_, err = repo.CreateWordSet(set3)
		if err != nil {
			t.Fatalf("Failed to create set 3: %v", err)
		}

		sets, err := repo.ListWordSets(nil, 10, 0)
		if err != nil {
			t.Fatalf("ListWordSets() error = %v", err)
		}
		if len(sets) < 2 {
			t.Errorf("Expected at least 2 published word sets, got %d", len(sets))
		}
	})

	t.Run("List word sets with includeUnpublished", func(t *testing.T) {
		sets, err := repo.ListWordSets(nil, 10, 0, true)
		if err != nil {
			t.Fatalf("ListWordSets() error = %v", err)
		}
		if len(sets) < 3 {
			t.Errorf("Expected at least 3 word sets (including unpublished), got %d", len(sets))
		}
	})

	t.Run("List word sets by category", func(t *testing.T) {
		categoryRepo := NewWordSetCategoryRepository(db, logger)
		category := &models.WordSetCategory{
			Name:      "Filter Category",
			SortOrder: 1,
		}
		categoryID, err := categoryRepo.CreateCategory(category)
		if err != nil {
			t.Fatalf("Failed to create category: %v", err)
		}

		set := &models.WordSet{
			CategoryID:  &categoryID,
			Title:       "Set in Category",
			IsPublished: true,
			SortOrder:   1,
		}
		_, err = repo.CreateWordSet(set)
		if err != nil {
			t.Fatalf("Failed to create set: %v", err)
		}

		sets, err := repo.ListWordSets(&categoryID, 10, 0)
		if err != nil {
			t.Fatalf("ListWordSets() error = %v", err)
		}
		if len(sets) < 1 {
			t.Errorf("Expected at least 1 word set in category, got %d", len(sets))
		}
	})
}

func TestWordSetRepository_UpdateWordSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)

	t.Run("Update word set", func(t *testing.T) {
		wordSet := &models.WordSet{
			Title:       "Original Title",
			IsPublished: true,
			SortOrder:   1,
		}
		id, err := repo.CreateWordSet(wordSet)
		if err != nil {
			t.Fatalf("Failed to create word set: %v", err)
		}

		updatedSet := &models.WordSet{
			ID:          id,
			Title:       "Updated Title",
			IsPublished: false,
			SortOrder:   2,
		}
		err = repo.UpdateWordSet(updatedSet)
		if err != nil {
			t.Fatalf("UpdateWordSet() error = %v", err)
		}

		// Verify update
		got, err := repo.GetWordSet(id)
		if err != nil {
			t.Fatalf("GetWordSet() error = %v", err)
		}
		if got.Title != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got %q", got.Title)
		}
		if got.IsPublished {
			t.Error("Expected IsPublished to be false")
		}
		if got.SortOrder != 2 {
			t.Errorf("Expected sort order 2, got %d", got.SortOrder)
		}
	})
}

func TestWordSetRepository_DeleteWordSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)

	t.Run("Delete word set", func(t *testing.T) {
		wordSet := &models.WordSet{
			Title:       "Set to Delete",
			IsPublished: true,
			SortOrder:   1,
		}
		id, err := repo.CreateWordSet(wordSet)
		if err != nil {
			t.Fatalf("Failed to create word set: %v", err)
		}

		err = repo.DeleteWordSet(id)
		if err != nil {
			t.Fatalf("DeleteWordSet() error = %v", err)
		}

		// Verify deletion
		got, err := repo.GetWordSet(id)
		if err != nil {
			t.Fatalf("GetWordSet() error = %v", err)
		}
		if got != nil {
			t.Error("Word set should be deleted")
		}
	})
}

func TestWordSetRepository_GetWordSetProgress(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)
	knowledgeRepo := NewUserWordKnowledgeRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12349)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word set
	wordSet := &models.WordSet{
		Title:       "Progress Test Set",
		IsPublished: true,
		SortOrder:   1,
	}
	setID, err := repo.CreateWordSet(wordSet)
	if err != nil {
		t.Fatalf("Failed to create word set: %v", err)
	}

	t.Run("Get progress for empty word set", func(t *testing.T) {
		progress, err := repo.GetWordSetProgress(setID, user.ID)
		if err != nil {
			t.Fatalf("GetWordSetProgress() error = %v", err)
		}
		if progress == nil {
			t.Fatal("GetWordSetProgress() should not return nil")
		}
		if progress.TotalWords != 0 {
			t.Errorf("Expected 0 total words, got %d", progress.TotalWords)
		}
		if progress.ProgressPercent != 0.0 {
			t.Errorf("Expected 0.0 progress percent, got %f", progress.ProgressPercent)
		}
	})

	t.Run("Get progress with words", func(t *testing.T) {
		// Create word cards
		wordCard1 := &models.WordCard{
			Word:       "word1",
			Definition: "definition1",
		}
		wordCardID1, err := wordRepo.UpsertWordCardLemma(wordCard1)
		if err != nil {
			t.Fatalf("Failed to create word card 1: %v", err)
		}

		wordCard2 := &models.WordCard{
			Word:       "word2",
			Definition: "definition2",
		}
		wordCardID2, err := wordRepo.UpsertWordCardLemma(wordCard2)
		if err != nil {
			t.Fatalf("Failed to create word card 2: %v", err)
		}

		// Add words to set
		err = repo.SetWordSetItems(setID, []int64{wordCardID1, wordCardID2})
		if err != nil {
			t.Fatalf("Failed to set word set items: %v", err)
		}

		// Mark one word as known
		err = knowledgeRepo.MarkKnown(user.ID, wordCardID1)
		if err != nil {
			t.Fatalf("Failed to mark word as known: %v", err)
		}

		progress, err := repo.GetWordSetProgress(setID, user.ID)
		if err != nil {
			t.Fatalf("GetWordSetProgress() error = %v", err)
		}
		if progress == nil {
			t.Fatal("GetWordSetProgress() should not return nil")
		}
		if progress.TotalWords != 2 {
			t.Errorf("Expected 2 total words, got %d", progress.TotalWords)
		}
		if progress.KnownWords != 1 {
			t.Errorf("Expected 1 known word, got %d", progress.KnownWords)
		}
		if progress.UnknownWords != 1 {
			t.Errorf("Expected 1 unknown word, got %d", progress.UnknownWords)
		}
	})
}

func TestWordSetRepository_SetWordSetItems(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)

	// Create word set
	wordSet := &models.WordSet{
		Title:       "Items Test Set",
		IsPublished: true,
		SortOrder:   1,
	}
	setID, err := repo.CreateWordSet(wordSet)
	if err != nil {
		t.Fatalf("Failed to create word set: %v", err)
	}

	t.Run("Set word set items", func(t *testing.T) {
		// Create word cards
		wordCard1 := &models.WordCard{
			Word:       "item1",
			Definition: "definition1",
		}
		wordCardID1, err := wordRepo.UpsertWordCardLemma(wordCard1)
		if err != nil {
			t.Fatalf("Failed to create word card 1: %v", err)
		}

		wordCard2 := &models.WordCard{
			Word:       "item2",
			Definition: "definition2",
		}
		wordCardID2, err := wordRepo.UpsertWordCardLemma(wordCard2)
		if err != nil {
			t.Fatalf("Failed to create word card 2: %v", err)
		}

		err = repo.SetWordSetItems(setID, []int64{wordCardID1, wordCardID2})
		if err != nil {
			t.Fatalf("SetWordSetItems() error = %v", err)
		}

		// Verify items were set
		progress, err := repo.GetWordSetProgress(setID, 1)
		if err != nil {
			t.Fatalf("GetWordSetProgress() error = %v", err)
		}
		if progress.TotalWords != 2 {
			t.Errorf("Expected 2 total words, got %d", progress.TotalWords)
		}
	})

	t.Run("Replace word set items", func(t *testing.T) {
		// Create new word card
		wordCard3 := &models.WordCard{
			Word:       "item3",
			Definition: "definition3",
		}
		wordCardID3, err := wordRepo.UpsertWordCardLemma(wordCard3)
		if err != nil {
			t.Fatalf("Failed to create word card 3: %v", err)
		}

		// Replace items with just the new one
		err = repo.SetWordSetItems(setID, []int64{wordCardID3})
		if err != nil {
			t.Fatalf("SetWordSetItems() error = %v", err)
		}

		// Verify items were replaced
		progress, err := repo.GetWordSetProgress(setID, 1)
		if err != nil {
			t.Fatalf("GetWordSetProgress() error = %v", err)
		}
		if progress.TotalWords != 1 {
			t.Errorf("Expected 1 total word, got %d", progress.TotalWords)
		}
	})

	t.Run("Set empty word set items", func(t *testing.T) {
		err = repo.SetWordSetItems(setID, []int64{})
		if err != nil {
			t.Fatalf("SetWordSetItems() error = %v", err)
		}

		// Verify items were cleared
		progress, err := repo.GetWordSetProgress(setID, 1)
		if err != nil {
			t.Fatalf("GetWordSetProgress() error = %v", err)
		}
		if progress.TotalWords != 0 {
			t.Errorf("Expected 0 total words, got %d", progress.TotalWords)
		}
	})
}

func TestWordSetRepository_GetWordSetWords(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWordSetTestDB(t)
	defer db.Close()

	repo := NewWordSetRepository(db, logger)
	userRepo := NewUserRepository(db, logger)
	wordRepo := NewWordRepository(db, logger)

	// Create user
	user, err := userRepo.GetOrCreateUser(12350)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create word set
	wordSet := &models.WordSet{
		Title:       "Words Test Set",
		IsPublished: true,
		SortOrder:   1,
	}
	setID, err := repo.CreateWordSet(wordSet)
	if err != nil {
		t.Fatalf("Failed to create word set: %v", err)
	}

	t.Run("Get words from empty set", func(t *testing.T) {
		words, err := repo.GetWordSetWords(setID, user.ID)
		if err != nil {
			t.Fatalf("GetWordSetWords() error = %v", err)
		}
		if len(words) != 0 {
			t.Errorf("Expected 0 words, got %d", len(words))
		}
	})

	t.Run("Get words from set with words", func(t *testing.T) {
		// Create word cards
		wordCard1 := &models.WordCard{
			Word:       "getword1",
			Definition: "definition1",
		}
		wordCardID1, err := wordRepo.UpsertWordCardLemma(wordCard1)
		if err != nil {
			t.Fatalf("Failed to create word card 1: %v", err)
		}

		wordCard2 := &models.WordCard{
			Word:       "getword2",
			Definition: "definition2",
		}
		wordCardID2, err := wordRepo.UpsertWordCardLemma(wordCard2)
		if err != nil {
			t.Fatalf("Failed to create word card 2: %v", err)
		}

		// Add words to set
		err = repo.SetWordSetItems(setID, []int64{wordCardID1, wordCardID2})
		if err != nil {
			t.Fatalf("Failed to set word set items: %v", err)
		}

		words, err := repo.GetWordSetWords(setID, user.ID)
		if err != nil {
			t.Fatalf("GetWordSetWords() error = %v", err)
		}
		if len(words) != 2 {
			t.Errorf("Expected 2 words, got %d", len(words))
		}
		if words[0].Word != "getword1" && words[0].Word != "getword2" {
			t.Errorf("Unexpected word: %q", words[0].Word)
		}
	})
}
