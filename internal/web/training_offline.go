package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

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
			"app_code":        userLC.AppCode,
			"native_lang":     userLC.NativeLang,
			"target_lang":     userLC.TargetLang,
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
			tl := learning.TargetLangNameRUPrepositional(userLC.TargetLang)
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
			tl := learning.TargetLangNameRUPrepositional(userLC.TargetLang)
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
		"app_code":        userLC.AppCode,
		"native_lang":     userLC.NativeLang,
		"target_lang":     userLC.TargetLang,
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
	lc := r.config.Learning
	if code := r.requestedCourseCodeForUser(req, getUserIDFromContext(req.Context())); code != "" {
		lc = learningConfigForCourse(lc, code)
	}

	displayWord := card.TrainingCard.WordEN
	if card.TrainingCard.DisplayWord != nil && *card.TrainingCard.DisplayWord != "" {
		displayWord = *card.TrainingCard.DisplayWord
	}
	var tl string
	switch lang {
	case "ru":
		tl = learning.TargetLangNameRUAccusative(lc.TargetLang)
	case "es":
		tl = learning.TargetLangNameES(lc.TargetLang)
	default:
		tl = learning.TargetLangNameEN(lc.TargetLang)
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
		item.Morph = buildCompactMorphFromWordCard(lc.TargetLang, wordCard, card.TrainingCard.POS)
	}
	return item
}

// Sync each answer atomically. A failed answer stays retryable; a committed
// client_attempt_id can never advance SRS a second time, even across replicas.
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
	unlock := r.trainingLocks.lock(userID)
	defer unlock()
	var sessionID int64
	synced := 0
	results := make([]map[string]interface{}, 0, len(payload.Attempts))
	for _, attempt := range payload.Attempts {
		result := map[string]interface{}{"client_attempt_id": attempt.ClientAttemptID}
		event, err := r.syncOfflineWordAttempt(req, userID, &sessionID, len(payload.Attempts), attempt)
		if err != nil {
			result["synced"] = false
			result["error"] = err.Error()
		} else {
			result["synced"] = true
			if event == nil {
				result["duplicate"] = true
			} else {
				result["is_correct"] = event.IsCorrect
				synced++
				r.recordLinglowWordReviewEvent(req.Context(), event.CourseCode, event.ID, event)
			}
		}
		results = append(results, result)
	}
	if sessionID != 0 {
		if err := r.trainingService.FinishSession(sessionID, synced); err != nil {
			r.logger.Warn("finish offline session", zap.Error(err))
		}
	}
	writeJSON(w, map[string]interface{}{"results": results, "synced": synced})
}

