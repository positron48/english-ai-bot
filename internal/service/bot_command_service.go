package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// BotCommandService handles telegram bot commands
type BotCommandService struct {
	bot          *tgbotapi.BotAPI
	userRepo     *repository.UserRepository
	logger       *zap.Logger
	helpMessage  string
	startMessage string
	unknownCommandMessage string
}

// NewBotCommandService creates a new bot command service
func NewBotCommandService(
	bot *tgbotapi.BotAPI,
	userRepo *repository.UserRepository,
	logger *zap.Logger,
	helpMessage string,
	startMessage string,
	unknownCommandMessage string,
) *BotCommandService {
	return &BotCommandService{
		bot:                   bot,
		userRepo:              userRepo,
		logger:                logger,
		helpMessage:           helpMessage,
		startMessage:          startMessage,
		unknownCommandMessage: unknownCommandMessage,
	}
}

// HandleUpdate processes a telegram update
func (s *BotCommandService) HandleUpdate(update tgbotapi.Update) {
	s.logger.Debug("handling telegram update",
		zap.Int("update_id", update.UpdateID),
		zap.Bool("has_message", update.Message != nil),
		zap.Bool("has_callback", update.CallbackQuery != nil),
	)

	// Handle callback queries (button clicks)
	if update.CallbackQuery != nil {
		s.handleCallbackQuery(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		s.logger.Debug("update has no message, skipping")
		return
	}

	message := update.Message
	telegramID := message.From.ID

	s.logger.Debug("processing message",
		zap.Int64("telegram_id", telegramID),
		zap.String("text", message.Text),
		zap.Bool("is_command", message.IsCommand()),
	)

	// Handle commands FIRST - before any other processing
	if message.IsCommand() {
		s.logger.Info("processing command", zap.String("command", message.Command()))
		s.handleCommand(message)
		return
	}

	// Handle start with parameter (e.g., /start unsubscribe)
	if message.Text != "" && strings.HasPrefix(message.Text, "/start ") {
		param := strings.TrimPrefix(message.Text, "/start ")
		if param == "unsubscribe" {
			s.handleUnsubscribe(telegramID, message.Chat.ID)
			return
		}
	}

	// If it's not a command, we don't handle it here
	// It should be handled by AI service or other handlers
	s.logger.Debug("message is not a command, skipping")
}

// handleCallbackQuery processes callback queries from inline buttons
func (s *BotCommandService) handleCallbackQuery(query *tgbotapi.CallbackQuery) {
	data := query.Data
	telegramID := query.From.ID
	chatID := query.Message.Chat.ID

	switch data {
	case "notification_unsubscribe":
		s.handleUnsubscribe(telegramID, chatID)
		// Answer callback query
		callback := tgbotapi.NewCallback(query.ID, "Вы отписаны от уведомлений")
		s.bot.Request(callback)
	case "train_start":
		// This is handled elsewhere, but we can acknowledge it here
		callback := tgbotapi.NewCallback(query.ID, "")
		s.bot.Request(callback)
	default:
		// Unknown callback
		callback := tgbotapi.NewCallback(query.ID, "")
		s.bot.Request(callback)
	}
}

// handleCommand processes bot commands
func (s *BotCommandService) handleCommand(message *tgbotapi.Message) {
	command := message.Command()
	telegramID := message.From.ID
	commandLower := strings.ToLower(command)

	s.logger.Info("handling command",
		zap.String("command", command),
		zap.String("command_lower", commandLower),
		zap.Int64("telegram_id", telegramID),
		zap.String("args", message.CommandArguments()),
		zap.String("full_text", message.Text),
	)

	switch commandLower {
	case "start":
		s.handleStart(message)
	case "help":
		s.handleHelp(message)
	case "unsubscribe":
		s.handleUnsubscribe(telegramID, message.Chat.ID)
	case "notification":
		args := message.CommandArguments()
		s.handleNotification(telegramID, message.Chat.ID, args)
	default:
		s.logger.Warn("unknown command received",
			zap.String("command", command),
			zap.String("command_lower", commandLower),
			zap.String("full_text", message.Text),
		)
		s.handleUnknownCommand(message)
	}
}

// handleStart handles /start command
func (s *BotCommandService) handleStart(message *tgbotapi.Message) {
	// Check if there's a parameter
	text := message.Text
	if strings.HasPrefix(text, "/start ") {
		param := strings.TrimPrefix(text, "/start ")
		if param == "unsubscribe" {
			s.handleUnsubscribe(message.From.ID, message.Chat.ID)
			return
		}
	}

	// Use message from config or default
	startMsg := s.startMessage
	if startMsg == "" {
		startMsg = "🤖 Привет! Я бот для изучения английского языка.\n\n" +
			"💡 Я буду отправлять вам уведомления о карточках, которые нужно повторить.\n\n" +
			"Используйте /help для получения дополнительной информации."
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, startMsg)
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("failed to send start message", zap.Error(err))
	}
}

