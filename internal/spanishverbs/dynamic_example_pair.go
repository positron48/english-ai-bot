package spanishverbs

import (
	"fmt"
	"strings"
	"unicode"
)

// ExampleModeRuntime marks verb cloze cards whose Spanish/Russian lines are built at request time (no DB examples).
const ExampleModeRuntime = "runtime"

// pickVariant returns a stable index in [0, n) from seed.
func pickVariant(seed int64, n int) int {
	if n <= 0 {
		return 0
	}
	x := seed ^ (seed >> 33)
	if x < 0 {
		x = -x
	}
	return int(x % int64(n))
}

// SpanishSubjectCapital returns a capitalized subject pronoun for finite forms (Jehle person 1–3, number singular|plural).
func SpanishSubjectCapital(person, number string) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch p {
	case "1":
		if n == "plural" {
			return "Nosotros"
		}
		return "Yo"
	case "2":
		if n == "plural" {
			return "Vosotros"
		}
		return "Tú"
	case "3":
		if n == "plural" {
			return "Ellos"
		}
		return "Él"
	default:
		return "Yo"
	}
}

// SpanishSubjectInSentence returns a subject pronoun with lowercase first letter for mid-sentence use.
func SpanishSubjectInSentence(person, number string) string {
	s := SpanishSubjectCapital(person, number)
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// RussianSubjectLower returns a lowercase Russian pronoun for mid-sentence use.
func RussianSubjectLower(person, number string) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch p {
	case "1":
		if n == "plural" {
			return "мы"
		}
		return "я"
	case "2":
		if n == "plural" {
			return "вы"
		}
		return "ты"
	case "3":
		if n == "plural" {
			return "они"
		}
		return "он"
	default:
		return "я"
	}
}

func startsWithReflexiveClitic(surface string) bool {
	fields := strings.Fields(strings.TrimSpace(surface))
	if len(fields) < 2 {
		return false
	}
	switch strings.ToLower(fields[0]) {
	case "me", "te", "se", "nos", "os":
		return true
	default:
		return false
	}
}

func moodTenseKey(mood, tense string) string {
	return strings.TrimSpace(strings.ToLower(mood)) + "|" + strings.TrimSpace(strings.ToLower(tense))
}

type esPair struct {
	esFmt string // first %s = subject (may be ""), second %s = surface form
}

func finiteTemplates(mood, tense string) []esPair {
	key := moodTenseKey(mood, tense)
	switch key {
	case "indicativo|presente":
		return []esPair{
			{"A veces %s %s por aquí, sin complicarse."},
			{"Muchas veces %s %s así, con naturalidad."},
			{"Normalmente %s %s en estas situaciones."},
			{"Por lo general %s %s cuando hace falta."},
			{"Hoy %s %s y nadie se extraña."},
			{"En el día a día %s %s sin pensarlo mucho."},
		}
	case "indicativo|futuro":
		return []esPair{
			{"El año que viene %s %s con más calma."},
			{"Pronto %s %s y lo veremos."},
			{"Más adelante %s %s sin problema."},
			{"En cuanto pueda %s %s sin prisas."},
		}
	case "indicativo|imperfecto":
		return []esPair{
			{"Antes %s %s a menudo por aquí."},
			{"En aquella época %s %s casi todos los días."},
			{"Cuando era más joven %s %s con frecuencia."},
			{"Solía pasar que %s %s sin darse cuenta."},
		}
	case "indicativo|pretérito", "indicativo|preterito":
		return []esPair{
			{"Ese día %s %s de repente."},
			{"Ayer por la tarde %s %s sin avisar."},
			{"La semana pasada %s %s una sola vez."},
			{"En aquel momento %s %s y listo."},
		}
	case "indicativo|condicional":
		return []esPair{
			{"Con más tiempo %s %s sin dudarlo."},
			{"En tu lugar %s %s igual."},
			{"Si hubiera apoyo %s %s encantado."},
		}
	case "indicativo|pretérito perfecto", "indicativo|preterito perfecto":
		return []esPair{
			{"Esta semana %s %s varias veces."},
			{"Últimamente %s %s más de lo habitual."},
			{"En los últimos días %s %s con constancia."},
		}
	case "indicativo|pluscuamperfecto":
		return []esPair{
			{"Para entonces %s %s ya hacía tiempo."},
			{"Cuando llegó la noticia %s %s todo listo."},
			{"Antes de salir %s %s en silencio."},
		}
	case "indicativo|futuro perfecto":
		return []esPair{
			{"Para mañana a mediodía %s %s seguro."},
			{"Cuando vuelvas %s %s ya todo hecho."},
		}
	case "indicativo|pretérito anterior", "indicativo|preterito anterior":
		return []esPair{
			{"Apenas terminó la reunión %s %s."},
			{"En cuanto cerró la puerta %s %s."},
		}
	case "indicativo|condicional perfecto":
		return []esPair{
			{"De haber sabido antes %s %s distinto."},
			{"Con más aviso %s %s sin problema."},
		}
	case "subjuntivo|presente":
		return []esPair{
			{"Espero que %s %s pronto."},
			{"Quiero que %s %s sin demora."},
			{"Prefiero que %s %s con calma."},
			{"No creo que %s %s tarde."},
		}
	case "subjuntivo|imperfecto":
		return []esPair{
			{"No creía que %s %s a tiempo."},
			{"Dudaba de que %s %s tan pronto."},
			{"Parecía imposible que %s %s así."},
		}
	case "subjuntivo|futuro":
		return []esPair{
			{"Cuando %s %s, ya hablaremos."},
			{"Si algún día %s %s, lo notarás."},
		}
	case "subjuntivo|pretérito perfecto", "subjuntivo|preterito perfecto":
		return []esPair{
			{"Me alegra que %s %s."},
			{"Es posible que %s %s ya."},
		}
	case "subjuntivo|pluscuamperfecto":
		return []esPair{
			{"Como si %s %s toda la vida."},
			{"Parecía que %s %s desde siempre."},
		}
	case "subjuntivo|futuro perfecto":
		return []esPair{
			{"Cuando %s %s, será tarde para arrepentirse."},
			{"Si para entonces %s %s, mejor."},
		}
	default:
		return []esPair{{"%s %s aquí, en contexto neutro."}}
	}
}

