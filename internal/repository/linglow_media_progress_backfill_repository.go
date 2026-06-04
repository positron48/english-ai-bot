package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/config"
)

type LinglowMediaProgressBackfillRepository struct {
	db *sql.DB
}

type LinglowMediaProgressBackfillOptions struct {
	Commit bool
	Source string
	Limit  int
}

type LinglowMediaProgressBackfillSummary struct {
	Source        string
	LegacyTotal   int64
	MirroredTotal int64
	Missing       int64
	Processed     int64
	Inserted      int64
	UnmappedTotal int64
}

func NewLinglowMediaProgressBackfillRepository(db *sql.DB) *LinglowMediaProgressBackfillRepository {
	return &LinglowMediaProgressBackfillRepository{db: db}
}

func (r *LinglowMediaProgressBackfillRepository) Backfill(ctx context.Context, lc config.LearningConfig, opts LinglowMediaProgressBackfillOptions) ([]LinglowMediaProgressBackfillSummary, error) {
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	sources, err := normalizeMediaSources(opts.Source)
	if err != nil {
		return nil, err
	}
	out := make([]LinglowMediaProgressBackfillSummary, 0, len(sources))
	for _, source := range sources {
		summary, err := r.audit(ctx, courseCode, source)
		if err != nil {
			return nil, err
		}
		if opts.Commit {
			if source == "reading" {
				if err := r.backfillReading(ctx, courseCode, opts.Limit, &summary); err != nil {
					return nil, err
				}
			} else {
				if err := r.backfillSpeaking(ctx, courseCode, opts.Limit, &summary); err != nil {
					return nil, err
				}
			}
			refreshed, err := r.audit(ctx, courseCode, source)
			if err != nil {
				return nil, err
			}
			summary.LegacyTotal = refreshed.LegacyTotal
			summary.MirroredTotal = refreshed.MirroredTotal
			summary.Missing = refreshed.Missing
			summary.UnmappedTotal = refreshed.UnmappedTotal
		}
		out = append(out, summary)
	}
	return out, nil
}

func normalizeMediaSources(source string) ([]string, error) {
	switch strings.TrimSpace(source) {
	case "", "all":
		return []string{"reading", "speaking"}, nil
	case "reading":
		return []string{"reading"}, nil
	case "speaking":
		return []string{"speaking"}, nil
	default:
		return nil, fmt.Errorf("unknown source %q", source)
	}
}

func (r *LinglowMediaProgressBackfillRepository) audit(ctx context.Context, courseCode, source string) (LinglowMediaProgressBackfillSummary, error) {
	s := LinglowMediaProgressBackfillSummary{Source: source}
	var queries mediaBackfillQueries
	if source == "reading" {
		queries = readingMediaQueries
	} else {
		queries = speakingMediaQueries
	}
	if err := r.db.QueryRowContext(ctx, queries.legacyTotal, courseCode).Scan(&s.LegacyTotal); err != nil {
		return s, fmt.Errorf("count %s legacy total: %w", source, err)
	}
	if err := r.db.QueryRowContext(ctx, queries.mirroredTotal, courseCode).Scan(&s.MirroredTotal); err != nil {
		return s, fmt.Errorf("count %s mirrored total: %w", source, err)
	}
	if err := r.db.QueryRowContext(ctx, queries.missing, courseCode).Scan(&s.Missing); err != nil {
		return s, fmt.Errorf("count %s missing: %w", source, err)
	}
	if err := r.db.QueryRowContext(ctx, queries.unmapped, courseCode).Scan(&s.UnmappedTotal); err != nil {
		return s, fmt.Errorf("count %s unmapped: %w", source, err)
	}
	return s, nil
}

