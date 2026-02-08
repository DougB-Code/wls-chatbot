// openai_compat.go handles OpenAI-compatible HTTP flows shared by provider adapters.
// internal/features/providers/adapters/httpcompat/openai_compat.go
package providerhttp

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

	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/core"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/gateway"
)

// ListOpenAICompatModels fetches models from an OpenAI-compatible API.
func ListOpenAICompatModels(ctx context.Context, client Client, baseURL string, headers map[string]string) ([]providercore.Model, error) {

	baseURL = normalizeCompatBaseURL(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("base URL required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	setHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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

	models := make([]providercore.Model, 0, len(payload.Data))
	for _, item := range payload.Data {
		if item.ID == "" {
			continue
		}
		models = append(models, providercore.Model{
			ID:   item.ID,
			Name: item.ID,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	return models, nil
}

// ChatOpenAICompat executes OpenAI-compatible chat completion requests.
func ChatOpenAICompat(ctx context.Context, client Client, baseURL string, headers map[string]string, messages []providergateway.ProviderMessage, opts providergateway.ChatOptions) (<-chan providergateway.Chunk, error) {

	baseURL = normalizeCompatBaseURL(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("base URL required")
	}

	body, err := MarshalOpenAICompatBody(opts.Model, messages, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	setHeaders(req, headers)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, &APIError{Code: resp.StatusCode, Message: string(body)}
	}

	chunks := make(chan providergateway.Chunk, 100)

	go func() {
		defer close(chunks)
		defer func() { _ = resp.Body.Close() }()

		if opts.Stream {
			streamOpenAICompatResponse(resp.Body, chunks)
		} else {
			parseOpenAICompatResponse(resp.Body, chunks)
		}
	}()

	return chunks, nil
}

// MarshalOpenAICompatBody builds and marshals an OpenAI-compatible request body.
func MarshalOpenAICompatBody(model string, messages []providergateway.ProviderMessage, opts providergateway.ChatOptions) ([]byte, error) {

	reqBody := buildOpenAICompatBody(model, messages, opts)
	return json.Marshal(reqBody)
}

// buildOpenAICompatBody constructs an OpenAI-compatible request payload.
func buildOpenAICompatBody(model string, messages []providergateway.ProviderMessage, opts providergateway.ChatOptions) map[string]interface{} {

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
		"model":    model,
		"messages": apiMessages,
		"stream":   opts.Stream,
	}

	if opts.Temperature > 0 {
		reqBody["temperature"] = opts.Temperature
	}
	if opts.MaxTokens > 0 {
		reqBody["max_tokens"] = opts.MaxTokens
	}

	return reqBody
}

// normalizeCompatBaseURL trims trailing slashes from OpenAI-compatible base URLs.
func normalizeCompatBaseURL(baseURL string) string {

	trimmed := strings.TrimSpace(baseURL)
	return strings.TrimRight(trimmed, "/")
}

// setHeaders adds configured headers when provided.
func setHeaders(req *http.Request, headers map[string]string) {

	if req == nil || len(headers) == 0 {
		return
	}
	for name, value := range headers {
		if strings.TrimSpace(name) == "" || strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(name, value)
	}
}

// streamOpenAICompatResponse parses SSE responses into chunks.
func streamOpenAICompatResponse(body io.Reader, chunks chan<- providergateway.Chunk) {

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			chunks <- providergateway.Chunk{FinishReason: "stop"}
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
			Usage *providergateway.UsageStats `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			continue
		}

		if len(resp.Choices) > 0 {
			choice := resp.Choices[0]
			chunk := providergateway.Chunk{
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
		chunks <- providergateway.Chunk{Error: err}
	}
}

// parseOpenAICompatResponse parses a non-streaming response into a chunk.
func parseOpenAICompatResponse(body io.Reader, chunks chan<- providergateway.Chunk) {

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage *providergateway.UsageStats `json:"usage"`
	}

	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		chunks <- providergateway.Chunk{Error: err}
		return
	}

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		chunks <- providergateway.Chunk{
			Content:      choice.Message.Content,
			FinishReason: choice.FinishReason,
			Usage:        resp.Usage,
		}
	}
}
