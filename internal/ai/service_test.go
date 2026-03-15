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

func TestSetTrainingPrompt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	service := NewService("http://test.ai", "test-model", "test-key", "test prompt", logger)

	prompt := "Generate training card for word:"
	service.SetTrainingPrompt(prompt)

	if service.trainingPrompt != prompt {
		t.Errorf("SetTrainingPrompt() failed, got %v, want %v", service.trainingPrompt, prompt)
	}
}
