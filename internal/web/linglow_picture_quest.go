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

// pictureOpeningTrigger is the (unstored) prompt that makes Lumi greet the learner first.
const pictureOpeningTrigger = "(The learner has just opened the picture. Greet them and invite them to describe what they see.)"

// handleLinglowPictureQuests lists the picture quests available in a district.
// @Summary      Квесты «опиши картинку» в районе
// @Tags         Linglow
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/linglow/picture-quests [get]
func (r *Router) handleLinglowPictureQuests(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.pictureQuestRepo == nil {
		http.Error(w, "Picture quests are not available", http.StatusServiceUnavailable)
		return
	}
	districtCode := strings.TrimSpace(req.URL.Query().Get("district_code"))
	if districtCode == "" {
		http.Error(w, "district_code is required", http.StatusBadRequest)
		return
	}
	// archive=true returns only passed quests; otherwise only not-yet-passed ones,
	// so the main list never eagerly loads finished cards.
	archive := strings.EqualFold(strings.TrimSpace(req.URL.Query().Get("archive")), "true")
	ctx := req.Context()
	courseCode := r.conversationCourseCode(ctx, userID, req.URL.Query().Get("course_code"))
	courseID, userCourseID, _, err := r.resolveConversationCourse(ctx, userID, courseCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		r.logger.Error("picture quest course resolve", zap.Error(err))
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
		r.logger.Error("picture quest district lookup", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	quests, err := r.pictureQuestRepo.ListQuestsByDistrict(ctx, courseID, districtID)
	if err != nil {
		r.logger.Error("list picture quests", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	out := make([]map[string]interface{}, 0, len(quests))
	for i := range quests {
		q := &quests[i]
		tasks, err := r.pictureQuestRepo.ListTasks(ctx, q.ID)
		if err != nil {
			r.logger.Error("list picture quest tasks", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		questPassed, fullyDone := r.pictureQuestProgressFlags(ctx, userCourseID, q, tasks)
		// Split active list vs archive: passed quests live only in the archive.
		if questPassed != archive {
			continue
		}
		sessionStatus := ""
		switch {
		case fullyDone:
			sessionStatus = "completed"
		case questPassed:
			sessionStatus = "passed"
		default:
			sessionStatus = r.latestPictureSessionStatus(ctx, userCourseID, q.ID)
		}
		out = append(out, map[string]interface{}{
			"code":           q.Code,
			"title":          q.Title,
			"cefr_level":     q.CEFRLevel,
			"image_url":      q.ImageURL,
			"tasks":          pictureTasksJSON(tasks, nil),
			"session_status": sessionStatus,
			"quest_passed":   questPassed,
			"all_tasks_done": fullyDone,
		})
	}
	writeJSON(w, map[string]interface{}{"quests": out})
}

// pictureQuestProgressFlags derives mandatory-pass and 100% completion for a quest.
func (r *Router) pictureQuestProgressFlags(ctx context.Context, userCourseID int64, quest *repository.PictureQuest, tasks []repository.PictureQuestTask) (questPassed, fullyDone bool) {
	var sessionID int64
	var status string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, status FROM picture_quest_sessions
		WHERE user_course_id = ? AND quest_id = ?
		ORDER BY started_at DESC, id DESC LIMIT 1`, userCourseID, quest.ID).Scan(&sessionID, &status)
	if err == nil {
		completed, _ := r.pictureQuestRepo.GetCompletedTaskIDs(ctx, sessionID)
		questPassed = allRequiredPictureTasksDone(tasks, completed)
		fullyDone = status == "completed" || allPictureTasksDone(tasks, completed)
	}
	if !questPassed {
		ever, _ := r.pictureQuestRepo.QuestEverPassed(ctx, userCourseID, quest.Code)
		questPassed = ever
	}
	return questPassed, fullyDone
}

// latestPictureSessionStatus returns the most recent session status for a quest, or "" if none.
func (r *Router) latestPictureSessionStatus(ctx context.Context, userCourseID, questID int64) string {
	var status string
	err := r.db.QueryRowContext(ctx, `
		SELECT status FROM picture_quest_sessions
		WHERE user_course_id = ? AND quest_id = ?
		ORDER BY started_at DESC, id DESC LIMIT 1`, userCourseID, questID).Scan(&status)
	if err != nil {
		return ""
	}
	return status
}

// handleLinglowPictureQuestSessions starts (or resumes) a session for a quest.
// @Summary      Начать сессию «опиши картинку»
// @Tags         Linglow
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/linglow/picture-quests/sessions [post]
func (r *Router) handleLinglowPictureQuestSessions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.pictureQuestRepo == nil || r.aiSvc() == nil {
		http.Error(w, "Picture quests are not available", http.StatusServiceUnavailable)
		return
	}
	var body struct {
		CourseCode string `json:"course_code"`
		QuestCode  string `json:"quest_code"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.QuestCode) == "" {
		http.Error(w, "quest_code is required", http.StatusBadRequest)
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
		r.logger.Error("picture quest course resolve", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	quest, err := r.pictureQuestRepo.GetQuestByCode(ctx, courseID, body.QuestCode)
	if err != nil {
		r.logger.Error("get picture quest", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
		return
	}

	session, created, err := r.pictureQuestRepo.StartSession(ctx, userCourseID, quest.ID)
	if err != nil {
		r.logger.Error("start picture quest session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tasks, err := r.pictureQuestRepo.ListTasks(ctx, quest.ID)
	if err != nil {
		r.logger.Error("list picture quest tasks", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if created {
		r.generatePictureOpeningLine(ctx, courseCode, targetLang, quest, session.ID)
	}

	r.writePictureSessionState(w, ctx, session.ID, userCourseID, quest, tasks)
}

// generatePictureOpeningLine produces Lumi's first greeting and stores it as seq 1. The opening
// prompt deliberately omits the task list so the greeting never hints at the quest objectives.
func (r *Router) generatePictureOpeningLine(ctx context.Context, courseCode, targetLang string, quest *repository.PictureQuest, sessionID int64) {
	svc := r.aiSvc()
	if svc == nil || !svc.HasPictureQuestPrompts(courseCode) {
		return
	}
	systemPrompt := buildPictureOpeningPrompt(svc.PictureLumiPromptForCourse(courseCode), targetLang, quest)
	result, err := svc.ConversationTurnSplit(ctx, ai.ConversationTurnSplitInput{
		NPCPrompt:           systemPrompt,
		UserMessage:         pictureOpeningTrigger,
		MaxTokens:           400,
		EvaluateQuest:       false,
		EvaluateCorrections: false,
	})
	if err != nil {
		r.logger.Warn("picture opening line generation failed", zap.Error(err))
		return
	}
	if strings.TrimSpace(result.VisibleContent) == "" {
		return
	}
	if err := r.pictureQuestRepo.AppendMessage(ctx, sessionID, 1, "assistant", result.VisibleContent, result.PromptTokens, result.CompletionTokens); err != nil {
		r.logger.Warn("store picture opening message failed", zap.Error(err))
		return
	}
	_ = r.pictureQuestRepo.BumpSessionCounters(ctx, sessionID, 0, result.PromptTokens+result.CompletionTokens)
}

// handleLinglowPictureQuestSessionByID dispatches GET /{id}, POST /{id}/message|end|reset.
func (r *Router) handleLinglowPictureQuestSessionByID(w http.ResponseWriter, req *http.Request) {
	userID := getUserIDFromContext(req.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if r.pictureQuestRepo == nil {
		http.Error(w, "Picture quests are not available", http.StatusServiceUnavailable)
		return
	}
	path := strings.TrimPrefix(req.URL.Path, "/api/linglow/picture-quests/sessions/")
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
		r.handlePictureQuestMessage(w, req, sessionID, userID)
	case len(parts) == 2 && parts[1] == "end" && req.Method == http.MethodPost:
		r.handlePictureQuestEnd(w, req, sessionID, userID)
	case len(parts) == 2 && parts[1] == "reset" && req.Method == http.MethodPost:
		r.handlePictureQuestReset(w, req, sessionID, userID)
	case len(parts) == 1 && req.Method == http.MethodGet:
		r.handlePictureQuestGet(w, req, sessionID, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getPictureSessionForUser returns a session owned by the given user (via any of their user_courses).
func (r *Router) getPictureSessionForUser(ctx context.Context, sessionID, userID int64) (*repository.PictureQuestSession, error) {
	var userCourseID int64
	err := r.db.QueryRowContext(ctx, `
		SELECT uc.id
		FROM picture_quest_sessions s
		JOIN user_courses uc ON uc.id = s.user_course_id
		WHERE s.id = ? AND uc.user_id = ?`, sessionID, userID).Scan(&userCourseID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("picture session for user: %w", err)
	}
	return r.pictureQuestRepo.GetSession(ctx, sessionID, userCourseID)
}

func (r *Router) handlePictureQuestGet(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	ctx := req.Context()
	session, err := r.getPictureSessionForUser(ctx, sessionID, userID)
	if err != nil {
		r.logger.Error("get picture quest session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	quest, err := r.pictureQuestRepo.GetQuestByID(ctx, session.QuestID)
	if err != nil || quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
		return
	}
	tasks, err := r.pictureQuestRepo.ListTasks(ctx, quest.ID)
	if err != nil {
		r.logger.Error("list picture quest tasks", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	r.writePictureSessionState(w, ctx, session.ID, session.UserCourseID, quest, tasks)
}

func (r *Router) handlePictureQuestEnd(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	ctx := req.Context()
	var body struct {
		Status string `json:"status"`
	}
	_ = json.NewDecoder(req.Body).Decode(&body)
	status := strings.TrimSpace(body.Status)
	if status != "completed" {
		status = "abandoned"
	}
	session, err := r.getPictureSessionForUser(ctx, sessionID, userID)
	if err != nil || session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	if err := r.pictureQuestRepo.CloseSession(ctx, session.ID, session.UserCourseID, status); err != nil {
		r.logger.Error("close picture quest session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{"status": status})
}

// handlePictureQuestReset abandons the current session and starts a fresh one for the same quest
// (new opening line, empty history), letting the learner restart a description.
func (r *Router) handlePictureQuestReset(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	ctx := req.Context()
	session, err := r.getPictureSessionForUser(ctx, sessionID, userID)
	if err != nil || session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	quest, err := r.pictureQuestRepo.GetQuestByID(ctx, session.QuestID)
	if err != nil || quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
		return
	}
	// Free up the current open session (no-op if it is already closed).
	_ = r.pictureQuestRepo.CloseSession(ctx, session.ID, session.UserCourseID, "abandoned")

	fresh, created, err := r.pictureQuestRepo.StartSession(ctx, session.UserCourseID, quest.ID)
	if err != nil {
		r.logger.Error("reset start picture quest session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tasks, err := r.pictureQuestRepo.ListTasks(ctx, quest.ID)
	if err != nil {
		r.logger.Error("list picture quest tasks", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if created {
		courseCode := r.courseCodeByID(ctx, quest.CourseID)
		targetLang := r.courseTargetLang(ctx, quest.CourseID)
		r.generatePictureOpeningLine(ctx, courseCode, targetLang, quest, fresh.ID)
	}
	r.writePictureSessionState(w, ctx, fresh.ID, session.UserCourseID, quest, tasks)
}

func (r *Router) handlePictureQuestMessage(w http.ResponseWriter, req *http.Request, sessionID, userID int64) {
	if r.aiSvc() == nil {
		http.Error(w, "Picture quests are not available", http.StatusServiceUnavailable)
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

	session, err := r.getPictureSessionForUser(ctx, sessionID, userID)
	if err != nil {
		r.logger.Error("get picture quest session", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	quest, err := r.pictureQuestRepo.GetQuestByID(ctx, session.QuestID)
	if err != nil || quest == nil {
		http.Error(w, "Quest not found", http.StatusNotFound)
		return
	}
	tasks, err := r.pictureQuestRepo.ListTasks(ctx, quest.ID)
	if err != nil {
		r.logger.Error("list picture quest tasks", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if session.Status != "open" {
		http.Error(w, "Session is not open", http.StatusConflict)
		return
	}

	courseCode := r.courseCodeByID(ctx, quest.CourseID)
	svc := r.aiSvc()
	if !svc.HasPictureQuestPrompts(courseCode) {
		http.Error(w, "Picture quests are not configured for this course", http.StatusServiceUnavailable)
		return
	}

	// Budget guard: refuse new turns once turn or token caps are hit.
	if session.TurnCount >= quest.MaxTurns || session.TokensUsed >= quest.TokenBudget {
		_ = r.pictureQuestRepo.CloseSession(ctx, session.ID, session.UserCourseID, "abandoned")
		completed, _ := r.pictureQuestRepo.GetCompletedTaskIDs(ctx, session.ID)
		writeJSON(w, map[string]interface{}{
			"reply":            "",
			"tasks":            pictureTasksJSON(tasks, completed),
			"turn_count":       session.TurnCount,
			"max_turns":        quest.MaxTurns,
			"quest_passed":     false,
			"budget_exhausted": true,
			"status":           "abandoned",
		})
		return
	}

	// Persist the user message.
	userSeq, err := r.pictureQuestRepo.NextSeq(ctx, session.ID)
	if err != nil {
		r.logger.Error("next seq", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := r.pictureQuestRepo.AppendMessage(ctx, session.ID, userSeq, "user", userText, 0, 0); err != nil {
		r.logger.Error("store picture quest user message", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Daily chat credit on every user message (matches the existing daily 'chat' task semantics).
	if courseCode != "" {
		_ = r.linglowEventRepo.RecordChatMessage(ctx, repository.ChatMessageInput{
			UserID:     userID,
			CourseCode: courseCode,
			MessageLen: len(userText),
			SentAt:     time.Now(),
		})
	}

	// Build history (trimmed) and current completion state.
	history, err := r.pictureQuestHistory(ctx, session.ID)
	if err != nil {
		r.logger.Error("picture quest history", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	completedBefore, _ := r.pictureQuestRepo.GetCompletedTaskIDs(ctx, session.ID)
	// Only nudge late in the budget so Lumi doesn't feel pushy early on.
	nudge := session.TurnCount >= (quest.MaxTurns*85/100) && !allRequiredPictureTasksDone(tasks, completedBefore)

	targetLang := r.courseTargetLang(ctx, quest.CourseID)
	result, err := svc.ConversationTurnSplit(ctx, ai.ConversationTurnSplitInput{
		QuestPrompt: buildPictureQuestEvalPrompt(
			svc.PictureQuestPromptForCourse(courseCode),
			targetLang, quest, tasks, completedBefore,
		),
		CorrectionPrompt: buildCorrectionPrompt(
			svc.ConversationCorrectionPromptForCourse(courseCode),
			targetLang, quest.CEFRLevel,
		),
		NPCPrompt: buildLumiReplyPrompt(
			svc.PictureLumiPromptForCourse(courseCode),
			targetLang, quest, nudge,
		),
		History:             history,
		UserMessage:         userText,
		MaxTokens:           500,
		EvaluateQuest:       len(tasks) > 0,
		EvaluateCorrections: true,
	})
	if err != nil {
		r.logger.Error("picture quest turn", zap.Error(err))
		http.Error(w, "AI provider error", http.StatusBadGateway)
		return
	}

	// Guard against the model resurfacing mistakes from earlier turns: keep only corrections whose
	// original fragment actually appears in the learner's latest message.
	result.Corrections = filterCorrectionsToMessage(result.Corrections, userText)

	// Persist Lumi's reply together with any error corrections for the learner's last message.
	correctionsJSON := "[]"
	if len(result.Corrections) > 0 {
		if b, mErr := json.Marshal(result.Corrections); mErr == nil {
			correctionsJSON = string(b)
		}
	}
	replySeq := userSeq + 1
	if err := r.pictureQuestRepo.AppendMessageWithCorrections(ctx, session.ID, replySeq, "assistant", result.VisibleContent, result.PromptTokens, result.CompletionTokens, correctionsJSON); err != nil {
		r.logger.Error("store picture quest reply", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tokens := result.PromptTokens + result.CompletionTokens
	if tokens == 0 {
		tokens = (len(userText) + len(result.VisibleContent)) / 4
	}
	_ = r.pictureQuestRepo.BumpSessionCounters(ctx, session.ID, 1, tokens)

	// Apply task completions (validated against this quest's task codes).
	if len(result.CompletedTaskCodes) > 0 {
		taskIDByCode := make(map[string]int64, len(tasks))
		for _, t := range tasks {
			taskIDByCode[t.Code] = t.ID
		}
		if err := r.pictureQuestRepo.MarkTasksCompleted(ctx, session.ID, taskIDByCode, result.CompletedTaskCodes, replySeq); err != nil {
			r.logger.Warn("mark picture quest tasks completed failed", zap.Error(err))
		}
	}

	completedAfter, _ := r.pictureQuestRepo.GetCompletedTaskIDs(ctx, session.ID)
	passedBefore := allRequiredPictureTasksDone(tasks, completedBefore)
	questPassed := allRequiredPictureTasksDone(tasks, completedAfter)
	status := "open"
	if questPassed && !passedBefore {
		// Record the win exactly once (the turn the last required task is met). The session stays
		// open so the learner can still finish optional tasks for the ★.
		if err := r.pictureQuestRepo.RecordQuestCompletion(ctx, session.UserCourseID, quest.LearningItemID, quest.Code, session.ID); err != nil {
			r.logger.Warn("record picture quest completion failed", zap.Error(err))
		}
	}
	// Close the session only when the WHOLE checklist (including optional tasks) is done — that is
	// the ★ 100% state.
	if allPictureTasksDone(tasks, completedAfter) {
		if err := r.pictureQuestRepo.CloseSession(ctx, session.ID, session.UserCourseID, "completed"); err != nil {
			r.logger.Warn("close completed picture quest session failed", zap.Error(err))
		}
		status = "completed"
	}

	writeJSON(w, map[string]interface{}{
		"reply":            result.VisibleContent,
		"corrections":      result.Corrections,
		"tasks":            pictureTasksJSON(tasks, completedAfter),
		"turn_count":       session.TurnCount + 1,
		"max_turns":        quest.MaxTurns,
		"quest_passed":     questPassed,
		"budget_exhausted": false,
		"status":           status,
	})
}

// writePictureSessionState writes the full session snapshot (messages + tasks) used by start/get.
func (r *Router) writePictureSessionState(w http.ResponseWriter, ctx context.Context, sessionID, userCourseID int64, quest *repository.PictureQuest, tasks []repository.PictureQuestTask) {
	messages, err := r.pictureQuestRepo.ListMessages(ctx, sessionID)
	if err != nil {
		r.logger.Error("list picture quest messages", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	completed, _ := r.pictureQuestRepo.GetCompletedTaskIDs(ctx, sessionID)
	session, _ := r.pictureQuestRepo.GetSession(ctx, sessionID, userCourseID)

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
		"session_id":   sessionID,
		"quest_code":   quest.Code,
		"title":        quest.Title,
		"cefr_level":   quest.CEFRLevel,
		"image_url":    quest.ImageURL,
		"opening_line": openingLine,
		"messages":     msgs,
		"tasks":        pictureTasksJSON(tasks, completed),
		"turn_count":   turnCount,
		"max_turns":    quest.MaxTurns,
		"status":       status,
		"quest_passed": allRequiredPictureTasksDone(tasks, completed),
	})
}

func (r *Router) pictureQuestHistory(ctx context.Context, sessionID int64) ([]ai.Message, error) {
	messages, err := r.pictureQuestRepo.ListMessages(ctx, sessionID)
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

// pictureTasksJSON renders the checklist; completedByID may be nil (all incomplete).
func pictureTasksJSON(tasks []repository.PictureQuestTask, completedByID map[int64]bool) []map[string]interface{} {
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

// allPictureTasksDone reports whether every task (required AND optional) is completed.
func allPictureTasksDone(tasks []repository.PictureQuestTask, completedByID map[int64]bool) bool {
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

// allRequiredPictureTasksDone reports whether every required task is completed.
func allRequiredPictureTasksDone(tasks []repository.PictureQuestTask, completedByID map[int64]bool) bool {
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

// buildPictureQuestEvalPrompt assembles the picture-quest task evaluation prompt. The admin-written
// image description is included as the ground truth the learner's statements are judged against.
func buildPictureQuestEvalPrompt(base, targetLang string, quest *repository.PictureQuest, tasks []repository.PictureQuestTask, completedByID map[int64]bool) string {
	var b strings.Builder
	if strings.TrimSpace(base) != "" {
		b.WriteString(base)
		b.WriteString("\n\n")
	} else {
		b.WriteString("Evaluate picture-description task completion from the learner's latest message. Output JSON only: {\"completed_task_codes\":[...],\"all_done\":bool}.\n\n")
	}
	fmt.Fprintf(&b, "CONTEXT\n- Target language code: %s. CEFR level: %s.\n", targetLang, quest.CEFRLevel)
	fmt.Fprintf(&b, "\nPICTURE (ground truth — the learner sees the image, not this text):\n%s\n", strings.TrimSpace(quest.ImageDescription))
	if len(tasks) > 0 {
		b.WriteString("\nTASKS (evaluate the learner's latest message against these):\n")
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
		b.WriteString("\nUse the task codes in square brackets in completed_task_codes. Re-check each 'not done yet' task independently; one message can complete several at once.\n")
	}
	return b.String()
}

// buildLumiReplyPrompt assembles Lumi's reply prompt (picture description included, no task list).
func buildLumiReplyPrompt(base, targetLang string, quest *repository.PictureQuest, nudge bool) string {
	var b strings.Builder
	if strings.TrimSpace(base) != "" {
		b.WriteString(base)
		b.WriteString("\n\n")
	} else {
		b.WriteString("You are Lumi, the friendly mascot of a language-learning app, looking at a picture together with the learner. Speak only in the target language, keep replies short and simple, never coach the learner or reveal hidden tasks. Reply with your line(s) only — no JSON.\n\n")
	}
	fmt.Fprintf(&b, "CONTEXT\n- Target language code: %s. CEFR level: %s.\n", targetLang, quest.CEFRLevel)
	fmt.Fprintf(&b, "\nPICTURE (what you both see; never reveal this text itself):\n%s\n", strings.TrimSpace(quest.ImageDescription))
	if nudge {
		b.WriteString("\nThe conversation has stalled. You may ask one very simple, open question about the picture to keep it going — but do not tell the learner what to say.\n")
	}
	return b.String()
}

// buildPictureOpeningPrompt assembles the prompt for Lumi's first greeting. It omits the task list
// entirely so the opening line cannot reveal or hint at quest objectives.
func buildPictureOpeningPrompt(base, targetLang string, quest *repository.PictureQuest) string {
	var b strings.Builder
	if strings.TrimSpace(base) != "" {
		b.WriteString(base)
		b.WriteString("\n\n")
	}
	fmt.Fprintf(&b, "CONTEXT\n- Target language code: %s. CEFR level: %s.\n", targetLang, quest.CEFRLevel)
	fmt.Fprintf(&b, "\nPICTURE (what you both see; never reveal this text itself):\n%s\n", strings.TrimSpace(quest.ImageDescription))
	b.WriteString("\nGreet the learner warmly with ONE short, simple opening line and invite them to describe the picture. " +
		"Do NOT mention or hint at any task, checklist or goal beyond describing the picture, and do NOT describe the picture yourself yet.\n")
	return b.String()
}
