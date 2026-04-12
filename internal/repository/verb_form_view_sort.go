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

// Indicative: present → preterite → imperfect → future → conditional → compound tenses.
var indicativeTenseStudyOrder = []string{
	"presente",
	"pretérito",
	"preterito",
	"imperfecto",
	"futuro",
	"condicional",
	"pretérito perfecto",
	"pluscuamperfecto",
	"futuro perfecto",
	"pretérito anterior",
	"condicional perfecto",
}

var subjunctiveTenseStudyOrder = []string{
	"presente",
	"imperfecto",
	"futuro",
	"pretérito perfecto",
	"pluscuamperfecto",
	"futuro perfecto",
}

func tenseRankSpanish(mood, tense string) int {
	m := strings.ToLower(strings.TrimSpace(mood))
	t := strings.ToLower(strings.TrimSpace(tense))
	var order []string
	switch m {
	case "indicativo":
		order = indicativeTenseStudyOrder
	case "subjuntivo":
		order = subjunctiveTenseStudyOrder
	default:
		order = []string{"presente"}
	}
	for i, key := range order {
		if tenseKeyMatches(t, key) {
			return i
		}
	}
	// Unknown tense: after all listed slots, stable sub-order by label
	return len(order)
}

func tenseKeyMatches(tense, key string) bool {
	if tense == key {
		return true
	}
	// ASCII-only variant in data
	if key == "pretérito" && tense == "preterito" {
		return true
	}
	if key == "preterito" && tense == "pretérito" {
		return true
	}
	return false
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
