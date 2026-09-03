package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"tgbot-skeleton/internal/placement"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

// Result links follow current publication state, including the parent section.
// Historical answers and grading remain pinned; withdrawn course links disappear.
func (r *Router) placementGrammar(course string) *service.GrammarService {
	if g := r.grammarServices[grammarBundleForCourse(course)]; g != nil {
		return g
	}
	if r.grammarService != nil && r.grammarService.AttemptRepo.CourseCode() == course {
		return r.grammarService
	}
	return nil
}
func (r *Router) filterPlacementLinks(req *http.Request, v *service.PlacementSessionView) error {
	if v == nil || v.Result == nil {
		return nil
	}
	g := r.placementGrammar(v.CourseCode)
	allowed := map[string]bool{}
	if g != nil {
		sections, err := g.ContentRepo.GetSections()
		if err != nil {
			return err
		}
		parents, err := g.PublishRepo.GetPublishedItemsByType("section")
		if err != nil {
			return err
		}
		chapters, err := g.PublishRepo.GetPublishedItemsByType("chapter")
		if err != nil {
			return err
		}
		for _, sec := range sections.Sections {
			if p := parents[sec.SectionID]; p == nil || !p.IsPublished {
				continue
			}
			for _, id := range sec.ChapterIDs {
				if p := chapters[id]; p != nil && p.IsPublished {
					allowed[id] = true
				}
			}
		}
	}
	filter := func(ids []string) []string {
		out := []string{}
		for _, id := range ids {
			if allowed[id] {
				out = append(out, id)
			}
		}
		return out
	}
	for i := range v.Result.Review {
		v.Result.Review[i].ChapterIDs = filter(v.Result.Review[i].ChapterIDs)
	}
	recommended := []placement.Skill{}
	for _, sk := range v.Result.RecommendedSkills {
		sk.ChapterIDs = filter(sk.ChapterIDs)
		if len(sk.ChapterIDs) > 0 {
			recommended = append(recommended, sk)
		}
	}
	v.Result.RecommendedSkills = recommended
	// Prefer a usable lesson for each topic; a published lesson can still be
	// locked by the course sequence. Preserve the topic even when all are locked.
	checked := map[string]bool{}
	v.AvailableChapterIDs = []string{}
	choose := func(ids []string) error {
		for _, id := range ids {
			accessible, exists := checked[id]
			if !exists {
				var err error
				accessible, err = g.CanAccessChapter(req.Context(), getUserIDFromContext(req.Context()), id)
				if err != nil {
					return err
				}
				checked[id] = accessible
				if accessible {
					v.AvailableChapterIDs = append(v.AvailableChapterIDs, id)
				}
			}
			if accessible {
				break
			}
		}
		return nil
	}
	for _, sk := range recommended {
		if err := choose(sk.ChapterIDs); err != nil {
			return err
		}
	}
	for _, q := range v.Result.Review {
		if err := choose(q.ChapterIDs); err != nil {
			return err
		}
	}
	return nil
}
func (r *Router) placementRespond(w http.ResponseWriter, req *http.Request, v *service.PlacementSessionView) {
	if err := r.filterPlacementLinks(req, v); err != nil {
		placementError(w, err)
		return
	}
	placementJSON(w, 200, v)
}

func (r *Router) placementService() *service.PlacementService {
	return service.NewPlacementService(r.db, r.config.Learning.ContentSource)
}

// Grammar/placement course selection does not depend on the word catalog's contents.
func (r *Router) placementCourseCode(req *http.Request, user int64) string {
	if code := req.URL.Query().Get("course_code"); code != "" {
		return code
	}
	if r.courseRepo != nil {
		if code, err := r.courseRepo.ResolveCurrentCourseCode(req.Context(), user, r.defaultCourseCode()); err == nil {
			return code
		}
	}
	return r.defaultCourseCode()
}
func placementJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func placementDecode(w http.ResponseWriter, req *http.Request, v interface{}) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, req.Body, 8192))
	d.DisallowUnknownFields()
	err := d.Decode(v)
	if err == nil {
		var extra interface{}
		err = d.Decode(&extra)
		if errors.Is(err, io.EOF) {
			return true
		}
	}
	placementJSON(w, 400, map[string]string{"code": "placement_invalid_request", "message": "Некорректный запрос теста."})
	return false
}
func placementError(w http.ResponseWriter, err error) {
	status, code, msg := 503, "placement_unavailable", "Тест временно недоступен. Попробуйте позже."
	switch {
	case errors.Is(err, repository.ErrPlacementNotFound):
		status, code, msg = 404, "placement_not_found", "Попытка не найдена."
	case errors.Is(err, repository.ErrPlacementExpired):
		status, code, msg = 410, "placement_expired", "Срок этой попытки истёк. Начните новый тест."
	case errors.Is(err, repository.ErrPlacementConflict):
		status, code, msg = 409, "placement_conflict", "Состояние теста изменилось. Обновите страницу и продолжите."
	case errors.Is(err, repository.ErrPlacementAnswer):
		status, code, msg = 400, "placement_invalid_answer", "Проверьте выбранный ответ."
	case errors.Is(err, repository.ErrCourseNotFound):
		status, code, msg = 404, "placement_course_not_found", "Курс не найден."
	}
	placementJSON(w, status, map[string]string{"code": code, "message": msg})
}

