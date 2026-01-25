package web

import (
	"database/sql"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func TestAuthMiddleware_GenerateJWTToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	userRepo := repository.NewUserRepository(db, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)

	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-bot-token")

	token, err := middleware.GenerateJWTToken(12345, 12345)
	if err != nil {
		t.Fatalf("GenerateJWTToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateJWTToken() should return non-empty token")
	}
}

func TestAuthMiddleware_GenerateTokenPair(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	userRepo := repository.NewUserRepository(db, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)

	accessCategoryRepo := repository.NewUserAccessCategoryRepository(db, logger)
	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-bot-token")

	accessToken, refreshToken, err := middleware.GenerateTokenPair(67890, 67890)
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
