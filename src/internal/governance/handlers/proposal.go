package handlers

import (
	"context"
	"fmt"

	"github.com/carv-protocol/d.a.t.a/src/internal/governance"

	"github.com/google/uuid"
)

// ProposalHandler handles proposal related operations
type ProposalHandler struct {
	BaseHandler
}

// NewProposalHandler creates a new proposal handler
func NewProposalHandler(registry governance.Registry) *ProposalHandler {
	return &ProposalHandler{
		BaseHandler: NewBaseHandler("proposal", "Manage governance proposals", registry),
	}
}

// Execute executes the proposal handler with given parameters
func (h *ProposalHandler) Execute(ctx context.Context, params map[string]interface{}) error {
	action, ok := params["action"].(string)
	if !ok {
		return fmt.Errorf("action parameter is required")
	}

	switch action {
	case "create":
		return h.handleCreate(ctx, params)
	case "get":
		return h.handleGet(ctx, params)
	case "list":
		return h.handleList(ctx, params)
	case "update_status":
		return h.handleUpdateStatus(ctx, params)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (h *ProposalHandler) handleCreate(ctx context.Context, params map[string]interface{}) error {
	title, ok := params["title"].(string)
	if !ok {
		return fmt.Errorf("title parameter is required")
	}

	description, ok := params["description"].(string)
	if !ok {
		return fmt.Errorf("description parameter is required")
	}

	creator, ok := params["creator"].(string)
	if !ok {
		return fmt.Errorf("creator parameter is required")
	}

	platform, ok := params["platform"].(string)
	if !ok {
		return fmt.Errorf("platform parameter is required")
	}

	startTime, ok := params["start_time"].(int64)
	if !ok {
		return fmt.Errorf("start_time parameter is required")
	}

	endTime, ok := params["end_time"].(int64)
	if !ok {
		return fmt.Errorf("end_time parameter is required")
	}

	_, err := h.registry.CreateProposal(ctx, title, description, creator, platform, startTime, endTime)
	return err
}

func (h *ProposalHandler) handleGet(ctx context.Context, params map[string]interface{}) error {
	idStr, ok := params["id"].(string)
	if !ok {
		return fmt.Errorf("id parameter is required")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	_, err = h.registry.GetProposal(ctx, id)
	return err
}

func (h *ProposalHandler) handleList(ctx context.Context, params map[string]interface{}) error {
	status, _ := params["status"].(string)
	_, err := h.registry.ListProposals(ctx, governance.ProposalStatus(status))
	return err
}

func (h *ProposalHandler) handleUpdateStatus(ctx context.Context, params map[string]interface{}) error {
	idStr, ok := params["id"].(string)
	if !ok {
		return fmt.Errorf("id parameter is required")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return fmt.Errorf("invalid proposal id: %v", err)
	}

	status, ok := params["status"].(string)
	if !ok {
		return fmt.Errorf("status parameter is required")
	}

	return h.registry.UpdateProposalStatus(ctx, id, governance.ProposalStatus(status))
}
