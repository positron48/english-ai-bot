package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func (r *Router) sentenceRepo() *repository.SentenceCompositionRepository {
	return repository.NewSentenceCompositionRepository(r.db, r.logger)
}

// activeSentenceSet returns the user's current (non-completed) set for their course, or nil.
func (r *Router) activeSentenceSet(req *http.Request, userID int64) (*models.SentenceSet, error) {
	courseCode := r.requestedCourseCodeForUser(req, userID)
	if courseCode == "" {
		courseCode = r.defaultCourseCode()
	}
	set, err := r.sentenceRepo().LatestSet(userID, courseCode)
	if err != nil {
		return nil, err
	}
	if set == nil || set.Status == models.SentenceSetCompleted {
		return nil, nil
	}
	return set, nil
}

func sentenceItemJSON(it models.SentenceItem, includePrompt bool) map[string]interface{} {
	m := map[string]interface{}{
		"id":        it.ID,
		"position":  it.Position,
		"attempted": it.AttemptedAt != nil,
	}
	if includePrompt {
		m["prompt_ru"] = it.PromptRU
	}
	if it.Outcome != nil {
		m["outcome"] = *it.Outcome
	}
	if it.ErrorCount != nil {
		m["error_count"] = *it.ErrorCount
	}
	if it.UserInput != nil {
		m["user_input"] = *it.UserInput
	}
	if it.GradingJSON != nil && *it.GradingJSON != "" {
		m["grading"] = json.RawMessage(*it.GradingJSON)
	}
	return m
}

func setProgress(items []models.SentenceItem) (total, attempted int) {
	total = len(items)
	for _, it := range items {
		if it.AttemptedAt != nil {
			attempted++
		}
	}
	return
}

// handleSentenceTrainingToday reports the availability of today's set (read-only; no consumption).
func (r *Router) handleSentenceTrainingToday(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	set, err := r.activeSentenceSet(req, userID)
	if err != nil {
		r.logger.Error("sentence today: load set", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if set == nil {
		writeJSON(w, map[string]interface{}{"available": false})
		return
	}
	items, err := r.sentenceRepo().GetItems(set.ID)
	if err != nil {
		r.logger.Error("sentence today: load items", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	total, attempted := setProgress(items)
	writeJSON(w, map[string]interface{}{
		"available":       true,
		"set_id":          set.ID,
		"status":          set.Status,
		"generation_date": set.GenerationDate,
		"total":           total,
		"attempted":       attempted,
		"remaining":       total - attempted,
		"stars":           set.StarCount,
		"passed":          set.PassedCount,
		"failed":          set.FailedCount,
	})
}

// handleSentenceTrainingStart marks the active set as started (consumption marker) and returns
// its items (with any prior results, so the session can resume).
func (r *Router) handleSentenceTrainingStart(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	set, err := r.activeSentenceSet(req, userID)
	if err != nil {
		r.logger.Error("sentence start: load set", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if set == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "no_set_available"})
		return
	}
	// Mark started only when still "ready" (sets started_at: unblocks future regeneration).
	if set.Status == models.SentenceSetReady {
		if err := r.sentenceRepo().MarkStarted(set.ID); err != nil {
			r.logger.Error("sentence start: mark started", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	items, err := r.sentenceRepo().GetItems(set.ID)
	if err != nil {
		r.logger.Error("sentence start: load items", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	out := make([]map[string]interface{}, 0, len(items))
	for _, it := range items {
		out = append(out, sentenceItemJSON(it, true))
	}
	writeJSON(w, map[string]interface{}{
		"set_id": set.ID,
		"status": "started",
		"items":  out,
	})
}

// handleSentenceTrainingCurrent returns the next un-attempted item, or done.
func (r *Router) handleSentenceTrainingCurrent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	set, err := r.activeSentenceSet(req, userID)
	if err != nil {
		r.logger.Error("sentence current: load set", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if set == nil {
		writeJSON(w, map[string]interface{}{"done": true})
		return
	}
	items, err := r.sentenceRepo().GetItems(set.ID)
	if err != nil {
		r.logger.Error("sentence current: load items", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	total, attempted := setProgress(items)
	for _, it := range items {
		if it.AttemptedAt == nil {
			m := sentenceItemJSON(it, true)
			m["total"] = total
			m["attempted"] = attempted
			m["remaining"] = total - attempted
			writeJSON(w, m)
			return
		}
	}
	writeJSON(w, map[string]interface{}{"done": true, "total": total, "attempted": attempted})
}

// handleSentenceTrainingAnswer grades one submission (single shot) and records the result.
func (r *Router) handleSentenceTrainingAnswer(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		ItemID    int64  `json:"item_id"`
		UserInput string `json:"user_input"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.ItemID == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	repo := r.sentenceRepo()
	item, set, err := repo.GetItemWithSet(body.ItemID)
	if err != nil {
		r.logger.Error("sentence answer: load item", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if item == nil || set == nil || set.UserID != userID {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if item.AttemptedAt != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "already_attempted"})
		return
	}

	aiSvc := r.aiSvc()
	if aiSvc == nil {
		http.Error(w, "AI service unavailable", http.StatusServiceUnavailable)
		return
	}
	// Grade with the default model (5.5-nano by config) — the grading prompt is tuned for it.
	grade, err := aiSvc.GradeSentenceForCourse(req.Context(), set.CourseCode, item.PromptRU, item.ReferenceES, strings.TrimSpace(body.UserInput))
	if err != nil {
		r.logger.Error("sentence answer: grading failed", zap.Error(err))
		http.Error(w, "Grading failed", http.StatusBadGateway)
		return
	}
	// Outcome is derived from the error count server-side so it always matches the score rule.
	outcome := models.OutcomeForErrorCount(grade.ErrorCount)
	grade.Outcome = outcome
	gradingJSON, _ := json.Marshal(grade)

	updated, err := repo.RecordAttempt(item.ID, strings.TrimSpace(body.UserInput), grade.ErrorCount, outcome, string(gradingJSON))
	if err != nil {
		r.logger.Error("sentence answer: record attempt", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Progress/stats: count as a "sentences" mode attempt (feeds streak & per-mode skills).
	r.bumpSentenceStats(req, userID, set.CourseCode, outcome)

	items, _ := repo.GetItems(set.ID)
	total, attempted := setProgress(items)
	writeJSON(w, map[string]interface{}{
		"grading":     grade,
		"outcome":     outcome,
		"error_count": grade.ErrorCount,
		"done":        updated.Status == models.SentenceSetCompleted,
		"set": map[string]interface{}{
			"status":    updated.Status,
			"stars":     updated.StarCount,
			"passed":    updated.PassedCount,
			"failed":    updated.FailedCount,
			"total":     total,
			"attempted": attempted,
		},
	})
}

func (r *Router) bumpSentenceStats(req *http.Request, userID int64, courseCode, outcome string) {
	statsRepo := repository.NewLinglowDailyStatsRepository(r.db)
	ucID, err := statsRepo.ResolveUserCourseID(req.Context(), userID, courseCode)
	if err != nil {
		return
	}
	correct := 0
	if outcome != models.SentenceOutcomeFailed {
		correct = 1
	}
	_ = statsRepo.Bump(req.Context(), repository.DailyBump{
		UserCourseID: ucID,
		Day:          repository.LocalDayFromTime(time.Now()),
		Mode:         "sentences",
		Attempts:     1,
		Correct:      correct,
	})
}
