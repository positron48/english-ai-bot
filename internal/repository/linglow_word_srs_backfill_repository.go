package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/config"
)

type LinglowWordSRSBackfillRepository struct {
	db *sql.DB
}

type LinglowWordSRSBackfillOptions struct {
	Commit bool
	Limit  int
}

type LinglowWordSRSBackfillSummary struct {
	CourseCode    string
	LegacyTotal   int64
	MappedTotal   int64
	SRSTotal      int64
	Missing       int64
	Processed     int64
	Upserted      int64
	UnmappedTotal int64
}

func NewLinglowWordSRSBackfillRepository(db *sql.DB) *LinglowWordSRSBackfillRepository {
	return &LinglowWordSRSBackfillRepository{db: db}
}

func (r *LinglowWordSRSBackfillRepository) Backfill(ctx context.Context, lc config.LearningConfig, opts LinglowWordSRSBackfillOptions) (*LinglowWordSRSBackfillSummary, error) {
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	summary, err := r.audit(ctx, courseCode)
	if err != nil {
		return nil, err
	}
	if !opts.Commit {
		return summary, nil
	}

	query := missingWordSRSRowsSQL
	args := []interface{}{courseCode}
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query missing word srs rows: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var row wordSRSBackfillRow
		if err := rows.Scan(
			&row.UserCourseID,
			&row.LearningItemID,
			&row.State,
			&row.EF,
			&row.NextDueAt,
			&row.LastReviewAt,
			&row.Reps,
			&row.LapseCount,
			&row.IntervalDays,
			&row.LearningStep,
			&row.LastQuality,
			&row.StatsJSON,
			&row.UserCardID,
			&row.TrainingCardID,
			&row.WordCardID,
			&row.Direction,
		); err != nil {
			return nil, fmt.Errorf("scan word srs row: %w", err)
		}
		if err := r.upsertWordSRSRow(ctx, row); err != nil {
			return nil, err
		}
		summary.Processed++
		summary.Upserted++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate word srs rows: %w", err)
	}
	refreshed, err := r.audit(ctx, courseCode)
	if err != nil {
		return nil, err
	}
	refreshed.Processed = summary.Processed
	refreshed.Upserted = summary.Upserted
	return refreshed, nil
}

