package spanishverbs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJehleVerbDatabaseCSV_Minimal(t *testing.T) {
	csv := `"infinitive","infinitive_english","mood","mood_english","tense","tense_english","verb_english","form_1s","form_2s","form_3s","form_1p","form_2p","form_3p","gerund","gerund_english","pastparticiple","pastparticiple_english"
"hablar","to speak","Indicativo","Indicative","Presente","Present","I speak","hablo","hablas","habla","hablamos","habláis","hablan","hablando","speaking","hablado","spoken"
`
	lemmas, err := ParseJehleVerbDatabaseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(lemmas) != 1 {
		t.Fatalf("lemmas: got %d want 1", len(lemmas))
	}
	if lemmas[0].Lemma != "hablar" {
		t.Fatalf("lemma: %q", lemmas[0].Lemma)
	}
	if len(lemmas[0].Forms) != 6 {
		t.Fatalf("forms: got %d want 6", len(lemmas[0].Forms))
	}
	var seenPresent1s bool
	for _, f := range lemmas[0].Forms {
		if f.Tense == "presente" && f.Mood == "indicativo" && f.Person == "1" && f.Number == "singular" && f.SurfaceForm == "hablo" {
			seenPresent1s = true
		}
	}
	if !seenPresent1s {
		t.Fatalf("missing hablo presente 1s: %#v", lemmas[0].Forms)
	}
}

func TestParseJehleVerbDatabaseCSV_HaberSupplementBundled(t *testing.T) {
	path := filepath.Join("..", "..", "resources", "verbs", "jehle_supplement_aux_haber.csv")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("read supplement: %v", err)
	}
	lemmas, err := ParseJehleVerbDatabaseCSV(strings.NewReader(string(b)))
	if err != nil {
		t.Fatal(err)
	}
	if len(lemmas) != 1 || lemmas[0].Lemma != "haber" {
		t.Fatalf("lemmas: %#v", lemmas)
	}
	if len(lemmas[0].Forms) < 90 {
		t.Fatalf("expected many finite forms, got %d", len(lemmas[0].Forms))
	}
	var hePresent bool
	for _, f := range lemmas[0].Forms {
		if f.Mood == "indicativo" && f.Tense == "presente" && f.Person == "1" && f.Number == "singular" && f.SurfaceForm == "he" {
			hePresent = true
		}
	}
	if !hePresent {
		t.Fatal("missing presente 1s he")
	}
}
