package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func backupLearningEnv() map[string]string {
	keys := []string{
		"LEARNING_PAIR",
		"LEARNING_NATIVE_LANG",
		"LEARNING_TARGET_LANG",
		"LEARNING_APP_CODE",
		"GRAMMAR_BUNDLE_ID",
		"GRAMMAR_BUNDLE_DIR",
	}
	m := make(map[string]string, len(keys))
	for _, k := range keys {
		m[k] = os.Getenv(k)
	}
	return m
}

func restoreLearningEnv(m map[string]string) {
	for k, v := range m {
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
}

func unsetLearningEnvForDefaults() {
	for _, k := range []string{
		"LEARNING_PAIR",
		"LEARNING_NATIVE_LANG",
		"LEARNING_TARGET_LANG",
		"LEARNING_APP_CODE",
		"GRAMMAR_BUNDLE_ID",
		"GRAMMAR_BUNDLE_DIR",
	} {
		os.Unsetenv(k)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Save original env vars - include all that might affect defaults
	originalEnv := map[string]string{
		"AI_URL":                     os.Getenv("AI_URL"),
		"AI_API_KEY":                 os.Getenv("AI_API_KEY"),
		"AI_PROMPT":                  os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":             os.Getenv("AI_PROMPT_FILE"),
		"AI_MODEL":                   os.Getenv("AI_MODEL"),
		"WEBAPP_JWT_SECRET":          os.Getenv("WEBAPP_JWT_SECRET"),
		"TELEGRAM_TOKEN":             os.Getenv("TELEGRAM_TOKEN"),
		"TRAINING_WORKER_BATCH_SIZE": os.Getenv("TRAINING_WORKER_BATCH_SIZE"),
		"TRAINING_LLM_WORKERS":       os.Getenv("TRAINING_LLM_WORKERS"),
		"SERVER_ADDRESS":             os.Getenv("SERVER_ADDRESS"),
		"LOG_LEVEL":                  os.Getenv("LOG_LEVEL"),
		"DATABASE_PATH":              os.Getenv("DATABASE_PATH"),
		"DATABASE_DRIVER":            os.Getenv("DATABASE_DRIVER"),
		"DATABASE_URL":               os.Getenv("DATABASE_URL"),
	}

	learningEnv := backupLearningEnv()
	// Restore original env vars after test
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Set required env vars and clear optional ones to get defaults
	unsetLearningEnvForDefaults()
	os.Setenv("AI_URL", "http://test-ai.local")
	os.Setenv("AI_API_KEY", "test-api-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Unsetenv("AI_MODEL")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	os.Unsetenv("TRAINING_WORKER_BATCH_SIZE")
	os.Unsetenv("TRAINING_LLM_WORKERS")
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DATABASE_PATH")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check defaults
	if cfg.Server.Address != ":8184" {
		t.Errorf("Expected default server address :8184, got %s", cfg.Server.Address)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default logging level info, got %s", cfg.Logging.Level)
	}

	if cfg.AI.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected default AI model gpt-3.5-turbo, got %s", cfg.AI.Model)
	}

	if cfg.Database.Driver != "postgres" {
		t.Errorf("Expected database driver postgres, got %s", cfg.Database.Driver)
	}

	if !cfg.Training.WorkerEnabled {
		t.Error("Expected training worker enabled by default")
	}

	if cfg.Training.WorkerInterval != "30s" {
		t.Errorf("Expected default worker interval 30s, got %s", cfg.Training.WorkerInterval)
	}

	if cfg.Training.WorkerBatchSize != 5 {
		t.Errorf("Expected default worker batch size 5, got %d", cfg.Training.WorkerBatchSize)
	}

	if cfg.Training.LLMWorkers != 4 {
		t.Errorf("Expected default LLM workers 4, got %d", cfg.Training.LLMWorkers)
	}

	if cfg.WebApp.OTPTTLSeconds != 300 {
		t.Errorf("Expected default OTP TTL 300, got %d", cfg.WebApp.OTPTTLSeconds)
	}

	if cfg.WebApp.SessionTTLHours != 720 {
		t.Errorf("Expected default session TTL 720, got %d", cfg.WebApp.SessionTTLHours)
	}

	if cfg.WebApp.JWTTTLHours != 24 {
		t.Errorf("Expected default JWT TTL 24, got %d", cfg.WebApp.JWTTTLHours)
	}

	if cfg.Learning.Pair != "ru-en" || cfg.Learning.NativeLang != "ru" || cfg.Learning.TargetLang != "en" {
		t.Errorf("Expected default learning ru→en, got pair=%q native=%q target=%q", cfg.Learning.Pair, cfg.Learning.NativeLang, cfg.Learning.TargetLang)
	}
	if cfg.Learning.AppCode != "english" || cfg.Learning.GrammarBundleID != "en" {
		t.Errorf("Expected default app_code=english grammar_bundle_id=en, got app_code=%q grammar_bundle_id=%q", cfg.Learning.AppCode, cfg.Learning.GrammarBundleID)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"DATABASE_URL":      os.Getenv("DATABASE_URL"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("DATABASE_URL")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	unsetLearningEnvForDefaults()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing DATABASE_URL")
	}
	if err != nil && !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("Expected error to mention DATABASE_URL, got %v", err)
	}
}

func TestLoad_MissingAIURL(t *testing.T) {
	// Clear relevant env vars
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Unsetenv("AI_URL")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing AI_URL")
	}
}

func TestLoad_MissingAIAPIKey(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Unsetenv("AI_API_KEY")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing AI_API_KEY")
	}
}

