// provider.go defines provider identity and configuration contracts.
// internal/features/ai/providers/ports/core/provider.go
package core

import (
	"context"

	"github.com/MadeByDoug/wls-chatbot/internal/core/logger"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/gateway"
)

// Model represents an AI model.
type Model struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ContextWindow     int    `json:"contextWindow"`
	SupportsStreaming bool   `json:"supportsStreaming"`
	SupportsTools     bool   `json:"supportsTools"`
	SupportsVision    bool   `json:"supportsVision"`
}

// ProviderCredentials stores credential values by field name.
type ProviderCredentials map[string]string

// CredentialField describes an input required to configure a provider.
type CredentialField struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Required    bool   `json:"required"`
	Secret      bool   `json:"secret"`
	Placeholder string `json:"placeholder,omitempty"`
	Help        string `json:"help,omitempty"`
}

const (
	CredentialAPIKey            = "api_key"
	CredentialAccountID         = "account_id"
	CredentialGatewayID         = "gateway_id"
	CredentialToken             = "token"
	CredentialCloudflareToken   = "cloudflare_api_token"
	CredentialOpenRouterReferer = "openrouter_referer"
	CredentialOpenRouterTitle   = "openrouter_title"
)

// ProviderConfig holds provider configuration.
type ProviderConfig struct {
	Name         string              `json:"name"`
	DisplayName  string              `json:"displayName"`
	APIKey       string              `json:"apiKey,omitempty"`
	BaseURL      string              `json:"baseUrl,omitempty"`
	DefaultModel string              `json:"defaultModel"`
	Models       []Model             `json:"models"`
	Credentials  ProviderCredentials `json:"credentials,omitempty"`
	Logger       logger.Logger       `json:"-"`
}

// Provider describes the full provider contract used by app services.
type Provider interface {
	providergateway.Provider
	Name() string
	DisplayName() string
	Models() []Model
	ListResources(ctx context.Context) ([]Model, error)
	CredentialFields() []CredentialField
	Configure(config ProviderConfig) error
}