func imperativeAffirmativeTemplates(person, number string) []string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch {
	case p == "2" && n == "singular":
		return []string{
			"¡%s, por favor!",
			"%s, sin prisa.",
			"%s cuando quieras.",
			"%s, con calma.",
		}
	case p == "3" && n == "singular":
		return []string{
			"Que %s, si hace falta.",
			"%s, si puede.",
		}
	case p == "1" && n == "plural":
		return []string{
			"%s mañana sin falta.",
			"%s con calma.",
		}
	case p == "2" && n == "plural":
		return []string{
			"%s sin correr.",
			"%s cuando podáis.",
		}
	case p == "3" && n == "plural":
		return []string{
			"%s todos a la vez.",
			"Que %s si pueden.",
		}
	default:
		return []string{"%s."}
	}
}

func imperativeNegativeTemplates(person, number string) []esPair {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch {
	case p == "2" && n == "singular":
		return []esPair{
			{"%s todavía; mejor espera."},
			{"%s ahora: hay tiempo."},
		}
	case p == "3" && n == "singular":
		return []esPair{
			{"%s sin prisa."},
		}
	case p == "1" && n == "plural":
		return []esPair{
			{"%s hoy; lo hablamos luego."},
		}
	case p == "2" && n == "plural":
		return []esPair{
			{"%s todavía; calmémonos."},
		}
	case p == "3" && n == "plural":
		return []esPair{
			{"%s aún; no hay apuro."},
		}
	default:
		return []esPair{{"%s."}}
	}
}

func formatFiniteES(t esPair, subj, form string) string {
	return fmt.Sprintf(t.esFmt, subj, form)
}

// russianHintForMoodTense is a short Russian grammar note (not a full lexical translation of the verb).
func russianHintForMoodTense(mood, tense string) string {
	key := moodTenseKey(mood, tense)
	switch key {
	case "indicativo|presente":
		return "Настоящее изъявительное: действие сейчас или как привычка."
	case "indicativo|futuro":
		return "Будущее изъявительное: действие позже."
	case "indicativo|imperfecto":
		return "Прошедшее несовершенное: длительность или повтор в прошлом."
	case "indicativo|pretérito", "indicativo|preterito":
		return "Претерито: завершённое действие в прошлом."
	case "indicativo|condicional":
		return "Условное наклонение: «бы» в гипотезах."
	case "indicativo|pretérito perfecto", "indicativo|preterito perfecto":
		return "Претерито перфекто: связь с настоящим."
	case "indicativo|pluscuamperfecto":
		return "Плюсквамперфекто: «ещё раньше» относительно другого прошлого."
	case "indicativo|futuro perfecto":
		return "Будущее перфекто: завершённость к будущему моменту."
	case "indicativo|pretérito anterior", "indicativo|preterito anterior":
		return "Претерито антериор (книжн.): действие сразу после другого."
	case "indicativo|condicional perfecto":
		return "Условное перфекто: «бы сделал» к прошлому."
	case "subjuntivo|presente":
		return "Сослагательное настоящее: после «quiero que…», «espero que…»."
	case "subjuntivo|imperfecto":
		return "Сослагательное прошедшее: сомнение, нереальность."
	case "subjuntivo|futuro":
		return "Сослагательное будущее (редко): «когда бы ни…»."
	case "subjuntivo|pretérito perfecto", "subjuntivo|preterito perfecto":
		return "Сослагательное перфекто: прошлое относительно главной клаузы."
	case "subjuntivo|pluscuamperfecto":
		return "Сослагательное плюсквамперфекто: «как будто уже»."
	case "subjuntivo|futuro perfecto":
		return "Сослагательное будущее перфекто (редко)."
	case "imperativo afirmativo|presente":
		return "Повелительное: просьба или приказ (утвердительная форма)."
	case "imperativo negativo|presente":
		return "Отрицательное повелительное: с «no» + сослагательное."
	default:
		return "Испанская глагольная форма в учебной рамке."
	}
}

