package core

import (
	"context"
	"fmt"
	"sync"
)

// Registry manages plugin registration and lifecycle
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
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
func (r *Registry) GetActions() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actions []Action
	for _, p := range r.plugins {
		actions = append(actions, p.Actions()...)
	}
	return actions
}

// InitPlugin initializes a specific plugin with the given options
func (r *Registry) InitPlugin(ctx context.Context, name string, opts map[string]interface{}) error {
	r.mu.RLock()
	plugin, exists := r.plugins[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	return plugin.Init(ctx, opts)
}

// InitAll initializes all registered plugins with the given options
func (r *Registry) InitAll(ctx context.Context, opts map[string]interface{}) error {
	r.mu.RLock()
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	r.mu.RUnlock()

	for _, p := range plugins {
		if err := p.Init(ctx, opts); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", p.Name(), err)
		}
	}
	return nil
}

// StartAll starts all registered plugins
func (r *Registry) StartAll(ctx context.Context) error {
	r.mu.RLock()
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	r.mu.RUnlock()

	for _, p := range plugins {
		if err := p.Start(ctx); err != nil {
			return fmt.Errorf("failed to start plugin %s: %w", p.Name(), err)
		}
	}
	return nil
}

// StopAll stops all registered plugins
func (r *Registry) StopAll(ctx context.Context) error {
	r.mu.RLock()
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}
	r.mu.RUnlock()

	var errs []error
	for _, p := range plugins {
		if err := p.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop plugin %s: %w", p.Name(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping plugins: %v", errs)
	}
	return nil
}

// GetProviders returns all providers from all plugins
func (r *Registry) GetProviders() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []Provider
	for _, p := range r.plugins {
		providers = append(providers, p.Providers()...)
	}
	return providers
}

// GetEvaluators returns all evaluators from all plugins
func (r *Registry) GetEvaluators() []Evaluator {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var evaluators []Evaluator
	for _, p := range r.plugins {
		evaluators = append(evaluators, p.Evaluators()...)
	}
	return evaluators
}

// GetServices returns all services from all plugins
func (r *Registry) GetServices() []Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var services []Service
	for _, p := range r.plugins {
		services = append(services, p.Services()...)
	}
	return services
}

// GetClients returns all clients from all plugins
func (r *Registry) GetClients() []Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var clients []Client
	for _, p := range r.plugins {
		clients = append(clients, p.Clients()...)
	}
	return clients
}
