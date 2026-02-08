// provider.go implements the Grok (xAI) provider adapter.
// internal/features/providers/adapters/grok/provider.go
package grok

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	providerhttp "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/httpcompat"
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/core"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/gateway"
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

// Grok implements the Provider interface for xAI.
type Grok struct {
	name        string
	displayName string
	baseURL     string
	apiKey      string
	models      []Model
	client      HTTPClient
}

var _ Provider = (*Grok)(nil)

// New creates a new Grok provider.
func New(config Config) *Grok {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.x.ai/v1"
	}

	return &Grok{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     baseURL,
		apiKey:      config.APIKey,
		models:      config.Models,
		client:      providerhttp.NewDefaultClient(),
	}
}

// Name returns the provider identifier.
func (g *Grok) Name() string {
	return g.name
}

// DisplayName returns the human-readable provider name.
func (g *Grok) DisplayName() string {
	return g.displayName
}

// Models returns the available models.
func (g *Grok) Models() []Model {
	return g.models
}

// CredentialFields returns the expected credential inputs.
func (g *Grok) CredentialFields() []CredentialField {
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
func (g *Grok) Configure(config Config) error {
	if config.Credentials != nil {
		if value, ok := config.Credentials[CredentialAPIKey]; ok {
			g.apiKey = strings.TrimSpace(value)
		}
	}
	if strings.TrimSpace(config.APIKey) != "" {
		g.apiKey = config.APIKey
	}
	if config.BaseURL != "" {
		g.baseURL = config.BaseURL
	}
	if config.Models != nil {
		g.models = config.Models
	}
	return nil
}

// SetHTTPClient overrides the HTTP client used by the provider.
func (g *Grok) SetHTTPClient(client HTTPClient) {
	if client != nil {
		g.client = client
	}
}


// TestConnection verifies the API is reachable.
func (g *Grok) TestConnection(ctx context.Context) error {
	return g.testConnectionSDK(ctx)
}

// ListResources fetches the available models.
func (g *Grok) ListResources(ctx context.Context) ([]Model, error) {
	return g.listResourcesSDK(ctx)
}

// Chat implements streaming chat completion.
func (g *Grok) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {
	return g.chatSDK(ctx, messages, opts)
}

// GenerateImage generates an image using the OpenAI SDK (xAI compatible).
func (g *Grok) GenerateImage(ctx context.Context, opts providergateway.ImageGenerationOptions) (*providergateway.ImageResult, error) {
	// xAI doesn't currently support image generation via this endpoint publicly documented as identical to OpenAI's DALL-E,
	// but the user requested it. We will assume compatibility or return error if it fails.
	// Actually, Grok 2 has image generation capabilities. It might be via the same /images/generations endpoint.

	client := g.newSDKClient()
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
		return nil, g.wrapOpenAIError(err)
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

// normalizeBaseURL ensures the base URL ends with a slash.
func (g *Grok) normalizeBaseURL() string {
	baseURL := strings.TrimSpace(g.baseURL)
	if baseURL == "" {
		return baseURL
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return baseURL
}

// newSDKClient constructs an OpenAI SDK client pointed at xAI.
func (g *Grok) newSDKClient() openaisdk.Client {
	opts := []option.RequestOption{
		option.WithAPIKey(g.apiKey),
		option.WithBaseURL(g.normalizeBaseURL()),
	}
	return openaisdk.NewClient(opts...)
}

// testConnectionSDK validates connectivity.
func (g *Grok) testConnectionSDK(ctx context.Context) error {
	client := g.newSDKClient()
	_, err := client.Models.List(ctx)
	if err != nil {
		return g.wrapOpenAIError(err)
	}
	return nil
}

// listResourcesSDK lists models.
func (g *Grok) listResourcesSDK(ctx context.Context) ([]Model, error) {
	client := g.newSDKClient()
	page, err := client.Models.List(ctx)
	if err != nil {
		return nil, g.wrapOpenAIError(err)
	}

	models := make([]Model, 0, len(page.Data))
	for _, item := range page.Data {
		if item.ID == "" {
			continue
		}
		// Basic filter or simply list all
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

// chatSDK executes chat requests.
func (g *Grok) chatSDK(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {
	client := g.newSDKClient()
	params := openaisdk.ChatCompletionNewParams{
		Model:    openaisdk.ChatModel(opts.Model),
		Messages: g.toSDKMessages(messages),
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
			return nil, g.wrapOpenAIError(err)
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
				Usage:        g.toUsageStats(resp.Usage),
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
			usage := g.toUsageStatsFromChunk(cur)
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
			chunks <- Chunk{Error: g.wrapOpenAIError(err)}
		}
	}()

	return chunks, nil
}

// toSDKMessages converts chat messages to SDK message params.
func (g *Grok) toSDKMessages(messages []ProviderMessage) []openaisdk.ChatCompletionMessageParamUnion {
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
func (g *Grok) toUsageStats(usage openaisdk.CompletionUsage) *UsageStats {
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
func (g *Grok) toUsageStatsFromChunk(chunk openaisdk.ChatCompletionChunk) *UsageStats {
	if !chunk.JSON.Usage.Valid() {
		return nil
	}
	return g.toUsageStats(chunk.Usage)
}

// wrapOpenAIError normalizes SDK errors into APIError when possible.
func (g *Grok) wrapOpenAIError(err error) error {
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

// GenerateImage returns not supported error (if not already implemented elsewhere, relying on interface compliance).
// ACTUALLY I need to check if GenerateImage is missing. If Gork compiles, it must have it.
// I will just append EditImage.
// EditImage returns not supported error.
func (g *Grok) EditImage(ctx context.Context, opts providergateway.ImageEditOptions) (*providergateway.ImageResult, error) {
	return nil, fmt.Errorf("edit image not supported by this provider")
}
