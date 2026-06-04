package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/learning"
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
	UserID               int64
	SessionID            int64
	Queue                []*models.TrainingQueueItem
	CurrentIndex         int
	CorrectCount         int // deprecated: stats are taken only from review_events (each mode creates one event per answer)
	ShownAt              time.Time
	OptionsShownAt       *time.Time
	Options              []string
	CorrectAnswer        string
	RecentCorrectAnswers []string
}

// WebTrainingHandler handles web training sessions
type WebTrainingHandler struct {
	trainingService         *service.TrainingService
	srsService              *service.SRSService
	optionsService          *service.OptionsService
	sessionRepo             *repository.SessionRepository
	logger                  *zap.Logger
	optionsDelayMS          int
	wrongAnswerDelaySeconds int
	sessions                map[int64]*WebTrainingState
	sessionsMutex           sync.RWMutex
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
		trainingService:         trainingService,
		srsService:              srsService,
		optionsService:          optionsService,
		sessionRepo:             sessionRepo,
		logger:                  logger,
		optionsDelayMS:          optionsDelayMS,
		wrongAnswerDelaySeconds: wrongAnswerDelaySeconds,
		sessions:                make(map[int64]*WebTrainingState),
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
		var concreteSRS *service.SRSService
		if srs, ok := r.srsService.(*service.SRSService); ok {
			concreteSRS = srs
		}
		var concreteOpts *service.OptionsService
		if opts, ok := r.optionsService.(*service.OptionsService); ok {
			concreteOpts = opts
		}
		r.webTrainingHandler = NewWebTrainingHandler(
			r.trainingService,
			concreteSRS,
			concreteOpts,
			sessionRepo,
			r.logger,
			r.config.Training.OptionsDelayMS,
			r.config.Training.WrongAnswerDelaySeconds,
		)
	}

	// Build session config from user settings (spell mode and threshold)
	var sessionConfig *service.SessionConfig
	if r.userRepo != nil {
		if userRepo, ok := r.userRepo.(*repository.UserRepository); ok {
			user, _ := userRepo.GetUserByID(userID)
			if user != nil && user.SettingsJSON != "" {
				var settings models.UserSettings
				if json.Unmarshal([]byte(user.SettingsJSON), &settings) == nil {
					spellEnabled := true
					if settings.SpellModeEnabled != nil {
						spellEnabled = *settings.SpellModeEnabled
					}
					spellThreshold := 50
					if settings.SpellMasteringThreshold != nil {
						t := *settings.SpellMasteringThreshold
						if t < 0 {
							t = 0
						}
						if t > 100 {
							t = 100
						}
						spellThreshold = t
					}
					typeEnabled := true
					if settings.TypeModeEnabled != nil {
						typeEnabled = *settings.TypeModeEnabled
					}
					typeThreshold := 70
					if settings.TypeMasteringThreshold != nil {
						t := *settings.TypeMasteringThreshold
						if t < 0 {
							t = 0
						}
						if t > 100 {
							t = 100
						}
						typeThreshold = t
					}
					sessionConfig = &service.SessionConfig{
						MaxCardsPerSession:      models.DefaultMaxCardsPerSession,
						MaxNewPerSession:        models.DefaultMaxNewPerSession,
						AlgoVersion:             "srs_v2_delayed_mcq_sm2_autoquality",
						SpellEnabled:            spellEnabled,
						SpellMasteringThreshold: spellThreshold,
						TypeEnabled:             typeEnabled,
						TypeMasteringThreshold:  typeThreshold,
					}
				}
			}
		}
	}

	// Start session
	session, queue, err := r.trainingService.StartSession(userID, models.SourceManual, sessionConfig)
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
		UserID:               userID,
		SessionID:            session.ID,
		Queue:                queue,
		CurrentIndex:         0,
		RecentCorrectAnswers: make([]string, 0, 2),
	}

	r.webTrainingHandler.sessionsMutex.Lock()
	r.webTrainingHandler.sessions[userID] = state
	r.webTrainingHandler.sessionsMutex.Unlock()

	// Show first card
	r.showTrainingCard(w, req, state)
}

