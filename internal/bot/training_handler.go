package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// TrainingHandler handles training sessions
type TrainingHandler struct {
	bot              *tgbotapi.BotAPI
	trainingService  *service.TrainingService
	srsService       *service.SRSService
	optionsService   *service.OptionsService
	logger           *zap.Logger
	sessions         map[int64]*SessionState
	sessionsMutex    sync.RWMutex
}

// SessionState holds the state of an active training session
type SessionState struct {
	UserID         int64
	SessionID      int64
	Queue          []*models.UserCardWithTraining
	CurrentIndex   int
	ShownAt        time.Time
	OptionsShownAt *time.Time
	Options        []string
	CorrectAnswer  string
}

// NewTrainingHandler creates a new training handler
func NewTrainingHandler(
	bot *tgbotapi.BotAPI,
	trainingService *service.TrainingService,
	srsService *service.SRSService,
	optionsService *service.OptionsService,
	logger *zap.Logger,
) *TrainingHandler {
	return &TrainingHandler{
		bot:             bot,
		trainingService: trainingService,
		srsService:      srsService,
		optionsService:  optionsService,
		logger:          logger,
		sessions:        make(map[int64]*SessionState),
	}
}

// StartTraining starts a training session for a user
func (h *TrainingHandler) StartTraining(ctx context.Context, chatID, userID int64, source models.SessionSource) error {
	// Start session
	session, queue, err := h.trainingService.StartSession(userID, source)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	// Create session state
	state := &SessionState{
		UserID:       userID,
		SessionID:    session.ID,
		Queue:        queue,
		CurrentIndex: 0,
	}

	// Store session
	h.sessionsMutex.Lock()
	h.sessions[chatID] = state
	h.sessionsMutex.Unlock()

	// Send welcome message
	welcomeMsg := fmt.Sprintf("Начинаем тренировку! Сегодня %d карточек.", len(queue))
	h.sendMessage(chatID, welcomeMsg)

	// Show first card
	time.Sleep(500 * time.Millisecond)
	return h.showCard(chatID)
}

// showCard shows the current card
func (h *TrainingHandler) showCard(chatID int64) error {
	h.sessionsMutex.RLock()
	state, exists := h.sessions[chatID]
	h.sessionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("no active session")
	}

	if state.CurrentIndex >= len(state.Queue) {
		// Session finished
		return h.finishSession(chatID)
	}

	card := state.Queue[state.CurrentIndex]
	
	// Update state
	h.sessionsMutex.Lock()
	state.ShownAt = time.Now()
	state.OptionsShownAt = nil
	h.sessionsMutex.Unlock()

	// Generate options
	options, correctAnswer, err := h.optionsService.GenerateOptions(card, models.DefaultOptionCount)
	if err != nil {
		h.logger.Error("failed to generate options", zap.Error(err))
		return h.skipCard(chatID, "Ошибка генерации вариантов")
	}

	h.sessionsMutex.Lock()
	state.Options = options
	state.CorrectAnswer = correctAnswer
	h.sessionsMutex.Unlock()

	// Build question text
	var questionText string
	if card.UserCard.Direction == models.DirectionRUtoEN {
		questionText = fmt.Sprintf(
			"🇬🇧 Переведите на английский:\n\n*%s*",
			card.TrainingCard.MeaningRU,
		)
		if card.TrainingCard.Hint != "" {
			questionText += fmt.Sprintf("\n\n💡 Подсказка: _%s_", card.TrainingCard.Hint)
		}
	} else {
		questionText = fmt.Sprintf(
			"🇷🇺 Что означает слово:\n\n*%s* %s",
			card.TrainingCard.WordEN,
			card.TrainingCard.Transcription,
		)
		if card.TrainingCard.ExampleEN != "" {
			questionText += fmt.Sprintf("\n\n📝 Пример: _%s_", card.TrainingCard.ExampleEN)
		}
	}

	// Add progress
	questionText = fmt.Sprintf(
		"Карточка %d из %d\n\n%s",
		state.CurrentIndex+1,
		len(state.Queue),
		questionText,
	)

	// Create keyboard with "Show options" button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Показать варианты", "show_options"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, questionText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Schedule automatic options reveal after delay
	go h.autoRevealOptions(chatID, models.OptionsDelayMS)

	return nil
}

