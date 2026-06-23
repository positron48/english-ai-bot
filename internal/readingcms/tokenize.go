package readingcms

import (
	"regexp"
	"strings"
	"unicode"
)

var wordTokenRE = regexp.MustCompile(`\w+|[^\w\s]`)

func Tokenize(text string) []map[string]interface{} {
	parts := wordTokenRE.FindAllString(text, -1)
	tokens := make([]map[string]interface{}, 0, len(parts))
	idx := 0
	for _, part := range parts {
		isWord := true
		for _, r := range part {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				isWord = false
				break
			}
		}
		lemma := part
		if isWord {
			lemma = toLowerRunes(part)
		}
		tokens = append(tokens, map[string]interface{}{
			"surface":   part,
			"lemma":     lemma,
			"clickable": isWord,
			"token_idx": idx,
		})
		idx++
	}
	return tokens
}

func toLowerRunes(s string) string {
	var b strings.Builder
	for _, r := range s {
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
