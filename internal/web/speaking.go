package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/speakingsync"

	"go.uber.org/zap"
)

// SyncSpeakingCatalogFromBundle loads speaking content from grammar bundle into DB.
func (r *Router) SyncSpeakingCatalogFromBundle(ctx context.Context) error {
	if r == nil || r.speakingCatalogRepo == nil {
		return nil
	}
	return speakingsync.SyncFromBundle(ctx, r.config, r.speakingCatalogRepo, r.logger)
}

func (r *Router) handleLearningSpeakingAvailability(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tier := r.getUserTierFromDB(req.Context())
	canAccess := r.speakingModeEnabled() && r.userAllowsFeature(req.Context(), "speaking")
	levels := []string{}
	if r.speakingCatalogRepo != nil {
		if snap, err := r.speakingCatalogRepo.LoadSnapshot(); err == nil && snap != nil {
			for _, c := range snap.Categories {
				if c != nil && c.Level != "" {
					levels = appendUniqueString(levels, c.Level)
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available":          r.speakingModeEnabled() && r.speakingCatalogRepo != nil,
		"subscription_tier":  string(tier),
		"can_access":         canAccess,
		"features":           models.UserFeaturesForTier(tier),
		"levels":             levels,
	})
}

func appendUniqueString(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

func (r *Router) handleLearningSpeakingCategories(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if getUserIDFromContext(req.Context()) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.speakingCatalogRepo == nil {
		http.Error(w, "Speaking not available", http.StatusNotFound)
		return
	}
	snap, err := r.speakingCatalogRepo.LoadSnapshot()
	if err != nil {
		r.logger.Error("speaking categories", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	type catResp struct {
		CategoryID string `json:"category_id"`
		Title      string `json:"title"`
		Level      string `json:"level"`
		Order      int    `json:"order"`
		TaskCount  int    `json:"task_count"`
	}
	out := make([]catResp, 0)
	for _, c := range snap.Categories {
		if c == nil {
			continue
		}
		out = append(out, catResp{
			CategoryID: c.CategoryID,
			Title:      c.Title,
			Level:      c.Level,
			Order:      c.Order,
			TaskCount:  len(c.TaskIDs),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"categories": out})
}

func (r *Router) handleLearningSpeakingCategoryTasks(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if getUserIDFromContext(req.Context()) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	categoryID := strings.TrimPrefix(req.URL.Path, "/api/learning/speaking/categories/")
	categoryID = strings.TrimSuffix(categoryID, "/tasks")
	categoryID = strings.Trim(categoryID, "/")
	if categoryID == "" || strings.Contains(categoryID, "/") || !strings.HasSuffix(strings.TrimPrefix(req.URL.Path, "/api/learning/speaking/categories/"), "/tasks") {
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}
	tasks, err := r.speakingCatalogRepo.ListTasksByCategory(categoryID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func (r *Router) handleLearningSpeakingSessions(w http.ResponseWriter, req *http.Request) {
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
		CategoryID string `json:"category_id"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	categoryID := strings.TrimSpace(body.CategoryID)
	if categoryID == "" {
		http.Error(w, "category_id required", http.StatusBadRequest)
		return
	}
	snap, err := r.speakingCatalogRepo.LoadSnapshot()
	if err != nil || snap.Categories[categoryID] == nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}
	taskIDs := append([]string(nil), snap.Categories[categoryID].TaskIDs...)
	limit := r.config.Speaking.SessionTaskCount
	if limit <= 0 {
		limit = 5
	}
	if len(taskIDs) > limit {
		taskIDs = taskIDs[:limit]
	}
	if len(taskIDs) == 0 {
		http.Error(w, "No tasks in category", http.StatusNotFound)
		return
	}
	session, err := r.speakingSessionRepo.CreateSession(userID, categoryID, taskIDs)
	if err != nil {
		r.logger.Error("create speaking session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r.speakingSessionJSON(session))
}

func (r *Router) handleLearningSpeakingSessionByID(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/speaking/sessions/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}
	sessionID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid session id", http.StatusBadRequest)
		return
	}
	if len(parts) == 2 && parts[1] == "submit" && req.Method == http.MethodPost {
		r.handleLearningSpeakingSubmit(w, req, sessionID, userID)
		return
	}
	if len(parts) == 2 && parts[1] == "next" && req.Method == http.MethodPost {
		r.handleLearningSpeakingNext(w, req, sessionID, userID)
		return
	}
	if len(parts) == 1 && req.Method == http.MethodGet {
		session, err := r.speakingSessionRepo.GetSession(sessionID, userID)
		if err != nil || session == nil {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(r.speakingSessionJSON(session))
		return
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (r *Router) speakingSessionJSON(session *repository.SpeakingSession) map[string]interface{} {
	var currentTask interface{}
	if session.CurrentTaskIndex >= 0 && session.CurrentTaskIndex < len(session.TaskIDs) {
		taskID := session.TaskIDs[session.CurrentTaskIndex]
		if doc, ok, _ := r.speakingCatalogRepo.GetTaskPublic(taskID); ok {
			currentTask = doc
		}
	}
	out := map[string]interface{}{
		"id":                 session.ID,
		"category_id":        session.CategoryID,
		"status":             session.Status,
		"task_ids":           session.TaskIDs,
		"current_task_index": session.CurrentTaskIndex,
		"current_task":       currentTask,
		"total_tasks":        len(session.TaskIDs),
	}
	if session.CompletedAt != nil {
		out["completed_at"] = session.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return out
}

func (r *Router) handleLearningSpeakingSubmit(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	if err := req.ParseMultipartForm(4 << 20); err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}
	taskID := strings.TrimSpace(req.FormValue("task_id"))
	mode := strings.TrimSpace(req.FormValue("mode"))
	if mode == "" {
		mode = "initial"
	}
	if taskID == "" {
		http.Error(w, "task_id required", http.StatusBadRequest)
		return
	}
	session, err := r.speakingSessionRepo.GetSession(sessionID, userID)
	if err != nil || session == nil || session.Status != "active" {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	file, header, err := req.FormFile("audio")
	if err != nil {
		http.Error(w, "audio file required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	audio, err := io.ReadAll(io.LimitReader(file, int64(r.config.Speaking.MaxAudioMB)*1024*1024+1))
	if err != nil {
		http.Error(w, "Failed to read audio", http.StatusBadRequest)
		return
	}
	maxBytes := r.config.Speaking.MaxAudioMB * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = 2 * 1024 * 1024
	}
	if len(audio) > maxBytes {
		http.Error(w, "Audio too large", http.StatusBadRequest)
		return
	}
	format := "webm"
	if header != nil {
		ct := strings.ToLower(header.Header.Get("Content-Type"))
		if strings.Contains(ct, "wav") {
			format = "wav"
		} else if strings.Contains(ct, "mpeg") || strings.Contains(ct, "mp3") {
			format = "mp3"
		}
	}

	taskFull, ok, err := r.speakingCatalogRepo.GetTaskFull(taskID)
	if err != nil || !ok {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	attemptCount, err := r.speakingSessionRepo.CountAttempts(sessionID, taskID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	maxAttempts := taskFull.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = r.config.Speaking.MaxAttemptsDefault
	}
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	attemptNo := attemptCount + 1

	if r.speakingEvaluator == nil {
		http.Error(w, "Speaking evaluator not configured", http.StatusServiceUnavailable)
		return
	}
	evalResult, evalErr := r.speakingEvaluator.Evaluate(req.Context(), taskFull, audio, format, attemptNo, mode)
	if evalErr != nil {
		r.logger.Error("speaking evaluation failed", zap.Error(evalErr))
		http.Error(w, "Evaluation failed", http.StatusBadGateway)
		return
	}

	meaning := evalResult.MeaningScore
	grammar := evalResult.GrammarScore
	pron := evalResult.PronunciationScore
	fluency := evalResult.FluencyScore
	isAcceptable := evalResult.IsAcceptable
	rec := &repository.SpeakingAttemptRecord{
		UserID:             userID,
		SessionID:          sessionID,
		TaskID:             taskID,
		AttemptNo:          attemptNo,
		Mode:               mode,
		UnderstoodAnswer:   evalResult.UnderstoodAnswer,
		MeaningScore:       &meaning,
		GrammarScore:       &grammar,
		PronunciationScore: &pron,
		FluencyScore:       &fluency,
		IsAcceptable:       &isAcceptable,
		AudioQuality:       evalResult.AudioQuality,
		FeedbackRU:         evalResult.ShortFeedbackRU,
		BetterVersion:      evalResult.BetterVersion,
		RepeatTask:         evalResult.RepeatTask,
	}
	attemptID, err := r.speakingSessionRepo.SaveAttempt(rec)
	if err != nil {
		r.logger.Error("save speaking attempt", zap.Error(err))
	} else {
		statsRepo := repository.NewLinglowDailyStatsRepository(r.db)
		if userCourseID, ucErr := statsRepo.ResolveUserCourseID(req.Context(), userID, r.currentCourseCodeForUser(req.Context(), userID)); ucErr == nil {
			correct := 0
			if isAcceptable {
				correct = 1
			}
			_ = statsRepo.Bump(req.Context(), repository.DailyBump{
				UserCourseID: userCourseID,
				Mode:         "speaking",
				Attempts:     1,
				Correct:      correct,
			})
		}
	}

	canAdvance := isAcceptable || attemptNo >= maxAttempts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"attempt_id":          attemptID,
		"attempt_no":          attemptNo,
		"max_attempts":        maxAttempts,
		"can_advance":         canAdvance,
		"understood_answer":   evalResult.UnderstoodAnswer,
		"meaning_score":       evalResult.MeaningScore,
		"grammar_score":       evalResult.GrammarScore,
		"pronunciation_score": evalResult.PronunciationScore,
		"fluency_score":       evalResult.FluencyScore,
		"is_acceptable":       evalResult.IsAcceptable,
		"audio_quality":       evalResult.AudioQuality,
		"short_feedback_ru":   evalResult.ShortFeedbackRU,
		"better_version":      evalResult.BetterVersion,
		"repeat_task":         evalResult.RepeatTask,
	})
}

func (r *Router) handleLearningSpeakingNext(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	session, err := r.speakingSessionRepo.GetSession(sessionID, userID)
	if err != nil || session == nil || session.Status != "active" {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	if session.CurrentTaskIndex >= len(session.TaskIDs)-1 {
		if err := r.speakingSessionRepo.AdvanceSession(sessionID, userID, session.CurrentTaskIndex+1, true); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		if err := r.speakingSessionRepo.AdvanceSession(sessionID, userID, session.CurrentTaskIndex+1, false); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
	updated, _ := r.speakingSessionRepo.GetSession(sessionID, userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r.speakingSessionJSON(updated))
}