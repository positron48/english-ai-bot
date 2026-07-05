package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ConversationScenario is an authored NPC role-play scenario tied to a district/location.
type ConversationScenario struct {
	ID             int64
	CourseID       int64
	DistrictID     sql.NullInt64
	LocationID     sql.NullInt64
	LearningItemID sql.NullInt64
	Code           string
	PlaceType      string
	CEFRLevel      string
	Title          string
	NPCName        string
	NPCPersona     string
	SceneSetup     string
	IsQuest        bool
	MaxTurns       int
	TokenBudget    int
	// NPCCode groups several scenarios under one recurring NPC ("" = standalone).
	NPCCode string
	// PrerequisiteCode is the scenario code that must be completed before this one
	// unlocks for a learner ("" = always available).
	PrerequisiteCode string
	// ImageURL is an optional banner image shown at the start of the quest dialog.
	ImageURL string
}

// ConversationTask is one ordered task inside a quest scenario.
type ConversationTask struct {
	ID                 int64
	ScenarioID         int64
	Code               string
	SortOrder          int
	IsRequired         bool
	Title              string
	CompletionCriteria string
}

// ConversationSession is one live (or finished) run of a scenario by a user.
type ConversationSession struct {
	ID           int64
	UserCourseID int64
	ScenarioID   int64
	Status       string
	TurnCount    int
	TokensUsed   int
	StartedAt    time.Time
	CompletedAt  sql.NullTime
}

// ConversationMessage is a single stored chat message (visible text only).
type ConversationMessage struct {
	ID        int64
	Seq       int
	Role      string
	Content   string
	// CorrectionsJSON is the raw JSON array of error corrections attached to an assistant
	// message ("[]" when none). The web layer parses it for the client.
	CorrectionsJSON string
	CreatedAt       time.Time
}

// ConversationRepository handles conversation scenarios, sessions, messages and task progress.
type ConversationRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewConversationRepository creates a ConversationRepository.
func NewConversationRepository(db *sql.DB, logger *zap.Logger) *ConversationRepository {
	return &ConversationRepository{db: db, logger: logger}
}

const conversationScenarioColumns = `
	id, course_id, district_id, location_id, learning_item_id, code, place_type, cefr_level,
	title, npc_name, npc_persona, scene_setup, is_quest, max_turns, token_budget, npc_code, prerequisite_code, image_url`

