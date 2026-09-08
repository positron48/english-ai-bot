package verbtraining

import "strings"

// CoreLemmas supplies examples even before the learner adds vocabulary.
var CoreLemmas = []string{"ser", "estar", "ir", "tener", "hacer", "poder", "decir", "haber", "dar", "ver", "venir", "poner", "querer", "saber", "salir", "hablar", "trabajar", "comer", "beber", "vivir", "abrir", "tomar", "aprender", "recibir"}

var tenseAliases = map[string]string{
	"preterito_indefinido": "pretérito", "preterito": "pretérito",
	"preterito_imperfecto": "imperfecto", "futuro_simple": "futuro",
	"condicional_simple": "condicional", "preterito_perfecto_compuesto": "pretérito perfecto",
	"preterito_perfecto": "pretérito perfecto", "preterito_pluscuamperfecto": "pluscuamperfecto",
	"preterito_anterior": "pretérito anterior", "futuro_perfecto": "futuro perfecto", "condicional_perfecto": "condicional perfecto",
}

func CanonicalTense(tense string) string {
	tense = strings.ToLower(strings.TrimSpace(tense))
	if canonical, ok := tenseAliases[tense]; ok {
		return canonical
	}
	return tense
}

func CanonicalScope(scope string) string {
	tense, mood, ok := ParseScope(scope)
	if !ok {
		return ""
	}
	return "es." + CanonicalTense(tense) + "." + mood
}

// ExpandScopes accepts both imported artifact and Jehle dictionary spelling.
func ExpandScopes(scopes []string) []string {
	seen := map[string]bool{}
	out := []string{}
	add := func(s string) {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	for _, scope := range scopes {
		canonical := CanonicalScope(scope)
		add(canonical)
		tense, mood, ok := ParseScope(canonical)
		if !ok {
			continue
		}
		for alias, value := range tenseAliases {
			if value == tense {
				add("es." + alias + "." + mood)
			}
		}
	}
	return out
}
