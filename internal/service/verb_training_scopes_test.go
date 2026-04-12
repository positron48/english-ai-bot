package service

import (
	"testing"

	"tgbot-skeleton/internal/config"
	"tgbot-skeleton/internal/models"
	"tgbot-skeleton/internal/repository"

	"go.uber.org/zap"
)

func TestIsSpanishVerbTrainingScope(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"es.presente.indicativo", true},
		{"es.pretérito.indicativo", true},
		{"es.grammar.past_preterito.foo", false},
		{"es.grammar.orientation_alphabet", false},
		{"en.presente.indicativo", false},
		{"es.presente", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isSpanishVerbTrainingScope(tt.in); got != tt.want {
			t.Errorf("isSpanishVerbTrainingScope(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestResolveVerbScopes_grammarBundleStageFallsBackToDefault(t *testing.T) {
	g := "es.grammar.past_preterito_perfecto.haber_plus_participio"
	settings := &models.UserSettings{GrammarStage: &g}
	learning := config.LearningConfig{TargetLang: "es"}
	got := ResolveVerbScopes(settings, learning)
	want := models.DefaultSpanishVerbScopes()
	if len(got) != len(want) {
		t.Fatalf("ResolveVerbScopes: got %v, want %v", got, want)
	}
	if len(want) > 0 && got[0] != want[0] {
		t.Fatalf("ResolveVerbScopes: got %v, want %v", got, want)
	}
}

func TestResolveVerbScopes_validGrammarStageUsed(t *testing.T) {
	g := "es.presente.indicativo"
	settings := &models.UserSettings{GrammarStage: &g}
	learning := config.LearningConfig{TargetLang: "es"}
	got := ResolveVerbScopes(settings, learning)
	if len(got) != 1 || got[0] != "es.presente.indicativo" {
		t.Fatalf("got %v", got)
	}
}

func TestEnsureVerbFormUserCards_DisabledIsNoOp(t *testing.T) {
	repo := &repository.VerbFormsRepository{}
	svc := NewVerbTrainingService(repo, config.LearningConfig{TargetLang: "en"}, config.TrainingConfig{SpanishVerbFormsEnabled: true}, zap.NewNop())
	if err := svc.EnsureVerbFormUserCards(1, []string{"es.presente.indicativo"}); err != nil {
		t.Fatal(err)
	}
}

func TestResolveVerbScopes_enabledVerbScopesWin(t *testing.T) {
	g := "es.grammar.foo"
	settings := &models.UserSettings{
		GrammarStage:        &g,
		EnabledVerbScopes: []string{"es.futuro.indicativo"},
	}
	learning := config.LearningConfig{TargetLang: "es"}
	got := ResolveVerbScopes(settings, learning)
	if len(got) != 1 || got[0] != "es.futuro.indicativo" {
		t.Fatalf("got %v", got)
	}
}

func TestResolveVerbScopes_enabledVerbScopesFiltersJunkKeepsValid(t *testing.T) {
	settings := &models.UserSettings{
		EnabledVerbScopes: []string{"es.grammar.bad", "es.presente.indicativo"},
	}
	got := ResolveVerbScopes(settings, config.LearningConfig{TargetLang: "es"})
	if len(got) != 1 || got[0] != "es.presente.indicativo" {
		t.Fatalf("got %v", got)
	}
}

func TestResolveVerbScopes_onlyInvalidEnabledScopesFallsBack(t *testing.T) {
	g := "es.grammar.past_preterito.foo"
	settings := &models.UserSettings{
		GrammarStage:      &g,
		EnabledVerbScopes: []string{"es.grammar.orientation", "es.grammar.past_preterito.foo"},
	}
	got := ResolveVerbScopes(settings, config.LearningConfig{TargetLang: "es"})
	want := models.DefaultSpanishVerbScopes()
	if len(got) != len(want) || (len(want) > 0 && got[0] != want[0]) {
		t.Fatalf("got %v want %v", got, want)
	}
}
