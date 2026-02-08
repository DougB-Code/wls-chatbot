// provider_facade.go adapts provider orchestrator capabilities into app contracts.
// internal/app/provider_facade.go
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/MadeByDoug/wls-chatbot/internal/app/contracts"
	providerfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/provider"
)

// NewProviderManagement adapts provider orchestrator operations into app contracts.
func NewProviderManagement(orchestrator *providerfeature.Orchestrator) ProviderManagement {

	return &providerManagement{orchestrator: orchestrator}
}

// providerManagement exposes provider orchestrator operations through app contracts.
type providerManagement struct {
	orchestrator *providerfeature.Orchestrator
}

// GetProviders returns provider statuses.
func (m *providerManagement) GetProviders(context.Context) ([]contracts.ProviderInfo, error) {

	if m.orchestrator == nil {
		return nil, fmt.Errorf("app providers: orchestrator not configured")
	}
	return mapProviderInfos(m.orchestrator.GetProviders()), nil
}

// TestProvider checks provider connectivity.
func (m *providerManagement) TestProvider(ctx context.Context, name string) error {

	if m.orchestrator == nil {
		return fmt.Errorf("app providers: orchestrator not configured")
	}
	return m.orchestrator.TestProvider(ctx, name)
}

// AddProvider configures credentials for a known provider.
func (m *providerManagement) AddProvider(ctx context.Context, request contracts.AddProviderRequest) (contracts.ProviderInfo, error) {

	if m.orchestrator == nil {
		return contracts.ProviderInfo{}, fmt.Errorf("app providers: orchestrator not configured")
	}
	if strings.TrimSpace(request.Name) == "" {
		return contracts.ProviderInfo{}, fmt.Errorf("app providers: provider name required")
	}
	if strings.TrimSpace(request.Type) != "" || strings.TrimSpace(request.BaseURL) != "" ||
		strings.TrimSpace(request.DefaultModel) != "" || len(request.EnabledModels) > 0 {
		return contracts.ProviderInfo{}, fmt.Errorf("app providers: dynamic provider registration fields are not supported")
	}

	info, err := m.orchestrator.ConnectProvider(ctx, request.Name, providerfeature.ProviderCredentials(request.Credentials))
	if err != nil {
		return contracts.ProviderInfo{}, err
	}
	return mapProviderInfo(info), nil
}

// UpdateProvider updates provider metadata when supported.
func (*providerManagement) UpdateProvider(context.Context, contracts.UpdateProviderRequest) (contracts.ProviderInfo, error) {

	return contracts.ProviderInfo{}, fmt.Errorf("app providers: provider metadata updates are not configured")
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
func (m *providerManagement) UpdateProviderCredentials(_ context.Context, request contracts.UpdateProviderCredentialsRequest) error {

	if m.orchestrator == nil {
		return fmt.Errorf("app providers: orchestrator not configured")
	}
	if strings.TrimSpace(request.ProviderName) == "" {
		return fmt.Errorf("app providers: provider name required")
	}
	return m.orchestrator.ConfigureProvider(request.ProviderName, providerfeature.ProviderCredentials(request.Credentials))
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
func (m *providerManagement) GetActiveProvider() *contracts.ProviderInfo {

	if m.orchestrator == nil {
		return nil
	}
	info := m.orchestrator.GetActiveProvider()
	if info == nil {
		return nil
	}
	mapped := mapProviderInfo(*info)
	return &mapped
}

// mapProviderInfo converts provider feature DTOs into app contracts.
func mapProviderInfo(info providerfeature.Info) contracts.ProviderInfo {

	var status *contracts.ProviderStatus
	if info.Status != nil {
		status = &contracts.ProviderStatus{
			OK:        info.Status.OK,
			Message:   info.Status.Message,
			CheckedAt: info.Status.CheckedAt,
		}
	}

	return contracts.ProviderInfo{
		Name:             info.Name,
		DisplayName:      info.DisplayName,
		CredentialFields: mapProviderCredentialFields(info.CredentialFields),
		CredentialValues: copyProviderValues(info.CredentialValues),
		Models:           mapProviderModels(info.Models),
		Resources:        mapProviderModels(info.Resources),
		IsConnected:      info.IsConnected,
		IsActive:         info.IsActive,
		Status:           status,
	}
}

// mapProviderInfos converts provider feature DTOs into app contracts.
func mapProviderInfos(infos []providerfeature.Info) []contracts.ProviderInfo {

	if len(infos) == 0 {
		return nil
	}

	mapped := make([]contracts.ProviderInfo, 0, len(infos))
	for _, info := range infos {
		mapped = append(mapped, mapProviderInfo(info))
	}
	return mapped
}

// mapProviderCredentialFields converts provider credential metadata into app contracts.
func mapProviderCredentialFields(fields []providerfeature.CredentialField) []contracts.ProviderCredentialField {

	if len(fields) == 0 {
		return nil
	}

	mapped := make([]contracts.ProviderCredentialField, 0, len(fields))
	for _, field := range fields {
		mapped = append(mapped, contracts.ProviderCredentialField{
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

// mapProviderModels converts provider model metadata into app contracts.
func mapProviderModels(models []providerfeature.Model) []contracts.ProviderModel {

	if len(models) == 0 {
		return nil
	}

	mapped := make([]contracts.ProviderModel, 0, len(models))
	for _, model := range models {
		mapped = append(mapped, contracts.ProviderModel{
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

	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
