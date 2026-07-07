package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func strPtr(s string) *string {
	return &s
}

func TestShowTrainingCard_SpellChallengeResponse(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	logger, _ := zap.NewDevelopment()
	router := NewRouter(logger, &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 1500, WrongAnswerDelaySeconds: 3}}, db, nil, nil, nil, nil)

	state := &WebTrainingState{
		SessionID: 99,
		Queue: []*models.TrainingQueueItem{{
			Type: "spell",
			Spell: &models.SpellChallenge{
				WordRU:          "шпионить",
				Prefix:          "to ",
				ShuffledLetters: []string{"s", "p", "y"},
				DisplayWord:     "to spy",
			},
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["type"] != "spell" {
		t.Fatalf("expected type=spell, got %v", payload["type"])
	}
	if payload["prefix"] != "to " {
		t.Fatalf("expected prefix=to, got %v", payload["prefix"])
	}
}

func TestShowTrainingCard_TypeChallengeResponse(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	logger, _ := zap.NewDevelopment()
	router := NewRouter(logger, &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 1500, WrongAnswerDelaySeconds: 3}}, db, nil, nil, nil, nil)

	state := &WebTrainingState{
		SessionID: 100,
		Queue: []*models.TrainingQueueItem{{
			Type: "type",
			TypeChallenge: &models.TypeChallenge{
				WordRU:      "говорить",
				DisplayWord: "to speak",
			},
		}},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload["type"] != "type" {
		t.Fatalf("expected type=type, got %v", payload["type"])
	}
	if payload["prefix"] != "to " {
		t.Fatalf("expected prefix=to, got %v", payload["prefix"])
	}
	if payload["hint_first_letter"] != "s" {
		t.Fatalf("expected first hint letter s, got %v", payload["hint_first_letter"])
	}
	if payload["hint_length"] != float64(5) {
		t.Fatalf("expected hint length 5, got %v", payload["hint_length"])
	}
}

func TestShowTrainingCard_CardTypeWithNilCard(t *testing.T) {
	db, _, _, _, _ := setupTrainingIntegrationTestDB(t)
	logger, _ := zap.NewDevelopment()
	router := NewRouter(logger, &config.Config{}, db, nil, nil, nil, nil)

	state := &WebTrainingState{
		Queue: []*models.TrainingQueueItem{{Type: "card", Card: nil}},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestShowTrainingCard_NormalCardResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, trainingCardRepo, _, _ := setupTrainingIntegrationTestDB(t)

	var wordCardID int64
	if err := db.QueryRow(
		"INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id",
		"term",
		"term",
	).Scan(&wordCardID); err != nil {
		t.Fatalf("create word card: %v", err)
	}

	if _, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID:    wordCardID,
		WordEN:        "term",
		SenseIndex:    0,
		WordRU:        "термин",
		MeaningEN:     "term meaning",
		Transcription: "[t3rm]",
		DistractorsRU: `["значение","понятие","идея"]`,
		DistractorsEN: `["meaning","concept","idea"]`,
		ExampleEN:     "Example sentence",
		DisplayWord:   strPtr("term"),
	}); err != nil {
		t.Fatalf("create training card: %v", err)
	}

	optionsService := service.NewOptionsService(trainingCardRepo, logger, "en")
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 1234, WrongAnswerDelaySeconds: 3}}
	router := NewRouter(logger, cfg, db, nil, nil, optionsService, nil)

	primaryCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{
			ID:        42,
			Direction: models.DirectionENtoRU,
		},
		TrainingCard: models.TrainingCard{
			WordCardID:    wordCardID,
			WordEN:        "term",
			WordRU:        "термин",
			Transcription: "[t3rm]",
			ExampleEN:     "Example sentence",
			DistractorsRU: `["значение","понятие","идея"]`,
			DisplayWord:   strPtr("term"),
		},
	}
	otherCard := &models.UserCardWithTraining{
		UserCard: models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: 999,
			WordEN:     "second",
			WordRU:     "второй",
		},
	}

	state := &WebTrainingState{
		UserID:    7,
		SessionID: 777,
		Queue: []*models.TrainingQueueItem{
			{Type: "card", Card: primaryCard},
			{Type: "card", Card: otherCard},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["user_card_id"] != float64(42) {
		t.Fatalf("expected user_card_id=42, got %v", payload["user_card_id"])
	}
	if payload["delay_ms"] != float64(1234) {
		t.Fatalf("expected delay_ms=1234, got %v", payload["delay_ms"])
	}
	if payload["display_word"] != "term" {
		t.Fatalf("expected display_word=term, got %v", payload["display_word"])
	}
	if payload["word_en"] != "term" {
		t.Fatalf("expected word_en=term, got %v", payload["word_en"])
	}
	if payload["example_en"] != "Example sentence" {
		t.Fatalf("expected example_en in response, got %v", payload["example_en"])
	}

	if len(state.Options) < 2 {
		t.Fatalf("expected generated options to be stored in state, got %v", state.Options)
	}
	if state.CorrectAnswer != "термин" {
		t.Fatalf("expected correct answer to be stored, got %q", state.CorrectAnswer)
	}
}

func TestExtractSessionWordsFromQueue_FiltersByPOSAndRecent(t *testing.T) {
	r := &Router{}
	verb := "verb"
	noun := "noun"

	current := &models.UserCardWithTraining{
		UserCard:     models.UserCard{Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{WordCardID: 1, POS: &verb},
	}
	queue := []*models.TrainingQueueItem{
		{Type: "card", Card: current},
		{Type: "card", Card: &models.UserCardWithTraining{TrainingCard: models.TrainingCard{WordCardID: 2, WordEN: "jump", POS: &verb, DisplayWord: strPtr("to jump")}}},
		{Type: "card", Card: &models.UserCardWithTraining{TrainingCard: models.TrainingCard{WordCardID: 3, WordEN: "cat", POS: &noun}}},
		{Type: "card", Card: &models.UserCardWithTraining{TrainingCard: models.TrainingCard{WordCardID: 4, WordEN: "jump", POS: &verb, DisplayWord: strPtr("to jump")}}},
	}

	words := r.extractSessionWordsFromQueue(queue, 0, current, []string{"to jump", "jump"})
	if len(words) != 0 {
		t.Fatalf("expected no words because recent answers excluded and others filtered, got %v", words)
	}

	words = r.extractSessionWordsFromQueue(queue, 0, current, nil)
	if len(words) != 1 || words[0] != "to jump" {
		t.Fatalf("expected one matching distractor 'to jump', got %v", words)
	}
}

func TestExtractSessionWords_FiltersAndDeduplicates(t *testing.T) {
	r := &Router{}
	verb := "verb"

	current := &models.UserCardWithTraining{
		UserCard: models.UserCard{Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{
			WordCardID: 1,
			WordEN:     "run",
			WordRU:     "бежать",
			POS:        &verb,
		},
	}

	queue := []*models.UserCardWithTraining{
		current,
		{TrainingCard: models.TrainingCard{WordCardID: 2, WordEN: "jump", WordRU: "прыгать", POS: &verb}},
		{TrainingCard: models.TrainingCard{WordCardID: 3, WordEN: "race", WordRU: "бежать", POS: &verb}},
		{TrainingCard: models.TrainingCard{WordCardID: 4, WordEN: "sprint", WordRU: "прыгать", POS: &verb}},
		{TrainingCard: models.TrainingCard{WordCardID: 5, WordEN: "walk", WordRU: "идти", POS: strPtr("noun")}},
	}

	words := r.extractSessionWords(queue, 0, current, []string{"прыгать"})
	if len(words) != 0 {
		t.Fatalf("expected all candidates filtered out, got %v", words)
	}

	words = r.extractSessionWords(queue, 0, current, nil)
	if len(words) != 1 || words[0] != "прыгать" {
		t.Fatalf("expected one deduplicated word 'прыгать', got %v", words)
	}
}

// extractSessionWordsFromQueue: currentIndex >= len(queue)
func TestExtractSessionWordsFromQueue_IndexOutOfRange(t *testing.T) {
	r := &Router{}
	queue := []*models.TrainingQueueItem{
		{Type: "card", Card: &models.UserCardWithTraining{TrainingCard: models.TrainingCard{WordCardID: 1, WordEN: "a", WordRU: "а"}}},
	}
	card := queue[0].Card
	words := r.extractSessionWordsFromQueue(queue, 1, card, nil)
	if len(words) != 0 {
		t.Fatalf("expected empty when currentIndex >= len(queue), got %v", words)
	}
	words = r.extractSessionWordsFromQueue(queue, 10, card, nil)
	if len(words) != 0 {
		t.Fatalf("expected empty when currentIndex 10 >= len(queue) 1, got %v", words)
	}
}

// showTrainingCard: CurrentIndex >= len(Queue) -> finishTrainingSession
func TestShowTrainingCard_SessionFinished(t *testing.T) {
	logger := zap.NewNop()
	db, _, _, _, sessionRepo := setupTrainingIntegrationTestDB(t)
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 3}}
	trainingService := service.NewTrainingService(nil, nil, sessionRepo, nil, config.DefaultLearningConfig(), logger)
	router := NewRouter(logger, cfg, db, trainingService, nil, nil, nil)
	router.webTrainingHandler = NewWebTrainingHandler(trainingService, nil, nil, sessionRepo, logger, 2000, 3)
	state := &WebTrainingState{
		UserID:       1,
		SessionID:    999,
		Queue:        []*models.TrainingQueueItem{},
		CurrentIndex: 0,
	}
	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["complete"] != true {
		t.Fatalf("expected complete=true, got %v", resp["complete"])
	}
}

// showTrainingCard: normal card with Direction RUtoEN (question text branch)
func TestShowTrainingCard_NormalCardRUtoEN(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _, trainingCardRepo, _, _ := setupTrainingIntegrationTestDB(t)
	var wordCardID int64
	if err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "run", "run").Scan(&wordCardID); err != nil {
		t.Fatalf("create word card: %v", err)
	}
	if _, err := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: wordCardID, WordEN: "run", SenseIndex: 0, WordRU: "бежать", MeaningEN: "run",
		DistractorsRU: `[]`, DistractorsEN: `[]`, DisplayWord: strPtr("run"),
	}); err != nil {
		t.Fatalf("create training card: %v", err)
	}
	optionsService := service.NewOptionsService(trainingCardRepo, logger, "en")
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 1000, WrongAnswerDelaySeconds: 3}}
	router := NewRouter(logger, cfg, db, nil, nil, optionsService, nil)
	card := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 1, Direction: models.DirectionRUtoEN},
		TrainingCard: models.TrainingCard{WordCardID: wordCardID, WordEN: "run", WordRU: "бежать", DisplayWord: strPtr("run"), DistractorsRU: `[]`, DistractorsEN: `[]`},
	}
	state := &WebTrainingState{
		UserID: 1, SessionID: 1, CurrentIndex: 0,
		Queue: []*models.TrainingQueueItem{{Type: "card", Card: card}},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/training/current", nil)
	w := httptest.NewRecorder()
	router.showTrainingCard(w, req, state)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !strings.Contains(payload["question"].(string), "Translate to English") {
		t.Fatalf("expected RUtoEN question, got %v", payload["question"])
	}
}

