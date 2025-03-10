package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/internal/governance"

	"github.com/google/uuid"
)

// VoteAction represents the action for handling votes
type VoteAction struct {
	name        string
	description string
	registry    governance.Registry
	examples    []string
	similes     []string
}

// NewVoteAction creates a new vote action
func NewVoteAction(registry governance.Registry) *VoteAction {
	return &VoteAction{
		name:        "governance_vote",
		description: "Vote on governance proposals",
		registry:    registry,
		examples: []string{
			"Vote yes on proposal 123e4567-e89b-12d3-a456-426614174000",
			"Vote no on proposal 123e4567-e89b-12d3-a456-426614174000",
			"Get votes for proposal 123e4567-e89b-12d3-a456-426614174000",
			"Tally votes for proposal 123e4567-e89b-12d3-a456-426614174000",
		},
		similes: []string{
			"vote yes",
			"vote no",
			"approve proposal",
			"reject proposal",
			"support proposal",
			"oppose proposal",
			"get votes",
			"count votes",
			"tally votes",
		},
	}
}

// Execute executes the vote action with given parameters
func (a *VoteAction) Execute(ctx context.Context, params map[string]interface{}) error {
	action, ok := params["action"].(string)
	if !ok {
		return fmt.Errorf("action parameter is required")
	}

	switch action {
	case "vote":
		return a.handleVote(ctx, params)
	case "get_votes":
		return a.handleGetVotes(ctx, params)
	case "tally":
		return a.handleTallyVotes(ctx, params)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (a *VoteAction) handleVote(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	optionStr, ok := params["option"].(string)
	if !ok {
		return fmt.Errorf("option parameter is required")
	}

	userID, ok := params["user_id"].(string)
	if !ok {
		return fmt.Errorf("user_id parameter is required")
	}

	platform, ok := params["platform"].(string)
	if !ok {
		return fmt.Errorf("platform parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	var option governance.VoteOption
	switch strings.ToLower(optionStr) {
	case "yes":
		option = governance.VoteOptionYes
	case "no":
		option = governance.VoteOptionNo
	default:
		return fmt.Errorf("invalid option: %s, must be 'yes' or 'no'", optionStr)
	}

	if err := a.registry.CastVote(ctx, proposalID, userID, platform, option); err != nil {
		return fmt.Errorf("failed to cast vote: %v", err)
	}

	// Get current vote tally
	tally, err := a.registry.TallyVotes(ctx, proposalID)
	if err != nil {
		// Just log the error and continue
		fmt.Printf("Failed to tally votes: %v\n", err)
	}

	params["result"] = map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Vote cast successfully for proposal %s", proposalIDStr),
		"tally": map[string]interface{}{
			"yes": tally[governance.VoteOptionYes],
			"no":  tally[governance.VoteOptionNo],
		},
	}

	return nil
}

func (a *VoteAction) handleGetVotes(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	votes, err := a.registry.GetVotes(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get votes: %v", err)
	}

	result := make([]map[string]interface{}, 0, len(votes))
	for _, v := range votes {
		result = append(result, map[string]interface{}{
			"voter":      v.Voter,
			"option":     v.Option,
			"created_at": v.CreatedAt,
		})
	}

	params["result"] = map[string]interface{}{
		"success": true,
		"votes":   result,
	}

	return nil
}

func (a *VoteAction) handleTallyVotes(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	tally, err := a.registry.TallyVotes(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to tally votes: %v", err)
	}

	// Get proposal details
	proposal, err := a.registry.GetProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %v", err)
	}

	// Calculate percentages
	totalVotes := tally[governance.VoteOptionYes] + tally[governance.VoteOptionNo]
	var yesPercent, noPercent float64
	if totalVotes > 0 {
		yesPercent = float64(tally[governance.VoteOptionYes]) / float64(totalVotes) * 100
		noPercent = float64(tally[governance.VoteOptionNo]) / float64(totalVotes) * 100
	}

	params["result"] = map[string]interface{}{
		"success": true,
		"proposal": map[string]interface{}{
			"id":     proposal.ID.String(),
			"title":  proposal.Title,
			"status": proposal.Status,
		},
		"tally": map[string]interface{}{
			"total":       totalVotes,
			"yes":         tally[governance.VoteOptionYes],
			"no":          tally[governance.VoteOptionNo],
			"yes_percent": yesPercent,
			"no_percent":  noPercent,
		},
	}

	return nil
}

func (a *VoteAction) Name() string {
	return a.name
}

func (a *VoteAction) Description() string {
	return a.description
}

func (a *VoteAction) Type() string {
	return "governance"
}

func (a *VoteAction) GetExamples() []string {
	return a.examples
}

func (a *VoteAction) GetSimiles() []string {
	return a.similes
}

// ParametersPrompt returns a prompt for the action parameters
func (a *VoteAction) ParametersPrompt() string {
	return fmt.Sprintf("Please provide parameters for the %s action", a.name)
}

// Validate validates the action parameters
func (a *VoteAction) Validate(params map[string]interface{}) error {
	// Basic validation could be implemented here
	return nil
}
