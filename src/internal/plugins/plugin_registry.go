package plugins

import (
	"fmt"
	"sync"

	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
)

// Registry manages plugin registration and lifecycle
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

func NewPluginRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register registers a plugin with the registry
func (r *Registry) Register(p Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if p == nil {
		return fmt.Errorf("cannot register nil plugin")
	}

	name := p.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	r.plugins[name] = p
	return nil
}

// GetPlugin returns a plugin by name
func (r *Registry) GetPlugin(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.plugins[name]
	return p, exists
}

// GetPlugins returns all registered plugins
func (r *Registry) GetPlugins() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

// GetActions returns all actions from all plugins
func (r *Registry) GetActions() []actions.IAction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actions []actions.IAction
	for _, p := range r.plugins {
		actions = append(actions, p.Actions()...)
	}
	return actions
}

func (r *Registry) GetProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []Provider
	for _, p := range r.plugins {
		providers = append(providers, p.Providers()...)
	}

	return providers
}
