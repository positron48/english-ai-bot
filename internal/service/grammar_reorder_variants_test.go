package service

import "testing"

func TestCompareQuestionAnswerReorderVariants(t *testing.T) {
	s := &GrammarService{}
	q := map[string]interface{}{"type": "reorder", "correct_answer": "Vamos a enviarte los documentos hoy.", "accepted_answers": []interface{}{"Hoy vamos a enviarte los documentos.", "Vamos a enviarte hoy los documentos.", "", 17}}
	for _, answer := range []string{"Vamos a enviarte los documentos hoy.", "  hoy  vamos a enviarte los documentos. ", "Vamos a enviarte hoy los documentos."} {
		if !s.compareQuestionAnswer(answer, q) {
			t.Errorf("rejected valid variant %q", answer)
		}
	}
	for _, answer := range []string{"documentos hoy los enviarte a vamos.", "", "Hoy vamos a enviar los documentos."} {
		if s.compareQuestionAnswer(answer, q) {
			t.Errorf("accepted invalid answer %q", answer)
		}
	}
	q["type"] = "fill_blank"
	if s.compareQuestionAnswer("Hoy vamos a enviarte los documentos.", q) {
		t.Fatal("variants must only apply to reorder")
	}
}
