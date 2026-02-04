// define secret storage contracts for provider credentials.
// internal/features/settings/ports/secret_store.go
package ports

// SecretStore manages provider API key storage.
type SecretStore interface {
	SaveProviderKey(providerName, apiKey string) error
	GetProviderKey(providerName string) (string, error)
	HasProviderKey(providerName string) bool
	DeleteProviderKey(providerName string) error
}
