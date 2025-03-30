package conf

import (
	"fmt"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	TwitterModeAPI     TwitterMode = "api"
	TwitterModeScraper TwitterMode = "scraper"

	ThoughtStepTypeTask   ThoughtStepType = "task"
	ThoughtStepTypeAction ThoughtStepType = "action"

	DatabasePostgres DatabaseType = "postgres"
	DatabaseSqlite   DatabaseType = "sqlite"
)

var (
	ErrInvalidLLMConfig = errors.New("invalid LLM configuration")
	ErrInvalidDBConfig  = errors.New("invalid database configuration")
)

type (
	TwitterMode     string
	ThoughtStepType string
	DatabaseType    string
)

type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
}

type CarvConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

type TwitterConfig struct {
	Mode          TwitterMode `mapstructure:"mode"`     // Mode of operation: "api" or "scraper"
	Username      string      `mapstructure:"username"` // Twitter username
	Password      string      `mapstructure:"password"` // Twitter password
	APIKey        string      `mapstructure:"api_key"`
	APIKeySecret  string      `mapstructure:"api_key_secret"`
	AccessToken   string      `mapstructure:"access_token"`
	TokenSecret   string      `mapstructure:"token_secret"`
	MonitorWindow int         `mapstructure:"monitor_window"` // Duration in minutes, e.g. 20
}

type DiscordConfig struct {
	APIToken string `mapstructure:"api_token"`
}

type TelegramConfig struct {
	Token     string `mapstructure:"bot_token"`  // Bot token from BotFather
	ChannelID int64  `mapstructure:"channel_id"` // Default channel ID for broadcasts
	Debug     bool   `mapstructure:"debug"`      // Enable debug mode
}

type PromptTemplates struct {
	System struct {
		BaseTemplate string            `mapstructure:"base_template"`
		InfoFormat   map[string]string `mapstructure:"info_format"`
	} `mapstructure:"system"`

	Message struct {
		Analysis string `mapstructure:"analysis"`
		Action   string `mapstructure:"action"`
	} `mapstructure:"message"`

	ThoughtSteps map[ThoughtStepType]struct {
		Initial     string `mapstructure:"initial"`
		Exploration string `mapstructure:"exploration"`
		Analysis    string `mapstructure:"analysis"`
		Reconsider  string `mapstructure:"reconsider"`
		Refinement  string `mapstructure:"refinement"`
		Concrete    string `mapstructure:"concrete"`
	} `mapstructure:"thought_steps"`
}

type PlatformGovernanceConfig struct {
	AdminIDs        []string `mapstructure:"admin_ids"`
	MinTokenBalance float64  `mapstructure:"min_token_balance"`
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

type Character struct {
	Name string `mapstructure:"name"`
	Path string `mapstructure:"path"`
}

type Config struct {
	Settings struct {
		ShutdownTimeout int `mapstructure:"shutdown_timeout"`
	} `mapstructure:"settings"`

	Character `mapstructure:"character"`

	Database struct {
		Type DatabaseType `mapstructure:"type"`
		Path string       `mapstructure:"path"`
	} `mapstructure:"database"`

	LLMConfig `mapstructure:"llm_config"`

	Data struct {
		CarvConfig `mapstructure:"carv"`
	} `mapstructure:"data"`

	Social struct {
		TwitterConfig  `mapstructure:"twitter"`
		DiscordConfig  `mapstructure:"discord"`
		TelegramConfig `mapstructure:"telegram"`
	} `mapstructure:"social"`

	Governance struct {
		DefaultAdminIDs        []string                  `mapstructure:"admin_ids"`
		DefaultMinTokenBalance float64                   `mapstructure:"min_token_balance"`
		Discord                *PlatformGovernanceConfig `mapstructure:"discord"`
		Twitter                *PlatformGovernanceConfig `mapstructure:"twitter"`
		Telegram               *PlatformGovernanceConfig `mapstructure:"telegram"`
	} `mapstructure:"governance"`

	Token struct {
		Network      string `mapstructure:"network"`
		Ticker       string `mapstructure:"ticker"`
		ContractAddr string `mapstructure:"contract_addr"`
	} `mapstructure:"token"`

	Web struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"web"`

	UserTemplates    *PromptTemplates `mapstructure:"user_templates"`
	DefaultTemplates *PromptTemplates `mapstructure:"default_templates"`

	Plugins map[string]PluginConfig `mapstructure:"plugins"`
}

// LoadConfig loads and validates the application configuration
func LoadConfig(confPath string) (*Config, error) {
	setDefaultConfig()

	if err := loadYamlConfig(confPath); err != nil {
		return nil, fmt.Errorf("error load yaml config: %w", err)
	}

	if err := loadEnvConfig(confPath); err != nil {
		return nil, fmt.Errorf("error load env config: %w", err)
	}

	fillDefaultOptions()

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := validateConfig(&conf, confPath); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &conf, nil
}

func loadYamlConfig(confPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(confPath)

	// Load main config
	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return fmt.Errorf("error reading config file: %w", err)
		}
		logger.GetLogger().Infoln("No config file found, using defaults")
	} else {
		logger.GetLogger().Infof("Loaded config from file: %s", viper.ConfigFileUsed())
	}

	return nil
}

func loadEnvConfig(confPath string) error {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(confPath)

	if err := viper.MergeInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil
		}
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

	return nil
}

func setDefaultConfig() {
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./data/data.db")
	viper.SetDefault("llm_config.provider", "openai")
	viper.SetDefault("llm_config.base_url", "https://api.openai.com/v1")
	viper.SetDefault("llm_config.model", "gpt-4o")                // Default model for OpenAI
	viper.SetDefault("shutdown_timeout", 30)                      // shutdown timeout in seconds
	viper.SetDefault("plugin.plugins", map[string]PluginConfig{}) // Default empty plugins map
}

func fillDefaultOptions() {
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

// loadDefaultTemplates loads templates from default_templates.yaml
func loadDefaultTemplates(configPath string) (*PromptTemplates, error) {
	defaultViper := viper.New()
	defaultViper.SetConfigName("default_templates")
	defaultViper.SetConfigType("yaml")
	defaultViper.AddConfigPath(configPath)

	if err := defaultViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading default templates: %w", err)
	}

	var defaultTemplates PromptTemplates
	if err := defaultViper.UnmarshalKey("default_templates", &defaultTemplates); err != nil {
		return nil, fmt.Errorf("error unmarshaling default templates: %w", err)
	}

	return &defaultTemplates, nil
}

func validateConfig(conf *Config, confPath string) error {
	// Check if user templates are defined, if not load default templates
	if conf.UserTemplates == nil {
		logger.GetLogger().Infoln("User templates not defined, loading default templates")
		defaultTemplates, err := loadDefaultTemplates(confPath)
		if err != nil {
			return fmt.Errorf("failed to load default templates: %w", err)
		}
		conf.DefaultTemplates = defaultTemplates
		conf.UserTemplates = conf.DefaultTemplates
	} else {
		logger.GetLogger().Infoln("Using user-defined templates")
	}

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
