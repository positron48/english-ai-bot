package spanishverbs

import (
	"strings"
	"tgbot-skeleton/internal/verbtraining"
)

// Rule describes an actual form, not a lemma-wide regular/irregular flag.
type Rule struct {
	ID      string `json:"id"`
	Regular bool   `json:"regular"`
	Pattern string `json:"pattern,omitempty"`
	Ending  string `json:"ending,omitempty"`
	Stem    string `json:"stem,omitempty"`
}

func FormRule(lemma, mood, tense, person, number, surface string) Rule {
	tense = verbtraining.CanonicalTense(tense)
	key := "es." + tense + "." + mood
	suffix := ""
	if len(lemma) >= 2 {
		suffix = lemma[len(lemma)-2:]
	}
	index := map[string]int{"1singular": 0, "2singular": 1, "3singular": 2, "1plural": 3, "2plural": 4, "3plural": 5}[person+number]
	endings := map[string]map[string][]string{
		"presente":   {"ar": {"o", "as", "a", "amos", "áis", "an"}, "er": {"o", "es", "e", "emos", "éis", "en"}, "ir": {"o", "es", "e", "imos", "ís", "en"}},
		"pretérito":  {"ar": {"é", "aste", "ó", "amos", "asteis", "aron"}, "er": {"í", "iste", "ió", "imos", "isteis", "ieron"}, "ir": {"í", "iste", "ió", "imos", "isteis", "ieron"}},
		"imperfecto": {"ar": {"aba", "abas", "aba", "ábamos", "abais", "aban"}, "er": {"ía", "ías", "ía", "íamos", "íais", "ían"}, "ir": {"ía", "ías", "ía", "íamos", "íais", "ían"}},
	}
	var ending, stem string
	if mood == "indicativo" && len(lemma) >= 2 {
		if forms := endings[tense][suffix]; len(forms) == 6 {
			ending = forms[index]
			stem = lemma[:len(lemma)-2]
		}
		if tense == "futuro" {
			ending = []string{"é", "ás", "á", "emos", "éis", "án"}[index]
			stem = lemma
		}
		if tense == "condicional" {
			ending = []string{"ía", "ías", "ía", "íamos", "íais", "ían"}[index]
			stem = lemma
		}
	}
	if mood == "subjuntivo" && len(lemma) >= 2 {
		stem = lemma[:len(lemma)-2]
		switch tense {
		case "presente":
			switch suffix {
			case "ar":
				ending = []string{"e", "es", "e", "emos", "éis", "en"}[index]
			case "er", "ir":
				ending = []string{"a", "as", "a", "amos", "áis", "an"}[index]
			}
		case "imperfecto":
			switch suffix {
			case "ar":
				ending = []string{"ara", "aras", "ara", "áramos", "arais", "aran"}[index]
			case "er", "ir":
				ending = []string{"iera", "ieras", "iera", "iéramos", "ierais", "ieran"}[index]
			}
		case "futuro":
			switch suffix {
			case "ar":
				ending = []string{"are", "ares", "are", "áremos", "areis", "aren"}[index]
			case "er", "ir":
				ending = []string{"iere", "ieres", "iere", "iéremos", "iereis", "ieren"}[index]
			}
		}
	}
	if ending != "" && stem+ending == strings.ToLower(surface) {
		return Rule{ID: key + "." + suffix + "." + person + "." + number, Regular: true, Ending: ending, Stem: stem}
	}
	if ending != "" {
		for _, change := range []struct{ from, to, pattern string }{{"e", "ie", "e_ie"}, {"o", "ue", "o_ue"}, {"e", "i", "e_i"}} {
			if i := strings.LastIndex(stem, change.from); i >= 0 && stem[:i]+change.to+stem[i+len(change.from):]+ending == surface {
				return Rule{ID: key + ".form." + lemma + "." + person + "." + number, Pattern: change.pattern, Ending: ending}
			}
		}
	}
	if ending != "" {
		for _, change := range []struct{ from, to, pattern string }{{"c", "qu", "spelling_c_qu"}, {"g", "gu", "spelling_g_gu"}, {"z", "c", "spelling_z_c"}, {"g", "j", "spelling_g_j"}} {
			if strings.HasSuffix(stem, change.from) && strings.TrimSuffix(stem, change.from)+change.to+ending == surface {
				return Rule{ID: key + "." + suffix + "." + change.pattern + "." + person + "." + number, Regular: true, Pattern: change.pattern, Ending: ending}
			}
		}
	}
	pattern := ""
	if mood == "indicativo" && tense == "presente" {
		switch lemma {
		case "ser":
			pattern = "ser_present"
		case "ir":
			pattern = "ir_present"
		case "estar":
			pattern = "estar_present"
		}
		if person == "1" && number == "singular" && strings.HasSuffix(surface, "go") {
			pattern = "yo_go"
		}
		if person == "1" && number == "singular" && strings.HasSuffix(surface, "zco") {
			pattern = "yo_zco"
		}
	}
	if mood == "indicativo" && tense == "pretérito" && (lemma == "ser" || lemma == "ir") {
		pattern = "ser_ir_preterite"
	}
	switch tense {
	case "pretérito perfecto", "pluscuamperfecto", "pretérito anterior", "futuro perfecto", "condicional perfecto":
		if strings.Contains(surface, " ") {
			pattern = "compound"
			if suffix == "ar" || suffix == "er" || suffix == "ir" {
				participle := lemma[:len(lemma)-2] + "ido"
				group := "ido"
				if suffix == "ar" {
					participle = lemma[:len(lemma)-2] + "ado"
					group = "ado"
				}
				words := strings.Fields(surface)
				if words[len(words)-1] == participle {
					return Rule{ID: key + ".compound." + group + "." + person + "." + number, Regular: true, Pattern: pattern}
				}
			}
		}
	}
	return Rule{ID: key + ".form." + lemma + "." + person + "." + number, Pattern: pattern}
}
