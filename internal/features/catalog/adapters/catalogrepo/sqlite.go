// sqlite.go persists model catalog, endpoints, roles, and provider metadata in SQLite.
// internal/features/catalog/adapters/catalogrepo/sqlite.go
package catalogrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/providers/interfaces/core"
	"github.com/google/uuid"
)

const catalogSchema = `
CREATE TABLE IF NOT EXISTS catalog_providers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    adapter_type TEXT NOT NULL,
    trust_mode TEXT NOT NULL,
    base_url TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_test_at INTEGER,
    last_test_ok INTEGER CHECK (last_test_ok IN (0, 1)),
    last_error TEXT,
    last_discovery_at INTEGER
);

CREATE TABLE IF NOT EXISTS provider_inputs (
    provider_id TEXT NOT NULL,
    input_key TEXT NOT NULL,
    input_value TEXT NOT NULL,
    PRIMARY KEY (provider_id, input_key),
    FOREIGN KEY (provider_id) REFERENCES catalog_providers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS catalog_endpoints (
    id TEXT PRIMARY KEY,
    provider_id TEXT NOT NULL,
    display_name TEXT NOT NULL,
    adapter_type TEXT NOT NULL,
    base_url TEXT NOT NULL,
    route_kind TEXT NOT NULL,
    origin_provider TEXT NOT NULL,
    origin_route_label TEXT NOT NULL,
    auth_json TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    last_test_at INTEGER,
    last_test_ok INTEGER CHECK (last_test_ok IN (0, 1)),
    last_error TEXT,
    FOREIGN KEY (provider_id) REFERENCES catalog_providers(id) ON DELETE CASCADE,
    UNIQUE (provider_id, route_kind, origin_provider, origin_route_label, base_url)
);

CREATE TABLE IF NOT EXISTS model_catalog_entries (
    id TEXT PRIMARY KEY,
    endpoint_id TEXT NOT NULL,
    model_id TEXT NOT NULL,
    display_name TEXT,
    first_seen_at INTEGER NOT NULL,
    last_seen_at INTEGER NOT NULL,
    availability_state TEXT NOT NULL,
    approved INTEGER NOT NULL CHECK (approved IN (0, 1)),
    missed_refreshes INTEGER NOT NULL,
    source TEXT NOT NULL DEFAULT 'discovered',
    metadata_json TEXT,
    FOREIGN KEY (endpoint_id) REFERENCES catalog_endpoints(id) ON DELETE CASCADE,
    UNIQUE (endpoint_id, model_id)
);

CREATE TABLE IF NOT EXISTS model_capabilities (
    model_catalog_entry_id TEXT PRIMARY KEY,
    supports_streaming INTEGER NOT NULL CHECK (supports_streaming IN (0, 1)),
    supports_tool_calling INTEGER NOT NULL CHECK (supports_tool_calling IN (0, 1)),
    supports_structured_output INTEGER NOT NULL CHECK (supports_structured_output IN (0, 1)),
    supports_vision INTEGER NOT NULL CHECK (supports_vision IN (0, 1)),
    capabilities_source TEXT NOT NULL,
    capabilities_as_of INTEGER NOT NULL,
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS model_capabilities_input_modalities (
    model_catalog_entry_id TEXT NOT NULL,
    modality TEXT NOT NULL,
    PRIMARY KEY (model_catalog_entry_id, modality),
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS model_capabilities_output_modalities (
    model_catalog_entry_id TEXT NOT NULL,
    modality TEXT NOT NULL,
    PRIMARY KEY (model_catalog_entry_id, modality),
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS model_system_profile (
    model_catalog_entry_id TEXT PRIMARY KEY,
    latency_tier TEXT NOT NULL,
    cost_tier TEXT NOT NULL,
    reliability_tier TEXT NOT NULL,
    system_profile_source TEXT NOT NULL,
    system_profile_as_of INTEGER NOT NULL,
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS model_system_tags (
    model_catalog_entry_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (model_catalog_entry_id, tag),
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS model_user_addenda (
    model_catalog_entry_id TEXT PRIMARY KEY,
    notes TEXT,
    user_addenda_source TEXT NOT NULL,
    user_addenda_as_of INTEGER NOT NULL,
    latency_tier_override TEXT,
    cost_tier_override TEXT,
    reliability_tier_override TEXT,
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS model_user_tags (
    model_catalog_entry_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (model_catalog_entry_id, tag),
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    requires_streaming INTEGER NOT NULL CHECK (requires_streaming IN (0, 1)),
    requires_tool_calling INTEGER NOT NULL CHECK (requires_tool_calling IN (0, 1)),
    requires_structured_output INTEGER NOT NULL CHECK (requires_structured_output IN (0, 1)),
    requires_vision INTEGER NOT NULL CHECK (requires_vision IN (0, 1)),
    max_cost_tier TEXT,
    max_latency_tier TEXT,
    min_reliability_tier TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS role_required_input_modalities (
    role_id TEXT NOT NULL,
    modality TEXT NOT NULL,
    PRIMARY KEY (role_id, modality),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS role_required_output_modalities (
    role_id TEXT NOT NULL,
    modality TEXT NOT NULL,
    PRIMARY KEY (role_id, modality),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS role_assignments (
    role_id TEXT NOT NULL,
    model_catalog_entry_id TEXT NOT NULL,
    assigned_by TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    enabled INTEGER NOT NULL CHECK (enabled IN (0, 1)),
    PRIMARY KEY (role_id, model_catalog_entry_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (model_catalog_entry_id) REFERENCES model_catalog_entries(id) ON DELETE CASCADE
);
`

// Repository stores model catalog and role data in SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a SQLite-backed model catalog repository.
func NewRepository(db *sql.DB) (*Repository, error) {

	if db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	if _, err := db.Exec(catalogSchema); err != nil {
		return nil, fmt.Errorf("catalog repo: ensure schema: %w", err)
	}
	if err := validateSchema(db); err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

// validateSchema ensures required columns exist for strict-schema operation.
func validateSchema(db *sql.DB) error {

	requiredColumns := []struct {
		table  string
		column string
	}{
		{table: "model_capabilities_input_modalities", column: "model_catalog_entry_id"},
		{table: "model_capabilities_input_modalities", column: "modality"},
		{table: "model_capabilities_output_modalities", column: "model_catalog_entry_id"},
		{table: "model_capabilities_output_modalities", column: "modality"},
	}

	for _, required := range requiredColumns {
		exists, err := schemaHasColumn(db, required.table, required.column)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf(
				"catalog repo: incompatible schema (%s.%s missing); remove workspace database and restart",
				required.table,
				required.column,
			)
		}
	}
	return nil
}

