// provider.go defines canonical provider DTOs for the application facade.
// internal/app/contracts/provider.go
package contracts

// ProviderInfo contains application-layer provider status data.
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

// AddProviderRequest contains inputs for configuring a provider.
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

// UpdateProviderCredentialsRequest contains credentials to merge for a provider.
type UpdateProviderCredentialsRequest struct {
	ProviderName string            `json:"providerName"`
	Credentials  map[string]string `json:"credentials"`
}

// ProviderCredentialField describes one provider input requirement.
type ProviderCredentialField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"`
	Placeholder string `json:"placeholder,omitempty"`
	Help        string `json:"help,omitempty"`
}

// ProviderModel describes a model exposed by a provider.
type ProviderModel struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ContextWindow     int    `json:"contextWindow"`
	SupportsStreaming bool   `json:"supportsStreaming"`
	SupportsTools     bool   `json:"supportsTools"`
	SupportsVision    bool   `json:"supportsVision"`
}

// ProviderStatus describes the last known provider health-check result.
type ProviderStatus struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message,omitempty"`
	CheckedAt int64  `json:"checkedAt"`
}
