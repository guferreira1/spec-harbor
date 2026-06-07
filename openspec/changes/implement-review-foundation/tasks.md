# Tasks: Implement Review Foundation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-review-foundation/`.
- [x] Inspect the current CLI command registry and error flow before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing domain, use case, port, filesystem adapter, and test patterns before editing.
- [x] Inspect the existing validation and archive implementations for report formatting, non-zero exit handling, and filesystem port patterns.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to `specharbor review <change-id>` and the deterministic review foundation described by this change.
- [x] Do not implement Git diff analysis, GitHub/GitLab API calls, semantic code review, semantic Markdown review beyond task checkbox counting, AI-assisted review, agent-assisted review, architecture scanning, archive-readiness enforcement, auto-fix behavior, file updates, scan, config, release notes, changelog generation, or changes to validation, generation, archive, or prompt behavior.

## Phase 1: Domain Concepts

- [x] Add review domain types under `internal/core/domain`.
- [x] Represent review status values: `approved`, `needs-work`, and `invalid`.
- [x] Represent review finding severity values: `error` and `warning`.
- [x] Represent review finding codes: `project_root_unavailable`, `change_directory_missing`, `required_file_missing`, `tasks_file_unreadable`, `tasks_not_found`, and `incomplete_task`.
- [x] Represent review findings with severity, code, message, relative path, and optional subject.
- [x] Represent task summary counts for total, completed, and incomplete tasks.
- [x] Represent a structured review result containing change id, checked path, status, required files, task summary, and findings.
- [x] Ensure review result status calculation returns `invalid` for error findings, `needs-work` for incomplete task warnings, and `approved` only when all detected tasks are completed.
- [x] Reuse `domain.RequiredOpenSpecChangeFiles()` for required file policy instead of duplicating the required file list in review-specific code.
- [x] Add helpers only where they keep result creation, task summary creation, defensive copying, or status calculation small and explicit.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, and external process execution.

## Phase 2: Ports

- [x] Add a small review-specific filesystem port under `internal/core/ports`.
- [x] Include only operations needed by this change: directory existence, file existence, and file reading.
- [x] Use a behavior-specific name such as `ReviewFileSystem`.
- [x] Keep the review filesystem port separate from initialization, validation, generation, archive, prompt rendering, AI provider, workflow dispatcher, source-control, release, and agent contracts.
- [x] Ensure the review use case depends on the review filesystem port instead of `internal/adapters/filesystem`.

## Phase 3: Review Use Case

- [x] Add the review use case under `internal/core/usecase`.
- [x] Make the use case accept project root and change id as input.
- [x] Validate that the use case dependency is present.
- [x] Validate that project root is non-empty after trimming whitespace.
- [x] Validate that change id is non-empty after trimming whitespace.
- [x] Reject unsafe change ids before performing filesystem checks.
- [x] Build the checked relative path as `openspec/changes/<change-id>`.
- [x] Load required file policy from `domain.RequiredOpenSpecChangeFiles()`.
- [x] Check OpenSpec project availability by verifying that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory through the review filesystem port.
- [x] Return an invalid review result with `project_root_unavailable` when OpenSpec project availability cannot be verified.
- [x] Check that `openspec/changes/<change-id>/` exists as a directory.
- [x] Return an invalid review result with `change_directory_missing` when the change directory is missing or is not a directory.
- [x] Check every required OpenSpec change file through the review filesystem port when the change directory exists.
- [x] Return one `required_file_missing` error finding per missing required file.
- [x] Stop before reading `tasks.md` when required files are missing.
- [x] Read `openspec/changes/<change-id>/tasks.md` through the review filesystem port.
- [x] Return an invalid review result with `tasks_file_unreadable` when `tasks.md` cannot be read after required file checks passed.
- [x] Parse `tasks.md` using deterministic task checkbox parsing.
- [x] Return an invalid review result with `tasks_not_found` when no Markdown task checkboxes are detected.
- [x] Return one `incomplete_task` warning finding per incomplete task.
- [x] Return a needs-work review result when one or more detected tasks are incomplete and no error findings exist.
- [x] Return an approved review result when all required files exist, task checkboxes are detected, and every detected task is completed.
- [x] Return errors for dependency failures and filesystem execution failures that prevent checks from completing, except unreadable `tasks.md`, which must be represented as a structured invalid result.
- [x] Do not print from the use case.
- [x] Do not call `os`, terminal IO, provider SDKs, network APIs, source-control APIs, external agents, external processes, or workflow tools from the use case.
- [x] Keep the structure ready for future review collaborators without adding unused exported strategy, factory, registry, chain, provider, source-control, workflow, AI, agent, scanner, or auto-fix abstractions.

