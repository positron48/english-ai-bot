package ai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestPolzaAISmoke calls the live Polza chat/completions API when RUN_POLZA_AI=1.
//
//	RUN_POLZA_AI=1 go test ./internal/ai -run TestPolzaAISmoke -v -count=1 -timeout 120s
func TestPolzaAISmoke(t *testing.T) {
	if os.Getenv("RUN_POLZA_AI") != "1" {
		t.Skip("manual smoke; set RUN_POLZA_AI=1 and POLZA_AI_API_KEY to run")
	}

	apiKey := strings.TrimSpace(os.Getenv("POLZA_AI_API_KEY"))
	if apiKey == "" {
		t.Fatal("POLZA_AI_API_KEY is required")
	}

	baseURL := strings.TrimSpace(os.Getenv("POLZA_AI_URL"))
	if baseURL == "" {
		baseURL = "https://polza.ai/api/v1"
	}

	model := strings.TrimSpace(os.Getenv("POLZA_TEST_MODEL"))
	if model == "" {
		model = "openai/gpt-4o-mini"
	}

	svc := NewServiceWithTimeout(baseURL, model, apiKey, "", 90*time.Second, zap.NewNop())
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	reply, err := svc.ChatSystemUser(ctx, "You are a helpful assistant.", "Ответь одним словом: работает?")
	if err != nil {
		t.Fatalf("ChatSystemUser: %v", err)
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		t.Fatal("empty reply from Polza")
	}
	t.Logf("polza reply: %q", reply)
}
