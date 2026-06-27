package web

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"tgbot-skeleton/internal/grammartrainingpack"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
	"tgbot-skeleton/internal/spanishverbs"
	"tgbot-skeleton/internal/verbtraining"

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
	return strings.EqualFold(r.config.Learning.TargetLang, "es")
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

func writeVerbFormsGroupedResponse(w http.ResponseWriter, wordCardID int64, rows []repository.VerbFormViewRow) {
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

func (r *Router) handleVerbTrainingLemmaForms(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if getUserIDFromContext(req.Context()) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !r.verbFormsEnabled() {
		http.NotFound(w, req)
		return
	}
	lemma := strings.TrimSpace(req.URL.Query().Get("lemma"))
	if lemma == "" {
		http.Error(w, "missing lemma", http.StatusBadRequest)
		return
	}
	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	rows, err := repo.ListVerbFormViewRowsForLemma(lemma, "es")
	if err != nil {
		r.logger.Error("verb forms by lemma", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if len(rows) == 0 {
		http.Error(w, "no forms for lemma", http.StatusNotFound)
		return
	}
	writeVerbFormsGroupedResponse(w, 0, rows)
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
	writeVerbFormsGroupedResponse(w, wordCardID, rows)
}

func (r *Router) getUserVerbScopes(userID int64) []string {
	if strings.ToLower(strings.TrimSpace(r.config.Learning.TargetLang)) != "es" {
		return nil
	}
	if fsys, err := grammartrainingpack.PackFS(r.config.Learning.GrammarBundleID); err == nil {
		if gates, err := verbtraining.LoadUnlockGates(fsys); err == nil && gates != nil {
			allowed := map[string]bool{}
			if r.grammarService != nil && userID > 0 {
				for chapterID := range gates.Chapters {
					canAccess, err := r.grammarService.CanAccessChapter(context.Background(), userID, chapterID)
					if err != nil {
						continue
					}
					if canAccess {
						allowed[chapterID] = true
					}
				}
			}
			scopes := gates.EnabledScopes(allowed)
			if len(scopes) > 0 {
				return scopes
			}
		}
	}
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

func isVerbClozeCardType(cardType string) bool {
	return strings.EqualFold(strings.TrimSpace(cardType), models.VerbCardTypeCloze)
}

func promptString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strings.TrimSpace(strconv.FormatInt(int64(t), 10))
	case json.Number:
		return strings.TrimSpace(t.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", t))
	}
}

func stringSliceFromPrompt(m map[string]interface{}, key string) []string {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case []string:
		out := make([]string, 0, len(t))
		for _, s := range t {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, x := range t {
			s := strings.TrimSpace(fmt.Sprint(x))
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// hydrateVerbClozePrompt fills Spanish cloze line and Russian support text when the stored prompt is minimal
// (e.g. legacy/runtime rows with empty question). Typed input mode uses the same payload as choice mode, so both must show
// a sentence-style cloze plus translation like imported lemma artifacts (question + example_translation).
func hydrateVerbClozePrompt(prompt map[string]interface{}, surfaceAns string, seed int64) {
	if prompt == nil {
		return
	}
	lemma := promptString(prompt, "lemma")
	mood := promptString(prompt, "mood")
	tense := promptString(prompt, "tense")
	person := promptString(prompt, "person")
	number := promptString(prompt, "number")
	ruGloss := promptString(prompt, "ru_gloss")
	verbClass := promptString(prompt, "verb_class")
	allowed := stringSliceFromPrompt(prompt, "allowed_template_ids")

	q := strings.TrimSpace(promptString(prompt, "question"))
	tr := strings.TrimSpace(promptString(prompt, "example_translation"))
	surfaceAns = strings.TrimSpace(surfaceAns)

	switch {
	case q == "" && surfaceAns != "":
		es, ru := spanishverbs.GenerateVerbExamplePair(seed, lemma, mood, tense, person, number, surfaceAns, ruGloss, verbClass, allowed)
		if strings.TrimSpace(es) != "" {
			prompt["question"] = es
		}
		if tr == "" && strings.TrimSpace(ru) != "" {
			prompt["example_translation"] = ru
		}
	case q != "" && tr == "" && surfaceAns != "":
		prompt["example_translation"] = spanishverbs.BuildRussianLiteraryLine(
			q, lemma, surfaceAns, mood, tense, ruGloss, seed)
	}
}

func ensureVerbClozeQuestionLine(prompt map[string]interface{}) {
	if prompt == nil {
		return
	}
	if strings.TrimSpace(promptString(prompt, "question")) != "" {
		return
	}
	lemma := promptString(prompt, "lemma")
	if lemma == "" {
		return
	}
	prompt["question"] = spanishverbs.BuildVerbTrainingClozeQuestion(
		promptString(prompt, "person"),
		promptString(prompt, "number"),
		lemma,
		promptString(prompt, "mood"),
		promptString(prompt, "tense"),
	)
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
		r.finishVerbTrainingSessionResponse(w, state)
		return
	}
	item := state.Queue[state.Index]
	var prompt map[string]interface{}
	_ = json.Unmarshal([]byte(item.PromptJSON), &prompt)
	if prompt == nil {
		prompt = map[string]interface{}{}
	}

	var answer map[string]string
	_ = json.Unmarshal([]byte(item.AnswerJSON), &answer)
	surfaceAns := strings.TrimSpace(answer["surface_form"])

	cardSeed := state.SessionID ^ item.UserVerbCardID ^ int64(state.Index)
	if isVerbClozeCardType(item.CardType) {
		hydrateVerbClozePrompt(prompt, surfaceAns, cardSeed)
	}

	if isVerbClozeCardType(item.CardType) && surfaceAns != "" {
		if q := strings.TrimSpace(promptString(prompt, "question")); q != "" {
			prompt["question"] = service.MaskClozeVerbSurfaceInQuestion(q, surfaceAns)
		}
	}
	if isVerbClozeCardType(item.CardType) {
		ensureVerbClozeQuestionLine(prompt)
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
		inputMode = "typed"
		options = []string{}
	}
	if inputMode == "choice" && len(options) >= 2 {
		seed := state.SessionID ^ item.UserVerbCardID ^ int64(state.Index)
		rng := rand.New(rand.NewSource(seed))
		rng.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	}

	lemma := promptString(prompt, "lemma")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":         state.SessionID,
		"user_verb_card_id":  item.UserVerbCardID,
		"word_card_id":       item.WordCardID,
		"lemma":              lemma,
		"card_type":          item.CardType,
		"prompt":             prompt,
		"options":            options,
		"input_mode":         inputMode,
		"typed_min_reps":     minR,
		"card_index":         state.Index + 1,
		"total_cards":        len(state.Queue),
		"direction":          "verb_forms",
		"supports_immediate": true,
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
	sessionComplete := state.Index >= len(state.Queue)
	webVerbSessionsMu.Lock()
	if sessionComplete {
		delete(webVerbSessions, userID)
	} else {
		webVerbSessions[userID] = state
	}
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
	if sessionComplete {
		repo := repository.NewVerbFormsRepository(r.db, r.logger)
		if err := repo.FinishVerbSession(state.SessionID, state.Index); err != nil {
			r.logger.Error("failed to finish verb session", zap.Error(err))
		}
		totalCards, correctCards, err := repo.GetVerbSessionStats(state.SessionID)
		if err != nil {
			r.logger.Error("failed to get verb session stats", zap.Error(err))
			totalCards = state.Index
			correctCards = 0
		}
		feedback["complete"] = true
		feedback["cards_completed"] = state.Index
		feedback["total_cards"] = totalCards
		feedback["correct_cards"] = correctCards
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(feedback)
}

func (r *Router) finishVerbTrainingSessionResponse(w http.ResponseWriter, state *webVerbTrainingState) {
	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	if err := repo.FinishVerbSession(state.SessionID, state.Index); err != nil {
		r.logger.Error("failed to finish verb session", zap.Error(err))
	}
	totalCards, correctCards, err := repo.GetVerbSessionStats(state.SessionID)
	if err != nil {
		r.logger.Error("failed to get verb session stats", zap.Error(err))
		totalCards = state.Index
		correctCards = 0
	}
	webVerbSessionsMu.Lock()
	delete(webVerbSessions, state.UserID)
	webVerbSessionsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"finished":        true,
		"complete":        true,
		"total":           totalCards,
		"total_cards":     totalCards,
		"correct_cards":   correctCards,
		"cards_completed": state.Index,
	})
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
	totalCards, err := repo.CountUserVerbClozeCards(userID)
	if err != nil {
		r.logger.Error("count user verb cloze cards", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"due":             len(queue),
		"total_cards":     totalCards,
		"max_per_session": r.config.Training.VerbFormsMaxCards,
		"enabled":         true,
		"pool_ready":      totalCards > 0,
	})
}

func (r *Router) handleInternalVerbTrainingPending(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !r.authorizeInternalService(req) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	limit := 200
	if raw := strings.TrimSpace(req.URL.Query().Get("limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		limit = v
	}
	var cursor int64
	if raw := strings.TrimSpace(req.URL.Query().Get("cursor")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			http.Error(w, "Invalid cursor", http.StatusBadRequest)
			return
		}
		cursor = v
	}
	// Default (forms_gap_only): lemmas with fewer than full V1 cloze cards in verb_training_cards (none/partial pack).
	// all=1 / forms_gap_only=0: every qualifying infinitive with a verb vocabulary card (ignore pack completeness).
	formsGapOnly := true
	if raw := strings.TrimSpace(req.URL.Query().Get("all")); raw == "1" || strings.EqualFold(raw, "true") {
		formsGapOnly = false
	}
	if raw := strings.TrimSpace(req.URL.Query().Get("forms_gap_only")); raw == "0" || strings.EqualFold(raw, "false") {
		formsGapOnly = false
	}
	repo := repository.NewVerbFormsRepository(r.db, r.logger)
	rows, err := repo.ListPendingVerbTrainingLemmas(limit, cursor, formsGapOnly)
	if err != nil {
		r.logger.Error("list pending verb lemmas", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	items := make([]map[string]interface{}, 0, len(rows))
	nextCursor := int64(0)
	for _, row := range rows {
		nextCursor = row.WordCardID
		items = append(items, map[string]interface{}{
			"lemma": row.Lemma,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"items":       items,
		"count":       len(items),
		"next_cursor": nextCursor,
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
