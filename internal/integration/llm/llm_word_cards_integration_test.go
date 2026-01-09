//go:build integration

package llm

import (
	"context"
	"testing"
)

func TestLLM_WordCards(t *testing.T) {
	// Load word prompt
	wordPrompt, err := loadPromptFromFile("prompts/english-teacher.txt")
	if err != nil {
		t.Fatalf("Failed to load word prompt: %v", err)
	}

	// Setup AI service
	aiService := setupAIService(t, wordPrompt, "")

	// Load test cases
	testDataPath := "internal/integration/llm/testdata/llm_word_cases.json"
	cases := loadTestCases(t, testDataPath)

	if len(cases) == 0 {
		t.Skip("No test cases found")
	}

	ctx := context.Background()

	// Run each test case
	for _, tc := range cases {
		runWordTestCase(ctx, t, aiService, wordPrompt, tc)
	}
}