// autoRevealOptions automatically reveals options after delay
func (h *TrainingHandler) autoRevealOptions(chatID int64, delayMS int) {
	time.Sleep(time.Duration(delayMS) * time.Millisecond)

	h.sessionsMutex.RLock()
	state, exists := h.sessions[chatID]
	if !exists || state.OptionsShownAt != nil {
		h.sessionsMutex.RUnlock()
		return
	}
	h.sessionsMutex.RUnlock()

	// Show options
	h.ShowOptions(chatID, false)
}

// ShowOptions shows answer options
func (h *TrainingHandler) ShowOptions(chatID int64, earlyReveal bool) error {
	h.sessionsMutex.Lock()
	state, exists := h.sessions[chatID]
	if !exists {
		h.sessionsMutex.Unlock()
		return fmt.Errorf("no active session")
	}

	// Check if options already shown
	if state.OptionsShownAt != nil {
		h.sessionsMutex.Unlock()
		return nil
	}

	now := time.Now()
	state.OptionsShownAt = &now
	options := state.Options
	h.sessionsMutex.Unlock()

	// Build keyboard with options
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, option := range options {
		callbackData := fmt.Sprintf("answer_%d", i)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(option, callbackData),
		))
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	optionsText := "Выберите правильный вариант:"
	if earlyReveal {
		optionsText = "Варианты ответа:"
	}

	msg := tgbotapi.NewMessage(chatID, optionsText)
	msg.ReplyMarkup = keyboard

	if _, err := h.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send options: %w", err)
	}

	return nil
}

// HandleAnswer handles user's answer
func (h *TrainingHandler) HandleAnswer(chatID int64, optionIndex int) error {
	h.sessionsMutex.RLock()
	state, exists := h.sessions[chatID]
	if !exists {
		h.sessionsMutex.RUnlock()
		return fmt.Errorf("no active session")
	}

	if state.CurrentIndex >= len(state.Queue) {
		h.sessionsMutex.RUnlock()
		return nil
	}

	card := state.Queue[state.CurrentIndex]
	answeredAt := time.Now()
	shownAt := state.ShownAt
	optionsShownAt := state.OptionsShownAt
	options := state.Options
	correctAnswer := state.CorrectAnswer
	h.sessionsMutex.RUnlock()

	// Validate option index
	if optionIndex < 0 || optionIndex >= len(options) {
		return fmt.Errorf("invalid option index")
	}

	chosenOption := options[optionIndex]
	isCorrect := chosenOption == correctAnswer

	// Calculate timings
	var tDelayMS int
	var earlyReveal bool
	if optionsShownAt != nil {
		tDelayMS = int(optionsShownAt.Sub(shownAt).Milliseconds())
		earlyReveal = tDelayMS < models.OptionsDelayMS
	} else {
		// User answered before options were shown (shouldn't happen)
		tDelayMS = models.OptionsDelayMS
		earlyReveal = false
	}

	answerTimeMS := int(answeredAt.Sub(*optionsShownAt).Milliseconds())

	// Create attempt data
	attemptData := models.AttemptData{
		Correct:      isCorrect,
		EarlyReveal:  earlyReveal,
		AnswerTimeMS: answerTimeMS,
		TDelayMS:     tDelayMS,
		OptionCount:  len(options),
		ChosenOption: chosenOption,
	}

	// Grade card
	if err := h.srsService.GradeCard(&card.UserCard, attemptData); err != nil {
		h.logger.Error("failed to grade card", zap.Error(err))
	}

	// Record wrong answer if incorrect
	if !isCorrect {
		if err := h.srsService.RecordWrongAnswer(&card.UserCard, chosenOption); err != nil {
			h.logger.Error("failed to record wrong answer", zap.Error(err))
		}
	}

	// Send feedback
	h.sendFeedback(chatID, &card.TrainingCard, isCorrect, chosenOption, correctAnswer)

	// Move to next card
	h.sessionsMutex.Lock()
	state.CurrentIndex++
	h.sessionsMutex.Unlock()

	// Show next card after delay
	time.Sleep(2 * time.Second)
	return h.showCard(chatID)
}

