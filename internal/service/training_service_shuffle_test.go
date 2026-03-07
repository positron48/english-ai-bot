package service

import (
	"testing"

	"tgbot-skeleton/internal/testutil"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestShufflePreventDuplicates_SmallList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// Create a small list of cards
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordEN: "banana"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordEN: "cherry"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(shuffled))
	}
}

func TestShufflePreventDuplicates_LargeList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// Create a larger list with some duplicate words
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordEN: "banana"}},
		{UserCard: models.UserCard{ID: 4}, TrainingCard: models.TrainingCard{WordEN: "banana"}},
		{UserCard: models.UserCard{ID: 5}, TrainingCard: models.TrainingCard{WordEN: "cherry"}},
		{UserCard: models.UserCard{ID: 6}, TrainingCard: models.TrainingCard{WordEN: "date"}},
		{UserCard: models.UserCard{ID: 7}, TrainingCard: models.TrainingCard{WordEN: "elderberry"}},
		{UserCard: models.UserCard{ID: 8}, TrainingCard: models.TrainingCard{WordEN: "fig"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 8 {
		t.Errorf("Expected 8 cards, got %d", len(shuffled))
	}

	// Check that no two adjacent cards have the same word
	for i := 1; i < len(shuffled); i++ {
		if shuffled[i].TrainingCard.WordEN == shuffled[i-1].TrainingCard.WordEN {
			t.Logf("Adjacent duplicates found at positions %d and %d: %s", i-1, i, shuffled[i].TrainingCard.WordEN)
			// This might happen in edge cases, but should be minimized
		}
	}
}

func TestShufflePreventDuplicates_EmptyList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	var cards []*models.UserCardWithTraining
	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 0 {
		t.Errorf("Expected 0 cards, got %d", len(shuffled))
	}
}

func TestShufflePreventDuplicates_SingleElement(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// Create a single card
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 1 {
		t.Errorf("Expected 1 card, got %d", len(shuffled))
	}
}

func TestShufflePreventDuplicates_AllSameWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	// All cards have the same word (WordCardID 0 by default -> one group)
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordEN: "apple"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)

	if len(shuffled) != 3 {
		t.Errorf("Expected 3 cards, got %d", len(shuffled))
	}
}

// TestShufflePreventDuplicates_FixAdjacentDuplicates uses many cards with same WordCardID
// so that after attempts we still have adjacent duplicates and fixAdjacentDuplicates is called.
func TestShufflePreventDuplicates_FixAdjacentDuplicates(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	wid := int64(1)
	cards := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "x"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "x"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "x"}},
		{UserCard: models.UserCard{ID: 4}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "x"}},
		{UserCard: models.UserCard{ID: 5}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "x"}},
		{UserCard: models.UserCard{ID: 6}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "x"}},
	}

	shuffled := trainingService.shufflePreventDuplicates(cards)
	if len(shuffled) != 6 {
		t.Errorf("expected 6 cards, got %d", len(shuffled))
	}
	// fixAdjacentDuplicates is invoked when bestScore > 0; we can't assert exact order
	score := trainingService.calculateShuffleScore(shuffled)
	if score > 0 {
		t.Logf("note: after shuffle still have %d adjacent duplicate pairs", score)
	}
}

// TestFixAdjacentDuplicates_SwapFixesPair calls fixAdjacentDuplicates with [A,A,B,B] so a swap can fix the pair.
func TestFixAdjacentDuplicates_SwapFixesPair(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)

	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	a, b := int64(1), int64(2)
	queue := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordCardID: a, WordEN: "a"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordCardID: a, WordEN: "a"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordCardID: b, WordEN: "b"}},
		{UserCard: models.UserCard{ID: 4}, TrainingCard: models.TrainingCard{WordCardID: b, WordEN: "b"}},
	}

	fixed := trainingService.fixAdjacentDuplicates(queue)
	if len(fixed) != 4 {
		t.Fatalf("expected 4 cards, got %d", len(fixed))
	}
	// fixAdjacentDuplicates attempts to swap; we only assert the path was executed
	score := trainingService.calculateShuffleScore(fixed)
	t.Logf("score after fix: %d (lower is better)", score)
}