## Phase 4: Filesystem Adapter

- [x] Use `internal/adapters/filesystem` as the concrete implementation of the review filesystem port.
- [x] Add only compatibility code needed for the review port.
- [x] Add file reading support if the existing local filesystem adapter does not already provide it.
- [x] Ensure directory existence checks distinguish directories from files.
- [x] Ensure file existence checks distinguish files from directories.
- [x] Ensure file reading returns the contents of `tasks.md` without adding task parsing or review policy to the adapter.
- [x] Add or update adapter tests for review-required file reading behavior.
- [x] Ensure the filesystem adapter does not contain OpenSpec required-file policy, task checkbox parsing policy, review status calculation, review finding construction, CLI report formatting, provider logic, source-control logic, workflow logic, architecture scanning, or auto-fix logic.

## Phase 5: Task Parsing

- [x] Implement deterministic task checkbox parsing in core as pure logic.
- [x] Count completed tasks when a line starts with `- [x]` after leading whitespace is trimmed.
- [x] Count completed tasks when a line starts with `- [X]` after leading whitespace is trimmed.
- [x] Count incomplete tasks when a line starts with `- [ ]` after leading whitespace is trimmed.
- [x] Extract incomplete task text from the remainder of the task line for finding messages.
- [x] Ignore lines that do not start with a supported checkbox marker after leading whitespace is trimmed.
- [x] Count nested Markdown task lines when leading whitespace is present before the task marker.
- [x] Return total, completed, incomplete, and incomplete task text values to the review use case.
- [x] Keep task parsing free of filesystem access, terminal IO, external Markdown libraries, AI calls, Git access, provider SDKs, and semantic review beyond checkbox counting.

## Phase 6: CLI Wiring and Reporting

- [x] Replace the `review` placeholder in `internal/adapters/cli/cli.go`.
- [x] Parse `specharbor review <change-id>`.
- [x] Reject missing change id.
- [x] Reject unsupported flags.
- [x] Reject extra positional arguments.
- [x] Reject leading-dash change ids through flag parsing.
- [x] Obtain the current working directory for the project root.
- [x] Construct the review use case with the local filesystem adapter.
- [x] Invoke the use case with project root and change id.
- [x] Print a human-readable review report from the structured result.
- [x] Ensure the approved report includes change id, checked path, status, task counts, and `Findings: 0`.
- [x] Ensure needs-work and invalid reports include change id, checked path, status, task counts, and a findings list.
- [x] Match the expected approved, needs-work, and invalid report shapes from `design.md`.
- [x] Return zero only when the review status is `approved`.
- [x] Return a non-zero exit after printing the report when the review status is `needs-work` or `invalid`.
- [x] Return argument and execution errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping unless minimal error handling changes are strictly required.
- [x] Preserve existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, `archive`, and unknown command behavior.
- [x] Keep the CLI adapter free of required-file policy, task parsing policy, status calculation, filesystem policy, provider logic, source-control logic, workflow logic, architecture scanning, and semantic review.

## Phase 7: Tests

