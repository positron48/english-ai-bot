package web

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// adminPictureQuest is the JSON shape of a picture quest in the admin UI.
type adminPictureQuest struct {
	ID               int64                   `json:"id"`
	Code             string                  `json:"code"`
	Title            string                  `json:"title"`
	CEFRLevel        string                  `json:"cefr_level"`
	ImageURL         string                  `json:"image_url"`
	ImageDescription string                  `json:"image_description"`
	MaxTurns         int                     `json:"max_turns"`
	TokenBudget      int                     `json:"token_budget"`
	SortOrder        int                     `json:"sort_order"`
	Status           string                  `json:"status"`
	Tasks            []adminPictureQuestTask `json:"tasks"`
}

type adminPictureQuestTask struct {
	ID                 int64  `json:"id"`
	Code               string `json:"code"`
	Title              string `json:"title"`
	CompletionCriteria string `json:"completion_criteria"`
	IsRequired         bool   `json:"is_required"`
	SortOrder          int    `json:"sort_order"`
}

// handleAdminPictureQuests handles GET (list) and POST (create) of picture quests.
// @Summary      Admin: список/создание квестов «опиши картинку»
// @Tags         Admin
// @Router       /api/admin/picture-quests [get]
func (r *Router) handleAdminPictureQuests(w http.ResponseWriter, req *http.Request) {
	if r.pictureQuestRepo == nil {
		http.Error(w, "Picture quest repository is not available", http.StatusServiceUnavailable)
		return
	}
	switch req.Method {
	case http.MethodGet:
		r.adminPictureQuestsList(w, req)
	case http.MethodPost:
		r.adminPictureQuestCreate(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) adminPictureQuestsList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	courseCode := r.adminConversationCourseCode(req)
	courseID, err := r.pictureQuestRepo.CourseIDByCode(ctx, courseCode)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	quests, err := r.pictureQuestRepo.ListQuestsForCourseAdmin(ctx, courseID)
	if err != nil {
		r.logger.Error("admin list picture quests", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	out := make([]adminPictureQuest, 0, len(quests))
	for i := range quests {
		q := &quests[i]
		tasks, err := r.pictureQuestRepo.ListTasks(ctx, q.ID)
		if err != nil {
			r.logger.Error("admin list picture quest tasks", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		taskOut := make([]adminPictureQuestTask, 0, len(tasks))
		for _, t := range tasks {
			taskOut = append(taskOut, adminPictureQuestTask{
				ID: t.ID, Code: t.Code, Title: t.Title,
				CompletionCriteria: t.CompletionCriteria, IsRequired: t.IsRequired, SortOrder: t.SortOrder,
			})
		}
		out = append(out, adminPictureQuest{
			ID: q.ID, Code: q.Code, Title: q.Title, CEFRLevel: q.CEFRLevel,
			ImageURL: q.ImageURL, ImageDescription: q.ImageDescription,
			MaxTurns: q.MaxTurns, TokenBudget: q.TokenBudget,
			SortOrder: q.SortOrder, Status: q.Status,
			Tasks: taskOut,
		})
	}

	levels, err := r.pictureQuestRepo.ListCourseLevels(ctx, courseID)
	if err != nil {
		r.logger.Error("admin list course levels", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	levelOut := make([]map[string]string, 0, len(levels))
	for _, l := range levels {
		levelOut = append(levelOut, map[string]string{"level_code": l.LevelCode, "title": l.Title})
	}

	writeJSON(w, map[string]interface{}{
		"course_code": courseCode,
		"quests":      out,
		"levels":      levelOut,
	})
}

// pictureQuestPayload is the request body for create/update.
type pictureQuestPayload struct {
	Code             string `json:"code"`
	Title            string `json:"title"`
	CEFRLevel        string `json:"cefr_level"`
	ImageURL         string `json:"image_url"`
	ImageDescription string `json:"image_description"`
	MaxTurns         int    `json:"max_turns"`
	TokenBudget      int    `json:"token_budget"`
	SortOrder        int    `json:"sort_order"`
	Status           string `json:"status"`
}

func (p *pictureQuestPayload) toInput(courseCode string) (repository.AdminPictureQuestInput, error) {
	in := repository.AdminPictureQuestInput{
		CourseCode:       courseCode,
		CEFRLevel:        strings.TrimSpace(p.CEFRLevel),
		Code:             strings.TrimSpace(p.Code),
		Title:            strings.TrimSpace(p.Title),
		ImageURL:         strings.TrimSpace(p.ImageURL),
		ImageDescription: strings.TrimSpace(p.ImageDescription),
		MaxTurns:         p.MaxTurns,
		TokenBudget:      p.TokenBudget,
		SortOrder:        p.SortOrder,
		Status:           strings.TrimSpace(p.Status),
	}
	if in.Code == "" || in.Title == "" || in.CEFRLevel == "" {
		return in, errors.New("code, title and cefr_level are required")
	}
	if in.ImageURL == "" {
		return in, errors.New("image_url is required — upload the picture first")
	}
	if in.ImageDescription == "" {
		return in, errors.New("image_description is required — it is the ground truth the model judges the learner against")
	}
	if in.MaxTurns <= 0 {
		in.MaxTurns = 30
	}
	if in.TokenBudget <= 0 {
		// Each turn replays the system prompt (incl. image description) + recent history.
		in.TokenBudget = 40000
	}
	switch in.Status {
	case "draft", "active", "locked", "archived":
	default:
		in.Status = "draft"
	}
	return in, nil
}

func (r *Router) adminPictureQuestCreate(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	courseCode := r.adminConversationCourseCode(req)
	var p pictureQuestPayload
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	in, err := p.toInput(courseCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := r.pictureQuestRepo.CreateQuest(ctx, in)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicatePictureQuestCode) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No district/conversation location for this level — check the course level code", http.StatusBadRequest)
			return
		}
		r.logger.Error("admin create picture quest", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"id": id, "success": true})
}

// handleAdminPictureQuestByID handles PUT/DELETE /picture-quests/{id} and POST /picture-quests/{id}/tasks.
func (r *Router) handleAdminPictureQuestByID(w http.ResponseWriter, req *http.Request) {
	if r.pictureQuestRepo == nil {
		http.Error(w, "Picture quest repository is not available", http.StatusServiceUnavailable)
		return
	}
	path := strings.Trim(strings.TrimPrefix(req.URL.Path, "/api/admin/picture-quests/"), "/")
	parts := strings.Split(path, "/")
	if parts[0] == "tasks" {
		r.handleAdminPictureQuestTaskByID(w, req)
		return
	}
	if parts[0] == "import" {
		r.handleAdminPictureQuestImport(w, req)
		return
	}
	if parts[0] == "prompt-template" {
		r.handleAdminPictureQuestPromptTemplate(w, req)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid quest id", http.StatusBadRequest)
		return
	}
	ctx := req.Context()

	switch {
	case len(parts) == 1 && req.Method == http.MethodPut:
		var p pictureQuestPayload
		if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		courseCode := r.adminConversationCourseCode(req)
		in, perr := p.toInput(courseCode)
		if perr != nil {
			http.Error(w, perr.Error(), http.StatusBadRequest)
			return
		}
		if err := r.pictureQuestRepo.UpdateQuest(ctx, id, in); err != nil {
			if errors.Is(err, repository.ErrDuplicatePictureQuestCode) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			r.logger.Error("admin update picture quest", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})

	case len(parts) == 1 && req.Method == http.MethodDelete:
		if err := r.pictureQuestRepo.DeleteQuest(ctx, id); err != nil {
			r.logger.Error("admin delete picture quest", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})

	case len(parts) == 2 && parts[1] == "tasks" && req.Method == http.MethodPost:
		r.adminPictureQuestTaskCreate(w, req, id)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (p *taskPayload) toPictureInput() (repository.AdminPictureTaskInput, error) {
	in := repository.AdminPictureTaskInput{
		Code:               strings.TrimSpace(p.Code),
		Title:              strings.TrimSpace(p.Title),
		CompletionCriteria: strings.TrimSpace(p.CompletionCriteria),
		IsRequired:         p.IsRequired,
		SortOrder:          p.SortOrder,
	}
	if in.Code == "" || in.Title == "" || in.CompletionCriteria == "" {
		return in, errors.New("code, title and completion_criteria are required")
	}
	return in, nil
}

func (r *Router) adminPictureQuestTaskCreate(w http.ResponseWriter, req *http.Request, questID int64) {
	var p taskPayload
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	in, err := p.toPictureInput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := r.pictureQuestRepo.CreateTask(req.Context(), questID, in)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicatePictureTaskCode) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		r.logger.Error("admin create picture quest task", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"id": id, "success": true})
}

// handleAdminPictureQuestTaskByID handles PUT/DELETE /picture-quests/tasks/{id}.
func (r *Router) handleAdminPictureQuestTaskByID(w http.ResponseWriter, req *http.Request) {
	idStr := strings.Trim(strings.TrimPrefix(req.URL.Path, "/api/admin/picture-quests/tasks/"), "/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid task id", http.StatusBadRequest)
		return
	}
	ctx := req.Context()

	switch req.Method {
	case http.MethodPut:
		var p taskPayload
		if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		in, perr := p.toPictureInput()
		if perr != nil {
			http.Error(w, perr.Error(), http.StatusBadRequest)
			return
		}
		if err := r.pictureQuestRepo.UpdateTask(ctx, id, in); err != nil {
			if errors.Is(err, repository.ErrDuplicatePictureTaskCode) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			r.logger.Error("admin update picture quest task", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	case http.MethodDelete:
		if err := r.pictureQuestRepo.DeleteTask(ctx, id); err != nil {
			r.logger.Error("admin delete picture quest task", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
