package web

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/spanishverbs"

	"go.uber.org/zap"
)

type webVerbTrainingState struct {
	UserID    int64
	SessionID int64
	Queue     []repository.VerbQueueCard
	Index     int
}

var (
	webVerbSessionsMu sync.RWMutex
	webVerbSessions   = map[int64]*webVerbTrainingState{}
)

func (r *Router) verbFormsEnabled() bool {
	return strings.EqualFold(r.config.Learning.TargetLang, "es") && r.config.Training.SpanishVerbFormsEnabled
}

// writeVerbTrainingDisabled responds when verb-form training is off (wrong target language or SPANISH_VERB_FORMS_ENABLED=false).
// Uses 403 + JSON so clients do not confuse this with a missing route (previously 404).
func (r *Router) writeVerbTrainingDisabled(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"code": "verb_training_disabled",
	})
}

func (r *Router) handleVocabVerbForms(w http.ResponseWriter, req *http.Request, userID int64, wordCardID int64) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.verbFormsEnabled() {
		http.NotFound(w, req)
		return
	}
	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	rows, err := repo.GetUserVerbForms(userID, wordCardID)
	if err != nil {
		r.logger.Error("failed to get user verb forms", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	grouped := map[string]map[string][]repository.VerbFormViewRow{}
	for _, row := range rows {
		if grouped[row.Mood] == nil {
			grouped[row.Mood] = map[string][]repository.VerbFormViewRow{}
		}
		grouped[row.Mood][row.Tense] = append(grouped[row.Mood][row.Tense], row)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"word_card_id": wordCardID,
		"forms":        rows,
		"grouped":      grouped,
	})
}

func (r *Router) getUserVerbScopes(userID int64) []string {
	if r.userRepo == nil {
		return models.DefaultSpanishVerbScopes()
	}
	userRepo, ok := r.userRepo.(*repository.UserRepository)
	if !ok {
		return models.DefaultSpanishVerbScopes()
	}
	user, err := userRepo.GetUserByID(userID)
	if err != nil || user == nil || user.SettingsJSON == "" {
		return models.DefaultSpanishVerbScopes()
	}
	var settings models.UserSettings
	if err := json.Unmarshal([]byte(user.SettingsJSON), &settings); err != nil {
		return models.DefaultSpanishVerbScopes()
	}
	return service.ResolveVerbScopes(&settings, r.config.Learning)
}

func (r *Router) newVerbTrainingService() *service.VerbTrainingService {
	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	return service.NewVerbTrainingService(repo, r.config.Learning, r.config.Training, r.logger)
}

// ensureVerbFormUserCardsAfterVocab mirrors word training: once user_cards exist for vocabulary,
// materialize verb form cards (verb_training_cards + user_verb_cards) for Spanish.
func (r *Router) ensureVerbFormUserCardsAfterVocab(userID int64) {
	if !r.verbFormsEnabled() {
		return
	}
	vs := r.newVerbTrainingService()
	scopes := r.getUserVerbScopes(userID)
	if err := vs.EnsureVerbFormUserCards(userID, scopes); err != nil {
		r.logger.Warn("ensure verb form user cards after vocabulary change",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
	}
}

func (r *Router) handleVerbTrainingStart(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vs := r.newVerbTrainingService()
	if !vs.Enabled() {
		r.writeVerbTrainingDisabled(w)
		return
	}
	session, err := vs.StartSession(userID, r.getUserVerbScopes(userID))
	if err != nil {
		r.logger.Error("failed to start verb training", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	state := &webVerbTrainingState{
		UserID:    userID,
		SessionID: session.SessionID,
		Queue:     session.Queue,
		Index:     0,
	}
	webVerbSessionsMu.Lock()
	webVerbSessions[userID] = state
	webVerbSessionsMu.Unlock()
	r.writeCurrentVerbCard(w, state)
}

func (r *Router) handleVerbTrainingCurrent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	webVerbSessionsMu.RLock()
	state := webVerbSessions[userID]
	webVerbSessionsMu.RUnlock()
	if state == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "no active verb training session"})
		return
	}
	r.writeCurrentVerbCard(w, state)
}

