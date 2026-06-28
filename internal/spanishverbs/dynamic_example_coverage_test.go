package spanishverbs

import (
	"strings"
	"testing"
)

func TestPickVariant(t *testing.T) {
	if got := pickVariant(42, 0); got != 0 {
		t.Fatalf("n<=0: got %d", got)
	}
	if got := pickVariant(-1<<62, 5); got < 0 || got >= 5 {
		t.Fatalf("negative seed: got %d", got)
	}
	seen := map[int]struct{}{}
	for seed := int64(0); seed < 20; seed++ {
		seen[pickVariant(seed, 6)] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatalf("expected variant spread, got %v", seen)
	}
}

func TestSpanishSubjectCapital(t *testing.T) {
	tests := []struct {
		person, number, want string
	}{
		{"1", "singular", "Yo"},
		{"1", "plural", "Nosotros"},
		{"2", "singular", "Tú"},
		{"2", "plural", "Vosotros"},
		{"3", "singular", "Él"},
		{"3", "plural", "Ellos"},
		{"9", "singular", "Yo"},
		{" 1 ", " PLURAL ", "Nosotros"},
	}
	for _, tc := range tests {
		if got := SpanishSubjectCapital(tc.person, tc.number); got != tc.want {
			t.Fatalf("SpanishSubjectCapital(%q,%q)=%q want %q", tc.person, tc.number, got, tc.want)
		}
	}
}

func TestRussianSubjectLower(t *testing.T) {
	tests := []struct {
		person, number, want string
	}{
		{"1", "singular", "я"},
		{"1", "plural", "мы"},
		{"2", "singular", "ты"},
		{"2", "plural", "вы"},
		{"3", "singular", "он"},
		{"3", "plural", "они"},
		{"x", "singular", "я"},
	}
	for _, tc := range tests {
		if got := RussianSubjectLower(tc.person, tc.number); got != tc.want {
			t.Fatalf("RussianSubjectLower(%q,%q)=%q want %q", tc.person, tc.number, got, tc.want)
		}
	}
}

func TestSpanishSubjectInSentence_allPersons(t *testing.T) {
	tests := []struct {
		person, number, want string
	}{
		{"1", "singular", "yo"},
		{"2", "plural", "vosotros"},
		{"3", "singular", "él"},
	}
	for _, tc := range tests {
		if got := SpanishSubjectInSentence(tc.person, tc.number); got != tc.want {
			t.Fatalf("SpanishSubjectInSentence(%q,%q)=%q want %q", tc.person, tc.number, got, tc.want)
		}
	}
}

func TestStartsWithReflexiveClitic(t *testing.T) {
	tests := []struct {
		surface string
		want    bool
	}{
		{"me lavo", true},
		{"Me lavo", true},
		{"te callas", true},
		{"se va", true},
		{"nos vamos", true},
		{"os quedáis", true},
		{"hablo", false},
		{"", false},
		{"me", false},
	}
	for _, tc := range tests {
		if got := startsWithReflexiveClitic(tc.surface); got != tc.want {
			t.Fatalf("startsWithReflexiveClitic(%q)=%v want %v", tc.surface, got, tc.want)
		}
	}
}

func TestFiniteTemplates_allMoodTenseKeys(t *testing.T) {
	keys := []struct{ mood, tense string }{
		{"indicativo", "presente"},
		{"indicativo", "futuro"},
		{"indicativo", "imperfecto"},
		{"indicativo", "pretérito"},
		{"indicativo", "preterito"},
		{"indicativo", "condicional"},
		{"indicativo", "pretérito perfecto"},
		{"indicativo", "preterito perfecto"},
		{"indicativo", "pluscuamperfecto"},
		{"indicativo", "futuro perfecto"},
		{"indicativo", "pretérito anterior"},
		{"indicativo", "preterito anterior"},
		{"indicativo", "condicional perfecto"},
		{"subjuntivo", "presente"},
		{"subjuntivo", "imperfecto"},
		{"subjuntivo", "futuro"},
		{"subjuntivo", "pretérito perfecto"},
		{"subjuntivo", "preterito perfecto"},
		{"subjuntivo", "pluscuamperfecto"},
		{"subjuntivo", "futuro perfecto"},
		{"unknown", "tense"},
	}
	for _, k := range keys {
		t.Run(k.mood+"|"+k.tense, func(t *testing.T) {
			pairs := finiteTemplates(k.mood, k.tense)
			if len(pairs) == 0 {
				t.Fatal("expected at least one template")
			}
			for _, p := range pairs {
				out := formatFiniteES(p, "yo", "hablo")
				if !strings.Contains(out, "hablo") {
					t.Fatalf("formatted %q missing surface", out)
				}
			}
		})
	}
}

