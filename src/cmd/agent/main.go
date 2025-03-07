package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/commands"
	"github.com/carv-protocol/d.a.t.a/src/internal/conf"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/internal/social"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	dataPlugin "github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a"
	"github.com/carv-protocol/d.a.t.a/src/web"

	"github.com/google/uuid"
)

var FlagConfig string

type pluginFactory func(llmClient llm.Client, config *plugins.Config) (plugins.Plugin, error)

func init() {
	flag.StringVar(&FlagConfig, "conf", "./src/config", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	config, err := conf.LoadConfig(FlagConfig)
	if err != nil {
		logger.GetLogger().Fatalf("Failed to load config: %v", err)
	}

	// Initialize components
	agent, err := initializeAgent(ctx, config)
	if err != nil {
		logger.GetLogger().Fatalf("Failed to initialize agent: %v", err)
	}

	// Start the agent
	if err = agent.Start(); err != nil {
		logger.GetLogger().Fatalf("Failed to start agent: %v", err)
	}

	web.Start(config.Web.Port)

	// Wait for shutdown signal
	<-handleShutdown(ctx, agent, config.Settings.ShutdownTimeout)
}

func initializeAgent(ctx context.Context, config *conf.Config) (*core.Agent, error) {
	// Setup database
	var store database.Store
	switch config.Database.Type {
	case conf.DatabasePostgres:
		store = adapters.NewPostgresStore(config.Database.Path)
	case conf.DatabaseSqlite:
		store = adapters.NewSQLiteStore(config.Database.Path)
	default:
		return nil, fmt.Errorf("unknown database type: %s", config.Database.Type)
	}

	if err := store.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize components
	llmClient := llm.NewClient((*conf.LLMConfig)(&config.LLMConfig))
	carvClient := carv.NewClient(config.Data.CarvConfig.APIKey, config.Data.CarvConfig.BaseURL)
	memoryManager, err := memory.NewManager(store)
	if err != nil {
		return nil, fmt.Errorf("failed to new manager: %w", err)
	}
	tokenManager := token.NewTokenManager(carvClient, &core.TokenInfo{
		Network:      config.Token.Network,
		Ticker:       config.Token.Ticker,
		ContractAddr: config.Token.ContractAddr,
	})
	stakeholderManager := token.NewStakeholderManager(memoryManager)

	// Initialize governance registry
	governanceRegistry, err := initializeGovernance(config, tokenManager)

	// Load character
	character, err := characters.NewCharacter(config.Character, store)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %w", err)
	}

	// Initialize plugins
	pluginRegistry := initializePlugins(config)

	// Initialize command manager
	commandManager := initializeCommands(config, governanceRegistry, llmClient, character)

	promptTemplates := config.UserTemplates
	if config.UserTemplates == nil {
		promptTemplates = config.DefaultTemplates
	}

	// Create agent
	agentConfig := core.AgentConfig{
		ID:              uuid.New(),
		Character:       character,
		CommandRegistry: commandManager,
		LLMClient:       llmClient,
		Model:           config.LLMConfig.Model,
		Stakeholders:    stakeholderManager,
		SocialClient: social.NewSocialClient(
			&config.Social.TwitterConfig,
			&config.Social.DiscordConfig,
			&config.Social.TelegramConfig,
		),
		PromptTemplates: promptTemplates,
		TokenManager:    tokenManager,
		PluginRegistry:  pluginRegistry,
	}

	agent, err := core.NewAgent(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}

func initializePlugins(config *conf.Config) *plugins.Registry {
	registry := plugins.NewPluginRegistry()

	// Initialize built-in plugins
	builtinPlugins := map[string]pluginFactory{
		"d.a.t.a": dataPlugin.NewPlugin,
	}

	// Load plugins from configuration
	for name, pluginConfig := range config.Plugins {
		// Skip disabled plugins
		if !pluginConfig.Enabled {
			continue
		}

		// Check dependencies
		if err := checkPluginDependencies(pluginConfig, config.Plugins); err != nil {
			logger.GetLogger().Errorf("Failed to load plugin %s: %v", name, err)
			continue
		}

		// Get plugin factory
		factory, exists := builtinPlugins[name]
		if !exists {
			logger.GetLogger().Errorf("Plugin %s not found in built-in plugins", name)
			continue
		}

		// Create plugin instance
		plugin, err := factory(llm.NewClient((*conf.LLMConfig)(&config.LLMConfig)), &plugins.Config{
			Name:        name,
			Description: pluginConfig.Description,
			Options:     pluginConfig.Options,
		})

		// Register plugin
		if err != nil {
			logger.GetLogger().Errorf("Failed to register plugin %s: %v", name, err)
			continue
		}

		if err = registry.Register(plugin); err != nil {
			logger.GetLogger().Errorf("Failed to register plugin %s: %v", name, err)
		}
	}

	return registry
}

// initializeGovernance initializes the governance registry with platform-specific configurations
func initializeGovernance(config *conf.Config, tokenManager *token.TokenManager) (governance.Registry, error) {
	// Initialize governance registry
	governanceRegistry := governance.NewMemoryRegistry(tokenManager)

	// Set up admin configuration from config
	adminConfig := &governance.AdminConfig{
		Admins:          config.Governance.DefaultAdminIDs,
		MinTokenBalance: config.Governance.DefaultMinTokenBalance,
		PlatformConfigs: make(map[string]*governance.PlatformConfig),
	}

	// Add platform-specific configurations
	if config.Governance.Discord != nil {
		adminConfig.PlatformConfigs["discord"] = &governance.PlatformConfig{
			Admins:          config.Governance.Discord.AdminIDs,
			MinTokenBalance: config.Governance.Discord.MinTokenBalance,
		}
	}

	if config.Governance.Twitter != nil {
		adminConfig.PlatformConfigs["twitter"] = &governance.PlatformConfig{
			Admins:          config.Governance.Twitter.AdminIDs,
			MinTokenBalance: config.Governance.Twitter.MinTokenBalance,
		}
	}

	if config.Governance.Telegram != nil {
		adminConfig.PlatformConfigs["telegram"] = &governance.PlatformConfig{
			Admins:          config.Governance.Telegram.AdminIDs,
			MinTokenBalance: config.Governance.Telegram.MinTokenBalance,
		}
	}

	// Update the governance registry with the admin configuration
	err := governanceRegistry.UpdateAdminConfig(adminConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin config: %v", err)
	}

	return governanceRegistry, nil
}

func initializeCommands(config *conf.Config, governanceRegistry governance.Registry, llmClient llm.Client, character *characters.Character) *commands.Registry {
	commandManager := commands.NewRegistry()
	commands.InitializeCommands(
		commandManager,
		governanceRegistry,
		llmClient,
		config.LLMConfig.Model,
		social.NewSocialClient(
			nil,
			&config.Social.DiscordConfig,
			nil,
		),
		character,
	)
	return commandManager
}

// checkPluginDependencies verifies that all plugin dependencies are enabled
func checkPluginDependencies(config conf.PluginConfig, plugins map[string]conf.PluginConfig) error {
	for _, dep := range config.Dependencies {
		depConfig, exists := plugins[dep]
		if !exists {
			return fmt.Errorf("dependency %s not found", dep)
		}
		if !depConfig.Enabled {
			return fmt.Errorf("dependency %s is disabled", dep)
		}
	}
	return nil
}

func handleShutdown(ctx context.Context, agent *core.Agent, timeoutSeconds int) chan struct{} {
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.GetLogger().Infoln("Shutdown signal received, initiating graceful shutdown...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer cancel()

		web.Stop()

		if err := agent.Shutdown(shutdownCtx); err != nil {
			logger.GetLogger().Errorf("Error during shutdown: %v", err)
		}

		close(done)
	}()

	return done
}
