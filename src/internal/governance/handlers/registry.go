package handlers

import (
	"sync"
)

// HandlerRegistry implements the Registry interface
type HandlerRegistry struct {
	handlers map[string]Handler
	mu       sync.RWMutex
}

// NewHandlerRegistry creates a new handler registry
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers: make(map[string]Handler),
	}
}

// Register registers a new handler
func (r *HandlerRegistry) Register(h Handler) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[h.Name()] = h
	return nil
}

// Get returns a handler by name
func (r *HandlerRegistry) Get(name string) (Handler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, exists := r.handlers[name]
	return h, exists
}

// ListHandlers returns all registered handlers
func (r *HandlerRegistry) ListHandlers() []Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handlers := make([]Handler, 0, len(r.handlers))
	for _, h := range r.handlers {
		handlers = append(handlers, h)
	}
	return handlers
}