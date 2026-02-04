// Implement the provider interface for OpenAI.
// internal/features/settings/adapters/provider/openai.go
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// Configure updates the provider configuration.
func (o *OpenAI) Configure(config Config) error {

	o.apiKey = config.APIKey
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

	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{Code: resp.StatusCode, Message: string(body)}
	}

	return nil
}

// ListResources fetches the available models from an OpenAI-compatible API.
func (o *OpenAI) ListResources(ctx context.Context) ([]Model, error) {

	if o.usesOpenAISDK() {
		return o.listResourcesSDK(ctx)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{Code: resp.StatusCode, Message: string(body)}
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(payload.Data))
	for _, item := range payload.Data {
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

// Chat implements streaming chat completion.
func (o *OpenAI) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	if o.usesOpenAISDK() {
		return o.chatSDK(ctx, messages, opts)
	}

	// Convert messages to OpenAI format
	apiMessages := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		content := msg.Content
		if strings.TrimSpace(content) == "" {
			continue
		}
		apiMessages = append(apiMessages, map[string]interface{}{
			"role":    string(msg.Role),
			"content": content,
		})
	}

	reqBody := map[string]interface{}{
		"model":    opts.Model,
		"messages": apiMessages,
		"stream":   opts.Stream,
	}

	if opts.Temperature > 0 {
		reqBody["temperature"] = opts.Temperature
	}
	if opts.MaxTokens > 0 {
		reqBody["max_tokens"] = opts.MaxTokens
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, &APIError{Code: resp.StatusCode, Message: string(body)}
	}

	chunks := make(chan Chunk, 100)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		if opts.Stream {
			o.streamResponse(resp.Body, chunks)
		} else {
			o.parseResponse(resp.Body, chunks)
		}
	}()

	return chunks, nil
}

// usesOpenAISDK decides whether to use the OpenAI SDK.
func (o *OpenAI) usesOpenAISDK() bool {

	return strings.Contains(strings.ToLower(o.baseURL), "api.openai.com")
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

// streamResponse parses SSE responses into chunks.
func (o *OpenAI) streamResponse(body io.Reader, chunks chan<- Chunk) {

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			chunks <- Chunk{FinishReason: "stop"}
			return
		}

		var resp struct {
			Model   string `json:"model"`
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage *UsageStats `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			continue
		}

		if len(resp.Choices) > 0 {
			choice := resp.Choices[0]
			chunk := Chunk{
				Content:      choice.Delta.Content,
				Model:        resp.Model,
				FinishReason: choice.FinishReason,
			}
			if resp.Usage != nil {
				chunk.Usage = resp.Usage
			}
			chunks <- chunk
		}
	}
	if err := scanner.Err(); err != nil {
		chunks <- Chunk{Error: err}
	}
}

// parseResponse parses a non-streaming response into a chunk.
func (o *OpenAI) parseResponse(body io.Reader, chunks chan<- Chunk) {

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *UsageStats `json:"usage"`
	}

	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		chunks <- Chunk{Error: err}
		return
	}

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		chunks <- Chunk{
			Content:      choice.Message.Content,
			FinishReason: choice.FinishReason,
			Usage:        resp.Usage,
		}
	}
}

// Ensure OpenAI implements Provider.
var _ Provider = (*OpenAI)(nil)
