package testkit

import (
	"testing"
)

func TestTestConfig(t *testing.T) {
	cfg := testConfig()
	if cfg == nil {
		t.Fatal("testConfig() returned nil")
	}
	if cfg.WebApp.JWTSecret != "test-jwt-secret-for-integration" {
		t.Errorf("WebApp.JWTSecret = %q, want test-jwt-secret-for-integration", cfg.WebApp.JWTSecret)
	}
	if cfg.WebApp.JWTTTLHours != 24 {
		t.Errorf("WebApp.JWTTTLHours = %d, want 24", cfg.WebApp.JWTTTLHours)
	}
	if cfg.WebApp.RateLimitAppAPIPerUser != 2000 {
		t.Errorf("RateLimitAppAPIPerUser = %d, want 2000", cfg.WebApp.RateLimitAppAPIPerUser)
	}
	if cfg.Admin.TelegramID != 0 {
		t.Errorf("Admin.TelegramID = %d, want 0", cfg.Admin.TelegramID)
	}
}

func TestWithAIService(t *testing.T) {
	c := &harnessConfig{}
	opt := WithAIService("http://mock:8080", "api-key", "You are a test.")
	opt(c)
	if c.aiServiceURL != "http://mock:8080" {
		t.Errorf("aiServiceURL = %q, want http://mock:8080", c.aiServiceURL)
	}
	if c.aiAPIKey != "api-key" {
		t.Errorf("aiAPIKey = %q, want api-key", c.aiAPIKey)
	}
	if c.aiPrompt != "You are a test." {
		t.Errorf("aiPrompt = %q, want You are a test.", c.aiPrompt)
	}
}

func TestHarness_GetConnection(t *testing.T) {
	h := &Harness{DB: nil}
	got := h.GetConnection()
	if got != nil {
		t.Errorf("GetConnection() = %v, want nil", got)
	}
}
