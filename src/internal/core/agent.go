// agent.go
package core

import (
	"context"
	"fmt"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/data"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/internal/tools"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Agent struct {
	ID            uuid.UUID
	cognitive     *CognitiveEngine
	memoryManager memory.Manager
	character     *characters.Character
	taskManager   *tasks.Manager
	dataManager   data.Manager
	actionManager actions.Manager
	plugins       *plugins.PluginRegistry
	config        AgentConfig
	logger        *zap.SugaredLogger
	toolManager   *tools.Manager
	stakeholders  *token.StakeholderManager
	socialClient  SocialClient
	Goals         []Goal
	ctx           context.Context
	cancel        context.CancelFunc
}

// SystemState represents the complete state of the agent system
type SystemState struct {
	// General system information
	Timestamp   time.Time
	AgentStates *AgentState

	// Token and stakeholder information
	// TokenState             *token.TokenState
	StakeholderPreferences map[string]interface{}
	// ActiveVotes            map[string][]token.Vote

	Character        *characters.Character
	AvailableTools   []tools.Tool
	AvailableActions []actions.Action
	// Task and action information
	ActiveTasks    []*tasks.Task
	PendingActions []*actions.Action
}

type Goal struct {
	ID          string
	Name        string
	Description string
	Weight      float64
}

type AgentConfig struct {
	ID            uuid.UUID
	Character     *characters.Character
	LLMClient     llm.Client
	DataManager   data.Manager
	MemoryManager memory.Manager
	TaskManager   *tasks.Manager
	Stakeholders  *token.StakeholderManager
	ActionManager actions.Manager
	SocialClient  SocialClient
	ToolsManager  *tools.Manager
	Training      struct {
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
	CognitiveConfig *CognitiveConfig
}

type AgentStatus string

const (
	AgentStatusIdle       AgentStatus = "IDLE"
	AgentStatusProcessing AgentStatus = "PROCESSING"
	AgentStatusPaused     AgentStatus = "PAUSED"
	AgentStatusError      AgentStatus = "ERROR"
)

// AgentState represents the state of an individual agent
type AgentState struct {
	ID             string
	Status         AgentStatus
	CurrentTask    *tasks.Task
	Goals          []Goal
	LastActionTime time.Time
}

func NewAgent(config AgentConfig) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	logger, _ := zap.NewProduction()

	return &Agent{
		ID:            config.ID,
		cognitive:     NewCognitiveEngine(config.LLMClient, config.Character, logger.Sugar(), config.CognitiveConfig),
		memoryManager: config.MemoryManager,
		character:     config.Character,
		dataManager:   config.DataManager,
		config:        config,
		logger:        logger.Sugar(),
		toolManager:   config.ToolsManager,
		stakeholders:  config.Stakeholders,
		ctx:           ctx,
		socialClient:  config.SocialClient,
		cancel:        cancel,
		taskManager:   config.TaskManager,
		actionManager: config.ActionManager,
	}
}

func (a *Agent) GenerateTasks(ctx context.Context, state *SystemState) ([]*tasks.Task, error) {
	tasks, err := a.cognitive.GenerateTasks(ctx, state)
	if err != nil {
		a.logger.Errorw("Failed to evaluate task", "error", err)
		return nil, err
	}

	return tasks.Tasks, nil
}

func (a *Agent) ExecuteTask(ctx context.Context, task *tasks.Task, state *SystemState) (*TaskResult, error) {
	// Generate actions using cognitive engine
	actionGen, err := a.cognitive.GenerateActions(ctx, task, state)
	if err != nil {
		return nil, fmt.Errorf("failed to generate actions: %w", err)
	}

	// Execute actions with continuous verification
	var results []error
	for _, action := range actionGen.Actions {
		// Execute action
		err := a.executeAction(ctx, action)
		results = append(results, err)

		// results = append(results, result)

		// Update stakeholders on significant progress
		// if a.isSignificantProgress(result) {
		// 	a.notifyStakeholders(ctx, task, result)
		// }
	}

	return &TaskResult{
		TaskID:    task.ID,
		Task:      task,
		Actions:   actionGen.Actions,
		Timestamp: time.Now(),
		Result:    results,
	}, nil
}

func (a *Agent) GetState() *AgentState {
	return &AgentState{
		ID:             a.ID.String(),
		Status:         AgentStatusIdle,
		CurrentTask:    nil,
		Goals:          a.Goals,
		LastActionTime: time.Now(),
	}
}

func (a *Agent) executeAction(ctx context.Context, action actions.Action) error {
	a.logger.Infow("Executing action", "type", action.Type)

	// Execute action
	if err := action.Execute(); err != nil {
		return err
	}

	return nil
}
