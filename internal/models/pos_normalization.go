package models

import "strings"

// CanonicalWordPOS maps noisy/multilingual POS labels to canonical tokens.
func CanonicalWordPOS(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(s, "noun"), strings.HasPrefix(s, "sustantivo"):
		return "noun"
	case strings.HasPrefix(s, "verb"), strings.HasPrefix(s, "verbo"):
		return "verb"
	case strings.HasPrefix(s, "adjective"), strings.HasPrefix(s, "adjetivo"):
		return "adjective"
	case strings.HasPrefix(s, "adverb"), strings.HasPrefix(s, "adverbio"):
		return "adverb"
	case strings.HasPrefix(s, "pronoun"), strings.HasPrefix(s, "pronombre"):
		return "pronoun"
	case strings.HasPrefix(s, "preposition"), strings.HasPrefix(s, "preposición"), strings.HasPrefix(s, "preposicion"):
		return "preposition"
	case strings.HasPrefix(s, "conjunction"), strings.HasPrefix(s, "conjunción"), strings.HasPrefix(s, "conjuncion"):
		return "conjunction"
	case strings.HasPrefix(s, "interjection"), strings.HasPrefix(s, "interjección"), strings.HasPrefix(s, "interjeccion"):
		return "interjection"
	case strings.HasPrefix(s, "article"), strings.HasPrefix(s, "artículo"), strings.HasPrefix(s, "articulo"):
		return "article"
	default:
		return s
	}
}

func IsNounPOS(raw string) bool {
	return CanonicalWordPOS(raw) == "noun"
}

func IsVerbPOS(raw string) bool {
	return CanonicalWordPOS(raw) == "verb"
}

// InferNounGenderFromPOSText extracts noun gender hints from noisy POS text.
func InferNounGenderFromPOSText(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	switch {
	case strings.Contains(s, "femenin"), strings.Contains(s, "feminine"):
		return "f"
	case strings.Contains(s, "masculin"), strings.Contains(s, "masculine"):
		return "m"
	case strings.Contains(s, "neutr"), strings.Contains(s, "neutro"):
		return "n"
	case strings.Contains(s, "común"), strings.Contains(s, "comun"), strings.Contains(s, "common gender"), strings.Contains(s, "m/f"):
		return "mf"
	default:
		return ""
	}
}

func NormalizeNounGenderValue(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "m", "f", "mf", "n":
		return s
	default:
		return ""
	}
}

// IsLikelySimpleSpanishGenderPair validates only safe o<->a noun pairs.
// Irregular pairs (actor/actriz) are intentionally excluded to avoid hallucinations.
func IsLikelySimpleSpanishGenderPair(lemma, opposite string) bool {
	l := strings.ToLower(strings.TrimSpace(lemma))
	o := strings.ToLower(strings.TrimSpace(opposite))
	if l == "" || o == "" || l == o {
		return false
	}
	if strings.HasSuffix(l, "o") && o == l[:len(l)-1]+"a" {
		return true
	}
	if strings.HasSuffix(l, "a") && o == l[:len(l)-1]+"o" {
		return true
	}
	return false
}
