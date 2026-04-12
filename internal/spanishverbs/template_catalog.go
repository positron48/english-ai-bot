package spanishverbs

import (
	"fmt"
	"strings"
)

// VerbClass is a coarse semantic bucket for template compatibility.
const (
	VerbClassMotion  = "motion"
	VerbClassSpeech  = "speech"
	VerbClassTransfer = "transfer"
	VerbClassGeneric = "generic"
)

// CatalogTemplate is one ES+RU pair pattern for a lemma (runtime or DB-backed).
type CatalogTemplate struct {
	ID         string
	VerbClass  string // e.g. VerbClassMotion
	Mood       string // lowercase, empty = any
	Tense      string // lowercase, empty = any
	EsSuffix   string // appended after "{subj} {form} " — e.g. "a casa"
	RuPattern  string // fmt: two %s — subject + second arg (see RuSecond)
	LemmaMatch string // lowercase lemma, e.g. "ir", or "__freq100__" for IsFreq100VerbLemma (except ir)
	// RuSecond: "" — second arg is Russian motion verb for lemma "ir" (RussianIrVerbForSpanishIndicativo) or RussianIrPresent fallback for other lemmas. "gloss" — ru_gloss.
	RuSecond string
}

// RussianSubjectCap returns "Я", "Ты", … for sentence-initial Russian.
func RussianSubjectCap(person, number string) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch p {
	case "1":
		if n == "plural" {
			return "Мы"
		}
		return "Я"
	case "2":
		if n == "plural" {
			return "Вы"
		}
		return "Ты"
	case "3":
		if n == "plural" {
			return "Они"
		}
		return "Он"
	default:
		return "Я"
	}
}

func moodTenseMatchCatalog(mood, tense, tplMood, tplTense string) bool {
	tplMood = strings.TrimSpace(strings.ToLower(tplMood))
	tplTense = strings.TrimSpace(strings.ToLower(tplTense))
	if tplMood == "" && tplTense == "" {
		return true
	}
	got := moodTenseKey(mood, tense)
	candidates := []string{moodTenseKey(tplMood, tplTense)}
	// Accept pretérito / preterito spelling drift vs template row.
	if tplTense == "pretérito" || tplTense == "preterito" {
		candidates = append(candidates,
			moodTenseKey(tplMood, "pretérito"),
			moodTenseKey(tplMood, "preterito"),
		)
	}
	for _, c := range candidates {
		if c == got {
			return true
		}
	}
	return false
}

// externalCatalogTemplates merged from DB at runtime (optional).
var externalCatalogTemplates []CatalogTemplate

// RegisterExternalTemplates replaces externally loaded templates (nil clears).
func RegisterExternalTemplates(t []CatalogTemplate) {
	if len(t) == 0 {
		externalCatalogTemplates = nil
		return
	}
	externalCatalogTemplates = append([]CatalogTemplate(nil), t...)
}

func allCatalogTemplates() []CatalogTemplate {
	seen := map[string]struct{}{}
	out := make([]CatalogTemplate, 0, len(irCatalogTemplates)+len(freq100CatalogTemplates)+len(externalCatalogTemplates))
	for _, t := range irCatalogTemplates {
		seen[strings.ToLower(t.ID)] = struct{}{}
		out = append(out, t)
	}
	for _, t := range freq100CatalogTemplates {
		seen[strings.ToLower(t.ID)] = struct{}{}
		out = append(out, t)
	}
	for _, t := range externalCatalogTemplates {
		if _, dup := seen[strings.ToLower(t.ID)]; dup {
			continue
		}
		out = append(out, t)
	}
	return out
}

func catalogTemplateMatchesLemma(t CatalogTemplate, lemma string) bool {
	lm := strings.TrimSpace(strings.ToLower(t.LemmaMatch))
	if lm == lemmaMatchFreq100 {
		if lemma == "ir" {
			return false
		}
		// Spanish tails like "a veces así" after a bare auxiliary (he/has/…) are ungrammatical for *haber*.
		if isFreq100SpanishTailExcludedLemma(lemma) {
			return false
		}
		return IsFreq100VerbLemma(lemma)
	}
	return lm == lemma
}

// filterCatalogTemplates returns templates matching lemma, mood, tense, verbClass and optional allowedIDs (empty = all).
func filterCatalogTemplates(lemma, mood, tense, verbClass, ruGloss string, allowedIDs []string) []CatalogTemplate {
	lemma = strings.TrimSpace(strings.ToLower(lemma))
	verbClass = strings.TrimSpace(strings.ToLower(verbClass))
	allow := map[string]struct{}{}
	for _, id := range allowedIDs {
		id = strings.TrimSpace(strings.ToLower(id))
		if id != "" {
			allow[id] = struct{}{}
		}
	}
	var out []CatalogTemplate
	for _, t := range allCatalogTemplates() {
		if !catalogTemplateMatchesLemma(t, lemma) {
			continue
		}
		if catalogRuSecondIsGloss(t) && strings.TrimSpace(ruGloss) == "" {
			continue
		}
		if verbClass != "" && t.VerbClass != "" && strings.ToLower(t.VerbClass) != verbClass {
			continue
		}
		if len(allow) > 0 {
			if _, ok := allow[strings.ToLower(t.ID)]; !ok {
				continue
			}
		}
		if !moodTenseMatchCatalog(mood, tense, t.Mood, t.Tense) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func catalogRuSecondIsGloss(t CatalogTemplate) bool {
	switch strings.TrimSpace(strings.ToLower(t.RuSecond)) {
	case "gloss":
		return true
	default:
		return false
	}
}

// TryGenerateCatalogPair builds ES + simple RU from catalog; ok=false if no template matched.
func TryGenerateCatalogPair(seed int64, lemma, mood, tense, person, number, surfaceForm, ruGloss, verbClass string, allowedTemplateIDs []string) (es string, ru string, ok bool) {
	cands := filterCatalogTemplates(lemma, mood, tense, verbClass, ruGloss, allowedTemplateIDs)
	if len(cands) == 0 {
		return "", "", false
	}
	t := cands[pickVariant(seed, len(cands))]
	subjES := SpanishSubjectCapital(person, number)
	esCore := fmt.Sprintf("%s %s %s", subjES, strings.TrimSpace(surfaceForm), strings.TrimSpace(t.EsSuffix))
	es = fmt.Sprintf("%s (%s)", esCore, strings.TrimSpace(strings.ToLower(lemma)))
	ruSubj := RussianSubjectCap(person, number)
	var ruLine string
	if catalogRuSecondIsGloss(t) {
		ruSecond := RussianCatalogSecondFromGloss(ruGloss, mood, tense, person, number)
		ruLine = fmt.Sprintf(t.RuPattern, ruSubj, ruSecond)
	} else {
		ruVerb := RussianIrPresent(person, number)
		if strings.TrimSpace(strings.ToLower(lemma)) == "ir" {
			ruVerb = RussianIrVerbForSpanishIndicativo(mood, tense, person, number)
		}
		ruLine = fmt.Sprintf(t.RuPattern, ruSubj, ruVerb)
	}
	return es, strings.TrimSpace(ruLine), true
}
