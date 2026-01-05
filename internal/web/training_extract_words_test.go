package web

import (
	"testing"

	"tgbot-skeleton/internal/models"
)

func TestExtractSessionWords(t *testing.T) {
	// Create a mock router for testing
	router := &Router{}

	// Test case 1: Empty queue
	queue := []*models.UserCardWithTraining{}
	words := router.extractSessionWords(queue, 0, models.DirectionENtoRU, []string{})
	if len(words) != 0 {
		t.Errorf("Expected empty words for empty queue, got %d", len(words))
	}

	// Test case 2: Single card
	trainingCard1 := &models.TrainingCard{
		WordCardID: 1,
		WordEN:     "test",
		WordRU:     "тест",
	}
	userCard1 := &models.UserCard{
		ID:     1,
		UserID: 1,
	}
	queue = []*models.UserCardWithTraining{
		{UserCard: *userCard1, TrainingCard: *trainingCard1},
	}
	words = router.extractSessionWords(queue, 0, models.DirectionENtoRU, []string{})
	if len(words) != 0 {
		t.Errorf("Expected empty words for single card, got %d", len(words))
	}

	// Test case 3: Multiple cards, DirectionENtoRU
	trainingCard2 := &models.TrainingCard{
		WordCardID: 2,
		WordEN:     "hello",
		WordRU:     "привет",
	}
	userCard2 := &models.UserCard{
		ID:     2,
		UserID: 1,
	}
	trainingCard3 := &models.TrainingCard{
		WordCardID: 3,
		WordEN:     "world",
		WordRU:     "мир",
	}
	userCard3 := &models.UserCard{
		ID:     3,
		UserID: 1,
	}
	queue = []*models.UserCardWithTraining{
		{UserCard: *userCard1, TrainingCard: *trainingCard1},
		{UserCard: *userCard2, TrainingCard: *trainingCard2},
		{UserCard: *userCard3, TrainingCard: *trainingCard3},
	}
	words = router.extractSessionWords(queue, 0, models.DirectionENtoRU, []string{})
	expectedCount := 2 // Should get WordRU from cards 2 and 3
	if len(words) != expectedCount {
		t.Errorf("Expected %d words, got %d: %v", expectedCount, len(words), words)
	}
	// Check that we got Russian words
	found := false
	for _, word := range words {
		if word == "привет" || word == "мир" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find Russian words, got: %v", words)
	}

	// Test case 4: DirectionRUtoEN
	words = router.extractSessionWords(queue, 0, models.DirectionRUtoEN, []string{})
	expectedCount = 2 // Should get WordEN from cards 2 and 3
	if len(words) != expectedCount {
		t.Errorf("Expected %d words for RUtoEN, got %d: %v", expectedCount, len(words), words)
	}

	// Test case 5: Cards with same WordCardID should be excluded
	trainingCard4 := &models.TrainingCard{
		WordCardID: 1, // Same as trainingCard1
		WordEN:     "test2",
		WordRU:     "тест2",
	}
	userCard4 := &models.UserCard{
		ID:     4,
		UserID: 1,
	}
	queue = []*models.UserCardWithTraining{
		{UserCard: *userCard1, TrainingCard: *trainingCard1},
		{UserCard: *userCard2, TrainingCard: *trainingCard2},
		{UserCard: *userCard4, TrainingCard: *trainingCard4}, // Same WordCardID as card 1
	}
	words = router.extractSessionWords(queue, 0, models.DirectionENtoRU, []string{})
	// Should only get word from card 2, not from card 4 (same WordCardID as current)
	expectedCount = 1
	if len(words) != expectedCount {
		t.Errorf("Expected %d words (excluding same WordCardID), got %d: %v", expectedCount, len(words), words)
	}

	// Test case 6: With recent correct answers
	words = router.extractSessionWords(queue, 0, models.DirectionENtoRU, []string{"привет"})
	// Should exclude "привет" from results
	if len(words) != 0 {
		t.Errorf("Expected 0 words (all excluded by recent correct), got %d: %v", len(words), words)
	}

	// Test case 7: CurrentIndex >= len(queue)
	words = router.extractSessionWords(queue, 10, models.DirectionENtoRU, []string{})
	if len(words) != 0 {
		t.Errorf("Expected empty words for invalid index, got %d", len(words))
	}

	// Test case 8: Duplicate words should be excluded
	trainingCard5 := &models.TrainingCard{
		WordCardID: 5,
		WordEN:     "duplicate",
		WordRU:     "дубликат",
	}
	userCard5 := &models.UserCard{
		ID:     5,
		UserID: 1,
	}
	trainingCard6 := &models.TrainingCard{
		WordCardID: 6,
		WordEN:     "duplicate",
		WordRU:     "дубликат", // Same word
	}
	userCard6 := &models.UserCard{
		ID:     6,
		UserID: 1,
	}
	queue = []*models.UserCardWithTraining{
		{UserCard: *userCard1, TrainingCard: *trainingCard1},
		{UserCard: *userCard5, TrainingCard: *trainingCard5},
		{UserCard: *userCard6, TrainingCard: *trainingCard6},
	}
	words = router.extractSessionWords(queue, 0, models.DirectionENtoRU, []string{})
	// Should only get one instance of "дубликат"
	expectedCount = 1
	if len(words) != expectedCount {
		t.Errorf("Expected %d words (duplicates excluded), got %d: %v", expectedCount, len(words), words)
	}
}
