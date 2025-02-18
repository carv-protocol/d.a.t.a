package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/tools/wallet"
	"github.com/spf13/viper"
)

type Config struct {
	Settings struct {
		ShutdownTimeout int `mapstructure:"shutdown_timeout"`
	} `mapstructure:"settings"`

	Character struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"character"`

	Database struct {
		Type string `mapstructure:"type"`
		Path string `mapstructure:"path"`
	} `mapstructure:"database"`

	llm.LLMConfig `mapstructure:"llm_config"`

	Data struct {
		carv.CarvConfig `mapstructure:"carv"`
	} `mapstructure:"data"`

	Social struct {
		clients.TwitterConfig  `mapstructure:"twitter"`
		clients.DiscordConfig  `mapstructure:"discord"`
		clients.TelegramConfig `mapstructure:"telegram"`
	} `mapstructure:"social"`

	Token struct {
		Network      string `mapstructure:"network"`
		Ticker       string `mapstructure:"ticker"`
		ContractAddr string `mapstructure:"contract_addr"`
	} `mapstructure:"token"`

	Wallet wallet.Config `mapstructure:"wallet"`

	Plugin struct {
		Plugins map[string]PluginConfig `mapstructure:"plugins"`
	} `mapstructure:"plugin"`
}

type PluginConfig struct {
	Name         string                 `mapstructure:"name"`
	Enabled      bool                   `mapstructure:"enabled"`
	Version      string                 `mapstructure:"version"`
	Author       string                 `mapstructure:"author"`
	Description  string                 `mapstructure:"description"`
	Dependencies []string               `mapstructure:"dependencies"`
	Options      map[string]interface{} `mapstructure:"options"`
}

func setDefaultConfig() {
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./data/data.db")
	viper.SetDefault("social.twitter.mode", "api")
	viper.SetDefault("llm_config.provider", "openai")
	viper.SetDefault("llm_config.base_url", "https://api.openai.com/v1")
	viper.SetDefault("llm_config.model", "gpt-4o")                // Default model for OpenAI
	viper.SetDefault("shutdown_timeout", 30)                      // shutdown timeout in seconds
	viper.SetDefault("plugin.plugins", map[string]PluginConfig{}) // Default empty plugins map
}

func loadEnvConfig() error {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./src")

	if err := viper.MergeInConfig(); err != nil {
		return fmt.Errorf("error reading .env: %w", err)
	}

	// Map environment variables to config
	envMappings := map[string]string{
		"LLM_PROVIDER":           "llm_config.provider",
		"LLM_API_KEY":            "llm_config.api_key",
		"TWITTER_BEARER_TOKEN":   "social.twitter.bearer_token",
		"TWITTER_API_KEY":        "social.twitter.api_key",
		"TWITTER_API_KEY_SECRET": "social.twitter.api_key_secret",
		"TWITTER_ACCESS_TOKEN":   "social.twitter.access_token",
		"TWITTER_TOKEN_SECRET":   "social.twitter.token_secret",
		"TWITTER_MONITOR_WINDOW": "social.twitter.monitor_window",
		"DISCORD_API_TOKEN":      "social.discord.api_token",
		"TELEGRAM_BOT_TOKEN":     "social.telegram.bot_token",
		"CARV_DATA_BASE_URL":     "data.carv.base_url",
		"CARV_DATA_API_KEY":      "data.carv.api_key",
		"WALLET_PRIVATE_KEY":     "wallet.private_key",
	}

	// override config values with environment variables
	for env, conf := range envMappings {
		viper.Set(conf, viper.Get(env))
	}

	// Set provider-specific defaults if not already set
	provider := viper.GetString("llm_config.provider")

	switch provider {
	case "deepseek":
		if !viper.IsSet("llm_config.base_url") {
			viper.Set("llm_config.base_url", "https://api.deepseek.com")
		}
		if !viper.IsSet("llm_config.model") {
			viper.Set("llm_config.model", "deepseek-chat")
		}
	case "openai":
		if !viper.IsSet("llm_config.base_url") {
			viper.Set("llm_config.base_url", "https://api.openai.com/v1")
		}
		if !viper.IsSet("llm_config.model") {
			viper.Set("llm_config.model", "gpt-3.5-turbo")
		}
	}

	// Special handling for Telegram channel ID
	if channelID := viper.GetString("TELEGRAM_CHANNEL_ID"); channelID != "" {
		viper.Set("social.telegram.channel_id", strings.Trim(channelID, "${}"))
	}

	return nil
}

// loadConfig loads and validates the application configuration
func loadConfig() (*Config, error) {
	// Set default configuration paths
	configPaths := []string{
		"./src/config",
		".",
		"./config",
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}

	// Set defaults
	setDefaultConfig()

	// Load main config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Printf("No config file found, using defaults")
	} else {
		log.Printf("Loaded config from file: %s", viper.ConfigFileUsed())
	}

	// Load environment variables
	if err := loadEnvConfig(); err != nil {
		return nil, fmt.Errorf("error loading environment config: %w", err)
	}

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := validateConfig(&conf); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &conf, nil
}

func validateConfig(conf *Config) error {
	if conf.LLMConfig.APIKey == "" {
		return fmt.Errorf("%w: missing API key", ErrInvalidLLMConfig)
	}
	if conf.LLMConfig.Provider == "" {
		return fmt.Errorf("%w: missing provider", ErrInvalidLLMConfig)
	}
	if conf.LLMConfig.Model == "" {
		return fmt.Errorf("%w: missing model", ErrInvalidLLMConfig)
	}
	if conf.Database.Path == "" {
		return ErrInvalidDBConfig
	}
	return nil
}
