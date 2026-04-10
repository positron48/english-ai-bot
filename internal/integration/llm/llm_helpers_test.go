//go:build integration

package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"

	"tgbot-skeleton/internal/ai"

	"go.uber.org/zap"
)

// TestCase represents a single test case
type TestCase struct {
	Word   string `json:"word"`
	Expect string `json:"expect"` // "ok" or "reject"
	Note   string `json:"note,omitempty"`
}

// findProjectRoot finds the project root directory by looking for go.mod file
func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	dir := wd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found (go.mod not found from %s)", wd)
}

// loadTestCases loads test cases from a JSON file
func loadTestCases(t *testing.T, filePath string) []TestCase {
	// If path is relative, make it relative to project root
	if !filepath.IsAbs(filePath) {
		root, err := findProjectRoot()
		if err != nil {
			t.Fatalf("Failed to find project root: %v", err)
		}
		filePath = filepath.Join(root, filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test cases file %s: %v", filePath, err)
	}

	var cases []TestCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("Failed to parse test cases JSON: %v", err)
	}

	return cases
}

// loadPromptFromFile loads a prompt from a file
func loadPromptFromFile(filePath string) (string, error) {
	// If path is relative, make it relative to project root
	if !filepath.IsAbs(filePath) {
		root, err := findProjectRoot()
		if err != nil {
			return "", fmt.Errorf("failed to find project root: %w", err)
		}
		filePath = filepath.Join(root, filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file: %w", err)
	}
	return strings.TrimSpace(string(content)), nil
}

// setupAIService creates an AI service for testing
func setupAIService(t *testing.T, wordPrompt, trainingPrompt string) *ai.Service {
	// Check required env vars first
	aiURL := os.Getenv("AI_URL")
	aiAPIKey := os.Getenv("AI_API_KEY")
	if aiURL == "" {
		t.Skip("AI_URL not set, skipping integration test")
	}
	if aiAPIKey == "" {
		t.Skip("AI_API_KEY not set, skipping integration test")
	}

	// Get model from env or use default
	aiModel := os.Getenv("AI_MODEL")
	if aiModel == "" {
		aiModel = "gpt-3.5-turbo" // Default from config
	}

	logger, _ := zap.NewDevelopment()
	service := ai.NewService(aiURL, aiModel, aiAPIKey, wordPrompt, logger)
	if trainingPrompt != "" {
		service.SetTrainingPrompt(trainingPrompt)
	}

	return service
}

// cleanJSONResponse removes markdown code fences from JSON response
func cleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)
	return response
}

// containsCyrillic checks if a string contains any Cyrillic characters
func containsCyrillic(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}

