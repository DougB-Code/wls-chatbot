// Implement the provider interface for Cloudflare AI Gateway.
// internal/features/settings/adapters/provider/cloudflare.go
package provider

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Cloudflare implements the Provider interface for Cloudflare AI Gateway.
type Cloudflare struct {
	name        string
	displayName string
	baseURL     string
	accountID   string
	gatewayID   string
	upstreamAPIKey   string
	cloudflareToken  string
	gatewayToken     string
	logger      Logger
	models      []Model
	client      HTTPClient
}

var _ Provider = (*Cloudflare)(nil)

// NewCloudflare creates a new Cloudflare provider.
func NewCloudflare(config Config) *Cloudflare {

	provider := &Cloudflare{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     config.BaseURL,
		models:      config.Models,
		client:      defaultHTTPClient(),
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
			Name:     CredentialAccountID,
			Label:    "Account ID",
			Required: true,
			Secret:   false,
			Placeholder: "Cloudflare account ID",
			Help:        "The Cloudflare account identifier that owns the AI Gateway.",
		},
		{
			Name:     CredentialGatewayID,
			Label:    "Gateway ID",
			Required: true,
			Secret:   false,
			Placeholder: "AI Gateway ID",
			Help:        "The gateway identifier from your Cloudflare AI Gateway settings.",
		},
		{
			Name:     CredentialCloudflareToken,
			Label:    "Cloudflare API Token",
			Required: false,
			Secret:   true,
			Placeholder: "Cloudflare API token",
			Help:        "Required for Cloudflare-hosted Workers AI models (for example @cf/...).",
		},
		{
			Name:     CredentialAPIKey,
			Label:    "Upstream API Key (optional)",
			Required: false,
			Secret:   true,
			Placeholder: "Provider API key",
			Help:        "Required for non-Workers AI models (for example openai/... or anthropic/...).",
		},
		{
			Name:     CredentialToken,
			Label:    "Gateway Auth Token (optional)",
			Required: false,
			Secret:   true,
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
		if value, ok := config.Credentials[CredentialGatewayID]; ok {
			c.gatewayID = strings.TrimSpace(value)
		}
		if value, ok := config.Credentials[CredentialAPIKey]; ok {
			c.upstreamAPIKey = strings.TrimSpace(value)
		}
		if value, ok := config.Credentials[CredentialCloudflareToken]; ok {
			c.cloudflareToken = strings.TrimSpace(value)
		}
		if value, ok := config.Credentials[CredentialToken]; ok {
			c.gatewayToken = strings.TrimSpace(value)
		}
	}
	if strings.TrimSpace(config.APIKey) != "" {
		c.upstreamAPIKey = strings.TrimSpace(config.APIKey)
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
		c.client = defaultHTTPClient()
	}
	return c.client
}

// TestConnection verifies the API is reachable.
func (c *Cloudflare) TestConnection(ctx context.Context) error {

	_, err := c.ListResources(ctx)
	return err
}

// ListResources fetches the available models from Cloudflare AI Gateway.
func (c *Cloudflare) ListResources(ctx context.Context) ([]Model, error) {

	baseURL := c.resolveBaseURL()
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("base URL required")
	}

	c.logDebug(
		"Cloudflare list resources request",
		LogField{Key: "provider", Value: c.name},
		LogField{Key: "baseUrl", Value: baseURL},
		LogField{Key: "gatewayAuth", Value: formatBool(strings.TrimSpace(c.gatewayToken) != "")},
		LogField{Key: "cloudflareToken", Value: formatBool(strings.TrimSpace(c.cloudflareToken) != "")},
		LogField{Key: "upstreamApiKey", Value: formatBool(strings.TrimSpace(c.upstreamAPIKey) != "")},
	)

	tokens := c.resourceAuthTokens()
	if len(tokens) == 0 {
		headers := c.gatewayHeaders()
		c.logRawRequest("GET", baseURL+"/models", headers, nil)
		return listOpenAICompatModels(ctx, c.httpClient(), baseURL, headers)
	}

	seen := make(map[string]struct{})
	var combined []Model
	var lastErr error
	for _, token := range tokens {
		headers := c.gatewayHeaders()
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Authorization"] = "Bearer " + token
		c.logRawRequest("GET", baseURL+"/models", headers, nil)
		models, err := listOpenAICompatModels(ctx, c.httpClient(), baseURL, headers)
		if err != nil {
			lastErr = err
			continue
		}
		combined = mergeModels(combined, seen, models)
	}

	if len(combined) > 0 {
		sort.Slice(combined, func(i, j int) bool {
			return combined[i].ID < combined[j].ID
		})
		return combined, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return combined, nil
}

