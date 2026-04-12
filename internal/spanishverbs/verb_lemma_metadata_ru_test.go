package spanishverbs

import (
	"strings"
	"testing"
)

func TestRuGlossFromLemmaMetadataJSON_nested(t *testing.T) {
	raw := `{"ru":{"gloss":"бросать","gloss_source":"x"}}`
	if g := RuGlossFromLemmaMetadataJSON(raw); g != "бросать" {
		t.Fatalf("got %q", g)
	}
}

func TestRuGlossFromLemmaMetadataJSON_flat(t *testing.T) {
	raw := `{"ru_gloss":"идти"}`
	if g := RuGlossFromLemmaMetadataJSON(raw); g != "идти" {
		t.Fatalf("got %q", g)
	}
}

func TestMergeRuGlossIntoLemmaMetadataJSON(t *testing.T) {
	got, err := MergeRuGlossIntoLemmaMetadataJSON(`{"x":1}`, "бросать", "llm-test")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "бросать") || !strings.Contains(got, "llm-test") {
		t.Fatalf("%s", got)
	}
}

func TestBuildRussianLiteraryLine_withGloss(t *testing.T) {
	ru := BuildRussianLiteraryLine("A veces tú hablas.", "hablar", "hablas", "indicativo", "presente", "говорить", 11)
	if ru == "" || !strings.Contains(ru, "говорить") {
		t.Fatalf("ru: %q", ru)
	}
}

func TestVerbClassFromLemmaMetadataJSON(t *testing.T) {
	raw := `{"verb_class":"motion","ru":{"gloss":"идти"}}`
	if g := VerbClassFromLemmaMetadataJSON(raw); g != "motion" {
		t.Fatalf("got %q", g)
	}
}

func TestAllowedTemplateIDsFromLemmaMetadataJSON(t *testing.T) {
	raw := `{"allowed_template_ids":["ir_a_casa","ir_al_parque"]}`
	ids := AllowedTemplateIDsFromLemmaMetadataJSON(raw)
	if len(ids) != 2 || ids[0] != "ir_a_casa" {
		t.Fatalf("got %#v", ids)
	}
}

func TestMergeVerbClassIntoLemmaMetadataJSON(t *testing.T) {
	got, err := MergeVerbClassIntoLemmaMetadataJSON(`{"ru":{"gloss":"идти"}}`, "motion", "rule-v1")
	if err != nil {
		t.Fatal(err)
	}
	if VerbClassFromLemmaMetadataJSON(got) != "motion" {
		t.Fatalf("%s", got)
	}
	if RuGlossFromLemmaMetadataJSON(got) != "идти" {
		t.Fatalf("ru gloss lost: %s", got)
	}
}

func TestMergeAllowedTemplateIDsIntoLemmaMetadataJSON(t *testing.T) {
	got, err := MergeAllowedTemplateIDsIntoLemmaMetadataJSON(`{"verb_class":"motion"}`, []string{"ir_a_casa", "ir_a_casa"}, "rule-v1")
	if err != nil {
		t.Fatal(err)
	}
	ids := AllowedTemplateIDsFromLemmaMetadataJSON(got)
	if len(ids) != 1 || ids[0] != "ir_a_casa" {
		t.Fatalf("got %#v from %s", ids, got)
	}
}
