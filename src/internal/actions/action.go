package actions

import (
	"context"
	"fmt"
	"time"
)

// ActionType defines the type of action to be executed
type ActionType string

// ActionDetails represents an action waiting to be executed
type ActionDetails struct {
	ID           string
	Name         string
	Type         ActionType
	Parameters   map[string]interface{}
	Priority     float64
	Deadline     time.Time
	Dependencies []string // IDs of actions that must complete first
}

func (a *ActionDetails) Execute(ctx context.Context) error {
	// TODO: Implement me
	return nil
}

type ManagerImpl struct {
	actions map[string]IAction
}

func NewManager() *ManagerImpl {
	return &ManagerImpl{
		actions: make(map[string]IAction),
	}
}

func (m *ManagerImpl) Register(action IAction) error {
	name := action.Name()
	if _, exists := m.actions[action.Name()]; exists {
		return fmt.Errorf("action %s already registered", name)
	}

	m.actions[action.Name()] = action
	return nil
}

func (m *ManagerImpl) GetAvailableActions() []IAction {
	actions := make([]IAction, 0, len(m.actions))
	for _, action := range m.actions {
		actions = append(actions, action)
	}
	return actions
}

func (m *ManagerImpl) GetAction(name string) IAction {
	return m.actions[name]
}
