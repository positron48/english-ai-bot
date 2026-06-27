package service

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/testutil"
)

func setupOptionsServiceTestDB(t *testing.T) (*sql.DB, *repository.TrainingCardRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)

	return db, trainingCardRepo
}

func TestOptionsService_GenerateOptions_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	// Create a word card
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "generate", "to generate")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training cards for the same word
	distractorsEN, _ := json.Marshal([]string{"create", "make", "build"})
	card1 := &models.TrainingCard{
		WordCardID:    1,
		WordEN:        "generate",
		SenseIndex:    0,
		WordRU:        "генерировать",
		MeaningEN:     "to generate",
		DistractorsEN: string(distractorsEN),
	}
	card1ID, err := trainingCardRepo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger, "en")

	userCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{
			ID:               1,
			Direction:        models.DirectionRUtoEN,
			WrongAnswersJSON: "",
		},
		TrainingCard: models.TrainingCard{
			ID:            card1ID,
			WordCardID:    1,
			WordEN:        "generate",
			WordRU:        "генерировать",
			DistractorsEN: string(distractorsEN),
		},
	}

	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{}, make(map[string]bool), make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions() error = %v", err)
	}
	if len(options) < 2 {
		t.Errorf("Expected at least 2 options, got %d", len(options))
	}
	if correctAnswer != "generate" {
		t.Errorf("Expected correct answer 'generate', got %q", correctAnswer)
	}
	// Verify correct answer is in options
	found := false
	for _, opt := range options {
		if opt == correctAnswer {
			found = true
			break
		}
	}
	if !found {
		t.Error("Correct answer should be in options")
	}
}

func TestOptionsService_GenerateOptions_UsesCardDistractors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	// Create a word card
	_, err := db.Exec("INSERT INTO word_cards (word, definition) VALUES ($1, $2)", "spy", "to spy")
	if err != nil {
		t.Fatalf("Failed to create word card: %v", err)
	}

	// Create training card with specific distractors
	cardDistractors := []string{"watch", "observe", "monitor"}
	distractorsEN, _ := json.Marshal(cardDistractors)
	card1 := &models.TrainingCard{
		WordCardID:    1,
		WordEN:        "spy",
		SenseIndex:    0,
		WordRU:        "шпионить",
		MeaningEN:     "to spy",
		DistractorsEN: string(distractorsEN),
	}
	card1ID, err := trainingCardRepo.CreateTrainingCard(card1)
	if err != nil {
		t.Fatalf("Failed to create training card: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger, "en")

	userCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{
			ID:               1,
			Direction:        models.DirectionRUtoEN,
			WrongAnswersJSON: "",
		},
		TrainingCard: models.TrainingCard{
			ID:            card1ID,
			WordCardID:    1,
			WordEN:        "spy",
			WordRU:        "шпионить",
			DistractorsEN: string(distractorsEN),
		},
	}

	// Generate options without session words to ensure card distractors are used
	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{}, make(map[string]bool), make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions() error = %v", err)
	}
	if len(options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(options))
	}
	if correctAnswer != "spy" {
		t.Errorf("Expected correct answer 'spy', got %q", correctAnswer)
	}

	// Verify that at least one card distractor is in the options
	cardDistractorsSet := make(map[string]bool)
	for _, d := range cardDistractors {
		cardDistractorsSet[d] = true
	}

	foundCardDistractor := false
	for _, opt := range options {
		if opt != correctAnswer && cardDistractorsSet[opt] {
			foundCardDistractor = true
			break
		}
	}

	if !foundCardDistractor {
		t.Errorf("Expected at least one card distractor (%v) in options, got: %v", cardDistractors, options)
	}

	// For 4 options, we should have at least 1 card distractor
	// Let's count how many card distractors are used
	cardDistractorsUsed := 0
	for _, opt := range options {
		if opt != correctAnswer && cardDistractorsSet[opt] {
			cardDistractorsUsed++
		}
	}
	if cardDistractorsUsed < 1 {
		t.Errorf("Expected at least 1 card distractor to be used, found %d", cardDistractorsUsed)
	}
}

func TestOptionsService_GenerateOptions_DirectionENtoRU(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	var wordCardID int64
	err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "generate", "to generate").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	distractorsRU, _ := json.Marshal([]string{"создавать", "делать", "строить"})
	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "generate",
		SenseIndex:    0,
		WordRU:        "генерировать",
		MeaningEN:     "to generate",
		DistractorsRU: string(distractorsRU),
	}
	cardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger, "en")
	userCard := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 1, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{ID: cardID, WordCardID: wordCardID, WordEN: "generate", WordRU: "генерировать", DistractorsRU: string(distractorsRU)},
	}

	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{}, make(map[string]bool), make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions: %v", err)
	}
	if correctAnswer != "генерировать" {
		t.Errorf("expected correct answer WordRU for EN->RU, got %q", correctAnswer)
	}
	found := false
	for _, o := range options {
		if o == correctAnswer {
			found = true
			break
		}
	}
	if !found {
		t.Error("correct answer should be in options")
	}
	if len(options) < 2 {
		t.Errorf("expected at least 2 options, got %d", len(options))
	}
}

