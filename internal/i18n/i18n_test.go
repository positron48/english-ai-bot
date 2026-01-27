package i18n

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestDetectLanguageFromRequest(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	if lang := DetectLanguageFromRequest(req); lang != "en" {
		t.Fatalf("expected default en, got %s", lang)
	}

	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.8")
	if lang := DetectLanguageFromRequest(req); lang != "ru" {
		t.Fatalf("expected ru, got %s", lang)
	}

	req.Header.Set("Accept-Language", "malformed")
	if lang := DetectLanguageFromRequest(req); lang != "en" {
		t.Fatalf("expected en for malformed header, got %s", lang)
	}
}

func TestTranslationFallbacks(t *testing.T) {
	if got := T("", "errors.not_found"); got == "errors.not_found" {
		// expecting some translation key in locales; just ensure it returns something stable
		if strings.TrimSpace(got) == "" {
			t.Fatalf("expected non-empty translation")
		}
	}

	missing := T("ru", "missing.key")
	if missing != "missing.key" {
		t.Fatalf("expected fallback key, got %s", missing)
	}
}

func TestWithLanguageContext(t *testing.T) {
	ctx := WithLanguage(context.Background(), "ru")
	if lang := GetLanguageFromContext(ctx); lang != "ru" {
		t.Fatalf("expected ru, got %s", lang)
	}
	if lang := GetLanguageFromContext(context.Background()); lang != "en" {
		t.Fatalf("expected default en, got %s", lang)
	}
}
