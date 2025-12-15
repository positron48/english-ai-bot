package bot

import (
	"context"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Handler handles Telegram updates
type Handler struct {
	bot         *tgbotapi.BotAPI
	logger      *zap.Logger
	aiService   *ai.Service
	wordService *service.WordService
	config      *config.Config
}

// NewHandler creates a new handler
func NewHandler(bot *tgbotapi.BotAPI, logger *zap.Logger, aiService *ai.Service, wordService *service.WordService, config *config.Config) *Handler {
	return &Handler{
		bot:         bot,
		logger:      logger,
		aiService:   aiService,
		wordService: wordService,
		config:      config,
	}
}

// HandleUpdate handles incoming Telegram updates
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
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

	h.logger.Info("handling command", zap.String("command", command))

	switch command {
	case "start":
		h.sendMessage(chatID, h.config.Bot.StartMessage)
	case "help":
		h.sendMessage(chatID, h.config.Bot.HelpMessage)
	default:
		h.sendMessage(chatID, h.config.Bot.UnknownCommandMessage)
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
