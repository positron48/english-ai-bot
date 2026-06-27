package web

import (
	"context"
	"testing"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestGetUserVerbScopes_UsesUnlockGatesAlwaysUnlocked(t *testing.T) {
	r := NewRouter(zap.NewNop(), &config.Config{
		Learning: config.LearningConfig{
			TargetLang:      "es",
			GrammarBundleID: "es",
		},
		Training: config.TrainingConfig{
			SpanishVerbFormsEnabled: true,
		},
	}, nil, nil, nil, nil, nil)

	scopes := r.getUserVerbScopes(context.Background(), 0)
	if len(scopes) == 0 {
		t.Fatalf("expected non-empty scopes")
	}
	found := false
	for _, s := range scopes {
		if s == "es.presente.indicativo" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected always-unlocked scope es.presente.indicativo, got %v", scopes)
	}
}

