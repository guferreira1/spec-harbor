# Proposal: Implement Init Command

## Problem

`specharbor init` is currently a placeholder. Users need a first real CLI workflow that can prepare an existing project for SpecHarbor and OpenSpec without manual directory setup.

The command must initialize only the expected project structure, preserve existing user files, and keep initialization behavior outside the CLI adapter so the implementation follows the repository architecture rules.

## Goal

Implement `specharbor init` so it initializes the SpecHarbor/OpenSpec structure in the current working directory.

The command must create:

- `openspec/project.md`
- `openspec/specs/`
- `openspec/changes/`
- `.specharbor/config.yml`
- `.specharbor/rules/`
- default rule files under `.specharbor/rules/`

## Scope

- Replace the placeholder `init` command with real initialization behavior.
- Add a project initialization use case under `internal/core/usecase`.
- Add the small ports needed by that use case under `internal/core/ports`.
- Add concrete filesystem and default-template adapters under `internal/adapters`.
- Wire the CLI adapter to the use case.
- Generate default `openspec/project.md`, `.specharbor/config.yml`, and rule files.
- Do not overwrite existing files by default.
- Return a clear success message for first initialization.
- Return a clear success message when the project is already initialized.
- Add focused tests for the use case, adapters, and CLI behavior.

## Out of Scope

- Flags such as `--force`, `--template`, `--path`, or `--dry-run`.
- Interactive prompts.
- AI-assisted or agent-assisted spec generation.
- OpenSpec change generation.
- Validation of existing OpenSpec content.
- Migration or repair of malformed existing files.
- Archive, review, scan, prompt, generate, config, or validate command behavior.
- Moving unrelated production files.

## Success Criteria

- Running `specharbor init` in an empty directory creates the required structure.
- Running `specharbor init` again does not overwrite files and prints a clear already-initialized message.
- Running `specharbor init` in a partially initialized directory creates missing files and directories while preserving existing files.
- The implementation follows Hexagonal Architecture: use case orchestration is in core, filesystem and template details are adapters, and CLI only parses input and formats output.
- `go test ./...` succeeds.
