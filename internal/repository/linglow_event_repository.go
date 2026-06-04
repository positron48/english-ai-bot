package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/config"
)

// LinglowEventRepository writes dual-write progress data into Linglow v2 tables.
type LinglowEventRepository struct {
	db *sql.DB
}

func NewLinglowEventRepository(db *sql.DB) *LinglowEventRepository {
	return &LinglowEventRepository{db: db}
}

// GrammarTestEventInput describes a legacy grammar_test_attempts row that should be mirrored.
type GrammarTestEventInput struct {
	UserID          int64
	AttemptID       int64
	ScopeType       string
	ScopeID         string
	Score           int
	Passed          bool
	TotalQuestions  int
	AnswersJSON     string
	ResultsJSON     string
	ClientAttemptID string
	AnsweredAt      time.Time
}

// GrammarTrainingEventInput describes a legacy grammar_attempts row that should be mirrored.
type GrammarTrainingEventInput struct {
	UserID          int64
	AttemptID       int64
	ChapterID       string
	TheoryBlockID   string
	ConceptID       string
	QuestionID      string
	IsCorrect       bool
	AnswerJSON      string
	CorrectJSON     string
	ClientAttemptID string
	AnsweredAt      time.Time
}

// WordReviewEventInput describes a legacy review_events row that should be mirrored.
type WordReviewEventInput struct {
	UserID          int64
	ReviewEventID   int64
	UserCardID      int64
	ClientAttemptID string
	Direction       string
	IsCorrect       bool
	Quality         int
	OptionsJSON     string
	ChosenOption    string
	MetricsJSON     string
	SRSBeforeJSON   string
	SRSAfterJSON    string
	AnsweredAt      time.Time
}

// RecordGrammarTestAttempt mirrors a legacy grammar test attempt into exercise_attempts and learning_events.
func (r *LinglowEventRepository) RecordGrammarTestAttempt(ctx context.Context, lc config.LearningConfig, input GrammarTestEventInput) (int64, error) {
	if input.UserID == 0 {
		return 0, fmt.Errorf("user id is empty")
	}
	if input.AttemptID == 0 {
		return 0, fmt.Errorf("attempt id is empty")
	}
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return 0, fmt.Errorf("course code is empty")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin linglow event write: %w", err)
	}
	defer tx.Rollback()

	var userCourseID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT uc.id
		FROM user_courses uc
		JOIN courses c ON c.id = uc.course_id
		WHERE uc.user_id = ? AND c.code = ?
	`, input.UserID, courseCode).Scan(&userCourseID); err != nil {
		return 0, fmt.Errorf("get user course: %w", err)
	}

	sourcePK := fmt.Sprintf("%d", input.AttemptID)
	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM exercise_attempts
		WHERE user_course_id = ? AND source_table = 'grammar_test_attempts' AND source_pk = ?
	`, userCourseID, sourcePK).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("check existing exercise attempt: %w", err)
	}
	if err == nil {
		return existingID, tx.Commit()
	}

	var learningItemID interface{}
	if strings.TrimSpace(input.ScopeType) == "chapter" && strings.TrimSpace(input.ScopeID) != "" {
		var id int64
		err := tx.QueryRowContext(ctx, `
			SELECT li.id
			FROM learning_items li
			JOIN courses c ON c.id = li.course_id
			WHERE c.code = ? AND li.source_kind = 'grammar_chapter' AND li.source_id = ?
		`, courseCode, strings.TrimSpace(input.ScopeID)).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("lookup grammar learning item: %w", err)
		}
		if err == nil {
			learningItemID = id
		}
	}

	answerJSON := normalizeJSON(input.AnswersJSON)
	resultJSON := normalizeJSON(input.ResultsJSON)
	promptJSON, _ := json.Marshal(map[string]interface{}{
		"scope_type": input.ScopeType,
		"scope_id":   input.ScopeID,
	})
	eventJSON, _ := json.Marshal(map[string]interface{}{
		"scope_type":      input.ScopeType,
		"scope_id":        input.ScopeID,
		"score":           input.Score,
		"passed":          input.Passed,
		"total_questions": input.TotalQuestions,
	})

	var clientAttemptID interface{}
	if strings.TrimSpace(input.ClientAttemptID) != "" {
		clientAttemptID = strings.TrimSpace(input.ClientAttemptID)
	}
	answeredAt := input.AnsweredAt
	if answeredAt.IsZero() {
		answeredAt = time.Now()
	}

	var exerciseID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO exercise_attempts (
			user_course_id, learning_item_id, mode, client_attempt_id,
			started_at, answered_at, is_correct, score,
			prompt_json, answer_json, result_json, source_table, source_pk
		)
		VALUES (?, ?, 'grammar_test', ?, ?, ?, ?, ?, CAST(? AS jsonb), CAST(? AS jsonb), CAST(? AS jsonb), 'grammar_test_attempts', ?)
		RETURNING id
	`, userCourseID, learningItemID, clientAttemptID, answeredAt, answeredAt, input.Passed, input.Score, string(promptJSON), answerJSON, resultJSON, sourcePK).Scan(&exerciseID)
	if err != nil {
		return 0, fmt.Errorf("insert exercise attempt: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO learning_events (
			user_course_id, learning_item_id, exercise_attempt_id,
			event_type, event_time, mode, source_table, source_pk, event_json
		)
		VALUES (?, ?, ?, 'grammar_test_submitted', ?, 'grammar_test', 'grammar_test_attempts', ?, CAST(? AS jsonb))
	`, userCourseID, learningItemID, exerciseID, answeredAt, sourcePK, string(eventJSON)); err != nil {
		return 0, fmt.Errorf("insert learning event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit linglow event write: %w", err)
	}
	return exerciseID, nil
}