// schemaHasColumn reports whether a table contains a specific column.
func schemaHasColumn(db *sql.DB, table, column string) (bool, error) {

	query := fmt.Sprintf("PRAGMA table_info(%s)", table)
	rows, err := db.Query(query)
	if err != nil {
		return false, fmt.Errorf("catalog repo: table info %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, dataType string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, fmt.Errorf("catalog repo: table info scan %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("catalog repo: table info rows %s: %w", table, err)
	}
	return false, nil
}

// ProviderRecord describes a stored provider entity.
type ProviderRecord struct {
	ID              string
	Name            string
	DisplayName     string
	AdapterType     string
	TrustMode       string
	BaseURL         string
	LastTestAt      int64
	LastTestOK      bool
	LastError       string
	LastDiscoveryAt int64
}

// EndpointRecord describes a stored endpoint.
type EndpointRecord struct {
	ID               string
	ProviderID       string
	ProviderName     string
	DisplayName      string
	AdapterType      string
	BaseURL          string
	RouteKind        string
	OriginProvider   string
	OriginRouteLabel string
	AuthJSON         string
	LastTestAt       int64
	LastTestOK       bool
	LastError        string
}

// ModelEntryRecord describes a stored model catalog entry.
type ModelEntryRecord struct {
	ID                string
	EndpointID        string
	ModelID           string
	DisplayName       string
	FirstSeenAt       int64
	LastSeenAt        int64
	AvailabilityState string
	Approved          bool
	MissedRefreshes   int
	Source            string // "seed", "user", or "discovered"
	MetadataJSON      string
}

// RoleRecord describes a stored role.
type RoleRecord struct {
	ID                       string
	Name                     string
	RequiresStreaming        bool
	RequiresToolCalling      bool
	RequiresStructuredOutput bool
	RequiresVision           bool
	MaxCostTier              string
	MaxLatencyTier           string
	MinReliabilityTier       string
	RequiredInputModalities  []string
	RequiredOutputModalities []string
}

// RoleAssignmentRecord describes a role assignment.
type RoleAssignmentRecord struct {
	RoleID              string
	ModelCatalogEntryID string
	AssignedBy          string
	CreatedAt           int64
	Enabled             bool
}

// EnsureProvider upserts a provider and returns its record.
func (r *Repository) EnsureProvider(ctx context.Context, provider ProviderRecord) (ProviderRecord, error) {

	if r == nil || r.db == nil {
		return ProviderRecord{}, fmt.Errorf("catalog repo: db required")
	}

	now := time.Now().UnixMilli()
	provider.Name = strings.TrimSpace(provider.Name)
	if provider.Name == "" {
		return ProviderRecord{}, fmt.Errorf("catalog repo: provider name required")
	}

	if provider.ID == "" {
		provider.ID = newUUID()
	}

	if provider.TrustMode == "" {
		provider.TrustMode = "user_managed"
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO catalog_providers (id, name, display_name, adapter_type, trust_mode, base_url, created_at, updated_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)
         ON CONFLICT(name) DO UPDATE SET
           display_name = excluded.display_name,
           adapter_type = excluded.adapter_type,
           trust_mode = excluded.trust_mode,
           base_url = excluded.base_url,
           updated_at = excluded.updated_at`,
		provider.ID,
		provider.Name,
		provider.DisplayName,
		provider.AdapterType,
		provider.TrustMode,
		provider.BaseURL,
		now,
		now,
	)
	if err != nil {
		return ProviderRecord{}, fmt.Errorf("catalog repo: ensure provider: %w", err)
	}

	row := r.db.QueryRowContext(ctx, `SELECT id, name, display_name, adapter_type, trust_mode, base_url, last_test_at, last_test_ok, last_error, last_discovery_at FROM catalog_providers WHERE name = ?`, provider.Name)
	var record ProviderRecord
	var lastTestAt sql.NullInt64
	var lastTestOK sql.NullInt64
	var lastError sql.NullString
	var lastDiscoveryAt sql.NullInt64
	if err := row.Scan(
		&record.ID,
		&record.Name,
		&record.DisplayName,
		&record.AdapterType,
		&record.TrustMode,
		&record.BaseURL,
		&lastTestAt,
		&lastTestOK,
		&lastError,
		&lastDiscoveryAt,
	); err != nil {
		return ProviderRecord{}, fmt.Errorf("catalog repo: ensure provider load: %w", err)
	}
	if lastTestAt.Valid {
		record.LastTestAt = lastTestAt.Int64
	}
	record.LastTestOK = lastTestOK.Valid && lastTestOK.Int64 == 1
	if lastError.Valid {
		record.LastError = lastError.String
	}
	if lastDiscoveryAt.Valid {
		record.LastDiscoveryAt = lastDiscoveryAt.Int64
	}
	return record, nil
}

// GetProviderByName loads a provider by name.
func (r *Repository) GetProviderByName(ctx context.Context, name string) (ProviderRecord, error) {

	if r == nil || r.db == nil {
		return ProviderRecord{}, fmt.Errorf("catalog repo: db required")
	}

	row := r.db.QueryRowContext(ctx, `SELECT id, name, display_name, adapter_type, trust_mode, base_url, last_test_at, last_test_ok, last_error, last_discovery_at FROM catalog_providers WHERE name = ?`, name)
	var record ProviderRecord
	var lastTestAt sql.NullInt64
	var lastTestOK sql.NullInt64
	var lastError sql.NullString
	var lastDiscoveryAt sql.NullInt64
	err := row.Scan(
		&record.ID,
		&record.Name,
		&record.DisplayName,
		&record.AdapterType,
		&record.TrustMode,
		&record.BaseURL,
		&lastTestAt,
		&lastTestOK,
		&lastError,
		&lastDiscoveryAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ProviderRecord{}, nil
	}
	if err != nil {
		return ProviderRecord{}, fmt.Errorf("catalog repo: get provider: %w", err)
	}
	if lastTestAt.Valid {
		record.LastTestAt = lastTestAt.Int64
	}
	record.LastTestOK = lastTestOK.Valid && lastTestOK.Int64 == 1
	if lastError.Valid {
		record.LastError = lastError.String
	}
	if lastDiscoveryAt.Valid {
		record.LastDiscoveryAt = lastDiscoveryAt.Int64
	}
	return record, nil
}

// ListProviders returns all providers in the catalog.
func (r *Repository) ListProviders(ctx context.Context) ([]ProviderRecord, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, name, display_name, adapter_type, trust_mode, base_url, last_test_at, last_test_ok, last_error, last_discovery_at FROM catalog_providers ORDER BY display_name`)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list providers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var providers []ProviderRecord
	for rows.Next() {
		var record ProviderRecord
		var lastTestAt sql.NullInt64
		var lastTestOK sql.NullInt64
		var lastError sql.NullString
		var lastDiscoveryAt sql.NullInt64
		if err := rows.Scan(
			&record.ID,
			&record.Name,
			&record.DisplayName,
			&record.AdapterType,
			&record.TrustMode,
			&record.BaseURL,
			&lastTestAt,
			&lastTestOK,
			&lastError,
			&lastDiscoveryAt,
		); err != nil {
			return nil, fmt.Errorf("catalog repo: list providers scan: %w", err)
		}
		if lastTestAt.Valid {
			record.LastTestAt = lastTestAt.Int64
		}
		record.LastTestOK = lastTestOK.Valid && lastTestOK.Int64 == 1
		if lastError.Valid {
			record.LastError = lastError.String
		}
		if lastDiscoveryAt.Valid {
			record.LastDiscoveryAt = lastDiscoveryAt.Int64
		}
		providers = append(providers, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: list providers rows: %w", err)
	}
	return providers, nil
}

// UpdateProviderStatus records provider health or discovery metadata.
func (r *Repository) UpdateProviderStatus(ctx context.Context, providerID string, lastTestAt int64, lastTestOK bool, lastError string, lastDiscoveryAt int64) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if providerID == "" {
		return fmt.Errorf("catalog repo: provider id required")
	}

	_, err := r.db.ExecContext(
		ctx,
		`UPDATE catalog_providers
         SET last_test_at = ?, last_test_ok = ?, last_error = ?, last_discovery_at = ?, updated_at = ?
         WHERE id = ?`,
		nullIfZero(lastTestAt),
		boolToInt(lastTestOK),
		strings.TrimSpace(lastError),
		nullIfZero(lastDiscoveryAt),
		time.Now().UnixMilli(),
		providerID,
	)
	if err != nil {
		return fmt.Errorf("catalog repo: update provider status: %w", err)
	}
	return nil
}

// SaveProviderInputs stores non-secret provider inputs.
func (r *Repository) SaveProviderInputs(providerName string, inputs map[string]string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}

	ctx := context.Background()
	provider, err := r.GetProviderByName(ctx, providerName)
	if err != nil {
		return err
	}
	if provider.ID == "" {
		return fmt.Errorf("catalog repo: provider not found: %s", providerName)
	}

	return withTx(r.db, func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM provider_inputs WHERE provider_id = ?`, provider.ID); err != nil {
			return fmt.Errorf("catalog repo: clear inputs: %w", err)
		}
		for key, value := range inputs {
			trimmedKey := strings.TrimSpace(key)
			trimmedValue := strings.TrimSpace(value)
			if trimmedKey == "" || trimmedValue == "" {
				continue
			}
			if providercore.IsSensitiveCredentialName(trimmedKey) {
				return fmt.Errorf("catalog repo: secret-like input key is not allowed: %s", trimmedKey)
			}
			if _, err := tx.Exec(`INSERT INTO provider_inputs (provider_id, input_key, input_value) VALUES (?, ?, ?)`, provider.ID, trimmedKey, trimmedValue); err != nil {
				return fmt.Errorf("catalog repo: insert input: %w", err)
			}
		}
		return nil
	})
}

// LoadProviderInputs returns stored provider inputs.
func (r *Repository) LoadProviderInputs(providerName string) (map[string]string, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	ctx := context.Background()
	provider, err := r.GetProviderByName(ctx, providerName)
	if err != nil {
		return nil, err
	}
	if provider.ID == "" {
		return nil, fmt.Errorf("catalog repo: provider not found: %s", providerName)
	}

	rows, err := r.db.Query(`SELECT input_key, input_value FROM provider_inputs WHERE provider_id = ?`, provider.ID)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: load inputs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	inputs := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("catalog repo: load inputs scan: %w", err)
		}
		inputs[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: load inputs rows: %w", err)
	}
	if len(inputs) == 0 {
		return nil, nil
	}
	return inputs, nil
}

// UpsertEndpoint inserts or updates an endpoint record.
func (r *Repository) UpsertEndpoint(ctx context.Context, endpoint EndpointRecord) (EndpointRecord, error) {

	if r == nil || r.db == nil {
		return EndpointRecord{}, fmt.Errorf("catalog repo: db required")
	}
	if endpoint.ProviderID == "" {
		return EndpointRecord{}, fmt.Errorf("catalog repo: provider id required")
	}
	if endpoint.DisplayName == "" {
		return EndpointRecord{}, fmt.Errorf("catalog repo: endpoint display name required")
	}
	if endpoint.AdapterType == "" {
		return EndpointRecord{}, fmt.Errorf("catalog repo: endpoint adapter type required")
	}

	now := time.Now().UnixMilli()
	if endpoint.ID == "" {
		endpoint.ID = newUUID()
	}
	endpoint.RouteKind = normalizeOptional(endpoint.RouteKind)
	endpoint.OriginProvider = normalizeOptional(endpoint.OriginProvider)
	endpoint.OriginRouteLabel = normalizeOptional(endpoint.OriginRouteLabel)
	authJSON, err := sanitizeEndpointAuthJSON(endpoint.AuthJSON)
	if err != nil {
		return EndpointRecord{}, err
	}

	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO catalog_endpoints (id, provider_id, display_name, adapter_type, base_url, route_kind, origin_provider, origin_route_label, auth_json, created_at, updated_at, last_test_at, last_test_ok, last_error)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
         ON CONFLICT(provider_id, route_kind, origin_provider, origin_route_label, base_url) DO UPDATE SET
           display_name = excluded.display_name,
           adapter_type = excluded.adapter_type,
           auth_json = excluded.auth_json,
           last_test_at = excluded.last_test_at,
           last_test_ok = excluded.last_test_ok,
           last_error = excluded.last_error,
           updated_at = excluded.updated_at`,
		endpoint.ID,
		endpoint.ProviderID,
		endpoint.DisplayName,
		endpoint.AdapterType,
		endpoint.BaseURL,
		endpoint.RouteKind,
		endpoint.OriginProvider,
		endpoint.OriginRouteLabel,
		authJSON,
		now,
		now,
		nullIfZero(endpoint.LastTestAt),
		boolToInt(endpoint.LastTestOK),
		strings.TrimSpace(endpoint.LastError),
	)
	if err != nil {
		return EndpointRecord{}, fmt.Errorf("catalog repo: upsert endpoint: %w", err)
	}

	row := r.db.QueryRowContext(ctx, `SELECT id, provider_id, display_name, adapter_type, base_url, route_kind, origin_provider, origin_route_label, auth_json, last_test_at, last_test_ok, last_error FROM catalog_endpoints WHERE provider_id = ? AND route_kind = ? AND origin_provider = ? AND origin_route_label = ? AND base_url = ?`, endpoint.ProviderID, endpoint.RouteKind, endpoint.OriginProvider, endpoint.OriginRouteLabel, endpoint.BaseURL)
	var record EndpointRecord
	var authJSON sql.NullString
	var lastTestAt sql.NullInt64
	var lastTestOK sql.NullInt64
	var lastError sql.NullString
	if err := row.Scan(
		&record.ID,
		&record.ProviderID,
		&record.DisplayName,
		&record.AdapterType,
		&record.BaseURL,
		&record.RouteKind,
		&record.OriginProvider,
		&record.OriginRouteLabel,
		&authJSON,
		&lastTestAt,
		&lastTestOK,
		&lastError,
	); err != nil {
		return EndpointRecord{}, fmt.Errorf("catalog repo: load endpoint: %w", err)
	}
	if authJSON.Valid {
		record.AuthJSON = authJSON.String
	}
	if lastTestAt.Valid {
		record.LastTestAt = lastTestAt.Int64
	}
	record.LastTestOK = lastTestOK.Valid && lastTestOK.Int64 == 1
	if lastError.Valid {
		record.LastError = lastError.String
	}
	return record, nil
}

