package service

import (
	"fmt"
	"strings"
	"unicode"

	"tgbot-skeleton/internal/models"
)

// ValidateTrainingCardResponse validates the LLM response for training card generation
// Returns an error message if validation fails, empty string if valid
func ValidateTrainingCardResponse(wordCard *models.WordCard, resp *models.TrainingCardResponse) string {
	if len(resp.Senses) == 0 {
		return "" // This is handled separately
	}

	lemma := strings.ToLower(wordCard.Word)
	var errors []string

	for senseIdx, sense := range resp.Senses {
		// Get POS for this sense: use from sense, or fallback to word_card if available
		pos := sense.POS
		if pos == "" && wordCard.POS != nil {
			pos = *wordCard.POS
		}

		// R1: distractors_en не должны содержать кириллицу
		for i, distractor := range sense.DistractorsEN {
			if containsCyrillic(distractor) {
				errors = append(errors, fmt.Sprintf("R1 sense=%d distractor_en[%d]=%q contains Cyrillic", senseIdx, i, truncate(distractor, 50)))
			}
		}

		// R2: distractors_ru не должны содержать латиницу
		for i, distractor := range sense.DistractorsRU {
			if containsLatin(distractor) {
				errors = append(errors, fmt.Sprintf("R2 sense=%d distractor_ru[%d]=%q contains Latin", senseIdx, i, truncate(distractor, 50)))
			}
		}

		// R3: distractors_en, если pos == "verb" - должны начинаться на "to "
		if pos == "verb" {
			for i, distractor := range sense.DistractorsEN {
				trimmed := strings.TrimSpace(distractor)
				if !strings.HasPrefix(trimmed, "to ") {
					errors = append(errors, fmt.Sprintf("R3 sense=%d distractor_en[%d]=%q should start with 'to ' (pos=verb)", senseIdx, i, truncate(distractor, 50)))
				}
			}
		}

		// R4: distractors_en, если pos != "verb" - не должны начинаться на "to "
		if pos != "verb" && pos != "" {
			for i, distractor := range sense.DistractorsEN {
				trimmed := strings.TrimSpace(distractor)
				if strings.HasPrefix(trimmed, "to ") {
					errors = append(errors, fmt.Sprintf("R4 sense=%d distractor_en[%d]=%q should not start with 'to ' (pos=%s)", senseIdx, i, truncate(distractor, 50), pos))
				}
			}
		}

		// R5: distractors_en не должны содержать lemma в своих значениях (substring, case-insensitive)
		lemmaLower := strings.ToLower(lemma)
		for i, distractor := range sense.DistractorsEN {
			distractorLower := strings.ToLower(distractor)
			if strings.Contains(distractorLower, lemmaLower) {
				errors = append(errors, fmt.Sprintf("R5 sense=%d distractor_en[%d]=%q contains lemma %q", senseIdx, i, truncate(distractor, 50), lemma))
			}
		}

		// R6: distractors_ru не должны содержать word_ru в своих значениях (substring, case-insensitive)
		wordRULower := strings.ToLower(sense.WordRU)
		for i, distractor := range sense.DistractorsRU {
			distractorLower := strings.ToLower(distractor)
			if strings.Contains(distractorLower, wordRULower) {
				errors = append(errors, fmt.Sprintf("R6 sense=%d distractor_ru[%d]=%q contains word_ru %q", senseIdx, i, truncate(distractor, 50), sense.WordRU))
			}
		}
	}

	if len(errors) > 0 {
		// Limit error message length to avoid database issues
		errorMsg := "validation_failed: " + strings.Join(errors, "; ")
		if len(errorMsg) > 1000 {
			errorMsg = errorMsg[:1000] + "..."
		}
		return errorMsg
	}

	return ""
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

// containsLatin checks if a string contains any Latin characters
func containsLatin(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Latin, r) {
			return true
		}
	}
	return false
}

// truncate truncates a string to maxLen characters, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
