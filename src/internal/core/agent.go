// agent.go
package core

import (
	"context"
	"fmt"
	"plugin"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/commands"
	"github.com/carv-protocol/d.a.t.a/src/internal/types"
	pluginCore "github.com/carv-protocol/d.a.t.a/src/plugins/core"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Agent struct {
	ID             uuid.UUID
	cognitive      *CognitiveEngine
	character      *characters.Character
	taskManager    TaskManager
	actionManager  actions.ActionManager
	logger         *zap.SugaredLogger
	toolManager    ToolManager
	stakeholders   StakeholderManager
	TokenManager   TokenManager
	socialClient   SocialClient
	pluginRegistry *pluginCore.Registry
	Goals          []Goal
	ctx            context.Context
	cancel         context.CancelFunc
	CommandManager *commands.Registry
}

// SystemState represents the complete state of the agent system
type SystemState struct {
	// General system information
	Timestamp   time.Time
	AgentStates *AgentState

	// Token and stakeholder information
	// TokenState             *TokenState
	StakeholderPreferences map[string]interface{}
	// ActiveVotes            map[string][]Vote

	Character           *characters.Character
	AvailableTools      []Tool
	AvailableActions    []actions.IAction
	AvailableEvaluators []pluginCore.Evaluator
	// Task and action information
	ActiveTasks     []*Task
	PendingActions  []actions.IAction
	NativeTokenInfo *types.TokenInfo
	ProviderStates  []*pluginCore.ProviderState
}

type Goal struct {
	ID          string
	Name        string
	Description string
	Weight      float64
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
	CurrentTask    *Task
	Goals          []Goal
	LastActionTime time.Time
}

// Main system routines
func (a *Agent) Start() error {
	a.logger.Info("Starting agent system")

	for _, account := range a.character.PriorityAccounts {
		_, err := a.stakeholders.FetchOrCreateStakeholder(
			a.ctx,
			account.ID,
			account.Platform,
			types.StakeholderTypePriority,
		)
		if err != nil {
			return err
		}
	}

	// Start periodic task evaluation
	go func() {
		a.runPeriodicEvaluation()
	}()

	// Start social media monitoring
	go func() {
		a.monitorSocialInputs()
	}()

	return nil
}

func (a *Agent) RegisterPlugin(p *plugin.Plugin) {
	// TODO: implement me
}

// Periodic task evaluation
func (a *Agent) runPeriodicEvaluation() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// a.evaluateAndExecuteTasks()
	for {
		select {
		case <-ticker.C:
			// TODO: enable execution
			// a.evaluateAndExecuteTasks()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *Agent) evaluateAndExecuteTasks() error {
	a.logger.Info("Evaluating and executing tasks")

	// Get current system state
	state := a.getCurrentState()

	tasks, _ := a.GenerateTasks(context.Background(), state)
	a.logger.Infof("Generated tasks: %d", len(tasks))

	for _, task := range tasks {
		// Check if stakeholder input is needed
		// if task.RequiresStakeholderInput {
		// 	a.requestStakeholderFeedback(task)
		// 	continue
		// }

		// Execute task
		_, err := a.ExecuteTask(a.ctx, task, state)
		if err != nil {
			return err
		}

		// Report results
		// a.reportTaskResults(task, result)
	}

	return nil
}

// In your agent_system.go
func (a *Agent) getCurrentState() *SystemState {
	pref, _ := a.stakeholders.GetAggregatedPreferences(a.ctx)

	nativeToken, _ := a.TokenManager.NativeTokenInfo(a.ctx)
	tasks, _ := a.taskManager.GetTasks(a.ctx)

	// Get plugin actions and provider states
	var pluginActions []actions.IAction
	var providerStates []*pluginCore.ProviderState
	var pluginEvaluators []pluginCore.Evaluator

	if a.pluginRegistry != nil {
		// Collect actions from plugins
		for _, plugin := range a.pluginRegistry.GetPlugins() {
			for _, action := range plugin.Actions() {
				adapter := pluginCore.NewActionAdapter(a.ctx, action)
				pluginActions = append(pluginActions, adapter)
			}
		}

		// Collect provider states
		for _, provider := range a.pluginRegistry.GetProviders() {
			if state, err := provider.GetProviderState(a.ctx); err == nil {
				providerStates = append(providerStates, state)
			} else {
				a.logger.Warnw("Failed to get provider state",
					"provider", provider.Name(),
					"error", err,
				)
			}
		}
	}

	// print all available actions
	for _, action := range pluginActions {
		a.logger.Infof("Available action: %s", action.Name())
	}

	// print all evaluators
	for _, evaluator := range pluginEvaluators {
		a.logger.Infof("Available evaluator: %s", evaluator.Name())
	}

	// print all provider states
	for _, state := range providerStates {
		a.logger.Infof("Provider state: %+v", state)
	}

	return &SystemState{
		Character:              a.character,
		AvailableTools:         a.toolManager.AvailableTools(),
		AvailableActions:       append(a.toolManager.AvailableActions(), pluginActions...),
		AvailableEvaluators:    pluginEvaluators,
		Timestamp:              time.Now(),
		AgentStates:            a.GetState(),
		StakeholderPreferences: pref,
		ActiveTasks:            tasks,
		NativeTokenInfo:        nativeToken,
		ProviderStates:         providerStates,
	}
}

// processCommand handles command messages starting with '/'
func (a *Agent) processCommand(msg *types.SocialMessage) error {
	a.logger.Infof("Processing command: %s", msg.Content)
	if a.CommandManager == nil {
		a.logger.Warn("Command manager is not initialized")
		return nil
	}

	// Get the command name (first word after /)
	args := strings.Fields(msg.Content)
	if len(args) == 0 {
		return nil
	}
	cmdName := args[0][1:] // Remove leading '/' from first word

	cmd, exists := a.CommandManager.Get(cmdName)
	if !exists {
		a.logger.Warnw("Command not found", "command", cmdName)
		a.socialClient.SendMessage(a.ctx, types.SocialMessage{
			Platform: msg.Platform,
			Type:     "Response",
			Content:  fmt.Sprintf("Command not found: `%s`. Use `/help` to see available commands.", cmdName),
			Metadata: msg.Metadata,
		})
		return nil
	}

	// Convert core.SocialMessage to types.SocialMessage
	typesMsg := &types.SocialMessage{
		Platform: msg.Platform,
		FromUser: msg.FromUser,
		Content:  msg.Content,
		Metadata: msg.Metadata,
	}

	if err := cmd.Execute(a.ctx, typesMsg); err != nil {
		a.logger.Errorw("Failed to execute command", "command", cmdName, "error", err)
		return err
	}

	a.socialClient.SendMessage(a.ctx, types.SocialMessage{
		Platform: msg.Platform,
		Type:     "Response",
		Content:  typesMsg.Content,
		Metadata: msg.Metadata,
	})

	return nil
}

// processMessage handles regular (non-command) messages
func (a *Agent) processMessage(msg *types.SocialMessage) error {
	var err error
	defer func() {
		if err != nil {
			a.logger.Errorw("Error processing message", "error", err)
			a.socialClient.SendMessage(a.ctx, types.SocialMessage{
				Platform: msg.Platform,
				Type:     "Response",
				Content:  "Something went wrong. Please try again later.",
				Metadata: msg.Metadata,
			})
		}
	}()

	state := a.getCurrentState()

	stakeholder, err := a.stakeholders.FetchOrCreateStakeholder(
		a.ctx,
		msg.FromUser,
		msg.Platform,
		types.StakeholderTypeUser,
	)
	if err != nil {
		a.logger.Errorw("Error fetching stakeholder", "error", err)
		return err
	}

	a.logger.Infof("Priority accounts: %t", stakeholder.Type == types.StakeholderTypePriority)

	balance, _ := a.TokenManager.FetchNativeTokenBalance(a.ctx, msg.FromUser, msg.Platform)
	if balance != nil {
		a.logger.Infof("Native token balance: %f", balance.Balance)
		stakeholder.TokenBalance = balance
	}

	processedMsg, err := a.cognitive.processMessage(a.ctx, state, msg, stakeholder)
	if err != nil {
		a.logger.Errorw("Error processing message", "error", err)
		return err
	}

	if processedMsg.ShouldGenerateAction {
		for _, action := range processedMsg.Actions {
			var actionImpl actions.IAction
			actionImpl, err := a.toolManager.GetAction(action.ActionType, action.ActionName)
			if err != nil {
				// If action not found in toolManager, try to find it in pluginRegistry
				if a.pluginRegistry != nil {
					for _, plugin := range a.pluginRegistry.GetPlugins() {
						for _, pluginAction := range plugin.Actions() {
							if pluginAction.Type() == action.ActionType && pluginAction.Name() == action.ActionName {
								actionImpl = pluginCore.NewActionAdapter(a.ctx, pluginAction)
								break
							}
						}
						if actionImpl != nil {
							break
						}
					}
				}

				if actionImpl == nil {
					a.logger.Errorw("Error getting action", "error", err)
					return err
				}
				a.logger.Infof("Action found in pluginRegistry: %s", actionImpl.Name())
			} else {
				a.logger.Infof("Action found in toolManager: %s", actionImpl.Name())
			}

			params, err := a.cognitive.generateActionParameters(a.ctx, state, msg, stakeholder, actionImpl)
			if err != nil {
				a.logger.Errorw("Error generating action parameters", "error", err)
				return err
			}

			if moreInfoNeeded, ok := params["more_info_needed"].(bool); ok && moreInfoNeeded {
				a.logger.Infof("More info needed, relying on message: %s", params["rely_message"])
				processedMsg.ResponseMsg = params["rely_message"].(string)
				processedMsg.ShouldReply = true
				continue
			}

			if err = a.executeAction(a.ctx, actionImpl, params); err != nil {
				a.logger.Errorw("Error executing action", "error", err)
				return err
			}
		}
	}

	a.logger.Infof("Processed message: %+v", processedMsg)
	err = a.stakeholders.AddHistoricalMsg(
		a.ctx,
		msg.FromUser,
		msg.Platform,
		[]string{
			fmt.Sprintf("%s: %s", msg.FromUser, msg.Content),
			fmt.Sprintf("%s: %s", state.Character.Name, processedMsg.ResponseMsg),
		},
	)
	if err != nil {
		a.logger.Errorw("Error adding historical message", "error", err)
		return err
	}

	if processedMsg.ShouldReply {
		// If we didn't send a response with analysis, send the original response
		a.socialClient.SendMessage(a.ctx, types.SocialMessage{
			Platform: msg.Platform,
			Type:     "Response",
			Content:  processedMsg.ResponseMsg,
			Metadata: msg.Metadata,
		})
	}

	return nil
}

// monitorSocialInputs handles incoming social media messages
func (a *Agent) monitorSocialInputs() {
	msgQueue := a.socialClient.GetMessageChannel()
	go a.socialClient.MonitorMessages(a.ctx)
	for {
		select {
		case msg := <-msgQueue:
			// Route message to appropriate handler based on type
			if len(msg.Content) > 0 && msg.Content[0] == '/' {
				if err := a.processCommand(&msg); err != nil {
					a.logger.Errorw("Error processing command", "error", err)
				}
			} else {
				if err := a.processMessage(&msg); err != nil {
					a.logger.Errorw("Error processing message", "error", err)
				}
			}
		case <-a.ctx.Done():
			return
		}
	}
}

// executeAction executes a generic action
func (a *Agent) executeAction(ctx context.Context, action actions.IAction, params map[string]interface{}) error {
	a.logger.Infow("Executing action", "type", action.Type(), "params", params)
	return action.Execute(ctx, params)
}

func (a *Agent) Shutdown(ctx context.Context) error {
	a.cancel()
	return nil
}

func NewAgent(config AgentConfig) (*Agent, error) {
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid agent config: %w", err)
	}

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()

	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		ID:             config.ID,
		character:      config.Character,
		cognitive:      NewCognitiveEngine(config.LLMClient, config.Model, config.Character, sugar, config.PromptTemplates),
		taskManager:    config.TaskManager,
		actionManager:  config.ActionManager,
		logger:         sugar,
		toolManager:    config.ToolsManager,
		stakeholders:   config.Stakeholders,
		TokenManager:   config.TokenManager,
		socialClient:   config.SocialClient,
		pluginRegistry: config.PluginRegistry,
		ctx:            ctx,
		cancel:         cancel,
		CommandManager: config.CommandManager,
	}

	return agent, nil
}

func (a *Agent) GenerateTasks(ctx context.Context, state *SystemState) ([]*Task, error) {
	tasks, err := a.cognitive.GenerateTasks(ctx, state)
	if err != nil {
		a.logger.Errorw("Failed to evaluate task", "error", err)
		return nil, err
	}

	return tasks.Tasks, nil
}

func (a *Agent) ExecuteTask(ctx context.Context, task *Task, state *SystemState) (*TaskResult, error) {
	// Generate actions using cognitive engine
	actionGen, err := a.cognitive.GenerateActions(ctx, task, state)
	if err != nil {
		return nil, fmt.Errorf("failed to generate actions: %w", err)
	}

	// Execute actions with continuous verification
	var results []error
	// for _, action := range actionGen.Actions {
	// Execute action
	// err := a.executeAction(ctx, action)
	// results = append(results, err)

	// results = append(results, result)

	// Update stakeholders on significant progress
	// if a.isSignificantProgress(result) {
	// 	a.notifyStakeholders(ctx, task, result)
	// }
	// }

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
