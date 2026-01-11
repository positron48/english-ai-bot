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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R4 (non-verb distractors_en with 'to ' prefix), got empty")
	}
	if !strings.Contains(errorMsg, "R4") {
		t.Errorf("Expected error message to contain 'R4', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_ExactMatch(t *testing.T) {
	// "test" для "test" - должно быть ошибкой (точное совпадение)
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"test", "exam", "quiz"}, // First exactly matches "test"
				DistractorsRU: []string{"проверка", "экзамен", "викторина"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R5 (distractors_en exactly matches lemma), got empty")
	}
	if !strings.Contains(errorMsg, "R5") {
		t.Errorf("Expected error message to contain 'R5', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_DiffersByOneCharacter(t *testing.T) {
	// "tests" для "test" - должно быть ошибкой (отличается на 1 символ)
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"tests", "exam", "quiz"}, // First differs by 1 character (added 's')
				DistractorsRU: []string{"проверка", "экзамен", "викторина"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R5 (distractors_en differs from lemma by 1 character), got empty")
	}
	if !strings.Contains(errorMsg, "R5") {
		t.Errorf("Expected error message to contain 'R5', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_DiffersByOneCharacterSubstitution(t *testing.T) {
	// "tast" для "test" - должно быть ошибкой (отличается на 1 символ - замена)
	wordCard := &models.WordCard{
		Word: "test",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "test",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "тест",
				DistractorsEN: []string{"tast", "exam", "quiz"}, // First differs by 1 character (substitution)
				DistractorsRU: []string{"проверка", "экзамен", "викторина"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R5 (distractors_en differs from lemma by 1 character), got empty")
	}
	if !strings.Contains(errorMsg, "R5") {
		t.Errorf("Expected error message to contain 'R5', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_DiffersByMoreThanOneCharacter(t *testing.T) {
	// "woman" для "man" - должно быть ок (отличается более чем на 1 символ)
	wordCard := &models.WordCard{
		Word: "man",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "man",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "мужчина",
				DistractorsEN: []string{"woman", "person", "human"}, // "woman" отличается более чем на 1 символ от "man"
				DistractorsRU: []string{"женщина", "человек", "личность"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for R5 (distractors_en differs from lemma by more than 1 character), got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_DiffersOnlyByFirstChar(t *testing.T) {
	// "billion" для "million" - должно быть валидно (отличается только первый символ)
	wordCard := &models.WordCard{
		Word: "million",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "million",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "миллион",
				DistractorsEN: []string{"billion", "thousand", "hundred"}, // First differs only by first character (m -> b)
				DistractorsRU: []string{"миллиард", "тысяча", "сотня"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for R5 (distractors_en differs from lemma only by first character), got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_VerbDistractorExactMatch(t *testing.T) {
	// "to be" для "be" - должно быть ошибкой (точное совпадение после удаления "to ")
	wordCard := &models.WordCard{
		Word: "be",
		POS:  stringPtr("verb"),
	}
	resp := &models.TrainingCardResponse{
		WordEN: "be",
		Senses: []models.TrainingCardSense{
			{
				POS:           "verb",
				WordRU:        "быть",
				DistractorsEN: []string{"to be", "to have", "to do"}, // "to be" после "to " точно совпадает с "be"
				DistractorsRU: []string{"быть", "иметь", "делать"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R5 (verb distractor exactly matches lemma after 'to '), got empty")
	}
	if !strings.Contains(errorMsg, "R5") {
		t.Errorf("Expected error message to contain 'R5', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_VerbDistractorDiffersByOneCharacter(t *testing.T) {
	// "to bes" для "be" - должно быть ошибкой (отличается на 1 символ после удаления "to ")
	wordCard := &models.WordCard{
		Word: "be",
		POS:  stringPtr("verb"),
	}
	resp := &models.TrainingCardResponse{
		WordEN: "be",
		Senses: []models.TrainingCardSense{
			{
				POS:           "verb",
				WordRU:        "быть",
				DistractorsEN: []string{"to bes", "to have", "to do"}, // "to bes" после "to " отличается на 1 символ от "be"
				DistractorsRU: []string{"быть", "иметь", "делать"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R5 (verb distractor differs from lemma by 1 character after 'to '), got empty")
	}
	if !strings.Contains(errorMsg, "R5") {
		t.Errorf("Expected error message to contain 'R5', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_VerbDistractorDiffersByMoreThanOneCharacter(t *testing.T) {
	// "to rebell" для "be" - должно быть ок (отличается более чем на 1 символ после удаления "to ")
	wordCard := &models.WordCard{
		Word: "be",
		POS:  stringPtr("verb"),
	}
	resp := &models.TrainingCardResponse{
		WordEN: "be",
		Senses: []models.TrainingCardSense{
			{
				POS:           "verb",
				WordRU:        "быть",
				DistractorsEN: []string{"to rebell", "to have", "to do"}, // "to rebell" после "to " отличается более чем на 1 символ от "be"
				DistractorsRU: []string{"восставать", "иметь", "делать"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for R5 (verb distractor differs from lemma by more than 1 character after 'to '), got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_ContainsWordButDiffersByMoreThanOneCharacter(t *testing.T) {
	// Пример: "court" для "court" - но если бы было "courtroom" для "court", это ок
	// Или "man" в "woman" - отличается более чем на 1 символ, поэтому ок
	// Этот тест проверяет, что даже если слово содержится внутри, но отличается более чем на 1 символ - это ок
	wordCard := &models.WordCard{
		Word: "court",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "court",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "суд",
				DistractorsEN: []string{"courtroom", "judge", "lawyer"}, // "courtroom" содержит "court", но отличается более чем на 1 символ
				DistractorsRU: []string{"судья", "заседатель", "адвокат"}, // "судья" отличается более чем на 1 символ от "суд"
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for R5 (distractor contains word but differs by more than 1 character), got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R5_ReactionForAction(t *testing.T) {
	// "reaction" для "action" - должно быть ок (отличается более чем на 1 символ)
	wordCard := &models.WordCard{
		Word: "action",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "action",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "действие",
				DistractorsEN: []string{"reaction", "motion", "movement"}, // "reaction" содержит "action", но отличается более чем на 1 символ
				DistractorsRU: []string{"реакция", "движение", "перемещение"},
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for R5 (reaction for action - differs by more than 1 character), got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R6_ExactMatch(t *testing.T) {
	// "тест" для "тест" - должно быть ошибкой (точное совпадение)
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
				DistractorsRU: []string{"тест", "экзамен", "викторина"}, // First exactly matches "тест"
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R6 (distractors_ru exactly matches word_ru), got empty")
	}
	if !strings.Contains(errorMsg, "R6") {
		t.Errorf("Expected error message to contain 'R6', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R6_DiffersByOneCharacter(t *testing.T) {
	// "тес" для "тест" - должно быть ошибкой (отличается на 1 символ - удаление)
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
				DistractorsRU: []string{"тес", "экзамен", "викторина"}, // First differs by 1 character (removed 'т')
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg == "" {
		t.Error("Expected validation error for R6 (distractors_ru differs from word_ru by 1 character), got empty")
	}
	if !strings.Contains(errorMsg, "R6") {
		t.Errorf("Expected error message to contain 'R6', got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R6_DiffersByMoreThanOneCharacter(t *testing.T) {
	// "судья" для "суд" - должно быть ок (отличается более чем на 1 символ)
	wordCard := &models.WordCard{
		Word: "court",
	}
	resp := &models.TrainingCardResponse{
		WordEN: "court",
		Senses: []models.TrainingCardSense{
			{
				POS:           "noun",
				WordRU:        "суд",
				DistractorsEN: []string{"judge", "lawyer", "jury"},
				DistractorsRU: []string{"судья", "адвокат", "присяжный"}, // "судья" отличается более чем на 1 символ от "суд"
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for R6 (distractors_ru differs from word_ru by more than 1 character), got: %s", errorMsg)
	}
}

func TestValidateTrainingCardResponse_R6_ContainsWordButDiffersByMoreThanOneCharacter(t *testing.T) {
	// "тестирование" для "тест" - должно быть ок (содержит "тест", но отличается более чем на 1 символ)
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
				DistractorsRU: []string{"тестирование", "экзамен", "викторина"}, // "тестирование" содержит "тест", но отличается более чем на 1 символ
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected no validation error for R6 (distractor contains word but differs by more than 1 character), got: %s", errorMsg)
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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
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
				DistractorsEN: []string{"правильный", "to check", "tests"}, // R1, R4, R5 (tests differs by 1 char from test)
				DistractorsRU: []string{"test", "тесты", "три"},        // R2, R6 (тесты differs by 1 char from тест)
			},
		},
	}

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
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

	errorMsg := ValidateTrainingCardResponse(wordCard, resp)
	if errorMsg != "" {
		t.Errorf("Expected empty error message for empty senses (handled separately), got: %s", errorMsg)
	}
}

// Helper function - uses existing stringPtr from word_service_integration_test.go
// For exact substring matching in error messages, we use strings.Contains directly
