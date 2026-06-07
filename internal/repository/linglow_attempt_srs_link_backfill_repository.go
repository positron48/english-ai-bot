package repository

import (
	"context"
	"database/sql"
	"fmt"

	"tgbot-skeleton/internal/config"
)

type LinglowAttemptSRSLinkBackfillRepository struct {
	db *sql.DB
}

type LinglowAttemptSRSLinkBackfillOptions struct {
	Commit bool
	Limit  int
}

type LinglowAttemptSRSLinkBackfillSummary struct {
	Source        string
	Missing       int64
	Updated       int64
	UnmappedTotal int64
}

func NewLinglowAttemptSRSLinkBackfillRepository(db *sql.DB) *LinglowAttemptSRSLinkBackfillRepository {
	return &LinglowAttemptSRSLinkBackfillRepository{db: db}
}

func (r *LinglowAttemptSRSLinkBackfillRepository) Backfill(ctx context.Context, lc config.LearningConfig, opts LinglowAttemptSRSLinkBackfillOptions) ([]LinglowAttemptSRSLinkBackfillSummary, error) {
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	summaries := make([]LinglowAttemptSRSLinkBackfillSummary, 0, 2)

	word, err := r.backfillSource(ctx, courseCode, "word-training", opts, wordAttemptSRSLinksMissingSQL, wordAttemptSRSLinksNullTotalSQL, wordAttemptSRSLinksUpdateSQL, wordAttemptSRSLinksUpdateAllSQL, "")
	if err != nil {
		return nil, err
	}
	summaries = append(summaries, word)

	grammar, err := r.backfillSource(ctx, courseCode, "grammar-training", opts, grammarAttemptSRSLinksMissingSQL, grammarAttemptSRSLinksNullTotalSQL, grammarAttemptSRSLinksUpdateSQL, grammarAttemptSRSLinksUpdateAllSQL, grammarLearningEventsItemSyncSQL)
	if err != nil {
		return nil, err
	}
	summaries = append(summaries, grammar)
	return summaries, nil
}

func (r *LinglowAttemptSRSLinkBackfillRepository) backfillSource(
	ctx context.Context,
	courseCode, source string,
	opts LinglowAttemptSRSLinkBackfillOptions,
	missingSQL, nullTotalSQL, updateBatchSQL, updateAllSQL, afterUpdateSQL string,
) (LinglowAttemptSRSLinkBackfillSummary, error) {
	summary, err := r.audit(ctx, courseCode, source, missingSQL, nullTotalSQL)
	if err != nil {
		return LinglowAttemptSRSLinkBackfillSummary{}, err
	}
	if !opts.Commit {
		return summary, nil
	}
	for {
		updated, err := r.execBatchUpdate(ctx, courseCode, updateBatchSQL, updateAllSQL, opts.Limit)
		if err != nil {
			return LinglowAttemptSRSLinkBackfillSummary{}, fmt.Errorf("update %s attempt srs links: %w", source, err)
		}
		summary.Updated += updated
		if updated == 0 {
			break
		}
		if opts.Limit > 0 {
			refreshed, err := r.audit(ctx, courseCode, source, missingSQL, nullTotalSQL)
			if err != nil {
				return LinglowAttemptSRSLinkBackfillSummary{}, err
			}
			summary.Missing = refreshed.Missing
			summary.UnmappedTotal = refreshed.UnmappedTotal
		}
	}
	if afterUpdateSQL != "" {
		if _, err := r.db.ExecContext(ctx, afterUpdateSQL, courseCode); err != nil {
			return LinglowAttemptSRSLinkBackfillSummary{}, fmt.Errorf("sync grammar learning event items: %w", err)
		}
	}
	refreshed, err := r.audit(ctx, courseCode, source, missingSQL, nullTotalSQL)
	if err != nil {
		return LinglowAttemptSRSLinkBackfillSummary{}, err
	}
	summary.Missing = refreshed.Missing
	summary.UnmappedTotal = refreshed.UnmappedTotal
	return summary, nil
}

func (r *LinglowAttemptSRSLinkBackfillRepository) audit(ctx context.Context, courseCode, source, missingSQL, nullTotalSQL string) (LinglowAttemptSRSLinkBackfillSummary, error) {
	missing, err := queryCount(ctx, r.db, missingSQL, courseCode)
	if err != nil {
		return LinglowAttemptSRSLinkBackfillSummary{}, fmt.Errorf("count missing %s srs links: %w", source, err)
	}
	nullTotal, err := queryCount(ctx, r.db, nullTotalSQL, courseCode)
	if err != nil {
		return LinglowAttemptSRSLinkBackfillSummary{}, fmt.Errorf("count null %s srs links: %w", source, err)
	}
	unmapped := nullTotal - missing
	if unmapped < 0 {
		unmapped = 0
	}
	return LinglowAttemptSRSLinkBackfillSummary{Source: source, Missing: missing, UnmappedTotal: unmapped}, nil
}

func (r *LinglowAttemptSRSLinkBackfillRepository) execBatchUpdate(ctx context.Context, courseCode, updateBatchSQL, updateAllSQL string, limit int) (int64, error) {
	if limit > 0 {
		return execRowsAffected(ctx, r.db, updateBatchSQL, courseCode, limit)
	}
	return execRowsAffected(ctx, r.db, updateAllSQL, courseCode)
}

