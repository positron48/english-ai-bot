package repository

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"tgbot-skeleton/internal/database"

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
	// Format time as UTC string to avoid timezone issues
	expiresAtStr := session.ExpiresAt.UTC().Format("2006-01-02 15:04:05")
	
	// Log token preview for debugging
	tokenPreview := session.Token
	if len(session.Token) > 16 {
		tokenPreview = session.Token[:8] + "..." + session.Token[len(session.Token)-8:]
	}
	r.logger.Info("creating session in database",
		zap.Int64("user_id", session.UserID),
		zap.String("token_preview", tokenPreview),
		zap.String("token_length", fmt.Sprintf("%d", len(session.Token))),
		zap.Time("expires_at", session.ExpiresAt),
		zap.String("expires_at_str", expiresAtStr))
	
	query := `INSERT INTO web_sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)`
	id, err := database.InsertAndReturnID(r.db, query, session.UserID, session.Token, expiresAtStr)
	if err != nil {
		r.logger.Error("failed to insert session into database",
			zap.Int64("user_id", session.UserID),
			zap.String("token_preview", tokenPreview),
			zap.Error(err))
		return fmt.Errorf("failed to create session: %w", err)
	}

	session.ID = id
	r.logger.Info("session created successfully in database",
		zap.Int64("session_id", id),
		zap.Int64("user_id", session.UserID),
		zap.String("token_preview", tokenPreview))
	return nil
}

// GetSessionByToken gets a session by token
func (r *WebSessionRepository) GetSessionByToken(token string) (*WebSession, error) {
	query := `SELECT id, user_id, session_token, expires_at, created_at, last_seen_at
			  FROM web_sessions WHERE session_token = ?`

	var session WebSession
	var createdAt, lastSeenAt, expiresAt string

	// Log token preview for debugging (first 8 and last 8 chars)
	tokenPreview := token
	if len(token) > 16 {
		tokenPreview = token[:8] + "..." + token[len(token)-8:]
	}
	r.logger.Info("looking up session by token", zap.String("token_preview", tokenPreview), zap.String("token_length", fmt.Sprintf("%d", len(token))))

	err := r.db.QueryRow(query, token).Scan(
		&session.ID,
		&session.UserID,
		&session.Token,
		&expiresAt,
		&createdAt,
		&lastSeenAt,
	)

	if err == sql.ErrNoRows {
		r.logger.Warn("session not found in database", zap.String("token_preview", tokenPreview))
		// Check if there are any sessions at all
		var count int
		countQuery := `SELECT COUNT(*) FROM web_sessions`
		if err2 := r.db.QueryRow(countQuery).Scan(&count); err2 == nil {
			r.logger.Info("total sessions in database", zap.Int("count", count))
		}
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query session", zap.String("token_preview", tokenPreview), zap.Error(err))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Parse times as UTC to match how we store them
	// Try multiple date formats
	loc, _ := time.LoadLocation("UTC")
	
	// Helper function to parse time with multiple format attempts
	parseTime := func(timeStr, fieldName string) time.Time {
		// Try ISO 8601 format first
		if t, err := time.Parse("2006-01-02T15:04:05Z", timeStr); err == nil {
			return t.UTC()
		}
		// Try ISO 8601 with timezone offset
		if t, err := time.Parse("2006-01-02T15:04:05-07:00", timeStr); err == nil {
			return t.UTC()
		}
		// Try our custom format
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, loc); err == nil {
			return t.UTC()
		}
		r.logger.Warn("failed to parse time", zap.String("field", fieldName), zap.String("value", timeStr))
		return time.Time{}
	}
	
	session.ExpiresAt = parseTime(expiresAt, "expires_at")
	session.CreatedAt = parseTime(createdAt, "created_at")
	session.LastSeenAt = parseTime(lastSeenAt, "last_seen_at")

	r.logger.Info("session found in database",
		zap.Int64("session_id", session.ID),
		zap.Int64("user_id", session.UserID),
		zap.String("token_preview", tokenPreview),
		zap.Time("expires_at", session.ExpiresAt))

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
