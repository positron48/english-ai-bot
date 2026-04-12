package spanishverbs

import (
	"strings"
	"testing"
)

func TestGenerateVerbExamplePair_indicativoPresente_containsSurface(t *testing.T) {
	const lemma = "hablar"
	es, ru := GenerateVerbExamplePair(101, lemma, "indicativo", "presente", "2", "singular", "hablas", "", "", nil)
	if !ExampleContainsSurface(es, "hablas") {
		t.Fatalf("es missing surface: %q", es)
	}
	if !strings.Contains(es, "hablas") {
		t.Fatalf("es: %q", es)
	}
	if ru == "" || !strings.Contains(ru, "«") || !strings.Contains(ru, "»") {
		t.Fatalf("ru: %q", ru)
	}
}

func TestGenerateVerbExamplePair_imperativoAfirmativo2s(t *testing.T) {
	es, ru := GenerateVerbExamplePair(7, "hablar", "imperativo afirmativo", "presente", "2", "singular", "habla", "", "", nil)
	if !ExampleContainsSurface(es, "habla") {
		t.Fatalf("es: %q", es)
	}
	if ru == "" {
		t.Fatal("empty ru")
	}
}

func TestGenerateVerbExamplePair_subjuntivoPresente(t *testing.T) {
	es, _ := GenerateVerbExamplePair(3, "venir", "subjuntivo", "presente", "3", "singular", "venga", "", "", nil)
	if !strings.Contains(strings.ToLower(es), "venga") {
		t.Fatalf("es: %q", es)
	}
}

func TestGenerateVerbExamplePair_compoundForm(t *testing.T) {
	es, _ := GenerateVerbExamplePair(99, "haber", "indicativo", "pretérito perfecto", "1", "singular", "he habido", "", "", nil)
	if !ExampleContainsSurface(es, "he habido") {
		t.Fatalf("es should contain compound form: %q", es)
	}
}

func TestGenerateVerbExamplePair_haber_presente_bare_es_and_compact_ru(t *testing.T) {
	es, ru := GenerateVerbExamplePair(5, "haber", "indicativo", "presente", "2", "singular", "has", "иметь", "", nil)
	if strings.ToLower(es) != "tú has." {
		t.Fatalf("es: want short grammatical frame, got %q", es)
	}
	if !strings.Contains(ru, "«имеешь»") && !strings.Contains(ru, "имеешь") {
		t.Fatalf("ru: want conjugated иметь, got %q", ru)
	}
}

func TestGenerateVerbExamplePair_variantsRotate(t *testing.T) {
	seen := map[string]struct{}{}
	for _, seed := range []int64{1, 2, 3, 4, 5, 9, 17, 33, 65, 129} {
		es, _ := GenerateVerbExamplePair(seed, "andar", "indicativo", "presente", "1", "singular", "ando", "", "", nil)
		seen[es] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatalf("expected variety, got %d distinct sentences: %v", len(seen), seen)
	}
}

func TestSpanishSubjectInSentence(t *testing.T) {
	if g := SpanishSubjectInSentence("3", "singular"); g != "él" {
		t.Fatalf("got %q", g)
	}
}

func TestGenerateVerbExamplePair_querer_seed202604121337_slots(t *testing.T) {
	const seed = int64(202604121337)
	ru := "хотеть, любить"
	for _, tc := range []struct {
		name, mood, tense, person, number, es string
	}{
		{"presente_2s", "indicativo", "presente", "2", "singular", "quieres"},
		{"imperfecto_1s", "indicativo", "imperfecto", "1", "singular", "quería"},
		{"preterito_3s", "indicativo", "pretérito", "3", "singular", "quiso"},
		{"futuro_1p", "indicativo", "futuro", "1", "plural", "querremos"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			es, ruOut := GenerateVerbExamplePair(seed, "querer", tc.mood, tc.tense, tc.person, tc.number, tc.es, ru, "", nil)
			if !ExampleContainsSurface(es, tc.es) {
				t.Fatalf("es=%q want surface %q", es, tc.es)
			}
			if ruOut == "" {
				t.Fatal("empty RU")
			}
			t.Logf("ES: %s\nRU: %s", es, ruOut)
		})
	}
}
