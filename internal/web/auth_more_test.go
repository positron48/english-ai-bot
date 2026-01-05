package web

import (
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestAuthMiddleware_GenerateJWTToken(t *testing.T) {
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

	token, err := middleware.GenerateJWTToken(12345)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateJWTToken() should return non-empty token")
	}
}

func TestAuthMiddleware_GenerateTokenPair(t *testing.T) {
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

	accessToken, refreshToken, err := middleware.GenerateTokenPair(67890)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	if accessToken == "" {
		t.Error("GenerateTokenPair() should return non-empty access token")
	}
	if refreshToken == "" {
		t.Error("GenerateTokenPair() should return non-empty refresh token")
	}
}
