package spanishverbs

import "strings"

// RussianInfinitiveIr is the dictionary headword for Spanish motion verb "ir" in RU examples.
const RussianInfinitiveIr = "идти"

// RussianIrVerbForSpanishIndicativo returns the finite Russian verb phrase (no subject) aligned with
// Spanish indicativo tense: present → несов. наст.; imperfecto → несов. прош.; pretérito → сов. прош. (пошёл);
// futuro → будущее (буду + инф.). Uses masculine defaults for past (шёл/пошёл) where gender varies.
// For non-indicativo mood, falls back to present.
func RussianIrVerbForSpanishIndicativo(mood, tense, person, number string) string {
	if strings.TrimSpace(strings.ToLower(mood)) != "indicativo" {
		return RussianIrPresent(person, number)
	}
	t := strings.TrimSpace(strings.ToLower(tense))
	switch t {
	case "presente":
		return RussianIrPresent(person, number)
	case "imperfecto":
		return RussianIrImperfectPast(person, number)
	case "pretérito", "preterito":
		return RussianIrPreteritePerfective(person, number)
	case "futuro":
		return RussianIrFutureComposite(person, number)
	default:
		return RussianIrPresent(person, number)
	}
}

// RussianIrPresent is imperfective present of идти (already used for legacy ir templates).
func RussianIrPresent(person, number string) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch p {
	case "1":
		if n == "plural" {
			return "идём"
		}
		return "иду"
	case "2":
		if n == "plural" {
			return "идёте"
		}
		return "идёшь"
	case "3":
		if n == "plural" {
			return "идут"
		}
		return "идёт"
	default:
		return "иду"
	}
}

// RussianIrImperfectPast is imperfective past of идти (идти → шёл…), masculine default.
func RussianIrImperfectPast(person, number string) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch p {
	case "1":
		if n == "plural" {
			return "шли"
		}
		return "шёл"
	case "2":
		if n == "plural" {
			return "шли"
		}
		return "шёл"
	case "3":
		if n == "plural" {
			return "шли"
		}
		return "шёл"
	default:
		return "шёл"
	}
}

// RussianIrPreteritePerfective is perfective past of пойти (пошёл…), matching Spanish pretérito as completed motion.
func RussianIrPreteritePerfective(person, number string) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	switch p {
	case "1":
		if n == "plural" {
			return "пошли"
		}
		return "пошёл"
	case "2":
		if n == "plural" {
			return "пошли"
		}
		return "пошёл"
	case "3":
		if n == "plural" {
			return "пошли"
		}
		return "пошёл"
	default:
		return "пошёл"
	}
}

// RussianIrFutureComposite is analytical future with imperfective infinitive идти.
func RussianIrFutureComposite(person, number string) string {
	p := strings.TrimSpace(strings.ToLower(person))
	n := strings.TrimSpace(strings.ToLower(number))
	aux := ""
	switch p {
	case "1":
		if n == "plural" {
			aux = "будем"
		} else {
			aux = "буду"
		}
	case "2":
		if n == "plural" {
			aux = "будете"
		} else {
			aux = "будешь"
		}
	case "3":
		if n == "plural" {
			aux = "будут"
		} else {
			aux = "будет"
		}
	default:
		aux = "буду"
	}
	return aux + " " + RussianInfinitiveIr
}
