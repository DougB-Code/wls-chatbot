<!-- docs/requirements/llm-connections-model-catalog-roles.md -->
<!-- LLM provider connections and model catalog requirements -->
# LLM Connections, Endpoints, Model Catalog, Capabilities, and Roles
Date: 2026-02-05  
Status: Draft (Implementation Target)

## 1. Objective

Implement a production-quality foundation for configuring and using multiple LLM providers and gateway/aggregator surfaces (initially Cloudflare and OpenRouter), with a clean abstraction over disparate APIs/SDKs.

The system must:
- Manage credentials and connectivity for multiple provider/gateway connections.
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

## 3. Core Concepts and Glossary

### 3.1 Connection
A Connection represents a configured integration to a platform (provider or gateway/aggregator):
- Authentication material (API key, token, etc.)
- Base URL(s) and connection-level settings
- Connection adapter type (e.g., OpenAI, Anthropic, Cloudflare, OpenRouter)

A Connection may produce one or more Endpoints.

### 3.2 Endpoint
An Endpoint is a concrete callable target that the application uses to make API calls:
- Adapter type
- Base URL (actual host)
- Credentials used for requests
- Optional “route” metadata (e.g., Cloudflare gateway routing to upstream provider)

Endpoints are the unit for:
- Health tests
- Model discovery
- Model capability verification

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
Connections operate in a Trust Mode:
- user_managed: user supplies keys; discovered models are assignable immediately.
- operator_managed: reserved for future; discovered models require approval before assignment.

Implementation must include this flag for forward compatibility, but only user_managed behavior is required now.

## 4. Supported Integrations (Initial)

### 4.1 Required adapters
- Cloudflare (supports hosted OSS models and gateway routing)
- OpenRouter (aggregator/route surface)
- At least one direct provider adapter must exist for end-to-end testing (choose one: OpenAI or Anthropic). The requirements below assume multiple direct providers will be added later.

### 4.2 Cloudflare behaviors
Cloudflare Connection must support:
- “Workers AI” hosted model access (provider-hosted surface).
- “Gateway routing” configuration that can produce endpoint routes to upstream providers.
- Ability for users to provide:
  - Cloudflare auth/token
  - Optional upstream provider keys per configured route (e.g., OpenAI key for CF-routed OpenAI usage)

### 4.3 OpenRouter behaviors
OpenRouter Connection must support:
- API key configuration.
- Model discovery from OpenRouter.
- Model invocations via OpenRouter endpoint.

## 5. Data Model Requirements

### 5.1 Connection entity
Must include:
- id (stable identifier)
- display_name
- adapter_type (enum)
- trust_mode (enum: user_managed | operator_managed)
- auth (typed by adapter, securely stored)
- base_url (optional override)
- created_at, updated_at
- status summary fields (optional cached):
  - last_test_at
  - last_test_ok (bool)
  - last_error (string)
  - last_discovery_at

### 5.2 Endpoint entity
Must include:
- id (stable identifier)
- connection_id (FK)
- display_name
- adapter_type (derived or explicit)
- base_url (resolved)
- auth (typed; may include nested upstream auth for gateway routes)
- route metadata (optional):
  - route_kind (enum; e.g., direct | gateway_route | hosted)
  - origin_provider (optional, for gateways/aggregators)
- health status fields:
  - last_test_at
  - last_test_ok
  - last_error

### 5.3 Model Catalog Entry entity
Must include:
- id (stable identifier)
- endpoint_id (FK)
- model_id (provider model identifier string)
- display_name (optional)
- discovery:
  - first_seen_at
  - last_seen_at
  - availability_state (enum: available | unavailable | deprecated | unknown)
- provider metadata snapshot (optional raw JSON stored for debugging)

### 5.4 Capability metadata entity (layered)
Must include:
- model_catalog_entry_id (FK)
- intrinsic (non-overridable):
  - supported_input_modalities: set
  - supported_output_modalities: set
  - supports_streaming: bool
  - supports_tool_calling: bool
  - supports_structured_output: bool
  - supports_vision: bool (optional derived flag)
  - other hard constraints as required by adapter reality
- system_profile (overridable defaults):
  - latency_tier: enum (fast | standard | slow | unknown)
  - cost_tier: enum (cheap | standard | expensive | unknown)
  - reliability_tier: enum (stable | preview | unknown)
  - recommended_tags: set (e.g., summary, code, reasoning)
