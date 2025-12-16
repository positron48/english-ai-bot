package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Handler handles Telegram updates
type Handler struct {
	bot             *tgbotapi.BotAPI
	logger          *zap.Logger
	aiService       *ai.Service
	wordService     *service.WordService
	trainingHandler *TrainingHandler
	userRepo        *repository.UserRepository
	cbService       *service.CircuitBreakerService
	config          *config.Config
}

// NewHandler creates a new handler
func NewHandler(
	bot *tgbotapi.BotAPI,
	logger *zap.Logger,
	aiService *ai.Service,
	wordService *service.WordService,
	trainingHandler *TrainingHandler,
	userRepo *repository.UserRepository,
	cbService *service.CircuitBreakerService,
	config *config.Config,
) *Handler {
	return &Handler{
		bot:             bot,
		logger:          logger,
		aiService:       aiService,
		wordService:     wordService,
		trainingHandler: trainingHandler,
		userRepo:        userRepo,
		cbService:       cbService,
		config:          config,
	}
}

// HandleUpdate handles incoming Telegram updates
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	// Handle callback queries (button presses)
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	message := update.Message
	h.logger.Info("received message",
		zap.Int64("chat_id", message.Chat.ID),
		zap.String("text", message.Text),
		zap.String("username", message.From.UserName),
	)

	// Handle commands
	if message.IsCommand() {
		h.handleCommand(ctx, message)
		return
	}

	// Handle regular messages
	h.handleMessage(ctx, message)
}

// handleCommand handles bot commands
func (h *Handler) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	command := message.Command()
	chatID := message.Chat.ID
	userID := message.From.ID

	h.logger.Info("handling command",
		zap.String("command", command),
		zap.Int64("user_id", userID),
	)

	switch command {
	case "start":
		h.sendMessage(chatID, h.config.Bot.StartMessage)
	case "help":
		h.sendMessage(chatID, h.config.Bot.HelpMessage)
	case "train":
		h.handleTrainCommand(ctx, chatID, userID)
	case "get_id":
		h.handleGetIDCommand(chatID, userID)
	case "reset_circuit":
		h.handleResetCircuitCommand(chatID, userID)
	default:
		h.sendMessage(chatID, h.config.Bot.UnknownCommandMessage)
	}
}

// handleTrainCommand handles /train command
func (h *Handler) handleTrainCommand(ctx context.Context, chatID, userID int64) {
	// Ensure user exists
	if _, err := h.userRepo.GetOrCreateUser(userID); err != nil {
		h.logger.Error("failed to get/create user", zap.Error(err))
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	// Start training
	if err := h.trainingHandler.StartTraining(ctx, chatID, userID, models.SourceManual); err != nil {
		h.logger.Error("failed to start training", zap.Error(err))
		if strings.Contains(err.Error(), "no cards available") {
			h.sendMessage(chatID, "У вас пока нет карточек для тренировки. Сначала запросите несколько слов!")
		} else {
			h.sendMessage(chatID, "Не удалось начать тренировку. Попробуйте позже.")
		}
	}
}

// handleGetIDCommand handles /get_id command
func (h *Handler) handleGetIDCommand(chatID, userID int64) {
	message := fmt.Sprintf("Your Telegram ID: `%d`", userID)
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown
	h.bot.Send(msg)
}

// handleResetCircuitCommand handles /reset_circuit command (admin only)
func (h *Handler) handleResetCircuitCommand(chatID, userID int64) {
	// Check if user is admin
	if h.config.Admin.TelegramID == 0 || userID != h.config.Admin.TelegramID {
		// Silently ignore for non-admins
		return
	}

	// Reset circuit breaker
	if err := h.cbService.Reset(); err != nil {
		h.logger.Error("failed to reset circuit breaker", zap.Error(err))
		h.sendMessage(chatID, "❌ Не удалось сбросить circuit breaker")
		return
	}

	h.sendMessage(chatID, "✅ Circuit breaker сброшен. Воркер возобновит работу.")
}

// handleCallbackQuery handles callback queries from inline keyboards
func (h *Handler) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	data := query.Data

	// Acknowledge callback
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := h.bot.Request(callback); err != nil {
		h.logger.Error("failed to acknowledge callback", zap.Error(err))
	}

	h.logger.Info("handling callback query",
		zap.Int64("chat_id", chatID),
		zap.String("data", data),
	)

	// Handle different callback types
	if data == "train_start" {
		// Start training from notification
		h.handleTrainCommand(ctx, chatID, query.From.ID)
	} else if data == "show_options" {
		// Show options for current card
		if err := h.trainingHandler.ShowOptions(chatID, true); err != nil {
			h.logger.Error("failed to show options", zap.Error(err))
		}
	} else if strings.HasPrefix(data, "answer_") {
		// Handle answer selection
		optionIndexStr := strings.TrimPrefix(data, "answer_")
		optionIndex, err := strconv.Atoi(optionIndexStr)
		if err != nil {
			h.logger.Error("invalid option index", zap.Error(err))
			return
		}

		if err := h.trainingHandler.HandleAnswer(chatID, optionIndex); err != nil {
			h.logger.Error("failed to handle answer", zap.Error(err))
		}
	}
}

// handleMessage handles regular text messages
func (h *Handler) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID
	text := message.Text

	if text == "" {
		h.sendMessage(chatID, h.config.Bot.EmptyMessage)
		return
	}

	h.logger.Info("processing user message",
		zap.Int64("chat_id", chatID),
		zap.Int64("user_id", userID),
		zap.String("text", text),
	)

	// Send typing indicator
	h.sendTyping(chatID)

	var response string
	var err error

	// Check if it's a single word - use word service (DB + AI)
	if h.wordService.IsSingleWord(text) {
		h.logger.Debug("detected single word request",
			zap.String("word", text),
		)
		response, err = h.wordService.GetWordDefinition(ctx, userID, text)
	} else {
		// Regular message - use AI service directly
		response, err = h.aiService.GenerateResponse(ctx, text)
	}

	if err != nil {
		h.logger.Error("failed to get response", zap.Error(err))
		h.sendMessage(chatID, h.config.Bot.ErrorMessage)
		return
	}

	// Convert Markdown to Telegram format and send response
	telegramResponse := utils.ConvertMarkdownToTelegram(response)
	h.sendMessage(chatID, telegramResponse)
}

// sendMessage sends a message to the specified chat
func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Error("failed to send message", zap.Error(err))
	}
}

// sendTyping sends a typing indicator to the specified chat
func (h *Handler) sendTyping(chatID int64) {
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := h.bot.Request(action); err != nil {
		h.logger.Error("failed to send typing indicator", zap.Error(err))
	}
}
