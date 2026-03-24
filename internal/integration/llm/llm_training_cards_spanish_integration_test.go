//go:build integration

package llm

import (
	"context"
	"testing"
)

func TestLLM_TrainingCards_ES(t *testing.T) {
	// Load Spanish training-card prompt (RU -> ES)
	trainingPrompt, err := loadPromptFromFile("prompts/training-card-ru-es.txt")
	if err != nil {
		t.Fatalf("Failed to load Spanish training prompt: %v", err)
	}

	// Word prompt is still required by service constructor, but not used in GenerateTrainingCard.
	aiService := setupAIService(t, "You are a helpful assistant.", trainingPrompt)

	testDataPath := "internal/integration/llm/testdata/llm_training_cases_es.json"
	cases := loadTestCases(t, testDataPath)
	if len(cases) == 0 {
		t.Skip("No Spanish training test cases found")
	}

	ctx := context.Background()
	for _, tc := range cases {
		runTrainingTestCase(ctx, t, aiService, tc)
	}
}

