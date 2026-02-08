// provider.go defines provider transport contracts for backend adapters.
// internal/core/interfaces/ai/provider.go
package ai

import (
	"context"

	modelinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/interfaces"
)

type ProviderModel = modelinterfaces.ProviderModel
type ProviderModelMutationInterface = modelinterfaces.ProviderModelMutationInterface

// ProviderInterface defines provider query capabilities shared across transports.
type ProviderInterface interface {
	GetProviders(ctx context.Context) ([]ProviderInfo, error)
	TestProvider(ctx context.Context, name string) error
}

// ProviderMutationInterface defines provider CRUD capabilities.
type ProviderMutationInterface interface {
	AddProvider(ctx context.Context, request AddProviderRequest) (ProviderInfo, error)
	UpdateProvider(ctx context.Context, request UpdateProviderRequest) (ProviderInfo, error)
	RemoveProvider(ctx context.Context, name string) error
	UpdateProviderCredentials(ctx context.Context, request UpdateProviderCredentialsRequest) error
}

// ProviderManagementInterface defines full provider and provider-model CRUD capabilities.
type ProviderManagementInterface interface {
	ProviderInterface
	ProviderMutationInterface
	ProviderModelMutationInterface
}

// ProviderInfo contains transport-safe provider status data.
type ProviderInfo struct {
	Name             string                    `json:"name"`
	DisplayName      string                    `json:"displayName"`
	CredentialFields []ProviderCredentialField `json:"credentialFields,omitempty"`
	CredentialValues map[string]string         `json:"credentialValues,omitempty"`
	Models           []ProviderModel           `json:"models"`
	Resources        []ProviderModel           `json:"resources"`
	IsConnected      bool                      `json:"isConnected"`
	IsActive         bool                      `json:"isActive"`
	Status           *ProviderStatus           `json:"status,omitempty"`
}

// AddProviderRequest contains inputs for creating a provider definition.
type AddProviderRequest struct {
	Name          string            `json:"name"`
	DisplayName   string            `json:"displayName"`
	Type          string            `json:"type"`
	BaseURL       string            `json:"baseUrl,omitempty"`
	DefaultModel  string            `json:"defaultModel,omitempty"`
	Credentials   map[string]string `json:"credentials,omitempty"`
	EnabledModels []ProviderModel   `json:"enabledModels,omitempty"`
}

// UpdateProviderRequest contains mutable provider definition fields.
type UpdateProviderRequest struct {
	Name         string  `json:"name"`
	DisplayName  *string `json:"displayName,omitempty"`
	BaseURL      *string `json:"baseUrl,omitempty"`
	DefaultModel *string `json:"defaultModel,omitempty"`
}

// UpdateProviderCredentialsRequest contains credential values to merge for a provider.
type UpdateProviderCredentialsRequest struct {
	ProviderName string            `json:"providerName"`
	Credentials  map[string]string `json:"credentials"`
}

// ProviderCredentialField describes a provider input requirement.
type ProviderCredentialField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"`
	Placeholder string `json:"placeholder,omitempty"`
	Help        string `json:"help,omitempty"`
}

// ProviderStatus describes the last known provider health check.
type ProviderStatus struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message,omitempty"`
	CheckedAt int64  `json:"checkedAt"`
}
