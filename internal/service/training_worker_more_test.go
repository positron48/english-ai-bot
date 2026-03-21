package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestTrainingWorker_hasMissingData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)

	t.Run("Word card with all data", func(t *testing.T) {
		pos := "noun"
		transcription := "/test/"
		definitionRU := "тест"
		wordCard := &models.WordCard{
			Word:          "test",
			Definition:    "test",
			POS:           &pos,
			Transcription: &transcription,
			DefinitionRU:  &definitionRU,
		}

		result := worker.hasMissingData(wordCard)
		if result {
			t.Error("hasMissingData() should return false when all data is present")
		}
	})

	t.Run("Word card missing POS", func(t *testing.T) {
		transcription := "/test/"
		definitionRU := "тест"
		wordCard := &models.WordCard{
			Word:          "test",
			Definition:    "test",
			POS:           nil,
			Transcription: &transcription,
			DefinitionRU:  &definitionRU,
		}

		result := worker.hasMissingData(wordCard)
		if !result {
			t.Error("hasMissingData() should return true when POS is missing")
		}
	})

	t.Run("Word card missing Transcription", func(t *testing.T) {
		pos := "noun"
		definitionRU := "тест"
		wordCard := &models.WordCard{
			Word:          "test",
			Definition:    "test",
			POS:           &pos,
			Transcription: nil,
			DefinitionRU:  &definitionRU,
		}

		result := worker.hasMissingData(wordCard)
		if !result {
			t.Error("hasMissingData() should return true when Transcription is missing")
		}
	})

	t.Run("Word card missing DefinitionRU", func(t *testing.T) {
		pos := "noun"
		transcription := "/test/"
		wordCard := &models.WordCard{
			Word:          "test",
			Definition:    "test",
			POS:           &pos,
			Transcription: &transcription,
			DefinitionRU:  nil,
		}

		result := worker.hasMissingData(wordCard)
		if !result {
			t.Error("hasMissingData() should return true when DefinitionRU is missing")
		}
	})

	t.Run("Word card missing all data", func(t *testing.T) {
		wordCard := &models.WordCard{
			Word:          "test",
			Definition:    "test",
			POS:           nil,
			Transcription: nil,
			DefinitionRU:  nil,
		}

		result := worker.hasMissingData(wordCard)
		if !result {
			t.Error("hasMissingData() should return true when all data is missing")
		}
	})
}

func TestTrainingWorker_Stop(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)

	// Stop should not panic
	worker.Stop()

	// Verify stopChan is closed
	select {
	case <-worker.stopChan:
		// Channel is closed, which is expected
	default:
		t.Error("stopChan should be closed after Stop()")
	}
}

func TestTrainingWorker_Start_ContextCancellation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	wordRepo := repository.NewWordRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)
	cbService := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 5, logger)

	aiService := ai.NewService("", "", "", "", logger)
	worker := NewTrainingWorker(
		aiService,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userRepo,
		nil,
		cbService,
		nil,
		0,
		1,
		1,
		100*time.Millisecond,
		"",
		config.DefaultLearningConfig(),
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())

	// Start worker in goroutine
	done := make(chan bool)
	go func() {
		worker.Start(ctx)
		done <- true
	}()

	// Cancel context after short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for worker to stop
	select {
	case <-done:
		// Worker stopped, which is expected
	case <-time.After(1 * time.Second):
		t.Error("Worker should stop when context is cancelled")
	}
}

func TestTrainingWorker_Start_StopChan(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	wordRepo := repository.NewWordRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)
	cbService := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 5, logger)

	aiService := ai.NewService("", "", "", "", logger)
	worker := NewTrainingWorker(
		aiService,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userRepo,
		nil,
		cbService,
		nil,
		0,
		1,
		1,
		100*time.Millisecond,
		"",
		config.DefaultLearningConfig(),
		logger,
	)

	ctx := context.Background()

	// Start worker in goroutine
	done := make(chan bool)
	go func() {
		worker.Start(ctx)
		done <- true
	}()

	// Stop worker after short delay
	time.Sleep(50 * time.Millisecond)
	worker.Stop()

	// Wait for worker to stop
	select {
	case <-done:
		// Worker stopped, which is expected
	case <-time.After(1 * time.Second):
		t.Error("Worker should stop when Stop() is called")
	}
}

