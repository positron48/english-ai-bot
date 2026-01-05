package web

import (
	"database/sql"
	"testing"
	"time"

	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func setupAuthOTPTestDB(t *testing.T) (*sql.DB, *repository.WebOTPRepository, *repository.UserRepository) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_id INTEGER UNIQUE NOT NULL,
		telegram_username TEXT,
		username TEXT,
		timezone TEXT DEFAULT '',
		preferred_training_time TEXT DEFAULT '',
		settings_json TEXT DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS web_otps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		code_hash TEXT NOT NULL,
		expires_at TEXT NOT NULL,
		consumed_at TEXT,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTables)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	otpRepo := repository.NewWebOTPRepository(db, logger)
	userRepo := repository.NewUserRepository(db, logger)

	return db, otpRepo, userRepo
}

func TestHandleAuthOTP_GenerateOTP(t *testing.T) {
	db, otpRepo, userRepo := setupAuthOTPTestDB(t)
	defer db.Close()

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
	db, otpRepo, userRepo := setupAuthOTPTestDB(t)
	defer db.Close()

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