func TestLoad_MissingAIPrompt(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Unsetenv("AI_PROMPT")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing AI_PROMPT")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":                os.Getenv("AI_URL"),
		"AI_API_KEY":            os.Getenv("AI_API_KEY"),
		"AI_PROMPT":             os.Getenv("AI_PROMPT"),
		"WEBAPP_JWT_SECRET":     os.Getenv("WEBAPP_JWT_SECRET"),
		"WEBAPP_SESSION_SECRET": os.Getenv("WEBAPP_SESSION_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("WEBAPP_JWT_SECRET")
	os.Unsetenv("WEBAPP_SESSION_SECRET")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for missing JWT secret")
	}
}

func TestLoad_WithPromptFile(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Create temp prompt file
	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	promptContent := "This is a test prompt from file"
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create prompt file: %v", err)
	}

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Unsetenv("AI_PROMPT")
	os.Setenv("AI_PROMPT_FILE", promptFile)
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AI.Prompt != promptContent {
		t.Errorf("Expected prompt from file, got %s", cfg.AI.Prompt)
	}
}

func TestLoad_WithPromptFile_TemplateSubstitutions(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	tmpDir := t.TempDir()
	promptFile := filepath.Join(tmpDir, "prompt.txt")
	promptContent := "pair={{pair}} native={{native_lang}} target={{target_lang}}"
	if err := os.WriteFile(promptFile, []byte(promptContent), 0644); err != nil {
		t.Fatalf("Failed to create prompt file: %v", err)
	}

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Unsetenv("AI_PROMPT")
	os.Setenv("AI_PROMPT_FILE", promptFile)
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := "pair=ru-en native=ru target=en"
	if cfg.AI.Prompt != want {
		t.Errorf("Expected rendered prompt %q, got %q", want, cfg.AI.Prompt)
	}
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	// Create a temp dir with invalid YAML config so ReadInConfig fails with non-ConfigFileNotFound error
	tmpDir := t.TempDir()
	invalidYAML := "invalid: [unclosed bracket"
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"DATABASE_URL":      os.Getenv("DATABASE_URL"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	unsetLearningEnvForDefaults()

	_, err = Load()
	if err == nil {
		t.Error("Expected error when config file has invalid YAML")
	}
	if err != nil && !strings.Contains(err.Error(), "config file") {
		t.Errorf("Expected error to mention config file, got %v", err)
	}
}

func TestLoad_InvalidPromptFile(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Unsetenv("AI_PROMPT")
	os.Setenv("AI_PROMPT_FILE", "/nonexistent/path/prompt.txt")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	_, err := Load()
	if err == nil {
		t.Error("Expected error for invalid prompt file")
	}
}

func TestLoad_UnmarshalError(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":                     os.Getenv("AI_URL"),
		"AI_API_KEY":                 os.Getenv("AI_API_KEY"),
		"AI_PROMPT":                  os.Getenv("AI_PROMPT"),
		"DATABASE_URL":               os.Getenv("DATABASE_URL"),
		"WEBAPP_JWT_SECRET":          os.Getenv("WEBAPP_JWT_SECRET"),
		"TRAINING_WORKER_BATCH_SIZE": os.Getenv("TRAINING_WORKER_BATCH_SIZE"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	// Invalid value for int field causes Unmarshal to fail
	os.Setenv("TRAINING_WORKER_BATCH_SIZE", "not_a_number")
	unsetLearningEnvForDefaults()

	_, err := Load()
	if err == nil {
		t.Error("Expected error when env value cannot be unmarshaled into config")
	}
	if err != nil && !strings.Contains(err.Error(), "unmarshaling") {
		t.Errorf("Expected error to mention unmarshaling, got %v", err)
	}
}

func TestLoad_BotMessageNewlines(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
		"BOT_START_MESSAGE": os.Getenv("BOT_START_MESSAGE"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	os.Setenv("BOT_START_MESSAGE", "Hello\\nWorld")
	unsetLearningEnvForDefaults()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Bot.StartMessage != "Hello\nWorld" {
		t.Errorf("Expected newlines to be processed, got %q", cfg.Bot.StartMessage)
	}
}

func TestLoad_SessionSecretFallback(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":                os.Getenv("AI_URL"),
		"AI_API_KEY":            os.Getenv("AI_API_KEY"),
		"AI_PROMPT":             os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":        os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET":     os.Getenv("WEBAPP_JWT_SECRET"),
		"WEBAPP_SESSION_SECRET": os.Getenv("WEBAPP_SESSION_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Unsetenv("WEBAPP_JWT_SECRET")
	os.Setenv("WEBAPP_SESSION_SECRET", "session-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, should allow session secret fallback", err)
	}

	if cfg.WebApp.SessionSecret != "session-secret" {
		t.Errorf("Expected session secret, got %s", cfg.WebApp.SessionSecret)
	}
}

func TestLoad_CustomEnvValues(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"AI_MODEL":          os.Getenv("AI_MODEL"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
		"SERVER_ADDRESS":    os.Getenv("SERVER_ADDRESS"),
		"LOG_LEVEL":         os.Getenv("LOG_LEVEL"),
		"DATABASE_PATH":     os.Getenv("DATABASE_PATH"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://custom-ai.local")
	os.Setenv("AI_API_KEY", "custom-api-key")
	os.Setenv("AI_PROMPT", "custom prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("AI_MODEL", "gpt-4")
	os.Setenv("WEBAPP_JWT_SECRET", "custom-jwt-secret")
	os.Setenv("SERVER_ADDRESS", ":9000")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://custom:custom@localhost:5432/customdb?sslmode=disable")
	unsetLearningEnvForDefaults()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AI.URL != "http://custom-ai.local" {
		t.Errorf("Expected custom AI URL, got %s", cfg.AI.URL)
	}

	if cfg.AI.APIKey != "custom-api-key" {
		t.Errorf("Expected custom API key, got %s", cfg.AI.APIKey)
	}

	if cfg.AI.Model != "gpt-4" {
		t.Errorf("Expected custom AI model, got %s", cfg.AI.Model)
	}

	if cfg.Server.Address != ":9000" {
		t.Errorf("Expected custom server address, got %s", cfg.Server.Address)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected debug log level, got %s", cfg.Logging.Level)
	}

	if cfg.Database.URL != "postgres://custom:custom@localhost:5432/customdb?sslmode=disable" {
		t.Errorf("Expected custom database URL, got %s", cfg.Database.URL)
	}
}

func TestLoad_RateLimitDefaults(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
	}

	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.WebApp.RateLimitAuthRequestOTPPerIP != 10 {
		t.Errorf("Expected rate limit 10, got %d", cfg.WebApp.RateLimitAuthRequestOTPPerIP)
	}

	if cfg.WebApp.RateLimitAppAPIPerUser != 300 {
		t.Errorf("Expected rate limit 300, got %d", cfg.WebApp.RateLimitAppAPIPerUser)
	}

	if cfg.WebApp.RateLimitWindowMinutes != 1 {
		t.Errorf("Expected window 1, got %d", cfg.WebApp.RateLimitWindowMinutes)
	}

	if cfg.WebApp.RateLimitBurstMultiplier != 2 {
		t.Errorf("Expected burst multiplier 2, got %d", cfg.WebApp.RateLimitBurstMultiplier)
	}
}

func TestLoad_InvalidLearningPair(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":               os.Getenv("AI_URL"),
		"AI_API_KEY":           os.Getenv("AI_API_KEY"),
		"AI_PROMPT":            os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":       os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET":    os.Getenv("WEBAPP_JWT_SECRET"),
		"DATABASE_URL":         os.Getenv("DATABASE_URL"),
		"LEARNING_PAIR":        os.Getenv("LEARNING_PAIR"),
		"LEARNING_NATIVE_LANG": os.Getenv("LEARNING_NATIVE_LANG"),
		"LEARNING_TARGET_LANG": os.Getenv("LEARNING_TARGET_LANG"),
		"LEARNING_APP_CODE":    os.Getenv("LEARNING_APP_CODE"),
		"GRAMMAR_BUNDLE_ID":    os.Getenv("GRAMMAR_BUNDLE_ID"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	os.Setenv("LEARNING_PAIR", "ru-es")
	os.Setenv("LEARNING_NATIVE_LANG", "ru")
	os.Setenv("LEARNING_TARGET_LANG", "en")
	os.Setenv("LEARNING_APP_CODE", "english")
	os.Setenv("GRAMMAR_BUNDLE_ID", "en")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for inconsistent LEARNING_PAIR vs target lang")
	}
	if !strings.Contains(err.Error(), "learning config") || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected learning pair validation error, got %v", err)
	}
}

func TestLoad_ValidSpanishLearningPair(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":               os.Getenv("AI_URL"),
		"AI_API_KEY":           os.Getenv("AI_API_KEY"),
		"AI_PROMPT":            os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":       os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET":    os.Getenv("WEBAPP_JWT_SECRET"),
		"DATABASE_URL":         os.Getenv("DATABASE_URL"),
		"LEARNING_PAIR":        os.Getenv("LEARNING_PAIR"),
		"LEARNING_NATIVE_LANG": os.Getenv("LEARNING_NATIVE_LANG"),
		"LEARNING_TARGET_LANG": os.Getenv("LEARNING_TARGET_LANG"),
		"LEARNING_APP_CODE":    os.Getenv("LEARNING_APP_CODE"),
		"GRAMMAR_BUNDLE_ID":    os.Getenv("GRAMMAR_BUNDLE_ID"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	os.Setenv("LEARNING_PAIR", "ru-es")
	os.Setenv("LEARNING_NATIVE_LANG", "ru")
	os.Setenv("LEARNING_TARGET_LANG", "es")
	os.Setenv("LEARNING_APP_CODE", "spanish")
	os.Setenv("GRAMMAR_BUNDLE_ID", "es")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Learning.Pair != "ru-es" || cfg.Learning.NativeLang != "ru" || cfg.Learning.TargetLang != "es" {
		t.Fatalf("learning pair: got pair=%q native=%q target=%q", cfg.Learning.Pair, cfg.Learning.NativeLang, cfg.Learning.TargetLang)
	}
	if cfg.Learning.AppCode != "spanish" || cfg.Learning.GrammarBundleID != "es" {
		t.Fatalf("learning app: got app_code=%q grammar_bundle_id=%q", cfg.Learning.AppCode, cfg.Learning.GrammarBundleID)
	}
}

func TestLoad_WhitespaceOnlyGrammarBundleID(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":            os.Getenv("AI_URL"),
		"AI_API_KEY":        os.Getenv("AI_API_KEY"),
		"AI_PROMPT":         os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":    os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET": os.Getenv("WEBAPP_JWT_SECRET"),
		"DATABASE_URL":      os.Getenv("DATABASE_URL"),
		"GRAMMAR_BUNDLE_ID": os.Getenv("GRAMMAR_BUNDLE_ID"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()
	// Viper keeps numeric defaults when env is unset; empty string may not override.
	// Whitespace-only still binds and trims to empty in ValidateLearningConfig.
	os.Setenv("GRAMMAR_BUNDLE_ID", "   ")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for whitespace-only GRAMMAR_BUNDLE_ID")
	}
	if !strings.Contains(err.Error(), "learning config") || !strings.Contains(err.Error(), "GRAMMAR_BUNDLE_ID") {
		t.Fatalf("expected GRAMMAR_BUNDLE_ID validation error, got %v", err)
	}
}

func TestLoad_InvalidEmbeddedGrammarBundleID(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":               os.Getenv("AI_URL"),
		"AI_API_KEY":           os.Getenv("AI_API_KEY"),
		"AI_PROMPT":            os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":       os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET":    os.Getenv("WEBAPP_JWT_SECRET"),
		"DATABASE_URL":         os.Getenv("DATABASE_URL"),
		"GRAMMAR_BUNDLE_ID":    os.Getenv("GRAMMAR_BUNDLE_ID"),
		"GRAMMAR_BUNDLE_DIR":   os.Getenv("GRAMMAR_BUNDLE_DIR"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()
	os.Unsetenv("GRAMMAR_BUNDLE_DIR")
	os.Setenv("GRAMMAR_BUNDLE_ID", "fr")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for unknown embedded GRAMMAR_BUNDLE_ID")
	}
	if !strings.Contains(err.Error(), "learning config") {
		t.Fatalf("expected learning config error, got %v", err)
	}
}

func TestLoad_GrammarBundleDir_MissingSections(t *testing.T) {
	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_URL":             os.Getenv("AI_URL"),
		"AI_API_KEY":         os.Getenv("AI_API_KEY"),
		"AI_PROMPT":          os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":     os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET":  os.Getenv("WEBAPP_JWT_SECRET"),
		"DATABASE_URL":       os.Getenv("DATABASE_URL"),
		"GRAMMAR_BUNDLE_DIR": os.Getenv("GRAMMAR_BUNDLE_DIR"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	tmp := t.TempDir()
	os.Setenv("AI_URL", "http://test.local")
	os.Setenv("AI_API_KEY", "test-key")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Unsetenv("AI_PROMPT_FILE")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	unsetLearningEnvForDefaults()
	os.Setenv("GRAMMAR_BUNDLE_DIR", tmp)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for GRAMMAR_BUNDLE_DIR without sections.json")
	}
	if !strings.Contains(err.Error(), "sections.json") {
		t.Fatalf("expected sections.json error, got %v", err)
	}
}

func TestLoad_LoadsLanguageSpecificDotenvOverridesBase(t *testing.T) {
	learningEnv := backupLearningEnv()
	keys := []string{
		"AI_URL", "AI_API_KEY", "AI_PROMPT", "AI_PROMPT_FILE",
		"WEBAPP_JWT_SECRET", "DATABASE_DRIVER", "DATABASE_URL",
		"LEARNING_PAIR", "LEARNING_NATIVE_LANG", "LEARNING_TARGET_LANG",
		"LEARNING_APP_CODE", "GRAMMAR_BUNDLE_ID",
	}
	original := make(map[string]string, len(keys))
	for _, k := range keys {
		original[k] = os.Getenv(k)
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range original {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	baseEnv := strings.Join([]string{
		"DATABASE_URL=postgres://base/base",
		"AI_URL=http://base.local",
		"AI_API_KEY=base-key",
		"AI_PROMPT=base prompt",
		"WEBAPP_JWT_SECRET=base-secret",
		"LEARNING_PAIR=ru-es",
		"LEARNING_NATIVE_LANG=ru",
		"LEARNING_TARGET_LANG=es",
		"LEARNING_APP_CODE=spanish",
		"GRAMMAR_BUNDLE_ID=es",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte(baseEnv), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	langEnv := "DATABASE_URL=postgres://es/es\n"
	if err := os.WriteFile(filepath.Join(tmp, ".env.es"), []byte(langEnv), 0644); err != nil {
		t.Fatalf("write .env.es: %v", err)
	}

	for _, k := range keys {
		os.Unsetenv(k)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.URL != "postgres://es/es" {
		t.Fatalf("expected DATABASE_URL from .env.es, got %q", cfg.Database.URL)
	}
}

func TestLoad_DotenvDoesNotOverrideInitialShellEnv(t *testing.T) {
	learningEnv := backupLearningEnv()
	keys := []string{
		"AI_URL", "AI_API_KEY", "AI_PROMPT", "AI_PROMPT_FILE",
		"WEBAPP_JWT_SECRET", "DATABASE_DRIVER", "DATABASE_URL",
		"LEARNING_PAIR", "LEARNING_NATIVE_LANG", "LEARNING_TARGET_LANG",
		"LEARNING_APP_CODE", "GRAMMAR_BUNDLE_ID",
	}
	original := make(map[string]string, len(keys))
	for _, k := range keys {
		original[k] = os.Getenv(k)
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range original {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()

	baseEnv := strings.Join([]string{
		"DATABASE_URL=postgres://base/base",
		"AI_URL=http://base.local",
		"AI_API_KEY=base-key",
		"AI_PROMPT=base prompt",
		"WEBAPP_JWT_SECRET=base-secret",
		"LEARNING_PAIR=ru-es",
		"LEARNING_NATIVE_LANG=ru",
		"LEARNING_TARGET_LANG=es",
		"LEARNING_APP_CODE=spanish",
		"GRAMMAR_BUNDLE_ID=es",
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(tmp, ".env"), []byte(baseEnv), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, ".env.es"), []byte("DATABASE_URL=postgres://es/es\n"), 0644); err != nil {
		t.Fatalf("write .env.es: %v", err)
	}

	for _, k := range keys {
		os.Unsetenv(k)
	}
	// Simulate shell/CI explicit value; must win over dotenv files.
	os.Setenv("DATABASE_URL", "postgres://shell/shell")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.URL != "postgres://shell/shell" {
		t.Fatalf("expected shell DATABASE_URL to win, got %q", cfg.Database.URL)
	}
}

func TestLoad_PolzaProvider(t *testing.T) {
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		_ = os.Chdir(origDir)
	}()

	learningEnv := backupLearningEnv()
	originalEnv := map[string]string{
		"AI_PROVIDER":            os.Getenv("AI_PROVIDER"),
		"AI_URL":                 os.Getenv("AI_URL"),
		"AI_API_KEY":             os.Getenv("AI_API_KEY"),
		"POLZA_AI_URL":           os.Getenv("POLZA_AI_URL"),
		"POLZA_AI_API_KEY":       os.Getenv("POLZA_AI_API_KEY"),
		"OPENROUTER_SOCKS5_PROXY": os.Getenv("OPENROUTER_SOCKS5_PROXY"),
		"AI_PROMPT":              os.Getenv("AI_PROMPT"),
		"AI_PROMPT_FILE":         os.Getenv("AI_PROMPT_FILE"),
		"WEBAPP_JWT_SECRET":      os.Getenv("WEBAPP_JWT_SECRET"),
		"DATABASE_URL":           os.Getenv("DATABASE_URL"),
		"SPEAKING_EVAL_API_KEY":  os.Getenv("SPEAKING_EVAL_API_KEY"),
		"SPEAKING_EVAL_BASE_URL": os.Getenv("SPEAKING_EVAL_BASE_URL"),
	}
	defer func() {
		restoreLearningEnv(learningEnv)
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	unsetLearningEnvForDefaults()
	os.Setenv("AI_PROVIDER", AIProviderPolza)
	os.Setenv("AI_URL", "https://openrouter.ai/api/v1")
	os.Setenv("AI_API_KEY", "openrouter-key")
	os.Setenv("POLZA_AI_API_KEY", "polza-key")
	os.Setenv("OPENROUTER_SOCKS5_PROXY", "51.254.98.124:1080")
	os.Setenv("AI_PROMPT", "test prompt")
	os.Setenv("WEBAPP_JWT_SECRET", "test-secret")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test?sslmode=disable")
	os.Unsetenv("SPEAKING_EVAL_API_KEY")
	os.Unsetenv("SPEAKING_EVAL_BASE_URL")
	os.Unsetenv("POLZA_AI_URL")
	os.Unsetenv("TTS_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AI.Provider != AIProviderPolza {
		t.Fatalf("AI provider = %q, want %q", cfg.AI.Provider, AIProviderPolza)
	}
	if cfg.AI.URL != defaultPolzaURL {
		t.Fatalf("AI url = %q, want %q", cfg.AI.URL, defaultPolzaURL)
	}
	if cfg.AI.APIKey != "polza-key" {
		t.Fatalf("AI api key = %q, want polza-key", cfg.AI.APIKey)
	}
	if cfg.AI.Socks5Proxy != "" {
		t.Fatalf("AI socks5 = %q, want empty", cfg.AI.Socks5Proxy)
	}
	if cfg.Speaking.EvalAPIKey != "openrouter-key" {
		t.Fatalf("Speaking eval api key = %q, want openrouter-key", cfg.Speaking.EvalAPIKey)
	}
	if cfg.Speaking.Socks5Proxy != "51.254.98.124:1080" {
		t.Fatalf("Speaking socks5 = %q, want proxy addr", cfg.Speaking.Socks5Proxy)
	}
}

