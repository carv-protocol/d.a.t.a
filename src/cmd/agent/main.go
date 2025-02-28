package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/commands"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/social"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/internal/tools"
	"github.com/carv-protocol/d.a.t.a/src/internal/types"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	pluginCore "github.com/carv-protocol/d.a.t.a/src/plugins/core"
	dataPlugin "github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a"
	customTools "github.com/carv-protocol/d.a.t.a/src/tools"
	dataTool "github.com/carv-protocol/d.a.t.a/src/tools/d.a.t.a"
	"github.com/carv-protocol/d.a.t.a/src/tools/wallet"
	"github.com/carv-protocol/d.a.t.a/src/web"

	"github.com/google/uuid"
)

// Config validation errors
var (
	ErrInvalidLLMConfig = errors.New("invalid LLM configuration")
	ErrInvalidDBConfig  = errors.New("invalid database configuration")
	FlagConfig          string
)

func init() {
	flag.StringVar(&FlagConfig, "conf", "./src/config", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	config, err := loadConfig(FlagConfig)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize components
	agent, err := initializeAgent(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}

	// Start the agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	web.Start(config.Web.Port)

	// Wait for shutdown signal
	<-handleShutdown(ctx, agent, config.Settings.ShutdownTimeout)
}

func initializeAgent(ctx context.Context, config *Config) (*core.Agent, error) {
	// Setup database
	var store database.Store
	switch config.Database.Type {
	case "postgres":
		store = adapters.NewPostgresStore(config.Database.Path)
	case "sqlite":
		store = adapters.NewSQLiteStore(config.Database.Path)
	default:
		return nil, fmt.Errorf("unknown database type: %s", config.Database.Type)
	}

	if err := store.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize components
	llmClient := llm.NewClient((*llm.LLMConfig)(&config.LLMConfig))
	carvClient := carv.NewClient(config.Data.CarvConfig.APIKey, config.Data.CarvConfig.BaseURL)
	memoryManager := memory.NewManager(store)
	tokenManager := token.NewTokenManager(carvClient, &types.TokenInfo{
		Network:      config.Token.Network,
		Ticker:       config.Token.Ticker,
		ContractAddr: config.Token.ContractAddr,
	})
	stakeholderManager := token.NewStakeholderManager(memoryManager)

	// Initialize governance registry
	governanceRegistry, err := initializeGovernance(config, tokenManager)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize governance: %w", err)
	}

	// Load character
	character, err := characters.LoadFromFile(config.Character.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %w", err)
	}

	// Initialize tools
	toolsManager := initializeTools(config)

	// Initialize plugins
	pluginRegistry := initializePlugins(config)

	// Initialize commands
	actionManager, err := registerPlugins(ctx, pluginRegistry, config, governanceRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to register plugins: %w", err)
	}

	// Initialize command manager
	commandManager := initializeCommands(config, governanceRegistry, llmClient)

	promptTemplates := config.UserTemplates
	if config.UserTemplates == nil {
		promptTemplates = config.DefaultTemplates
	}

	// Create agent
	agentConfig := core.AgentConfig{
		ID:           uuid.New(),
		Character:    character,
		LLMClient:    llmClient,
		Model:        config.LLMConfig.Model,
		Stakeholders: stakeholderManager,
		ToolsManager: toolsManager,
		TokenManager: tokenManager,
		SocialClient: social.NewSocialClient(
			&config.Social.TwitterConfig,
			&config.Social.DiscordConfig,
			&config.Social.TelegramConfig,
		),
		ActionManager:   actionManager,
		PluginRegistry:  pluginRegistry,
		CommandManager:  commandManager,
		PromptTemplates: promptTemplates,
		TaskManager:     tasks.NewManager(tasks.NewTaskStore(store)),
	}

	return core.NewAgent(agentConfig)
}

