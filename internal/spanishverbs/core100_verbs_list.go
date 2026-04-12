package spanishverbs

import "strings"

// Freq100VerbLemmas — первые 100 лемм с pos VERB или AUX в порядке файла
// resources/wordsets/spanish_word_freq_pos_ud_top6000.csv (частотный корпус).
// Используется для детерминированных cloze-шаблонов без LLM в рантайме.
var Freq100VerbLemmas = []string{
	"ser", "haber", "tener", "estar", "poder", "hacer", "decir", "dar", "ir", "deber",
	"encontrar", "ver", "llegar", "llevar", "querer", "seguir", "poner", "pasar", "dejar", "ganar",
	"conocer", "quedar", "mantener", "saber", "presentar", "considerar", "volver", "realizar", "asegurar", "permitir",
	"jugar", "contar", "recibir", "pedir", "conseguir", "producir", "convertir", "tratar", "explicar", "lograr",
	"comenzar", "incluir", "parecer", "creer", "perder", "decidir", "vivir", "afirmar", "salir", "llamar",
	"mostrar", "utilizar", "existir", "señalar", "trabajar", "tomar", "hablar", "destacar", "anunciar", "esperar",
	"formar", "crear", "partir", "intentar", "participar", "situar", "recordar", "añadir", "alcanzar", "ofrecer",
	"acabar", "buscar", "ubicar", "nacer", "reconocer", "pensar", "indicar", "empezar", "venir", "informar",
	"suponer", "entrar", "aparecer", "servir", "obtener", "dirigir", "celebrar", "cambiar", "declarar", "sufrir",
	"evitar", "abrir", "morir", "cumplir", "representar", "continuar", "resultar", "establecer", "dedicar", "iniciar",
}

var freq100VerbLemmaSet map[string]struct{}

func init() {
	freq100VerbLemmaSet = make(map[string]struct{}, len(Freq100VerbLemmas))
	for _, lem := range Freq100VerbLemmas {
		freq100VerbLemmaSet[lem] = struct{}{}
	}
}

// IsFreq100VerbLemma reports whether lemma (lowercase) is in the frequency top-100 verb/aux list.
func IsFreq100VerbLemma(lemma string) bool {
	k := strings.ToLower(strings.TrimSpace(lemma))
	_, ok := freq100VerbLemmaSet[k]
	return ok
}
