package spanishverbs

import (
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/jus1d/gomorphy"
)

// Russian morphology for support lines uses github.com/jus1d/gomorphy (MIT), which embeds
// the OpenCorpora dictionary (https://opencorpora.org/, CC BY-SA 4.0). We only read tags/forms here.

var (
	russianMorphOnce sync.Once
	russianMorph     *gomorphy.Analyzer
	russianMorphErr  error
)

func defaultRussianMorph() (*gomorphy.Analyzer, error) {
	russianMorphOnce.Do(func() {
		russianMorph, russianMorphErr = gomorphy.Default()
	})
	return russianMorph, russianMorphErr
}

var ruCyrillicTokenRE = regexp.MustCompile(`[\p{Cyrillic}-]+`)

// normalizeRussianMorphLemma trims, lowercases, and fixes common Latin/Cyrillic homoglyphs
// in lemmas/tokens so OpenCorpora analysis sees the intended Russian headword.
func normalizeRussianMorphLemma(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'a', 'A':
			r = 'а'
		case 'e', 'E':
			r = 'е'
		case 'o', 'O':
			r = 'о'
		case 'p', 'P':
			r = 'р'
		case 'c', 'C':
			r = 'с'
		case 'x', 'X':
			r = 'х'
		case 'y', 'Y':
			r = 'у'
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return strings.TrimSpace(b.String())
}

// RussianInfinitiveFromRuGloss picks a Russian infinitive token from a DB/embedded gloss
// (may contain parentheses or short gloss after comma).
func RussianInfinitiveFromRuGloss(ruGloss string) string {
	raw := strings.TrimSpace(ruGloss)
	if raw == "" {
		return ""
	}
	s := raw
	if i := strings.IndexByte(s, '('); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.TrimSpace(s)
	s = normalizeRussianMorphLemma(s)
	a, err := defaultRussianMorph()
	if err != nil {
		return ""
	}
	pickFirst := func(scan string) string {
		scan = normalizeRussianMorphLemma(strings.TrimSpace(scan))
		for _, tok := range ruCyrillicTokenRE.FindAllString(scan, -1) {
			tok = normalizeRussianMorphLemma(tok)
			if looksLikeRussianInfinitiveLemma(a, tok) {
				return tok
			}
		}
		return ""
	}
	inf := pickFirst(s)
	if inf == "быть" {
		// Glosses like «быть (вспомог.), иметь» — prefer a non-быть infinitive for conjugation.
		rawNorm := normalizeRussianMorphLemma(raw)
		for _, tok := range ruCyrillicTokenRE.FindAllString(rawNorm, -1) {
			tok = normalizeRussianMorphLemma(tok)
			if tok == "быть" || !looksLikeRussianInfinitiveLemma(a, tok) {
				continue
			}
			return tok
		}
	}
	return inf
}

func normalizeSpanishTenseKey(tense string) string {
	t := strings.TrimSpace(strings.ToLower(tense))
	switch t {
	case "pretérito", "preterito":
		return "preterito"
	case "presente", "imperfecto", "futuro":
		return t
	default:
		return strings.ReplaceAll(t, " ", "_")
	}
}

func opencorporaPersonNumberTags(person, number string) (numTag, perTag string) {
	switch strings.TrimSpace(strings.ToLower(number)) {
	case "plural":
		numTag = "plur"
	default:
		numTag = "sing"
	}
	switch strings.TrimSpace(strings.ToLower(person)) {
	case "1":
		perTag = "1per"
	case "2":
		perTag = "2per"
	case "3":
		perTag = "3per"
	default:
		perTag = "1per"
	}
	return numTag, perTag
}

func pickVerbFormByTag(a *gomorphy.Analyzer, lemma string, wantParts ...string) string {
	forms := a.WordForms(lemma)
	if len(forms) == 0 {
		return ""
	}
	type tagged struct {
		form string
		tag  string
	}
	var hits []tagged
	for _, f := range forms {
		tag := a.Tag(f)
		if strings.Contains(tag, "Arch") {
			continue
		}
		if !strings.Contains(tag, "VERB") && !strings.Contains(tag, "INFN") {
			continue
		}
		ok := true
		for _, p := range wantParts {
			if p == "" || !strings.Contains(tag, p) {
				ok = false
				break
			}
		}
		if ok {
			hits = append(hits, tagged{form: f, tag: tag})
		}
	}
	if len(hits) == 0 {
		return ""
	}
	// Prefer standard dictionary forms over colloquial/erroneous OpenCorpora alternates (Infr/Erro/Dist).
	var preferred []tagged
	for _, h := range hits {
		if strings.Contains(h.tag, "Infr") || strings.Contains(h.tag, "Erro") || strings.Contains(h.tag, "Dist") {
			continue
		}
		preferred = append(preferred, h)
	}
	use := preferred
	if len(use) == 0 {
		use = hits
	}
	return use[0].form
}

