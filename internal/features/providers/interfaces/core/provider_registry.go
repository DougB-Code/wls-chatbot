// provider_registry.go defines provider registry contracts.
// internal/features/providers/interfaces/core/provider_registry.go
package core

// ProviderRegistry manages provider instances and active selection.
type ProviderRegistry interface {
	Register(p Provider)
	Get(name string) Provider
	GetActive() Provider
	SetActive(name string) bool
	List() []Provider
	ListConfigs() []ProviderConfig
}