// Chat implements streaming chat completion.
func (c *Cloudflare) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	baseURL := c.resolveBaseURL()
	model, workersAI, err := c.resolveModel(opts.Model, baseURL)
	if err != nil {
		return nil, err
	}
	headers, err := c.headersForMode(workersAI)
	if err != nil {
		return nil, err
	}
	headers["Content-Type"] = "application/json"
	c.logDebug(
		"Cloudflare chat request",
		LogField{Key: "provider", Value: c.name},
		LogField{Key: "baseUrl", Value: baseURL},
		LogField{Key: "model", Value: opts.Model},
		LogField{Key: "resolvedModel", Value: model},
		LogField{Key: "mode", Value: resolveModeLabel(workersAI)},
		LogField{Key: "authorization", Value: formatBool(hasHeader(headers, "Authorization"))},
		LogField{Key: "gatewayAuth", Value: formatBool(hasHeader(headers, "cf-aig-authorization"))},
	)
	body, err := marshalOpenAICompatBody(model, messages, opts)
	if err == nil {
		c.logRawRequest("POST", baseURL+"/chat/completions", headers, body)
	}
	opts.Model = model
	return chatOpenAICompat(ctx, c.httpClient(), baseURL, headers, messages, opts)
}

// resolveBaseURL builds the gateway base URL for OpenAI-compatible endpoints.
func (c *Cloudflare) resolveBaseURL() string {

	if strings.TrimSpace(c.baseURL) != "" {
		return c.baseURL
	}
	if strings.TrimSpace(c.accountID) == "" || strings.TrimSpace(c.gatewayID) == "" {
		return ""
	}
	return fmt.Sprintf("https://gateway.ai.cloudflare.com/v1/%s/%s/compat", c.accountID, c.gatewayID)
}

// logDebug writes debug output if a logger is configured.
func (c *Cloudflare) logDebug(message string, fields ...LogField) {

	if c.logger == nil {
		return
	}
	c.logger.Debug(message, fields...)
}

// logRawRequest emits raw request details when enabled.
func (c *Cloudflare) logRawRequest(method, url string, headers map[string]string, body []byte) {

	if c.logger == nil || !shouldLogRawRequests() {
		return
	}

	headerText := formatHeaders(headers)
	bodyText := ""
	if len(body) > 0 {
		bodyText = string(body)
	}

	c.logger.Debug(
		"Cloudflare raw request",
		LogField{Key: "provider", Value: c.name},
		LogField{Key: "method", Value: method},
		LogField{Key: "url", Value: url},
		LogField{Key: "headers", Value: headerText},
		LogField{Key: "body", Value: bodyText},
	)
}

// gatewayHeaders returns headers for Cloudflare AI Gateway authentication.
func (c *Cloudflare) gatewayHeaders() map[string]string {

	if strings.TrimSpace(c.gatewayToken) == "" {
		return nil
	}
	return map[string]string{
		"cf-aig-authorization": "Bearer " + c.gatewayToken,
	}
}