func scanScenario(row interface{ Scan(...interface{}) error }) (*ConversationScenario, error) {
	var s ConversationScenario
	if err := row.Scan(
		&s.ID, &s.CourseID, &s.DistrictID, &s.LocationID, &s.LearningItemID, &s.Code, &s.PlaceType, &s.CEFRLevel,
		&s.Title, &s.NPCName, &s.NPCPersona, &s.SceneSetup, &s.IsQuest, &s.MaxTurns, &s.TokenBudget,
		&s.NPCCode, &s.PrerequisiteCode, &s.ImageURL,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

// ListScenariosForDistrict returns the active scenarios available in a district (its conversation
// location), ordered by sort_order.
func (r *ConversationRepository) ListScenariosForDistrict(ctx context.Context, courseID, districtID int64) ([]ConversationScenario, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+conversationScenarioColumns+`
		FROM conversation_scenarios
		WHERE course_id = ? AND district_id = ? AND status = 'active'
		ORDER BY sort_order, id`, courseID, districtID)
	if err != nil {
		return nil, fmt.Errorf("list scenarios: %w", err)
	}
	defer rows.Close()

	var out []ConversationScenario
	for rows.Next() {
		s, err := scanScenario(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	return out, rows.Err()
}

// GetScenarioByID returns one scenario by its primary key (any status).
func (r *ConversationRepository) GetScenarioByID(ctx context.Context, id int64) (*ConversationScenario, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+conversationScenarioColumns+`
		FROM conversation_scenarios
		WHERE id = ?`, id)
	s, err := scanScenario(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get scenario by id: %w", err)
	}
	return s, nil
}

// GetScenarioByCode returns one active scenario by its code within a course.
func (r *ConversationRepository) GetScenarioByCode(ctx context.Context, courseID int64, code string) (*ConversationScenario, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+conversationScenarioColumns+`
		FROM conversation_scenarios
		WHERE course_id = ? AND code = ? AND status = 'active'`, courseID, code)
	s, err := scanScenario(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get scenario: %w", err)
	}
	return s, nil
}

// ListTasks returns the ordered tasks of a scenario.
func (r *ConversationRepository) ListTasks(ctx context.Context, scenarioID int64) ([]ConversationTask, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, scenario_id, code, sort_order, is_required, title, completion_criteria
		FROM conversation_tasks
		WHERE scenario_id = ?
		ORDER BY sort_order, id`, scenarioID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var out []ConversationTask
	for rows.Next() {
		var t ConversationTask
		if err := rows.Scan(&t.ID, &t.ScenarioID, &t.Code, &t.SortOrder, &t.IsRequired, &t.Title, &t.CompletionCriteria); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

const conversationSessionColumns = `id, user_course_id, scenario_id, status, turn_count, tokens_used, started_at, completed_at`

func scanSession(row interface{ Scan(...interface{}) error }) (*ConversationSession, error) {
	var s ConversationSession
	if err := row.Scan(&s.ID, &s.UserCourseID, &s.ScenarioID, &s.Status, &s.TurnCount, &s.TokensUsed, &s.StartedAt, &s.CompletedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetOpenSession returns the open session for (user_course, scenario) if one exists.
func (r *ConversationRepository) GetOpenSession(ctx context.Context, userCourseID, scenarioID int64) (*ConversationSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+conversationSessionColumns+`
		FROM conversation_sessions
		WHERE user_course_id = ? AND scenario_id = ? AND status = 'open'`, userCourseID, scenarioID)
	s, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get open session: %w", err)
	}
	return s, nil
}

// GetSession returns a session owned by the given user_course.
func (r *ConversationRepository) GetSession(ctx context.Context, sessionID, userCourseID int64) (*ConversationSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+conversationSessionColumns+`
		FROM conversation_sessions
		WHERE id = ? AND user_course_id = ?`, sessionID, userCourseID)
	s, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return s, nil
}

// GetSessionForUser returns a session owned by the given user (via any of their user_courses).
func (r *ConversationRepository) GetSessionForUser(ctx context.Context, sessionID, userID int64) (*ConversationSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT cs.id, cs.user_course_id, cs.scenario_id, cs.status, cs.turn_count, cs.tokens_used, cs.started_at, cs.completed_at
		FROM conversation_sessions cs
		JOIN user_courses uc ON uc.id = cs.user_course_id
		WHERE cs.id = ? AND uc.user_id = ?`, sessionID, userID)
	s, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session for user: %w", err)
	}
	return s, nil
}

