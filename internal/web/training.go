package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

// getTrainingDelaysForUser returns options delay in ms and wrong-answer delay in seconds from user settings, with config defaults when not set.
func (r *Router) getTrainingDelaysForUser(userID int64) (optionsDelayMS int, wrongAnswerDelaySeconds int) {
	optionsDelayMS = r.config.Training.OptionsDelayMS
	wrongAnswerDelaySeconds = r.config.Training.WrongAnswerDelaySeconds
	if r.userRepo == nil {
		return optionsDelayMS, wrongAnswerDelaySeconds
	}
	userRepo, ok := r.userRepo.(*repository.UserRepository)
	if !ok {
		return optionsDelayMS, wrongAnswerDelaySeconds
	}
	user, err := userRepo.GetUserByID(userID)
	if err != nil || user == nil || user.SettingsJSON == "" {
		return optionsDelayMS, wrongAnswerDelaySeconds
	}
	var settings models.UserSettings
	if err := json.Unmarshal([]byte(user.SettingsJSON), &settings); err != nil {
		return optionsDelayMS, wrongAnswerDelaySeconds
	}
	if settings.OptionsDelaySeconds != nil {
		optionsDelayMS = *settings.OptionsDelaySeconds * 1000
	}
	if settings.WrongAnswerDelaySeconds != nil {
		wrongAnswerDelaySeconds = *settings.WrongAnswerDelaySeconds
	}
	return optionsDelayMS, wrongAnswerDelaySeconds
}

// WebTrainingState holds the state of a web training session
type WebTrainingState struct {
	UserID              int64
	SessionID           int64
	Queue               []*models.UserCardWithTraining
	CurrentIndex        int
	ShownAt             time.Time
	OptionsShownAt      *time.Time
	Options             []string
	CorrectAnswer       string
	RecentCorrectAnswers []string
}

// WebTrainingHandler handles web training sessions
type WebTrainingHandler struct {
	trainingService      *service.TrainingService
	srsService           *service.SRSService
	optionsService       *service.OptionsService
	sessionRepo          *repository.SessionRepository
	logger               *zap.Logger
	optionsDelayMS       int
	wrongAnswerDelaySeconds int
	sessions              map[int64]*WebTrainingState
	sessionsMutex         sync.RWMutex
}

// NewWebTrainingHandler creates a new web training handler
func NewWebTrainingHandler(
	trainingService *service.TrainingService,
	srsService *service.SRSService,
	optionsService *service.OptionsService,
	sessionRepo *repository.SessionRepository,
	logger *zap.Logger,
	optionsDelayMS int,
	wrongAnswerDelaySeconds int,
) *WebTrainingHandler {
	return &WebTrainingHandler{
		trainingService:        trainingService,
		srsService:             srsService,
		optionsService:         optionsService,
		sessionRepo:            sessionRepo,
		logger:                 logger,
		optionsDelayMS:         optionsDelayMS,
		wrongAnswerDelaySeconds: wrongAnswerDelaySeconds,
		sessions:               make(map[int64]*WebTrainingState),
	}
}

// handleTrainingStart starts a new training session
// @Summary      Начать тренировку
// @Description  Создает новую сессию тренировки и возвращает первую карточку для изучения
// @Tags         Training
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Данные первой карточки"
// @Failure      400  {object}  map[string]interface{}  "Нет доступных карточек для тренировки"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/training/start [post]
func (r *Router) handleTrainingStart(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create web training handler if not exists
	if r.webTrainingHandler == nil {
		sessionRepo := repository.NewSessionRepository(r.db, r.logger)
		r.webTrainingHandler = NewWebTrainingHandler(
			r.trainingService,
			r.srsService,
			r.optionsService,
			sessionRepo,
			r.logger,
			r.config.Training.OptionsDelayMS,
			r.config.Training.WrongAnswerDelaySeconds,
		)
	}

	// Start session
	session, queue, err := r.trainingService.StartSession(userID, models.SourceManual)
	if err != nil {
		r.logger.Error("failed to start session", zap.Error(err))
		lang := i18n.GetLanguageFromContext(req.Context())
		if err.Error() == "no cards available for training" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": i18n.T(lang, "errors.noCardsAvailable"),
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": i18n.T(lang, "errors.internalError"),
		})
		return
	}

	// Create web state
	state := &WebTrainingState{
		UserID:              userID,
		SessionID:           session.ID,
		Queue:               queue,
		CurrentIndex:        0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	r.webTrainingHandler.sessionsMutex.Lock()
	r.webTrainingHandler.sessions[userID] = state
	r.webTrainingHandler.sessionsMutex.Unlock()

	// Show first card
	r.showTrainingCard(w, req, state)
}