func (r *LinglowMediaProgressBackfillRepository) backfillReading(ctx context.Context, courseCode string, limit int, summary *LinglowMediaProgressBackfillSummary) error {
	query := readingMissingRowsSQL
	args := []interface{}{courseCode}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query missing reading rows: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var row readingMediaBackfillRow
		if err := rows.Scan(&row.UserCourseID, &row.LearningItemID, &row.UserID, &row.ChapterID, &row.ReadAt); err != nil {
			return fmt.Errorf("scan reading row: %w", err)
		}
		if err := r.insertReading(ctx, row); err != nil {
			return err
		}
		summary.Processed++
		summary.Inserted++
	}
	return rows.Err()
}

func (r *LinglowMediaProgressBackfillRepository) backfillSpeaking(ctx context.Context, courseCode string, limit int, summary *LinglowMediaProgressBackfillSummary) error {
	query := speakingMissingRowsSQL
	args := []interface{}{courseCode}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query missing speaking rows: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var row speakingMediaBackfillRow
		if err := rows.Scan(
			&row.UserCourseID, &row.LearningItemID, &row.AttemptID, &row.UserID, &row.SessionID, &row.TaskID,
			&row.AttemptNo, &row.Mode, &row.UnderstoodAnswer, &row.MeaningScore, &row.GrammarScore,
			&row.PronunciationScore, &row.FluencyScore, &row.IsAcceptable, &row.AudioQuality,
			&row.FeedbackRU, &row.BetterVersion, &row.RepeatTask, &row.CreatedAt,
		); err != nil {
			return fmt.Errorf("scan speaking row: %w", err)
		}
		if err := r.insertSpeaking(ctx, row); err != nil {
			return err
		}
		summary.Processed++
		summary.Inserted++
	}
	return rows.Err()
}

type readingMediaBackfillRow struct {
	UserCourseID   int64
	LearningItemID int64
	UserID         int64
	ChapterID      string
	ReadAt         sql.NullTime
}

type speakingMediaBackfillRow struct {
	UserCourseID       int64
	LearningItemID     int64
	AttemptID          int64
	UserID             int64
	SessionID          int64
	TaskID             string
	AttemptNo          int
	Mode               string
	UnderstoodAnswer   sql.NullString
	MeaningScore       sql.NullInt64
	GrammarScore       sql.NullInt64
	PronunciationScore sql.NullInt64
	FluencyScore       sql.NullInt64
	IsAcceptable       sql.NullBool
	AudioQuality       sql.NullString
	FeedbackRU         sql.NullString
	BetterVersion      sql.NullString
	RepeatTask         sql.NullString
	CreatedAt          sql.NullTime
}

func (r *LinglowMediaProgressBackfillRepository) insertReading(ctx context.Context, row readingMediaBackfillRow) error {
	readAt := nullTime(row.ReadAt)
	sourcePK := fmt.Sprintf("%d:%s", row.UserID, row.ChapterID)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exerciseID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO exercise_attempts (
			user_course_id, learning_item_id, mode, started_at, answered_at, is_correct, score,
			prompt_json, answer_json, result_json, source_table, source_pk
		)
		VALUES (?, ?, 'reading_completion', ?, ?, true, 100,
			jsonb_build_object('text_id', CAST(? AS text)),
			'{}'::jsonb,
			jsonb_build_object('completed', true),
			'reading_text_progress',
			?
		)
		RETURNING id
	`, row.UserCourseID, row.LearningItemID, readAt, readAt, row.ChapterID, sourcePK).Scan(&exerciseID); err != nil {
		return fmt.Errorf("insert reading exercise attempt: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO learning_events (
			user_course_id, learning_item_id, exercise_attempt_id, event_type, event_time,
			mode, source_table, source_pk, event_json
		)
		VALUES (?, ?, ?, 'reading_text_completed', ?, 'reading_completion', 'reading_text_progress',
			?,
			jsonb_build_object('text_id', CAST(? AS text), 'completed', true)
		)
	`, row.UserCourseID, row.LearningItemID, exerciseID, readAt, sourcePK, row.ChapterID); err != nil {
		return fmt.Errorf("insert reading learning event: %w", err)
	}
	return tx.Commit()
}

