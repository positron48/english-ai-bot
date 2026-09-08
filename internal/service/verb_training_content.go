package service

import (
	"encoding/json"
	"io/fs"
	"regexp"
	"strings"
	"sync"

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
	prompt["question"] = MaskClozeVerbSurfaceInQuestion(question, row.SurfaceForm)
	prompt["content_version"] = 2
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
