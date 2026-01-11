package service

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestTrainingWorker_Stop_WithoutStart(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	worker := NewTrainingWorker(
		(*ai.Service)(nil),
		(*repository.WordRepository)(nil),
		(*repository.TrainingCardRepository)(nil),
		(*repository.UserCardRepository)(nil),
		(*repository.UserRepository)(nil),
		(*CircuitBreakerService)(nil),
		(*tgbotapi.BotAPI)(nil),
		0,
		10,
		4,
		5*time.Minute,
		"",
		logger,
	)

	// Stop without start should not panic
	worker.Stop()
}