// initializeGovernance initializes the governance registry with platform-specific configurations
func initializeGovernance(config *Config, tokenManager *token.TokenManager) (governance.Registry, error) {
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

func initializeTools(config *Config) *tools.Manager {
	toolsManager := tools.NewManager()

	walletTool, err := wallet.NewWalletTool(&config.Wallet)
	if err != nil {
		log.Fatalf("Failed to create wallet tool: %v", err)
	}

	toolsManager.Register(&customTools.TwitterTool{})
	toolsManager.Register(walletTool)
	toolsManager.Register(&dataTool.CARVDataTool{})

	return toolsManager
}

func initializePlugins(config *Config) *pluginCore.Registry {
	registry := pluginCore.NewRegistry()

	// Initialize built-in plugins
	builtinPlugins := map[string]pluginCore.PluginFactory{
		"d.a.t.a": dataPlugin.NewPlugin,
	}

	// Load plugins from configuration
	for name, pluginConfig := range config.Plugin.Plugins {
		// Skip disabled plugins
		if !pluginConfig.Enabled {
			continue
		}

		// Check dependencies
		if err := checkPluginDependencies(pluginConfig, config.Plugin.Plugins); err != nil {
			log.Printf("Failed to load plugin %s: %v", name, err)
			continue
		}

		// Get plugin factory - case insensitive matching
		var factory pluginCore.PluginFactory
		var exists bool
		for builtinName, builtinFactory := range builtinPlugins {
			if strings.EqualFold(name, builtinName) {
				factory = builtinFactory
				exists = true
				break
			}
		}

		if !exists {
			log.Printf("Plugin %s not found in built-in plugins", name)
			continue
		}

		// Create plugin instance
		plugin := factory(llm.NewClient((*llm.LLMConfig)(&config.LLMConfig)))

		// Verify metadata
		if err := verifyPluginMetadata(plugin, pluginConfig); err != nil {
			log.Printf("Plugin metadata verification failed for %s: %v", name, err)
			continue
		}

		// Register plugin
		if err := registry.Register(plugin); err != nil {
			log.Printf("Failed to register plugin %s: %v", name, err)
			continue
		}
	}

	return registry
}

func initializeCommands(config *Config, governanceRegistry governance.Registry, llmClient llm.Client) *commands.Registry {
	commandManager := commands.NewRegistry()
	commands.InitializeCommands(commandManager, governanceRegistry, llmClient, config.LLMConfig.Model, social.NewSocialClient(
		&config.Social.TwitterConfig,
		&config.Social.DiscordConfig,
		&config.Social.TelegramConfig,
	))
	return commandManager
}

// registerPlugins initializes and starts all plugins, then registers their actions
func registerPlugins(ctx context.Context, pluginRegistry *pluginCore.Registry, config *Config, governanceRegistry governance.Registry) (actions.ActionManager, error) {
	// Initialize each plugin with its own options
	for configName, pluginConfig := range config.Plugin.Plugins {
		if !pluginConfig.Enabled {
			continue
		}

		// Find registered plugin (case insensitive)
		var plugin pluginCore.Plugin
		var found bool
		for _, p := range pluginRegistry.GetPlugins() {
			if strings.EqualFold(p.Name(), configName) {
				plugin = p
				found = true
				break
			}
		}

		if !found {
			log.Printf("Plugin %s is not registered, skipping initialization", configName)
			continue
		}

		// Prepare plugin options (copy original configuration options)
		pluginOptions := make(map[string]interface{})
		for k, v := range pluginConfig.Options {
			pluginOptions[k] = v
		}

		// Add special dependencies for d.a.t.aGovernance
		if strings.EqualFold(plugin.Name(), "d.a.t.aGovernance") {
			pluginOptions["governance_registry"] = governanceRegistry
		}
		// Future can add special dependencies for other plugins here

		// Initialize plugin
		if err := pluginRegistry.InitPlugin(ctx, plugin.Name(), pluginOptions); err != nil {
			return nil, fmt.Errorf("failed to initialize plugin %s: %w", plugin.Name(), err)
		}
	}

	// Start all plugins
	if err := pluginRegistry.StartAll(ctx); err != nil {
		return nil, fmt.Errorf("failed to start plugins: %w", err)
	}

	// Initialize action manager and register actions
	actionManager := actions.NewManager()
	for _, pluginAction := range pluginRegistry.GetActions() {
		log.Printf("Registering action %s", pluginAction.Name())
		adapter := pluginCore.NewActionAdapter(ctx, pluginAction)
		if err := actionManager.Register(adapter); err != nil {
			return nil, fmt.Errorf("failed to register action %s: %w", pluginAction.Name(), err)
		}
	}

	return actionManager, nil
}

// checkPluginDependencies verifies that all plugin dependencies are enabled
func checkPluginDependencies(config PluginConfig, plugins map[string]PluginConfig) error {
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

// verifyPluginMetadata verifies that plugin metadata matches configuration
func verifyPluginMetadata(plugin pluginCore.Plugin, config PluginConfig) error {
	if !strings.EqualFold(plugin.Name(), config.Name) {
		return fmt.Errorf("plugin name mismatch: got %s, want %s", plugin.Name(), config.Name)
	}
	if plugin.Version() != config.Version {
		return fmt.Errorf("plugin version mismatch: got %s, want %s", plugin.Version(), config.Version)
	}
	return nil
}

func handleShutdown(ctx context.Context, agent *core.Agent, timeoutSeconds int) chan struct{} {
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, initiating graceful shutdown...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer cancel()

		web.Stop()

		if err := agent.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		close(done)
	}()

	return done
}
