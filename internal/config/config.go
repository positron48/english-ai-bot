package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"tgbot-skeleton/internal/ai"
	"tgbot-skeleton/internal/grammarbundle"
)

var langCodeSegment = regexp.MustCompile(`^[a-z]{2,8}$`)

// Config represents the application configuration
type Config struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	Server   ServerConfig   `mapstructure:"server"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	AI       AIConfig       `mapstructure:"ai"`
	TTS      TTSConfig      `mapstructure:"tts"`
	Bot      BotConfig      `mapstructure:"bot"`
	Database DatabaseConfig `mapstructure:"database"`
	Training TrainingConfig `mapstructure:"training"`
	Admin    AdminConfig    `mapstructure:"admin"`
	WebApp   WebAppConfig   `mapstructure:"webapp"`
	Learning LearningConfig `mapstructure:"learning"`
	Linglow  LinglowConfig  `mapstructure:"linglow"`
	Speaking SpeakingConfig `mapstructure:"speaking"`
}

// LearningConfig holds language pair and bundle identity for multilang deployments.
type LearningConfig struct {
	Pair             string `mapstructure:"pair"`
	NativeLang       string `mapstructure:"native_lang"`
	TargetLang       string `mapstructure:"target_lang"`
	AppCode          string `mapstructure:"app_code"`
	GrammarBundleID  string `mapstructure:"grammar_bundle_id"`
	GrammarBundleDir string `mapstructure:"grammar_bundle_dir"` // optional: filesystem root with sections.json (overrides embedded bundle)
	ContentSource    string `mapstructure:"content_source"`     // bundle | db
}

// LinglowConfig holds migration flags for the Linglow v2 architecture.
type LinglowConfig struct {
	EventsWriteEnabled bool `mapstructure:"events_write_enabled"`
	SRSReadEnabled     bool `mapstructure:"srs_read_enabled"`
	SRSWriteEnabled    bool `mapstructure:"srs_write_enabled"`
}

// DefaultLearningConfig returns the canonical RU→EN English instance defaults (matches viper defaults in Load).
func DefaultLearningConfig() LearningConfig {
	return LearningConfig{
		Pair:             "ru-en",
		NativeLang:       "ru",
		TargetLang:       "en",
		AppCode:          "english",
		GrammarBundleID:  "en",
		GrammarBundleDir: "",
		ContentSource:    "bundle",
	}
}

// ValidateLearningConfig checks LEARNING_* / GRAMMAR_BUNDLE_ID consistency.
func ValidateLearningConfig(lc LearningConfig) error {
	pair := strings.TrimSpace(strings.ToLower(lc.Pair))
	native := strings.TrimSpace(strings.ToLower(lc.NativeLang))
	target := strings.TrimSpace(strings.ToLower(lc.TargetLang))
	appCode := strings.TrimSpace(lc.AppCode)
	bundleID := strings.TrimSpace(lc.GrammarBundleID)

	if pair == "" {
		return fmt.Errorf("LEARNING_PAIR is required (e.g. ru-en)")
	}
	if strings.Count(pair, "-") != 1 {
		return fmt.Errorf("LEARNING_PAIR %q must contain exactly one hyphen between two language codes (e.g. ru-en)", lc.Pair)
	}
	parts := strings.SplitN(pair, "-", 2)
	pNative, pTarget := parts[0], parts[1]
	if !langCodeSegment.MatchString(pNative) || !langCodeSegment.MatchString(pTarget) {
		return fmt.Errorf("LEARNING_PAIR %q: each side must be a lowercase language code (2–8 letters, e.g. ru-en)", lc.Pair)
	}

	if native == "" || target == "" {
		return fmt.Errorf("LEARNING_NATIVE_LANG and LEARNING_TARGET_LANG are required")
	}
	if !langCodeSegment.MatchString(native) {
		return fmt.Errorf("LEARNING_NATIVE_LANG %q must be a lowercase language code (2–8 letters)", lc.NativeLang)
	}
	if !langCodeSegment.MatchString(target) {
		return fmt.Errorf("LEARNING_TARGET_LANG %q must be a lowercase language code (2–8 letters)", lc.TargetLang)
	}

	if pNative != native || pTarget != target {
		return fmt.Errorf("LEARNING_PAIR %q does not match LEARNING_NATIVE_LANG=%q and LEARNING_TARGET_LANG=%q (expected %s-%s)",
			lc.Pair, lc.NativeLang, lc.TargetLang, native, target)
	}

	if appCode == "" {
		return fmt.Errorf("LEARNING_APP_CODE is required")
	}
	if bundleID == "" {
		return fmt.Errorf("GRAMMAR_BUNDLE_ID is required")
	}
	return nil
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	Token           string `mapstructure:"token"`
	APIBaseURL      string `mapstructure:"api_base_url"`
	Socks5ProxyAddr string `mapstructure:"socks5_proxy_addr"`
	Debug           bool   `mapstructure:"debug"`
	UpdatesTimeout  int    `mapstructure:"updates_timeout"`
	WebhookEnable   bool   `mapstructure:"webhook_enable"`
	WebhookURL      string `mapstructure:"webhook_url"`
	WebhookDomain   string `mapstructure:"webhook_domain"`
	WebhookPath     string `mapstructure:"webhook_path"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Address string `mapstructure:"address"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

// AIConfig holds AI provider configuration
type AIConfig struct {
	URL            string `mapstructure:"url"`
	Model          string `mapstructure:"model"`
	ModelHigh      string `mapstructure:"model_high"`
	APIKey         string `mapstructure:"api_key"`
	Prompt         string `mapstructure:"prompt"`
	PromptFile     string `mapstructure:"prompt_file"`
	RequestTimeout string `mapstructure:"request_timeout"` // e.g. 120s, 3m; HTTP client timeout for chat/completions (default 30s)
}

// TTSConfig holds text-to-speech/pronunciation audio configuration
type TTSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	Provider           string `mapstructure:"provider"` // auto | dictionary | openrouter | external
	ExternalOnly       bool   `mapstructure:"external_only"`
	AudioDir           string `mapstructure:"audio_dir"`
	PublicBasePath     string `mapstructure:"public_base_path"`
	Model              string `mapstructure:"model"`
	Voice              string `mapstructure:"voice"`
	BaseURL            string `mapstructure:"base_url"`
	APIKey             string `mapstructure:"api_key"`
	RequestTimeout     string `mapstructure:"request_timeout"`
	PrefetchEnabled    bool   `mapstructure:"prefetch_enabled"`
	PrefetchWorkers    int    `mapstructure:"prefetch_workers"`
	BackfillInterval   string `mapstructure:"backfill_interval"`
	BackfillBatchSize  int    `mapstructure:"backfill_batch_size"`
	RetryBaseDelay     string `mapstructure:"retry_base_delay"`
	RetryMaxDelay      string `mapstructure:"retry_max_delay"`
	MaxRetries         int    `mapstructure:"max_retries"`
	DictionaryBaseURL  string `mapstructure:"dictionary_base_url"`
	DictionaryEnabled  bool   `mapstructure:"dictionary_enabled"`
	DictionaryMinDelay string `mapstructure:"dictionary_min_delay"`
	// ChatPronunciationPrompt is the user message for OpenRouter chat+audio TTS. Use {word} — replaced with the word in backticks. Empty = built-in English default.
	ChatPronunciationPrompt string `mapstructure:"chat_pronunciation_prompt"`
	InternalEnabled         bool   `mapstructure:"internal_enabled"`
	InternalTokensJSON      string `mapstructure:"internal_tokens_json"`
	InternalMaxPendingLimit int    `mapstructure:"internal_max_pending_limit"`
	InternalMaxUploadMB     int    `mapstructure:"internal_max_upload_mb"`
}

// SpeakingConfig holds speaking evaluation mode settings.
type SpeakingConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	EvalModel          string `mapstructure:"eval_model"`
	EvalBaseURL        string `mapstructure:"eval_base_url"`
	EvalAPIKey         string `mapstructure:"eval_api_key"`
	EvalTimeout        string `mapstructure:"eval_timeout"`
	MaxAudioMB         int    `mapstructure:"max_audio_mb"`
	MaxAttemptsDefault int    `mapstructure:"max_attempts_default"`
	AcceptMeaningScore int    `mapstructure:"accept_meaning_score"`
	SessionTaskCount   int    `mapstructure:"session_task_count"`
}

// BotConfig holds bot messages and behavior configuration
type BotConfig struct {
	StartMessage          string `mapstructure:"start_message"`
	HelpMessage           string `mapstructure:"help_message"`
	UnknownCommandMessage string `mapstructure:"unknown_command_message"`
	ErrorMessage          string `mapstructure:"error_message"`
	EmptyMessage          string `mapstructure:"empty_message"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	Path   string `mapstructure:"path"`
	URL    string `mapstructure:"url"`
}

