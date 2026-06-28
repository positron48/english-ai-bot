package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSpanishGenderLexicon(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lexicon.tsv")
	content := "lemma\tgender\tarticle\topposite_gender_word\tsource\tnotes\n" +
		"Casa\tf\tla\tcaso\tunit\tfeminine noun\n" +
		"perro\tm\tel\tperra\tunit\tmasculine noun\n" +
		"\tm\tel\t\tunit\tskipped empty lemma\n" +
		"invalid\tbad\tel\t\tunit\tskipped invalid gender\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	lex, err := LoadSpanishGenderLexicon(path)
	if err != nil {
		t.Fatalf("LoadSpanishGenderLexicon: %v", err)
	}
	if len(lex) != 2 {
		t.Fatalf("entries: %d", len(lex))
	}
	casa, ok := lex["casa"]
	if !ok || casa.Gender != "f" || casa.Article != "la" || casa.OppositeGenderWord != "caso" {
		t.Fatalf("casa entry: %+v ok=%v", casa, ok)
	}
}

func TestLoadSpanishGenderLexicon_missingColumn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.tsv")
	if err := os.WriteFile(path, []byte("lemma\tgender\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadSpanishGenderLexicon(path)
	if err == nil || !strings.Contains(err.Error(), "missing required column") {
		t.Fatalf("got %v", err)
	}
}

func TestLoadSpanishGenderLexiconDefault_customPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.tsv")
	content := "lemma\tgender\tarticle\topposite_gender_word\tsource\tnotes\n" +
		"gato\tm\tel\tgata\tunit\tnote\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Setenv("SPANISH_GENDER_LEXICON_PATH", path)
	lex, loadedPath, err := LoadSpanishGenderLexiconDefault()
	if err != nil {
		t.Fatalf("LoadSpanishGenderLexiconDefault: %v", err)
	}
	if loadedPath != path {
		t.Fatalf("loaded path: got %q want %q", loadedPath, path)
	}
	if entry, ok := lex["gato"]; !ok || entry.Gender != "m" {
		t.Fatalf("gato entry: %+v ok=%v", entry, ok)
	}
}

func TestLoadSpanishGenderLexiconDefault_bundledFile(t *testing.T) {
	t.Setenv("SPANISH_GENDER_LEXICON_PATH", "")
	for _, p := range defaultSpanishGenderLexiconPaths() {
		if _, err := os.Stat(p); err != nil {
			t.Skipf("bundled lexicon %q not available from test cwd", p)
		}
	}
	lex, path, err := LoadSpanishGenderLexiconDefault()
	if err != nil {
		t.Fatalf("LoadSpanishGenderLexiconDefault: %v", err)
	}
	if path == "" || len(lex) == 0 {
		t.Fatalf("expected bundled lexicon, path=%q len=%d", path, len(lex))
	}
}

func TestDefaultSpanishGenderLexiconPaths(t *testing.T) {
	paths := defaultSpanishGenderLexiconPaths()
	if len(paths) < 2 {
		t.Fatalf("expected default paths, got %#v", paths)
	}
}
