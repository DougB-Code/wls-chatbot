// defaults_test.go verifies default configuration values.
// internal/app/config/defaults_test.go
package config

import "testing"

// TestDefaultConfigProvidesProviders verifies the default providers are populated.
func TestDefaultConfigProvidesProviders(t *testing.T) {

	cfg := DefaultConfig()

	if len(cfg.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(cfg.Providers))
	}

	openai := cfg.Providers[0]
	if openai.Type != "openai" || openai.Name != "openai" {
		t.Fatalf("unexpected openai provider: %+v", openai)
	}
	if openai.BaseURL == "" {
		t.Fatalf("expected openai baseUrl to be set")
	}

	gemini := cfg.Providers[1]
	if gemini.Type != "gemini" || gemini.Name != "gemini" {
		t.Fatalf("unexpected gemini provider: %+v", gemini)
	}
	if gemini.BaseURL == "" {
		t.Fatalf("expected gemini baseUrl to be set")
	}
}
