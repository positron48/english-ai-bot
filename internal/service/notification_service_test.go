package service

import (
	"testing"

	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

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

func TestNotificationService_CheckAndSendNotifications_NoUsers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	userCardRepo := repository.NewUserCardRepository(db, logger)
	nudgeRepo := repository.NewNudgeRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	service := NewNotificationService(nil, userRepo, userCardRepo, nudgeRepo, sessionRepo, logger)
	// No users in DB — checkAndSendNotifications should not panic and should return without sending
	service.checkAndSendNotifications()
}
