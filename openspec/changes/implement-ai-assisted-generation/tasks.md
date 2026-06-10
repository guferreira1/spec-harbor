# Tasks: Implement AI-Assisted Generation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-ai-assisted-generation/`.
- [x] Inspect current generation, validation, agent-assisted authoring, runner, filesystem, template, CLI, and architecture test patterns before editing.
- [x] Keep implementation limited to `specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> [--overwrite]`.
- [x] Do not implement provider APIs, remote AI calls, local model APIs, OAuth, credentials, remote execution, cloud execution, IDE automation, marketplace integrations, live runner output apply, source-control automation, workflow automation, patch application, auto-commit, auto-push, PR creation, merge automation, or automatic archive.
- [x] Do not modify production code outside the approved implementation surface.
- [x] Preserve existing blank, template, guided, agent-assisted, validate, prompt, review, archive, config, scan, init, workflow, help, and version behavior.
- [x] Run the baseline test suite before implementation.

## Phase 1: Domain Parser and Models

- [x] Add or reuse the `ai-assisted` generation mode in domain where needed.
- [x] Add domain policy for allowed AI-generated OpenSpec filenames based on `domain.RequiredOpenSpecChangeFiles()`.
- [x] Add domain models for AI output file blocks, parsed generated files, parse findings, parse result, and AI-assisted generation result.
- [x] Add stable parse finding codes for unknown file blocks, duplicate blocks, missing blocks, empty content, path traversal, absolute filenames, nested filenames, malformed block syntax, text outside blocks, and unclosed blocks.
- [x] Implement the strict line-oriented parser for `---FILE: <filename>---` and `---END FILE---`.
- [x] Reject unknown filenames.
- [x] Reject duplicate file blocks.
- [x] Reject missing required file blocks.
- [x] Reject empty or whitespace-only generated content.
- [x] Reject path traversal.
- [x] Reject absolute paths.
- [x] Reject nested paths and filename separators.
- [x] Reject malformed starts, orphan end markers, unclosed blocks, and non-whitespace text outside blocks.
- [x] Ensure parser output ordering follows `domain.RequiredOpenSpecChangeFiles()`.
- [x] Ensure parser logic is deterministic, local, and pure over strings.
- [x] Ensure parser does not execute content, interpret shell commands, parse patches, or fetch URLs.
- [x] Defensively copy slices and maps exposed by domain constructors or accessors.

## Phase 2: Domain Tests

