package wordsetimport

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMustHaveBlueprint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "must-have.yaml")
	content := []byte(`
must_have:
  title: "Must Have — Espanol"
  subcategories:
    - id: basics
      title: "Base"
      sets:
        - id: greetings
          title: "Greetings"
          words:
            - { es: "Hola", ru: "привет" }
            - { es: "hola", ru: "привет" }
            - { es: " adiós ", ru: "пока" }
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	root, err := loadMustHaveBlueprint(path)
	if err != nil {
		t.Fatalf("load must-have blueprint: %v", err)
	}
	if root.Title != "Must Have — Espanol" {
		t.Fatalf("unexpected root title: %q", root.Title)
	}
	if len(root.Subcategories) != 1 || len(root.Subcategories[0].Sets) != 1 {
		t.Fatalf("unexpected structure: %+v", root.Subcategories)
	}

	lemmas := normalizeMustHaveWords(root.Subcategories[0].Sets[0].Words, "es")
	if len(lemmas) != 2 {
		t.Fatalf("expected 2 unique lemmas, got %d (%v)", len(lemmas), lemmas)
	}
	if lemmas[0] != "hola" || lemmas[1] != "adiós" {
		t.Fatalf("unexpected lemmas order/content: %v", lemmas)
	}
}

func TestLoadMustHaveBlueprint_EnglishWords(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "must-have-en.yaml")
	content := []byte(`
must_have:
  title: "Must Have — English"
  subcategories:
    - id: basics
      title: "Base"
      sets:
        - id: greetings
          title: "Greetings"
          words:
            - { en: "Hello", ru: "привет" }
            - { en: "hello", ru: "привет" }
            - { en: "can't", ru: "не могу" }
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	root, err := loadMustHaveBlueprint(path)
	if err != nil {
		t.Fatalf("load must-have blueprint: %v", err)
	}
	lemmas := normalizeMustHaveWords(root.Subcategories[0].Sets[0].Words, "en")
	if len(lemmas) != 2 {
		t.Fatalf("expected 2 unique lemmas, got %d (%v)", len(lemmas), lemmas)
	}
	if lemmas[0] != "hello" || lemmas[1] != "can't" {
		t.Fatalf("unexpected lemmas order/content: %v", lemmas)
	}
}

func TestLoadMustHaveBlueprint_RequiresTitle(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "must-have.yaml")
	content := []byte(`
must_have:
  subcategories: []
`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	if _, err := loadMustHaveBlueprint(path); err == nil {
		t.Fatal("expected error for empty must_have.title")
	}
}

