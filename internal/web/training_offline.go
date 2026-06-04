package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/learning"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

type offlineWordTrainingCard struct {
	Question       string                 `json:"question"`
	UserCardID     int64                  `json:"user_card_id"`
	TrainingCardID int64                  `json:"training_card_id"`
	WordCardID     int64                  `json:"word_card_id"`
	Direction      string                 `json:"direction"`
	WordEN         string                 `json:"word_en,omitempty"`
	WordTarget     string                 `json:"word_target,omitempty"`
	WordRU         string                 `json:"word_ru,omitempty"`
	WordNative     string                 `json:"word_native,omitempty"`
	DisplayWord    string                 `json:"display_word,omitempty"`
	DisplayTarget  string                 `json:"display_target,omitempty"`
	Transcription  string                 `json:"transcription,omitempty"`
	ExampleEN      string                 `json:"example_en,omitempty"`
	ExampleTarget  string                 `json:"example_target,omitempty"`
	Hint           string                 `json:"hint,omitempty"`
	WordCategory   string                 `json:"word_category,omitempty"`
	Morph          *models.WordMorphInfo  `json:"morph,omitempty"`
	Options        []string               `json:"options"`
	CorrectAnswer  string                 `json:"correct_answer"`
	SRS            map[string]interface{} `json:"srs"`
}

type offlineWordTrainingAttempt struct {
	ClientAttemptID string    `json:"client_attempt_id"`
	UserCardID      int64     `json:"user_card_id"`
	TrainingCardID  int64     `json:"training_card_id"`
	Direction       string    `json:"direction"`
	Mode            string    `json:"mode"`
	ShownAt         time.Time `json:"shown_at"`
	OptionsShownAt  time.Time `json:"options_shown_at"`
	AnsweredAt      time.Time `json:"answered_at"`
	TDelayMS        int       `json:"t_delay_ms"`
	AnswerTimeMS    int       `json:"answer_time_ms"`
	EarlyReveal     bool      `json:"early_reveal"`
	Options         []string  `json:"options"`
	ChosenOption    string    `json:"chosen_option"`
	CorrectAnswer   string    `json:"correct_answer"`
}