// UpdateEndpointStatus updates endpoint health metadata.
func (r *Repository) UpdateEndpointStatus(ctx context.Context, endpointID string, lastTestAt int64, lastTestOK bool, lastError string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if endpointID == "" {
		return fmt.Errorf("catalog repo: endpoint id required")
	}

	_, err := r.db.ExecContext(
		ctx,
		`UPDATE catalog_endpoints SET last_test_at = ?, last_test_ok = ?, last_error = ?, updated_at = ? WHERE id = ?`,
		nullIfZero(lastTestAt),
		boolToInt(lastTestOK),
		strings.TrimSpace(lastError),
		time.Now().UnixMilli(),
		endpointID,
	)
	if err != nil {
		return fmt.Errorf("catalog repo: update endpoint status: %w", err)
	}
	return nil
}

// ListEndpoints returns all endpoints for UI display.
func (r *Repository) ListEndpoints(ctx context.Context) ([]EndpointRecord, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT e.id, e.provider_id, p.name, e.display_name, e.adapter_type, e.base_url, e.route_kind, e.origin_provider, e.origin_route_label, e.auth_json, e.last_test_at, e.last_test_ok, e.last_error FROM catalog_endpoints e JOIN catalog_providers p ON p.id = e.provider_id ORDER BY p.display_name, e.display_name`)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list endpoints: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var endpoints []EndpointRecord
	for rows.Next() {
		var record EndpointRecord
		var authJSON sql.NullString
		var lastTestAt sql.NullInt64
		var lastTestOK sql.NullInt64
		var lastError sql.NullString
		if err := rows.Scan(
			&record.ID,
			&record.ProviderID,
			&record.ProviderName,
			&record.DisplayName,
			&record.AdapterType,
			&record.BaseURL,
			&record.RouteKind,
			&record.OriginProvider,
			&record.OriginRouteLabel,
			&authJSON,
			&lastTestAt,
			&lastTestOK,
			&lastError,
		); err != nil {
			return nil, fmt.Errorf("catalog repo: list endpoints scan: %w", err)
		}
		if authJSON.Valid {
			record.AuthJSON = authJSON.String
		}
		if lastTestAt.Valid {
			record.LastTestAt = lastTestAt.Int64
		}
		record.LastTestOK = lastTestOK.Valid && lastTestOK.Int64 == 1
		if lastError.Valid {
			record.LastError = lastError.String
		}
		endpoints = append(endpoints, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: list endpoints rows: %w", err)
	}
	return endpoints, nil
}

// GetEndpoint loads an endpoint by ID.
func (r *Repository) GetEndpoint(ctx context.Context, endpointID string) (EndpointRecord, error) {

	if r == nil || r.db == nil {
		return EndpointRecord{}, fmt.Errorf("catalog repo: db required")
	}

	row := r.db.QueryRowContext(ctx, `SELECT e.id, e.provider_id, p.name, e.display_name, e.adapter_type, e.base_url, e.route_kind, e.origin_provider, e.origin_route_label, e.auth_json, e.last_test_at, e.last_test_ok, e.last_error FROM catalog_endpoints e JOIN catalog_providers p ON p.id = e.provider_id WHERE e.id = ?`, endpointID)
	var record EndpointRecord
	var authJSON sql.NullString
	var lastTestAt sql.NullInt64
	var lastTestOK sql.NullInt64
	var lastError sql.NullString
	err := row.Scan(
		&record.ID,
		&record.ProviderID,
		&record.ProviderName,
		&record.DisplayName,
		&record.AdapterType,
		&record.BaseURL,
		&record.RouteKind,
		&record.OriginProvider,
		&record.OriginRouteLabel,
		&authJSON,
		&lastTestAt,
		&lastTestOK,
		&lastError,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return EndpointRecord{}, nil
	}
	if err != nil {
		return EndpointRecord{}, fmt.Errorf("catalog repo: get endpoint: %w", err)
	}
	if authJSON.Valid {
		record.AuthJSON = authJSON.String
	}
	if lastTestAt.Valid {
		record.LastTestAt = lastTestAt.Int64
	}
	record.LastTestOK = lastTestOK.Valid && lastTestOK.Int64 == 1
	if lastError.Valid {
		record.LastError = lastError.String
	}
	return record, nil
}

// ListModelEntriesByEndpoint returns model catalog entries for an endpoint.
func (r *Repository) ListModelEntriesByEndpoint(ctx context.Context, endpointID string) ([]ModelEntryRecord, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, endpoint_id, model_id, display_name, first_seen_at, last_seen_at, availability_state, approved, missed_refreshes, metadata_json FROM model_catalog_entries WHERE endpoint_id = ?`, endpointID)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list models: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []ModelEntryRecord
	for rows.Next() {
		var entry ModelEntryRecord
		var approved int
		if err := rows.Scan(
			&entry.ID,
			&entry.EndpointID,
			&entry.ModelID,
			&entry.DisplayName,
			&entry.FirstSeenAt,
			&entry.LastSeenAt,
			&entry.AvailabilityState,
			&approved,
			&entry.MissedRefreshes,
			&entry.MetadataJSON,
		); err != nil {
			return nil, fmt.Errorf("catalog repo: list models scan: %w", err)
		}
		entry.Approved = approved == 1
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: list models rows: %w", err)
	}
	return entries, nil
}

