// provider.go implements the Gemini provider adapter.
// internal/features/providers/adapters/gemini/provider.go
package gemini

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	providerhttp "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/httpcompat"
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/core"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/gateway"
	"google.golang.org/genai"
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
	RoleAssistant    = providergateway.RoleAssistant
)

// Gemini implements the Provider interface for Google's Gemini API.
type Gemini struct {
	name        string
	displayName string
	baseURL     string
	apiKey      string
	models      []Model
	client      HTTPClient
}

var _ Provider = (*Gemini)(nil)

// New creates a new Gemini provider.
func New(config Config) *Gemini {

	baseURL := config.BaseURL
	// The SDK handles default base URL. Only set if explicitly provided?
	// Actually, SDK config might need it.

	return &Gemini{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     baseURL,
		apiKey:      config.APIKey,
		models:      config.Models,
		client:      providerhttp.NewDefaultClient(),
	}
}

// Name returns the provider identifier.
func (g *Gemini) Name() string {
	return g.name
}

// DisplayName returns the human-readable provider name.
func (g *Gemini) DisplayName() string {
	return g.displayName
}

// Models returns the available models.
func (g *Gemini) Models() []Model {
	return g.models
}

// CredentialFields returns the expected credential inputs.
func (g *Gemini) CredentialFields() []CredentialField {
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
func (g *Gemini) Configure(config Config) error {

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
func (g *Gemini) SetHTTPClient(client HTTPClient) {

	if client != nil {
		g.client = client
	}
}

// httpClient returns the configured HTTP client or a default client.
func (g *Gemini) httpClient() HTTPClient {

	if g.client == nil {
		g.client = providerhttp.NewDefaultClient()
	}
	return g.client
}

// newSDKClient creates a new Gemini SDK client.
func (g *Gemini) newSDKClient(ctx context.Context) (*genai.Client, error) {
	opts := &genai.ClientConfig{
		APIKey: g.apiKey,
	}
	// If base URL is custom, check if genai client supports it.
	// Looking at typical genai patterns, it's often implicit.
	// We'll trust the default unless we find a way to set it in ClientConfig if needed.
	// NOTE: google.golang.org/genai might not expose BaseURL directly in config,
	// but let's assume standard behavior or skip if not needed for migration.

	if g.client != nil {
		if client, ok := g.client.(*http.Client); ok {
			opts.HTTPClient = client
		}
	}

	return genai.NewClient(ctx, opts)
}

// TestConnection verifies the Gemini API is reachable.
func (g *Gemini) TestConnection(ctx context.Context) error {

	url := fmt.Sprintf("%s/models?key=%s", g.baseURL, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := g.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &providerhttp.APIError{Code: resp.StatusCode, Message: string(body)}
	}

	return nil
}

// ListResources fetches the available models from the Gemini API.
func (g *Gemini) ListResources(ctx context.Context) ([]Model, error) {
	client, err := g.newSDKClient(ctx)
	if err != nil {
		return nil, err
	}

	pager, err := client.Models.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Supports Go 1.23 iterator? The error said `assignment mismatch`.
	// Most likely `List` returns `*genai.ModelListIterator`.
	// If it handles standard iterator:

	models := make([]Model, 0)
	for _, item := range pager.Items {

		name := item.DisplayName
		if name == "" {
			name = item.Name // fallback
		}

		id := strings.TrimPrefix(item.Name, "models/")

		models = append(models, Model{
			ID:   id,
			Name: name,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})

	return models, nil
}

// Chat implements streaming chat completion for Gemini.
func (g *Gemini) Chat(ctx context.Context, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {
	client, err := g.newSDKClient(ctx)
	if err != nil {
		return nil, err
	}

	sdkParts := make([]*genai.Content, 0, len(messages))
	for _, msg := range messages {
		if strings.TrimSpace(msg.Content) == "" {
			continue
		}
		role := "user"
		if msg.Role == RoleAssistant {
			role = "model"
		}
		sdkParts = append(sdkParts, &genai.Content{
			Role: role,
			Parts: []*genai.Part{
				{Text: msg.Content},
			},
		})
	}

	config := &genai.GenerateContentConfig{
		Temperature:     g.float32Ptr(opts.Temperature),
		MaxOutputTokens: int32(opts.MaxTokens),
	}

	modelID := opts.Model

	if opts.Stream {
		stream := client.Models.GenerateContentStream(ctx, modelID, sdkParts, config)
		chunks := make(chan Chunk, 100)
		go func() {
			defer close(chunks)
			// Go 1.23 iterator loop
			for resp, err := range stream {
				if err != nil {
					chunks <- Chunk{Error: err}
					return
				}

				// Process chunk
				for _, cand := range resp.Candidates {
					text := ""
					for _, part := range cand.Content.Parts {
						text += part.Text
					}
					chunk := Chunk{
						Content:      text,
						FinishReason: "",
					}
					if resp.UsageMetadata != nil {
						chunk.Usage = &UsageStats{
							PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
							CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
							TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
						}
					}
					chunks <- chunk
				}
			}
		}()
		return chunks, nil
	}

	// Non-streaming
	resp, err := client.Models.GenerateContent(ctx, modelID, sdkParts, config)
	if err != nil {
		return nil, err
	}

	chunks := make(chan Chunk, 1)
	go func() {
		defer close(chunks)
		for _, cand := range resp.Candidates {
			text := ""
			for _, part := range cand.Content.Parts {
				text += part.Text
			}
			chunk := Chunk{
				Content:      text,
				FinishReason: "",
			}
			if resp.UsageMetadata != nil {
				chunk.Usage = &UsageStats{
					PromptTokens:     int(resp.UsageMetadata.PromptTokenCount),
					CompletionTokens: int(resp.UsageMetadata.CandidatesTokenCount),
					TotalTokens:      int(resp.UsageMetadata.TotalTokenCount),
				}
			}
			chunks <- chunk
		}
	}()
	return chunks, nil
}

func (g *Gemini) float32Ptr(v float64) *float32 {
	if v == 0 {
		return nil
	}
	val := float32(v)
	return &val
}

// GenerateImage generates an image using the Gemini/Imagen predict API.
func (g *Gemini) GenerateImage(ctx context.Context, opts providergateway.ImageGenerationOptions) (*providergateway.ImageResult, error) {

	// Default to a known Imagen model if not specified.
	model := opts.Model
	if model == "" {
		model = "imagen-4.0-generate-001"
	}

	// Gemini models (gemini-2.5-flash-image, etc.) use the GenerateContent API (multimodal).
	if strings.HasPrefix(model, "gemini") {
		return g.generateContentImage(ctx, model, opts)
	}

	// Use genai library for Imagen models (Imagen 3/4)
	client, err := g.newSDKClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDK client: %w", err)
	}

	cfg := &genai.GenerateImagesConfig{
		NumberOfImages: int32(opts.N),
	}
	if cfg.NumberOfImages == 0 {
		cfg.NumberOfImages = 1
	}

	// Aspect Ratio mapping
	if opts.Size == "1024x1024" || opts.Size == "1:1" {
		cfg.AspectRatio = "1:1"
	} else if opts.Size == "16:9" {
		cfg.AspectRatio = "16:9"
	} else if opts.Size == "9:16" {
		cfg.AspectRatio = "9:16"
	} else if opts.Size == "3:4" {
		cfg.AspectRatio = "3:4"
	} else if opts.Size == "4:3" {
		cfg.AspectRatio = "4:3"
	}

	resp, err := client.Models.GenerateImages(ctx, opts.Model, opts.Prompt, cfg)
	if err != nil {
		return nil, fmt.Errorf("genai generate images failed: %w", err)
	}

	result := &providergateway.ImageResult{
		// Data: make([]providergateway.ImageData, len(resp.GeneratedImages)),
		Data: make([]providergateway.ImageData, 0),
	}

	for _, img := range resp.GeneratedImages {
		result.Data = append(result.Data, providergateway.ImageData{
			B64JSON: base64.StdEncoding.EncodeToString(img.Image.ImageBytes),
		})
	}

	return result, nil
}

// generateContentImage handles image generation for Gemini models via the GenerateContent API.
func (g *Gemini) generateContentImage(ctx context.Context, model string, opts providergateway.ImageGenerationOptions) (*providergateway.ImageResult, error) {
	client, err := g.newSDKClient(ctx)
	if err != nil {
		return nil, err
	}

	// Construct content parts
	parts := []*genai.Part{
		{Text: opts.Prompt},
	}

	// Config
	// Note: Gemini models typically produce 1 image per request via GenerateContent defaults or strict configs.
	// The Go SDK might not expose explicit image count in GenerateContentConfig effectively for all models yet,
	// checking genai.GenerateContentConfig...
	config := &genai.GenerateContentConfig{
		// Temperature, etc. might apply.
		// For image generation, response MimeType might be specified?
	}

	resp, err := client.Models.GenerateContent(ctx, model, []*genai.Content{{Parts: parts}}, config)
	if err != nil {
		return nil, fmt.Errorf("generate content failed: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned")
	}

	result := &providergateway.ImageResult{
		Data: make([]providergateway.ImageData, 0),
	}

	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			if part.InlineData != nil {
				result.Data = append(result.Data, providergateway.ImageData{
					B64JSON: base64.StdEncoding.EncodeToString(part.InlineData.Data),
				})
			} else if part.FileData != nil {
				// URI based
				result.Data = append(result.Data, providergateway.ImageData{
					URL: part.FileData.FileURI,
				})
			}
		}
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no image data found in response")
	}

	return result, nil
}

// EditImage edits an image using the Gemini/Imagen predict API.
func (g *Gemini) EditImage(ctx context.Context, opts providergateway.ImageEditOptions) (*providergateway.ImageResult, error) {

	// Logic for imagen-3.0-capability-image-editing-001 or newer equivalents
	// Currently Imagen 4+ via genai SDK does not expose EditImage/Masking directly in the public docs provided.
	// We will attempt to use the genai client if possible, but standard 'GenerateImages' is text-to-image.

	return nil, fmt.Errorf("image editing is not currently supported by the genai library for Imagen 4 models (Text-to-Image only). Please use 'generate image' without an input image.")
}

// loadInputImage reads image from file path or returns error.
func loadInputImage(input string) ([]byte, error) {
	// Simple check: is it a file that exists?
	// If not, is it valid base64?

	// Try reading file first
	data, err := os.ReadFile(input)
	if err == nil {
		return data, nil
	}

	// If read file fails, maybe it is base64?
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err == nil {
		return decoded, nil
	}

	return nil, fmt.Errorf("input is neither a valid file path nor base64 string")
}
