# AGENTS.md

This repository follows OpenSpec, Hexagonal Architecture, SOLID, Clean Code, and pragmatic design patterns.

## Mandatory workflow

Before implementing a meaningful change:

1. Read `openspec/project.md`.
2. Read `openspec/specs/architecture/spec.md`.
3. Read the active change under `openspec/changes/<change-id>`.
4. Implement only the described scope.
5. Update `tasks.md`.
6. Run `go test ./...`.

## Architecture rules

- Domain code belongs in `internal/core/domain`.
- Ports belong in `internal/core/ports`.
- Use cases belong in `internal/core/usecase`.
- Concrete implementations belong in `internal/adapters`.
- Core must not import adapters.
- Use cases must depend on interfaces.
- CLI must not contain business rules.

## AI and agent rules

- Direct model APIs belong to AI provider adapters.
- Coding tools belong to agent target adapters.
- External execution tools belong to workflow connector adapters.
- Spec generation must support API-based and agent-assisted workflows.
- Agent-assisted workflows must not require provider API keys.

## Code quality rules

- Keep functions small.
- Return errors instead of panicking.
- Avoid global mutable state.
- Keep interfaces small.
- Do not hardcode secrets.
