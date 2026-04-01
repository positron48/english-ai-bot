//go:build integration

package llm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/service"
)

func TestLLM_TrainingCards(t *testing.T) {
	// Load training prompt
	trainingPrompt, err := loadPromptFromFile("prompts/training-card-generator.txt")
	if err != nil {
		t.Fatalf("Failed to load training prompt: %v", err)
	}

	// Load word prompt (needed for AI service initialization, but we'll use training prompt)
	wordPrompt, err := loadPromptFromFile("prompts/english-teacher.txt")
	if err != nil {
		t.Fatalf("Failed to load word prompt: %v", err)
	}

	// Setup AI service
	aiService := setupAIService(t, wordPrompt, trainingPrompt)

	// Load test cases
	testDataPath := "internal/integration/llm/testdata/llm_training_cases.json"
	cases := loadTestCases(t, testDataPath)

	if len(cases) == 0 {
		t.Skip("No test cases found")
	}

	ctx := context.Background()

	// Run each test case
	for _, tc := range cases {
		runTrainingTestCaseWithBusinessValidation(ctx, t, aiService, tc)
	}
}

// runTrainingTestCaseWithBusinessValidation runs a training test case and also validates using business logic
func runTrainingTestCaseWithBusinessValidation(ctx context.Context, t *testing.T, aiService *ai.Service, tc TestCase) {
	runTrainingTestCaseWithBusinessValidationForTarget(ctx, t, aiService, tc, "en")
}

func runTrainingTestCaseWithBusinessValidationForTarget(ctx context.Context, t *testing.T, aiService *ai.Service, tc TestCase, targetLang string) {
	t.Run(tc.Word, func(t *testing.T) {
		// Generate response
		response, err := aiService.GenerateTrainingCard(ctx, tc.Word)
		if err != nil {
			t.Fatalf("LLM generation failed: %v", err)
		}

		// Clean JSON
		response = cleanJSONResponse(response)

		// Parse JSON into structured model for business validation
		var trainingResp models.TrainingCardResponse
		if err := json.Unmarshal([]byte(response), &trainingResp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v\nResponse: %s", err, response)
		}

		// Parse JSON into map for schema validation (needs []interface{} format)
		var respMap map[string]interface{}
		if err := json.Unmarshal([]byte(response), &respMap); err != nil {
			t.Fatalf("Failed to parse JSON response as map: %v\nResponse: %s", err, response)
		}

		// Basic schema validation based on expect
		var validationErrors []string
		if tc.Expect == "ok" {
			validationErrors = validateTrainingResponseOK(t, tc.Word, respMap)
		} else if tc.Expect == "reject" {
			validationErrors = validateTrainingResponseReject(t, tc.Word, respMap)
		} else {
			t.Fatalf("Invalid expect value: %q (must be 'ok' or 'reject')", tc.Expect)
		}

		// If expect="ok", also run business validation
		if tc.Expect == "ok" && len(trainingResp.Senses) > 0 {
			// Create a minimal WordCard for validation
			wordCard := &models.WordCard{
				Word: tc.Word,
			}
			// Try to extract POS from first sense if available
			if len(trainingResp.Senses) > 0 && trainingResp.Senses[0].POS != "" {
				pos := trainingResp.Senses[0].POS
				wordCard.POS = &pos
			}

			// Run business validation (this is the same validation used in production)
			validationError := service.ValidateTrainingCardResponse(targetLang, wordCard, &trainingResp)
			if validationError != "" {
				validationErrors = append(validationErrors, "business_validation: "+validationError)
			}
		}

		if len(validationErrors) > 0 {
			t.Errorf("Validation failed for word %q (expect=%q):\n%s\nFull response: %s",
				tc.Word, tc.Expect, strings.Join(validationErrors, "\n"), response)
		}
	})
}
