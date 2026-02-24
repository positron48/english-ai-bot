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

// GetTrainingStreak returns the current streak (consecutive days with at least one session, ending yesterday)
// and whether the user trained yesterday. All dates are interpreted in the user's timezone.
// asOfLocalDate is "today" in user's local date (YYYY-MM-DD). Streak is counted backwards from yesterday.
func (r *SessionRepository) GetTrainingStreak(userID int64, timezone string, asOfLocalDate string) (streak int, trainedYesterday bool, err error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	asOf, err := time.ParseInLocation("2006-01-02", asOfLocalDate, loc)
	if err != nil {
		return 0, false, fmt.Errorf("invalid asOfLocalDate %q: %w", asOfLocalDate, err)
	}
	// Fetch sessions from the last 90 days (UTC window to cover any TZ)
	sinceUTC := asOf.AddDate(0, 0, -90).UTC()
	query := `SELECT started_at FROM training_sessions WHERE user_id = ? AND started_at >= ? ORDER BY started_at ASC`
	rows, err := r.db.Query(query, userID, sinceUTC)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get sessions for streak: %w", err)
	}
	defer rows.Close()

	datesSet := make(map[string]struct{})
	for rows.Next() {
		var startedAt time.Time
		if err := rows.Scan(&startedAt); err != nil {
			return 0, false, fmt.Errorf("scan started_at: %w", err)
		}
		localT := startedAt.In(loc)
		datesSet[localT.Format("2006-01-02")] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return 0, false, err
	}

	yesterday := asOf.AddDate(0, 0, -1).Format("2006-01-02")
	trainedYesterday = false
	if _, ok := datesSet[yesterday]; ok {
		trainedYesterday = true
	}

	// Streak: count consecutive days ending at yesterday
	streak = 0
	for d := asOf.AddDate(0, 0, -1); ; d = d.AddDate(0, 0, -1) {
		key := d.Format("2006-01-02")
		if _, ok := datesSet[key]; !ok {
			break
		}
		streak++
	}
	return streak, trainedYesterday, nil
}
