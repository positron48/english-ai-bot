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

// PictureQuest is an authored "describe the picture" quest tied to a district by CEFR level.
// The learner chats with Lumi; ImageDescription is the model-facing ground truth of what is
// on the picture (the model never sees the image bytes).
type PictureQuest struct {
	ID               int64
	CourseID         int64
	DistrictID       sql.NullInt64
	LocationID       sql.NullInt64
	LearningItemID   sql.NullInt64
	Code             string
	CEFRLevel        string
	Title            string
	ImageURL         string
	ImageDescription string
	MaxTurns         int
	TokenBudget      int
}

// PictureQuestTask is one ordered task inside a picture quest.
type PictureQuestTask struct {
	ID                 int64
	QuestID            int64
	Code               string
	SortOrder          int
	IsRequired         bool
	Title              string
	CompletionCriteria string
}

// PictureQuestSession is one live (or finished) run of a picture quest by a user.
type PictureQuestSession struct {
	ID           int64
	UserCourseID int64
	QuestID      int64
	Status       string
	TurnCount    int
	TokensUsed   int
	StartedAt    time.Time
	CompletedAt  sql.NullTime
}

// PictureQuestMessage is a single stored chat message (visible text only).
type PictureQuestMessage struct {
	ID      int64
	Seq     int
	Role    string
	Content string
	// CorrectionsJSON is the raw JSON array of error corrections attached to an assistant
	// message ("[]" when none). The web layer parses it for the client.
	CorrectionsJSON string
	CreatedAt       time.Time
}

// PictureQuestRepository handles picture quests, sessions, messages and task progress.
type PictureQuestRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPictureQuestRepository creates a PictureQuestRepository.
func NewPictureQuestRepository(db *sql.DB, logger *zap.Logger) *PictureQuestRepository {
	return &PictureQuestRepository{db: db, logger: logger}
}

const pictureQuestColumns = `
	id, course_id, district_id, location_id, learning_item_id, code, cefr_level,
	title, image_url, image_description, max_turns, token_budget`

func scanPictureQuest(row interface{ Scan(...interface{}) error }) (*PictureQuest, error) {
	var q PictureQuest
	if err := row.Scan(
		&q.ID, &q.CourseID, &q.DistrictID, &q.LocationID, &q.LearningItemID, &q.Code, &q.CEFRLevel,
		&q.Title, &q.ImageURL, &q.ImageDescription, &q.MaxTurns, &q.TokenBudget,
	); err != nil {
		return nil, err
	}
	return &q, nil
}

// ListQuestsByDistrict returns the active picture quests of a district, ordered by sort_order.
func (r *PictureQuestRepository) ListQuestsByDistrict(ctx context.Context, courseID, districtID int64) ([]PictureQuest, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+pictureQuestColumns+`
		FROM picture_quests
		WHERE course_id = ? AND district_id = ? AND status = 'active'
		ORDER BY sort_order, id`, courseID, districtID)
	if err != nil {
		return nil, fmt.Errorf("list picture quests: %w", err)
	}
	defer rows.Close()

	var out []PictureQuest
	for rows.Next() {
		q, err := scanPictureQuest(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *q)
	}
	return out, rows.Err()
}

// GetQuestByID returns one quest by its primary key (any status).
func (r *PictureQuestRepository) GetQuestByID(ctx context.Context, id int64) (*PictureQuest, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+pictureQuestColumns+`
		FROM picture_quests
		WHERE id = ?`, id)
	q, err := scanPictureQuest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get picture quest by id: %w", err)
	}
	return q, nil
}

// GetQuestByCode returns one active quest by its code within a course.
func (r *PictureQuestRepository) GetQuestByCode(ctx context.Context, courseID int64, code string) (*PictureQuest, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+pictureQuestColumns+`
		FROM picture_quests
		WHERE course_id = ? AND code = ? AND status = 'active'`, courseID, code)
	q, err := scanPictureQuest(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get picture quest: %w", err)
	}
	return q, nil
}

