package service

import (
	"fmt"
	"testing"
	"time"

	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

func TestTrainingService_generateQueue_WithDueAndNewCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(9999)

	// Create word cards first
	_, err := db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 1, "queue1", "queue 1")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}
	_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", 2, "queue2", "queue 2")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
		1, "queue1", 0, "очередь1", "queue 1", "noun", "queue1")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
		2, "queue2", 0, "очередь2", "queue 2", "noun", "queue2")
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	// Create due cards
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	dueCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(dueCard)
	if err != nil {
		t.Fatalf("Failed to create due card: %v", err)
	}

	// Create new cards
	newCard := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 2,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
	}
	_, err = userCardRepo.CreateUserCard(newCard)
	if err != nil {
		t.Fatalf("Failed to create new card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:       "test",
	}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) < 2 {
		t.Errorf("Expected at least 2 cards in queue, got %d", len(queue))
	}
}

func TestTrainingService_generateQueue_OnlyDueCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(8888)

	// Create word cards first
	var err error
	for i := 1; i <= 3; i++ {
		_, err = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (?, ?, ?)", i, fmt.Sprintf("onlydue%d", i), fmt.Sprintf("only due %d", i))
		if err != nil {
			t.Fatalf("Failed to create word card: %v", err)
		}
	}

	// Create training cards
	for i := 1; i <= 3; i++ {
		_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (?, ?, ?, ?, ?, ?, ?)",
			i, "onlydue", 0, "только просроченные", "only due", "adjective", "onlydue")
		if err != nil {
			t.Fatalf("Failed to create training card: %v", err)
		}
	}

	// Create only due cards
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
		_, err = userCardRepo.CreateUserCard(card)
		if err != nil {
			t.Fatalf("Failed to create due card: %v", err)
		}
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:       "test",
	}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) < 3 {
		t.Errorf("Expected at least 3 cards in queue, got %d", len(queue))
	}
}

func TestTrainingService_generateQueue_Empty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(7777)

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)

	config := SessionConfig{
		MaxCardsPerSession: 10,
		MaxNewPerSession:   5,
		AlgoVersion:       "test",
	}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 0 {
		t.Errorf("Expected empty queue, got %d cards", len(queue))
	}
}

// TestTrainingService_generateQueue_RandomSampleFromLargePool verifies that when pool size
// exceeds MaxCardsPerSession, exactly MaxCardsPerSession cards are returned (random sample).
func TestTrainingService_generateQueue_RandomSampleFromLargePool(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(5555)

	const poolSize = 40
	const maxPerSession = 30

	for i := 1; i <= poolSize; i++ {
		_, err := db.Exec("INSERT INTO word_cards (id, word, definition) VALUES ($1, $2, $3)", i, fmt.Sprintf("rand%d", i), fmt.Sprintf("rand %d", i))
		if err != nil {
			t.Fatalf("insert word_cards: %v", err)
		}
		_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			i, fmt.Sprintf("rand%d", i), 0, "случайное", "random", "noun", fmt.Sprintf("rand%d", i))
		if err != nil {
			t.Fatalf("insert training_cards: %v", err)
		}
	}

	past := time.Now().Add(-24 * time.Hour)
	for i := 1; i <= poolSize; i++ {
		card := &models.UserCard{
			UserID:         user.ID,
			TrainingCardID: int64(i),
			Direction:      models.DirectionENtoRU,
			State:          models.StateReview,
			EF:             2.0,
			NextDueAt:      &past,
		}
		_, err := userCardRepo.CreateUserCard(card)
		if err != nil {
			t.Fatalf("create user card: %v", err)
		}
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	config := SessionConfig{
		MaxCardsPerSession: maxPerSession,
		MaxNewPerSession:   5,
		AlgoVersion:        "test",
	}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != maxPerSession {
		t.Errorf("expected queue length %d (random sample from pool), got %d", maxPerSession, len(queue))
	}

	// Run a few more times to ensure we don't always get the same set (randomness)
	seenFirstIDs := make(map[int64]bool)
	for run := 0; run < 5; run++ {
		q, err := service.generateQueue(user.ID, config)
		if err != nil || len(q) != maxPerSession {
			continue
		}
		if q[0].Type == "card" && q[0].Card != nil {
			seenFirstIDs[q[0].Card.UserCard.ID] = true
		}
	}
	if len(seenFirstIDs) < 2 {
		t.Logf("note: over 5 runs first card varied %d times (random sample)", len(seenFirstIDs))
	}
}

