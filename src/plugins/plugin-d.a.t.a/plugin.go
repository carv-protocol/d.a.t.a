package data

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	"github.com/carv-protocol/d.a.t.a/src/plugins/core"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/actions"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/providers"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/types"
	"go.uber.org/zap"
)

// init registers the plugin factory
func init() {
	core.RegisterPlugin("d.a.t.a", NewPlugin)
}

// Required configuration keys
const (
	ConfigKeyAPIURL    = "api_url"    // maps to CarvConfig.BaseURL
	ConfigKeyAuthToken = "auth_token" // maps to CarvConfig.APIKey
	ConfigKeyChain     = "chain"      // maps to Token.Network
	ConfigKeyLLM       = "llm"        // LLM configuration section
)

// Plugin implements the core.Plugin interface for data functionality
type Plugin struct {
	config     *core.PluginConfig
	llmClient  llm.Client
	logger     *zap.SugaredLogger
	actions    []core.Action
	providers  []core.Provider
	evaluators []core.Evaluator
	services   []core.Service
	clients    []core.Client
}

// NewPlugin creates a new data plugin
func NewPlugin(llmClient llm.Client) core.Plugin {
	return &Plugin{
		llmClient: llmClient,
		logger:    logger.GetLogger().With(zap.String("plugin", "d.a.t.a")),
		config: &core.PluginConfig{
			Metadata: core.PluginMetadata{
				Name:        "d.a.t.a",
				Description: "Data interaction plugin",
				Version:     "1.0.0",
				Author:      "CARV Protocol",
				License:     "MIT",
				Homepage:    "https://github.com/carv-protocol/d.a.t.a",
				Repository:  "https://github.com/carv-protocol/d.a.t.a",
			},
			Options: make(map[string]interface{}),
		},
	}
}

// Name implements core.Plugin interface
func (p *Plugin) Name() string {
	return p.config.Metadata.Name
}

// Description implements core.Plugin interface
func (p *Plugin) Description() string {
	return p.config.Metadata.Description
}

// Version implements core.Plugin interface
func (p *Plugin) Version() string {
	return p.config.Metadata.Version
}

// Actions implements core.Plugin interface
func (p *Plugin) Actions() []core.Action {
	return p.actions
}

// Providers implements core.Plugin interface
func (p *Plugin) Providers() []core.Provider {
	return p.providers
}

// Evaluators implements core.Plugin interface
func (p *Plugin) Evaluators() []core.Evaluator {
	return p.evaluators
}

// Services implements core.Plugin interface
func (p *Plugin) Services() []core.Service {
	return p.services
}

// Clients implements core.Plugin interface
func (p *Plugin) Clients() []core.Client {
	return p.clients
}

// validateConfig validates the plugin configuration
func (p *Plugin) validateConfig(opts map[string]interface{}) error {
	required := []string{ConfigKeyAPIURL, ConfigKeyAuthToken, ConfigKeyChain, ConfigKeyLLM}
	for _, key := range required {
		val, ok := opts[key]
		if !ok {
			return fmt.Errorf("missing required configuration: %s", key)
		}
		if key == ConfigKeyLLM {
			// Try both map[interface{}]interface{} and map[string]interface{}
			llmConfig, ok := val.(map[string]interface{})
			if !ok {
				// Try the old type as fallback
				if llmConfig2, ok := val.(map[interface{}]interface{}); ok {
					// Convert to map[string]interface{}
					llmConfig = make(map[string]interface{})
					for k, v := range llmConfig2 {
						if kStr, ok := k.(string); ok {
							llmConfig[kStr] = v
						}
					}
				} else {
					return fmt.Errorf("invalid configuration value for %s: must be a map", key)
				}
			}
			if model, ok := llmConfig["model"].(string); !ok || model == "" {
				return fmt.Errorf("invalid or missing model in LLM configuration")
			}
		} else if strVal, ok := val.(string); !ok || strVal == "" {
			return fmt.Errorf("invalid configuration value for %s: must be a non-empty string", key)
		}
	}
	return nil
}

// DatabaseProviderAdapter adapts DatabaseProvider to core.Provider interface
type DatabaseProviderAdapter struct {
	provider types.DatabaseProvider
	logger   *zap.SugaredLogger
}

func (a *DatabaseProviderAdapter) Name() string {
	return "ethereum_database_provider"
}

func (a *DatabaseProviderAdapter) Type() string {
	return "database"
}

func (a *DatabaseProviderAdapter) GetProviderState(ctx context.Context) (*core.ProviderState, error) {
	providerImpl, ok := a.provider.(*providers.DatabaseProviderImpl)
	if !ok {
		return nil, fmt.Errorf("invalid provider type: expected *DatabaseProviderImpl")
	}
	return providerImpl.GetProviderState(ctx)
}

func (a *DatabaseProviderAdapter) GetData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return a.provider.ProcessQuery(ctx, params)
}

// FetchTransactionActionAdapter adapts FetchTransactionAction to core.Action interface
type FetchTransactionActionAdapter struct {
	action core.FetchTransactionAction
	logger *zap.SugaredLogger
}

func NewFetchTransactionActionAdapter(action core.FetchTransactionAction, logger *zap.SugaredLogger) *FetchTransactionActionAdapter {
	return &FetchTransactionActionAdapter{
		action: action,
		logger: logger,
	}
}

func (a *FetchTransactionActionAdapter) Name() string {
	return a.action.Name()
}

func (a *FetchTransactionActionAdapter) Description() string {
	return a.action.Description()
}

func (a *FetchTransactionActionAdapter) Type() string {
	return a.action.Type()
}

