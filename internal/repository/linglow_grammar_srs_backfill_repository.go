package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tgbot-skeleton/internal/config"
)

type LinglowGrammarSRSBackfillRepository struct {
	db *sql.DB
}

type LinglowGrammarSRSBackfillOptions struct {
	Commit       bool
	Resync       bool
	PruneOrphans bool
	Limit        int
}

type LinglowGrammarSRSBackfillSummary struct {
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

func NewLinglowGrammarSRSBackfillRepository(db *sql.DB) *LinglowGrammarSRSBackfillRepository {
	return &LinglowGrammarSRSBackfillRepository{db: db}
}

func (r *LinglowGrammarSRSBackfillRepository) Backfill(ctx context.Context, lc config.LearningConfig, opts LinglowGrammarSRSBackfillOptions) (*LinglowGrammarSRSBackfillSummary, error) {
	courseCode := CourseCodeForLearning(lc)
	targetLang := strings.TrimSpace(strings.ToLower(lc.TargetLang))
	legacyCourseID := strings.TrimSpace(strings.ToLower(lc.GrammarBundleID))
	if legacyCourseID == "" {
		legacyCourseID = targetLang
	}
	if courseCode == "" || targetLang == "" || legacyCourseID == "" {
		return nil, fmt.Errorf("learning config is incomplete")
	}
	summary, err := r.audit(ctx, courseCode, targetLang, legacyCourseID)
	if err != nil {
		return nil, err
	}
	if !opts.Commit {
		return summary, nil
	}
	query := missingGrammarSRSRowsSQL
	if opts.Resync {
		query = resyncGrammarSRSRowsSQL
	}
	args := []interface{}{courseCode, targetLang, legacyCourseID}
	if opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query missing grammar srs rows: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var row grammarSRSBackfillRow
		if err := rows.Scan(
			&row.UserCourseID,
			&row.LearningItemID,
			&row.State,
			&row.Ease,
			&row.NextReviewAt,
			&row.LastReviewAt,
			&row.ReviewCount,
			&row.LapseCount,
			&row.CorrectCount,
			&row.WrongCount,
			&row.CorrectStreak,
			&row.WrongStreak,
			&row.IntervalDays,
			&row.MasteryScore,
			&row.MemoryID,
			&row.ChapterID,
			&row.TheoryBlockID,
			&row.ConceptID,
		); err != nil {
			return nil, fmt.Errorf("scan grammar srs row: %w", err)
		}
		if err := r.upsertGrammarSRSRow(ctx, row); err != nil {
			return nil, err
		}
		summary.Processed++
		summary.Upserted++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate grammar srs rows: %w", err)
	}
	if opts.Resync && opts.PruneOrphans {
		pruned, err := r.pruneOrphanGrammarSRSItems(ctx, courseCode, targetLang, legacyCourseID)
		if err != nil {
			return nil, err
		}
		summary.PrunedOrphans = pruned
	}
	refreshed, err := r.audit(ctx, courseCode, targetLang, legacyCourseID)
	if err != nil {
		return nil, err
	}
	refreshed.Processed = summary.Processed
	refreshed.Upserted = summary.Upserted
	refreshed.PrunedOrphans = summary.PrunedOrphans
	return refreshed, nil
}

func (r *LinglowGrammarSRSBackfillRepository) pruneOrphanGrammarSRSItems(ctx context.Context, courseCode, targetLang, legacyCourseID string) (int64, error) {
	res, err := r.db.ExecContext(ctx, pruneOrphanGrammarSRSItemsSQL, courseCode, targetLang, legacyCourseID)
	if err != nil {
		return 0, fmt.Errorf("prune orphan grammar srs items: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune orphan grammar srs rows affected: %w", err)
	}
	return affected, nil
}

