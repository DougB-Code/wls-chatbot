<!-- docs/requirements/llm-connections-model-catalog-roles.md -->
<!-- LLM provider connections and model catalog requirements -->
# LLM Connections, Endpoints, Model Catalog, Capabilities, and Roles
Date: 2026-02-05  
Status: Draft (Implementation Target)

## 1. Objective

Implement a production-quality foundation for configuring and using multiple LLM providers and gateway/aggregator surfaces (initially Cloudflare and OpenRouter), with a clean abstraction over disparate APIs/SDKs.

The system must:
- Manage credentials and connectivity for multiple providers and gateways.
- Produce concrete callable Endpoints that represent a specific route + credentials.
- Discover models per Endpoint into a unified Model Catalog.
- Store capability metadata with provenance and validation rules.
- Define Roles with crisp semantic contracts and validate model-to-role assignments.

Chat model picker UX is out of scope. Only data/logic needed to support role assignment must be implemented.

## 2. Non-Goals

- Acting as a provider proxy / reselling tokens (operator-managed key brokering).
- Enterprise AI gateway support beyond Cloudflare + OpenRouter.
- Perfect cost accounting. Cost/latency are tier-based and advisory.
- Fully automated continuous verification in CI.
- Capability verification implementation in phase 1 (deferred to phase 2).

## 3. Core Concepts and Glossary

Implementation alignment: "Connection" is implemented as the existing Provider configuration and lifecycle. Do not introduce a parallel Connection abstraction in code.

### 3.1 Provider (Connection semantics)
A Provider represents a configured integration to a platform (provider or gateway/aggregator):
- Authentication material (API key, token, etc.)
- Base URL(s) and provider-level settings
- Provider adapter type (e.g., OpenAI, Anthropic, Cloudflare, OpenRouter)

A Provider may produce one or more Endpoints.

### 3.2 Endpoint
An Endpoint is a concrete callable target that the application uses to make API calls:
- Adapter type
- Base URL (actual host)
- Credentials used for requests
- Optional “route” metadata (e.g., Cloudflare gateway routing to upstream provider)

Direct providers typically produce a single Endpoint; gateway providers (Cloudflare) can produce multiple Endpoints.

Endpoints are the unit for:
- Health tests
- Model discovery
- Model capability verification (phase 2)

### 3.3 Model Catalog Entry
A Model Catalog Entry is a model identifier discovered via a specific Endpoint.
Catalog keys must include endpoint identity to avoid collisions.

### 3.4 Capability Metadata (Layered)
Capability metadata is stored in three layers:
- Intrinsic (hard, non-overridable technical constraints)
- System Profile (system-provided defaults; overridable)
- User Addenda (user-provided tags/preferences; additive)

### 3.5 Role
A Role is a semantic contract describing required modalities/features and optional constraints.
Models can be assigned to roles only if they satisfy role-required intrinsic capabilities.

### 3.6 Trust Mode
Providers operate in a Trust Mode:
- user_managed: user supplies keys; discovered models are assignable immediately.
- operator_managed: reserved for future; discovered models require approval before assignment.

Implementation must include this flag for forward compatibility, but only user_managed behavior is required now.

## 4. Supported Integrations (Initial)

### 4.1 Required adapters
- Cloudflare (supports hosted OSS models and gateway routing)
- OpenRouter (aggregator/route surface)
- At least one direct provider adapter must exist for end-to-end testing (choose one: OpenAI or Anthropic). The requirements below assume multiple direct providers will be added later.

### 4.2 Cloudflare behaviors
Cloudflare Provider must support:
- “Workers AI” hosted model access (provider-hosted surface).
- “Gateway routing” configuration that produces one Endpoint per upstream route.
- Ability for users to provide:
  - Cloudflare auth/token
  - Optional upstream provider keys per configured route (e.g., OpenAI key for CF-routed OpenAI usage)
  - Each gateway route persists its own upstream auth and is modeled as its own Endpoint.

### 4.3 OpenRouter behaviors
OpenRouter Provider must support:
- API key configuration.
- Optional `referer` and `title` inputs for OpenRouter analytics/rate-limit headers.
  - Persist in `provider_inputs` as `openrouter_referer` and `openrouter_title`.
  - Use `openrouter_referer` as the `HTTP-Referer` header and `openrouter_title` as the `X-Title` header on OpenRouter requests.
- Model discovery from OpenRouter.
- Model invocations via OpenRouter endpoint.

## 5. Data Model Requirements

