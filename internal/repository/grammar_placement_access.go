package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ForCourse returns an independent view; services for different languages must
// never mutate the shared repository's selected course.
func (r *GrammarAttemptRepository) ForCourse(courseCode string) *GrammarAttemptRepository {
	clone := *r
	clone.courseCode = strings.ToLower(strings.TrimSpace(courseCode))
	return &clone
}

func (r *GrammarAttemptRepository) CourseCode() string { return r.courseCode }

// legacyPlacementSections deliberately does not attribute an unidentifiable
// old result (including an empty admin assignment) to either language.
func legacyPlacementSections(opened []string, courseCode string) []string {
	prefix := strings.SplitN(courseCode, "_", 2)[0] + ".grammar."
	filtered := make([]string, 0, len(opened))
	for _, id := range opened {
		if strings.HasPrefix(id, prefix) {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

func (r *GrammarAttemptRepository) getCoursePlacement(ctx context.Context, userID int64) (*PlacementTestResult, error) {
	var result PlacementTestResult
	var openedJSON string
	var completedAt sql.NullString
	var cleared bool
	err := r.db.QueryRowContext(ctx, `
		SELECT uc.user_id, a.user_course_id, a.score, a.total_questions,
		       a.opened_sections_json, a.completed_at, a.admin_override, a.source, a.cleared
		FROM grammar_placement_access a
		JOIN user_courses uc ON uc.id = a.user_course_id
		JOIN courses c ON c.id = uc.course_id
		WHERE uc.user_id = ? AND c.code = ?`, userID, r.courseCode).Scan(
		&result.UserID, &result.UserCourseID, &result.Score, &result.TotalQuestions,
		&openedJSON, &completedAt, &result.AdminOverride, &result.Source, &cleared)
	if err == sql.ErrNoRows {
		legacy, legacyErr := r.ForCourse("").GetPlacementTestResult(userID)
		var syntaxErr *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		if errors.As(legacyErr, &syntaxErr) || errors.As(legacyErr, &typeErr) {
			// An unreadable legacy row cannot establish course access. Keep it
			// in the archive without breaking user lists or a new diagnostic.
			return nil, nil
		}
		if legacyErr != nil || legacy == nil {
			return nil, legacyErr
		}
		legacy.OpenedSections = legacyPlacementSections(legacy.OpenedSections, r.courseCode)
		if len(legacy.OpenedSections) == 0 {
			return nil, nil
		}
		return legacy, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get course placement: %w", err)
	}
	if cleared {
		return nil, nil
	}
	if err := json.Unmarshal([]byte(openedJSON), &result.OpenedSections); err != nil {
		return nil, fmt.Errorf("parse course placement: %w", err)
	}
	if completedAt.Valid {
		result.CompletedAt = parseTimestampFlex(completedAt.String)
	}
	return &result, nil
}

func (r *GrammarAttemptRepository) saveCoursePlacement(userID int64, score, total int, opened []string, source string, cleared bool) error {
	ctx := context.Background()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userCourseID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO user_courses (user_id, course_id)
		SELECT ?, id FROM courses WHERE code = ?
		ON CONFLICT (user_id, course_id) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING id`, userID, r.courseCode).Scan(&userCourseID)
	if err != nil {
		return fmt.Errorf("resolve placement course %s: %w", r.courseCode, err)
	}
	if source == "admin" {
		if opened == nil {
			opened = []string{}
		}
		raw, err := json.Marshal(opened)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO grammar_placement_access
			(user_course_id, score, total_questions, opened_sections_json, completed_at, source, admin_override, cleared)
			VALUES (?, ?, ?, ?::jsonb, CURRENT_TIMESTAMP, 'admin', true, ?)
			ON CONFLICT (user_course_id) DO UPDATE SET
			 score = EXCLUDED.score, total_questions = EXCLUDED.total_questions,
			 opened_sections_json = EXCLUDED.opened_sections_json, completed_at = CURRENT_TIMESTAMP,
			 source = 'admin', admin_override = true, cleared = EXCLUDED.cleared`,
			userCourseID, score, total, string(raw), cleared)
		if err != nil {
			return fmt.Errorf("save admin course placement: %w", err)
		}
	} else if err := savePlacementAccessTx(ctx, tx, userCourseID, score, total, opened, source); err != nil {
		return err
	}
	return tx.Commit()
}

// SaveDiagnosticPlacementAccessTx joins a diagnostic's completion transaction.
// Scores describe the latest attempt, while section access only grows. Explicit
// admin resets/assignments are the only operations allowed to narrow access.
func SaveDiagnosticPlacementAccessTx(ctx context.Context, tx *sql.Tx, userCourseID int64, score, total int, opened []string) error {
	return savePlacementAccessTx(ctx, tx, userCourseID, score, total, opened, "diagnostic")
}

func savePlacementAccessTx(ctx context.Context, tx *sql.Tx, userCourseID int64, score, total int, opened []string, source string) error {
	// Preserve attributable pre-migration access before the first new attempt.
	if err := importLegacyPlacementAccessTx(ctx, tx, userCourseID); err != nil {
		return err
	}
	if opened == nil {
		opened = []string{}
	}
	raw, err := json.Marshal(opened)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO grammar_placement_access
		(user_course_id, score, total_questions, opened_sections_json, completed_at, source, admin_override, cleared)
		VALUES (?, ?, ?, ?::jsonb, CURRENT_TIMESTAMP, ?, false, false)
		ON CONFLICT (user_course_id) DO UPDATE SET
		 score = EXCLUDED.score, total_questions = EXCLUDED.total_questions,
		 opened_sections_json = (
		   SELECT COALESCE(jsonb_agg(section_id ORDER BY section_id), '[]'::jsonb)
		   FROM (SELECT DISTINCT jsonb_array_elements_text(
		     grammar_placement_access.opened_sections_json || EXCLUDED.opened_sections_json
		   ) AS section_id) combined
		 ),
		 completed_at = CURRENT_TIMESTAMP, source = EXCLUDED.source,
		 admin_override = grammar_placement_access.admin_override AND NOT grammar_placement_access.cleared,
		 cleared = false`, userCourseID, score, total, string(raw), source)
	if err != nil {
		return fmt.Errorf("save diagnostic course access: %w", err)
	}
	return nil
}

func importLegacyPlacementAccessTx(ctx context.Context, tx *sql.Tx, userCourseID int64) error {
	var courseCode, raw string
	var score, total int
	var completedAt sql.NullString
	var adminOverride bool
	err := tx.QueryRowContext(ctx, `
		SELECT c.code, old.score, old.total_questions, old.opened_sections_json,
		       old.completed_at, COALESCE(old.admin_override, false)
		FROM user_courses uc JOIN courses c ON c.id = uc.course_id
		JOIN grammar_placement_test old ON old.user_id = uc.user_id
		WHERE uc.id = ?
		AND NOT EXISTS (SELECT 1 FROM grammar_placement_access a WHERE a.user_course_id = uc.id)`,
		userCourseID).Scan(&courseCode, &score, &total, &raw, &completedAt, &adminOverride)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read legacy placement access: %w", err)
	}
	var opened []string
	if err := json.Unmarshal([]byte(raw), &opened); err != nil {
		return nil // Unattributable legacy data remains in the archive.
	}
	opened = legacyPlacementSections(opened, courseCode)
	if len(opened) == 0 {
		return nil
	}
	encoded, _ := json.Marshal(opened)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO grammar_placement_access
		(user_course_id, score, total_questions, opened_sections_json, completed_at, source, admin_override, cleared)
		VALUES (?, ?, ?, ?::jsonb, COALESCE(?::timestamptz, CURRENT_TIMESTAMP), 'legacy', ?, false)
		ON CONFLICT (user_course_id) DO NOTHING`, userCourseID, score, total, string(encoded), completedAt, adminOverride)
	if err != nil {
		return fmt.Errorf("preserve legacy placement access: %w", err)
	}
	return nil
}
