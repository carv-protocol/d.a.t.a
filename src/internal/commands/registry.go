package commands

import (
	"sync"

	"github.com/carv-protocol/d.a.t.a/src/internal/commands/handlers"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
)

// Registry implements both handlers.Registry and core.CommandRegistry interfaces
type Registry struct {
	commands map[string]handlers.Command
	mu       sync.RWMutex
}

// HandlerRegistry is used to implement handlers.Registry
type HandlerRegistry Registry

// NewRegistry create a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]handlers.Command),
	}
}

// Register register a command
func (r *Registry) Register(cmd handlers.Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name()] = cmd
	return nil
}

// Get get a command by name, implements core.CommandRegistry
func (r *Registry) Get(name string) (core.CommandInterface, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, exists := r.commands[name]
	return cmd, exists
}

// Get get a command by name, implements handlers.Registry
func (r *HandlerRegistry) Get(name string) (handlers.Command, bool) {
	reg := (*Registry)(r)
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	cmd, exists := reg.commands[name]
	return cmd, exists
}

// Register register a command
func (r *HandlerRegistry) Register(cmd handlers.Command) error {
	reg := (*Registry)(r)
	return reg.Register(cmd)
}

// ListCommands list all registered commands
func (r *HandlerRegistry) ListCommands() []handlers.Command {
	reg := (*Registry)(r)
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	cmds := make([]handlers.Command, 0, len(reg.commands))
	for _, cmd := range reg.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}

// AsHandlerRegistry returns the registry as a handlers.Registry
func (r *Registry) AsHandlerRegistry() handlers.Registry {
	return (*HandlerRegistry)(r)
}
