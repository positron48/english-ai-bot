package learning

import "strings"

func normalizeTargetLangCode(code string) string {
	c := strings.ToLower(strings.TrimSpace(code))
	if c == "" {
		return "en"
	}
	return c
}

// TargetLangNameRUAccusative is used in phrases like «Переведите на …» (short adjective form).
func TargetLangNameRUAccusative(code string) string {
	switch normalizeTargetLangCode(code) {
	case "en":
		return "английский"
	case "es":
		return "испанский"
	case "de":
		return "немецкий"
	case "fr":
		return "французский"
	case "it":
		return "итальянский"
	case "pt":
		return "португальский"
	default:
		return normalizeTargetLangCode(code)
	}
}

// TargetLangNameRUPrepositional is used after «на» in phrases like «слово на …», «на …» (изучаемом языке).
func TargetLangNameRUPrepositional(code string) string {
	switch normalizeTargetLangCode(code) {
	case "en":
		return "английском"
	case "es":
		return "испанском"
	case "de":
		return "немецком"
	case "fr":
		return "французском"
	case "it":
		return "итальянском"
	case "pt":
		return "португальском"
	default:
		return normalizeTargetLangCode(code)
	}
}

// TargetLangNameEN returns an English name for ISO-like language codes (for non-UI diagnostics or bilingual payloads).
func TargetLangNameEN(code string) string {
	switch normalizeTargetLangCode(code) {
	case "en":
		return "English"
	case "es":
		return "Spanish"
	case "de":
		return "German"
	case "fr":
		return "French"
	case "it":
		return "Italian"
	case "pt":
		return "Portuguese"
	default:
		return strings.ToLower(normalizeTargetLangCode(code))
	}
}

// TargetLangNameES returns a Spanish name for common language codes (for UI text in Spanish).
func TargetLangNameES(code string) string {
	switch normalizeTargetLangCode(code) {
	case "en":
		return "inglés"
	case "es":
		return "español"
	case "de":
		return "alemán"
	case "fr":
		return "francés"
	case "it":
		return "italiano"
	case "pt":
		return "portugués"
	default:
		return strings.ToLower(normalizeTargetLangCode(code))
	}
}
