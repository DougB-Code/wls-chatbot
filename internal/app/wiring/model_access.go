// reconcile enabled model selections from configuration.
// internal/app/wiring/model_access.go
package wiring

import (
	"strings"

	"github.com/MadeByDoug/wls-chatbot/internal/app/config"
	"github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

// ResolveEnabledModelsFromConfig returns enabled models from configuration only.
func ResolveEnabledModelsFromConfig(cfg config.AppConfig, providerName string, defaultModel string) []ports.Model {
	providerConfig := findProviderConfig(&cfg, providerName)
	enabledIDs := []string{}
	if providerConfig != nil {
		enabledIDs = normalizeModelIDs(enabledModelIDs(providerConfig.Models))
	}
	trimmedDefault := strings.TrimSpace(defaultModel)
	if trimmedDefault != "" && !containsString(enabledIDs, trimmedDefault) {
		enabledIDs = append(enabledIDs, trimmedDefault)
	}
	return buildFallbackModels(enabledIDs)
}

// findProviderConfig returns the provider config entry for a name.
func findProviderConfig(cfg *config.AppConfig, providerName string) *config.ProviderConfig {
	for i := range cfg.Providers {
		if cfg.Providers[i].Name == providerName {
			return &cfg.Providers[i]
		}
	}
	return nil
}

// enabledModelIDs extracts enabled model IDs from config.
func enabledModelIDs(models []config.ModelConfig) []string {
	enabled := make([]string, 0, len(models))
	for _, model := range models {
		if !model.Enabled {
			continue
		}
		enabled = append(enabled, model.ID)
	}
	return enabled
}

// normalizeModelIDs trims and deduplicates model IDs.
func normalizeModelIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	normalized := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

// buildFallbackModels constructs model structs from IDs.
func buildFallbackModels(ids []string) []ports.Model {
	models := make([]ports.Model, 0, len(ids))
	for _, id := range ids {
		models = append(models, ports.Model{
			ID:   id,
			Name: id,
		})
	}
	return models
}

// containsString checks whether a slice contains a value.
func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
