package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	Server   ServerConfig   `mapstructure:"server"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	AI       AIConfig       `mapstructure:"ai"`
	Bot      BotConfig      `mapstructure:"bot"`
	Database DatabaseConfig `mapstructure:"database"`
	Training TrainingConfig `mapstructure:"training"`
	Admin    AdminConfig    `mapstructure:"admin"`
	WebApp   WebAppConfig   `mapstructure:"webapp"`
}

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	Token          string `mapstructure:"token"`
	APIBaseURL     string `mapstructure:"api_base_url"`
	Debug          bool   `mapstructure:"debug"`
	UpdatesTimeout int    `mapstructure:"updates_timeout"`
	WebhookEnable  bool   `mapstructure:"webhook_enable"`
	WebhookURL     string `mapstructure:"webhook_url"`
	WebhookDomain  string `mapstructure:"webhook_domain"`
	WebhookPath    string `mapstructure:"webhook_path"`
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
	URL        string `mapstructure:"url"`
	Model      string `mapstructure:"model"`
	ModelHigh  string `mapstructure:"model_high"`
	APIKey     string `mapstructure:"api_key"`
	Prompt     string `mapstructure:"prompt"`
	PromptFile string `mapstructure:"prompt_file"`
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
	Path string `mapstructure:"path"`
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
}

// AdminConfig holds admin configuration
type AdminConfig struct {
	TelegramID int64 `mapstructure:"telegram_id"`
}

// WebAppConfig holds web app configuration
type WebAppConfig struct {
	PublicURL         string `mapstructure:"public_url"`
	SessionSecret     string `mapstructure:"session_secret"`
	JWTSecret         string `mapstructure:"jwt_secret"`
	OTPTTLSeconds     int    `mapstructure:"otp_ttl_seconds"`
	SessionTTLHours   int    `mapstructure:"session_ttl_hours"`
	JWTTTLHours       int    `mapstructure:"jwt_ttl_hours"`
	RefreshTTLHours   int    `mapstructure:"refresh_ttl_hours"`
	ViteDevServerURL  string `mapstructure:"vite_dev_server_url"`
	
	// Rate limiting configuration
	RateLimitAuthRequestOTPPerIP        int `mapstructure:"rate_limit_auth_request_otp_per_ip"`
	RateLimitAuthRequestOTPPerIPUser    int `mapstructure:"rate_limit_auth_request_otp_per_ip_user"`
	RateLimitAuthOTPPerIP               int `mapstructure:"rate_limit_auth_otp_per_ip"`
	RateLimitAuthOTPPerIPUser           int `mapstructure:"rate_limit_auth_otp_per_ip_user"`
	RateLimitAuthTelegramUnsafePerIP   int `mapstructure:"rate_limit_auth_telegram_unsafe_per_ip"`
	RateLimitAuthTelegramUnsafePerIPUser int `mapstructure:"rate_limit_auth_telegram_unsafe_per_ip_user"`
	RateLimitAuthTelegramPerIP         int `mapstructure:"rate_limit_auth_telegram_per_ip"`
	RateLimitAuthRefreshPerIP          int `mapstructure:"rate_limit_auth_refresh_per_ip"`
	RateLimitAppAPIPerUser             int `mapstructure:"rate_limit_app_api_per_user"`
	RateLimitAppChatPerUser            int `mapstructure:"rate_limit_app_chat_per_user"`
	RateLimitWindowMinutes             int `mapstructure:"rate_limit_window_minutes"`
	RateLimitBurstMultiplier           int `mapstructure:"rate_limit_burst_multiplier"`
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, ignore error
		_ = err
	}

	// Set default values
	viper.SetDefault("telegram.debug", false)
	viper.SetDefault("telegram.updates_timeout", 30)
	viper.SetDefault("telegram.webhook_enable", false)
	viper.SetDefault("telegram.webhook_path", "/webhook")
	viper.SetDefault("server.address", ":8184")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("ai.model", "gpt-3.5-turbo")
	viper.SetDefault("database.path", "./data/words.db")
	
	// Training defaults
	viper.SetDefault("training.worker_enabled", true)
	viper.SetDefault("training.worker_interval", "30s")
	viper.SetDefault("training.worker_batch_size", 5)
	viper.SetDefault("training.llm_workers", 4)
	viper.SetDefault("training.prompt_file", "prompts/training-card-generator.txt")
	viper.SetDefault("training.circuit_breaker_threshold", 5)
	viper.SetDefault("training.circuit_breaker_auto_reset_hours", 24)
	viper.SetDefault("training.options_delay_ms", 5000)
	viper.SetDefault("training.wrong_answer_delay_seconds", 5)
	
	// Admin defaults
	viper.SetDefault("admin.telegram_id", 0)
	
	// WebApp defaults
	viper.SetDefault("webapp.public_url", "")
	viper.SetDefault("webapp.session_secret", "")
	viper.SetDefault("webapp.jwt_secret", "")
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

	// Bot message defaults
	viper.SetDefault("bot.start_message", "🤖 Hello! I'm a universal AI assistant.\n\n💡 Just send me a message and I'll help you with any questions!\n\nUse /help for additional information.")
	viper.SetDefault("bot.help_message", "📚 Помощь по использованию бота-преподавателя английского:\n\n🔤 **Одно слово** → Карточка слова с:\n• Частью речи\n• Транскрипцией IPA\n• Определением на русском\n• Примерами использования\n• Формами неправильных глаголов\n\n📝 **Английский текст** → Исправления:\n• Поиск ошибок (орфография, грамматика, пунктуация)\n• Подробные объяснения\n• Исправленная версия\n\n🇷🇺 **Русский текст** → Перевод:\n• Естественный перевод на английский\n• Анализ сложных фраз\n• Сохранение тона и стиля\n\n🔧 **Доступные команды:**\n• /start - Начать работу с ботом\n• /help - Показать эту справку\n• /unsubscribe - Отписаться от уведомлений\n• /notification [daily|never|N] - Настроить периодичность уведомлений\n  - daily - ежедневно\n  - never - никогда\n  - N - каждые N дней (например, /notification 3)\n\n💬 Просто отправьте текст или слово - я сразу помогу!")
	viper.SetDefault("bot.unknown_command_message", "❓ Unknown command. Use /help to get information about bot capabilities.")
	viper.SetDefault("bot.error_message", "Sorry, an error occurred while processing your message. Please try again.")
	viper.SetDefault("bot.empty_message", "Please send a text message.")

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicitly bind environment variables to viper keys
	_ = viper.BindEnv("telegram.token", "TELEGRAM_TOKEN")
	_ = viper.BindEnv("telegram.api_base_url", "TELEGRAM_API_BASE_URL")
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
	_ = viper.BindEnv("bot.start_message", "BOT_START_MESSAGE")
	_ = viper.BindEnv("bot.help_message", "BOT_HELP_MESSAGE")
	_ = viper.BindEnv("bot.unknown_command_message", "BOT_UNKNOWN_COMMAND_MESSAGE")
	_ = viper.BindEnv("bot.error_message", "BOT_ERROR_MESSAGE")
	_ = viper.BindEnv("bot.empty_message", "BOT_EMPTY_MESSAGE")
	_ = viper.BindEnv("database.path", "DATABASE_PATH")
	_ = viper.BindEnv("training.worker_enabled", "TRAINING_WORKER_ENABLED")
	_ = viper.BindEnv("training.worker_interval", "TRAINING_WORKER_INTERVAL")
	_ = viper.BindEnv("training.worker_batch_size", "TRAINING_WORKER_BATCH_SIZE")
	_ = viper.BindEnv("training.llm_workers", "TRAINING_LLM_WORKERS")
	_ = viper.BindEnv("training.prompt_file", "TRAINING_PROMPT_FILE")
	_ = viper.BindEnv("training.circuit_breaker_threshold", "CIRCUIT_BREAKER_THRESHOLD")
	_ = viper.BindEnv("training.circuit_breaker_auto_reset_hours", "CIRCUIT_BREAKER_AUTO_RESET_HOURS")
	_ = viper.BindEnv("training.options_delay_ms", "TRAINING_OPTIONS_DELAY_MS")
	_ = viper.BindEnv("training.wrong_answer_delay_seconds", "TRAINING_WRONG_ANSWER_DELAY_SECONDS")
	_ = viper.BindEnv("admin.telegram_id", "ADMIN_TELEGRAM_ID")
	_ = viper.BindEnv("webapp.public_url", "WEBAPP_PUBLIC_URL")
	_ = viper.BindEnv("webapp.session_secret", "WEBAPP_SESSION_SECRET")
	_ = viper.BindEnv("webapp.jwt_secret", "WEBAPP_JWT_SECRET")
	_ = viper.BindEnv("webapp.otp_ttl_seconds", "WEBAPP_OTP_TTL_SECONDS")
	_ = viper.BindEnv("webapp.session_ttl_hours", "WEBAPP_SESSION_TTL_HOURS")
	_ = viper.BindEnv("webapp.jwt_ttl_hours", "WEBAPP_JWT_TTL_HOURS")
	_ = viper.BindEnv("webapp.refresh_ttl_hours", "WEBAPP_REFRESH_TTL_HOURS")
	_ = viper.BindEnv("webapp.vite_dev_server_url", "WEBAPP_VITE_DEV_SERVER_URL")
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
	_ = viper.BindEnv("webapp.rate_limit_window_minutes", "WEBAPP_RATE_LIMIT_WINDOW_MINUTES")
	_ = viper.BindEnv("webapp.rate_limit_burst_multiplier", "WEBAPP_RATE_LIMIT_BURST_MULTIPLIER")

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

	// Validate required fields
	// Note: Telegram token is optional - app can work without it (web-only mode)
	if config.AI.URL == "" {
		return nil, fmt.Errorf("ai url is required")
	}
	if config.AI.APIKey == "" {
		return nil, fmt.Errorf("ai api key is required")
	}
	if config.AI.Prompt == "" {
		return nil, fmt.Errorf("ai prompt is required (either AI_PROMPT or AI_PROMPT_FILE must be set)")
	}
	
	// Validate JWT secret (can use session secret as fallback)
	if config.WebApp.JWTSecret == "" && config.WebApp.SessionSecret == "" {
		return nil, fmt.Errorf("JWT secret is required (set WEBAPP_JWT_SECRET or WEBAPP_SESSION_SECRET)")
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
