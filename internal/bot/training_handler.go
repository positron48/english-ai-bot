package bot

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

// Test hooks for coverage (set by tests, must be nil in production).
var (
	testHookSaveSessionStateFail  func() error                      // if set, saveSessionState returns this before UpdateSessionState
	testHookGenerateOptionsErr    func() error                      // if set, showCard treats as GenerateOptions error and skips card
	testHookGenerateOptionsResult func() ([]string, string, error)  // if set, showCard uses this instead of calling GenerateOptions (to cover err branch)
	testHookRecordWrongAnswerErr  func() error                      // if set, HandleAnswer treats as RecordWrongAnswer error
	testHookMarshalSessionState   func(interface{}) ([]byte, error) // if set, saveSessionState uses it instead of json.Marshal
	testHookMarshalStateDataRaw   func(interface{}) ([]byte, error) // if set, restoreSession uses it for stateDataRaw
	testHookRestoreQueueErr       func() error                      // if set, restoreSession returns this before RestoreQueue
	testHookFinishSessionWarnErr  func() error                      // if set, restoreSession uses it to cover FinishSession error log (empty or completed session)
)

// TrainingHandler handles training sessions
type TrainingHandler struct {
	bot             *tgbotapi.BotAPI
	trainingService *service.TrainingService
	srsService      *service.SRSService
	optionsService  *service.OptionsService
	sessionRepo     interface {
		CreateReviewEvent(event *models.ReviewEvent) (int64, error)
		GetSessionStats(sessionID int64) (totalCards int, correctCards int, err error)
	}
	logger                  *zap.Logger
	sessions                map[int64]*SessionState
	sessionsMutex           sync.RWMutex
	optionsDelayMS          int
	wrongAnswerDelaySeconds int
	db                      *sql.DB
	linglowEventRepo        *repository.LinglowEventRepository
	learning                config.LearningConfig
	linglowEventsEnabled    bool
}

// SessionState holds the state of an active training session
type SessionState struct {
	UserID               int64
	SessionID            int64
	Queue                []*models.TrainingQueueItem
	CurrentIndex         int
	ShownAt              time.Time
	OptionsShownAt       *time.Time
	Options              []string
	CorrectAnswer        string
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
	sessionRepo interface {
		CreateReviewEvent(event *models.ReviewEvent) (int64, error)
		GetSessionStats(sessionID int64) (totalCards int, correctCards int, err error)
	},
	logger *zap.Logger,
	optionsDelayMS int,
	wrongAnswerDelaySeconds int,
	db *sql.DB,
) *TrainingHandler {
	return &TrainingHandler{
		bot:                     bot,
		trainingService:         trainingService,
		srsService:              srsService,
		optionsService:          optionsService,
		sessionRepo:             sessionRepo,
		logger:                  logger,
		sessions:                make(map[int64]*SessionState),
		optionsDelayMS:          optionsDelayMS,
		wrongAnswerDelaySeconds: wrongAnswerDelaySeconds,
		db:                      db,
	}
}

// SetLinglowEventWriter enables optional non-blocking dual-write into Linglow v2 event tables.
func (h *TrainingHandler) SetLinglowEventWriter(repo *repository.LinglowEventRepository, learning config.LearningConfig, enabled bool) {
	h.linglowEventRepo = repo
	h.learning = learning
	h.linglowEventsEnabled = enabled
}

