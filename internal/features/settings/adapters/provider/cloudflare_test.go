// cloudflare_test.go verifies Cloudflare AI Gateway adapter behavior.
// internal/features/settings/adapters/provider/cloudflare_test.go
package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCloudflareChatAddsAuthHeaders verifies auth headers are sent when configured.
func TestCloudflareChatAddsAuthHeaders(t *testing.T) {

	var gotGatewayAuth string
	var gotUpstreamAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotGatewayAuth = r.Header.Get("cf-aig-authorization")
		gotUpstreamAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	provider := NewCloudflare(Config{
		Name:        "cloudflare",
		DisplayName: "Cloudflare",
		BaseURL:     server.URL,
		Credentials: ProviderCredentials{
			CredentialToken:  "gateway-token",
			CredentialAPIKey: "upstream-key",
		},
	})
	provider.SetHTTPClient(server.Client())

	chunks, err := provider.Chat(context.Background(), []ProviderMessage{
		{Role: RoleUser, Content: "Hello"},
	}, ChatOptions{Model: "openai/gpt-5", Stream: false})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}

	var content string
	for chunk := range chunks {
		if chunk.Error != nil {
			t.Fatalf("chunk error: %v", chunk.Error)
		}
		if chunk.Content != "" {
			content = chunk.Content
		}
	}

	if gotGatewayAuth != "Bearer gateway-token" {
		t.Fatalf("expected gateway auth header, got %q", gotGatewayAuth)
	}
	if gotUpstreamAuth != "Bearer upstream-key" {
		t.Fatalf("expected upstream auth header, got %q", gotUpstreamAuth)
	}
	if content != "ok" {
		t.Fatalf("expected content, got %q", content)
	}
}

// TestCloudflareChatWorkersAIModelPrefixes verifies Workers AI models are normalized.
func TestCloudflareChatWorkersAIModelPrefixes(t *testing.T) {

	var gotAuth string
	var gotModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		var payload struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		gotModel = payload.Model
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}]}`))
	}))
	defer server.Close()

	provider := NewCloudflare(Config{
		Name:        "cloudflare",
		DisplayName: "Cloudflare",
		BaseURL:     server.URL,
		Credentials: ProviderCredentials{
			CredentialCloudflareToken: "cf-token",
		},
	})
	provider.SetHTTPClient(server.Client())

	chunks, err := provider.Chat(context.Background(), []ProviderMessage{
		{Role: RoleUser, Content: "Hello"},
	}, ChatOptions{Model: "@cf/llama-3-8b", Stream: false})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}

	for chunk := range chunks {
		if chunk.Error != nil {
			t.Fatalf("chunk error: %v", chunk.Error)
		}
	}

	if gotAuth != "Bearer cf-token" {
		t.Fatalf("expected cloudflare auth header, got %q", gotAuth)
	}
	if gotModel != "workers-ai/@cf/llama-3-8b" {
		t.Fatalf("expected workers ai model prefix, got %q", gotModel)
	}
}

// TestCloudflareListResourcesOmitsAuthHeader verifies no auth header is sent without a token.
func TestCloudflareListResourcesOmitsAuthHeader(t *testing.T) {

	var gotGatewayAuth string
	var gotUpstreamAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotGatewayAuth = r.Header.Get("cf-aig-authorization")
		gotUpstreamAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"openai/gpt-5.2"}]}`))
	}))
	defer server.Close()

	provider := NewCloudflare(Config{
		Name:        "cloudflare",
		DisplayName: "Cloudflare",
		BaseURL:     server.URL,
	})
	provider.SetHTTPClient(server.Client())

	models, err := provider.ListResources(context.Background())
	if err != nil {
		t.Fatalf("list resources: %v", err)
	}

	if gotGatewayAuth != "" {
		t.Fatalf("expected no gateway auth header, got %q", gotGatewayAuth)
	}
	if gotUpstreamAuth != "" {
		t.Fatalf("expected no upstream auth header, got %q", gotUpstreamAuth)
	}
	if len(models) != 1 || models[0].ID != "openai/gpt-5.2" {
		t.Fatalf("unexpected models: %+v", models)
	}
}
