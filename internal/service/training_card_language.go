package service

import (
	"strings"

	"tgbot-skeleton/internal/models"
)

// TargetLangFromCourseCode extracts the target language from a course code ("es_ru" -> "es").
func TargetLangFromCourseCode(courseCode string) string {
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(courseCode)), "_", 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0]
	}
	return ""
}

// NativeLangFromCourseCode extracts the native language from a course code ("es_ru" -> "ru").
func NativeLangFromCourseCode(courseCode string) string {
	parts := strings.SplitN(strings.ToLower(strings.TrimSpace(courseCode)), "_", 2)
	if len(parts) == 2 && parts[1] != "" {
		return parts[1]
	}
	return ""
}

// TargetLangForCourse returns the learning target language for a word card: course code wins
// when present, otherwise the instance default (e.g. Linglow admin generating es_ru cards).
func TargetLangForCourse(courseCode, fallbackTargetLang string) string {
	if tl := TargetLangFromCourseCode(courseCode); tl != "" {
		return tl
	}
	return fallbackTargetLang
}

// NativeLangForCourse returns the learner native language for a word card.
func NativeLangForCourse(courseCode, fallbackNativeLang string) string {
	if nl := NativeLangFromCourseCode(courseCode); nl != "" {
		return nl
	}
	return fallbackNativeLang
}

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

// NormalizeTargetVerbDisplay strips legacy English "to …" from non-English verb displays.
func NormalizeTargetVerbDisplay(targetLang, pos, word string) string {
	return normalizeTargetVerbDisplay(targetLang, pos, word)
}

// NormalizeTargetVerbDisplays applies NormalizeTargetVerbDisplay to each distractor.
func NormalizeTargetVerbDisplays(targetLang, pos string, words []string) []string {
	return normalizeTargetVerbDisplays(targetLang, pos, words)
}
