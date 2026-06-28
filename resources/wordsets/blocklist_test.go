package wordsets

import "testing"

func TestNormalizeLemmaImport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in, want string
	}{
		{"  Hello  ", "hello"},
		{"CAFÉ", "café"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := NormalizeLemmaImport(tc.in); got != tc.want {
			t.Fatalf("NormalizeLemmaImport(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsLemmaBlockedForLang(t *testing.T) {
	t.Parallel()
	if IsLemmaBlockedForLang("en", "zzz-nonexistent-blocklist-entry-xyz") {
		t.Fatal("unexpected english block")
	}
	if IsLemmaBlockedForLang("es", "zzz-nonexistent-blocklist-entry-xyz") {
		t.Fatal("unexpected spanish block")
	}
	if IsLemmaBlocked("") {
		t.Fatal("empty lemma should not be blocked")
	}
}

func TestIsVowellessAbbrevASCII(t *testing.T) {
	t.Parallel()
	tests := []struct {
		lemma string
		want  bool
	}{
		{"cm", true},
		{"mm", true},
		{"abc", false},
		{"a", false},
		{"toolong", false},
		{"c1", false},
		{"bcdfg", true},
	}
	for _, tc := range tests {
		if got := IsVowellessAbbrevASCII(tc.lemma); got != tc.want {
			t.Fatalf("IsVowellessAbbrevASCII(%q) = %v, want %v", tc.lemma, got, tc.want)
		}
	}
}

func TestBlocklistsLoaded(t *testing.T) {
	t.Parallel()
	if len(blockedSpanishLemmas) == 0 {
		t.Fatal("spanish blocklist empty")
	}
	// english_lemma_blocklist.txt may contain only comments (intentionally minimal).
	_ = blockedEnglishLemmas
}
