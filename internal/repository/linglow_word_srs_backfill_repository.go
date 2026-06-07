package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
)

type LinglowWordSRSBackfillRepository struct {
	db *sql.DB
}

type LinglowWordSRSBackfillOptions struct {
	Commit       bool
	Resync       bool
	PruneOrphans bool
	Limit        int
}

type LinglowWordSRSBackfillSummary struct {
	CourseCode    string
	LegacyTotal   int64
	MappedTotal   int64
	SRSTotal      int64
	Missing       int64
	Processed     int64
	Upserted      int64
	PrunedOrphans int64
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
	if opts.Resync {
		query = resyncWordSRSRowsSQL
	}
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
	if opts.Resync && opts.PruneOrphans {
		pruned, err := r.pruneOrphanWordSRSItems(ctx, courseCode)
		if err != nil {
			return nil, err
		}
		summary.PrunedOrphans = pruned
	}
	refreshed, err := r.audit(ctx, courseCode)
	if err != nil {
		return nil, err
	}
	refreshed.Processed = summary.Processed
	refreshed.Upserted = summary.Upserted
	refreshed.PrunedOrphans = summary.PrunedOrphans
	return refreshed, nil
}

func (r *LinglowWordSRSBackfillRepository) SyncWordLearningItemForUser(ctx context.Context, lc config.LearningConfig, userID, wordCardID int64) error {
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" || userID == 0 || wordCardID == 0 {
		return nil
	}
	rows, err := r.db.QueryContext(ctx, syncWordSRSForWordCardSQL, courseCode, userID, wordCardID)
	if err != nil {
		return fmt.Errorf("query word srs mirror row: %w", err)
	}
	defer rows.Close()
	var synced bool
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
			return fmt.Errorf("scan word srs mirror row: %w", err)
		}
		if err := r.upsertWordSRSRow(ctx, row); err != nil {
			return err
		}
		synced = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate word srs mirror rows: %w", err)
	}
	if !synced {
		return nil
	}
	return nil
}

