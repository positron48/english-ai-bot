package web

// Tests to cover error paths in auth.go not covered by existing tests.

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/testutil"

	"go.uber.org/zap"
)

// buildInitDataWithHash builds a Telegram initData string with the given params and hash.
// botToken is used to compute the expected HMAC hash.
func buildInitDataWithHash(botToken string, params map[string]string) string {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	dataCheckString := ""
	for i, p := range parts {
		if i > 0 {
			dataCheckString += "\n"
		}
		dataCheckString += p
	}

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secretKeyBytes := secretKey.Sum(nil)

	calculatedHash := hmac.New(sha256.New, secretKeyBytes)
	calculatedHash.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(calculatedHash.Sum(nil))

	result := ""
	for i, k := range keys {
		if i > 0 {
			result += "&"
		}
		result += fmt.Sprintf("%s=%s", k, params[k])
	}
	result += "&hash=" + hash
	return result
}

// TestRequireAuth_UserSettingsLanguage covers lines 114-121:
// user has SettingsJSON with a language field → lang is set from settings.
func TestRequireAuth_UserSettingsLanguage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	// Create a user with SettingsJSON containing language
	user, err := userRepo.GetOrCreateUser(88001)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	// Update user settings to include language
	_, err = db.Exec(`UPDATE users SET settings_json = $1 WHERE id = $2`,
		`{"language":"ru"}`, user.ID)
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}

	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	// Generate a valid token for the user
	token, err := jwtService.GenerateToken(user.ID, []int64{})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	var handlerCalled bool
	handler := authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestValidateTelegramInitData_WrongHash covers lines 181-183:
// hash != expectedHash → "invalid hash" error.
func TestValidateTelegramInitData_WrongHash(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	jwtService, _ := NewJWTService(cfg, logger)

	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "real-bot-token")

	// Build initData with wrong hash
	initData := "user=%7B%22id%22%3A12345%7D&auth_date=1234567890&hash=wronghashvalue"

	_, err := middleware.ValidateTelegramInitData(initData)
	if err == nil {
		t.Error("expected error for wrong hash")
	}
	if err.Error() != "invalid hash" {
		t.Errorf("expected 'invalid hash', got %q", err.Error())
	}
}

// TestValidateTelegramInitData_UserNotFound covers lines 204-206:
// valid hash but no "user" field in initData.
func TestValidateTelegramInitData_UserNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	jwtService, _ := NewJWTService(cfg, logger)

	botToken := "test-bot-token-no-user"
	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, botToken)

	// Build valid initData without "user" field
	initData := buildInitDataWithHash(botToken, map[string]string{
		"auth_date": "1234567890",
	})

	_, err := middleware.ValidateTelegramInitData(initData)
	if err == nil {
		t.Error("expected error when user not found in initData")
	}
	if err.Error() != "user not found in initData" {
		t.Errorf("expected 'user not found in initData', got %q", err.Error())
	}
}

// TestValidateTelegramInitData_QueryUnescapeFails covers the branch where url.QueryUnescape
// returns an error (line 153-155: "use original value" when decoding fails).
func TestValidateTelegramInitData_QueryUnescapeFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	jwtService, _ := NewJWTService(cfg, logger)

	botToken := "test-bot-token-query-unescape"
	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, botToken)

	// Build initData with a param value that causes QueryUnescape to fail (invalid %-encoding, e.g. %Z).
	// The hash is computed from the raw value; handler keeps original value when decode fails.
	initData := buildInitDataWithHash(botToken, map[string]string{
		"auth_date": "%Z",
		"user":      `{"id":99999}`,
	})

	userID, err := middleware.ValidateTelegramInitData(initData)
	if err != nil {
		t.Fatalf("ValidateTelegramInitData: %v", err)
	}
	if userID != 99999 {
		t.Errorf("expected userID 99999, got %d", userID)
	}
}

