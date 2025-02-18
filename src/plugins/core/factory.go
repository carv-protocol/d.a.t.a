package core

import (
	"fmt"
	"sync"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
)

// PluginFactory is a function type that creates a new plugin instance
type PluginFactory func(llmClient llm.Client) Plugin

// pluginRegistry stores registered plugin factories
var (
	pluginFactories = make(map[string]PluginFactory)
	registryMutex   sync.RWMutex
)

// RegisterPlugin registers a plugin factory for the given plugin name
func RegisterPlugin(name string, factory PluginFactory) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if factory == nil {
		panic(fmt.Sprintf("plugin factory for %s is nil", name))
	}

	if _, exists := pluginFactories[name]; exists {
		panic(fmt.Sprintf("plugin %s already registered", name))
	}

	pluginFactories[name] = factory
}

// CreatePlugin creates a new plugin instance using the registered factory
func CreatePlugin(name string, llmClient llm.Client) (Plugin, error) {
	registryMutex.RLock()
	factory, exists := pluginFactories[name]
	registryMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no factory registered for plugin: %s", name)
	}

	return factory(llmClient), nil
}

// GetRegisteredPlugins returns a list of all registered plugin names
func GetRegisteredPlugins() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	plugins := make([]string, 0, len(pluginFactories))
	for name := range pluginFactories {
		plugins = append(plugins, name)
	}
	return plugins
}
