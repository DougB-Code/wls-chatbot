// Implement the provider interface for OpenAI.
// internal/features/settings/adapters/provider/openai.go
package provider

import (
	"context"
	"errors"
	"sort"
	"strings"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAI implements the Provider interface for OpenAI-compatible APIs.
// Works with OpenAI, Groq, and other OpenAI-compatible endpoints.
type OpenAI struct {
	name        string
	displayName string
	baseURL     string
	apiKey      string
	models      []Model
	client      HTTPClient
}

var _ Provider = (*OpenAI)(nil)

// NewOpenAI creates a new OpenAI provider.
func NewOpenAI(config Config) *OpenAI {

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAI{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     baseURL,
		apiKey:      config.APIKey,
		models:      config.Models,
		client:      defaultHTTPClient(),
	}
}

// Name returns the provider identifier.
func (o *OpenAI) Name() string {

	return o.name
}

// DisplayName returns the human-readable provider name.
func (o *OpenAI) DisplayName() string {

	return o.displayName
}

// Models returns the available models.
func (o *OpenAI) Models() []Model {

	return o.models
}

// CredentialFields returns the expected credential inputs.
func (o *OpenAI) CredentialFields() []CredentialField {

	return []CredentialField{
		{
			Name:     CredentialAPIKey,
			Label:    "API Key",
			Required: true,
			Secret:   true,
		},
	}
}

// Configure updates the provider configuration.
func (o *OpenAI) Configure(config Config) error {

	if config.Credentials != nil {
		if value, ok := config.Credentials[CredentialAPIKey]; ok {
			o.apiKey = strings.TrimSpace(value)
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
func (o *OpenAI) SetHTTPClient(client HTTPClient) {

	if client != nil {
		o.client = client
	}
}

// httpClient returns the configured HTTP client or a default client.
func (o *OpenAI) httpClient() HTTPClient {

	if o.client == nil {
		o.client = defaultHTTPClient()
	}
	return o.client
}

// TestConnection verifies the API is reachable.
func (o *OpenAI) TestConnection(ctx context.Context) error {

	if o.usesOpenAISDK() {
		return o.testConnectionSDK(ctx)
	}

	headers := o.authHeaders()
	_, err := listOpenAICompatModels(ctx, o.httpClient(), o.baseURL, headers)
	return err
}

// ListResources fetches the available models from an OpenAI-compatible API.
func (o *OpenAI) ListResources(ctx context.Context) ([]Model, error) {

	if o.usesOpenAISDK() {
		return o.listResourcesSDK(ctx)
	}

	headers := o.authHeaders()
	return listOpenAICompatModels(ctx, o.httpClient(), o.baseURL, headers)
}

// Chat implements streaming chat completion.
func (o *OpenAI) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	if o.usesOpenAISDK() {
		return o.chatSDK(ctx, messages, opts)
	}

	headers := o.authHeaders()
	return chatOpenAICompat(ctx, o.httpClient(), o.baseURL, headers, messages, opts)
}

// usesOpenAISDK decides whether to use the OpenAI SDK.
func (o *OpenAI) usesOpenAISDK() bool {

	return strings.Contains(strings.ToLower(o.baseURL), "api.openai.com")
}

// authHeaders returns configured request headers.
func (o *OpenAI) authHeaders() map[string]string {

	if strings.TrimSpace(o.apiKey) == "" {
		return nil
	}
	return map[string]string{
		"Authorization": "Bearer " + o.apiKey,
	}
}

// normalizeBaseURL ensures the base URL ends with a slash.
func (o *OpenAI) normalizeBaseURL() string {

	baseURL := strings.TrimSpace(o.baseURL)
	if baseURL == "" {
		return baseURL
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL
}

// newSDKClient constructs an OpenAI SDK client.
func (o *OpenAI) newSDKClient() openai.Client {

	opts := []option.RequestOption{option.WithAPIKey(o.apiKey)}
	if o.baseURL != "" {
		opts = append(opts, option.WithBaseURL(o.normalizeBaseURL()))
	}
	return openai.NewClient(opts...)
}

// testConnectionSDK validates connectivity using the OpenAI SDK.
func (o *OpenAI) testConnectionSDK(ctx context.Context) error {

	client := o.newSDKClient()
	_, err := client.Models.List(ctx)
	if err != nil {
		return o.wrapOpenAIError(err)
	}
	return nil
}

// listResourcesSDK lists models using the OpenAI SDK.
func (o *OpenAI) listResourcesSDK(ctx context.Context) ([]Model, error) {

	client := o.newSDKClient()
	page, err := client.Models.List(ctx)
	if err != nil {
		return nil, o.wrapOpenAIError(err)
	}

	models := make([]Model, 0, len(page.Data))
	for _, item := range page.Data {
		if item.ID == "" {
			continue
		}
		models = append(models, Model{
			ID:   item.ID,
			Name: item.ID,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	return models, nil
}

// chatSDK executes chat requests with the OpenAI SDK.
func (o *OpenAI) chatSDK(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	client := o.newSDKClient()
	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(opts.Model),
		Messages: o.toSDKMessages(messages),
	}
	if opts.Temperature > 0 {
		params.Temperature = openai.Float(opts.Temperature)
	}
	if opts.MaxTokens > 0 {
		params.MaxTokens = openai.Int(int64(opts.MaxTokens))
	}
	if opts.Stream {
		params.StreamOptions = openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		}
	}

	if !opts.Stream {
		resp, err := client.Chat.Completions.New(ctx, params)
		if err != nil {
			return nil, o.wrapOpenAIError(err)
		}
		chunks := make(chan Chunk, 1)
		go func() {
			defer close(chunks)
			if len(resp.Choices) == 0 {
				return
			}
			choice := resp.Choices[0]
			chunk := Chunk{
				Content:      choice.Message.Content,
				FinishReason: choice.FinishReason,
				Usage:        o.toUsageStats(resp.Usage),
			}
			chunks <- chunk
		}()
		return chunks, nil
	}

	stream := client.Chat.Completions.NewStreaming(ctx, params)
	chunks := make(chan Chunk, 100)
	go func() {
		defer close(chunks)
		defer stream.Close()

		for stream.Next() {
			cur := stream.Current()
			usage := o.toUsageStatsFromChunk(cur)
			content := ""
			finishReason := ""

			if len(cur.Choices) > 0 {
				choice := cur.Choices[0]
				content = choice.Delta.Content
				finishReason = choice.FinishReason
			}

			if content == "" && finishReason == "" && usage == nil {
				continue
			}

			chunks <- Chunk{
				Content:      content,
				Model:        cur.Model,
				FinishReason: finishReason,
				Usage:        usage,
			}
		}

		if err := stream.Err(); err != nil {
			chunks <- Chunk{Error: o.wrapOpenAIError(err)}
		}
	}()

	return chunks, nil
}

// toSDKMessages converts chat messages to SDK message params.
func (o *OpenAI) toSDKMessages(messages []ProviderMessage) []openai.ChatCompletionMessageParamUnion {

	result := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		content := msg.Content
		if strings.TrimSpace(content) == "" {
			continue
		}

		switch msg.Role {
		case RoleSystem:
			result = append(result, openai.SystemMessage(content))
		case RoleAssistant:
			result = append(result, openai.AssistantMessage(content))
		case RoleUser:
			result = append(result, openai.UserMessage(content))
		default:
			result = append(result, openai.UserMessage(content))
		}
	}
	return result
}

// toUsageStats converts SDK usage to provider usage stats.
func (o *OpenAI) toUsageStats(usage openai.CompletionUsage) *UsageStats {

	if usage.TotalTokens == 0 && usage.PromptTokens == 0 && usage.CompletionTokens == 0 {
		return nil
	}
	return &UsageStats{
		PromptTokens:     int(usage.PromptTokens),
		CompletionTokens: int(usage.CompletionTokens),
		TotalTokens:      int(usage.TotalTokens),
	}
}

// toUsageStatsFromChunk extracts usage stats from a streaming chunk.
func (o *OpenAI) toUsageStatsFromChunk(chunk openai.ChatCompletionChunk) *UsageStats {

	if !chunk.JSON.Usage.Valid() {
		return nil
	}
	return o.toUsageStats(chunk.Usage)
}

// wrapOpenAIError normalizes SDK errors into APIError when possible.
func (o *OpenAI) wrapOpenAIError(err error) error {

	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		return &APIError{Code: apiErr.StatusCode, Message: apiErr.Message}
	}
	return err
}

// Ensure OpenAI implements Provider.
var _ Provider = (*OpenAI)(nil)