// showTrainingCard shows the current training card
func (r *Router) showTrainingCard(w http.ResponseWriter, req *http.Request, state *WebTrainingState) {
	if state.CurrentIndex >= len(state.Queue) {
		// Session finished
		r.finishTrainingSession(w, req, state)
		return
	}

	card := state.Queue[state.CurrentIndex]
	
	// Extract session words for distractors (filtered by POS)
	sessionWords := r.extractSessionWords(state.Queue, state.CurrentIndex, card, state.RecentCorrectAnswers)
	
	// Extract WordEN and WordRU from all cards in session for distractor filtering
	sessionWordENs := make(map[string]bool)
	sessionWordRUs := make(map[string]bool)
	for i, sessionCard := range state.Queue {
		if i == state.CurrentIndex {
			continue
		}
		if sessionCard.TrainingCard.WordEN != "" {
			sessionWordENs[sessionCard.TrainingCard.WordEN] = true
		}
		if sessionCard.TrainingCard.WordRU != "" {
			sessionWordRUs[sessionCard.TrainingCard.WordRU] = true
		}
	}
	
	// Update state
	state.ShownAt = time.Now()
	state.OptionsShownAt = nil
	
	// Generate options
	options, correctAnswer, err := r.optionsService.GenerateOptions(card, models.DefaultOptionCount, sessionWords, sessionWordENs, sessionWordRUs)
	if err != nil {
		r.logger.Error("failed to generate options", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	state.Options = options
	state.CorrectAnswer = correctAnswer

	// Build question
	var questionText string
	if card.UserCard.Direction == models.DirectionRUtoEN {
		questionText = fmt.Sprintf("Переведите на английский: <strong>%s</strong>", card.TrainingCard.WordRU)
	} else {
		transcriptionHTML := ""
		if card.TrainingCard.Transcription != "" {
			transcriptionHTML = fmt.Sprintf(` <span class="transcription">%s</span>`, card.TrainingCard.Transcription)
		}
		// Use display_word if available (e.g., "to spy" for verbs)
		displayWord := card.TrainingCard.WordEN
		if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
			displayWord = *card.TrainingCard.DisplayWord
		}
		questionText = fmt.Sprintf("Что означает слово: <strong>%s</strong>%s", displayWord, transcriptionHTML)
	}

	optionsDelayMS, _ := r.getTrainingDelaysForUser(state.UserID)
	// Return card data as JSON
	response := map[string]interface{}{
		"question":     questionText,
		"card_index":    state.CurrentIndex + 1,
		"total_cards":   len(state.Queue),
		"session_id":    state.SessionID,
		"user_card_id":  card.UserCard.ID,
		"delay_ms":      optionsDelayMS,
		"direction":     string(card.UserCard.Direction),
	}
	
	// Add example_en if available (for showing example usage button)
	if card.TrainingCard.ExampleEN != "" {
		response["example_en"] = card.TrainingCard.ExampleEN
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// extractSessionWords extracts words from other cards in the session for use as distractors
// Excludes cards with the same WordCardID to avoid showing correct answers from other cards of the same word
// Filters by POS to ensure only words with matching part of speech are included
func (r *Router) extractSessionWords(queue []*models.UserCardWithTraining, currentIndex int, currentCard *models.UserCardWithTraining, recentCorrect []string) []string {
	if currentIndex >= len(queue) {
		return []string{}
	}
	
	currentWordCardID := currentCard.TrainingCard.WordCardID
	direction := currentCard.UserCard.Direction
	
	// Get POS of current card for filtering
	currentPOS := ""
	if currentCard.TrainingCard.POS != nil && *currentCard.TrainingCard.POS != "" {
		currentPOS = *currentCard.TrainingCard.POS
	}
	
	var sessionWords []string
	excludeSet := make(map[string]bool)
	seenWords := make(map[string]bool) // Track duplicates
	
	// Add recent correct answers to exclude set
	for _, word := range recentCorrect {
		excludeSet[word] = true
	}

	// Collect words from other cards (excluding cards with the same WordCardID)
	for i, card := range queue {
		if i == currentIndex {
			continue
		}
		
		// Skip cards from the same word (same WordCardID)
		// This prevents showing correct answers from other cards of the same word
		if card.TrainingCard.WordCardID == currentWordCardID {
			continue
		}
		
		// Filter by POS if current card has POS
		if currentPOS != "" {
			if card.TrainingCard.POS == nil || *card.TrainingCard.POS != currentPOS {
				continue
			}
		}
		
		// Exclude words that have the same English or Russian spelling as the current card
		// This prevents showing correct answers from other words with the same spelling
		// (e.g., "bug" and "beetle" both mean "жук" in Russian)
		if card.TrainingCard.WordEN == currentCard.TrainingCard.WordEN && currentCard.TrainingCard.WordEN != "" {
			continue
		}
		if card.TrainingCard.WordRU == currentCard.TrainingCard.WordRU && currentCard.TrainingCard.WordRU != "" {
			continue
		}
		
		var word string
		if direction == models.DirectionRUtoEN {
			// For RU->EN, use DisplayWord if available (e.g., "to spy" for verbs), otherwise WordEN
			if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
				word = *card.TrainingCard.DisplayWord
			} else {
				word = card.TrainingCard.WordEN
			}
		} else {
			word = card.TrainingCard.WordRU
		}
		
		// Exclude recent correct answers and duplicates
		if word != "" && !excludeSet[word] && !seenWords[word] {
			sessionWords = append(sessionWords, word)
			seenWords[word] = true
		}
	}

	return sessionWords
}

// handleTrainingCurrent shows the current card (for page refresh)
// @Summary      Получить текущую карточку
// @Description  Возвращает текущую карточку активной сессии тренировки (для обновления страницы)
// @Tags         Training
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Данные текущей карточки"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Router       /api/training/current [get]
func (r *Router) handleTrainingCurrent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.webTrainingHandler == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active":  false,
			"message": "No active session",
		})
		return
	}

	r.webTrainingHandler.sessionsMutex.RLock()
	state, exists := r.webTrainingHandler.sessions[userID]
	r.webTrainingHandler.sessionsMutex.RUnlock()

	if !exists || state == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active":  false,
			"message": "No active session",
		})
		return
	}

	r.showTrainingCard(w, req, state)
}

