package service

import (
	"strings"

	"tgbot-skeleton/internal/models"
)

func normalizeTargetVerbDisplay(targetLang, pos, word string) string {
	word = strings.TrimSpace(word)
	if word == "" {
		return word
	}
	if englishTargetUsesToInfinitive(targetLang) || !models.IsVerbPOS(pos) {
		return word
	}
	return strings.TrimSpace(strings.TrimPrefix(word, "to "))
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
