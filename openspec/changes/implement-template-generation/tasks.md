# Tasks: Implement Template Generation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-template-generation/`.
- [x] Inspect the existing blank generation implementation before editing.
- [x] Inspect current CLI parsing and reporting for `specharbor generate`.
- [x] Inspect current generation domain concepts, ports, use case, filesystem adapter, template/content adapter, and tests.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep the implementation limited to built-in template generation for `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Preserve existing `specharbor generate <change-id> --blank` behavior.
- [x] Do not implement guided, AI-assisted, agent-assisted, hybrid, custom template, remote template, source-control, workflow, provider, interactive, or config-backed generation.
- [x] Do not modify README files, docs outside this OpenSpec change, CI files, `.github/workflows/ci.yml`, or `.specharbor/config.yml`.
- [x] Do not change `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, or CI behavior.

## Phase 1: Domain Concepts

- [x] Add or extend generation mode values in `internal/core/domain` to represent template generation.
- [x] Add or extend a domain value for built-in template names.
- [x] Support only `feature`, `bugfix`, `docs`, and `refactor` as valid built-in template names.
- [x] Reject empty, unknown, or unsupported template names through clear domain or use case validation.
- [x] Keep unsafe change id validation consistent with existing blank generation behavior.
- [x] Reuse the shared required OpenSpec change file policy, currently `domain.RequiredOpenSpecChangeFiles()`, for the files created by template generation.
- [x] Include the selected template name in structured generation results when template generation succeeds, if the existing result shape supports extension cleanly.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, external-agent tooling, workflow SDKs, and concrete template engines.

## Phase 2: Ports

- [x] Reuse or extend the generation filesystem port so template generation performs all filesystem checks and writes through ports.
- [x] Keep write-if-absent behavior available through the filesystem port.
- [x] Add a small template content port under `internal/core/ports` if one does not already exist.
- [x] Ensure the template content port can provide content for a selected built-in template and required filename.
- [x] Ensure the template content port can report supported template names or otherwise allow clear unknown-template validation before target writes.
- [x] Keep template content ports separate from prompt rendering, validation, project initialization, AI providers, source-control integrations, workflow dispatchers, config repositories, and external execution.
- [x] Ensure the generation use case depends only on domain concepts and ports.

## Phase 3: Template Content Adapter

- [x] Add or extend a concrete adapter under `internal/adapters/templates` for deterministic built-in template content.
- [x] Provide `feature` template content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Provide `bugfix` template content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Provide `docs` template content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Provide `refactor` template content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Ensure `feature` proposal content includes sections for Summary, Problem, Proposed Solution, Scope, Out of Scope, and Success Criteria.
- [x] Ensure `feature` design content includes Architecture Notes, Domain, Ports, Use Case, Adapters, CLI, Testing, and Validation sections.
- [x] Ensure `feature` tasks content includes baseline, domain, ports, use case, adapters, CLI, tests, and verification tasks.
- [x] Ensure `feature` acceptance criteria focus on behavior and scope.
- [x] Ensure `feature` risks focus on scope creep, architecture boundaries, and backwards compatibility.
- [x] Ensure `bugfix` proposal content includes Current Behavior, Expected Behavior, Impact, Scope, Out of Scope, and Success Criteria.
- [x] Ensure `bugfix` design content includes Root Cause, Fix Approach, Boundaries, Regression Testing, and Validation sections.
- [x] Ensure `bugfix` tasks content includes reproduce, test, fix, regression, and verification tasks.
- [x] Ensure `bugfix` acceptance criteria focus on corrected behavior and no regressions.
- [x] Ensure `bugfix` risks focus on regression risk and over-fixing.
- [x] Ensure `docs` proposal content includes Documentation Goal, Audience, Files to Update, Scope, Out of Scope, and Success Criteria.
- [x] Ensure `docs` design content includes Documentation Structure, Source of Truth, Accuracy Rules, and Validation sections.
- [x] Ensure `docs` tasks content includes inventory, README/docs updates, command verification, and Markdown verification tasks.
- [x] Ensure `docs` acceptance criteria focus on Markdown-only scope and command accuracy.
- [x] Ensure `docs` risks focus on stale docs and overstating planned behavior.
- [x] Ensure `refactor` proposal content includes Refactor Goal, Current Pain, Non-Functional Goal, Scope, Out of Scope, and Success Criteria.
- [x] Ensure `refactor` design content includes Boundaries, Migration Plan, Compatibility, Testing, and Validation sections.
- [x] Ensure `refactor` tasks content includes baseline tests, small refactor steps, regression tests, and verification tasks.
- [x] Ensure `refactor` acceptance criteria focus on unchanged external behavior.
- [x] Ensure `refactor` risks focus on accidental behavior changes and broad diffs.
- [x] Ensure all generated `tasks.md` content uses unchecked tasks only.
- [x] Ensure all template content is deterministic, generic, human-readable, and safe to commit.
- [x] Return clear adapter errors for unknown template names and unknown required filenames.
- [x] Do not add runtime template discovery, custom template paths, remote templates, provider prompts, agent prompts, network access, source-control access, workflow access, or config integration.

## Phase 4: Generation Use Case Updates