func (a *FetchTransactionActionAdapter) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	a.logger.Infow("Executing fetch transaction action", "params", params)
	result, err := a.action.Execute(ctx, params)
	if err != nil {
		a.logger.Errorw("Failed to execute fetch transaction action", "error", err)
		return nil, err
	}
	a.logger.Infow("Fetch transaction action executed successfully", "result", result)
	return result, nil
}

// GetAction returns the underlying FetchTransactionAction
func (a *FetchTransactionActionAdapter) GetAction() core.FetchTransactionAction {
	return a.action
}

// Init implements core.Plugin interface
func (p *Plugin) Init(ctx context.Context, opts map[string]interface{}) error {
	// Store configuration
	if p.config == nil {
		p.config = &core.PluginConfig{
			Options: make(map[string]interface{}),
		}
	}

	// Merge options
	for k, v := range opts {
		p.config.Options[k] = v
	}

	// Validate configuration
	if err := p.validateConfig(p.config.Options); err != nil {
		p.logger.Errorw("Failed to validate configuration", "error", err)
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize provider
	llmConfig, ok := p.config.Options[ConfigKeyLLM].(map[string]interface{})
	if !ok {
		// Try converting from map[interface{}]interface{}
		llmConfig2, ok := p.config.Options[ConfigKeyLLM].(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("invalid LLM configuration type: expected map[string]interface{} or map[interface{}]interface{}")
		}
		// Convert to map[string]interface{}
		llmConfig = make(map[string]interface{})
		for k, v := range llmConfig2 {
			if kStr, ok := k.(string); ok {
				llmConfig[kStr] = v
			}
		}
	}

	model, ok := llmConfig["model"].(string)
	if !ok || model == "" {
		return fmt.Errorf("invalid or missing model in LLM configuration")
	}

	// Create provider using factory
	provider := providers.NewDatabaseProvider(
		"ethereum_database_provider",
		p.config.Options[ConfigKeyAPIURL].(string),
		p.config.Options[ConfigKeyAuthToken].(string),
		p.config.Options[ConfigKeyChain].(string),
		getDefaultDatabaseSchema(),
		getDefaultQueryExamples(),
		p.llmClient,
		model,
		p.logger,
	)
	providerAdapter := &DatabaseProviderAdapter{
		provider: provider,
		logger:   p.logger,
	}
	p.providers = append(p.providers, providerAdapter)

	// Create action using factory
	action := actions.NewFetchTransactionAction(provider)
	actionAdapter := NewFetchTransactionActionAdapter(action, p.logger)
	p.actions = append(p.actions, actionAdapter)

	return nil
}

// Start implements core.Plugin interface
func (p *Plugin) Start(ctx context.Context) error {
	// Start all services
	for _, service := range p.services {
		if err := service.Start(ctx); err != nil {
			p.logger.Errorw("Failed to start service",
				"service", service.Name(),
				"error", err,
			)
			return fmt.Errorf("failed to start service %s: %w", service.Name(), err)
		}
	}

	// Connect all clients
	for _, client := range p.clients {
		if err := client.Connect(ctx); err != nil {
			p.logger.Errorw("Failed to connect client",
				"client", client.Name(),
				"error", err,
			)
			return fmt.Errorf("failed to connect client %s: %w", client.Name(), err)
		}
	}

	p.logger.Info("d.a.t.a plugin started successfully")
	return nil
}

// Stop implements core.Plugin interface
func (p *Plugin) Stop(ctx context.Context) error {
	p.logger.Info("Stopping data plugin")

	var errs []error

	// Stop all services
	for _, service := range p.services {
		if err := service.Stop(ctx); err != nil {
			p.logger.Errorw("Failed to stop service",
				"service", service.Name(),
				"error", err,
			)
			errs = append(errs, fmt.Errorf("failed to stop service %s: %w", service.Name(), err))
		}
	}

	// Close all clients
	for _, client := range p.clients {
		if err := client.Close(ctx); err != nil {
			p.logger.Errorw("Failed to close client",
				"client", client.Name(),
				"error", err,
			)
			errs = append(errs, fmt.Errorf("failed to close client %s: %w", client.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to stop plugin cleanly: %v", errs)
	}

	p.logger.Info("Data plugin stopped successfully")
	return nil
}

// getDefaultDatabaseSchema returns the default database schema
func getDefaultDatabaseSchema() string {
	return `
CREATE EXTERNAL TABLE transactions(
    hash string,
    nonce bigint,
    block_hash string,
    block_number bigint,
    block_timestamp timestamp,
    date string,
    transaction_index bigint,
    from_address string,
    to_address string,
    value double,
    gas bigint,
    gas_price bigint,
    input string,
    max_fee_per_gas bigint,
    max_priority_fee_per_gas bigint,
    transaction_type bigint
) PARTITIONED BY (date string);
`
}

// getDefaultQueryExamples returns default query examples
func getDefaultQueryExamples() string {
	return `
Common Query Examples:

1. Find Most Active Addresses in Last 7 Days:
SELECT from_address, COUNT(*) as tx_count 
FROM eth.transactions 
WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d')
GROUP BY from_address 
ORDER BY tx_count DESC 
LIMIT 10;

2. Analyze Gas Price Trends:
SELECT date_trunc('hour', block_timestamp) as hour,
       avg(gas_price) as avg_gas_price
FROM eth.transactions
WHERE date >= date_sub(current_date(), 1)
GROUP BY 1
ORDER BY 1;

3. Find latest transactions:
SELECT * FROM eth.transactions 
WHERE date >= date_format(date_add('day', -7, current_date), '%Y-%m-%d')
ORDER BY block_timestamp DESC 
LIMIT 3;
`
}