// TrainingConfig holds training system configuration
type TrainingConfig struct {
	WorkerEnabled           bool   `mapstructure:"worker_enabled"`
	WorkerInterval          string `mapstructure:"worker_interval"`
	WorkerBatchSize         int    `mapstructure:"worker_batch_size"`
	LLMWorkers              int    `mapstructure:"llm_workers"`
	PromptFile              string `mapstructure:"prompt_file"`
	CircuitBreakerThreshold int    `mapstructure:"circuit_breaker_threshold"`
	CircuitBreakerAutoReset int    `mapstructure:"circuit_breaker_auto_reset_hours"`
	OptionsDelayMS          int    `mapstructure:"options_delay_ms"`
	WrongAnswerDelaySeconds int    `mapstructure:"wrong_answer_delay_seconds"`
	SpanishVerbFormsEnabled bool   `mapstructure:"spanish_verb_forms_enabled"`
	VerbFormsMaxCards       int    `mapstructure:"verb_forms_max_cards_per_session"`
	VerbFormsMaxNew         int    `mapstructure:"verb_forms_max_new_per_session"`
	// VerbFormsTypedMinReps: after this many successful SRS reps (and not in "learning"), verb-form cards become eligible for typing the whole form (see VerbFormsTypedChancePercent).
	VerbFormsTypedMinReps int `mapstructure:"verb_forms_typed_min_reps"`
	// VerbFormsTypedChancePercent: when eligible for typed mode, each card uses typed with this probability (0–100); otherwise multiple choice. Default 50 (like spell/type mix in word training).
	VerbFormsTypedChancePercent int `mapstructure:"verb_forms_typed_chance_percent"`
}