func TestImperativeAffirmativeTemplates(t *testing.T) {
	cases := []struct {
		person, number string
		minLen         int
	}{
		{"2", "singular", 2},
		{"3", "singular", 1},
		{"1", "plural", 1},
		{"2", "plural", 1},
		{"3", "plural", 1},
		{"1", "singular", 1},
	}
	for _, tc := range cases {
		t.Run(tc.person+"_"+tc.number, func(t *testing.T) {
			tpls := imperativeAffirmativeTemplates(tc.person, tc.number)
			if len(tpls) < tc.minLen {
				t.Fatalf("got %d templates", len(tpls))
			}
			es, _ := GenerateVerbExamplePair(1, "hablar", "imperativo afirmativo", "presente", tc.person, tc.number, "habla", "", "", nil)
			if !ExampleContainsSurface(es, "habla") {
				t.Fatalf("es=%q", es)
			}
		})
	}
}

func TestImperativeNegativeTemplates(t *testing.T) {
	cases := []struct {
		person, number, surface string
	}{
		{"2", "singular", "no hables"},
		{"3", "singular", "no hable"},
		{"1", "plural", "no hablemos"},
		{"2", "plural", "no habléis"},
		{"3", "plural", "no hablen"},
		{"1", "singular", "no hables"},
	}
	for _, tc := range cases {
		t.Run(tc.person+"_"+tc.number, func(t *testing.T) {
			pairs := imperativeNegativeTemplates(tc.person, tc.number)
			if len(pairs) == 0 {
				t.Fatal("expected templates")
			}
			es, ru := GenerateVerbExamplePair(3, "hablar", "imperativo negativo", "presente", tc.person, tc.number, tc.surface, "говорить", "", nil)
			if !ExampleContainsSurface(es, tc.surface) {
				t.Fatalf("es=%q want surface %q", es, tc.surface)
			}
			if ru == "" {
				t.Fatal("empty ru")
			}
		})
	}
}

func TestRussianHintForMoodTense(t *testing.T) {
	hints := map[string]string{
		"indicativo|presente":              "Настоящее изъявительное",
		"indicativo|futuro":                "Будущее изъявительное",
		"indicativo|imperfecto":            "Прошедшее несовершенное",
		"indicativo|pretérito":             "Претерито",
		"indicativo|condicional":           "Условное наклонение",
		"indicativo|pretérito perfecto":    "Претерито перфекто",
		"indicativo|pluscuamperfecto":      "Плюсквамперфекто",
		"indicativo|futuro perfecto":       "Будущее перфекто",
		"indicativo|pretérito anterior":    "Претерито антериор",
		"indicativo|condicional perfecto":  "Условное перфекто",
		"subjuntivo|presente":              "Сослагательное настоящее",
		"subjuntivo|imperfecto":            "Сослагательное прошедшее",
		"subjuntivo|futuro":                "Сослагательное будущее",
		"subjuntivo|pretérito perfecto":    "Сослагательное перфекто",
		"subjuntivo|pluscuamperfecto":      "Сослагательное плюсквамперфекто",
		"subjuntivo|futuro perfecto":       "Сослагательное будущее перфекто",
		"imperativo afirmativo|presente":   "Повелительное",
		"imperativo negativo|presente":     "Отрицательное повелительное",
		"other|other":                      "учебной рамке",
	}
	for key, fragment := range hints {
		parts := strings.SplitN(key, "|", 2)
		got := russianHintForMoodTense(parts[0], parts[1])
		if !strings.Contains(got, fragment) {
			t.Fatalf("russianHintForMoodTense(%q)=%q want fragment %q", key, got, fragment)
		}
	}
}

func TestBuildRussianExampleLine(t *testing.T) {
	line := BuildRussianExampleLine("yo hablo", "hablar", "hablo", "indicativo", "presente", 7)
	if !strings.Contains(line, "«yo hablo»") || !strings.Contains(line, "hablar") {
		t.Fatalf("line=%q", line)
	}
}

func TestExampleContainsSurface(t *testing.T) {
	if ExampleContainsSurface("", "x") || ExampleContainsSurface("x", "") {
		t.Fatal("empty inputs should be false")
	}
	if !ExampleContainsSurface("Yo Hablo", "hablo") {
		t.Fatal("case insensitive match expected")
	}
}

func TestGenerateVerbExamplePair_reflexiveClitic(t *testing.T) {
	es, ru := GenerateVerbExamplePair(11, "lavar", "indicativo", "presente", "1", "singular", "me lavo", "", "", nil)
	if !ExampleContainsSurface(es, "me lavo") {
		t.Fatalf("es=%q", es)
	}
	if ru == "" {
		t.Fatal("empty ru")
	}
}