func (r *LinglowGrammarSRSBackfillRepository) audit(ctx context.Context, courseCode, targetLang, legacyCourseID string) (*LinglowGrammarSRSBackfillSummary, error) {
	s := &LinglowGrammarSRSBackfillSummary{CourseCode: courseCode}
	args := []interface{}{courseCode, targetLang, legacyCourseID}
	if err := r.db.QueryRowContext(ctx, grammarSRSAuditLegacyTotalSQL, args...).Scan(&s.LegacyTotal); err != nil {
		return nil, fmt.Errorf("count legacy grammar srs: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, grammarSRSAuditMappedTotalSQL, args...).Scan(&s.MappedTotal); err != nil {
		return nil, fmt.Errorf("count mapped grammar srs: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, grammarSRSAuditSRSTotalSQL, courseCode).Scan(&s.SRSTotal); err != nil {
		return nil, fmt.Errorf("count linglow grammar srs: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, grammarSRSAuditMissingSQL, args...).Scan(&s.Missing); err != nil {
		return nil, fmt.Errorf("count missing grammar srs: %w", err)
	}
	s.UnmappedTotal = s.LegacyTotal - s.MappedTotal
	if s.UnmappedTotal < 0 {
		s.UnmappedTotal = 0
	}
	return s, nil
}

type grammarSRSBackfillRow struct {
	UserCourseID   int64
	LearningItemID int64
	State          string
	Ease           float64
	NextReviewAt   sql.NullTime
	LastReviewAt   sql.NullTime
	ReviewCount    int
	LapseCount     int
	CorrectCount   int
	WrongCount     int
	CorrectStreak  int
	WrongStreak    int
	IntervalDays   int
	MasteryScore   int
	MemoryID       int64
	ChapterID      string
	TheoryBlockID  string
	ConceptID      sql.NullString
}

func (r *LinglowGrammarSRSBackfillRepository) upsertGrammarSRSRow(ctx context.Context, row grammarSRSBackfillRow) error {
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO srs_items (
			user_course_id, learning_item_id, state, stability, difficulty, due_at, last_review_at,
			reps, lapse_count, stats_json, created_at, updated_at
		)
		VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, jsonb_build_object(
				'legacy', jsonb_build_object(
					'source_table', 'grammar_theory_memory',
					'memory_id', CAST(? AS bigint),
					'chapter_id', CAST(? AS text),
					'theory_block_id', CAST(? AS text),
					'concept_id', CAST(? AS text),
					'correct_count', CAST(? AS integer),
					'wrong_count', CAST(? AS integer),
					'correct_streak', CAST(? AS integer),
					'wrong_streak', CAST(? AS integer),
					'interval_days', CAST(? AS integer),
					'mastery_score', CAST(? AS integer)
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
	`, row.UserCourseID, row.LearningItemID, normalizeLinglowSRSState(row.State), row.Ease, row.Ease, row.NextReviewAt, row.LastReviewAt, row.ReviewCount, row.LapseCount,
		row.MemoryID, row.ChapterID, row.TheoryBlockID, nullString(row.ConceptID), row.CorrectCount, row.WrongCount, row.CorrectStreak, row.WrongStreak, row.IntervalDays, row.MasteryScore); err != nil {
		return fmt.Errorf("upsert grammar srs item user_course=%d item=%d: %w", row.UserCourseID, row.LearningItemID, err)
	}
	return nil
}

const grammarSRSAuditLegacyTotalSQL = `
	SELECT COUNT(*)
	FROM grammar_theory_memory gtm
	JOIN user_courses ucourse ON ucourse.user_id = gtm.user_id
	JOIN courses c ON c.id = ucourse.course_id
	WHERE c.code = ? AND lower(gtm.language) = ? AND lower(gtm.course_id) = ?`

const grammarSRSAuditMappedTotalSQL = `
	SELECT COUNT(*)
	FROM grammar_theory_memory gtm
	JOIN user_courses ucourse ON ucourse.user_id = gtm.user_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.course_id = c.id
		AND li.source_kind = 'grammar_theory_block'
		AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
	WHERE c.code = ? AND lower(gtm.language) = ? AND lower(gtm.course_id) = ?`

const grammarSRSAuditSRSTotalSQL = `
	SELECT COUNT(*)
	FROM srs_items si
	JOIN user_courses ucourse ON ucourse.id = si.user_course_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.id = si.learning_item_id AND li.source_kind = 'grammar_theory_block'
	WHERE c.code = ?`

const grammarSRSAuditMissingSQL = `
	SELECT COUNT(*)
	FROM grammar_theory_memory gtm
	JOIN user_courses ucourse ON ucourse.user_id = gtm.user_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.course_id = c.id
		AND li.source_kind = 'grammar_theory_block'
		AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
	WHERE c.code = ? AND lower(gtm.language) = ? AND lower(gtm.course_id) = ?
		AND NOT EXISTS (
			SELECT 1 FROM srs_items si
			WHERE si.user_course_id = ucourse.id AND si.learning_item_id = li.id
		)`

const missingGrammarSRSRowsSQL = `
	SELECT ucourse.id AS user_course_id, li.id AS learning_item_id,
		gtm.state, gtm.ease, gtm.next_review_at, gtm.last_review_at,
		gtm.review_count, gtm.lapse_count, gtm.correct_count, gtm.wrong_count,
		gtm.correct_streak, gtm.wrong_streak, gtm.interval_days, gtm.mastery_score,
		gtm.id AS memory_id, gtm.chapter_id, gtm.theory_block_id, gtm.concept_id
	FROM grammar_theory_memory gtm
	JOIN user_courses ucourse ON ucourse.user_id = gtm.user_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.course_id = c.id
		AND li.source_kind = 'grammar_theory_block'
		AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
	WHERE c.code = ? AND lower(gtm.language) = ? AND lower(gtm.course_id) = ?
		AND NOT EXISTS (
			SELECT 1 FROM srs_items si
			WHERE si.user_course_id = ucourse.id AND si.learning_item_id = li.id
		)
		ORDER BY gtm.id`

const resyncGrammarSRSRowsSQL = `
	SELECT ucourse.id AS user_course_id, li.id AS learning_item_id,
		gtm.state, gtm.ease, gtm.next_review_at, gtm.last_review_at,
		gtm.review_count, gtm.lapse_count, gtm.correct_count, gtm.wrong_count,
		gtm.correct_streak, gtm.wrong_streak, gtm.interval_days, gtm.mastery_score,
		gtm.id AS memory_id, gtm.chapter_id, gtm.theory_block_id, gtm.concept_id
	FROM grammar_theory_memory gtm
	JOIN user_courses ucourse ON ucourse.user_id = gtm.user_id
	JOIN courses c ON c.id = ucourse.course_id
	JOIN learning_items li ON li.course_id = c.id
		AND li.source_kind = 'grammar_theory_block'
		AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
	WHERE c.code = ? AND lower(gtm.language) = ? AND lower(gtm.course_id) = ?
	ORDER BY gtm.id`

const pruneOrphanGrammarSRSItemsSQL = `
	UPDATE srs_items si
	SET state = 'suspended',
		due_at = NULL,
		updated_at = CURRENT_TIMESTAMP
	FROM learning_items li, user_courses uc, courses c
	WHERE si.learning_item_id = li.id
		AND li.source_kind = 'grammar_theory_block'
		AND si.user_course_id = uc.id
		AND uc.course_id = c.id
		AND c.code = ?
		AND si.state IN ('learning', 'review', 'relearning')
		AND (si.due_at IS NULL OR si.due_at <= CURRENT_TIMESTAMP)
		AND NOT EXISTS (
			SELECT 1
			FROM grammar_theory_memory gtm
			WHERE gtm.user_id = uc.user_id
				AND lower(gtm.language) = ?
				AND lower(gtm.course_id) = ?
				AND li.source_id = gtm.chapter_id || ':' || gtm.theory_block_id
		)`