// ListTasks returns the ordered tasks of a quest.
func (r *PictureQuestRepository) ListTasks(ctx context.Context, questID int64) ([]PictureQuestTask, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, quest_id, code, sort_order, is_required, title, completion_criteria
		FROM picture_quest_tasks
		WHERE quest_id = ?
		ORDER BY sort_order, id`, questID)
	if err != nil {
		return nil, fmt.Errorf("list picture quest tasks: %w", err)
	}
	defer rows.Close()

	var out []PictureQuestTask
	for rows.Next() {
		var t PictureQuestTask
		if err := rows.Scan(&t.ID, &t.QuestID, &t.Code, &t.SortOrder, &t.IsRequired, &t.Title, &t.CompletionCriteria); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

const pictureQuestSessionColumns = `id, user_course_id, quest_id, status, turn_count, tokens_used, started_at, completed_at`

func scanPictureQuestSession(row interface{ Scan(...interface{}) error }) (*PictureQuestSession, error) {
	var s PictureQuestSession
	if err := row.Scan(&s.ID, &s.UserCourseID, &s.QuestID, &s.Status, &s.TurnCount, &s.TokensUsed, &s.StartedAt, &s.CompletedAt); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetOpenSession returns the open session for (user_course, quest) if one exists.
func (r *PictureQuestRepository) GetOpenSession(ctx context.Context, userCourseID, questID int64) (*PictureQuestSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+pictureQuestSessionColumns+`
		FROM picture_quest_sessions
		WHERE user_course_id = ? AND quest_id = ? AND status = 'open'`, userCourseID, questID)
	s, err := scanPictureQuestSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get open picture quest session: %w", err)
	}
	return s, nil
}

// GetSession returns a session owned by the given user_course.
func (r *PictureQuestRepository) GetSession(ctx context.Context, sessionID, userCourseID int64) (*PictureQuestSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+pictureQuestSessionColumns+`
		FROM picture_quest_sessions
		WHERE id = ? AND user_course_id = ?`, sessionID, userCourseID)
	s, err := scanPictureQuestSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get picture quest session: %w", err)
	}
	return s, nil
}

// StartSession returns the existing open session for the quest or creates a new one. The second
// return value is true when a fresh session was created.
func (r *PictureQuestRepository) StartSession(ctx context.Context, userCourseID, questID int64) (*PictureQuestSession, bool, error) {
	if existing, err := r.GetOpenSession(ctx, userCourseID, questID); err != nil {
		return nil, false, err
	} else if existing != nil {
		return existing, false, nil
	}

	var id int64
	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO picture_quest_sessions (user_course_id, quest_id, status)
		VALUES (?, ?, 'open')
		RETURNING id`, userCourseID, questID).Scan(&id); err != nil {
		return nil, false, fmt.Errorf("create picture quest session: %w", err)
	}
	s, err := r.GetSession(ctx, id, userCourseID)
	if err != nil {
		return nil, false, err
	}
	return s, true, nil
}

// NextSeq returns the next 1-based message sequence number for a session.
func (r *PictureQuestRepository) NextSeq(ctx context.Context, sessionID int64) (int, error) {
	var maxSeq sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `
		SELECT MAX(seq) FROM picture_quest_messages WHERE session_id = ?`, sessionID).Scan(&maxSeq); err != nil {
		return 0, fmt.Errorf("next seq: %w", err)
	}
	return int(maxSeq.Int64) + 1, nil
}

// AppendMessage stores one message. promptTok/complTok may be 0 when unknown.
func (r *PictureQuestRepository) AppendMessage(ctx context.Context, sessionID int64, seq int, role, content string, promptTok, complTok int) error {
	return r.AppendMessageWithCorrections(ctx, sessionID, seq, role, content, promptTok, complTok, "")
}

// AppendMessageWithCorrections stores a message together with an optional JSON array of error
// corrections (used for assistant replies). An empty correctionsJSON stores "[]".
func (r *PictureQuestRepository) AppendMessageWithCorrections(ctx context.Context, sessionID int64, seq int, role, content string, promptTok, complTok int, correctionsJSON string) error {
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
		INSERT INTO picture_quest_messages (session_id, seq, role, content, prompt_tokens, completion_tokens, corrections)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, sessionID, seq, role, content, pt, ct, correctionsJSON); err != nil {
		return fmt.Errorf("append picture quest message: %w", err)
	}
	return nil
}

// ListMessages returns all visible messages of a session ordered by seq.
func (r *PictureQuestRepository) ListMessages(ctx context.Context, sessionID int64) ([]PictureQuestMessage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, seq, role, content, COALESCE(corrections, '[]'::jsonb), created_at
		FROM picture_quest_messages
		WHERE session_id = ?
		ORDER BY seq`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list picture quest messages: %w", err)
	}
	defer rows.Close()

	var out []PictureQuestMessage
	for rows.Next() {
		var m PictureQuestMessage
		if err := rows.Scan(&m.ID, &m.Seq, &m.Role, &m.Content, &m.CorrectionsJSON, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// BumpSessionCounters increments turn and token counters for a session.
func (r *PictureQuestRepository) BumpSessionCounters(ctx context.Context, sessionID int64, addTurns, addTokens int) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE picture_quest_sessions
		SET turn_count = turn_count + ?, tokens_used = tokens_used + ?
		WHERE id = ?`, addTurns, addTokens, sessionID); err != nil {
		return fmt.Errorf("bump picture quest session counters: %w", err)
	}
	return nil
}

