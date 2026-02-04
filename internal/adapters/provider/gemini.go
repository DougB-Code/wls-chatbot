// Implement the provider interface for Gemini.
// internal/adapters/provider/gemini.go
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
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

// NewGemini creates a new Gemini provider.
func NewGemini(config Config) *Gemini {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta" // @TODO this can't be hard coded
	}

	return &Gemini{
		name:        config.Name,
		displayName: config.DisplayName,
		baseURL:     baseURL,
		apiKey:      config.APIKey,
		models:      config.Models,
		client:      defaultHTTPClient(),
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

// Configure updates the provider configuration.
func (g *Gemini) Configure(config Config) error {
	g.apiKey = config.APIKey
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
		g.client = defaultHTTPClient()
	}
	return g.client
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{Code: resp.StatusCode, Message: string(body)}
	}

	return nil
}

// ListResources fetches the available models from the Gemini API.
func (g *Gemini) ListResources(ctx context.Context) ([]Model, error) {
	url := fmt.Sprintf("%s/models?key=%s", g.baseURL, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &APIError{Code: resp.StatusCode, Message: string(body)}
	}

	var payload struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	models := make([]Model, 0, len(payload.Models))
	for _, item := range payload.Models {
		if item.Name == "" {
			continue
		}
		id := strings.TrimPrefix(item.Name, "models/")
		if id == "" {
			id = item.Name
		}
		name := item.DisplayName
		if name == "" {
			name = id
		}
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

	// Convert messages to Gemini format
	contents := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		content := msg.Content
		if strings.TrimSpace(content) == "" {
			continue
		}

		role := "user"
		if msg.Role == RoleAssistant {
			role = "model"
		}

		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{
				{"text": content},
			},
		})
	}

	reqBody := map[string]interface{}{
		"contents": contents,
	}

	if opts.Temperature > 0 || opts.MaxTokens > 0 {
		genConfig := map[string]interface{}{}
		if opts.Temperature > 0 {
			genConfig["temperature"] = opts.Temperature
		}
		if opts.MaxTokens > 0 {
			genConfig["maxOutputTokens"] = opts.MaxTokens
		}
		reqBody["generationConfig"] = genConfig
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := "generateContent"
	if opts.Stream {
		endpoint = "streamGenerateContent"
	}

	url := fmt.Sprintf("%s/models/%s:%s?key=%s", g.baseURL, opts.Model, endpoint, g.apiKey)
	if opts.Stream {
		url += "&alt=sse"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient().Do(req)
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
			g.streamResponse(resp.Body, chunks)
		} else {
			g.parseResponse(resp.Body, chunks)
		}
	}()

	return chunks, nil
}

// streamResponse parses Gemini SSE responses into chunks.
func (g *Gemini) streamResponse(body io.Reader, chunks chan<- Chunk) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var resp struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
				FinishReason string `json:"finishReason"`
			} `json:"candidates"`
			UsageMetadata *struct {
				PromptTokenCount     int `json:"promptTokenCount"`
				CandidatesTokenCount int `json:"candidatesTokenCount"`
				TotalTokenCount      int `json:"totalTokenCount"`
			} `json:"usageMetadata"`
		}

		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			continue
		}

		if len(resp.Candidates) > 0 {
			candidate := resp.Candidates[0]
			text := ""
			if len(candidate.Content.Parts) > 0 {
				text = candidate.Content.Parts[0].Text
			}

			chunk := Chunk{
				Content:      text,
				FinishReason: strings.ToLower(candidate.FinishReason),
			}

			if resp.UsageMetadata != nil {
				chunk.Usage = &UsageStats{
					PromptTokens:     resp.UsageMetadata.PromptTokenCount,
					CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
					TotalTokens:      resp.UsageMetadata.TotalTokenCount,
				}
			}

			chunks <- chunk
		}
	}
	if err := scanner.Err(); err != nil {
		chunks <- Chunk{Error: err}
	}
}

// parseResponse parses a non-streaming Gemini response into a chunk.
func (g *Gemini) parseResponse(body io.Reader, chunks chan<- Chunk) {
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata *struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		chunks <- Chunk{Error: err}
		return
	}

	if len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		text := ""
		if len(candidate.Content.Parts) > 0 {
			text = candidate.Content.Parts[0].Text
		}

		chunk := Chunk{
			Content:      text,
			FinishReason: strings.ToLower(candidate.FinishReason),
		}

		if resp.UsageMetadata != nil {
			chunk.Usage = &UsageStats{
				PromptTokens:     resp.UsageMetadata.PromptTokenCount,
				CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
				TotalTokens:      resp.UsageMetadata.TotalTokenCount,
			}
		}

		chunks <- chunk
	}
}

// Ensure Gemini implements Provider.
var _ Provider = (*Gemini)(nil)
