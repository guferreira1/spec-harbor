# Proposal: Refactor to Hexagonal Architecture

## Problem

SpecHarbor has an architecture specification that defines Hexagonal Architecture as the target structure, but the project started with early-stage feature-oriented package locations:

- `internal/ai`
- `internal/cli`
- `internal/generator`
- `internal/prompt`
- `internal/scanner`
- `internal/version`

As new spec authoring modes, validators, AI providers, coding agent targets, workflow connectors, templates, and project scanners are added, the current layout may encourage business rules to leak into CLI, provider-specific code to leak into use cases, and direct dependencies on file system, terminal, or external APIs.

## Goal

Implement the first narrow migration slice from the early package structure toward the target Hexagonal Architecture:

- `internal/core/domain`
- `internal/core/ports`
- `internal/core/usecase`
- `internal/adapters`
- `internal/platform`

This change reorganizes existing packages where their current contents already have a clear architectural home. It preserves current CLI behavior, updates package declarations and imports, and avoids introducing new behavior or placeholder abstractions.

## Scope

- Move existing pure value definitions from `internal/scanner`, `internal/generator`, `internal/prompt`, and `internal/ai` into `internal/core/domain`.
- Move version metadata from `internal/version` into `internal/platform/version`.
- Move the CLI delivery package from `internal/cli` into `internal/adapters/cli`.
- Update package declarations for moved files.
- Update imports that reference moved packages.
- Preserve existing CLI behavior for `help`, `version`, placeholder commands, and unknown commands.
- Keep the change limited to package reorganization, package declaration updates, and import updates.
- Preserve separation between AI providers, agent targets, and workflow connectors without adding new implementations.

## Out of Scope

- Adding new CLI commands or command behavior.
- Adding provider, agent, workflow, scanner, validator, or template implementations.
- Adding OpenSpec file operations or config storage.
- Adding `internal/core/ports` interfaces before use cases need them.
- Adding `internal/core/usecase` orchestration while commands only provide CLI parsing and placeholder output.
- Adding placeholder abstractions, compatibility wrappers, or empty packages only to satisfy the target architecture.
- Adding architecture enforcement tooling.

## Implemented Reorganization

This change implements the following package moves:

| Current location | Target location | Migration intent |
| --- | --- | --- |
| `cmd/specharbor` | `cmd/specharbor` | Remain the executable entrypoint and perform only process-level bootstrapping. |
| `internal/cli` | `internal/adapters/cli` | Become the CLI adapter responsible for command parsing, terminal output, and invoking future use cases. |
| `internal/scanner/project.go` | `internal/core/domain/project_context.go` | Move `ProjectContext` into the domain as a pure project context value. |
| `internal/generator/mode.go` | `internal/core/domain/generation_mode.go` | Move `GenerationMode` and its existing constants into the domain. |
| `internal/prompt/agent.go` | `internal/core/domain/agent.go` | Move `Agent` and its existing constants into the domain. |
| `internal/ai/provider.go` | `internal/core/domain/ai_provider.go` | Move `Provider` and its existing constants into the domain. |
| `internal/version` | `internal/platform/version` | Treat version metadata as platform-level technical information used by adapters and bootstrapping code. |

This change does not introduce `internal/core/ports`, `internal/core/usecase`, scanner adapters, spec author adapters, prompt adapters beyond the existing CLI delivery package, or AI adapters. Those packages should be added only when concrete behavior requires them.

## Success Criteria

- The package reorganization maps each moved file to the target architecture responsibility described above.
- Package declarations and imports are updated for all moved files.
- Existing CLI behavior is preserved.
- The core domain remains independent from adapters, CLI, terminal IO, file system access, provider SDKs, and external APIs.
- Agent-assisted authoring remains a no-API-key workflow.
- The change does not add new behavior, new providers, new agents, new scanners, new validators, new templates, new workflow connectors, or OpenSpec file operations.
- The change does not add placeholder ports, use cases, adapters, wrappers, or empty packages.
