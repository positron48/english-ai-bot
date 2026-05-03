package verbtraining

import (
	"testing"
)

func makeValidArtifact() LemmaArtifact {
	cards := make([]GeneratedCard, 0, len(ExpectedScopesV1)*6)
	for _, scope := range ExpectedScopesV1 {
		tense, mood, _ := ParseScope(scope)
		slots := []struct {
			p string
			n string
		}{
			{"1", "singular"},
			{"2", "singular"},
			{"3", "singular"},
			{"1", "plural"},
			{"2", "plural"},
			{"3", "plural"},
		}
		for _, slot := range slots {
			cards = append(cards, GeneratedCard{
				Scope:         scope,
				Mood:          mood,
				Tense:         tense,
				Person:        slot.p,
				Number:        slot.n,
				SurfaceForm:   "form_" + tense + "_" + mood + "_" + slot.p + "_" + slot.n,
				Question:      "___ question " + tense + "_" + mood + "_" + slot.p + "_" + slot.n + " (tener)",
				TranslationRU: "перевод " + tense + "_" + mood + "_" + slot.p + "_" + slot.n,
				Options: []string{
					"bad1_" + tense + "_" + mood + "_" + slot.p + "_" + slot.n,
					"bad2_" + tense + "_" + mood + "_" + slot.p + "_" + slot.n,
					"form_" + tense + "_" + mood + "_" + slot.p + "_" + slot.n,
					"bad3_" + tense + "_" + mood + "_" + slot.p + "_" + slot.n,
				},
			})
		}
	}
	return LemmaArtifact{
		Version:    ArtifactVersionV1,
		Language:   "es",
		Lemma:      "tener",
		WordCardID: 0,
		Cards:      cards,
	}
}

func TestLemmaArtifactValidateStrictCoverage_OK(t *testing.T) {
	a := makeValidArtifact()
	if err := a.ValidateStrictCoverage(); err != nil {
		t.Fatalf("ValidateStrictCoverage: %v", err)
	}
}

func TestLemmaArtifactValidateStrictCoverage_MissingSlot(t *testing.T) {
	a := makeValidArtifact()
	if len(a.Cards) == 0 {
		t.Fatal("test setup")
	}
	a.Cards = a.Cards[1:]
	if err := a.ValidateStrictCoverage(); err == nil {
		t.Fatal("expected missing-slot error")
	}
}

func TestLemmaArtifactValidateStrictCoverage_SameBlankDifferentRUallowed(t *testing.T) {
	a := makeValidArtifact()
	// Два разных лица в одном scope: одинаковый cloze, разные переводы — допускается.
	a.Cards[25].Question = a.Cards[24].Question
	if err := a.ValidateStrictCoverage(); err != nil {
		t.Fatalf("same blank with different RU should pass: %v", err)
	}
}

func TestLemmaArtifactValidateStrictCoverage_DuplicateQuestionTranslationPair(t *testing.T) {
	a := makeValidArtifact()
	a.Cards[25].Question = a.Cards[24].Question
	a.Cards[25].TranslationRU = a.Cards[24].TranslationRU
	if err := a.ValidateStrictCoverage(); err == nil {
		t.Fatal("expected duplicate question+translation_ru error")
	}
}

func TestLemmaArtifactValidateStrictCoverage_TranslationNeedsCyrillic(t *testing.T) {
	a := makeValidArtifact()
	a.Cards[0].TranslationRU = "only latin here"
	if err := a.ValidateStrictCoverage(); err == nil {
		t.Fatal("expected translation_ru_full must contain Cyrillic")
	}
}

func TestLemmaArtifactValidateStrictCoverage_NoCyrillicInSpanish(t *testing.T) {
	a := makeValidArtifact()
	a.Cards[0].Question = "Привет _ здесь (tener)"
	if err := a.ValidateStrictCoverage(); err == nil {
		t.Fatal("expected Spanish question must not contain Cyrillic")
	}
}

