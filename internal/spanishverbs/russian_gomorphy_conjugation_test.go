package spanishverbs

import (
	"strings"
	"testing"
)

func TestRussianInfinitiveFromRuGloss_parentheses(t *testing.T) {
	if got := RussianInfinitiveFromRuGloss("быть (состояние/место)"); got != "быть" {
		t.Fatalf("got %q", got)
	}
}

func TestRussianInfinitiveFromRuGloss_prefersImetWhenListedWithByt(t *testing.T) {
	if got := RussianInfinitiveFromRuGloss("быть (вспомог.), иметь"); got != "иметь" {
		t.Fatalf("got %q", got)
	}
}

func TestRussianInfinitiveFromRuGloss_firstField(t *testing.T) {
	if got := RussianInfinitiveFromRuGloss("иметь (вспомог.), быть"); got != "иметь" {
		t.Fatalf("got %q", got)
	}
}

func TestRussianInfinitiveFromRuGloss_commaSpacing(t *testing.T) {
	for _, g := range []string{"хотеть,любить", "хотеть, любить"} {
		if lem := RussianInfinitiveFromRuGloss(g); lem != "хотеть" {
			t.Fatalf("gloss %q: want lemma хотеть, got %q", g, lem)
		}
		if got := RussianCatalogSecondFromGloss(g, "indicativo", "presente", "2", "singular"); got != "хочешь" {
			t.Fatalf("gloss %q: want хочешь, got %q", g, got)
		}
	}
}

func TestRussianCatalogSecondFromGloss_homoglyphLatinEInHotet(t *testing.T) {
	// Latin 'e' (U+0065) in the first gloss field must still resolve to хотеть → хочешь.
	g := "хотeть, любить"
	if got := RussianInfinitiveFromRuGloss(g); got != "хотеть" {
		t.Fatalf("lemma: got %q", got)
	}
	if got := RussianCatalogSecondFromGloss(g, "indicativo", "presente", "2", "singular"); got != "хочешь" {
		t.Fatalf("conjugated: got %q", got)
	}
}

func TestRussianVerbFormForSpanishIndicativo_xotet_present2sg(t *testing.T) {
	got, ok := RussianVerbFormForSpanishIndicativo("хотеть", "presente", "2", "singular")
	if !ok || got != "хочешь" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestRussianVerbFormForSpanishIndicativo_govorit_present2sg(t *testing.T) {
	got, ok := RussianVerbFormForSpanishIndicativo("говорить", "presente", "2", "singular")
	if !ok || got != "говоришь" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestRussianVerbFormForSpanishIndicativo_govorit_futuro1pl(t *testing.T) {
	got, ok := RussianVerbFormForSpanishIndicativo("говорить", "futuro", "1", "plural")
	if !ok || got != "будем говорить" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestRussianCatalogSecondFromGloss_hablarLine(t *testing.T) {
	got := RussianCatalogSecondFromGloss("говорить", "indicativo", "presente", "3", "plural")
	if got != "говорят" {
		t.Fatalf("got %q", got)
	}
}

func TestFreq100TemplateCodes_count(t *testing.T) {
	c := Freq100TemplateCodes()
	if len(c) != 32 {
		t.Fatalf("want 8 frames × 4 tenses = 32, got %d", len(c))
	}
	if !containsString(c, "fp100_a_veces_asi_futuro") {
		t.Fatalf("missing futuro id, sample: %v", c[:5])
	}
}

func containsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func TestTryGenerateCatalogPair_freq100_futuro_conjugatedRU(t *testing.T) {
	es, ru, ok := TryGenerateCatalogPair(0, "hablar", "indicativo", "futuro", "1", "singular", "hablaré", "говорить", "", nil)
	if !ok {
		t.Fatal("expected freq-100 catalog for futuro")
	}
	if !strings.Contains(ru, "«буду говорить»") && !strings.Contains(ru, "буду говорить") {
		t.Fatalf("expected analytic future in RU, got %q es=%q", ru, es)
	}
}