func (r *LinglowMediaProgressBackfillRepository) insertSpeaking(ctx context.Context, row speakingMediaBackfillRow) error {
	createdAt := nullTime(row.CreatedAt)
	score := averageNullableScores(row.MeaningScore, row.GrammarScore, row.PronunciationScore, row.FluencyScore)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exerciseID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO exercise_attempts (
			user_course_id, learning_item_id, mode, started_at, answered_at, is_correct, score,
			prompt_json, answer_json, result_json, source_table, source_pk
		)
		VALUES (?, ?, 'speaking_attempt', ?, ?, ?, ?,
			jsonb_build_object('task_id', CAST(? AS text), 'session_id', CAST(? AS bigint), 'attempt_no', CAST(? AS integer), 'mode', CAST(? AS text)),
			jsonb_build_object('understood_answer', CAST(? AS text)),
			jsonb_build_object(
				'meaning_score', CAST(? AS integer),
				'grammar_score', CAST(? AS integer),
				'pronunciation_score', CAST(? AS integer),
				'fluency_score', CAST(? AS integer),
				'audio_quality', CAST(? AS text),
				'feedback_ru', CAST(? AS text),
				'better_version', CAST(? AS text),
				'repeat_task', CAST(? AS text)
			),
			'speaking_attempts',
			CAST(? AS text)
		)
		RETURNING id
	`, row.UserCourseID, row.LearningItemID, createdAt, createdAt, nullableBool(row.IsAcceptable), score,
		row.TaskID, row.SessionID, row.AttemptNo, row.Mode, nullString(row.UnderstoodAnswer),
		nullableInt(row.MeaningScore), nullableInt(row.GrammarScore), nullableInt(row.PronunciationScore), nullableInt(row.FluencyScore),
		nullString(row.AudioQuality), nullString(row.FeedbackRU), nullString(row.BetterVersion), nullString(row.RepeatTask), fmt.Sprintf("%d", row.AttemptID)).Scan(&exerciseID); err != nil {
		return fmt.Errorf("insert speaking exercise attempt: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO learning_events (
			user_course_id, learning_item_id, exercise_attempt_id, event_type, event_time,
			mode, source_table, source_pk, event_json
		)
		VALUES (?, ?, ?, 'speaking_attempt_submitted', ?, 'speaking_attempt', 'speaking_attempts', CAST(? AS text),
			jsonb_build_object('task_id', CAST(? AS text), 'attempt_no', CAST(? AS integer), 'score', CAST(? AS integer))
		)
	`, row.UserCourseID, row.LearningItemID, exerciseID, createdAt, fmt.Sprintf("%d", row.AttemptID), row.TaskID, row.AttemptNo, score); err != nil {
		return fmt.Errorf("insert speaking learning event: %w", err)
	}
	return tx.Commit()
}

func nullableBool(v sql.NullBool) interface{} {
	if !v.Valid {
		return nil
	}
	return v.Bool
}

func nullableInt(v sql.NullInt64) interface{} {
	if !v.Valid {
		return nil
	}
	return int(v.Int64)
}

func averageNullableScores(scores ...sql.NullInt64) interface{} {
	total := 0
	count := 0
	for _, score := range scores {
		if score.Valid {
			total += int(score.Int64)
			count++
		}
	}
	if count == 0 {
		return nil
	}
	return total / count
}

type mediaBackfillQueries struct {
	legacyTotal   string
	mirroredTotal string
	missing       string
	unmapped      string
}