func TestHandleTrainingPrefetchNext_DoesNotMutateCurrentCardAnswerState(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, userRepo, trainingCardRepo, _, _ := setupTrainingIntegrationTestDB(t)
	user, _ := userRepo.GetOrCreateUser(700099)

	var firstWordID, secondWordID int64
	if err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "first", "first").Scan(&firstWordID); err != nil {
		t.Fatalf("create first word: %v", err)
	}
	if err := db.QueryRow("INSERT INTO word_cards (word, definition) VALUES ($1, $2) RETURNING id", "second", "second").Scan(&secondWordID); err != nil {
		t.Fatalf("create second word: %v", err)
	}
	firstCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: firstWordID, WordEN: "first", SenseIndex: 0, WordRU: "первый", MeaningEN: "first",
		DistractorsRU: `["а","б","в"]`, DistractorsEN: `["a","b","c"]`,
	})
	secondCardID, _ := trainingCardRepo.CreateTrainingCard(&models.TrainingCard{
		WordCardID: secondWordID, WordEN: "second", SenseIndex: 0, WordRU: "второй", MeaningEN: "second",
		DistractorsRU: `["г","д","е"]`, DistractorsEN: `["d","e","f"]`,
	})

	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 1000, WrongAnswerDelaySeconds: 3}}
	optionsService := service.NewOptionsService(trainingCardRepo, logger, "en")
	router := NewRouter(logger, cfg, db, nil, nil, optionsService, nil)

	first := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 101, UserID: user.ID, TrainingCardID: firstCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{ID: firstCardID, WordCardID: firstWordID, WordEN: "first", WordRU: "первый", DistractorsRU: `["а","б","в"]`},
	}
	second := &models.UserCardWithTraining{
		UserCard:     models.UserCard{ID: 102, UserID: user.ID, TrainingCardID: secondCardID, Direction: models.DirectionENtoRU},
		TrainingCard: models.TrainingCard{ID: secondCardID, WordCardID: secondWordID, WordEN: "second", WordRU: "второй", DistractorsRU: `["г","д","е"]`},
	}
	state := &WebTrainingState{
		UserID: user.ID, SessionID: 1, CurrentIndex: 0,
		Queue: []*models.TrainingQueueItem{{Type: "card", Card: first}, {Type: "card", Card: second}},
	}
	router.webTrainingHandler = &WebTrainingHandler{sessions: map[int64]*WebTrainingState{user.ID: state}}

	currentReq := setUserIDInContext(httptest.NewRequest(http.MethodGet, "/api/training/current", nil), user.ID)
	currentW := httptest.NewRecorder()
	router.showTrainingCard(currentW, currentReq, state)
	if currentW.Code != http.StatusOK {
		t.Fatalf("show first card status=%d body=%s", currentW.Code, currentW.Body.String())
	}
	currentOptions := append([]string(nil), state.Options...)
	currentCorrect := state.CorrectAnswer

	prefetchReq := setUserIDInContext(httptest.NewRequest(http.MethodPost, "/api/training/prefetch-next", nil), user.ID)
	prefetchW := httptest.NewRecorder()
	router.handleTrainingPrefetchNext(prefetchW, prefetchReq)
	if prefetchW.Code != http.StatusOK {
		t.Fatalf("prefetch status=%d body=%s", prefetchW.Code, prefetchW.Body.String())
	}
	if state.CurrentIndex != 0 {
		t.Fatalf("prefetch changed current index to %d", state.CurrentIndex)
	}
	if state.CorrectAnswer != currentCorrect {
		t.Fatalf("prefetch changed correct answer from %q to %q", currentCorrect, state.CorrectAnswer)
	}
	if len(state.Options) != len(currentOptions) {
		t.Fatalf("prefetch changed options length from %d to %d", len(currentOptions), len(state.Options))
	}
	for i := range currentOptions {
		if state.Options[i] != currentOptions[i] {
			t.Fatalf("prefetch changed current options: before=%v after=%v", currentOptions, state.Options)
		}
	}
	if state.PrefetchedCards == nil || state.PrefetchedCards[1] == nil {
		t.Fatalf("expected second card to be stored in prefetch cache")
	}

	state.CurrentIndex = 1
	nextW := httptest.NewRecorder()
	router.showTrainingCard(nextW, currentReq, state)
	if nextW.Code != http.StatusOK {
		t.Fatalf("show prefetched card status=%d body=%s", nextW.Code, nextW.Body.String())
	}
	if state.CorrectAnswer != "второй" {
		t.Fatalf("expected prefetched correct answer for second card, got %q", state.CorrectAnswer)
	}
	if state.PrefetchedCards[1] != nil {
		t.Fatalf("expected consumed prefetched card to be removed")
	}
}

