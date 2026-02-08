// catalog.go loads and validates the canonical model catalog.
// pkg/models/catalog.go
package models

import (
	_ "embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed models.yaml
var embeddedCatalogYAML []byte

// Catalog represents the root structure of the canonical models catalog.
type Catalog struct {
	Providers []Provider `yaml:"providers"`
	Families  []Family   `yaml:"families"`
	Models    []Model    `yaml:"models"`
}

// Provider represents provider metadata used to bootstrap runtime configuration.
type Provider struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	DisplayName string `yaml:"display_name"`
	BaseURL     string `yaml:"base_url"`
}

// Family represents a model family with shared provider and capability metadata.
type Family struct {
	ID           string       `yaml:"id"`
	Provider     string       `yaml:"provider"`
	Modalities   Modalities   `yaml:"modalities"`
	Capabilities Capabilities `yaml:"capabilities"`
	SystemTags   []string     `yaml:"system_tags"`
}

// Model represents a concrete model entry bound to a family.
type Model struct {
	ID         string `yaml:"id"`
	Family     string `yaml:"family"`
	IsSnapshot bool   `yaml:"is_snapshot"`
	LinkAlias  string `yaml:"link_alias,omitempty"`
}

// Modalities describes model input and output modality support.
type Modalities struct {
	Input  []string `yaml:"input"`
	Output []string `yaml:"output"`
}

// Capabilities describes capability flags shared by a model family.
type Capabilities struct {
	Streaming        bool `yaml:"streaming"`
	ToolCalling      bool `yaml:"tool_calling"`
	StructuredOutput bool `yaml:"structured_output"`
	Vision           bool `yaml:"vision"`
}

// EmbeddedYAML returns a copy of the embedded canonical catalog bytes.
func EmbeddedYAML() []byte {

	copyBytes := make([]byte, len(embeddedCatalogYAML))
	copy(copyBytes, embeddedCatalogYAML)
	return copyBytes
}

// LoadEmbedded parses and validates the embedded canonical catalog.
func LoadEmbedded() (*Catalog, error) {

	return Parse(embeddedCatalogYAML)
}

// Parse decodes and validates canonical catalog content.
func Parse(data []byte) (*Catalog, error) {

	var catalog Catalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("unmarshal models catalog: %w", err)
	}
	if err := validateCatalog(&catalog); err != nil {
		return nil, err
	}
	return &catalog, nil
}

// validateCatalog enforces structural consistency rules for canonical catalog data.
func validateCatalog(catalog *Catalog) error {

	if catalog == nil {
		return fmt.Errorf("models catalog: catalog required")
	}

	providerByID := make(map[string]Provider, len(catalog.Providers))
	providerNameSeen := make(map[string]struct{}, len(catalog.Providers))
	for _, provider := range catalog.Providers {
		providerID := strings.TrimSpace(provider.ID)
		if providerID == "" {
			return fmt.Errorf("models catalog: provider id required")
		}
		if _, exists := providerByID[providerID]; exists {
			return fmt.Errorf("models catalog: duplicate provider id %q", providerID)
		}
		if strings.TrimSpace(provider.Name) == "" {
			return fmt.Errorf("models catalog: provider %q missing name", providerID)
		}
		if strings.TrimSpace(provider.Type) == "" {
			return fmt.Errorf("models catalog: provider %q missing type", providerID)
		}
		if strings.TrimSpace(provider.DisplayName) == "" {
			return fmt.Errorf("models catalog: provider %q missing display_name", providerID)
		}

		providerName := strings.TrimSpace(provider.Name)
		if _, exists := providerNameSeen[providerName]; exists {
			return fmt.Errorf("models catalog: duplicate provider name %q", providerName)
		}
		providerNameSeen[providerName] = struct{}{}
		providerByID[providerID] = provider
	}

	familyByID := make(map[string]struct{}, len(catalog.Families))
	for _, family := range catalog.Families {
		familyID := strings.TrimSpace(family.ID)
		if familyID == "" {
			return fmt.Errorf("models catalog: family id required")
		}
		if _, exists := familyByID[familyID]; exists {
			return fmt.Errorf("models catalog: duplicate family id %q", familyID)
		}
		providerID := strings.TrimSpace(family.Provider)
		if providerID == "" {
			return fmt.Errorf("models catalog: family %q missing provider", familyID)
		}
		if _, exists := providerByID[providerID]; !exists {
			return fmt.Errorf("models catalog: family %q references unknown provider %q", familyID, providerID)
		}
		familyByID[familyID] = struct{}{}
	}

	modelByID := make(map[string]struct{}, len(catalog.Models))
	for _, model := range catalog.Models {
		modelID := strings.TrimSpace(model.ID)
		if modelID == "" {
			return fmt.Errorf("models catalog: model id required")
		}
		if _, exists := modelByID[modelID]; exists {
			return fmt.Errorf("models catalog: duplicate model id %q", modelID)
		}
		modelByID[modelID] = struct{}{}

		familyID := strings.TrimSpace(model.Family)
		if familyID == "" {
			return fmt.Errorf("models catalog: model %q missing family", modelID)
		}
		if _, exists := familyByID[familyID]; !exists {
			return fmt.Errorf("models catalog: model %q references unknown family %q", modelID, familyID)
		}
	}

	return nil
}
