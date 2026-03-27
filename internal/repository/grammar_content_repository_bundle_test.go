package repository

import (
	"path/filepath"
	"reflect"
	"testing"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

// Stable EN chapter id used to assert ES bundle routing does not accidentally serve the English course.
const sampleENChapterID = "en.grammar.first_sentences_be_as.personal_pronouns_am_is"

// First published ES chapter (replaces retired demo placeholder es.grammar.mvp_orientation.intro_placeholder).
const sampleESChapterID = "es.grammar.orientation_alphabet_sounds.spanish_alphabet_letter_names"

func TestNewGrammarContentRepositoryForLearning_DefaultENMatchesEmbeddedAlias(t *testing.T) {
	rFor, err := NewGrammarContentRepositoryForLearning(config.DefaultLearningConfig(), zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	rLegacy := NewGrammarContentRepository(zap.NewNop())
	i1, err1 := rFor.GetIndex()
	i2, err2 := rLegacy.GetIndex()
	if err1 != nil || err2 != nil {
		t.Fatalf("index err: %v %v", err1, err2)
	}
	if !reflect.DeepEqual(i1, i2) {
		t.Fatalf("index mismatch: DefaultLearningConfig forLearning vs NewGrammarContentRepository (legacy FS)")
	}
}

func TestNewGrammarContentRepositoryForLearning_ENExplicitAndCaseInsensitiveMatchLegacy(t *testing.T) {
	rLegacy := NewGrammarContentRepository(zap.NewNop())
	want, err := rLegacy.GetIndex()
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"en", "EN"} {
		id := id
		t.Run(id, func(t *testing.T) {
			lc := config.DefaultLearningConfig()
			lc.GrammarBundleID = id
			r, err := NewGrammarContentRepositoryForLearning(lc, zap.NewNop())
			if err != nil {
				t.Fatal(err)
			}
			got, err := r.GetIndex()
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("index mismatch for GrammarBundleID=%q", id)
			}
		})
	}
}

func TestNewGrammarContentRepositoryForLearning_ESMVPChapter(t *testing.T) {
	lc := config.DefaultLearningConfig()
	lc.GrammarBundleID = "es"
	r, err := NewGrammarContentRepositoryForLearning(lc, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	if r.ChapterExists(sampleENChapterID) {
		t.Fatalf("ES bundle must not expose EN chapter %q", sampleENChapterID)
	}
	ch, err := r.GetChapter(sampleESChapterID)
	if err != nil {
		t.Fatal(err)
	}
	if ch.TargetLanguage != "es" {
		t.Fatalf("target_language: %q", ch.TargetLanguage)
	}
	sec, err := r.GetSections()
	if err != nil {
		t.Fatal(err)
	}
	if len(sec.Sections) == 0 {
		t.Fatal("expected ES grammar sections")
	}
}

func TestNewGrammarContentRepositoryForLearning_FilesystemBundleDir(t *testing.T) {
	dir, err := filepath.Abs(filepath.Join("..", "grammarbundle", "en"))
	if err != nil {
		t.Fatal(err)
	}
	lc := config.LearningConfig{
		Pair:               "ru-en",
		NativeLang:         "ru",
		TargetLang:         "en",
		AppCode:            "english",
		GrammarBundleID:    "en",
		GrammarBundleDir:   dir,
	}
	r, err := NewGrammarContentRepositoryForLearning(lc, zap.NewNop())
	if err != nil {
		t.Fatal(err)
	}
	sec, err := r.GetSections()
	if err != nil {
		t.Fatal(err)
	}
	if len(sec.Sections) == 0 {
		t.Fatal("expected english sections from dir")
	}
}