// handleTrainingReveal reveals the options
// @Summary      Показать варианты ответов
// @Description  Показывает варианты ответов для текущей карточки тренировки
// @Tags         Training
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Варианты ответов"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Нет активной сессии"
// @Failure      405  {string}  string  "Метод не разрешен"
// @Router       /api/training/reveal [post]
func (r *Router) handleTrainingReveal(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.webTrainingHandler == nil {
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	r.webTrainingHandler.sessionsMutex.Lock()
	state, exists := r.webTrainingHandler.sessions[userID]
	if !exists || state == nil {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	// Mark options as shown
	now := time.Now()
	state.OptionsShownAt = &now
	r.webTrainingHandler.sessionsMutex.Unlock()

	// Return options as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"options":      state.Options,
		"user_card_id": state.Queue[state.CurrentIndex].UserCard.ID,
	})
}

// handleTrainingAnswer handles the user's answer
// @Summary      Отправить ответ на карточку
// @Description  Обрабатывает ответ пользователя на карточку, обновляет SRS состояние и возвращает обратную связь
// @Tags         Training
// @Accept       application/x-www-form-urlencoded
// @Produce      application/json
// @Security     ApiKeyAuth
// @Param        option_index  formData  int  true  "Индекс выбранного варианта ответа"
// @Param        user_card_id  formData  int64  true  "ID карточки пользователя"
// @Success      200  {object}  map[string]interface{}  "Обратная связь и следующий шаг"
// @Failure      400  {string}  string  "Неверный запрос (неверные параметры)"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      404  {string}  string  "Нет активной сессии"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/training/answer [post]
func (r *Router) handleTrainingAnswer(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := req.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	optionIndexStr := req.FormValue("option_index")
	userCardIDStr := req.FormValue("user_card_id")

	optionIndex, err := strconv.Atoi(optionIndexStr)
	if err != nil {
		http.Error(w, "Invalid option_index", http.StatusBadRequest)
		return
	}

	userCardID, err := strconv.ParseInt(userCardIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user_card_id", http.StatusBadRequest)
		return
	}

	if r.webTrainingHandler == nil {
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	r.webTrainingHandler.sessionsMutex.Lock()
	state, exists := r.webTrainingHandler.sessions[userID]
	if !exists || state == nil || state.CurrentIndex >= len(state.Queue) {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	card := state.Queue[state.CurrentIndex]
	if card.UserCard.ID != userCardID {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "Card mismatch", http.StatusBadRequest)
		return
	}

	answeredAt := time.Now()
	shownAt := state.ShownAt
	optionsShownAt := state.OptionsShownAt
	options := state.Options
	correctAnswer := state.CorrectAnswer

	if optionIndex < 0 || optionIndex >= len(options) {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "Invalid option index", http.StatusBadRequest)
		return
	}

	chosenOption := options[optionIndex]
	isCorrect := chosenOption == correctAnswer

	// Calculate timings
	var tDelayMS int
	var earlyReveal bool
	if optionsShownAt != nil {
		tDelayMS = int(optionsShownAt.Sub(shownAt).Milliseconds())
		earlyReveal = tDelayMS < r.config.Training.OptionsDelayMS
	} else {
		tDelayMS = r.config.Training.OptionsDelayMS
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

	// Grade card (this updates the card)
	if err := r.srsService.GradeCard(&card.UserCard, attemptData); err != nil {
		r.logger.Error("failed to grade card", zap.Error(err))
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

	// Create review event
	optionsJSON, _ := json.Marshal(options)
	quality := models.CalculateQuality(attemptData)
	
	reviewEvent := &models.ReviewEvent{
		SessionID:      &state.SessionID,
		UserID:         userID,
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

	if _, err := r.webTrainingHandler.sessionRepo.CreateReviewEvent(reviewEvent); err != nil {
		r.logger.Error("failed to create review event", zap.Error(err))
	}

	// Record wrong answer if incorrect
	if !isCorrect {
		if err := r.srsService.RecordWrongAnswer(&card.UserCard, chosenOption); err != nil {
			r.logger.Error("failed to record wrong answer", zap.Error(err))
		}
	} else {
		// Track correct answer
		if state.RecentCorrectAnswers == nil {
			state.RecentCorrectAnswers = make([]string, 0, 2)
		}
		state.RecentCorrectAnswers = append([]string{correctAnswer}, state.RecentCorrectAnswers...)
		if len(state.RecentCorrectAnswers) > 2 {
			state.RecentCorrectAnswers = state.RecentCorrectAnswers[:2]
		}
	}

	// Move to next card
	state.CurrentIndex++
	r.webTrainingHandler.sessionsMutex.Unlock()

	// Show feedback and next card
	r.showTrainingFeedback(w, req, state, isCorrect, chosenOption, correctAnswer, card.TrainingCard)
}

// showTrainingFeedback shows feedback and moves to next card
func (r *Router) showTrainingFeedback(w http.ResponseWriter, req *http.Request, state *WebTrainingState, isCorrect bool, chosenOption, correctAnswer string, trainingCard models.TrainingCard) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	feedback := map[string]interface{}{
		"is_correct":     isCorrect,
		"chosen_option":  chosenOption,
		"correct_answer": correctAnswer,
	}

	if !isCorrect {
		if trainingCard.Hint != "" {
			feedback["hint"] = trainingCard.Hint
		}
		if trainingCard.ExampleEN != "" {
			feedback["example"] = trainingCard.ExampleEN
		}
		_, wrongAnswerDelaySeconds := r.getTrainingDelaysForUser(state.UserID)
		feedback["delay_seconds"] = wrongAnswerDelaySeconds
	}

	json.NewEncoder(w).Encode(feedback)
}

// finishTrainingSession finishes the training session
func (r *Router) finishTrainingSession(w http.ResponseWriter, req *http.Request, state *WebTrainingState) {
	// Finish session in DB
	if err := r.trainingService.FinishSession(state.SessionID, state.CurrentIndex); err != nil {
		r.logger.Error("failed to finish session", zap.Error(err))
	}

	// Get session statistics
	sessionRepo := repository.NewSessionRepository(r.db, r.logger)
	totalCards, correctCards, err := sessionRepo.GetSessionStats(state.SessionID)
	if err != nil {
		r.logger.Error("failed to get session stats", zap.Error(err))
		// Use fallback values
		totalCards = state.CurrentIndex
		correctCards = 0
	}

	// Remove from memory
	if r.webTrainingHandler != nil {
		r.webTrainingHandler.sessionsMutex.Lock()
		delete(r.webTrainingHandler.sessions, state.UserID)
		r.webTrainingHandler.sessionsMutex.Unlock()
	}

	// Show completion message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"complete":        true,
		"cards_completed": state.CurrentIndex,
		"total_cards":     totalCards,
		"correct_cards":   correctCards,
	})
}

