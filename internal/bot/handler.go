package bot

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	bot                *tgbotapi.BotAPI
	logger             *zap.Logger
	aiService          *ai.Service
	wordService        *service.WordService
	trainingHandler    *TrainingHandler
	userRepo           *repository.UserRepository
	trainingCardRepo   *repository.TrainingCardRepository
	userCardRepo       *repository.UserCardRepository
	cbService          *service.CircuitBreakerService
	config             *config.Config
	db                 *sql.DB
}

// NewHandler creates a new handler
func NewHandler(
	bot *tgbotapi.BotAPI,
	logger *zap.Logger,
	aiService *ai.Service,
	wordService *service.WordService,
	trainingHandler *TrainingHandler,
	userRepo *repository.UserRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo *repository.UserCardRepository,
	cbService *service.CircuitBreakerService,
	config *config.Config,
	db *sql.DB,
) *Handler {
	return &Handler{
		bot:              bot,
		logger:           logger,
		aiService:        aiService,
		wordService:      wordService,
		trainingHandler:  trainingHandler,
		userRepo:         userRepo,
		trainingCardRepo: trainingCardRepo,
		userCardRepo:     userCardRepo,
		cbService:        cbService,
		config:           config,
		db:               db,
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
	case "stats":
		h.handleStatsCommand(ctx, chatID, userID)
	case "get_id":
		h.handleGetIDCommand(chatID, userID)
	case "reset_circuit":
		h.handleResetCircuitCommand(chatID, userID)
	case "delete_train":
		h.handleDeleteTrainCommand(chatID, userID, message.CommandArguments())
	case "delete_train_all":
		h.handleDeleteTrainAllCommand(chatID, userID)
	case "get_train_data":
		h.handleGetTrainDataCommand(chatID, userID, message.CommandArguments())
	default:
		h.sendMessage(chatID, h.config.Bot.UnknownCommandMessage)
	}
}

// handleTrainCommand handles /train command
func (h *Handler) handleTrainCommand(ctx context.Context, chatID, userID int64) {
	// Ensure user exists and get internal user ID
	user, err := h.userRepo.GetOrCreateUser(userID)
	if err != nil {
		h.logger.Error("failed to get/create user", zap.Error(err))
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}
	// Note: Username is updated in handleMessage when user sends messages

	h.logger.Info("starting training for user",
		zap.Int64("telegram_id", userID),
		zap.Int64("internal_user_id", user.ID),
	)

	// Start training with internal user ID
	if err := h.trainingHandler.StartTraining(ctx, chatID, user.ID, models.SourceManual); err != nil {
		h.logger.Error("failed to start training", zap.Error(err))
		if strings.Contains(err.Error(), "no cards available") {
			h.sendMessage(chatID, "У вас пока нет карточек для тренировки. Сначала запросите несколько слов!")
		} else {
			h.sendMessage(chatID, "Не удалось начать тренировку. Попробуйте позже.")
		}
	}
}

