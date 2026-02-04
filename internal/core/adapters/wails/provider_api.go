// expose provider endpoints to the frontend via the bridge.
// internal/core/adapters/wails/provider_api.go
package wails

import "github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"

// GetProviders returns all available providers with their status.
func (b *Bridge) GetProviders() []provider.Info {

	return b.providers.GetProviders()
}

// ConnectProvider connects and configures a provider with the given API key.
func (b *Bridge) ConnectProvider(name, apiKey string) (provider.Info, error) {

	return b.providers.ConnectProvider(b.ctxOrBackground(), name, apiKey)
}

// ConfigureProvider updates a provider's API key without full connection flow.
func (b *Bridge) ConfigureProvider(name, apiKey string) error {

	return b.providers.ConfigureProvider(name, apiKey)
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
