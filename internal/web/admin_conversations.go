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

// adminConversationScenario is the JSON shape of a scenario in the admin UI.
type adminConversationScenario struct {
	ID          int64                   `json:"id"`
	Code        string                  `json:"code"`
	Title       string                  `json:"title"`
	CEFRLevel   string                  `json:"cefr_level"`
	PlaceType   string                  `json:"place_type"`
	NPCName     string                  `json:"npc_name"`
	NPCPersona  string                  `json:"npc_persona"`
	SceneSetup  string                  `json:"scene_setup"`
	IsQuest          bool                    `json:"is_quest"`
	MaxTurns         int                     `json:"max_turns"`
	TokenBudget      int                     `json:"token_budget"`
	NPCCode          string                  `json:"npc_code"`
	PrerequisiteCode string                  `json:"prerequisite_code"`
	ImageURL         string                  `json:"image_url"`
	SortOrder        int                     `json:"sort_order"`
	Status           string                  `json:"status"`
	Tasks            []adminConversationTask `json:"tasks"`
}

type adminConversationTask struct {
	ID                 int64  `json:"id"`
	Code               string `json:"code"`
	Title              string `json:"title"`
	CompletionCriteria string `json:"completion_criteria"`
	IsRequired         bool   `json:"is_required"`
	SortOrder          int    `json:"sort_order"`
}