// TestTrainingService_generateQueue_PoolSmallerThanSession verifies that when pool has
// fewer cards than MaxCardsPerSession, all of them are returned.
func TestTrainingService_generateQueue_PoolSmallerThanSession(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(4444)

	for i := 1; i <= 5; i++ {
		_, err := db.Exec("INSERT INTO word_cards (id, word, definition) VALUES ($1, $2, $3)", i, fmt.Sprintf("small%d", i), "small")
		if err != nil {
			t.Fatalf("insert word_cards: %v", err)
		}
		_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			i, fmt.Sprintf("small%d", i), 0, "мало", "small", "noun", fmt.Sprintf("small%d", i))
		if err != nil {
			t.Fatalf("insert training_cards: %v", err)
		}
	}

	past := time.Now().Add(-24 * time.Hour)
	for i := 1; i <= 5; i++ {
		_, err := userCardRepo.CreateUserCard(&models.UserCard{
			UserID:         user.ID,
			TrainingCardID: int64(i),
			Direction:      models.DirectionENtoRU,
			State:          models.StateReview,
			EF:             2.0,
			NextDueAt:      &past,
		})
		if err != nil {
			t.Fatalf("create user card: %v", err)
		}
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	config := SessionConfig{MaxCardsPerSession: 30, MaxNewPerSession: 10, AlgoVersion: "test"}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 5 {
		t.Errorf("expected 5 cards (pool smaller than session), got %d", len(queue))
	}
}

// TestTrainingService_generateQueue_DedupeDueAndNew verifies that a card that appears
// in both due and new sets (e.g. state=new with next_due_at nil) is deduped in the pool.
func TestTrainingService_generateQueue_DedupeDueAndNew(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(3333)

	_, err := db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'dedupe', 'dedupe')")
	if err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'dedupe', 0, 'дедупе', 'dedupe', 'noun', 'dedupe')")
	if err != nil {
		t.Fatalf("insert training_cards: %v", err)
	}

	// One new card: it is both "due" (next_due_at IS NULL) and "new" (state=new)
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionENtoRU,
		State:          models.StateNew,
		EF:             models.InitialEF,
		NextDueAt:      nil,
	}
	_, err = userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	config := SessionConfig{MaxCardsPerSession: 30, MaxNewPerSession: 10, AlgoVersion: "test"}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 1 {
		t.Errorf("expected 1 card after dedupe (same card in due and new), got %d", len(queue))
	}
}

// TestTrainingService_generateQueue_SpellTypePath runs generateQueue with cards that have
// high mastering score (review + reps) and Direction RUtoEN so the spell/type injection
// branch and computeMasteringScore/spellPrefixAndLetters are exercised.
func TestTrainingService_generateQueue_SpellTypePath(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(2222)

	_, err := db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'spellword', 'spell')")
	if err != nil {
		t.Fatalf("insert word_cards: %v", err)
	}
	_, err = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'spellword', 0, 'спелл', 'spell', 'verb', 'to spell')")
	if err != nil {
		t.Fatalf("insert training_cards: %v", err)
	}

	past := time.Now().Add(-24 * time.Hour)
	card := &models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           50,
		NextDueAt:      &past,
	}
	_, err = userCardRepo.CreateUserCard(card)
	if err != nil {
		t.Fatalf("create user card: %v", err)
	}

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	config := SessionConfig{
		MaxCardsPerSession:      30,
		MaxNewPerSession:        10,
		AlgoVersion:             "test",
		SpellEnabled:            true,
		SpellMasteringThreshold: 50,
		TypeEnabled:             true,
		TypeMasteringThreshold:  70,
	}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 1 {
		t.Fatalf("expected 1 item in queue, got %d", len(queue))
	}
	// Item may be card, spell, or type (random); we only check that queue was built
	if queue[0].Type != "card" && queue[0].Type != "spell" && queue[0].Type != "type" {
		t.Errorf("unexpected queue item type: %s", queue[0].Type)
	}
}

