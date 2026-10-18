package web

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/dailyjourney"
	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"
)

// The two chat modes reuse the existing authored content, prompts, checker,
// task progress and ordinary UI. Sessions are explicitly bound to a journey step.
type dailyJourneyChatAdapter struct {
	router *Router
	mode   dailyjourney.Mode
	run    func(context.Context, ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error)
}
type journeyChatAccessKey struct{}
type journeyChatAccess struct {
	user, courseID, uc int64
	course, target     string
	levels             map[string]bool
}
type journeyChatRef struct {
	Code     string `json:"code"`
	District string `json:"district"`
}
type journeyChatContent struct {
	Quest        *repository.PictureQuest         `json:"quest,omitempty"`
	Scenario     *repository.ConversationScenario `json:"scenario,omitempty"`
	Tasks        []repository.ConversationTask    `json:"tasks"`
	District     string                           `json:"district"`
	UserCourseID int64                            `json:"-"`
	Target       string                           `json:"-"`
}
type journeyChatMessage struct {
	Role        string              `json:"role"`
	Content     string              `json:"content"`
	Corrections []ai.ChatCorrection `json:"corrections,omitempty"`
}
type journeyChatState struct {
	SessionID int64                `json:"session_id"`
	Messages  []journeyChatMessage `json:"messages"`
	Completed map[int64]bool       `json:"completed"`
	Turns     int                  `json:"turns"`
	Tokens    int                  `json:"tokens"`
	Passed    bool                 `json:"passed"`
	Done      bool                 `json:"done"`
	Exhausted bool                 `json:"exhausted"`
}
type journeyChatPrepared struct {
	LegacyPrompt string
	Input        ai.ConversationTurnSplitInput
	Opening      bool
	Noop         bool
	Text         string
}
type journeyChatResult struct {
	Turn    ai.ChatTurnResult
	Opening bool
	Noop    bool
	Text    string
}

