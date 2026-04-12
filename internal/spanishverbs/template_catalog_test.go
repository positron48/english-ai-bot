package spanishverbs

import (
	"strings"
	"testing"
)

func TestIrTemplateCodes_nonEmpty(t *testing.T) {
	c := IrTemplateCodes()
	if len(c) < 32 {
		t.Fatalf("expected 8 motion frames × 4 tenses, got %d: %v", len(c), c[:min(5, len(c))])
	}
}

func TestTryGenerateCatalogPair_irPresent(t *testing.T) {
	es, ru, ok := TryGenerateCatalogPair(7, "ir", "indicativo", "presente", "2", "singular", "vas", "", "", nil)
	if !ok {
		t.Fatal("expected catalog hit for ir presente")
	}
	if !ExampleContainsSurface(es, "vas") {
		t.Fatalf("es missing surface: %q", es)
	}
	if strings.Contains(ru, "«") {
		t.Fatalf("expected compact RU, got %q", ru)
	}
	if !strings.Contains(strings.ToLower(es), "(ir)") {
		t.Fatalf("expected lemma suffix: %q", es)
	}
}

func TestTryGenerateCatalogPair_verbClassMismatch(t *testing.T) {
	_, _, ok := TryGenerateCatalogPair(7, "ir", "indicativo", "presente", "2", "singular", "vas", "", "speech", nil)
	if ok {
		t.Fatal("motion templates should not match verb_class=speech")
	}
}

func TestTryGenerateCatalogPair_allowedIDsRestrict(t *testing.T) {
	es, _, ok := TryGenerateCatalogPair(100, "ir", "indicativo", "presente", "1", "singular", "voy", "", "", []string{"ir_a_casa_presente"})
	if !ok {
		t.Fatal("expected catalog")
	}
	if !strings.Contains(strings.ToLower(es), "casa") {
		t.Fatalf("expected only ir_a_casa tail, got %q", es)
	}
}

func TestGenerateVerbExamplePair_irCatalogCompactRU(t *testing.T) {
	es, ru := GenerateVerbExamplePair(100, "ir", "indicativo", "presente", "1", "singular", "voy", "идти", "", nil)
	if !ExampleContainsSurface(es, "voy") {
		t.Fatalf("es: %q", es)
	}
	if strings.Contains(ru, "«") {
		t.Fatalf("catalog path should use short RU, got %q", ru)
	}
}

func TestGenerateVerbExamplePair_irSpeechClassFallback(t *testing.T) {
	_, ru := GenerateVerbExamplePair(100, "ir", "indicativo", "presente", "1", "singular", "voy", "", "speech", nil)
	if !strings.Contains(ru, "«") {
		t.Fatalf("expected generic literary RU with quotes, got %q", ru)
	}
}

func TestFreq100VerbLemmas_count(t *testing.T) {
	if len(Freq100VerbLemmas) != 100 {
		t.Fatalf("want 100 lemmas, got %d", len(Freq100VerbLemmas))
	}
}

func TestTryGenerateCatalogPair_freq100_hablar(t *testing.T) {
	es, ru, ok := TryGenerateCatalogPair(3, "hablar", "indicativo", "presente", "2", "singular", "hablas", "говорить", "", nil)
	if !ok {
		t.Fatal("expected freq-100 catalog")
	}
	if !ExampleContainsSurface(es, "hablas") {
		t.Fatalf("es: %q", es)
	}
	if !strings.Contains(ru, "«говоришь»") {
		t.Fatalf("expected conjugated RU in quotes, got %q", ru)
	}
}

func TestIrNeverUsesFp100SpanishTails(t *testing.T) {
	for seed := int64(0); seed < 60; seed++ {
		es, _, ok := TryGenerateCatalogPair(seed, "ir", "indicativo", "presente", "1", "singular", "voy", "", "", nil)
		if !ok {
			t.Fatalf("seed %d: expected ir catalog", seed)
		}
		if strings.Contains(strings.ToLower(es), "contexto neutro") || strings.Contains(strings.ToLower(es), "en el día a día") {
			t.Fatalf("seed %d: ir must not use fp100 tails: %q", seed, es)
		}
	}
}

func TestTryGenerateCatalogPair_irPreterito_russianPerfective(t *testing.T) {
	es, ru, ok := TryGenerateCatalogPair(11, "ir", "indicativo", "pretérito", "1", "singular", "fui", "", "", nil)
	if !ok {
		t.Fatal("expected ir catalog for pretérito")
	}
	if !ExampleContainsSurface(es, "fui") {
		t.Fatalf("es: %q", es)
	}
	if !strings.Contains(ru, "пошёл") {
		t.Fatalf("want perfective past in RU, got %q", ru)
	}
}

func TestTryGenerateCatalogPair_freq100_requiresGloss(t *testing.T) {
	_, _, ok := TryGenerateCatalogPair(1, "hablar", "indicativo", "presente", "2", "singular", "hablas", "", "", nil)
	if ok {
		t.Fatal("fp100 gloss templates should not match without ru_gloss")
	}
}

func TestTryGenerateCatalogPair_freq100_excludes_haber(t *testing.T) {
	_, _, ok := TryGenerateCatalogPair(0, "haber", "indicativo", "presente", "2", "singular", "has", "иметь", "", nil)
	if ok {
		t.Fatal("fp100 lexical tails must not match lemma haber (auxiliary)")
	}
}

func TestTryGenerateCatalogPair_querer_futuro1pl_querremos(t *testing.T) {
	const seed = int64(202604121337)
	ruGloss := "хотеть, любить"
	es, ru, ok := TryGenerateCatalogPair(seed, "querer", "indicativo", "futuro", "1", "plural", "querremos", ruGloss, "", nil)
	if !ok {
		t.Fatal("expected freq-100 catalog for querer futuro 1p")
	}
	if !ExampleContainsSurface(es, "querremos") {
		t.Fatalf("es missing futuro surface querremos: %q", es)
	}
	if strings.Contains(strings.ToLower(es), "queremos") && !strings.Contains(strings.ToLower(es), "querremos") {
		t.Fatalf("es must not substitute present nosotros queremos for futuro: %q", es)
	}
	if strings.Count(strings.ToLower(es), "rr") < 1 {
		t.Fatalf("expected double-r stem in querremos, got: %q", es)
	}
	if !strings.Contains(ru, "будем") || !strings.Contains(ru, "хотеть") {
		t.Fatalf("expected analytic RU future, got %q", ru)
	}
}

func TestTryGenerateCatalogPair_querer_presente2s_xocesh(t *testing.T) {
	const seed = int64(202604121337)
	ruGloss := "хотеть, любить"
	es, ru, ok := TryGenerateCatalogPair(seed, "querer", "indicativo", "presente", "2", "singular", "quieres", ruGloss, "", nil)
	if !ok {
		t.Fatal("expected freq-100 catalog for querer presente 2s")
	}
	if !ExampleContainsSurface(es, "quieres") {
		t.Fatalf("es: %q", es)
	}
	if !strings.Contains(ru, "«хочешь»") && !strings.Contains(ru, "хочешь") {
		t.Fatalf("expected RU present 2sg хочешь, got %q", ru)
	}
	if strings.Contains(ru, "хотешь") {
		t.Fatalf("unexpected wrong RU form хотешь in %q", ru)
	}
}
