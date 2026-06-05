package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func (r *Router) defaultCourseCode() string {
	if r == nil || r.config == nil {
		return ""
	}
	return repository.CourseCodeForLearning(r.config.Learning)
}

func (r *Router) handleCourses(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	courses, err := r.courseRepo.ListCoursesForUser(req.Context(), userID, r.defaultCourseCode())
	if err != nil {
		r.logger.Error("failed to list courses", zap.Error(err), zap.Int64("user_id", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"courses": courses})
}

func (r *Router) handleCurrentCourse(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	current, err := r.courseRepo.GetCurrentCourse(req.Context(), userID, r.defaultCourseCode())
	if err != nil {
		r.logger.Error("failed to get current course", zap.Error(err), zap.Int64("user_id", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, current)
}

func (r *Router) handleSelectCourse(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		CourseCode string `json:"course_code"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	body.CourseCode = strings.TrimSpace(strings.ToLower(body.CourseCode))
	if body.CourseCode == "" {
		http.Error(w, "course_code is required", http.StatusBadRequest)
		return
	}
	current, err := r.courseRepo.SelectCurrentCourse(req.Context(), userID, body.CourseCode)
	if err != nil {
		r.logger.Warn("failed to select current course", zap.Error(err), zap.Int64("user_id", userID), zap.String("course_code", body.CourseCode))
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	writeJSON(w, current)
}

// handleLearningCourse returns the Linglow v2 course map for compatibility with the old route.
func (r *Router) handleLearningCourse(w http.ResponseWriter, req *http.Request) {
	r.handleCourseMap(w, req)
}

// handleLinglowCity returns the Linglow v2 course map for the selected or requested course.
func (r *Router) handleLinglowCity(w http.ResponseWriter, req *http.Request) {
	r.handleCourseMap(w, req)
}

func (r *Router) handleCourseMap(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}

	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	explicitCourseCode := req.URL.Query().Get("course_code")
	courseMap, err := r.courseRepo.GetCourseMapForUser(req.Context(), userID, r.defaultCourseCode(), explicitCourseCode)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to get learning course map", zap.Error(err), zap.Int64("user_id", userID), zap.String("course_code", explicitCourseCode))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(courseMap); err != nil {
		r.logger.Error("failed to encode learning course map", zap.Error(err))
	}
}

func (r *Router) handleLinglowDailyRoute(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	limit := 8
	if parsed, ok := parsePositiveLimit(w, req, limit); ok {
		limit = parsed
	} else {
		return
	}
	explicitCourseCode := req.URL.Query().Get("course_code")
	route, err := r.courseRepo.GetDailyRouteForUser(req.Context(), userID, r.defaultCourseCode(), explicitCourseCode, limit)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to get daily route", zap.Error(err), zap.Int64("user_id", userID), zap.String("course_code", explicitCourseCode))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, route)
}

func (r *Router) handleLinglowReview(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	limit, ok := parsePositiveLimit(w, req, 20)
	if !ok {
		return
	}
	explicitCourseCode := req.URL.Query().Get("course_code")
	queue, err := r.courseRepo.GetReviewQueueForUser(req.Context(), userID, r.defaultCourseCode(), explicitCourseCode, limit)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to get review queue", zap.Error(err), zap.Int64("user_id", userID), zap.String("course_code", explicitCourseCode))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, queue)
}

func (r *Router) handleLinglowProgress(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	explicitCourseCode := req.URL.Query().Get("course_code")
	progress, err := r.courseRepo.GetProgressForUser(req.Context(), userID, r.defaultCourseCode(), explicitCourseCode)
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("failed to get Linglow progress", zap.Error(err), zap.Int64("user_id", userID), zap.String("course_code", explicitCourseCode))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, progress)
}

func (r *Router) handleLinglowExerciseAttempts(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.courseRepo == nil {
		http.Error(w, "Course repository is not available", http.StatusServiceUnavailable)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		CourseCode      string          `json:"course_code"`
		LearningItemID  int64           `json:"learning_item_id"`
		SRSItemID       int64           `json:"srs_item_id"`
		Mode            string          `json:"mode"`
		ClientAttemptID string          `json:"client_attempt_id"`
		IsCorrect       *bool           `json:"is_correct"`
		Score           *int            `json:"score"`
		Quality         *int            `json:"quality"`
		Prompt          json.RawMessage `json:"prompt"`
		Answer          json.RawMessage `json:"answer"`
		Result          json.RawMessage `json:"result"`
		AnsweredAt      string          `json:"answered_at"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Mode) == "" {
		http.Error(w, "mode is required", http.StatusBadRequest)
		return
	}
	promptJSON, ok := rawObjectJSON(w, body.Prompt, "prompt")
	if !ok {
		return
	}
	answerJSON, ok := rawObjectJSON(w, body.Answer, "answer")
	if !ok {
		return
	}
	resultJSON, ok := rawObjectJSON(w, body.Result, "result")
	if !ok {
		return
	}
	var answeredAt time.Time
	if strings.TrimSpace(body.AnsweredAt) != "" {
		parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(body.AnsweredAt))
		if err != nil {
			http.Error(w, "answered_at must be RFC3339", http.StatusBadRequest)
			return
		}
		answeredAt = parsed
	}
	result, err := r.courseRepo.RecordExerciseAttempt(req.Context(), repository.ExerciseAttemptInput{
		UserID:          userID,
		DefaultCourse:   r.defaultCourseCode(),
		ExplicitCourse:  body.CourseCode,
		LearningItemID:  body.LearningItemID,
		SRSItemID:       body.SRSItemID,
		Mode:            body.Mode,
		ClientAttemptID: body.ClientAttemptID,
		IsCorrect:       body.IsCorrect,
		Score:           body.Score,
		Quality:         body.Quality,
		PromptJSON:      promptJSON,
		AnswerJSON:      answerJSON,
		ResultJSON:      resultJSON,
		AnsweredAt:      answeredAt,
		UpdateSRS:       r.config != nil && r.config.Linglow.SRSWriteEnabled,
	})
	if err != nil {
		if errors.Is(err, repository.ErrCourseNotFound) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Warn("failed to record Linglow exercise attempt", zap.Error(err), zap.Int64("user_id", userID), zap.String("course_code", body.CourseCode))
		http.Error(w, "Invalid exercise attempt", http.StatusBadRequest)
		return
	}
	writeJSON(w, result)
}

func parsePositiveLimit(w http.ResponseWriter, req *http.Request, defaultLimit int) (int, bool) {
	rawLimit := strings.TrimSpace(req.URL.Query().Get("limit"))
	if rawLimit == "" {
		return defaultLimit, true
	}
	parsed, err := strconv.Atoi(rawLimit)
	if err != nil || parsed < 1 {
		http.Error(w, "Invalid limit", http.StatusBadRequest)
		return 0, false
	}
	return parsed, true
}

func rawObjectJSON(w http.ResponseWriter, raw json.RawMessage, field string) (string, bool) {
	if len(raw) == 0 {
		return "{}", true
	}
	var object map[string]interface{}
	if err := json.Unmarshal(raw, &object); err != nil {
		http.Error(w, field+" must be a JSON object", http.StatusBadRequest)
		return "", false
	}
	return string(raw), true
}

func writeJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