func (a *dailyJourneyChatAdapter) picture() bool { return a.mode == dailyjourney.Picture }
func (a *dailyJourneyChatAdapter) tables() (string, string) {
	if a.picture() {
		return "picture_quest_sessions", "quest_id"
	}
	return "conversation_sessions", "scenario_id"
}
func (c journeyChatContent) id() int64 {
	if c.Quest != nil {
		return c.Quest.ID
	}
	return c.Scenario.ID
}
func (c journeyChatContent) code() string {
	if c.Quest != nil {
		return c.Quest.Code
	}
	return c.Scenario.Code
}
func (c journeyChatContent) limits() (int, int) {
	if c.Quest != nil {
		return c.Quest.MaxTurns, c.Quest.TokenBudget
	}
	return c.Scenario.MaxTurns, c.Scenario.TokenBudget
}
func chatVersion(c *journeyChatContent) string {
	raw, _ := json.Marshal(c)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func chatPassed(tasks []repository.ConversationTask, done map[int64]bool) bool {
	return allRequiredDone(tasks, done)
}
func (a *dailyJourneyChatAdapter) configured(course string) bool {
	if a.run != nil {
		return true
	}
	s := a.router.aiSvc()
	if s == nil {
		return false
	}
	if a.picture() {
		return s.HasPictureQuestPrompts(course)
	}
	return s.HasSplitConversationPrompts(course) || strings.TrimSpace(s.ConversationPromptForCourse(course)) != ""
}
func (a *dailyJourneyChatAdapter) access(ctx context.Context, q database.ContextDBTX, user int64, course string) (int64, int64, string, map[string]bool, error) {
	if cached, ok := ctx.Value(journeyChatAccessKey{}).(journeyChatAccess); ok && cached.user == user && cached.course == course {
		return cached.courseID, cached.uc, cached.target, cached.levels, nil
	}
	var courseID, uc int64
	var target, tier string
	err := q.QueryRowContext(ctx, `SELECT c.id,uc.id,c.target_lang,COALESCE(u.subscription_tier,'free') FROM user_courses uc JOIN courses c ON c.id=uc.course_id JOIN users u ON u.id=uc.user_id WHERE uc.user_id=? AND c.code=? AND uc.status='active' AND c.status='active'`, user, course).Scan(&courseID, &uc, &target, &tier)
	if err == sql.ErrNoRows {
		return 0, 0, "", nil, dailyjourney.ErrAccess
	}
	if err != nil {
		return 0, 0, "", nil, err
	}
	feature := "conversation"
	if a.picture() {
		feature = "picture_description"
	}
	if !models.TierAllowsFeature(models.ParseUserTier(tier), feature) {
		return 0, 0, "", nil, service.ErrDailyJourneyContentUnavailable
	}
	levels, err := service.DailyJourneyOpenLevels(ctx, a.router.db, q, user, course)
	return courseID, uc, target, levels, err
}
func (a *dailyJourneyChatAdapter) content(ctx context.Context, q database.ContextDBTX, user int64, course string, ref journeyChatRef, started bool) (*journeyChatContent, error) {
	courseID, uc, target, levels, err := a.access(ctx, q, user, course)
	if err != nil {
		return nil, err
	}
	if !a.configured(course) {
		return nil, service.ErrDailyJourneyContentUnavailable
	}
	c := &journeyChatContent{UserCourseID: uc, Target: target}
	var district sql.NullInt64
	var level string
	if a.picture() {
		repo := repository.NewPictureQuestRepository(a.router.db, a.router.logger).WithQueries(q)
		quest, err := repo.GetQuestByCode(ctx, courseID, ref.Code)
		if err != nil {
			return nil, err
		}
		if quest == nil || strings.TrimSpace(quest.ImageURL) == "" || strings.TrimSpace(quest.ImageDescription) == "" {
			return nil, service.ErrDailyJourneyContentUnavailable
		}
		c.Quest = quest
		district = quest.DistrictID
		level = quest.CEFRLevel
		tasks, err := repo.ListTasks(ctx, quest.ID)
		if err != nil {
			return nil, err
		}
		for _, task := range tasks {
			c.Tasks = append(c.Tasks, repository.ConversationTask{ID: task.ID, Code: task.Code, Title: task.Title, IsRequired: task.IsRequired, SortOrder: task.SortOrder, CompletionCriteria: task.CompletionCriteria})
		}
	} else {
		repo := repository.NewConversationRepository(a.router.db, a.router.logger).WithQueries(q)
		scenario, err := repo.GetScenarioByCode(ctx, courseID, ref.Code)
		if err != nil {
			return nil, err
		}
		if scenario == nil || !scenario.IsQuest {
			return nil, service.ErrDailyJourneyContentUnavailable
		}
		c.Scenario = scenario
		district = scenario.DistrictID
		level = scenario.CEFRLevel
		c.Tasks, err = repo.ListTasks(ctx, scenario.ID)
		if err != nil {
			return nil, err
		}
		passed, err := repo.LatestPassedScenarioCodes(ctx, uc, courseID)
		if err != nil {
			return nil, err
		}
		if !started && passed[scenario.Code] {
			return nil, service.ErrDailyJourneyContentUnavailable
		}
		passedAt, err := repo.PassedAtByScenarioCode(ctx, uc, courseID)
		if err != nil {
			return nil, err
		}
		if locked, _ := a.router.scenarioLockState(scenario, passed, passedAt, time.Now()); locked {
			return nil, service.ErrDailyJourneyContentUnavailable
		}
	}
	if !district.Valid || !levels[level] {
		return nil, service.ErrDailyJourneyContentUnavailable
	}
	var districtLevel string
	err = q.QueryRowContext(ctx, `SELECT code,level_code FROM districts WHERE id=? AND course_id=? AND status='active'`, district.Int64, courseID).Scan(&c.District, &districtLevel)
	if err == sql.ErrNoRows {
		return nil, service.ErrDailyJourneyContentUnavailable
	}
	if err != nil {
		return nil, err
	}
	if !levels[districtLevel] || ref.District != "" && ref.District != c.District {
		return nil, service.ErrDailyJourneyContentUnavailable
	}
	required := 0
	for _, t := range c.Tasks {
		if t.IsRequired {
			required++
		}
		if t.Code == "" || strings.TrimSpace(t.CompletionCriteria) == "" {
			return nil, service.ErrDailyJourneyContentUnavailable
		}
	}
	turns, budget := c.limits()
	if required == 0 || turns < 1 || budget < 1 {
		return nil, service.ErrDailyJourneyContentUnavailable
	}
	return c, nil
}
func (a *dailyJourneyChatAdapter) Candidates(ctx context.Context, q database.ContextDBTX, user int64, course string) ([]dailyjourney.Step, error) {
	if !a.configured(course) {
		return nil, nil
	}
	courseID, uc, target, levels, err := a.access(ctx, q, user, course)
	if errors.Is(err, service.ErrDailyJourneyContentUnavailable) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, journeyChatAccessKey{}, journeyChatAccess{user: user, courseID: courseID, uc: uc, course: course, target: target, levels: levels})
	table := "conversation_scenarios"
	if a.picture() {
		table = "picture_quests"
	}
	rows, err := q.QueryContext(ctx, `SELECT code FROM `+table+` WHERE course_id=? AND status='active' ORDER BY sort_order,id`, courseID)
	if err != nil {
		return nil, err
	}
	codes := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			rows.Close()
			return nil, err
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	out, repeat := []dailyjourney.Step{}, []dailyjourney.Step{}
	for _, code := range codes {
		c, err := a.content(ctx, q, user, course, journeyChatRef{Code: code}, false)
		if errors.Is(err, service.ErrDailyJourneyContentUnavailable) {
			continue
		}
		if err != nil {
			return nil, err
		}
		ref, _ := json.Marshal(journeyChatRef{Code: code, District: c.District})
		turns, _ := c.limits()
		route := "PlaceChat"
		params := map[string]string{"districtCode": c.District, "scenarioCode": code}
		title := c.ScenarioTitle()
		if a.picture() {
			route = "PictureQuestChat"
			params = map[string]string{"questCode": code}
			title = c.Quest.Title
		}
		step := dailyjourney.Step{Mode: a.mode, Title: title, Goal: fmt.Sprintf("Закончить попытку: до %d реплик", turns), TargetCount: turns, ContentRef: ref, ContentVersion: chatVersion(c), Launch: &dailyjourney.Launch{RouteName: route, Params: params}}
		if a.picture() {
			passed, err := repository.NewPictureQuestRepository(a.router.db, a.router.logger).WithQueries(q).QuestEverPassed(ctx, uc, code)
			if err != nil {
				return nil, err
			}
			if passed {
				repeat = append(repeat, step)
				continue
			}
		}
		out = append(out, step)
	}
	if len(out) == 0 && a.picture() {
		out = repeat
	}
	return out, nil
}
func (c journeyChatContent) ScenarioTitle() string {
	if c.Scenario != nil {
		return c.Scenario.Title
	}
	return ""
}
func (a *dailyJourneyChatAdapter) bound(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) (*journeyChatContent, *journeyChatState, error) {
	var ref journeyChatRef
	if err := json.Unmarshal(step.ContentRef, &ref); err != nil {
		return nil, nil, err
	}
	state := &journeyChatState{}
	if len(step.State) > 0 {
		if err := json.Unmarshal(step.State, state); err != nil {
			return nil, nil, err
		}
	}
	c, err := a.content(ctx, q, user, course, ref, state.SessionID != 0)
	if err != nil {
		return nil, nil, err
	}
	if chatVersion(c) != step.ContentVersion {
		return nil, nil, service.ErrDailyJourneyContentUnavailable
	}
	if state.Completed == nil {
		state.Completed = map[int64]bool{}
	}
	return c, state, nil
}
func (a *dailyJourneyChatAdapter) Validate(ctx context.Context, q database.ContextDBTX, user int64, course string, step *dailyjourney.Step) error {
	_, _, err := a.bound(ctx, q, user, course, step)
	return err
}
func (a *dailyJourneyChatAdapter) RequiresAI(action dailyjourney.Action) bool {
	return action.Type == "start" || action.Type == "message"
}
func (a *dailyJourneyChatAdapter) Apply(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) error {
	_, state, err := a.bound(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return err
	}
	if action.Type != "ack_feedback" || !state.Done {
		return dailyjourney.ErrInvalidAction
	}
	step.Status = "completed"
	step.ResultDetails, _ = json.Marshal(map[string]interface{}{"attempt_finished": true, "passed": state.Passed, "turns": state.Turns, "budget_exhausted": state.Exhausted})
	return nil
}
func chatActionText(action dailyjourney.Action) (string, error) {
	var payload struct {
		Text string `json:"text"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(action.Payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return "", dailyjourney.ErrInvalidAction
	}
	text := strings.TrimSpace(strings.ReplaceAll(payload.Text, ai.ControlSentinel, ""))
	if text == "" || len(text) > 8000 {
		return "", dailyjourney.ErrInvalidAction
	}
	return text, nil
}
func (a *dailyJourneyChatAdapter) PrepareAI(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action) (json.RawMessage, error) {
	c, state, err := a.bound(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return nil, err
	}
	prepared := journeyChatPrepared{Opening: action.Type == "start"}
	if prepared.Opening && state.SessionID != 0 {
		prepared.Noop = true
		return json.Marshal(prepared)
	}
	if state.Done {
		return nil, dailyjourney.ErrInvalidAction
	}
	if !prepared.Opening && state.SessionID == 0 {
		return nil, dailyjourney.ErrInvalidAction
	}
	svc := a.router.aiSvc()
	var questPrompt, correctionPrompt, npcPrompt string
	if svc != nil {
		correctionPrompt = svc.ConversationCorrectionPromptForCourse(j.CourseCode)
		if a.picture() {
			questPrompt = svc.PictureQuestPromptForCourse(j.CourseCode)
			npcPrompt = svc.PictureLumiPromptForCourse(j.CourseCode)
		} else {
			questPrompt = svc.ConversationQuestPromptForCourse(j.CourseCode)
			npcPrompt = svc.ConversationNPCPromptForCourse(j.CourseCode)
		}
	}
	input := ai.ConversationTurnSplitInput{MaxTokens: 500, EvaluateQuest: !prepared.Opening, EvaluateCorrections: !prepared.Opening}
	turns, _ := c.limits()
	nudge := state.Turns >= turns*85/100
	if a.picture() {
		tasks := []repository.PictureQuestTask{}
		for _, t := range c.Tasks {
			tasks = append(tasks, repository.PictureQuestTask{ID: t.ID, Code: t.Code, Title: t.Title, IsRequired: t.IsRequired, CompletionCriteria: t.CompletionCriteria})
		}
		input.QuestPrompt = buildPictureQuestEvalPrompt(questPrompt, c.Target, c.Quest, tasks, state.Completed)
		input.NPCPrompt = buildLumiReplyPrompt(npcPrompt, c.Target, c.Quest, nudge)
		if prepared.Opening {
			input.NPCPrompt = buildPictureOpeningPrompt(npcPrompt, c.Target, c.Quest)
			input.UserMessage = pictureOpeningTrigger
		}
	} else {
		input.QuestPrompt = buildQuestEvalPrompt(questPrompt, c.Target, c.Scenario, c.Tasks, state.Completed)
		input.NPCPrompt = buildNPCReplyPrompt(npcPrompt, c.Target, c.Scenario, nudge)
		if prepared.Opening {
			input.NPCPrompt = buildOpeningNPCPrompt(npcPrompt, c.Target, c.Scenario)
			input.UserMessage = openingTrigger
		}
	}
	level := ""
	if c.Quest != nil {
		level = c.Quest.CEFRLevel
	} else {
		level = c.Scenario.CEFRLevel
	}
	input.CorrectionPrompt = buildCorrectionPrompt(correctionPrompt, c.Target, level)
	if !prepared.Opening {
		text, err := chatActionText(action)
		if err != nil {
			return nil, err
		}
		prepared.Text = text
		input.UserMessage = text
		start := max(0, len(state.Messages)-conversationHistoryLimit)
		for _, m := range state.Messages[start:] {
			input.History = append(input.History, ai.Message{Role: m.Role, Content: m.Content})
		}
	}
	if !a.picture() && svc != nil && !svc.HasSplitConversationPrompts(j.CourseCode) {
		if prepared.Opening {
			prepared.LegacyPrompt = buildOpeningSystemPrompt(svc.ConversationPromptForCourse(j.CourseCode), c.Target, c.Scenario)
		} else {
			prepared.LegacyPrompt = buildConversationSystemPrompt(svc.ConversationPromptForCourse(j.CourseCode), c.Target, c.Scenario, c.Tasks, state.Completed, nudge)
		}
	}
	prepared.Input = input
	return json.Marshal(prepared)
}
func (a *dailyJourneyChatAdapter) RunAI(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var prepared journeyChatPrepared
	if err := json.Unmarshal(input, &prepared); err != nil {
		return nil, err
	}
	result := journeyChatResult{Opening: prepared.Opening, Text: prepared.Text, Noop: prepared.Noop}
	if prepared.Noop {
		return json.Marshal(result)
	}
	run := a.run
	if run == nil {
		if a.router.aiSvc() == nil {
			return nil, fmt.Errorf("chat provider unavailable")
		}
		run = func(ctx context.Context, in ai.ConversationTurnSplitInput) (*ai.ChatTurnResult, error) {
			return a.router.aiSvc().ConversationTurnSplit(ctx, in)
		}
	}
	var turn *ai.ChatTurnResult
	var err error
	if prepared.LegacyPrompt != "" && a.run == nil {
		turn, err = a.router.aiSvc().ConversationTurn(ctx, prepared.LegacyPrompt, prepared.Input.History, prepared.Input.UserMessage, prepared.Input.MaxTokens)
	} else {
		turn, err = run(ctx, prepared.Input)
	}
	if err != nil {
		return nil, err
	}
	if turn == nil || strings.TrimSpace(turn.VisibleContent) == "" || turn.PromptTokens < 0 || turn.CompletionTokens < 0 {
		return nil, fmt.Errorf("empty or invalid chat result")
	}
	result.Turn = *turn
	result.Turn.Raw = ""
	return json.Marshal(result)
}
func (a *dailyJourneyChatAdapter) ApplyAI(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step, action dailyjourney.Action, raw json.RawMessage) error {
	c, state, err := a.bound(ctx, tx, user, j.CourseCode, step)
	if err != nil {
		return err
	}
	var result journeyChatResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return err
	}
	if result.Opening != (action.Type == "start") {
		return dailyjourney.ErrInvalidAction
	}
	if result.Noop {
		if state.SessionID == 0 {
			return dailyjourney.ErrInvalidAction
		}
		return nil
	}
	if state.Done {
		return dailyjourney.ErrInvalidAction
	}
	table, fk := a.tables()
	pic := repository.NewPictureQuestRepository(a.router.db, a.router.logger).WithQueries(tx)
	npc := repository.NewConversationRepository(a.router.db, a.router.logger).WithQueries(tx)
	if result.Opening {
		if state.SessionID != 0 {
			return dailyjourney.ErrInvalidAction
		}
		if err := tx.QueryRowContext(ctx, `INSERT INTO `+table+`(user_course_id,`+fk+`,daily_journey_step_id) VALUES(?,?,?) RETURNING id`, c.UserCourseID, c.id(), step.ID).Scan(&state.SessionID); err != nil {
			return err
		}
	} else {
		text, err := chatActionText(action)
		if err != nil {
			return err
		}
		if text != result.Text || state.SessionID == 0 {
			return dailyjourney.ErrInvalidAction
		}
		state.Messages = append(state.Messages, journeyChatMessage{Role: "user", Content: text})
		seq := len(state.Messages)
		if a.picture() {
			err = pic.AppendMessage(ctx, state.SessionID, seq, "user", text, 0, 0)
		} else {
			err = npc.AppendMessage(ctx, state.SessionID, seq, "user", text, 0, 0)
		}
		if err != nil {
			return err
		}
		if err := repository.NewLinglowEventRepository(a.router.db).RecordChatMessageTx(ctx, tx, repository.ChatMessageInput{UserID: user, CourseCode: j.CourseCode, MessageLen: len(text), SentAt: time.Now()}); err != nil {
			return err
		}
		result.Turn.Corrections = filterCorrectionsToMessage(result.Turn.Corrections, text)
		state.Turns++
	}
	state.Messages = append(state.Messages, journeyChatMessage{Role: "assistant", Content: result.Turn.VisibleContent, Corrections: result.Turn.Corrections})
	seq := len(state.Messages)
	corrections, _ := json.Marshal(result.Turn.Corrections)
	if a.picture() {
		err = pic.AppendMessageWithCorrections(ctx, state.SessionID, seq, "assistant", result.Turn.VisibleContent, result.Turn.PromptTokens, result.Turn.CompletionTokens, string(corrections))
	} else {
		err = npc.AppendMessageWithCorrections(ctx, state.SessionID, seq, "assistant", result.Turn.VisibleContent, result.Turn.PromptTokens, result.Turn.CompletionTokens, string(corrections))
	}
	if err != nil {
		return err
	}
	tokens := result.Turn.PromptTokens + result.Turn.CompletionTokens
	if tokens == 0 {
		tokens = (len(result.Text) + len(result.Turn.VisibleContent)) / 4
	}
	state.Tokens += tokens
	if !result.Opening {
		taskIDs := map[string]int64{}
		for _, t := range c.Tasks {
			taskIDs[t.Code] = t.ID
		}
		for _, code := range result.Turn.CompletedTaskCodes {
			if id, ok := taskIDs[code]; ok {
				state.Completed[id] = true
			}
		}
		if a.picture() {
			err = pic.MarkTasksCompleted(ctx, state.SessionID, taskIDs, result.Turn.CompletedTaskCodes, seq)
		} else {
			err = npc.MarkTasksCompleted(ctx, state.SessionID, taskIDs, result.Turn.CompletedTaskCodes, seq)
		}
		if err != nil {
			return err
		}
	}
	passed := chatPassed(c.Tasks, state.Completed)
	if passed && !state.Passed {
		if a.picture() {
			err = pic.RecordQuestCompletion(ctx, c.UserCourseID, c.Quest.LearningItemID, c.code(), state.SessionID)
		} else {
			err = npc.RecordQuestCompletion(ctx, c.UserCourseID, c.Scenario.LearningItemID, c.code(), state.SessionID)
		}
		if err != nil {
			return err
		}
	}
	state.Passed = passed
	turns, budget := c.limits()
	state.Exhausted = state.Turns >= turns || state.Tokens >= budget
	state.Done = state.Passed || state.Exhausted
	status := "open"
	if state.Done {
		status = "abandoned"
		if allTasksDone(c.Tasks, state.Completed) {
			status = "completed"
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE `+table+` SET turn_count=?,tokens_used=?,status=?,completed_at=CASE WHEN ?<>'open' THEN CURRENT_TIMESTAMP ELSE completed_at END WHERE id=? AND daily_journey_step_id=?`, state.Turns, state.Tokens, status, status, state.SessionID, step.ID); err != nil {
		return err
	}
	step.CompletedCount = state.Turns
	step.State, err = json.Marshal(state)
	return err
}
func (a *dailyJourneyChatAdapter) Exercise(ctx context.Context, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) (interface{}, error) {
	c, state, err := a.bound(ctx, a.router.db, user, j.CourseCode, step)
	if err != nil {
		return nil, err
	}
	turns, _ := c.limits()
	status := "open"
	if state.Done {
		status = "abandoned"
		if state.Passed {
			status = "completed"
		}
	}
	opening := ""
	if len(state.Messages) > 0 {
		opening = state.Messages[0].Content
	}
	out := map[string]interface{}{"session_id": state.SessionID, "messages": state.Messages, "tasks": tasksJSON(c.Tasks, state.Completed), "turn_count": state.Turns, "max_turns": turns, "status": status, "quest_passed": state.Passed, "budget_exhausted": state.Exhausted, "attempt_finished": state.Done, "opening_line": opening}
	if a.picture() {
		out["quest_code"] = c.Quest.Code
		out["title"] = c.Quest.Title
		out["cefr_level"] = c.Quest.CEFRLevel
		out["image_url"] = c.Quest.ImageURL
	} else {
		sc := c.Scenario
		out["scenario_code"] = sc.Code
		out["title"] = sc.Title
		out["npc_name"] = sc.NPCName
		out["npc_image_url"] = ""
		if images, err := repository.NewConversationRepository(a.router.db, a.router.logger).GetNPCImages(ctx, sc.CourseID); err != nil {
			return nil, err
		} else {
			out["npc_image_url"] = images[sc.NPCCode]
		}
		out["place_type"] = sc.PlaceType
		out["cefr_level"] = sc.CEFRLevel
		out["is_quest"] = true
		out["scene_setup"] = sc.SceneSetup
		out["image_url"] = sc.ImageURL
	}
	return out, nil
}

func (a *dailyJourneyChatAdapter) Skip(ctx context.Context, tx *sql.Tx, user int64, j *dailyjourney.Journey, step *dailyjourney.Step) error {
	var state journeyChatState
	if err := json.Unmarshal(step.State, &state); err != nil {
		return err
	}
	if state.SessionID == 0 {
		return nil
	}
	table, _ := a.tables()
	_, err := tx.ExecContext(ctx, `UPDATE `+table+` SET status='abandoned',completed_at=CURRENT_TIMESTAMP WHERE id=? AND daily_journey_step_id=? AND status='open'`, state.SessionID, step.ID)
	return err
}