- [x] Test allowed generated file names are exactly `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Test unknown file names are rejected.
- [x] Test duplicate file blocks are rejected.
- [x] Test path traversal is rejected.
- [x] Test absolute paths are rejected.
- [x] Test nested paths and separators are rejected.
- [x] Test successful parser output for all required files.
- [x] Test malformed parser input, including orphan end markers, unclosed blocks, and text outside blocks.
- [x] Test missing required file blocks.
- [x] Test empty generated content behavior.
- [x] Test parse finding codes, messages, filenames, and line numbers.
- [x] Test generation result model fields.
- [x] Test defensive copying where applicable.

## Phase 3: Ports and Adapter Behavior

- [x] Add small core-owned ports for AI-assisted source reads and safe OpenSpec target writes.
- [x] Include local source AI output file reads in the port contract.
- [x] Include target directory/file existence checks in the port contract.
- [x] Include creating the target change directory in the port contract.
- [x] Include write-if-absent behavior in the port contract.
- [x] Include explicit overwrite behavior in the port contract.
- [x] Implement the local filesystem adapter behavior for the new port or ports.
- [x] Ensure source file reads treat `--from-file` as a local path only and never fetch URLs.
- [x] Ensure target writes receive only use-case-constructed approved relative OpenSpec paths.
- [x] Prevent target path traversal and absolute target paths at the adapter boundary as a defense in depth.
- [x] Report missing source files clearly.
- [x] Report write errors clearly.

## Phase 4: Use Case

- [x] Add the AI-assisted generation use case under `internal/core/usecase`.
- [x] Validate dependencies.
- [x] Validate project root.
- [x] Validate change id with `domain.NewChangeID`.
- [x] Validate source file path is present.
- [x] Require `openspec/project.md` and `openspec/changes/` before reading and writing.
- [x] Read source AI output through the source-read port.
- [x] Parse source content with the domain parser.
- [x] Return parse failures as structured errors or results without writing files.
- [x] Create the target change directory if it is missing and the parsed input is valid.
- [x] Fail before writes if the target path exists and is not a directory.
- [x] Preflight all target filenames before writes.
- [x] Skip existing files by default.
- [x] Overwrite existing files only when `Overwrite` is true.
- [x] Write only required OpenSpec files under `openspec/changes/<change-id>/`.
- [x] Record generated files.
- [x] Record skipped existing files.
- [x] Record overwritten files.
- [x] Avoid expected partial writes through parse and preflight checks.
- [x] Report runtime write failures clearly without deleting or rolling back user files automatically.
- [x] Run existing validation after all planned writes succeed.
- [x] Return validation result in the AI-assisted generation result.
- [x] Ensure unsafe change ids fail before source reads and filesystem writes.
- [x] Keep output formatting out of the use case.
- [x] Keep `os`, `os/exec`, adapters, CLI packages, provider SDKs, network APIs, source-control SDKs, workflow SDKs, and external-agent SDKs out of core.

## Phase 5: Use Case Tests

- [x] Valid AI output generates the required OpenSpec files.
- [x] Malformed AI output writes nothing.
- [x] Unknown filenames write nothing.
- [x] Duplicate file blocks write nothing.
- [x] Missing required blocks write nothing.
- [x] Empty generated file content writes nothing.
- [x] Existing files are not overwritten by default and are reported as skipped.
- [x] `--overwrite` replaces existing required files and reports overwritten files.
- [x] Writes are limited to `openspec/changes/<change-id>/`.
- [x] Missing change directory is created only after successful parse.
- [x] Existing target path that is not a directory fails before writes.
- [x] Validation runs after successful generation.
- [x] Validation warnings are reported in the result.
- [x] Validation errors are reported in the result and affect exit behavior through CLI mapping.
- [x] Unsafe change id fails before source reads and filesystem writes.
- [x] Source read errors write nothing.
- [x] Write errors are reported clearly.

## Phase 6: CLI

- [x] Add `--ai-assisted` parsing to `specharbor generate`.
- [x] Add required `--from-file <agent-output-file>` parsing for `--ai-assisted`.
- [x] Add optional `--overwrite` parsing for `--ai-assisted`.
- [x] Reject `--from-file` without `--ai-assisted`.
- [x] Reject `--overwrite` without `--ai-assisted`.
- [x] Reject `--ai-assisted` combined with `--blank`, `--template`, `--guided`, or `--agent-assisted`.
- [x] Reject `--ai-assisted --execute`.
- [x] Reject agent-assisted and guided-only flags when combined with `--ai-assisted`.
- [x] Reject duplicate `--ai-assisted`, `--from-file`, and `--overwrite` flags.
- [x] Preserve existing unsupported flag and extra argument behavior.
- [x] Wire the new use case with local filesystem and validation dependencies.
- [x] Print deterministic success output with change id, source file, target path, directory status, overwrite status, generated files, skipped files, overwritten files, validation status, validation counts, findings, and safety notes.
- [x] Print deterministic parse failure output with parse codes, messages, filenames, line numbers where available, and no-files-written status.
- [x] Map parser, source read, argument, and write errors to non-zero exit.
- [x] Map validation warnings-only results to exit `0`.
- [x] Map validation errors after successful writes to non-zero exit after printing the report.

## Phase 7: Documentation

- [x] Update `README.md` to list AI-assisted from-file generation once implemented.
- [x] Update `README.md` status and examples without claiming provider API integration.
- [x] Update `docs/usage.md` with command syntax, examples, output format, validation behavior, overwrite behavior, and safety boundaries.
- [x] Update `docs/generation-modes.md` to move AI-assisted generation from planned to implemented after this feature is implemented.
- [x] Update any existing AI-assisted generation docs if present.
- [x] Document the strict `---FILE: name---` and `---END FILE---` format.
- [x] Document that all five required file blocks are mandatory.
- [x] Document that existing files are skipped by default.
- [x] Document that `--overwrite` is explicit.
- [x] Document that validation runs after successful writes.
- [x] Document that malformed AI output writes nothing.
- [x] Document that no provider APIs, remote AI services, local model APIs, OAuth, credentials, production code writes, source-control automation, workflow automation, PR creation, merge, or archive behavior is introduced.

## Phase 8: CLI, Adapter, Regression, and Architecture Tests

- [x] Test CLI command parsing for the supported AI-assisted command.
- [x] Test missing `--from-file`.
- [x] Test unsupported flags.
- [x] Test duplicate flags.
- [x] Test extra positional args.
- [x] Test mode conflicts.
- [x] Test success output.
- [x] Test parse error output.
- [x] Test skipped existing output.
- [x] Test overwrite output.
- [x] Test validation report integration.
- [x] Test exit-code behavior for success, validation warnings, validation errors, parser errors, source read errors, and write errors.
- [x] Test local source AI output file reads.
- [x] Test missing source file reporting.
- [x] Test approved target writes.
- [x] Test target traversal prevention and absolute target path prevention.
- [x] Test blank generation remains unchanged.
- [x] Test template generation remains unchanged.
- [x] Test guided generation remains unchanged.
- [x] Test agent-assisted dry-run remains unchanged.
- [x] Test agent-assisted run-and-report remains unchanged.
- [x] Test validate remains unchanged except integration usage.
- [x] Test prompt, review, archive, config, scan, init, workflow, help, and version remain unchanged.
- [x] Add architecture tests proving core does not import adapters, CLI packages, `os`, `os/exec`, provider SDKs, network APIs, source-control SDKs, workflow SDKs, or external-agent SDKs.
- [x] Add architecture tests proving no provider/network/source-control/workflow/agent SDKs are introduced for this feature.
- [x] Add architecture tests proving no production code write path, arbitrary path write path, patch application abstraction, auto-commit, auto-push, PR, merge, or archive automation is introduced.

## Phase 9: Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-ai-assisted-generation`.
- [x] Manually verify successful AI-assisted generation from a strict local source file.
- [x] Manually verify malformed AI output writes nothing.
- [x] Manually verify default skip behavior for existing files.
- [x] Manually verify explicit overwrite behavior.
- [x] Run `git status --short`.
- [x] Inspect the implementation diff.
- [x] Update this `tasks.md` only for work actually completed.

## Phase 10: Symlink Overwrite Safety Fix

- [x] Reject AI-assisted generated target files that already exist as symlinks before planning writes.
- [x] Recheck AI-assisted generated target files immediately before create, skip, or overwrite actions.
- [x] Reject symlink parent directories on generated write paths so writes cannot escape through an existing symlinked change directory.
- [x] Keep the filesystem-specific `Lstat` and symlink checks inside the local filesystem adapter.
- [x] Add adapter tests for symlink target rejection, symlink parent rejection, missing-file behavior, regular-file overwrite, unsafe paths, and unchanged external symlink targets.
- [x] Add use case tests proving symlink safety failures write nothing and do not run validation.
- [x] Add CLI regression coverage for `--ai-assisted --from-file --overwrite` with a symlinked output file.
- [x] Update docs to state that symlink output targets are rejected.
