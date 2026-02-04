// register and retrieve provider instances.
// internal/adapters/provider/registry.go
package provider

import (
	"sync"

	"github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

// Registry manages available providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	active    string
	order     []string
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {

	return &Registry{
		providers: make(map[string]Provider),
	}
}

var _ ports.ProviderRegistry = (*Registry)(nil)

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[p.Name()]; !ok {
		r.order = append(r.order, p.Name())
	}
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) Provider {

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// GetActive returns the currently active provider.
func (r *Registry) GetActive() Provider {

	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.active == "" {
		return nil
	}
	return r.providers[r.active]
}

// SetActive sets the active provider, or clears it when name is empty.
func (r *Registry) SetActive(name string) bool {

	r.mu.Lock()
	defer r.mu.Unlock()
	if name == "" {
		r.active = ""
		return true
	}
	if _, ok := r.providers[name]; ok {
		r.active = name
		return true
	}
	return false
}

// List returns all registered providers.
func (r *Registry) List() []Provider {

	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]Provider, 0, len(r.providers))
	for _, name := range r.order {
		if p, ok := r.providers[name]; ok {
			providers = append(providers, p)
		}
	}
	return providers
}

// ListConfigs returns configurations for all providers.
func (r *Registry) ListConfigs() []Config {

	r.mu.RLock()
	defer r.mu.RUnlock()

	configs := make([]Config, 0, len(r.providers))
	for _, name := range r.order {
		if p, ok := r.providers[name]; ok {
			configs = append(configs, Config{
				Name:        p.Name(),
				DisplayName: p.DisplayName(),
				Models:      p.Models(),
			})
		}
	}
	return configs
}
