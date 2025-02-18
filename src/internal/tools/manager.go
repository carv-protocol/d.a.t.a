package tools

import (
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

type Manager struct {
	tools map[string]core.Tool
	// actions[type][name]
	actions map[string]map[string]actions.IAction
}

// TODO: keep one of tool or plugin
func NewManager() *Manager {
	return &Manager{
		tools:   make(map[string]core.Tool),
		actions: make(map[string]map[string]actions.IAction),
	}
}

// Register a plugin
func (m *Manager) Register(tool core.Tool) error {
	if _, exists := m.tools[tool.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", tool.Name())
	}

	m.tools[tool.Name()] = tool
	for _, action := range tool.AvailableActions() {
		if _, exists := m.actions[action.Type()]; !exists {
			m.actions[action.Type()] = make(map[string]actions.IAction)
		}
		m.actions[action.Type()][action.Name()] = action
	}
	return nil
}

func (m *Manager) AvailableTools() []core.Tool {
	var tools []core.Tool
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (m *Manager) AvailableActions() []actions.IAction {
	var actions []actions.IAction
	for _, tool := range m.tools {
		actions = append(actions, tool.AvailableActions()...)
	}
	return actions
}

func (m *Manager) GetAction(actionType string, actionName string) (actions.IAction, error) {
	actions, ok := m.actions[actionType]
	if !ok {
		return nil, fmt.Errorf("action type %s not found", actionType)
	}
	action, ok := actions[actionName]
	if !ok {
		return nil, fmt.Errorf("action name %s not found", actionName)
	}

	return action, nil
}
