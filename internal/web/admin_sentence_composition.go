package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// handleAdminSentenceCompositionUsers returns one summary row per user that has any
// sentence-composition set, for the admin "results by users" screen.
// GET /api/admin/sentence-composition/users
func (r *Router) handleAdminSentenceCompositionUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	overviews, err := r.sentenceRepo().ListUserOverviews()
	if err != nil {
		r.logger.Error("admin sentence: list user overviews", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{
		"users":   overviews,
		"enabled": r.sentenceWorker != nil,
	})
}

// handleAdminSentenceCompositionUserDetail returns a user's sets with full item detail,
// or force-generates a new set for that user.
//
//	GET  /api/admin/sentence-composition/users/{id}          -> sets + items
//	POST /api/admin/sentence-composition/users/{id}/generate -> force generate a new set
//	DELETE /api/admin/sentence-composition/users/{id}/sets/{setID} -> remove a ready set
func (r *Router) handleAdminSentenceCompositionUserDetail(w http.ResponseWriter, req *http.Request) {
	rest := strings.TrimPrefix(req.URL.Path, "/api/admin/sentence-composition/users/")
	rest = strings.Trim(rest, "/")
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	// POST .../generate — force a new set.
	if len(parts) == 2 && parts[1] == "generate" {
		r.adminForceGenerateSentenceSet(w, req, userID)
		return
	}
	if len(parts) == 3 && parts[1] == "sets" {
		setID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || setID <= 0 {
			http.Error(w, "Invalid set id", http.StatusBadRequest)
			return
		}
		r.adminDeleteReadySentenceSet(w, req, userID, setID)
		return
	}
	if len(parts) != 1 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	repo := r.sentenceRepo()
	sets, err := repo.ListSetsByUser(userID, 50)
	if err != nil {
		r.logger.Error("admin sentence: list sets", zap.Error(err), zap.Int64("user_id", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]interface{}, 0, len(sets))
	for _, set := range sets {
		items, err := repo.GetItems(set.ID)
		if err != nil {
			r.logger.Error("admin sentence: load items", zap.Error(err), zap.Int64("set_id", set.ID))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		itemsJSON := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			m := sentenceItemJSON(it, true)
			m["reference_es"] = it.ReferenceES
			itemsJSON = append(itemsJSON, m)
		}
		total, attempted := setProgress(items)
		out = append(out, map[string]interface{}{
			"id":              set.ID,
			"course_code":     set.CourseCode,
			"generation_date": set.GenerationDate,
			"status":          set.Status,
			"scopes":          set.Scopes,
			"total":           total,
			"attempted":       attempted,
			"stars":           set.StarCount,
			"passed":          set.PassedCount,
			"failed":          set.FailedCount,
			"started_at":      set.StartedAt,
			"completed_at":    set.CompletedAt,
			"created_at":      set.CreatedAt,
			"items":           itemsJSON,
		})
	}
	writeJSON(w, map[string]interface{}{
		"user_id":           userID,
		"sets":              out,
		"enabled":           r.sentenceWorker != nil,
		"available_courses": r.sentenceCoursesForUser(req, userID),
	})
}

func (r *Router) sentenceCoursesForUser(req *http.Request, userID int64) []string {
	if r.sentenceWorker == nil {
		return nil
	}
	courses, err := r.sentenceWorker.SentenceCourseCodesForUser(req.Context(), userID)
	if err != nil {
		r.logger.Warn("admin sentence: list available courses", zap.Error(err), zap.Int64("user_id", userID))
		return nil
	}
	return courses
}

func (r *Router) adminForceGenerateSentenceSet(w http.ResponseWriter, req *http.Request, userID int64) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.sentenceWorker == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "sentence composition is disabled"})
		return
	}
	var input struct {
		CourseCode string `json:"course_code"`
	}
	if req.Body != nil {
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil && !errors.Is(err, io.EOF) {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	var setID int64
	var err error
	if strings.TrimSpace(input.CourseCode) == "" {
		setID, err = r.sentenceWorker.ForceGenerateForUser(req.Context(), userID)
	} else {
		setID, err = r.sentenceWorker.ForceGenerateForUserCourse(req.Context(), userID, input.CourseCode)
	}
	if err != nil {
		r.logger.Warn("admin sentence: force generate failed", zap.Error(err), zap.Int64("user_id", userID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if setID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not enough well-learned words to build a set"})
		return
	}
	writeJSON(w, map[string]interface{}{"generated": true, "set_id": setID})
}

func (r *Router) adminDeleteReadySentenceSet(w http.ResponseWriter, req *http.Request, userID, setID int64) {
	if req.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	deleted, err := r.sentenceRepo().DeleteReadySet(userID, setID)
	if err != nil {
		r.logger.Error("admin sentence: delete ready set", zap.Error(err), zap.Int64("user_id", userID), zap.Int64("set_id", setID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "Ready set not found", http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]bool{"deleted": true})
}
