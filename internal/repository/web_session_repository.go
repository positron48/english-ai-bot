package repository

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// WebSession represents a web session
type WebSession struct {
	ID         int64
	UserID     int64
	Token      string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastSeenAt time.Time
}

// WebSessionRepository handles database operations for web sessions
type WebSessionRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewWebSessionRepository creates a new web session repository
func NewWebSessionRepository(db *sql.DB, logger *zap.Logger) *WebSessionRepository {
	return &WebSessionRepository{
		db:     db,
		logger: logger,
	}
}

// CreateSession creates a new web session
func (r *WebSessionRepository) CreateSession(session *WebSession) error {
	query := `INSERT INTO web_sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, session.UserID, session.Token, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get session ID: %w", err)
	}

	session.ID = id
	r.logger.Debug("created web session", zap.Int64("session_id", id), zap.Int64("user_id", session.UserID))
	return nil
}

// GetSessionByToken gets a session by token
func (r *WebSessionRepository) GetSessionByToken(token string) (*WebSession, error) {
	query := `SELECT id, user_id, session_token, expires_at, created_at, last_seen_at
			  FROM web_sessions WHERE session_token = ?`

	var session WebSession
	var createdAt, lastSeenAt, expiresAt string

	err := r.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&expiresAt,
		&createdAt,
		&lastSeenAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAt)
	session.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	session.LastSeenAt, _ = time.Parse("2006-01-02 15:04:05", lastSeenAt)

	return &session, nil
}

// DeleteSession deletes a session by token
func (r *WebSessionRepository) DeleteSession(token string) error {
	query := `DELETE FROM web_sessions WHERE session_token = ?`
	_, err := r.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// UpdateLastSeen updates the last seen timestamp for a session from request
func (r *WebSessionRepository) UpdateLastSeen(req *http.Request) error {
	cookie, err := req.Cookie("session")
	if err != nil {
		return err
	}

	query := `UPDATE web_sessions SET last_seen_at = CURRENT_TIMESTAMP WHERE session_token = ?`
	_, err = r.db.Exec(query, cookie.Value)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *WebSessionRepository) CleanupExpiredSessions() error {
	query := `DELETE FROM web_sessions WHERE expires_at < CURRENT_TIMESTAMP`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

