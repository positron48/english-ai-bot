package service

import (
	"testing"

	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestNewNotificationService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	bot := (*tgbotapi.BotAPI)(nil)
	userRepo := (*repository.UserRepository)(nil)
	userCardRepo := (*repository.UserCardRepository)(nil)
	nudgeRepo := (*repository.NudgeRepository)(nil)
	sessionRepo := (*repository.SessionRepository)(nil)

	service := NewNotificationService(bot, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)
	_ = service // Verify service is created
}
