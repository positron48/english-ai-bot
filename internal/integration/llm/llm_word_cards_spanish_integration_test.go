//go:build integration

package llm

import (
	"context"
	"testing"
)

func TestLLM_WordCards_ES(t *testing.T) {
	// Load Spanish word prompt (RU -> ES teacher)
	wordPrompt, err := loadPromptFromFile("prompts/teacher-ru-es.txt")
	if err != nil {
		t.Fatalf("Failed to load Spanish word prompt: %v", err)
	}

	aiService := setupAIService(t, wordPrompt, "")

	// Spanish regression pack (known unstable scenarios)
	testDataPath := "internal/integration/llm/testdata/llm_word_cases_es.json"
	cases := loadTestCases(t, testDataPath)
	if len(cases) == 0 {
		t.Skip("No Spanish word test cases found")
	}

	ctx := context.Background()
	for _, tc := range cases {
		runWordTestCase(ctx, t, aiService, wordPrompt, tc)
	}
}

