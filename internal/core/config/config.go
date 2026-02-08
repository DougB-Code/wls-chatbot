// config.go defines application configuration entities and helpers.
// internal/core/config/config.go
package config

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	modelcatalog "github.com/MadeByDoug/wls-chatbot/pkg/models"
)

// AppConfig represents the root application configuration.
type AppConfig struct {
	Providers []ProviderConfig `json:"providers"`
}

// UpdateFrequency describes how often provider resources are refreshed.
type UpdateFrequency string

const (
	UpdateFrequencyManual UpdateFrequency = "manual"
	UpdateFrequencyHourly UpdateFrequency = "hourly"
	UpdateFrequencyDaily  UpdateFrequency = "daily"
	UpdateFrequencyWeekly UpdateFrequency = "weekly"
)

// ProviderConfig describes a configured provider.
type ProviderConfig struct {
	Type            string            `json:"type"`
	Name            string            `json:"name"`
	DisplayName     string            `json:"displayName"`
	BaseURL         string            `json:"baseUrl"`
	DefaultModel    string            `json:"defaultModel"`
	UpdateFrequency UpdateFrequency   `json:"updateFrequency"`
	Models          []ModelConfig     `json:"models"`
	Inputs          map[string]string `json:"inputs,omitempty"`
}

// ModelConfig describes a provider model toggle in configuration.
type ModelConfig struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
}

// LoadConfig loads configuration from the provided store.
func LoadConfig(store Store) (AppConfig, error) {

	if store == nil {
		return AppConfig{}, fmt.Errorf("load config: store required")
	}

	cfg, err := store.Load()
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, ErrConfigNotFound) {
		cfg, cfgErr := defaultConfigFromCatalog()
		if cfgErr != nil {
			return AppConfig{}, fmt.Errorf("load config: default from catalog: %w", cfgErr)
		}
		if saveErr := store.Save(cfg); saveErr != nil {
			return AppConfig{}, fmt.Errorf("load config: save default: %w", saveErr)
		}
		return cfg, nil
	}

	return AppConfig{}, fmt.Errorf("load config: %w", err)
}

// ResolveUpdateFrequencies parses provider update frequencies from configuration.
func ResolveUpdateFrequencies(cfg AppConfig) (map[string]time.Duration, error) {

	frequencies := make(map[string]time.Duration)
	for _, providerConfig := range cfg.Providers {
		if providerConfig.UpdateFrequency == "" {
			continue
		}
		frequency, err := parseUpdateFrequency(providerConfig.UpdateFrequency)
		if err != nil {
			return nil, fmt.Errorf("provider %s updateFrequency: %w", providerConfig.Name, err)
		}
		if frequency <= 0 {
			continue
		}
		frequencies[providerConfig.Name] = frequency
	}

	if len(frequencies) == 0 {
		return nil, nil
	}

	return frequencies, nil
}

// parseUpdateFrequency converts enum values into durations.
func parseUpdateFrequency(value UpdateFrequency) (time.Duration, error) {

	switch value {
	case UpdateFrequencyManual:
		return 0, nil
	case UpdateFrequencyHourly:
		return time.Hour, nil
	case UpdateFrequencyDaily:
		return 24 * time.Hour, nil
	case UpdateFrequencyWeekly:
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown update frequency: %s", value)
	}
}

// defaultConfigFromCatalog builds default provider/model config from the canonical pkg/models catalog.
func defaultConfigFromCatalog() (AppConfig, error) {

	catalog, err := modelcatalog.LoadEmbedded()
	if err != nil {
		return AppConfig{}, err
	}

	providerByCatalogID := make(map[string]modelcatalog.Provider, len(catalog.Providers))
	for _, provider := range catalog.Providers {
		providerByCatalogID[strings.TrimSpace(provider.ID)] = provider
	}

	familyProviderByID := make(map[string]string, len(catalog.Families))
	for _, family := range catalog.Families {
		catalogProviderID := strings.TrimSpace(family.Provider)
		provider, ok := providerByCatalogID[catalogProviderID]
		if !ok {
			return AppConfig{}, fmt.Errorf("unsupported catalog provider: %s", family.Provider)
		}
		familyProviderByID[strings.TrimSpace(family.ID)] = strings.TrimSpace(provider.ID)
	}

	modelsByProvider := make(map[string][]ModelConfig)
	modelSeenByProvider := make(map[string]map[string]struct{})
	defaultModelByProvider := make(map[string]string)
	for _, model := range catalog.Models {
		providerName, ok := familyProviderByID[strings.TrimSpace(model.Family)]
		if !ok {
			return AppConfig{}, fmt.Errorf("catalog model %s references unknown family %s", model.ID, model.Family)
		}

		modelID := strings.TrimSpace(model.ID)
		if modelID == "" {
			continue
		}
		if _, seen := modelSeenByProvider[providerName]; !seen {
			modelSeenByProvider[providerName] = make(map[string]struct{})
		}
		if _, seen := modelSeenByProvider[providerName][modelID]; seen {
			continue
		}
		modelSeenByProvider[providerName][modelID] = struct{}{}
		modelsByProvider[providerName] = append(modelsByProvider[providerName], ModelConfig{ID: modelID, Enabled: true})
		if defaultModelByProvider[providerName] == "" {
			defaultModelByProvider[providerName] = modelID
		}
	}

	providerNames := make([]string, 0, len(modelsByProvider))
	for providerName := range modelsByProvider {
		providerNames = append(providerNames, providerName)
	}
	sort.Strings(providerNames)

	providers := make([]ProviderConfig, 0, len(providerNames))
	for _, providerID := range providerNames {
		provider := providerByCatalogID[providerID]
		name := strings.TrimSpace(provider.Name)
		if name == "" {
			return AppConfig{}, fmt.Errorf("catalog provider %s missing name", providerID)
		}
		providerType := strings.TrimSpace(provider.Type)
		if providerType == "" {
			return AppConfig{}, fmt.Errorf("catalog provider %s missing type", providerID)
		}
		displayName := strings.TrimSpace(provider.DisplayName)
		if displayName == "" {
			return AppConfig{}, fmt.Errorf("catalog provider %s missing display_name", providerID)
		}
		baseURL := strings.TrimSpace(provider.BaseURL)

		providers = append(providers, ProviderConfig{
			Type:            providerType,
			Name:            name,
			DisplayName:     displayName,
			BaseURL:         baseURL,
			DefaultModel:    defaultModelByProvider[providerID],
			UpdateFrequency: UpdateFrequencyManual,
			Models:          modelsByProvider[providerID],
			Inputs:          map[string]string{},
		})
	}

	return AppConfig{Providers: providers}, nil
}