// validateWordResponseOK validates a word response when expect="ok"
func validateWordResponseOK(t *testing.T, word string, resp map[string]interface{}) []string {
	var errors []string

	// Check required fields are non-empty
	if lemma, ok := resp["lemma"].(string); !ok || lemma == "" {
		errors = append(errors, "lemma is missing or empty")
	}
	if pos, ok := resp["pos"].(string); !ok || pos == "" {
		errors = append(errors, "pos is missing or empty")
	}
	if transcription, ok := resp["transcription"].(string); !ok || transcription == "" {
		errors = append(errors, "transcription is missing or empty")
	}
	defRU, _ := resp["definition_ru"].(string)
	defNat, _ := resp["definition_native"].(string)
	if strings.TrimSpace(defRU) == "" && strings.TrimSpace(defNat) == "" {
		errors = append(errors, "definition_ru or definition_native is missing or empty")
	}

	// Check hint is empty
	if hint, ok := resp["hint"].(string); ok && hint != "" {
		errors = append(errors, fmt.Sprintf("hint should be empty for ok response, got: %q", hint))
	}

	// Check error is false
	errorVal := resp["error"]
	if errorBool, ok := errorVal.(bool); ok && errorBool {
		errors = append(errors, "error should be false for ok response")
	} else if errorStr, ok := errorVal.(string); ok && strings.ToLower(strings.TrimSpace(errorStr)) == "true" {
		errors = append(errors, "error should be false for ok response")
	}

	// Check examples
	examples, ok := resp["examples"].([]interface{})
	if !ok {
		errors = append(errors, "examples is missing or not an array")
	} else {
		if len(examples) < 2 || len(examples) > 3 {
			errors = append(errors, fmt.Sprintf("examples should have 2-3 items, got %d", len(examples)))
		}
		for i, ex := range examples {
			exMap, ok := ex.(map[string]interface{})
			if !ok {
				errors = append(errors, fmt.Sprintf("examples[%d] is not an object", i))
				continue
			}
			exText, _ := exMap["example_target"].(string)
			if strings.TrimSpace(exText) == "" {
				exText, _ = exMap["example_en"].(string)
			}
			if strings.TrimSpace(exText) == "" {
				errors = append(errors, fmt.Sprintf("examples[%d].example_target (or legacy example_en) is missing or empty", i))
			}
			gloss, _ := exMap["gloss_native"].(string)
			if strings.TrimSpace(gloss) == "" {
				gloss, _ = exMap["gloss_ru"].(string)
			}
			if strings.TrimSpace(gloss) == "" {
				errors = append(errors, fmt.Sprintf("examples[%d].gloss_native (or legacy gloss_ru) is missing or empty", i))
			}
		}
	}

	// Check verb_forms: should be present only if pos == "verb"
	pos, _ := resp["pos"].(string)
	verbForms, hasVerbForms := resp["verb_forms"]
	if pos == "verb" {
		if !hasVerbForms || verbForms == nil {
			errors = append(errors, "verb_forms should be present for verb")
		} else {
			// Validate verb_forms structure
			vfMap, ok := verbForms.(map[string]interface{})
			if !ok {
				errors = append(errors, "verb_forms is not an object")
			} else {
				if v1, ok := vfMap["v1"].(string); !ok || v1 == "" {
					errors = append(errors, "verb_forms.v1 is missing or empty")
				}
				if v2, ok := vfMap["v2"].(string); !ok || v2 == "" {
					errors = append(errors, "verb_forms.v2 is missing or empty")
				}
				if v3, ok := vfMap["v3"].(string); !ok || v3 == "" {
					errors = append(errors, "verb_forms.v3 is missing or empty")
				}
			}
		}
	} else {
		// For non-verbs, verb_forms should be absent OR empty object {} is acceptable
		if hasVerbForms && verbForms != nil {
			// Check if it's an empty object
			vfMap, ok := verbForms.(map[string]interface{})
			if !ok || len(vfMap) > 0 {
				// Not an empty object or not a map at all
				errors = append(errors, fmt.Sprintf("verb_forms should not be present for pos=%q (or should be empty object {})", pos))
			}
			// Empty object {} is acceptable, so no error in that case
		}
	}

	return errors
}

// validateWordResponseReject validates a word response when expect="reject"
func validateWordResponseReject(t *testing.T, word string, resp map[string]interface{}) []string {
	var errors []string

	// Check error is true
	errorVal := resp["error"]
	errorIsTrue := false
	if errorBool, ok := errorVal.(bool); ok && errorBool {
		errorIsTrue = true
	} else if errorStr, ok := errorVal.(string); ok && strings.ToLower(strings.TrimSpace(errorStr)) == "true" {
		errorIsTrue = true
	}
	if !errorIsTrue {
		errors = append(errors, "error should be true for reject response")
	}

	// Check hint is non-empty and contains Cyrillic
	hint, ok := resp["hint"].(string)
	if !ok || hint == "" {
		errors = append(errors, "hint should be non-empty for reject response")
	} else if !containsCyrillic(hint) {
		errors = append(errors, fmt.Sprintf("hint should contain Cyrillic characters, got: %q", hint))
	}

	// Check examples is empty array
	examples, ok := resp["examples"].([]interface{})
	if !ok {
		errors = append(errors, "examples should be an array")
	} else if len(examples) != 0 {
		errors = append(errors, fmt.Sprintf("examples should be empty for reject response, got %d items", len(examples)))
	}

	// Check verb_forms is absent
	if verbForms, hasVerbForms := resp["verb_forms"]; hasVerbForms && verbForms != nil {
		errors = append(errors, "verb_forms should not be present for reject response")
	}

	return errors
}

// validateWordResponseTranslate validates a word response when expect="translate"
// Response should NOT be JSON, should be plain text with Latin characters
func validateWordResponseTranslate(t *testing.T, word string, response string) []string {
	var errors []string

	// Try to parse as JSON - should FAIL
	var tempJSON map[string]interface{}
	if err := json.Unmarshal([]byte(response), &tempJSON); err == nil {
		// Successfully parsed as JSON - this is wrong!
		errors = append(errors, "response should NOT be JSON for translate mode, but valid JSON was returned")
	}

	// Check response contains Latin characters
	if !containsLatin(response) {
		errors = append(errors, "response should contain Latin characters (English translation)")
	}

	// Response should not be empty
	response = strings.TrimSpace(response)
	if response == "" {
		errors = append(errors, "response should not be empty")
	}

	return errors
}

