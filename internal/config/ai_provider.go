package config

import (
	"fmt"
	"strings"
)

const (
	AIProviderOpenRouter = "openrouter"
	AIProviderPolza      = "polza"

	defaultOpenRouterURL = "https://openrouter.ai/api/v1"
	defaultPolzaURL      = "https://polza.ai/api/v1"
)

// resolveAIProvider maps AI_PROVIDER to the effective text-LLM URL, API key, and SOCKS5
// settings. Speaking/TTS fallbacks must run before this so they keep the raw OpenRouter
// AI_API_KEY and OPENROUTER_SOCKS5_PROXY when AI_PROVIDER=polza.
func resolveAIProvider(cfg *AIConfig) error {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = AIProviderOpenRouter
	}

	switch provider {
	case AIProviderOpenRouter:
		if strings.TrimSpace(cfg.URL) == "" {
			return fmt.Errorf("AI_URL is required when AI_PROVIDER=openrouter")
		}
		if strings.TrimSpace(cfg.APIKey) == "" {
			return fmt.Errorf("AI_API_KEY is required when AI_PROVIDER=openrouter")
		}
	case AIProviderPolza:
		polzaURL := strings.TrimSpace(cfg.PolzaURL)
		if polzaURL == "" {
			polzaURL = defaultPolzaURL
		}
		cfg.URL = polzaURL

		polzaKey := strings.TrimSpace(cfg.PolzaAPIKey)
		if polzaKey == "" {
			return fmt.Errorf("POLZA_AI_API_KEY is required when AI_PROVIDER=polza")
		}
		cfg.APIKey = polzaKey
		cfg.Socks5Proxy = ""
	default:
		return fmt.Errorf("AI_PROVIDER %q is not supported (use openrouter or polza)", provider)
	}

	cfg.Provider = provider
	return nil
}
