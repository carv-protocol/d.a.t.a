package governance

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	govActions "github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a-governance/actions"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a-governance/provider"

	"go.uber.org/zap"
)

// Required configuration keys
const (
	ConfigKeyAPIURL       = "api_url"       // maps to CarvConfig.BaseURL
	ConfigKeyAuthToken    = "auth_token"    // maps to CarvConfig.APIKey
	ConfigKeyChain        = "chain"         // maps to Token.Network
	ConfigKeyRegistryType = "registry_type" // type of governance registry to use
)

// Plugin implements the core.Plugin interface for governance functionality
type governancePlugin struct {
	llmClient  llm.Client
	metadata   plugins.PluginMetadata
	logger     *zap.SugaredLogger
	actions    []actions.IAction
	providers  []plugins.Provider
	evaluators []plugins.Evaluator
	services   []plugins.Service
}

// NewPlugin creates a new governance plugin
func NewPlugin(llmClient llm.Client, config *plugins.Config) (plugins.Plugin, error) {
	logger := logger.GetLogger().With(zap.String("plugin", "d.a.t.a-governance"))

	// Validate configuration
	if err := validateConfig(config.Options); err != nil {
		return nil, fmt.Errorf("invalid plugin configuration: %w", err)
	}

	// In a real implementation, we would initialize the governance registry based on config
	// For now, we'll use a memory-based implementation for demonstration
	govRegistry := governance.NewMemoryRegistry(nil)

	// Create provider
	proposalProvider := provider.NewProposalInformationProvider(govRegistry)

	// Create actions
	proposalAction := govActions.NewProposalAction(govRegistry)
	voteAction := govActions.NewVoteAction(govRegistry)

	return &governancePlugin{
		llmClient: llmClient,
		logger:    logger,
		providers: []plugins.Provider{proposalProvider},
		actions:   []actions.IAction{proposalAction, voteAction},
		metadata: plugins.PluginMetadata{
			Name:        "d.a.t.a-governance",
			Description: "Governance interaction plugin",
			Version:     "1.0.0",
			Author:      "CARV",
			License:     "Proprietary",
		},
	}, nil
}

// Name returns the plugin name
func (p *governancePlugin) Name() string {
	return p.metadata.Name
}

// Description returns the plugin description
func (p *governancePlugin) Description() string {
	return p.metadata.Description
}

// Version returns the plugin version
func (p *governancePlugin) Version() string {
	return p.metadata.Version
}

// Actions returns the plugin actions
func (p *governancePlugin) Actions() []actions.IAction {
	return p.actions
}

// Providers returns the plugin providers
func (p *governancePlugin) Providers() []plugins.Provider {
	return p.providers
}

// Evaluators returns the plugin evaluators
func (p *governancePlugin) Evaluators() []plugins.Evaluator {
	return p.evaluators
}

// Services returns the plugin services
func (p *governancePlugin) Services() []plugins.Service {
	return p.services
}

// validateConfig validates the plugin configuration
func validateConfig(opts map[string]interface{}) error {
	// For now, we won't enforce strict configuration requirements
	// In a real implementation, we would validate all required configuration fields

	// Example of validation:
	// if _, ok := opts[ConfigKeyAPIURL]; !ok {
	//    return fmt.Errorf("missing required config: %s", ConfigKeyAPIURL)
	// }

	return nil
}

// Init initializes the plugin
func (p *governancePlugin) Init(ctx context.Context) error {
	p.logger.Info("Initializing d.a.t.a-governance plugin")

	// Additional initialization if needed

	return nil
}

// Start starts the plugin
func (p *governancePlugin) Start(ctx context.Context) error {
	p.logger.Info("d.a.t.a-governance plugin started successfully")
	return nil
}

// Stop stops the plugin
func (p *governancePlugin) Stop(ctx context.Context) error {
	p.logger.Info("d.a.t.a-governance plugin stopped successfully")
	return nil
}
