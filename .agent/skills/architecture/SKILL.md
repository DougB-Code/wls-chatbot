## Application Architecture

Clean Architecture + SOLID maps cleanly to **role-based folder rules** (direction by “policy vs details”)

---

## A folder taxonomy that encodes Clean Architecture

* `main.go`: Composition root entrypoint(s). Knows everything.

* `internal/app/...`: Wiring/assembly. Reads config, builds concrete implementations, connects them.

* `internal/core/domain/...`: Entities + domain rules (most stable).

* `internal/core/usecase/...`: Application services / orchestrations.

* `internal/core/ports/...`: Interfaces the core needs (repositories, clocks, loggers, publishers, etc.). **Defined by core.**

* `internal/adapters/...`: Implementations of ports + delivery mechanisms:
  * `internal/adapters/http/...`
  * `internal/adapters/db/...`
  * `internal/adapters/fs/...`
  * `internal/adapters/cli/...`

* `internal/platform/...`: Shared technical utilities that are not “policy” (small helpers, primitives). Keep it boring.

---

## File naming conventions (canonical)

### Global rules

* Filenames are lowercase `snake_case`.
* One primary responsibility per file; name the file for the primary type or responsibility.
* Avoid generic catch-alls like `common.go`, `utils.go`, or `helpers.go`.
* Prefer explicit, role-based names over vague abstractions (`client.go`, `adapter.go`, `repository.go`, `cache.go`, `mapper.go`).

### Core domain (`internal/core/domain/...`)

* Entity/value files named after the type they define.
  * Example: `conversation.go`, `message.go`, `model.go`.
* Shared enums and small structs for the package go in `types.go`.
* Package errors go in `errors.go`.

### Core ports (`internal/core/ports/...`)

* One port per file; filename matches the port role.
  * Examples: `chat_repository.go`, `provider.go`, `clock.go`, `logger.go`, `events.go`.
* Avoid `interfaces.go` or `ports.go` that collects unrelated interfaces.

### Core use cases (`internal/core/usecase/...`)

* Each use case package has **one** primary coordinator type named `Service`.
  * File name: `service.go`.
* If a package needs more than one distinct coordinator, name each explicitly and use `<name>_service.go`.
* Do **not** create both `Service` and `Orchestrator` for the same responsibility. If both exist, their responsibilities must be distinct and non-overlapping.
* Supporting logic can be split into focused files: `stream_manager.go`, `validation.go`, `builder.go`, `policy.go`.

### Adapters (`internal/adapters/...`)

* Adapter entrypoints use role-based names:
  * External API wrappers: `client.go` or `<vendor>_client.go`.
  * Port implementations: `<port>_adapter.go` or `<vendor>_provider.go` (if it implements `ports.Provider`).
  * Persistence: `repository.go` or `<entity>_repository.go`.
  * Caches: `cache_fs.go`, `cache_redis.go` (explicit storage type).
* Avoid `service.go` in adapters to prevent confusion with use case `Service`.

### App wiring (`internal/app/...`)

* Wiring modules: `wiring.go` or `<feature>_wiring.go`.
* Composition roots for a subsystem: `app.go`, `providers.go`, `transport.go`.
* Config stays pure: `config.go`, `paths.go`, `save.go`.

---

## Use case file roles

### `service.go`

**Purpose:** the primary application API for the use case.

* Encapsulates the core application rules for a single capability.
* Operates on domain entities and depends only on `core/ports`.
* Owns invariants and state transitions for that use case.
* Does **not** contain transport, persistence, or framework logic.

### `orchestration.go`

**Purpose:** coordinates multi-step workflows that compose multiple services or long-running processes.

* Calls one or more use case services and sequences their work.
* Handles cross-cutting concerns like event emission, streaming lifecycles, or sagas.
* Depends only on `core/ports`, never on adapters.
* Must **not** duplicate the service API; it should only add coordination behavior.

**Rule:** `orchestration.go` is optional. Use it only when there is meaningful coordination logic that doesn’t belong in `service.go`.

---

## The import rules (the “folder law”)

Think of it as an allowlist. Everything else is forbidden.

### Allowed dependency direction

* `main.go` → can import anything (entrypoints)
* `internal/app` → can import anything (wiring)
* `internal/adapters/*` → may import:
  * `internal/core/*` (to call use cases / implement ports)
  * `internal/platform/*` (small utilities)
  * stdlib / third-party frameworks (HTTP, SQL, UI libs, etc.)
