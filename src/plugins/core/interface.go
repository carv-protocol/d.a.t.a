package core

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Basic information
	Name() string
	Description() string
	Version() string

	// Components
	Actions() []Action
	Providers() []Provider
	Services() []Service
	Clients() []Client

	// Lifecycle
	Init(ctx context.Context, opts map[string]interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// Action defines the interface for plugin actions
type Action interface {
	// Name returns the unique name of the action
	Name() string

	// Description returns the description of the action
	Description() string

	// Type returns the type of the action
	Type() string

	// Execute executes the action with given parameters
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// FetchTransactionAction defines the interface for fetch transaction action
type FetchTransactionAction interface {
	Action
	ExecuteWithParams(ctx context.Context, query string, params map[string]interface{}) (interface{}, error)
}

// FetchTransactionActionAdapter defines the interface for fetch transaction action adapter
type FetchTransactionActionAdapter interface {
	Action
	GetAction() FetchTransactionAction
}

// ProviderState represents the current state of a provider
type ProviderState struct {
	// Name of the provider
	Name string `json:"name"`

	// Type of the provider
	Type string `json:"type"`

	// Current state of the provider
	State string `json:"state"`

	// Additional metadata specific to the provider type
	Metadata map[string]interface{} `json:"metadata"`

	// Any error state
	Error string `json:"error,omitempty"`
}

// Provider interface defines methods that must be implemented by all providers
type Provider interface {
	// Name returns the name of the provider
	Name() string

	// Type returns the type of the provider
	Type() string

	// GetState returns the current state of the provider
	GetProviderState(ctx context.Context) (*ProviderState, error)
}

// SystemState represents the interface for system state
type SystemState interface {
	GetCharacter() interface{}
	GetAvailableTools() []interface{}
	GetAvailableActions() []actions.IAction
	GetAvailableEvaluators() []Evaluator
	GetStakeholderPreferences() map[string]interface{}
	GetActiveProviderStates() []*ProviderState
}

// SocialMessage represents the interface for social messages
type SocialMessage interface {
	GetPlatform() string
	GetFromUser() string
	GetContent() string
	GetMetadata() map[string]interface{}
}

// Stakeholder represents the interface for stakeholders
type Stakeholder interface {
	GetID() string
	GetType() string
	GetTokenBalance() interface{}
	GetHistoricalMsgs() []string
}

// EvaluationContext contains all the context needed for evaluation
type EvaluationContext struct {
	Action           actions.IAction
	State            SystemState
	Message          SocialMessage
	Stakeholder      Stakeholder
	ResponseMsg      string
	AvailableActions []actions.IAction
}

// EvaluationResult contains the result of an evaluation
type EvaluationResult struct {
	// Whether the response message should be modified
	ShouldModifyResponse bool
	// The modified response message if ShouldModifyResponse is true
	ModifiedResponse string
	// Additional actions suggested by the evaluator
	SuggestedActions []actions.IAction
	// Parameters for the suggested actions
	ActionParams map[string]interface{}
}

// Evaluator interface defines methods that must be implemented by evaluators
type Evaluator interface {
	// Name returns the name of the evaluator
	Name() string
	// Evaluate evaluates the current state and returns an evaluation result
	Evaluate(ctx context.Context, evalCtx *EvaluationContext) (*EvaluationResult, error)
}

// Service defines the interface for plugin services
type Service interface {
	// Name returns the unique name of the service
	Name() string

	// Start starts the service
	Start(ctx context.Context) error

	// Stop stops the service
	Stop(ctx context.Context) error
}

// Client defines the interface for plugin clients
type Client interface {
	// Name returns the unique name of the client
	Name() string

	// Connect establishes connection
	Connect(ctx context.Context) error

	// Close closes the connection
	Close(ctx context.Context) error
}

// PluginMetadata contains plugin metadata
type PluginMetadata struct {
	Name        string
	Description string
	Version     string
	Author      string
	License     string
	Homepage    string
	Repository  string
}

// PluginConfig contains plugin configuration
type PluginConfig struct {
	// Plugin metadata
	Metadata PluginMetadata

	// Plugin options
	Options map[string]interface{}

	// Dependencies
	Dependencies []string
}
