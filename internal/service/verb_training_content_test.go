package service

import (
	"strings"
	"testing"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"
	"tgbot-skeleton/internal/verbtraining"
)

func TestPracticeContent_ArtifactAndAttestedChoices(t *testing.T) {
	row := repository.LinkedVerbFormRow{Lemma: "hablar", Mood: "indicativo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "hablo"}
	p := map[string]interface{}{}
	options := enrichVerbCard(row, p, []string{"hablo", "hablas", "habla", "hablamos", "habláis", "hablan"})
	if p["example_source"] != verbtraining.CardSourceLLMJSON || p["example_translation"] == "" {
		t.Fatalf("artifact not used: %v", p)
	}
	if strings.Contains(p["question"].(string), "hablo") {
		t.Fatal("answer leaked")
	}
	if len(options) != 4 {
		t.Fatal(options)
	}
	row.SurfaceForm = "unmatched"
	enrichVerbCard(row, map[string]interface{}{}, nil)
}
func TestPracticeContent_SyncretismAndAccents(t *testing.T) {
	options := RealVerbFormOptions("hablaba", []string{"hablaba", "hablabas", "hablaba"}, 7)
	if len(options) != 2 {
		t.Fatal(options)
	}
	if got := MaskClozeVerbSurfaceInQuestion("Él habló y habló ayer.", "habló"); got != "Él ____ y ____ ayer." {
		t.Fatal(got)
	}
	if got := MaskClozeVerbSurfaceInQuestion("comemos", "come"); got != "comemos" {
		t.Fatal(got)
	}
}
func TestPracticeContent_RulesTransferOnlyBetweenMatchingForms(t *testing.T) {
	a := spanishverbs.FormRule("hablar", "indicativo", "presente", "2", "singular", "hablas")
	b := spanishverbs.FormRule("trabajar", "indicativo", "presente", "2", "singular", "trabajas")
	if !a.Regular || a.ID != b.ID || a.Ending != "as" {
		t.Fatalf("%+v %+v", a, b)
	}
	c := spanishverbs.FormRule("tener", "indicativo", "presente", "2", "singular", "tienes")
	if c.Regular || c.Pattern != "e_ie" {
		t.Fatal(c)
	}
	d := spanishverbs.FormRule("hablar", "subjuntivo", "imperfecto", "1", "plural", "habláramos")
	if !d.Regular {
		t.Fatal(d)
	}
}
func TestPracticeContent_AssistanceDoesNotIncreaseMastery(t *testing.T) {
	c := &repository.VerbUserCardSRS{Reps: 3, State: "review", EF: 2.5, IntervalDays: 8}
	_, quality := AdvanceVerbSRS(c, true, true)
	if quality != 3 || c.Reps != 3 || c.IntervalDays != 8 {
		t.Fatal(c, quality)
	}
	AdvanceVerbSRS(c, false, true)
	if c.Reps != 0 || c.State != "learning" {
		t.Fatal(c)
	}
}

func TestPracticeContent_CompoundAndSpellingRules(t *testing.T) {
	a := spanishverbs.FormRule("comer", "indicativo", "pretérito perfecto", "1", "singular", "he comido")
	b := spanishverbs.FormRule("vivir", "indicativo", "pretérito perfecto", "1", "singular", "he vivido")
	if !a.Regular || a.ID != b.ID || a.Pattern != "compound" {
		t.Fatal(a, b)
	}
	c := spanishverbs.FormRule("hacer", "indicativo", "pretérito perfecto", "1", "singular", "he hecho")
	if c.Regular || c.ID == a.ID {
		t.Fatal(c)
	}
	d := spanishverbs.FormRule("buscar", "indicativo", "pretérito", "1", "singular", "busqué")
	if !d.Regular || d.Pattern != "spelling_c_qu" {
		t.Fatal(d)
	}
}

func TestPracticeContent_ContextEligibility(t *testing.T) {
	for _, lemma := range []string{"hablar", "haber", "soler"} {
		row := repository.LinkedVerbFormRow{Lemma: lemma, Mood: "indicativo", Tense: "pretérito", Person: "1", Number: "singular"}
		for _, c := range artifactCards(lemma) {
			if c.Mood == row.Mood && verbtraining.CanonicalTense(c.Tense) == row.Tense && c.Person == row.Person && c.Number == row.Number {
				row.SurfaceForm = c.SurfaceForm
				break
			}
		}
		p := map[string]interface{}{}
		enrichVerbCard(row, p, nil)
		if p["practice_eligible"] != (lemma == "hablar") {
			t.Fatalf("%s: %v", lemma, p)
		}
	}
}

func TestPracticeContent_ContrastsUseOnlyAvailableForms(t *testing.T) {
	card := repository.VerbQueueCard{UserVerbCardID: 42, PromptJSON: `{"lemma":"hablar","mood":"indicativo","tense":"presente","person":"1","number":"singular"}`, DistractorsJSON: `["hablo","hablas","habla","hablan"]`}
	rows := []repository.LinkedVerbFormRow{}
	for _, f := range []struct{ tense, form string }{{"presente", "hablo"}, {"pretérito", "hablé"}, {"imperfecto", "hablaba"}} {
		rows = append(rows, repository.LinkedVerbFormRow{Lemma: "hablar", Mood: "indicativo", Tense: f.tense, Person: "1", Number: "singular", SurfaceForm: f.form})
	}
	contrastVerbOptions(&card, rows)
	options := ParseStringJSONArray(card.DistractorsJSON)
	for _, wanted := range []string{"hablo", "hablé", "hablaba"} {
		if !strings.Contains(card.DistractorsJSON, `"`+wanted+`"`) {
			t.Fatal(options)
		}
	}
	if strings.Contains(card.DistractorsJSON, "hablaré") || len(options) != 4 {
		t.Fatal(options)
	}
	card.DistractorsJSON = `["hablo","hablas","habla","hablan"]`
	contrastVerbOptions(&card, rows[:1])
	if strings.Contains(card.DistractorsJSON, "hablé") {
		t.Fatal("locked tense leaked")
	}
}

func TestPracticeContent_CompoundTenseContrastsWithUnlockedSimpleTense(t *testing.T) {
	card := repository.VerbQueueCard{
		UserVerbCardID:  43,
		PromptJSON:      `{"lemma":"hablar","mood":"indicativo","tense":"pretérito perfecto","person":"1","number":"singular"}`,
		DistractorsJSON: `["he hablado","has hablado","ha hablado","hemos hablado"]`,
	}
	rows := []repository.LinkedVerbFormRow{
		{Lemma: "hablar", Mood: "indicativo", Tense: "presente", Person: "1", Number: "singular", SurfaceForm: "hablo"},
		{Lemma: "hablar", Mood: "indicativo", Tense: "pretérito", Person: "1", Number: "singular", SurfaceForm: "hablé"},
		{Lemma: "hablar", Mood: "indicativo", Tense: "pretérito perfecto", Person: "1", Number: "singular", SurfaceForm: "he hablado"},
		{Lemma: "hablar", Mood: "indicativo", Tense: "pluscuamperfecto", Person: "1", Number: "singular", SurfaceForm: "había hablado"},
		{Lemma: "hablar", Mood: "indicativo", Tense: "futuro perfecto", Person: "1", Number: "singular", SurfaceForm: "habré hablado"},
		{Lemma: "hablar", Mood: "indicativo", Tense: "condicional perfecto", Person: "1", Number: "singular", SurfaceForm: "habría hablado"},
	}
	contrastVerbOptions(&card, rows)
	containsSimple := strings.Contains(card.DistractorsJSON, `"hablo"`) || strings.Contains(card.DistractorsJSON, `"hablé"`)
	if !containsSimple || !strings.Contains(card.DistractorsJSON, `"he hablado"`) {
		t.Fatal(ParseStringJSONArray(card.DistractorsJSON))
	}
}
