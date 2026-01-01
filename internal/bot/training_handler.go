package bot

import (
	"context"
	"encoding/json"
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
	bot                   *tgbotapi.BotAPI
	trainingService       *service.TrainingService
	srsService            *service.SRSService
	optionsService        *service.OptionsService
	logger                *zap.Logger
	sessions              map[int64]*SessionState
	sessionsMutex          sync.RWMutex
	optionsDelayMS         int
	wrongAnswerDelaySeconds int
}

// SessionState holds the state of an active training session
type SessionState struct {
	UserID              int64
	SessionID           int64
	Queue               []*models.UserCardWithTraining
	CurrentIndex        int
	ShownAt             time.Time
	OptionsShownAt      *time.Time
	Options             []string
	CorrectAnswer       string
	RecentCorrectAnswers []string // Last N correct answers to exclude from distractors
}

// SessionStateData holds serializable session state data
type SessionStateData struct {
	UserCardIDs  []int64 `json:"user_card_ids"`
	CurrentIndex int     `json:"current_index"`
}

// NewTrainingHandler creates a new training handler
func NewTrainingHandler(
	bot *tgbotapi.BotAPI,
	trainingService *service.TrainingService,
	srsService *service.SRSService,
	optionsService *service.OptionsService,
	logger *zap.Logger,
	optionsDelayMS int,
	wrongAnswerDelaySeconds int,
) *TrainingHandler {
	return &TrainingHandler{
		bot:                    bot,
		trainingService:        trainingService,
		srsService:             srsService,
		optionsService:         optionsService,
		logger:                 logger,
		sessions:               make(map[int64]*SessionState),
		optionsDelayMS:          optionsDelayMS,
		wrongAnswerDelaySeconds: wrongAnswerDelaySeconds,
	}
}

// StartTraining starts a training session for a user
func (h *TrainingHandler) StartTraining(ctx context.Context, chatID, userID int64, source models.SessionSource) error {
	// Check if there's an active session in memory
	h.sessionsMutex.RLock()
	existingState, existsInMemory := h.sessions[chatID]
	h.sessionsMutex.RUnlock()

	if existsInMemory && existingState != nil {
		// Session already exists in memory, just show current card
		return h.showCard(chatID)
	}

	// Try to restore session from database
	restored, err := h.restoreSession(chatID, userID)
	if err != nil {
		h.logger.Warn("failed to restore session", zap.Error(err))
	}
	if restored {
		// Session restored, show current card
		return h.showCard(chatID)
	}

	// Start new session
	session, queue, err := h.trainingService.StartSession(userID, source)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	// Create session state
	state := &SessionState{
		UserID:              userID,
		SessionID:           session.ID,
		Queue:               queue,
		CurrentIndex:        0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	// Store session
	h.sessionsMutex.Lock()
	h.sessions[chatID] = state
	h.sessionsMutex.Unlock()

	// Save state to database
	if err := h.saveSessionState(state); err != nil {
		h.logger.Warn("failed to save session state", zap.Error(err))
	}

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

	if !exists || state == nil {
		return fmt.Errorf("no active session")
	}

	if state.CurrentIndex >= len(state.Queue) {
		// Session finished
		return h.finishSession(chatID)
	}

	card := state.Queue[state.CurrentIndex]
	
	// Extract session words from other cards in the queue (for mixing into distractors)
	// Exclude recent correct answers to avoid "freshness recognition"
	sessionWords := h.extractSessionWords(state.Queue, state.CurrentIndex, card.UserCard.Direction, state.RecentCorrectAnswers)
	
	// Update state
	h.sessionsMutex.Lock()
	state.ShownAt = time.Now()
	state.OptionsShownAt = nil
	h.sessionsMutex.Unlock()

	// Generate options
	options, correctAnswer, err := h.optionsService.GenerateOptions(card, models.DefaultOptionCount, sessionWords)
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
			card.TrainingCard.WordRU,
		)
	} else {
		questionText = fmt.Sprintf(
			"🇷🇺 Что означает слово:\n\n*%s* %s",
			card.TrainingCard.WordEN,
			card.TrainingCard.Transcription,
		)
	}

	// Add progress
	questionText = fmt.Sprintf(
		"Карточка %d из %d\n\n%s",
		state.CurrentIndex+1,
		len(state.Queue),
		questionText,
	)

	msg := tgbotapi.NewMessage(chatID, questionText)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := h.bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Save state to database
	if err := h.saveSessionState(state); err != nil {
		h.logger.Warn("failed to save session state", zap.Error(err))
	}

	// Schedule automatic options reveal after delay
	go h.autoRevealOptions(chatID, h.optionsDelayMS)

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
	if !exists || state == nil {
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

	// Save state to database
	if err := h.saveSessionState(state); err != nil {
		h.logger.Warn("failed to save session state", zap.Error(err))
	}

	return nil
}

