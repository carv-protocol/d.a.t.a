package data

import (
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
	walletactions "github.com/carv-protocol/d.a.t.a/src/plugins/plugin-evm-wallet/actions"

	"go.uber.org/zap"
)

// Required configuration keys
const (
	ConfigPrivateKey = "private_key"
	ConfigNetwork    = "network"
	ConfigRPCURL     = "rpc_url"
	ConfigChainID    = "chain_id"
	ConfigTimeout    = "timeout"
)

// Plugin implements the core.Plugin interface for data functionality
type evmPlugin struct {
	name        string
	description string
	version     string
	actions     []actions.IAction
	logger      *zap.SugaredLogger
}

// NewPlugin creates a new data plugin
func NewPlugin(llmClient llm.Client, config *plugins.Config) (plugins.Plugin, error) {
	if err := validateConfig(config.Options); err != nil {
		return nil, fmt.Errorf("invalid plugin configuration: %w", err)
	}

	transferAllERC20Action, err := walletactions.NewTransferAllERC20Action(
		config.Options[ConfigPrivateKey].(string),
		config.Options[ConfigNetwork].(string),
		config.Options[ConfigRPCURL].(string),
		config.Options[ConfigChainID].(int64),
		time.Duration(config.Options[ConfigTimeout].(int64)),
		"TransferAllERC20Action",
	)
	if err != nil {
		return nil, err
	}

	return &evmPlugin{
		name:        "evm-wallet",
		description: "EVM Wallet Plugin supports EVM wallet actions, such as transferring ERC20 tokens",
		logger:      logger.GetLogger().With(zap.String("plugin", "evm-wallet")),
		actions:     []actions.IAction{transferAllERC20Action},
	}, nil
}

// Name implements core.Plugin interface
func (p *evmPlugin) Name() string {
	return p.name
}

// Description implements core.Plugin interface
func (p *evmPlugin) Description() string {
	return p.description
}

// Version implements core.Plugin interface
func (p *evmPlugin) Version() string {
	return p.version
}

// Actions implements core.Plugin interface
func (p *evmPlugin) Actions() []actions.IAction {
	return p.actions
}

// Providers implements core.Plugin interface
func (p *evmPlugin) Providers() []plugins.Provider {
	return nil
}

// Evaluators implements core.Plugin interface
func (p *evmPlugin) Evaluators() []plugins.Evaluator {
	return nil
}

// Services implements core.Plugin interface
func (p *evmPlugin) Services() []plugins.Service {
	return nil
}

// validateConfig validates the plugin configuration
func validateConfig(opts map[string]interface{}) error {
	required := []string{ConfigPrivateKey, ConfigNetwork, ConfigRPCURL, ConfigChainID, ConfigTimeout}
	for _, key := range required {
		_, ok := opts[key]
		if !ok {
			return fmt.Errorf("missing required configuration: %s", key)
		}
	}
	return nil
}
