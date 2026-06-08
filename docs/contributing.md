# Contributing

SpecHarbor uses OpenSpec changes to keep implementation work scoped and reviewable. Meaningful feature, architecture, behavior, test, CI, or documentation work should start with an OpenSpec change under `openspec/changes/<change-id>/`.

## Required Reading

Before implementing a meaningful change, contributors and coding agents should read:

- `AGENTS.md`
- `.specharbor/rules/global.md`
- the relevant role rule under `.specharbor/rules/`
- `openspec/project.md`
- `openspec/specs/architecture/spec.md`
- the active change under `openspec/changes/<change-id>/`

## OpenSpec Files

Each active change should explain the work through:

- `proposal.md`: problem, goal, scope, out-of-scope work, and success criteria.
- `design.md`: technical approach and tradeoffs.
- `tasks.md`: implementation and verification checklist.
- `acceptance-criteria.md`: observable completion criteria.
- `risks.md`: risks and mitigations.

Keep these files aligned. Update `tasks.md` only for work that was actually completed.

## Scope

Implementation must stay within the active change. Do not add unrelated commands, refactors, docs, tests, or CI changes because they are convenient.

Do not mix documentation, code, tests, and CI in one broad change unless the active OpenSpec change explicitly requires that mix.

## Architecture Boundaries

SpecHarbor follows Hexagonal Architecture:

- Domain code belongs in `internal/core/domain`.
- Ports belong in `internal/core/ports`.
- Use cases belong in `internal/core/usecase`.
- Concrete implementations belong in `internal/adapters`.
- Core must not import adapters.
- Use cases must depend on interfaces.
- CLI code must not contain business rules.

Use `openspec/specs/architecture/spec.md` as the source of detail instead of duplicating the full architecture spec here.

## Verification

Run tests after implementation:

```bash
go test ./...
```

Inspect diffs before finalizing work:

```bash
git status --short
git diff --stat
git diff --name-only
git diff
```

Check that the diff matches the active change and that `tasks.md` does not claim work that was not completed.

## Branch Hygiene

Use separate branches or worktrees for separate OpenSpec changes. Prefer one active change id per branch.

Avoid multiple agents editing the same files without explicit coordination. If parallel feature branches are merged, reconcile the docs afterward so command lists and status labels match the stabilized branch.

Before starting and before finalizing, run:

```bash
git status --short
```

Do not revert, overwrite, or clean up unrelated dirty worktree changes. Treat unrelated local changes as someone else's work unless coordination says otherwise.

Do not describe scan, config, CI, or other recently worked features as complete until the relevant branch is merged into the branch being documented and the behavior is verified there.
