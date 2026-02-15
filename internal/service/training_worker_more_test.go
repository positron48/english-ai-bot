package service

import (
	"context"
	"testing"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestTrainingWorker_hasMissingData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		nil, nil, nil, nil, nil, nil, nil, 0, 0, 0, 0, "", logger,
	)

	t.Run("Word card with all data", func(t *testing.T) {
		pos := "noun"
		transcription := "/test/"
		definitionRU := "тест"
		wordCard := &models.WordCard{
			Word:        "test",
			Definition:  "test",
			POS:         &pos,
			Transcription: &transcription,
			DefinitionRU: &definitionRU,
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
			Word:        "test",
			Definition:  "test",
			POS:         nil,
			Transcription: &transcription,
			DefinitionRU: &definitionRU,
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
			Word:        "test",
			Definition:  "test",
			POS:         &pos,
			Transcription: nil,
			DefinitionRU: &definitionRU,
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
			Word:        "test",
			Definition:  "test",
			POS:         &pos,
			Transcription: &transcription,
			DefinitionRU: nil,
		}

		result := worker.hasMissingData(wordCard)
		if !result {
			t.Error("hasMissingData() should return true when DefinitionRU is missing")
		}
	})

	t.Run("Word card missing all data", func(t *testing.T) {
		wordCard := &models.WordCard{
			Word:        "test",
			Definition:  "test",
			POS:         nil,
			Transcription: nil,
			DefinitionRU: nil,
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
		nil, nil, nil, nil, nil, nil, nil, 0, 0, 0, 0, "", logger,
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
