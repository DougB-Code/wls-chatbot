// define provider resource cache contracts.
// internal/core/ports/provider_cache.go
package ports

// ProviderCacheEntry represents cached resources for a provider.
type ProviderCacheEntry struct {
	UpdatedAt int64   `json:"updatedAt"`
	Models    []Model `json:"models"`
}

// ProviderCacheSnapshot contains cached resources by provider name.
type ProviderCacheSnapshot map[string]ProviderCacheEntry

// ProviderCache manages persistence of provider resources.
type ProviderCache interface {
	Load() (ProviderCacheSnapshot, error)
	Save(snapshot ProviderCacheSnapshot) error
}
