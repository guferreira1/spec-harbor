# Tasks: Implement Generation Foundation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-generation-foundation/`.
- [x] Inspect the current CLI command registry and error flow before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing domain, use case, port, filesystem adapter, template adapter, and test patterns before editing.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to `specharbor generate <change-id> --blank` and the generation foundation described by this change.
- [x] Do not implement guided generation, template generation, AI-assisted generation, agent-assisted generation, hybrid generation, archive, review, scan, config, provider integrations, external-agent integrations, workflow integrations, validation changes, or auto-fix behavior.

## Phase 1: Domain Concepts

- [x] Add or reuse generation mode values under `internal/core/domain`, with blank generation as the only implemented mode for this change.
- [x] Add generation result concepts under `internal/core/domain`.
- [x] Represent the requested change id, generation mode, relative change path, created files, skipped existing files, and change-directory status in the structured result.
- [x] Reuse the existing domain-level required OpenSpec change file policy, currently `domain.RequiredOpenSpecChangeFiles()`, for the required blank OpenSpec change files.
- [x] Do not duplicate the required OpenSpec change file list in generation-specific code for iteration or policy decisions.
- [x] Keep required OpenSpec change file policy out of the CLI adapter.
- [x] Add helpers only where they keep result creation, required-file access, or defensive copying small and explicit.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, and workflow SDKs.

## Phase 2: Ports

- [x] Add a small generation-specific filesystem port under `internal/core/ports`.
- [x] Include only operations needed by this change: directory existence, file existence, directory creation, and write-if-absent file creation.
- [x] Keep the generation filesystem port separate from initialization, validation, prompt rendering, AI provider, workflow dispatcher, spec generation, and review contracts.
- [x] Add a small blank generation content port under `internal/core/ports` if starter content is loaded from an adapter.
- [x] Ensure the generation use case depends on ports instead of `internal/adapters/filesystem` or `internal/adapters/templates`.

## Phase 3: Generation Use Case

- [x] Add the generation use case under `internal/core/usecase`.
- [x] Make the use case accept project root, change id, and generation mode as input.
- [x] Validate that project root is non-empty after trimming whitespace.
- [x] Validate that change id is non-empty after trimming whitespace.
- [x] Validate that the requested mode is blank.
- [x] Reject unsupported generation modes.
- [x] Reject unsafe change ids before performing filesystem writes.
- [x] Build the change relative path as `openspec/changes/<change-id>`.
- [x] Check OpenSpec project availability by verifying that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory through the generation filesystem port.
- [x] Return a clear execution error telling the user to run `specharbor init` first when OpenSpec project availability cannot be verified.
- [x] Do not create `openspec/`, `openspec/project.md`, or `openspec/changes/` from the generation use case.
- [x] Create the target change directory when missing.
- [x] Continue without error when the target change directory already exists.
- [x] Treat partially existing change directories as recoverable.
- [x] Load starter content for each required blank OpenSpec file.
- [x] Write each required file through a write-if-absent filesystem operation.
- [x] Record created files when files are written.
- [x] Record skipped existing files when files already exist.
- [x] Create only missing required files and never overwrite existing file contents.
- [x] Return a structured generation result.
- [x] Return errors for dependency failures, invalid input, filesystem execution errors, and content-loading errors.
- [x] Do not print from the use case.
- [x] Do not call `os`, terminal IO, provider SDKs, network APIs, external agents, external processes, or workflow tools from the use case.
- [x] Keep the structure ready for future generation strategy extraction without adding unused exported strategy, factory, registry, or chain abstractions.

## Phase 4: Filesystem Adapter

- [x] Use `internal/adapters/filesystem` as the concrete implementation of the generation filesystem port.
- [x] Add only compatibility code needed for the generation port, if the existing local filesystem adapter is not already sufficient.
- [x] Ensure file creation uses defensive write-if-absent behavior and does not overwrite existing files.
- [x] Add or update adapter tests for directory creation, file existence, directory existence, and write-if-absent behavior as needed by generation.
- [x] Ensure the filesystem adapter does not contain required-file policy, starter content, generation result construction, CLI report formatting, provider logic, or agent logic.

## Phase 5: Blank Starter Content

