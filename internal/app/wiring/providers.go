// build provider instances from application configuration.
// internal/app/wiring/providers.go
package wiring

import (
	"fmt"

	"github.com/MadeByDoug/wls-chatbot/internal/app/config"
	provideradapter "github.com/MadeByDoug/wls-chatbot/internal/adapters/provider"
	providerusecase "github.com/MadeByDoug/wls-chatbot/internal/core/usecase/provider"
	"github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

// LoadProvidersFromStore loads providers from configuration storage.
func LoadProvidersFromStore(store config.Store, secrets ports.SecretStore) ([]ports.Provider, error) {
	cfg, err := config.LoadConfig(store)
	if err != nil {
		return nil, err
	}
	return ProvidersFromConfig(cfg, secrets)
}

// ProvidersFromConfig constructs providers from configuration.
func ProvidersFromConfig(cfg config.AppConfig, secrets ports.SecretStore) ([]ports.Provider, error) {
	providers := make([]ports.Provider, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		apiKey := ""
		if secrets != nil {
			if key, err := secrets.GetProviderKey(p.Name); err == nil {
				apiKey = key
			}
		}
		enabledModels := ResolveEnabledModelsFromConfig(cfg, p.Name, p.DefaultModel)
		providerConfig := provideradapter.Config{
			Name:         p.Name,
			DisplayName:  p.DisplayName,
			APIKey:       apiKey,
			BaseURL:      p.BaseURL,
			DefaultModel: p.DefaultModel,
			Models:       enabledModels,
		}
		switch p.Type {
		case "openai":
			providers = append(providers, provideradapter.NewOpenAI(providerConfig))
		case "gemini":
			providers = append(providers, provideradapter.NewGemini(providerConfig))
		default:
			return nil, fmt.Errorf("unknown provider type: %s", p.Type)
		}
	}
	return providers, nil
}

// BuildProviderService wires provider adapters into the provider use case.
func BuildProviderService(cfg config.AppConfig, cache ports.ProviderCache, secrets ports.SecretStore, logger ports.Logger) (*providerusecase.Service, ports.ProviderRegistry, error) {
	registry := provideradapter.NewRegistry()
	providerConfigs, providerErr := ProvidersFromConfig(cfg, secrets)
	if providerErr == nil {
		for _, p := range providerConfigs {
			registry.Register(p)
		}
	}

	updateFrequency, frequencyErr := config.ResolveUpdateFrequencies(cfg)
	service := providerusecase.NewService(registry, cache, secrets, updateFrequency, logger)
	if providerErr != nil {
		return service, registry, providerErr
	}
	if frequencyErr != nil {
		return service, registry, frequencyErr
	}
	return service, registry, nil
}
