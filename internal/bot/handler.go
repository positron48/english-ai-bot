package bot

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

// userRepoInterface is used by Handler for user lookup and username updates (allows tests to inject mocks).
type userRepoInterface interface {
	GetOrCreateUser(telegramID int64) (*models.User, error)
	UpdateUsername(telegramID int64, username string) error
	UpdateUserSettings(userID int64, settingsJSON string) error
}

// Handler handles Telegram updates
type Handler struct {
	bot               *tgbotapi.BotAPI
	logger            *zap.Logger
	aiService         *ai.Service
	wordServices      map[string]*service.WordService
	defaultCourse     string
	courseRepo        *repository.CourseRepository
	userRepo          userRepoInterface
	trainingCardRepo  *repository.TrainingCardRepository
	userCardRepo      *repository.UserCardRepository
	cbService         *service.CircuitBreakerService
	config            *config.Config
	db                *sql.DB
	botCommandService *service.BotCommandService
}

// NewHandler creates a new handler
func NewHandler(
	bot *tgbotapi.BotAPI,
	logger *zap.Logger,
	aiService *ai.Service,
	wordServices map[string]*service.WordService,
	defaultCourse string,
	courseRepo *repository.CourseRepository,
	userRepo *repository.UserRepository,
	trainingCardRepo *repository.TrainingCardRepository,
	userCardRepo *repository.UserCardRepository,
	cbService *service.CircuitBreakerService,
	config *config.Config,
	db *sql.DB,
) *Handler {
	// Initialize bot command service
	var botCommandService *service.BotCommandService
	if bot != nil {
		botCommandService = service.NewBotCommandService(
			bot,
			userRepo,
			logger,
			config.Bot.HelpMessage,
			config.Bot.StartMessage,
			config.Bot.UnknownCommandMessage,
		)
	}

	return &Handler{
		bot:               bot,
		logger:            logger,
		aiService:         aiService,
		wordServices:      wordServices,
		defaultCourse:     defaultCourse,
		courseRepo:        courseRepo,
		userRepo:          userRepo,
		trainingCardRepo:  trainingCardRepo,
		userCardRepo:      userCardRepo,
		cbService:         cbService,
		config:            config,
		db:                db,
		botCommandService: botCommandService,
	}
}

// resolveUserCourse returns the user's currently selected course code and its WordService.
// It reuses the shared current-course mechanism (same as the website), falling back to the
// deployment default course when the user has no selection or the lookup fails.
func (h *Handler) resolveUserCourse(ctx context.Context, internalUserID int64) (string, *service.WordService) {
	courseCode := h.defaultCourse
	if h.courseRepo != nil && internalUserID != 0 {
		if resolved, err := h.courseRepo.ResolveCurrentCourseCode(ctx, internalUserID, h.defaultCourse); err != nil {
			h.logger.Warn("failed to resolve current course, using default",
				zap.Int64("user_id", internalUserID), zap.Error(err))
		} else if resolved != "" {
			courseCode = resolved
		}
	}
	if ws, ok := h.wordServices[courseCode]; ok {
		return courseCode, ws
	}
	// Fallback to the default course's word service.
	return h.defaultCourse, h.wordServices[h.defaultCourse]
}

// HandleUpdate handles incoming Telegram updates
func (h *Handler) HandleUpdate(ctx context.Context, update tgbotapi.Update) {
	// Migration shim: this instance is retired, just point users to the new bot
	// and skip all normal handling (no DB access, no AI calls).
	if h.config.Migration.Enabled {
		h.handleMigrationRedirect(update)
		return
	}

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

	// Handle commands - use BotCommandService first for notification commands
	if message.IsCommand() {
		command := message.Command()
		// Let BotCommandService handle notification-related commands
		if h.botCommandService != nil && (command == "unsubscribe" || command == "notification" || command == "start" || command == "help") {
			h.botCommandService.HandleUpdate(update)
			return
		}
		// Other commands handled by existing handler
		h.handleCommand(ctx, message)
		return
	}

	// Handle regular messages
	h.handleMessage(ctx, message)
}

