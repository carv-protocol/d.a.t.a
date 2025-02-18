package core

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"

	"go.uber.org/zap"
)

type promptGeneratorFunc func(StepPurpose, []*ThoughtStep) string

type StepPurpose string

const (
	PurposeInitial     StepPurpose = "initial"
	PurposeExploration StepPurpose = "exploration"
	PurposeAnalysis    StepPurpose = "analysis"
	PurposeReconsider  StepPurpose = "reconsider"
	PurposeRefinement  StepPurpose = "refinement"
	PurposeConcrete    StepPurpose = "concrete"
)

type CognitiveEngine struct {
	llm           llm.Client
	model         string
	maxSteps      int
	minConfidence float64
	character     *characters.Character
	logger        *zap.SugaredLogger
}

type CognitiveConfig struct {
	NumIterations      int
	SamplesPerBatch    int
	MinRewardThreshold float64
	Temperature        float64
	MaxChainLength     int
	StabilityWindow    int
}

// ThoughtChain represents a sequence of reasoning steps
type ThoughtChain struct {
	Steps []*ThoughtStep
	// Confidence      float64
	Reflection      string
	FinalConclusion string
	Timestamp       time.Time
}

// ThoughtStep represents a single step in the reasoning process
type ThoughtStep struct {
	Type         string
	Content      string // The actual thought content
	RawLLMOutput string // Original LLM output for analysis
	Confidence   float64
	Evidence     []string
	// Verification         string
	Alternatives         []string
	ContributesToOutcome bool
	Purpose              StepPurpose
	Metadata             map[string]interface{}
	Timestamp            time.Time
}

func NewCognitiveEngine(llmClient llm.Client, model string, character *characters.Character, logger *zap.SugaredLogger) *CognitiveEngine {
	return &CognitiveEngine{
		llm:           llmClient,
		model:         model,
		maxSteps:      3,
		minConfidence: 0.7,
		character:     character,
		logger:        logger,
	}
}

// GenerateThoughtChain creates a DeepSeek-style reasoning chain
func (e *CognitiveEngine) GenerateThoughtChain(
	ctx context.Context,
	state *SystemState,
	input interface{},
	prefs map[string]interface{},
	promptGenerator promptGeneratorFunc,
) (*ThoughtChain, error) {
	e.logger.Info("Generating thought chain")
	chain := &ThoughtChain{
		Steps:     make([]*ThoughtStep, 0),
		Timestamp: time.Now(),
	}

	// Generate reasoning steps
	for i := 0; i < e.maxSteps; i++ {
		// Determine appropriate step purpose based on progress
		purpose := e.determineStepPurpose(i)

		step, err := e.generateThoughtStep(ctx, state, chain, purpose, promptGenerator)
		if err != nil {
			return nil, err
		}

		// Detect "aha moment"
		if AhaMomentDetection := e.detectAhaMoment(
			ctx, step, chain.Steps, step.Alternatives, prefs,
		); purpose != PurposeConcrete && AhaMomentDetection.Triggered {
			// Generate reconsideration step
			step, err = e.generateThoughtStep(ctx, state, chain, PurposeReconsider, promptGenerator)
			if err != nil {
				return nil, err
			}
		}

		e.logger.Infof("Generated step: %d, %s", i, step.Content)
		chain.Steps = append(chain.Steps, step)

		// Check if we need more steps
		if e.isConclusive(chain) {
			break
		}
	}

	return chain, nil
}

// determineStepPurpose decides appropriate purpose for current step
func (e *CognitiveEngine) determineStepPurpose(stepIndex int) StepPurpose {
	if stepIndex == 0 {
		return PurposeInitial
	}
	if stepIndex == e.maxSteps-1 {
		return PurposeConcrete
	}

	totalSteps := float64(e.maxSteps)
	progress := float64(stepIndex+1) / totalSteps

	switch {
	case progress < 0.3:
		return PurposeExploration
	case progress < 0.5:
		return PurposeAnalysis
	case progress < 0.7:
		return PurposeRefinement
	default:
		return PurposeConcrete
	}
}

// doesStepContributeToOutcome determines if step contributes to final actions/tasks
func (e *CognitiveEngine) doesStepContributeToOutcome(purpose StepPurpose, chain *ThoughtChain) bool {
	// Concrete steps always contribute
	if purpose == PurposeConcrete {
		return true
	}

	// Reconsideration steps that improve the solution contribute
	if purpose == PurposeReconsider {
		return true
	}

	// Late refinement steps often contribute
	if purpose == PurposeRefinement && len(chain.Steps) > 5 {
		return true
	}

	return false
}

