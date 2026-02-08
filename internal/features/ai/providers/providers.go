// providers.go builds settings provider dependencies from configuration.
// internal/features/settings/module/providers.go
package providers

import (
	"fmt"
	"strings"

	config "github.com/MadeByDoug/wls-chatbot/internal/core/config"
	corelogger "github.com/MadeByDoug/wls-chatbot/internal/core/logger"
	modelaccess "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model"
	anthropicadapter "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/anthropic"
	cloudflareadapter "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/cloudflare"
	geminiadapter "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/gemini"
	grokadapter "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/grok"
	openaiadapter "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/openai"
	openrouteradapter "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/openrouter"
	providerregistry "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/adapters/registry"
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/interfaces/core"
	providerusecase "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/provider"
)

// LoadProvidersFromStore loads providers from configuration storage.
func LoadProvidersFromStore(store config.Store, secrets providercore.SecretStore) ([]providercore.Provider, error) {

	cfg, err := config.LoadConfig(store)
	if err != nil {
		return nil, err
	}
	return ProvidersFromConfig(cfg, secrets, nil)
}

// ProvidersFromConfig constructs providers from configuration.
func ProvidersFromConfig(cfg config.AppConfig, secrets providercore.SecretStore, logger corelogger.Logger) ([]providercore.Provider, error) {

	providers := make([]providercore.Provider, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		credentials := buildProviderCredentials(p, secrets)
		apiKey := strings.TrimSpace(credentials[providercore.CredentialAPIKey])
		enabledModels := modelaccess.ResolveEnabledModelsFromConfig(cfg, p.Name, p.DefaultModel)
		providerConfig := providercore.ProviderConfig{
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
			providers = append(providers, openaiadapter.New(providerConfig))
		case "anthropic":
			providers = append(providers, anthropicadapter.New(providerConfig))
		case "gemini":
			providers = append(providers, geminiadapter.New(providerConfig))
		case "grok":
			providers = append(providers, grokadapter.New(providerConfig))
		case "cloudflare":
			providers = append(providers, cloudflareadapter.New(providerConfig))
		case "openrouter":
			providers = append(providers, openrouteradapter.New(providerConfig))
		default:
			return nil, fmt.Errorf("unknown provider type: %s", p.Type)
		}
	}
	return providers, nil
}

// buildProviderCredentials merges config inputs with stored secrets.
func buildProviderCredentials(cfg config.ProviderConfig, secrets providercore.SecretStore) providercore.ProviderCredentials {

	credentials := make(providercore.ProviderCredentials)
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
	case "openai", "anthropic", "gemini", "grok":
		return []string{providercore.CredentialAPIKey}
	case "openrouter":
		return []string{providercore.CredentialAPIKey}
	case "cloudflare":
		return []string{
			providercore.CredentialCloudflareToken,
			providercore.CredentialAPIKey,
			providercore.CredentialToken,
		}
	default:
		return nil
	}
}

// BuildProviderService wires provider adapters into the provider use case.
func BuildProviderService(cfg config.AppConfig, cache providercore.ProviderCache, secrets providercore.SecretStore, inputs providercore.ProviderInputsStore, logger corelogger.Logger) (*providerusecase.Service, providercore.ProviderRegistry, error) {

	registry := providerregistry.New()
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