// mockCircuitBreakerForWorker returns error from IsOpen to trigger processCards error path.
type mockCircuitBreakerForWorker struct {
	isOpenFunc       func() (bool, error)
	recordFailureErr error
	recordSuccessErr error
	getStateFunc     func() (bool, int, string, error)
}

func (m *mockCircuitBreakerForWorker) IsOpen() (bool, error) {
	if m.isOpenFunc != nil {
		return m.isOpenFunc()
	}
	return false, nil
}

func (m *mockCircuitBreakerForWorker) RecordFailure(_ string) error {
	return m.recordFailureErr
}
func (m *mockCircuitBreakerForWorker) RecordSuccess() error {
	return m.recordSuccessErr
}
func (m *mockCircuitBreakerForWorker) GetState() (bool, int, string, error) {
	if m.getStateFunc != nil {
		return m.getStateFunc()
	}
	return false, 0, "", nil
}

// TestTrainingWorker_processCards_CircuitBreakerIsOpenError covers processCards when IsOpen returns an error (log and return).
func TestTrainingWorker_processCards_CircuitBreakerIsOpenError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	wordRepo := repository.NewWordRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)
	mockCB := &mockCircuitBreakerForWorker{
		isOpenFunc: func() (bool, error) {
			return false, fmt.Errorf("mock circuit check error")
		},
	}
	worker := NewTrainingWorker(
		ai.NewService("", "", "", "", logger),
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userRepo,
		nil,
		mockCB,
		nil,
		0,
		2,
		1,
		time.Hour,
		"",
		config.DefaultLearningConfig(),
		logger,
	)
	worker.processCards(context.Background())
	// must not panic; processCards returns early after logging error
}

func TestTrainingWorker_processCards_CircuitBreakerOpen(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	wordRepo := repository.NewWordRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)
	cbRepo := repository.NewCircuitBreakerRepository(db, logger)
	cbService := NewCircuitBreakerService(cbRepo, 1, logger)
	_ = cbService.RecordFailure("test error") // opens circuit (threshold 1)

	aiService := ai.NewService("", "", "", "", logger)
	worker := NewTrainingWorker(
		aiService,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userRepo,
		nil,
		cbService,
		nil,
		0,
		2,
		1,
		time.Hour,
		"",
		config.DefaultLearningConfig(),
		logger,
	)

	// processCards should return early when circuit is open (no panic, no call to GetWordCardsWithoutTrainingCards)
	worker.processCards(context.Background())
}

func TestTrainingWorker_processCards_NoPendingCards(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)

	wordRepo := repository.NewWordRepository(db, logger)
	trainingCardRepo := repository.NewTrainingCardRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)
	cbService := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 5, logger)

	aiService := ai.NewService("", "", "", "", logger)
	worker := NewTrainingWorker(
		aiService,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userRepo,
		nil,
		cbService,
		nil,
		0,
		2,
		1,
		time.Hour,
		"",
		config.DefaultLearningConfig(),
		logger,
	)

	// No word_cards without training_cards — processCards should return without error
	worker.processCards(context.Background())
}

// TestTrainingWorker_processCards_ZeroWorkersFallback covers processCards when llmWorkers <= 0 (workers fallback to 1).
func TestTrainingWorker_processCards_ZeroWorkersFallback(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			content := `{"input_word":"one","lemma":"one","pos":"noun","transcription":"wʌn","definition_ru":"один"}`
			return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
		}
		content := `{"word_en":"one","lemma":"one","transcription":"","senses":[{"pos":"noun","word_ru":"один","meaning_en":"one","example_en":"","example_ru":"","distractors_ru":["два","три"],"distractors_en":["two","three"],"hint":""}]}`
		return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
	})
	worker, wordRepo, trainingCardRepo, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	logger, _ := zap.NewDevelopment()
	worker.cbService = NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger)
	worker.llmWorkers = 0 // triggers workers = 1 fallback
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "one", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	worker.processCards(context.Background())
	cards, _ := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if len(cards) < 1 {
		t.Errorf("expected 1 training card, got %d", len(cards))
	}
}

