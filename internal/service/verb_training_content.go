package service

import (
	"encoding/json"
	"io/fs"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"tgbot-skeleton/internal/grammartrainingpack"
	"tgbot-skeleton/internal/repository"
	"tgbot-skeleton/internal/spanishverbs"
	"tgbot-skeleton/internal/verbtraining"
)

var verbArtifactCache sync.Map

func artifactCards(lemma string) []verbtraining.GeneratedCard {
	if cached, ok := verbArtifactCache.Load(lemma); ok {
		return cached.([]verbtraining.GeneratedCard)
	}
	cards := []verbtraining.GeneratedCard{}
	if pack, err := grammartrainingpack.PackFS("es"); err == nil {
		if raw, err := fs.ReadFile(pack, "verb_forms/lemmas/"+lemma+".json"); err == nil {
			var artifact verbtraining.LemmaArtifact
			if json.Unmarshal(raw, &artifact) == nil {
				cards = artifact.Cards
			}
		}
	}
	verbArtifactCache.Store(lemma, cards)
	return cards
}

func enrichVerbCard(row repository.LinkedVerbFormRow, prompt map[string]interface{}, forms []string) []string {
	// Artifact text takes precedence over the generic runtime fallback. Match the
	// actual dictionary answer as well as the slot: stale artifacts cannot change it.
	for _, c := range artifactCards(row.Lemma) {
		if verbtraining.CanonicalTense(c.Tense) == verbtraining.CanonicalTense(row.Tense) && c.Mood == row.Mood && c.Person == row.Person && c.Number == row.Number && strings.EqualFold(c.SurfaceForm, row.SurfaceForm) {
			prompt["question"] = regexp.MustCompile(`_+`).ReplaceAllString(c.Question, "____")
			prompt["example_translation"] = c.TranslationRU
			prompt["example_source"] = verbtraining.CardSourceLLMJSON
			break
		}
	}
	if prompt["example_source"] != verbtraining.CardSourceLLMJSON {
		prompt["question"] = SpanishVerbRecallContrastQuestion(row.Lemma, row.Person, row.Number, row.VerbFormDictID, false)
		prompt["example_translation"] = ""
	}
	question, _ := prompt["question"].(string)
	prompt["question"] = strings.TrimSpace(strings.TrimSuffix(MaskClozeVerbSurfaceInQuestion(question, row.SurfaceForm), "("+row.Lemma+")"))
	prompt["content_version"] = 4
	prompt["practice_eligible"] = contextualVerbExample(row, prompt)
	prompt["tense"] = verbtraining.CanonicalTense(row.Tense)
	rule := spanishverbs.FormRule(row.Lemma, row.Mood, row.Tense, row.Person, row.Number, row.SurfaceForm)
	prompt["rule"] = rule
	wrong := []string{}
	for _, form := range forms {
		if !spanishverbs.AcceptedVerbAnswer(row.SurfaceForm, form, row.Mood, row.Tense) {
			wrong = append(wrong, form)
		}
	}
	return RealVerbFormOptions(row.SurfaceForm, wrong, stableVerbTrainingSeed(row))
}

