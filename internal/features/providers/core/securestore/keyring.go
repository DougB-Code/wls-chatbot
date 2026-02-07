// keyring.go accesses OS keyring storage for provider secrets.
// internal/features/providers/core/securestore/keyring.go
package securestore

import (
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/providers/interfaces/core"
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

var _ providercore.SecretStore = (*KeyringStore)(nil)

// SaveProviderSecret stores a provider secret field in the OS keychain.
func (s *KeyringStore) SaveProviderSecret(providerName, fieldName, value string) error {
	return keyring.Set(s.serviceName, s.credentialKey(providerName, fieldName), value)
}

// GetProviderSecret retrieves a provider secret field from the OS keychain.
func (s *KeyringStore) GetProviderSecret(providerName, fieldName string) (string, error) {
	return keyring.Get(s.serviceName, s.credentialKey(providerName, fieldName))
}

// HasProviderSecret returns true when a provider secret field is stored.
func (s *KeyringStore) HasProviderSecret(providerName, fieldName string) bool {
	_, err := keyring.Get(s.serviceName, s.credentialKey(providerName, fieldName))
	return err == nil
}

// DeleteProviderSecret removes a stored provider secret field.
func (s *KeyringStore) DeleteProviderSecret(providerName, fieldName string) error {
	return keyring.Delete(s.serviceName, s.credentialKey(providerName, fieldName))
}

// credentialKey builds the keyring entry key for a provider secret field.
func (s *KeyringStore) credentialKey(providerName, fieldName string) string {

	return providerName + ":" + fieldName
}