func (r *Router) writeCurrentVerbCard(w http.ResponseWriter, state *webVerbTrainingState) {
	if state.Index >= len(state.Queue) {
		repo := repository.NewVerbFormsRepository(r.db, r.logger)
		_ = repo.FinishVerbSession(state.SessionID, len(state.Queue))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"finished": true,
			"total":    len(state.Queue),
		})
		return
	}
	item := state.Queue[state.Index]
	var prompt map[string]interface{}
	_ = json.Unmarshal([]byte(item.PromptJSON), &prompt)

	var answer map[string]string
	_ = json.Unmarshal([]byte(item.AnswerJSON), &answer)
	surfaceAns := strings.TrimSpace(answer["surface_form"])
	if item.CardType == models.VerbCardTypeCloze && prompt != nil && surfaceAns != "" {
		mode, _ := prompt["example_mode"].(string)
		if mode == spanishverbs.ExampleModeRuntime {
			lemma, _ := prompt["lemma"].(string)
			mood, _ := prompt["mood"].(string)
			tense, _ := prompt["tense"].(string)
			person, _ := prompt["person"].(string)
			number, _ := prompt["number"].(string)
			ruGloss, _ := prompt["ru_gloss"].(string)
			prompt["question"] = spanishverbs.BuildVerbTrainingClozeQuestion(person, number, lemma, mood, tense)
			prompt["example_translation"] = spanishverbs.PlainRussianVerbTrainingHintLine(lemma, person, number, ruGloss, mood, tense)
		} else if q, ok := prompt["question"].(string); ok && strings.TrimSpace(q) != "" {
			prompt["question"] = service.MaskClozeVerbSurfaceInQuestion(q, surfaceAns)
		}
	}

	options := service.ParseStringJSONArray(item.DistractorsJSON)
	if options == nil {
		options = []string{}
	}

	vrepo := repository.NewVerbFormsRepository(r.db, r.logger)
	srs, err := vrepo.GetVerbUserCardSRS(item.UserVerbCardID)
	if err != nil {
		r.logger.Warn("verb training srs lookup failed", zap.Error(err))
		srs = nil
	}
	minR := r.config.Training.VerbFormsTypedMinReps
	if minR < 1 {
		minR = 2
	}
	chance := r.config.Training.VerbFormsTypedChancePercent
	if chance < 0 {
		chance = 0
	}
	if chance > 100 {
		chance = 100
	}
	eligibleTyped := srs != nil && srs.Reps >= minR && srs.State != "learning"
	inputMode := "choice"
	if eligibleTyped && chance > 0 {
		seed := state.SessionID ^ item.UserVerbCardID ^ int64(state.Index)
		if rand.New(rand.NewSource(seed)).Intn(100) < chance {
			inputMode = "typed"
			options = []string{}
		}
	}
	if inputMode == "choice" && len(options) < 2 {
		lemma := ""
		if prompt != nil {
			if v, ok := prompt["lemma"].(string); ok {
				lemma = strings.TrimSpace(v)
			}
		}
		options = service.BuildVerbFormMultipleChoiceOptions(surfaceAns, lemma, item.UserVerbCardID)
	}
	if inputMode == "choice" && len(options) > service.VerbChoiceOptionCount {
		options = service.CapVerbMultipleChoiceOptions(surfaceAns, options, item.UserVerbCardID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":           state.SessionID,
		"user_verb_card_id":    item.UserVerbCardID,
		"card_type":            item.CardType,
		"prompt":               prompt,
		"options":              options,
		"input_mode":           inputMode,
		"typed_min_reps":       minR,
		"card_index":           state.Index + 1,
		"total_cards":          len(state.Queue),
		"direction":            "verb_forms",
		"supports_immediate":   true,
	})
}

func (r *Router) handleVerbTrainingAnswer(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	webVerbSessionsMu.RLock()
	state := webVerbSessions[userID]
	webVerbSessionsMu.RUnlock()
	if state == nil || state.Index >= len(state.Queue) {
		http.Error(w, "No active session", http.StatusBadRequest)
		return
	}

	var payload struct {
		UserVerbCardID int64  `json:"user_verb_card_id"`
		Answer         string `json:"answer"`
		Skip           bool   `json:"skip"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	item := state.Queue[state.Index]
	if payload.UserVerbCardID != 0 && payload.UserVerbCardID != item.UserVerbCardID {
		http.Error(w, "Card mismatch", http.StatusBadRequest)
		return
	}
	var answer map[string]string
	_ = json.Unmarshal([]byte(item.AnswerJSON), &answer)
	surface := strings.TrimSpace(answer["surface_form"])
	expectedLower := strings.ToLower(surface)
	got := strings.ToLower(strings.TrimSpace(payload.Answer))
	isCorrect := !payload.Skip && expectedLower != "" && expectedLower == got
	chosen := strings.TrimSpace(payload.Answer)
	if payload.Skip {
		chosen = ""
	}

	vs := r.newVerbTrainingService()
	if err := vs.Grade(userID, state.SessionID, item.UserVerbCardID, isCorrect); err != nil {
		r.logger.Error("failed to grade verb card", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	state.Index++
	webVerbSessionsMu.Lock()
	webVerbSessions[userID] = state
	webVerbSessionsMu.Unlock()

	feedback := map[string]interface{}{
		"is_correct":     isCorrect,
		"correct":        isCorrect,
		"chosen_option":  chosen,
		"correct_answer": surface,
		"expected":       expectedLower,
		"next":           state.Index < len(state.Queue),
	}
	if !isCorrect {
		_, wrongAnswerDelaySeconds := r.getTrainingDelaysForUser(userID)
		feedback["delay_seconds"] = wrongAnswerDelaySeconds
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(feedback)
}

func (r *Router) handleVerbTrainingUpcoming(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !r.verbFormsEnabled() {
		r.writeVerbTrainingDisabled(w)
		return
	}
	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	queue, err := repo.GetVerbQueue(userID, time.Now(), r.config.Training.VerbFormsMaxCards, r.config.Training.VerbFormsMaxNew)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"due":             len(queue),
		"max_per_session": r.config.Training.VerbFormsMaxCards,
		"enabled":         true,
	})
}

func parseWordCardIDForVerbForms(path string) (int64, bool) {
	if !strings.HasPrefix(path, "/api/vocab/") {
		return 0, false
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/vocab/"), "/")
	if len(parts) != 2 || parts[1] != "verb-forms" {
		return 0, false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
