package service

import (
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestNewTrainingWorker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	aiService := (*ai.Service)(nil)
	wordRepo := (*repository.WordRepository)(nil)
	trainingCardRepo := (*repository.TrainingCardRepository)(nil)
	userCardRepo := (*repository.UserCardRepository)(nil)
	userRepo := (*repository.UserRepository)(nil)
	cbService := (*CircuitBreakerService)(nil)
	bot := (*tgbotapi.BotAPI)(nil)

	worker := NewTrainingWorker(
		aiService,
		wordRepo,
		trainingCardRepo,
		userCardRepo,
		userRepo,
		cbService,
		bot,
		0,
		5,
		4,
		0,
		"",
		logger,
	)
	_ = worker // Verify worker is created
}
