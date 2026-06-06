# Proposal: Refactor to Hexagonal Architecture

## Problem

SpecHarbor has an architecture specification that defines Hexagonal Architecture as the target structure, but the current Go packages still use early-stage feature-oriented locations:

- `internal/ai`
- `internal/cli`
- `internal/generator`
- `internal/prompt`
- `internal/scanner`
- `internal/version`

As new spec authoring modes, validators, AI providers, coding agent targets, workflow connectors, templates, and project scanners are added, the current layout may encourage business rules to leak into CLI, provider-specific code to leak into use cases, and direct dependencies on file system, terminal, or external APIs.

## Goal

Plan an incremental migration from the current package structure toward the target Hexagonal Architecture:

- `internal/core/domain`
- `internal/core/ports`
- `internal/core/usecase`
- `internal/adapters`
- `internal/platform`

The migration should preserve the current behavior while establishing package boundaries that match the architecture specification.

## Scope

- Define how current packages should migrate to the target structure.
- Define the intended responsibilities of each target package group.
- Define an incremental migration sequence for future implementation.
- Keep CLI behavior thin and free of business rules.
- Preserve separation between AI providers, agent targets, and workflow connectors.
- Keep this change limited to OpenSpec planning artifacts.

## Out of Scope

- Implementing Go code.
- Moving Go files.
- Renaming packages.
- Changing command behavior.
- Adding provider, agent, workflow, scanner, validator, or template implementations.
- Adding architecture enforcement tooling.

## Target Direction

Current packages should migrate as follows:

| Current location | Target location | Migration intent |
| --- | --- | --- |
| `cmd/specharbor` | `cmd/specharbor` | Remain the executable entrypoint and perform only process-level bootstrapping. |
| `internal/cli` | `internal/adapters/cli` | Become the CLI adapter responsible for command parsing, terminal output, and invoking use cases. |
| `internal/scanner` | `internal/core/domain`, `internal/core/ports`, `internal/core/usecase`, `internal/adapters/scanner` | Move project context concepts into domain, scanner contracts into ports, scan orchestration into use cases, and concrete detection into scanner adapters. |
| `internal/generator` | `internal/core/domain`, `internal/core/ports`, `internal/core/usecase`, `internal/adapters/specauthor`, `internal/adapters/templates` | Move generation or authoring modes into domain, generation contracts into ports, orchestration into use cases, and concrete authoring strategies into adapters. |
| `internal/prompt` | `internal/core/domain`, `internal/core/ports`, `internal/core/usecase`, `internal/adapters/prompt` | Move agent target concepts into domain, prompt generator contracts into ports, prompt orchestration into use cases, and concrete agent prompt renderers into adapters. |
| `internal/ai` | `internal/core/domain`, `internal/core/ports`, `internal/adapters/ai` | Move provider identity concepts into domain, provider contracts into ports, and provider-specific API integrations into AI adapters. |
| `internal/version` | `internal/platform/version` | Treat version metadata as platform-level technical information used by adapters and bootstrapping code. |

## Success Criteria

- The migration plan clearly maps each existing package to the target architecture.
- The plan keeps core packages independent from adapters, CLI, terminal IO, file system access, provider SDKs, and external APIs.
- Future implementation can be completed incrementally without a large one-step rewrite.
- Agent-assisted authoring remains a no-API-key workflow.
- The change does not modify Go code or move files.