func TestGenerateVerbExamplePair_catalogPath(t *testing.T) {
	es, ru := GenerateVerbExamplePair(7, "ir", "indicativo", "presente", "2", "singular", "vas", "идти", "", nil)
	if !ExampleContainsSurface(es, "vas") {
		t.Fatalf("catalog es=%q", es)
	}
	if ru == "" {
		t.Fatal("empty ru")
	}
}

func TestGenerateVerbExamplePair_catalogSkippedWhenSurfaceMismatch(t *testing.T) {
	// Catalog matches "ir" but sentence may not contain a deliberately wrong surface token.
	es, ru := GenerateVerbExamplePair(7, "ir", "indicativo", "presente", "2", "singular", "zzzsurfacezzz", "", "", nil)
	if !ExampleContainsSurface(es, "zzzsurfacezzz") {
		t.Fatalf("fallback es=%q", es)
	}
	if ru == "" {
		t.Fatal("empty ru")
	}
}

func TestGenerateVerbExamplePair_reflexiveFallbackBareSurface(t *testing.T) {
	// Force fallback path when composed sentence still misses surface (edge surface token).
	es, _ := GenerateVerbExamplePair(0, "x", "otro", "tiempo", "1", "singular", "me x", "", "", nil)
	if !strings.HasPrefix(es, "me x") {
		t.Fatalf("es=%q", es)
	}
}

func TestPickVariant_negativeSeed(t *testing.T) {
	seed := int64(-9223372036854775808)
	got := pickVariant(seed, 3)
	if got < 0 || got >= 3 {
		t.Fatalf("got %d", got)
	}
}

func TestGenerateVerbExamplePair_unknownMoodUsesDefaultFinite(t *testing.T) {
	es, ru := GenerateVerbExamplePair(2, "andar", "otro", "tiempo", "2", "singular", "andarías", "", "", nil)
	if !ExampleContainsSurface(es, "andarías") {
		t.Fatalf("es=%q", es)
	}
	if ru == "" {
		t.Fatal("empty ru")
	}
}

func TestGenerateVerbExamplePair_ruGlossCompactLine(t *testing.T) {
	es, ru := GenerateVerbExamplePair(4, "hablar", "indicativo", "presente", "2", "singular", "hablas", "говорить", "", nil)
	if !ExampleContainsSurface(es, "hablas") {
		t.Fatalf("es=%q", es)
	}
	if !strings.Contains(ru, "«") {
		t.Fatalf("expected compact RU with gloss, got %q", ru)
	}
}

func TestGenerateVerbExamplePair_catalogFallbackWhenSurfaceMissing(t *testing.T) {
	// Catalog may match but with wrong surface — runtime path should still contain surface.
	es, ru := GenerateVerbExamplePair(7, "ir", "indicativo", "presente", "2", "singular", "___wrong___", "", "", nil)
	if !ExampleContainsSurface(es, "___wrong___") {
		t.Fatalf("fallback es=%q", es)
	}
	if ru == "" {
		t.Fatal("empty ru")
	}
}

func TestGenerateVerbExamplePair_allIndicativoTensesRotate(t *testing.T) {
	tenses := []string{
		"presente", "futuro", "imperfecto", "pretérito", "condicional",
		"pretérito perfecto", "pluscuamperfecto", "futuro perfecto",
		"pretérito anterior", "condicional perfecto",
	}
	for i, tense := range tenses {
		es, ru := GenerateVerbExamplePair(int64(i+100), "caminar", "indicativo", tense, "1", "singular", "camino", "", "", nil)
		if !ExampleContainsSurface(es, "camino") {
			t.Fatalf("tense %q es=%q", tense, es)
		}
		if ru == "" {
			t.Fatalf("tense %q empty ru", tense)
		}
	}
}

func TestGenerateVerbExamplePair_subjuntivoAllTenses(t *testing.T) {
	tenses := []string{"presente", "imperfecto", "futuro", "pretérito perfecto", "pluscuamperfecto", "futuro perfecto"}
	for i, tense := range tenses {
		es, ru := GenerateVerbExamplePair(int64(i+200), "venir", "subjuntivo", tense, "3", "plural", "vengan", "", "", nil)
		if !ExampleContainsSurface(es, "vengan") {
			t.Fatalf("tense %q es=%q", tense, es)
		}
		if ru == "" {
			t.Fatalf("tense %q empty ru", tense)
		}
	}
}

func TestMoodTenseKey(t *testing.T) {
	if got := moodTenseKey(" Indicativo ", " Presente "); got != "indicativo|presente" {
		t.Fatalf("got %q", got)
	}
}