// Rare literary paradigms and metalinguistic prompts do not teach tense from context.
// Haber is practiced as an auxiliary in compound forms of the other verbs.
func contextualVerbExample(row repository.LinkedVerbFormRow, prompt map[string]interface{}) bool {
	q, _ := prompt["question"].(string)
	translation, _ := prompt["example_translation"].(string)
	tense := verbtraining.CanonicalTense(row.Tense)
	if row.Lemma == "haber" || tense == "pretérito anterior" || (row.Mood == "subjuntivo" && (tense == "futuro" || tense == "futuro perfecto")) {
		return false
	}
	if row.Lemma == "soler" && tense != "presente" && tense != "imperfecto" {
		return false
	}
	if row.Lemma == "consistir" && row.Person != "3" {
		return false
	}
	lower := strings.ToLower(q)
	for _, marker := range []string{"metaling", "pronombre", "jurídic", "arcaic", "se lee:", "ejemplo gramatical"} {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	return translation != "" && utf8.RuneCountInString(q) <= 110 && strings.Count(q, "____") == 1
}

// Build per-session options: stored training cards are shared by users, whereas
// unlocked scopes are personal. Different tenses keep the same grammatical person.
func contrastVerbOptions(card *repository.VerbQueueCard, rows []repository.LinkedVerbFormRow) {
	var p struct{ Lemma, Mood, Tense, Person, Number string }
	if json.Unmarshal([]byte(card.PromptJSON), &p) != nil {
		return
	}
	// Translation supplies temporal meaning and past aspect. Mix simple and
	// compound indicative forms so the learner must identify the tense as well
	// as the person. Only forms from the user's unlocked scopes are present in rows.
	family := map[string]bool{
		"presente": true, "pretérito": true, "imperfecto": true, "futuro": true, "condicional": true,
		"pretérito perfecto": true, "pluscuamperfecto": true, "futuro perfecto": true, "condicional perfecto": true,
	}
	if p.Mood != "indicativo" || !family[verbtraining.CanonicalTense(p.Tense)] {
		return
	}
	var correct string
	for _, row := range rows {
		if row.Lemma == p.Lemma && row.Mood == p.Mood && row.Person == p.Person && row.Number == p.Number && verbtraining.CanonicalTense(row.Tense) == verbtraining.CanonicalTense(p.Tense) {
			correct = row.SurfaceForm
			break
		}
	}
	if correct == "" {
		return
	}
	different := []string{}
	simple := []string{}
	for _, row := range rows {
		// Present can also express a scheduled future (mañana hablo).
		if verbtraining.CanonicalTense(p.Tense) == "futuro" && verbtraining.CanonicalTense(row.Tense) == "presente" {
			continue
		}
		if row.Lemma == p.Lemma && row.Mood == p.Mood && row.Person == p.Person && row.Number == p.Number && family[verbtraining.CanonicalTense(row.Tense)] && verbtraining.CanonicalTense(row.Tense) != verbtraining.CanonicalTense(p.Tense) && !spanishverbs.AcceptedVerbAnswer(correct, row.SurfaceForm, p.Mood, p.Tense) {
			different = append(different, row.SurfaceForm)
			if isSimpleIndicativeTense(row.Tense) {
				simple = append(simple, row.SurfaceForm)
			}
		}
	}
	// A compound-tense question must contain at least one simple-tense contrast
	// when the learner has one unlocked. Otherwise it degrades into choosing he/has/ha.
	if !isSimpleIndicativeTense(p.Tense) && len(simple) > 0 {
		preferred := simple[int(card.UserVerbCardID%int64(len(simple)))]
		prioritized := []string{preferred}
		for _, option := range different {
			if option != preferred {
				prioritized = append(prioritized, option)
			}
		}
		different = prioritized
		if len(different) > VerbChoiceOptionCount-1 {
			different = different[:VerbChoiceOptionCount-1]
		}
	}
	// Select the cross-tense contrasts first, then fill with same-tense persons.
	chosen := RealVerbFormOptions(correct, different, card.UserVerbCardID)
	for _, option := range ParseStringJSONArray(card.DistractorsJSON) {
		if len(chosen) >= VerbChoiceOptionCount {
			break
		}
		found := false
		for _, existing := range chosen {
			if existing == option {
				found = true
			}
		}
		if !found {
			chosen = append(chosen, option)
		}
	}
	raw, _ := json.Marshal(RealVerbFormOptions(correct, chosen, card.UserVerbCardID))
	card.DistractorsJSON = string(raw)
}

func isSimpleIndicativeTense(tense string) bool {
	switch verbtraining.CanonicalTense(tense) {
	case "presente", "pretérito", "imperfecto", "futuro", "condicional":
		return true
	default:
		return false
	}
}