var readingMediaQueries = mediaBackfillQueries{
	legacyTotal: `
		SELECT COUNT(*)
		FROM reading_text_progress rtp
		JOIN user_courses uc ON uc.user_id = rtp.user_id
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = ?`,
	mirroredTotal: `
		SELECT COUNT(*)
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = ? AND ea.source_table = 'reading_text_progress'`,
	missing: `
		SELECT COUNT(*)
		FROM reading_text_progress rtp
		JOIN user_courses uc ON uc.user_id = rtp.user_id
		JOIN courses c ON c.id = uc.course_id
		JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'reading_text' AND li.source_id = rtp.chapter_id
		WHERE c.code = ?
			AND NOT EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = uc.id AND ea.source_table = 'reading_text_progress'
					AND ea.source_pk = CAST(rtp.user_id AS text) || ':' || rtp.chapter_id
			)`,
	unmapped: `
		SELECT COUNT(*)
		FROM reading_text_progress rtp
		JOIN user_courses uc ON uc.user_id = rtp.user_id
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = ?
			AND NOT EXISTS (
				SELECT 1 FROM learning_items li
				WHERE li.course_id = c.id AND li.source_kind = 'reading_text' AND li.source_id = rtp.chapter_id
			)`,
}

var speakingMediaQueries = mediaBackfillQueries{
	legacyTotal: `
		SELECT COUNT(*)
		FROM speaking_attempts sa
		JOIN user_courses uc ON uc.user_id = sa.user_id
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = ?`,
	mirroredTotal: `
		SELECT COUNT(*)
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = ? AND ea.source_table = 'speaking_attempts'`,
	missing: `
		SELECT COUNT(*)
		FROM speaking_attempts sa
		JOIN user_courses uc ON uc.user_id = sa.user_id
		JOIN courses c ON c.id = uc.course_id
		JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'speaking_task' AND li.source_id = sa.task_id
		WHERE c.code = ?
			AND NOT EXISTS (
				SELECT 1 FROM exercise_attempts ea
				WHERE ea.user_course_id = uc.id AND ea.source_table = 'speaking_attempts'
					AND ea.source_pk = CAST(sa.id AS text)
			)`,
	unmapped: `
		SELECT COUNT(*)
		FROM speaking_attempts sa
		JOIN user_courses uc ON uc.user_id = sa.user_id
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = ?
			AND NOT EXISTS (
				SELECT 1 FROM learning_items li
				WHERE li.course_id = c.id AND li.source_kind = 'speaking_task' AND li.source_id = sa.task_id
			)`,
}

const readingMissingRowsSQL = `
	SELECT uc.id AS user_course_id, li.id AS learning_item_id, rtp.user_id, rtp.chapter_id, rtp.read_at
	FROM reading_text_progress rtp
	JOIN user_courses uc ON uc.user_id = rtp.user_id
	JOIN courses c ON c.id = uc.course_id
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'reading_text' AND li.source_id = rtp.chapter_id
	WHERE c.code = ?
		AND NOT EXISTS (
			SELECT 1 FROM exercise_attempts ea
			WHERE ea.user_course_id = uc.id AND ea.source_table = 'reading_text_progress'
				AND ea.source_pk = CAST(rtp.user_id AS text) || ':' || rtp.chapter_id
		)
	ORDER BY rtp.read_at, rtp.user_id, rtp.chapter_id`

const speakingMissingRowsSQL = `
	SELECT uc.id AS user_course_id, li.id AS learning_item_id,
		sa.id, sa.user_id, sa.session_id, sa.task_id, sa.attempt_no, sa.mode,
		sa.understood_answer, sa.meaning_score, sa.grammar_score, sa.pronunciation_score, sa.fluency_score,
		sa.is_acceptable, sa.audio_quality, sa.feedback_ru, sa.better_version, sa.repeat_task, sa.created_at
	FROM speaking_attempts sa
	JOIN user_courses uc ON uc.user_id = sa.user_id
	JOIN courses c ON c.id = uc.course_id
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'speaking_task' AND li.source_id = sa.task_id
	WHERE c.code = ?
		AND NOT EXISTS (
			SELECT 1 FROM exercise_attempts ea
			WHERE ea.user_course_id = uc.id AND ea.source_table = 'speaking_attempts'
				AND ea.source_pk = CAST(sa.id AS text)
		)
	ORDER BY sa.id`