// UpsertModelEntry inserts or updates a catalog entry.
func (r *Repository) UpsertModelEntry(ctx context.Context, entry ModelEntryRecord) (ModelEntryRecord, error) {

	if r == nil || r.db == nil {
		return ModelEntryRecord{}, fmt.Errorf("catalog repo: db required")
	}
	if entry.EndpointID == "" || entry.ModelID == "" {
		return ModelEntryRecord{}, fmt.Errorf("catalog repo: model entry requires endpoint and model id")
	}

	if entry.ID == "" {
		entry.ID = newUUID()
	}
	if entry.Source == "" {
		entry.Source = "discovered"
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO model_catalog_entries (id, endpoint_id, model_id, display_name, first_seen_at, last_seen_at, availability_state, approved, missed_refreshes, source, metadata_json)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
         ON CONFLICT(endpoint_id, model_id) DO UPDATE SET
           display_name = excluded.display_name,
           last_seen_at = excluded.last_seen_at,
           availability_state = excluded.availability_state,
           approved = excluded.approved,
           missed_refreshes = excluded.missed_refreshes,
           metadata_json = excluded.metadata_json`,
		entry.ID,
		entry.EndpointID,
		entry.ModelID,
		entry.DisplayName,
		entry.FirstSeenAt,
		entry.LastSeenAt,
		entry.AvailabilityState,
		boolToInt(entry.Approved),
		entry.MissedRefreshes,
		entry.Source,
		entry.MetadataJSON,
	)
	if err != nil {
		return ModelEntryRecord{}, fmt.Errorf("catalog repo: upsert model: %w", err)
	}

	row := r.db.QueryRowContext(ctx, `SELECT id, endpoint_id, model_id, display_name, first_seen_at, last_seen_at, availability_state, approved, missed_refreshes, source, metadata_json FROM model_catalog_entries WHERE endpoint_id = ? AND model_id = ?`, entry.EndpointID, entry.ModelID)
	var updated ModelEntryRecord
	var approved int
	if err := row.Scan(
		&updated.ID,
		&updated.EndpointID,
		&updated.ModelID,
		&updated.DisplayName,
		&updated.FirstSeenAt,
		&updated.LastSeenAt,
		&updated.AvailabilityState,
		&approved,
		&updated.MissedRefreshes,
		&updated.Source,
		&updated.MetadataJSON,
	); err != nil {
		return ModelEntryRecord{}, fmt.Errorf("catalog repo: load model: %w", err)
	}
	updated.Approved = approved == 1
	return updated, nil
}

// UpdateMissingModelEntry updates a missing model entry.
func (r *Repository) UpdateMissingModelEntry(ctx context.Context, entryID string, missedRefreshes int, availabilityState string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if entryID == "" {
		return fmt.Errorf("catalog repo: model entry id required")
	}

	_, err := r.db.ExecContext(
		ctx,
		`UPDATE model_catalog_entries SET missed_refreshes = ?, availability_state = ? WHERE id = ?`,
		missedRefreshes,
		availabilityState,
		entryID,
	)
	if err != nil {
		return fmt.Errorf("catalog repo: update missing model: %w", err)
	}
	return nil
}

// EnsureModelCapabilities inserts model capabilities if missing.
func (r *Repository) EnsureModelCapabilities(ctx context.Context, entryID string, supportsStreaming, supportsToolCalling, supportsStructuredOutput, supportsVision bool, inputModalities, outputModalities []string, source string, asOf int64) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if entryID == "" {
		return fmt.Errorf("catalog repo: model entry id required")
	}

	return withTx(r.db, func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`INSERT INTO model_capabilities (model_catalog_entry_id, supports_streaming, supports_tool_calling, supports_structured_output, supports_vision, capabilities_source, capabilities_as_of)
             VALUES (?, ?, ?, ?, ?, ?, ?)
             ON CONFLICT(model_catalog_entry_id) DO UPDATE SET
               supports_streaming = excluded.supports_streaming,
               supports_tool_calling = excluded.supports_tool_calling,
               supports_structured_output = excluded.supports_structured_output,
               supports_vision = excluded.supports_vision,
               capabilities_source = excluded.capabilities_source,
               capabilities_as_of = excluded.capabilities_as_of`,
			entryID,
			boolToInt(supportsStreaming),
			boolToInt(supportsToolCalling),
			boolToInt(supportsStructuredOutput),
			boolToInt(supportsVision),
			source,
			asOf,
		)
		if err != nil {
			return fmt.Errorf("catalog repo: insert capabilities: %w", err)
		}

		if err := replaceModalities(tx, "model_capabilities_input_modalities", entryID, inputModalities); err != nil {
			return err
		}
		if err := replaceModalities(tx, "model_capabilities_output_modalities", entryID, outputModalities); err != nil {
			return err
		}
		return nil
	})
}

// EnsureModelSystemProfile inserts system profile defaults if missing.
func (r *Repository) EnsureModelSystemProfile(ctx context.Context, entryID string, latencyTier, costTier, reliabilityTier, source string, asOf int64) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if entryID == "" {
		return fmt.Errorf("catalog repo: model entry id required")
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO model_system_profile (model_catalog_entry_id, latency_tier, cost_tier, reliability_tier, system_profile_source, system_profile_as_of)
         VALUES (?, ?, ?, ?, ?, ?)
         ON CONFLICT(model_catalog_entry_id) DO UPDATE SET
           latency_tier = excluded.latency_tier,
           cost_tier = excluded.cost_tier,
           reliability_tier = excluded.reliability_tier,
           system_profile_source = excluded.system_profile_source,
           system_profile_as_of = excluded.system_profile_as_of`,
		entryID,
		latencyTier,
		costTier,
		reliabilityTier,
		source,
		asOf,
	)
	if err != nil {
		return fmt.Errorf("catalog repo: insert system profile: %w", err)
	}
	return nil
}

