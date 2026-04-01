package service

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type circuitBreakerStateGetter interface {
	GetState() (bool, int, string, error)
}

// NotifyCircuitBreakerOpened sends the same alert as native circuit opening.
func NotifyCircuitBreakerOpened(
	bot *tgbotapi.BotAPI,
	adminTelegramID int64,
	cbStateGetter circuitBreakerStateGetter,
	logger *zap.Logger,
) {
	if adminTelegramID == 0 {
		logger.Debug("admin telegram ID not set, skipping notification")
		return
	}
	if cbStateGetter == nil {
		logger.Warn("cannot send circuit breaker notification: state getter is nil")
		return
	}

	_, failureCount, lastError, err := cbStateGetter.GetState()
	if err != nil {
		logger.Error("failed to get circuit breaker state", zap.Error(err))
		return
	}

	message := fmt.Sprintf(
		"⚠️ Circuit Breaker ОТКРЫТ\n\n"+
			"Воркер генерации карточек остановлен.\n"+
			"Причина: %d последовательных ошибок LLM.\n\n"+
			"Последняя ошибка: %s\n"+
			"Время: %s\n\n"+
			"Для сброса используйте /reset_circuit",
		failureCount,
		lastError,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	if bot == nil {
		logger.Warn("cannot send admin notification: Telegram bot not initialized",
			zap.Int64("admin_id", adminTelegramID),
		)
		return
	}

	msg := tgbotapi.NewMessage(adminTelegramID, message)
	if _, err := bot.Send(msg); err != nil {
		logger.Error("failed to send admin notification", zap.Error(err))
		return
	}

	logger.Info("sent circuit breaker notification to admin",
		zap.Int64("admin_id", adminTelegramID),
	)
}
