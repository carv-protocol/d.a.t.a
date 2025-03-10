package actions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/internal/governance"

	"github.com/google/uuid"
)

// ProposalAction represents the action for handling proposals
type ProposalAction struct {
	name        string
	description string
	registry    governance.Registry
	examples    []string
	similes     []string
}

// NewProposalAction creates a new proposal action
func NewProposalAction(registry governance.Registry) *ProposalAction {
	return &ProposalAction{
		name:        "governance_proposal",
		description: "Create and manage governance proposals",
		registry:    registry,
		examples: []string{
			"Create a proposal to increase token supply",
			"List all active proposals",
			"Get details of proposal 123e4567-e89b-12d3-a456-426614174000",
			"Set minimum balance for voting to 1000 tokens",
		},
		similes: []string{
			"create proposal",
			"new proposal",
			"submit proposal",
			"list proposals",
			"show proposals",
			"view proposal",
			"proposal details",
		},
	}
}

// Execute executes the proposal action with given parameters
func (a *ProposalAction) Execute(ctx context.Context, params map[string]interface{}) error {
	action, ok := params["action"].(string)
	if !ok {
		return fmt.Errorf("action parameter is required")
	}

	switch action {
	case "create":
		return a.handleCreate(ctx, params)
	case "list":
		return a.handleList(ctx, params)
	case "get":
		return a.handleGet(ctx, params)
	case "admin":
		return a.handleAdmin(ctx, params)
	case "force-end":
		return a.handleForceEnd(ctx, params)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (a *ProposalAction) handleCreate(ctx context.Context, params map[string]interface{}) error {
	title, ok := params["title"].(string)
	if !ok {
		return fmt.Errorf("title parameter is required")
	}

	description, ok := params["description"].(string)
	if !ok {
		return fmt.Errorf("description parameter is required")
	}

	userID, ok := params["user_id"].(string)
	if !ok {
		return fmt.Errorf("user_id parameter is required")
	}

	platform, ok := params["platform"].(string)
	if !ok {
		return fmt.Errorf("platform parameter is required")
	}

	// Get duration in seconds, default to 7 days
	durationStr, ok := params["duration"].(string)
	var durationSec int64 = 7 * 24 * 60 * 60 // 7 days in seconds
	if ok {
		var err error
		durationSec, err = strconv.ParseInt(durationStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid duration: %v", err)
		}
	}

	// Calculate start and end times
	now := time.Now().Unix()
	endTime := now + durationSec

	// Create proposal
	proposal, err := a.registry.CreateProposal(ctx, title, description, userID, platform, now, endTime)
	if err != nil {
		return fmt.Errorf("failed to create proposal: %v", err)
	}

	// Store result in params
	params["result"] = map[string]interface{}{
		"success":     true,
		"proposal_id": proposal.ID.String(),
		"title":       proposal.Title,
		"description": proposal.Description,
		"creator":     proposal.Creator,
		"start_time":  proposal.StartTime,
		"end_time":    proposal.EndTime,
		"status":      proposal.Status,
	}
	return nil
}

func (a *ProposalAction) handleList(ctx context.Context, params map[string]interface{}) error {
	statusStr, ok := params["status"].(string)
	if !ok {
		statusStr = "active" // Default to active proposals
	}

	var status governance.ProposalStatus
	switch strings.ToLower(statusStr) {
	case "active":
		status = governance.ProposalStatusActive
	case "pending":
		status = governance.ProposalStatusPending
	case "passed":
		status = governance.ProposalStatusPassed
	case "rejected":
		status = governance.ProposalStatusRejected
	case "all":
		status = "" // Empty string means all proposals
	default:
		return fmt.Errorf("invalid status: %s", statusStr)
	}

	proposals, err := a.registry.ListProposals(ctx, status)
	if err != nil {
		return fmt.Errorf("failed to list proposals: %v", err)
	}

	result := make([]map[string]interface{}, 0, len(proposals))
	for _, p := range proposals {
		result = append(result, map[string]interface{}{
			"id":          p.ID.String(),
			"title":       p.Title,
			"description": p.Description,
			"creator":     p.Creator,
			"start_time":  p.StartTime,
			"end_time":    p.EndTime,
			"status":      p.Status,
		})
	}

	// Store result in params
	params["result"] = map[string]interface{}{
		"success":   true,
		"proposals": result,
	}
	return nil
}

func (a *ProposalAction) handleGet(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	proposal, err := a.registry.GetProposal(ctx, proposalID)
	if err != nil {
		return fmt.Errorf("failed to get proposal: %v", err)
	}

	// Get votes
	tally, err := a.registry.TallyVotes(ctx, proposalID)
	if err != nil {
		// Just log the error and continue
		fmt.Printf("Failed to tally votes: %v\n", err)
	}

	// Store result in params
	params["result"] = map[string]interface{}{
		"success":     true,
		"id":          proposal.ID.String(),
		"title":       proposal.Title,
		"description": proposal.Description,
		"creator":     proposal.Creator,
		"start_time":  proposal.StartTime,
		"end_time":    proposal.EndTime,
		"status":      proposal.Status,
		"votes": map[string]interface{}{
			"yes": tally[governance.VoteOptionYes],
			"no":  tally[governance.VoteOptionNo],
		},
	}
	return nil
}

func (a *ProposalAction) handleAdmin(ctx context.Context, params map[string]interface{}) error {
	adminAction, ok := params["admin_action"].(string)
	if !ok {
		return fmt.Errorf("admin_action parameter is required")
	}

	switch adminAction {
	case "set_admin":
		userID, ok := params["user_id"].(string)
		if !ok {
			return fmt.Errorf("user_id parameter is required")
		}

		if err := a.registry.SetAdmin(userID); err != nil {
			return fmt.Errorf("failed to set admin: %v", err)
		}

		params["result"] = map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("User %s is now an admin", userID),
		}
		return nil

	case "remove_admin":
		userID, ok := params["user_id"].(string)
		if !ok {
			return fmt.Errorf("user_id parameter is required")
		}

		if err := a.registry.RemoveAdmin(userID); err != nil {
			return fmt.Errorf("failed to remove admin: %v", err)
		}

		params["result"] = map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("User %s is no longer an admin", userID),
		}
		return nil

	case "set_min_balance":
		balanceStr, ok := params["balance"].(string)
		if !ok {
			return fmt.Errorf("balance parameter is required")
		}

		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return fmt.Errorf("invalid balance: %v", err)
		}

		platform, ok := params["platform"].(string)
		if ok {
			// Set platform-specific minimum balance
			if err := a.registry.SetMinTokenBalanceForPlatform(platform, balance); err != nil {
				return fmt.Errorf("failed to set minimum balance for platform %s: %v", platform, err)
			}

			params["result"] = map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Minimum token balance for platform %s set to %f", platform, balance),
			}
			return nil
		}

		// Set global minimum balance
		if err := a.registry.SetMinTokenBalance(balance); err != nil {
			return fmt.Errorf("failed to set minimum balance: %v", err)
		}

		params["result"] = map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Minimum token balance set to %f", balance),
		}
		return nil

	default:
		return fmt.Errorf("unknown admin action: %s", adminAction)
	}
}

