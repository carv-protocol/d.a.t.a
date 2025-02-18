package core

import (
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	pluginCore "github.com/carv-protocol/d.a.t.a/src/plugins/core"
	"github.com/google/uuid"
)

// AgentConfig represents the configuration for creating a new agent
type AgentConfig struct {
	ID             uuid.UUID
	Character      *characters.Character
	LLMClient      llm.Client
	Model          string
	TaskManager    TaskManager
	ActionManager  actions.ActionManager
	ToolsManager   ToolManager
	Stakeholders   StakeholderManager
	TokenManager   TokenManager
	SocialClient   SocialClient
	PluginRegistry *pluginCore.Registry
	Training       struct {
		Enabled       bool
		MaxIterations int
		BatchSize     int
		StopThreshold float64
	}
	Inference struct {
		Temperature    float64
		MaxChainLength int
		MinConfidence  float64
	}

	SystemConfig struct {
		MaxConcurrentTasks int
		TaskTimeout        time.Duration
		MessageBatchSize   int
		MonitorInterval    time.Duration
		Temperature        float64
		MaxChainLength     int
	}
}

func validateConfig(config *AgentConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.ID == uuid.Nil {
		return fmt.Errorf("agent ID is required")
	}
	if config.Model == "" {
		return fmt.Errorf("model name is required")
	}
	// Add more validation as needed
	return nil
}
