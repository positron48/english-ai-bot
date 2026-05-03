package verbtraining

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"
)

const (
	ArtifactVersionV1 = "v1"
	CardSourceLLMJSON = "llm_verb_forms_json"
)

// ExpectedScopesV1 is the required set for full Spanish verb training coverage.
var ExpectedScopesV1 = []string{
	"es.presente.indicativo",
	"es.preterito_imperfecto.indicativo",
	"es.preterito_indefinido.indicativo",
	"es.futuro_simple.indicativo",
	"es.condicional_simple.indicativo",
	"es.preterito_perfecto_compuesto.indicativo",
	"es.preterito_pluscuamperfecto.indicativo",
	"es.preterito_anterior.indicativo",
	"es.futuro_perfecto.indicativo",
	"es.condicional_perfecto.indicativo",
	"es.presente.subjuntivo",
	"es.preterito_imperfecto.subjuntivo",
	"es.futuro_simple.subjuntivo",
	"es.preterito_perfecto.subjuntivo",
	"es.preterito_pluscuamperfecto.subjuntivo",
	"es.futuro_perfecto.subjuntivo",
}

var expectedSlots = []struct {
	Person string
	Number string
}{
	{Person: "1", Number: "singular"},
	{Person: "2", Number: "singular"},
	{Person: "3", Number: "singular"},
	{Person: "1", Number: "plural"},
	{Person: "2", Number: "plural"},
	{Person: "3", Number: "plural"},
}

// FullCoverageClozeCardCountV1 is how many cloze_form rows in verb_training_cards equal one full lemma artifact (16 scopes × 6 persons).
func FullCoverageClozeCardCountV1() int {
	return len(ExpectedScopesV1) * len(expectedSlots)
}

// LemmaArtifact is one generated file in training_pack/verb_forms/lemmas/<lemma>.json.
type LemmaArtifact struct {
	Version    string          `json:"version"`
	Language   string          `json:"language"`
	Lemma      string          `json:"lemma"`
	WordCardID int64           `json:"word_card_id,omitempty"`
	Generated  string          `json:"generated_at"`
	Cards      []GeneratedCard `json:"cards"`
}

type GeneratedCard struct {
	Scope         string   `json:"scope"`
	Mood          string   `json:"mood"`
	Tense         string   `json:"tense"`
	Person        string   `json:"person"`
	Number        string   `json:"number"`
	SurfaceForm   string   `json:"surface_form"`
	Question      string   `json:"question_es_with_blank"`
	TranslationRU string   `json:"translation_ru_full"`
	Options       []string `json:"options"`
}

type UnlockGates struct {
	Version        string              `json:"version"`
	AlwaysUnlocked []string            `json:"always_unlocked"`
	Chapters       map[string][]string `json:"chapters"`
}

func hasCyrillic(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Cyrillic, r) {
			return true
		}
	}
	return false
}

func normUniqKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Join(strings.Fields(s), " ")
}