// EnsureModelUserAddenda inserts a user addenda row if missing.
func (r *Repository) EnsureModelUserAddenda(ctx context.Context, entryID string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if entryID == "" {
		return fmt.Errorf("catalog repo: model entry id required")
	}

	now := time.Now().UnixMilli()
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO model_user_addenda (model_catalog_entry_id, notes, user_addenda_source, user_addenda_as_of)
         VALUES (?, ?, ?, ?)
         ON CONFLICT(model_catalog_entry_id) DO NOTHING`,
		entryID,
		"",
		"manual",
		now,
	)
	if err != nil {
		return fmt.Errorf("catalog repo: insert user addenda: %w", err)
	}
	return nil
}

// ModelCapabilitiesRecord holds model capabilities and modalities.
type ModelCapabilitiesRecord struct {
	SupportsStreaming        bool
	SupportsToolCalling      bool
	SupportsStructuredOutput bool
	SupportsVision           bool
	InputModalities          []string
	OutputModalities         []string
}

// GetModelEffectiveCostTier returns the resolved cost tier for a model entry.
func (r *Repository) GetModelEffectiveCostTier(ctx context.Context, entryID string) (string, error) {

	if r == nil || r.db == nil {
		return "", fmt.Errorf("catalog repo: db required")
	}
	if strings.TrimSpace(entryID) == "" {
		return "", fmt.Errorf("catalog repo: model entry id required")
	}

	var costTier sql.NullString
	err := r.db.QueryRowContext(
		ctx,
		`SELECT COALESCE(ua.cost_tier_override, sp.cost_tier)
         FROM model_system_profile sp
         LEFT JOIN model_user_addenda ua ON ua.model_catalog_entry_id = sp.model_catalog_entry_id
         WHERE sp.model_catalog_entry_id = ?`,
		entryID,
	).Scan(&costTier)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("catalog repo: get model cost tier: %w", err)
	}
	if costTier.Valid {
		return costTier.String, nil
	}
	return "", nil
}

// GetModelCapabilities loads model capabilities for a model entry.
func (r *Repository) GetModelCapabilities(ctx context.Context, entryID string) (ModelCapabilitiesRecord, error) {

	if r == nil || r.db == nil {
		return ModelCapabilitiesRecord{}, fmt.Errorf("catalog repo: db required")
	}

	var record ModelCapabilitiesRecord
	var supportsStreaming, supportsToolCalling, supportsStructuredOutput, supportsVision int
	err := r.db.QueryRowContext(
		ctx,
		`SELECT supports_streaming, supports_tool_calling, supports_structured_output, supports_vision FROM model_capabilities WHERE model_catalog_entry_id = ?`,
		entryID,
	).Scan(&supportsStreaming, &supportsToolCalling, &supportsStructuredOutput, &supportsVision)
	if errors.Is(err, sql.ErrNoRows) {
		return ModelCapabilitiesRecord{}, nil
	}
	if err != nil {
		return ModelCapabilitiesRecord{}, fmt.Errorf("catalog repo: get capabilities: %w", err)
	}
	record.SupportsStreaming = supportsStreaming == 1
	record.SupportsToolCalling = supportsToolCalling == 1
	record.SupportsStructuredOutput = supportsStructuredOutput == 1
	record.SupportsVision = supportsVision == 1

	inputModalities, err := loadModalities(r.db, "model_capabilities_input_modalities", "model_catalog_entry_id", entryID)
	if err != nil {
		return ModelCapabilitiesRecord{}, err
	}
	outputModalities, err := loadModalities(r.db, "model_capabilities_output_modalities", "model_catalog_entry_id", entryID)
	if err != nil {
		return ModelCapabilitiesRecord{}, err
	}
	record.InputModalities = inputModalities
	record.OutputModalities = outputModalities
	return record, nil
}

// ModelSummaryRecord returns model entries with capabilities.
type ModelSummaryRecord struct {
	ModelEntryRecord
	ModelCapabilitiesRecord
	CostTier string
}

// ListModelSystemTags returns system tags grouped by model entry id.
func (r *Repository) ListModelSystemTags(ctx context.Context) (map[string][]string, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT model_catalog_entry_id, tag FROM model_system_tags ORDER BY model_catalog_entry_id, tag`)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list model system tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tagsByEntryID := make(map[string][]string)
	for rows.Next() {
		var entryID string
		var tag string
		if err := rows.Scan(&entryID, &tag); err != nil {
			return nil, fmt.Errorf("catalog repo: list model system tags scan: %w", err)
		}
		tagsByEntryID[entryID] = append(tagsByEntryID[entryID], tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: list model system tags rows: %w", err)
	}

	return tagsByEntryID, nil
}

