package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestNewAuthMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(nil, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)

	middleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-bot-token")
	_ = middleware // Verify middleware is created
}

func TestAuthMiddleware_ValidateTelegramInitData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(nil, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	botToken := "test-bot-token-12345"

	middleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, botToken)

	// Create valid initData
	// Format: key1=value1&key2=value2&hash=...
	userID := int64(12345)
	authDate := "1234567890"
	
	// Create user JSON
	userJSON, _ := json.Marshal(map[string]int64{"id": userID})
	userEncoded := url.QueryEscape(string(userJSON))
	
	// Build params map (without hash)
	params := map[string]string{
		"auth_date": authDate,
		"user":      string(userJSON), // Not URL-encoded in data_check_string
	}
	
	// Build data_check_string (sorted keys, not URL-encoded values)
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	var dataCheckParts []string
	for _, k := range keys {
		dataCheckParts = append(dataCheckParts, k+"="+params[k])
	}
	dataCheckString := strings.Join(dataCheckParts, "\n")
	
	// Calculate hash
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secretKeyBytes := secretKey.Sum(nil)
	
	hash := hmac.New(sha256.New, secretKeyBytes)
	hash.Write([]byte(dataCheckString))
	hashHex := hex.EncodeToString(hash.Sum(nil))
	
	// Build initData with URL-encoded values
	initData := "auth_date=" + url.QueryEscape(authDate) + "&user=" + userEncoded + "&hash=" + hashHex

	validatedID, err := middleware.ValidateTelegramInitData(initData)
	if err != nil {
		t.Fatalf("ValidateTelegramInitData() error = %v", err)
	}
	if validatedID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, validatedID)
	}
}

func TestAuthMiddleware_ValidateTelegramInitData_InvalidHash(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(nil, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)

	middleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-bot-token")

	// Invalid initData with wrong hash
	initData := "auth_date=1234567890&user=%7B%22id%22%3A12345%7D&hash=invalidhash"

	_, err := middleware.ValidateTelegramInitData(initData)
	if err == nil {
		t.Error("ValidateTelegramInitData() should return error for invalid hash")
	}
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(nil, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)

	middleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-bot-token")

	// Generate a valid token
	userID := int64(999)
	token, err := jwtService.GenerateToken(userID, "user")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Create a request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Test handler that checks user ID from context
	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		ctxUserID := r.Context().Value(userIDKey)
		if ctxUserID == nil {
			t.Error("User ID should be in context")
		}
		if ctxUserID.(int64) != userID {
			t.Errorf("Expected UserID %d in context, got %d", userID, ctxUserID.(int64))
		}
	}

	protectedHandler := middleware.RequireAuth(testHandler)
	protectedHandler(w, req)

	if !handlerCalled {
		t.Error("Handler should be called with valid token")
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_RequireAuth_NoToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(nil, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)

	middleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-bot-token")

	// Create a request without token
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}

	protectedHandler := middleware.RequireAuth(testHandler)
	protectedHandler(w, req)

	if handlerCalled {
		t.Error("Handler should not be called without token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_RequireAuth_InvalidToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	userRepo := repository.NewUserRepository(nil, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)

	middleware := NewAuthMiddleware(userRepo, jwtService, logger, cfg, "test-bot-token")

	// Create a request with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}

	protectedHandler := middleware.RequireAuth(testHandler)
	protectedHandler(w, req)

	if handlerCalled {
		t.Error("Handler should not be called with invalid token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
