package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

// conversationHistoryLimit caps how many prior messages are replayed to the model per turn.
const conversationHistoryLimit = 12

// openingTrigger is the (unstored) prompt that makes the NPC greet the learner first.
const openingTrigger = "(The learner has just entered. Greet them in character and start the scene.)"

func (r *Router) aiSvc() *ai.Service {
	if r.aiService == nil {
		return nil
	}
	svc, _ := r.aiService.(*ai.Service)
	return svc
}

// conversationCourseCode resolves the effective course code for a conversation request.
func (r *Router) conversationCourseCode(ctx context.Context, userID int64, explicit string) string {
	if c := strings.TrimSpace(explicit); c != "" {
		return c
	}
	if c := r.currentCourseCodeForUser(ctx, userID); c != "" {
		return c
	}
	return r.defaultCourseCode()
}

// resolveConversationCourse returns the course id, user_course id and target language for a user/course.
func (r *Router) resolveConversationCourse(ctx context.Context, userID int64, courseCode string) (courseID, userCourseID int64, targetLang string, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT c.id, uc.id, c.target_lang
		FROM courses c
		JOIN user_courses uc ON uc.course_id = c.id AND uc.user_id = ?
		WHERE c.code = ?`, userID, courseCode).Scan(&courseID, &userCourseID, &targetLang)
	return
}

// handleLinglowConversationScenarios lists the scenarios available in a district.
// @Summary      Сценарии общения в районе
// @Tags         Linglow
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/linglow/conversation/scenarios [get]
func (r *Router) handleLinglowConversationScenarios(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.conversationRepo == nil {
		http.Error(w, "Conversation repository is not available", http.StatusServiceUnavailable)
		return
	}
	districtCode := strings.TrimSpace(req.URL.Query().Get("district_code"))
	if districtCode == "" {
		http.Error(w, "district_code is required", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	courseCode := r.conversationCourseCode(ctx, userID, req.URL.Query().Get("course_code"))
	courseID, userCourseID, _, err := r.resolveConversationCourse(ctx, userID, courseCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("conversation course resolve", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var districtID int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT id FROM districts WHERE course_id = ? AND code = ?`, courseID, districtCode).Scan(&districtID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "District not found", http.StatusNotFound)
			return
		}
		r.logger.Error("conversation district lookup", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	scenarios, err := r.conversationRepo.ListScenariosForDistrict(ctx, courseID, districtID)
	if err != nil {
		r.logger.Error("list scenarios", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Resolve which scenarios the learner has already completed so prerequisite-gated
	// scenarios (NPC chains) can be marked locked until their predecessor is done.
	completedCodes, err := r.conversationRepo.LatestCompletedScenarioCodes(ctx, userCourseID, courseID)
	if err != nil {
		r.logger.Error("completed scenario codes", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]interface{}, 0, len(scenarios))
	for i := range scenarios {
		sc := &scenarios[i]
		tasks, err := r.conversationRepo.ListTasks(ctx, sc.ID)
		if err != nil {
			r.logger.Error("list tasks", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		locked := sc.PrerequisiteCode != "" && !completedCodes[sc.PrerequisiteCode]
		out = append(out, map[string]interface{}{
			"code":              sc.Code,
			"title":             sc.Title,
			"npc_name":          sc.NPCName,
			"npc_code":          sc.NPCCode,
			"place_type":        sc.PlaceType,
			"cefr_level":        sc.CEFRLevel,
			"is_quest":          sc.IsQuest,
			"prerequisite_code": sc.PrerequisiteCode,
			"locked":            locked,
			"tasks":             tasksJSON(tasks, nil),
			"session_status":    r.latestSessionStatus(ctx, userCourseID, sc.ID),
		})
	}
	writeJSON(w, map[string]interface{}{"scenarios": out})
}

// latestSessionStatus returns the most recent session status for a scenario, or "" if none.
func (r *Router) latestSessionStatus(ctx context.Context, userCourseID, scenarioID int64) string {
	var status string
	err := r.db.QueryRowContext(ctx, `
		SELECT status FROM conversation_sessions
		WHERE user_course_id = ? AND scenario_id = ?
		ORDER BY started_at DESC, id DESC LIMIT 1`, userCourseID, scenarioID).Scan(&status)
	if err != nil {
		return ""
	}
	return status
}

// handleLinglowConversationSessions starts (or resumes) a session for a scenario.
// @Summary      Начать сессию общения
// @Tags         Linglow
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/linglow/conversation/sessions [post]
func (r *Router) handleLinglowConversationSessions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.conversationRepo == nil || r.aiSvc() == nil {
		http.Error(w, "Conversation is not available", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		CourseCode   string `json:"course_code"`
		ScenarioCode string `json:"scenario_code"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.ScenarioCode) == "" {
		http.Error(w, "scenario_code is required", http.StatusBadRequest)
		return
	}
	ctx := req.Context()
	courseCode := r.conversationCourseCode(ctx, userID, body.CourseCode)
	courseID, userCourseID, targetLang, err := r.resolveConversationCourse(ctx, userID, courseCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("conversation course resolve", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	scenario, err := r.conversationRepo.GetScenarioByCode(ctx, courseID, body.ScenarioCode)
	if err != nil {
		r.logger.Error("get scenario", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if scenario == nil {
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	// Enforce prerequisite chains: a gated scenario cannot be started until its predecessor
	// has been completed (mirrors the locked flag in the scenario list).
	if scenario.PrerequisiteCode != "" {
		completedCodes, err := r.conversationRepo.LatestCompletedScenarioCodes(ctx, userCourseID, courseID)
		if err != nil {
			r.logger.Error("completed scenario codes", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if !completedCodes[scenario.PrerequisiteCode] {
			http.Error(w, "Scenario is locked: complete the previous one first", http.StatusForbidden)
			return
		}
	}

	session, created, err := r.conversationRepo.StartSession(ctx, userCourseID, scenario.ID)
	if err != nil {
		r.logger.Error("start session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tasks, err := r.conversationRepo.ListTasks(ctx, scenario.ID)
	if err != nil {
		r.logger.Error("list tasks", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// On a fresh session, generate the NPC opening line and store it as the first message.
	if created {
		systemPrompt := buildConversationSystemPrompt(r.aiSvc().ConversationPromptForCourse(courseCode), targetLang, scenario, tasks, nil, false)
		result, err := r.aiSvc().ConversationTurn(ctx, systemPrompt, nil, openingTrigger, 400)
		if err != nil {
			r.logger.Warn("opening line generation failed", zap.Error(err))
		} else if strings.TrimSpace(result.VisibleContent) != "" {
			if err := r.conversationRepo.AppendMessage(ctx, session.ID, 1, "assistant", result.VisibleContent, result.PromptTokens, result.CompletionTokens); err != nil {
				r.logger.Warn("store opening message failed", zap.Error(err))
			} else {
				_ = r.conversationRepo.BumpSessionCounters(ctx, session.ID, 0, result.PromptTokens+result.CompletionTokens)
			}
		}
	}

	r.writeSessionState(w, ctx, userID, session.ID, userCourseID, scenario, tasks)
}

// handleLinglowConversationSessionByID dispatches GET /{id}, POST /{id}/message, POST /{id}/end.
func (r *Router) handleLinglowConversationSessionByID(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.conversationRepo == nil {
		http.Error(w, "Conversation is not available", http.StatusServiceUnavailable)
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/api/linglow/conversation/sessions/")
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

	switch {
	case len(parts) == 2 && parts[1] == "message" && req.Method == http.MethodPost:
		r.handleConversationMessage(w, req, sessionID, userID)
	case len(parts) == 2 && parts[1] == "end" && req.Method == http.MethodPost:
		r.handleConversationEnd(w, req, sessionID, userID)
	case len(parts) == 1 && req.Method == http.MethodGet:
		r.handleConversationGet(w, req, sessionID, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) handleConversationGet(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	ctx := req.Context()
	session, err := r.conversationRepo.GetSessionForUser(ctx, sessionID, userID)
	if err != nil {
		r.logger.Error("get session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	scenario, err := r.conversationRepo.GetScenarioByID(ctx, session.ScenarioID)
	if err != nil || scenario == nil {
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}
	tasks, err := r.conversationRepo.ListTasks(ctx, scenario.ID)
	if err != nil {
		r.logger.Error("list tasks", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	r.writeSessionState(w, ctx, userID, session.ID, session.UserCourseID, scenario, tasks)
}

func (r *Router) handleConversationEnd(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	ctx := req.Context()
	var body struct {
		Status string `json:"status"`
	}
	_ = json.NewDecoder(req.Body).Decode(&body)
	status := strings.TrimSpace(body.Status)
	if status != "completed" {
		status = "abandoned"
	}
	session, err := r.conversationRepo.GetSessionForUser(ctx, sessionID, userID)
	if err != nil || session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	if err := r.conversationRepo.CloseSession(ctx, session.ID, session.UserCourseID, status); err != nil {
		r.logger.Error("close session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"status": status})
}

func (r *Router) handleConversationMessage(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	if r.aiSvc() == nil {
		http.Error(w, "Conversation is not available", http.StatusServiceUnavailable)
		return
	}
	ctx := req.Context()
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// Strip the control sentinel so a user cannot self-complete tasks via injection.
	userText := strings.TrimSpace(strings.ReplaceAll(body.Text, ai.ControlSentinel, ""))
	if userText == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}

	session, err := r.conversationRepo.GetSessionForUser(ctx, sessionID, userID)
	if err != nil {
		r.logger.Error("get session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	scenario, err := r.conversationRepo.GetScenarioByID(ctx, session.ScenarioID)
	if err != nil || scenario == nil {
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}
	tasks, err := r.conversationRepo.ListTasks(ctx, scenario.ID)
	if err != nil {
		r.logger.Error("list tasks", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session.Status != "open" {
		http.Error(w, "Session is not open", http.StatusConflict)
		return
	}

	// Budget guard: refuse new turns once turn or token caps are hit.
	if session.TurnCount >= scenario.MaxTurns || session.TokensUsed >= scenario.TokenBudget {
		_ = r.conversationRepo.CloseSession(ctx, session.ID, session.UserCourseID, "abandoned")
		completed, _ := r.conversationRepo.GetCompletedTaskIDs(ctx, session.ID)
		writeJSON(w, map[string]interface{}{
			"reply":            "",
			"tasks":            tasksJSON(tasks, completed),
			"turn_count":       session.TurnCount,
			"max_turns":        scenario.MaxTurns,
			"quest_passed":     false,
			"budget_exhausted": true,
			"status":           "abandoned",
		})
		return
	}

	// Persist the user message.
	userSeq, err := r.conversationRepo.NextSeq(ctx, session.ID)
	if err != nil {
		r.logger.Error("next seq", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := r.conversationRepo.AppendMessage(ctx, session.ID, userSeq, "user", userText, 0, 0); err != nil {
		r.logger.Error("store user message", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Daily chat credit on every user message (matches the existing daily 'chat' task semantics).
	courseCode := r.courseCodeByID(ctx, scenario.CourseID)
	if courseCode != "" {
		_ = r.linglowEventRepo.RecordChatMessage(ctx, repository.ChatMessageInput{
			UserID:     userID,
			CourseCode: courseCode,
			MessageLen: len(userText),
			SentAt:     time.Now(),
		})
	}

	// Build history (trimmed) and current completion state.
	history, err := r.conversationHistory(ctx, session.ID)
	if err != nil {
		r.logger.Error("conversation history", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	completedBefore, _ := r.conversationRepo.GetCompletedTaskIDs(ctx, session.ID)
	nudge := scenario.IsQuest && session.TurnCount >= (scenario.MaxTurns*6/10) && !allRequiredDone(tasks, completedBefore)

	targetLang := r.courseTargetLang(ctx, scenario.CourseID)
	systemPrompt := buildConversationSystemPrompt(r.aiSvc().ConversationPromptForCourse(courseCode), targetLang, scenario, tasks, completedBefore, nudge)

	result, err := r.aiSvc().ConversationTurn(ctx, systemPrompt, history, userText, 500)
	if err != nil {
		r.logger.Error("conversation turn", zap.Error(err))
		http.Error(w, "AI provider error", http.StatusBadGateway)
		return
	}

	// Persist the NPC reply together with any error corrections for the learner's last message.
	correctionsJSON := "[]"
	if len(result.Corrections) > 0 {
		if b, mErr := json.Marshal(result.Corrections); mErr == nil {
			correctionsJSON = string(b)
		}
	}
	replySeq := userSeq + 1
	if err := r.conversationRepo.AppendMessageWithCorrections(ctx, session.ID, replySeq, "assistant", result.VisibleContent, result.PromptTokens, result.CompletionTokens, correctionsJSON); err != nil {
		r.logger.Error("store reply", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tokens := result.PromptTokens + result.CompletionTokens
	if tokens == 0 {
		tokens = (len(userText) + len(result.VisibleContent)) / 4
	}
	_ = r.conversationRepo.BumpSessionCounters(ctx, session.ID, 1, tokens)

	// Apply task completions (validated against this scenario's task codes).
	if scenario.IsQuest && len(result.CompletedTaskCodes) > 0 {
		taskIDByCode := make(map[string]int64, len(tasks))
		for _, t := range tasks {
			taskIDByCode[t.Code] = t.ID
		}
		if err := r.conversationRepo.MarkTasksCompleted(ctx, session.ID, taskIDByCode, result.CompletedTaskCodes, replySeq); err != nil {
			r.logger.Warn("mark tasks completed failed", zap.Error(err))
		}
	}

	completedAfter, _ := r.conversationRepo.GetCompletedTaskIDs(ctx, session.ID)
	passedBefore := scenario.IsQuest && allRequiredDone(tasks, completedBefore)
	questPassed := scenario.IsQuest && allRequiredDone(tasks, completedAfter)
	status := "open"
	if questPassed && !passedBefore {
		// Record the win exactly once (the turn the last required task is met). We do NOT close
		// the session here: the learner should still be able to say goodbye and finish optional
		// tasks (the abrupt close was cutting the conversation off mid-scene).
		if err := r.conversationRepo.RecordQuestCompletion(ctx, session.UserCourseID, scenario.LearningItemID, scenario.Code, session.ID); err != nil {
			r.logger.Warn("record quest completion failed", zap.Error(err))
		}
	}
	// Close the session only when the WHOLE checklist (including optional tasks like the farewell)
	// is done, giving the scene a natural ending instead of stopping the moment quests pass.
	if scenario.IsQuest && allTasksDone(tasks, completedAfter) {
		if err := r.conversationRepo.CloseSession(ctx, session.ID, session.UserCourseID, "completed"); err != nil {
			r.logger.Warn("close completed session failed", zap.Error(err))
		}
		status = "completed"
	}

	writeJSON(w, map[string]interface{}{
		"reply":            result.VisibleContent,
		"corrections":      result.Corrections,
		"tasks":            tasksJSON(tasks, completedAfter),
		"turn_count":       session.TurnCount + 1,
		"max_turns":        scenario.MaxTurns,
		"quest_passed":     questPassed,
		"budget_exhausted": false,
		"status":           status,
	})
}

// writeSessionState writes the full session snapshot (messages + tasks) used by start/get.
func (r *Router) writeSessionState(w http.ResponseWriter, ctx context.Context, userID, sessionID, userCourseID int64, scenario *repository.ConversationScenario, tasks []repository.ConversationTask) {
	messages, err := r.conversationRepo.ListMessages(ctx, sessionID)
	if err != nil {
		r.logger.Error("list messages", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	completed, _ := r.conversationRepo.GetCompletedTaskIDs(ctx, sessionID)
	session, _ := r.conversationRepo.GetSession(ctx, sessionID, userCourseID)

	msgs := make([]map[string]interface{}, 0, len(messages))
	openingLine := ""
	for _, m := range messages {
		entry := map[string]interface{}{
			"role":    m.Role,
			"content": m.Content,
		}
		if cj := strings.TrimSpace(m.CorrectionsJSON); cj != "" && cj != "[]" {
			entry["corrections"] = json.RawMessage(cj)
		}
		msgs = append(msgs, entry)
		if openingLine == "" && m.Role == "assistant" {
			openingLine = m.Content
		}
	}

	status := "open"
	turnCount := 0
	if session != nil {
		status = session.Status
		turnCount = session.TurnCount
	}

	writeJSON(w, map[string]interface{}{
		"session_id":    sessionID,
		"scenario_code": scenario.Code,
		"title":         scenario.Title,
		"npc_name":      scenario.NPCName,
		"place_type":    scenario.PlaceType,
		"cefr_level":    scenario.CEFRLevel,
		"is_quest":      scenario.IsQuest,
		"scene_setup":   scenario.SceneSetup,
		"opening_line":  openingLine,
		"messages":      msgs,
		"tasks":         tasksJSON(tasks, completed),
		"turn_count":    turnCount,
		"max_turns":     scenario.MaxTurns,
		"status":        status,
		"quest_passed":  scenario.IsQuest && allRequiredDone(tasks, completed),
	})
}

func (r *Router) conversationHistory(ctx context.Context, sessionID int64) ([]ai.Message, error) {
	messages, err := r.conversationRepo.ListMessages(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if len(messages) > conversationHistoryLimit {
		messages = messages[len(messages)-conversationHistoryLimit:]
	}
	out := make([]ai.Message, 0, len(messages))
	for _, m := range messages {
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		out = append(out, ai.Message{Role: m.Role, Content: m.Content})
	}
	return out, nil
}

func (r *Router) courseCodeByID(ctx context.Context, courseID int64) string {
	var code string
	if err := r.db.QueryRowContext(ctx, `SELECT code FROM courses WHERE id = ?`, courseID).Scan(&code); err != nil {
		return ""
	}
	return code
}

func (r *Router) courseTargetLang(ctx context.Context, courseID int64) string {
	var lang string
	if err := r.db.QueryRowContext(ctx, `SELECT target_lang FROM courses WHERE id = ?`, courseID).Scan(&lang); err != nil {
		return ""
	}
	return lang
}

// tasksJSON renders the checklist; completedByID may be nil (all incomplete).
func tasksJSON(tasks []repository.ConversationTask, completedByID map[int64]bool) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, map[string]interface{}{
			"code":      t.Code,
			"title":     t.Title,
			"required":  t.IsRequired,
			"completed": completedByID[t.ID],
		})
	}
	return out
}

// allTasksDone reports whether every task (required AND optional) is completed.
func allTasksDone(tasks []repository.ConversationTask, completedByID map[int64]bool) bool {
	if len(tasks) == 0 {
		return false
	}
	for _, t := range tasks {
		if !completedByID[t.ID] {
			return false
		}
	}
	return true
}

// allRequiredDone reports whether every required task is completed.
func allRequiredDone(tasks []repository.ConversationTask, completedByID map[int64]bool) bool {
	hasRequired := false
	for _, t := range tasks {
		if !t.IsRequired {
			continue
		}
		hasRequired = true
		if !completedByID[t.ID] {
			return false
		}
	}
	return hasRequired
}

// buildConversationSystemPrompt assembles the per-turn system prompt from the course base prompt
// and the runtime scenario details.
func buildConversationSystemPrompt(base, targetLang string, scenario *repository.ConversationScenario, tasks []repository.ConversationTask, completedByID map[int64]bool, nudge bool) string {
	var b strings.Builder
	if strings.TrimSpace(base) != "" {
		b.WriteString(base)
		b.WriteString("\n\n")
	} else {
		// Built-in fallback when no course prompt file is registered.
		b.WriteString("You are an NPC in a language-learning role-play game. Speak only in the target language, stay in character, keep replies short and simple, and after each reply output a line '###CONTROL###' followed by a JSON object {\"completed_task_codes\":[...],\"all_done\":bool,\"corrections\":[{\"original\":\"...\",\"corrected\":\"...\",\"explanation\":\"...\"}]}. In corrections, list real mistakes in the learner's latest message with a short explanation in Russian; use an empty array when there are none.\n\n")
	}

	fmt.Fprintf(&b, "SCENE\n- You are %s, %s.\n- Setting: %s\n- Target language code: %s. CEFR level: %s.\n",
		scenario.NPCName, scenario.NPCPersona, scenario.SceneSetup, targetLang, scenario.CEFRLevel)

	if scenario.IsQuest && len(tasks) > 0 {
		b.WriteString("\nTASKS the learner should accomplish (do NOT read this list aloud):\n")
		for _, t := range tasks {
			done := completedByID[t.ID]
			req := "required"
			if !t.IsRequired {
				req = "optional"
			}
			status := "not done yet"
			if done {
				status = "ALREADY DONE"
			}
			fmt.Fprintf(&b, "- [%s] (%s, %s) %s\n", t.Code, req, status, t.CompletionCriteria)
		}
		b.WriteString("\nUse the task codes in square brackets in completed_task_codes. Only mark a task when the learner has genuinely done it in their own words. Do not re-mark tasks already done.\n")
		b.WriteString("IMPORTANT: mark each task in completed_task_codes on the SAME turn the learner accomplishes it — as soon as it happens. Never wait until the end of the conversation to report several tasks at once. Each reply must report any tasks newly satisfied by the learner's latest message.\n")
		if nudge {
			b.WriteString("\nThe learner is taking a while. Gently steer them, in very simple language, toward a remaining task without solving it for them.\n")
		}
	} else {
		b.WriteString("\nThis is a free-chat scene with no tasks. Keep an easy, friendly conversation going. Always still output the control line with an empty completed_task_codes array.\n")
	}

	return b.String()
}