type endpointAuthPayload struct {
	CredentialFields []endpointAuthCredentialField `json:"credential_fields,omitempty"`
}

type endpointAuthCredentialField struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Secret   bool   `json:"secret"`
}

// ListModelSummaries returns models with intrinsic metadata.
func (r *Repository) ListModelSummaries(ctx context.Context) ([]ModelSummaryRecord, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, endpoint_id, model_id, display_name, first_seen_at, last_seen_at, availability_state, approved, missed_refreshes, source, metadata_json FROM model_catalog_entries ORDER BY model_id`)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list model summaries: %w", err)
	}

	entries := make([]ModelEntryRecord, 0)
	for rows.Next() {
		var entry ModelEntryRecord
		var approved int
		if err := rows.Scan(
			&entry.ID,
			&entry.EndpointID,
			&entry.ModelID,
			&entry.DisplayName,
			&entry.FirstSeenAt,
			&entry.LastSeenAt,
			&entry.AvailabilityState,
			&approved,
			&entry.MissedRefreshes,
			&entry.Source,
			&entry.MetadataJSON,
		); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("catalog repo: list model summaries scan: %w", err)
		}
		entry.Approved = approved == 1
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("catalog repo: list model summaries rows: %w", err)
	}

	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("catalog repo: list model summaries close: %w", err)
	}

	// Load per-model details only after closing the result cursor to avoid blocking
	// when the datastore is configured with a single open SQLite connection.
	summaries := make([]ModelSummaryRecord, 0, len(entries))
	for _, entry := range entries {
		capabilities, err := r.GetModelCapabilities(ctx, entry.ID)
		if err != nil {
			return nil, err
		}
		costTier, err := r.GetModelEffectiveCostTier(ctx, entry.ID)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, ModelSummaryRecord{
			ModelEntryRecord:        entry,
			ModelCapabilitiesRecord: capabilities,
			CostTier:                costTier,
		})
	}

	return summaries, nil
}

// UpsertRole inserts or updates a role.
func (r *Repository) UpsertRole(ctx context.Context, role RoleRecord) (RoleRecord, error) {

	if r == nil || r.db == nil {
		return RoleRecord{}, fmt.Errorf("catalog repo: db required")
	}
	role.Name = strings.TrimSpace(role.Name)
	if role.Name == "" {
		return RoleRecord{}, fmt.Errorf("catalog repo: role name required")
	}
	now := time.Now().UnixMilli()
	if role.ID == "" {
		role.ID = newUUID()
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO roles (id, name, requires_streaming, requires_tool_calling, requires_structured_output, requires_vision, max_cost_tier, max_latency_tier, min_reliability_tier, created_at, updated_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
         ON CONFLICT(name) DO UPDATE SET
           requires_streaming = excluded.requires_streaming,
           requires_tool_calling = excluded.requires_tool_calling,
           requires_structured_output = excluded.requires_structured_output,
           requires_vision = excluded.requires_vision,
           max_cost_tier = excluded.max_cost_tier,
           max_latency_tier = excluded.max_latency_tier,
           min_reliability_tier = excluded.min_reliability_tier,
           updated_at = excluded.updated_at`,
		role.ID,
		role.Name,
		boolToInt(role.RequiresStreaming),
		boolToInt(role.RequiresToolCalling),
		boolToInt(role.RequiresStructuredOutput),
		boolToInt(role.RequiresVision),
		normalizeOptional(role.MaxCostTier),
		normalizeOptional(role.MaxLatencyTier),
		normalizeOptional(role.MinReliabilityTier),
		now,
		now,
	)
	if err != nil {
		return RoleRecord{}, fmt.Errorf("catalog repo: upsert role: %w", err)
	}

	record, err := r.GetRoleByName(ctx, role.Name)
	if err != nil {
		return RoleRecord{}, err
	}
	if err := r.replaceRoleModalities(ctx, record.ID, role.RequiredInputModalities, role.RequiredOutputModalities); err != nil {
		return RoleRecord{}, err
	}
	record.RequiredInputModalities = role.RequiredInputModalities
	record.RequiredOutputModalities = role.RequiredOutputModalities
	return record, nil
}