// resolveModel normalizes model names and resolves Workers AI mode.
func (c *Cloudflare) resolveModel(model, baseURL string) (string, bool, error) {

	trimmed := strings.TrimSpace(model)
	if trimmed == "" {
		return "", false, fmt.Errorf("model required")
	}

	if isWorkersAIBaseURL(baseURL) {
		if strings.HasPrefix(trimmed, workersAICompatPrefix) {
			trimmed = strings.TrimPrefix(trimmed, workersAICompatPrefix)
		}
		if !strings.HasPrefix(trimmed, workersAIModelPrefix) {
			return "", false, fmt.Errorf("workers AI base URL requires @cf/ model")
		}
		return trimmed, true, nil
	}

	if strings.HasPrefix(trimmed, workersAIModelPrefix) {
		return workersAICompatPrefix + trimmed, true, nil
	}
	if strings.HasPrefix(trimmed, workersAICompatPrefix) {
		return trimmed, true, nil
	}
	return trimmed, false, nil
}

// headersForMode builds request headers for the selected mode.
func (c *Cloudflare) headersForMode(workersAI bool) (map[string]string, error) {

	headers := c.gatewayHeaders()
	if headers == nil {
		headers = make(map[string]string)
	}

	token := ""
	if workersAI {
		token = strings.TrimSpace(c.cloudflareToken)
		if token == "" {
			return nil, fmt.Errorf("cloudflare API token required for Workers AI models")
		}
	} else {
		token = strings.TrimSpace(c.upstreamAPIKey)
		if token == "" {
			return nil, fmt.Errorf("upstream API key required for non-Workers AI models")
		}
	}
	headers["Authorization"] = "Bearer " + token
	return headers, nil
}

// resourceAuthTokens returns unique auth tokens for listing models.
func (c *Cloudflare) resourceAuthTokens() []string {

	trimmedCloudflare := strings.TrimSpace(c.cloudflareToken)
	trimmedUpstream := strings.TrimSpace(c.upstreamAPIKey)
	if trimmedCloudflare == "" && trimmedUpstream == "" {
		return nil
	}

	tokens := make([]string, 0, 2)
	if trimmedCloudflare != "" {
		tokens = append(tokens, trimmedCloudflare)
	}
	if trimmedUpstream != "" && trimmedUpstream != trimmedCloudflare {
		tokens = append(tokens, trimmedUpstream)
	}
	return tokens
}

// mergeModels merges model lists while preserving unique IDs.
func mergeModels(target []Model, seen map[string]struct{}, incoming []Model) []Model {

	for _, model := range incoming {
		if model.ID == "" {
			continue
		}
		if _, ok := seen[model.ID]; ok {
			continue
		}
		seen[model.ID] = struct{}{}
		target = append(target, model)
	}
	return target
}

const (
	workersAIModelPrefix  = "@cf/"
	workersAICompatPrefix = "workers-ai/"
)

// hasHeader reports whether a header is present with a non-empty value.
func hasHeader(headers map[string]string, name string) bool {

	if headers == nil {
		return false
	}
	value, ok := headers[name]
	return ok && strings.TrimSpace(value) != ""
}

// formatBool formats a boolean as a string.
func formatBool(value bool) string {

	if value {
		return "true"
	}
	return "false"
}

// resolveModeLabel returns a label for the selected mode.
func resolveModeLabel(workersAI bool) string {

	if workersAI {
		return "workers_ai"
	}
	return "upstream"
}

// formatHeaders renders headers as a stable string for logging.
func formatHeaders(headers map[string]string) string {

	if len(headers) == 0 {
		return ""
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		value := headers[name]
		parts = append(parts, fmt.Sprintf("%s: %s", name, value))
	}
	return strings.Join(parts, "; ")
}

// shouldLogRawRequests returns true when raw request logging is enabled.
func shouldLogRawRequests() bool {

	value := strings.TrimSpace(strings.ToLower(os.Getenv("WLS_LOG_RAW_HTTP")))
	return value == "1" || value == "true" || value == "yes"
}

// isWorkersAIBaseURL reports whether the base URL targets Workers AI directly.
func isWorkersAIBaseURL(baseURL string) bool {

	return strings.Contains(strings.ToLower(baseURL), "/workers-ai/")
}

// Ensure Cloudflare implements Provider.
var _ Provider = (*Cloudflare)(nil)