// MarkTasksCompleted records (monotonically) that the given task codes are complete for a session.
// Unknown codes (not in taskIDByCode) are ignored. Already-completed tasks keep their original
// completion timestamp and sequence.
func (r *PictureQuestRepository) MarkTasksCompleted(ctx context.Context, sessionID int64, taskIDByCode map[string]int64, codes []string, seq int) error {
	for _, code := range codes {
		taskID, ok := taskIDByCode[code]
		if !ok {
			continue
		}
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO picture_quest_task_progress (session_id, task_id, completed, completed_at, completed_in_seq)
			VALUES (?, ?, true, CURRENT_TIMESTAMP, ?)
			ON CONFLICT (session_id, task_id) DO UPDATE SET
				completed = true,
				completed_at = COALESCE(picture_quest_task_progress.completed_at, EXCLUDED.completed_at),
				completed_in_seq = COALESCE(picture_quest_task_progress.completed_in_seq, EXCLUDED.completed_in_seq)`,
			sessionID, taskID, seq); err != nil {
			return fmt.Errorf("mark picture quest task complete: %w", err)
		}
	}
	return nil
}

// GetCompletedTaskIDs returns the set of completed task ids for a session.
func (r *PictureQuestRepository) GetCompletedTaskIDs(ctx context.Context, sessionID int64) (map[int64]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT task_id FROM picture_quest_task_progress
		WHERE session_id = ? AND completed = true`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get completed picture quest tasks: %w", err)
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

// RecordQuestCompletion writes the progress signals for a passed picture quest: an idempotent
// exercise_attempts row (so district/location progress aggregates it), a learning_events analytics
// row, and a daily 'chat' stat bump. learningItemID may be NULL.
func (r *PictureQuestRepository) RecordQuestCompletion(ctx context.Context, userCourseID int64, learningItemID sql.NullInt64, questCode string, sessionID int64) error {
	clientAttemptID := fmt.Sprintf("picsess-%d", sessionID)
	var liArg interface{}
	if learningItemID.Valid {
		liArg = learningItemID.Int64
	}
	resultJSON := fmt.Sprintf(`{"picture_quest_code":%q}`, questCode)

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO exercise_attempts
			(user_course_id, learning_item_id, mode, client_attempt_id, answered_at, is_correct, score, result_json)
		VALUES (?, ?, 'chat', ?, CURRENT_TIMESTAMP, true, 100, CAST(? AS jsonb))
		ON CONFLICT (user_course_id, client_attempt_id) WHERE client_attempt_id IS NOT NULL DO NOTHING`,
		userCourseID, liArg, clientAttemptID, resultJSON); err != nil {
		return fmt.Errorf("record picture quest exercise attempt: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO learning_events (user_course_id, event_type, event_time, mode, source_table, event_json)
		VALUES (?, 'picture_quest_completed', CURRENT_TIMESTAMP, 'chat', 'picture_quest_sessions', CAST(? AS jsonb))`,
		userCourseID, resultJSON); err != nil {
		return fmt.Errorf("record picture quest event: %w", err)
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

// QuestEverPassed reports whether mandatory tasks were completed for a quest.
func (r *PictureQuestRepository) QuestEverPassed(ctx context.Context, userCourseID int64, questCode string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM exercise_attempts
			WHERE user_course_id = ? AND mode = 'chat'
				AND result_json->>'picture_quest_code' = ?
		)`, userCourseID, questCode).Scan(&exists)
	return exists, err
}

// PassedQuestCodes returns quest codes where the learner finished mandatory tasks
// (exercise_attempts from RecordQuestCompletion) or fully completed the session.
func (r *PictureQuestRepository) PassedQuestCodes(ctx context.Context, userCourseID, courseID int64) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT q.code
		FROM picture_quests q
		WHERE q.course_id = ?
		AND (
			EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = ?
					AND ea.mode = 'chat'
					AND ea.result_json->>'picture_quest_code' = q.code
			)
			OR EXISTS (
				SELECT 1 FROM picture_quest_sessions sess
				WHERE sess.quest_id = q.id
					AND sess.user_course_id = ?
					AND sess.status = 'completed'
			)
		)`, courseID, userCourseID, userCourseID)
	if err != nil {
		return nil, fmt.Errorf("passed picture quest codes: %w", err)
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

// CompletedQuestCodes returns the set of quest codes the user_course has at least one
// 'completed' session for (all tasks incl. optional — the ★ state), within a course.
func (r *PictureQuestRepository) CompletedQuestCodes(ctx context.Context, userCourseID, courseID int64) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT q.code
		FROM picture_quest_sessions sess
		JOIN picture_quests q ON q.id = sess.quest_id
		WHERE sess.user_course_id = ? AND q.course_id = ? AND sess.status = 'completed'`,
		userCourseID, courseID)
	if err != nil {
		return nil, fmt.Errorf("completed picture quest codes: %w", err)
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
func (r *PictureQuestRepository) CloseSession(ctx context.Context, sessionID, userCourseID int64, status string) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE picture_quest_sessions
		SET status = ?, completed_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_course_id = ? AND status = 'open'`, status, sessionID, userCourseID); err != nil {
		return fmt.Errorf("close picture quest session: %w", err)
	}
	return nil
}