// handleHelp handles /help command
func (s *BotCommandService) handleHelp(message *tgbotapi.Message) {
	helpMsg := s.helpMessage
	if helpMsg == "" {
		helpMsg = "📚 Помощь по боту:\n\n" +
			"🔧 **Доступные команды:**\n" +
			"• /start - Начать работу с ботом\n" +
			"• /help - Показать эту справку\n" +
			"• /unsubscribe - Отписаться от уведомлений\n" +
			"• /notification [daily|never|N] - Настроить периодичность уведомлений\n" +
			"  - daily - ежедневно\n" +
			"  - never - никогда\n" +
			"  - N - каждые N дней (например, /notification 3)\n\n" +
			"💡 Бот будет отправлять вам уведомления о карточках, которые нужно повторить."
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, helpMsg)
	msg.ParseMode = "Markdown"
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("failed to send help message", zap.Error(err))
	}
}

// handleUnsubscribe handles /unsubscribe command
func (s *BotCommandService) handleUnsubscribe(telegramID int64, chatID int64) {
	user, err := s.userRepo.GetUserByTelegramID(telegramID)
	if err != nil {
		s.logger.Error("failed to get user", zap.Error(err))
		msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка. Попробуйте позже.")
		s.bot.Send(msg)
		return
	}

	if user == nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Пользователь не найден. Используйте /start для регистрации.")
		s.bot.Send(msg)
		return
	}

	// Parse current settings
	var settings models.UserSettings
	if user.SettingsJSON != "" {
		if err := json.Unmarshal([]byte(user.SettingsJSON), &settings); err != nil {
			s.logger.Warn("failed to parse user settings", zap.Error(err))
		}
	}

	// Set notification frequency to "never"
	settings.NotificationFrequency = "never"

	// Save settings
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		s.logger.Error("failed to marshal settings", zap.Error(err))
		msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при сохранении настроек.")
		s.bot.Send(msg)
		return
	}

	if err := s.userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		s.logger.Error("failed to update user settings", zap.Error(err))
		msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при сохранении настроек.")
		s.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "✅ Вы отписаны от уведомлений.\n\n"+
		"Чтобы снова получать уведомления, используйте команду:\n"+
		"/notification daily")
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("failed to send unsubscribe message", zap.Error(err))
	}
}

