package wordsets

import (
	_ "embed"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

//go:embed spanish_lemma_blocklist.txt
var spanishLemmaBlocklistRaw string

var blockedSpanishLemmas map[string]struct{}

func init() {
	blockedSpanishLemmas = parseLemmaBlocklist(spanishLemmaBlocklistRaw)
}

func parseLemmaBlocklist(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[NormalizeLemmaImport(line)] = struct{}{}
	}
	return out
}

// NormalizeLemmaImport lowercases and applies NFC (matches CSV clean script).
func NormalizeLemmaImport(s string) string {
	return strings.ToLower(norm.NFC.String(strings.TrimSpace(s)))
}

// IsLemmaBlocked reports whether lemma is listed in spanish_lemma_blocklist.txt (after NormalizeLemmaImport).
func IsLemmaBlocked(lemma string) bool {
	_, ok := blockedSpanishLemmas[lemma]
	return ok
}

// IsVowellessAbbrevASCII is true for 2–5 letter a–z tokens with no vowels (y counts as vowel). Catches cm, mm, psc, etc.
func IsVowellessAbbrevASCII(lemma string) bool {
	n := utf8.RuneCountInString(lemma)
	if n < 2 || n > 5 {
		return false
	}
	for _, c := range lemma {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	for _, c := range lemma {
		switch c {
		case 'a', 'e', 'i', 'o', 'u', 'y':
			return false
		}
	}
	return true
}
