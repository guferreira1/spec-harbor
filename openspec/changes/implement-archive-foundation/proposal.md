# Proposal: Implement Archive Foundation

## Problem

SpecHarbor can initialize an OpenSpec project, render role-specific implementation prompts, validate active changes, and generate blank change packages. Completed changes still remain under `openspec/changes/`, so users do not have a first-class way to move finished work into the standard archive area.

Archiving is a core product capability for OpenSpec-based development workflows. Future archive behavior may update living specs, generate changelogs, verify completion state, integrate with source control hosts, or produce AI-assisted summaries. The first implementation should establish the archive foundation without mixing CLI parsing, filesystem movement policy, future workflow integrations, or report formatting into the wrong layers.

## Goal

Implement the first archive capability:

```text
specharbor archive <change-id>
```

The command moves an active OpenSpec change from:

```text
openspec/changes/<change-id>/
```

to the current archive date directory:

```text
openspec/archive/YYYY-MM-DD/<change-id>/
```

Example:

```text
openspec/changes/implement-archive-foundation/
```

becomes:

```text
openspec/archive/2026-06-06/implement-archive-foundation/
```

The implementation must return a structured archive result and print a concise human-readable CLI report.

## Scope

- Replace the `archive` placeholder with real behavior for `specharbor archive <change-id>`.
- Validate that exactly one change id is provided.
- Reject unsupported flags.
- Reject extra positional arguments.
- Reject unsafe change ids, including absolute-path input, dot segments, path separators, drive-prefix-like input, leading-dash input, and traversal-like input.
- Obtain the project root from the current working directory in the CLI adapter.
- Validate that the project root is available.
- Derive the archive date for the CLI command from the current local date, formatted as `YYYY-MM-DD`.
- Validate the archive date value before building archive paths.
- Verify OpenSpec project availability through an archive-specific filesystem port by checking:
  - `openspec/project.md` exists as a file;
  - `openspec/changes/` exists as a directory.
- Verify that `openspec/changes/<change-id>/` exists as a directory, not merely as any filesystem path.
- Return a clear execution error when `openspec/changes/<change-id>/` is missing or exists but is not a directory.
- Create `openspec/archive/` when missing.
- Return a clear execution error when `openspec/archive` exists but is not a directory.
- Create `openspec/archive/YYYY-MM-DD/` when missing.
- Return a clear execution error when `openspec/archive/YYYY-MM-DD` exists but is not a directory.
- Move `openspec/changes/<change-id>/` to `openspec/archive/YYYY-MM-DD/<change-id>/`.
- Avoid overwriting any existing archive destination path, whether it is a file or a directory.
- Return a structured archive result containing the change id, source path, archive path, archive date, and moved directory information.
- Print a human-readable CLI report from the structured archive result.
- Keep CLI archive date derivation small and testable without adding broad or unused clock abstractions.
- Keep archive orchestration in `internal/core/usecase`.
- Keep archive concepts and result types in `internal/core/domain`.
- Add a small archive-specific filesystem port under `internal/core/ports`.
- Perform concrete filesystem behavior through `internal/adapters/filesystem`.
- Add focused tests for domain behavior, use case orchestration, filesystem adapter compatibility, CLI parsing, CLI reporting, and regressions for existing commands.

## Out of Scope

- Updating living specs automatically.
- Generating changelogs.
- Validating completion state before archive.
- Checking that tasks are fully completed.
- Checking git merge state.
- GitHub or GitLab integration.
- Release notes.
- AI-assisted archive summaries.
- Archive rollback.
- Archive metadata files.
- Validation behavior changes.
- Generation behavior changes.
- Prompt behavior changes.
- Review command implementation.
- Scan command implementation.
- Config command implementation.
- AI calls.
- Semantic Markdown parsing.
- External-agent integrations.
- Auto-fix behavior.
- Machine-readable output formats.
- User-provided archive dates or archive flags such as `--date`, `--force`, `--dry-run`, or `--metadata`.
- Updating the living architecture spec.
- Adding unused archive registries, factories, chains, provider abstractions, agent abstractions, workflow abstractions, source-control abstractions, or release abstractions.

## Success Criteria

- Running `specharbor archive <change-id>` in an initialized SpecHarbor/OpenSpec project moves the active change directory into `openspec/archive/YYYY-MM-DD/<change-id>/`.
- The archive date uses the current local date formatted as `YYYY-MM-DD`.
- `openspec/archive/` and `openspec/archive/YYYY-MM-DD/` are created when missing.
- `openspec/archive` and `openspec/archive/YYYY-MM-DD` must be directories when they already exist.
- Existing archive destination paths are not overwritten, whether they are files or directories.
- Missing change id, unsupported flags, and extra positional arguments return clear CLI errors.
- Unsafe change ids are rejected before filesystem writes or moves occur.
- Missing OpenSpec project structure is rejected before archive directory creation or movement.
- Missing source change directories and non-directory source paths are rejected before archive directory creation or movement.
- The use case returns a structured archive result with the change id, source path, archive path, archive date, and moved directory information.
- The CLI prints a concise human-readable archive report from the structured result.
- Archive follows Hexagonal Architecture: CLI parses and reports, the use case orchestrates, domain models archive concepts and results, ports define filesystem dependencies, and adapters perform concrete filesystem behavior.
- The implementation does not call AI providers, local model APIs, external agents, workflow tools, source-control host APIs, network APIs, or external processes.
- Existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, and unknown command behavior is preserved.
- `go test ./...` succeeds.