// HandleAnswer handles user's answer
func (h *TrainingHandler) HandleAnswer(chatID int64, optionIndex int) error {
	h.sessionsMutex.RLock()
	state, exists := h.sessions[chatID]
	h.sessionsMutex.RUnlock()

	if !exists || state == nil {
		return fmt.Errorf("no active session")
	}

	if state.CurrentIndex >= len(state.Queue) {
		return nil
	}

	card := state.Queue[state.CurrentIndex]
	answeredAt := time.Now()
	shownAt := state.ShownAt
	optionsShownAt := state.OptionsShownAt
	options := state.Options
	correctAnswer := state.CorrectAnswer

	// Check if card state is initialized
	// This can happen if session was restored but showCard wasn't called yet
	if len(options) == 0 || correctAnswer == "" {
		// Card wasn't shown yet, show it now
		h.logger.Warn("card state not initialized, showing card",
			zap.Int64("chat_id", chatID),
			zap.Int("current_index", state.CurrentIndex),
		)
		if err := h.showCard(chatID); err != nil {
			return fmt.Errorf("failed to show card: %w", err)
		}
		// Return error to indicate that user should wait for options
		return fmt.Errorf("card is being shown, please wait for options to appear")
	}

	// Validate option index
	if optionIndex < 0 || optionIndex >= len(options) {
		return fmt.Errorf("invalid option index: %d (valid range: 0-%d)", optionIndex, len(options)-1)
	}

	chosenOption := options[optionIndex]
	isCorrect := chosenOption == correctAnswer

	// Calculate timings
	var tDelayMS int
	var earlyReveal bool
	if optionsShownAt != nil {
		tDelayMS = int(optionsShownAt.Sub(shownAt).Milliseconds())
		earlyReveal = tDelayMS < h.optionsDelayMS
	} else {
		// User answered before options were shown (shouldn't happen)
		tDelayMS = h.optionsDelayMS
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
	} else {
		// Track correct answer for exclusion from future distractors
		h.sessionsMutex.Lock()
		if state.RecentCorrectAnswers == nil {
			state.RecentCorrectAnswers = make([]string, 0, 2)
		}
		// Add to front, keep only last 2
		state.RecentCorrectAnswers = append([]string{correctAnswer}, state.RecentCorrectAnswers...)
		if len(state.RecentCorrectAnswers) > 2 {
			state.RecentCorrectAnswers = state.RecentCorrectAnswers[:2]
		}
		h.sessionsMutex.Unlock()
	}

	// Send feedback
	h.sendFeedback(chatID, &card.TrainingCard, isCorrect, chosenOption, correctAnswer)

	// Move to next card
	h.sessionsMutex.Lock()
	state.CurrentIndex++
	h.sessionsMutex.Unlock()

	// Save state to database
	if err := h.saveSessionState(state); err != nil {
		h.logger.Warn("failed to save session state", zap.Error(err))
	}

	// Show next card after delay (only for wrong answers)
	if !isCorrect {
		time.Sleep(time.Duration(h.wrongAnswerDelaySeconds) * time.Second)
	}
	return h.showCard(chatID)
}