func normalizeJSON(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "{}"
	}
	return raw
}

func normalizeJSONArray(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "[]"
	}
	return raw
}

// RecordGrammarTrainingAttempt mirrors a legacy grammar_attempts SRS row into exercise_attempts and learning_events.
func (r *LinglowEventRepository) RecordGrammarTrainingAttempt(ctx context.Context, lc config.LearningConfig, input GrammarTrainingEventInput) (int64, error) {
	if input.UserID == 0 {
		return 0, fmt.Errorf("user id is empty")
	}
	if input.AttemptID == 0 {
		return 0, fmt.Errorf("attempt id is empty")
	}
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return 0, fmt.Errorf("course code is empty")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin linglow grammar training write: %w", err)
	}
	defer tx.Rollback()

	var userCourseID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT uc.id
		FROM user_courses uc
		JOIN courses c ON c.id = uc.course_id
		WHERE uc.user_id = ? AND c.code = ?
	`, input.UserID, courseCode).Scan(&userCourseID); err != nil {
		return 0, fmt.Errorf("get user course: %w", err)
	}

	sourcePK := fmt.Sprintf("%d", input.AttemptID)
	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM exercise_attempts
		WHERE user_course_id = ? AND source_table = 'grammar_attempts' AND source_pk = ?
	`, userCourseID, sourcePK).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("check existing grammar training exercise attempt: %w", err)
	}
	if err == nil {
		return existingID, tx.Commit()
	}

	var learningItemID interface{}
	if strings.TrimSpace(input.ChapterID) != "" && strings.TrimSpace(input.TheoryBlockID) != "" {
		var id int64
		err := tx.QueryRowContext(ctx, `
			SELECT li.id
			FROM learning_items li
			JOIN courses c ON c.id = li.course_id
			WHERE c.code = ? AND li.source_kind = 'grammar_theory_block' AND li.source_id = ?
		`, courseCode, strings.TrimSpace(input.ChapterID)+":"+strings.TrimSpace(input.TheoryBlockID)).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("lookup grammar theory block learning item: %w", err)
		}
		if err == nil {
			learningItemID = id
		}
	}
	if learningItemID == nil && strings.TrimSpace(input.ChapterID) != "" {
		var id int64
		err := tx.QueryRowContext(ctx, `
			SELECT li.id
			FROM learning_items li
			JOIN courses c ON c.id = li.course_id
			WHERE c.code = ? AND li.source_kind = 'grammar_chapter' AND li.source_id = ?
		`, courseCode, strings.TrimSpace(input.ChapterID)).Scan(&id)
		if err != nil && err != sql.ErrNoRows {
			return 0, fmt.Errorf("lookup grammar chapter learning item: %w", err)
		}
		if err == nil {
			learningItemID = id
		}
	}
	srsItemID := lookupSRSItemID(ctx, tx, userCourseID, learningItemID)

	promptJSON, _ := json.Marshal(map[string]interface{}{
		"chapter_id":      input.ChapterID,
		"theory_block_id": input.TheoryBlockID,
		"concept_id":      input.ConceptID,
		"question_id":     input.QuestionID,
	})
	resultJSON, _ := json.Marshal(map[string]interface{}{
		"is_correct":      input.IsCorrect,
		"correct_payload": json.RawMessage(normalizeJSON(input.CorrectJSON)),
	})
	eventJSON, _ := json.Marshal(map[string]interface{}{
		"chapter_id":      input.ChapterID,
		"theory_block_id": input.TheoryBlockID,
		"concept_id":      input.ConceptID,
		"question_id":     input.QuestionID,
		"is_correct":      input.IsCorrect,
	})

	var clientAttemptID interface{}
	if strings.TrimSpace(input.ClientAttemptID) != "" {
		clientAttemptID = strings.TrimSpace(input.ClientAttemptID)
	}
	answeredAt := input.AnsweredAt
	if answeredAt.IsZero() {
		answeredAt = time.Now()
	}

	var exerciseID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO exercise_attempts (
			user_course_id, learning_item_id, srs_item_id, mode, client_attempt_id,
			started_at, answered_at, is_correct,
			prompt_json, answer_json, result_json, source_table, source_pk
		)
		VALUES (?, ?, ?, 'grammar_training', ?, ?, ?, ?, CAST(? AS jsonb), CAST(? AS jsonb), CAST(? AS jsonb), 'grammar_attempts', ?)
		RETURNING id
	`, userCourseID, learningItemID, srsItemID, clientAttemptID, answeredAt, answeredAt, input.IsCorrect, string(promptJSON), normalizeJSON(input.AnswerJSON), string(resultJSON), sourcePK).Scan(&exerciseID)
	if err != nil {
		return 0, fmt.Errorf("insert grammar training exercise attempt: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO learning_events (
			user_course_id, learning_item_id, exercise_attempt_id,
			event_type, event_time, mode, source_table, source_pk, event_json
		)
		VALUES (?, ?, ?, 'grammar_training_answered', ?, 'grammar_training', 'grammar_attempts', ?, CAST(? AS jsonb))
	`, userCourseID, learningItemID, exerciseID, answeredAt, sourcePK, string(eventJSON)); err != nil {
		return 0, fmt.Errorf("insert grammar training learning event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit linglow grammar training write: %w", err)
	}
	return exerciseID, nil
}