func (r *LinglowWordSRSBackfillRepository) pruneOrphanWordSRSItems(ctx context.Context, courseCode string) (int64, error) {
	res, err := r.db.ExecContext(ctx, pruneOrphanWordSRSItemsSQL, courseCode)
	if err != nil {
		return 0, fmt.Errorf("prune orphan word srs items: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune orphan word srs rows affected: %w", err)
	}
	return affected, nil
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
	state := normalizeLinglowSRSStateForBackfill(row.State, row.NextDueAt)
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

func legacySRSDueAt(dueAt sql.NullTime) bool {
	if !dueAt.Valid {
		return true
	}
	return !dueAt.Time.After(time.Now().UTC())
}

// normalizeLinglowSRSStateForBackfill maps legacy due "new" cards into active canonical SRS.
func normalizeLinglowSRSStateForBackfill(state string, dueAt sql.NullTime) string {
	normalized := normalizeLinglowSRSState(state)
	if normalized != "new" {
		return normalized
	}
	if legacySRSDueAt(dueAt) {
		return "learning"
	}
	return "new"
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

const wordSRSAggregatedRowsSQL = `
	WITH card_rows AS (
		SELECT
			ucourse.id AS user_course_id,
			li.id AS learning_item_id,
			CASE WHEN uwk.status = 'known' THEN 'mastered' ELSE COALESCE(NULLIF(ucards.state, ''), 'new') END AS state,
			ucards.ef,
			CASE WHEN uwk.status = 'known' THEN NULL ELSE ucards.next_due_at END AS next_due_at,
			ucards.last_review_at,
			ucards.reps,
			ucards.lapse_count,
			ucards.interval_days,
			ucards.learning_step,
			ucards.last_quality,
			ucards.stats_json,
			ucards.id AS user_card_id,
			tc.id AS training_card_id,
			tc.word_card_id,
			ucards.direction,
			CASE
				WHEN uwk.status = 'known' THEN FALSE
				WHEN ucards.next_due_at IS NULL OR ucards.next_due_at <= CURRENT_TIMESTAMP THEN TRUE
				ELSE FALSE
			END AS is_legacy_due
		FROM user_cards ucards
		JOIN training_cards tc ON tc.id = ucards.training_card_id
		JOIN user_courses ucourse ON ucourse.user_id = ucards.user_id
		JOIN courses c ON c.id = ucourse.course_id
		JOIN learning_items li ON li.course_id = c.id
			AND li.source_kind = 'word_card'
			AND li.source_id = CAST(tc.word_card_id AS TEXT)
		LEFT JOIN user_word_knowledge uwk
			ON uwk.user_id = ucards.user_id AND uwk.word_card_id = tc.word_card_id
		WHERE c.code = ?
	), ranked AS (
		SELECT
			card_rows.*,
			ROW_NUMBER() OVER (
				PARTITION BY user_course_id, learning_item_id
				ORDER BY
					CASE WHEN state = 'mastered' THEN 1 ELSE 0 END,
					CASE WHEN is_legacy_due THEN 0 ELSE 1 END,
					CASE lower(trim(state))
						WHEN 'relearning' THEN 0
						WHEN 'learning' THEN 1
						WHEN 'new' THEN 2
						WHEN 'review' THEN 3
						ELSE 4
					END,
					next_due_at NULLS FIRST,
					user_card_id
			) AS rn
		FROM card_rows
	)`

const missingWordSRSRowsSQL = wordSRSAggregatedRowsSQL + `
	SELECT
		user_course_id, learning_item_id, state, ef, next_due_at, last_review_at,
		reps, lapse_count, interval_days, learning_step, last_quality, stats_json,
		user_card_id, training_card_id, word_card_id, direction
	FROM ranked
	WHERE rn = 1
		AND NOT EXISTS (
			SELECT 1 FROM srs_items si
			WHERE si.user_course_id = ranked.user_course_id AND si.learning_item_id = ranked.learning_item_id
		)
	ORDER BY user_course_id, learning_item_id`

const resyncWordSRSRowsSQL = wordSRSAggregatedRowsSQL + `
	SELECT
		user_course_id, learning_item_id, state, ef, next_due_at, last_review_at,
		reps, lapse_count, interval_days, learning_step, last_quality, stats_json,
		user_card_id, training_card_id, word_card_id, direction
	FROM ranked
	WHERE rn = 1
	ORDER BY user_course_id, learning_item_id`

const syncWordSRSForWordCardSQL = `
	WITH card_rows AS (
		SELECT
			ucourse.id AS user_course_id,
			li.id AS learning_item_id,
			CASE WHEN uwk.status = 'known' THEN 'mastered' ELSE COALESCE(NULLIF(ucards.state, ''), 'new') END AS state,
			ucards.ef,
			CASE WHEN uwk.status = 'known' THEN NULL ELSE ucards.next_due_at END AS next_due_at,
			ucards.last_review_at,
			ucards.reps,
			ucards.lapse_count,
			ucards.interval_days,
			ucards.learning_step,
			ucards.last_quality,
			ucards.stats_json,
			ucards.id AS user_card_id,
			tc.id AS training_card_id,
			tc.word_card_id,
			ucards.direction,
			CASE
				WHEN uwk.status = 'known' THEN FALSE
				WHEN ucards.next_due_at IS NULL OR ucards.next_due_at <= CURRENT_TIMESTAMP THEN TRUE
				ELSE FALSE
			END AS is_legacy_due
		FROM user_cards ucards
		JOIN training_cards tc ON tc.id = ucards.training_card_id
		JOIN user_courses ucourse ON ucourse.user_id = ucards.user_id
		JOIN courses c ON c.id = ucourse.course_id
		JOIN learning_items li ON li.course_id = c.id
			AND li.source_kind = 'word_card'
			AND li.source_id = CAST(tc.word_card_id AS TEXT)
		LEFT JOIN user_word_knowledge uwk
			ON uwk.user_id = ucards.user_id AND uwk.word_card_id = tc.word_card_id
		WHERE c.code = ? AND ucards.user_id = ? AND tc.word_card_id = ?
	), ranked AS (
		SELECT
			card_rows.*,
			ROW_NUMBER() OVER (
				PARTITION BY user_course_id, learning_item_id
				ORDER BY
					CASE WHEN state = 'mastered' THEN 1 ELSE 0 END,
					CASE WHEN is_legacy_due THEN 0 ELSE 1 END,
					CASE lower(trim(state))
						WHEN 'relearning' THEN 0
						WHEN 'learning' THEN 1
						WHEN 'new' THEN 2
						WHEN 'review' THEN 3
						ELSE 4
					END,
					next_due_at NULLS FIRST,
					user_card_id
			) AS rn
		FROM card_rows
	)
	SELECT
		user_course_id, learning_item_id, state, ef, next_due_at, last_review_at,
		reps, lapse_count, interval_days, learning_step, last_quality, stats_json,
		user_card_id, training_card_id, word_card_id, direction
	FROM ranked
	WHERE rn = 1`

const pruneOrphanWordSRSItemsSQL = `
	UPDATE srs_items si
	SET state = 'suspended',
		due_at = NULL,
		updated_at = CURRENT_TIMESTAMP
	FROM learning_items li, user_courses uc, courses c
	WHERE si.learning_item_id = li.id
		AND li.source_kind = 'word_card'
		AND si.user_course_id = uc.id
		AND uc.course_id = c.id
		AND c.code = ?
		AND si.state IN ('learning', 'review', 'relearning')
		AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		AND NOT EXISTS (
			SELECT 1
			FROM user_cards ucards
			JOIN training_cards tc ON tc.id = ucards.training_card_id
			WHERE ucards.user_id = uc.user_id
				AND li.source_id = CAST(tc.word_card_id AS TEXT)
		)`