// sendFeedback sends feedback about the answer
func (h *TrainingHandler) sendFeedback(chatID int64, card *models.TrainingCard, isCorrect bool, chosen, correct string) {
	var message string

	if isCorrect {
		message = "✅ Правильно!"
	} else {
		message = fmt.Sprintf("❌ Неправильно\n\nВы выбрали: %s\nПравильный ответ: *%s*\n\n%s — %s",
			chosen, correct, card.WordEN, card.WordRU)
		
		// Show hint only after wrong answer
		if card.Hint != "" {
			message += fmt.Sprintf("\n\n💡 Подсказка: _%s_", card.Hint)
		}
		
		// Add example only for wrong answers
		if card.ExampleEN != "" {
			message += fmt.Sprintf("\n\n📝 %s\n_%s_", card.ExampleEN, card.ExampleRU)
		}
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

// extractSessionWords extracts correct answers from other cards in the session
// to be used as distractors (prevents guessing by word recognition)
// recentCorrectAnswers: list of recent correct answers to exclude (to avoid "freshness recognition")
// Excludes cards with the same WordCardID to avoid showing correct answers from other cards of the same word
func (h *TrainingHandler) extractSessionWords(queue []*models.UserCardWithTraining, currentIndex int, direction models.CardDirection, recentCorrectAnswers []string) []string {
	if currentIndex >= len(queue) {
		return []string{}
	}
	
	currentCard := queue[currentIndex]
	currentWordCardID := currentCard.TrainingCard.WordCardID
	
	sessionWords := make([]string, 0, len(queue))
	
	// Create a set of excluded words for fast lookup
	excludedSet := make(map[string]bool)
	for _, word := range recentCorrectAnswers {
		excludedSet[word] = true
	}
	
	// Track duplicates
	seenWords := make(map[string]bool)
	
	for i, card := range queue {
		// Skip current card
		if i == currentIndex {
			continue
		}
		
		// Skip cards from the same word (same WordCardID)
		// This prevents showing correct answers from other cards of the same word
		if card.TrainingCard.WordCardID == currentWordCardID {
			continue
		}
		
		// Extract correct answer based on direction
		var word string
		if direction == models.DirectionRUtoEN {
			// For RU->EN, collect English words
			word = card.TrainingCard.WordEN
		} else {
			// For EN->RU, collect Russian meanings
			word = card.TrainingCard.WordRU
		}
		
		// Add word if it's not in the excluded set and not a duplicate
		if word != "" && !excludedSet[word] && !seenWords[word] {
			sessionWords = append(sessionWords, word)
			seenWords[word] = true
		}
	}
	
	return sessionWords
}

// sendMessage is a helper to send text messages
func (h *TrainingHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	if _, err := h.bot.Send(msg); err != nil {
		h.logger.Error("failed to send message", zap.Error(err))
	}
}

// saveSessionState saves session state to database
func (h *TrainingHandler) saveSessionState(state *SessionState) error {
	// Extract user card IDs from queue
	userCardIDs := make([]int64, 0, len(state.Queue))
	for _, card := range state.Queue {
		userCardIDs = append(userCardIDs, card.UserCard.ID)
	}

	// Create state data
	stateData := SessionStateData{
		UserCardIDs:  userCardIDs,
		CurrentIndex: state.CurrentIndex,
	}

	// Get session from database to merge with existing session_json
	session, err := h.trainingService.GetSession(state.SessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}

	// Parse existing session_json to preserve config
	var sessionData map[string]interface{}
	if session.SessionJSON != "" {
		if err := json.Unmarshal([]byte(session.SessionJSON), &sessionData); err != nil {
			// If parsing fails, create new map
			sessionData = make(map[string]interface{})
		}
	} else {
		sessionData = make(map[string]interface{})
	}

	// Add state data
	sessionData["state"] = stateData

	// Serialize back to JSON
	mergedJSON, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal merged session data: %w", err)
	}

	// Update session in database
	return h.trainingService.UpdateSessionState(state.SessionID, string(mergedJSON))
}

// restoreSession restores session from database
func (h *TrainingHandler) restoreSession(chatID, userID int64) (bool, error) {
	// Get active session from database
	activeSession, err := h.trainingService.GetActiveSession(userID)
	if err != nil {
		return false, fmt.Errorf("failed to get active session: %w", err)
	}
	if activeSession == nil {
		return false, nil
	}

	// Parse session_json to get state
	var sessionData map[string]interface{}
	if activeSession.SessionJSON == "" {
		return false, nil
	}

	if err := json.Unmarshal([]byte(activeSession.SessionJSON), &sessionData); err != nil {
		return false, fmt.Errorf("failed to parse session_json: %w", err)
	}

	// Extract state data
	stateDataRaw, ok := sessionData["state"]
	if !ok {
		return false, nil
	}

	stateDataJSON, err := json.Marshal(stateDataRaw)
	if err != nil {
		return false, fmt.Errorf("failed to marshal state data: %w", err)
	}

	var stateData SessionStateData
	if err := json.Unmarshal(stateDataJSON, &stateData); err != nil {
		return false, fmt.Errorf("failed to unmarshal state data: %w", err)
	}

	// Restore queue from user card IDs
	queue, err := h.trainingService.RestoreQueue(userID, stateData.UserCardIDs)
	if err != nil {
		return false, fmt.Errorf("failed to restore queue: %w", err)
	}

	if len(queue) == 0 {
		// Queue is empty, finish session
		if err := h.trainingService.FinishSession(activeSession.ID, activeSession.DoneCount); err != nil {
			h.logger.Warn("failed to finish empty session", zap.Error(err))
		}
		return false, nil
	}

	// Validate current index
	currentIndex := stateData.CurrentIndex
	if currentIndex < 0 {
		currentIndex = 0
	}
	if currentIndex >= len(queue) {
		// Session was already finished, mark it as such
		if err := h.trainingService.FinishSession(activeSession.ID, len(queue)); err != nil {
			h.logger.Warn("failed to finish completed session", zap.Error(err))
		}
		return false, nil
	}

	// Create session state (without card state - will be set when showCard is called)
	state := &SessionState{
		UserID:              userID,
		SessionID:           activeSession.ID,
		Queue:               queue,
		CurrentIndex:        currentIndex,
		RecentCorrectAnswers: make([]string, 0, 2),
		// Options, CorrectAnswer, ShownAt, OptionsShownAt will be set when showCard is called
	}

	// Store session in memory
	h.sessionsMutex.Lock()
	h.sessions[chatID] = state
	h.sessionsMutex.Unlock()

	h.logger.Info("restored session from database",
		zap.Int64("session_id", activeSession.ID),
		zap.Int64("user_id", userID),
		zap.Int("current_index", state.CurrentIndex),
		zap.Int("queue_length", len(queue)),
	)

	// Note: showCard will be called by StartTraining after restoreSession returns true
	// This ensures that Options, CorrectAnswer, etc. are properly initialized
	return true, nil
}

// RestoreSession restores session from database (public method for handler)
func (h *TrainingHandler) RestoreSession(chatID, userID int64) (bool, error) {
	return h.restoreSession(chatID, userID)
}

