package learning

import "testing"

func TestTargetLangNameRUAccusative(t *testing.T) {
	if got := TargetLangNameRUAccusative("en"); got != "английский" {
		t.Fatalf("en: got %q", got)
	}
	if got := TargetLangNameRUAccusative("es"); got != "испанский" {
		t.Fatalf("es: got %q", got)
	}
	if got := TargetLangNameRUAccusative(""); got != "английский" {
		t.Fatalf("empty defaults to en: got %q", got)
	}
}

func TestTargetLangNameRUPrepositional(t *testing.T) {
	if got := TargetLangNameRUPrepositional("en"); got != "английском" {
		t.Fatalf("en: got %q", got)
	}
	if got := TargetLangNameRUPrepositional(""); got != "английском" {
		t.Fatalf("empty defaults to en: got %q", got)
	}
}

func TestTargetLangNameEN(t *testing.T) {
	if got := TargetLangNameEN("en"); got != "English" {
		t.Fatalf("en: got %q", got)
	}
	if got := TargetLangNameEN(""); got != "English" {
		t.Fatalf("empty defaults to en: got %q", got)
	}
}
