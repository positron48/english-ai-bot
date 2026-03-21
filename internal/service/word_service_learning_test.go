package service

import (
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestWordService_typoOrInvalidWordMessage_ENKeepsEnglishHint(t *testing.T) {
	logger := zap.NewNop()
	en := NewWordService(nil, nil, nil, nil, config.DefaultLearningConfig(), logger)
	msg := en.typoOrInvalidWordMessage()
	if !strings.Contains(msg, "английское") {
		t.Fatalf("expected Russian hint mentioning English for target=en, got %q", msg)
	}
}

func TestWordService_typoOrInvalidWordMessage_NonENGeneric(t *testing.T) {
	logger := zap.NewNop()
	lc := config.LearningConfig{
		Pair: "ru-es", NativeLang: "ru", TargetLang: "es",
		AppCode: "spanish", GrammarBundleID: "es",
	}
	if err := config.ValidateLearningConfig(lc); err != nil {
		t.Fatal(err)
	}
	svc := NewWordService(nil, nil, nil, nil, lc, logger)
	msg := svc.typoOrInvalidWordMessage()
	if strings.Contains(msg, "английское") {
		t.Fatalf("non-EN target should not mention English, got %q", msg)
	}
	if !strings.Contains(msg, "несуществующее") {
		t.Fatalf("expected generic typo hint, got %q", msg)
	}
}
