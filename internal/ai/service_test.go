package ai

import (
	"testing"

	"go.uber.org/zap"
)

func TestStripLLMJSONFences(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{`{"a":1}`, `{"a":1}`},
		{"  {\"a\":1}  ", `{"a":1}`},
		{"```json\n{\"a\":1}\n```", `{"a":1}`},
		{"```JSON\n{\"a\":1}\n```", `{"a":1}`},
		{"```\n{\"a\":1}\n```", `{"a":1}`},
	}
	for _, tt := range tests {
		got := stripLLMJSONFences(tt.in)
		if got != tt.want {
			t.Errorf("stripLLMJSONFences(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

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
