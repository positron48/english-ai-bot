package service

import (
	"fmt"
	"strings"
)

// SpanishVerbRecallContrastQuestion is the on-card prompt for recall/contrast verb-form training:
// subject pronoun + gap + infinitive in parentheses (e.g. "Tú ... (ser)").
// contrast uses different wording so recall and contrast cards for the same slot are never identical.
func SpanishVerbRecallContrastQuestion(lemma, person, number string, stableSeed int64, contrast bool) string {
	lemma = strings.TrimSpace(strings.ToLower(lemma))
	pron := spanishSubjectPronoun(person, number, stableSeed)
	if pron == "" {
		s := fmt.Sprintf("%s (%s/%s)", lemma, strings.TrimSpace(person), strings.TrimSpace(number))
		if contrast {
			return "Elige la forma verbal: " + s
		}
		return s
	}
	if contrast {
		return fmt.Sprintf("Elige la forma: %s … (%s)", pron, lemma)
	}
	return fmt.Sprintf("%s ... (%s)", pron, lemma)
}

func spanishSubjectPronoun(person, number string, stableSeed int64) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	variants := map[string][]string{
		"1|singular": {"Yo"},
		"2|singular": {"Tú"},
		"3|singular": {"Él", "Ella", "Usted"},
		"1|plural":   {"Nosotros", "Nosotras"},
		"2|plural":   {"Vosotros", "Vosotras"},
		"3|plural":   {"Ellos", "Ellas", "Ustedes"},
	}
	list := variants[p+"|"+n]
	if len(list) == 0 {
		return ""
	}
	u := uint64(stableSeed)
	u ^= u >> 33
	u *= 0xff51afd7ed558ccd
	u ^= u >> 33
	u *= 0xc4ceb9fe1a85ec53
	u ^= u >> 33
	return list[int(u%uint64(len(list)))]
}
