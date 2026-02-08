// catalog.go derives model test plans from the canonical model catalog.
// pkg/models/modeltest/catalog.go
package modeltest

import (
	"fmt"
	"strings"

	modelcatalog "github.com/MadeByDoug/wls-chatbot/pkg/models"
)

// LoadEmbeddedCatalogPlan loads the canonical model catalog and derives a test plan.
func LoadEmbeddedCatalogPlan() (*TestPlan, error) {

	catalog, err := modelcatalog.LoadEmbedded()
	if err != nil {
		return nil, err
	}
	return BuildPlanFromCatalog(catalog)
}

// BuildPlanFromCatalog derives a test plan from canonical catalog data.
func BuildPlanFromCatalog(catalog *modelcatalog.Catalog) (*TestPlan, error) {

	if catalog == nil {
		return nil, fmt.Errorf("modeltest catalog plan: catalog required")
	}

	familyByID := make(map[string]modelcatalog.Family, len(catalog.Families))
	for _, family := range catalog.Families {
		familyByID[family.ID] = family
	}

	providerByName := make(map[string]*ProviderConfig)
	providerOrder := make([]string, 0)

	for _, model := range catalog.Models {
		family, exists := familyByID[model.Family]
		if !exists {
			return nil, fmt.Errorf("modeltest catalog plan: unknown family %q for model %q", model.Family, model.ID)
		}

		providerName := normalizeProviderName(family.Provider)
		providerType, baseURL, err := providerConfigForName(providerName)
		if err != nil {
			return nil, fmt.Errorf("modeltest catalog plan: model %q: %w", model.ID, err)
		}

		provider := providerByName[providerName]
		if provider == nil {
			provider = &ProviderConfig{
				Name:    providerName,
				Type:    providerType,
				BaseURL: baseURL,
			}
			providerByName[providerName] = provider
			providerOrder = append(providerOrder, providerName)
		}

		provider.Models = append(provider.Models, ModelConfig{
			ID:           model.ID,
			Capabilities: deriveModelCapabilities(family),
		})
	}

	plan := &TestPlan{Providers: make([]ProviderConfig, 0, len(providerOrder))}
	for _, providerName := range providerOrder {
		plan.Providers = append(plan.Providers, *providerByName[providerName])
	}
	return plan, nil
}

// normalizeProviderName maps catalog provider aliases into modeltest provider names.
func normalizeProviderName(value string) string {

	name := strings.TrimSpace(strings.ToLower(value))
	switch name {
	case "google":
		return "gemini"
	case "xai":
		return "grok"
	default:
		return name
	}
}

// providerConfigForName resolves modeltest provider config fields from provider name.
func providerConfigForName(name string) (ProviderType, string, error) {

	switch name {
	case "openai":
		return ProviderTypeOpenAI, "https://api.openai.com", nil
	case "grok":
		return ProviderTypeOpenAI, "https://api.x.ai", nil
	case "openrouter":
		return ProviderTypeOpenAI, "https://openrouter.ai/api", nil
	case "gemini":
		return ProviderTypeGemini, "https://generativelanguage.googleapis.com", nil
	case "anthropic":
		return ProviderTypeAnthropic, "https://api.anthropic.com", nil
	default:
		return "", "", fmt.Errorf("unsupported provider %q", name)
	}
}

// deriveModelCapabilities derives modeltest capabilities from canonical family metadata.
func deriveModelCapabilities(family modelcatalog.Family) []Capability {

	capabilities := make([]Capability, 0, 4)

	if hasValue(family.Modalities.Output, "text") {
		capabilities = append(capabilities, CapabilityChat)
	}
	if hasValue(family.Modalities.Output, "image") || hasValue(family.SystemTags, "image_gen") {
		capabilities = append(capabilities, CapabilityImageGen)
	}
	if hasValue(family.SystemTags, "image_edit") || hasValue(family.SystemTags, "image_editing") {
		capabilities = append(capabilities, CapabilityImageEdit)
	}

	capabilities = append(capabilities, CapabilityTestConnection)
	return dedupeCapabilities(capabilities)
}

// hasValue reports whether values contain target after case-insensitive normalization.
func hasValue(values []string, target string) bool {

	normalizedTarget := strings.TrimSpace(strings.ToLower(target))
	for _, value := range values {
		if strings.TrimSpace(strings.ToLower(value)) == normalizedTarget {
			return true
		}
	}
	return false
}

// dedupeCapabilities removes duplicate capability values while preserving order.
func dedupeCapabilities(values []Capability) []Capability {

	seen := make(map[Capability]struct{}, len(values))
	deduped := make([]Capability, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		deduped = append(deduped, value)
	}
	return deduped
}