func TestOptionsService_GenerateOptions_InvalidDistractorsJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	var wordCardID int64
	err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "test", "def").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "test",
		SenseIndex:    0,
		WordRU:        "тест",
		MeaningEN:     "test",
		DistractorsEN: `{invalid json`, // triggers Warn and empty slice
		DistractorsRU: `["другой","третий"]`,
	}
	cardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger, "en")
	userCard := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 1, Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{ID: cardID, WordCardID: wordCardID, WordEN: "test", WordRU: "тест", DistractorsEN: `{invalid json`, DistractorsRU: `["другой","третий"]`},
	}

	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{}, make(map[string]bool), make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions: %v", err)
	}
	// Fallback distractors should fill; correct answer must be present
	if correctAnswer != "test" {
		t.Errorf("expected correct answer 'test', got %q", correctAnswer)
	}
	found := false
	for _, o := range options {
		if o == correctAnswer {
			found = true
			break
		}
	}
	if !found {
		t.Error("correct answer should be in options")
	}
	if len(options) < 2 {
		t.Errorf("expected at least 2 options (with fallback), got %d", len(options))
	}
}

func TestOptionsService_GenerateOptions_DisplayWordForRUtoEN(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	var wordCardID int64
	err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "spy", "to spy").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	displayWord := "to spy"
	distractorsEN, _ := json.Marshal([]string{"to watch", "to see", "to look"})
	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "spy",
		SenseIndex:    0,
		WordRU:        "шпионить",
		MeaningEN:     "to spy",
		DisplayWord:   &displayWord,
		DistractorsEN: string(distractorsEN),
	}
	cardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger, "en")
	userCard := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 1, Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{ID: cardID, WordCardID: wordCardID, WordEN: "spy", WordRU: "шпионить", DisplayWord: &displayWord, DistractorsEN: string(distractorsEN)},
	}

	_, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{}, make(map[string]bool), make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions: %v", err)
	}
	if correctAnswer != "to spy" {
		t.Errorf("expected correct answer DisplayWord 'to spy' for RU->EN, got %q", correctAnswer)
	}
}

func TestOptionsService_GenerateOptions_SessionWordsExcluded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	var wordCardID int64
	err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "apple", "fruit").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	distractorsEN, _ := json.Marshal([]string{"orange", "banana", "grape"})
	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "apple",
		SenseIndex:    0,
		WordRU:        "яблоко",
		MeaningEN:     "apple",
		DistractorsEN: string(distractorsEN),
	}
	cardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger, "en")
	userCard := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 1, Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{ID: cardID, WordCardID: wordCardID, WordEN: "apple", WordRU: "яблоко", DistractorsEN: string(distractorsEN)},
	}

	// Exclude "orange" as session word (English) so it should not appear as distractor
	sessionWordENs := map[string]bool{"orange": true}
	options, _, err := service.GenerateOptions(userCard, 4, []string{}, sessionWordENs, make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions: %v", err)
	}
	for _, o := range options {
		if o == "orange" {
			t.Error("session word 'orange' should be excluded from options")
		}
	}
	if len(options) < 2 {
		t.Errorf("expected at least 2 options, got %d", len(options))
	}
}

func TestOptionsService_GenerateOptions_WrongAnswersIncluded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	var wordCardID int64
	err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "generate", "to generate").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	distractorsEN, _ := json.Marshal([]string{"create", "make", "build"})
	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "generate",
		SenseIndex:    0,
		WordRU:        "генерировать",
		MeaningEN:     "to generate",
		DistractorsEN: string(distractorsEN),
	}
	cardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	wrongAnswersJSON := `[{"option":"produce"},{"option":"create"}]`
	service := NewOptionsService(trainingCardRepo, logger, "en")
	userCard := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 1, Direction: models.DirectionRUtoEN, WrongAnswersJSON: wrongAnswersJSON},
		TrainingCard: models.TrainingCard{ID: cardID, WordCardID: wordCardID, WordEN: "generate", WordRU: "генерировать", DistractorsEN: string(distractorsEN)},
	}

	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{}, make(map[string]bool), make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions: %v", err)
	}
	if correctAnswer != "generate" {
		t.Errorf("expected correct answer 'generate', got %q", correctAnswer)
	}
	// Wrong answers (produce, create) can appear in pool; at least one non-correct option expected
	if len(options) < 2 {
		t.Errorf("expected at least 2 options, got %d", len(options))
	}
}

func TestOptionsService_GenerateOptions_SpanishVerbStripsToPrefix(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, trainingCardRepo := setupOptionsServiceTestDB(t)

	var wordCardID int64
	err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "hablar", "hablar").Scan(&wordCardID)
	if err != nil {
		t.Fatalf("insert word_card: %v", err)
	}
	displayWord := "to hablar"
	pos := "verbo"
	distractorsEN, _ := json.Marshal([]string{"to comer", "to vivir", "to correr"})
	card := &models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "hablar",
		SenseIndex:    0,
		WordRU:        "говорить",
		MeaningEN:     "hablar",
		POS:           &pos,
		DisplayWord:   &displayWord,
		DistractorsEN: string(distractorsEN),
	}
	cardID, err := trainingCardRepo.CreateTrainingCard(card)
	if err != nil {
		t.Fatalf("CreateTrainingCard: %v", err)
	}

	service := NewOptionsService(trainingCardRepo, logger, "es")
	userCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{ID: 1, Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{
			ID: cardID, WordCardID: wordCardID, WordEN: "hablar", WordRU: "говорить",
			POS: &pos, DisplayWord: &displayWord, DistractorsEN: string(distractorsEN),
		},
	}

	options, correctAnswer, err := service.GenerateOptions(userCard, 4, []string{}, make(map[string]bool), make(map[string]bool))
	if err != nil {
		t.Fatalf("GenerateOptions: %v", err)
	}
	if correctAnswer != "hablar" {
		t.Fatalf("correct answer = %q, want hablar", correctAnswer)
	}
	for _, o := range options {
		if strings.HasPrefix(o, "to ") {
			t.Fatalf("option %q must not use English to-prefix for Spanish", o)
		}
	}
}
