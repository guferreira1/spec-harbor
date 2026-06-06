# Tasks: Implement Archive Foundation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-archive-foundation/`.
- [x] Inspect the current CLI command registry and error flow before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing domain, use case, port, filesystem adapter, and test patterns before editing.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to `specharbor archive <change-id>` and the archive foundation described by this change.
- [x] Do not implement living spec updates, changelog generation, completion checks, task completion checks, git merge checks, GitHub/GitLab integrations, release notes, AI-assisted archive summaries, rollback, metadata files, review, scan, config, validation changes, generation changes, prompt changes, external-agent integrations, or auto-fix behavior.

## Phase 1: Domain Concepts

- [x] Add archive result concepts under `internal/core/domain`.
- [x] Represent the requested change id, source path, archive path, archive date, and moved directory information in the structured result.
- [x] Add small moved directory concepts if they keep archive reporting clear.
- [x] Add helpers only where they keep result creation or defensive copying small and explicit.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, and concrete filesystem packages.

## Phase 2: Ports

- [x] Add a small archive-specific filesystem port under `internal/core/ports`.
- [x] Include only operations needed by this change: directory existence, file existence, path existence, directory creation, and directory movement.
- [x] Keep the archive filesystem port separate from initialization, validation, generation, prompt rendering, AI provider, workflow dispatcher, source-control, release, and review contracts.
- [x] Ensure the archive use case depends on the archive filesystem port instead of `internal/adapters/filesystem`.

## Phase 3: Archive Use Case

- [x] Add the archive use case under `internal/core/usecase`.
- [x] Make the use case accept project root, change id, and archive date as input.
- [x] Validate that project root is non-empty after trimming whitespace.
- [x] Validate that change id is non-empty after trimming whitespace.
- [x] Validate that archive date is non-empty and formatted as `YYYY-MM-DD`.
- [x] Reject unsafe change ids before performing filesystem writes or moves.
- [x] Build the source relative path as `openspec/changes/<change-id>`.
- [x] Build the archive date directory as `openspec/archive/<archive-date>`.
- [x] Build the archive relative path as `openspec/archive/<archive-date>/<change-id>`.
- [x] Check OpenSpec project availability by verifying that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory through the archive filesystem port.
- [x] Return a clear execution error telling the user to run `specharbor init` first when OpenSpec project availability cannot be verified.
- [x] Verify that the source path `openspec/changes/<change-id>/` exists as a directory, not merely as any filesystem path.
- [x] Return a clear execution error when `openspec/changes/<change-id>/` is missing or exists but is not a directory.
- [x] Return a clear execution error when `openspec/archive` exists but is not a directory.
- [x] Return a clear execution error when `openspec/archive/<archive-date>` exists but is not a directory.
- [x] Check whether `openspec/archive/<archive-date>/<change-id>` already exists as a file or directory before moving.
- [x] Return a clear execution error when the archive destination path already exists.
- [x] Do not overwrite existing archived content, whether it is a file or a directory.
- [x] Create `openspec/archive/` when missing.
- [x] Create `openspec/archive/<archive-date>/` when missing.
- [x] Move the source change directory to the archive path through the archive filesystem port.
- [x] Return a structured archive result.
- [x] Return errors for dependency failures, invalid input, filesystem check errors, directory creation errors, and move errors.
- [x] Do not print from the use case.
- [x] Do not call `os`, terminal IO, provider SDKs, network APIs, source-control APIs, external agents, external processes, or workflow tools from the use case.
- [x] Keep the structure ready for future archive collaborators without adding unused exported strategy, factory, registry, chain, provider, source-control, workflow, release, AI, or rollback abstractions.

## Phase 4: Filesystem Adapter

- [x] Use `internal/adapters/filesystem` as the concrete implementation of the archive filesystem port.
- [x] Add only compatibility code needed for the archive port.
- [x] Add path existence support if the existing local filesystem adapter does not already provide it.
- [x] Add directory move support for relative paths under the project root.
- [x] Ensure directory existence checks distinguish directories from files.
- [x] Ensure path existence checks detect both files and directories.
- [x] Ensure directory movement does not overwrite an existing destination file or directory.
- [x] Preserve nested files and directories when moving an active change directory.
- [x] Add or update adapter tests for path existence and directory movement.
- [x] Ensure the filesystem adapter does not contain OpenSpec path policy, archive date policy, archive result construction, CLI report formatting, completion checks, changelog logic, release logic, provider logic, source-control logic, workflow logic, or rollback logic.

