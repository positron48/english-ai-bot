package web

import (
	"testing"

	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/service"

	"go.uber.org/zap"
)

func TestNewWebTrainingHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	trainingService := (*service.TrainingService)(nil)
	srsService := (*service.SRSService)(nil)
	optionsService := (*service.OptionsService)(nil)
	sessionRepo := (*repository.SessionRepository)(nil)

	handler := NewWebTrainingHandler(
		trainingService,
		srsService,
		optionsService,
		sessionRepo,
		logger,
		2000,
		3,
	)
	_ = handler // Verify handler is created
}
