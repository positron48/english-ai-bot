package web

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/repository"

	"tgbot-skeleton/internal/testutil"
	"go.uber.org/zap"
)

func setupAuthOTPTestDB(t *testing.T) (*sql.DB, *repository.WebOTPRepository, *repository.UserRepository) {
	t.Helper()
	db := testutil.SetupTestDB(t)

	logger, _ := zap.NewDevelopment()
	otpRepo := repository.NewWebOTPRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)

	return db, otpRepo, userRepo
}

func TestHandleAuthOTP_GenerateOTP(t *testing.T) {
	_, otpRepo, userRepo := setupAuthOTPTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(12345)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Generate OTP
	code, otp, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP() error = %v", err)
	}
	if code == "" {
		t.Error("GenerateOTP() should return non-empty code")
	}
	if otp == nil {
		t.Fatal("GenerateOTP() should not return nil OTP")
	}
	if otp.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, otp.UserID)
	}
}

func TestHandleAuthOTP_ValidateOTP(t *testing.T) {
	_, otpRepo, userRepo := setupAuthOTPTestDB(t)

	// Create a user
	user, err := userRepo.GetOrCreateUser(67890)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Generate OTP
	code, otp, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate OTP: %v", err)
	}

	// Validate OTP
	validated, err := otpRepo.ValidateOTP(user.ID, code)
	if err != nil {
		t.Fatalf("ValidateOTP() error = %v", err)
	}
	if validated == nil {
		t.Fatal("ValidateOTP() should not return nil")
	}
	if validated.ID != otp.ID {
		t.Errorf("Expected OTP ID %d, got %d", otp.ID, validated.ID)
	}
}
