package repository

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func setupWebOTPTestDB(t *testing.T) *sql.DB {
	return testutil.SetupTestDB(t)
}

func TestNewWebOTPRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebOTPTestDB(t)
	defer db.Close()

	repo := NewWebOTPRepository(db, logger)
	_ = repo // Verify repository is created
}

func TestWebOTPRepository_GenerateOTP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebOTPTestDB(t)
	defer db.Close()

	repo := NewWebOTPRepository(db, logger)

	code, otp, err := repo.GenerateOTP(123, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP() error = %v", err)
	}
	if code == "" {
		t.Error("GenerateOTP() should return non-empty code")
	}
	if otp == nil {
		t.Fatal("GenerateOTP() should not return nil OTP")
	}
	if otp.UserID != 123 {
		t.Errorf("Expected UserID 123, got %d", otp.UserID)
	}
	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}
}

func TestWebOTPRepository_ValidateOTP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebOTPTestDB(t)
	defer db.Close()

	repo := NewWebOTPRepository(db, logger)

	// Generate an OTP
	code, otp, err := repo.GenerateOTP(456, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	// Validate the OTP
	validated, err := repo.ValidateOTP(456, code)
	if err != nil {
		t.Fatalf("ValidateOTP() error = %v", err)
	}
	if validated == nil {
		t.Fatal("ValidateOTP() should not return nil")
	}
	if validated.ID != otp.ID {
		t.Errorf("Expected OTP ID %d, got %d", otp.ID, validated.ID)
	}

	// Try to validate again (should fail - already consumed)
	_, err = repo.ValidateOTP(456, code)
	if err == nil {
		t.Error("ValidateOTP() should return error for consumed OTP")
	}
}

func TestWebOTPRepository_ValidateOTP_InvalidCode(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebOTPTestDB(t)
	defer db.Close()

	repo := NewWebOTPRepository(db, logger)

	// Try to validate invalid code
	_, err := repo.ValidateOTP(789, "000000")
	if err == nil {
		t.Error("ValidateOTP() should return error for invalid code")
	}
}

func TestWebOTPRepository_CleanupExpiredOTPs(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := setupWebOTPTestDB(t)
	defer db.Close()

	repo := NewWebOTPRepository(db, logger)

	// Generate an expired OTP (negative duration)
	_, _, err := repo.GenerateOTP(999, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired OTP: %v", err)
	}

	// Cleanup expired OTPs
	err = repo.CleanupExpiredOTPs()
	if err != nil {
		t.Fatalf("CleanupExpiredOTPs() error = %v", err)
	}
}
