package repository

import (
	"testing"
)

func TestSortVerbFormPreviewRowsSpanish_order(t *testing.T) {
	rows := []VerbFormPreviewRow{
		{Mood: "subjuntivo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "haya"},
		{Mood: "indicativo", Tense: "futuro", Person: "1", Number: "singular", SurfaceForm: "habré"},
		{Mood: "indicativo", Tense: "presente", Person: "3", Number: "plural", SurfaceForm: "han"},
		{Mood: "indicativo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "he"},
	}
	SortVerbFormPreviewRowsSpanish(rows)
	if rows[0].SurfaceForm != "he" || rows[1].SurfaceForm != "han" {
		t.Fatalf("preview sort order: %#v", rows)
	}
}

func TestSortVerbFormViewRowsSpanish_order(t *testing.T) {
	rows := []VerbFormViewRow{
		{Mood: "subjuntivo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "haya"},
		{Mood: "indicativo", Tense: "futuro", Person: "1", Number: "singular", SurfaceForm: "habré"},
		{Mood: "indicativo", Tense: "presente", Person: "3", Number: "plural", SurfaceForm: "han"},
		{Mood: "indicativo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "he"},
		{Mood: "indicativo", Tense: "presente", Person: "2", Number: "singular", SurfaceForm: "has"},
		{Mood: "imperativo negativo", Tense: "presente", Person: "2", Number: "singular", SurfaceForm: "no hayas"},
		{Mood: "imperativo afirmativo", Tense: "presente", Person: "2", Number: "singular", SurfaceForm: "ha"},
	}
	SortVerbFormViewRowsSpanish(rows)
	// indicativo presente: 1sg, 2sg, 3pl before indicativo futuro
	if rows[0].SurfaceForm != "he" || rows[1].SurfaceForm != "has" || rows[2].SurfaceForm != "han" {
		t.Fatalf("indicativo presente person order wrong: %#v", surfaceSeq(rows[:3]))
	}
	if rows[3].Mood != "indicativo" || rows[3].Tense != "futuro" {
		t.Fatalf("expected futuro after presente block, got mood=%s tense=%s", rows[3].Mood, rows[3].Tense)
	}
	// subjunctive after indicative
	idxSub := indexOfSurface(rows, "haya")
	idxInd := indexOfSurface(rows, "he")
	if idxSub <= idxInd {
		t.Fatalf("subjuntivo should follow indicativo")
	}
	// imperativo afirmativo before negativo
	idxAff := indexOfSurface(rows, "ha")
	idxNeg := indexOfSurface(rows, "no hayas")
	if idxAff <= 0 || idxNeg <= 0 || idxAff >= idxNeg {
		t.Fatalf("imperativo afirmativo should precede negativo: aff=%d neg=%d", idxAff, idxNeg)
	}
}

func surfaceSeq(rows []VerbFormViewRow) []string {
	out := make([]string, len(rows))
	for i := range rows {
		out[i] = rows[i].SurfaceForm
	}
	return out
}

func indexOfSurface(rows []VerbFormViewRow, form string) int {
	for i := range rows {
		if rows[i].SurfaceForm == form {
			return i
		}
	}
	return -1
}
