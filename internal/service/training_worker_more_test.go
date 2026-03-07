package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestTrainingWorker_hasMissingData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, 0, 0, "", logger,
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
		nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, 0, 0, "", logger,
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
		logger,
	)

	// No word_cards without training_cards — processCards should return without error
	worker.processCards(context.Background())
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
		0, 0, 0, "", logger,
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
		0, 0, 0, "", logger,
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
		999, 0, 0, 0, "", logger,
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

func TestTrainingWorker_notifyAdminValidationError_SkipsWhenAdminIDZero(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, nil,
		0, 0, 0, 0, "", logger,
	)
	worker.notifyAdminValidationError("word", "validation error")
}

func TestTrainingWorker_notifyAdminValidationError_SkipsWhenBotNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, nil,
		12345, 0, 0, 0, "", logger,
	)
	worker.notifyAdminValidationError("word", "validation error")
}

func TestTrainingWorker_notifyAdminValidationError_SendsWhenBotSet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &mockTelegramClientWorker{}
	bot := newTestBotWorker(client)

	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, bot,
		888, 0, 0, 0, "", logger,
	)
	worker.notifyAdminValidationError("apple", "R1 sense=0 distractor_en[0] contains Cyrillic")

	if len(client.lastBody) == 0 {
		t.Error("expected bot.Send to be called with non-empty body")
	}
	if !bytes.Contains(client.lastBody, []byte("apple")) {
		t.Errorf("expected message to contain word, got body: %s", client.lastBody)
	}
}