func formatPreviousSteps(steps []*ThoughtStep) string {
	if len(steps) == 0 {
		return "No previous steps"
	}

	var formatted string
	for i, step := range steps {
		formatted += fmt.Sprintf("Step %d (%s):\n%s\n\n",
			i+1, step.Type, step.Content)
	}
	return formatted
}

// GenerateActions uses chain-of-thought for action planning
func (e *CognitiveEngine) GenerateActions(
	ctx context.Context,
	task *Task,
	state *SystemState,
) (*ActionGeneration, error) {
	// Build action context
	actionContext := map[string]interface{}{
		"task":        task,
		"preferences": state.StakeholderPreferences,
		"goal":        "generate detailed action plan",
	}

	// Generate thought chain for action planning
	chain, err := e.GenerateThoughtChain(
		ctx,
		state,
		actionContext,
		state.StakeholderPreferences,
		generateActionsPromptFunc(state, task, state.AvailableActions),
	)
	if err != nil {
		return nil, err
	}

	// Convert thought chain to actions
	actions, _ := convertThoughtChainToActions(chain)

	return &ActionGeneration{
		Actions: actions,
		Chain:   chain,
	}, nil
}

// GenerateTasks uses chain-of-thought for tasks planning
func (e *CognitiveEngine) GenerateTasks(
	ctx context.Context,
	state *SystemState,
) (*TaskGeneration, error) {
	// Build action context
	taskContext := map[string]interface{}{
		"state": state,
		"goal":  "generate detailed tasks plan",
	}

	// Generate thought chain for action planning
	chain, err := e.GenerateThoughtChain(
		ctx,
		state,
		taskContext,
		state.StakeholderPreferences,
		generateTasksPromptFunc(state))
	if err != nil {
		return nil, err
	}

	// Convert thought chain to actions
	task, err := convertThoughtChainToTasks(chain)
	if err != nil {
		return nil, err
	}

	return &TaskGeneration{
		Tasks: []*Task{task},
		Chain: chain,
	}, nil
}

func (e *CognitiveEngine) generateThoughtStep(
	ctx context.Context,
	state *SystemState,
	chain *ThoughtChain,
	purpose StepPurpose,
	promptGenerator func(StepPurpose, []*ThoughtStep) string,
) (*ThoughtStep, error) {
	prompt := promptGenerator(purpose, chain.Steps)

	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: e.model,
		Messages: []llm.Message{
			{Role: "system", Content: buildSystemPrompt(state, nil)},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	return &ThoughtStep{
		// Core reasoning content
		Content:              extractThinkingContent(response),
		Evidence:             extractEvidence(response),
		Alternatives:         extractAlternatives(response),
		Purpose:              purpose,
		ContributesToOutcome: e.doesStepContributeToOutcome(purpose, chain),
	}, nil
}

// isConclusive determines if the reasoning chain has reached a satisfactory conclusion
func (e *CognitiveEngine) isConclusive(chain *ThoughtChain) bool {
	// Check minimum confidence threshold
	// if chain.Confidence < e.minConfidence {
	// 	return false
	// }

	// Must have at least one step
	if len(chain.Steps) == 0 {
		return false
	}

	// Check if we've addressed all key aspects
	// aspectsCovered := e.checkAspectsCoverage(chain)
	// if !aspectsCovered {
	// 	return false
	// }

	// Verify last step completion
	lastStep := chain.Steps[len(chain.Steps)-1]

	return lastStep.Purpose == PurposeConcrete
}

// Helper functions

func (e *CognitiveEngine) identifyLogicalIssues(thinking string) []string {
	var issues []string

	// Common logical fallacies and issues to check
	checks := map[string]string{
		"circular_reasoning":   `\b(because.*therefore.*because|therefore.*because.*therefore)\b`,
		"false_assumption":     `\b(must|always|never|everyone|nobody)\b`,
		"causal_fallacy":       `\b(leads to|causes|results in)\b`,
		"hasty_generalization": `\b(all|none|every|no one)\b`,
	}

	for issueType, pattern := range checks {
		if strings.Contains(thinking, pattern) {
			issues = append(issues, issueType)
		}
	}

	return issues
}

func (e *CognitiveEngine) evaluateAlternative(alternative string) float64 {
	var score float64 = 1.0

	// Evaluate completeness
	if !strings.Contains(alternative, "benefits") || !strings.Contains(alternative, "drawbacks") {
		score *= 0.8
	}

	// Check for concrete steps
	if !containsConcreteSteps(alternative) {
		score *= 0.7
	}

	// Assess feasibility
	if !assessFeasibility(alternative) {
		score *= 0.6
	}

	return score
}

// Utility functions

func parseAlternatives(response string) []string {
	// Extract alternatives between <think> tags
	alternatives := make([]string, 0)

	// Split response by <think> tags
	parts := strings.Split(response, "<think>")
	for _, part := range parts[1:] { // Skip first empty part
		if idx := strings.Index(part, "</think>"); idx != -1 {
			alt := strings.TrimSpace(part[:idx])
			alternatives = append(alternatives, alt)
		}
	}

	return alternatives
}

func containsConcreteSteps(alternative string) bool {
	// Check for numbered steps or action words
	return strings.Contains(alternative, "1.") ||
		strings.Contains(alternative, "First") ||
		strings.Contains(alternative, "Initially") ||
		strings.Contains(alternative, "Start by")
}

func assessFeasibility(alternative string) bool {
	// Check for implementation details and resource considerations
	return strings.Contains(alternative, "implement") ||
		strings.Contains(alternative, "resource") ||
		strings.Contains(alternative, "require") ||
		strings.Contains(alternative, "need")
}

func containsAspect(step string, aspect string) bool {
	aspectPatterns := map[string][]string{
		"problem_definition": {"problem", "challenge", "objective", "goal"},
		"methodology":        {"method", "approach", "strategy", "process"},
		"validation":         {"verify", "validate", "check", "confirm"},
		"risks":              {"risk", "challenge", "issue", "concern"},
		"outcomes":           {"result", "outcome", "impact", "effect"},
	}

	patterns, exists := aspectPatterns[aspect]
	if !exists {
		return false
	}

	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(step), pattern) {
			return true
		}
	}
	return false
}

