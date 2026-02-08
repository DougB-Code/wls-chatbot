// capabilities.go implements capability-specific test functions.
package modeltest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Capability identifies a testable capability.
type Capability string

const (
	CapabilityChat           Capability = "chat"
	CapabilityImageGen       Capability = "image_gen"
	CapabilityImageEdit      Capability = "image_edit"
	CapabilityTestConnection Capability = "test_connection"
)

// AllCapabilities returns all available test capabilities.
func AllCapabilities() []Capability {
	return []Capability{
		CapabilityChat,
		CapabilityImageGen,
		CapabilityImageEdit,
		CapabilityTestConnection,
	}
}

// TestResult contains the outcome of a capability test.
type TestResult struct {
	Provider   string      `json:"provider"`
	Capability Capability  `json:"capability"`
	Model      string      `json:"model"`
	Success    bool        `json:"success"`
	Error      string      `json:"error,omitempty"`
	Recording  *Recording  `json:"recording,omitempty"`
}

// CapabilityTester tests a specific capability.
type CapabilityTester interface {
	Test(ctx context.Context, client *Client, model string) TestResult
	Capability() Capability
}

// ChatTester tests chat completion capability.
type ChatTester struct{}

func (t *ChatTester) Capability() Capability { return CapabilityChat }

func (t *ChatTester) Test(ctx context.Context, client *Client, model string) TestResult {
	result := TestResult{
		Provider:   client.Config().ProviderName,
		Capability: CapabilityChat,
		Model:      model,
	}

	req := t.buildRequest(client.Config().ProviderType, model)
	path := t.endpoint(client.Config().ProviderType, model)

	resp, err := client.Do(ctx, http.MethodPost, path, req)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		return result
	}

	result.Success = true
	return result
}

func (t *ChatTester) endpoint(providerType ProviderType, model string) string {
	switch providerType {
	case ProviderTypeOpenAI:
		return "/v1/chat/completions"
	case ProviderTypeGemini:
		return fmt.Sprintf("/v1beta/models/%s:generateContent", model)
	case ProviderTypeAnthropic:
		return "/v1/messages"
	default:
		return "/v1/chat/completions"
	}
}

func (t *ChatTester) buildRequest(providerType ProviderType, model string) interface{} {
	switch providerType {
	case ProviderTypeGemini:
		return map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]string{
						{"text": "Say 'test successful' and nothing else."},
					},
				},
			},
		}
	case ProviderTypeAnthropic:
		return map[string]interface{}{
			"model":      model,
			"max_tokens": 10,
			"messages": []map[string]string{
				{"role": "user", "content": "Say 'test successful' and nothing else."},
			},
		}
	default: // OpenAI-compatible
		return ChatRequest{
			Model: model,
			Messages: []ChatMessage{
				{Role: "user", Content: "Say 'test successful' and nothing else."},
			},
			MaxTokens: 10,
		}
	}
}

// ImageGenTester tests image generation capability.
type ImageGenTester struct{}

func (t *ImageGenTester) Capability() Capability { return CapabilityImageGen }

func (t *ImageGenTester) Test(ctx context.Context, client *Client, model string) TestResult {
	result := TestResult{
		Provider:   client.Config().ProviderName,
		Capability: CapabilityImageGen,
		Model:      model,
	}

	req := t.buildRequest(client.Config().ProviderType, model)
	path := t.endpoint(client.Config().ProviderType, model)

	resp, err := client.Do(ctx, http.MethodPost, path, req)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		return result
	}

	result.Success = true
	return result
}

func (t *ImageGenTester) endpoint(providerType ProviderType, model string) string {
	switch providerType {
	case ProviderTypeGemini:
		return fmt.Sprintf("/v1beta/models/%s:predict", model)
	default:
		return "/v1/images/generations"
	}
}

func (t *ImageGenTester) buildRequest(providerType ProviderType, model string) interface{} {
	switch providerType {
	case ProviderTypeGemini:
		return map[string]interface{}{
			"instances": []map[string]string{
				{"prompt": "A simple red square on white background"},
			},
			"parameters": map[string]interface{}{
				"sampleCount": 1,
			},
		}
	default:
		return ImageGenRequest{
			Model:          model,
			Prompt:         "A simple red square on white background",
			N:              1,
			Size:           "256x256",
			ResponseFormat: "b64_json",
		}
	}
}

// TestConnectionTester tests provider connectivity.
type TestConnectionTester struct{}

func (t *TestConnectionTester) Capability() Capability { return CapabilityTestConnection }

func (t *TestConnectionTester) Test(ctx context.Context, client *Client, model string) TestResult {
	result := TestResult{
		Provider:   client.Config().ProviderName,
		Capability: CapabilityTestConnection,
		Model:      model,
	}

	path := t.endpoint(client.Config().ProviderType)
	resp, err := client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))
		return result
	}

	result.Success = true
	return result
}

func (t *TestConnectionTester) endpoint(providerType ProviderType) string {
	switch providerType {
	case ProviderTypeGemini:
		return "/v1beta/models"
	case ProviderTypeAnthropic:
		return "/v1/messages" // Anthropic doesn't have a list models endpoint
	default:
		return "/v1/models"
	}
}

// GetTester returns the appropriate tester for a capability.
func GetTester(cap Capability) CapabilityTester {
	switch cap {
	case CapabilityChat:
		return &ChatTester{}
	case CapabilityImageGen:
		return &ImageGenTester{}
	case CapabilityTestConnection:
		return &TestConnectionTester{}
	default:
		return nil
	}
}

// ParseResponse attempts to parse a response body as JSON.
func ParseResponse(body []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