// StartSession returns the existing open session for the scenario or creates a new one. The second
// return value is true when a fresh session was created.
func (r *ConversationRepository) StartSession(ctx context.Context, userCourseID, scenarioID int64) (*ConversationSession, bool, error) {
	if existing, err := r.GetOpenSession(ctx, userCourseID, scenarioID); err != nil {
		return nil, false, err
	} else if existing != nil {
		return existing, false, nil
	}

	var id int64
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO conversation_sessions (user_course_id, scenario_id, status)
		VALUES (?, ?, 'open')
		RETURNING id`, userCourseID, scenarioID).Scan(&id); err != nil {
		return nil, false, fmt.Errorf("create session: %w", err)
	}
	s, err := r.GetSession(ctx, id, userCourseID)
	if err != nil {
		return nil, false, err
	}
	return s, true, nil
}

// NextSeq returns the next 1-based message sequence number for a session.
func (r *ConversationRepository) NextSeq(ctx context.Context, sessionID int64) (int, error) {
	var maxSeq sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `
		SELECT MAX(seq) FROM conversation_messages WHERE session_id = ?`, sessionID).Scan(&maxSeq); err != nil {
		return 0, fmt.Errorf("next seq: %w", err)
	}
	return int(maxSeq.Int64) + 1, nil
}

// AppendMessage stores one message. promptTok/complTok may be 0 when unknown.
func (r *ConversationRepository) AppendMessage(ctx context.Context, sessionID int64, seq int, role, content string, promptTok, complTok int) error {
	return r.AppendMessageWithCorrections(ctx, sessionID, seq, role, content, promptTok, complTok, "")
}

// AppendMessageWithCorrections stores a message together with an optional JSON array of error
// corrections (used for assistant replies). An empty correctionsJSON stores "[]".
func (r *ConversationRepository) AppendMessageWithCorrections(ctx context.Context, sessionID int64, seq int, role, content string, promptTok, complTok int, correctionsJSON string) error {
	var pt, ct interface{}
	if promptTok > 0 {
		pt = promptTok
	}
	if complTok > 0 {
		ct = complTok
	}
	if strings.TrimSpace(correctionsJSON) == "" {
		correctionsJSON = "[]"
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO conversation_messages (session_id, seq, role, content, prompt_tokens, completion_tokens, corrections)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, sessionID, seq, role, content, pt, ct, correctionsJSON); err != nil {
		return fmt.Errorf("append message: %w", err)
	}
	return nil
}

// ListMessages returns all visible messages of a session ordered by seq.
func (r *ConversationRepository) ListMessages(ctx context.Context, sessionID int64) ([]ConversationMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, seq, role, content, COALESCE(corrections, '[]'::jsonb), created_at
		FROM conversation_messages
		WHERE session_id = ?
		ORDER BY seq`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var out []ConversationMessage
	for rows.Next() {
		var m ConversationMessage
		if err := rows.Scan(&m.ID, &m.Seq, &m.Role, &m.Content, &m.CorrectionsJSON, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// BumpSessionCounters increments turn and token counters for a session.
func (r *ConversationRepository) BumpSessionCounters(ctx context.Context, sessionID int64, addTurns, addTokens int) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE conversation_sessions
		SET turn_count = turn_count + ?, tokens_used = tokens_used + ?
		WHERE id = ?`, addTurns, addTokens, sessionID); err != nil {
		return fmt.Errorf("bump session counters: %w", err)
	}
	return nil
}

// MarkTasksCompleted records (monotonically) that the given task codes are complete for a session.
// Unknown codes (not in taskIDByCode) are ignored. Already-completed tasks keep their original
// completion timestamp and sequence.
func (r *ConversationRepository) MarkTasksCompleted(ctx context.Context, sessionID int64, taskIDByCode map[string]int64, codes []string, seq int) error {
	for _, code := range codes {
		taskID, ok := taskIDByCode[code]
		if !ok {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO conversation_task_progress (session_id, task_id, completed, completed_at, completed_in_seq)
			VALUES (?, ?, true, CURRENT_TIMESTAMP, ?)
			ON CONFLICT (session_id, task_id) DO UPDATE SET
				completed = true,
				completed_at = COALESCE(conversation_task_progress.completed_at, EXCLUDED.completed_at),
				completed_in_seq = COALESCE(conversation_task_progress.completed_in_seq, EXCLUDED.completed_in_seq)`,
			sessionID, taskID, seq); err != nil {
			return fmt.Errorf("mark task complete: %w", err)
		}
	}
	return nil
}