// handleMigrationRedirect replies to any incoming update with the migration notice.
func (h *Handler) handleMigrationRedirect(update tgbotapi.Update) {
	var chatID int64
	switch {
	case update.Message != nil:
		chatID = update.Message.Chat.ID
	case update.CallbackQuery != nil && update.CallbackQuery.Message != nil:
		chatID = update.CallbackQuery.Message.Chat.ID
	default:
		return
	}
	h.sendMessage(chatID, h.config.Migration.Message)
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
	case "language":
		h.handleLanguageCommand(ctx, chatID, userID)
	case "get_id":
		h.handleGetIDCommand(chatID, userID)
	case "reset_circuit":
		h.handleResetCircuitCommand(chatID, userID)
	default:
		h.sendMessage(chatID, h.config.Bot.UnknownCommandMessage)
	}
}

// handleGetIDCommand handles /get_id command
func (h *Handler) handleGetIDCommand(chatID, userID int64) {
	message := fmt.Sprintf("Your Telegram ID: `%d`", userID)
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown
	h.bot.Send(msg)
}

// handleLanguageCommand handles /language: shows the active courses as inline buttons so the
// user can switch the language the bot works in. The current course is marked.
func (h *Handler) handleLanguageCommand(ctx context.Context, chatID, userID int64) {
	user, err := h.userRepo.GetOrCreateUser(userID)
	if err != nil {
		h.logger.Error("failed to get/create user", zap.Error(err))
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	courses, err := h.courseRepo.ListCoursesForUser(ctx, user.ID, h.defaultCourse)
	if err != nil || len(courses) == 0 {
		if err != nil {
			h.logger.Error("failed to list courses", zap.Error(err))
		}
		h.sendMessage(chatID, "Сейчас нет доступных языков для выбора.")
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range courses {
		label := courseDisplayName(c)
		if c.IsCurrent {
			label = "✅ " + label
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "setlang:"+c.Code),
		))
	}

	msg := tgbotapi.NewMessage(chatID, "🌐 Выберите язык, с которым будет работать бот:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Error("failed to send language menu", zap.Error(err))
	}
}

// handleSetLanguage handles the setlang:<course_code> callback and persists the user's choice
// via the shared current-course mechanism (kept in sync with the website).
func (h *Handler) handleSetLanguage(ctx context.Context, query *tgbotapi.CallbackQuery, courseCode string) {
	chatID := query.Message.Chat.ID
	user, err := h.userRepo.GetOrCreateUser(query.From.ID)
	if err != nil {
		h.logger.Error("failed to get/create user", zap.Error(err))
		h.sendMessage(chatID, "Произошла ошибка. Попробуйте позже.")
		return
	}

	current, err := h.courseRepo.SelectCurrentCourse(ctx, user.ID, courseCode)
	if err != nil {
		h.logger.Error("failed to select course", zap.String("course_code", courseCode), zap.Error(err))
		h.sendMessage(chatID, "Не удалось переключить язык. Попробуйте позже.")
		return
	}

	confirmation := fmt.Sprintf("Язык переключён на %s", courseDisplayName(current.Course))
	edit := tgbotapi.NewEditMessageText(chatID, query.Message.MessageID, confirmation)
	if _, err := h.bot.Send(edit); err != nil {
		h.logger.Warn("failed to edit language confirmation, sending new message", zap.Error(err))
		h.sendMessage(chatID, confirmation)
	}
}

// courseDisplayName renders a friendly course label with a flag based on the target language.
func courseDisplayName(c repository.CourseSummary) string {
	title := strings.TrimSpace(c.Title)
	if title == "" {
		title = c.Code
	}
	if flag := languageFlag(c.TargetLanguage); flag != "" {
		return flag + " " + title
	}
	return title
}