func queryCount(ctx context.Context, db *sql.DB, query string, args ...interface{}) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func execRowsAffected(ctx context.Context, db *sql.DB, query string, args ...interface{}) (int64, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return affected, nil
}

const wordAttemptSRSLinksNullTotalSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	WHERE c.code = ? AND ea.mode = 'word_training' AND ea.srs_item_id IS NULL`

const wordAttemptSRSLinksMissingSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	JOIN review_events re ON ea.source_table = 'review_events' AND re.id = CAST(ea.source_pk AS BIGINT)
	JOIN user_cards user_card ON user_card.id = re.user_card_id
	JOIN training_cards tc ON tc.id = user_card.training_card_id
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'word_card' AND li.source_id = CAST(tc.word_card_id AS TEXT)
	JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
	WHERE c.code = ? AND ea.mode = 'word_training' AND ea.srs_item_id IS NULL`

const wordAttemptSRSLinksUpdateSQL = `
	UPDATE exercise_attempts ea
	SET learning_item_id = src.learning_item_id,
		srs_item_id = src.srs_item_id
	FROM (
		SELECT ea.id AS attempt_id, li.id AS learning_item_id, si.id AS srs_item_id
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		JOIN review_events re ON ea.source_table = 'review_events' AND re.id = CAST(ea.source_pk AS BIGINT)
		JOIN user_cards user_card ON user_card.id = re.user_card_id
		JOIN training_cards tc ON tc.id = user_card.training_card_id
		JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'word_card' AND li.source_id = CAST(tc.word_card_id AS TEXT)
		JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
		WHERE c.code = ?
			AND ea.mode = 'word_training'
			AND ea.srs_item_id IS NULL
		LIMIT ?
	) src
	WHERE ea.id = src.attempt_id`

const wordAttemptSRSLinksUpdateAllSQL = `
	UPDATE exercise_attempts ea
	SET learning_item_id = src.learning_item_id,
		srs_item_id = src.srs_item_id
	FROM (
		SELECT ea.id AS attempt_id, li.id AS learning_item_id, si.id AS srs_item_id
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		JOIN review_events re ON ea.source_table = 'review_events' AND re.id = CAST(ea.source_pk AS BIGINT)
		JOIN user_cards user_card ON user_card.id = re.user_card_id
		JOIN training_cards tc ON tc.id = user_card.training_card_id
		JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'word_card' AND li.source_id = CAST(tc.word_card_id AS TEXT)
		JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
		WHERE c.code = ?
			AND ea.mode = 'word_training'
			AND ea.srs_item_id IS NULL
	) src
	WHERE ea.id = src.attempt_id`

const grammarAttemptSRSLinksNullTotalSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	WHERE c.code = ? AND ea.mode = 'grammar_training' AND ea.srs_item_id IS NULL`

const grammarAttemptSRSLinksMissingSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	JOIN grammar_attempts ga ON ea.source_table = 'grammar_attempts' AND ga.id = CAST(ea.source_pk AS BIGINT)
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'grammar_theory_block' AND li.source_id = ga.chapter_id || ':' || ga.theory_block_id
	JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
	WHERE c.code = ? AND ea.mode = 'grammar_training' AND ea.srs_item_id IS NULL`

const grammarAttemptSRSLinksUpdateSQL = `
	UPDATE exercise_attempts ea
	SET learning_item_id = src.learning_item_id,
		srs_item_id = src.srs_item_id
	FROM (
		SELECT ea.id AS attempt_id, li.id AS learning_item_id, si.id AS srs_item_id
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		JOIN grammar_attempts ga ON ea.source_table = 'grammar_attempts' AND ga.id = CAST(ea.source_pk AS BIGINT)
		JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'grammar_theory_block' AND li.source_id = ga.chapter_id || ':' || ga.theory_block_id
		JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
		WHERE c.code = ?
			AND ea.mode = 'grammar_training'
			AND ea.srs_item_id IS NULL
		LIMIT ?
	) src
	WHERE ea.id = src.attempt_id`

const grammarAttemptSRSLinksUpdateAllSQL = `
	UPDATE exercise_attempts ea
	SET learning_item_id = src.learning_item_id,
		srs_item_id = src.srs_item_id
	FROM (
		SELECT ea.id AS attempt_id, li.id AS learning_item_id, si.id AS srs_item_id
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		JOIN grammar_attempts ga ON ea.source_table = 'grammar_attempts' AND ga.id = CAST(ea.source_pk AS BIGINT)
		JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'grammar_theory_block' AND li.source_id = ga.chapter_id || ':' || ga.theory_block_id
		JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
		WHERE c.code = ?
			AND ea.mode = 'grammar_training'
			AND ea.srs_item_id IS NULL
	) src
	WHERE ea.id = src.attempt_id`

const grammarLearningEventsItemSyncSQL = `
	UPDATE learning_events le
	SET learning_item_id = ea.learning_item_id
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	WHERE le.exercise_attempt_id = ea.id
		AND c.code = ?
		AND ea.mode = 'grammar_training'
		AND ea.learning_item_id IS NOT NULL
		AND le.learning_item_id IS DISTINCT FROM ea.learning_item_id`