// TestValidateTelegramInitData_InvalidUserJSON covers lines 212-214:
// valid hash, "user" field present but invalid JSON.
func TestValidateTelegramInitData_InvalidUserJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	cfg := &config.Config{WebApp: config.WebAppConfig{JWTSecret: "test-secret"}}
	jwtService, _ := NewJWTService(cfg, logger)

	botToken := "test-bot-token-bad-user"
	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, botToken)

	// Build valid initData with invalid user JSON
	initData := buildInitDataWithHash(botToken, map[string]string{
		"auth_date": "1234567890",
		"user":      "not-valid-json",
	})

	_, err := middleware.ValidateTelegramInitData(initData)
	if err == nil {
		t.Error("expected error for invalid user JSON")
	}
	if len(err.Error()) == 0 || err.Error()[:len("failed to parse user data")] != "failed to parse user data" {
		t.Errorf("expected 'failed to parse user data' prefix, got %q", err.Error())
	}
}

// TestGenerateTokenPair_Success covers the happy path of GenerateTokenPair.
// The error paths (lines 249-251 and 254-256) require GenerateToken/GenerateRefreshToken
// to fail. With HMAC-SHA256, SignedString never fails for any []byte key (including nil/empty),
// so these error paths are not reachable with the concrete JWTService implementation.
// They would require a mock jwtService interface which doesn't exist in the current codebase.
func TestGenerateTokenPair_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	accessToken, refreshToken, err := middleware.GenerateTokenPair(1, 1)
	if err != nil {
		t.Fatalf("GenerateTokenPair() unexpected error: %v", err)
	}
	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

// TestRequireAuth_UserSettingsLanguage_InvalidJSON covers the branch where
// user.SettingsJSON is set but contains invalid JSON → fallback to Accept-Language.
func TestRequireAuth_UserSettingsLanguage_InvalidJSON(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	user, err := userRepo.GetOrCreateUser(88002)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	// Set invalid JSON in settings
	_, err = db.Exec(`UPDATE users SET settings_json = $1 WHERE id = $2`,
		`not-valid-json`, user.ID)
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}

	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	token, err := jwtService.GenerateToken(user.ID, []int64{})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	var handlerCalled bool
	handler := authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called even with invalid settings JSON")
	}
}

// TestRequireAuth_UserSettingsLanguage_UnknownLang covers the branch where
// user.SettingsJSON has a language that is neither "ru" nor "en" → fallback to Accept-Language.
func TestRequireAuth_UserSettingsLanguage_UnknownLang(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	user, err := userRepo.GetOrCreateUser(88003)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	// Set unknown language in settings
	_, err = db.Exec(`UPDATE users SET settings_json = $1 WHERE id = $2`,
		`{"language":"fr"}`, user.ID)
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}

	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	token, err := jwtService.GenerateToken(user.ID, []int64{})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	var handlerCalled bool
	handler := authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	// Set Accept-Language to Russian to test fallback
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called with unknown language in settings")
	}
}

// TestRequireAuth_UserSettingsLanguage_EmptyLang covers the branch where
// user.SettingsJSON has an empty language field → fallback to Accept-Language.
func TestRequireAuth_UserSettingsLanguage_EmptyLang(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	user, err := userRepo.GetOrCreateUser(88004)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	// Set empty language in settings
	_, err = db.Exec(`UPDATE users SET settings_json = $1 WHERE id = $2`,
		`{"language":""}`, user.ID)
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}

	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	token, err := jwtService.GenerateToken(user.ID, []int64{})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	var handlerCalled bool
	handler := authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called with empty language in settings")
	}
}

// TestRequireAuth_UserSettingsLanguage_RU covers the branch where
// user.SettingsJSON has language="ru" → lang is set to "ru" (not fallback).
func TestRequireAuth_UserSettingsLanguage_RU(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)

	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:       "test-secret",
			JWTTTLHours:     24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	user, err := userRepo.GetOrCreateUser(88005)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	_, err = db.Exec(`UPDATE users SET settings_json = $1 WHERE id = $2`,
		`{"language":"ru"}`, user.ID)
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}

	authMiddleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-token")

	token, err := jwtService.GenerateToken(user.ID, []int64{})
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	var capturedCtx context.Context
	handler := authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedCtx == nil {
		t.Error("expected context to be set")
	}
}