func (r *Router) handleTrainingOfflinePack(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.optionsService == nil {
		http.Error(w, "Options service unavailable", http.StatusInternalServerError)
		return
	}

	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)
	now := time.Now()
	dueCards, err := userCardRepo.GetDueCards(userID, now, models.MaxDuePoolSize)
	if err != nil {
		r.logger.Error("failed to get offline due word cards", zap.Int64("user_id", userID), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	newCards, err := userCardRepo.GetNewCards(userID, 300)
	if err != nil {
		r.logger.Error("failed to get offline new word cards", zap.Int64("user_id", userID), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	seen := make(map[int64]bool)
	pool := make([]*models.UserCard, 0, len(dueCards)+len(newCards))
	for _, card := range append(dueCards, newCards...) {
		if card == nil || seen[card.ID] {
			continue
		}
		seen[card.ID] = true
		pool = append(pool, card)
	}

	queue := make([]*models.UserCardWithTraining, 0, len(pool))
	for _, userCard := range pool {
		trainingCard, err := trainingCardRepo.GetTrainingCard(userCard.TrainingCardID)
		if err != nil || trainingCard == nil {
			r.logger.Warn("failed to load offline training card", zap.Int64("user_card_id", userCard.ID), zap.Int64("training_card_id", userCard.TrainingCardID), zap.Error(err))
			continue
		}
		queue = append(queue, &models.UserCardWithTraining{UserCard: *userCard, TrainingCard: *trainingCard})
	}

	cards := make([]offlineWordTrainingCard, 0, len(queue))
	lang := i18n.GetLanguageFromContext(req.Context())
	for i, card := range queue {
		options, correctAnswer, err := r.optionsService.GenerateOptions(card, models.DefaultOptionCount, r.extractSessionWords(queue, i, card, nil), collectWordENs(queue, i), collectWordRUs(queue, i))
		if err != nil {
			r.logger.Warn("failed to generate offline word options", zap.Int64("user_card_id", card.UserCard.ID), zap.Error(err))
			continue
		}
		cards = append(cards, r.buildOfflineWordTrainingCard(req, lang, card, options, correctAnswer))
	}

	response := map[string]interface{}{
		"app_code":        r.config.Learning.AppCode,
		"native_lang":     r.config.Learning.NativeLang,
		"target_lang":     r.config.Learning.TargetLang,
		"generated_at":    time.Now().UTC().Format(time.RFC3339),
		"algo_version":    "word_training_offline_v1_mcq",
		"total_cards":     len(cards),
		"available_count": len(cards),
		"cards":           cards,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func collectWordENs(queue []*models.UserCardWithTraining, excludeIndex int) map[string]bool {
	out := make(map[string]bool)
	for i, card := range queue {
		if i == excludeIndex || card == nil {
			continue
		}
		if card.TrainingCard.WordEN != "" {
			out[card.TrainingCard.WordEN] = true
		}
	}
	return out
}

func collectWordRUs(queue []*models.UserCardWithTraining, excludeIndex int) map[string]bool {
	out := make(map[string]bool)
	for i, card := range queue {
		if i == excludeIndex || card == nil {
			continue
		}
		if card.TrainingCard.WordRU != "" {
			out[card.TrainingCard.WordRU] = true
		}
	}
	return out
}

func (r *Router) buildOfflineWordTrainingCard(req *http.Request, lang string, card *models.UserCardWithTraining, options []string, correctAnswer string) offlineWordTrainingCard {
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
	question := ""
	if card.UserCard.Direction == models.DirectionRUtoEN {
		question = fmt.Sprintf(i18n.T(lang, "training.translateTo"), tl, card.TrainingCard.WordRU)
	} else {
		transcriptionHTML := ""
		if card.TrainingCard.Transcription != "" {
			transcriptionHTML = fmt.Sprintf(` <span class="transcription">%s</span>`, card.TrainingCard.Transcription)
		}
		question = fmt.Sprintf(i18n.T(lang, "training.whatMeansWord"), displayWord, transcriptionHTML)
	}

	item := offlineWordTrainingCard{
		Question:       question,
		UserCardID:     card.UserCard.ID,
		TrainingCardID: card.TrainingCard.ID,
		WordCardID:     card.TrainingCard.WordCardID,
		Direction:      string(card.UserCard.Direction),
		Options:        options,
		CorrectAnswer:  correctAnswer,
		ExampleEN:      card.TrainingCard.ExampleEN,
		ExampleTarget:  card.TrainingCard.ExampleTarget,
		Hint:           card.TrainingCard.Hint,
		SRS: map[string]interface{}{
			"state":         card.UserCard.State,
			"ef":            card.UserCard.EF,
			"reps":          card.UserCard.Reps,
			"interval_days": card.UserCard.IntervalDays,
			"learning_step": card.UserCard.LearningStep,
			"lapse_count":   card.UserCard.LapseCount,
		},
	}
	if card.TrainingCard.POS != nil {
		item.WordCategory = *card.TrainingCard.POS
	}
	if card.UserCard.Direction == models.DirectionENtoRU {
		item.WordEN = card.TrainingCard.WordEN
		item.WordTarget = card.TrainingCard.WordTarget
		item.DisplayWord = displayWord
		item.DisplayTarget = displayWord
		item.Transcription = card.TrainingCard.Transcription
	} else {
		item.WordRU = card.TrainingCard.WordRU
		item.WordNative = card.TrainingCard.WordNative
	}
	wordRepo := repository.NewWordRepository(r.db, r.logger)
	if wordCard, err := wordRepo.GetWordCardByID(card.TrainingCard.WordCardID); err == nil {
		item.Morph = buildCompactMorphFromWordCard(r.config.Learning.TargetLang, wordCard, card.TrainingCard.POS)
	}
	return item
}

func (r *Router) handleTrainingOfflineSyncAttempts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var payload struct {
		Attempts []offlineWordTrainingAttempt `json:"attempts"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	sessionRepo := repository.NewSessionRepository(r.db, r.logger)
	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	trainingCardRepo := repository.NewTrainingCardRepository(r.db, r.logger)
	masteringRepo := repository.NewUserWordMasteringRepository(r.db, r.logger)

	sessionID := int64(0)
	syncedCount := 0
	results := make([]map[string]interface{}, 0, len(payload.Attempts))
	for _, attempt := range payload.Attempts {
		result := map[string]interface{}{"client_attempt_id": attempt.ClientAttemptID}
		if attempt.ClientAttemptID == "" || attempt.UserCardID == 0 {
			result["synced"] = false
			result["error"] = "invalid_attempt"
			results = append(results, result)
			continue
		}
		exists, err := sessionRepo.HasReviewEventClientAttempt(userID, attempt.ClientAttemptID)
		if err != nil {
			result["synced"] = false
			result["error"] = "idempotency_check_failed"
			results = append(results, result)
			continue
		}
		if exists {
			result["synced"] = true
			result["duplicate"] = true
			results = append(results, result)
			continue
		}

		userCard, err := userCardRepo.GetUserCard(attempt.UserCardID)
		if err != nil || userCard == nil || userCard.UserID != userID {
			result["synced"] = false
			result["error"] = "user_card_not_found"
			results = append(results, result)
			continue
		}
		trainingCard, err := trainingCardRepo.GetTrainingCard(userCard.TrainingCardID)
		if err != nil || trainingCard == nil {
			result["synced"] = false
			result["error"] = "training_card_not_found"
			results = append(results, result)
			continue
		}
		if attempt.TrainingCardID != 0 && attempt.TrainingCardID != userCard.TrainingCardID {
			result["synced"] = false
			result["error"] = "training_card_mismatch"
			results = append(results, result)
			continue
		}
		if sessionID == 0 {
			session := &models.TrainingSession{UserID: userID, Source: models.SourceManual, PlannedCount: len(payload.Attempts), SessionJSON: `{"offline_sync":true}`}
			newID, err := sessionRepo.CreateSession(session)
			if err != nil {
				result["synced"] = false
				result["error"] = "session_create_failed"
				results = append(results, result)
				continue
			}
			sessionID = newID
		}

		chosenOption := attempt.ChosenOption
		correctAnswer := attempt.CorrectAnswer
		if correctAnswer == "" {
			_, generatedCorrect, _ := r.optionsService.GenerateOptions(&models.UserCardWithTraining{UserCard: *userCard, TrainingCard: *trainingCard}, models.DefaultOptionCount, nil, nil, nil)
			correctAnswer = generatedCorrect
		}
		isCorrect := chosenOption == correctAnswer
		shownAt := fallbackTime(attempt.ShownAt)
		answeredAt := fallbackTime(attempt.AnsweredAt)
		optionsShownAt := attempt.OptionsShownAt
		if optionsShownAt.IsZero() {
			optionsShownAt = shownAt
		}
		answerTimeMS := attempt.AnswerTimeMS
		if answerTimeMS <= 0 {
			answerTimeMS = int(answeredAt.Sub(optionsShownAt).Milliseconds())
			if answerTimeMS < 0 {
				answerTimeMS = 0
			}
		}
		tDelayMS := attempt.TDelayMS
		if tDelayMS <= 0 {
			tDelayMS = int(optionsShownAt.Sub(shownAt).Milliseconds())
			if tDelayMS < 0 {
				tDelayMS = 0
			}
		}

		attemptData := models.AttemptData{Correct: isCorrect, EarlyReveal: attempt.EarlyReveal, AnswerTimeMS: answerTimeMS, TDelayMS: tDelayMS, OptionCount: len(attempt.Options), ChosenOption: chosenOption}
		srsBefore := models.SRSState{State: userCard.State, EF: userCard.EF, Reps: userCard.Reps, IntervalDays: userCard.IntervalDays, LearningStep: userCard.LearningStep, LapseCount: userCard.LapseCount}
		srsBeforeJSON, _ := json.Marshal(srsBefore)
		if err := r.srsService.GradeCard(userCard, attemptData); err != nil {
			result["synced"] = false
			result["error"] = "grade_failed"
			results = append(results, result)
			continue
		}
		srsAfter := models.SRSState{State: userCard.State, EF: userCard.EF, Reps: userCard.Reps, IntervalDays: userCard.IntervalDays, LearningStep: userCard.LearningStep, LapseCount: userCard.LapseCount}
		srsAfterJSON, _ := json.Marshal(srsAfter)
		optionsJSON, _ := json.Marshal(attempt.Options)
		metricsJSON, _ := json.Marshal(map[string]interface{}{"offline_sync": true, "answer_time_ms": answerTimeMS, "total_time_ms": int(answeredAt.Sub(shownAt).Milliseconds()), "mode": "card"})
		quality := models.CalculateQuality(attemptData)
		reviewEvent := &models.ReviewEvent{
			SessionID:       &sessionID,
			UserID:          userID,
			UserCardID:      userCard.ID,
			ClientAttemptID: attempt.ClientAttemptID,
			Direction:       userCard.Direction,
			ShownAt:         shownAt,
			OptionsShownAt:  &optionsShownAt,
			AnsweredAt:      &answeredAt,
			TDelayMS:        tDelayMS,
			EarlyReveal:     attempt.EarlyReveal,
			OptionCount:     len(attempt.Options),
			OptionsJSON:     string(optionsJSON),
			ChosenOption:    chosenOption,
			IsCorrect:       isCorrect,
			Quality:         int(quality),
			MetricsJSON:     string(metricsJSON),
			SRSBeforeJSON:   string(srsBeforeJSON),
			SRSAfterJSON:    string(srsAfterJSON),
		}
		if reviewEventID, err := sessionRepo.CreateReviewEvent(reviewEvent); err != nil {
			result["synced"] = false
			result["error"] = "review_event_create_failed"
			results = append(results, result)
			continue
		} else {
			r.recordLinglowWordReviewEvent(req.Context(), reviewEventID, reviewEvent)
		}
		if !isCorrect {
			if err := r.srsService.RecordWrongAnswer(userCard, chosenOption); err != nil {
				r.logger.Warn("failed to record offline wrong answer", zap.Int64("user_card_id", userCard.ID), zap.Error(err))
			}
		}
		syncedCount++
		result["synced"] = true
		result["is_correct"] = isCorrect
		results = append(results, result)
	}
	if sessionID != 0 {
		if err := r.trainingService.FinishSession(sessionID, syncedCount); err != nil {
			r.logger.Warn("failed to finish offline word training sync session", zap.Int64("session_id", sessionID), zap.Error(err))
		}
		if syncedCount > 0 {
			pairs, err := masteringRepo.GetWordCardIDsBySessionID(sessionID)
			if err == nil {
				r.logger.Info("synced offline word training attempts", zap.Int64("user_id", userID), zap.Int("synced", syncedCount), zap.Int("words", len(pairs)))
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": results, "synced": syncedCount})
}

func fallbackTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now()
	}
	return value
}
