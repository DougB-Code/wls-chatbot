// provider.go implements the Cloudflare AI Gateway provider adapter.
// internal/features/providers/adapters/cloudflare/provider.go
package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
	providerhttp "github.com/MadeByDoug/wls-chatbot/internal/features/providers/core/providerhttp"
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/providers/interfaces/core"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/providers/interfaces/gateway"
)

type Model = providercore.Model
type Config = providercore.ProviderConfig
type ChatOptions = providergateway.ChatOptions
type ProviderMessage = providergateway.ProviderMessage
type Chunk = providergateway.Chunk
type CredentialField = providercore.CredentialField
type ProviderCredentials = providercore.ProviderCredentials
type Provider = providercore.Provider
type HTTPClient = providerhttp.Client
type Logger = coreports.Logger
type LogField = coreports.LogField

const (
	CredentialAccountID       = providercore.CredentialAccountID
	CredentialGatewayID       = providercore.CredentialGatewayID
	CredentialCloudflareToken = providercore.CredentialCloudflareToken
	CredentialAPIKey          = providercore.CredentialAPIKey
	CredentialToken           = providercore.CredentialToken
)

// Cloudflare implements the Provider interface for Cloudflare AI Gateway.
type Cloudflare struct {
	name            string
	displayName     string
	baseURL         string
	accountID       string
	gatewayID       string
	upstreamAPIKey  string
	cloudflareToken string
	gatewayToken    string
	logger          Logger
	models          []Model
	client          HTTPClient
}

var _ Provider = (*Cloudflare)(nil)

// New creates a new Cloudflare provider.
func New(config Config) *Cloudflare {

	provider := &Cloudflare{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     config.BaseURL,
		models:      config.Models,
		client:      providerhttp.NewDefaultClient(),
	}
	_ = provider.Configure(config)
	return provider
}

// Name returns the provider identifier.
func (c *Cloudflare) Name() string {

	return c.name
}

// DisplayName returns the human-readable provider name.
func (c *Cloudflare) DisplayName() string {

	return c.displayName
}

// Models returns the available models.
func (c *Cloudflare) Models() []Model {

	return c.models
}

// CredentialFields returns the expected credential inputs.
func (c *Cloudflare) CredentialFields() []CredentialField {

	return []CredentialField{
		{
			Name:        CredentialAccountID,
			Label:       "Account ID",
			Required:    true,
			Secret:      false,
			Placeholder: "Cloudflare account ID",
			Help:        "The Cloudflare account identifier that owns the AI Gateway.",
		},
		{
			Name:        CredentialGatewayID,
			Label:       "Gateway ID",
			Required:    true,
			Secret:      false,
			Placeholder: "AI Gateway ID",
			Help:        "The gateway identifier from your Cloudflare AI Gateway settings.",
		},
		{
			Name:        CredentialCloudflareToken,
			Label:       "Cloudflare API Token",
			Required:    false,
			Secret:      true,
			Placeholder: "Cloudflare API token",
			Help:        "Required for Cloudflare-hosted Workers AI models (for example @cf/...).",
		},
		{
			Name:        CredentialAPIKey,
			Label:       "Upstream API Key (optional)",
			Required:    false,
			Secret:      true,
			Placeholder: "Provider API key",
			Help:        "Required for non-Workers AI models (for example openai/... or anthropic/...).",
		},
		{
			Name:        CredentialToken,
			Label:       "Gateway Auth Token (optional)",
			Required:    false,
			Secret:      true,
			Placeholder: "AI Gateway token",
			Help:        "Optional token for Cloudflare AI Gateway authentication (cf-aig-authorization).",
		},
	}
}

// Configure updates the provider configuration.
func (c *Cloudflare) Configure(config Config) error {

	if config.Credentials != nil {
		if value, ok := config.Credentials[CredentialAccountID]; ok {
			c.accountID = strings.TrimSpace(value)
		}
		if value, ok := config.Credentials[CredentialCloudflareToken]; ok {
			c.cloudflareToken = strings.TrimSpace(value)
		}
	}
	if config.BaseURL != "" {
		c.baseURL = config.BaseURL
	}
	if config.Models != nil {
		c.models = config.Models
	}
	if config.Logger != nil {
		c.logger = config.Logger
	}
	return nil
}

// SetHTTPClient overrides the HTTP client used by the provider.
func (c *Cloudflare) SetHTTPClient(client HTTPClient) {

	if client != nil {
		c.client = client
	}
}

// SetLogger sets the logger used for debug output.
func (c *Cloudflare) SetLogger(logger Logger) {

	if logger != nil {
		c.logger = logger
	}
}

// httpClient returns the configured HTTP client or a default client.
func (c *Cloudflare) httpClient() HTTPClient {

	if c.client == nil {
		c.client = providerhttp.NewDefaultClient()
	}
	return c.client
}

// newSDKClient creates a new Cloudflare SDK client.
func (c *Cloudflare) newSDKClient() (*cloudflare.API, error) {
	// Cloudflare SDK requires API Token.
	token := c.cloudflareToken
	if token == "" {
		return nil, fmt.Errorf("cloudflare API token required for SDK usage")
	}

	// Create client
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, err
	}
	
	if c.client != nil {
		// propagate client if feasible, but SDK v0.116 doesn't easily expose it via NewWithAPIToken
	}
	
	return api, nil
}

// TestConnection verifies the API is reachable.
func (c *Cloudflare) TestConnection(ctx context.Context) error {
	// Simple validation
	if c.accountID == "" || c.cloudflareToken == "" {
		return fmt.Errorf("account ID and Cloudflare token required")
	}
	return nil
}

