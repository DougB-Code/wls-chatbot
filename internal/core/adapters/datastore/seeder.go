// seeder.go imports model catalog seed data into the datastore.
// internal/core/adapters/datastore/seeder.go
package datastore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ModelData represents the root structure of models.yaml
type ModelData struct {
	Families []Family `yaml:"families"`
	Models   []Model  `yaml:"models"`
}

// Family represents a model family with shared capabilities
type Family struct {
	ID           string       `yaml:"id"`
	Provider     string       `yaml:"provider"`
	Modalities   Modalities   `yaml:"modalities"`
	Capabilities Capabilities `yaml:"capabilities"`
	SystemTags   []string     `yaml:"system_tags"`
}

// Model represents a specific model snapshot or alias
type Model struct {
	ID         string `yaml:"id"`
	Family     string `yaml:"family"`
	IsSnapshot bool   `yaml:"is_snapshot"`
	LinkAlias  string `yaml:"link_alias,omitempty"`
}

// Modalities defines input and output modalities
type Modalities struct {
	Input  []string `yaml:"input"`
	Output []string `yaml:"output"`
}

// Capabilities defines model capabilities
type Capabilities struct {
	Streaming        bool `yaml:"streaming"`
	ToolCalling      bool `yaml:"tool_calling"`
	StructuredOutput bool `yaml:"structured_output"`
	Vision           bool `yaml:"vision"`
}

// seededProvider describes a provider row used by model seeding.
type seededProvider struct {
	ID          string
	Name        string
	DisplayName string
	AdapterType string
	BaseURL     string
}

// SeedModels populates the database with model data from YAML
func SeedModels(db *sql.DB, data []byte) error {
	var catalog ModelData
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("unmarshal models.yaml: %w", err)
	}

	families := make(map[string]Family)
	for _, f := range catalog.Families {
		families[f.ID] = f
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Rollback is a no-op if tx.Commit() was called
	}()

	now := time.Now().UnixMilli()

	// Prepare statements
	stmtEndpoint, err := tx.Prepare(`
		INSERT INTO catalog_endpoints (
			id, provider_id, display_name, adapter_type, base_url, route_kind, 
			origin_provider, origin_route_label, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider_id = excluded.provider_id,
			display_name = excluded.display_name,
			adapter_type = excluded.adapter_type,
			base_url = excluded.base_url,
			route_kind = excluded.route_kind,
			origin_provider = excluded.origin_provider,
			origin_route_label = excluded.origin_route_label,
			updated_at = excluded.updated_at
	`)
	if err != nil {
		return err
	}
	defer func() { _ = stmtEndpoint.Close() }()

	stmtEntry, err := tx.Prepare(`
		INSERT INTO model_catalog_entries (
			id, endpoint_id, model_id, display_name, first_seen_at, last_seen_at, 
			availability_state, approved, missed_refreshes, source, metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(endpoint_id, model_id) DO UPDATE SET last_seen_at = ?
	`)
	if err != nil {
		return err
	}
	defer func() { _ = stmtEntry.Close() }()

	stmtCapabilities, err := tx.Prepare(`
		INSERT OR REPLACE INTO model_capabilities (
			model_catalog_entry_id, supports_streaming, supports_tool_calling, 
			supports_structured_output, supports_vision, capabilities_source, capabilities_as_of
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer func() { _ = stmtCapabilities.Close() }()

	stmtSysTags, err := tx.Prepare(`INSERT OR IGNORE INTO model_system_tags (model_catalog_entry_id, tag) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = stmtSysTags.Close() }()

	stmtInputMod, err := tx.Prepare(`INSERT OR IGNORE INTO model_capabilities_input_modalities (model_catalog_entry_id, modality) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = stmtInputMod.Close() }()

	stmtOutputMod, err := tx.Prepare(`INSERT OR IGNORE INTO model_capabilities_output_modalities (model_catalog_entry_id, modality) VALUES (?, ?)`)
	if err != nil {
		return err
	}
	defer func() { _ = stmtOutputMod.Close() }()

	for _, model := range catalog.Models {
		fam, ok := families[model.Family]
		if !ok {
			return fmt.Errorf("unknown family %q for model %q", model.Family, model.ID)
		}
		provider, err := resolveSeededProvider(tx, fam.Provider)
		if err != nil {
			return fmt.Errorf("resolve provider for family %q: %w", model.Family, err)
		}

		// 1. Ensure Endpoint Exists (provider_id is constructed from provider name)
		endpointID := "ep_" + provider.Name + "_default"

		if _, err := stmtEndpoint.Exec(
			endpointID, provider.ID, provider.DisplayName+" Default", provider.AdapterType, provider.BaseURL, "llm",
			provider.Name, "default", now, now,
		); err != nil {
			return fmt.Errorf("insert endpoint: %w", err)
		}

		// 2. Insert Model Entry
		entryID := "entry_" + provider.Name + "_" + model.ID
		metadata := map[string]interface{}{
			"family":      model.Family,
			"is_snapshot": model.IsSnapshot,
		}
		if model.LinkAlias != "" {
			metadata["link_alias"] = model.LinkAlias
		}
		metaBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata for model %q: %w", model.ID, err)
		}

		if _, err := stmtEntry.Exec(
			entryID, endpointID, model.ID, model.ID, now, now,
			"available", 1, 0, "seed", string(metaBytes), now,
		); err != nil {
			return fmt.Errorf("insert entry %s: %w", model.ID, err)
		}

		// 3. Insert Capabilities
		if _, err := stmtCapabilities.Exec(
			entryID, fam.Capabilities.Streaming, fam.Capabilities.ToolCalling,
			fam.Capabilities.StructuredOutput, fam.Capabilities.Vision, "seed_v1", now,
		); err != nil {
			return fmt.Errorf("insert capabilities %s: %w", model.ID, err)
		}

		// 4. Insert Tags & Modalities
		for _, tag := range fam.SystemTags {
			if _, err := stmtSysTags.Exec(entryID, tag); err != nil {
				return err
			}
		}
		for _, mod := range fam.Modalities.Input {
			if _, err := stmtInputMod.Exec(entryID, mod); err != nil {
				return err
			}
		}
		for _, mod := range fam.Modalities.Output {
			if _, err := stmtOutputMod.Exec(entryID, mod); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// resolveSeededProvider maps YAML provider identifiers to catalog providers.
func resolveSeededProvider(tx *sql.Tx, providerName string) (seededProvider, error) {

	name := strings.TrimSpace(strings.ToLower(providerName))
	switch name {
	case "google":
		name = "gemini"
	case "xai":
		name = "grok"
	}

	var provider seededProvider
	if err := tx.QueryRow(
		`SELECT id, name, display_name, adapter_type, base_url
		 FROM catalog_providers
		 WHERE name = ?`,
		name,
	).Scan(
		&provider.ID,
		&provider.Name,
		&provider.DisplayName,
		&provider.AdapterType,
		&provider.BaseURL,
	); err != nil {
		if err == sql.ErrNoRows {
			return seededProvider{}, fmt.Errorf("provider %q not found in catalog_providers", name)
		}
		return seededProvider{}, err
	}

	if provider.BaseURL == "" {
		return seededProvider{}, fmt.Errorf("provider %q has empty base_url", provider.Name)
	}

	return provider, nil
}