// TestTrainingWorker_processCards_WithPendingCards runs processCards when there is one pending word card;
// worker uses mock AI and creates training cards.
func TestTrainingWorker_processCards_WithPendingCards(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// GenerateResponse (fillWordCardData)
			content := `{"input_word":"processme","lemma":"processme","pos":"noun","transcription":"prəˈses","definition_ru":"обработать"}`
			return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
		}
		content := `{"word_en":"processme","lemma":"processme","transcription":"","senses":[{"pos":"noun","word_ru":"тест","meaning_en":"process me","example_en":"","example_ru":"","distractors_ru":["яблоко","груша"],"distractors_en":["orange","banana"],"hint":""}]}`
		return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
	})
	worker, wordRepo, trainingCardRepo, _, _, db, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	logger, _ := zap.NewDevelopment()
	worker.cbService = NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db.GetConnection(), logger), 5, logger)
	// Cover cardsToFetch when llmWorkers > batchSize
	worker.batchSize = 2
	worker.llmWorkers = 5
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "processme", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	worker.processCards(context.Background())
	cards, err := trainingCardRepo.GetTrainingCardsByWordCardID(cardID)
	if err != nil {
		t.Fatalf("GetTrainingCardsByWordCardID: %v", err)
	}
	if len(cards) < 1 {
		t.Errorf("expected at least 1 training card after processCards, got %d", len(cards))
	}
}

// TestTrainingWorker_processCards_RecordFailureError covers processCards when RecordFailure returns error.
func TestTrainingWorker_processCards_RecordFailureError(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			content := `{"input_word":"failcard","lemma":"failcard","pos":"noun","transcription":"","definition_ru":""}`
			return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
		}
		// Invalid JSON so processCard fails
		return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: "not json"}}}}), nil
	})
	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	worker.cbService = &mockCircuitBreakerForWorker{recordFailureErr: fmt.Errorf("record failure error")}
	worker.interval = time.Hour
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "failcard", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_ = cardID
	worker.processCards(context.Background())
	// Must not panic; logs "failed to record circuit breaker failure"
}

// TestTrainingWorker_processCards_RecordSuccessError covers processCards when RecordSuccess returns error.
func TestTrainingWorker_processCards_RecordSuccessError(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			content := `{"input_word":"ok","lemma":"ok","pos":"noun","transcription":"","definition_ru":"ок"}`
			return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
		}
		content := `{"word_en":"ok","lemma":"ok","transcription":"","senses":[{"pos":"noun","word_ru":"ок","meaning_en":"ok","example_en":"","example_ru":"","distractors_ru":["да","нет"],"distractors_en":["yes","no"],"hint":""}]}`
		return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
	})
	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	worker.cbService = &mockCircuitBreakerForWorker{recordSuccessErr: fmt.Errorf("record success error")}
	worker.interval = time.Hour
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "ok", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_ = cardID
	worker.processCards(context.Background())
	// Must not panic; logs "failed to record circuit breaker success"
}

// TestTrainingWorker_processCards_CardFailsCircuitOpensNotifyAdmin covers processCards when a card fails,
// circuit opens (IsOpen returns true after failure), and notifyAdmin is called.
func TestTrainingWorker_processCards_CardFailsCircuitOpensNotifyAdmin(t *testing.T) {
	callCount := 0
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			content := `{"input_word":"open","lemma":"open","pos":"verb","transcription":"","definition_ru":""}`
			return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
		}
		return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: "invalid"}}}}), nil
	})
	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	isOpenCallCount := 0
	mockCB := &mockCircuitBreakerForWorker{
		isOpenFunc: func() (bool, error) {
			isOpenCallCount++
			if isOpenCallCount >= 2 {
				return true, nil
			}
			return false, nil
		},
		getStateFunc: func() (bool, int, string, error) {
			return true, 1, "last error", nil
		},
	}
	worker.cbService = mockCB
	client := &mockTelegramClientWorker{}
	worker.bot = newTestBotWorker(client)
	worker.adminTelegramID = 999
	worker.interval = time.Hour
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "open", Definition: ""})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	_ = cardID
	worker.processCards(context.Background())
	if len(client.lastBody) == 0 {
		t.Error("expected notifyAdmin to send message when circuit opens")
	}
	if !bytes.Contains(client.lastBody, []byte("Circuit")) {
		t.Errorf("expected circuit breaker message, got body: %s", client.lastBody)
	}
}