// showTrainingCard shows the current training card or spell challenge
func (r *Router) showTrainingCard(w http.ResponseWriter, req *http.Request, state *WebTrainingState) {
	if state.CurrentIndex >= len(state.Queue) {
		// Session finished
		r.finishTrainingSession(w, req, state)
		return
	}

	item := state.Queue[state.CurrentIndex]

	// Spell challenge: compose the word from letters
	if item.Type == "spell" && item.Spell != nil {
		state.ShownAt = time.Now()
		state.OptionsShownAt = nil
		tl := learning.TargetLangNameRUPrepositional(r.config.Learning.TargetLang)
		response := map[string]interface{}{
			"type":           "spell",
			"question":       fmt.Sprintf("Составьте слово на %s: <strong>%s</strong>", tl, item.Spell.WordRU),
			"word_ru":        item.Spell.WordRU,
			"word_native":    item.Spell.WordNative,
			"word_target":    item.Spell.WordTarget,
			"prefix":         item.Spell.Prefix,
			"letters":        item.Spell.ShuffledLetters,
			"correct_answer": item.Spell.DisplayWord,
			"card_index":     state.CurrentIndex + 1,
			"total_cards":    len(state.Queue),
			"session_id":     state.SessionID,
			"user_card_id":   0,
			"delay_ms":       0,
			"direction":      "spell",
			"word_card_id":   item.Spell.WordCardID,
		}
		if item.Spell.ReplacedUserCardID > 0 {
			response["user_card_id"] = item.Spell.ReplacedUserCardID
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Type challenge: type the word (no letter hints)
	if item.Type == "type" && item.TypeChallenge != nil {
		state.ShownAt = time.Now()
		state.OptionsShownAt = nil
		displayWord := item.TypeChallenge.DisplayWord
		prefix := ""
		wordForHint := displayWord
		if strings.HasPrefix(displayWord, "to ") && len(displayWord) > 3 {
			prefix = "to "
			wordForHint = displayWord[3:]
		}
		runes := []rune(wordForHint)
		hintFirstLetter := ""
		hintLength := 0
		if len(runes) > 0 {
			hintFirstLetter = string(runes[0])
			hintLength = len(runes)
		}
		tl := learning.TargetLangNameRUPrepositional(r.config.Learning.TargetLang)
		response := map[string]interface{}{
			"type":              "type",
			"question":          fmt.Sprintf("Введите слово на %s: <strong>%s</strong>", tl, item.TypeChallenge.WordRU),
			"word_ru":           item.TypeChallenge.WordRU,
			"word_native":       item.TypeChallenge.WordNative,
			"word_target":       item.TypeChallenge.WordTarget,
			"correct_answer":    displayWord,
			"prefix":            prefix,
			"hint_first_letter": hintFirstLetter,
			"hint_length":       hintLength,
			"card_index":        state.CurrentIndex + 1,
			"total_cards":       len(state.Queue),
			"session_id":        state.SessionID,
			"user_card_id":      0,
			"delay_ms":          0,
			"direction":         "type",
			"word_card_id":      item.TypeChallenge.WordCardID,
		}
		if item.TypeChallenge.ReplacedUserCardID > 0 {
			response["user_card_id"] = item.TypeChallenge.ReplacedUserCardID
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Normal card
	card := item.Card
	if card == nil {
		r.logger.Error("queue item is card type but Card is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Extract session words for distractors (only from card items)
	sessionWords := r.extractSessionWordsFromQueue(state.Queue, state.CurrentIndex, card, state.RecentCorrectAnswers)

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
	lang := i18n.GetLanguageFromContext(req.Context())
	displayWord := card.TrainingCard.WordEN
	if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
		displayWord = *card.TrainingCard.DisplayWord
	}
	var tl string
	switch lang {
	case "ru":
		tl = learning.TargetLangNameRUAccusative(r.config.Learning.TargetLang)
	case "es":
		tl = learning.TargetLangNameES(r.config.Learning.TargetLang)
	default:
		tl = learning.TargetLangNameEN(r.config.Learning.TargetLang)
	}
	if card.UserCard.Direction == models.DirectionRUtoEN {
		questionText = fmt.Sprintf(i18n.T(lang, "training.translateTo"), tl, card.TrainingCard.WordRU)
	} else {
		transcriptionHTML := ""
		if card.TrainingCard.Transcription != "" {
			transcriptionHTML = fmt.Sprintf(` <span class="transcription">%s</span>`, card.TrainingCard.Transcription)
		}
		questionText = fmt.Sprintf(i18n.T(lang, "training.whatMeansWord"), displayWord, transcriptionHTML)
	}

	optionsDelayMS, _ := r.getTrainingDelaysForUser(state.UserID)
	var morph *models.WordMorphInfo
	wordRepo := repository.NewWordRepository(r.db, r.logger)
	if wordCard, err := wordRepo.GetWordCardByID(card.TrainingCard.WordCardID); err == nil {
		morph = buildCompactMorphFromWordCard(r.config.Learning.TargetLang, wordCard, card.TrainingCard.POS)
	}
	// Return card data as JSON
	response := map[string]interface{}{
		"question":         questionText,
		"card_index":       state.CurrentIndex + 1,
		"total_cards":      len(state.Queue),
		"session_id":       state.SessionID,
		"user_card_id":     card.UserCard.ID,
		"delay_ms":         optionsDelayMS,
		"direction":        string(card.UserCard.Direction),
		"word_card_id":     card.TrainingCard.WordCardID,
		"training_card_id": card.TrainingCard.ID,
		"word_category": func() string {
			if card.TrainingCard.POS != nil {
				return *card.TrainingCard.POS
			}
			return ""
		}(),
	}
	if morph != nil {
		response["morph"] = morph
	}
	if card.UserCard.Direction == models.DirectionENtoRU {
		response["word_en"] = card.TrainingCard.WordEN
		response["word_target"] = card.TrainingCard.WordTarget
		response["display_word"] = displayWord
		response["display_target"] = displayWord
		response["transcription"] = card.TrainingCard.Transcription
	}
	if card.UserCard.Direction == models.DirectionRUtoEN {
		response["word_ru"] = card.TrainingCard.WordRU
		response["word_native"] = card.TrainingCard.WordNative
	}

	// Add example_en if available (for showing example usage button)
	if card.TrainingCard.ExampleEN != "" {
		response["example_en"] = card.TrainingCard.ExampleEN
		response["example_target"] = card.TrainingCard.ExampleTarget
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// extractSessionWordsFromQueue extracts words from other card items in the session for use as distractors
func (r *Router) extractSessionWordsFromQueue(queue []*models.TrainingQueueItem, currentIndex int, currentCard *models.UserCardWithTraining, recentCorrect []string) []string {
	if currentIndex >= len(queue) {
		return []string{}
	}
	currentWordCardID := currentCard.TrainingCard.WordCardID
	direction := currentCard.UserCard.Direction
	currentPOS := ""
	if currentCard.TrainingCard.POS != nil && *currentCard.TrainingCard.POS != "" {
		currentPOS = *currentCard.TrainingCard.POS
	}
	var sessionWords []string
	excludeSet := make(map[string]bool)
	seenWords := make(map[string]bool)
	for _, word := range recentCorrect {
		excludeSet[word] = true
	}
	for i, qi := range queue {
		if i == currentIndex || qi.Type != "card" || qi.Card == nil {
			continue
		}
		card := qi.Card
		if card.TrainingCard.WordCardID == currentWordCardID {
			continue
		}
		var word string
		if direction == models.DirectionRUtoEN {
			word = card.TrainingCard.WordEN
			if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
				word = *card.TrainingCard.DisplayWord
			}
		} else {
			word = card.TrainingCard.WordRU
		}
		if word == "" || excludeSet[word] || seenWords[word] {
			continue
		}
		if currentPOS != "" && (card.TrainingCard.POS == nil || *card.TrainingCard.POS != currentPOS) {
			continue
		}
		seenWords[word] = true
		sessionWords = append(sessionWords, word)
	}
	return sessionWords
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

	// Reveal only applies to card type
	item := state.Queue[state.CurrentIndex]
	if item.Type != "card" || item.Card == nil {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "Reveal not applicable to this card type", http.StatusBadRequest)
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
		"user_card_id": item.Card.UserCard.ID,
	})
}

// gradeReplacedCardForSpellType grades the user_card that was replaced by a spell/type challenge so SRS is updated and the card won't stay due.
// mode is "spell" or "type", wordLen is the length of the word (for type: longer = higher time multiplier).
func (r *Router) gradeReplacedCardForSpellType(userID int64, userCardID int64, isCorrect bool, chosenOption string, shownAt, answeredAt time.Time, sessionID int64, mode string, wordLen int) {
	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	userCard, err := userCardRepo.GetUserCard(userCardID)
	if err != nil || userCard == nil {
		r.logger.Warn("failed to load replaced user card for spell/type grade", zap.Int64("user_card_id", userCardID), zap.Error(err))
		return
	}
	answerTimeMS := 0
	if !answeredAt.IsZero() {
		answerTimeMS = int(answeredAt.Sub(shownAt).Milliseconds())
		if answerTimeMS < 0 {
			answerTimeMS = 0
		}
	}
	attemptData := models.AttemptData{
		Correct:        isCorrect,
		EarlyReveal:    false,
		AnswerTimeMS:   answerTimeMS,
		TDelayMS:       0,
		OptionCount:    1,
		ChosenOption:   chosenOption,
		TimeMultiplier: models.TimeMultiplierForMode(mode, wordLen),
	}
	srsBefore := models.SRSState{
		State:        userCard.State,
		EF:           userCard.EF,
		Reps:         userCard.Reps,
		IntervalDays: userCard.IntervalDays,
		LearningStep: userCard.LearningStep,
		LapseCount:   userCard.LapseCount,
	}
	srsBeforeJSON, _ := json.Marshal(srsBefore)
	if err := r.srsService.GradeCard(userCard, attemptData); err != nil {
		r.logger.Error("failed to grade replaced card after spell/type", zap.Int64("user_card_id", userCardID), zap.Error(err))
		return
	}
	srsAfter := models.SRSState{
		State:        userCard.State,
		EF:           userCard.EF,
		Reps:         userCard.Reps,
		IntervalDays: userCard.IntervalDays,
		LearningStep: userCard.LearningStep,
		LapseCount:   userCard.LapseCount,
	}
	srsAfterJSON, _ := json.Marshal(srsAfter)
	quality := models.CalculateQuality(attemptData)
	metricsJSON, _ := json.Marshal(map[string]interface{}{
		"spell_or_type":  true,
		"answer_time_ms": answerTimeMS,
		"mode":           mode,
		"word_len":       wordLen,
	})
	answeredAtPtr := answeredAt
	reviewEvent := &models.ReviewEvent{
		SessionID:      &sessionID,
		UserID:         userID,
		UserCardID:     userCardID,
		Direction:      userCard.Direction,
		ShownAt:        shownAt,
		OptionsShownAt: nil,
		AnsweredAt:     &answeredAtPtr,
		TDelayMS:       0,
		EarlyReveal:    false,
		OptionCount:    1,
		OptionsJSON:    "[]",
		ChosenOption:   chosenOption,
		IsCorrect:      isCorrect,
		Quality:        int(quality),
		MetricsJSON:    string(metricsJSON),
		SRSBeforeJSON:  string(srsBeforeJSON),
		SRSAfterJSON:   string(srsAfterJSON),
	}
	if reviewEventID, err := r.webTrainingHandler.sessionRepo.CreateReviewEvent(reviewEvent); err != nil {
		r.logger.Error("failed to create review event for spell/type", zap.Error(err))
	} else {
		r.recordLinglowWordReviewEvent(context.Background(), reviewEventID, reviewEvent)
	}
	if !isCorrect {
		if err := r.srsService.RecordWrongAnswer(userCard, chosenOption); err != nil {
			r.logger.Error("failed to record wrong answer for spell/type", zap.Error(err))
		}
	}
}

// handleTrainingSpellAnswer handles the answer for a spell (compose word) challenge
func (r *Router) handleTrainingSpellAnswer(w http.ResponseWriter, req *http.Request, userID int64, userAnswer string) {
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
	item := state.Queue[state.CurrentIndex]
	if item.Type != "spell" || item.Spell == nil {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "Not a spell challenge", http.StatusBadRequest)
		return
	}
	correctNorm := strings.TrimSpace(strings.ToLower(item.Spell.DisplayWord))
	userNorm := userAnswer
	if userNorm == "" {
		userNorm = " "
	}
	isCorrect := userNorm == correctNorm
	correctAnswer := item.Spell.DisplayWord
	replacedUserCardID := item.Spell.ReplacedUserCardID
	shownAt := state.ShownAt
	sessionID := state.SessionID
	answeredAt := time.Now()
	state.CurrentIndex++
	r.webTrainingHandler.sessionsMutex.Unlock()

	// Grade the replaced user_card so it gets next_due_at updated and doesn't reappear next session
	if replacedUserCardID != 0 {
		wordLen := len(item.Spell.DisplayWord)
		r.gradeReplacedCardForSpellType(userID, replacedUserCardID, isCorrect, userAnswer, shownAt, answeredAt, sessionID, "spell", wordLen)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	feedback := map[string]interface{}{
		"is_correct":     isCorrect,
		"chosen_option":  userAnswer,
		"correct_answer": correctAnswer,
	}
	_, wrongAnswerDelaySeconds := r.getTrainingDelaysForUser(userID)
	if !isCorrect {
		feedback["delay_seconds"] = wrongAnswerDelaySeconds
	}
	json.NewEncoder(w).Encode(feedback)
}

// handleTrainingTypeAnswer handles the answer for a type-the-word challenge (no letter hints)
func (r *Router) handleTrainingTypeAnswer(w http.ResponseWriter, req *http.Request, userID int64, userAnswer string) {
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
	item := state.Queue[state.CurrentIndex]
	if item.Type != "type" || item.TypeChallenge == nil {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "Not a type challenge", http.StatusBadRequest)
		return
	}
	correctNorm := strings.TrimSpace(strings.ToLower(item.TypeChallenge.DisplayWord))
	userNorm := userAnswer
	if userNorm == "" {
		userNorm = " "
	}
	isCorrect := userNorm == correctNorm
	correctAnswer := item.TypeChallenge.DisplayWord
	replacedUserCardID := item.TypeChallenge.ReplacedUserCardID
	shownAt := state.ShownAt
	sessionID := state.SessionID
	answeredAt := time.Now()
	state.CurrentIndex++
	r.webTrainingHandler.sessionsMutex.Unlock()

	// Grade the replaced user_card so it gets next_due_at updated and doesn't reappear next session
	if replacedUserCardID != 0 {
		wordLen := len(item.TypeChallenge.DisplayWord)
		r.gradeReplacedCardForSpellType(userID, replacedUserCardID, isCorrect, userAnswer, shownAt, answeredAt, sessionID, "type", wordLen)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	feedback := map[string]interface{}{
		"is_correct":     isCorrect,
		"chosen_option":  userAnswer,
		"correct_answer": correctAnswer,
	}
	_, wrongAnswerDelaySeconds := r.getTrainingDelaysForUser(userID)
	if !isCorrect {
		feedback["delay_seconds"] = wrongAnswerDelaySeconds
	}
	json.NewEncoder(w).Encode(feedback)
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

	answerText := req.FormValue("answer_text")
	// answer_text present (including empty for "skip" in spell/type) → spell or type handler
	if req.Form.Has("answer_text") && r.webTrainingHandler != nil {
		r.webTrainingHandler.sessionsMutex.Lock()
		state, exists := r.webTrainingHandler.sessions[userID]
		if exists && state != nil && state.CurrentIndex < len(state.Queue) {
			item := state.Queue[state.CurrentIndex]
			r.webTrainingHandler.sessionsMutex.Unlock()
			text := strings.TrimSpace(strings.ToLower(answerText))
			if item.Type == "type" {
				r.handleTrainingTypeAnswer(w, req, userID, text)
				return
			}
			if item.Type == "spell" {
				r.handleTrainingSpellAnswer(w, req, userID, text)
				return
			}
		} else {
			r.webTrainingHandler.sessionsMutex.Unlock()
		}
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

	item := state.Queue[state.CurrentIndex]
	if item.Type != "card" || item.Card == nil {
		r.webTrainingHandler.sessionsMutex.Unlock()
		http.Error(w, "Not a card answer", http.StatusBadRequest)
		return
	}
	card := item.Card
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

	// Timings: for quality we use only time from options shown to answer (delay before options not counted).
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
		if answerTimeMS < 0 {
			answerTimeMS = 0
		}
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

	// Grade card (this updates the card in DB). If this fails, do not create review_event
	// so that review_events count stays in sync with actual user_cards state.
	if err := r.srsService.GradeCard(&card.UserCard, attemptData); err != nil {
		r.logger.Error("failed to grade card, progress not saved",
			zap.Int64("user_card_id", card.UserCard.ID),
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		r.webTrainingHandler.sessionsMutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error":          "failed_to_save_progress",
			"is_correct":     isCorrect,
			"chosen_option":  chosenOption,
			"correct_answer": correctAnswer,
		})
		return
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

	// Create review event only if user_card still exists (it may have been deleted e.g. by admin during session)
	optionsJSON, _ := json.Marshal(options)
	quality := models.CalculateQuality(attemptData)

	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	existingCard, err := userCardRepo.GetUserCard(card.UserCard.ID)
	if err != nil {
		r.logger.Warn("failed to check user card before review event", zap.Int64("user_card_id", card.UserCard.ID), zap.Error(err))
	} else if existingCard == nil {
		r.logger.Warn("user card no longer exists, skipping review event (e.g. deleted during session)",
			zap.Int64("user_card_id", card.UserCard.ID),
			zap.Int64("user_id", userID),
			zap.Int64("session_id", state.SessionID),
		)
	} else {
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
		if reviewEventID, err := r.webTrainingHandler.sessionRepo.CreateReviewEvent(reviewEvent); err != nil {
			r.logger.Error("failed to create review event", zap.Error(err))
		} else {
			r.recordLinglowWordReviewEvent(req.Context(), reviewEventID, reviewEvent)
		}
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

func (r *Router) recordLinglowWordReviewEvent(ctx context.Context, reviewEventID int64, event *models.ReviewEvent) {
	if r == nil || r.config == nil || !r.config.Linglow.EventsWriteEnabled || r.linglowEventRepo == nil || event == nil || reviewEventID == 0 {
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
	if _, err := r.linglowEventRepo.RecordWordReviewEvent(ctx, r.config.Learning, input); err != nil {
		r.logger.Warn("failed to dual-write linglow word review event",
			zap.Int64("user_id", event.UserID),
			zap.Int64("review_event_id", reviewEventID),
			zap.Int64("user_card_id", event.UserCardID),
			zap.Error(err),
		)
	}
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
			feedback["example_target"] = trainingCard.ExampleTarget
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

	// Session statistics: one source of truth — review_events (card, spell, type each create exactly one event per answer)
	sessionRepo := repository.NewSessionRepository(r.db, r.logger)
	totalCards, correctCards, err := sessionRepo.GetSessionStats(state.SessionID)
	if err != nil {
		r.logger.Error("failed to get session stats", zap.Error(err))
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
