package service

import (
	"fmt"
	"strings"
	"unicode"

	"tgbot-skeleton/internal/models"
)

// englishTargetUsesToInfinitive is true when the learned (target) language is English — verb
// distractors are expected as "to …". For Spanish, French, etc. this rule does not apply.
func englishTargetUsesToInfinitive(targetLang string) bool {
	return strings.EqualFold(strings.TrimSpace(targetLang), "en")
}

// ValidateTrainingCardResponse validates the LLM response for training card generation
// Returns an error message if validation fails, empty string if valid.
// targetLang is the LEARNING_TARGET_LANG code (e.g. en, es); English-specific "to " checks apply only when targetLang is en.
func ValidateTrainingCardResponse(targetLang string, wordCard *models.WordCard, resp *models.TrainingCardResponse) string {
	if len(resp.Senses) == 0 {
		return "" // This is handled separately
	}

	lemma := strings.ToLower(wordCard.Word)
	var errors []string

	// R0: top-level word/display fields must not contain comma-separated forms
	if strings.Contains(resp.WordEN, ",") {
		errors = append(errors, fmt.Sprintf("R0 word_target=%q should not contain comma-separated values", truncate(resp.WordEN, 50)))
	}

	for senseIdx, sense := range resp.Senses {
		if strings.Contains(sense.DisplayWord, ",") {
			errors = append(errors, fmt.Sprintf("R0 sense=%d display_word=%q should not contain comma-separated values", senseIdx, truncate(sense.DisplayWord, 50)))
		}

		// Get POS for this sense: use from sense, or fallback to word_card if available
		pos := sense.POS
		if pos == "" && wordCard.POS != nil {
			pos = *wordCard.POS
		}

		// R1: distractors_en не должны содержать кириллицу
		for i, distractor := range sense.DistractorsEN {
			if ContainsCyrillic(distractor) {
				errors = append(errors, fmt.Sprintf("R1 sense=%d distractor_en[%d]=%q contains Cyrillic", senseIdx, i, truncate(distractor, 50)))
			}
		}

		// R2: distractors_ru не должны содержать латиницу
		for i, distractor := range sense.DistractorsRU {
			if containsLatin(distractor) {
				errors = append(errors, fmt.Sprintf("R2 sense=%d distractor_ru[%d]=%q contains Latin", senseIdx, i, truncate(distractor, 50)))
			}
		}

		enInfinitive := englishTargetUsesToInfinitive(targetLang)

		// R3: distractors_en, если pos == "verb" — для английского должны начинаться на "to "
		if enInfinitive && pos == "verb" {
			for i, distractor := range sense.DistractorsEN {
				trimmed := strings.TrimSpace(distractor)
				if !strings.HasPrefix(trimmed, "to ") {
					errors = append(errors, fmt.Sprintf("R3 sense=%d distractor_en[%d]=%q should start with 'to ' (pos=verb)", senseIdx, i, truncate(distractor, 50)))
				}
			}
		}

		// R4: distractors_en, если pos != "verb" — для английского не должны начинаться на "to "
		if enInfinitive && pos != "verb" && pos != "" {
			for i, distractor := range sense.DistractorsEN {
				trimmed := strings.TrimSpace(distractor)
				if strings.HasPrefix(trimmed, "to ") {
					errors = append(errors, fmt.Sprintf("R4 sense=%d distractor_en[%d]=%q should not start with 'to ' (pos=%s)", senseIdx, i, truncate(distractor, 50), pos))
				}
			}
		}

		// R5: distractors_en не должны в точности совпадать с lemma или отличаться от него на 1 символ
		// Исключение: если отличается только первый символ, это валидно (например, million/billion)
		// Для глаголов отбрасываем "to " перед проверкой
		lemmaLower := strings.ToLower(lemma)
		for i, distractor := range sense.DistractorsEN {
			distractorLower := strings.ToLower(distractor)
			// Для английских глаголов удаляем "to " из начала дескриптора перед проверкой
			if enInfinitive && pos == "verb" {
				distractorLower = strings.TrimPrefix(distractorLower, "to ")
				distractorLower = strings.TrimSpace(distractorLower)
			}
			// Проверяем точное совпадение
			if distractorLower == lemmaLower {
				errors = append(errors, fmt.Sprintf("R5 sense=%d distractor_en[%d]=%q exactly matches lemma %q", senseIdx, i, truncate(distractor, 50), lemma))
				continue
			}
			// Проверяем, отличается ли дескриптор от леммы на 1 символ (расстояние Левенштейна = 1)
			if levenshteinDistance(distractorLower, lemmaLower) == 1 {
				// Проверяем, отличается ли только первый символ
				if differsOnlyByFirstChar(distractorLower, lemmaLower) {
					// Это валидно - отличается только первый символ (например, million/billion)
					continue
				}
				// Отличается не первый символ - это невалидно
				errors = append(errors, fmt.Sprintf("R5 sense=%d distractor_en[%d]=%q differs from lemma %q by 1 character", senseIdx, i, truncate(distractor, 50), lemma))
			}
		}

		// R6: distractors_ru не должны в точности совпадать с word_ru
		wordRULower := strings.ToLower(sense.WordRU)
		for i, distractor := range sense.DistractorsRU {
			distractorLower := strings.ToLower(distractor)
			// Проверяем точное совпадение
			if distractorLower == wordRULower {
				errors = append(errors, fmt.Sprintf("R6 sense=%d distractor_ru[%d]=%q exactly matches word_ru %q", senseIdx, i, truncate(distractor, 50), sense.WordRU))
			}
		}

		// R7: meaning_target should not trivially mirror lemma.
		// Generic anti-hallucination check (no hardcoded words):
		// reject if meaning_target is the same as lemma or differs by one character.
		meaningTargetLower := strings.ToLower(strings.TrimSpace(sense.MeaningEN))
		lemmaNormalized := strings.ToLower(strings.TrimSpace(lemma))
		if meaningTargetLower != "" && lemmaNormalized != "" {
			if meaningTargetLower == lemmaNormalized {
				errors = append(errors, fmt.Sprintf("R7 sense=%d meaning_target=%q exactly matches lemma %q", senseIdx, truncate(sense.MeaningEN, 50), lemma))
			} else if levenshteinDistance(meaningTargetLower, lemmaNormalized) == 1 {
				errors = append(errors, fmt.Sprintf("R7 sense=%d meaning_target=%q differs from lemma %q by 1 character", senseIdx, truncate(sense.MeaningEN, 50), lemma))
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

// ContainsCyrillic checks if a string contains any Cyrillic characters
func ContainsCyrillic(s string) bool {
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

// levenshteinDistance calculates the Levenshtein distance between two strings
// Returns the minimum number of single-character edits (insertions, deletions, or substitutions)
// required to change one string into the other
// Works with runes (Unicode characters), not bytes, to properly handle multi-byte characters
func levenshteinDistance(s1, s2 string) int {
	// Convert strings to rune slices for proper Unicode handling
	r1 := []rune(s1)
	r2 := []rune(s2)

	if len(r1) == 0 {
		return len(r2)
	}
	if len(r2) == 0 {
		return len(r1)
	}

	// Create a 2D slice to store distances
	dp := make([][]int, len(r1)+1)
	for i := range dp {
		dp[i] = make([]int, len(r2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(r1); i++ {
		dp[i][0] = i
	}
	for j := 0; j <= len(r2); j++ {
		dp[0][j] = j
	}

	// Fill the dp table
	for i := 1; i <= len(r1); i++ {
		for j := 1; j <= len(r2); j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}
			dp[i][j] = min3(
				dp[i-1][j]+1,      // deletion
				dp[i][j-1]+1,      // insertion
				dp[i-1][j-1]+cost, // substitution
			)
		}
	}

	return dp[len(r1)][len(r2)]
}

// min3 returns the minimum of three integers
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// differsOnlyByFirstChar checks if two strings differ only by the first character
// Returns true if strings have the same length and all characters except the first are identical
func differsOnlyByFirstChar(s1, s2 string) bool {
	r1 := []rune(s1)
	r2 := []rune(s2)

	// Must have same length
	if len(r1) != len(r2) {
		return false
	}

	// Must have at least 2 characters (otherwise it's just one character difference)
	if len(r1) < 2 {
		return false
	}

	// First characters must be different
	if r1[0] == r2[0] {
		return false
	}

	// All other characters must be identical
	for i := 1; i < len(r1); i++ {
		if r1[i] != r2[i] {
			return false
		}
	}

	return true
}
