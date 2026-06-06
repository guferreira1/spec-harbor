# Tasks: Implement Validation Foundation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-validation-foundation/`.
- [x] Inspect the current CLI command registry and error flow before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing core use case, port, filesystem adapter, and test patterns before editing.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to `specharbor validate <change-id>` and the validation foundation described by this change.
- [x] Do not implement AI calls, semantic validation, generate, archive, review, scan, config, provider integrations, external-agent integrations, or auto-fix behavior.

## Phase 1: Domain Concepts

- [x] Add validation domain types under `internal/core/domain`.
- [x] Represent validation status values: `valid` and `invalid`.
- [x] Represent validation finding severity, starting with `error`.
- [x] Represent validation findings with code, message, relative path, and optional subject or filename.
- [x] Represent a structured validation result containing change id, checked path, required files, findings, and status.
- [x] Represent the required OpenSpec change files without putting that policy in the CLI adapter.
- [x] Add helpers only where they keep result creation or status calculation small and explicit.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, and workflow SDKs.

## Phase 2: Ports

- [x] Add a small validation-specific filesystem port under `internal/core/ports`.
- [x] Include only the operations needed by this change: directory existence and file existence.
- [x] Keep the validation filesystem port separate from initialization, prompt rendering, AI provider, workflow dispatcher, spec generation, and review contracts.
- [x] Ensure use cases depend on the validation port instead of `internal/adapters/filesystem`.

## Phase 3: Validation Use Case

- [x] Add the validation use case under `internal/core/usecase`.
- [x] Make the use case accept project root and change id as input.
- [x] Validate that project root is non-empty after trimming whitespace.
- [x] Validate that change id is non-empty after trimming whitespace.
- [x] Build the checked relative path as `openspec/changes/<change-id>`.
- [x] Check OpenSpec project availability by verifying that `openspec/project.md` and `openspec/changes/` exist through the filesystem port.
- [x] Return an invalid result with a `project_root_unavailable` finding when OpenSpec project availability cannot be verified.
- [x] Check target change directory existence through the filesystem port.
- [x] Return an invalid result with a `change_directory_missing` finding when the change directory is missing.
- [x] Check every required file through the filesystem port when the change directory exists.
- [x] Return one `required_file_missing` finding per missing required file.
- [x] Return a valid result when all required files exist.
- [x] Return errors for dependency failures, invalid input, and filesystem execution errors.
- [x] Do not print from the use case.
- [x] Do not call `os`, terminal IO, provider SDKs, external APIs, external agents, or workflow tools from the use case.
- [x] Keep the structure ready for future validator extraction without adding unused exported chain or registry abstractions.

## Phase 4: Filesystem Adapter

- [x] Use `internal/adapters/filesystem` as the concrete implementation of the validation filesystem port.
- [x] Add only compatibility code needed for the validation port, if the existing local filesystem adapter is not already sufficient.
- [x] Add or update adapter tests to cover directory and file existence behavior needed by validation.
- [x] Ensure the filesystem adapter does not contain required-file policy, validation finding construction, or CLI report formatting.

## Phase 5: CLI Wiring and Reporting

- [x] Replace the `validate` placeholder in `internal/adapters/cli/cli.go`.
- [x] Parse `specharbor validate <change-id>`.
- [x] Reject missing change id.
- [x] Reject unsupported flags.
- [x] Reject extra positional arguments.
- [x] Obtain the current working directory for the project root.
- [x] Construct the validation use case with the local filesystem adapter.
- [x] Print a human-readable valid report when validation succeeds.
- [x] Print a human-readable invalid report when validation returns findings.
- [x] Ensure invalid validation results make the CLI return a non-zero exit code after printing the report, so `specharbor validate` can be used in CI.
- [x] Ensure invalid reports list missing required files by filename.
- [x] Ensure missing change directory reports identify `openspec/changes/<change-id>`.
- [x] Return argument and execution errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping unless minimal error handling changes are strictly required.
- [x] Preserve existing `help`, `version`, `init`, `prompt`, and unknown command behavior.

## Phase 6: Tests

- [x] Add use case tests with fake filesystem ports for valid and invalid results.
- [x] Test that a complete change returns status `valid` with no findings.
- [x] Test that missing OpenSpec project structure returns status `invalid` with `project_root_unavailable`.
- [x] Test that a missing change directory returns status `invalid` with `change_directory_missing`.
- [x] Test that missing required files return status `invalid` with `required_file_missing` findings.
- [x] Test that existing required files are not reported as missing.
- [x] Test that empty project root is rejected.
- [x] Test that empty change id is rejected.
- [x] Test that filesystem errors are returned as errors.
- [x] Test that the local filesystem adapter satisfies the validation filesystem port.
- [x] Add CLI tests for a valid report.
- [x] Add CLI tests for an invalid report with missing required files.
- [x] Add CLI tests that invalid validation results print the report and return a non-zero exit code.
- [x] Add CLI tests for a missing change directory report.
- [x] Add CLI tests for missing change id, unsupported flags, and extra positional arguments.
- [x] Preserve or add regression coverage for `help`, `version`, `init`, `prompt`, and unknown command behavior.

### Test Engineer Follow-up

- [x] Add CLI coverage for missing OpenSpec project structure reporting `project_root_unavailable` and returning a non-zero exit code.
- [x] Broaden use case coverage so filesystem execution failures remain errors across project, changes directory, change directory, and required-file checks.
- [x] Add local filesystem adapter coverage for distinguishing files from directories during validation existence checks.

## Phase 7: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor validate <change-id>` prints a valid report for a complete change.
- [x] Manually verify `specharbor validate <change-id>` prints an invalid report and exits non-zero when required files are missing.
- [x] Manually verify `specharbor validate <change-id>` prints an invalid report and exits non-zero when the change directory is missing.
- [x] Manually verify `specharbor validate` returns a clear argument error.
- [x] Manually verify the command does not call AI providers, external agents, workflow tools, or external processes.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
