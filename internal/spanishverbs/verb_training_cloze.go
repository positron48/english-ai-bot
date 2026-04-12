package spanishverbs

import (
	"fmt"
	"strings"
)

// ClozeSpanishMoodTenseLabel is a short footer label after the lemma in verb-training cloze, e.g. "presente" or "presente, subjuntivo".
func ClozeSpanishMoodTenseLabel(mood, tense string) string {
	m := strings.TrimSpace(strings.ToLower(mood))
	t := strings.TrimSpace(tense)
	if t == "" {
		t = "?"
	}
	if m == "" || m == "indicativo" {
		return strings.ToLower(t)
	}
	return strings.ToLower(t) + ", " + m
}

// BuildVerbTrainingClozeQuestion returns a line like "Yo ___ (haber, presente)" for web verb training.
func BuildVerbTrainingClozeQuestion(person, number, lemma, mood, tense string) string {
	subj := SpanishSubjectCapital(person, number)
	lem := strings.ToLower(strings.TrimSpace(lemma))
	if lem == "" {
		lem = "?"
	}
	slot := ClozeSpanishMoodTenseLabel(mood, tense)
	return fmt.Sprintf("%s ___ (%s, %s)", subj, lem, slot)
}

// PlainRussianVerbTrainingTranslation returns only the Russian verb form (or analytic future phrase),
// without subject prefix or guillemets — for a simple subtitle under the Spanish cloze.
func PlainRussianVerbTrainingTranslation(person, number, ruGloss, mood, tense string) string {
	if strings.TrimSpace(ruGloss) == "" {
		return ""
	}
	s := strings.TrimSpace(RussianCatalogSecondFromGloss(ruGloss, mood, tense, person, number))
	s = strings.Trim(s, " \t\n\r«»\"'")
	return s
}

// PlainRussianVerbTrainingTranslationForLemma applies lemma-specific gloss overrides so a mis-linked
// ru_gloss (e.g. from another verb) does not produce absurd forms for high-frequency auxiliaries.
func PlainRussianVerbTrainingTranslationForLemma(spanishLemma, person, number, ruGloss, mood, tense string) string {
	g := strings.TrimSpace(ruGloss)
	switch strings.ToLower(strings.TrimSpace(spanishLemma)) {
	case "haber":
		g = "иметь"
	default:
	}
	return PlainRussianVerbTrainingTranslation(person, number, g, mood, tense)
}

// PlainRussianVerbTrainingHintLine returns a short hint line: lowercase Russian pronoun + verb form
// (e.g. "они имеют", "я буду говорить") aligned with the Spanish person/number in the cloze.
func PlainRussianVerbTrainingHintLine(spanishLemma, person, number, ruGloss, mood, tense string) string {
	form := PlainRussianVerbTrainingTranslationForLemma(spanishLemma, person, number, ruGloss, mood, tense)
	if form == "" {
		return ""
	}
	pron := strings.ToLower(strings.TrimSpace(RussianSubjectCap(person, number)))
	if pron == "" {
		return form
	}
	return pron + " " + form
}
