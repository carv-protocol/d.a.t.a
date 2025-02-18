package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/social"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/internal/tools"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	pluginCore "github.com/carv-protocol/d.a.t.a/src/plugins/core"
	dataPlugin "github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a"
	customTools "github.com/carv-protocol/d.a.t.a/src/tools"
	dataTool "github.com/carv-protocol/d.a.t.a/src/tools/d.a.t.a"
	"github.com/carv-protocol/d.a.t.a/src/tools/wallet"

	"github.com/google/uuid"
)

// Config validation errors
var (
	ErrInvalidLLMConfig = errors.New("invalid LLM configuration")
	ErrInvalidDBConfig  = errors.New("invalid database configuration")
)

func main() {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	config, err := loadConfig()
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

	// Wait for shutdown signal
	<-handleShutdown(ctx, agent, config.Settings.ShutdownTimeout)
}

func initializeAgent(ctx context.Context, config *Config) (*core.Agent, error) {
	// Setup database
	store := adapters.NewSQLiteStore(config.Database.Path)
	if err := store.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize components
	llmClient := llm.NewClient((*llm.LLMConfig)(&config.LLMConfig))
	carvClient := carv.NewClient(config.Data.CarvConfig.APIKey, config.Data.CarvConfig.BaseURL)
	memoryManager := memory.NewManager(store)
	tokenManager := token.NewTokenManager(carvClient, &core.TokenInfo{
		Network:      config.Token.Network,
		Ticker:       config.Token.Ticker,
		ContractAddr: config.Token.ContractAddr,
	})
	stakeholderManager := token.NewStakeholderManager(memoryManager)

	// Load character
	character, err := characters.LoadFromFile(config.Character.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %w", err)
	}

	// Initialize tools
	toolsManager := initializeTools(config)

	// Initialize plugins
	pluginRegistry := initializePlugins(config)

	actionManager, err := registerPlugins(ctx, pluginRegistry, config)
	if err != nil {
		return nil, fmt.Errorf("failed to register plugins: %w", err)
	}

	// Create agent
	agentConfig := core.AgentConfig{
		ID:           uuid.New(),
		Character:    character,
		LLMClient:    llmClient,
		Model:        config.LLMConfig.Model,
		Stakeholders: stakeholderManager,
		ToolsManager: toolsManager,
		SocialClient: social.NewSocialClient(
			&config.Social.TwitterConfig,
			&config.Social.DiscordConfig,
			&config.Social.TelegramConfig,
		),
		TaskManager:    tasks.NewManager(tasks.NewTaskStore(store)),
		ActionManager:  actionManager,
		TokenManager:   tokenManager,
		PluginRegistry: pluginRegistry,
	}

	agent, err := core.NewAgent(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
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

		// Get plugin factory
		factory, exists := builtinPlugins[name]
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

// returns the action manager
func registerPlugins(ctx context.Context, pluginRegistry *pluginCore.Registry, config *Config) (actions.ActionManager, error) {
	// Initialize each plugin with its own options
	for name, pluginConfig := range config.Plugin.Plugins {
		if !pluginConfig.Enabled {
			continue
		}

		// Check if plugin is registered
		if _, exists := pluginRegistry.GetPlugin(name); !exists {
			log.Printf("Plugin %s is not registered, skipping initialization", name)
			continue
		}

		if err := pluginRegistry.InitPlugin(ctx, name, pluginConfig.Options); err != nil {
			return nil, fmt.Errorf("failed to initialize plugin %s: %w", name, err)
		}
	}

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
	if plugin.Name() != config.Name {
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

		if err := agent.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		close(done)
	}()

	return done
}
