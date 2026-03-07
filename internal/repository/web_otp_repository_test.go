package repository

import (
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

func TestWebOTPRepository_GenerateOTP(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	// Create user first
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	otpRepo := NewWebOTPRepository(conn, logger)

	code, _, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute) // 5 minute TTL
	if err != nil {
		t.Fatalf("GenerateOTP() error = %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Expected 6-digit code, got %d digits", len(code))
	}
}

func TestWebOTPRepository_ValidateOTP_Valid(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	otpRepo := NewWebOTPRepository(conn, logger)

	code, _, _ := otpRepo.GenerateOTP(user.ID, 5*time.Minute)

	otp, err := otpRepo.ValidateOTP(user.ID, code)
	if err != nil {
		t.Fatalf("ValidateOTP() error = %v", err)
	}

	if otp == nil {
		t.Error("Expected valid OTP")
	}

	if otp != nil && otp.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, otp.UserID)
	}
}

func TestWebOTPRepository_ValidateOTP_Invalid(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	otpRepo := NewWebOTPRepository(conn, logger)

	_, _, _ = otpRepo.GenerateOTP(user.ID, 5*time.Minute)

	otp, err := otpRepo.ValidateOTP(user.ID, "000000")
	// Should return error for invalid code
	if err == nil {
		t.Error("Expected error for invalid OTP code")
	}

	if otp != nil {
		t.Error("Expected nil OTP for invalid code")
	}
}

func TestWebOTPRepository_ValidateOTP_Consumed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12345)

	otpRepo := NewWebOTPRepository(conn, logger)

	code, _, _ := otpRepo.GenerateOTP(user.ID, 5*time.Minute)

	// First validation should succeed
	otp1, _ := otpRepo.ValidateOTP(user.ID, code)
	if otp1 == nil {
		t.Error("First validation should succeed")
	}

	// Second validation should fail (OTP consumed)
	otp2, _ := otpRepo.ValidateOTP(user.ID, code)
	if otp2 != nil {
		t.Error("Second validation should fail (OTP consumed)")
	}
}

func TestWebOTPRepository_CleanupExpiredOTPs(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	otpRepo := NewWebOTPRepository(conn, logger)

	err := otpRepo.CleanupExpiredOTPs()
	if err != nil {
		t.Fatalf("CleanupExpiredOTPs() error = %v", err)
	}
}

func TestWebOTPRepository_CleanupExpiredOTPs_RemovesExpired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)

	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12399)

	// Insert expired OTP directly (past expires_at)
	_, err := conn.Exec(`INSERT INTO web_otps (user_id, code_hash, expires_at) VALUES (?, ?, '2020-01-01 00:00:00')`,
		user.ID, "dummyhash")
	if err != nil {
		t.Fatalf("setup: insert expired OTP: %v", err)
	}

	otpRepo := NewWebOTPRepository(conn, logger)
	err = otpRepo.CleanupExpiredOTPs()
	if err != nil {
		t.Fatalf("CleanupExpiredOTPs() error = %v", err)
	}

	var count int
	_ = conn.QueryRow(`SELECT COUNT(*) FROM web_otps WHERE user_id = ?`, user.ID).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 OTPs after cleanup, got %d", count)
	}
}

func TestWebOTPRepository_GenerateOTP_CodeFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12398)
	otpRepo := NewWebOTPRepository(conn, logger)

	code, _, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP() error = %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got len %d", len(code))
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Errorf("expected only digits, got %q", code)
			break
		}
	}
}

// TestWebOTPRepository_ValidateOTP_Expired covers the expired OTP path.
func TestWebOTPRepository_ValidateOTP_Expired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(12397)
	otpRepo := NewWebOTPRepository(conn, logger)

	code, _, err := otpRepo.GenerateOTP(user.ID, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("GenerateOTP() error = %v", err)
	}
	// Set expires_at to the past so validation returns "OTP expired"
	_, err = conn.Exec(`UPDATE web_otps SET expires_at = $1 WHERE user_id = $2`, time.Now().Add(-1*time.Hour).UTC().Format("2006-01-02 15:04:05"), user.ID)
	if err != nil {
		t.Fatalf("UPDATE expires_at: %v", err)
	}
	// Use the same code - row exists but is expired
	otp, err := otpRepo.ValidateOTP(user.ID, code)
	if err == nil {
		t.Error("ValidateOTP() expected error for expired OTP")
	}
	if otp != nil {
		t.Error("ValidateOTP() expected nil OTP when expired")
	}
	if err != nil && err.Error() != "OTP expired" {
		t.Logf("expected 'OTP expired' error, got: %v", err)
	}
}