- user_addenda (additive):
  - user_tags: set
  - notes: string
- provenance per field or per layer:
  - intrinsic_source: enum (verified | declared)
  - system_profile_source: enum (verified | summarized | manual)
  - user_addenda_source: manual
  - as_of timestamps:
    - intrinsic_as_of
    - system_profile_as_of

### 5.5 Role entity
Must include:
- id (stable identifier)
- name (stable string; used in configs)
- required_input_modalities: set
- required_output_modalities: set
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
- Mark models not seen in N refreshes as unavailable or unknown (policy configurable; default: unknown).

### 7.3 Discovery policy (user-managed)
In user_managed trust mode:
- Discovered models are assignable immediately (subject to role validation).
- No approval workflow required.

In operator_managed trust mode (future):
- Schema must support a boolean approved flag (even if always true/ignored for now).
- Implementation may stub enforcement.

## 8. Capability Verification (On-demand, paid suite)

### 8.1 Verification runner
Implement an on-demand capability verification command:
- Runs a suite of conformance tests against a selected model catalog entry.
- Stores results, updates intrinsic capabilities when verified.
- Captures verification timestamp and test-suite version hash.

### 8.2 Suite scope (minimum)
For text-capable models:
- basic completion
- streaming support (if claimed)
- tool calling support (if claimed)
- structured output support (if claimed)

For multimodal models (if present):
- vision input test (image understanding)
- image output test (image generation) if applicable

Embeddings / rerankers / audio are optional in the starter unless already integrated, but the framework must allow adding suites later.

### 8.3 CI behavior
- Verification must NOT run as part of standard unit tests.
- Verification is invoked explicitly via UI action or CLI command.

## 9. UI Requirements (Implementation Scope)

### 9.1 Connections view
- List Connections with status.
- Create/Edit Connection.
- For Cloudflare: support configuring both hosted access and gateway routing.
- Provide “Test” action.
- Provide “Manage Endpoints” (inline section or navigation).

### 9.2 Endpoints view (may be embedded under Connections)
- List Endpoints with status.
- “Refresh Models” action per Endpoint.
- Show last test and last error.

### 9.3 Models view
- Unified list/table of Model Catalog Entries across endpoints.
- Must include endpoint/provider column.
- Show verification “as of” dates.
- Actions:
  - Run capability verification
  - Edit user addenda metadata
  - Assign to role (with validation feedback)

### 9.4 Roles view
- Create/Edit Roles (contract fields).
- Assign/Unassign models to roles.
- Validation errors must be shown clearly.

## 10. Security and Safety

- Secrets (API keys/tokens) must be stored securely and never logged.
- Redact secrets from UI and API responses by default.
- Model verification samples must be redactable; do not store raw prompts/responses containing secrets unless explicitly opted-in.

## 11. Diagnostics and Observability

- Provide structured error codes for:
  - connection/auth failures
  - discovery failures
  - role validation failures
  - verification failures
- Record timestamps and last error messages at Connection and Endpoint levels.

## 12. Acceptance Criteria (Must Pass)

### Connections / Endpoints
- Can create Cloudflare connection, test it, and produce at least one endpoint.
- Can create OpenRouter connection, test it, and produce endpoint.
- Can create at least one direct provider connection (OpenAI or Anthropic), test it.

### Discovery
- Refresh models per endpoint populates model catalog entries.
- Removing or changing credentials causes test/discovery to fail with actionable error.

### Capabilities
- Intrinsic capability constraints are enforced (cannot override to invalid modalities/features).
- System profile overrides are additive; system data remains visible/auditable.

### Roles
- Roles define required modalities/features.
- Assigning an incompatible model is rejected with specific missing constraints.
- Assigning a compatible model succeeds and persists.

### Verification
- On-demand verification runs for a selected model and stores:
  - pass/fail results
  - timestamp
  - suite version hash
- Verification does not run during standard unit tests.

## 13. Future Extensions (Informative, Not Required Now)

- Operator-managed trust mode enforcement (approval gates).
- Enterprise gateways (Kong/Apigee/LiteLLM Proxy/Helicone/Portkey) as additional gateway adapters.
- Policy-based automatic routing across role-assigned models.
- Fine-grained governance controls (PII handling, external fetch rules) at role/endpoint levels.
