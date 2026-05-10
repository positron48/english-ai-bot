package ai

import (
	"testing"
	"time"

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

func TestParseHTTPTimeout(t *testing.T) {
	t.Parallel()
	if d := ParseHTTPTimeout(""); d != 0 {
		t.Fatalf("empty: got %v want 0", d)
	}
	if d := ParseHTTPTimeout("2m"); d != 2*60*time.Second {
		t.Fatalf("2m: got %v", d)
	}
	if d := ParseHTTPTimeout("bogus"); d != 0 {
		t.Fatalf("invalid: got %v want 0", d)
	}
}

func TestNewServiceWithTimeout(t *testing.T) {
	t.Parallel()
	logger := zap.NewNop()
	s := NewServiceWithTimeout("http://x", "m", "k", "p", 5*time.Minute, logger)
	if s.client.Timeout != 5*time.Minute {
		t.Fatalf("timeout: got %v", s.client.Timeout)
	}
	s2 := NewServiceWithTimeout("http://x", "m", "k", "p", 0, logger)
	if s2.client.Timeout != DefaultHTTPTimeout {
		t.Fatalf("zero timeout: got %v want default", s2.client.Timeout)
	}
}

func TestNewService(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	service := NewService("http://test.ai", "test-model", "test-key", "test prompt", logger)
	if service.client.Timeout != DefaultHTTPTimeout {
		t.Fatalf("default client timeout: got %v want %v", service.client.Timeout, DefaultHTTPTimeout)
	}
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
