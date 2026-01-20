package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// GrammarAttemptRepository handles database operations for grammar test attempts
type GrammarAttemptRepository struct {
	db     *sql.DB
	logger *zap.Logger
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
	ID            int64
	UserID        int64
	ScopeType     string // "chapter" or "category"
	ScopeID       string
	StartedAt     time.Time
	FinishedAt    *time.Time
	Score         int
	Passed        bool
	TotalQuestions int
	AnswersJSON   string
	ResultsJSON   string
	CourseVersion  *string
}

// CreateAttempt creates a new test attempt
func (r *GrammarAttemptRepository) CreateAttempt(attempt *TestAttempt) (int64, error) {
	query := `INSERT INTO grammar_test_attempts 
			  (user_id, scope_type, scope_id, started_at, finished_at, score, passed, total_questions, answers_json, results_json, course_version)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	var finishedAt interface{}
	if attempt.FinishedAt != nil {
		finishedAt = attempt.FinishedAt.Format("2006-01-02 15:04:05")
	}

	passed := 0
	if attempt.Passed {
		passed = 1
	}

	result, err := r.db.Exec(query,
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
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create attempt: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get attempt ID: %w", err)
	}

	return id, nil
}

// UpdateProgress updates grammar progress for a chapter
func (r *GrammarAttemptRepository) UpdateProgress(userID int64, chapterID string, score int, passed bool) error {
	query := `INSERT INTO grammar_progress (user_id, chapter_id, best_score, passed_at, last_attempt_at)
			  VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
			  ON CONFLICT(user_id, chapter_id) DO UPDATE SET
			  	best_score = MAX(best_score, excluded.best_score),
			  	passed_at = CASE WHEN excluded.best_score >= 50 THEN COALESCE(passed_at, CURRENT_TIMESTAMP) ELSE passed_at END,
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
		progress.PassedAt, _ = time.Parse("2006-01-02 15:04:05", passedAt.String)
		progress.Passed = true
	}

	if lastAttemptAt.Valid {
		progress.LastAttemptAt, _ = time.Parse("2006-01-02 15:04:05", lastAttemptAt.String)
	}

	return &progress, nil
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
			attempt.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt.String)
		}
		if finishedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", finishedAt.String)
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
