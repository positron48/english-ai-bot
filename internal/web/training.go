package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

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
		if err.Error() == "no cards available for training" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `<div class="error">No cards available for training. Request some words first!</div>`)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
	
	// Extract session words for distractors
	sessionWords := r.extractSessionWords(state.Queue, state.CurrentIndex, card.UserCard.Direction, state.RecentCorrectAnswers)
	
	// Update state
	state.ShownAt = time.Now()
	state.OptionsShownAt = nil
	
	// Generate options
	options, correctAnswer, err := r.optionsService.GenerateOptions(card, models.DefaultOptionCount, sessionWords)
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
		questionText = fmt.Sprintf("Что означает слово: <strong>%s</strong> %s", card.TrainingCard.WordEN, card.TrainingCard.Transcription)
	}

	// Render card template
	r.renderTemplate(w, "training_card.html", map[string]interface{}{
		"Title":        "Training",
		"Question":     questionText,
		"CardIndex":    state.CurrentIndex + 1,
		"TotalCards":   len(state.Queue),
		"SessionID":    state.SessionID,
		"UserCardID":   card.UserCard.ID,
		"DelayMS":      r.config.Training.OptionsDelayMS,
	})
}

// extractSessionWords extracts words from other cards in the session for use as distractors
func (r *Router) extractSessionWords(queue []*models.UserCardWithTraining, currentIndex int, direction models.CardDirection, recentCorrect []string) []string {
	var sessionWords []string
	excludeSet := make(map[string]bool)
	
	// Add recent correct answers to exclude set
	for _, word := range recentCorrect {
		excludeSet[word] = true
	}

	// Collect words from other cards
	for i, card := range queue {
		if i == currentIndex {
			continue
		}
		
		var word string
		if direction == models.DirectionRUtoEN {
			word = card.TrainingCard.WordEN
		} else {
			word = card.TrainingCard.WordRU
		}
		
		if !excludeSet[word] {
			sessionWords = append(sessionWords, word)
		}
	}

	return sessionWords
}

// handleTrainingCurrent shows the current card (for page refresh)
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
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	r.webTrainingHandler.sessionsMutex.RLock()
	state, exists := r.webTrainingHandler.sessions[userID]
	r.webTrainingHandler.sessionsMutex.RUnlock()

	if !exists || state == nil {
		http.Error(w, "No active session", http.StatusNotFound)
		return
	}

	r.showTrainingCard(w, req, state)
}

// handleTrainingReveal reveals the options
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

	// Return options HTML
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	
	var optionsHTML strings.Builder
	for i, option := range state.Options {
		optionsHTML.WriteString(fmt.Sprintf(
			`<button class="option-btn" hx-post="/app/training/answer" hx-vals='{"option_index": %d, "user_card_id": %d}' hx-target="#training-content" hx-swap="innerHTML">%s</button>`,
			i, state.Queue[state.CurrentIndex].UserCard.ID, option,
		))
	}
	
	fmt.Fprintf(w, `<div class="options">%s</div>`, optionsHTML.String())
}

// handleTrainingAnswer handles the user's answer
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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	var feedbackHTML strings.Builder
	if isCorrect {
		feedbackHTML.WriteString(`<div class="feedback correct">✅ Correct!</div>`)
	} else {
		feedbackHTML.WriteString(fmt.Sprintf(
			`<div class="feedback incorrect">❌ Incorrect. Correct answer: <strong>%s</strong></div>`,
			correctAnswer,
		))
		if trainingCard.ExampleEN != "" {
			feedbackHTML.WriteString(fmt.Sprintf(
				`<div class="example">Example: %s</div>`,
				trainingCard.ExampleEN,
			))
		}
	}

	// If incorrect, add delay before next card
	if !isCorrect {
		feedbackHTML.WriteString(fmt.Sprintf(
			`<div class="next-card-delay" hx-trigger="load delay:%ds" hx-get="/app/training/current" hx-target="#training-content" hx-swap="innerHTML"></div>`,
			r.config.Training.WrongAnswerDelaySeconds,
		))
	} else {
		// If correct, show next card immediately
		feedbackHTML.WriteString(`<div hx-trigger="load" hx-get="/app/training/current" hx-target="#training-content" hx-swap="innerHTML"></div>`)
	}

	fmt.Fprint(w, feedbackHTML.String())
}

// finishTrainingSession finishes the training session
func (r *Router) finishTrainingSession(w http.ResponseWriter, req *http.Request, state *WebTrainingState) {
	// Finish session in DB
	if err := r.trainingService.FinishSession(state.SessionID, state.CurrentIndex); err != nil {
		r.logger.Error("failed to finish session", zap.Error(err))
	}

	// Remove from memory
	if r.webTrainingHandler != nil {
		r.webTrainingHandler.sessionsMutex.Lock()
		delete(r.webTrainingHandler.sessions, state.UserID)
		r.webTrainingHandler.sessionsMutex.Unlock()
	}

	// Show completion message
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<div class="session-complete">
		<h2>Training Complete!</h2>
		<p>You've completed %d cards.</p>
		<a href="/app/dashboard" class="btn btn-primary">Back to Dashboard</a>
	</div>`, state.CurrentIndex)
}

