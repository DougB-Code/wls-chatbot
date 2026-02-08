// Package modeltest provides infrastructure for testing LLM provider capabilities
// and capturing request/response pairs as golden files for mock-based regression testing.
//
// This package is self-contained with no dependencies on internal packages,
// allowing it to be used independently or distributed as a standalone binary.
package modeltest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ProviderType identifies the API format used by a provider.
type ProviderType string

const (
	ProviderTypeOpenAI    ProviderType = "openai"
	ProviderTypeGemini    ProviderType = "gemini"
	ProviderTypeAnthropic ProviderType = "anthropic"
)

// ClientConfig configures a provider HTTP client.
type ClientConfig struct {
	ProviderName string
	ProviderType ProviderType
	BaseURL      string
	APIKey       string
	Headers      map[string]string
}

// Client is a provider-agnostic HTTP client for LLM APIs.
type Client struct {
	config    ClientConfig
	transport http.RoundTripper
}

// NewClient creates a new provider client.
func NewClient(config ClientConfig) *Client {
	return &Client{
		config:    config,
		transport: http.DefaultTransport,
	}
}

// SetTransport overrides the HTTP transport (for recording/mocking).
func (c *Client) SetTransport(t http.RoundTripper) {
	c.transport = t
}

// Transport returns the current transport.
func (c *Client) Transport() http.RoundTripper {
	return c.transport
}

// Config returns the client configuration.
func (c *Client) Config() ClientConfig {
	return c.config
}

// Do executes an HTTP request with provider-specific authentication.
func (c *Client) Do(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := strings.TrimSuffix(c.config.BaseURL, "/") + "/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set content type for JSON requests
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Apply provider-specific authentication
	c.applyAuth(req)

	// Apply custom headers
	for k, v := range c.config.Headers {
		req.Header.Set(k, v)
	}

	httpClient := &http.Client{Transport: c.transport}
	return httpClient.Do(req)
}

// applyAuth adds provider-specific authentication headers.
func (c *Client) applyAuth(req *http.Request) {
	switch c.config.ProviderType {
	case ProviderTypeOpenAI:
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	case ProviderTypeGemini:
		// Gemini uses query parameter for API key
		q := req.URL.Query()
		q.Set("key", c.config.APIKey)
		req.URL.RawQuery = q.Encode()
	case ProviderTypeAnthropic:
		req.Header.Set("x-api-key", c.config.APIKey)
		req.Header.Set("anthropic-version", "2023-06-01")
	}
}

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatMessage represents a single message in a chat.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ImageGenRequest represents an image generation request.
type ImageGenRequest struct {
	Model          string `json:"model,omitempty"`
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	Quality        string `json:"quality,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
}
