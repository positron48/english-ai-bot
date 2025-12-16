package service

import (
	"context"
	"fmt"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// NotificationService handles daily training notifications
type NotificationService struct {
	bot              *tgbotapi.BotAPI
	userRepo         *repository.UserRepository
	userCardRepo     *repository.UserCardRepository
	nudgeRepo        *repository.NudgeRepository
	sessionRepo      *repository.SessionRepository
	logger           *zap.Logger
	stopChan         chan struct{}
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	bot *tgbotapi.BotAPI,
	userRepo *repository.UserRepository,
	userCardRepo *repository.UserCardRepository,
	nudgeRepo *repository.NudgeRepository,
	sessionRepo *repository.SessionRepository,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		bot:          bot,
		userRepo:     userRepo,
		userCardRepo: userCardRepo,
		nudgeRepo:    nudgeRepo,
		sessionRepo:  sessionRepo,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}
}

// Start starts the notification scheduler
func (s *NotificationService) Start(ctx context.Context) {
	s.logger.Info("starting notification service")

	// Check immediately on start
	go s.checkAndSendNotifications()

	// Then check every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("notification service stopped")
			return
		case <-s.stopChan:
			s.logger.Info("notification service stopped")
			return
		case <-ticker.C:
			s.checkAndSendNotifications()
		}
	}
}

// Stop stops the notification service
func (s *NotificationService) Stop() {
	close(s.stopChan)
}

// checkAndSendNotifications checks and sends notifications to users
func (s *NotificationService) checkAndSendNotifications() {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		s.logger.Error("failed to get users", zap.Error(err))
		return
	}

	now := time.Now()
	
	for _, user := range users {
		// Parse user's timezone
		loc, err := time.LoadLocation(user.Timezone)
		if err != nil {
			s.logger.Warn("invalid timezone, using UTC",
				zap.Int64("user_id", user.ID),
				zap.String("timezone", user.Timezone),
			)
			loc = time.UTC
		}

		// Get current time in user's timezone
		userNow := now.In(loc)
		
		// Parse preferred training time
		preferredTime, err := time.Parse("15:04", user.PreferredTrainingTime)
		if err != nil {
			s.logger.Warn("invalid preferred time, using 19:00",
				zap.Int64("user_id", user.ID),
				zap.String("preferred_time", user.PreferredTrainingTime),
			)
			preferredTime, _ = time.Parse("15:04", "19:00")
		}

		// Check if it's time to send notification
		// Send if current hour matches preferred hour
		if userNow.Hour() != preferredTime.Hour() {
			continue
		}

		// Check if we should send notification
		if err := s.sendNotificationIfNeeded(user, userNow); err != nil {
			s.logger.Error("failed to send notification",
				zap.Int64("user_id", user.ID),
				zap.Error(err),
			)
		}
	}
}

// sendNotificationIfNeeded sends a notification to a user if conditions are met
func (s *NotificationService) sendNotificationIfNeeded(user *models.User, userNow time.Time) error {
	localDate := userNow.Format("2006-01-02")

	// Check if already sent nudge today
	hasNudge, err := s.nudgeRepo.HasNudgeToday(user.ID, localDate)
	if err != nil {
		return fmt.Errorf("failed to check nudge: %w", err)
	}
	if hasNudge {
		return nil // Already sent today
	}

	// Check if user already trained today
	sessionCount, err := s.sessionRepo.GetTodaySessionCount(user.ID, localDate)
	if err != nil {
		return fmt.Errorf("failed to check session count: %w", err)
	}
	if sessionCount > 0 {
		return nil // Already trained today
	}

	// Get due count
	dueCount, err := s.userCardRepo.GetDueCount(user.ID, userNow)
	if err != nil {
		return fmt.Errorf("failed to get due count: %w", err)
	}

	if dueCount == 0 {
		return nil // No cards due
	}

	// Estimate time
	avgSecondsPerCard := 15 // Default estimate
	estimatedMinutes := (dueCount * avgSecondsPerCard) / 60
	if estimatedMinutes < 1 {
		estimatedMinutes = 1
	}

	// Send notification
	message := fmt.Sprintf(
		"К повторению: %d карточек (~%d минут). Начать тренировку?",
		dueCount,
		estimatedMinutes,
	)

	// Create inline keyboard with "Start" button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Начать", "train_start"),
		),
	)

	msg := tgbotapi.NewMessage(user.TelegramID, message)
	msg.ReplyMarkup = keyboard

	sentMsg, err := s.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Record nudge
	nudge := &models.TrainingNudge{
		UserID:         user.ID,
		LocalDate:      localDate,
		DueCountAtSend: dueCount,
	}
	msgID := sentMsg.MessageID
	nudge.MessageID = &msgID

	if _, err := s.nudgeRepo.CreateNudge(nudge); err != nil {
		s.logger.Warn("failed to record nudge", zap.Error(err))
	}

	s.logger.Info("sent training notification",
		zap.Int64("user_id", user.ID),
		zap.Int64("telegram_id", user.TelegramID),
		zap.Int("due_count", dueCount),
	)

	return nil
}

