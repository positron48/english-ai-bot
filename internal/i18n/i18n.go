package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/text/language"
)

//go:embed locales
var localesFS embed.FS

var (
	translations  map[string]map[string]interface{}
	supportedTags = []language.Tag{
		language.English,
		language.Russian,
	}
	matcher = language.NewMatcher(supportedTags)
)

func init() {
	translations = make(map[string]map[string]interface{})

	// Load English translations
	enData, err := localesFS.ReadFile("locales/en.json")
	if err == nil {
		var enMap map[string]interface{}
		if err := json.Unmarshal(enData, &enMap); err == nil {
			translations["en"] = enMap
		}
	}

	// Load Russian translations
	ruData, err := localesFS.ReadFile("locales/ru.json")
	if err == nil {
		var ruMap map[string]interface{}
		if err := json.Unmarshal(ruData, &ruMap); err == nil {
			translations["ru"] = ruMap
		}
	}
}

// DetectLanguageFromRequest detects language from Accept-Language header
func DetectLanguageFromRequest(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang == "" {
		return "en"
	}

	// Parse Accept-Language header
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil || len(tags) == 0 {
		return "en"
	}

	// Match against supported languages
	matched, _, _ := matcher.Match(tags...)

	// Get base language
	base, _ := matched.Base()
	lang := base.String()

	// Map to supported languages
	if lang == "ru" {
		return "ru"
	}
	return "en"
}

// T translates a key for the given language
func T(lang string, key string) string {
	if lang == "" {
		lang = "en"
	}

	// Get translations for language
	langTranslations, ok := translations[lang]
	if !ok {
		// Fallback to English
		langTranslations, ok = translations["en"]
		if !ok {
			return key
		}
	}

	// Split key by dots (e.g., "errors.unauthorized")
	parts := strings.Split(key, ".")
	current := langTranslations

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - should be the value
			if val, ok := current[part]; ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
			break
		}

		// Navigate nested map
		if next, ok := current[part]; ok {
			if nextMap, ok := next.(map[string]interface{}); ok {
				current = nextMap
			} else {
				break
			}
		} else {
			break
		}
	}

	// Fallback to English if not found
	if lang != "en" {
		return T("en", key)
	}

	return key
}

// GetLanguageFromContext extracts language from context
func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(languageKey).(string); ok {
		return lang
	}
	return "en"
}

// Context key type
type contextKey string

const languageKey contextKey = "language"

// WithLanguage adds language to context
func WithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, languageKey, lang)
}
