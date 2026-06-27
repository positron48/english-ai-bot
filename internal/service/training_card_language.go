package service

import (
	"strings"

	"tgbot-skeleton/internal/models"
)

func normalizeTargetVerbDisplay(targetLang, pos, word string) string {
	word = strings.TrimSpace(word)
	if word == "" || englishTargetUsesToInfinitive(targetLang) {
		return word
	}
	// Strip legacy English "to …" markers from Spanish/non-English verbs (incl. POS "verbo")
	// and from distractors stored with an English-style prefix even when POS is missing.
	if models.IsVerbPOS(pos) || strings.HasPrefix(word, "to ") {
		return strings.TrimSpace(strings.TrimPrefix(word, "to "))
	}
	return word
}

func normalizeTargetVerbDisplays(targetLang, pos string, words []string) []string {
	if len(words) == 0 {
		return words
	}
	out := make([]string, len(words))
	for i, word := range words {
		out[i] = normalizeTargetVerbDisplay(targetLang, pos, word)
	}
	return out
}
