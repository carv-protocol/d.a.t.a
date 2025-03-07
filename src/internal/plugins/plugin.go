package plugins

import (
	"context"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
)

// Plugin defines behavior for extending D.A.T.A
// Hook Event Summary:
// 1. BeforeAnalysis: Invoked before the cognitive analysis starts.
//    - Purpose: Modify or inspect the input data before analysis.
//    - Input: `Input` (raw user input).
// 2. AfterAnalysis: Invoked after cognitive analysis completes.
//    - Purpose: Add metadata or post-process the analysis results.
//    - Input: `*Analysis` (analysis object generated from input).
// 3. FilterAction: Invoked during action evaluation to filter out invalid or unwanted actions.
//    - Purpose: Enable or disable actions based on the current analysis.
//    - Input: `*Action`, `*Analysis` (current action and related analysis).
//    - Output: `bool` (allow or reject action), `error`.
// 4. BeforeActionExecution: Invoked before an action is executed.
//    - Purpose: Prepare or modify the action before execution.
//    - Input: `*Action` (the action about to be executed).
// 5. AfterActionExecution: Invoked after an action is executed.
//    - Purpose: Inspect or log the result of the executed action.
//    - Input: `*Action`, `*ActionResult` (executed action and its result).

// Plugin interface
type Plugin interface {
	Name() string
	Description() string
	Providers() []Provider
	Actions() []actions.IAction
	Evaluators() []Evaluator
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

// Evaluator defines the interface for plugin evaluators
type Evaluator interface {
	// Name returns the unique name of the evaluator
	Name() string

	// Evaluate performs evaluation with given parameters
	Evaluate(ctx context.Context, params map[string]interface{}) (interface{}, error)
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

// Config contains plugin configuration
type Config struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`

	// Plugin options
	Options map[string]interface{} `mapstructure:"options"`
}