// GetRoleByID loads a role by ID.
func (r *Repository) GetRoleByID(ctx context.Context, roleID string) (RoleRecord, error) {

	if r == nil || r.db == nil {
		return RoleRecord{}, fmt.Errorf("catalog repo: db required")
	}

	row := r.db.QueryRowContext(ctx, `SELECT id, name, requires_streaming, requires_tool_calling, requires_structured_output, requires_vision, max_cost_tier, max_latency_tier, min_reliability_tier FROM roles WHERE id = ?`, roleID)
	record, err := scanRole(row)
	if errors.Is(err, sql.ErrNoRows) {
		return RoleRecord{}, nil
	}
	if err != nil {
		return RoleRecord{}, err
	}

	inputModalities, err := loadModalities(r.db, "role_required_input_modalities", "role_id", record.ID)
	if err != nil {
		return RoleRecord{}, err
	}
	outputModalities, err := loadModalities(r.db, "role_required_output_modalities", "role_id", record.ID)
	if err != nil {
		return RoleRecord{}, err
	}
	record.RequiredInputModalities = inputModalities
	record.RequiredOutputModalities = outputModalities
	return record, nil
}

// GetRoleByName loads a role by name.
func (r *Repository) GetRoleByName(ctx context.Context, name string) (RoleRecord, error) {

	if r == nil || r.db == nil {
		return RoleRecord{}, fmt.Errorf("catalog repo: db required")
	}

	row := r.db.QueryRowContext(ctx, `SELECT id, name, requires_streaming, requires_tool_calling, requires_structured_output, requires_vision, max_cost_tier, max_latency_tier, min_reliability_tier FROM roles WHERE name = ?`, name)
	record, err := scanRole(row)
	if errors.Is(err, sql.ErrNoRows) {
		return RoleRecord{}, nil
	}
	if err != nil {
		return RoleRecord{}, err
	}

	inputModalities, err := loadModalities(r.db, "role_required_input_modalities", "role_id", record.ID)
	if err != nil {
		return RoleRecord{}, err
	}
	outputModalities, err := loadModalities(r.db, "role_required_output_modalities", "role_id", record.ID)
	if err != nil {
		return RoleRecord{}, err
	}
	record.RequiredInputModalities = inputModalities
	record.RequiredOutputModalities = outputModalities
	return record, nil
}

