package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestLoadDailyJourneyDefaultAndRollback(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		enabled     bool
	}{{"default", "", true}, {"rollback", "false", false}, {"explicit", "true", true}} {
		t.Run(tc.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)
			t.Setenv("AI_PROVIDER", "openrouter")
			t.Setenv("AI_URL", "https://test-ai.invalid/v1")
			t.Setenv("AI_API_KEY", "test-only")
			t.Setenv("AI_PROMPT", "Test prompt")
			t.Setenv("AI_PROMPT_FILE", "")
			t.Setenv("WEBAPP_JWT_SECRET", "test-only")
			t.Setenv("DATABASE_DRIVER", "postgres")
			t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
			t.Setenv("LINGLOW_DAILY_JOURNEY_ENABLED", tc.value)
			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Linglow.DailyJourneyEnabled != tc.enabled {
				t.Fatalf("enabled=%v want=%v", cfg.Linglow.DailyJourneyEnabled, tc.enabled)
			}
		})
	}
}