// handleTrainingUpcoming gets upcoming cards count by date for the next 7 days
// @Summary      Получить количество карточек на ближайшую неделю
// @Description  Возвращает количество карточек, которые появятся в каждый день ближайшей недели
// @Tags         Training
// @Accept       json
// @Produce      application/json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "Карта дат и количества карточек"
// @Failure      401  {string}  string  "Неавторизован"
// @Failure      500  {string}  string  "Внутренняя ошибка сервера"
// @Router       /api/training/upcoming [get]
func (r *Router) handleTrainingUpcoming(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's timezone or use UTC
	userRepo := repository.NewUserRepository(r.db, r.logger)
	user, err := userRepo.GetUserByID(userID)
	if err != nil {
		r.logger.Error("failed to get user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get current time in user's timezone
	now := time.Now()
	if user != nil && user.Timezone != "" {
		loc, err := time.LoadLocation(user.Timezone)
		if err != nil {
			r.logger.Warn("failed to load timezone, using UTC", zap.String("timezone", user.Timezone), zap.Error(err))
		} else {
			now = now.In(loc)
		}
	}

	// Start from today at 00:00:00
	startDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Get upcoming cards by date
	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	upcomingCards, err := userCardRepo.GetUpcomingCardsByDate(userID, startDate)
	if err != nil {
		r.logger.Error("failed to get upcoming cards", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Format response with dates and labels
	response := make(map[string]interface{})
	
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		dateStr := date.Format("2006-01-02")
		
		// Format date as dd.mm
		label := date.Format("02.01")
		
		count := upcomingCards[dateStr]
		response[dateStr] = map[string]interface{}{
			"date":  dateStr,
			"label": label,
			"count": count,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

