package spanishverbs

import (
	"fmt"
	"strings"
)

// ruLiteraryTemplates are book-style Russian lines (4 args: ruGloss, spanishSentence, surfaceForm, lemma).
var ruLiteraryTemplates = []string{
	`Литературно тот же узел смысла — **%s**: испанское «%s»; в центре форма «%s» (инф. *%s*).`,
	`По-русски это звучало бы как разговор о **%s**; испанский оригинал: «%s» — ключ «%s» (*%s*).`,
	`Передавая нюанс **%s**, можно сопоставить с «%s»; ударение на «%s» (*%s*).`,
	`Смысловой мост: **%s** — то поле значений, куда ложится «%s» (форма «%s», *%s*).`,
	`Если искать русский оттенок — **%s**; параллель в испанском: «%s», форма «%s» (*%s*).`,
	`Читательски ближе всего **%s**: та же сцена в «%s», с опорой на «%s» (*%s*).`,
	`В русской прозе то же часто обозначают через **%s**; здесь: «%s», ядро «%s» (*%s*).`,
	`Сопоставимо с русским **%s**; испанская ткань фразы — «%s», центр — «%s» (*%s*).`,
	`Тональность **%s** — ближайший русский якорь к «%s»; форма «%s» (*%s*).`,
	`По смыслу, не по буквам: **%s**; испанский вариант «%s», форма «%s» (*%s*).`,
	`Русская перекладина смысла — **%s**; она же слышится за «%s» (форма «%s», *%s*).`,
	`Как в учебнике перевода: **%s** ≈ то, что делает «%s»; маркер «%s» (*%s*).`,
}

// CompactRussianVerbTrainingLine is a short RU line aligned with catalog read-mode cards:
// subject + quoted finite verb (or analytic future) from ru_gloss; empty if gloss is missing.
func CompactRussianVerbTrainingLine(person, number, ruGloss, mood, tense string) string {
	if strings.TrimSpace(ruGloss) == "" {
		return ""
	}
	second := RussianCatalogSecondFromGloss(ruGloss, mood, tense, person, number)
	if strings.TrimSpace(second) == "" {
		return ""
	}
	return fmt.Sprintf("%s: «%s».", RussianSubjectCap(person, number), second)
}

// BuildRussianLiteraryLine returns a Russian line: literary gloss-based text when ruGloss is set, else grammar meta line.
func BuildRussianLiteraryLine(es, lemma, surface, mood, tense, ruGloss string, seed int64) string {
	g := strings.TrimSpace(ruGloss)
	if g == "" {
		return BuildRussianExampleLine(es, lemma, surface, mood, tense, seed)
	}
	lemma = strings.TrimSpace(strings.ToLower(lemma))
	surface = strings.TrimSpace(surface)
	i := pickVariant(seed, len(ruLiteraryTemplates))
	return fmt.Sprintf(ruLiteraryTemplates[i], g, es, surface, lemma)
}
