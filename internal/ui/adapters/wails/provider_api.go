// provider_api.go exposes provider endpoints to the frontend via the bridge.
// internal/ui/adapters/wails/provider_api.go
package wails

import (
	"fmt"

	appcontracts "github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	provider "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/provider"
)

// GetProviders returns all available providers with their status.
func (b *Bridge) GetProviders() []provider.Info {

	if b.app == nil || b.app.Providers == nil {
		return nil
	}

	infos, err := b.app.Providers.GetProviders(b.ctxOrBackground())
	if err != nil {
		return nil
	}
	return mapAppProviderInfos(infos)
}

// ConnectProvider connects and configures a provider with the given credentials.
func (b *Bridge) ConnectProvider(name string, credentials provider.ProviderCredentials) (provider.Info, error) {

	if b.app == nil || b.app.Providers == nil {
		return provider.Info{}, fmt.Errorf("provider interface not configured")
	}

	info, err := b.app.Providers.AddProvider(b.ctxOrBackground(), appcontracts.AddProviderRequest{
		Name:        name,
		Credentials: map[string]string(credentials),
	})
	if err != nil {
		return provider.Info{}, err
	}
	return mapAppProviderInfo(info), nil
}

// ConfigureProvider updates a provider's credentials without full connection flow.
func (b *Bridge) ConfigureProvider(name string, credentials provider.ProviderCredentials) error {

	if b.app == nil || b.app.Providers == nil {
		return fmt.Errorf("provider interface not configured")
	}
	return b.app.Providers.UpdateProviderCredentials(b.ctxOrBackground(), appcontracts.UpdateProviderCredentialsRequest{
		ProviderName: name,
		Credentials:  map[string]string(credentials),
	})
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
func (b *Bridge) GetActiveProvider() *provider.Info {

	if b.app == nil || b.app.Providers == nil {
		return nil
	}

	info := b.app.Providers.GetActiveProvider()
	if info == nil {
		return nil
	}
	mapped := mapAppProviderInfo(*info)
	return &mapped
}

// mapAppProviderInfo converts app provider DTOs into Wails provider DTOs.
func mapAppProviderInfo(info appcontracts.ProviderInfo) provider.Info {

	var status *provider.Status
	if info.Status != nil {
		status = &provider.Status{
			OK:        info.Status.OK,
			Message:   info.Status.Message,
			CheckedAt: info.Status.CheckedAt,
		}
	}

	return provider.Info{
		Name:             info.Name,
		DisplayName:      info.DisplayName,
		CredentialFields: mapAppCredentialFields(info.CredentialFields),
		CredentialValues: copyProviderValues(info.CredentialValues),
		Models:           mapAppProviderModels(info.Models),
		Resources:        mapAppProviderModels(info.Resources),
		IsConnected:      info.IsConnected,
		IsActive:         info.IsActive,
		Status:           status,
	}
}

// mapAppProviderInfos converts app provider DTOs into Wails provider DTOs.
func mapAppProviderInfos(infos []appcontracts.ProviderInfo) []provider.Info {

	if len(infos) == 0 {
		return nil
	}

	mapped := make([]provider.Info, 0, len(infos))
	for _, info := range infos {
		mapped = append(mapped, mapAppProviderInfo(info))
	}
	return mapped
}

// mapAppCredentialFields converts app credential field DTOs into Wails provider DTOs.
func mapAppCredentialFields(fields []appcontracts.ProviderCredentialField) []provider.CredentialField {

	if len(fields) == 0 {
		return nil
	}

	mapped := make([]provider.CredentialField, 0, len(fields))
	for _, field := range fields {
		mapped = append(mapped, provider.CredentialField{
			Name:        field.Name,
			Label:       field.Label,
			Required:    field.Required,
			Secret:      field.Secret,
			Placeholder: field.Placeholder,
			Help:        field.Help,
		})
	}
	return mapped
}

// mapAppProviderModels converts app provider model DTOs into Wails provider DTOs.
func mapAppProviderModels(models []appcontracts.ProviderModel) []provider.Model {

	if len(models) == 0 {
		return nil
	}

	mapped := make([]provider.Model, 0, len(models))
	for _, model := range models {
		mapped = append(mapped, provider.Model{
			ID:                model.ID,
			Name:              model.Name,
			ContextWindow:     model.ContextWindow,
			SupportsStreaming: model.SupportsStreaming,
			SupportsTools:     model.SupportsTools,
			SupportsVision:    model.SupportsVision,
		})
	}
	return mapped
}

// copyProviderValues duplicates provider credential values.
func copyProviderValues(values map[string]string) map[string]string {

	if len(values) == 0 {
		return nil
	}

	duplicated := make(map[string]string, len(values))
	for key, value := range values {
		duplicated[key] = value
	}
	return duplicated
}