- [x] Update the generation use case under `internal/core/usecase` to support template generation.
- [x] Validate that required use case dependencies are present.
- [x] Validate project root the same way blank generation does.
- [x] Validate change id the same way blank generation does.
- [x] Validate that template generation has a non-empty selected template name.
- [x] Validate the selected template name before target directory or file writes.
- [x] Reject unknown template names with a clear error before target directory or file writes.
- [x] Preserve existing OpenSpec project availability checks.
- [x] Preserve the existing clear error when `openspec/project.md` or `openspec/changes/` is missing.
- [x] Preserve the existing behavior that generation does not create `openspec/`, `openspec/project.md`, or `openspec/changes/`.
- [x] Build the change relative path as `openspec/changes/<change-id>`.
- [x] Create the target change directory when missing.
- [x] Continue without error when the target change directory already exists.
- [x] Treat partially existing change directories as recoverable.
- [x] Obtain required filenames from the shared domain required-file policy.
- [x] Obtain starter content from the template content port for each selected template and required file.
- [x] Write each missing required file through write-if-absent behavior.
- [x] Skip existing files without overwriting content.
- [x] Record created files and skipped existing files in the structured result.
- [x] Keep output printing out of the use case.
- [x] Keep providers, agents, source-control APIs, workflow connectors, network calls, external process execution, and config access out of the use case.
- [x] Avoid adding unused public strategy registries, factories, chains, provider abstractions, agent abstractions, source-control abstractions, workflow abstractions, or remote template abstractions.

## Phase 5: CLI Parsing and Reporting

- [x] Update `internal/adapters/cli` so `specharbor generate <change-id> --template <template-name>` invokes template generation.
- [x] Preserve `specharbor generate <change-id> --blank` parsing and output behavior.
- [x] Reject `specharbor generate <change-id> --template` with a clear missing-template-name error.
- [x] Reject `specharbor generate <change-id> --template ""` or equivalent empty template input when representable by the parser.
- [x] Reject unknown template names with a clear error.
- [x] Reject commands that provide both `--blank` and `--template`.
- [x] Reject unsupported flags.
- [x] Reject extra positional arguments.
- [x] Reject duplicate generation mode flags.
- [x] Preserve unsafe change id error behavior.
- [x] Obtain the current working directory as project root in the CLI adapter.
- [x] Construct the generation use case with the local filesystem adapter and built-in template content adapter.
- [x] Print a concise deterministic success report for template generation.
- [x] Include the selected template name in the template generation report.
- [x] Include the change id, relative path, directory status, created count, skipped existing count, created files, and skipped files in the report.
- [x] Avoid absolute local paths, debug output, provider details, agent details, source-control details, workflow details, network details, and validation summaries in the report.
- [x] Return argument and execution errors without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to existing process bootstrapping unless a minimal error-handling adjustment is strictly required.

## Phase 6: Tests

- [x] Add domain tests for template generation mode and built-in template name validation.
- [x] Add template content adapter tests for every supported template and every required file.
- [x] Test that each generated template file has non-empty useful Markdown content.
- [x] Test that generated template `tasks.md` files contain unchecked tasks only.
- [x] Test that unknown template names return clear errors.
- [x] Test that unknown required filenames return clear errors.
- [x] Add use case tests for `--template feature` behavior.
- [x] Add use case tests for `--template bugfix` behavior.
- [x] Add use case tests for `--template docs` behavior.
- [x] Add use case tests for `--template refactor` behavior.
- [x] Test that template generation creates the target change directory when missing.
- [x] Test that template generation creates all required files when missing.
- [x] Test that template generation creates only missing files in an existing change directory.
- [x] Test that existing files are skipped and their contents are preserved.
- [x] Test that unknown templates are rejected before target writes occur.
- [x] Test that missing template names are rejected clearly.
- [x] Test that unsafe change ids are rejected before filesystem writes and consistently with blank generation.
- [x] Test that missing OpenSpec project structure is rejected before target writes.
- [x] Test that blank generation still behaves exactly as before.
- [x] Add CLI tests for successful `--template feature` generation.
- [x] Add CLI tests for successful `--template bugfix` generation.
- [x] Add CLI tests for successful `--template docs` generation.
- [x] Add CLI tests for successful `--template refactor` generation.
- [x] Add CLI tests for clear errors on unknown templates.
- [x] Add CLI tests for clear errors on missing template names.
- [x] Add CLI tests for clear errors when `--blank` and `--template` are provided together.
- [x] Add CLI tests for unsupported flags and extra arguments.
- [x] Add CLI tests or regression coverage proving `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `help`, `version`, and unknown command behavior are not changed.

## Phase 7: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor generate <change-id> --blank` remains unchanged.
- [x] Manually verify `specharbor generate <change-id> --template feature` creates deterministic feature starter files.
- [x] Manually verify `specharbor generate <change-id> --template bugfix` creates deterministic bugfix starter files.
- [x] Manually verify `specharbor generate <change-id> --template docs` creates deterministic docs starter files.
- [x] Manually verify `specharbor generate <change-id> --template refactor` creates deterministic refactor starter files.
- [x] Manually verify running template generation again skips existing files and does not overwrite their contents.
- [x] Manually verify an existing partial change directory is completed by creating only missing required files.
- [x] Manually verify unknown template names return a clear error.
- [x] Manually verify missing template names return a clear error.
- [x] Manually verify providing both `--blank` and `--template` returns a clear error.
- [x] Manually verify unsupported flags and extra arguments return clear errors.
- [x] Manually verify unsafe change ids are rejected before filesystem writes.
- [x] Manually verify no provider APIs, local model APIs, external agents, source-control APIs, workflow tools, network APIs, external processes, or config integrations are used.
- [x] Inspect `git status --short`, `git diff --stat`, `git diff --name-only`, and `git diff`.
- [x] Confirm no README, docs outside this OpenSpec change, CI, `.github/workflows/ci.yml`, or `.specharbor/config.yml` changes are included.
- [x] Update this `tasks.md` by checking off only implementation tasks actually completed.