## Phase 5: CLI Wiring and Reporting

- [x] Replace the `archive` placeholder in `internal/adapters/cli/cli.go`.
- [x] Parse `specharbor archive <change-id>`.
- [x] Reject missing change id.
- [x] Reject unsupported flags.
- [x] Reject extra positional arguments.
- [x] Obtain the current working directory for the project root.
- [x] Derive the current local archive date formatted as `YYYY-MM-DD`.
- [x] Keep archive date derivation small and testable without introducing a broad or unused clock abstraction.
- [x] Construct the archive use case with the local filesystem adapter.
- [x] Invoke the use case with project root, change id, and archive date.
- [x] Print a human-readable archive report from the structured result.
- [x] Ensure the report includes the change id, relative source path, relative archive path, archive date, `Moved: yes`, and moved directory line.
- [x] Match the expected success report shape from `design.md`.
- [x] Return argument and execution errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping unless minimal error handling changes are strictly required.
- [x] Preserve existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, and unknown command behavior.

## Phase 6: Tests

- [x] Add domain tests for archive result behavior.
- [x] Add use case tests with a fake archive filesystem for successful archive movement.
- [x] Test that the use case creates `openspec/archive/` when missing.
- [x] Test that the use case creates `openspec/archive/YYYY-MM-DD/` when missing.
- [x] Test that the use case moves `openspec/changes/<change-id>/` to `openspec/archive/YYYY-MM-DD/<change-id>/`.
- [x] Test that the structured result contains change id, source path, archive path, archive date, and moved directory information.
- [x] Test that empty project root is rejected.
- [x] Test that empty change id is rejected.
- [x] Test that malformed archive date is rejected.
- [x] Test that unsafe change ids are rejected before filesystem writes or moves.
- [x] Test that missing `openspec/project.md` is rejected before archive directory creation or movement.
- [x] Test that missing `openspec/changes/` is rejected before archive directory creation or movement.
- [x] Test that missing source change directory is rejected before archive directory creation or movement.
- [x] Test that a source change path that exists as a file is rejected before archive directory creation or movement.
- [x] Test that `openspec/archive` existing as a file is rejected before archive directory creation or movement.
- [x] Test that `openspec/archive/YYYY-MM-DD` existing as a file is rejected before archive directory creation or movement.
- [x] Test that an existing archive target directory is rejected before moving.
- [x] Test that an existing archive target file is rejected before moving.
- [x] Test that existing archived content is not overwritten, whether it is a file or a directory.
- [x] Test that filesystem check, create, and move errors are returned as errors.
- [x] Test that the local filesystem adapter satisfies the archive filesystem port.
- [x] Test that the local filesystem adapter moves directories and preserves nested contents.
- [x] Add CLI tests for a successful archive report.
- [x] Add CLI tests proving the successful archive report uses `Moved: yes`.
- [x] Add CLI or helper tests for current local archive date formatting without introducing a broad clock abstraction.
- [x] Add CLI tests for missing change id, unsupported flags, extra positional arguments, and unsafe change ids.
- [x] Preserve or add regression coverage for `help`, `version`, `init`, `prompt`, `validate`, `generate`, and unknown command behavior.

## Phase 7: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor archive <change-id>` moves an active change directory into `openspec/archive/YYYY-MM-DD/<change-id>/`.
- [x] Manually verify `openspec/archive/` and `openspec/archive/YYYY-MM-DD/` are created when missing.
- [x] Manually verify `openspec/archive` existing as a file returns a clear execution error.
- [x] Manually verify `openspec/archive/YYYY-MM-DD` existing as a file returns a clear execution error.
- [x] Manually verify an existing archive destination file or directory is not overwritten.
- [x] Manually verify missing change id returns a clear argument error.
- [x] Manually verify unsupported flags and extra positional arguments are rejected.
- [x] Manually verify unsafe change ids are rejected before filesystem writes or moves.
- [x] Manually verify missing OpenSpec project structure is rejected before filesystem writes or moves.
- [x] Manually verify missing OpenSpec project structure tells the user to run `specharbor init` first.
- [x] Manually verify a missing source change directory returns a clear execution error.
- [x] Manually verify a source change path that exists as a file returns a clear execution error.
- [x] Manually verify the command does not call AI providers, local model APIs, external agents, workflow tools, source-control APIs, network APIs, or external processes.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
