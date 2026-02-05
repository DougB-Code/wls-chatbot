// build provider instances from application configuration.
// internal/features/settings/wiring/providers.go
package wiring

import (
	"fmt"
	"strings"

	coreports "github.com/MadeByDoug/wls-chatbot/internal/core/ports"
	provideradapter "github.com/MadeByDoug/wls-chatbot/internal/features/settings/adapters/provider"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/config"
	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/ports"
	providerusecase "github.com/MadeByDoug/wls-chatbot/internal/features/settings/usecase"
)

// LoadProvidersFromStore loads providers from configuration storage.
func LoadProvidersFromStore(store config.Store, secrets ports.SecretStore) ([]ports.Provider, error) {
	cfg, err := config.LoadConfig(store)
	if err != nil {
		return nil, err
	}
	return ProvidersFromConfig(cfg, secrets, nil)
}

// ProvidersFromConfig constructs providers from configuration.
func ProvidersFromConfig(cfg config.AppConfig, secrets ports.SecretStore, logger coreports.Logger) ([]ports.Provider, error) {
	providers := make([]ports.Provider, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		credentials := buildProviderCredentials(p, secrets)
		apiKey := strings.TrimSpace(credentials[provideradapter.CredentialAPIKey])
		enabledModels := ResolveEnabledModelsFromConfig(cfg, p.Name, p.DefaultModel)
		providerConfig := provideradapter.Config{
			Name:         p.Name,
			DisplayName:  p.DisplayName,
			APIKey:       apiKey,
			BaseURL:      p.BaseURL,
			DefaultModel: p.DefaultModel,
			Models:       enabledModels,
			Credentials:  credentials,
			Logger:       logger,
		}
		switch p.Type {
		case "openai":
			providers = append(providers, provideradapter.NewOpenAI(providerConfig))
		case "anthropic":
			providers = append(providers, provideradapter.NewAnthropic(providerConfig))
		case "gemini":
			providers = append(providers, provideradapter.NewGemini(providerConfig))
		case "cloudflare":
			providers = append(providers, provideradapter.NewCloudflare(providerConfig))
		default:
			return nil, fmt.Errorf("unknown provider type: %s", p.Type)
		}
	}
	return providers, nil
}

// buildProviderCredentials merges config inputs with stored secrets.
func buildProviderCredentials(cfg config.ProviderConfig, secrets ports.SecretStore) provideradapter.ProviderCredentials {

	credentials := make(provideradapter.ProviderCredentials)
	for key, value := range cfg.Inputs {
		if strings.TrimSpace(value) == "" {
			continue
		}
		credentials[key] = value
	}

	secretFields := providerSecretFields(cfg.Type)
	if secrets != nil && len(secretFields) > 0 {
		for _, field := range secretFields {
			if value, err := secrets.GetProviderSecret(cfg.Name, field); err == nil && strings.TrimSpace(value) != "" {
				credentials[field] = value
			}
		}
	}

	if len(credentials) == 0 {
		return nil
	}
	return credentials
}

// providerSecretFields returns secret credential field names for a provider type.
func providerSecretFields(providerType string) []string {

	switch providerType {
	case "openai", "anthropic", "gemini":
		return []string{provideradapter.CredentialAPIKey}
	case "cloudflare":
		return []string{
			provideradapter.CredentialCloudflareToken,
			provideradapter.CredentialAPIKey,
			provideradapter.CredentialToken,
		}
	default:
		return nil
	}
}

// BuildProviderService wires provider adapters into the provider use case.
func BuildProviderService(cfg config.AppConfig, cache ports.ProviderCache, secrets ports.SecretStore, inputs ports.ProviderInputsStore, logger coreports.Logger) (*providerusecase.Service, ports.ProviderRegistry, error) {
	registry := provideradapter.NewRegistry()
	providerConfigs, providerErr := ProvidersFromConfig(cfg, secrets, logger)
	if providerErr == nil {
		for _, p := range providerConfigs {
			registry.Register(p)
		}
	}

	updateFrequency, frequencyErr := config.ResolveUpdateFrequencies(cfg)
	service := providerusecase.NewService(registry, cache, secrets, inputs, updateFrequency, logger)
	if providerErr != nil {
		return service, registry, providerErr
	}
	if frequencyErr != nil {
		return service, registry, frequencyErr
	}
	return service, registry, nil
}
