package service

import (
	"strings"
	"testing"
)

func TestSpanishVerbRecallContrastQuestion_2singular(t *testing.T) {
	got := SpanishVerbRecallContrastQuestion("ser", "2", "singular", 1, false)
	if !strings.HasPrefix(got, "Tú ...") || !strings.Contains(got, "(ser)") {
		t.Fatalf("got %q", got)
	}
}

func TestSpanishVerbRecallContrastQuestion_3singular_rotates(t *testing.T) {
	seen := map[string]bool{}
	for seed := int64(0); seed < 500; seed++ {
		got := SpanishVerbRecallContrastQuestion("hablar", "3", "singular", seed, false)
		for _, p := range []string{"Él", "Ella", "Usted"} {
			if strings.HasPrefix(got, p+" ...") {
				seen[p] = true
			}
		}
	}
	if len(seen) < 2 {
		t.Fatalf("expected variety among él/ella/usted, saw %v", seen)
	}
}

func TestSpanishVerbRecallContrastQuestion_1plural(t *testing.T) {
	got := SpanishVerbRecallContrastQuestion("ir", "1", "plural", 42, false)
	if !strings.Contains(got, "...") || !strings.Contains(got, "(ir)") {
		t.Fatalf("got %q", got)
	}
	if !strings.HasPrefix(got, "Nosotros ...") && !strings.HasPrefix(got, "Nosotras ...") {
		t.Fatalf("got %q", got)
	}
}

func TestSpanishVerbRecallContrastQuestion_unknownPerson(t *testing.T) {
	got := SpanishVerbRecallContrastQuestion("ser", "x", "y", 1, false)
	if !strings.Contains(got, "x") || !strings.Contains(got, "y") {
		t.Fatalf("fallback expected, got %q", got)
	}
}

func TestSpanishVerbRecallContrastQuestion_recallVsContrastDistinct(t *testing.T) {
	seed := int64(7)
	recall := SpanishVerbRecallContrastQuestion("hablar", "1", "singular", seed, false)
	contrast := SpanishVerbRecallContrastQuestion("hablar", "1", "singular", seed, true)
	if recall == contrast {
		t.Fatalf("recall and contrast prompts must differ, both %q", recall)
	}
	if !strings.Contains(recall, "...") {
		t.Fatalf("recall: %q", recall)
	}
	if !strings.Contains(contrast, "Elige la forma") || !strings.Contains(contrast, "…") {
		t.Fatalf("contrast: %q", contrast)
	}
}
