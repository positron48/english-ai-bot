package spanishverbs

import (
	"golang.org/x/text/unicode/norm"
	"strings"
	"tgbot-skeleton/internal/verbtraining"
)

func AcceptedVerbAnswer(expected, answer, mood, tense string) bool {
	normalize := func(s string) string { return strings.ToLower(norm.NFC.String(strings.TrimSpace(s))) }
	expected, answer = normalize(expected), normalize(answer)
	if expected == "" {
		return false
	}
	if expected == answer {
		return true
	}
	tense = verbtraining.CanonicalTense(tense)
	if mood != "subjuntivo" || (tense != "imperfecto" && tense != "pluscuamperfecto") {
		return false
	}
	pairs := [][2]string{{"ramos", "semos"}, {"rais", "seis"}, {"ras", "ses"}, {"ran", "sen"}, {"ra", "se"}}
	words := strings.Fields(expected)
	for i, word := range words {
		if len([]rune(word)) < 4 {
			continue
		}
		for _, pair := range pairs {
			for direction := 0; direction < 2; direction++ {
				suffix, replacement := pair[direction], pair[1-direction]
				if strings.HasSuffix(word, suffix) {
					alternative := append([]string(nil), words...)
					alternative[i] = strings.TrimSuffix(word, suffix) + replacement
					if strings.Join(alternative, " ") == answer {
						return true
					}
				}
			}
		}
	}
	return false
}