- [x] Add or extend an adapter under `internal/adapters/templates` for blank OpenSpec change starter content if content is not kept directly in the use case.
- [x] Provide deterministic starter content for `proposal.md`.
- [x] Provide deterministic starter content for `design.md`.
- [x] Provide deterministic starter content for `tasks.md`.
- [x] Provide deterministic starter content for `acceptance-criteria.md`.
- [x] Provide deterministic starter content for `risks.md`.
- [x] Ensure generated `tasks.md` uses unchecked tasks only.
- [x] Return a clear error for unknown blank file paths.
- [x] Add tests that every required blank file has non-empty useful Markdown content.
- [x] Do not add custom template paths, user template loading, provider prompts, agent prompts, or semantic rendering.

## Phase 6: CLI Wiring and Reporting

- [x] Replace the `generate` placeholder in `internal/adapters/cli/cli.go`.
- [x] Parse `specharbor generate <change-id> --blank`.
- [x] Accept `--blank` before or after the change id if the parser is intentionally order-independent.
- [x] Reject missing change id.
- [x] Reject missing `--blank`.
- [x] Reject duplicate `--blank`.
- [x] Reject unsupported flags.
- [x] Reject extra positional arguments.
- [x] Obtain the current working directory for the project root.
- [x] Construct the generation use case with the local filesystem adapter and blank content adapter.
- [x] Invoke the use case with blank generation mode.
- [x] Print a human-readable generation report from the structured result.
- [x] Ensure the report includes the change id, relative change path, change-directory status, created file count, skipped existing file count, and relevant filenames.
- [x] Match the expected success report shape for a newly generated change, including `Directory: created`, created count `5`, skipped count `0`, and the created file list.
- [x] Match the expected success report shape for rerunning the same command, including `Directory: existing`, created count `0`, skipped count `5`, and the skipped existing file list.
- [x] Return argument and execution errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping unless minimal error handling changes are strictly required.
- [x] Preserve existing `help`, `version`, `init`, `prompt`, `validate`, and unknown command behavior.

## Phase 7: Tests

- [x] Add domain tests for generation result behavior.
- [x] Add use case tests with fake filesystem and fake content ports for a newly generated blank change.
- [x] Test that the use case creates the target change directory when missing.
- [x] Test that the use case fills in missing files when the target change directory already exists.
- [x] Test that existing files are skipped and not overwritten.
- [x] Test that created files and skipped existing files are returned in the structured result.
- [x] Test that the required files come from `domain.RequiredOpenSpecChangeFiles()` or the existing shared domain-level required OpenSpec change file policy.
- [x] Test that empty project root is rejected.
- [x] Test that empty change id is rejected.
- [x] Test that unsafe change ids are rejected before filesystem writes.
- [x] Test that unsupported generation modes are rejected.
- [x] Test that missing `openspec/project.md` is rejected before target writes.
- [x] Test that missing `openspec/changes/` is rejected before target writes.
- [x] Test that missing OpenSpec project structure reports a clear execution error telling the user to run `specharbor init` first.
- [x] Test that generation does not create `openspec/`, `openspec/project.md`, or `openspec/changes/`.
- [x] Test that filesystem errors are returned as errors.
- [x] Test that content-loading errors are returned as errors.
- [x] Test that the local filesystem adapter satisfies the generation filesystem port.
- [x] Add content adapter tests for every required blank file.
- [x] Add CLI tests for a successful blank generation report.
- [x] Add CLI tests for skipped existing files in the generation report.
- [x] Add CLI tests for missing change id, missing `--blank`, duplicate `--blank`, unsupported flags, extra positional arguments, and unsafe change ids.
- [x] Preserve or add regression coverage for `help`, `version`, `init`, `prompt`, `validate`, and unknown command behavior.

## Phase 8: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor generate <change-id> --blank` creates the blank change directory and files in an initialized project.
- [x] Manually verify running the command again skips existing files and does not overwrite their contents.
- [x] Manually verify missing change id returns a clear argument error.
- [x] Manually verify missing `--blank` returns a clear argument error.
- [x] Manually verify unsupported flags and extra positional arguments are rejected.
- [x] Manually verify unsafe change ids are rejected before filesystem writes.
- [x] Manually verify missing OpenSpec project structure is rejected before filesystem writes.
- [x] Manually verify missing OpenSpec project structure tells the user to run `specharbor init` first.
- [x] Manually verify the command does not call AI providers, local model APIs, external agents, workflow tools, network APIs, or external processes.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
