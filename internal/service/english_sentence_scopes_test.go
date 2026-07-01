package service

import (
	"reflect"
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"
)

func TestResolveEnglishSentenceScopes_NoProgressFallback(t *testing.T) {
	got := ResolveEnglishSentenceScopes(testEnglishSections(), nil, nil)
	want := DefaultEnglishSentenceScopes()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestResolveEnglishSentenceScopes_PassedPresentSimpleChapter(t *testing.T) {
	got := ResolveEnglishSentenceScopes(testEnglishSections(), nil, map[string]*repository.ChapterProgress{
		"en.grammar.first_actions_present_simple.present_simple_statements_i": {Passed: true, BestScore: 80},
	})
	if !containsString(got, "en.present_simple") {
		t.Fatalf("expected en.present_simple in %v", got)
	}
	if containsString(got, "en.past_simple") {
		t.Fatalf("did not expect en.past_simple in %v", got)
	}
}

func TestResolveEnglishSentenceScopes_PlacementA2UnlocksThroughA2(t *testing.T) {
	got := ResolveEnglishSentenceScopes(testEnglishSections(), &repository.PlacementTestResult{
		OpenedSections: []string{"en.grammar.now_present_continuous_contrast"},
	}, nil)
	for _, want := range []string{"en.be_present", "en.present_simple", "en.present_continuous", "en.past_simple", "en.future_will", "en.modals"} {
		if !containsString(got, want) {
			t.Fatalf("expected %s in %v", want, got)
		}
	}
	if containsString(got, "en.present_perfect") {
		t.Fatalf("did not expect B1 present perfect from A2 placement in %v", got)
	}
}

func TestResolveEnglishSentenceScopes_UnrelatedProgressDoesNotUnlockTense(t *testing.T) {
	got := ResolveEnglishSentenceScopes(testEnglishSections(), nil, map[string]*repository.ChapterProgress{
		"en.grammar.describing_precisely_adjectives_adverbs.adjective_order": {Passed: true, BestScore: 90},
	})
	want := DefaultEnglishSentenceScopes()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want fallback %v", got, want)
	}
}

func TestSentenceScopes_SpanishStillUsesEnabledVerbScopes(t *testing.T) {
	got := ResolveVerbScopes(&models.UserSettings{
		EnabledVerbScopes: []string{"es.presente.indicativo", "es.pretérito.indicativo"},
	}, config.LearningConfig{TargetLang: "es"})
	want := []string{"es.presente.indicativo", "es.pretérito.indicativo"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func testEnglishSections() *repository.SectionsData {
	return &repository.SectionsData{Sections: []repository.Section{
		{SectionID: "en.grammar.first_sentences_be_as", Level: "A1"},
		{SectionID: "en.grammar.first_actions_present_simple", Level: "A1"},
		{SectionID: "en.grammar.now_present_continuous_contrast", Level: "A2"},
		{SectionID: "en.grammar.past_1_past_simple", Level: "A2"},
		{SectionID: "en.grammar.past_2_background_process", Level: "A2"},
		{SectionID: "en.grammar.future_plans", Level: "A2"},
		{SectionID: "en.grammar.core_modality_ability_permission", Level: "A2"},
		{SectionID: "en.grammar.perfect_aspect_experience_result", Level: "B1"},
		{SectionID: "en.grammar.voice_focus_passive_and", Level: "B2"},
		{SectionID: "en.grammar.conditionals_hypotheticals", Level: "B2"},
	}}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
