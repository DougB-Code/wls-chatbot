# AGENTS.md

Wails Lit Starter ChatBot is a foundation project with LLM provider integration and a basic chat interface, built using Wails and Lit.

## Your North Star

- You're writing production grade software. This is not a weekend hackathon. Plan and code accordingly. 
- DO  NOT be sycophantic. Push back HARD when directions don't align with industry best practives, this AGENTS.md file, or any of the Agent Skills. 
- If you do not know an answer, it's OK to speculate, but you must ALWAYS tell us when you are speculating. 
- Be explicit over implicit.
- Report on all deviations from this AGENTS.md file or any Agent Skill you are currently using. 

## Planning code changes

- ALWAYS preview the existing project standards and conventions, both in this document plus in code, and make sure your implementation plan aligns with project standards and conventions.

- ALWAYS identify the boundary context of your change(s), clearly articulating where they fit into the existing application structure BEFORE making a change.

- NEVER add backwards compatibiity support unless explicitly asked. 

- NEVER add add competing / duplicate implementations.

- ALWAYS adhere to SOLID principles. The Single Responsibility Principle (SRP) is the most important principle to adhere to.

- NEVER mock or stub implementation code (I/O in test code is OK) unless asked.

- ALWAYS identify and fill in gaps in the directions to ensure a sound implementation plan.

- ALWAYS default to being restrictive (ex: the sensible default is for no access to be enabled).

## Code architecture

- All code MUST be co-located and organized together, according to natural boundary contexts.

- All IO code MUST be isolated into its own module and be easily mockable.

- All functions, files, classes, and modules *ABSOLUTELY MUST* follow the single responsibility principle and NEVER mix responsibilities.

## Code comments 

- ALWAYS add the relative path (from the project root) to the top of each file.
- ALWAYS document each folder/package/module/class/file/function with its **single** responsibility as the doc title.

```go-example
// config.go loads application configuration files
// internal/config/config.go
//
// LoadProvidersFromRoot loads providers from the YAML config in the project root.
func LoadProvidersFromRoot() {}
``` 

```ts-example
/**
 * AiGateway.ts renders provider connection settings and emit provider actions.
 */
```

- NEVER add code comments indicating where you made changes.

## Go Backend Architecture

```
internal/
├── app/           # Application facades (thin pass-through to features)
├── core/          # Cross-cutting concerns (config, datastore, events, logger)
├── features/      # Feature-first clean architecture
│   └── ai/
│       ├── chat/        # Conversations, messages, streaming
│       ├── image/       # Image generation/editing
│       ├── model/       # Model catalog
│       └── providers/   # Provider adapters (OpenAI, Gemini, Grok, etc.)
└── ui/
    └── adapters/
        ├── wails/   # Frontend bridge (Wails bindings)
        └── cli/     # CLI commands
```

**Data flow:** `UI Adapters` → `App Facades` → `Feature Layer` → `Provider Adapters`

**Key principles:**
- Features own their domain types—no duplicate DTOs across layers
- App facades are pass-through; they return feature types directly
- Provider adapters implement a common interface (`Provider`) for chat/image operations

## Go conventions:
- Add a single line space after the function signature if the first line is a comment.

## TypeScript conventions:
- Use `camelCase` for variables, functions, params, properties.
- Use `PascalCase` for classes, interfaces, type aliases, enums.
- Enum members: `PascalCase` or `UPPER_CASE` (pick one per enum).
- `const` can be `camelCase` or `UPPER_CASE`.
- Leading `_` allowed for intentionally unused params or private-ish fields; trailing `_` disallowed.

## Wails conventions
- Use generated bindings from `frontend/wailsjs` for backend calls and logging.
- Do not access `window.go` directly or add global `wails.d.ts` bridge typings.