// sendFeedback sends feedback about the answer
func (h *TrainingHandler) sendFeedback(chatID int64, card *models.TrainingCard, isCorrect bool, chosen, correct string) {
	var message string

	if isCorrect {
		message = fmt.Sprintf("✅ Правильно!\n\n*%s* — %s", card.WordEN, card.MeaningRU)
	} else {
		message = fmt.Sprintf("❌ Неправильно\n\nВы выбрали: %s\nПравильный ответ: *%s*\n\n%s — %s",
			chosen, correct, card.WordEN, card.MeaningRU)
	}

	// Add example
	if card.ExampleEN != "" {
		message += fmt.Sprintf("\n\n📝 %s\n_%s_", card.ExampleEN, card.ExampleRU)
	}

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Error("failed to send feedback", zap.Error(err))
	}
}

// skipCard skips the current card
func (h *TrainingHandler) skipCard(chatID int64, reason string) error {
	h.sendMessage(chatID, fmt.Sprintf("⏭ Пропускаем карточку: %s", reason))

	h.sessionsMutex.Lock()
	state := h.sessions[chatID]
	if state != nil {
		state.CurrentIndex++
	}
	h.sessionsMutex.Unlock()

	time.Sleep(1 * time.Second)
	return h.showCard(chatID)
}

// finishSession finishes the training session
func (h *TrainingHandler) finishSession(chatID int64) error {
	h.sessionsMutex.Lock()
	state, exists := h.sessions[chatID]
	if !exists {
		h.sessionsMutex.Unlock()
		return nil
	}

	sessionID := state.SessionID
	doneCount := state.CurrentIndex
	delete(h.sessions, chatID)
	h.sessionsMutex.Unlock()

	// Finish session in database
	if err := h.trainingService.FinishSession(sessionID, doneCount); err != nil {
		h.logger.Error("failed to finish session", zap.Error(err))
	}

	// Send completion message
	message := fmt.Sprintf(
		"🎉 Тренировка завершена!\n\nВы прошли %d карточек.\n\nОтличная работа! До встречи завтра.",
		doneCount,
	)

	h.sendMessage(chatID, message)

	return nil
}

// CancelSession cancels an active training session
func (h *TrainingHandler) CancelSession(chatID int64) error {
	h.sessionsMutex.Lock()
	state, exists := h.sessions[chatID]
	if !exists {
		h.sessionsMutex.Unlock()
		return fmt.Errorf("no active session")
	}

	sessionID := state.SessionID
	doneCount := state.CurrentIndex
	delete(h.sessions, chatID)
	h.sessionsMutex.Unlock()

	// Finish session in database
	if err := h.trainingService.FinishSession(sessionID, doneCount); err != nil {
		h.logger.Error("failed to finish session", zap.Error(err))
	}

	h.sendMessage(chatID, "Тренировка отменена. Возвращайтесь когда будете готовы!")

	return nil
}

// HasActiveSession checks if user has an active session
func (h *TrainingHandler) HasActiveSession(chatID int64) bool {
	h.sessionsMutex.RLock()
	defer h.sessionsMutex.RUnlock()
	_, exists := h.sessions[chatID]
	return exists
}

// sendMessage is a helper to send text messages
func (h *TrainingHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Error("failed to send message", zap.Error(err))
	}
}

