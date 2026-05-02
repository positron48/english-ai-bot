package verbtraining

import (
	"testing"
	"testing/fstest"
)

func TestLoadUnlockGatesAndEnabledScopes(t *testing.T) {
	fsys := fstest.MapFS{
		UnlockGatesPath: {
			Data: []byte(`{
  "version":"v1",
  "always_unlocked":["es.presente.indicativo"],
  "chapters":{
    "es.grammar.chapter.one":["es.futuro_simple.indicativo"],
    "es.grammar.chapter.two":["es.presente.subjuntivo"]
  }
}`),
		},
	}
	g, err := LoadUnlockGates(fsys)
	if err != nil {
		t.Fatalf("LoadUnlockGates: %v", err)
	}
	scopes := g.EnabledScopes(map[string]bool{"es.grammar.chapter.two": true})
	set := map[string]bool{}
	for _, s := range scopes {
		set[s] = true
	}
	if !set["es.presente.indicativo"] {
		t.Fatalf("always unlocked scope is missing")
	}
	if !set["es.presente.subjuntivo"] {
		t.Fatalf("chapter-unlocked scope is missing")
	}
	if set["es.futuro_simple.indicativo"] {
		t.Fatalf("unexpected scope from locked chapter")
	}
}