// handleStatsCommand handles /stats command
func (h *Handler) handleStatsCommand(ctx context.Context, chatID, userID int64) {
	// Ensure user exists and get internal user ID
	user, err := h.userRepo.GetOrCreateUser(userID)
	if err != nil {
		h.logger.Error("failed to get/create user", zap.Error(err))
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	internalUserID := user.ID
	now := time.Now()

	// Get new cards count (exclude orphaned cards - those with non-existent training_cards or word_cards)
	// Excludes words marked as "known" in user_word_knowledge (same as GetNewCards)
	newQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state = 'new'
		AND NOT EXISTS (
			SELECT 1 FROM user_word_knowledge uwk 
			WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
		)`
	var newCount int
	err = h.db.QueryRow(newQuery, internalUserID, internalUserID).Scan(&newCount)
	if err != nil {
		h.logger.Error("failed to get new cards count", zap.Error(err))
		newCount = 0
	}

	// Get due count (cards ready for review, excluding new cards and orphaned cards)
	// Excludes words marked as "known" in user_word_knowledge (same as GetDueCards)
	dueQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state != 'new' AND (uc.next_due_at IS NULL OR uc.next_due_at <= ?)
		AND NOT EXISTS (
			SELECT 1 FROM user_word_knowledge uwk 
			WHERE uwk.user_id = ? AND uwk.word_card_id = tc.word_card_id AND uwk.status = 'known'
		)`
	var dueCount int
	err = h.db.QueryRow(dueQuery, internalUserID, now, internalUserID).Scan(&dueCount)
	if err != nil {
		h.logger.Error("failed to get due count", zap.Error(err))
		dueCount = 0
	}

	// Get learning cards count (exclude orphaned cards)
	learningQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state = 'learning'`
	var learningCount int
	err = h.db.QueryRow(learningQuery, internalUserID).Scan(&learningCount)
	if err != nil {
		h.logger.Error("failed to get learning count", zap.Error(err))
		learningCount = 0
	}

	// Get review cards count (exclude orphaned cards)
	reviewQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ? AND uc.state = 'review'`
	var reviewCount int
	err = h.db.QueryRow(reviewQuery, internalUserID).Scan(&reviewCount)
	if err != nil {
		h.logger.Error("failed to get review count", zap.Error(err))
		reviewCount = 0
	}

	// Calculate available cards for training
	availableForTraining := dueCount
	if newCount > 0 {
		availableForTraining += newCount
	}

	// Get total cards count (exclude orphaned cards)
	totalQuery := `SELECT COUNT(*) 
		FROM user_cards uc
		INNER JOIN training_cards tc ON uc.training_card_id = tc.id
		INNER JOIN word_cards wc ON tc.word_card_id = wc.id
		WHERE uc.user_id = ?`
	var totalCards int
	err = h.db.QueryRow(totalQuery, internalUserID).Scan(&totalCards)
	if err != nil {
		h.logger.Error("failed to get total cards count", zap.Error(err))
		totalCards = 0
	}

	// Get accuracy (last 30 days)
	monthAgo := now.AddDate(0, 0, -30)
	var totalReviews int
	var correctReviews int
	accuracyQuery := `SELECT 
		COUNT(*) as total,
		SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END) as correct
		FROM review_events 
		WHERE user_id = ? AND answered_at >= ? AND answered_at IS NOT NULL`
	err = h.db.QueryRow(accuracyQuery, internalUserID, monthAgo).Scan(&totalReviews, &correctReviews)
	if err != nil {
		h.logger.Error("failed to get accuracy", zap.Error(err))
		totalReviews = 0
		correctReviews = 0
	}

	var accuracyPercent float64
	if totalReviews > 0 {
		accuracyPercent = float64(correctReviews) / float64(totalReviews) * 100
	}

	// Format message
	var message strings.Builder
	message.WriteString("📊 *Статистика*\n\n")
	message.WriteString("*Карточки:*\n")
	message.WriteString(fmt.Sprintf("• Доступно для тренировки: *%d*\n", availableForTraining))
	message.WriteString(fmt.Sprintf("• Новые: *%d*\n", newCount))
	message.WriteString(fmt.Sprintf("• В изучении: *%d*\n", learningCount))
	message.WriteString(fmt.Sprintf("• На повторении: *%d*\n", reviewCount))
	message.WriteString(fmt.Sprintf("• Всего карточек: *%d*\n\n", totalCards))
	message.WriteString("*Точность (30 дней):*\n")
	if totalReviews > 0 {
		message.WriteString(fmt.Sprintf("• Правильных ответов: *%.1f%%*\n", accuracyPercent))
		message.WriteString(fmt.Sprintf("• Всего ответов: *%d*\n", totalReviews))
	} else {
		message.WriteString("• Пока нет данных\n")
	}

	h.sendMessage(chatID, message.String())
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

// handleDeleteTrainCommand handles /delete_train command (admin only)
func (h *Handler) handleDeleteTrainCommand(chatID, userID int64, wordEN string) {
	// Check if user is admin
	if h.config.Admin.TelegramID == 0 || userID != h.config.Admin.TelegramID {
		// Silently ignore for non-admins
		return
	}

	// Validate word
	wordEN = strings.TrimSpace(wordEN)
	if wordEN == "" {
		h.sendMessage(chatID, "❌ Укажите слово для удаления: `/delete_train word`")
		return
	}

	// Delete training cards
	rowsAffected, err := h.trainingCardRepo.DeleteTrainingCardsByWordEN(wordEN)
	if err != nil {
		h.logger.Error("failed to delete training cards",
			zap.String("word_en", wordEN),
			zap.Error(err),
		)
		h.sendMessage(chatID, fmt.Sprintf("❌ Ошибка при удалении карточек для слова `%s`", wordEN))
		return
	}

	if rowsAffected == 0 {
		h.sendMessage(chatID, fmt.Sprintf("ℹ️ Тренировочные карточки для слова `%s` не найдены", wordEN))
		return
	}

	h.logger.Info("deleted training cards",
		zap.String("word_en", wordEN),
		zap.Int64("rows_affected", rowsAffected),
		zap.Int64("admin_id", userID),
	)

	h.sendMessage(chatID, fmt.Sprintf("✅ Удалено тренировочных карточек для слова `%s`: %d", wordEN, rowsAffected))
}

// handleDeleteTrainAllCommand handles /delete_train_all command (admin only)
func (h *Handler) handleDeleteTrainAllCommand(chatID, userID int64) {
	// Check if user is admin
	if h.config.Admin.TelegramID == 0 || userID != h.config.Admin.TelegramID {
		// Silently ignore for non-admins
		return
	}

	h.logger.Warn("admin requested deletion of all training cards",
		zap.Int64("admin_id", userID),
	)

	// Delete all training cards (cascades to user_cards and review_events)
	rowsAffected, err := h.trainingCardRepo.DeleteAllTrainingCards()
	if err != nil {
		h.logger.Error("failed to delete all training cards",
			zap.Error(err),
		)
		h.sendMessage(chatID, "❌ Ошибка при удалении всех тренировочных карточек")
		return
	}

	// Clean up any orphaned user_cards that might remain
	orphanedCount, err := h.userCardRepo.DeleteOrphanedUserCards()
	if err != nil {
		h.logger.Warn("failed to delete orphaned user cards",
			zap.Error(err),
		)
	}

	h.logger.Info("deleted all training cards",
		zap.Int64("rows_affected", rowsAffected),
		zap.Int64("orphaned_user_cards", orphanedCount),
		zap.Int64("admin_id", userID),
	)

	message := fmt.Sprintf("✅ Удалено всех тренировочных карточек: %d\n\n"+
		"Также автоматически удалены:\n"+
		"• Все пользовательские карточки (user_cards)\n"+
		"• Вся история ответов (review_events)", rowsAffected)
	if orphanedCount > 0 {
		message += fmt.Sprintf("\n• Дополнительно очищено висячих записей: %d", orphanedCount)
	}
	h.sendMessage(chatID, message)
}

// handleGetTrainDataCommand handles /get_train_data command (admin only)
func (h *Handler) handleGetTrainDataCommand(chatID, userID int64, wordEN string) {
	// Check if user is admin
	if h.config.Admin.TelegramID == 0 || userID != h.config.Admin.TelegramID {
		// Silently ignore for non-admins
		return
	}

	// Validate word
	wordEN = strings.TrimSpace(wordEN)
	if wordEN == "" {
		h.sendMessage(chatID, "❌ Укажите слово: `/get_train_data word`")
		return
	}

	// Get training cards
	cards, err := h.trainingCardRepo.GetTrainingCardsByWordEN(wordEN)
	if err != nil {
		h.logger.Error("failed to get training cards",
			zap.String("word_en", wordEN),
			zap.Error(err),
		)
		h.sendMessage(chatID, fmt.Sprintf("❌ Ошибка при получении данных для слова `%s`", wordEN))
		return
	}

	if len(cards) == 0 {
		h.sendMessage(chatID, fmt.Sprintf("ℹ️ Тренировочные карточки для слова `%s` не найдены", wordEN))
		return
	}

	// Format response
	var message strings.Builder
	message.WriteString(fmt.Sprintf("📊 Данные по тренировочным карточкам для слова `%s`:\n\n", wordEN))
	message.WriteString(fmt.Sprintf("Всего карточек: %d\n\n", len(cards)))

	for i, card := range cards {
		message.WriteString(fmt.Sprintf("*Карточка #%d* (ID: `%d`)\n", i+1, card.ID))
		message.WriteString(fmt.Sprintf("• Word Card ID: `%d`\n", card.WordCardID))
		message.WriteString(fmt.Sprintf("• Sense Index: `%d`\n", card.SenseIndex))
		if card.Transcription != "" {
			message.WriteString(fmt.Sprintf("• Transcription: `%s`\n", card.Transcription))
		}
		message.WriteString(fmt.Sprintf("• Word EN: `%s`\n", card.WordEN))
		message.WriteString(fmt.Sprintf("• Word RU: %s\n", card.WordRU))
		message.WriteString(fmt.Sprintf("• Meaning EN: %s\n", card.MeaningEN))
		if card.ExampleEN != "" {
			message.WriteString(fmt.Sprintf("• Example EN: %s\n", card.ExampleEN))
		}
		if card.ExampleRU != "" {
			message.WriteString(fmt.Sprintf("• Example RU: %s\n", card.ExampleRU))
		}
		if card.Hint != "" {
			message.WriteString(fmt.Sprintf("• Hint: _%s_\n", card.Hint))
		}
		if card.DistractorsRU != "" {
			message.WriteString(fmt.Sprintf("• Distractors RU: `%s`\n", card.DistractorsRU))
		}
		if card.DistractorsEN != "" {
			message.WriteString(fmt.Sprintf("• Distractors EN: `%s`\n", card.DistractorsEN))
		}
		message.WriteString(fmt.Sprintf("• Created: `%s`\n", card.CreatedAt.Format("2006-01-02 15:04:05")))
		if i < len(cards)-1 {
			message.WriteString("\n---\n\n")
		}
	}

	h.logger.Info("retrieved training cards data",
		zap.String("word_en", wordEN),
		zap.Int("count", len(cards)),
		zap.Int64("admin_id", userID),
	)

	h.sendMessage(chatID, message.String())
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
	} else if strings.HasPrefix(data, "answer_") {
		// Ensure user exists and get internal user ID
		user, err := h.userRepo.GetOrCreateUser(query.From.ID)
		if err != nil {
			h.logger.Error("failed to get/create user", zap.Error(err))
			h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже.")
			return
		}
		// Update username if it changed
		if query.From.UserName != "" && user != nil && user.TelegramUsername != query.From.UserName {
			if err := h.userRepo.UpdateUsername(query.From.ID, query.From.UserName); err != nil {
				h.logger.Warn("failed to update username", zap.Error(err))
			}
		}

		// Try to restore session if not in memory
		if !h.trainingHandler.HasActiveSession(chatID) {
			_, restoreErr := h.trainingHandler.RestoreSession(chatID, user.ID)
			if restoreErr != nil {
				h.logger.Warn("failed to restore session", zap.Error(restoreErr))
			}
		}

		// Handle answer selection
		optionIndexStr := strings.TrimPrefix(data, "answer_")
		optionIndex, err := strconv.Atoi(optionIndexStr)
		if err != nil {
			h.logger.Error("invalid option index", zap.Error(err))
			return
		}

		if err := h.trainingHandler.HandleAnswer(chatID, optionIndex); err != nil {
			h.logger.Error("failed to handle answer", zap.Error(err))
			if strings.Contains(err.Error(), "no active session") {
				h.sendMessage(chatID, "Сессия не найдена. Начните новую тренировку командой /train")
			}
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

	// Ensure user exists in database (for training cards)
	user, userErr := h.userRepo.GetOrCreateUser(userID)
	if userErr != nil {
		h.logger.Error("failed to get/create user", zap.Error(userErr))
	} else if user != nil && message.From.UserName != "" && user.TelegramUsername != message.From.UserName {
		// Update username if it changed
		if err := h.userRepo.UpdateUsername(userID, message.From.UserName); err != nil {
			h.logger.Warn("failed to update username", zap.Error(err))
		}
	}

	// Send typing indicator
	h.sendTyping(chatID)

	var response string
	var err error

	// Check if it's a single word - use word service (DB + AI)
	// Use internal user.ID instead of telegram_id to match web chat behavior
	if h.wordService.IsSingleWord(text) {
		h.logger.Debug("detected single word request",
			zap.String("word", text),
		)
		// Use internal user ID (not telegram_id) for consistency with web chat
		internalUserID := userID // Default to telegram_id if user not found
		if user != nil {
			internalUserID = user.ID
		}
		response, err = h.wordService.GetWordDefinition(ctx, internalUserID, text)
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