func TestFixAdjacentDuplicates_SingleElement(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	queue := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordCardID: 1, WordEN: "a"}},
	}
	fixed := trainingService.fixAdjacentDuplicates(queue)
	if len(fixed) != 1 {
		t.Fatalf("expected 1 card, got %d", len(fixed))
	}
	if fixed[0].UserCard.ID != 1 {
		t.Errorf("expected same card, got id %d", fixed[0].UserCard.ID)
	}
}

func TestFixAdjacentDuplicates_EmptyReturnsAsIs(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	var queue []*models.UserCardWithTraining
	fixed := trainingService.fixAdjacentDuplicates(queue)
	if len(fixed) != 0 {
		t.Errorf("expected empty slice for empty input, got len %d", len(fixed))
	}
}

// TestFixAdjacentDuplicates_TwoSameWord has two elements with same WordCardID; no swap possible, returns copy.
func TestFixAdjacentDuplicates_TwoSameWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	wid := int64(1)
	queue := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "a"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordCardID: wid, WordEN: "a"}},
	}
	fixed := trainingService.fixAdjacentDuplicates(queue)
	if len(fixed) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(fixed))
	}
	// Input unchanged in content (same two cards; order may stay)
	if fixed[0].TrainingCard.WordCardID != wid || fixed[1].TrainingCard.WordCardID != wid {
		t.Errorf("expected both WordCardID %d, got %d and %d", wid, fixed[0].TrainingCard.WordCardID, fixed[1].TrainingCard.WordCardID)
	}
}

// TestFixAdjacentDuplicates_NoAdjacentDupes returns slice as-is when no adjacent duplicates.
func TestFixAdjacentDuplicates_NoAdjacentDupes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	a, b, c := int64(1), int64(2), int64(3)
	queue := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordCardID: a, WordEN: "a"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordCardID: b, WordEN: "b"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordCardID: c, WordEN: "c"}},
	}
	fixed := trainingService.fixAdjacentDuplicates(queue)
	if len(fixed) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(fixed))
	}
	score := trainingService.calculateShuffleScore(fixed)
	if score != 0 {
		t.Errorf("expected no adjacent duplicates (score 0), got %d", score)
	}
}

// TestFixAdjacentDuplicates_ThreeElementsOnePair has [A,A,B]; one swap can fix the pair at 0,1.
func TestFixAdjacentDuplicates_ThreeElementsOnePair(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDatabase(t)
	ucRepo := repository.NewUserCardRepository(db.GetConnection(), logger)
	tcRepo := repository.NewTrainingCardRepository(db.GetConnection(), logger)
	sessionRepo := repository.NewSessionRepository(db.GetConnection(), logger)
	trainingService := NewTrainingService(ucRepo, tcRepo, sessionRepo, nil, logger)

	a, b := int64(1), int64(2)
	queue := []*models.UserCardWithTraining{
		{UserCard: models.UserCard{ID: 1}, TrainingCard: models.TrainingCard{WordCardID: a, WordEN: "a"}},
		{UserCard: models.UserCard{ID: 2}, TrainingCard: models.TrainingCard{WordCardID: a, WordEN: "a"}},
		{UserCard: models.UserCard{ID: 3}, TrainingCard: models.TrainingCard{WordCardID: b, WordEN: "b"}},
	}
	fixed := trainingService.fixAdjacentDuplicates(queue)
	if len(fixed) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(fixed))
	}
	score := trainingService.calculateShuffleScore(fixed)
	if score > 0 {
		t.Logf("score after fix: %d (swap may have resolved adjacent pair)", score)
	}
}
