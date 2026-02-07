// registry.go registers and retrieves provider instances.
// internal/features/providers/core/registry/registry.go
package registry

import (
	"sync"

	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/providers/interfaces/core"
)

// Registry manages available providers.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]providercore.Provider
	active    string
	order     []string
}

// New creates a new provider registry.
func New() *Registry {

	return &Registry{
		providers: make(map[string]providercore.Provider),
	}
}

var _ providercore.ProviderRegistry = (*Registry)(nil)

// Register adds a provider to the registry.
func (r *Registry) Register(p providercore.Provider) {

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.providers[p.Name()]; !ok {
		r.order = append(r.order, p.Name())
	}
	r.providers[p.Name()] = p
}

// Get retrieves a provider by name.
func (r *Registry) Get(name string) providercore.Provider {

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.providers[name]
}

// GetActive returns the currently active provider.
func (r *Registry) GetActive() providercore.Provider {

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
func (r *Registry) List() []providercore.Provider {

	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]providercore.Provider, 0, len(r.providers))
	for _, name := range r.order {
		if p, ok := r.providers[name]; ok {
			providers = append(providers, p)
		}
	}
	return providers
}

// ListConfigs returns configurations for all providers.
func (r *Registry) ListConfigs() []providercore.ProviderConfig {

	r.mu.RLock()
	defer r.mu.RUnlock()

	configs := make([]providercore.ProviderConfig, 0, len(r.providers))
	for _, name := range r.order {
		if p, ok := r.providers[name]; ok {
			configs = append(configs, providercore.ProviderConfig{
				Name:        p.Name(),
				DisplayName: p.DisplayName(),
				Models:      p.Models(),
			})
		}
	}
	return configs
}
