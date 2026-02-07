// defaults_test.go verifies default configuration values.
// internal/features/settings/config/defaults_test.go
package config

import "testing"

// TestDefaultConfigProvidesProviders verifies the default providers are populated.
func TestDefaultConfigProvidesProviders(t *testing.T) {

	cfg := DefaultConfig()

	if len(cfg.Providers) != 7 {
		t.Fatalf("expected 7 providers, got %d", len(cfg.Providers))
	}

	openai := cfg.Providers[0]
	if openai.Type != "openai" || openai.Name != "openai" {
		t.Fatalf("unexpected openai provider: %+v", openai)
	}
	if openai.BaseURL == "" {
		t.Fatalf("expected openai baseUrl to be set")
	}

	grok := cfg.Providers[1]
	if grok.Type != "openai" || grok.Name != "grok" {
		t.Fatalf("unexpected grok provider: %+v", grok)
	}
	if grok.BaseURL == "" {
		t.Fatalf("expected grok baseUrl to be set")
	}

	mistral := cfg.Providers[2]
	if mistral.Type != "openai" || mistral.Name != "mistral" {
		t.Fatalf("unexpected mistral provider: %+v", mistral)
	}
	if mistral.BaseURL == "" {
		t.Fatalf("expected mistral baseUrl to be set")
	}

	anthropic := cfg.Providers[3]
	if anthropic.Type != "anthropic" || anthropic.Name != "anthropic" {
		t.Fatalf("unexpected anthropic provider: %+v", anthropic)
	}
	if anthropic.BaseURL == "" {
		t.Fatalf("expected anthropic baseUrl to be set")
	}

	gemini := cfg.Providers[4]
	if gemini.Type != "gemini" || gemini.Name != "gemini" {
		t.Fatalf("unexpected gemini provider: %+v", gemini)
	}
	if gemini.BaseURL == "" {
		t.Fatalf("expected gemini baseUrl to be set")
	}

	cloudflare := cfg.Providers[5]
	if cloudflare.Type != "cloudflare" || cloudflare.Name != "cloudflare" {
		t.Fatalf("unexpected cloudflare provider: %+v", cloudflare)
	}

	openrouter := cfg.Providers[6]
	if openrouter.Type != "openrouter" || openrouter.Name != "openrouter" {
		t.Fatalf("unexpected openrouter provider: %+v", openrouter)
	}
	if openrouter.BaseURL == "" {
		t.Fatalf("expected openrouter baseUrl to be set")
	}
}
