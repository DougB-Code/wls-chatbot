-- schema.sql aggregates all application tables.

-- Config Store
CREATE TABLE IF NOT EXISTS app_config (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	config_json TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);

-- Catalog Repo
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

-- Chat Repo
CREATE TABLE IF NOT EXISTS chat_conversations (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	provider TEXT NOT NULL,
	model TEXT NOT NULL,
	temperature REAL NOT NULL,
	max_tokens INTEGER NOT NULL,
	system_prompt TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	is_archived INTEGER NOT NULL CHECK (is_archived IN (0, 1))
);

CREATE TABLE IF NOT EXISTS chat_messages (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	role TEXT NOT NULL,
	timestamp INTEGER NOT NULL,
	is_streaming INTEGER NOT NULL CHECK (is_streaming IN (0, 1)),
	provider TEXT,
	model TEXT,
	tokens_in INTEGER,
	tokens_out INTEGER,
	tokens_total INTEGER,
	latency_ms INTEGER,
	finish_reason TEXT,
	status_code INTEGER,
	error_message TEXT,
	FOREIGN KEY (conversation_id) REFERENCES chat_conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_conversation_ts
ON chat_messages (conversation_id, timestamp DESC);

CREATE TABLE IF NOT EXISTS chat_message_blocks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	message_id TEXT NOT NULL,
	block_index INTEGER NOT NULL,
	block_type TEXT NOT NULL,
	content TEXT NOT NULL,
	language TEXT,
	is_collapsed INTEGER NOT NULL CHECK (is_collapsed IN (0, 1)),
	artifact_id TEXT,
	artifact_name TEXT,
	artifact_type TEXT,
	artifact_content TEXT,
	artifact_language TEXT,
	artifact_version INTEGER,
	artifact_created_at INTEGER,
	artifact_updated_at INTEGER,
	action_id TEXT,
	action_tool_name TEXT,
	action_description TEXT,
	action_status TEXT,
	action_result TEXT,
	action_started_at INTEGER,
	action_completed_at INTEGER,
	FOREIGN KEY (message_id) REFERENCES chat_messages(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_message_blocks_order
ON chat_message_blocks (message_id, block_index);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	message TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	read_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_notifications_created_at
ON notifications (created_at DESC);