// validateWordResponseCorrection validates a word response when expect="correction"
// Response should NOT be JSON, should be plain text with Latin characters
func validateWordResponseCorrection(t *testing.T, word string, response string) []string {
	var errors []string

	// Try to parse as JSON - should FAIL
	var tempJSON map[string]interface{}
	if err := json.Unmarshal([]byte(response), &tempJSON); err == nil {
		// Successfully parsed as JSON - this is wrong!
		errors = append(errors, "response should NOT be JSON for correction mode, but valid JSON was returned")
	}

	// Check response contains Latin characters
	if !containsLatin(response) {
		errors = append(errors, "response should contain Latin characters (English correction)")
	}

	// Response should not be empty
	response = strings.TrimSpace(response)
	if response == "" {
		errors = append(errors, "response should not be empty")
	}

	return errors
}

// containsLatin checks if a string contains any Latin characters
func containsLatin(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Latin, r) {
			return true
		}
	}
	return false
}

// validateTrainingResponseOK validates a training response when expect="ok"
func validateTrainingResponseOK(t *testing.T, word string, resp map[string]interface{}) []string {
	var errors []string

	// Check transcription is present for accepted cards
	if transcription, ok := resp["transcription"].(string); !ok || strings.TrimSpace(transcription) == "" {
		errors = append(errors, "transcription is missing or empty")
	}

	// Check error is empty
	if errorVal, ok := resp["error"].(string); ok && errorVal != "" {
		errors = append(errors, fmt.Sprintf("error should be empty for ok response, got: %q", errorVal))
	}

	// Check senses is non-empty array
	senses, ok := resp["senses"].([]interface{})
	if !ok {
		errors = append(errors, "senses is missing or not an array")
	} else if len(senses) == 0 {
		errors = append(errors, "senses should have at least 1 item for ok response")
	} else {
		// Validate each sense
		for i, sense := range senses {
			senseMap, ok := sense.(map[string]interface{})
			if !ok {
				errors = append(errors, fmt.Sprintf("senses[%d] is not an object", i))
				continue
			}

			// Check required fields
			if pos, ok := senseMap["pos"].(string); !ok || pos == "" {
				errors = append(errors, fmt.Sprintf("senses[%d].pos is missing or empty", i))
			}
			if displayWord, ok := senseMap["display_word"].(string); !ok || displayWord == "" {
				errors = append(errors, fmt.Sprintf("senses[%d].display_word is missing or empty", i))
			}
			wordNative, _ := senseMap["word_native"].(string)
			wordRU, _ := senseMap["word_ru"].(string)
			if strings.TrimSpace(wordNative) == "" && strings.TrimSpace(wordRU) == "" {
				errors = append(errors, fmt.Sprintf("senses[%d].word_native (or legacy word_ru) is missing or empty", i))
			}
			meaningTarget, _ := senseMap["meaning_target"].(string)
			meaningEN, _ := senseMap["meaning_en"].(string)
			if strings.TrimSpace(meaningTarget) == "" && strings.TrimSpace(meaningEN) == "" {
				errors = append(errors, fmt.Sprintf("senses[%d].meaning_target (or legacy meaning_en) is missing or empty", i))
			}

			// Check distractors
			distractorsRU, ok := senseMap["distractors_ru"].([]interface{})
			if !ok {
				errors = append(errors, fmt.Sprintf("senses[%d].distractors_ru is missing or not an array", i))
			} else if len(distractorsRU) != 3 {
				errors = append(errors, fmt.Sprintf("senses[%d].distractors_ru should have exactly 3 items, got %d", i, len(distractorsRU)))
			}

			distractorsTarget, okTarget := senseMap["distractors_target"].([]interface{})
			distractorsEN, okLegacy := senseMap["distractors_en"].([]interface{})
			if !okTarget && !okLegacy {
				errors = append(errors, fmt.Sprintf("senses[%d].distractors_target (or legacy distractors_en) is missing or not an array", i))
			} else {
				list := distractorsTarget
				if !okTarget {
					list = distractorsEN
				}
				if len(list) != 3 {
					errors = append(errors, fmt.Sprintf("senses[%d].distractors_target (or legacy distractors_en) should have exactly 3 items, got %d", i, len(list)))
				}
			}
		}
	}

	// Check lemma/word_en is non-empty
	lemma, hasLemma := resp["lemma"].(string)
	wordEN, hasWordEN := resp["word_en"].(string)
	if (!hasLemma || lemma == "") && (!hasWordEN || wordEN == "") {
		errors = append(errors, "lemma or word_en should be non-empty")
	}

	return errors
}

