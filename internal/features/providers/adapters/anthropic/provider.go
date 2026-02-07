// provider.go implements the Anthropic provider adapter.
// internal/features/providers/adapters/anthropic/provider.go
package anthropic

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
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
type UsageStats = providergateway.UsageStats
type Provider = providercore.Provider
type HTTPClient = providerhttp.Client

const (
	defaultAnthropicMaxTokens = 1024
	CredentialAPIKey          = providercore.CredentialAPIKey
	RoleUser                  = providergateway.RoleUser
	RoleAssistant             = providergateway.RoleAssistant
	RoleSystem                = providergateway.RoleSystem
)

// Anthropic implements the Provider interface for Anthropic.
type Anthropic struct {
	name        string
	displayName string
	baseURL     string
	apiKey      string
	models      []Model
	client      HTTPClient
}

var _ Provider = (*Anthropic)(nil)

// New creates a new Anthropic provider.
func New(config Config) *Anthropic {

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	return &Anthropic{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     baseURL,
		apiKey:      config.APIKey,
		models:      config.Models,
		client:      providerhttp.NewDefaultClient(),
	}
}

// Name returns the provider identifier.
func (a *Anthropic) Name() string {

	return a.name
}

// DisplayName returns the human-readable provider name.
func (a *Anthropic) DisplayName() string {

	return a.displayName
}

// Models returns the available models.
func (a *Anthropic) Models() []Model {

	return a.models
}

// CredentialFields returns the expected credential inputs.
func (a *Anthropic) CredentialFields() []CredentialField {

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
func (a *Anthropic) Configure(config Config) error {

	if config.Credentials != nil {
		if value, ok := config.Credentials[CredentialAPIKey]; ok {
			a.apiKey = strings.TrimSpace(value)
		}
	}
	if strings.TrimSpace(config.APIKey) != "" {
		a.apiKey = config.APIKey
	}
	if config.BaseURL != "" {
		a.baseURL = config.BaseURL
	}
	if config.Models != nil {
		a.models = config.Models
	}
	return nil
}

// SetHTTPClient overrides the HTTP client used by the provider.
func (a *Anthropic) SetHTTPClient(client HTTPClient) {

	if client != nil {
		a.client = client
	}
}

// httpClient returns the configured HTTP client or a default client.
func (a *Anthropic) httpClient() HTTPClient {

	if a.client == nil {
		a.client = providerhttp.NewDefaultClient()
	}
	return a.client
}

// TestConnection verifies the API is reachable.
func (a *Anthropic) TestConnection(ctx context.Context) error {

	_, err := a.ListResources(ctx)
	return err
}

// ListResources fetches the available models from Anthropic.
func (a *Anthropic) ListResources(ctx context.Context) ([]Model, error) {

	client := a.newSDKClient()
	page, err := client.Models.List(ctx, anthropicsdk.ModelListParams{})
	if err != nil {
		return nil, a.wrapAnthropicError(err)
	}

	models := make([]Model, 0, len(page.Data))
	for _, item := range page.Data {
		if item.ID == "" {
			continue
		}
		name := strings.TrimSpace(item.DisplayName)
		if name == "" {
			name = item.ID
		}
		models = append(models, Model{
			ID:   item.ID,
			Name: name,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	return models, nil
}

// Chat implements streaming chat completion.
func (a *Anthropic) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	if strings.TrimSpace(opts.Model) == "" {
		return nil, fmt.Errorf("model required")
	}

	apiMessages, systemBlocks := a.toAnthropicMessages(messages)
	if len(apiMessages) == 0 {
		return nil, fmt.Errorf("messages required")
	}

	params := anthropicsdk.MessageNewParams{
		Model:     anthropicsdk.Model(opts.Model),
		MaxTokens: int64(a.resolveMaxTokens(opts)),
		Messages:  apiMessages,
	}
	if len(systemBlocks) > 0 {
		params.System = systemBlocks
	}
	if opts.Temperature > 0 {
		params.Temperature = anthropicsdk.Float(opts.Temperature)
	}
	if len(opts.StopWords) > 0 {
		params.StopSequences = opts.StopWords
	}

	if opts.Stream {
		return a.chatStreaming(ctx, params)
	}
	return a.chatOnce(ctx, params)
}

// resolveMaxTokens returns a valid max token value for Anthropic.
func (a *Anthropic) resolveMaxTokens(opts ChatOptions) int {

	if opts.MaxTokens > 0 {
		return opts.MaxTokens
	}
	return defaultAnthropicMaxTokens
}

// toAnthropicMessages converts provider messages into Anthropic params.
func (a *Anthropic) toAnthropicMessages(messages []ProviderMessage) ([]anthropicsdk.MessageParam, []anthropicsdk.TextBlockParam) {

	apiMessages := make([]anthropicsdk.MessageParam, 0, len(messages))
	systemBlocks := make([]anthropicsdk.TextBlockParam, 0)

	for _, msg := range messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}

		if msg.Role == RoleSystem {
			systemBlocks = append(systemBlocks, anthropicsdk.TextBlockParam{Text: content})
			continue
		}

		block := anthropicsdk.NewTextBlock(content)
		if msg.Role == RoleAssistant {
			apiMessages = append(apiMessages, anthropicsdk.NewAssistantMessage(block))
			continue
		}

		apiMessages = append(apiMessages, anthropicsdk.NewUserMessage(block))
	}

	return apiMessages, systemBlocks
}

// chatOnce executes a non-streaming Anthropic chat request.
func (a *Anthropic) chatOnce(ctx context.Context, params anthropicsdk.MessageNewParams) (<-chan Chunk, error) {

	client := a.newSDKClient()
	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return nil, a.wrapAnthropicError(err)
	}

	content := a.textFromContentBlocks(resp.Content)
	chunks := make(chan Chunk, 1)
	go func() {
		defer close(chunks)
		chunks <- Chunk{
			Content:      content,
			Model:        string(resp.Model),
			FinishReason: a.mapStopReason(resp.StopReason),
			Usage:        a.toUsageStats(resp.Usage),
		}
	}()

	return chunks, nil
}

