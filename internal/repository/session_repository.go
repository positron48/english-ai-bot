package repository

import (
	"database/sql"
	"fmt"
	"time"

	"tgbot-skeleton/internal/database"
	"tgbot-skeleton/internal/models"

	"go.uber.org/zap"
)

// SessionRepository handles database operations for training sessions
type SessionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB, logger *zap.Logger) *SessionRepository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

// CreateSession creates a new training session
func (r *SessionRepository) CreateSession(session *models.TrainingSession) (int64, error) {
	query := `INSERT INTO training_sessions (
		user_id, source, planned_count, done_count, session_json
	) VALUES (?, ?, ?, ?, ?)`

	id, err := database.InsertAndReturnID(r.db, query,
		session.UserID, session.Source, session.PlannedCount,
		session.DoneCount, session.SessionJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}

	r.logger.Info("created training session",
		zap.Int64("session_id", id),
		zap.Int64("user_id", session.UserID),
		zap.String("source", string(session.Source)),
	)

	return id, nil
}

// GetSession gets a session by ID
func (r *SessionRepository) GetSession(id int64) (*models.TrainingSession, error) {
	query := `SELECT id, user_id, started_at, ended_at, source, 
			  planned_count, done_count, COALESCE(session_json, '')
			  FROM training_sessions WHERE id = ?`

	var session models.TrainingSession
	var startedAt string
	var endedAt sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&session.ID, &session.UserID, &startedAt, &endedAt,
		&session.Source, &session.PlannedCount, &session.DoneCount,
		&session.SessionJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
	if endedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", endedAt.String)
		session.EndedAt = &t
	}

	return &session, nil
}

// UpdateSession updates a session
func (r *SessionRepository) UpdateSession(session *models.TrainingSession) error {
	query := `UPDATE training_sessions SET
			  ended_at = ?, done_count = ?, session_json = ?
			  WHERE id = ?`

	_, err := r.db.Exec(query, session.EndedAt, session.DoneCount, session.SessionJSON, session.ID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// FinishSession marks a session as finished
func (r *SessionRepository) FinishSession(sessionID int64, doneCount int) error {
	query := `UPDATE training_sessions SET ended_at = CURRENT_TIMESTAMP, done_count = ? WHERE id = ?`
	_, err := r.db.Exec(query, doneCount, sessionID)
	if err != nil {
		return fmt.Errorf("failed to finish session: %w", err)
	}

	r.logger.Info("finished training session",
		zap.Int64("session_id", sessionID),
		zap.Int("done_count", doneCount),
	)

	return nil
}

// GetActiveSession gets the active session for a user (if any)
func (r *SessionRepository) GetActiveSession(userID int64) (*models.TrainingSession, error) {
	query := `SELECT id, user_id, started_at, ended_at, source, 
			  planned_count, done_count, COALESCE(session_json, '')
			  FROM training_sessions 
			  WHERE user_id = ? AND ended_at IS NULL
			  ORDER BY started_at DESC LIMIT 1`

	var session models.TrainingSession
	var startedAt string
	var endedAt sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&session.ID, &session.UserID, &startedAt, &endedAt,
		&session.Source, &session.PlannedCount, &session.DoneCount,
		&session.SessionJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %w", err)
	}

	session.StartedAt, _ = time.Parse("2006-01-02 15:04:05", startedAt)
	if endedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", endedAt.String)
		session.EndedAt = &t
	}

	return &session, nil
}

// CreateReviewEvent creates a review event
func (r *SessionRepository) CreateReviewEvent(event *models.ReviewEvent) (int64, error) {
	query := `INSERT INTO review_events (
		session_id, user_id, user_card_id, direction, shown_at,
		options_shown_at, answered_at, t_delay_ms, early_reveal,
		option_count, options_json, chosen_option, is_correct,
		quality, metrics_json, srs_before_json, srs_after_json
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	earlyReveal := 0
	if event.EarlyReveal {
		earlyReveal = 1
	}
	isCorrect := 0
	if event.IsCorrect {
		isCorrect = 1
	}

	id, err := database.InsertAndReturnID(r.db, query,
		event.SessionID, event.UserID, event.UserCardID, event.Direction,
		event.ShownAt, event.OptionsShownAt, event.AnsweredAt,
		event.TDelayMS, earlyReveal, event.OptionCount,
		event.OptionsJSON, event.ChosenOption, isCorrect,
		event.Quality, event.MetricsJSON, event.SRSBeforeJSON, event.SRSAfterJSON,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create review event: %w", err)
	}

	return id, nil
}

// GetSessionStats gets statistics for a training session
func (r *SessionRepository) GetSessionStats(sessionID int64) (totalCards int, correctCards int, err error) {
	query := `SELECT 
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN is_correct = 1 THEN 1 ELSE 0 END), 0) as correct
	FROM review_events 
	WHERE session_id = ? AND answered_at IS NOT NULL`
	
	err = r.db.QueryRow(query, sessionID).Scan(&totalCards, &correctCards)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get session stats: %w", err)
	}
	return totalCards, correctCards, nil
}

// GetTodaySessionCount checks if user has trained today
func (r *SessionRepository) GetTodaySessionCount(userID int64, localDate string) (int, error) {
	query := `SELECT COUNT(*) FROM training_sessions 
			  WHERE user_id = ? AND DATE(started_at) = ?`

	var count int
	err := r.db.QueryRow(query, userID, localDate).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get today session count: %w", err)
	}
	return count, nil
}
