package main

import (
	"fmt"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/tools/wallet"
	"github.com/spf13/viper"
	"log"
	"strings"
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

	Web struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"web"`

	UserTemplates    *core.PromptTemplates `mapstructure:"user_templates"`
	DefaultTemplates *core.PromptTemplates `mapstructure:"default_templates"`

	Plugin struct {
		Plugins map[string]PluginConfig `mapstructure:"plugins"`
	} `mapstructure:"plugin"`

	Limitation struct {
		Enable            bool    `mapstructure:"enable"`
		MinTokenBalance   float64 `mapstructure:"min_token_balance"`
		FrequencyFactor   float64 `mapstructure:"frequency_factor"`
		FrequencyDuration int64   `mapstructure:"frequency_duration"`
		TransferAmount    float64 `mapstructure:"transfer_amount"`
	} `mapstructure:"limitation"`
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
	viper.SetDefault("llm_config.provider", "openai")
	viper.SetDefault("llm_config.base_url", "https://api.openai.com/v1")
	viper.SetDefault("llm_config.model", "gpt-4o")                // Default model for OpenAI
	viper.SetDefault("shutdown_timeout", 30)                      // shutdown timeout in seconds
	viper.SetDefault("plugin.plugins", map[string]PluginConfig{}) // Default empty plugins map

	switch viper.GetString("llm_config.provider") {
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
}

// loadConfig loads and validates the application configuration
func loadConfig(confPath string) (*Config, error) {
	// Set default configuration paths
	configPaths := []string{confPath}

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

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Check if user templates are defined, if not load default templates
	if conf.UserTemplates == nil {
		log.Printf("User templates not defined, loading default templates")
		defaultTemplates, err := loadDefaultTemplates(configPaths)
		if err != nil {
			return nil, fmt.Errorf("failed to load default templates: %w", err)
		}
		conf.DefaultTemplates = defaultTemplates
		conf.UserTemplates = conf.DefaultTemplates
	} else {
		log.Printf("Using user-defined templates")
	}

	// Validate config
	if err := validateConfig(&conf); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &conf, nil
}

// loadDefaultTemplates loads templates from default_templates.yaml
func loadDefaultTemplates(configPaths []string) (*core.PromptTemplates, error) {
	defaultViper := viper.New()
	defaultViper.SetConfigName("default_templates")
	defaultViper.SetConfigType("yaml")

	for _, path := range configPaths {
		defaultViper.AddConfigPath(path)
	}

	if err := defaultViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading default templates: %w", err)
	}

	var defaultTemplates core.PromptTemplates
	if err := defaultViper.UnmarshalKey("default_templates", &defaultTemplates); err != nil {
		return nil, fmt.Errorf("error unmarshaling default templates: %w", err)
	}

	return &defaultTemplates, nil
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
	if conf.DefaultTemplates == nil && conf.UserTemplates == nil {
		return fmt.Errorf("missing prompt templates")
	}

	return nil
}
