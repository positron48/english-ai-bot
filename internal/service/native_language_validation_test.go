package service

import (
	"testing"

	"tgbot-skeleton/internal/config"
)

func TestValidDefinitionNativeForLearning(t *testing.T) {
	ruEs := config.LearningConfig{NativeLang: "ru", TargetLang: "es"}
	enEs := config.LearningConfig{NativeLang: "en", TargetLang: "es"}

	if validDefinitionNativeForLearning("puerto en la costa", ruEs) {
		t.Fatal("expected Spanish definition to be rejected for ru-es")
	}

	if !validDefinitionNativeForLearning("порт или стоянка для яхт", ruEs) {
		t.Fatal("expected Russian definition to be accepted for ru-es")
	}

	// For non ru-es pairs this strict rule is not applied.
	if !validDefinitionNativeForLearning("harbor on the coast", enEs) {
		t.Fatal("expected definition to be accepted for non ru-es pair")
	}
}

