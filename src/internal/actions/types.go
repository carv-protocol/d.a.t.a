package actions

import (
	"context"
	"time"
)

// IAction is an interface for actions that can be executed by the agent
type IAction interface {
	Name() string
	Description() string
	Execute(ctx context.Context, params map[string]interface{}) error
	Type() string
	Validate(params map[string]interface{}) error
	ParametersPrompt() string
}

// ActionManager is an interface for managing actions
type ActionManager interface {
	Register(action IAction) error
	GetAvailableActions() []IAction
	GetAction(name string) IAction
}

// ActionKind defines the type of action to be executed
type ActionKind string

// ActionInfo represents an action waiting to be executed
type ActionInfo struct {
	ID           string
	Name         string
	Kind         ActionKind
	Parameters   map[string]interface{}
	Priority     float64
	Deadline     time.Time
	Dependencies []string // IDs of actions that must complete first
}
