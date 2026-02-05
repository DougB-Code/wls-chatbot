// handle OpenAI-compatible HTTP flows shared by providers.
// internal/features/settings/adapters/provider/openai_compat.go
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

// listOpenAICompatModels fetches models from an OpenAI-compatible API.
func listOpenAICompatModels(ctx context.Context, client HTTPClient, baseURL string, headers map[string]string) ([]Model, error) {

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

// chatOpenAICompat executes OpenAI-compatible chat completion requests.
func chatOpenAICompat(ctx context.Context, client HTTPClient, baseURL string, headers map[string]string, messages []ProviderMessage, opts ChatOptions) (<-chan Chunk, error) {

	baseURL = normalizeCompatBaseURL(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("base URL required")
	}

	body, err := marshalOpenAICompatBody(opts.Model, messages, opts)
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
		resp.Body.Close()
		return nil, &APIError{Code: resp.StatusCode, Message: string(body)}
	}

	chunks := make(chan Chunk, 100)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		if opts.Stream {
			streamOpenAICompatResponse(resp.Body, chunks)
		} else {
			parseOpenAICompatResponse(resp.Body, chunks)
		}
	}()

	return chunks, nil
}

// marshalOpenAICompatBody builds and marshals a compat request body.
func marshalOpenAICompatBody(model string, messages []ProviderMessage, opts ChatOptions) ([]byte, error) {

	reqBody := buildOpenAICompatBody(model, messages, opts)
	return json.Marshal(reqBody)
}

// buildOpenAICompatBody constructs an OpenAI-compatible request payload.
func buildOpenAICompatBody(model string, messages []ProviderMessage, opts ChatOptions) map[string]interface{} {

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
func streamOpenAICompatResponse(body io.Reader, chunks chan<- Chunk) {

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

// parseOpenAICompatResponse parses a non-streaming response into a chunk.
func parseOpenAICompatResponse(body io.Reader, chunks chan<- Chunk) {

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
