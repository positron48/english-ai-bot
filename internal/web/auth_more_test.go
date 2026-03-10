package web

import (
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"

	"tgbot-skeleton/internal/testutil"
	"go.uber.org/zap"
)

func TestAuthMiddleware_GenerateJWTToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db := testutil.SetupTestDB(t)
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
	db := testutil.SetupTestDB(t)
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

// TestAuthMiddleware_GenerateJWTToken_GetUserCategoriesFails covers the branch when getUserCategories fails (uses empty categories).
func TestAuthMiddleware_GenerateJWTToken_GetUserCategoriesFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	goodDB := testutil.SetupTestDB(t)
	badDB := badDBConn(t)

	userRepo := repository.NewUserRepository(goodDB, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(badDB, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-bot-token")

	token, err := middleware.GenerateJWTToken(1, 1)
	if err != nil {
		t.Fatalf("GenerateJWTToken() should succeed with empty categories on GetUserCategories failure: %v", err)
	}
	if token == "" {
		t.Error("GenerateJWTToken() should return non-empty token")
	}
}

// TestAuthMiddleware_GenerateTokenPair_GetUserCategoriesFails covers the branch when getUserCategories fails (uses empty categories).
func TestAuthMiddleware_GenerateTokenPair_GetUserCategoriesFails(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	goodDB := testutil.SetupTestDB(t)
	badDB := badDBConn(t)

	userRepo := repository.NewUserRepository(goodDB, logger)
	accessCategoryRepo := repository.NewUserAccessCategoryRepository(badDB, logger)
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret:     "test-secret",
			JWTTTLHours:   24,
			RefreshTTLHours: 720,
		},
	}
	jwtService, _ := NewJWTService(cfg, logger)
	middleware := NewAuthMiddleware(userRepo, accessCategoryRepo, jwtService, logger, cfg, "test-bot-token")

	accessToken, refreshToken, err := middleware.GenerateTokenPair(1, 1)
	if err != nil {
		t.Fatalf("GenerateTokenPair() should succeed with empty categories on GetUserCategories failure: %v", err)
	}
	if accessToken == "" || refreshToken == "" {
		t.Error("GenerateTokenPair() should return non-empty tokens")
	}
}
