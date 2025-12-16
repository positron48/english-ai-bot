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
	PromptFile              string `mapstructure:"prompt_file"`
	CircuitBreakerThreshold int    `mapstructure:"circuit_breaker_threshold"`
	CircuitBreakerAutoReset int    `mapstructure:"circuit_breaker_auto_reset_hours"`
	OptionsDelayMS          int    `mapstructure:"options_delay_ms"`
}

// AdminConfig holds admin configuration
type AdminConfig struct {
	TelegramID int64 `mapstructure:"telegram_id"`
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
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("ai.model", "gpt-3.5-turbo")
	viper.SetDefault("database.path", "./data/words.db")
	
	// Training defaults
	viper.SetDefault("training.worker_enabled", true)
	viper.SetDefault("training.worker_interval", "30s")
	viper.SetDefault("training.worker_batch_size", 5)
	viper.SetDefault("training.prompt_file", "prompts/training-card-generator.txt")
	viper.SetDefault("training.circuit_breaker_threshold", 5)
	viper.SetDefault("training.circuit_breaker_auto_reset_hours", 24)
	viper.SetDefault("training.options_delay_ms", 5000)
	
	// Admin defaults
	viper.SetDefault("admin.telegram_id", 0)

	// Bot message defaults
	viper.SetDefault("bot.start_message", "🤖 Hello! I'm a universal AI assistant.\n\n💡 Just send me a message and I'll help you with any questions!\n\nUse /help for additional information.")
	viper.SetDefault("bot.help_message", "📚 AI Assistant Help:\n\n💬 **Any message** → Get a smart response:\n• Answer questions\n• Help with tasks\n• Explanations and advice\n• Creative ideas\n\n🔧 **Available commands:**\n• /start - Start working with the bot\n• /help - Show this help\n\n💡 Just send text - I'll help right away!")
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
	_ = viper.BindEnv("training.prompt_file", "TRAINING_PROMPT_FILE")
	_ = viper.BindEnv("training.circuit_breaker_threshold", "CIRCUIT_BREAKER_THRESHOLD")
	_ = viper.BindEnv("training.circuit_breaker_auto_reset_hours", "CIRCUIT_BREAKER_AUTO_RESET_HOURS")
	_ = viper.BindEnv("training.options_delay_ms", "TRAINING_OPTIONS_DELAY_MS")
	_ = viper.BindEnv("admin.telegram_id", "ADMIN_TELEGRAM_ID")

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
	if config.Telegram.Token == "" {
		return nil, fmt.Errorf("telegram token is required")
	}
	if config.AI.URL == "" {
		return nil, fmt.Errorf("ai url is required")
	}
	if config.AI.APIKey == "" {
		return nil, fmt.Errorf("ai api key is required")
	}
	if config.AI.Prompt == "" {
		return nil, fmt.Errorf("ai prompt is required (either AI_PROMPT or AI_PROMPT_FILE must be set)")
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