func russianVerbFormPostcorrect(lemma, tense, person, number, form string) string {
	form = strings.TrimSpace(form)
	lemma = normalizeRussianMorphLemma(lemma)
	if lemma == "хотеть" && normalizeSpanishTenseKey(tense) == "presente" &&
		strings.EqualFold(strings.TrimSpace(person), "2") &&
		strings.EqualFold(strings.TrimSpace(number), "singular") {
		if strings.EqualFold(form, "хотешь") {
			return "хочешь"
		}
	}
	return form
}

func isRussianInfinitiveShape(w string) bool {
	w = strings.TrimSpace(strings.ToLower(w))
	if len([]rune(w)) < 3 {
		return false
	}
	for _, r := range w {
		if r == '-' || r == 'ё' || (r >= 'а' && r <= 'я') {
			continue
		}
		return false
	}
	switch {
	case strings.HasSuffix(w, "ться"), strings.HasSuffix(w, "ти"), strings.HasSuffix(w, "чь"), strings.HasSuffix(w, "ть"):
		return true
	case strings.HasSuffix(w, "ать"), strings.HasSuffix(w, "еть"), strings.HasSuffix(w, "ить"),
		strings.HasSuffix(w, "оть"), strings.HasSuffix(w, "уть"), strings.HasSuffix(w, "ять"), strings.HasSuffix(w, "ыть"):
		return true
	default:
		return false
	}
}

func looksLikeRussianInfinitiveLemma(a *gomorphy.Analyzer, lemma string) bool {
	lemma = strings.TrimSpace(lemma)
	if !isRussianInfinitiveShape(lemma) {
		return false
	}
	tag := a.Tag(lemma)
	return strings.Contains(tag, "INFN")
}

// RussianVerbFormForSpanishIndicativo returns a finite Russian verb (or analytic future) aligned with
// Spanish indicativo person/number. ok is false if morphology is unavailable or lemma cannot be conjugated.
//
// Spanish pretérito / imperfecto both map to the imperfective past of the Russian lemma (OpenCorpora does not
// encode Spanish aspect on a single imperfective lemma; perfective pairs need a separate lexeme).
func RussianVerbFormForSpanishIndicativo(ruLemma string, tense, person, number string) (form string, ok bool) {
	ruLemma = normalizeRussianMorphLemma(ruLemma)
	if ruLemma == "" {
		return "", false
	}
	a, err := defaultRussianMorph()
	if err != nil {
		return "", false
	}
	if !looksLikeRussianInfinitiveLemma(a, ruLemma) {
		return "", false
	}
	numTag, perTag := opencorporaPersonNumberTags(person, number)
	if strings.EqualFold(ruLemma, "быть") && normalizeSpanishTenseKey(tense) == "presente" {
		if numTag == "sing" && perTag == "3per" {
			w := pickVerbFormByTag(a, ruLemma, "VERB", "pres", "indc", numTag, perTag)
			return w, w != ""
		}
		return "", false
	}
	switch normalizeSpanishTenseKey(tense) {
	case "presente":
		w := pickVerbFormByTag(a, ruLemma, "VERB", "pres", "indc", numTag, perTag)
		w = russianVerbFormPostcorrect(ruLemma, tense, person, number, w)
		return w, w != ""
	case "imperfecto", "preterito":
		if numTag == "plur" {
			w := pickVerbFormByTag(a, ruLemma, "VERB", "past", "indc", "plur")
			return w, w != ""
		}
		w := pickVerbFormByTag(a, ruLemma, "VERB", "past", "indc", "masc", "sing")
		return w, w != ""
	case "futuro":
		aux := pickVerbFormByTag(a, "быть", "VERB", "futr", "indc", numTag, perTag)
		if aux == "" {
			return "", false
		}
		return aux + " " + ruLemma, true
	default:
		return "", false
	}
}

// RussianCatalogSecondFromGloss returns the string to substitute for ru_gloss in freq100-style patterns
// (conjugated finite verb when morphology succeeds; otherwise cleaned infinitive or original gloss).
func RussianCatalogSecondFromGloss(ruGloss, mood, tense, person, number string) string {
	g := strings.TrimSpace(ruGloss)
	if g == "" {
		return ""
	}
	if strings.TrimSpace(strings.ToLower(mood)) != "indicativo" {
		return firstSentenceTokenRu(g)
	}
	lemma := RussianInfinitiveFromRuGloss(g)
	if lemma == "" {
		return firstSentenceTokenRu(g)
	}
	if w, ok := RussianVerbFormForSpanishIndicativo(lemma, tense, person, number); ok && strings.TrimSpace(w) != "" {
		return w
	}
	if lemma != "" {
		return lemma
	}
	return firstSentenceTokenRu(g)
}

func firstSentenceTokenRu(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		if unicode.IsSpace(r) || r == ',' || r == ';' {
			break
		}
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return strings.TrimSpace(s)
	}
	return out
}
