// provider_api.go exposes provider endpoints to the frontend via the bridge.
// internal/ui/adapters/wails/provider_api.go
package wails

import (
	"fmt"

	providerfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/app/provider"
)

// GetProviders returns all available providers with their status.
func (b *Bridge) GetProviders() []providerfeature.Info {

	if b.app == nil || b.app.Providers == nil {
		return nil
	}

	infos, err := b.app.Providers.GetProviders(b.ctxOrBackground())
	if err != nil {
		return nil
	}
	return infos
}

// ConnectProvider connects and configures a provider with the given credentials.
func (b *Bridge) ConnectProvider(name string, credentials providerfeature.ProviderCredentials) (providerfeature.Info, error) {

	if b.app == nil || b.app.Providers == nil {
		return providerfeature.Info{}, fmt.Errorf("provider interface not configured")
	}

	return b.app.Providers.AddProvider(b.ctxOrBackground(), name, credentials)
}

// ConfigureProvider updates a provider's credentials without full connection flow.
func (b *Bridge) ConfigureProvider(name string, credentials providerfeature.ProviderCredentials) error {

	if b.app == nil || b.app.Providers == nil {
		return fmt.Errorf("provider interface not configured")
	}
	return b.app.Providers.UpdateProviderCredentials(b.ctxOrBackground(), name, credentials)
}

// DisconnectProvider removes a provider's credentials and resets its state.
func (b *Bridge) DisconnectProvider(name string) error {

	if b.app == nil || b.app.Providers == nil {
		return fmt.Errorf("provider interface not configured")
	}
	return b.app.Providers.RemoveProvider(b.ctxOrBackground(), name)
}

// SetActiveProvider sets the active provider by name.
func (b *Bridge) SetActiveProvider(name string) bool {

	if b.app == nil || b.app.Providers == nil {
		return false
	}
	return b.app.Providers.SetActiveProvider(name)
}

// TestProvider tests the connection to a provider.
func (b *Bridge) TestProvider(name string) error {

	if b.app == nil || b.app.Providers == nil {
		return fmt.Errorf("provider interface not configured")
	}
	return b.app.Providers.TestProvider(b.ctxOrBackground(), name)
}

// RefreshProviderResources fetches the latest resources from a provider.
func (b *Bridge) RefreshProviderResources(name string) error {

	if b.app == nil || b.app.Providers == nil {
		return fmt.Errorf("provider interface not configured")
	}
	return b.app.Providers.RefreshProviderResources(b.ctxOrBackground(), name)
}

// GetActiveProvider returns the currently active provider, if any.
func (b *Bridge) GetActiveProvider() *providerfeature.Info {

	if b.app == nil || b.app.Providers == nil {
		return nil
	}
	return b.app.Providers.GetActiveProvider()
}