// chatStreaming executes a streaming Anthropic chat request.
func (a *Anthropic) chatStreaming(ctx context.Context, params anthropicsdk.MessageNewParams) (<-chan Chunk, error) {

	client := a.newSDKClient()
	stream := client.Messages.NewStreaming(ctx, params)
	chunks := make(chan Chunk, 100)

	go func() {
		defer close(chunks)
		defer func() { _ = stream.Close() }()

		message := anthropicsdk.Message{}
		for stream.Next() {
			event := stream.Current()
			if err := message.Accumulate(event); err != nil {
				chunks <- Chunk{Error: err}
				return
			}

			switch variant := event.AsAny().(type) {
			case anthropicsdk.ContentBlockDeltaEvent:
				delta := variant.Delta.AsAny()
				if textDelta, ok := delta.(anthropicsdk.TextDelta); ok {
					if text := textDelta.Text; text != "" {
						chunks <- Chunk{
							Content: text,
							Model:   string(message.Model),
						}
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			chunks <- Chunk{Error: a.wrapAnthropicError(err)}
			return
		}

		chunks <- Chunk{
			Model:        string(message.Model),
			FinishReason: a.mapStopReason(message.StopReason),
			Usage:        a.toUsageStats(message.Usage),
		}
	}()

	return chunks, nil
}

// textFromContentBlocks flattens Anthropic content blocks into plain text.
func (a *Anthropic) textFromContentBlocks(blocks []anthropicsdk.ContentBlockUnion) string {

	if len(blocks) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, block := range blocks {
		if block.Type != "text" {
			continue
		}
		if block.Text == "" {
			continue
		}
		builder.WriteString(block.Text)
	}
	return builder.String()
}

// toUsageStats converts Anthropic usage into provider usage stats.
func (a *Anthropic) toUsageStats(usage anthropicsdk.Usage) *UsageStats {

	if usage.InputTokens == 0 && usage.OutputTokens == 0 {
		return nil
	}
	prompt := int(usage.InputTokens)
	completion := int(usage.OutputTokens)
	return &UsageStats{
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      prompt + completion,
	}
}

// mapStopReason normalizes Anthropic stop reasons for the provider response.
func (a *Anthropic) mapStopReason(reason anthropicsdk.StopReason) string {

	if reason == "" {
		return ""
	}
	return string(reason)
}

// newSDKClient constructs an Anthropic SDK client.
func (a *Anthropic) newSDKClient() anthropicsdk.Client {

	opts := []option.RequestOption{
		option.WithAPIKey(a.apiKey),
		option.WithHTTPClient(a.httpClient()),
	}
	if a.baseURL != "" {
		opts = append(opts, option.WithBaseURL(a.baseURL))
	}
	return anthropicsdk.NewClient(opts...)
}

// wrapAnthropicError normalizes SDK errors into APIError when possible.
func (a *Anthropic) wrapAnthropicError(err error) error {

	var apiErr *anthropicsdk.Error
	if errors.As(err, &apiErr) {
		return &providerhttp.APIError{Code: apiErr.StatusCode, Message: apiErr.Error()}
	}
	return err
}

// GenerateImage generates an image. Anthropic does not support this capability.
func (a *Anthropic) GenerateImage(ctx context.Context, opts providergateway.ImageGenerationOptions) (*providergateway.ImageResult, error) {
	return nil, fmt.Errorf("anthropic provider does not support image generation")
}

// EditImage returns not supported error.
func (a *Anthropic) EditImage(ctx context.Context, opts providergateway.ImageEditOptions) (*providergateway.ImageResult, error) {
	return nil, fmt.Errorf("edit image not supported by this provider")
}
