package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/learning"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

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

type offlineWordTrainingQueueItem struct {
	Type            string                 `json:"type"`
	Question        string                 `json:"question,omitempty"`
	UserCardID      int64                  `json:"user_card_id"`
	TrainingCardID  int64                  `json:"training_card_id,omitempty"`
	WordCardID      int64                  `json:"word_card_id,omitempty"`
	Direction       string                 `json:"direction"`
	WordEN          string                 `json:"word_en,omitempty"`
	WordTarget      string                 `json:"word_target,omitempty"`
	WordRU          string                 `json:"word_ru,omitempty"`
	WordNative      string                 `json:"word_native,omitempty"`
	DisplayWord     string                 `json:"display_word,omitempty"`
	DisplayTarget   string                 `json:"display_target,omitempty"`
	Transcription   string                 `json:"transcription,omitempty"`
	ExampleEN       string                 `json:"example_en,omitempty"`
	ExampleTarget   string                 `json:"example_target,omitempty"`
	Hint            string                 `json:"hint,omitempty"`
	WordCategory    string                 `json:"word_category,omitempty"`
	Morph           *models.WordMorphInfo  `json:"morph,omitempty"`
	Options         []string               `json:"options,omitempty"`
	CorrectAnswer   string                 `json:"correct_answer"`
	Prefix          string                 `json:"prefix,omitempty"`
	Letters         []string               `json:"letters,omitempty"`
	HintFirstLetter string                 `json:"hint_first_letter,omitempty"`
	HintLength      int                    `json:"hint_length,omitempty"`
	SRS             map[string]interface{} `json:"srs,omitempty"`
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
	AnswerText      string    `json:"answer_text"`
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
	if r.optionsService == nil || r.trainingService == nil {
		http.Error(w, "Training service unavailable", http.StatusInternalServerError)
		return
	}

	courseCode := r.requestedCourseCodeForUser(req, userID)
	config := r.trainingSessionConfigForUser(userID)
	config.CourseCode = courseCode
	userLC := r.config.Learning
	if courseCode != "" {
		userLC = learningConfigForCourse(r.config.Learning, courseCode)
	}
	config.IsEnglishTarget = !strings.EqualFold(userLC.TargetLang, "es")
	queue, err := r.trainingService.GenerateQueue(userID, config)
	if err != nil {
		r.logger.Error("failed to generate offline word training queue", zap.Int64("user_id", userID), zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(queue) == 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"app_code":        r.config.Learning.AppCode,
			"native_lang":     r.config.Learning.NativeLang,
			"target_lang":     r.config.Learning.TargetLang,
			"generated_at":    time.Now().UTC().Format(time.RFC3339),
			"algo_version":    "word_training_offline_v2_queue",
			"total_cards":     0,
			"available_count": 0,
			"queue":           []offlineWordTrainingQueueItem{},
			"cards":           []offlineWordTrainingCard{},
		})
		return
	}

	lang := i18n.GetLanguageFromContext(req.Context())
	items := make([]offlineWordTrainingQueueItem, 0, len(queue))
	legacyCards := make([]offlineWordTrainingCard, 0, len(queue))
	cardQueue := make([]*models.UserCardWithTraining, 0, len(queue))
	for _, item := range queue {
		if item.Type == "card" && item.Card != nil {
			cardQueue = append(cardQueue, item.Card)
		}
	}
	for _, item := range queue {
		switch item.Type {
		case "spell":
			if item.Spell == nil {
				continue
			}
			tl := learning.TargetLangNameRUPrepositional(r.config.Learning.TargetLang)
			items = append(items, offlineWordTrainingQueueItem{
				Type:          "spell",
				Question:      fmt.Sprintf("Составьте слово на %s: <strong>%s</strong>", tl, item.Spell.WordRU),
				UserCardID:    item.Spell.ReplacedUserCardID,
				WordCardID:    item.Spell.WordCardID,
				Direction:     "spell",
				WordRU:        item.Spell.WordRU,
				WordNative:    item.Spell.WordNative,
				WordTarget:    item.Spell.WordTarget,
				Prefix:        item.Spell.Prefix,
				Letters:       item.Spell.ShuffledLetters,
				CorrectAnswer: item.Spell.DisplayWord,
			})
		case "type":
			if item.TypeChallenge == nil {
				continue
			}
			displayWord := item.TypeChallenge.DisplayWord
			prefix := ""
			wordForHint := displayWord
			isEnglishTarget := config.IsEnglishTarget
			if strings.HasPrefix(displayWord, "to ") && len(displayWord) > 3 {
				if isEnglishTarget {
					prefix = "to "
					wordForHint = displayWord[3:]
				} else {
					displayWord = displayWord[3:]
					wordForHint = displayWord
				}
			}
			runes := []rune(wordForHint)
			hintFirstLetter := ""
			hintLength := 0
			if len(runes) > 0 {
				hintFirstLetter = string(runes[0])
				hintLength = len(runes)
			}
			tl := learning.TargetLangNameRUPrepositional(r.config.Learning.TargetLang)
			items = append(items, offlineWordTrainingQueueItem{
				Type:            "type",
				Question:        fmt.Sprintf("Введите слово на %s: <strong>%s</strong>", tl, item.TypeChallenge.WordRU),
				UserCardID:      item.TypeChallenge.ReplacedUserCardID,
				WordCardID:      item.TypeChallenge.WordCardID,
				Direction:       "type",
				WordRU:          item.TypeChallenge.WordRU,
				WordNative:      item.TypeChallenge.WordNative,
				WordTarget:      item.TypeChallenge.WordTarget,
				Prefix:          prefix,
				HintFirstLetter: hintFirstLetter,
				HintLength:      hintLength,
				CorrectAnswer:   displayWord,
			})
		default:
			if item.Card == nil {
				continue
			}
			cardIndex := indexOfCardInQueue(cardQueue, item.Card.UserCard.ID)
			options, correctAnswer, err := r.optionsServiceForCourse(req.Context(), userID, courseCode).GenerateOptions(item.Card, models.DefaultOptionCount, r.extractSessionWords(cardQueue, cardIndex, item.Card, nil), collectWordENs(cardQueue, cardIndex), collectWordRUs(cardQueue, cardIndex))
			if err != nil {
				r.logger.Warn("failed to generate offline word options", zap.Int64("user_card_id", item.Card.UserCard.ID), zap.Error(err))
				continue
			}
			legacy := r.buildOfflineWordTrainingCard(req, lang, item.Card, options, correctAnswer)
			legacyCards = append(legacyCards, legacy)
			queueItem := offlineWordTrainingQueueItem{
				Type:           "card",
				Question:       legacy.Question,
				UserCardID:     legacy.UserCardID,
				TrainingCardID: legacy.TrainingCardID,
				WordCardID:     legacy.WordCardID,
				Direction:      legacy.Direction,
				WordEN:         legacy.WordEN,
				WordTarget:     legacy.WordTarget,
				WordRU:         legacy.WordRU,
				WordNative:     legacy.WordNative,
				DisplayWord:    legacy.DisplayWord,
				DisplayTarget:  legacy.DisplayTarget,
				Transcription:  legacy.Transcription,
				ExampleEN:      legacy.ExampleEN,
				ExampleTarget:  legacy.ExampleTarget,
				Hint:           legacy.Hint,
				WordCategory:   legacy.WordCategory,
				Morph:          legacy.Morph,
				Options:        legacy.Options,
				CorrectAnswer:  legacy.CorrectAnswer,
				SRS:            legacy.SRS,
			}
			items = append(items, queueItem)
		}
	}

	response := map[string]interface{}{
		"app_code":        r.config.Learning.AppCode,
		"native_lang":     r.config.Learning.NativeLang,
		"target_lang":     r.config.Learning.TargetLang,
		"generated_at":    time.Now().UTC().Format(time.RFC3339),
		"algo_version":    "word_training_offline_v2_queue",
		"total_cards":     len(items),
		"available_count": len(items),
		"queue":           items,
		"cards":           legacyCards,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func indexOfCardInQueue(queue []*models.UserCardWithTraining, userCardID int64) int {
	for i, card := range queue {
		if card != nil && card.UserCard.ID == userCardID {
			return i
		}
	}
	return -1
}

func (r *Router) trainingSessionConfigForUser(userID int64) service.SessionConfig {
	config := service.SessionConfig{
		MaxCardsPerSession:      models.DefaultMaxCardsPerSession,
		MaxNewPerSession:        models.DefaultMaxNewPerSession,
		AlgoVersion:             "word_training_offline_v2_queue",
		SpellEnabled:            true,
		SpellMasteringThreshold: 50,
		TypeEnabled:             true,
		TypeMasteringThreshold:  70,
	}
	if r.userRepo != nil {
		if userRepo, ok := r.userRepo.(*repository.UserRepository); ok {
			user, _ := userRepo.GetUserByID(userID)
			if user != nil && user.SettingsJSON != "" {
				var settings models.UserSettings
				if json.Unmarshal([]byte(user.SettingsJSON), &settings) == nil {
					if settings.SpellModeEnabled != nil {
						config.SpellEnabled = *settings.SpellModeEnabled
					}
					if settings.SpellMasteringThreshold != nil {
						t := *settings.SpellMasteringThreshold
						if t < 0 {
							t = 0
						}
						if t > 100 {
							t = 100
						}
						config.SpellMasteringThreshold = t
					}
					if settings.TypeModeEnabled != nil {
						config.TypeEnabled = *settings.TypeModeEnabled
					}
					if settings.TypeMasteringThreshold != nil {
						t := *settings.TypeMasteringThreshold
						if t < 0 {
							t = 0
						}
						if t > 100 {
							t = 100
						}
						config.TypeMasteringThreshold = t
					}
				}
			}
		}
	}
	return config
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

		mode := strings.TrimSpace(attempt.Mode)
		if mode == "" {
			mode = "card"
		}
		if mode == "spell" || mode == "type" {
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
			if err := r.syncOfflineSpellTypeAttempt(req, userID, sessionRepo, sessionID, attempt, mode, result); err != nil {
				result["synced"] = false
				result["error"] = err.Error()
				results = append(results, result)
				continue
			}
			syncedCount++
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
			syncCourseCode := r.requestedCourseCodeForUser(req, userID)
			_, generatedCorrect, _ := r.optionsServiceForCourse(req.Context(), userID, syncCourseCode).GenerateOptions(&models.UserCardWithTraining{UserCard: *userCard, TrainingCard: *trainingCard}, models.DefaultOptionCount, nil, nil, nil)
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

		attemptData := models.AttemptData{Correct: isCorrect, EarlyReveal: attempt.EarlyReveal, AnswerTimeMS: answerTimeMS, TDelayMS: tDelayMS, OptionCount: len(attempt.Options), ChosenOption: chosenOption, GradedAt: &answeredAt}
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
		metricsJSON, _ := json.Marshal(map[string]interface{}{"offline_sync": true, "answer_time_ms": answerTimeMS, "total_time_ms": int(answeredAt.Sub(shownAt).Milliseconds()), "mode": mode})
		quality := models.CalculateQuality(attemptData)
		reviewEvent := &models.ReviewEvent{
			SessionID:       &sessionID,
			UserID:          userID,
			UserCardID:      userCard.ID,
			CourseCode:      r.courseCodeForUserCard(userCard.ID),
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
			r.recordLinglowWordReviewEvent(req.Context(), reviewEvent.CourseCode, reviewEventID, reviewEvent)
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

func (r *Router) syncOfflineSpellTypeAttempt(req *http.Request, userID int64, sessionRepo *repository.SessionRepository, sessionID int64, attempt offlineWordTrainingAttempt, mode string, result map[string]interface{}) error {
	userCardRepo := repository.NewUserCardRepository(r.db, r.logger)
	userCard, err := userCardRepo.GetUserCard(attempt.UserCardID)
	if err != nil || userCard == nil || userCard.UserID != userID {
		return fmt.Errorf("user_card_not_found")
	}
	correctAnswer := strings.TrimSpace(attempt.CorrectAnswer)
	if correctAnswer == "" {
		return fmt.Errorf("correct_answer_required")
	}
	userAnswer := strings.TrimSpace(strings.ToLower(attempt.AnswerText))
	if userAnswer == "" && attempt.ChosenOption != "" {
		userAnswer = strings.TrimSpace(strings.ToLower(attempt.ChosenOption))
	}
	if userAnswer == "" {
		userAnswer = " "
	}
	isCorrect := userAnswer == strings.TrimSpace(strings.ToLower(correctAnswer))
	shownAt := fallbackTime(attempt.ShownAt)
	answeredAt := fallbackTime(attempt.AnsweredAt)
	answerTimeMS := attempt.AnswerTimeMS
	if answerTimeMS <= 0 {
		answerTimeMS = int(answeredAt.Sub(shownAt).Milliseconds())
		if answerTimeMS < 0 {
			answerTimeMS = 0
		}
	}
	wordLen := len(correctAnswer)
	attemptData := models.AttemptData{
		Correct:        isCorrect,
		EarlyReveal:    false,
		AnswerTimeMS:   answerTimeMS,
		TDelayMS:       0,
		OptionCount:    1,
		ChosenOption:   attempt.AnswerText,
		TimeMultiplier: models.TimeMultiplierForMode(mode, wordLen),
		GradedAt:       &answeredAt,
	}
	srsBefore := models.SRSState{
		State: userCard.State, EF: userCard.EF, Reps: userCard.Reps,
		IntervalDays: userCard.IntervalDays, LearningStep: userCard.LearningStep, LapseCount: userCard.LapseCount,
	}
	srsBeforeJSON, _ := json.Marshal(srsBefore)
	if err := r.srsService.GradeCard(userCard, attemptData); err != nil {
		return fmt.Errorf("grade_failed")
	}
	srsAfter := models.SRSState{
		State: userCard.State, EF: userCard.EF, Reps: userCard.Reps,
		IntervalDays: userCard.IntervalDays, LearningStep: userCard.LearningStep, LapseCount: userCard.LapseCount,
	}
	srsAfterJSON, _ := json.Marshal(srsAfter)
	quality := models.CalculateQuality(attemptData)
	metricsJSON, _ := json.Marshal(map[string]interface{}{
		"offline_sync": true, "spell_or_type": true, "answer_time_ms": answerTimeMS, "mode": mode, "word_len": wordLen,
	})
	reviewEvent := &models.ReviewEvent{
		SessionID:       &sessionID,
		UserID:          userID,
		UserCardID:      userCard.ID,
		CourseCode:      r.courseCodeForUserCard(userCard.ID),
		ClientAttemptID: attempt.ClientAttemptID,
		Direction:       userCard.Direction,
		ShownAt:         shownAt,
		AnsweredAt:      &answeredAt,
		TDelayMS:        0,
		EarlyReveal:     false,
		OptionCount:     1,
		OptionsJSON:     "[]",
		ChosenOption:    attempt.AnswerText,
		IsCorrect:       isCorrect,
		Quality:         int(quality),
		MetricsJSON:     string(metricsJSON),
		SRSBeforeJSON:   string(srsBeforeJSON),
		SRSAfterJSON:    string(srsAfterJSON),
	}
	if reviewEventID, err := sessionRepo.CreateReviewEvent(reviewEvent); err != nil {
		return fmt.Errorf("review_event_create_failed")
	} else {
		r.recordLinglowWordReviewEvent(req.Context(), reviewEvent.CourseCode, reviewEventID, reviewEvent)
	}
	if !isCorrect {
		if err := r.srsService.RecordWrongAnswer(userCard, attempt.AnswerText); err != nil {
			r.logger.Warn("failed to record offline spell/type wrong answer", zap.Int64("user_card_id", userCard.ID), zap.Error(err))
		}
	}
	result["synced"] = true
	result["is_correct"] = isCorrect
	return nil
}
