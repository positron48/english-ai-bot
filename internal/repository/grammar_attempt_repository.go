package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tgbot-skeleton/internal/database"

	"go.uber.org/zap"
)

// GrammarAttemptRepository handles database operations for grammar test attempts
type GrammarAttemptRepository struct {
	db         *sql.DB
	logger     *zap.Logger
	courseCode string
}

// NewGrammarAttemptRepository creates a new grammar attempt repository
func NewGrammarAttemptRepository(db *sql.DB, logger *zap.Logger) *GrammarAttemptRepository {
	return &GrammarAttemptRepository{
		db:     db,
		logger: logger,
	}
}

// TestAttempt represents a grammar test attempt
type TestAttempt struct {
	ID              int64
	UserID          int64
	ScopeType       string // "chapter" or "category"
	ScopeID         string
	StartedAt       time.Time
	FinishedAt      *time.Time
	Score           int
	Passed          bool
	TotalQuestions  int
	AnswersJSON     string
	ResultsJSON     string
	CourseVersion   *string
	ClientAttemptID *string
}

// CreateAttempt creates a new test attempt
func (r *GrammarAttemptRepository) CreateAttempt(attempt *TestAttempt) (int64, error) {
	query := `INSERT INTO grammar_test_attempts 
			  (user_id, scope_type, scope_id, started_at, finished_at, score, passed, total_questions, answers_json, results_json, course_version, client_attempt_id)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var finishedAt interface{}
	if attempt.FinishedAt != nil {
		finishedAt = attempt.FinishedAt.Format("2006-01-02 15:04:05")
	}

	passed := 0
	if attempt.Passed {
		passed = 1
	}

	id, err := database.InsertAndReturnID(r.db, query,
		attempt.UserID,
		attempt.ScopeType,
		attempt.ScopeID,
		attempt.StartedAt.Format("2006-01-02 15:04:05"),
		finishedAt,
		attempt.Score,
		passed,
		attempt.TotalQuestions,
		attempt.AnswersJSON,
		attempt.ResultsJSON,
		attempt.CourseVersion,
		attempt.ClientAttemptID,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create attempt: %w", err)
	}

	return id, nil
}

// HasClientAttempt checks whether an offline/client-originated attempt was already synced.
func (r *GrammarAttemptRepository) HasClientAttempt(userID int64, clientAttemptID string) (bool, error) {
	if clientAttemptID == "" {
		return false, nil
	}
	query := `SELECT COUNT(*) > 0 FROM grammar_test_attempts
			  WHERE user_id = ? AND client_attempt_id = ?`
	var exists bool
	if err := r.db.QueryRow(query, userID, clientAttemptID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check client attempt: %w", err)
	}
	return exists, nil
}

// UpdateProgress updates grammar progress for a chapter
func (r *GrammarAttemptRepository) UpdateProgress(userID int64, chapterID string, score int, passed bool) error {
	query := `INSERT INTO grammar_progress (user_id, chapter_id, best_score, passed_at, last_attempt_at)
			  VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
			  ON CONFLICT(user_id, chapter_id) DO UPDATE SET
			  	best_score = CASE
			  		WHEN excluded.best_score > grammar_progress.best_score THEN excluded.best_score
			  		ELSE grammar_progress.best_score
			  	END,
			  	passed_at = CASE WHEN excluded.best_score >= 50 THEN COALESCE(grammar_progress.passed_at, CURRENT_TIMESTAMP) ELSE grammar_progress.passed_at END,
			  	last_attempt_at = CURRENT_TIMESTAMP`

	var passedAt interface{}
	if passed && score > 50 {
		passedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	_, err := r.db.Exec(query, userID, chapterID, score, passedAt)
	if err != nil {
		return fmt.Errorf("failed to update progress: %w", err)
	}

	return nil
}

// UpdateCategoryTestProgress updates category test progress
// This is stored in grammar_test_attempts table with scope_type='category'
// We check if the category test was passed to unlock the next category
func (r *GrammarAttemptRepository) UpdateCategoryTestProgress(userID int64, sectionID string, score int, passed bool) error {
	// The attempt is already saved in CreateAttempt, so we just need to verify it exists
	// The CanAccessSection will check for passed category tests
	return nil
}

// GetCategoryTestProgress checks if a category test was passed (score >= 50%)
func (r *GrammarAttemptRepository) GetCategoryTestProgress(userID int64, sectionID string) (bool, error) {
	query := `SELECT COUNT(*) > 0 FROM grammar_test_attempts
			  WHERE user_id = ? AND scope_type = 'category' AND scope_id = ? AND score >= 50 AND passed = 1`

	var hasPassed bool
	err := r.db.QueryRow(query, userID, sectionID).Scan(&hasPassed)
	if err != nil {
		return false, fmt.Errorf("failed to get category test progress: %w", err)
	}

	return hasPassed, nil
}

// HasCategoryTestAttempt checks if user has any category test attempt (even if not passed)
func (r *GrammarAttemptRepository) HasCategoryTestAttempt(userID int64, sectionID string) (bool, error) {
	query := `SELECT COUNT(*) > 0 FROM grammar_test_attempts
			  WHERE user_id = ? AND scope_type = 'category' AND scope_id = ?`

	var hasAttempt bool
	err := r.db.QueryRow(query, userID, sectionID).Scan(&hasAttempt)
	if err != nil {
		return false, fmt.Errorf("failed to check category test attempt: %w", err)
	}

	return hasAttempt, nil
}

// GetCategoryTestBestScore gets the best score for a category test
func (r *GrammarAttemptRepository) GetCategoryTestBestScore(userID int64, sectionID string) (int, error) {
	query := `SELECT MAX(score) FROM grammar_test_attempts
			  WHERE user_id = ? AND scope_type = 'category' AND scope_id = ?`

	var bestScore sql.NullInt64
	err := r.db.QueryRow(query, userID, sectionID).Scan(&bestScore)
	if err != nil {
		return 0, fmt.Errorf("failed to get category test best score: %w", err)
	}

	if !bestScore.Valid {
		return 0, nil
	}

	return int(bestScore.Int64), nil
}

// GetCategoryTestBestScores returns best category-test scores for all sections attempted by a user.
func (r *GrammarAttemptRepository) GetCategoryTestBestScores(userID int64) (map[string]int, error) {
	query := `SELECT scope_id, MAX(score)
			  FROM grammar_test_attempts
			  WHERE user_id = ? AND scope_type = 'category'
			  GROUP BY scope_id`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query category test best scores: %w", err)
	}
	defer rows.Close()

	scores := make(map[string]int)
	for rows.Next() {
		var sectionID string
		var score sql.NullInt64
		if err := rows.Scan(&sectionID, &score); err != nil {
			return nil, fmt.Errorf("failed to scan category test best score: %w", err)
		}
		if score.Valid {
			scores[sectionID] = int(score.Int64)
		}
	}
	return scores, rows.Err()
}

// GetChapterProgress retrieves progress for a chapter
func (r *GrammarAttemptRepository) GetChapterProgress(userID int64, chapterID string) (*ChapterProgress, error) {
	query := `SELECT best_score, passed_at, last_attempt_at
			  FROM grammar_progress
			  WHERE user_id = ? AND chapter_id = ?`

	var progress ChapterProgress
	var passedAt sql.NullString
	var lastAttemptAt sql.NullString

	err := r.db.QueryRow(query, userID, chapterID).Scan(
		&progress.BestScore,
		&passedAt,
		&lastAttemptAt,
	)

	if err == sql.ErrNoRows {
		return &ChapterProgress{
			BestScore: 0,
			Passed:    false,
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get progress: %w", err)
	}

	if passedAt.Valid {
		progress.PassedAt = parseTimestampFlex(passedAt.String)
		progress.Passed = true
	}

	if lastAttemptAt.Valid {
		progress.LastAttemptAt = parseTimestampFlex(lastAttemptAt.String)
	}

	return &progress, nil
}

// GetAllChapterProgress returns all grammar chapter progress rows for a user.
func (r *GrammarAttemptRepository) GetAllChapterProgress(userID int64) (map[string]*ChapterProgress, error) {
	query := `SELECT chapter_id, best_score, passed_at, last_attempt_at
			  FROM grammar_progress
			  WHERE user_id = ?`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chapter progress: %w", err)
	}
	defer rows.Close()

	progressByChapter := make(map[string]*ChapterProgress)
	for rows.Next() {
		var chapterID string
		var progress ChapterProgress
		var passedAt sql.NullString
		var lastAttemptAt sql.NullString
		if err := rows.Scan(&chapterID, &progress.BestScore, &passedAt, &lastAttemptAt); err != nil {
			return nil, fmt.Errorf("failed to scan chapter progress: %w", err)
		}
		if passedAt.Valid {
			progress.PassedAt = parseTimestampFlex(passedAt.String)
			progress.Passed = true
		}
		if lastAttemptAt.Valid {
			progress.LastAttemptAt = parseTimestampFlex(lastAttemptAt.String)
		}
		progressByChapter[chapterID] = &progress
	}
	return progressByChapter, rows.Err()
}

// parseTimestampFlex tries multiple common timestamp formats returned by PostgreSQL.
func parseTimestampFlex(s string) time.Time {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999-07:00",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ChapterProgress represents user progress for a chapter
type ChapterProgress struct {
	BestScore     int
	Passed        bool
	PassedAt      time.Time
	LastAttemptAt time.Time
}

// GetUserAttempts retrieves attempts for a user
func (r *GrammarAttemptRepository) GetUserAttempts(userID int64, scopeType, scopeID string, limit int) ([]*TestAttempt, error) {
	query := `SELECT id, user_id, scope_type, scope_id, started_at, finished_at, score, passed, total_questions, answers_json, results_json, course_version
			  FROM grammar_test_attempts
			  WHERE user_id = ? AND scope_type = ? AND scope_id = ?
			  ORDER BY started_at DESC
			  LIMIT ?`

	rows, err := r.db.Query(query, userID, scopeType, scopeID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query attempts: %w", err)
	}
	defer rows.Close()

	var attempts []*TestAttempt
	for rows.Next() {
		attempt := &TestAttempt{}
		var startedAt, finishedAt sql.NullString
		var passed int
		var courseVersion sql.NullString

		err := rows.Scan(
			&attempt.ID,
			&attempt.UserID,
			&attempt.ScopeType,
			&attempt.ScopeID,
			&startedAt,
			&finishedAt,
			&attempt.Score,
			&passed,
			&attempt.TotalQuestions,
			&attempt.AnswersJSON,
			&attempt.ResultsJSON,
			&courseVersion,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt: %w", err)
		}

		attempt.Passed = passed == 1
		if startedAt.Valid {
			attempt.StartedAt = parseTimestampFlex(startedAt.String)
		}
		if finishedAt.Valid {
			t := parseTimestampFlex(finishedAt.String)
			attempt.FinishedAt = &t
		}
		if courseVersion.Valid {
			attempt.CourseVersion = &courseVersion.String
		}

		attempts = append(attempts, attempt)
	}

	return attempts, rows.Err()
}

// GetBestScore retrieves the best score for a scope
func (r *GrammarAttemptRepository) GetBestScore(userID int64, scopeType, scopeID string) (int, error) {
	query := `SELECT MAX(score) FROM grammar_test_attempts
			  WHERE user_id = ? AND scope_type = ? AND scope_id = ?`

	var bestScore sql.NullInt64
	err := r.db.QueryRow(query, userID, scopeType, scopeID).Scan(&bestScore)
	if err != nil {
		return 0, fmt.Errorf("failed to get best score: %w", err)
	}

	if !bestScore.Valid {
		return 0, nil
	}

	return int(bestScore.Int64), nil
}

// GetAverageTestScore calculates average score using only the latest attempt for each chapter
// Only counts chapters where at least one test was completed
func (r *GrammarAttemptRepository) GetAverageTestScore(userID int64) (int, error) {
	// Get the latest attempt score for each chapter (scope_type = 'chapter')
	query := `SELECT score FROM (
				SELECT scope_id, score, 
					   ROW_NUMBER() OVER (PARTITION BY scope_id ORDER BY finished_at DESC, started_at DESC) as rn
				FROM grammar_test_attempts
				WHERE user_id = ? AND scope_type = 'chapter' AND finished_at IS NOT NULL
			) ranked
			WHERE rn = 1`

	args := []interface{}{userID}
	if r.courseCode != "" {
		query = strings.Replace(query, "AND scope_type = 'chapter' AND finished_at IS NOT NULL", "AND scope_type = 'chapter' AND finished_at IS NOT NULL AND scope_id LIKE ?", 1)
		args = append(args, strings.SplitN(r.courseCode, "_", 2)[0]+".grammar.%")
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to query test scores: %w", err)
	}
	defer rows.Close()

	var scores []int
	for rows.Next() {
		var score int
		if err := rows.Scan(&score); err != nil {
			continue
		}
		scores = append(scores, score)
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan scores: %w", err)
	}

	// If no tests were completed, return 0
	if len(scores) == 0 {
		return 0, nil
	}

	// Calculate average
	sum := 0
	for _, score := range scores {
		sum += score
	}
	avg := float64(sum) / float64(len(scores))

	return int(avg + 0.5), nil // Round to nearest integer
}

// ParseAnswersJSON parses answers JSON string into a map
func ParseAnswersJSON(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return make(map[string]interface{}), nil
	}
	var answers map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &answers); err != nil {
		return nil, fmt.Errorf("failed to parse answers JSON: %w", err)
	}
	return answers, nil
}

// ParseResultsJSON parses results JSON string into a slice
func ParseResultsJSON(jsonStr string) ([]interface{}, error) {
	if jsonStr == "" {
		return []interface{}{}, nil
	}
	var results []interface{}
	if err := json.Unmarshal([]byte(jsonStr), &results); err != nil {
		return nil, fmt.Errorf("failed to parse results JSON: %w", err)
	}
	return results, nil
}

// PlacementTestResult represents a placement test result
type PlacementTestResult struct {
	UserID         int64
	UserCourseID   int64
	Score          int
	TotalQuestions int
	OpenedSections []string
	CompletedAt    time.Time
	AdminOverride  bool
	Source         string
}

// SavePlacementTestResult saves or updates placement test result
// Only updates if the new score is higher (better) than existing, unless the existing row was set by an admin override
// (then the user's attempt always replaces it).
func (r *GrammarAttemptRepository) SavePlacementTestResult(userID int64, score int, totalQuestions int, openedSections []string) error {
	if r.courseCode != "" {
		return r.saveCoursePlacement(userID, score, totalQuestions, openedSections, "legacy", false)
	}
	openedSectionsJSON, _ := json.Marshal(openedSections)

	// Check existing result
	existingScore := 0
	var existingOpenedSectionsJSON sql.NullString
	var existingAdminOverride bool
	checkQuery := `SELECT score, opened_sections_json, COALESCE(admin_override, false) FROM grammar_placement_test WHERE user_id = ?`
	err := r.db.QueryRow(checkQuery, userID).Scan(&existingScore, &existingOpenedSectionsJSON, &existingAdminOverride)

	if err == sql.ErrNoRows {
		// No existing result, insert new
		query := `INSERT INTO grammar_placement_test (user_id, score, total_questions, opened_sections_json, completed_at, admin_override)
				  VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, false)`
		_, err = r.db.Exec(query, userID, score, totalQuestions, string(openedSectionsJSON))
		if err != nil {
			return fmt.Errorf("failed to save placement test result: %w", err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to check existing placement test result: %w", err)
	}

	if !existingAdminOverride && score <= existingScore {
		return nil
	}

	query := `UPDATE grammar_placement_test 
			  SET score = ?, total_questions = ?, opened_sections_json = ?, completed_at = CURRENT_TIMESTAMP, admin_override = false
			  WHERE user_id = ?`
	_, err = r.db.Exec(query, score, totalQuestions, string(openedSectionsJSON), userID)
	if err != nil {
		return fmt.Errorf("failed to update placement test result: %w", err)
	}

	return nil
}

// UpsertPlacementByAdmin overwrites placement (opened sections / level) regardless of score; marks admin_override.
func (r *GrammarAttemptRepository) UpsertPlacementByAdmin(userID int64, score int, totalQuestions int, openedSections []string) error {
	if r.courseCode != "" {
		return r.saveCoursePlacement(userID, score, totalQuestions, openedSections, "admin", false)
	}
	openedSectionsJSON, err := json.Marshal(openedSections)
	if err != nil {
		return fmt.Errorf("failed to marshal opened sections: %w", err)
	}
	q := `INSERT INTO grammar_placement_test (user_id, score, total_questions, opened_sections_json, completed_at, admin_override)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, true)
		ON CONFLICT (user_id) DO UPDATE SET
			score = EXCLUDED.score,
			total_questions = EXCLUDED.total_questions,
			opened_sections_json = EXCLUDED.opened_sections_json,
			completed_at = CURRENT_TIMESTAMP,
			admin_override = true`
	_, err = r.db.Exec(q, userID, score, totalQuestions, string(openedSectionsJSON))
	if err != nil {
		return fmt.Errorf("failed to upsert admin placement: %w", err)
	}
	return nil
}

// DeletePlacementTestResult removes placement row for a user (full reset of placement-based unlocks).
func (r *GrammarAttemptRepository) DeletePlacementTestResult(userID int64) error {
	if r.courseCode != "" {
		return r.saveCoursePlacement(userID, 0, 0, nil, "admin", true)
	}
	_, err := r.db.Exec(`DELETE FROM grammar_placement_test WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete placement test result: %w", err)
	}
	return nil
}

// GetPlacementTestResult retrieves placement test result for a user
func (r *GrammarAttemptRepository) GetPlacementTestResult(userID int64) (*PlacementTestResult, error) {
	if r.courseCode != "" {
		return r.getCoursePlacement(context.Background(), userID)
	}
	query := `SELECT user_id, score, total_questions, opened_sections_json, completed_at, COALESCE(admin_override, false)
			  FROM grammar_placement_test
			  WHERE user_id = ?`

	var result PlacementTestResult
	var openedSectionsJSON sql.NullString
	var completedAt sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&result.UserID,
		&result.Score,
		&result.TotalQuestions,
		&openedSectionsJSON,
		&completedAt,
		&result.AdminOverride,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No result found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get placement test result: %w", err)
	}

	if openedSectionsJSON.Valid {
		if err := json.Unmarshal([]byte(openedSectionsJSON.String), &result.OpenedSections); err != nil {
			return nil, fmt.Errorf("failed to parse opened sections JSON: %w", err)
		}
	}

	if completedAt.Valid {
		result.CompletedAt = parseTimestampFlex(completedAt.String)
	}
	result.Source = "legacy"

	return &result, nil
}
