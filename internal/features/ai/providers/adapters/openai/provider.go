// provider.go implements the OpenAI provider adapter.
// internal/features/ai/providers/adapters/openai/provider.go
package openai

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	providerhttp "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/httpcompat"
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/core"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/gateway"
	openaisdk "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type Model = providercore.Model
type Config = providercore.ProviderConfig
type ChatOptions = providergateway.ChatOptions
type ProviderMessage = providergateway.ProviderMessage
type Chunk = providergateway.Chunk
type CredentialField = providercore.CredentialField
type UsageStats = providergateway.UsageStats
type Provider = providercore.Provider
type HTTPClient = providerhttp.Client

const (
	CredentialAPIKey = providercore.CredentialAPIKey
	RoleUser         = providergateway.RoleUser
	RoleAssistant    = providergateway.RoleAssistant
	RoleSystem       = providergateway.RoleSystem
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

// New creates a new OpenAI provider.
func New(config Config) *OpenAI {

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
		client:      providerhttp.NewDefaultClient(),
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
		o.client = providerhttp.NewDefaultClient()
	}
	return o.client
}

// TestConnection verifies the API is reachable.
func (o *OpenAI) TestConnection(ctx context.Context) error {

	if o.usesOpenAISDK() {
		return o.testConnectionSDK(ctx)
	}

	headers := o.authHeaders()
	_, err := providerhttp.ListOpenAICompatModels(ctx, o.httpClient(), o.baseURL, headers)
	return err
}

// ListResources fetches the available models from an OpenAI-compatible API.
func (o *OpenAI) ListResources(ctx context.Context) ([]Model, error) {

	if o.usesOpenAISDK() {
		return o.listResourcesSDK(ctx)
	}

	headers := o.authHeaders()
	return providerhttp.ListOpenAICompatModels(ctx, o.httpClient(), o.baseURL, headers)
}

// Chat implements streaming chat completion.
func (o *OpenAI) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	if o.usesOpenAISDK() {
		return o.chatSDK(ctx, messages, opts)
	}

	headers := o.authHeaders()
	return providerhttp.ChatOpenAICompat(ctx, o.httpClient(), o.baseURL, headers, messages, opts)
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
func (o *OpenAI) newSDKClient() openaisdk.Client {

	opts := []option.RequestOption{option.WithAPIKey(o.apiKey)}
	if o.baseURL != "" {
		opts = append(opts, option.WithBaseURL(o.normalizeBaseURL()))
	}
	return openaisdk.NewClient(opts...)
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
	params := openaisdk.ChatCompletionNewParams{
		Model:    openaisdk.ChatModel(opts.Model),
		Messages: o.toSDKMessages(messages),
	}
	if opts.Temperature > 0 {
		params.Temperature = openaisdk.Float(opts.Temperature)
	}
	if opts.MaxTokens > 0 {
		params.MaxTokens = openaisdk.Int(int64(opts.MaxTokens))
	}
	if opts.Stream {
		params.StreamOptions = openaisdk.ChatCompletionStreamOptionsParam{
			IncludeUsage: openaisdk.Bool(true),
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
		defer func() { _ = stream.Close() }()

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
func (o *OpenAI) toSDKMessages(messages []ProviderMessage) []openaisdk.ChatCompletionMessageParamUnion {

	result := make([]openaisdk.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, msg := range messages {
		content := msg.Content
		if strings.TrimSpace(content) == "" {
			continue
		}

		switch msg.Role {
		case RoleSystem:
			result = append(result, openaisdk.SystemMessage(content))
		case RoleAssistant:
			result = append(result, openaisdk.AssistantMessage(content))
		case RoleUser:
			result = append(result, openaisdk.UserMessage(content))
		default:
			result = append(result, openaisdk.UserMessage(content))
		}
	}
	return result
}

// toUsageStats converts SDK usage to provider usage stats.
func (o *OpenAI) toUsageStats(usage openaisdk.CompletionUsage) *UsageStats {

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
func (o *OpenAI) toUsageStatsFromChunk(chunk openaisdk.ChatCompletionChunk) *UsageStats {

	if !chunk.JSON.Usage.Valid() {
		return nil
	}
	return o.toUsageStats(chunk.Usage)
}

// wrapOpenAIError normalizes SDK errors into APIError when possible.
// wrapOpenAIError normalizes SDK errors into APIError when possible.
func (o *OpenAI) wrapOpenAIError(err error) error {

	var apiErr *openaisdk.Error
	if errors.As(err, &apiErr) {
		message := strings.TrimSpace(apiErr.Message)
		if message == "" {
			message = strings.TrimSpace(apiErr.RawJSON())
		}
		return &providerhttp.APIError{Code: apiErr.StatusCode, Message: message}
	}
	return err
}

// GenerateImage generates an image using the OpenAI SDK.
func (o *OpenAI) GenerateImage(ctx context.Context, opts providergateway.ImageGenerationOptions) (*providergateway.ImageResult, error) {

	client := o.newSDKClient()
	params := openaisdk.ImageGenerateParams{
		Prompt: opts.Prompt,
		Model:  openaisdk.ImageModel(opts.Model),
	}

	if opts.N > 0 {
		params.N = openaisdk.Int(int64(opts.N))
	}
	if opts.Size != "" {
		params.Size = openaisdk.ImageGenerateParamsSize(opts.Size)
	}
	if opts.Quality != "" {
		params.Quality = openaisdk.ImageGenerateParamsQuality(opts.Quality)
	}
	if opts.Style != "" {
		params.Style = openaisdk.ImageGenerateParamsStyle(opts.Style)
	}
	if opts.ResponseFormat != "" {
		params.ResponseFormat = openaisdk.ImageGenerateParamsResponseFormat(opts.ResponseFormat)
	}
	if opts.User != "" {
		params.User = openaisdk.String(opts.User)
	}

	resp, err := client.Images.Generate(ctx, params)
	if err != nil {
		return nil, o.wrapOpenAIError(err)
	}

	result := &providergateway.ImageResult{
		Created: resp.Created,
		Data:    make([]providergateway.ImageData, len(resp.Data)),
	}

	for i, d := range resp.Data {
		result.Data[i] = providergateway.ImageData{
			URL:           d.URL,
			B64JSON:       d.B64JSON,
			RevisedPrompt: d.RevisedPrompt,
		}
	}

	return result, nil
}

// EditImage returns not supported error.
// EditImage returns not supported error.
func (o *OpenAI) EditImage(ctx context.Context, opts providergateway.ImageEditOptions) (*providergateway.ImageResult, error) {
	return nil, fmt.Errorf("edit image not supported by this provider")
}
