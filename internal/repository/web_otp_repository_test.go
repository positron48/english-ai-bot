package repository

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/testutil"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestNewWebOTPRepository(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	repo := NewWebOTPRepository(conn, logger)
	if repo == nil {
		t.Error("NewWebOTPRepository() should not return nil")
	}
}

func TestWebOTPRepository_GenerateOTP_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	userID := int64(1)
	mock.ExpectQuery("INSERT INTO web_otps .+ RETURNING id").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert failed"))

	repo := NewWebOTPRepository(db, zap.NewNop())
	_, _, err = repo.GenerateOTP(userID, 5*time.Minute)
	if err == nil {
		t.Error("GenerateOTP() expected error when insert fails")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to create OTP") {
		t.Errorf("expected 'failed to create OTP' wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestWebOTPRepository_ValidateOTP_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("(?s)SELECT .+ FROM web_otps .+ consumed_at IS NULL").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("db connection error"))

	repo := NewWebOTPRepository(db, zap.NewNop())
	_, err = repo.ValidateOTP(1, "123456")
	if err == nil {
		t.Error("ValidateOTP() expected error when query fails")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to validate OTP") {
		t.Errorf("expected 'failed to validate OTP' wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestWebOTPRepository_ValidateOTP_UpdateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Future expires_at so OTP is not expired
	expiresAt := time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	createdAt := time.Now().UTC().Format("2006-01-02 15:04:05")
	cols := []string{"id", "user_id", "code_hash", "expires_at", "consumed_at", "created_at"}
	mock.ExpectQuery("(?s)SELECT .+ FROM web_otps .+ consumed_at IS NULL").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, 1, "hash", expiresAt, nil, createdAt))
	mock.ExpectExec("UPDATE web_otps SET consumed_at").
		WithArgs(1).
		WillReturnError(fmt.Errorf("update failed"))

	repo := NewWebOTPRepository(db, zap.NewNop())
	_, err = repo.ValidateOTP(1, "123456")
	if err == nil {
		t.Error("ValidateOTP() expected error when update fails")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to mark OTP as consumed") {
		t.Errorf("expected 'failed to mark OTP as consumed' wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestWebOTPRepository_ValidateOTP_TimeParseRFC3339(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Use RFC3339 format so ParseInLocation fails and fallback to RFC3339 is used
	expiresAt := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	createdAt := time.Now().UTC().Format(time.RFC3339)
	cols := []string{"id", "user_id", "code_hash", "expires_at", "consumed_at", "created_at"}
	mock.ExpectQuery("(?s)SELECT .+ FROM web_otps .+ consumed_at IS NULL").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, 1, "hash", expiresAt, nil, createdAt))
	mock.ExpectExec("UPDATE web_otps SET consumed_at").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewWebOTPRepository(db, zap.NewNop())
	otp, err := repo.ValidateOTP(1, "123456")
	if err != nil {
		t.Fatalf("ValidateOTP() unexpected error: %v", err)
	}
	if otp == nil || otp.ID != 1 {
		t.Errorf("expected valid OTP, got: %v", otp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestWebOTPRepository_ValidateOTP_ConsumedAtSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	expiresAt := time.Now().UTC().Add(1 * time.Hour).Format("2006-01-02 15:04:05")
	createdAt := time.Now().UTC().Format("2006-01-02 15:04:05")
	consumedAt := time.Now().UTC().Format("2006-01-02 15:04:05")
	cols := []string{"id", "user_id", "code_hash", "expires_at", "consumed_at", "created_at"}
	mock.ExpectQuery("(?s)SELECT .+ FROM web_otps .+ consumed_at IS NULL").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, 1, "hash", expiresAt, consumedAt, createdAt))
	mock.ExpectExec("UPDATE web_otps SET consumed_at").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repo := NewWebOTPRepository(db, zap.NewNop())
	otp, err := repo.ValidateOTP(1, "123456")
	if err != nil {
		t.Fatalf("ValidateOTP() unexpected error: %v", err)
	}
	if otp == nil {
		t.Fatal("expected non-nil OTP")
	}
	if otp.ConsumedAt == nil {
		t.Error("expected ConsumedAt to be set when consumed_at was returned from DB")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestWebOTPRepository_ValidateOTP_TimeParseWarn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// Invalid date format so all parse attempts fail and Warn is logged; zero time is used, then "OTP expired" is returned
	cols := []string{"id", "user_id", "code_hash", "expires_at", "consumed_at", "created_at"}
	mock.ExpectQuery("(?s)SELECT .+ FROM web_otps .+ consumed_at IS NULL").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1, 1, "hash", "invalid-date", nil, "invalid-date"))

	repo := NewWebOTPRepository(db, zap.NewNop())
	_, err = repo.ValidateOTP(1, "123456")
	if err == nil {
		t.Error("ValidateOTP() expected error when dates are invalid (zero time → expired)")
	}
	if err != nil && err.Error() != "OTP expired" {
		t.Logf("ValidateOTP with invalid dates returned: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

func TestWebOTPRepository_CleanupExpiredOTPs_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM web_otps WHERE").
		WillReturnError(fmt.Errorf("delete failed"))

	repo := NewWebOTPRepository(db, zap.NewNop())
	err = repo.CleanupExpiredOTPs()
	if err == nil {
		t.Error("CleanupExpiredOTPs() expected error when delete fails")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to cleanup expired OTPs") {
		t.Errorf("expected 'failed to cleanup expired OTPs' wrapped error, got: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("mock expectations: %v", err)
	}
}

// errReader is an io.Reader that always returns an error (used to cover generateRandomCode fallback).
type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("injected read error")
}

// highByteReader returns bytes >= 10 so generateRandomCode uses the modulo branch.
type highByteReader struct{ byte }

func (r highByteReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = r.byte
	return 1, nil
}

func TestWebOTPRepository_GenerateOTP_RandFallback(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(99991)
	otpRepo := NewWebOTPRepository(conn, logger)

	old := randReader
	randReader = errReader{}
	defer func() { randReader = old }()

	code, _, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP() with rand fallback error = %v", err)
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

func TestWebOTPRepository_GenerateOTP_RandModuloBranch(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	conn := testutil.SetupTestDB(t)
	userRepo := NewUserRepository(conn, logger)
	user, _ := userRepo.GetOrCreateUser(99992)
	otpRepo := NewWebOTPRepository(conn, logger)

	// Reader that returns 250 (>= 10) so n[0] % 10 is used
	old := randReader
	randReader = highByteReader{250}
	defer func() { randReader = old }()

	code, _, err := otpRepo.GenerateOTP(user.ID, 5*time.Minute)
	if err != nil {
		t.Fatalf("GenerateOTP() with high-byte rand error = %v", err)
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

// lowByteReader returns a single byte in 0-9 so generateRandomCode takes the "n[0] < 10" branch.
type lowByteReader struct{ b byte }

func (r lowByteReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = r.b
	return 1, nil
}

func TestGenerateRandomCode_LengthZero(t *testing.T) {
	got := generateRandomCode(0)
	if got != "" {
		t.Errorf("generateRandomCode(0) = %q, want %q", got, "")
	}
}

func TestGenerateRandomCode_LowByteBranch(t *testing.T) {
	old := randReader
	randReader = lowByteReader{b: 7}
	defer func() { randReader = old }()

	got := generateRandomCode(3)
	if got != "777" {
		t.Errorf("generateRandomCode(3) with lowByteReader(7) = %q, want %q", got, "777")
	}
}

func TestGenerateRandomCode_HighByteModuloBranch(t *testing.T) {
	old := randReader
	randReader = highByteReader{253} // 253 % 10 = 3
	defer func() { randReader = old }()

	got := generateRandomCode(2)
	if got != "33" {
		t.Errorf("generateRandomCode(2) with highByteReader(253) = %q, want %q", got, "33")
	}
}

func TestGenerateRandomCode_ErrReaderFallback(t *testing.T) {
	old := randReader
	randReader = errReader{}
	defer func() { randReader = old }()

	got := generateRandomCode(4)
	if len(got) != 4 {
		t.Errorf("generateRandomCode(4) with errReader: len = %d, want 4", len(got))
	}
	for _, c := range got {
		if c < '0' || c > '9' {
			t.Errorf("generateRandomCode(4) with errReader: got non-digit %q", got)
			break
		}
	}
}
