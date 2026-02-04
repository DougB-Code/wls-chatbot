// config.go defines application configuration types and helpers.
// internal/app/config/config.go
package config

import (
	"errors"
	"fmt"
	"time"
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
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	DisplayName     string          `json:"displayName"`
	BaseURL         string          `json:"baseUrl"`
	DefaultModel    string          `json:"defaultModel"`
	UpdateFrequency UpdateFrequency `json:"updateFrequency"`
	Models          []ModelConfig   `json:"models"`
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
		cfg = DefaultConfig()
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
