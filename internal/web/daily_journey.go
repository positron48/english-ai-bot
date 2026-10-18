package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/i18n"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

// SetDailyJourneyService is intentionally explicit: no fixture/catalog fallback
// may enable production. Until all subject adapters are ready the flag is inert.
func (r *Router) SetDailyJourneyService(s *service.DailyJourneyService) { r.dailyJourneyService = s }
func (r *Router) dailyJourneyEnabled() bool {
	return r.courseRepo != nil && r.config != nil && r.config.Linglow.DailyJourneyEnabled && r.dailyJourneyService.Ready()
}

func decodeJourneyBody(w http.ResponseWriter, req *http.Request, body any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 32<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(body); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return dailyjourney.ErrInvalidAction
	}
	return nil
}
func (r *Router) journeyError(w http.ResponseWriter, req *http.Request, err error) {
	status, code := http.StatusInternalServerError, "daily_journey_failed"
	switch {
	case errors.Is(err, dailyjourney.ErrNotFound), errors.Is(err, repository.ErrCourseNotFound):
		status = http.StatusNotFound
		code = "daily_journey_not_found"
	case errors.Is(err, dailyjourney.ErrProcessing):
		status = http.StatusServiceUnavailable
		code = "daily_journey_processing"
		w.Header().Set("Retry-After", "2")
	case errors.Is(err, dailyjourney.ErrReservationLost):
		status = http.StatusConflict
		code = "daily_journey_reload"
	case errors.Is(err, dailyjourney.ErrAIProvider):
		status = http.StatusBadGateway
		code = "daily_journey_ai_failed"
	case errors.Is(err, dailyjourney.ErrConflict):
		status = http.StatusConflict
		code = "daily_journey_reload"
	case errors.Is(err, dailyjourney.ErrActionKey):
		status = http.StatusConflict
		code = "daily_journey_action_key_reused"
	case errors.Is(err, dailyjourney.ErrAccess), errors.Is(err, service.ErrDailyJourneyContentUnavailable):
		status = http.StatusForbidden
		code = "daily_journey_access_denied"
	case errors.Is(err, dailyjourney.ErrInvalidAction):
		status = http.StatusUnprocessableEntity
		code = "daily_journey_invalid_action"
	default:
		if r.logger != nil {
			r.logger.Error("daily journey request failed", zap.Error(err))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": i18n.T(i18n.DetectLanguageFromRequest(req), "dailyJourney."+code), "reload": status == http.StatusConflict})
}
func (r *Router) handleDailyJourney(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if req.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !r.dailyJourneyEnabled() {
		writeJSON(w, map[string]any{"enabled": false, "journey": nil})
		return
	}
	course, err := r.courseRepo.ResolveRequestedCourseCode(req.Context(), userID, r.defaultCourseCode(), req.URL.Query().Get("course_code"))
	if err != nil {
		r.journeyError(w, req, err)
		return
	}
	j, err := r.dailyJourneyService.Current(req.Context(), userID, course)
	if err != nil {
		r.journeyError(w, req, err)
		return
	}
	if req.URL.Query().Get("summary_only") == "1" {
		writeJSON(w, map[string]any{"enabled": true, "courseCode": course, "journey": compactDailyJourney(j)})
		return
	}
	writeJSON(w, map[string]any{"enabled": true, "journey": j})
}
func (r *Router) handleDailyJourneys(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if !r.dailyJourneyEnabled() {
		http.Error(w, "Daily journey is not available", http.StatusServiceUnavailable)
		return
	}
	var course string
	if req.URL.Path == "/api/linglow/daily-journeys" {
		if req.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			CourseCode string `json:"course_code"`
		}
		if err := decodeJourneyBody(w, req, &body); err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}
		var err error
		course, err = r.courseRepo.ResolveRequestedCourseCode(req.Context(), userID, r.defaultCourseCode(), body.CourseCode)
		if err != nil {
			r.journeyError(w, req, err)
			return
		}
		j, err := r.dailyJourneyService.Start(req.Context(), userID, course)
		if err != nil {
			r.journeyError(w, req, err)
			return
		}
		writeJSON(w, j)
		return
	}
	parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/api/linglow/daily-journeys/"), "/")
	var err error
	course, err = r.courseRepo.ResolveRequestedCourseCode(req.Context(), userID, r.defaultCourseCode(), req.URL.Query().Get("course_code"))
	if err != nil {
		r.journeyError(w, req, err)
		return
	}
	if len(parts) == 1 && parts[0] != "" {
		if req.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		j, err := r.dailyJourneyService.Get(req.Context(), userID, course, parts[0])
		if err != nil {
			r.journeyError(w, req, err)
			return
		}
		writeJSON(w, j)
		return
	}
	if len(parts) == 4 && parts[0] != "" && parts[1] == "steps" && parts[2] != "" && parts[3] == "exercise" {
		if req.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		view, err := r.dailyJourneyService.Exercise(req.Context(), userID, course, parts[0], parts[2])
		if err != nil {
			r.journeyError(w, req, err)
			return
		}
		writeJSON(w, view)
		return
	}
	if len(parts) == 4 && parts[0] != "" && parts[1] == "steps" && parts[2] != "" && parts[3] == "actions" {
		if req.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var action dailyjourney.Action
		if err := decodeJourneyBody(w, req, &action); err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}
		j, err := r.dailyJourneyService.Act(req.Context(), userID, course, parts[0], parts[2], action)
		if err != nil {
			r.journeyError(w, req, err)
			return
		}
		r.BumpUserCache(userID)
		writeJSON(w, j)
		return
	}
	http.NotFound(w, req)
}

// Dashboard cache contains navigation progress only, never exercise state or feedback.
func compactDailyJourney(j *dailyjourney.Journey) any {
	if j == nil {
		return nil
	}
	steps := make([]map[string]any, 0, len(j.Steps))
	for _, step := range j.Steps {
		steps = append(steps, map[string]any{"id": step.ID, "title": step.Title, "status": step.Status, "mode": step.Mode})
	}
	return map[string]any{"id": j.ID, "courseCode": j.CourseCode, "date": j.Date, "status": j.Status, "revision": j.Revision, "steps": steps}
}
