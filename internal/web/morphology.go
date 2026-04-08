package web

import (
	"encoding/json"
	"strings"
	"sync"

	"tgbot-skeleton/internal/models"
)

var (
	spanishGenderLexiconOnce sync.Once
	spanishGenderLexiconData map[string]models.SpanishGenderLexiconEntry
)

func lookupSpanishGenderLexiconByLemma(lemma string) (models.SpanishGenderLexiconEntry, bool) {
	key := strings.ToLower(strings.TrimSpace(lemma))
	if key == "" {
		return models.SpanishGenderLexiconEntry{}, false
	}
	spanishGenderLexiconOnce.Do(func() {
		lex, _, err := models.LoadSpanishGenderLexiconDefault()
		if err == nil {
			spanishGenderLexiconData = lex
		}
	})
	if len(spanishGenderLexiconData) == 0 {
		return models.SpanishGenderLexiconEntry{}, false
	}
	e, ok := spanishGenderLexiconData[key]
	return e, ok
}

func normalizeNounGenderValue(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "m", "f", "mf", "n":
		return s
	default:
		return ""
	}
}

func nounArticleForTarget(targetLang, nounGender string) string {
	if strings.ToLower(strings.TrimSpace(targetLang)) != "es" {
		return ""
	}
	switch nounGender {
	case "m":
		return "el"
	case "f":
		return "la"
	case "mf":
		return "el/la"
	case "n":
		return "lo"
	default:
		return ""
	}
}

func parseVerbFormsJSON(raw *string) *models.WordInfoVerbForms {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	var forms models.WordInfoVerbForms
	if err := json.Unmarshal([]byte(*raw), &forms); err != nil {
		return nil
	}
	if forms.V1 == "" && forms.V2 == "" && forms.V3 == "" && forms.Gerund == "" && forms.ThirdPerson == "" {
		return nil
	}
	return &forms
}

func canonicalPOS(raw string) string {
	return models.CanonicalWordPOS(raw)
}

func buildCompactMorphFromWordCard(targetLang string, card *models.WordCard, fallbackPOS *string) *models.WordMorphInfo {
	if card == nil && fallbackPOS == nil {
		return nil
	}

	// For training cards, fallbackPOS comes from card.TrainingCard.POS and is
	// more specific than lemma-level POS in word_cards (important for polysemy).
	pos := ""
	if fallbackPOS != nil {
		pos = canonicalPOS(*fallbackPOS)
	}
	if pos == "" && card != nil && card.POS != nil {
		pos = canonicalPOS(*card.POS)
	}

	m := &models.WordMorphInfo{}
	if pos != "" {
		m.POS = pos
	}

	if pos == "noun" && card != nil {
		g := ""
		if card.NounGender != nil {
			g = normalizeNounGenderValue(*card.NounGender)
		}
		opp := ""
		if card.OppositeGenderWord != nil {
			opp = strings.ToLower(strings.TrimSpace(*card.OppositeGenderWord))
		}
		article := ""

		// Fallback to lexicon for noun cards where DB row misses noun metadata
		// (e.g. lemma has multiple POS and stored card was verb-first).
		if (g == "" || opp == "") && strings.EqualFold(strings.TrimSpace(targetLang), "es") {
			if lexEntry, ok := lookupSpanishGenderLexiconByLemma(card.Word); ok {
				if g == "" {
					g = normalizeNounGenderValue(lexEntry.Gender)
				}
				if opp == "" {
					opp = strings.ToLower(strings.TrimSpace(lexEntry.OppositeGenderWord))
				}
				article = strings.TrimSpace(lexEntry.Article)
			}
		}

		if g != "" {
			m.NounGender = g
			m.Article = nounArticleForTarget(targetLang, g)
			if m.Article == "" && article != "" {
				m.Article = article
			}
		}
		if opp != "" {
			m.OppositeGenderWord = opp
		}
	}
	if pos == "verb" && card != nil {
		m.VerbForms = parseVerbFormsJSON(card.VerbFormsJSON)
	}

	if m.POS == "" && m.NounGender == "" && m.Article == "" && m.OppositeGenderWord == "" && m.VerbForms == nil {
		return nil
	}
	return m
}
