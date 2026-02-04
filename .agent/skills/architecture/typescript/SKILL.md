---
name: architecture-typescript
description: Apply the frontend TypeScript architecture for this repo. Use when placing frontend files, shaping feature boundaries, and enforcing dependency direction.
metadata:
  short-description: Feature-first frontend architecture
---

# Frontend TypeScript Architecture (Feature-First, Clean Boundaries)

## Boundary context

The frontend should mirror the backend's **feature-first architecture** while using frontend-native layers.

- Primary boundary: `frontend/src/features/<feature>/...`
- Shared cross-feature boundary: `frontend/src/core` (or existing shared folders)
- Composition root: `frontend/src/main.ts` and `frontend/src/AppShell.ts`

## Canonical frontend topology

For each feature (for example `chat`, `settings`), organize by responsibility:

- `ui/`
  - Lit components and view rendering only.
- `application/`
  - Use-case orchestration and user-intent flows.
- `state/`
  - Feature-owned signals/state transitions.
- `domain/`
  - Feature types, pure rules, and selectors with no transport/framework side effects.
- `infrastructure/`
  - Wails bindings wrappers, event bridge wiring, persistence/IO adapters.

Shared/non-feature code:

- `app/` (or existing `main.ts` + `AppShell.ts`)
  - App bootstrap and top-level composition.
- `shell/`
  - Layout scaffolding shared across features.
- `components/`
  - Reusable UI components that are truly cross-feature.
- `styles/`
  - Global tokens/base styles only.
- `wailsjs/`
  - Generated bindings (read-only; generated source of backend API calls).

## Mapping from current folders

Current structure is partially feature-first and partially global:

- `frontend/src/features/*` -> mostly `ui` layer today.
- `frontend/src/policy/*` -> mostly `application` layer today.
- `frontend/src/store/*` -> mostly `state` layer today.
- `frontend/src/transport/*` -> mostly `infrastructure` layer today.

Target direction:

1. Keep old global folders working during migration.
2. Place **new feature logic inside the owning feature** first.
3. Shrink global `policy/store/transport` over time by moving code into `features/<feature>/...`.

## Dependency direction (enforced)

Inside a feature, dependency flow should point inward:

1. `ui` -> may import `application`, `state`, `domain`.
2. `application` -> may import `domain`, `state`, and feature-defined infrastructure ports/contracts.
3. `state` -> may import `domain` only.
4. `domain` -> pure TS logic/types; no transport imports.
5. `infrastructure` -> may import feature contracts and external APIs (`wailsjs`, runtime events, fetch, etc.).

Forbidden:

1. `domain` importing `infrastructure` or Lit UI.
2. `state` importing `wailsjs` directly.
3. `ui` components calling generated Wails bindings directly.
4. Cross-feature imports that bypass application contracts and create feature-to-feature coupling.

## Wails boundary rules

- Always call backend using generated bindings from `frontend/wailsjs/go/...`.
- Never access `window.go` directly.
- Keep direct Wails binding calls in `infrastructure` adapters, not in UI components.
- Surface backend operations through feature application functions.

## Naming and file conventions

- Follow project TypeScript conventions:
  - `camelCase`: variables/functions/params/properties.
  - `PascalCase`: classes/interfaces/type aliases/enums.
- Use explicit file names by responsibility:
  - UI: `ChatView.ts`, `MessageBubble.ts`
  - Application: `sendMessage.ts`, `chatPolicy.ts` (temporary during migration)
  - State: `chatSignals.ts`
  - Infrastructure: `chatTransport.ts`, `wailsEvents.ts`
  - Domain: `conversation.ts`, `message.ts`, `modelSelection.ts`
- Avoid catch-all files (`utils.ts`, `helpers.ts`) unless scope is tightly bounded to a feature layer.

## Change checklist for frontend TypeScript work

1. Identify the owning feature first (`chat`, `settings`, or a new feature).
2. Place rendering logic in `ui` and keep side effects out of components.
3. Place business/application flow in `application`.
4. Place backend/event IO in `infrastructure`.
5. Place state transitions in `state`, with pure derivations in `domain`.
6. Keep shared code in `components/shell/styles/core` only when genuinely cross-feature.

## Architecture smell checks

- Lit component imports `../../wailsjs/go/...` directly.
- Feature behavior implemented in global folders when a feature boundary already exists.
- A state module performing transport calls.
- A single file mixing UI rendering, event wiring, and backend API calls.
- New shared folder created for code used by only one feature.
