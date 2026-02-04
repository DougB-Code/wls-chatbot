---
name: architecture-go
description: Apply the Go clean architecture used in this repo. Use for Go file placement, package boundaries, and dependency direction decisions.
metadata:
  short-description: Feature-first clean architecture for Go
---

# Go Architecture (Feature-First Clean Architecture)

## Boundary context

This repository uses **clean architecture aligned to product features**, not a single global `core/domain` + `core/usecase` tree.

- Primary boundary: `internal/features/<feature>/...`
- Shared cross-feature boundary: `internal/core/...`
- Composition root: `main.go`

## Canonical folder topology

- `main.go`
  - App entrypoint and top-level dependency wiring.
- `internal/core/ports`
  - Cross-feature ports used by multiple features (for example event emitters, logger).
- `internal/core/adapters`
  - Technical/framework adapters shared across features (for example Wails bridge, datastore, logger adapter).
- `internal/features/chat`
  - `domain`: chat entities and domain behavior.
  - `ports`: chat feature interfaces (for example `ChatRepository`).
  - `usecase`: chat policy layer (`Service`, `Orchestrator`, focused helpers).
  - `adapters`: chat feature implementations of ports (for example sqlite repository).
- `internal/features/settings`
  - `config`: settings config model + load/save/path/default behavior.
  - `ports`: settings/provider feature interfaces.
  - `usecase`: provider policy layer (`Service`, `Orchestrator`, focused helpers).
  - `adapters`: provider/secret/config/cache implementations.
  - `wiring`: feature-local assembly from config + adapters into use cases.

## Layer responsibilities inside a feature

### `domain`

- Owns feature entities, value types, and invariants.
- Must remain framework-agnostic.
- Should not import feature adapters.

### `ports`

- Defines interfaces required by the feature use cases.
- Port names should reflect behavior (`ChatRepository`, `ProviderRegistry`, `SecretStore`), not technology.

### `usecase`

- Owns application policy for the feature.
- `service.go`: core feature operations and state transitions.
- `orchestration.go`: multi-step coordination, cross-port flow, event emission.
- Helper files stay narrow (`stream_manager.go`, `types.go`, etc.).

### `adapters`

- Implements feature ports using concrete technologies.
- Handles I/O, framework APIs, external SDKs, and persistence details.

### `config` and `wiring` (when needed)

- `config`: parse/persist config; no feature business policy.
- `wiring`: construct concrete feature dependencies and return use case-ready objects.

## Dependency direction (enforced)

Allowed (high level):

1. `main.go` -> may import all internal packages needed for composition.
2. `internal/core/adapters/*` -> may import `internal/core/ports` and feature usecases/ports when acting as a delivery adapter.
3. `internal/features/<feature>/usecase` -> may import:
   - same feature `domain`
   - same feature `ports`
   - `internal/core/ports` for shared concerns (logging/events)
4. `internal/features/<feature>/adapters` -> may import:
   - same feature `ports` and `domain` (as needed)
   - shared `internal/core/ports` when implementing shared contracts.

Forbidden:

1. Feature `domain` importing feature `adapters`.
2. Feature `ports` importing feature `adapters`.
3. `internal/core/ports` importing `internal/core/adapters` or feature adapters.
4. Adapter-to-adapter dependency chains across unrelated features. Prefer wiring/composition.

## Naming and file rules

- Use lowercase `snake_case` filenames.
- Keep one primary responsibility per file.
- Keep `service.go` for the main use case service and `orchestration.go` for workflow coordination when both are needed.
- Keep ports split by responsibility (avoid catch-all `interfaces.go`).
- Prefer explicit adapter names by role/technology (`sqlite.go`, `keyring.go`, `registry.go`, `http_client.go`).

## Change checklist for new Go code

1. Pick the owning feature first (`chat`, `settings`, or a new feature under `internal/features`).
2. Place business rules in `domain`/`usecase`, never in adapters.
3. Define or extend ports before adding adapter behavior.
4. Add adapter implementation in the same feature unless it is truly cross-feature and belongs in `internal/core/adapters`.
5. Wire dependencies in `main.go` or feature-local `wiring` package.

## Architecture smell checks

- A feature use case directly calling SQL/HTTP/SDK code.
- New shared code added to `internal/core` that is actually feature-specific.
- `orchestration.go` duplicating business logic already in `service.go`.
- A package with mixed responsibilities (policy + I/O in the same file).