func (r *LinglowWordSRSBackfillRepository) audit(ctx context.Context, courseCode string) (*LinglowWordSRSBackfillSummary, error) {
	s := &LinglowWordSRSBackfillSummary{CourseCode: courseCode}
	if err := r.db.QueryRowContext(ctx, wordSRSAuditLegacyTotalSQL, courseCode).Scan(&s.LegacyTotal); err != nil {
		return nil, fmt.Errorf("count legacy word srs: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, wordSRSAuditMappedTotalSQL, courseCode).Scan(&s.MappedTotal); err != nil {
		return nil, fmt.Errorf("count mapped word srs: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, wordSRSAuditSRSTotalSQL, courseCode).Scan(&s.SRSTotal); err != nil {
		return nil, fmt.Errorf("count linglow word srs: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, wordSRSAuditMissingSQL, courseCode).Scan(&s.Missing); err != nil {
		return nil, fmt.Errorf("count missing word srs: %w", err)
	}
	s.UnmappedTotal = s.LegacyTotal - s.MappedTotal
	if s.UnmappedTotal < 0 {
		s.UnmappedTotal = 0
	}
	return s, nil
}

type wordSRSBackfillRow struct {
	UserCourseID   int64
	LearningItemID int64
	State          string
	EF             sql.NullFloat64
	NextDueAt      sql.NullTime
	LastReviewAt   sql.NullTime
	Reps           int
	LapseCount     int
	IntervalDays   int
	LearningStep   int
	LastQuality    sql.NullInt64
	StatsJSON      sql.NullString
	UserCardID     int64
	TrainingCardID int64
	WordCardID     int64
	Direction      string
}

func (r *LinglowWordSRSBackfillRepository) upsertWordSRSRow(ctx context.Context, row wordSRSBackfillRow) error {
	state := normalizeLinglowSRSState(row.State)
	statsJSON := strings.TrimSpace(row.StatsJSON.String)
	if !row.StatsJSON.Valid || statsJSON == "" {
		statsJSON = "{}"
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO srs_items (
			user_course_id, learning_item_id, state, stability, difficulty, due_at, last_review_at,
			reps, lapse_count, stats_json, created_at, updated_at
		)
		VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, jsonb_build_object(
				'legacy', jsonb_build_object(
					'source_table', 'user_cards',
					'user_card_id', CAST(? AS bigint),
					'training_card_id', CAST(? AS bigint),
					'word_card_id', CAST(? AS bigint),
					'direction', CAST(? AS text),
					'interval_days', CAST(? AS integer),
					'learning_step', CAST(? AS integer),
					'last_quality', CAST(? AS integer),
					'stats_json', CAST(? AS jsonb)
				)
			), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON CONFLICT (user_course_id, learning_item_id) DO UPDATE SET
			state = excluded.state,
			stability = excluded.stability,
			difficulty = excluded.difficulty,
			due_at = excluded.due_at,
			last_review_at = excluded.last_review_at,
			reps = excluded.reps,
			lapse_count = excluded.lapse_count,
			stats_json = excluded.stats_json,
			updated_at = CURRENT_TIMESTAMP
	`, row.UserCourseID, row.LearningItemID, state, row.EF, row.EF, row.NextDueAt, row.LastReviewAt, row.Reps, row.LapseCount,
		row.UserCardID, row.TrainingCardID, row.WordCardID, row.Direction, row.IntervalDays, row.LearningStep, row.LastQuality, statsJSON); err != nil {
		return fmt.Errorf("upsert word srs item user_course=%d item=%d: %w", row.UserCourseID, row.LearningItemID, err)
	}
	return nil
}

func normalizeLinglowSRSState(state string) string {
	switch strings.TrimSpace(strings.ToLower(state)) {
	case "new":
		return "new"
	case "learning":
		return "learning"
	case "review":
		return "review"
	case "relearning":
		return "relearning"
	case "mastered", "known":
		return "mastered"
	case "suspended":
		return "suspended"
	default:
		return "new"
	}
}

const wordSRSAuditLegacyTotalSQL = `
	SELECT COUNT(*)
	FROM user_cards ucards
	JOIN user_courses ucourse ON ucourse.user_id = ucards.user_id
	JOIN courses c ON c.id = ucourse.course_id
	WHERE c.code = ?`

const wordSRSAuditMappedTotalSQL = `
	SELECT COUNT(*)
	FROM user_cards ucards
	JOIN training_cards tc ON tc.id = ucards.training_card_id
	JOIN user_courses ucourse ON ucourse.user_id = ucards.user_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.course_id = c.id
		AND li.source_kind = 'word_card'
		AND li.source_id = CAST(tc.word_card_id AS TEXT)
	WHERE c.code = ?`

const wordSRSAuditSRSTotalSQL = `
	SELECT COUNT(*)
	FROM srs_items si
	JOIN user_courses ucourse ON ucourse.id = si.user_course_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.id = si.learning_item_id AND li.source_kind = 'word_card'
	WHERE c.code = ?`

const wordSRSAuditMissingSQL = `
	SELECT COUNT(*)
	FROM user_cards ucards
	JOIN training_cards tc ON tc.id = ucards.training_card_id
	JOIN user_courses ucourse ON ucourse.user_id = ucards.user_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.course_id = c.id
		AND li.source_kind = 'word_card'
		AND li.source_id = CAST(tc.word_card_id AS TEXT)
	WHERE c.code = ?
		AND NOT EXISTS (
			SELECT 1 FROM srs_items si
			WHERE si.user_course_id = ucourse.id AND si.learning_item_id = li.id
		)`

const missingWordSRSRowsSQL = `
	SELECT ucourse.id AS user_course_id, li.id AS learning_item_id,
		ucards.state, ucards.ef, ucards.next_due_at, ucards.last_review_at,
		ucards.reps, ucards.lapse_count, ucards.interval_days, ucards.learning_step,
		ucards.last_quality, ucards.stats_json,
		ucards.id AS user_card_id, tc.id AS training_card_id, tc.word_card_id, ucards.direction
	FROM user_cards ucards
	JOIN training_cards tc ON tc.id = ucards.training_card_id
	JOIN user_courses ucourse ON ucourse.user_id = ucards.user_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.course_id = c.id
		AND li.source_kind = 'word_card'
		AND li.source_id = CAST(tc.word_card_id AS TEXT)
	WHERE c.code = ?
		AND NOT EXISTS (
			SELECT 1 FROM srs_items si
			WHERE si.user_course_id = ucourse.id AND si.learning_item_id = li.id
		)
	ORDER BY ucards.id`