// handleAdminConversationScenarios handles GET (list) and POST (create) of scenarios.
// @Summary      Admin: список/создание сценариев общения
// @Tags         Admin
// @Router       /api/admin/conversations/scenarios [get]
func (r *Router) handleAdminConversationScenarios(w http.ResponseWriter, req *http.Request) {
	if r.conversationRepo == nil {
		http.Error(w, "Conversation repository is not available", http.StatusServiceUnavailable)
		return
	}
	switch req.Method {
	case http.MethodGet:
		r.adminConversationsList(w, req)
	case http.MethodPost:
		r.adminConversationCreate(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) adminConversationCourseCode(req *http.Request) string {
	c := strings.TrimSpace(strings.ToLower(req.URL.Query().Get("course_code")))
	if c == "" {
		c = r.defaultCourseCode()
	}
	return c
}

func (r *Router) adminConversationsList(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	courseCode := r.adminConversationCourseCode(req)
	courseID, err := r.conversationRepo.CourseIDByCode(ctx, courseCode)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	scenarios, err := r.conversationRepo.ListScenariosForCourseAdmin(ctx, courseID)
	if err != nil {
		r.logger.Error("admin list conversation scenarios", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	out := make([]adminConversationScenario, 0, len(scenarios))
	for i := range scenarios {
		sc := &scenarios[i]
		tasks, err := r.conversationRepo.ListTasks(ctx, sc.ID)
		if err != nil {
			r.logger.Error("admin list conversation tasks", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		taskOut := make([]adminConversationTask, 0, len(tasks))
		for _, t := range tasks {
			taskOut = append(taskOut, adminConversationTask{
				ID: t.ID, Code: t.Code, Title: t.Title,
				CompletionCriteria: t.CompletionCriteria, IsRequired: t.IsRequired, SortOrder: t.SortOrder,
			})
		}
		out = append(out, adminConversationScenario{
			ID: sc.ID, Code: sc.Code, Title: sc.Title, CEFRLevel: sc.CEFRLevel, PlaceType: sc.PlaceType,
			NPCName: sc.NPCName, NPCPersona: sc.NPCPersona, SceneSetup: sc.SceneSetup, IsQuest: sc.IsQuest,
			MaxTurns: sc.MaxTurns, TokenBudget: sc.TokenBudget,
			NPCCode: sc.NPCCode, PrerequisiteCode: sc.PrerequisiteCode, ImageURL: sc.ImageURL,
			SortOrder: sc.SortOrder, Status: sc.Status,
			Tasks: taskOut,
		})
	}

	levels, err := r.conversationRepo.ListCourseLevels(ctx, courseID)
	if err != nil {
		r.logger.Error("admin list course levels", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	levelOut := make([]map[string]string, 0, len(levels))
	for _, l := range levels {
		levelOut = append(levelOut, map[string]string{"level_code": l.LevelCode, "title": l.Title})
	}

	npcImages, err := r.conversationRepo.GetNPCImages(ctx, courseID)
	if err != nil {
		r.logger.Error("admin get npc images", zap.Error(err))
		// non-fatal — return empty map
		npcImages = map[string]string{}
	}

	writeJSON(w, map[string]interface{}{
		"course_code": courseCode,
		"scenarios":   out,
		"levels":      levelOut,
		"npc_images":  npcImages,
	})
}

// scenarioPayload is the request body for create/update.
type scenarioPayload struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	CEFRLevel   string `json:"cefr_level"`
	PlaceType   string `json:"place_type"`
	NPCName     string `json:"npc_name"`
	NPCPersona  string `json:"npc_persona"`
	SceneSetup  string `json:"scene_setup"`
	IsQuest          bool   `json:"is_quest"`
	MaxTurns         int    `json:"max_turns"`
	TokenBudget      int    `json:"token_budget"`
	NPCCode          string `json:"npc_code"`
	PrerequisiteCode string `json:"prerequisite_code"`
	ImageURL         string `json:"image_url"`
	SortOrder        int    `json:"sort_order"`
	Status           string `json:"status"`
}

func (p *scenarioPayload) toInput(courseCode string) (repository.AdminScenarioInput, error) {
	in := repository.AdminScenarioInput{
		CourseCode:       courseCode,
		CEFRLevel:        strings.TrimSpace(p.CEFRLevel),
		Code:             strings.TrimSpace(p.Code),
		PlaceType:        strings.TrimSpace(p.PlaceType),
		Title:            strings.TrimSpace(p.Title),
		NPCName:          strings.TrimSpace(p.NPCName),
		NPCPersona:       strings.TrimSpace(p.NPCPersona),
		SceneSetup:       strings.TrimSpace(p.SceneSetup),
		IsQuest:          p.IsQuest,
		MaxTurns:         p.MaxTurns,
		TokenBudget:      p.TokenBudget,
		NPCCode:          strings.TrimSpace(p.NPCCode),
		PrerequisiteCode: strings.TrimSpace(p.PrerequisiteCode),
		ImageURL:         strings.TrimSpace(p.ImageURL),
		SortOrder:        p.SortOrder,
		Status:           strings.TrimSpace(p.Status),
	}
	if in.Code == "" || in.Title == "" || in.CEFRLevel == "" {
		return in, errors.New("code, title and cefr_level are required")
	}
	if in.PlaceType == "" {
		in.PlaceType = "cafe"
	}
	if in.NPCName == "" {
		in.NPCName = "NPC"
	}
	if in.MaxTurns <= 0 {
		in.MaxTurns = 30
	}
	if in.TokenBudget <= 0 {
		// Each turn replays the system prompt + recent history, so a turn can cost ~1-1.5k tokens.
		// Keep the budget generous so a normal quest is never cut off mid-scene.
		in.TokenBudget = 40000
	}
	switch in.Status {
	case "draft", "active", "locked", "archived":
	default:
		in.Status = "draft"
	}
	return in, nil
}

func (r *Router) adminConversationCreate(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	courseCode := r.adminConversationCourseCode(req)
	var p scenarioPayload
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	in, err := p.toInput(courseCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := r.conversationRepo.CreateScenario(ctx, in)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateScenarioCode) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "No district/conversation location for this level — check the course level code", http.StatusBadRequest)
			return
		}
		r.logger.Error("admin create scenario", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"id": id, "success": true})
}

// handleAdminConversationScenarioByID handles PUT/DELETE /scenarios/{id} and POST /scenarios/{id}/tasks.
func (r *Router) handleAdminConversationScenarioByID(w http.ResponseWriter, req *http.Request) {
	if r.conversationRepo == nil {
		http.Error(w, "Conversation repository is not available", http.StatusServiceUnavailable)
		return
	}
	path := strings.Trim(strings.TrimPrefix(req.URL.Path, "/api/admin/conversations/scenarios/"), "/")
	parts := strings.Split(path, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid scenario id", http.StatusBadRequest)
		return
	}
	ctx := req.Context()

	switch {
	case len(parts) == 1 && req.Method == http.MethodPut:
		var p scenarioPayload
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
		if err := r.conversationRepo.UpdateScenario(ctx, id, in); err != nil {
			if errors.Is(err, repository.ErrDuplicateScenarioCode) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			r.logger.Error("admin update scenario", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})

	case len(parts) == 1 && req.Method == http.MethodDelete:
		if err := r.conversationRepo.DeleteScenario(ctx, id); err != nil {
			r.logger.Error("admin delete scenario", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})

	case len(parts) == 2 && parts[1] == "tasks" && req.Method == http.MethodPost:
		r.adminConversationTaskCreate(w, req, id)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type taskPayload struct {
	Code               string `json:"code"`
	Title              string `json:"title"`
	CompletionCriteria string `json:"completion_criteria"`
	IsRequired         bool   `json:"is_required"`
	SortOrder          int    `json:"sort_order"`
}

func (p *taskPayload) toInput() (repository.AdminTaskInput, error) {
	in := repository.AdminTaskInput{
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

func (r *Router) adminConversationTaskCreate(w http.ResponseWriter, req *http.Request, scenarioID int64) {
	var p taskPayload
	if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	in, err := p.toInput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := r.conversationRepo.CreateTask(req.Context(), scenarioID, in)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateTaskCode) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		r.logger.Error("admin create task", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"id": id, "success": true})
}

// handleAdminConversationTaskByID handles PUT/DELETE /conversations/tasks/{id}.
func (r *Router) handleAdminConversationTaskByID(w http.ResponseWriter, req *http.Request) {
	if r.conversationRepo == nil {
		http.Error(w, "Conversation repository is not available", http.StatusServiceUnavailable)
		return
	}
	idStr := strings.Trim(strings.TrimPrefix(req.URL.Path, "/api/admin/conversations/tasks/"), "/")
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
		in, perr := p.toInput()
		if perr != nil {
			http.Error(w, perr.Error(), http.StatusBadRequest)
			return
		}
		if err := r.conversationRepo.UpdateTask(ctx, id, in); err != nil {
			if errors.Is(err, repository.ErrDuplicateTaskCode) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			r.logger.Error("admin update task", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	case http.MethodDelete:
		if err := r.conversationRepo.DeleteTask(ctx, id); err != nil {
			r.logger.Error("admin delete task", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdminConversationNPCImage handles PUT /api/admin/conversations/npcs/{npc_code}/image.
func (r *Router) handleAdminConversationNPCImage(w http.ResponseWriter, req *http.Request) {
	if r.conversationRepo == nil {
		http.Error(w, "Conversation repository is not available", http.StatusServiceUnavailable)
		return
	}
	if req.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.Trim(strings.TrimPrefix(req.URL.Path, "/api/admin/conversations/npcs/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "image" || parts[0] == "" {
		http.Error(w, "Invalid path: expected /npcs/{npc_code}/image", http.StatusBadRequest)
		return
	}
	npcCode := parts[0]

	var body struct {
		ImageURL string `json:"image_url"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	courseCode := r.adminConversationCourseCode(req)
	courseID, err := r.conversationRepo.CourseIDByCode(ctx, courseCode)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if err := r.conversationRepo.UpsertNPCImage(ctx, courseID, npcCode, strings.TrimSpace(body.ImageURL)); err != nil {
		r.logger.Error("upsert npc image", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"success": true})
}