- [x] Add domain tests for review status behavior.
- [x] Add domain tests for review result construction.
- [x] Add domain tests for task summary counts.
- [x] Add domain tests for defensive copies if review result slices are exposed.
- [x] Add task parser tests for completed `- [x]` tasks.
- [x] Add task parser tests for completed `- [X]` tasks.
- [x] Add task parser tests for incomplete `- [ ]` tasks.
- [x] Add task parser tests proving non-task lines are ignored.
- [x] Add task parser tests proving malformed checkbox lines are ignored.
- [x] Add task parser tests proving leading whitespace before task markers is supported.
- [x] Add task parser tests proving incomplete task text is extracted for finding messages.
- [x] Add use case tests for incomplete tasks without text using the fallback finding message.
- [x] Add use case tests with a fake review filesystem for approved review results.
- [x] Test that a complete change with all detected tasks completed returns status `approved`.
- [x] Test that a complete change with incomplete tasks returns status `needs-work`.
- [x] Test that incomplete tasks produce one `incomplete_task` warning per incomplete task.
- [x] Test that missing OpenSpec project structure returns status `invalid` with `project_root_unavailable`.
- [x] Test that a missing change directory returns status `invalid` with `change_directory_missing`.
- [x] Test that a change path that is not a directory returns status `invalid` with `change_directory_missing`.
- [x] Test that missing required files return status `invalid` with `required_file_missing` findings.
- [x] Test that required file checks use `domain.RequiredOpenSpecChangeFiles()`.
- [x] Test that required-file failures stop before reading `tasks.md`.
- [x] Test that unreadable `tasks.md` returns status `invalid` with `tasks_file_unreadable`.
- [x] Test that no detected task checkboxes returns status `invalid` with `tasks_not_found`.
- [x] Test that empty project root is rejected.
- [x] Test that empty change id is rejected.
- [x] Test that unsafe change ids are rejected before filesystem checks.
- [x] Test that filesystem check errors are returned as errors where no structured finding can be produced safely.
- [x] Test that the local filesystem adapter satisfies the review filesystem port.
- [x] Test that the local filesystem adapter reads `tasks.md` contents.
- [x] Test that the local filesystem adapter distinguishes files from directories for review checks.
- [x] Add CLI tests for an approved review report.
- [x] Add CLI tests for a needs-work review report listing incomplete tasks.
- [x] Add CLI tests for an invalid review report listing missing required files.
- [x] Align CLI approved and needs-work report tests with the `design.md` task counts and incomplete task messages.
- [x] Add CLI tests proving approved review returns zero.
- [x] Add CLI tests proving needs-work review prints the report and returns non-zero.
- [x] Add CLI tests proving invalid review prints the report and returns non-zero.
- [x] Add CLI tests for missing change id, unsupported flags, extra positional arguments, and unsafe change ids.
- [x] Expand CLI unsupported review flag tests for `--format`, `--github`, `--gitlab`, `--diff`, `--fix`, and `--agent`.
- [x] Preserve or add regression coverage for `help`, `version`, `init`, `prompt`, `validate`, `generate`, `archive`, and unknown command behavior.

## Phase 8: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor review <change-id>` prints an approved report and exits zero for a complete change with all detected tasks completed.
- [x] Manually verify `specharbor review <change-id>` prints a needs-work report and exits non-zero for a complete change with incomplete tasks.
- [x] Manually verify `specharbor review <change-id>` prints an invalid report and exits non-zero when a required file is missing.
- [x] Manually verify missing change id returns a clear argument error.
- [x] Manually verify unsupported flags and extra positional arguments are rejected.
- [x] Manually verify unsafe change ids are rejected before filesystem checks.
- [x] Manually verify missing OpenSpec project structure returns an invalid review result.
- [x] Manually verify missing change directory returns an invalid review result.
- [x] Manually verify unreadable `tasks.md` returns an invalid review result.
- [x] Manually verify `tasks.md` with no detected task checkboxes returns an invalid review result.
- [x] Manually verify the command does not call AI providers, local model APIs, external agents, workflow tools, source-control APIs, network APIs, Git commands, or external processes.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