// validateTrainingResponseReject validates a training response when expect="reject"
func validateTrainingResponseReject(t *testing.T, word string, resp map[string]interface{}) []string {
	var errors []string

	// Check error is non-empty string
	errorVal, ok := resp["error"].(string)
	if !ok || errorVal == "" {
		errors = append(errors, "error should be a non-empty string for reject response")
	}

	// Check senses is empty array
	senses, ok := resp["senses"].([]interface{})
	if !ok {
		errors = append(errors, "senses should be an array")
	} else if len(senses) != 0 {
		errors = append(errors, fmt.Sprintf("senses should be empty for reject response, got %d items", len(senses)))
	}

	// For rejected response transcription must be empty
	if transcription, ok := resp["transcription"].(string); !ok || strings.TrimSpace(transcription) != "" {
		errors = append(errors, "transcription should be empty for reject response")
	}

	return errors
}

// runWordTestCase runs a single word test case
func runWordTestCase(ctx context.Context, t *testing.T, aiService *ai.Service, wordPrompt string, tc TestCase) {
	t.Run(tc.Word, func(t *testing.T) {
		// Generate response
		response, err := aiService.GenerateResponse(ctx, tc.Word)
		if err != nil {
			t.Fatalf("LLM generation failed: %v", err)
		}

		// For translate and correction modes, response should NOT be JSON
		if tc.Expect == "translate" || tc.Expect == "correction" {
			// Don't clean JSON fences for these modes - they should be plain text
			response = strings.TrimSpace(response)

			var validationErrors []string
			if tc.Expect == "translate" {
				validationErrors = validateWordResponseTranslate(t, tc.Word, response)
			} else if tc.Expect == "correction" {
				validationErrors = validateWordResponseCorrection(t, tc.Word, response)
			}

			if len(validationErrors) > 0 {
				t.Errorf("Validation failed for word %q (expect=%q):\n%s\nFull response: %s",
					tc.Word, tc.Expect, strings.Join(validationErrors, "\n"), response)
			}
			return
		}

		// For ok and reject modes, response should be JSON
		// Clean JSON
		response = cleanJSONResponse(response)

		// Parse JSON
		var resp map[string]interface{}
		if err := json.Unmarshal([]byte(response), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v\nResponse: %s", err, response)
		}

		// Validate based on expect
		var validationErrors []string
		if tc.Expect == "ok" {
			validationErrors = validateWordResponseOK(t, tc.Word, resp)
		} else if tc.Expect == "reject" {
			validationErrors = validateWordResponseReject(t, tc.Word, resp)
		} else {
			t.Fatalf("Invalid expect value: %q (must be 'ok', 'reject', 'translate', or 'correction')", tc.Expect)
		}

		if len(validationErrors) > 0 {
			t.Errorf("Validation failed for word %q (expect=%q):\n%s\nFull response: %s",
				tc.Word, tc.Expect, strings.Join(validationErrors, "\n"), response)
		}
	})
}

// runTrainingTestCase runs a single training test case
func runTrainingTestCase(ctx context.Context, t *testing.T, aiService *ai.Service, tc TestCase) {
	t.Run(tc.Word, func(t *testing.T) {
		// Generate response
		response, err := aiService.GenerateTrainingCard(ctx, tc.Word)
		if err != nil {
			t.Fatalf("LLM generation failed: %v", err)
		}

		// Clean JSON (already done in GenerateTrainingCard, but do it again for safety)
		response = cleanJSONResponse(response)

		// Parse JSON
		var resp map[string]interface{}
		if err := json.Unmarshal([]byte(response), &resp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v\nResponse: %s", err, response)
		}

		// Validate based on expect
		var validationErrors []string
		if tc.Expect == "ok" {
			validationErrors = validateTrainingResponseOK(t, tc.Word, resp)
		} else if tc.Expect == "reject" {
			validationErrors = validateTrainingResponseReject(t, tc.Word, resp)
		} else {
			t.Fatalf("Invalid expect value: %q (must be 'ok' or 'reject')", tc.Expect)
		}

		if len(validationErrors) > 0 {
			t.Errorf("Validation failed for word %q (expect=%q):\n%s\nFull response: %s",
				tc.Word, tc.Expect, strings.Join(validationErrors, "\n"), response)
		}
	})
}
