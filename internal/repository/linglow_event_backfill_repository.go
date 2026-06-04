package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
)

const (
	LinglowBackfillSourceGrammarTests    = "grammar-tests"
	LinglowBackfillSourceGrammarTraining = "grammar-training"
	LinglowBackfillSourceWordReviews     = "word-reviews"
	LinglowBackfillSourceAll             = "all"
)

// LinglowEventBackfillRepository audits and backfills legacy attempts into Linglow event tables.
type LinglowEventBackfillRepository struct {
	db     *sql.DB
	events *LinglowEventRepository
}

type LinglowEventBackfillOptions struct {
	Commit bool
	Source string
	Limit  int
}

type LinglowEventBackfillSummary struct {
	Source        string
	LegacyTotal   int64
	MirroredTotal int64
	Missing       int64
	Processed     int64
	Inserted      int64
	Failed        int64
}

func NewLinglowEventBackfillRepository(db *sql.DB) *LinglowEventBackfillRepository {
	return &LinglowEventBackfillRepository{
		db:     db,
		events: NewLinglowEventRepository(db),
	}
}

func (r *LinglowEventBackfillRepository) Backfill(ctx context.Context, lc config.LearningConfig, opts LinglowEventBackfillOptions) ([]LinglowEventBackfillSummary, error) {
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return nil, fmt.Errorf("course code is empty")
	}
	sources, err := normalizeBackfillSources(opts.Source)
	if err != nil {
		return nil, err
	}

	summaries := make([]LinglowEventBackfillSummary, 0, len(sources))
	for _, source := range sources {
		summary, err := r.auditSource(ctx, courseCode, source)
		if err != nil {
			return nil, err
		}
		if opts.Commit {
			if err := r.backfillSource(ctx, lc, source, opts.Limit, &summary); err != nil {
				return nil, err
			}
			refreshed, err := r.auditSource(ctx, courseCode, source)
			if err != nil {
				return nil, err
			}
			summary.LegacyTotal = refreshed.LegacyTotal
			summary.MirroredTotal = refreshed.MirroredTotal
			summary.Missing = refreshed.Missing
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func normalizeBackfillSources(source string) ([]string, error) {
	switch strings.TrimSpace(source) {
	case "", LinglowBackfillSourceAll:
		return []string{
			LinglowBackfillSourceGrammarTests,
			LinglowBackfillSourceGrammarTraining,
			LinglowBackfillSourceWordReviews,
		}, nil
	case LinglowBackfillSourceGrammarTests:
		return []string{LinglowBackfillSourceGrammarTests}, nil
	case LinglowBackfillSourceGrammarTraining:
		return []string{LinglowBackfillSourceGrammarTraining}, nil
	case LinglowBackfillSourceWordReviews:
		return []string{LinglowBackfillSourceWordReviews}, nil
	default:
		return nil, fmt.Errorf("unknown source %q", source)
	}
}

func sourceTable(source string) string {
	switch source {
	case LinglowBackfillSourceGrammarTests:
		return "grammar_test_attempts"
	case LinglowBackfillSourceGrammarTraining:
		return "grammar_attempts"
	case LinglowBackfillSourceWordReviews:
		return "review_events"
	default:
		return ""
	}
}

func (r *LinglowEventBackfillRepository) auditSource(ctx context.Context, courseCode, source string) (LinglowEventBackfillSummary, error) {
	table := sourceTable(source)
	if table == "" {
		return LinglowEventBackfillSummary{}, fmt.Errorf("unknown source %q", source)
	}

	summary := LinglowEventBackfillSummary{Source: source}
	legacySQL := legacyCountSQL(source)
	if err := r.db.QueryRowContext(ctx, legacySQL, courseCode).Scan(&summary.LegacyTotal); err != nil {
		return summary, fmt.Errorf("count legacy %s: %w", source, err)
	}
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM exercise_attempts ea
		JOIN user_courses uc ON uc.id = ea.user_course_id
		JOIN courses c ON c.id = uc.course_id
		WHERE c.code = ? AND ea.source_table = ?
	`, courseCode, table).Scan(&summary.MirroredTotal); err != nil {
		return summary, fmt.Errorf("count mirrored %s: %w", source, err)
	}
	missingSQL := missingCountSQL(source)
	if err := r.db.QueryRowContext(ctx, missingSQL, courseCode).Scan(&summary.Missing); err != nil {
		return summary, fmt.Errorf("count missing %s: %w", source, err)
	}
	return summary, nil
}

func legacyCountSQL(source string) string {
	switch source {
	case LinglowBackfillSourceGrammarTests:
		return `
			SELECT COUNT(*)
			FROM grammar_test_attempts gta
			JOIN user_courses uc ON uc.user_id = gta.user_id
			JOIN courses c ON c.id = uc.course_id
			WHERE c.code = ?`
	case LinglowBackfillSourceGrammarTraining:
		return `
			SELECT COUNT(*)
			FROM grammar_attempts ga
			JOIN user_courses uc ON uc.user_id = ga.user_id
			JOIN courses c ON c.id = uc.course_id
			WHERE c.code = ?`
	case LinglowBackfillSourceWordReviews:
		return `
			SELECT COUNT(*)
			FROM review_events re
			JOIN user_courses uc ON uc.user_id = re.user_id
			JOIN courses c ON c.id = uc.course_id
			WHERE c.code = ?`
	default:
		return ""
	}
}

func missingCountSQL(source string) string {
	return strings.Replace(missingRowsSQL(source), "SELECT *", "SELECT COUNT(*)", 1)
}

func missingRowsSQL(source string) string {
	switch source {
	case LinglowBackfillSourceGrammarTests:
		return `
			SELECT *
			FROM (
				SELECT gta.id, gta.user_id, gta.scope_type, gta.scope_id, gta.score, gta.passed, gta.total_questions,
					gta.answers_json, gta.results_json, gta.client_attempt_id, COALESCE(gta.finished_at, gta.started_at) AS answered_at
				FROM grammar_test_attempts gta
				JOIN user_courses uc ON uc.user_id = gta.user_id
				JOIN courses c ON c.id = uc.course_id
				WHERE c.code = ?
					AND NOT EXISTS (
						SELECT 1 FROM exercise_attempts ea
						WHERE ea.user_course_id = uc.id
							AND ea.source_table = 'grammar_test_attempts'
							AND ea.source_pk = CAST(gta.id AS TEXT)
					)
				ORDER BY gta.id
			) rows`
	case LinglowBackfillSourceGrammarTraining:
		return `
			SELECT *
			FROM (
				SELECT ga.id, ga.user_id, ga.chapter_id, ga.theory_block_id, ga.concept_id, ga.question_id,
					ga.is_correct, ga.answer_payload_json, ga.correct_payload_json, ga.client_attempt_id, ga.answered_at
				FROM grammar_attempts ga
				JOIN user_courses uc ON uc.user_id = ga.user_id
				JOIN courses c ON c.id = uc.course_id
				WHERE c.code = ?
					AND NOT EXISTS (
						SELECT 1 FROM exercise_attempts ea
						WHERE ea.user_course_id = uc.id
							AND ea.source_table = 'grammar_attempts'
							AND ea.source_pk = CAST(ga.id AS TEXT)
					)
				ORDER BY ga.id
			) rows`
	case LinglowBackfillSourceWordReviews:
		return `
			SELECT *
			FROM (
				SELECT re.id, re.user_id, re.user_card_id, re.client_attempt_id, re.direction, COALESCE(re.is_correct, 0) AS is_correct,
					COALESCE(re.quality, 0) AS quality, re.options_json, re.chosen_option, re.metrics_json,
					re.srs_before_json, re.srs_after_json, re.answered_at
				FROM review_events re
				JOIN user_courses uc ON uc.user_id = re.user_id
				JOIN courses c ON c.id = uc.course_id
				WHERE c.code = ?
					AND NOT EXISTS (
						SELECT 1 FROM exercise_attempts ea
						WHERE ea.user_course_id = uc.id
							AND ea.source_table = 'review_events'
							AND ea.source_pk = CAST(re.id AS TEXT)
					)
				ORDER BY re.id
			) rows`
	default:
		return ""
	}
}

func (r *LinglowEventBackfillRepository) backfillSource(ctx context.Context, lc config.LearningConfig, source string, limit int, summary *LinglowEventBackfillSummary) error {
	query := missingRowsSQL(source)
	args := []interface{}{CourseCodeForLearning(lc)}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query missing %s: %w", source, err)
	}
	defer rows.Close()

	for rows.Next() {
		summary.Processed++
		if err := r.recordMissingRow(ctx, lc, source, rows); err != nil {
			summary.Failed++
			return fmt.Errorf("backfill %s row %d: %w", source, summary.Processed, err)
		}
		summary.Inserted++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate missing %s: %w", source, err)
	}
	return nil
}

func (r *LinglowEventBackfillRepository) recordMissingRow(ctx context.Context, lc config.LearningConfig, source string, rows *sql.Rows) error {
	switch source {
	case LinglowBackfillSourceGrammarTests:
		var in GrammarTestEventInput
		var passed int
		var answers, results, client sql.NullString
		var answeredAt sql.NullTime
		if err := rows.Scan(&in.AttemptID, &in.UserID, &in.ScopeType, &in.ScopeID, &in.Score, &passed, &in.TotalQuestions, &answers, &results, &client, &answeredAt); err != nil {
			return err
		}
		in.Passed = passed != 0
		in.AnswersJSON = nullString(answers)
		in.ResultsJSON = nullString(results)
		in.ClientAttemptID = nullString(client)
		in.AnsweredAt = nullTime(answeredAt)
		_, err := r.events.RecordGrammarTestAttempt(ctx, lc, in)
		return err
	case LinglowBackfillSourceGrammarTraining:
		var in GrammarTrainingEventInput
		var concept, answer, correct, client sql.NullString
		var answeredAt sql.NullTime
		if err := rows.Scan(&in.AttemptID, &in.UserID, &in.ChapterID, &in.TheoryBlockID, &concept, &in.QuestionID, &in.IsCorrect, &answer, &correct, &client, &answeredAt); err != nil {
			return err
		}
		in.ConceptID = nullString(concept)
		in.AnswerJSON = nullString(answer)
		in.CorrectJSON = nullString(correct)
		in.ClientAttemptID = nullString(client)
		in.AnsweredAt = nullTime(answeredAt)
		_, err := r.events.RecordGrammarTrainingAttempt(ctx, lc, in)
		return err
	case LinglowBackfillSourceWordReviews:
		var in WordReviewEventInput
		var isCorrect int
		var client, options, chosen, metrics, before, after sql.NullString
		var answeredAt sql.NullTime
		if err := rows.Scan(&in.ReviewEventID, &in.UserID, &in.UserCardID, &client, &in.Direction, &isCorrect, &in.Quality, &options, &chosen, &metrics, &before, &after, &answeredAt); err != nil {
			return err
		}
		in.IsCorrect = isCorrect != 0
		in.ClientAttemptID = nullString(client)
		in.OptionsJSON = nullString(options)
		in.ChosenOption = nullString(chosen)
		in.MetricsJSON = nullString(metrics)
		in.SRSBeforeJSON = nullString(before)
		in.SRSAfterJSON = nullString(after)
		in.AnsweredAt = nullTime(answeredAt)
		_, err := r.events.RecordWordReviewEvent(ctx, lc, in)
		return err
	default:
		return fmt.Errorf("unknown source %q", source)
	}
}

func nullString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func nullTime(v sql.NullTime) time.Time {
	if !v.Valid {
		return time.Time{}
	}
	return v.Time
}