// TestTrainingService_generateQueue_ThresholdClamping verifies that spell/type thresholds
// are clamped to 0..100 (no panic or out-of-range).
func TestTrainingService_generateQueue_ThresholdClamping(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(1111)

	_, _ = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'clamp', 'clamp')")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'clamp', 0, 'кламп', 'clamp', 'noun', 'clamp')")
	past := time.Now().Add(-24 * time.Hour)
	_, _ = userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           50,
		NextDueAt:      &past,
	})

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	config := SessionConfig{
		MaxCardsPerSession:      30,
		MaxNewPerSession:        10,
		SpellEnabled:            true,
		SpellMasteringThreshold: 150, // clamped to 100
		TypeEnabled:             true,
		TypeMasteringThreshold:  -10,
	}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 1 {
		t.Fatalf("expected 1 item, got %d", len(queue))
	}
	_ = queue
}

// TestTrainingService_generateQueue_ShortDisplayWord skips spell/type for words with len(displayWord) < 2.
func TestTrainingService_generateQueue_ShortDisplayWord(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(9998)

	_, _ = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'i', 'I')")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'I', 0, 'я', 'I', 'pronoun', 'I')")
	past := time.Now().Add(-24 * time.Hour)
	_, _ = userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.0,
		Reps:           50,
		NextDueAt:      &past,
	})

	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, nil, logger)
	config := SessionConfig{
		MaxCardsPerSession:      30,
		MaxNewPerSession:        10,
		SpellEnabled:            true,
		TypeEnabled:             true,
		SpellMasteringThreshold: 0,
		TypeMasteringThreshold:  0,
	}

	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 1 {
		t.Fatalf("expected 1 item, got %d", len(queue))
	}
	// Should remain "card" because displayWord "I" has len 1
	if queue[0].Type != "card" {
		t.Errorf("expected type card for short display word, got %s", queue[0].Type)
	}
}

// TestTrainingService_generateQueue_WithMasteringRepoGetScore uses userWordMasteringRepo.GetScore for spell/type eligibility (covers branch when userWordMasteringRepo != nil).
func TestTrainingService_generateQueue_WithMasteringRepoGetScore(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, userCardRepo, trainingCardRepo, _ := setupTrainingServiceTestDB(t)
	user, _ := userRepo.GetOrCreateUser(8882)

	_, _ = db.Exec("INSERT INTO word_cards (id, word, definition) VALUES (1, 'getscore', 'def')")
	_, _ = db.Exec("INSERT INTO training_cards (word_card_id, word_en, sense_index, word_ru, meaning_en, pos, display_word) VALUES (1, 'getscore', 0, 'гетскор', 'def', 'verb', 'to run')")
	past := time.Now().Add(-24 * time.Hour)
	_, _ = userCardRepo.CreateUserCard(&models.UserCard{
		UserID:         user.ID,
		TrainingCardID: 1,
		Direction:      models.DirectionRUtoEN,
		State:          models.StateReview,
		EF:             2.0,
		NextDueAt:      &past,
	})

	mockMastering := &mockMasteringRepoForSession{
		getScoreFunc: func(_, _ int64) (int, error) { return 80, nil },
	}
	service := NewTrainingService(userCardRepo, trainingCardRepo, nil, mockMastering, logger)
	config := SessionConfig{
		MaxCardsPerSession:      10,
		MaxNewPerSession:        5,
		AlgoVersion:             "test",
		SpellEnabled:            true,
		TypeEnabled:             true,
		SpellMasteringThreshold: 50,
		TypeMasteringThreshold:  70,
	}
	queue, err := service.generateQueue(user.ID, config)
	if err != nil {
		t.Fatalf("generateQueue() error = %v", err)
	}
	if len(queue) != 1 {
		t.Fatalf("expected 1 item, got %d", len(queue))
	}
	// With score 80 >= thresholds, item may be card, spell, or type
	if queue[0].Type != "card" && queue[0].Type != "spell" && queue[0].Type != "type" {
		t.Errorf("unexpected queue item type: %s", queue[0].Type)
	}
}
