package web

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"tgbot-skeleton/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func TestNewJWTService(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	
	tests := []struct {
		name      string
		cfg       *config.Config
		wantError bool
	}{
		{
			name: "Valid config with JWT secret",
			cfg: &config.Config{
				WebApp: config.WebAppConfig{
					JWTSecret:     "test-secret",
					JWTTTLHours:   24,
					RefreshTTLHours: 720,
				},
			},
			wantError: false,
		},
		{
			name: "Valid config with session secret fallback",
			cfg: &config.Config{
				WebApp: config.WebAppConfig{
					SessionSecret: "session-secret",
					JWTTTLHours:   24,
					RefreshTTLHours: 720,
				},
			},
			wantError: false,
		},
		{
			name: "Missing secret",
			cfg: &config.Config{
				WebApp: config.WebAppConfig{},
			},
			wantError: true,
		},
		{
			name: "Default TTL values",
			cfg: &config.Config{
				WebApp: config.WebAppConfig{
					JWTSecret: "test-secret",
				},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewJWTService(tt.cfg, logger)
			if (err != nil) != tt.wantError {
				t.Errorf("NewJWTService() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && service == nil {
				t.Error("NewJWTService() should not return nil on success")
			}
		})
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	service, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	token, err := service.GenerateToken(12345, []int64{})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateToken() should return a non-empty token")
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	service, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	token, err := service.GenerateRefreshToken(12345)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateRefreshToken() should return a non-empty token")
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	service, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	userID := int64(12345)
	categories := []int64{1, 2}
	token, err := service.GenerateToken(userID, categories)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	t.Run("Valid token", func(t *testing.T) {
		validatedID, validatedCategories, err := service.ValidateToken(token)
		if err != nil {
			t.Fatalf("ValidateToken() error = %v", err)
		}
		if validatedID != userID {
			t.Errorf("ValidateToken() = %d, want %d", validatedID, userID)
		}
		if len(validatedCategories) != len(categories) {
			t.Errorf("ValidateToken() categories length = %d, want %d", len(validatedCategories), len(categories))
		}
		for i, cat := range categories {
			if i < len(validatedCategories) && validatedCategories[i] != cat {
				t.Errorf("ValidateToken() categories[%d] = %d, want %d", i, validatedCategories[i], cat)
			}
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		_, _, err := service.ValidateToken("invalid.token.here")
		if err == nil {
			t.Error("ValidateToken() should return error for invalid token")
		}
	})

	t.Run("Empty token", func(t *testing.T) {
		_, _, err := service.ValidateToken("")
		if err == nil {
			t.Error("ValidateToken() should return error for empty token")
		}
	})
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}

	service, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	userID := int64(12345)
	token, err := service.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	t.Run("Valid refresh token", func(t *testing.T) {
		validatedID, err := service.ValidateRefreshToken(token)
		if err != nil {
			t.Fatalf("ValidateRefreshToken() error = %v", err)
		}
		if validatedID != userID {
			t.Errorf("ValidateRefreshToken() = %d, want %d", validatedID, userID)
		}
	})

	t.Run("Invalid refresh token", func(t *testing.T) {
		_, err := service.ValidateRefreshToken("invalid.token.here")
		if err == nil {
			t.Error("ValidateRefreshToken() should return error for invalid token")
		}
	})
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantError bool
	}{
		{
			name:      "Valid Bearer token",
			header:    "Bearer test-token-123",
			wantToken: "test-token-123",
			wantError: false,
		},
		{
			name:      "Empty header",
			header:    "",
			wantError: true,
		},
		{
			name:      "Missing Bearer prefix",
			header:    "test-token-123",
			wantError: true,
		},
		{
			name:      "Empty token after Bearer",
			header:    "Bearer ",
			wantError: true,
		},
		{
			name:      "Token with spaces",
			header:    "Bearer token with spaces",
			wantToken: "token with spaces",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.header)
			if (err != nil) != tt.wantError {
				t.Errorf("ExtractTokenFromHeader() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && token != tt.wantToken {
				t.Errorf("ExtractTokenFromHeader() = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestJWTService_TokenExpiration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key",
			JWTTTLHours:   1, // 1 hour
			RefreshTTLHours: 720,
		},
	}

	service, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	// Generate token
	token, err := service.GenerateToken(12345, []int64{})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	// Token should be valid immediately
	_, _, err = service.ValidateToken(token)
	if err != nil {
		t.Errorf("ValidateToken() should succeed for fresh token, got error: %v", err)
	}
}

func TestJWTService_ValidateToken_Expired(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	svc, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	// Build a token with expiry in the past (same structure as GenerateToken)
	roleJSON, _ := json.Marshal(RoleClaim{Categories: []int64{1}})
	expiredAt := time.Now().Add(-time.Hour)
	claims := &Claims{
		UserID: 99,
		Role:   roleJSON,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiredAt),
			IssuedAt:  jwt.NewNumericDate(expiredAt.Add(-time.Hour)),
			NotBefore: jwt.NewNumericDate(expiredAt.Add(-time.Hour)),
			Issuer:    "english-bot",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.WebApp.JWTSecret))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}

	_, _, err = svc.ValidateToken(tokenString)
	if err == nil {
		t.Error("ValidateToken() should return error for expired token")
	}
	if err != nil && !strings.Contains(err.Error(), "expired") {
		t.Errorf("ValidateToken() error should mention expired, got: %v", err)
	}
}

func TestJWTService_ValidateToken_LegacyRoleString(t *testing.T) {
	logger := zap.NewNop()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret-key",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	svc, err := NewJWTService(cfg, logger)
	if err != nil {
		t.Fatalf("NewJWTService: %v", err)
	}

	// Legacy token: role is a JSON string (e.g. "admin") instead of object
	roleJSON := json.RawMessage(`"admin"`)
	expiresAt := time.Now().Add(time.Hour)
	claims := &Claims{
		UserID: 42,
		Role:   roleJSON,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "english-bot",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.WebApp.JWTSecret))
	if err != nil {
		t.Fatalf("SignedString: %v", err)
	}

	userID, categories, err := svc.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken(legacy role) error: %v", err)
	}
	if userID != 42 {
		t.Errorf("ValidateToken() userID = %d, want 42", userID)
	}
	// Legacy format should return empty categories
	if len(categories) != 0 {
		t.Errorf("ValidateToken() legacy role should return empty categories, got %v", categories)
	}
}