func (e *CognitiveEngine) processMessage(
	ctx context.Context,
	state *SystemState,
	msg *SocialMessage,
	stakeholder *Stakeholder,
) (*ProcessedMessage, error) {
	prompt := buildMessagePrompt(state, msg, stakeholder)
	// Get LLM's analysis
	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: e.model,
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: buildSystemPrompt(state, stakeholder),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Parse LLM response into ProcessedMessage
	return ParseAnalysis(response)
}

func (e *CognitiveEngine) generateActionParameters(
	ctx context.Context,
	state *SystemState,
	msg *SocialMessage,
	stakeholder *Stakeholder,
	action actions.IAction,
) (map[string]interface{}, error) {
	prompt := generateActionParametersPrompt(state, msg, stakeholder, action)
	response, err := e.llm.CreateCompletion(ctx, llm.CompletionRequest{
		Model: e.model,
		Messages: []llm.Message{
			{Role: "system", Content: buildSystemPrompt(state, stakeholder)},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	parsedResponse, err := parseActionParameters(response)
	if err != nil {
		return nil, err
	}
	return parsedResponse, nil
}

// Helper functions
// ExtractThinkingContent extracts the core reasoning content from an LLM response.
func extractThinkingContent(response string) string {
	// Define a regex pattern to capture content within <think> tags
	pattern := `<think>(.*?)</think>`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(response)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// If no matches, return the response as-is, assuming no <think> tags were used
	return strings.TrimSpace(response)
}

func extractEvidence(response string) []string {
	// TODO: implement me
	// Extract evidence from response
	// Implementation details...
	return nil
}

func extractAlternatives(response string) []string {
	// TODO: implement me
	// Extract evidence from response
	// Implementation details...
	return nil
}

func extractAnwser(response string) []string {
	// TODO: implement me
	// Extract evidence from response
	// Implementation details...
	return nil
}

func calculateConfidence(response string) float64 {
	// TODO: implement me
	// Calculate confidence based on response
	// Implementation details...
	return 0.0
}

func generateAlternativeApproach(chain *ThoughtChain) string {
	// TODO: implement me
	// Generate alternative approach based on current chain
	// Implementation details...
	return ""
}

func ParseAnalysis(response string) (*ProcessedMessage, error) {
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var processedMsg ProcessedMessage
	if err := json.Unmarshal([]byte(response), &processedMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &processedMsg, nil
}

func parseActionParameters(response string) (map[string]interface{}, error) {
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(response), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return params, nil
}