func (h *TrainingHandler) recordLinglowWordReviewEvent(reviewEventID int64, event *models.ReviewEvent) {
	if h == nil || !h.linglowEventsEnabled || h.linglowEventRepo == nil || event == nil || reviewEventID == 0 {
		return
	}
	answeredAt := time.Now()
	if event.AnsweredAt != nil {
		answeredAt = *event.AnsweredAt
	}
	input := repository.WordReviewEventInput{
		UserID:          event.UserID,
		ReviewEventID:   reviewEventID,
		UserCardID:      event.UserCardID,
		ClientAttemptID: event.ClientAttemptID,
		Direction:       string(event.Direction),
		IsCorrect:       event.IsCorrect,
		Quality:         event.Quality,
		OptionsJSON:     event.OptionsJSON,
		ChosenOption:    event.ChosenOption,
		MetricsJSON:     event.MetricsJSON,
		SRSBeforeJSON:   event.SRSBeforeJSON,
		SRSAfterJSON:    event.SRSAfterJSON,
		AnsweredAt:      answeredAt,
	}
	if _, err := h.linglowEventRepo.RecordWordReviewEvent(context.Background(), h.learning, input); err != nil {
		h.logger.Warn("failed to dual-write linglow bot word review event",
			zap.Int64("user_id", event.UserID),
			zap.Int64("review_event_id", reviewEventID),
			zap.Int64("user_card_id", event.UserCardID),
			zap.Error(err),
		)
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
	session, queue, err := h.trainingService.StartSession(userID, source, nil)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	// Create session state
	state := &SessionState{
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	// Store session
	h.sessionsMutex.Lock()
	h.sessions[chatID] = state
	h.sessionsMutex.Unlock()

	// Save state to database
	if err := h.saveSessionState(state); err != nil {
		h.logger.Warn("failed to save session state", zap.Error(err)) // testHookSaveSessionStateFail can trigger this
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

	item := state.Queue[state.CurrentIndex]
	// Skip spell challenges in Telegram (only web supports them)
	if item.Type == "spell" {
		state.CurrentIndex++
		return h.showCard(chatID)
	}
	if item.Type != "card" || item.Card == nil {
		state.CurrentIndex++
		return h.showCard(chatID)
	}
	card := item.Card

	// Extract session words from other card items in the queue
	sessionWords := h.extractSessionWordsFromQueue(state.Queue, state.CurrentIndex, card, state.RecentCorrectAnswers)

	// Extract WordEN and WordRU from all cards in session for distractor filtering
	sessionWordENs := make(map[string]bool)
	sessionWordRUs := make(map[string]bool)
	for i, qi := range state.Queue {
		if i == state.CurrentIndex || qi.Type != "card" || qi.Card == nil {
			continue
		}
		c := qi.Card
		if c.TrainingCard.WordEN != "" {
			sessionWordENs[c.TrainingCard.WordEN] = true
		}
		if c.TrainingCard.WordRU != "" {
			sessionWordRUs[c.TrainingCard.WordRU] = true
		}
	}

	// Update state
	h.sessionsMutex.Lock()
	state.ShownAt = time.Now()
	state.OptionsShownAt = nil
	h.sessionsMutex.Unlock()

	// Generate options
	if testHookGenerateOptionsErr != nil {
		if err := testHookGenerateOptionsErr(); err != nil {
			h.logger.Error("failed to generate options", zap.Error(err))
			return h.skipCard(chatID, "Ошибка генерации вариантов")
		}
	}
	var options []string
	var correctAnswer string
	var err error
	if testHookGenerateOptionsResult != nil {
		options, correctAnswer, err = testHookGenerateOptionsResult()
	} else {
		options, correctAnswer, err = h.optionsService.GenerateOptions(card, models.DefaultOptionCount, sessionWords, sessionWordENs, sessionWordRUs)
	}
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
		// Use display_word if available (e.g., "to spy" for verbs)
		displayWord := card.TrainingCard.WordEN
		if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
			displayWord = *card.TrainingCard.DisplayWord
		}
		questionText = fmt.Sprintf(
			"🇷🇺 Что означает слово:\n\n*%s* %s",
			displayWord,
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

	item := state.Queue[state.CurrentIndex]
	if item.Type != "card" || item.Card == nil {
		return nil
	}
	card := item.Card
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

	var answerTimeMS int
	if optionsShownAt != nil {
		answerTimeMS = int(answeredAt.Sub(*optionsShownAt).Milliseconds())
	} else {
		answerTimeMS = 0
	}

	// Create attempt data
	attemptData := models.AttemptData{
		Correct:      isCorrect,
		EarlyReveal:  earlyReveal,
		AnswerTimeMS: answerTimeMS,
		TDelayMS:     tDelayMS,
		OptionCount:  len(options),
		ChosenOption: chosenOption,
	}

	// Capture SRS state before update
	srsBefore := models.SRSState{
		State:        card.UserCard.State,
		EF:           card.UserCard.EF,
		Reps:         card.UserCard.Reps,
		IntervalDays: card.UserCard.IntervalDays,
		LearningStep: card.UserCard.LearningStep,
		LapseCount:   card.UserCard.LapseCount,
	}
	srsBeforeJSON, _ := json.Marshal(srsBefore)

	// Grade card (persists SRS state to DB). If this fails, do not create review_event
	// so that review_events count stays in sync with user_cards state.
	if err := h.srsService.GradeCard(&card.UserCard, attemptData); err != nil {
		h.logger.Error("failed to grade card, progress not saved",
			zap.Int64("user_card_id", card.UserCard.ID),
			zap.Int64("user_id", state.UserID),
			zap.Error(err),
		)
		h.sendMessage(chatID, "⚠️ Не удалось сохранить прогресс. Попробуйте ответить ещё раз.")
		return h.showCard(chatID) // show same card again without advancing
	}

	// Capture SRS state after update
	srsAfter := models.SRSState{
		State:        card.UserCard.State,
		EF:           card.UserCard.EF,
		Reps:         card.UserCard.Reps,
		IntervalDays: card.UserCard.IntervalDays,
		LearningStep: card.UserCard.LearningStep,
		LapseCount:   card.UserCard.LapseCount,
	}
	srsAfterJSON, _ := json.Marshal(srsAfter)

	// Create metrics JSON
	metrics := map[string]interface{}{
		"answer_time_ms": answerTimeMS,
		"total_time_ms":  int(answeredAt.Sub(shownAt).Milliseconds()),
	}
	metricsJSON, _ := json.Marshal(metrics)

	// Record review event (needed for correct session stats)
	if h.sessionRepo != nil {
		optionsJSON, _ := json.Marshal(options)
		quality := models.CalculateQuality(attemptData)
		sessionID := state.SessionID

		reviewEvent := &models.ReviewEvent{
			SessionID:      &sessionID,
			UserID:         state.UserID,
			UserCardID:     card.UserCard.ID,
			Direction:      card.UserCard.Direction,
			ShownAt:        shownAt,
			OptionsShownAt: optionsShownAt,
			AnsweredAt:     &answeredAt,
			TDelayMS:       tDelayMS,
			EarlyReveal:    earlyReveal,
			OptionCount:    len(options),
			OptionsJSON:    string(optionsJSON),
			ChosenOption:   chosenOption,
			IsCorrect:      isCorrect,
			Quality:        int(quality),
			MetricsJSON:    string(metricsJSON),
			SRSBeforeJSON:  string(srsBeforeJSON),
			SRSAfterJSON:   string(srsAfterJSON),
		}

		if reviewEventID, err := h.sessionRepo.CreateReviewEvent(reviewEvent); err != nil {
			h.logger.Error("failed to create review event", zap.Error(err))
		} else {
			h.recordLinglowWordReviewEvent(reviewEventID, reviewEvent)
		}
	}

	// Record wrong answer if incorrect
	if !isCorrect {
		if testHookRecordWrongAnswerErr != nil {
			if err := testHookRecordWrongAnswerErr(); err != nil {
				h.logger.Error("failed to record wrong answer", zap.Error(err))
			}
		} else if err := h.srsService.RecordWrongAnswer(&card.UserCard, chosenOption); err != nil {
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

	// Get session statistics
	totalCards := doneCount
	correctCards := 0
	if h.sessionRepo != nil {
		total, correct, err := h.sessionRepo.GetSessionStats(sessionID)
		if err != nil {
			h.logger.Error("failed to get session stats", zap.Error(err))
		} else {
			totalCards = total
			correctCards = correct
		}
	}

	// Check how many cards are still available for training
	now := time.Now()
	var availableForTraining int
	if h.db != nil {
		// Get due count (cards ready for review, excluding new cards and orphaned cards)
		// Excludes words marked as "known" in user_word_knowledge
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
		err := h.db.QueryRow(dueQuery, state.UserID, now, state.UserID).Scan(&dueCount)
		if err != nil {
			h.logger.Error("failed to get due count", zap.Error(err))
			dueCount = 0
		}

		// Get new cards count (exclude orphaned cards)
		// Excludes words marked as "known" in user_word_knowledge
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
		err = h.db.QueryRow(newQuery, state.UserID, state.UserID).Scan(&newCount)
		if err != nil {
			h.logger.Error("failed to get new cards count", zap.Error(err))
			newCount = 0
		}

		availableForTraining = dueCount
		if newCount > 0 {
			availableForTraining += newCount
		}
	}

	// Send completion message with statistics
	var message string
	if totalCards > 0 {
		accuracy := (correctCards * 100) / totalCards
		message = fmt.Sprintf(
			"🎉 Тренировка завершена!\n\n"+
				"📊 Результаты:\n"+
				"• Всего карточек: %d\n"+
				"• Правильных ответов: %d\n"+
				"• Точность: %d%%",
			totalCards, correctCards, accuracy,
		)
	} else {
		message = fmt.Sprintf(
			"🎉 Тренировка завершена!\n\nВы прошли %d карточек.",
			doneCount,
		)
	}

	// Add message about available cards or goodbye message
	if availableForTraining > 0 {
		message += fmt.Sprintf(
			"\n\n💡 У вас еще %d карточек для тренировки.\n\nПродолжить тренировку? Используйте /train",
			availableForTraining,
		)
	} else {
		message += "\n\nОтличная работа! До встречи завтра."
	}

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

// extractSessionWordsFromQueue extracts correct answers from other card items in the session
func (h *TrainingHandler) extractSessionWordsFromQueue(queue []*models.TrainingQueueItem, currentIndex int, currentCard *models.UserCardWithTraining, recentCorrectAnswers []string) []string {
	if currentIndex >= len(queue) {
		return []string{}
	}
	currentWordCardID := currentCard.TrainingCard.WordCardID
	direction := currentCard.UserCard.Direction
	currentPOS := ""
	if currentCard.TrainingCard.POS != nil && *currentCard.TrainingCard.POS != "" {
		currentPOS = *currentCard.TrainingCard.POS
	}
	sessionWords := make([]string, 0, len(queue))
	excludedSet := make(map[string]bool)
	for _, word := range recentCorrectAnswers {
		excludedSet[word] = true
	}
	seenWords := make(map[string]bool)
	for i, qi := range queue {
		if i == currentIndex || qi.Type != "card" || qi.Card == nil {
			continue
		}
		card := qi.Card
		if card.TrainingCard.WordCardID == currentWordCardID {
			continue
		}
		if currentPOS != "" {
			if card.TrainingCard.POS == nil || *card.TrainingCard.POS != currentPOS {
				continue
			}
		}
		var word string
		if direction == models.DirectionRUtoEN {
			if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
				word = *card.TrainingCard.DisplayWord
			} else {
				word = card.TrainingCard.WordEN
			}
		} else {
			word = card.TrainingCard.WordRU
		}
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
	for _, qi := range state.Queue {
		if qi.Type == "card" && qi.Card != nil {
			userCardIDs = append(userCardIDs, qi.Card.UserCard.ID)
		}
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
	var mergedJSON []byte
	var errMarshal error
	if testHookMarshalSessionState != nil {
		mergedJSON, errMarshal = testHookMarshalSessionState(sessionData)
	} else {
		mergedJSON, errMarshal = json.Marshal(sessionData)
	}
	if errMarshal != nil {
		return fmt.Errorf("failed to marshal merged session data: %w", errMarshal)
	}

	if testHookSaveSessionStateFail != nil {
		if err := testHookSaveSessionStateFail(); err != nil {
			return err
		}
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

	var stateDataJSON []byte
	var errMarshalState error
	if testHookMarshalStateDataRaw != nil {
		stateDataJSON, errMarshalState = testHookMarshalStateDataRaw(stateDataRaw)
	} else {
		stateDataJSON, errMarshalState = json.Marshal(stateDataRaw)
	}
	if errMarshalState != nil {
		return false, fmt.Errorf("failed to marshal state data: %w", errMarshalState)
	}

	var stateData SessionStateData
	if err := json.Unmarshal(stateDataJSON, &stateData); err != nil {
		return false, fmt.Errorf("failed to unmarshal state data: %w", err)
	}

	// Restore queue from user card IDs (returns []*UserCardWithTraining)
	if testHookRestoreQueueErr != nil {
		if err := testHookRestoreQueueErr(); err != nil {
			return false, fmt.Errorf("failed to restore queue: %w", err)
		}
	}
	cardQueue, err := h.trainingService.RestoreQueue(userID, stateData.UserCardIDs)
	if err != nil {
		return false, fmt.Errorf("failed to restore queue: %w", err)
	}

	if len(cardQueue) == 0 {
		// Queue is empty, finish session
		if testHookFinishSessionWarnErr != nil {
			if err := testHookFinishSessionWarnErr(); err != nil {
				h.logger.Warn("failed to finish empty session", zap.Error(err))
			}
		} else if err := h.trainingService.FinishSession(activeSession.ID, activeSession.DoneCount); err != nil {
			h.logger.Warn("failed to finish empty session", zap.Error(err))
		}
		return false, nil
	}

	// Wrap in TrainingQueueItem (restored sessions only have cards, no spell items)
	queue := make([]*models.TrainingQueueItem, 0, len(cardQueue))
	for _, c := range cardQueue {
		queue = append(queue, &models.TrainingQueueItem{Type: "card", Card: c})
	}

	// Validate current index
	currentIndex := stateData.CurrentIndex
	if currentIndex < 0 {
		currentIndex = 0
	}
	if currentIndex >= len(queue) {
		// Session was already finished, mark it as such
		if testHookFinishSessionWarnErr != nil {
			if err := testHookFinishSessionWarnErr(); err != nil {
				h.logger.Warn("failed to finish completed session", zap.Error(err))
			}
		} else if err := h.trainingService.FinishSession(activeSession.ID, len(queue)); err != nil {
			h.logger.Warn("failed to finish completed session", zap.Error(err))
		}
		return false, nil
	}

	// Create session state (without card state - will be set when showCard is called)
	state := &SessionState{
		UserID:               userID,
		SessionID:            activeSession.ID,
		Queue:                queue,
		CurrentIndex:         currentIndex,
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
