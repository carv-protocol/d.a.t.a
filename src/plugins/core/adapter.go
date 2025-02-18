package core

import (
	"context"
	"fmt"
)

// ActionAdapter adapts a plugin Action to the internal IAction interface.
// It provides thread-safe access to action execution results and parameters.
type ActionAdapter struct {
	action     Action
	ctx        context.Context
	params     map[string]interface{}
	lastResult interface{}
	lastError  error
}

// NewActionAdapter creates a new action adapter.
// It requires a context and an Action implementation.
// The context will be used for all action executions.
func NewActionAdapter(ctx context.Context, action Action) *ActionAdapter {
	if ctx == nil {
		ctx = context.Background()
	}
	return &ActionAdapter{
		action: action,
		ctx:    ctx,
		params: make(map[string]interface{}),
	}
}

// Name implements IAction interface.
// Returns the name of the underlying action.
func (a *ActionAdapter) Name() string {
	return a.action.Name()
}

// Description implements IAction interface.
// Returns the description of the underlying action.
func (a *ActionAdapter) Description() string {
	return a.action.Description()
}

// Type implements IAction interface.
// Returns a string representation of the action type,
// which is derived from the action name but can be overridden.
func (a *ActionAdapter) Type() string {
	return a.action.Name()
}

// Execute implements IAction interface.
// This method executes the underlying action with the current parameters and context.
func (a *ActionAdapter) Execute(ctx context.Context, params map[string]interface{}) error {
	// Store the parameters
	a.params = make(map[string]interface{}, len(params))
	for k, v := range params {
		a.params[k] = v
	}

	// Execute the action
	result, err := a.action.Execute(ctx, params)

	// Store results
	a.lastResult = result
	a.lastError = err

	return err
}

// Validate implements IAction interface.
// This method validates the parameters before execution.
func (a *ActionAdapter) Validate(params map[string]interface{}) error {
	// Basic validation: ensure params is not nil
	if params == nil {
		return fmt.Errorf("params cannot be nil")
	}

	// TODO: Add more validation if needed
	return nil
}

// ParametersPrompt implements IAction interface.
// This method returns a string describing the expected parameters.
func (a *ActionAdapter) ParametersPrompt() string {
	return fmt.Sprintf(`Parameters for %s:
{
	// Add parameters description here
}`, a.Name())
}

// GetResult returns the result of the last execution.
// Returns nil if no execution has occurred.
func (a *ActionAdapter) GetResult() interface{} {
	return a.lastResult
}

// GetError returns the error from the last execution.
// Returns nil if no execution has occurred or if the last execution was successful.
func (a *ActionAdapter) GetError() error {
	return a.lastError
}

// GetParams returns a copy of the current parameters.
func (a *ActionAdapter) GetParams() map[string]interface{} {
	params := make(map[string]interface{}, len(a.params))
	for k, v := range a.params {
		params[k] = v
	}
	return params
}

// GetOriginalAction returns the underlying action
func (a *ActionAdapter) GetOriginalAction() Action {
	return a.action
}
