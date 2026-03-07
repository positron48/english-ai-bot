package repository

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"tgbot-skeleton/internal/database"

	"go.uber.org/zap"
)

// randReader is used by generateRandomCode; replaced in tests to cover error and modulo branches.
var randReader io.Reader = rand.Reader

// WebOTP represents a one-time password
type WebOTP struct {
	ID         int64
	UserID     int64
	CodeHash   string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

// WebOTPRepository handles database operations for OTPs
type WebOTPRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewWebOTPRepository creates a new OTP repository
func NewWebOTPRepository(db *sql.DB, logger *zap.Logger) *WebOTPRepository {
	return &WebOTPRepository{
		db:     db,
		logger: logger,
	}
}

// GenerateOTP generates a new OTP code and returns both plain and hashed versions
func (r *WebOTPRepository) GenerateOTP(userID int64, ttl time.Duration) (string, *WebOTP, error) {
	// Generate random 6-digit code
	code := generateRandomCode(6)
	
	// Hash the code
	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	// Use UTC for all time operations to avoid timezone issues
	now := time.Now().UTC()
	otp := &WebOTP{
		UserID:    userID,
		CodeHash:  codeHash,
		ExpiresAt: now.Add(ttl),
	}

	r.logger.Debug("generating OTP", 
		zap.Int64("user_id", userID),
		zap.Time("expires_at", otp.ExpiresAt),
		zap.Duration("ttl", ttl))

	// Format time as UTC string
	expiresAtStr := otp.ExpiresAt.Format("2006-01-02 15:04:05")
	query := `INSERT INTO web_otps (user_id, code_hash, expires_at) VALUES (?, ?, ?)`
	id, err := database.InsertAndReturnID(r.db, query, otp.UserID, otp.CodeHash, expiresAtStr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	otp.ID = id
	r.logger.Debug("generated OTP", zap.Int64("otp_id", id), zap.Int64("user_id", userID))
	
	return code, otp, nil
}

// ValidateOTP validates an OTP code
func (r *WebOTPRepository) ValidateOTP(userID int64, code string) (*WebOTP, error) {
	// Hash the provided code
	hash := sha256.Sum256([]byte(code))
	codeHash := hex.EncodeToString(hash[:])

	r.logger.Debug("validating OTP", 
		zap.Int64("user_id", userID),
		zap.String("code_length", fmt.Sprintf("%d", len(code))),
		zap.String("code_hash", codeHash))

	query := `SELECT id, user_id, code_hash, expires_at, consumed_at, created_at
			  FROM web_otps 
			  WHERE user_id = ? AND code_hash = ? AND consumed_at IS NULL`

	var otp WebOTP
	var expiresAt, createdAt string
	var consumedAt sql.NullString

	err := r.db.QueryRow(query, userID, codeHash).Scan(
		&otp.ID,
		&otp.UserID,
		&otp.CodeHash,
		&expiresAt,
		&consumedAt,
		&createdAt,
	)

	if err == sql.ErrNoRows {
		r.logger.Debug("OTP not found in database", 
			zap.Int64("user_id", userID),
			zap.String("code_hash", codeHash))
		// Check if there are any OTPs for this user (for debugging)
		var count int
		countQuery := `SELECT COUNT(*) FROM web_otps WHERE user_id = ? AND consumed_at IS NULL`
		if err := r.db.QueryRow(countQuery, userID).Scan(&count); err == nil {
			r.logger.Debug("available OTPs for user", zap.Int64("user_id", userID), zap.Int("count", count))
		}
		return nil, fmt.Errorf("invalid OTP code")
	}
	if err != nil {
		r.logger.Error("failed to query OTP", zap.Error(err))
		return nil, fmt.Errorf("failed to validate OTP: %w", err)
	}

	// Parse times as UTC to avoid timezone issues
	loc, _ := time.LoadLocation("UTC")
	var errParse error
	otp.ExpiresAt, errParse = time.ParseInLocation("2006-01-02 15:04:05", expiresAt, loc)
	if errParse != nil {
		otp.ExpiresAt, errParse = time.Parse("2006-01-02T15:04:05Z", expiresAt)
	}
	if errParse != nil {
		// Postgres TIMESTAMPTZ returns RFC3339 with offset (e.g. 2026-02-15T12:22:49+03:00)
		otp.ExpiresAt, errParse = time.Parse(time.RFC3339, expiresAt)
		if errParse != nil {
			r.logger.Warn("failed to parse expires_at", zap.String("expires_at", expiresAt), zap.Error(errParse))
		}
	}

	otp.CreatedAt, errParse = time.ParseInLocation("2006-01-02 15:04:05", createdAt, loc)
	if errParse != nil {
		otp.CreatedAt, errParse = time.Parse("2006-01-02T15:04:05Z", createdAt)
	}
	if errParse != nil {
		otp.CreatedAt, errParse = time.Parse(time.RFC3339, createdAt)
		if errParse != nil {
			r.logger.Warn("failed to parse created_at", zap.String("created_at", createdAt), zap.Error(errParse))
		}
	}
	if consumedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", consumedAt.String)
		otp.ConsumedAt = &t
	}

	// Use UTC for comparison to match how we store times
	now := time.Now().UTC()
	timeUntilExpiry := otp.ExpiresAt.Sub(now)
	
	r.logger.Debug("OTP found", 
		zap.Int64("otp_id", otp.ID),
		zap.Time("created_at", otp.CreatedAt),
		zap.Time("expires_at", otp.ExpiresAt),
		zap.Time("now", now),
		zap.String("time_until_expiry", timeUntilExpiry.String()),
		zap.Float64("seconds_until_expiry", timeUntilExpiry.Seconds()))

	// Check expiration - add 5 second grace period to account for clock skew
	if now.After(otp.ExpiresAt.Add(5 * time.Second)) {
		r.logger.Debug("OTP expired", 
			zap.Int64("otp_id", otp.ID),
			zap.Time("expires_at", otp.ExpiresAt),
			zap.Time("now", now),
			zap.String("expired_by", now.Sub(otp.ExpiresAt).String()))
		return nil, fmt.Errorf("OTP expired")
	}

	// Mark as consumed
	query = `UPDATE web_otps SET consumed_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = r.db.Exec(query, otp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to mark OTP as consumed: %w", err)
	}

	return &otp, nil
}

// generateRandomCode generates a random numeric code
func generateRandomCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		// Generate random digit 0-9
		var n [1]byte
		for {
			_, err := randReader.Read(n[:])
			if err != nil {
				// Fallback to simple modulo
				n[0] = byte(time.Now().UnixNano() % 10)
				break
			}
			if n[0] < 10 {
				break
			}
			// If >= 10, use modulo to get 0-9
			n[0] = n[0] % 10
			break
		}
		b[i] = '0' + n[0]
	}
	return string(b)
}

// CleanupExpiredOTPs removes expired OTPs
func (r *WebOTPRepository) CleanupExpiredOTPs() error {
	query := `DELETE FROM web_otps WHERE expires_at < CURRENT_TIMESTAMP OR consumed_at IS NOT NULL`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired OTPs: %w", err)
	}
	return nil
}
