---
name: wls-chatbot-structure
description: Enforce the wls-chatbot Go project layout: only main.go at the repo root, all other Go files live under internal/, and app/spine live under internal/app. Use when adding or moving Go files or adjusting entrypoint wiring.
metadata:
  short-description: Keep Go files under internal/
---

# WLS ChatBot Project Structure

## Canonical layout

- Root contains only `main.go` (entrypoint + Wails app wiring).
- All other Go code lives under `internal/`.
- App + spine types live under `internal/app/`.

## When adding or moving Go files

1. Choose/create an `internal/<domain>` folder.
2. Set the `package` name to match the folder.
3. Update imports and call sites in `main.go` and other packages.
4. If `main.go` needs to reference methods, ensure they are exported (e.g., `Startup`).

## Guardrail

- Do not add new `.go` files at the repo root.
