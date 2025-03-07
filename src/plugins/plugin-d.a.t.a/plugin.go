package data

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	walletactions "github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/actions"
	"github.com/carv-protocol/d.a.t.a/src/plugins/plugin-d.a.t.a/providers"

	"go.uber.org/zap"
)

// Required configuration keys
const (
	ConfigKeyAPIURL    = "api_url"    // maps to CarvConfig.BaseURL
	ConfigKeyAuthToken = "auth_token" // maps to CarvConfig.APIKey
	ConfigKeyChain     = "chain"      // maps to Token.Network
	ConfigKeyLLM       = "llm"        // LLM configuration section
)

// dataPlugin implements the core.Plugin interface for data functionality
type dataPlugin struct {
	llmClient  llm.Client
	metadata   plugins.PluginMetadata
	logger     *zap.SugaredLogger
	actions    []actions.IAction
	providers  []plugins.Provider
	evaluators []plugins.Evaluator
	services   []plugins.Service
}

// NewPlugin creates a new data plugin
func NewPlugin(llmClient llm.Client, config *plugins.Config) (plugins.Plugin, error) {
	logger := logger.GetLogger().With(zap.String("plugin", "d.a.t.a"))

	if err := validateConfig(config.Options); err != nil {
		return nil, fmt.Errorf("invalid plugin configuration: %w", err)
	}

	// Initialize provider
	llmConfig, ok := config.Options[ConfigKeyLLM].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid LLM configuration type: expected map[string]interface{}")
	}

	model, ok := llmConfig["model"].(string)
	if !ok || model == "" {
		return nil, fmt.Errorf("invalid or missing model in LLM configuration")
	}

	// Create provider using factory
	provider := providers.NewDatabaseProvider(
		"ethereum_database_provider",
		config.Options[ConfigKeyAPIURL].(string),
		config.Options[ConfigKeyAuthToken].(string),
		config.Options[ConfigKeyChain].(string),
		getDefaultDatabaseSchema(),
		getDefaultQueryExamples(),
		llmClient,
		model,
		logger,
	)

	// Create action using factory
	action := walletactions.NewFetchTransactionAction(provider)

	return &dataPlugin{
		llmClient: llmClient,
		logger:    logger,
		providers: []plugins.Provider{provider},
		actions:   []actions.IAction{action},
		metadata: plugins.PluginMetadata{
			Name:        "d.a.t.a",
			Description: "Data interaction plugin",
			Version:     "1.0.0",
			Author:      "CARV Protocol",
			License:     "MIT",
			Homepage:    "https://github.com/carv-protocol/d.a.t.a",
			Repository:  "https://github.com/carv-protocol/d.a.t.a",
		},
	}, nil
}

// Name implements core.Plugin interface
func (p *dataPlugin) Name() string {
	return p.metadata.Name
}

// Description implements core.Plugin interface
func (p *dataPlugin) Description() string {
	return p.metadata.Description
}

// Version implements core.Plugin interface
func (p *dataPlugin) Version() string {
	return p.metadata.Version
}

// Actions implements core.Plugin interface
func (p *dataPlugin) Actions() []actions.IAction {
	return p.actions
}

// Providers implements core.Plugin interface
func (p *dataPlugin) Providers() []plugins.Provider {
	return p.providers
}

// Evaluators implements core.Plugin interface
func (p *dataPlugin) Evaluators() []plugins.Evaluator {
	return p.evaluators
}

// Services implements core.Plugin interface
func (p *dataPlugin) Services() []plugins.Service {
	return p.services
}

// Clients implements core.Plugin interface
func (p *dataPlugin) Init(ctx context.Context) error {
	return nil
}

// validateConfig validates the plugin configuration
func validateConfig(opts map[string]interface{}) error {
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

// Start implements core.Plugin interface
func (p *dataPlugin) Start(ctx context.Context) error {
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

	p.logger.Info("d.a.t.a plugin started successfully")
	return nil
}

// Stop implements core.Plugin interface
func (p *dataPlugin) Stop(ctx context.Context) error {
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
