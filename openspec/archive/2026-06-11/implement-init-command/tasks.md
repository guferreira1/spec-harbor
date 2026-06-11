# Tasks: Implement Init Command

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-init-command/`.
- [x] Inspect current CLI files before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing default configuration and rule files: `.specharbor/config.example.yml` and `.specharbor/rules/*.md`; use the example config as a shape for safe defaults, not as an exact source for placeholder secrets or local-only values.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to the `specharbor init` behavior described by this change.
- [x] Do not implement flags, prompts, validation, scan, generate, prompt, review, archive, config, or AI-provider behavior.

## Phase 1: Core Contracts and Result Types

- [x] Add a small filesystem port under `internal/core/ports` for initialization needs only.
- [x] Ensure the filesystem port supports checking existence, creating directories, and writing files only when absent.
- [x] Add a small default-content/template port under `internal/core/ports` for initialization defaults only.
- [x] Add initialization input and result types under `internal/core/usecase` or `internal/core/domain`, keeping them specific to this use case.
- [x] Represent at least these result statuses: initialized and already initialized.
- [x] Include created and skipped item lists or counts in the structured result.
- [x] Keep core packages free of imports from `internal/adapters`, `cmd`, `os`, terminal IO, network APIs, and provider SDKs.

## Phase 2: Initialization Use Case

- [x] Add the project initialization use case under `internal/core/usecase`.
- [x] Make the use case accept the target project root as input instead of reading the current working directory itself.
- [x] Define the required directories: `openspec/`, `openspec/specs/`, `openspec/changes/`, `.specharbor/`, and `.specharbor/rules/`.
- [x] Define the required files: `openspec/project.md`, `.specharbor/config.yml`, and the default rule files under `.specharbor/rules/`.
- [x] Create missing directories through the filesystem port.
- [x] Write missing files through the filesystem port without overwriting existing files.
- [x] Use the template/default-content port to obtain contents for generated files.
- [x] Treat a project as already initialized only when every required directory and file already exists.
- [x] For a partially initialized project, create only the missing items and report existing items as skipped.
- [x] Return errors instead of panicking.
- [x] Do not print messages from the use case.

## Phase 3: Adapters

- [x] Add a local filesystem adapter under `internal/adapters/filesystem`.
- [x] Implement directory creation with standard library filesystem APIs.
- [x] Implement absent-only file creation defensively so existing files are not overwritten.
- [x] Add a default template adapter under `internal/adapters/templates` or another clearly named adapter package.
- [x] Provide generated content for `openspec/project.md`.
- [x] Generate `.specharbor/config.yml` from safe defaults based on the existing `.specharbor/config.example.yml` shape, without copying placeholder secrets, real credentials, or local-only values.
- [x] Provide generated content for default rule files matching the repository's current `.specharbor/rules/*.md` defaults.
- [x] Embed default rule templates from `.specharbor/rules/` with `go:embed` or an equivalent compiled-in mechanism so the installed CLI does not depend on repository-relative files at runtime.
- [x] Keep adapter code free of business decisions beyond concrete filesystem and template mechanics.

## Phase 4: CLI Wiring

- [x] Replace the `init` placeholder in `internal/adapters/cli/cli.go`.
- [x] Have the CLI adapter obtain the current working directory for the target project root.
- [x] Construct the initialization use case with the local filesystem and default template adapters.
- [x] Call the use case for `specharbor init`.
- [x] Print a concise initialized message for first-time or partial initialization that includes `SpecHarbor initialized`.
- [x] Print a clear already-initialized message when the use case returns that status that includes `SpecHarbor already initialized`.
- [x] Return non-success errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping.
- [x] Preserve existing `help`, `version`, and unknown command behavior.

## Phase 5: Tests

- [x] Add use case tests with fake ports for empty, partial, and already-initialized projects.
- [x] Test that existing file contents are not overwritten.
- [x] Test that missing files and directories are reported as created.
- [x] Test that existing files and directories are reported as skipped.
- [x] Add filesystem adapter tests using temporary directories.
- [x] Add template adapter tests verifying all required defaults are available.
- [x] Add CLI tests asserting first-time or partial `init` output includes `SpecHarbor initialized` and second-run already-initialized output includes `SpecHarbor already initialized`.
- [x] Add regression tests for `help`, `version`, and unknown command behavior if current coverage is missing.

## Phase 6: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor init` in a temporary empty directory creates the required structure.
- [x] Manually verify running `specharbor init` a second time reports the project is already initialized and does not overwrite files.
- [x] Manually verify a partial initialization creates missing items and preserves existing files.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
