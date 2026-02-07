// expose provider endpoints to the frontend via the bridge.
// internal/core/adapters/wails/provider_api.go
package wails

import (
	"fmt"

	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"
)

// GetProviders returns all available providers with their status.
func (b *Bridge) GetProviders() []provider.Info {

	if b.backend != nil {
		infos, err := b.backend.GetProviders(b.ctxOrBackground())
		if err == nil {
			return infos
		}
	}
	if b.providers == nil {
		return nil
	}
	return b.providers.GetProviders()
}

// ConnectProvider connects and configures a provider with the given credentials.
func (b *Bridge) ConnectProvider(name string, credentials provider.ProviderCredentials) (provider.Info, error) {

	return b.providers.ConnectProvider(b.ctxOrBackground(), name, credentials)
}

// ConfigureProvider updates a provider's credentials without full connection flow.
func (b *Bridge) ConfigureProvider(name string, credentials provider.ProviderCredentials) error {

	return b.providers.ConfigureProvider(name, credentials)
}

// DisconnectProvider removes a provider's credentials and resets its state.
func (b *Bridge) DisconnectProvider(name string) error {

	return b.providers.DisconnectProvider(name)
}

// SetActiveProvider sets the active provider by name.
func (b *Bridge) SetActiveProvider(name string) bool {

	return b.providers.SetActiveProvider(name)
}

// TestProvider tests the connection to a provider.
func (b *Bridge) TestProvider(name string) error {

	if b.backend != nil {
		return b.backend.TestProvider(b.ctxOrBackground(), name)
	}
	if b.providers == nil {
		return fmt.Errorf("provider orchestrator not configured")
	}
	return b.providers.TestProvider(b.ctxOrBackground(), name)
}

// RefreshProviderResources fetches the latest resources from a provider.
func (b *Bridge) RefreshProviderResources(name string) error {

	return b.providers.RefreshProviderResources(b.ctxOrBackground(), name)
}

// GetActiveProvider returns the currently active provider, if any.
func (b *Bridge) GetActiveProvider() *provider.Info {

	return b.providers.GetActiveProvider()
}