// ListRoles returns all roles.
func (r *Repository) ListRoles(ctx context.Context) ([]RoleRecord, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, name, requires_streaming, requires_tool_calling, requires_structured_output, requires_vision, max_cost_tier, max_latency_tier, min_reliability_tier FROM roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var roles []RoleRecord
	for rows.Next() {
		record, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		inputModalities, err := loadModalities(r.db, "role_required_input_modalities", "role_id", record.ID)
		if err != nil {
			return nil, err
		}
		outputModalities, err := loadModalities(r.db, "role_required_output_modalities", "role_id", record.ID)
		if err != nil {
			return nil, err
		}
		record.RequiredInputModalities = inputModalities
		record.RequiredOutputModalities = outputModalities
		roles = append(roles, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: list roles rows: %w", err)
	}
	return roles, nil
}

// DeleteRole removes a role by ID.
func (r *Repository) DeleteRole(ctx context.Context, roleID string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if roleID == "" {
		return fmt.Errorf("catalog repo: role id required")
	}

	_, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id = ?`, roleID)
	if err != nil {
		return fmt.Errorf("catalog repo: delete role: %w", err)
	}
	return nil
}

// ListRoleAssignments returns all role assignments.
func (r *Repository) ListRoleAssignments(ctx context.Context) ([]RoleAssignmentRecord, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT role_id, model_catalog_entry_id, assigned_by, created_at, enabled FROM role_assignments`)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list role assignments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var assignments []RoleAssignmentRecord
	for rows.Next() {
		var record RoleAssignmentRecord
		var enabled int
		if err := rows.Scan(&record.RoleID, &record.ModelCatalogEntryID, &record.AssignedBy, &record.CreatedAt, &enabled); err != nil {
			return nil, fmt.Errorf("catalog repo: list role assignments scan: %w", err)
		}
		record.Enabled = enabled == 1
		assignments = append(assignments, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: list role assignments rows: %w", err)
	}
	return assignments, nil
}

// UpsertRoleAssignment inserts or updates a role assignment.
func (r *Repository) UpsertRoleAssignment(ctx context.Context, assignment RoleAssignmentRecord) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if assignment.RoleID == "" || assignment.ModelCatalogEntryID == "" {
		return fmt.Errorf("catalog repo: role assignment requires role and model")
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO role_assignments (role_id, model_catalog_entry_id, assigned_by, created_at, enabled)
         VALUES (?, ?, ?, ?, ?)
         ON CONFLICT(role_id, model_catalog_entry_id) DO UPDATE SET
           assigned_by = excluded.assigned_by,
           created_at = excluded.created_at,
           enabled = excluded.enabled`,
		assignment.RoleID,
		assignment.ModelCatalogEntryID,
		assignment.AssignedBy,
		assignment.CreatedAt,
		boolToInt(assignment.Enabled),
	)
	if err != nil {
		return fmt.Errorf("catalog repo: upsert role assignment: %w", err)
	}
	return nil
}

// DeleteRoleAssignment removes a role assignment.
func (r *Repository) DeleteRoleAssignment(ctx context.Context, roleID, modelEntryID string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("catalog repo: db required")
	}
	if roleID == "" || modelEntryID == "" {
		return fmt.Errorf("catalog repo: role assignment requires role and model")
	}

	_, err := r.db.ExecContext(ctx, `DELETE FROM role_assignments WHERE role_id = ? AND model_catalog_entry_id = ?`, roleID, modelEntryID)
	if err != nil {
		return fmt.Errorf("catalog repo: delete role assignment: %w", err)
	}
	return nil
}

// ListModelLabels returns model labels for assignment display.
func (r *Repository) ListModelLabels(ctx context.Context) (map[string]string, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("catalog repo: db required")
	}

	rows, err := r.db.QueryContext(ctx, `SELECT m.id, m.model_id, e.display_name, p.display_name FROM model_catalog_entries m JOIN catalog_endpoints e ON e.id = m.endpoint_id JOIN catalog_providers p ON p.id = e.provider_id`)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: list model labels: %w", err)
	}
	defer func() { _ = rows.Close() }()

	labels := make(map[string]string)
	for rows.Next() {
		var id, modelID, endpointName, providerName string
		if err := rows.Scan(&id, &modelID, &endpointName, &providerName); err != nil {
			return nil, fmt.Errorf("catalog repo: list model labels scan: %w", err)
		}
		labels[id] = fmt.Sprintf("%s · %s · %s", providerName, endpointName, modelID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: list model labels rows: %w", err)
	}
	return labels, nil
}

// MarshalAuthJSON creates a JSON string for endpoint auth hints.
func MarshalAuthJSON(data map[string]interface{}) (string, error) {

	if data == nil {
		return "", nil
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("catalog repo: encode auth json: %w", err)
	}
	return string(payload), nil
}

// sanitizeEndpointAuthJSON validates and normalizes endpoint auth metadata shape.
func sanitizeEndpointAuthJSON(raw string) (string, error) {

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	decoder := json.NewDecoder(strings.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	var payload endpointAuthPayload
	if err := decoder.Decode(&payload); err != nil {
		return "", fmt.Errorf("catalog repo: invalid auth metadata: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("catalog repo: invalid auth metadata: trailing content")
	}

	for i := range payload.CredentialFields {
		payload.CredentialFields[i].Name = strings.TrimSpace(payload.CredentialFields[i].Name)
		if payload.CredentialFields[i].Name == "" {
			return "", fmt.Errorf("catalog repo: auth metadata credential field name required")
		}
		if providercore.IsSensitiveCredentialName(payload.CredentialFields[i].Name) && !payload.CredentialFields[i].Secret {
			return "", fmt.Errorf("catalog repo: credential field %q must be marked secret", payload.CredentialFields[i].Name)
		}
	}

	normalized, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("catalog repo: encode auth metadata: %w", err)
	}
	return string(normalized), nil
}

func newUUID() string {

	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}

func boolToInt(value bool) int {

	if value {
		return 1
	}
	return 0
}

func nullIfZero(value int64) interface{} {

	if value == 0 {
		return nil
	}
	return value
}

func normalizeOptional(value string) string {

	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return trimmed
}

func withTx(db *sql.DB, fn func(*sql.Tx) error) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func replaceModalities(tx *sql.Tx, table string, entryID string, modalities []string) error {

	if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE model_catalog_entry_id = ?", table), entryID); err != nil {
		return fmt.Errorf("catalog repo: clear modalities: %w", err)
	}
	for _, modality := range modalities {
		trimmed := strings.TrimSpace(modality)
		if trimmed == "" {
			continue
		}
		if _, err := tx.Exec(fmt.Sprintf("INSERT INTO %s (model_catalog_entry_id, modality) VALUES (?, ?)", table), entryID, trimmed); err != nil {
			return fmt.Errorf("catalog repo: insert modality: %w", err)
		}
	}
	return nil
}

func loadModalities(db *sql.DB, table, keyColumn, keyValue string) ([]string, error) {

	rows, err := db.Query(fmt.Sprintf("SELECT modality FROM %s WHERE %s = ? ORDER BY modality", table, keyColumn), keyValue)
	if err != nil {
		return nil, fmt.Errorf("catalog repo: load modalities: %w", err)
	}
	defer func() { _ = rows.Close() }()

	modalities := []string{}
	for rows.Next() {
		var modality string
		if err := rows.Scan(&modality); err != nil {
			return nil, fmt.Errorf("catalog repo: load modalities scan: %w", err)
		}
		modalities = append(modalities, modality)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repo: load modalities rows: %w", err)
	}
	if len(modalities) == 0 {
		return nil, nil
	}
	return modalities, nil
}

func scanRole(scanner interface{ Scan(...interface{}) error }) (RoleRecord, error) {

	var record RoleRecord
	var requiresStreaming, requiresToolCalling, requiresStructuredOutput, requiresVision int
	err := scanner.Scan(
		&record.ID,
		&record.Name,
		&requiresStreaming,
		&requiresToolCalling,
		&requiresStructuredOutput,
		&requiresVision,
		&record.MaxCostTier,
		&record.MaxLatencyTier,
		&record.MinReliabilityTier,
	)
	if err != nil {
		return RoleRecord{}, err
	}
	record.RequiresStreaming = requiresStreaming == 1
	record.RequiresToolCalling = requiresToolCalling == 1
	record.RequiresStructuredOutput = requiresStructuredOutput == 1
	record.RequiresVision = requiresVision == 1
	return record, nil
}

func (r *Repository) replaceRoleModalities(ctx context.Context, roleID string, inputModalities, outputModalities []string) error {

	return withTx(r.db, func(tx *sql.Tx) error {
		if _, err := tx.Exec(`DELETE FROM role_required_input_modalities WHERE role_id = ?`, roleID); err != nil {
			return fmt.Errorf("catalog repo: clear input modalities: %w", err)
		}
		for _, modality := range inputModalities {
			trimmed := strings.TrimSpace(modality)
			if trimmed == "" {
				continue
			}
			if _, err := tx.Exec(`INSERT INTO role_required_input_modalities (role_id, modality) VALUES (?, ?)`, roleID, trimmed); err != nil {
				return fmt.Errorf("catalog repo: insert input modality: %w", err)
			}
		}

		if _, err := tx.Exec(`DELETE FROM role_required_output_modalities WHERE role_id = ?`, roleID); err != nil {
			return fmt.Errorf("catalog repo: clear output modalities: %w", err)
		}
		for _, modality := range outputModalities {
			trimmed := strings.TrimSpace(modality)
			if trimmed == "" {
				continue
			}
			if _, err := tx.Exec(`INSERT INTO role_required_output_modalities (role_id, modality) VALUES (?, ?)`, roleID, trimmed); err != nil {
				return fmt.Errorf("catalog repo: insert output modality: %w", err)
			}
		}
		return nil
	})
}