// handleNotification handles /notification command
func (s *BotCommandService) handleNotification(telegramID int64, chatID int64, args string) {
	user, err := s.userRepo.GetUserByTelegramID(telegramID)
	if err != nil {
		s.logger.Error("failed to get user", zap.Error(err))
		msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка. Попробуйте позже.")
		s.bot.Send(msg)
		return
	}

	if user == nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Пользователь не найден. Используйте /start для регистрации.")
		s.bot.Send(msg)
		return
	}

	// Parse current settings
	var settings models.UserSettings
	if user.SettingsJSON != "" {
		if err := json.Unmarshal([]byte(user.SettingsJSON), &settings); err != nil {
			s.logger.Warn("failed to parse user settings", zap.Error(err))
		}
	}

	// Parse arguments
	args = strings.TrimSpace(args)
	if args == "" {
		// Show current settings
		currentFreq := settings.NotificationFrequency
		if currentFreq == "" {
			currentFreq = "daily"
		}
		msgText := fmt.Sprintf("📅 Текущая периодичность уведомлений: %s\n\n", currentFreq)
		if currentFreq != "daily" && currentFreq != "never" {
			if days, err := strconv.Atoi(currentFreq); err == nil {
				dayWord := pluralizeDays(days)
				msgText = fmt.Sprintf("📅 Текущая периодичность уведомлений: каждые %d %s\n\n", days, dayWord)
			}
		}
		msgText += "Используйте:\n"
		msgText += "• /notification daily - ежедневно\n"
		msgText += "• /notification never - никогда\n"
		msgText += "• /notification N - каждые N дней (например, /notification 3)"
		msg := tgbotapi.NewMessage(chatID, msgText)
		s.bot.Send(msg)
		return
	}

	// Set notification frequency
	argsLower := strings.ToLower(args)
	if argsLower == "daily" {
		settings.NotificationFrequency = "daily"
	} else if argsLower == "never" {
		settings.NotificationFrequency = "never"
	} else {
		// Try to parse as number
		days, err := strconv.Atoi(args)
		if err != nil || days < 1 {
			msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат. Используйте:\n"+
				"• /notification daily - ежедневно\n"+
				"• /notification never - никогда\n"+
				"• /notification N - каждые N дней (например, /notification 3)")
			s.bot.Send(msg)
			return
		}
		settings.NotificationFrequency = strconv.Itoa(days)
	}

	// Save settings
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		s.logger.Error("failed to marshal settings", zap.Error(err))
		msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при сохранении настроек.")
		s.bot.Send(msg)
		return
	}

	if err := s.userRepo.UpdateUserSettings(user.ID, string(settingsJSON)); err != nil {
		s.logger.Error("failed to update user settings", zap.Error(err))
		msg := tgbotapi.NewMessage(chatID, "❌ Произошла ошибка при сохранении настроек.")
		s.bot.Send(msg)
		return
	}

	// Send confirmation
	var msgText string
	if settings.NotificationFrequency == "daily" {
		msgText = "✅ Периодичность уведомлений установлена: ежедневно"
	} else if settings.NotificationFrequency == "never" {
		msgText = "✅ Уведомления отключены"
	} else {
		days, _ := strconv.Atoi(settings.NotificationFrequency)
		dayWord := pluralizeDays(days)
		msgText = fmt.Sprintf("✅ Периодичность уведомлений установлена: каждые %d %s", days, dayWord)
	}

	msg := tgbotapi.NewMessage(chatID, msgText)
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("failed to send notification message", zap.Error(err))
	}
}

// handleUnknownCommand handles unknown commands
func (s *BotCommandService) handleUnknownCommand(message *tgbotapi.Message) {
	unknownMsg := s.unknownCommandMessage
	if unknownMsg == "" {
		unknownMsg = "❓ Неизвестная команда. Используйте /help для получения информации о возможностях бота."
	}
	msg := tgbotapi.NewMessage(message.Chat.ID, unknownMsg)
	if _, err := s.bot.Send(msg); err != nil {
		s.logger.Error("failed to send unknown command message", zap.Error(err))
	}
}

// pluralizeDays returns the correct form of "день" based on the number
func pluralizeDays(n int) string {
	if n < 0 {
		n = -n
	}
	
	// Get last digit
	lastDigit := n % 10
	// Get last two digits
	lastTwoDigits := n % 100
	
	// Special cases for 11-14
	if lastTwoDigits >= 11 && lastTwoDigits <= 14 {
		return "дней"
	}
	
	// Cases based on last digit
	switch lastDigit {
	case 1:
		return "день"
	case 2, 3, 4:
		return "дня"
	default:
		return "дней"
	}
}
