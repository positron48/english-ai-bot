package web

import (
	"testing"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestVerbFormsEnabledForUser_LinglowSpanishCourse(t *testing.T) {
	r := NewRouter(zap.NewNop(), &config.Config{
		Learning: config.DefaultLearningConfig(),
		Training: config.TrainingConfig{SpanishVerbFormsEnabled: true},
	}, nil, nil, nil, nil, nil)

	if r.verbFormsEnabled() {
		t.Fatal("server default en must not enable verb forms")
	}

	esLC := learningConfigForCourse(config.DefaultLearningConfig(), "es_ru")
	if esLC.TargetLang != "es" {
		t.Fatalf("es_ru target = %q", esLC.TargetLang)
	}
	if !r.verbFormsEnabledForLearning(esLC) {
		t.Fatal("expected verb forms enabled for es_ru learning config when flag is on")
	}
}

func TestVerbFormsEnabledForLearning_RespectsFeatureFlag(t *testing.T) {
	r := NewRouter(zap.NewNop(), &config.Config{
		Learning: config.LearningConfig{TargetLang: "es"},
		Training: config.TrainingConfig{SpanishVerbFormsEnabled: false},
	}, nil, nil, nil, nil, nil)

	if r.verbFormsEnabledForLearning(config.LearningConfig{TargetLang: "es"}) {
		t.Fatal("expected disabled when SpanishVerbFormsEnabled is false")
	}
}
