package commands

import (
	"sync"

	"github.com/carv-protocol/d.a.t.a/src/internal/commands/handlers"
)

// Registry implements the handlers.Registry interface
type Registry struct {
	commands map[string]handlers.Command
	mu       sync.RWMutex
}

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

// Get get a command by name
func (r *Registry) Get(name string) (handlers.Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmd, exists := r.commands[name]
	return cmd, exists
}

// ListCommands list all registered commands
func (r *Registry) ListCommands() []handlers.Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cmds := make([]handlers.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}
