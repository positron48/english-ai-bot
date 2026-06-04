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
	word, err := r.audit(ctx, courseCode, "word-training", wordAttemptSRSLinksMissingSQL, wordAttemptSRSLinksUnmappedSQL)
	if err != nil {
		return nil, err
	}
	if opts.Commit {
		updated, err := execRowsAffected(ctx, r.db, wordAttemptSRSLinksUpdateSQL, courseCode)
		if err != nil {
			return nil, fmt.Errorf("update word attempt srs links: %w", err)
		}
		word.Updated = updated
		word.Missing, _ = queryCount(ctx, r.db, wordAttemptSRSLinksMissingSQL, courseCode)
	}
	summaries = append(summaries, word)

	grammar, err := r.audit(ctx, courseCode, "grammar-training", grammarAttemptSRSLinksMissingSQL, grammarAttemptSRSLinksUnmappedSQL)
	if err != nil {
		return nil, err
	}
	if opts.Commit {
		updated, err := execRowsAffected(ctx, r.db, grammarAttemptSRSLinksUpdateSQL, courseCode)
		if err != nil {
			return nil, fmt.Errorf("update grammar attempt srs links: %w", err)
		}
		if _, err := r.db.ExecContext(ctx, grammarLearningEventsItemSyncSQL, courseCode); err != nil {
			return nil, fmt.Errorf("sync grammar learning event items: %w", err)
		}
		grammar.Updated = updated
		grammar.Missing, _ = queryCount(ctx, r.db, grammarAttemptSRSLinksMissingSQL, courseCode)
	}
	summaries = append(summaries, grammar)
	return summaries, nil
}

func (r *LinglowAttemptSRSLinkBackfillRepository) audit(ctx context.Context, courseCode, source, missingSQL, unmappedSQL string) (LinglowAttemptSRSLinkBackfillSummary, error) {
	missing, err := queryCount(ctx, r.db, missingSQL, courseCode)
	if err != nil {
		return LinglowAttemptSRSLinkBackfillSummary{}, fmt.Errorf("count missing %s srs links: %w", source, err)
	}
	unmapped, err := queryCount(ctx, r.db, unmappedSQL, courseCode)
	if err != nil {
		return LinglowAttemptSRSLinkBackfillSummary{}, fmt.Errorf("count unmapped %s srs links: %w", source, err)
	}
	return LinglowAttemptSRSLinkBackfillSummary{Source: source, Missing: missing, UnmappedTotal: unmapped}, nil
}

func queryCount(ctx context.Context, db *sql.DB, query, courseCode string) (int64, error) {
	var count int64
	err := db.QueryRowContext(ctx, query, courseCode).Scan(&count)
	return count, err
}

func execRowsAffected(ctx context.Context, db *sql.DB, query, courseCode string) (int64, error) {
	result, err := db.ExecContext(ctx, query, courseCode)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return affected, nil
}

const wordAttemptSRSLinksMissingSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	JOIN review_events re ON ea.source_table = 'review_events' AND ea.source_pk = CAST(re.id AS TEXT)
	JOIN user_cards user_card ON user_card.id = re.user_card_id
	JOIN training_cards tc ON tc.id = user_card.training_card_id
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'word_card' AND li.source_id = CAST(tc.word_card_id AS TEXT)
	JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
	WHERE c.code = ? AND ea.mode = 'word_training' AND ea.srs_item_id IS NULL`

const wordAttemptSRSLinksUnmappedSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	WHERE c.code = ? AND ea.mode = 'word_training' AND ea.srs_item_id IS NULL
		AND NOT EXISTS (
			SELECT 1
			FROM review_events re
			JOIN user_cards user_card ON user_card.id = re.user_card_id
			JOIN training_cards tc ON tc.id = user_card.training_card_id
			JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'word_card' AND li.source_id = CAST(tc.word_card_id AS TEXT)
			JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
			WHERE ea.source_table = 'review_events' AND ea.source_pk = CAST(re.id AS TEXT)
		)`

const wordAttemptSRSLinksUpdateSQL = `
	UPDATE exercise_attempts ea
	SET learning_item_id = li.id,
		srs_item_id = si.id
	FROM user_courses uc
	JOIN courses c ON c.id = uc.course_id
	JOIN review_events re ON TRUE
	JOIN user_cards user_card ON user_card.id = re.user_card_id
	JOIN training_cards tc ON tc.id = user_card.training_card_id
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'word_card' AND li.source_id = CAST(tc.word_card_id AS TEXT)
	JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
	WHERE ea.user_course_id = uc.id
		AND c.code = ?
		AND ea.mode = 'word_training'
		AND ea.source_table = 'review_events'
		AND ea.source_pk = CAST(re.id AS TEXT)
		AND ea.srs_item_id IS NULL`

const grammarAttemptSRSLinksMissingSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	JOIN grammar_attempts ga ON ea.source_table = 'grammar_attempts' AND ea.source_pk = CAST(ga.id AS TEXT)
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'grammar_theory_block' AND li.source_id = ga.chapter_id || ':' || ga.theory_block_id
	JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
	WHERE c.code = ? AND ea.mode = 'grammar_training' AND ea.srs_item_id IS NULL`

const grammarAttemptSRSLinksUnmappedSQL = `
	SELECT COUNT(*)
	FROM exercise_attempts ea
	JOIN user_courses uc ON uc.id = ea.user_course_id
	JOIN courses c ON c.id = uc.course_id
	WHERE c.code = ? AND ea.mode = 'grammar_training' AND ea.srs_item_id IS NULL
		AND NOT EXISTS (
			SELECT 1
			FROM grammar_attempts ga
			JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'grammar_theory_block' AND li.source_id = ga.chapter_id || ':' || ga.theory_block_id
			JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
			WHERE ea.source_table = 'grammar_attempts' AND ea.source_pk = CAST(ga.id AS TEXT)
		)`

const grammarAttemptSRSLinksUpdateSQL = `
	UPDATE exercise_attempts ea
	SET learning_item_id = li.id,
		srs_item_id = si.id
	FROM user_courses uc
	JOIN courses c ON c.id = uc.course_id
	JOIN grammar_attempts ga ON TRUE
	JOIN learning_items li ON li.course_id = c.id AND li.source_kind = 'grammar_theory_block' AND li.source_id = ga.chapter_id || ':' || ga.theory_block_id
	JOIN srs_items si ON si.user_course_id = uc.id AND si.learning_item_id = li.id
	WHERE ea.user_course_id = uc.id
		AND c.code = ?
		AND ea.mode = 'grammar_training'
		AND ea.source_table = 'grammar_attempts'
		AND ea.source_pk = CAST(ga.id AS TEXT)
		AND ea.srs_item_id IS NULL`

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