### 5.0 Relational model and IDs
Must include:
- All primary keys use UUIDv7. Store as TEXT in SQLite and enforce foreign keys.
- Use normalized join tables for set-like fields (modalities, tags, role requirements). Avoid JSON for query-critical sets.
- Raw provider metadata snapshots may be stored as JSON text for debugging.
- Store timestamps as Unix milliseconds in integer columns.

### 5.0.1 Enumerations (Phase 1 scope)
Modalities:
- `text`
- `image`
- `audio`
- `video`

Origin providers (for `origin_provider` and gateway routes):
- `openai`
- `anthropic`
- `google`
- `mistral`
- `xai`
- `meta`
- `cohere`
- `openrouter`
- `cloudflare_workers_ai`
- `other` (use when the upstream is unknown or not listed)

### 5.1 Provider entity (Connection semantics)
Must include:
- id (UUIDv7)
- name (stable identifier)
- display_name
- adapter_type (enum)
- trust_mode (enum: user_managed | operator_managed)
- auth (typed by adapter, securely stored; secrets are stored outside the relational DB)
- base_url (optional override)
- inputs (non-secret key/value inputs like account IDs or OpenRouter headers; store in a normalized `provider_inputs` table keyed by provider_id)
- created_at, updated_at
- status summary fields (optional cached):
  - last_test_at
  - last_test_ok (bool)
  - last_error (string)
  - last_discovery_at

### 5.2 Endpoint entity
Must include:
- id (UUIDv7)
- provider_id (FK)
- display_name
- adapter_type (derived or explicit)
- base_url (resolved)
- auth (typed; may include nested upstream auth for gateway routes)
- route metadata (optional):
  - route_kind (enum; e.g., direct | gateway_route | hosted)
  - origin_provider (optional, for gateways/aggregators)
  - origin_route_label (optional, for gateway route identifiers)
- Cloudflare gateway routes must use route_kind = gateway_route and set origin_provider; Workers AI uses route_kind = hosted.
- health status fields:
  - last_test_at
  - last_test_ok
  - last_error

### 5.3 Model Catalog Entry entity
Must include:
- id (UUIDv7)
- endpoint_id (FK)
- model_id (provider model identifier string)
- display_name (optional)
- discovery:
  - first_seen_at
  - last_seen_at
  - availability_state (enum: available | unavailable | deprecated | unknown)
- provider metadata snapshot (optional raw JSON stored for debugging)
- unique constraint on (endpoint_id, model_id)

### 5.4 Capability metadata entities (layered)
Must include:
- `model_intrinsic` (1:1 with model_catalog_entry):
  - supports_streaming: bool
  - supports_tool_calling: bool
  - supports_structured_output: bool
  - supports_vision: bool (optional derived flag)
  - other hard constraints as required by adapter reality
  - intrinsic_source: enum (verified | declared)
  - intrinsic_as_of
- `model_intrinsic_input_modalities` (model_catalog_entry_id, modality)
- `model_intrinsic_output_modalities` (model_catalog_entry_id, modality)
- `model_system_profile` (1:1 with model_catalog_entry):
  - latency_tier: enum (fast | standard | slow | unknown)
  - cost_tier: enum (cheap | standard | expensive | unknown)
  - reliability_tier: enum (stable | preview | unknown)
  - system_profile_source: enum (verified | summarized | manual)
  - system_profile_as_of
- `model_system_tags` (model_catalog_entry_id, tag)
- `model_user_addenda` (1:1 with model_catalog_entry):
  - notes: string
  - user_addenda_source: manual
- `model_user_tags` (model_catalog_entry_id, tag)
- Original system_profile values must be retained even when user_addenda overrides advisory fields.

### 5.5 Role entity
Must include:
- id (UUIDv7)
- name (stable string; used in configs)
- required_input_modalities: set (normalized via `role_required_input_modalities`)
- required_output_modalities: set (normalized via `role_required_output_modalities`)
- required_features:
  - requires_streaming: bool
  - requires_tool_calling: bool
  - requires_structured_output: bool
  - requires_vision: bool
- optional_constraints:
  - max_cost_tier (optional)
  - max_latency_tier (optional)
  - min_reliability_tier (optional)

### 5.6 Role assignment entity
Must include:
- role_id (FK)
- model_catalog_entry_id (FK)
- assigned_by (user/system)
- created_at
- enabled (bool)
- unique constraint on (role_id, model_catalog_entry_id)

## 6. Capability Rules and Validation

### 6.1 Non-overridable intrinsic constraints
- Intrinsic capability fields are treated as hard technical truth.
- User cannot override intrinsic capability constraints.
- If the UI attempts to set a contradictory intrinsic field, reject the write.

### 6.2 Additive overrides
- User addenda must not delete intrinsic or system profile fields.
- User can override system_profile advisory fields (latency/cost/reliability) but the system must retain and display the original system values as well.

