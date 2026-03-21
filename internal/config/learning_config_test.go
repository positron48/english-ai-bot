package config

import (
	"strings"
	"testing"
)

func TestValidateLearningConfig_OK(t *testing.T) {
	tests := []struct {
		name string
		lc   LearningConfig
	}{
		{
			name: "defaults style RU-EN",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "ru", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
		},
		{
			name: "native and target env-style casing normalized in validator",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "RU", TargetLang: "EN",
				AppCode: "english", GrammarBundleID: "en",
			},
		},
		{
			name: "trim and case normalization inputs still valid before Load normalizes",
			lc: LearningConfig{
				Pair: " RU-EN ", NativeLang: "ru", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
		},
		{
			name: "RU-ES profile",
			lc: LearningConfig{
				Pair: "ru-es", NativeLang: "ru", TargetLang: "es",
				AppCode: "spanish", GrammarBundleID: "es",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateLearningConfig(tt.lc); err != nil {
				t.Fatalf("ValidateLearningConfig: %v", err)
			}
		})
	}
}

func TestValidateLearningConfig_Errors(t *testing.T) {
	tests := []struct {
		name    string
		lc      LearningConfig
		contain string
	}{
		{
			name:    "empty pair",
			lc:      LearningConfig{NativeLang: "ru", TargetLang: "en", AppCode: "x", GrammarBundleID: "y"},
			contain: "LEARNING_PAIR",
		},
		{
			name: "double hyphen",
			lc: LearningConfig{
				Pair: "ru-en-us", NativeLang: "ru", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "exactly one hyphen",
		},
		{
			name: "pair mismatch",
			lc: LearningConfig{
				Pair: "ru-es", NativeLang: "ru", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "does not match",
		},
		{
			name: "lang code too short",
			lc: LearningConfig{
				Pair: "a-en", NativeLang: "a", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "language code",
		},
		{
			name: "empty app code",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "ru", TargetLang: "en",
				AppCode: "", GrammarBundleID: "en",
			},
			contain: "LEARNING_APP_CODE",
		},
		{
			name: "empty grammar bundle",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "ru", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "",
			},
			contain: "GRAMMAR_BUNDLE_ID",
		},
		{
			name: "empty native lang",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "LEARNING_NATIVE_LANG",
		},
		{
			name: "empty target lang",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "ru", TargetLang: "",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "LEARNING_TARGET_LANG",
		},
		{
			name: "pair without hyphen",
			lc: LearningConfig{
				Pair: "ruen", NativeLang: "ru", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "exactly one hyphen",
		},
		{
			name: "native lang code too short",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "r", TargetLang: "en",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "LEARNING_NATIVE_LANG",
		},
		{
			name: "target lang code too short",
			lc: LearningConfig{
				Pair: "ru-en", NativeLang: "ru", TargetLang: "e",
				AppCode: "english", GrammarBundleID: "en",
			},
			contain: "LEARNING_TARGET_LANG",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLearningConfig(tt.lc)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.contain) {
				t.Fatalf("error %q should contain %q", err.Error(), tt.contain)
			}
		})
	}
}

func TestDefaultLearningConfig(t *testing.T) {
	d := DefaultLearningConfig()
	if err := ValidateLearningConfig(d); err != nil {
		t.Fatalf("DefaultLearningConfig should validate: %v", err)
	}
	if d.Pair != "ru-en" || d.NativeLang != "ru" || d.TargetLang != "en" || d.AppCode != "english" {
		t.Fatalf("unexpected defaults: %+v", d)
	}
	if d.GrammarBundleID != "en" {
		t.Fatalf("GrammarBundleID: got %q want en", d.GrammarBundleID)
	}
}