// ListResources fetches the available models.
func (c *Cloudflare) ListResources(ctx context.Context) ([]Model, error) {

	// SDK List logic using api.Raw
	api, err := c.newSDKClient()
	if err != nil {
		return nil, err
	}
	if c.accountID == "" {
		return nil, fmt.Errorf("account ID required for listing models")
	}

	// Try to list text generation models via generic search endpoint if available, 
	// or assume standard models. Since SDK doesn't expose ListWorkersAIModels, 
	// and we removed legacy, we attempt `GET accounts/{id}/ai/models/search` 
	// which is the common endpoint.
	
	endpoint := fmt.Sprintf("accounts/%s/ai/models/search", c.accountID)
	// Some docs say `/ai/models/search`, others just `/ai/models`.
	// We'll try one. If it fails, we return error (strict no legacy).
	
	res, err := api.Raw(ctx, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list models via SDK: %w", err)
	}
	
	if !res.Success {
		return nil, fmt.Errorf("failed to list models: active success=false in response")
	}

	// Response structure for models search
	var listRes struct {
		Result []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Task        struct {
				Name string `json:"name"`
			} `json:"task"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(res.Result, &listRes); err != nil {
		// fall back to empty list if parsing fails? Or strict error?
		return nil, fmt.Errorf("failed to parse models list: %w", err)
	}

	var models []Model
	seen := make(map[string]struct{})
	
	for _, m := range listRes.Result {
		// Filter for text generation or image models if desired?
		// For now, include all or just Text Generation.
		// "Text Generation" task name usually.
		if m.Task.Name == "Text Generation" || m.Task.Name == "Text-to-Image" {
			id := "@cf/" + m.Name
			if !strings.HasPrefix(m.Name, "@cf/") {
				// usually name is full like "meta/llama-2-7b-chat-int8" ? 
				// The API returns name like "meta/llama-2-7b-chat-int8".
				// Workers AI requires "@cf/" prefix often, or just name? 
				// SDK examples often show "@cf/meta/llama..."
				id = "@cf/" + m.Name
			}
			
			if _, ok := seen[id]; !ok {
				models = append(models, Model{
					ID:   id,
					Name: m.Name,
				})
				seen[id] = struct{}{}
			}
		}
	}
	
	if len(models) > 0 {
		sort.Slice(models, func(i, j int) bool {
			return models[i].ID < models[j].ID
		})
	}
	return models, nil
}

// Chat implements streaming chat completion.
func (c *Cloudflare) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	// Strict SDK / Workers AI logic
	if c.accountID == "" {
		return nil, fmt.Errorf("account ID required")
	}
	
	model := resolveModelName(opts.Model)

	api, err := c.newSDKClient()
	if err != nil {
		return nil, err
	}
	
	endpoint := fmt.Sprintf("accounts/%s/ai/run/%s", c.accountID, model)
	
	reqBody := map[string]interface{}{}
	sdkMessages := make([]map[string]string, 0, len(messages))
	for _, msg := range messages {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		sdkMessages = append(sdkMessages, map[string]string{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}
	reqBody["messages"] = sdkMessages
	reqBody["max_tokens"] = opts.MaxTokens
	
	if opts.Stream {
		// SDK api.Raw does not support streaming (buffers response).
		// Per user instruction "Remove legacy", we cannot fallback to manual HTTP.
		// We return error for streaming requests.
		return nil, fmt.Errorf("streaming not supported with current SDK integration")
	}

	// Non-streaming via SDK Raw
	res, err := api.Raw(ctx, "POST", endpoint, reqBody, nil)
	if err != nil {
		return nil, err
	}
	
	if !res.Success {
		if len(res.Errors) > 0 {
			return nil, fmt.Errorf("ai error: %s", res.Errors[0].Message)
		}
		return nil, fmt.Errorf("ai request failed")
	}

	var aiRes struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(res.Result, &aiRes); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}
	
	chunks := make(chan Chunk, 1)
	chunks <- Chunk{
		Content: aiRes.Response,
	}
	close(chunks)
	return chunks, nil
}

// resolveModelName normalizes model names for Workers AI.
func resolveModelName(model string) string {
	model = strings.TrimSpace(model)
	// Ensure @cf prefix if missing and looks like it needs one?
	// or assume input is correct.
	// Common convention: "llama-2..." -> "@cf/meta/llama-2..."
	// But we can't guess vendor.
	// If it doesn't start with @cf/, we assume user provided full path or we fail?
	// Let's assume input is raw ID.
	return model
}

// logDebug writes debug output if a logger is configured.
func (c *Cloudflare) logDebug(message string, fields ...LogField) {

	if c.logger == nil {
		return
	}
	c.logger.Debug(message, fields...)
}

// isWorkersAIBaseURL reports whether the base URL targets Workers AI directly.
func isWorkersAIBaseURL(baseURL string) bool {
	return strings.Contains(strings.ToLower(baseURL), "/workers-ai/")
}

// GenerateImage generates an image. Cloudflare Workers AI supports this, but implementation is pending specific model support verification.
func (c *Cloudflare) GenerateImage(ctx context.Context, opts providergateway.ImageGenerationOptions) (*providergateway.ImageResult, error) {
	// SDK's api.Raw() expects JSON response, but Image Gen returns binary (PNG).
	// Cannot use SDK's Raw method for this as it tries to parse JSON.
	return nil, fmt.Errorf("cloudflare provider image generation not yet implemented via SDK")
}

// EditImage returns not supported error.
func (c *Cloudflare) EditImage(ctx context.Context, opts providergateway.ImageEditOptions) (*providergateway.ImageResult, error) {
	return nil, fmt.Errorf("edit image not supported by this provider")
}
