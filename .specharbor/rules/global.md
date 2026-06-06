# Global Rules

These rules apply to every agent working in this repository.

## Required reading

Before acting on a meaningful change, read:

- `AGENTS.md`
- `openspec/project.md`
- `openspec/specs/architecture/spec.md`
- the active change under `openspec/changes/<change-id>/`

## Scope control

- Implement only the described scope.
- Do not introduce unrelated refactors.
- Do not add new behavior unless the active change requires it.
- Do not create placeholder abstractions only to satisfy a target architecture.
- Keep changes small, explicit, and reviewable.

## OpenSpec workflow

- Create or update OpenSpec before meaningful implementation work.
- Keep `tasks.md` executable by another agent.
- Update `tasks.md` only for work that was actually completed.
- Keep proposal, design, tasks, acceptance criteria, and risks aligned.

## Validation

- Run validation commands requested by the active change.
- For Go code changes, run `gofmt` and `go test ./...`.
- Report commands that could not be executed.

## Security

- Do not hardcode secrets, tokens, API keys, provider credentials, or local paths.
- Prefer environment variables or ignored local configuration for credentials.
