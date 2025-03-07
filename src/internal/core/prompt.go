package core

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/conf"
	"github.com/carv-protocol/d.a.t.a/src/internal/plugins"
)

func generateTasksPromptFunc(systemState *SystemState, promptTemplate *conf.PromptTemplates) promptGeneratorFunc {
	return func(stepPurpose StepPurpose, steps []*ThoughtStep) string {
		switch stepPurpose {
		case PurposeInitial:
			return fmt.Sprintf(
				promptTemplate.ThoughtSteps[conf.ThoughtStepTypeTask].Initial,
				systemState.Character.Name,
			)
		case PurposeAnalysis:
			// Purpose Analysis: Evaluate the tasks that have been generated to assess their feasibility, risks, and alignment with goals.
			return fmt.Sprintf(
				promptTemplate.ThoughtSteps[conf.ThoughtStepTypeTask].Analysis,
				formatPreviousSteps(steps),
			)
		case PurposeReconsider:
			return fmt.Sprintf(
				promptTemplate.ThoughtSteps[conf.ThoughtStepTypeTask].Reconsider,
				formatPreviousSteps(steps),
			)
		case PurposeRefinement:
			// Purpose Refinement: Improve and polish the tasks based on analysis and feedback.
			return fmt.Sprintf(
				promptTemplate.ThoughtSteps[conf.ThoughtStepTypeTask].Refinement,
				formatPreviousSteps(steps),
			)
		case PurposeConcrete:
			// Purpose Concrete: Finalize the tasks into fully executable plans with precise actions.
			return fmt.Sprintf(
				promptTemplate.ThoughtSteps[conf.ThoughtStepTypeTask].Refinement,
				formatPreviousSteps(steps),
			)
		}
		return ""
	}
}

func generateActionsPromptFunc(systemState *SystemState, actions []actions.IAction, prompts *conf.PromptTemplates) promptGeneratorFunc {
	return func(stepPurpose StepPurpose, steps []*ThoughtStep) string {
		switch stepPurpose {
		case PurposeInitial:
			// Initial Action Generation
			actionDescriptions := ""
			for _, action := range actions {
				actionDescriptions += fmt.Sprintf("\n- **%s**: %s", action.Name(), action.Description())
			}

			return fmt.Sprintf(
				prompts.ThoughtSteps[conf.ThoughtStepTypeAction].Initial,
				actionDescriptions,
			)

		case PurposeExploration:
		case PurposeAnalysis:
		case PurposeReconsider:
		case PurposeRefinement:
		case PurposeConcrete:
		}

		return ""
	}
}

func formatMap(data map[string]interface{}) string {
	var result string
	for key, value := range data {
		result += fmt.Sprintf("%s: %v\n", key, value)
	}
	return result
}

func buildMessagePrompt(state *SystemState, msg *SocialMessage, stakeholder *Stakeholder, prompts *conf.PromptTemplates) string {
	template := prompts.Message.Analysis
	return fmt.Sprintf(
		template,
		msg.Platform,
		msg.FromUser,
		msg.Content,
		getHistoricalMessages(stakeholder),
		strings.Join(state.Character.Style.Tone, ", "),
		strings.Join(state.Character.MessageExamples, "\n"),
		formatActions(state.AvailableActions),
	)
}

func buildSystemPrompt(state *SystemState, stakeholder *Stakeholder, prompts *conf.PromptTemplates) string {
	// Get prompt templates from config
	baseTemplate := prompts.System.BaseTemplate
	infoFormat := prompts.System.InfoFormat

	// Format priority account info
	priorityAccountInfo := ""
	if stakeholder != nil && stakeholder.Type == StakeholderTypePriority {
		priorityAccountInfo = infoFormat["priority_account"]
	}

	// Format token balance info
	tokenBalanceInfo := ""
	if stakeholder != nil {
		if stakeholder.TokenBalance != nil {
			tokenBalanceInfo = fmt.Sprintf(
				infoFormat["token_balance_exists"],
				stakeholder.TokenBalance.Balance,
			)
		} else {
			tokenBalanceInfo = infoFormat["token_balance_missing"]
		}
	}

	// Format the final prompt using the template
	return fmt.Sprintf(
		baseTemplate,
		state.Character.Name,
		state.Character.System,
		strings.Join(state.Character.Bio, "\n"),
		strings.Join(state.Character.Lore, "\n"),
		formatProviderStates(state.ProviderStates),
		strings.Join(state.Character.Style.Constraints, "\n"),
		priorityAccountInfo,
		tokenBalanceInfo,
	)
}

func formatActions(actions []actions.IAction) string {
	var result string
	for _, action := range actions {
		result += fmt.Sprintf("- {Action Type: %s, Action Name: %s, Action Description: %s}\n", action.Type(), action.Name(), action.Description())
	}
	return result
}

func generateActionParametersPrompt(state *SystemState, msg *SocialMessage, stakeholder *Stakeholder, action actions.IAction, prompts *conf.PromptTemplates) string {
	// Create a prompt that explains all the possible types and asks for structured analysis
	template := prompts.Message.Action

	return fmt.Sprintf(
		template,
		msg.Platform,
		msg.Content,
		getHistoricalMessages(stakeholder),
		action.Name(),
		action.Description(),
		action.ParametersPrompt(),
	)
}

func getHistoricalMessages(stakeholder *Stakeholder) string {
	if stakeholder == nil {
		return ""
	}

	return strings.Join(stakeholder.HistoricalMsgs, ";")
}

func formatProviderStates(states []*plugins.ProviderState) string {
	if len(states) == 0 {
		return "No additional information available from providers"
	}

	var result string
	for _, state := range states {
		result += fmt.Sprintf("- **%s** (%s):\n", state.Name, state.Type)
		result += fmt.Sprintf("  - Status: %s\n", state.State)
		if state.Metadata != nil {
			result += "  - Details:\n"
			for key, value := range state.Metadata {
				result += fmt.Sprintf("    - %s: %v\n", formatKey(key), value)
			}
		}
		result += "\n"
	}
	return result
}

func formatKey(key string) string {
	// Convert snake_case or camelCase to Title Case
	words := strings.FieldsFunc(key, func(r rune) bool {
		return r == '_' || unicode.IsUpper(r)
	})
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	return strings.Join(words, " ")
}
