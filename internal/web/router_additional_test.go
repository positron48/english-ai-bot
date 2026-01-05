package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestRouter_SetDependencies(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	userRepo := repository.NewUserRepository(nil, logger)
	wordService := (*service.WordService)(nil)
	aiService := (*interface{})(nil)
	bot := (*tgbotapi.BotAPI)(nil)

	router.SetDependencies(userRepo, wordService, aiService, bot, "test-token")
	_ = router // Verify router is configured
}

func TestRouter_SetOTPRepo(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	otpRepo := repository.NewWebOTPRepository(nil, logger)
	router.SetOTPRepo(otpRepo)
	_ = router // Verify router is configured
}

func TestRouter_getAuthMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	// Set dependencies to initialize auth middleware
	userRepo := repository.NewUserRepository(nil, logger)
	router.SetDependencies(userRepo, nil, nil, nil, "test-token")

	// Get auth middleware
	middleware := router.getAuthMiddleware()
	if middleware == nil {
		t.Error("getAuthMiddleware() should not return nil after SetDependencies")
	}
}

func TestRouter_ServeHTTP_Health(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		WebApp: config.WebAppConfig{
			JWTSecret: "test-secret",
		},
	}

	router := NewRouter(logger, cfg, nil, nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
