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

func TestT(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		key      string
		want     string
		wantKey  bool // if true, expect key to be returned as-is (missing key)
	}{
		// empty lang defaults to en
		{name: "empty lang uses en", lang: "", key: "errors.unauthorized", want: "Unauthorized"},
		// English
		{name: "en nested key", lang: "en", key: "errors.unauthorized", want: "Unauthorized"},
		{name: "en messages key", lang: "en", key: "messages.notificationSettingsUpdated", want: "Notification settings updated successfully"},
		{name: "en missing key returns key", lang: "en", key: "missing.key", wantKey: true},
		{name: "en missing top-level returns key", lang: "en", key: "nonexistent", wantKey: true},
		// Russian
		{name: "ru nested key", lang: "ru", key: "errors.unauthorized", want: "Неавторизован"},
		{name: "ru messages key", lang: "ru", key: "messages.notificationSettingsUpdated", want: "Настройки уведомлений успешно обновлены"},
		{name: "ru missing key returns key", lang: "ru", key: "missing.key", wantKey: true},
		// unknown language fallback to en
		{name: "unknown lang fallback to en", lang: "de", key: "errors.unauthorized", want: "Unauthorized"},
		{name: "unknown lang missing key returns key", lang: "fr", key: "absent.key", wantKey: true},
		// key that is a nested object (not a string) returns key
		{name: "key is object not string returns key", lang: "en", key: "errors", wantKey: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := T(tt.lang, tt.key)
			if tt.wantKey {
				if got != tt.key {
					t.Errorf("T(%q, %q) = %q, want key %q", tt.lang, tt.key, got, tt.key)
				}
				return
			}
			if got != tt.want {
				t.Errorf("T(%q, %q) = %q, want %q", tt.lang, tt.key, got, tt.want)
			}
		})
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