### 6.3 Effective capability view
The system must compute:
- effective_intrinsic = intrinsic
- effective_profile = merge(system_profile, user_addenda) where user values win for overlapping advisory fields

### 6.4 Role assignment validation
When assigning a model to a role:
- Validate role-required modalities are supported by model intrinsic capabilities.
- Validate required features are supported by model intrinsic capabilities.
- If validation fails:
  - Reject assignment
  - Return an error listing the missing requirements.

## 7. Discovery and Refresh

### 7.1 Endpoint health test
Implement an on-demand “Test Endpoint” action:
- Performs a low-cost request suitable for the adapter.
- Updates endpoint last_test_* fields.
- Produces actionable error messages.

### 7.2 Model discovery
Implement “Refresh Models” per Endpoint:
- Fetch model list from endpoint.
- Upsert Model Catalog Entries.
- Update last_seen_at for existing models.
- Run a refresh for all endpoints on app startup (fixed default; not configurable).
- Mark models not seen in 2 consecutive refreshes as unknown.

### 7.3 Discovery policy (user-managed)
In user_managed trust mode:
- Discovered models are assignable immediately (subject to role validation).
- No approval workflow required.

In operator_managed trust mode (future):
- Schema must support a boolean approved flag (even if always true/ignored for now).
- Implementation may stub enforcement.

## 8. Capability Verification (Phase 2, Deferred)

Phase 1: do not implement verification runners or UI actions. Ensure the schema can be extended in phase 2 without breaking changes.

Phase 2 (future) requirements:
- Runs a suite of conformance tests against a selected model catalog entry.
- Stores results, updates intrinsic capabilities when verified.
- Captures verification timestamp and test-suite version hash.
- Suite scope (minimum):
  - For text-capable models: basic completion, streaming support (if claimed), tool calling support (if claimed), structured output support (if claimed).
  - For multimodal models: vision input test (image understanding), image output test (image generation) if applicable.
- Embeddings / rerankers / audio are optional unless already integrated, but the framework must allow adding suites later.
- Verification must NOT run as part of standard unit tests and is invoked explicitly via UI action or CLI command in phase 2.

## 9. UI Requirements (Implementation Scope)

### 9.1 Consolidated Providers + Roles view
- Providers section (top of page):
  - List Providers with status.
  - Create/Edit Provider credentials (auth layout should mirror the current settings UX).
  - Actions: “Test” provider and “Refresh Models”.
  - For Cloudflare: configure hosted access and gateway routes inline; show each route as an Endpoint row under the Provider.
  - Under each Provider, show a read-only dropdown of discovered models and capabilities (endpoint-scoped).
- Roles section (below Providers):
  - Table layout with Role name, required modalities/features, and a model assignment control.
  - Assign/Unassign models to roles with inline validation feedback for missing requirements.
  - Show enabled/disabled state per assignment.

## 10. Security and Safety

- Secrets (API keys/tokens) must be stored securely and never logged.
- Redact secrets from UI and API responses by default.
- If verification is implemented in a later phase, verification samples must be redactable; do not store raw prompts/responses containing secrets unless explicitly opted-in.

## 11. Diagnostics and Observability

- Provide structured error codes for:
  - provider/auth failures
  - discovery failures
  - role validation failures
- Record timestamps and last error messages at Provider and Endpoint levels.

## 12. Acceptance Criteria (Must Pass)

### Providers / Endpoints
- Can create Cloudflare provider, test it, and produce at least one endpoint.
- Can create OpenRouter provider, test it, and produce endpoint.
- Can create at least one direct provider (OpenAI or Anthropic), test it.

### Discovery
- Refresh models per endpoint populates model catalog entries.
- App startup refresh populates/updates model catalog entries.
- Removing or changing credentials causes test/discovery to fail with actionable error.

### Capabilities
- Intrinsic capability constraints are enforced (cannot override to invalid modalities/features).
- System profile overrides are additive; system data remains visible/auditable.

### Roles
- Roles define required modalities/features.
- Assigning an incompatible model is rejected with specific missing constraints.
- Assigning a compatible model succeeds and persists.


## 13. Future Extensions (Informative, Not Required Now)

- Capability verification suite and runner (phase 2 scope described in section 8).
- Operator-managed trust mode enforcement (approval gates).
- Enterprise gateways (Kong/Apigee/LiteLLM Proxy/Helicone/Portkey) as additional gateway adapters.
- Policy-based automatic routing across role-assigned models.
- Fine-grained governance controls (PII handling, external fetch rules) at role/endpoint levels.
