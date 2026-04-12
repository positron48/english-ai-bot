package spanishverbs

import "testing"

func TestRussianIrVerbForSpanishIndicativo_preterito(t *testing.T) {
	got := RussianIrVerbForSpanishIndicativo("indicativo", "pretérito", "1", "singular")
	if got != "пошёл" {
		t.Fatalf("got %q", got)
	}
}

func TestRussianIrVerbForSpanishIndicativo_imperfecto(t *testing.T) {
	got := RussianIrVerbForSpanishIndicativo("indicativo", "imperfecto", "2", "singular")
	if got != "шёл" {
		t.Fatalf("got %q", got)
	}
}

func TestRussianIrVerbForSpanishIndicativo_futuro(t *testing.T) {
	got := RussianIrVerbForSpanishIndicativo("indicativo", "futuro", "1", "singular")
	if got != "буду идти" {
		t.Fatalf("got %q", got)
	}
}

func TestRussianIrVerbForSpanishIndicativo_preteritoASCII(t *testing.T) {
	got := RussianIrVerbForSpanishIndicativo("indicativo", "preterito", "3", "plural")
	if got != "пошли" {
		t.Fatalf("got %q", got)
	}
}

func TestIrMotionLexicon_coversTemplates(t *testing.T) {
	for _, row := range irCatalogTemplates {
		if _, ok := IrMotionSpanishSuffixToRuTail[row.EsSuffix]; !ok {
			t.Fatalf("missing lexicon for es suffix %q", row.EsSuffix)
		}
	}
}