* `internal/core/usecase` → may import:
  * `internal/core/domain`
  * `internal/core/ports`
  * stdlib (limited: `context`, `time`, `errors`, etc.)
* `internal/core/domain` → may import:
  * stdlib only (no IO frameworks)
* `internal/core/ports` → may import:
  * `internal/core/domain` (if needed for types)
  * stdlib only
* `internal/config` → stdlib only (and maybe a config parsing lib)

### Forbidden (the big ones)
* `internal/core/*` must **never** import `internal/adapters/*`
* `internal/core/*` must **never** import `internal/config`
* `internal/adapters/*` should not import each other (prefer composition to connect adapters)
* `internal/config` should not import `internal/adapters/*` or `internal/core/*` (keep it pure parsing)

This encodes the Clean Architecture “dependencies point inward” rule.

---

## Mapping SOLID → folder rules

### S — Single Responsibility Principle

**Folder rule:** each package/folder has one reason to change.

* `core/domain`: changes when business rules change
* `core/usecase`: changes when orchestration / application rules change
* `adapters/http`: changes when HTTP concerns change
* `adapters/db`: changes when storage changes
* `config`: changes when config formats/sources change

**Smell:** “god packages” like `internal/common` that everyone imports and that changes weekly.

---

### O — Open/Closed Principle

**Folder rule:** add new behavior by adding new adapters, not by editing core.

* Want gRPC in addition to HTTP? Add `internal/adapters/grpc`.
* Want Postgres in addition to SQLite? Add `internal/adapters/db/postgres`.

Core stays closed to infrastructure churn because it speaks through `core/ports`.

**Smell:** adding a new DB requires edits all over `core/usecase`.

---

### L — Liskov Substitution Principle

**Folder rule:** ports define behavioral contracts; adapters must be substitutable.

* `core/ports/UserRepo` has clear semantics (transactions, idempotency, error mapping).
* Adapters implement the same semantics across SQLite/Postgres/mock.

**Practical enforcement:**

* Contract tests per port (run the same test suite against each adapter implementation).
* Avoid adapter-only “special cases” that leak upward.

**Smell:** the use case has `if repoIsSQLite { ... }`.

---

### I — Interface Segregation Principle

**Folder rule:** ports stay small and specific; don’t create “mega ports”.

* Prefer:

  * `UserReader`, `UserWriter`
  * `TokenSigner`, `TokenVerifier`
* Over:

  * `UserRepositoryEverything`

A good heuristic: put each port in its own file, and keep it narrow enough that most adapters implement only what they need.

**Smell:** many adapters implement methods they don’t use or can’t implement correctly.

---

### D — Dependency Inversion Principle

**Folder rule:** interfaces live inward; implementations live outward.

* `core/ports` defines `Clock`, `Logger`, `Repo`, `Publisher`
* `adapters/*` implements them
* `app` wires them

This is the primary reason the folder rules work.

**Smell:** `core/usecase` imports an HTTP client, SQL driver, filesystem, or the config loader.

---

## How cross-cutting concerns fit without breaking rules

### Config values needed by adapters (HTTP timeouts, DB paths, feature flags)

**Rule:** adapters don’t import `config`; they receive values via constructors.

* `internal/config` parses config into a struct.
* `internal/app` maps config → adapter options.

Example shape:

* `adapters/http` defines `ServerConfig` (only what it needs).
* `app` converts `AppConfig.HTTP` → `ServerConfig`.

This keeps `adapters/http` reusable and testable.

### Cross-cutting services (logging/metrics/tracing/time/IDs)

**Rule:** define the interface in `core/ports`, inject it everywhere.

Adapters can use real implementations; tests can use fakes.

---

## Two viable “folder law” variants (trade-offs)

### Variant A: Strict role-based boundaries (recommended)

* Core is pure policy.
* Adapters are details.
* App/cmd wires.

**Pros:** maximum testability, minimal drift, clean substitution story
**Cons:** more explicit wiring code

### Variant B: “Service locator” / global runtime (not recommended except tiny apps)

* Everyone imports `internal/runtime` for config/log/etc.

**Pros:** less wiring
**Cons:** hidden dependencies, harder tests, encourages coupling, boundary decay

If you’re writing generally applicable guidance, endorse A and explicitly caution against B.