func (a *ProposalAction) handleForceEnd(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	statusStr, ok := params["status"].(string)
	if !ok {
		return fmt.Errorf("status parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	var status governance.ProposalStatus
	switch strings.ToLower(statusStr) {
	case "passed":
		status = governance.ProposalStatusPassed
	case "rejected":
		status = governance.ProposalStatusRejected
	default:
		return fmt.Errorf("invalid status: %s, must be 'passed' or 'rejected'", statusStr)
	}

	if err := a.registry.ForceEndProposal(proposalID, status); err != nil {
		return fmt.Errorf("failed to force end proposal: %v", err)
	}

	params["result"] = map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Proposal %s has been force ended with status %s", proposalIDStr, status),
	}
	return nil
}

func (a *ProposalAction) Name() string {
	return a.name
}

func (a *ProposalAction) Description() string {
	return a.description
}

func (a *ProposalAction) Type() string {
	return "governance"
}

func (a *ProposalAction) GetExamples() []string {
	return a.examples
}

func (a *ProposalAction) GetSimiles() []string {
	return a.similes
}

// ParametersPrompt returns a prompt for the action parameters
func (a *ProposalAction) ParametersPrompt() string {
	return fmt.Sprintf("Please provide parameters for the %s action", a.name)
}

// Validate validates the action parameters
func (a *ProposalAction) Validate(params map[string]interface{}) error {
	// Basic validation could be implemented here
	return nil
}
