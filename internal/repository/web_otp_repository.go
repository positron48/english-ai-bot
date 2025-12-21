package repository

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"go.uber.org/zap"
)

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

	otp := &WebOTP{
		UserID:    userID,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(ttl),
	}

	query := `INSERT INTO web_otps (user_id, code_hash, expires_at) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, otp.UserID, otp.CodeHash, otp.ExpiresAt)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get OTP ID: %w", err)
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
		return nil, fmt.Errorf("invalid OTP code")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to validate OTP: %w", err)
	}

	otp.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAt)
	otp.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	if consumedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", consumedAt.String)
		otp.ConsumedAt = &t
	}

	// Check expiration
	if time.Now().After(otp.ExpiresAt) {
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
			_, err := rand.Read(n[:])
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

