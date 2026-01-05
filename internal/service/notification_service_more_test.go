package service

import (
	"testing"

	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestNotificationService_Stop_WithoutStart(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewNotificationService(
		(*tgbotapi.BotAPI)(nil),
		(*repository.UserRepository)(nil),
		(*repository.UserCardRepository)(nil),
		(*repository.NudgeRepository)(nil),
		(*repository.SessionRepository)(nil),
		logger,
	)

	// Stop without start should not panic
	service.Stop()
}
