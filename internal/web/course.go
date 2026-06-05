package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	if rawLimit := strings.TrimSpace(req.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 {
			http.Error(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsed
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

func writeJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