// showTrainingFeedback: direct call with correct and incorrect (hint, example, delay_seconds)
func TestShowTrainingFeedback_Direct(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, userRepo, _, _, _ := setupTrainingHandlersTestDB(t)
	user, _ := userRepo.GetOrCreateUser(252525)
	cfg := &config.Config{Training: config.TrainingConfig{OptionsDelayMS: 2000, WrongAnswerDelaySeconds: 5}}
	router := &Router{config: cfg, userRepo: userRepo, logger: logger}
	state := &WebTrainingState{UserID: user.ID}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Correct answer: no delay_seconds
	router.showTrainingFeedback(w, req, state, true, "ok", "ok", models.TrainingCard{})
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["delay_seconds"] != nil {
		t.Errorf("correct answer should not have delay_seconds, got %v", resp["delay_seconds"])
	}

	// Incorrect with hint and example
	w = httptest.NewRecorder()
	router.showTrainingFeedback(w, req, state, false, "wrong", "right", models.TrainingCard{Hint: "hint", ExampleEN: "example"})
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["hint"] != "hint" || resp["example"] != "example" {
		t.Errorf("expected hint and example, got %v", resp)
	}
	if resp["delay_seconds"] != float64(5) {
		t.Errorf("expected delay_seconds=5, got %v", resp["delay_seconds"])
	}
}