var russianIntroVariants = []string{
	"Так звучит типичная испанская конструкция:",
	"В разговорной речи это может выглядеть так:",
	"Нейтральный пример на испанском:",
	"Повседневная формулировка:",
	"Шаблон для запоминания спряжения:",
	"Короткая модель фразы:",
	"Пример в естественном порядке слов:",
	"Бытовой контекст без лишних деталей:",
}

// BuildRussianExampleLine builds a Russian support line: intro, quoted Spanish, grammar hint, lemma and form.
func BuildRussianExampleLine(es, lemma, surface, mood, tense string, seed int64) string {
	intro := russianIntroVariants[pickVariant(seed, len(russianIntroVariants))]
	hint := russianHintForMoodTense(mood, tense)
	lemma = strings.TrimSpace(strings.ToLower(lemma))
	surface = strings.TrimSpace(surface)
	return fmt.Sprintf("%s «%s» — %s Инфинитив: «%s»; форма: «%s».", intro, es, hint, lemma, surface)
}

// GenerateVerbExamplePair returns a Spanish practice sentence containing surfaceForm and a Russian line.
// ruGloss comes from verb_lemmas.metadata_json, embedded defaults, or LLM batch; when empty, Russian is grammar-oriented.
// seed should vary per session/card so the same lemma gets different ES/RU frames over time.
// verbClass and allowedTemplateIDs tune the deterministic template catalog (e.g. lemma "ir"); empty values use catalog defaults.
func GenerateVerbExamplePair(seed int64, lemma, mood, tense, person, number, surfaceForm, ruGloss, verbClass string, allowedTemplateIDs []string) (string, string) {
	surfaceForm = strings.TrimSpace(surfaceForm)
	lemma = strings.TrimSpace(lemma)
	mood = strings.TrimSpace(strings.ToLower(mood))
	subj := SpanishSubjectInSentence(person, number)

	if esCat, ruCat, ok := TryGenerateCatalogPair(seed, lemma, mood, tense, person, number, surfaceForm, ruGloss, verbClass, allowedTemplateIDs); ok {
		if ExampleContainsSurface(esCat, surfaceForm) {
			return esCat, ruCat
		}
	}

	mImperAff := mood == "imperativo afirmativo"
	mImperNeg := mood == "imperativo negativo"

	var esOut string
	switch {
	case mImperAff:
		aff := imperativeAffirmativeTemplates(person, number)
		tpl := aff[pickVariant(seed, len(aff))]
		esOut = fmt.Sprintf(tpl, surfaceForm)
	case mImperNeg:
		neg := imperativeNegativeTemplates(person, number)
		t := neg[pickVariant(seed, len(neg))]
		esOut = fmt.Sprintf(t.esFmt, surfaceForm)
	default:
		fin := finiteTemplates(mood, tense)
		t := fin[pickVariant(seed, len(fin))]
		if startsWithReflexiveClitic(surfaceForm) {
			esFmt := strings.Replace(t.esFmt, "%s %s", "%s", 1)
			esOut = fmt.Sprintf(esFmt, surfaceForm)
		} else {
			esOut = formatFiniteES(t, subj, surfaceForm)
		}
	}

	// *haber* indicativo: forms (he/has/ha/…) are almost always auxiliary — do not glue random lexical tails.
	if strings.EqualFold(strings.TrimSpace(lemma), "haber") && mood == "indicativo" {
		esOut = fmt.Sprintf("%s %s.", subj, surfaceForm)
	}

	if !ExampleContainsSurface(esOut, surfaceForm) {
		if startsWithReflexiveClitic(surfaceForm) {
			esOut = surfaceForm + "."
		} else {
			esOut = fmt.Sprintf("%s %s.", subj, surfaceForm)
		}
	}

	var ruOut string
	if strings.TrimSpace(ruGloss) != "" {
		ruOut = CompactRussianVerbTrainingLine(person, number, ruGloss, mood, tense)
	}
	if ruOut == "" {
		ruOut = BuildRussianLiteraryLine(esOut, lemma, surfaceForm, mood, tense, ruGloss, seed^0x51ed_f00d)
	}
	return esOut, ruOut
}

// ExampleContainsSurface reports whether surface appears as a substring (case-insensitive) suitable for cloze masking.
func ExampleContainsSurface(es, surface string) bool {
	es = strings.TrimSpace(es)
	surface = strings.TrimSpace(surface)
	if es == "" || surface == "" {
		return false
	}
	return strings.Contains(strings.ToLower(es), strings.ToLower(surface))
}
