package repository

import (
	"sort"
	"strings"
)

// SortVerbFormViewRowsSpanish orders rows for learner-facing tables: indicativo → subjuntivo → imperativo,
// tenses in a typical study order, then persons yo → tú → él… → nosotros → vosotros → ellos….
// SortVerbFormPreviewRowsSpanish uses the same study order as SortVerbFormViewRowsSpanish.
func SortVerbFormPreviewRowsSpanish(rows []VerbFormPreviewRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if d := moodRankSpanish(a.Mood) - moodRankSpanish(b.Mood); d != 0 {
			return d < 0
		}
		if d := tenseRankSpanish(a.Mood, a.Tense) - tenseRankSpanish(b.Mood, b.Tense); d != 0 {
			return d < 0
		}
		if lt, rt := strings.ToLower(strings.TrimSpace(a.Tense)), strings.ToLower(strings.TrimSpace(b.Tense)); lt != rt {
			return lt < rt
		}
		if d := pronounStudyOrder(a.Person, a.Number) - pronounStudyOrder(b.Person, b.Number); d != 0 {
			return d < 0
		}
		return strings.ToLower(a.SurfaceForm) < strings.ToLower(b.SurfaceForm)
	})
}

func SortVerbFormViewRowsSpanish(rows []VerbFormViewRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if d := moodRankSpanish(a.Mood) - moodRankSpanish(b.Mood); d != 0 {
			return d < 0
		}
		if d := tenseRankSpanish(a.Mood, a.Tense) - tenseRankSpanish(b.Mood, b.Tense); d != 0 {
			return d < 0
		}
		if lt, rt := strings.ToLower(strings.TrimSpace(a.Tense)), strings.ToLower(strings.TrimSpace(b.Tense)); lt != rt {
			return lt < rt
		}
		if d := pronounStudyOrder(a.Person, a.Number) - pronounStudyOrder(b.Person, b.Number); d != 0 {
			return d < 0
		}
		return strings.ToLower(a.SurfaceForm) < strings.ToLower(b.SurfaceForm)
	})
}

// SortAdminVerbTrainingCardDetailsSpanish orders admin verb-pack rows like SortVerbFormViewRowsSpanish (study order).
func SortAdminVerbTrainingCardDetailsSpanish(rows []AdminVerbTrainingCardDetail) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if d := moodRankSpanish(a.Mood) - moodRankSpanish(b.Mood); d != 0 {
			return d < 0
		}
		if d := tenseRankSpanish(a.Mood, a.Tense) - tenseRankSpanish(b.Mood, b.Tense); d != 0 {
			return d < 0
		}
		if lt, rt := strings.ToLower(strings.TrimSpace(a.Tense)), strings.ToLower(strings.TrimSpace(b.Tense)); lt != rt {
			return lt < rt
		}
		if d := pronounStudyOrder(a.Person, a.Number) - pronounStudyOrder(b.Person, b.Number); d != 0 {
			return d < 0
		}
		return strings.ToLower(a.SurfaceForm) < strings.ToLower(b.SurfaceForm)
	})
}

func moodRankSpanish(mood string) int {
	m := strings.ToLower(strings.TrimSpace(mood))
	switch {
	case m == "indicativo":
		return 0
	case m == "subjuntivo":
		return 1
	case strings.Contains(m, "imperativo") && strings.Contains(m, "negativo"):
		return 3
	case strings.Contains(m, "imperativo"):
		return 2
	default:
		return 40
	}
}

// indicativeCanonicalTenseOrder matches verbtraining.ExpectedScopesV1 indicativo slots (lemma pack / local tooling).
var indicativeCanonicalTenseOrder = []string{
	"presente",
	"preterito_imperfecto",
	"preterito_indefinido",
	"futuro_simple",
	"condicional_simple",
	"preterito_perfecto_compuesto",
	"preterito_pluscuamperfecto",
	"preterito_anterior",
	"futuro_perfecto",
	"condicional_perfecto",
}