func languageFlag(targetLang string) string {
	switch strings.ToLower(strings.TrimSpace(targetLang)) {
	case "en":
		return "🇬🇧"
	case "es":
		return "🇪🇸"
	default:
		return ""
	}
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

	h.logger.Info("handling callback query",
		zap.Int64("chat_id", chatID),
		zap.String("data", data),
	)

	// Let BotCommandService handle notification-related callbacks first
	if h.botCommandService != nil && data == "notification_unsubscribe" {
		update := tgbotapi.Update{
			UpdateID:      0,
			CallbackQuery: query,
		}
		h.botCommandService.HandleUpdate(update)
		return
	}

	// Acknowledge callback
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := h.bot.Request(callback); err != nil {
		h.logger.Error("failed to acknowledge callback", zap.Error(err))
	}

	// Handle different callback types
	if strings.HasPrefix(data, "setlang:") {
		courseCode := strings.TrimPrefix(data, "setlang:")
		h.handleSetLanguage(ctx, query, courseCode)
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

	// Use internal user ID (not telegram_id) for consistency with web chat
	internalUserID := userID // Default to telegram_id if user not found
	if user != nil {
		internalUserID = user.ID
	}

	// Resolve the user's selected language/course so word cards and corrections use the
	// right per-course prompt. Falls back to the deployment default course.
	courseCode, wordService := h.resolveUserCourse(ctx, internalUserID)
	targetLang := repository.GrammarBundleIDForCourse(courseCode) // e.g. "es_ru" -> "es"

	var response string
	var err error

	// Check if it's a single word - use word service (DB + AI)
	if wordService.IsSingleWord(text) {
		h.logger.Debug("detected single word request",
			zap.String("word", text),
			zap.String("course_code", courseCode),
		)
		response, err = wordService.GetWordDefinition(ctx, internalUserID, text)
	} else {
		// Regular message - use AI service with the user's selected course prompt
		response, err = h.aiService.GenerateResponseForCourse(ctx, text, courseCode)
		if err == nil && looksLikeWordInfoJSON(response) {
			rendered, converted := renderWordInfoJSONAsMarkdown(response, targetLang)
			if converted {
				h.logger.Info("converted JSON word-card response to markdown",
					zap.String("input", text),
					zap.String("target_lang", targetLang),
				)
				response = rendered
			} else {
				h.logger.Warn("detected JSON word-card response but failed to convert; sending raw response")
			}
		}
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
		if isTelegramParseEntitiesError(err) {
			// Fallback to plain text when Markdown entities are malformed in LLM output.
			plain := tgbotapi.NewMessage(chatID, text)
			if _, fallbackErr := h.bot.Send(plain); fallbackErr != nil {
				h.logger.Error("failed to send message", zap.Error(fallbackErr))
				return
			}
			h.logger.Warn("markdown send failed, sent plain text fallback", zap.Error(err))
			return
		}
		h.logger.Error("failed to send message", zap.Error(err))
	}
}

func isTelegramParseEntitiesError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "can't parse entities") || strings.Contains(s, "cant parse entities")
}

func looksLikeWordInfoJSON(s string) bool {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return false
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		return false
	}
	_, hasLemma := obj["lemma"]
	_, hasPOS := obj["pos"]
	_, hasExamples := obj["examples"]
	return hasLemma && hasPOS && hasExamples
}

func renderWordInfoJSONAsMarkdown(raw, targetLang string) (string, bool) {
	var info models.WordInfoResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &info); err != nil {
		return "", false
	}
	models.SyncWordInfoResponseNeutralAliases(&info)

	definition := info.DefinitionNative
	if definition == "" {
		definition = info.DefinitionRU
	}
	pos := info.POS
	transcription := info.Transcription
	display := info.Lemma
	if display == "" {
		display = info.InputWord
	}
	card := &models.WordCard{
		Word:          info.Lemma,
		POS:           &pos,
		Transcription: &transcription,
		DefinitionRU:  &definition,
		DisplayEN:     &display,
	}
	if card.Word == "" {
		card.Word = strings.TrimSpace(info.InputWord)
	}
	if card.Word == "" || strings.TrimSpace(definition) == "" {
		return "", false
	}

	return utils.RenderWordCardMarkdownForTarget(card, info.Examples, info.VerbForms, targetLang), true
}

// sendTyping sends a typing indicator to the specified chat
func (h *Handler) sendTyping(chatID int64) {
	action := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	if _, err := h.bot.Request(action); err != nil {
		h.logger.Error("failed to send typing indicator", zap.Error(err))
	}
}
