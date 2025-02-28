package core

import (
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	pluginCore "github.com/carv-protocol/d.a.t.a/src/plugins/core"
)

// Implement pluginCore.SystemState interface
func (s *SystemState) GetCharacter() interface{} {
	return s.Character
}

func (s *SystemState) GetAvailableTools() []interface{} {
	tools := make([]interface{}, len(s.AvailableTools))
	for i, tool := range s.AvailableTools {
		tools[i] = tool
	}
	return tools
}

func (s *SystemState) GetAvailableActions() []actions.IAction {
	return s.AvailableActions
}

func (s *SystemState) GetAvailableEvaluators() []pluginCore.Evaluator {
	return s.AvailableEvaluators
}

func (s *SystemState) GetStakeholderPreferences() map[string]interface{} {
	return s.StakeholderPreferences
}

func (s *SystemState) GetActiveProviderStates() []*pluginCore.ProviderState {
	return s.ProviderStates
}
