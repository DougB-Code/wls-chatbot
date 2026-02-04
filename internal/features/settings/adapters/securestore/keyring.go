// access OS keyring storage for provider secrets.
// internal/features/settings/adapters/securestore/keyring.go
package securestore

import (
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/ports"
	"github.com/zalando/go-keyring"
)

// KeyringStore stores provider credentials in the OS keychain.
type KeyringStore struct {
	serviceName string
}

// NewKeyringStore creates a keyring-backed secret store scoped to a service name.
func NewKeyringStore(serviceName string) *KeyringStore {
	return &KeyringStore{serviceName: serviceName}
}

var _ ports.SecretStore = (*KeyringStore)(nil)

// SaveProviderKey stores the provider API key in the OS keychain.
func (s *KeyringStore) SaveProviderKey(providerName, apiKey string) error {
	return keyring.Set(s.serviceName, providerName, apiKey)
}

// GetProviderKey retrieves the provider API key from the OS keychain.
func (s *KeyringStore) GetProviderKey(providerName string) (string, error) {
	return keyring.Get(s.serviceName, providerName)
}

// HasProviderKey returns true when a provider API key is stored.
func (s *KeyringStore) HasProviderKey(providerName string) bool {
	_, err := keyring.Get(s.serviceName, providerName)
	return err == nil
}

// DeleteProviderKey removes a stored provider API key.
func (s *KeyringStore) DeleteProviderKey(providerName string) error {
	return keyring.Delete(s.serviceName, providerName)
}