// TestTrainingWorker_processCard_ValidationFailsNotifiesAdminWithBot covers processCard when validation fails
// and worker has bot and admin ID set, so notifyAdminValidationError sends.
func TestTrainingWorker_processCard_ValidationFailsNotifiesAdminWithBot(t *testing.T) {
	transport := rtFuncTW(func(req *http.Request) (*http.Response, error) {
		// Cyrillic in distractors_en triggers validation error
		content := `{"word_en":"run","lemma":"run","transcription":"","senses":[{"pos":"verb","word_ru":"бежать","meaning_en":"run","example_en":"","example_ru":"","distractors_ru":["идти","плыть"],"distractors_en":["бежать","to walk"],"hint":""}]}`
		return newJSONHTTPResponseTW(http.StatusOK, ai.ChatResponse{Choices: []ai.Choice{{Message: ai.Message{Content: content}}}}), nil
	})
	worker, wordRepo, _, _, _, _, cleanup := newTrainingWorker(t, transport)
	defer cleanup()
	worker.modelHigh = ""
	worker.adminTelegramID = 888
	client := &mockTelegramClientWorker{}
	worker.bot = newTestBotWorker(client)
	pos := "verb"
	cardID, err := wordRepo.UpsertWordCardLemma(&models.WordCard{Word: "run", Definition: "", POS: &pos})
	if err != nil {
		t.Fatalf("UpsertWordCardLemma: %v", err)
	}
	wordCard, _ := wordRepo.GetWordCardByID(cardID)
	err = worker.processCard(context.Background(), wordCard)
	if err != nil {
		t.Fatalf("processCard should return nil on validation failure: %v", err)
	}
	if len(client.lastBody) == 0 {
		t.Error("expected notifyAdminValidationError to send when bot set")
	}
	if !bytes.Contains(client.lastBody, []byte("run")) {
		t.Errorf("expected word in validation error notification, got: %s", client.lastBody)
	}
}

// TestTrainingWorker_getUsersForWord_NoUsers ensures empty slice when word has no request history.
func TestTrainingWorker_getUsersForWord_NoUsers(t *testing.T) {
	worker, _, _, _, _, _, cleanup := newTrainingWorker(t, nil)
	defer cleanup()
	users, err := worker.getUsersForWord("nonexistentword123")
	if err != nil {
		t.Fatalf("getUsersForWord: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users for unknown word, got %d", len(users))
	}
}

