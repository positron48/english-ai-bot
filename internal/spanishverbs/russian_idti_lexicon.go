package spanishverbs

// IrMotionSpanishSuffixToRuTail maps Spanish complement (EsSuffix in templates) to a short Russian goal phrase
// used after the conjugated motion verb (идти / шёл / пошёл / буду идти). Same order as buildIrCatalogTemplates.
// Spanish "ir" ↔ Russian headword: RussianInfinitiveIr ("идти"); completed motion in pretérito uses perfective past (пошёл).
var IrMotionSpanishSuffixToRuTail = map[string]string{
	"a casa.":         "домой.",
	"al trabajo.":     "на работу.",
	"a la escuela.":   "в школу.",
	"al mercado.":     "на рынок.",
	"al parque.":      "в парк.",
	"a pie.":          "пешком.",
	"de compras.":     "за покупками.",
	"de viaje.":       "в поездку.",
}
