package web

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// handleLearningCourse returns the Linglow v2 course map for the current app language.
func (r *Router) handleLearningCourse(w http.ResponseWriter, req *http.Request) {
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

	courseMap, err := r.courseRepo.GetCourseMapForLearning(req.Context(), r.config.Learning, userID)
	if err != nil {
		r.logger.Error("failed to get learning course map", zap.Error(err), zap.Int64("user_id", userID))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(courseMap); err != nil {
		r.logger.Error("failed to encode learning course map", zap.Error(err))
	}
}
