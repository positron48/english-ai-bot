package spanishverbs

import "strings"

// freq100CatalogTemplates: нейтральные «добивки» для топ-100 глаголов по частоте (см. Freq100VerbLemmas).
// LemmaMatch "__freq100__" — см. filterCatalogTemplates; лемма «ir» исключена (отдельный каталог ir).
// RuSecond "gloss" — RuPattern: первый %s = русское местоимение с заглавной, второй %s = спряжённая RU-форма
// по OpenCorpora (github.com/jus1d/gomorphy), иначе — инфинитив из ru_gloss.
const lemmaMatchFreq100 = "__freq100__"

var freq100CatalogTemplates []CatalogTemplate

func init() {
	freq100CatalogTemplates = buildFreq100CatalogTemplates()
}

func buildFreq100CatalogTemplates() []CatalogTemplate {
	frames := []struct {
		id, esSuffix, ruPattern string
	}{
		{id: "fp100_a_veces_asi", esSuffix: "a veces así.", ruPattern: "%s: «%s»."},
		{id: "fp100_sin_complicarse", esSuffix: "sin complicarse.", ruPattern: "%s: «%s»."},
		{id: "fp100_con_calma", esSuffix: "con calma.", ruPattern: "%s: «%s»."},
		{id: "fp100_dia_a_dia", esSuffix: "en el día a día.", ruPattern: "%s: «%s»."},
		{id: "fp100_por_aqui", esSuffix: "por aquí.", ruPattern: "%s: «%s»."},
		{id: "fp100_cuando_falta", esSuffix: "cuando hace falta.", ruPattern: "%s: «%s»."},
		{id: "fp100_muchas_veces", esSuffix: "muchas veces.", ruPattern: "%s: «%s»."},
		{id: "fp100_contexto_neutro", esSuffix: "en contexto neutro.", ruPattern: "%s: «%s»."},
	}
	tenses := irIndicativoTenses()
	out := make([]CatalogTemplate, 0, len(frames)*len(tenses))
	for _, fr := range frames {
		for _, ten := range tenses {
			slug := irTenseSlugForID(ten)
			id := fr.id
			if slug != "presente" {
				id = fr.id + "_" + slug
			}
			out = append(out, CatalogTemplate{
				ID:         id,
				LemmaMatch: lemmaMatchFreq100,
				Mood:       "indicativo",
				Tense:      ten,
				EsSuffix:   fr.esSuffix,
				RuPattern:  fr.ruPattern,
				RuSecond:   "gloss",
			})
		}
	}
	return out
}

// isFreq100SpanishTailExcludedLemma is true for lemmas whose indicativo forms are typically auxiliary-only
// and must not be concatenated with lexical adverbial tails (e.g. "Tú has a veces así").
func isFreq100SpanishTailExcludedLemma(lemma string) bool {
	switch strings.TrimSpace(strings.ToLower(lemma)) {
	case "haber":
		return true
	default:
		return false
	}
}

// Freq100TemplateCodes returns template IDs for the frequency-100 generic pack (offline tools / metadata).
func Freq100TemplateCodes() []string {
	out := make([]string, 0, len(freq100CatalogTemplates))
	for _, t := range freq100CatalogTemplates {
		out = append(out, t.ID)
	}
	return out
}