// GetCompletedTaskIDs returns the set of completed task ids for a session.
func (r *ConversationRepository) GetCompletedTaskIDs(ctx context.Context, sessionID int64) (map[int64]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT task_id FROM conversation_task_progress
		WHERE session_id = ? AND completed = true`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get completed tasks: %w", err)
	}
	defer rows.Close()

	out := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// RecordQuestCompletion writes the progress signals for a passed quest: an idempotent
// exercise_attempts row (so district/location progress aggregates it), a learning_events analytics
// row, and a daily 'chat' stat bump. learningItemID may be NULL.
func (r *ConversationRepository) RecordQuestCompletion(ctx context.Context, userCourseID int64, learningItemID sql.NullInt64, scenarioCode string, sessionID int64) error {
	clientAttemptID := fmt.Sprintf("convsess-%d", sessionID)
	var liArg interface{}
	if learningItemID.Valid {
		liArg = learningItemID.Int64
	}
	resultJSON := fmt.Sprintf(`{"scenario_code":%q}`, scenarioCode)

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO exercise_attempts
			(user_course_id, learning_item_id, mode, client_attempt_id, answered_at, is_correct, score, result_json)
		VALUES (?, ?, 'chat', ?, CURRENT_TIMESTAMP, true, 100, CAST(? AS jsonb))
		ON CONFLICT (user_course_id, client_attempt_id) WHERE client_attempt_id IS NOT NULL DO NOTHING`,
		userCourseID, liArg, clientAttemptID, resultJSON); err != nil {
		return fmt.Errorf("record quest exercise attempt: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO learning_events (user_course_id, event_type, event_time, mode, source_table, event_json)
		VALUES (?, 'conversation_quest_completed', CURRENT_TIMESTAMP, 'chat', 'conversation_sessions', CAST(? AS jsonb))`,
		userCourseID, resultJSON); err != nil {
		return fmt.Errorf("record quest event: %w", err)
	}

	_ = NewLinglowDailyStatsRepository(r.db).Bump(ctx, DailyBump{
		UserCourseID: userCourseID,
		Day:          LocalDayFromTime(time.Now()),
		Mode:         "chat",
		Attempts:     1,
		Correct:      1,
	})
	return nil
}

// LatestPassedScenarioCodes returns scenario codes where the learner finished mandatory tasks
// (exercise_attempts from RecordQuestCompletion) or fully completed the session.
func (r *ConversationRepository) LatestPassedScenarioCodes(ctx context.Context, userCourseID, courseID int64) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT sc.code
		FROM conversation_scenarios sc
		WHERE sc.course_id = ?
		AND (
			EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = ?
					AND ea.mode = 'chat'
					AND ea.result_json->>'scenario_code' = sc.code
			)
			OR EXISTS (
				SELECT 1 FROM conversation_sessions sess
				WHERE sess.scenario_id = sc.id
					AND sess.user_course_id = ?
					AND sess.status = 'completed'
			)
		)`, courseID, userCourseID, userCourseID)
	if err != nil {
		return nil, fmt.Errorf("passed scenario codes: %w", err)
	}
	defer rows.Close()
	out := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out[code] = true
	}
	return out, rows.Err()
}

// PassedAtByScenarioCode returns the first pass/completion time per scenario code.
// Cooldown windows start when mandatory tasks are done, not only after optional farewell.
func (r *ConversationRepository) PassedAtByScenarioCode(ctx context.Context, userCourseID, courseID int64) (map[string]time.Time, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, MIN(ts) FROM (
			SELECT sc.code AS code, sess.completed_at AS ts
			FROM conversation_sessions sess
			JOIN conversation_scenarios sc ON sc.id = sess.scenario_id
			WHERE sess.user_course_id = ? AND sc.course_id = ? AND sess.status = 'completed'
			UNION ALL
			SELECT ea.result_json->>'scenario_code' AS code, ea.answered_at AS ts
			FROM exercise_attempts ea
			WHERE ea.user_course_id = ? AND ea.mode = 'chat'
				AND ea.result_json->>'scenario_code' IS NOT NULL
		) sub
		WHERE code IS NOT NULL AND code <> ''
		GROUP BY code`, userCourseID, courseID, userCourseID)
	if err != nil {
		return nil, fmt.Errorf("passed at by scenario code: %w", err)
	}
	defer rows.Close()
	out := make(map[string]time.Time)
	for rows.Next() {
		var code string
		var t time.Time
		if err := rows.Scan(&code, &t); err != nil {
			return nil, err
		}
		out[code] = t
	}
	return out, rows.Err()
}

// ScenarioEverPassed reports whether mandatory tasks were completed for a scenario.
func (r *ConversationRepository) ScenarioEverPassed(ctx context.Context, userCourseID int64, scenarioCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM exercise_attempts
			WHERE user_course_id = ? AND mode = 'chat'
				AND result_json->>'scenario_code' = ?
		)`, userCourseID, scenarioCode).Scan(&exists)
	return exists, err
}