type placementStartRequest struct {
	CourseCode     string `json:"course_code"`
	IdempotencyKey string `json:"idempotency_key"`
	NewAttempt     bool   `json:"new_attempt"`
}

func (r *Router) handlePlacementSessions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		placementJSON(w, 405, map[string]string{"code": "method_not_allowed"})
		return
	}
	user := getUserIDFromContext(req.Context())
	if user == 0 {
		placementJSON(w, 401, map[string]string{"code": "unauthorized"})
		return
	}
	var body placementStartRequest
	if !placementDecode(w, req, &body) {
		return
	}
	if body.CourseCode == "" {
		body.CourseCode = r.placementCourseCode(req, user)
	}
	if body.CourseCode != "en_ru" && body.CourseCode != "es_ru" {
		placementError(w, repository.ErrCourseNotFound)
		return
	}
	v, err := r.placementService().Start(req.Context(), user, body.CourseCode, body.IdempotencyKey, body.NewAttempt)
	if err != nil {
		placementError(w, err)
		return
	}
	r.placementRespond(w, req, v)
}
func (r *Router) handlePlacementSession(w http.ResponseWriter, req *http.Request) {
	user := getUserIDFromContext(req.Context())
	if user == 0 {
		placementJSON(w, 401, map[string]string{"code": "unauthorized"})
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/api/learning/placement/sessions/")
	parts := strings.Split(path, "/")
	if len(parts) > 2 || len(parts[0]) != 32 {
		placementError(w, repository.ErrPlacementNotFound)
		return
	}
	svc := r.placementService()
	var v *service.PlacementSessionView
	var err error
	switch {
	case len(parts) == 1 && req.Method == http.MethodGet:
		v, err = svc.Get(req.Context(), user, parts[0])
	case len(parts) == 2 && parts[1] == "answers" && req.Method == http.MethodPost:
		var body struct {
			QuestionID string  `json:"question_id"`
			Answer     *string `json:"answer"`
		}
		if !placementDecode(w, req, &body) {
			return
		}
		v, err = svc.Answer(req.Context(), user, parts[0], body.QuestionID, body.Answer)
	case len(parts) == 2 && parts[1] == "finish" && req.Method == http.MethodPost:
		v, err = svc.Finish(req.Context(), user, parts[0], r.placementGrammar)
	default:
		placementJSON(w, 405, map[string]string{"code": "method_not_allowed"})
		return
	}
	if err != nil {
		placementError(w, err)
		return
	}
	r.placementRespond(w, req, v)
}
func (r *Router) handlePlacementResults(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		placementJSON(w, 405, map[string]string{"code": "method_not_allowed"})
		return
	}
	user := getUserIDFromContext(req.Context())
	if user == 0 {
		placementJSON(w, 401, map[string]string{"code": "unauthorized"})
		return
	}
	course := req.URL.Query().Get("course_code")
	if course == "" {
		course = r.placementCourseCode(req, user)
	}
	if course != "en_ru" && course != "es_ru" {
		placementError(w, repository.ErrCourseNotFound)
		return
	}
	v, err := r.placementService().History(req.Context(), user, course)
	if err != nil {
		placementError(w, err)
		return
	}
	for _, session := range v {
		if err := r.filterPlacementLinks(req, session); err != nil {
			placementError(w, err)
			return
		}
	}
	placementJSON(w, 200, map[string]interface{}{"sessions": v})
}

// Old clients must update rather than submit arbitrary chapter IDs as placement.
func (r *Router) handleLegacyPlacementRetired(w http.ResponseWriter, req *http.Request) {
	placementJSON(w, 410, map[string]string{"code": "placement_replaced", "message": "Откройте новый тест определения уровня.", "path": "/learning/placement-test"})
}
