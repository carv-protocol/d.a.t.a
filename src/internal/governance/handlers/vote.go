package handlers

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
	"github.com/google/uuid"
)

// VoteHandler handles voting related operations
type VoteHandler struct {
	BaseHandler
}

// NewVoteHandler creates a new vote handler
func NewVoteHandler(registry governance.Registry) *VoteHandler {
	return &VoteHandler{
		BaseHandler: NewBaseHandler("vote", "Manage governance voting", registry),
	}
}

// Execute executes the vote handler with given parameters
func (h *VoteHandler) Execute(ctx context.Context, params map[string]interface{}) error {
	action, ok := params["action"].(string)
	if !ok {
		return fmt.Errorf("action parameter is required")
	}

	switch action {
	case "cast":
		return h.handleCastVote(ctx, params)
	case "get":
		return h.handleGetVotes(ctx, params)
	case "tally":
		return h.handleTallyVotes(ctx, params)
	case "result":
		return h.handleGetResult(ctx, params)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (h *VoteHandler) handleCastVote(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	voter, ok := params["voter"].(string)
	if !ok {
		return fmt.Errorf("voter parameter is required")
	}

	platform, ok := params["platform"].(string)
	if !ok {
		return fmt.Errorf("platform parameter is required")
	}

	optionStr, ok := params["option"].(string)
	if !ok {
		return fmt.Errorf("option parameter is required")
	}

	option := governance.VoteOption(optionStr)
	return h.registry.CastVote(ctx, proposalID, voter, platform, option)
}

func (h *VoteHandler) handleGetVotes(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	_, err = h.registry.GetVotes(ctx, proposalID)
	return err
}

func (h *VoteHandler) handleTallyVotes(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	_, err = h.registry.TallyVotes(ctx, proposalID)
	return err
}

func (h *VoteHandler) handleGetResult(ctx context.Context, params map[string]interface{}) error {
	proposalIDStr, ok := params["proposal_id"].(string)
	if !ok {
		return fmt.Errorf("proposal_id parameter is required")
	}

	proposalID, err := uuid.Parse(proposalIDStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	_, err = h.registry.GetProposalResult(ctx, proposalID)
	return err
}
