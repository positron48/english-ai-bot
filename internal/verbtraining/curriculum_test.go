package verbtraining

import "testing"

func TestCurriculumCanonicalScopes(t *testing.T) {
	for _, pair := range [][2]string{{"es.preterito_indefinido.indicativo", "es.pretérito.indicativo"}, {"es.preterito_imperfecto.subjuntivo", "es.imperfecto.subjuntivo"}, {"es.preterito_perfecto_compuesto.indicativo", "es.pretérito perfecto.indicativo"}} {
		if CanonicalScope(pair[0]) != pair[1] {
			t.Fatal(pair)
		}
		found := false
		for _, scope := range ExpandScopes([]string{pair[1]}) {
			if scope == pair[0] {
				found = true
			}
		}
		if !found {
			t.Fatal(pair)
		}
	}
}