// AdminConfig holds admin configuration
type AdminConfig struct {
	TelegramID    int64 `mapstructure:"telegram_id"`
	DBQueryAccess bool  `mapstructure:"db_query_access"`
}

// WebAppConfig holds web app configuration
type WebAppConfig struct {
	PublicURL                 string `mapstructure:"public_url"`
	SessionSecret             string `mapstructure:"session_secret"`
	JWTSecret                 string `mapstructure:"jwt_secret"`
	ComplaintsServiceToken    string `mapstructure:"complaints_service_token"`
	InternalServiceTokensJSON string `mapstructure:"internal_service_tokens_json"`
	OTPTTLSeconds             int    `mapstructure:"otp_ttl_seconds"`
	SessionTTLHours           int    `mapstructure:"session_ttl_hours"`
	JWTTTLHours               int    `mapstructure:"jwt_ttl_hours"`
	RefreshTTLHours           int    `mapstructure:"refresh_ttl_hours"`
	ViteDevServerURL          string `mapstructure:"vite_dev_server_url"`
	AndroidCertFingerprints   string `mapstructure:"android_cert_fingerprints"`

	// Rate limiting configuration
	RateLimitAuthRequestOTPPerIP         int `mapstructure:"rate_limit_auth_request_otp_per_ip"`
	RateLimitAuthRequestOTPPerIPUser     int `mapstructure:"rate_limit_auth_request_otp_per_ip_user"`
	RateLimitAuthOTPPerIP                int `mapstructure:"rate_limit_auth_otp_per_ip"`
	RateLimitAuthOTPPerIPUser            int `mapstructure:"rate_limit_auth_otp_per_ip_user"`
	RateLimitAuthTelegramUnsafePerIP     int `mapstructure:"rate_limit_auth_telegram_unsafe_per_ip"`
	RateLimitAuthTelegramUnsafePerIPUser int `mapstructure:"rate_limit_auth_telegram_unsafe_per_ip_user"`
	RateLimitAuthTelegramPerIP           int `mapstructure:"rate_limit_auth_telegram_per_ip"`
	RateLimitAuthRefreshPerIP            int `mapstructure:"rate_limit_auth_refresh_per_ip"`
	RateLimitAppAPIPerUser               int `mapstructure:"rate_limit_app_api_per_user"`
	RateLimitAppChatPerUser              int `mapstructure:"rate_limit_app_chat_per_user"`
	RateLimitSpeakingPerUser             int `mapstructure:"rate_limit_speaking_per_user"`
	RateLimitWindowMinutes               int `mapstructure:"rate_limit_window_minutes"`
	RateLimitBurstMultiplier             int `mapstructure:"rate_limit_burst_multiplier"`
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	// Load .env chain:
	// 1) .env (base)
	// 2) .env.<learning target lang> (language overrides)
	// Do not override variables that were already present in the process environment.
	initialEnv := envKeysSnapshot()
	if err := godotenv.Load(".env"); err != nil {
		_ = err // optional
	}
	// Target lang can be provided by shell env or loaded from .env.
	targetLangHint := strings.ToLower(strings.TrimSpace(os.Getenv("LEARNING_TARGET_LANG")))
	if targetLangHint != "" && langCodeSegment.MatchString(targetLangHint) {
		_ = loadDotEnvFileWithoutOverridingInitialEnv(".env."+targetLangHint, initialEnv)
	}

	// Set default values
	viper.SetDefault("telegram.debug", false)
	viper.SetDefault("telegram.updates_timeout", 30)
	viper.SetDefault("telegram.webhook_enable", false)
	viper.SetDefault("telegram.webhook_path", "/webhook")
	viper.SetDefault("server.address", ":8184")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("ai.model", "gpt-3.5-turbo")
	viper.SetDefault("tts.enabled", true)
	viper.SetDefault("tts.provider", "auto")
	viper.SetDefault("tts.external_only", false)
	viper.SetDefault("tts.audio_dir", "/app/data/tts")
	viper.SetDefault("tts.public_base_path", "/media/tts")
	viper.SetDefault("tts.model", "openai/gpt-audio-mini")
	viper.SetDefault("tts.voice", "alloy")
	viper.SetDefault("tts.base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("tts.request_timeout", "45s")
	viper.SetDefault("tts.prefetch_enabled", true)
	viper.SetDefault("tts.prefetch_workers", 2)
	viper.SetDefault("tts.backfill_interval", "10m")
	viper.SetDefault("tts.backfill_batch_size", 200)
	viper.SetDefault("tts.retry_base_delay", "1m")
	viper.SetDefault("tts.retry_max_delay", "24h")
	viper.SetDefault("tts.max_retries", 10)
	viper.SetDefault("tts.dictionary_base_url", "https://api.dictionaryapi.dev/api/v2/entries/en")
	viper.SetDefault("tts.dictionary_enabled", true)
	viper.SetDefault("tts.dictionary_min_delay", "100ms")
	viper.SetDefault("tts.chat_pronunciation_prompt", "")
	viper.SetDefault("tts.internal_enabled", false)
	viper.SetDefault("tts.internal_tokens_json", "")
	viper.SetDefault("tts.internal_max_pending_limit", 500)
	viper.SetDefault("tts.internal_max_upload_mb", 10)
	viper.SetDefault("speaking.enabled", false)
	viper.SetDefault("speaking.eval_model", "openai/gpt-audio-mini")
	viper.SetDefault("speaking.eval_base_url", "https://openrouter.ai/api/v1")
	viper.SetDefault("speaking.eval_timeout", "60s")
	viper.SetDefault("speaking.max_audio_mb", 2)
	viper.SetDefault("speaking.max_attempts_default", 3)
	viper.SetDefault("speaking.accept_meaning_score", 3)
	viper.SetDefault("speaking.session_task_count", 5)
	viper.SetDefault("webapp.rate_limit_speaking_per_user", 30)
	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.path", "")

	// Training defaults
	viper.SetDefault("training.worker_enabled", true)
	viper.SetDefault("training.worker_interval", "30s")
	viper.SetDefault("training.worker_batch_size", 5)
	viper.SetDefault("training.llm_workers", 4)
	viper.SetDefault("training.prompt_file", "")
	viper.SetDefault("training.circuit_breaker_threshold", 5)
	viper.SetDefault("training.circuit_breaker_auto_reset_hours", 24)
	viper.SetDefault("training.options_delay_ms", 5000)
	viper.SetDefault("training.wrong_answer_delay_seconds", 5)
	viper.SetDefault("training.spanish_verb_forms_enabled", false)
	// Verb-form session: size of one run and how many brand-new (state=new) cards may be mixed in.
	// Defaults match a typical "presente indicativo" grid (~6 slots × several lemmas): one session can hold ~30 cards
	// and draw new cards in random order so a single lemma does not monopolize the slot when reviews are not due yet.
	viper.SetDefault("training.verb_forms_max_cards_per_session", 30)
	viper.SetDefault("training.verb_forms_max_new_per_session", 30)
	viper.SetDefault("training.verb_forms_typed_min_reps", 2)
	viper.SetDefault("training.verb_forms_typed_chance_percent", 50)

	// Admin defaults
	viper.SetDefault("admin.telegram_id", 0)
	viper.SetDefault("admin.db_query_access", false)

	// WebApp defaults
	viper.SetDefault("webapp.public_url", "")
	viper.SetDefault("webapp.session_secret", "")
	viper.SetDefault("webapp.jwt_secret", "")
	viper.SetDefault("webapp.complaints_service_token", "")
	viper.SetDefault("webapp.internal_service_tokens_json", "")
	viper.SetDefault("webapp.otp_ttl_seconds", 300)
	viper.SetDefault("webapp.session_ttl_hours", 720)
	viper.SetDefault("webapp.jwt_ttl_hours", 24)
	viper.SetDefault("webapp.refresh_ttl_hours", 720)
	viper.SetDefault("webapp.vite_dev_server_url", "http://localhost:5173")

	// Rate limiting defaults
	viper.SetDefault("webapp.rate_limit_auth_request_otp_per_ip", 10)
	viper.SetDefault("webapp.rate_limit_auth_request_otp_per_ip_user", 3)
	viper.SetDefault("webapp.rate_limit_auth_otp_per_ip", 20)
	viper.SetDefault("webapp.rate_limit_auth_otp_per_ip_user", 5)
	viper.SetDefault("webapp.rate_limit_auth_telegram_unsafe_per_ip", 30)
	viper.SetDefault("webapp.rate_limit_auth_telegram_unsafe_per_ip_user", 10)
	viper.SetDefault("webapp.rate_limit_auth_telegram_per_ip", 60)
	viper.SetDefault("webapp.rate_limit_auth_refresh_per_ip", 60)
	viper.SetDefault("webapp.rate_limit_app_api_per_user", 300)
	viper.SetDefault("webapp.rate_limit_app_chat_per_user", 60)
	viper.SetDefault("webapp.rate_limit_window_minutes", 1)
	viper.SetDefault("webapp.rate_limit_burst_multiplier", 2)

	// Learning / language pair (defaults: RU→EN English instance)
	viper.SetDefault("learning.pair", "ru-en")
	viper.SetDefault("learning.native_lang", "ru")
	viper.SetDefault("learning.target_lang", "en")
	viper.SetDefault("learning.app_code", "english")
	viper.SetDefault("learning.grammar_bundle_id", "en")
	viper.SetDefault("learning.content_source", "bundle")
	viper.SetDefault("linglow.events_write_enabled", false)

	// Bot message defaults
	viper.SetDefault("bot.start_message", "🇬🇧 Привет! Я ваш персональный преподаватель английского языка!\n\n📝 Что я умею:\n• Исправлять ошибки в английском тексте\n• Переводить с русского на английский\n• Создавать карточки слов с объяснениями\n\n💡 Как пользоваться:\n• Отправьте английский текст → получите исправления\n• Отправьте русский текст → получите перевод\n• Отправьте одно слово → получите карточку слова\n\nИспользуйте /help для подробной информации.")
	viper.SetDefault("bot.help_message", "📚 Помощь по использованию бота-преподавателя английского:\n\n🔤 **Одно слово** → Карточка слова с:\n• Частью речи\n• Транскрипцией IPA\n• Определением на русском\n• Примерами использования\n• Формами неправильных глаголов\n\n📝 **Английский текст** → Исправления:\n• Поиск ошибок (орфография, грамматика, пунктуация)\n• Подробные объяснения\n• Исправленная версия\n\n🇷🇺 **Русский текст** → Перевод:\n• Естественный перевод на английский\n• Анализ сложных фраз\n• Сохранение тона и стиля\n\n🔧 **Доступные команды:**\n• /help - Показать эту справку\n• /notification [daily|never|N] - Настроить периодичность уведомлений\n💬 Просто отправьте текст или слово - я сразу помогу!")
	viper.SetDefault("bot.unknown_command_message", "❓ Неизвестная команда. Используйте /help для получения информации о возможностях бота.")
	viper.SetDefault("bot.error_message", "Извините, при обработке сообщения произошла ошибка. Попробуйте позднее.")
	viper.SetDefault("bot.empty_message", "Пожалуйста, отправьте сообщение.")

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicitly bind environment variables to viper keys
	_ = viper.BindEnv("telegram.token", "TELEGRAM_TOKEN")
	_ = viper.BindEnv("telegram.api_base_url", "TELEGRAM_API_BASE_URL")
	_ = viper.BindEnv("telegram.socks5_proxy_addr", "TELEGRAM_SOCKS5_PROXY_ADDR")
	_ = viper.BindEnv("telegram.debug", "TELEGRAM_DEBUG")
	_ = viper.BindEnv("telegram.updates_timeout", "TELEGRAM_UPDATES_TIMEOUT")
	_ = viper.BindEnv("telegram.webhook_enable", "TELEGRAM_WEBHOOK_ENABLE")
	_ = viper.BindEnv("telegram.webhook_url", "TELEGRAM_WEBHOOK_URL")
	_ = viper.BindEnv("telegram.webhook_domain", "TELEGRAM_WEBHOOK_DOMAIN")
	_ = viper.BindEnv("telegram.webhook_path", "TELEGRAM_WEBHOOK_PATH")
	_ = viper.BindEnv("server.address", "SERVER_ADDRESS")
	_ = viper.BindEnv("logging.level", "LOG_LEVEL")
	_ = viper.BindEnv("ai.url", "AI_URL")
	_ = viper.BindEnv("ai.model", "AI_MODEL")
	_ = viper.BindEnv("ai.model_high", "AI_MODEL_HIGH")
	_ = viper.BindEnv("ai.api_key", "AI_API_KEY")
	_ = viper.BindEnv("ai.prompt", "AI_PROMPT")
	_ = viper.BindEnv("ai.prompt_file", "AI_PROMPT_FILE")
	_ = viper.BindEnv("ai.request_timeout", "AI_REQUEST_TIMEOUT")
	_ = viper.BindEnv("tts.enabled", "TTS_ENABLED", "TTS_ENABLE")
	_ = viper.BindEnv("tts.provider", "TTS_PROVIDER")
	_ = viper.BindEnv("tts.external_only", "TTS_EXTERNAL_ONLY")
	_ = viper.BindEnv("tts.audio_dir", "TTS_AUDIO_DIR")
	_ = viper.BindEnv("tts.public_base_path", "TTS_PUBLIC_BASE_PATH")
	_ = viper.BindEnv("tts.model", "TTS_MODEL")
	_ = viper.BindEnv("tts.voice", "TTS_VOICE")
	_ = viper.BindEnv("tts.base_url", "TTS_BASE_URL")
	_ = viper.BindEnv("tts.api_key", "TTS_API_KEY")
	_ = viper.BindEnv("tts.request_timeout", "TTS_REQUEST_TIMEOUT")
	_ = viper.BindEnv("tts.prefetch_enabled", "TTS_PREFETCH_ENABLED")
	_ = viper.BindEnv("tts.prefetch_workers", "TTS_PREFETCH_WORKERS")
	_ = viper.BindEnv("tts.backfill_interval", "TTS_BACKFILL_INTERVAL")
	_ = viper.BindEnv("tts.backfill_batch_size", "TTS_BACKFILL_BATCH_SIZE")
	_ = viper.BindEnv("tts.retry_base_delay", "TTS_RETRY_BASE_DELAY")
	_ = viper.BindEnv("tts.retry_max_delay", "TTS_RETRY_MAX_DELAY")
	_ = viper.BindEnv("tts.max_retries", "TTS_MAX_RETRIES")
	_ = viper.BindEnv("tts.dictionary_base_url", "TTS_DICTIONARY_BASE_URL")
	_ = viper.BindEnv("tts.dictionary_min_delay", "TTS_DICTIONARY_MIN_DELAY")
	_ = viper.BindEnv("tts.dictionary_enabled", "TTS_DICTIONARY_ENABLED")
	_ = viper.BindEnv("tts.chat_pronunciation_prompt", "TTS_CHAT_PRONUNCIATION_PROMPT")
	_ = viper.BindEnv("tts.internal_enabled", "TTS_INTERNAL_ENABLED")
	_ = viper.BindEnv("tts.internal_tokens_json", "TTS_INTERNAL_TOKENS_JSON")
	_ = viper.BindEnv("tts.internal_max_pending_limit", "TTS_INTERNAL_MAX_PENDING_LIMIT")
	_ = viper.BindEnv("tts.internal_max_upload_mb", "TTS_INTERNAL_MAX_UPLOAD_MB")
	_ = viper.BindEnv("speaking.enabled", "SPEAKING_MODE_ENABLED")
	_ = viper.BindEnv("speaking.eval_model", "SPEAKING_EVAL_MODEL")
	_ = viper.BindEnv("speaking.eval_base_url", "SPEAKING_EVAL_BASE_URL")
	_ = viper.BindEnv("speaking.eval_api_key", "SPEAKING_EVAL_API_KEY")
	_ = viper.BindEnv("speaking.eval_timeout", "SPEAKING_EVAL_TIMEOUT")
	_ = viper.BindEnv("speaking.max_audio_mb", "SPEAKING_MAX_AUDIO_MB")
	_ = viper.BindEnv("speaking.max_attempts_default", "SPEAKING_MAX_ATTEMPTS_DEFAULT")
	_ = viper.BindEnv("speaking.accept_meaning_score", "SPEAKING_ACCEPT_MEANING_SCORE")
	_ = viper.BindEnv("speaking.session_task_count", "SPEAKING_SESSION_TASK_COUNT")
	_ = viper.BindEnv("bot.start_message", "BOT_START_MESSAGE")
	_ = viper.BindEnv("bot.help_message", "BOT_HELP_MESSAGE")
	_ = viper.BindEnv("bot.unknown_command_message", "BOT_UNKNOWN_COMMAND_MESSAGE")
	_ = viper.BindEnv("bot.error_message", "BOT_ERROR_MESSAGE")
	_ = viper.BindEnv("bot.empty_message", "BOT_EMPTY_MESSAGE")
	_ = viper.BindEnv("database.path", "DATABASE_PATH")
	_ = viper.BindEnv("database.driver", "DATABASE_DRIVER")
	_ = viper.BindEnv("database.url", "DATABASE_URL")
	_ = viper.BindEnv("training.worker_enabled", "TRAINING_WORKER_ENABLED")
	_ = viper.BindEnv("training.worker_interval", "TRAINING_WORKER_INTERVAL")
	_ = viper.BindEnv("training.worker_batch_size", "TRAINING_WORKER_BATCH_SIZE")
	_ = viper.BindEnv("training.llm_workers", "TRAINING_LLM_WORKERS")
	_ = viper.BindEnv("training.prompt_file", "TRAINING_PROMPT_FILE")
	_ = viper.BindEnv("training.circuit_breaker_threshold", "CIRCUIT_BREAKER_THRESHOLD")
	_ = viper.BindEnv("training.circuit_breaker_auto_reset_hours", "CIRCUIT_BREAKER_AUTO_RESET_HOURS")
	_ = viper.BindEnv("training.options_delay_ms", "TRAINING_OPTIONS_DELAY_MS")
	_ = viper.BindEnv("training.wrong_answer_delay_seconds", "TRAINING_WRONG_ANSWER_DELAY_SECONDS")
	_ = viper.BindEnv("training.spanish_verb_forms_enabled", "SPANISH_VERB_FORMS_ENABLED")
	_ = viper.BindEnv("training.verb_forms_max_cards_per_session", "VERB_FORMS_MAX_CARDS_PER_SESSION")
	_ = viper.BindEnv("training.verb_forms_max_new_per_session", "VERB_FORMS_MAX_NEW_PER_SESSION")
	_ = viper.BindEnv("training.verb_forms_typed_min_reps", "VERB_FORMS_TYPED_MIN_REPS")
	_ = viper.BindEnv("training.verb_forms_typed_chance_percent", "VERB_FORMS_TYPED_CHANCE_PERCENT")
	_ = viper.BindEnv("admin.telegram_id", "ADMIN_TELEGRAM_ID")
	_ = viper.BindEnv("admin.db_query_access", "DB_QUERY_ACCESS")
	_ = viper.BindEnv("webapp.public_url", "WEBAPP_PUBLIC_URL")
	_ = viper.BindEnv("webapp.session_secret", "WEBAPP_SESSION_SECRET")
	_ = viper.BindEnv("webapp.jwt_secret", "WEBAPP_JWT_SECRET")
	_ = viper.BindEnv("webapp.complaints_service_token", "COMPLAINTS_SERVICE_TOKEN")
	_ = viper.BindEnv("webapp.internal_service_tokens_json", "WEBAPP_INTERNAL_SERVICE_TOKENS_JSON")
	_ = viper.BindEnv("webapp.otp_ttl_seconds", "WEBAPP_OTP_TTL_SECONDS")
	_ = viper.BindEnv("webapp.session_ttl_hours", "WEBAPP_SESSION_TTL_HOURS")
	_ = viper.BindEnv("webapp.jwt_ttl_hours", "WEBAPP_JWT_TTL_HOURS")
	_ = viper.BindEnv("webapp.refresh_ttl_hours", "WEBAPP_REFRESH_TTL_HOURS")
	_ = viper.BindEnv("webapp.vite_dev_server_url", "WEBAPP_VITE_DEV_SERVER_URL")
	_ = viper.BindEnv("webapp.android_cert_fingerprints", "WEBAPP_ANDROID_CERT_FINGERPRINTS")
	_ = viper.BindEnv("webapp.rate_limit_auth_request_otp_per_ip", "WEBAPP_RATE_LIMIT_AUTH_REQUEST_OTP_PER_IP")
	_ = viper.BindEnv("webapp.rate_limit_auth_request_otp_per_ip_user", "WEBAPP_RATE_LIMIT_AUTH_REQUEST_OTP_PER_IP_USER")
	_ = viper.BindEnv("webapp.rate_limit_auth_otp_per_ip", "WEBAPP_RATE_LIMIT_AUTH_OTP_PER_IP")
	_ = viper.BindEnv("webapp.rate_limit_auth_otp_per_ip_user", "WEBAPP_RATE_LIMIT_AUTH_OTP_PER_IP_USER")
	_ = viper.BindEnv("webapp.rate_limit_auth_telegram_unsafe_per_ip", "WEBAPP_RATE_LIMIT_AUTH_TELEGRAM_UNSAFE_PER_IP")
	_ = viper.BindEnv("webapp.rate_limit_auth_telegram_unsafe_per_ip_user", "WEBAPP_RATE_LIMIT_AUTH_TELEGRAM_UNSAFE_PER_IP_USER")
	_ = viper.BindEnv("webapp.rate_limit_auth_telegram_per_ip", "WEBAPP_RATE_LIMIT_AUTH_TELEGRAM_PER_IP")
	_ = viper.BindEnv("webapp.rate_limit_auth_refresh_per_ip", "WEBAPP_RATE_LIMIT_AUTH_REFRESH_PER_IP")
	_ = viper.BindEnv("webapp.rate_limit_app_api_per_user", "WEBAPP_RATE_LIMIT_APP_API_PER_USER")
	_ = viper.BindEnv("webapp.rate_limit_app_chat_per_user", "WEBAPP_RATE_LIMIT_APP_CHAT_PER_USER")
	_ = viper.BindEnv("webapp.rate_limit_speaking_per_user", "WEBAPP_RATE_LIMIT_SPEAKING_PER_USER")
	_ = viper.BindEnv("webapp.rate_limit_window_minutes", "WEBAPP_RATE_LIMIT_WINDOW_MINUTES")
	_ = viper.BindEnv("webapp.rate_limit_burst_multiplier", "WEBAPP_RATE_LIMIT_BURST_MULTIPLIER")
	_ = viper.BindEnv("learning.pair", "LEARNING_PAIR")
	_ = viper.BindEnv("learning.native_lang", "LEARNING_NATIVE_LANG")
	_ = viper.BindEnv("learning.target_lang", "LEARNING_TARGET_LANG")
	_ = viper.BindEnv("learning.app_code", "LEARNING_APP_CODE")
	_ = viper.BindEnv("learning.grammar_bundle_id", "GRAMMAR_BUNDLE_ID")
	_ = viper.BindEnv("learning.grammar_bundle_dir", "GRAMMAR_BUNDLE_DIR")
	_ = viper.BindEnv("learning.content_source", "CONTENT_SOURCE")
	_ = viper.BindEnv("linglow.events_write_enabled", "LINGLOW_EVENTS_WRITE_ENABLED")
	_ = viper.BindEnv("linglow.srs_read_enabled", "LINGLOW_SRS_READ_ENABLED")
	_ = viper.BindEnv("linglow.srs_write_enabled", "LINGLOW_SRS_WRITE_ENABLED")

	// Set config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use env vars and defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := ValidateLearningConfig(config.Learning); err != nil {
		return nil, fmt.Errorf("learning config: %w", err)
	}
	config.Learning.Pair = strings.ToLower(strings.TrimSpace(config.Learning.Pair))
	config.Learning.NativeLang = strings.ToLower(strings.TrimSpace(config.Learning.NativeLang))
	config.Learning.TargetLang = strings.ToLower(strings.TrimSpace(config.Learning.TargetLang))
	config.Learning.AppCode = strings.TrimSpace(config.Learning.AppCode)
	config.Learning.GrammarBundleID = strings.TrimSpace(config.Learning.GrammarBundleID)
	config.Learning.GrammarBundleDir = strings.TrimSpace(config.Learning.GrammarBundleDir)
	config.Learning.ContentSource = strings.ToLower(strings.TrimSpace(config.Learning.ContentSource))
	if config.Learning.ContentSource == "" {
		config.Learning.ContentSource = "bundle"
	}
	if config.Learning.ContentSource != "bundle" && config.Learning.ContentSource != "db" {
		return nil, fmt.Errorf("CONTENT_SOURCE must be bundle or db, got %q", config.Learning.ContentSource)
	}

	if config.Learning.GrammarBundleDir != "" {
		if err := validateGrammarBundleDir(config.Learning.GrammarBundleDir); err != nil {
			return nil, fmt.Errorf("learning config: %w", err)
		}
	} else if err := grammarbundle.ValidateEmbeddedBundleID(config.Learning.GrammarBundleID); err != nil {
		return nil, fmt.Errorf("learning config: %w", err)
	}

	if strings.TrimSpace(config.Training.PromptFile) == "" {
		config.Training.PromptFile = defaultTrainingPromptFile(config.Learning.TargetLang)
	}

	// Validate required fields before loading prompt file so missing env yields clear errors
	if config.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if config.AI.URL == "" {
		return nil, fmt.Errorf("ai url is required")
	}
	if config.AI.APIKey == "" {
		return nil, fmt.Errorf("ai api key is required")
	}
	config.AI.RequestTimeout = strings.TrimSpace(config.AI.RequestTimeout)
	if config.AI.RequestTimeout != "" {
		if _, err := time.ParseDuration(config.AI.RequestTimeout); err != nil {
			return nil, fmt.Errorf("AI_REQUEST_TIMEOUT: invalid duration %q (examples: 120s, 3m): %w", config.AI.RequestTimeout, err)
		}
	}
	if config.WebApp.JWTSecret == "" && config.WebApp.SessionSecret == "" {
		return nil, fmt.Errorf("JWT secret is required (set WEBAPP_JWT_SECRET or WEBAPP_SESSION_SECRET)")
	}

	// Load prompt from file if specified
	if config.AI.PromptFile != "" {
		promptFromFile, err := loadPromptFromFile(config.AI.PromptFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load prompt from file: %w", err)
		}
		config.AI.Prompt = promptFromFile
	}

	// Process newlines in bot messages
	config.Bot.StartMessage = processNewlines(config.Bot.StartMessage)
	config.Bot.HelpMessage = processNewlines(config.Bot.HelpMessage)
	config.Bot.UnknownCommandMessage = processNewlines(config.Bot.UnknownCommandMessage)
	config.Bot.ErrorMessage = processNewlines(config.Bot.ErrorMessage)
	config.Bot.EmptyMessage = processNewlines(config.Bot.EmptyMessage)

	// AI prompt required (from env or from file)
	if config.AI.Prompt == "" {
		return nil, fmt.Errorf("ai prompt is required (either AI_PROMPT or AI_PROMPT_FILE must be set)")
	}

	config.AI.Prompt = ai.PreparePrompt(config.AI.Prompt, config.Learning.NativeLang, config.Learning.TargetLang, config.Learning.Pair)

	if strings.TrimSpace(config.Speaking.EvalAPIKey) == "" {
		config.Speaking.EvalAPIKey = strings.TrimSpace(config.TTS.APIKey)
	}
	if strings.TrimSpace(config.Speaking.EvalAPIKey) == "" {
		config.Speaking.EvalAPIKey = strings.TrimSpace(config.AI.APIKey)
	}
	if strings.TrimSpace(config.Speaking.EvalBaseURL) == "" {
		config.Speaking.EvalBaseURL = strings.TrimSpace(config.TTS.BaseURL)
	}
	if strings.TrimSpace(config.Speaking.EvalBaseURL) == "" {
		config.Speaking.EvalBaseURL = strings.TrimSpace(config.AI.URL)
	}

	return &config, nil
}

