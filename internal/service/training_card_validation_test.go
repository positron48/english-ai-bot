package service

import (
	"strings"
	"testing"

	"tgbot-skeleton/internal/models"
)

func TestValidateTrainingCardResponse_R1_CyrillicInDistractorsEN(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"правильный", "wrong", "bad"}, // First contains Cyrillic
				DistractorsRU: []string{"один", "два", "три"},
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R1 (Cyrillic in distractors_en), got empty")
	}
	if !strings.Contains(errorMsg, "R1") {
		t.Errorf("Expected error message to contain 'R1', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R2_LatinInDistractorsRU(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"one", "two", "three"},
				DistractorsRU: []string{"один", "test", "три"}, // Second contains Latin
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R2 (Latin in distractors_ru), got empty")
	}
	if !strings.Contains(errorMsg, "R2") {
		t.Errorf("Expected error message to contain 'R2', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R3_VerbWithoutToPrefix(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "run",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "run",
		Senses: []models.TrainingCardSense{
			{
				POS:           "verb",
				WordRU:        "бежать",
				DistractorsEN: []string{"walk", "to jump", "to skip"}, // First doesn't start with "to "
				DistractorsRU: []string{"идти", "прыгать", "пропускать"},
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R3 (verb distractors_en without 'to ' prefix), got empty")
	}
	if !strings.Contains(errorMsg, "R3") {
		t.Errorf("Expected error message to contain 'R3', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R4_NonVerbWithToPrefix(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"to check", "exam", "quiz"}, // First starts with "to "
				DistractorsRU: []string{"проверка", "экзамен", "викторина"},
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R4 (non-verb distractors_en with 'to ' prefix), got empty")
	}
	if !strings.Contains(errorMsg, "R4") {
		t.Errorf("Expected error message to contain 'R4', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_DistractorsENContainsLemma(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"testing", "exam", "quiz"}, // First contains "test"
				DistractorsRU: []string{"проверка", "экзамен", "викторина"},
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R5 (distractors_en contains lemma), got empty")
	}
	if !strings.Contains(errorMsg, "R5") {
		t.Errorf("Expected error message to contain 'R5', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R6_DistractorsRUContainsWordRU(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"exam", "quiz", "check"},
				DistractorsRU: []string{"тестирование", "экзамен", "викторина"}, // First contains "тест"
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R6 (distractors_ru contains word_ru), got empty")
	}
	if !strings.Contains(errorMsg, "R6") {
		t.Errorf("Expected error message to contain 'R6', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_HappyPath(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "run",
		POS:  stringPtr("verb"),
	}
	resp := &models.TrainingCardResponse{
		WordEN: "run",
		Senses: []models.TrainingCardSense{
			{
				POS:           "verb",
				WordRU:        "бежать",
				DistractorsEN: []string{"to walk", "to jump", "to skip"},
				DistractorsRU: []string{"идти", "прыгать", "пропускать"},
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for valid response, got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_HappyPath_NonVerb(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
		POS:  stringPtr("noun"),
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"exam", "quiz", "check"},
				DistractorsRU: []string{"экзамен", "викторина", "проверка"},
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for valid response, got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_MultipleErrors(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"правильный", "to check", "testing"}, // R1, R4, R5
				DistractorsRU: []string{"test", "тестирование", "три"},        // R2, R6
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for multiple rule violations, got empty")
	}
	// Should contain multiple rule identifiers
	ruleCount := 0
	if strings.Contains(errorMsg, "R1") {
		ruleCount++
	}
	if strings.Contains(errorMsg, "R2") {
		ruleCount++
	}
	if strings.Contains(errorMsg, "R4") {
		ruleCount++
	}
	if strings.Contains(errorMsg, "R5") {
		ruleCount++
	}
	if strings.Contains(errorMsg, "R6") {
		ruleCount++
	}
	if ruleCount < 3 {
		t.Errorf("Expected at least 3 different rule violations, found %d. Error: %s", ruleCount, errorMsg)
	}
}

func TestValidateTrainingCardResponse_POSFromWordCard(t *testing.T) {
	// Test that POS is taken from wordCard when sense.POS is empty
	wordCard := &models.WordCard{
		Word: "run",
		POS:  stringPtr("verb"),
	}
	resp := &models.TrainingCardResponse{
		WordEN: "run",
		Senses: []models.TrainingCardSense{
			{
				POS:           "", // Empty, should use wordCard.POS
				WordRU:        "бежать",
				DistractorsEN: []string{"walk", "to jump", "to skip"}, // First doesn't start with "to "
				DistractorsRU: []string{"идти", "прыгать", "пропускать"},
			},
		},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R3 (verb without 'to '), got empty")
	}
	if !strings.Contains(errorMsg, "R3") {
		t.Errorf("Expected error message to contain 'R3', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_EmptySenses(t *testing.T) {
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{},
	}

	errorMsg := validateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected empty error message for empty senses (handled separately), got: %s", errorMsg)
	}
}

// Helper function - uses existing contains from word_service_integration_test.go
// For exact substring matching in error messages, we use strings.Contains directly