func ParseScope(scope string) (tense, mood string, ok bool) {
	scope = strings.ToLower(strings.TrimSpace(scope))
	if !strings.HasPrefix(scope, "es.") {
		return "", "", false
	}
	parts := strings.Split(scope, ".")
	if len(parts) != 3 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func (c *GeneratedCard) Normalize() {
	c.Scope = strings.ToLower(strings.TrimSpace(c.Scope))
	c.Mood = strings.ToLower(strings.TrimSpace(c.Mood))
	c.Tense = strings.ToLower(strings.TrimSpace(c.Tense))
	c.Person = strings.ToLower(strings.TrimSpace(c.Person))
	c.Number = strings.ToLower(strings.TrimSpace(c.Number))
	c.SurfaceForm = strings.ToLower(strings.TrimSpace(c.SurfaceForm))
	c.Question = strings.TrimSpace(c.Question)
	c.TranslationRU = strings.TrimSpace(c.TranslationRU)
	for i := range c.Options {
		c.Options[i] = strings.ToLower(strings.TrimSpace(c.Options[i]))
	}
}

func (a *LemmaArtifact) Normalize() {
	a.Version = strings.ToLower(strings.TrimSpace(a.Version))
	a.Language = strings.ToLower(strings.TrimSpace(a.Language))
	a.Lemma = strings.ToLower(strings.TrimSpace(a.Lemma))
	for i := range a.Cards {
		a.Cards[i].Normalize()
		if a.Cards[i].Scope == "" {
			a.Cards[i].Scope = "es." + a.Cards[i].Tense + "." + a.Cards[i].Mood
		}
	}
}

func (a *LemmaArtifact) ValidateStrictCoverage() error {
	a.Normalize()
	if a.Version == "" {
		a.Version = ArtifactVersionV1
	}
	if a.Version != ArtifactVersionV1 {
		return fmt.Errorf("unsupported artifact version: %s", a.Version)
	}
	if a.Language != "es" {
		return fmt.Errorf("artifact language must be es")
	}
	if a.Lemma == "" {
		return fmt.Errorf("artifact lemma is empty")
	}
	if len(a.Cards) == 0 {
		return fmt.Errorf("artifact cards are empty")
	}
	wantCards := len(ExpectedScopesV1) * len(expectedSlots)
	if len(a.Cards) != wantCards {
		return fmt.Errorf("expected exactly %d cards (full tense × person coverage), got %d", wantCards, len(a.Cards))
	}
	seenPairByScope := map[string]map[string]struct{}{}
	seenCoverage := map[string]map[string]struct{}{}
	for _, card := range a.Cards {
		if card.Scope == "" || card.Mood == "" || card.Tense == "" || card.Person == "" || card.Number == "" {
			return fmt.Errorf("card has empty grammar coordinates")
		}
		if card.SurfaceForm == "" {
			return fmt.Errorf("card has empty surface_form for scope=%s person=%s number=%s", card.Scope, card.Person, card.Number)
		}
		if card.Question == "" || card.TranslationRU == "" {
			return fmt.Errorf("card has empty question/translation for scope=%s person=%s number=%s", card.Scope, card.Person, card.Number)
		}
		if !hasCyrillic(card.TranslationRU) {
			return fmt.Errorf("translation_ru_full must contain Cyrillic for scope=%s person=%s number=%s", card.Scope, card.Person, card.Number)
		}
		if hasCyrillic(card.SurfaceForm) || hasCyrillic(card.Question) {
			return fmt.Errorf("spanish fields must not contain Cyrillic for scope=%s person=%s number=%s", card.Scope, card.Person, card.Number)
		}
		if len(card.Options) != 4 {
			return fmt.Errorf("card options must have exactly 4 entries for scope=%s person=%s number=%s", card.Scope, card.Person, card.Number)
		}
		hasCorrect := false
		for _, opt := range card.Options {
			if hasCyrillic(opt) {
				return fmt.Errorf("spanish option must not contain Cyrillic for scope=%s person=%s number=%s", card.Scope, card.Person, card.Number)
			}
			if opt == card.SurfaceForm {
				hasCorrect = true
			}
		}
		if !hasCorrect {
			return fmt.Errorf("options do not contain surface_form for scope=%s person=%s number=%s", card.Scope, card.Person, card.Number)
		}
		pairKey := normUniqKey(card.Question) + "\x00" + normUniqKey(card.TranslationRU)
		if _, ok := seenPairByScope[card.Scope]; !ok {
			seenPairByScope[card.Scope] = map[string]struct{}{}
		}
		if _, dup := seenPairByScope[card.Scope][pairKey]; dup {
			return fmt.Errorf("duplicate question+translation_ru pair inside scope=%s", card.Scope)
		}
		seenPairByScope[card.Scope][pairKey] = struct{}{}
		if _, ok := seenCoverage[card.Scope]; !ok {
			seenCoverage[card.Scope] = map[string]struct{}{}
		}
		slot := card.Person + ":" + card.Number
		if _, dup := seenCoverage[card.Scope][slot]; dup {
			return fmt.Errorf("duplicate slot in scope=%s for %s", card.Scope, slot)
		}
		seenCoverage[card.Scope][slot] = struct{}{}
	}
	for _, scope := range ExpectedScopesV1 {
		gotSlots, ok := seenCoverage[scope]
		if !ok {
			return fmt.Errorf("missing scope: %s", scope)
		}
		for _, slot := range expectedSlots {
			key := slot.Person + ":" + slot.Number
			if _, ok := gotSlots[key]; !ok {
				return fmt.Errorf("missing slot %s in scope %s", key, scope)
			}
		}
	}
	return nil
}

func EncodePromptJSON(lemma string, card GeneratedCard) (string, error) {
	p := map[string]interface{}{
		"type":                "cloze_form",
		"example_mode":        "llm_json",
		"example_source":      CardSourceLLMJSON,
		"lemma":               lemma,
		"mood":                card.Mood,
		"tense":               card.Tense,
		"person":              card.Person,
		"number":              card.Number,
		"expected_form":       card.SurfaceForm,
		"question":            card.Question,
		"example_translation": card.TranslationRU,
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