// LatestCompletedScenarioCodes returns the set of scenario codes the user_course has at least one
// 'completed' session for, within a course. Used to resolve prerequisite/unlock state.
func (r *ConversationRepository) LatestCompletedScenarioCodes(ctx context.Context, userCourseID, courseID int64) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT sc.code
		FROM conversation_sessions sess
		JOIN conversation_scenarios sc ON sc.id = sess.scenario_id
		WHERE sess.user_course_id = ? AND sc.course_id = ? AND sess.status = 'completed'`,
		userCourseID, courseID)
	if err != nil {
		return nil, fmt.Errorf("completed scenario codes: %w", err)
	}
	defer rows.Close()
	out := make(map[string]bool)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out[code] = true
	}
	return out, rows.Err()
}

// CloseSession sets a terminal status ('completed' or 'abandoned') on a session.
func (r *ConversationRepository) CloseSession(ctx context.Context, sessionID, userCourseID int64, status string) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE conversation_sessions
		SET status = ?, completed_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_course_id = ? AND status = 'open'`, status, sessionID, userCourseID); err != nil {
		return fmt.Errorf("close session: %w", err)
	}
	return nil
}

// GetNPCImages returns a map of npc_code → image_url for NPCs that have an image set in the course.
func (r *ConversationRepository) GetNPCImages(ctx context.Context, courseID int64) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT npc_code, image_url FROM conversation_npcs
		WHERE course_id = ? AND image_url <> ''`, courseID)
	if err != nil {
		return nil, fmt.Errorf("get npc images: %w", err)
	}
	defer rows.Close()
	out := make(map[string]string)
	for rows.Next() {
		var code, url string
		if err := rows.Scan(&code, &url); err != nil {
			return nil, err
		}
		out[code] = url
	}
	return out, rows.Err()
}

// CompletedAtByScenarioCode returns a map of scenario_code → latest completed_at for all
// scenarios this user_course has finished. Used to compute cooldown windows.
func (r *ConversationRepository) CompletedAtByScenarioCode(ctx context.Context, userCourseID, courseID int64) (map[string]time.Time, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT sc.code, MAX(sess.completed_at)
		FROM conversation_sessions sess
		JOIN conversation_scenarios sc ON sc.id = sess.scenario_id
		WHERE sess.user_course_id = ? AND sc.course_id = ? AND sess.status = 'completed'
		GROUP BY sc.code`, userCourseID, courseID)
	if err != nil {
		return nil, fmt.Errorf("completed at by scenario code: %w", err)
	}
	defer rows.Close()
	out := make(map[string]time.Time)
	for rows.Next() {
		var code string
		var t time.Time
		if err := rows.Scan(&code, &t); err != nil {
			return nil, err
		}
		out[code] = t
	}
	return out, rows.Err()
}

// UpsertNPCImage sets (or clears, when imageURL == "") the avatar image for an NPC in a course.
func (r *ConversationRepository) UpsertNPCImage(ctx context.Context, courseID int64, npcCode, imageURL string) error {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO conversation_npcs (course_id, npc_code, image_url, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (course_id, npc_code) DO UPDATE SET
			image_url = EXCLUDED.image_url,
			updated_at = CURRENT_TIMESTAMP`,
		courseID, npcCode, imageURL); err != nil {
		return fmt.Errorf("upsert npc image: %w", err)
	}
	return nil
}