// subjunctiveCanonicalTenseOrder matches verbtraining.ExpectedScopesV1 subjuntivo slots.
var subjunctiveCanonicalTenseOrder = []string{
	"presente",
	"preterito_imperfecto",
	"futuro_simple",
	"preterito_perfecto_compuesto",
	"preterito_pluscuamperfecto",
	"futuro_perfecto",
}

// canonicalSpanishVerbTense maps DB or legacy tense labels to the keys in *CanonicalTenseOrder slices.
func canonicalSpanishVerbTense(mood, tense string) string {
	m := strings.ToLower(strings.TrimSpace(mood))
	t := strings.ToLower(strings.TrimSpace(tense))
	tUs := strings.ReplaceAll(t, " ", "_")

	switch m {
	case "indicativo":
		switch {
		case tUs == "presente" || t == "presente":
			return "presente"
		case strings.Contains(tUs, "preterito_imperfecto"):
			return "preterito_imperfecto"
		case t == "imperfecto":
			return "preterito_imperfecto"
		case strings.Contains(tUs, "preterito_indefinido"):
			return "preterito_indefinido"
		case t == "preterito" || t == "pretérito":
			return "preterito_indefinido"
		case strings.Contains(tUs, "perfecto_compuesto"):
			return "preterito_perfecto_compuesto"
		case strings.Contains(tUs, "pluscuamperfect"):
			return "preterito_pluscuamperfecto"
		case strings.Contains(tUs, "preterito_anterior"):
			return "preterito_anterior"
		case strings.Contains(tUs, "futuro_perfecto"):
			return "futuro_perfecto"
		case strings.Contains(tUs, "condicional_perfecto"):
			return "condicional_perfecto"
		case strings.Contains(tUs, "futuro_simple"):
			return "futuro_simple"
		case t == "futuro":
			return "futuro_simple"
		case strings.Contains(tUs, "condicional_simple"):
			return "condicional_simple"
		case t == "condicional":
			return "condicional_simple"
		}
	case "subjuntivo":
		switch {
		case t == "presente" || tUs == "presente":
			return "presente"
		case strings.Contains(tUs, "preterito_imperfecto"):
			return "preterito_imperfecto"
		case t == "imperfecto":
			return "preterito_imperfecto"
		case strings.Contains(tUs, "futuro_simple"):
			return "futuro_simple"
		case t == "futuro":
			return "futuro_simple"
		case strings.Contains(tUs, "perfecto_compuesto"):
			return "preterito_perfecto_compuesto"
		case strings.Contains(tUs, "pluscuamperfect"):
			return "preterito_pluscuamperfecto"
		case strings.Contains(tUs, "futuro_perfecto"):
			return "futuro_perfecto"
		}
	}
	return tUs
}

func tenseRankSpanish(mood, tense string) int {
	m := strings.ToLower(strings.TrimSpace(mood))
	canonical := canonicalSpanishVerbTense(mood, tense)
	var order []string
	switch m {
	case "indicativo":
		order = indicativeCanonicalTenseOrder
	case "subjuntivo":
		order = subjunctiveCanonicalTenseOrder
	default:
		return len(indicativeCanonicalTenseOrder)
	}
	for i, k := range order {
		if canonical == k {
			return i
		}
	}
	return len(order)
}

// pronounStudyOrder: yo, tú, él/3sg, nosotros, vosotros, ellos (1–3 × sing/pl).
func pronounStudyOrder(person, number string) int {
	p := strings.TrimSpace(person)
	n := strings.ToLower(strings.TrimSpace(number))
	isPlural := strings.HasPrefix(n, "pl")
	var slot int
	switch p {
	case "1":
		slot = 0
	case "2":
		slot = 1
	case "3":
		slot = 2
	default:
		return 99
	}
	if isPlural {
		slot += 3
	}
	return slot
}
