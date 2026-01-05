package ai

import (
	"testing"

	"go.uber.org/zap"
)

func TestNewService(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://test.ai", "test-model", "test-key", "test prompt", logger)
	_ = service // Verify service is created
}
