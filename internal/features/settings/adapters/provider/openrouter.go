// openrouter.go implements the provider interface for OpenRouter.
// internal/features/settings/adapters/provider/openrouter.go
package provider

import (
	"context"
	"strings"
)

// OpenRouter implements the Provider interface for OpenRouter.
type OpenRouter struct {
	name        string
	displayName string
	baseURL     string
	apiKey      string
	referer     string
	title       string
	models      []Model
	client      HTTPClient
}

var _ Provider = (*OpenRouter)(nil)

// NewOpenRouter creates a new OpenRouter provider.
func NewOpenRouter(config Config) *OpenRouter {

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	return &OpenRouter{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     baseURL,
		apiKey:      config.APIKey,
		models:      config.Models,
		client:      defaultHTTPClient(),
	}
}

// Name returns the provider identifier.
func (o *OpenRouter) Name() string {

	return o.name
}

// DisplayName returns the human-readable provider name.
func (o *OpenRouter) DisplayName() string {

	return o.displayName
}

// Models returns the available models.
func (o *OpenRouter) Models() []Model {

	return o.models
}

// CredentialFields returns the expected credential inputs.
func (o *OpenRouter) CredentialFields() []CredentialField {

	return []CredentialField{
		{
			Name:     CredentialAPIKey,
			Label:    "API Key",
			Required: true,
			Secret:   true,
		},
		{
			Name:        CredentialOpenRouterReferer,
			Label:       "HTTP Referer (optional)",
			Required:    false,
			Secret:      false,
			Placeholder: "https://your-app.example",
			Help:        "Optional HTTP-Referer header used for OpenRouter analytics.",
		},
		{
			Name:        CredentialOpenRouterTitle,
			Label:       "App Title (optional)",
			Required:    false,
			Secret:      false,
			Placeholder: "Wails Chatbot",
			Help:        "Optional X-Title header used for OpenRouter analytics.",
		},
	}
}

// Configure updates the provider configuration.
func (o *OpenRouter) Configure(config Config) error {

	if config.Credentials != nil {
		if value, ok := config.Credentials[CredentialAPIKey]; ok {
			o.apiKey = strings.TrimSpace(value)
		}
		if value, ok := config.Credentials[CredentialOpenRouterReferer]; ok {
			o.referer = strings.TrimSpace(value)
		}
		if value, ok := config.Credentials[CredentialOpenRouterTitle]; ok {
			o.title = strings.TrimSpace(value)
		}
	}
	if strings.TrimSpace(config.APIKey) != "" {
		o.apiKey = config.APIKey
	}
	if config.BaseURL != "" {
		o.baseURL = config.BaseURL
	}
	if config.Models != nil {
		o.models = config.Models
	}
	return nil
}

// SetHTTPClient overrides the HTTP client used by the provider.
func (o *OpenRouter) SetHTTPClient(client HTTPClient) {

	if client != nil {
		o.client = client
	}
}

// httpClient returns the configured HTTP client or a default client.
func (o *OpenRouter) httpClient() HTTPClient {

	if o.client == nil {
		o.client = defaultHTTPClient()
	}
	return o.client
}

// TestConnection verifies the API is reachable.
func (o *OpenRouter) TestConnection(ctx context.Context) error {

	headers := o.authHeaders()
	_, err := listOpenAICompatModels(ctx, o.httpClient(), o.baseURL, headers)
	return err
}

// ListResources fetches the available models from OpenRouter.
func (o *OpenRouter) ListResources(ctx context.Context) ([]Model, error) {

	headers := o.authHeaders()
	return listOpenAICompatModels(ctx, o.httpClient(), o.baseURL, headers)
}

// Chat implements streaming chat completion.
func (o *OpenRouter) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	headers := o.authHeaders()
	return chatOpenAICompat(ctx, o.httpClient(), o.baseURL, headers, messages, opts)
}

func (o *OpenRouter) authHeaders() map[string]string {

	headers := make(map[string]string)
	if strings.TrimSpace(o.apiKey) != "" {
		headers["Authorization"] = "Bearer " + strings.TrimSpace(o.apiKey)
	}
	if strings.TrimSpace(o.referer) != "" {
		headers["HTTP-Referer"] = strings.TrimSpace(o.referer)
	}
	if strings.TrimSpace(o.title) != "" {
		headers["X-Title"] = strings.TrimSpace(o.title)
	}
	if len(headers) == 0 {
		return nil
	}
	return headers
}
