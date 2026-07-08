package config

import (
	"strings"
	"testing"
)

func TestResolveAIProvider_OpenRouter(t *testing.T) {
	cfg := &AIConfig{
		Provider:    AIProviderOpenRouter,
		URL:         "https://openrouter.ai/api/v1",
		APIKey:      "or-key",
		Socks5Proxy: "51.254.98.124:1080",
	}
	if err := resolveAIProvider(cfg); err != nil {
		t.Fatalf("resolveAIProvider() error = %v", err)
	}
	if cfg.Provider != AIProviderOpenRouter {
		t.Fatalf("provider = %q, want %q", cfg.Provider, AIProviderOpenRouter)
	}
	if cfg.URL != "https://openrouter.ai/api/v1" {
		t.Fatalf("url = %q", cfg.URL)
	}
	if cfg.APIKey != "or-key" {
		t.Fatalf("api key = %q", cfg.APIKey)
	}
	if cfg.Socks5Proxy != "51.254.98.124:1080" {
		t.Fatalf("socks5 = %q", cfg.Socks5Proxy)
	}
}

func TestResolveAIProvider_Polza(t *testing.T) {
	cfg := &AIConfig{
		Provider:    AIProviderPolza,
		URL:         "https://openrouter.ai/api/v1",
		APIKey:      "openrouter-key",
		PolzaAPIKey: "polza-key",
		Socks5Proxy: "51.254.98.124:1080",
	}
	if err := resolveAIProvider(cfg); err != nil {
		t.Fatalf("resolveAIProvider() error = %v", err)
	}
	if cfg.Provider != AIProviderPolza {
		t.Fatalf("provider = %q, want %q", cfg.Provider, AIProviderPolza)
	}
	if cfg.URL != defaultPolzaURL {
		t.Fatalf("url = %q, want %q", cfg.URL, defaultPolzaURL)
	}
	if cfg.APIKey != "polza-key" {
		t.Fatalf("api key = %q, want polza-key", cfg.APIKey)
	}
	if cfg.Socks5Proxy != "" {
		t.Fatalf("socks5 = %q, want empty for polza text LLM", cfg.Socks5Proxy)
	}
}

func TestResolveAIProvider_PolzaCustomURL(t *testing.T) {
	cfg := &AIConfig{
		Provider:    AIProviderPolza,
		PolzaURL:    "https://polza.example/api/v1",
		PolzaAPIKey: "polza-key",
	}
	if err := resolveAIProvider(cfg); err != nil {
		t.Fatalf("resolveAIProvider() error = %v", err)
	}
	if cfg.URL != "https://polza.example/api/v1" {
		t.Fatalf("url = %q", cfg.URL)
	}
}

func TestResolveAIProvider_Errors(t *testing.T) {
	tests := []struct {
		name string
		cfg  AIConfig
		want string
	}{
		{
			name: "openrouter missing url",
			cfg:  AIConfig{Provider: AIProviderOpenRouter, APIKey: "k"},
			want: "AI_URL is required",
		},
		{
			name: "openrouter missing key",
			cfg:  AIConfig{Provider: AIProviderOpenRouter, URL: "https://openrouter.ai/api/v1"},
			want: "AI_API_KEY is required",
		},
		{
			name: "polza missing key",
			cfg:  AIConfig{Provider: AIProviderPolza},
			want: "POLZA_AI_API_KEY is required",
		},
		{
			name: "unknown provider",
			cfg:  AIConfig{Provider: "other", URL: "http://x", APIKey: "k"},
			want: "not supported",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			err := resolveAIProvider(&cfg)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}
