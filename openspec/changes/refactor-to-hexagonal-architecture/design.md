# Design: Refactor to Hexagonal Architecture

## Overview

This change defines the migration path from the current early package structure to the architecture described in `openspec/specs/architecture/spec.md`.

The target design keeps stable business concepts and orchestration in `internal/core`, concrete external integrations in `internal/adapters`, and cross-cutting technical helpers in `internal/platform`.

## Target Package Responsibilities

### `internal/core/domain`

Contains pure SpecHarbor concepts and value objects.

Candidate concepts from current packages:

- `ProjectContext` from `internal/scanner`
- generation or authoring mode values from `internal/generator`
- agent target values from `internal/prompt`
- AI provider identity values from `internal/ai`
- future concepts such as OpenSpec changes, specs, validation findings, templates, workflow targets, and configuration values

Rules:

- No file system, terminal, network, environment, or provider SDK access.
- No dependency on adapters, CLI, or platform helpers that hide business behavior.
- No business behavior that depends on external execution.

### `internal/core/ports`

Contains small interfaces required by use cases.

Candidate ports:

- `ProjectScanner`
- `SpecAuthor`
- `SpecDraftProvider`
- `PromptDraftProvider`
- `AgentPromptGenerator`
- `SpecValidator`
- `TemplateRenderer`
- `ConfigRepository`
- `FileSystem`
- `WorkflowDispatcher`

Rules:

- Ports are owned by the use cases that consume them.
- Interfaces stay small and behavior-specific.
- Provider, scanner, prompt, validator, file system, and workflow implementation details must not appear in port contracts.

### `internal/core/usecase`

Contains application orchestration.

Candidate use cases:

- initialize SpecHarbor/OpenSpec project structure
- scan project context
- generate OpenSpec change
- generate agent-specific prompt
- validate OpenSpec change
- review implementation against a spec
- archive completed change
- manage configuration

Rules:

- Use cases may depend on `internal/core/domain` and `internal/core/ports`.
- Use cases must return structured results and errors.
- Use cases must not print to terminal, parse CLI arguments, call `os`, call HTTP clients, use provider SDKs, or render provider-specific payloads.

### `internal/adapters`

Contains concrete implementations of core ports and external delivery mechanisms.

Target adapter groups:

- `internal/adapters/cli`
- `internal/adapters/filesystem`
- `internal/adapters/scanner`
- `internal/adapters/specauthor`
- `internal/adapters/prompt`
- `internal/adapters/ai`
- `internal/adapters/workflow`
- `internal/adapters/validator`
- `internal/adapters/config`
- `internal/adapters/templates`

Rules:

- Adapters may depend on core domain and ports.
- Adapters must not define business rules.
- Provider-specific payloads, authentication, retries, rate limits, and response mapping stay inside provider adapters.
- Agent-specific prompt formats stay inside prompt adapters.

### `internal/platform`

Contains minimal cross-cutting technical utilities.

Candidate platform packages:

- `internal/platform/version`
- `internal/platform/logger`
- `internal/platform/errors`

Rules:

- Platform packages must stay technical and small.
- Platform must not become a place for business rules.
- Avoid global mutable state.

## Current Package Migration Plan

### `cmd/specharbor`

Keep `cmd/specharbor` as the process entrypoint.

Future implementation should use it only for:

- process startup
- wiring adapters to use cases
- passing command-line arguments into the CLI adapter
- translating top-level errors into process exit codes

It must not contain business rules.

### `internal/cli`

Move to `internal/adapters/cli`.

The CLI adapter should:

- parse commands and flags
- call use cases
- format command output
- map structured errors to user-facing messages

Business rules currently added to CLI commands during future development should instead move to `internal/core/usecase`.

### `internal/scanner`

Split by responsibility:

- `ProjectContext` becomes a domain value in `internal/core/domain`.
- Scanner interfaces move to `internal/core/ports`.
- Scan orchestration moves to `internal/core/usecase`.
- Concrete project file detection moves to `internal/adapters/scanner`.

This keeps project-scanning behavior available to use cases without making the core depend on file system details.

### `internal/generator`

Split by responsibility:

- Generation or authoring mode values move to `internal/core/domain`.
- Spec generation and authoring contracts move to `internal/core/ports`.
- Change generation orchestration moves to `internal/core/usecase`.
- Blank, guided, template, AI-provider, agent-assisted, and hybrid authoring implementations move to strategy adapters under `internal/adapters/specauthor`.
- Template rendering implementation moves to `internal/adapters/templates`.

Use cases should resolve authoring behavior through ports and injected registries or factories rather than provider-specific conditionals.

### `internal/prompt`

Split by responsibility:

- Agent target values move to `internal/core/domain`.
- Prompt generation contracts move to `internal/core/ports`.
- Prompt generation orchestration moves to `internal/core/usecase`.
- Codex, Claude Code, Cursor, generic, and future agent prompt renderers move to `internal/adapters/prompt`.

Agent-assisted authoring must remain independent from AI provider API keys.

### `internal/ai`

Split by responsibility:

- Provider identity values move to `internal/core/domain`.
- AI drafting contracts move to `internal/core/ports`.
- Concrete provider integrations move to `internal/adapters/ai`.

Provider-specific request formats, authentication, retry behavior, rate limits, error mapping, and response parsing must stay inside AI adapters.

### `internal/version`

Move to `internal/platform/version`.

Version information is technical metadata. It may be read by CLI adapters or bootstrapping code, but it must not become part of core business orchestration unless a future use case explicitly requires structured version reporting.

## Incremental Sequence

Future implementation should proceed in small steps:

1. Introduce the target directories without changing behavior.
2. Move pure value objects into `internal/core/domain`.
3. Add small ports owned by use cases as behavior is migrated.
4. Move orchestration out of CLI and feature packages into `internal/core/usecase`.
5. Move concrete implementations into matching adapter packages.
6. Update `cmd/specharbor` to wire use cases and adapters.
7. Add focused tests for domain, use cases, adapters, and CLI behavior.
8. Add architecture validation once package boundaries stabilize.

## Dependency Direction

The intended dependency direction is:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

Forbidden dependencies remain:

```text
core -> adapters
core -> cmd
usecase -> concrete provider
usecase -> os package directly
usecase -> HTTP client directly
usecase -> terminal directly
```

## Testing Strategy

The migration should preserve or add tests according to package responsibility:

- Domain tests should avoid mocks where possible.
- Use case tests should use fake or mocked ports.
- Adapter tests should focus on integration behavior and external mapping.
- CLI tests should verify command input/output and error presentation.
- Architecture tests or validators should be added after the boundary migration is stable.
