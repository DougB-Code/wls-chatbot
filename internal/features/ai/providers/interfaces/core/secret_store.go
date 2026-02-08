// secret_store.go defines secret storage contracts for provider credentials.
// internal/features/providers/interfaces/core/secret_store.go
package core

// SecretStore manages provider secret storage.
type SecretStore interface {
	SaveProviderSecret(providerName, fieldName, value string) error
	GetProviderSecret(providerName, fieldName string) (string, error)
	HasProviderSecret(providerName, fieldName string) bool
	DeleteProviderSecret(providerName, fieldName string) error
}