// TestTrainingWorker_getUsersForWord_MultipleUsers returns all distinct users who requested the word.
func TestTrainingWorker_getUsersForWord_MultipleUsers(t *testing.T) {
	worker, wordRepo, _, _, userRepo, _, cleanup := newTrainingWorker(t, nil)
	defer cleanup()
	u1, _ := userRepo.GetOrCreateUser(301)
	u2, _ := userRepo.GetOrCreateUser(302)
	word := "multiword"
	_ = wordRepo.AddWordRequestHistoryWithCard(u1.ID, word, nil, &word)
	_ = wordRepo.AddWordRequestHistoryWithCard(u2.ID, word, nil, &word)
	users, err := worker.getUsersForWord(word)
	if err != nil {
		t.Fatalf("getUsersForWord: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

// TestTrainingWorker_getUsersForWord_DeduplicateSameUser ensures the same user is returned once
// when they appear multiple times in word request history.
func TestTrainingWorker_getUsersForWord_DeduplicateSameUser(t *testing.T) {
	worker, wordRepo, _, _, userRepo, _, cleanup := newTrainingWorker(t, nil)
	defer cleanup()
	u, _ := userRepo.GetOrCreateUser(303)
	word := "dedupeword"
	_ = wordRepo.AddWordRequestHistoryWithCard(u.ID, word, nil, &word)
	_ = wordRepo.AddWordRequestHistoryWithCard(u.ID, word, nil, &word)
	users, err := worker.getUsersForWord(word)
	if err != nil {
		t.Fatalf("getUsersForWord: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 user (deduplicated), got %d", len(users))
	}
	if len(users) > 0 && users[0].ID != u.ID {
		t.Errorf("expected user ID %d, got %d", u.ID, users[0].ID)
	}
}

// mockTelegramClientWorker captures the last request body for assertions.
type mockTelegramClientWorker struct {
	lastBody []byte
}

func (c *mockTelegramClientWorker) Do(req *http.Request) (*http.Response, error) {
	c.lastBody, _ = io.ReadAll(req.Body)
	_ = req.Body.Close()
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"ok": true, "result": {"message_id": 1}}`)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func newTestBotWorker(client *mockTelegramClientWorker) *tgbotapi.BotAPI {
	bot := &tgbotapi.BotAPI{Token: "test", Client: client, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	return bot
}

func TestTrainingWorker_notifyAdmin_SkipsWhenAdminIDZero(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	cbService := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 5, logger)

	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, cbService, nil,
		0, // adminTelegramID
		0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdmin("some error")
	// No panic, no send (bot is nil anyway)
}

func TestTrainingWorker_notifyAdmin_SkipsWhenBotNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	cbService := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 5, logger)

	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, cbService, nil,
		12345, // adminTelegramID set
		0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdmin("some error")
	// No panic (bot is nil, we log and return)
}

func TestTrainingWorker_notifyAdmin_SendsWhenBotSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	cbService := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 5, logger)
	_ = cbService.RecordFailure("first")
	client := &mockTelegramClientWorker{}
	bot := newTestBotWorker(client)

	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, cbService, bot,
		999, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdmin("circuit open reason")

	if len(client.lastBody) == 0 {
		t.Error("expected bot.Send to be called with non-empty body")
	}
	// Body is form-encoded (e.g. text=...+Circuit+Breaker+...)
	if !bytes.Contains(client.lastBody, []byte("Circuit")) {
		t.Errorf("expected message to contain circuit breaker text, got body: %s", client.lastBody)
	}
}

// TestTrainingWorker_notifyAdmin_GetStateError covers notifyAdmin when GetState returns error.
func TestTrainingWorker_notifyAdmin_GetStateError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockCB := &mockCircuitBreakerForWorker{
		getStateFunc: func() (bool, int, string, error) {
			return false, 0, "", fmt.Errorf("get state failed")
		},
	}
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, mockCB, nil,
		12345, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdmin("some error")
	// No panic; logs error and returns
}

// TestTrainingWorker_notifyAdmin_SendFails covers notifyAdmin when bot.Send returns error.
func TestTrainingWorker_notifyAdmin_SendFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	cbService := NewCircuitBreakerService(repository.NewCircuitBreakerRepository(db, logger), 5, logger)
	_ = cbService.RecordFailure("first")
	client := &mockTelegramClientWorkerFail{}
	bot := &tgbotapi.BotAPI{Token: "test", Client: client, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, cbService, bot,
		999, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdmin("circuit open")
	// No panic; logs "failed to send admin notification"
}

func TestTrainingWorker_notifyAdminValidationError_SkipsWhenAdminIDZero(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, nil,
		0, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdminValidationError("word", "validation error")
}

func TestTrainingWorker_notifyAdminValidationError_SkipsWhenBotNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, nil,
		12345, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdminValidationError("word", "validation error")
}

func TestTrainingWorker_notifyAdminValidationError_SendsWhenBotSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &mockTelegramClientWorker{}
	bot := newTestBotWorker(client)

	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, bot,
		888, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdminValidationError("apple", "R1 sense=0 distractor_en[0] contains Cyrillic")

	if len(client.lastBody) == 0 {
		t.Error("expected bot.Send to be called with non-empty body")
	}
	if !bytes.Contains(client.lastBody, []byte("apple")) {
		t.Errorf("expected message to contain word, got body: %s", client.lastBody)
	}
}

// TestTrainingWorker_notifyAdminValidationError_LongMessageTruncated covers truncation when error > 500 chars.
func TestTrainingWorker_notifyAdminValidationError_LongMessageTruncated(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &mockTelegramClientWorker{}
	bot := newTestBotWorker(client)
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, bot,
		777, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	longError := strings.Repeat("x", 600)
	worker.notifyAdminValidationError("word", longError)
	if len(client.lastBody) == 0 {
		t.Error("expected bot.Send to be called")
	}
	// Message should contain truncated error (500 + "...")
	if !bytes.Contains(client.lastBody, []byte("...")) {
		t.Error("expected truncated validation error to end with ...")
	}
}

// mockTelegramClientWorkerFail returns error on Do to cover Send failure path.
type mockTelegramClientWorkerFail struct{}

func (c *mockTelegramClientWorkerFail) Do(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("telegram API error")
}

// TestTrainingWorker_notifyAdminValidationError_SendFails covers notifyAdminValidationError when bot.Send fails.
func TestTrainingWorker_notifyAdminValidationError_SendFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &mockTelegramClientWorkerFail{}
	bot := &tgbotapi.BotAPI{Token: "test", Client: client, Buffer: 1}
	bot.SetAPIEndpoint("http://example.com/bot%s/%s")
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, bot,
		666, 0, 0, 0, "", config.DefaultLearningConfig(), logger,
	)
	worker.notifyAdminValidationError("word", "validation error")
	// No panic; logs error and returns
}