func (r *Router) syncOfflineWordAttempt(req *http.Request, userID int64, sessionID *int64, planned int, attempt offlineWordTrainingAttempt) (*models.ReviewEvent, error) {
	if strings.TrimSpace(attempt.ClientAttemptID) == "" || attempt.UserCardID == 0 {
		return nil, fmt.Errorf("invalid_attempt")
	}
	mode := strings.TrimSpace(attempt.Mode)
	if mode == "" {
		mode = "card"
	}
	if mode != "card" && mode != "spell" && mode != "type" {
		return nil, fmt.Errorf("invalid_mode")
	}
	tx, err := r.db.BeginTx(req.Context(), nil)
	if err != nil {
		return nil, fmt.Errorf("transaction_failed")
	}
	defer tx.Rollback()
	var locked int64
	if err = tx.QueryRow(`SELECT id FROM users WHERE id = ? FOR UPDATE`, userID).Scan(&locked); err != nil {
		return nil, fmt.Errorf("user_not_found")
	}
	sessions := repository.NewSessionRepository(tx, r.logger)
	exists, err := sessions.HasReviewEventClientAttempt(userID, attempt.ClientAttemptID)
	if err != nil {
		return nil, fmt.Errorf("idempotency_check_failed")
	}
	if exists {
		return nil, nil
	}
	// The row lock also protects against other writers while the grade is computed.
	if err = tx.QueryRow(`SELECT id FROM user_cards WHERE id = ? AND user_id = ? FOR UPDATE`, attempt.UserCardID, userID).Scan(&locked); err != nil {
		return nil, fmt.Errorf("user_card_not_found")
	}
	cards := repository.NewUserCardRepository(tx, r.logger)
	card, err := cards.GetUserCard(attempt.UserCardID)
	if err != nil {
		return nil, fmt.Errorf("user_card_not_found")
	}
	training, err := repository.NewTrainingCardRepository(tx, r.logger).GetTrainingCard(card.TrainingCardID)
	if err != nil || training == nil {
		return nil, fmt.Errorf("training_card_not_found")
	}
	if attempt.TrainingCardID != 0 && attempt.TrainingCardID != training.ID {
		return nil, fmt.Errorf("training_card_mismatch")
	}
	var course string
	err = tx.QueryRow(`SELECT COALESCE(NULLIF(uc.course_code, ''), NULLIF(tc.course_code, ''), NULLIF(wc.course_code, ''), '') FROM user_cards uc JOIN training_cards tc ON tc.id = uc.training_card_id JOIN word_cards wc ON wc.id = tc.word_card_id WHERE uc.id = ?`, card.ID).Scan(&course)
	if err != nil {
		return nil, fmt.Errorf("course_lookup_failed")
	}
	lc := r.config.Learning
	if course != "" {
		lc = learningConfigForCourse(lc, course)
	}
	options := service.NewOptionsService(nil, r.logger, lc.TargetLang)
	correct := options.CorrectAnswer(&models.UserCardWithTraining{UserCard: *card, TrainingCard: *training})
	chosen := attempt.ChosenOption
	if mode != "card" {
		if card.Direction != models.DirectionRUtoEN {
			return nil, fmt.Errorf("invalid_direction")
		}
		correct = service.TrainingDisplayWord(training)
		if lc.TargetLang != "en" {
			correct = strings.TrimPrefix(correct, "to ")
		}
		chosen = attempt.AnswerText
		if chosen == "" {
			chosen = attempt.ChosenOption
		}
	}
	if strings.TrimSpace(correct) == "" {
		return nil, fmt.Errorf("correct_answer_unavailable")
	}
	isCorrect := chosen == correct
	if mode != "card" {
		isCorrect = strings.EqualFold(strings.TrimSpace(chosen), strings.TrimSpace(correct))
	}
	shown, answered := fallbackTime(attempt.ShownAt), fallbackTime(attempt.AnsweredAt)
	optionsShown := attempt.OptionsShownAt
	if optionsShown.IsZero() {
		optionsShown = shown
	}
	answerMS, delayMS := attempt.AnswerTimeMS, attempt.TDelayMS
	if answerMS <= 0 {
		answerMS = max(0, int(answered.Sub(optionsShown).Milliseconds()))
	}
	if delayMS <= 0 {
		delayMS = max(0, int(optionsShown.Sub(shown).Milliseconds()))
	}
	data := models.AttemptData{
		Correct: isCorrect, EarlyReveal: attempt.EarlyReveal,
		AnswerTimeMS: answerMS, TDelayMS: delayMS,
		OptionCount: len(attempt.Options), ChosenOption: chosen, GradedAt: &answered,
	}
	if mode != "card" {
		data.EarlyReveal = false
		data.TDelayMS = 0
		data.OptionCount = 1
		data.TimeMultiplier = models.TimeMultiplierForMode(mode, utf8.RuneCountInString(correct))
	}
	before, _ := json.Marshal(wordReviewSRSState(card))
	srs := service.NewSRSService(cards, lc, r.logger)
	if err = srs.GradeCard(card, data); err != nil {
		return nil, fmt.Errorf("grade_failed")
	}
	after, _ := json.Marshal(wordReviewSRSState(card))
	if !isCorrect {
		if err = srs.RecordWrongAnswer(card, chosen); err != nil {
			return nil, fmt.Errorf("wrong_answer_save_failed")
		}
	}
	sid := *sessionID
	if sid == 0 {
		sid, err = sessions.CreateSession(&models.TrainingSession{UserID: userID, Source: models.SourceManual, PlannedCount: planned, SessionJSON: `{"offline_sync":true}`})
		if err != nil {
			return nil, fmt.Errorf("session_create_failed")
		}
	}
	optionsJSON, _ := json.Marshal(attempt.Options)
	metrics, _ := json.Marshal(map[string]interface{}{"offline_sync": true, "mode": mode, "answer_time_ms": answerMS, "total_time_ms": max(0, int(answered.Sub(shown).Milliseconds())), "spell_or_type": mode != "card"})
	event := &models.ReviewEvent{
		SessionID: &sid, UserID: userID, UserCardID: card.ID,
		CourseCode: course, ClientAttemptID: attempt.ClientAttemptID,
		Direction: card.Direction, ShownAt: shown, OptionsShownAt: &optionsShown, AnsweredAt: &answered,
		TDelayMS: data.TDelayMS, EarlyReveal: data.EarlyReveal, OptionCount: data.OptionCount,
		OptionsJSON: string(optionsJSON), ChosenOption: chosen, IsCorrect: isCorrect,
		Quality: int(models.CalculateQuality(data)), MetricsJSON: string(metrics),
		SRSBeforeJSON: string(before), SRSAfterJSON: string(after),
	}
	id, err := sessions.CreateReviewEvent(event)
	if err != nil {
		return nil, fmt.Errorf("review_event_create_failed")
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit_failed")
	}
	event.ID = id
	*sessionID = sid
	return event, nil
}

func wordReviewSRSState(card *models.UserCard) models.SRSState {
	return models.SRSState{State: card.State, EF: card.EF, Reps: card.Reps, IntervalDays: card.IntervalDays, LearningStep: card.LearningStep, LapseCount: card.LapseCount}
}

func fallbackTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now()
	}
	return value
}
