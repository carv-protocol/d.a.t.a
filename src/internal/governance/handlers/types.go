package handlers

import (
	"context"

	// "github.com/google/uuid"
	"github.com/carv-protocol/d.a.t.a/src/internal/governance"
)

// Handler defines the basic interface for a governance handler
type Handler interface {
	// Name returns the name of the handler
	Name() string
	// Description returns the description of the handler
	Description() string
	// Execute executes the handler with given context and parameters
	Execute(ctx context.Context, params map[string]interface{}) error
}

// Registry defines the interface for handler registry
type Registry interface {
	// Register registers a new handler
	Register(h Handler) error
	// Get returns a handler by name
	Get(name string) (Handler, bool)
	// ListHandlers returns all registered handlers
	ListHandlers() []Handler
}

// BaseHandler provides basic implementation for Handler interface
type BaseHandler struct {
	name        string
	description string
	registry    governance.Registry
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(name, description string, registry governance.Registry) BaseHandler {
	return BaseHandler{
		name:        name,
		description: description,
		registry:    registry,
	}
}

// Name returns the name of the handler
func (h BaseHandler) Name() string {
	return h.name
}

// Description returns the description of the handler
func (h BaseHandler) Description() string {
	return h.description
}
