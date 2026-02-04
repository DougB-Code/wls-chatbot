// define provider registry contracts.
// internal/features/settings/ports/provider_registry.go
package ports

// ProviderRegistry manages provider instances and active selection.
type ProviderRegistry interface {
	Register(p Provider)
	Get(name string) Provider
	GetActive() Provider
	SetActive(name string) bool
	List() []Provider
	ListConfigs() []ProviderConfig
}
