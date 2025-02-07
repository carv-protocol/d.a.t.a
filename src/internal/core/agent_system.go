package core

import (
	"context"
	"fmt"
	"plugin"
	"sync"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/token"
)

// Main system routines
func (a *Agent) Start() error {
	a.logger.Info("Starting agent system")

	for _, account := range a.character.PriorityAccounts {
		_, err := a.stakeholders.FetchOrCreateStakeholder(
			a.ctx,
			account.ID,
			account.Platform,
			token.StakeholderTypePriority,
		)
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup

	// Start periodic task evaluation
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.runPeriodicEvaluation()
	}()

	// Start social media monitoring
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.monitorSocialInputs()
	}()

	wg.Wait()
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

	// a.dataManager.FetchTokenInfo(a.ctx, )

	return &SystemState{
		Character:              a.character,
		AvailableTools:         a.toolManager.AvailableTools(),
		AvailableActions:       a.actionManager.GetAvailableActions(),
		Timestamp:              time.Now(),
		AgentStates:            a.GetState(),
		StakeholderPreferences: pref,
		ActiveTasks:            a.taskManager.GetTasks(a.ctx),
	}
}

// Social media monitoring
func (a *Agent) monitorSocialInputs() {
	msgQueue := a.socialClient.GetMessageChannel()
	// TODO graceful shutdown
	go a.socialClient.MonitorMessages(a.ctx)
	for {
		select {
		case msg := <-msgQueue:
			a.processMessage(&msg)
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *Agent) processMessage(msg *SocialMessage) error {
	state := a.getCurrentState()

	stakeholder, err := a.stakeholders.FetchOrCreateStakeholder(
		a.ctx,
		msg.FromUser,
		msg.Platform,
		token.StakeholderTypeUser,
	)
	if err != nil {
		a.logger.Errorw("Error fetching stakeholder", "error", err)
		return err
	}

	// a.logger.Infof("Historical message: %+v", stakeholder.HistoricalMsgs)
	a.logger.Infof("Priority accounts: %t", msg.FromUser, msg.Platform, stakeholder.Type == token.StakeholderTypePriority)

	// a.dataManager.FetchStakeholderInfo(a.ctx, stakeholder.ID)
	processedMsg, err := a.cognitive.processMessage(a.ctx, state, msg, stakeholder)
	if err != nil {
		a.logger.Errorw("Error processing message", "error", err)
		return err
	}

	// TODO fetch or create stakeholder
	// a.stakeholders.FetchOrCreateStakeholder(a.ctx, msg.FromUser)

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

	// TODO: process task generation and action taking
	if processedMsg.ShouldReply {
		// Send response
		a.socialClient.SendMessage(SocialMessage{
			Platform: msg.Platform,
			Type:     "Response",
			Content:  processedMsg.ResponseMsg,
			Metadata: msg.Metadata,
		})
	}

	return nil
}

func (a *Agent) Shutdown(ctx context.Context) {
	a.cancel()
}
