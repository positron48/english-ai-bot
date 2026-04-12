package spanishverbs

import "strings"

// irCatalogTemplates is filled in init(): 8 motion frames × 4 indicativo tenses (presente, imperfecto, pretérito, futuro).
var irCatalogTemplates []CatalogTemplate

func init() {
	irCatalogTemplates = buildIrCatalogTemplates()
}

func irIndicativoTenses() []string {
	return []string{"presente", "imperfecto", "pretérito", "futuro"}
}

func irTenseSlugForID(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "pretérito", "preterito":
		return "preterito"
	case "presente":
		return "presente"
	case "imperfecto":
		return "imperfecto"
	case "futuro":
		return "futuro"
	default:
		return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(t)), " ", "_")
	}
}

func buildIrCatalogTemplates() []CatalogTemplate {
	tails := []struct {
		baseID, esSuffix, ruFmt string
	}{
		{"ir_a_casa", "a casa.", "%s %s домой."},
		{"ir_al_trabajo", "al trabajo.", "%s %s на работу."},
		{"ir_a_la_escuela", "a la escuela.", "%s %s в школу."},
		{"ir_al_mercado", "al mercado.", "%s %s на рынок."},
		{"ir_al_parque", "al parque.", "%s %s в парк."},
		{"ir_a_pie", "a pie.", "%s %s пешком."},
		{"ir_de_compras", "de compras.", "%s %s за покупками."},
		{"ir_de_viaje", "de viaje.", "%s %s в поездку."},
	}
	tenses := irIndicativoTenses()
	out := make([]CatalogTemplate, 0, len(tails)*len(tenses))
	for _, tail := range tails {
		for _, ten := range tenses {
			slug := irTenseSlugForID(ten)
			id := tail.baseID + "_" + slug
			out = append(out, CatalogTemplate{
				ID:         id,
				VerbClass:  VerbClassMotion,
				Mood:       "indicativo",
				Tense:      ten,
				EsSuffix:   tail.esSuffix,
				RuPattern:  tail.ruFmt,
				LemmaMatch: "ir",
				RuSecond:   "",
			})
		}
	}
	return out
}

// IrTemplateCodes returns catalog template IDs for lemma "ir" (offline tools / DB seeds).
func IrTemplateCodes() []string {
	out := make([]string, 0, len(irCatalogTemplates))
	for _, t := range irCatalogTemplates {
		out = append(out, t.ID)
	}
	return out
}
