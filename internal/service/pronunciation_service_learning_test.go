package service

import (
	"strings"
	"testing"

	"tgbot-skeleton/internal/config"

	"go.uber.org/zap"
)

func TestBuildPronunciationProviders_DictionaryBaseURLRespectsTargetLang(t *testing.T) {
	logger := zap.NewNop()
	baseCfg := config.TTSConfig{
		Provider:          "dictionary",
		DictionaryEnabled: true,
	}

	t.Run("default URL uses TargetLang en", func(t *testing.T) {
		lc := config.LearningConfig{
			Pair: "ru-en", NativeLang: "ru", TargetLang: "en",
			AppCode: "english", GrammarBundleID: "en",
		}
		provs := buildPronunciationProviders(baseCfg, lc, logger)
		if len(provs) != 1 {
			t.Fatalf("got %d providers, want 1", len(provs))
		}
		dp, ok := provs[0].(*dictionaryPronunciationProvider)
		if !ok {
			t.Fatalf("provider type %T", provs[0])
		}
		want := "https://api.dictionaryapi.dev/api/v2/entries/en"
		if dp.baseURL != want {
			t.Fatalf("baseURL=%q want %q", dp.baseURL, want)
		}
	})

	t.Run("default URL uses TargetLang es segment", func(t *testing.T) {
		lc := config.LearningConfig{
			Pair: "ru-es", NativeLang: "ru", TargetLang: "es",
			AppCode: "spanish", GrammarBundleID: "es",
		}
		provs := buildPronunciationProviders(baseCfg, lc, logger)
		dp := provs[0].(*dictionaryPronunciationProvider)
		want := "https://api.dictionaryapi.dev/api/v2/entries/es"
		if dp.baseURL != want {
			t.Fatalf("baseURL=%q want %q", dp.baseURL, want)
		}
	})

	t.Run("empty TargetLang falls back to en segment", func(t *testing.T) {
		lc := config.LearningConfig{TargetLang: ""}
		provs := buildPronunciationProviders(baseCfg, lc, logger)
		dp := provs[0].(*dictionaryPronunciationProvider)
		if !strings.HasSuffix(dp.baseURL, "/en") {
			t.Fatalf("baseURL=%q should end with /en", dp.baseURL)
		}
	})

	t.Run("explicit DictionaryBaseURL overrides TargetLang", func(t *testing.T) {
		cfg := baseCfg
		cfg.DictionaryBaseURL = "https://example.com/dict/api"
		lc := config.LearningConfig{
			Pair: "ru-es", NativeLang: "ru", TargetLang: "es",
			AppCode: "spanish", GrammarBundleID: "es",
		}
		provs := buildPronunciationProviders(cfg, lc, logger)
		dp := provs[0].(*dictionaryPronunciationProvider)
		if dp.baseURL != "https://example.com/dict/api" {
			t.Fatalf("baseURL=%q", dp.baseURL)
		}
	})
}

func TestBuildOpenRouterChatUserPrompt(t *testing.T) {
	t.Run("empty uses default", func(t *testing.T) {
		got := buildOpenRouterChatUserPrompt("", "hello")
		if !strings.Contains(got, "`hello`") || !strings.Contains(got, "pronunciation machine") {
			t.Fatalf("unexpected: %q", got)
		}
	})
	t.Run("{word} replacement", func(t *testing.T) {
		got := buildOpenRouterChatUserPrompt("Say in Spanish: {word}", "casa")
		want := "Say in Spanish: `casa`"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
	t.Run("no placeholder appends quoted word", func(t *testing.T) {
		got := buildOpenRouterChatUserPrompt("Spanish only.", "sol")
		want := "Spanish only. `sol`"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}