// RecordWordReviewEvent mirrors a legacy word review_events row into exercise_attempts and learning_events.
func (r *LinglowEventRepository) RecordWordReviewEvent(ctx context.Context, lc config.LearningConfig, input WordReviewEventInput) (int64, error) {
	if input.UserID == 0 {
		return 0, fmt.Errorf("user id is empty")
	}
	if input.ReviewEventID == 0 {
		return 0, fmt.Errorf("review event id is empty")
	}
	if input.UserCardID == 0 {
		return 0, fmt.Errorf("user card id is empty")
	}
	courseCode := CourseCodeForLearning(lc)
	if courseCode == "" {
		return 0, fmt.Errorf("course code is empty")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin linglow word review write: %w", err)
	}
	defer tx.Rollback()

	var userCourseID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT uc.id
		FROM user_courses uc
		JOIN courses c ON c.id = uc.course_id
		WHERE uc.user_id = ? AND c.code = ?
	`, input.UserID, courseCode).Scan(&userCourseID); err != nil {
		return 0, fmt.Errorf("get user course: %w", err)
	}

	sourcePK := fmt.Sprintf("%d", input.ReviewEventID)
	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM exercise_attempts
		WHERE user_course_id = ? AND source_table = 'review_events' AND source_pk = ?
	`, userCourseID, sourcePK).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("check existing word review exercise attempt: %w", err)
	}
	if err == nil {
		return existingID, tx.Commit()
	}

	var wordCardID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT tc.word_card_id
		FROM user_cards uc
		JOIN training_cards tc ON tc.id = uc.training_card_id
		WHERE uc.id = ? AND uc.user_id = ?
	`, input.UserCardID, input.UserID).Scan(&wordCardID); err != nil {
		return 0, fmt.Errorf("lookup word card: %w", err)
	}

	var learningItemID interface{}
	var itemID int64
	err = tx.QueryRowContext(ctx, `
		SELECT li.id
		FROM learning_items li
		JOIN courses c ON c.id = li.course_id
		WHERE c.code = ? AND li.source_kind = 'word_card' AND li.source_id = ?
	`, courseCode, fmt.Sprintf("%d", wordCardID)).Scan(&itemID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("lookup word learning item: %w", err)
	}
	if err == nil {
		learningItemID = itemID
	}
	srsItemID := lookupSRSItemID(ctx, tx, userCourseID, learningItemID)

	promptJSON, _ := json.Marshal(map[string]interface{}{
		"user_card_id": input.UserCardID,
		"word_card_id": wordCardID,
		"direction":    input.Direction,
		"options":      json.RawMessage(normalizeJSONArray(input.OptionsJSON)),
	})
	answerJSON, _ := json.Marshal(map[string]interface{}{
		"chosen_option": input.ChosenOption,
	})
	resultJSON, _ := json.Marshal(map[string]interface{}{
		"is_correct": input.IsCorrect,
		"quality":    input.Quality,
		"srs_before": json.RawMessage(normalizeJSON(input.SRSBeforeJSON)),
		"srs_after":  json.RawMessage(normalizeJSON(input.SRSAfterJSON)),
		"metrics":    json.RawMessage(normalizeJSON(input.MetricsJSON)),
	})
	eventJSON, _ := json.Marshal(map[string]interface{}{
		"user_card_id": input.UserCardID,
		"word_card_id": wordCardID,
		"direction":    input.Direction,
		"is_correct":   input.IsCorrect,
		"quality":      input.Quality,
	})

	var clientAttemptID interface{}
	if strings.TrimSpace(input.ClientAttemptID) != "" {
		clientAttemptID = strings.TrimSpace(input.ClientAttemptID)
	}
	answeredAt := input.AnsweredAt
	if answeredAt.IsZero() {
		answeredAt = time.Now()
	}

	var exerciseID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO exercise_attempts (
			user_course_id, learning_item_id, srs_item_id, mode, client_attempt_id,
			started_at, answered_at, is_correct, quality,
			prompt_json, answer_json, result_json, source_table, source_pk
		)
		VALUES (?, ?, ?, 'word_training', ?, ?, ?, ?, ?, CAST(? AS jsonb), CAST(? AS jsonb), CAST(? AS jsonb), 'review_events', ?)
		RETURNING id
	`, userCourseID, learningItemID, srsItemID, clientAttemptID, answeredAt, answeredAt, input.IsCorrect, input.Quality, string(promptJSON), string(answerJSON), string(resultJSON), sourcePK).Scan(&exerciseID)
	if err != nil {
		return 0, fmt.Errorf("insert word review exercise attempt: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO learning_events (
			user_course_id, learning_item_id, exercise_attempt_id,
			event_type, event_time, mode, source_table, source_pk, event_json
		)
		VALUES (?, ?, ?, 'word_training_answered', ?, 'word_training', 'review_events', ?, CAST(? AS jsonb))
	`, userCourseID, learningItemID, exerciseID, answeredAt, sourcePK, string(eventJSON)); err != nil {
		return 0, fmt.Errorf("insert word review learning event: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit linglow word review write: %w", err)
	}
	return exerciseID, nil
}

func lookupSRSItemID(ctx context.Context, tx *sql.Tx, userCourseID int64, learningItemID interface{}) interface{} {
	id, ok := learningItemID.(int64)
	if !ok || id == 0 {
		return nil
	}
	var srsItemID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM srs_items
		WHERE user_course_id = ? AND learning_item_id = ?
	`, userCourseID, id).Scan(&srsItemID); err != nil {
		return nil
	}
	return srsItemID
}