// loadPromptFromFile loads prompt content from a file
func loadPromptFromFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", filePath, err)
	}
	return strings.TrimSpace(string(content)), nil
}

func validateGrammarBundleDir(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("GRAMMAR_BUNDLE_DIR: %w", err)
	}
	st, err := os.Stat(filepath.Join(abs, "sections.json"))
	if err != nil {
		return fmt.Errorf("GRAMMAR_BUNDLE_DIR %q: %w", abs, err)
	}
	if st.IsDir() {
		return fmt.Errorf("GRAMMAR_BUNDLE_DIR %q: sections.json is not a file", abs)
	}
	return nil
}

// GetEnv returns environment variable value or default
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// processNewlines converts \n to actual newlines in bot messages
func processNewlines(text string) string {
	return strings.ReplaceAll(text, "\\n", "\n")
}

func defaultTrainingPromptFile(targetLang string) string {
	if strings.EqualFold(strings.TrimSpace(targetLang), "es") {
		return "prompts/training-card-ru-es.txt"
	}
	return "prompts/training-card-ru-en.txt"
}

func envKeysSnapshot() map[string]struct{} {
	out := make(map[string]struct{})
	for _, kv := range os.Environ() {
		if eq := strings.IndexByte(kv, '='); eq > 0 {
			out[kv[:eq]] = struct{}{}
		}
	}
	return out
}

// loadDotEnvFileWithoutOverridingInitialEnv loads dotenv file values and applies them
// only for keys that were not present in the process environment at startup.
// It allows language-specific .env.<lang> to override base .env values while respecting
// explicit shell/CI environment variables.
func loadDotEnvFileWithoutOverridingInitialEnv(path string, initialEnv map[string]struct{}) error {
	values, err := godotenv.Read(path)
	if err != nil {
		return err
	}
	for k, v := range values {
		if _, hadInitially := initialEnv[k]; hadInitially {
			continue
		}
		_ = os.Setenv(k, v)
	}
	return nil
}
