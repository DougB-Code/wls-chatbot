// provider_facade.go adapts provider orchestrator capabilities into app interfaces.
// internal/app/provider_facade.go
package app

import (
	"context"
	"fmt"
	"strings"

	providerfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/provider"
)

// NewProviderManagement adapts provider orchestrator operations into app interfaces.
func NewProviderManagement(orchestrator *providerfeature.Orchestrator) ProviderManagement {

	return &providerManagement{orchestrator: orchestrator}
}

// providerManagement exposes provider orchestrator operations through app interfaces.
type providerManagement struct {
	orchestrator *providerfeature.Orchestrator
}

// GetProviders returns provider statuses.
func (m *providerManagement) GetProviders(context.Context) ([]providerfeature.Info, error) {

	if m.orchestrator == nil {
		return nil, fmt.Errorf("app providers: orchestrator not configured")
	}
	return m.orchestrator.GetProviders(), nil
}

// TestProvider checks provider connectivity.
func (m *providerManagement) TestProvider(ctx context.Context, name string) error {

	if m.orchestrator == nil {
		return fmt.Errorf("app providers: orchestrator not configured")
	}
	return m.orchestrator.TestProvider(ctx, name)
}

// AddProvider configures credentials for a known provider.
func (m *providerManagement) AddProvider(ctx context.Context, name string, credentials providerfeature.ProviderCredentials) (providerfeature.Info, error) {

	if m.orchestrator == nil {
		return providerfeature.Info{}, fmt.Errorf("app providers: orchestrator not configured")
	}
	if strings.TrimSpace(name) == "" {
		return providerfeature.Info{}, fmt.Errorf("app providers: provider name required")
	}

	return m.orchestrator.ConnectProvider(ctx, name, credentials)
}

// RemoveProvider disconnects a provider.
func (m *providerManagement) RemoveProvider(_ context.Context, name string) error {

	if m.orchestrator == nil {
		return fmt.Errorf("app providers: orchestrator not configured")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("app providers: provider name required")
	}
	return m.orchestrator.DisconnectProvider(name)
}

// UpdateProviderCredentials updates provider credentials.
func (m *providerManagement) UpdateProviderCredentials(_ context.Context, name string, credentials providerfeature.ProviderCredentials) error {

	if m.orchestrator == nil {
		return fmt.Errorf("app providers: orchestrator not configured")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("app providers: provider name required")
	}
	return m.orchestrator.ConfigureProvider(name, credentials)
}

// SetActiveProvider sets the active provider.
func (m *providerManagement) SetActiveProvider(name string) bool {

	if m.orchestrator == nil {
		return false
	}
	return m.orchestrator.SetActiveProvider(name)
}

// RefreshProviderResources refreshes provider resources.
func (m *providerManagement) RefreshProviderResources(ctx context.Context, name string) error {

	if m.orchestrator == nil {
		return fmt.Errorf("app providers: orchestrator not configured")
	}
	return m.orchestrator.RefreshProviderResources(ctx, name)
}

// GetActiveProvider returns the active provider, when present.
func (m *providerManagement) GetActiveProvider() *providerfeature.Info {

	if m.orchestrator == nil {
		return nil
	}
	return m.orchestrator.GetActiveProvider()
}
