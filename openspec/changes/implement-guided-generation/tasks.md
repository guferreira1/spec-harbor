# Tasks: Implement Guided Generation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-guided-generation/`.
- [x] Inspect the existing blank generation implementation before editing.
- [x] Inspect the existing built-in template generation implementation before editing.
- [x] Inspect current CLI parsing and reporting for `specharbor generate`.
- [x] Inspect current generation domain concepts, ports, use case, filesystem adapter, template/content adapters, and tests.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep the implementation limited to deterministic non-interactive guided generation through explicit CLI flags.
- [x] Preserve existing `specharbor generate <change-id> --blank` behavior.
- [x] Preserve existing `specharbor generate <change-id> --template <template-name>` behavior for `feature`, `bugfix`, `docs`, and `refactor`.
- [x] Do not implement interactive prompts, AI-assisted generation, agent-assisted generation, hybrid generation, remote templates, custom templates, config-driven templates, template marketplace behavior, provider setup, source-control integration, workflow integration, network access, or external execution.
- [x] Do not modify README files, docs outside this OpenSpec change, CI files, `.github/workflows/ci.yml`, or `.specharbor/config.yml`.
- [x] Do not change `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, or CI behavior.

## Phase 1: Domain Concepts

- [x] Treat `domain.GuidedMode` as an implemented generation mode for `specharbor generate`.
- [x] Add or reuse a domain value for supported guided types.
- [x] Support only `feature`, `bugfix`, `docs`, and `refactor` as valid guided types.
- [x] Reject empty guided type values with a clear error.
- [x] Reject unknown guided type values with a clear error.
- [x] Add or reuse validation for required guided title input.
- [x] Add or reuse validation for required guided summary input.
- [x] Trim guided type, title, and summary for validation.
- [x] Ensure generated guided content uses the normalized title and summary values that passed validation.
- [x] Keep unsafe change id validation consistent with existing blank and template generation behavior.
- [x] Reuse the shared required OpenSpec change file policy, currently `domain.RequiredOpenSpecChangeFiles()`, for guided generation.
- [x] Extend structured generation results to represent guided mode and selected guided type when guided generation succeeds.
- [x] Preserve existing template result behavior and selected template reporting data.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, external-agent tooling, workflow SDKs, external process execution, and concrete template engines.

## Phase 2: Ports

- [x] Reuse or extend the generation filesystem port so guided generation performs all filesystem checks and writes through ports.
- [x] Preserve write-if-absent behavior through the filesystem port.
- [x] Add a small guided content port under `internal/core/ports` if one does not already exist.
- [x] Ensure the guided content port can provide content for a selected guided type, title, summary, and required filename.
- [x] Ensure the guided content port supports clear unknown-guided-type validation before target writes.
- [x] Ensure the guided content port returns clear errors for unknown required filenames.
- [x] Keep guided content ports separate from prompt rendering, validation, project initialization, AI providers, source-control integrations, workflow dispatchers, config repositories, terminal IO, and external execution.
- [x] Ensure the generation use case depends only on domain concepts and ports.

## Phase 3: Guided Content Adapter

- [x] Add or extend a concrete adapter, likely under `internal/adapters/templates`, for deterministic guided content.
- [x] Reuse the existing built-in template content strategy where it keeps guided content consistent and simple.
- [x] Provide guided `feature` content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Provide guided `bugfix` content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Provide guided `docs` content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Provide guided `refactor` content for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Ensure every guided `proposal.md` includes the supplied title and summary.
- [x] Ensure every guided `design.md` includes the supplied title and summary or uses them as explicit context for the design starter.
- [x] Ensure every guided `tasks.md` includes the supplied title and summary as context without marking any task complete.
- [x] Ensure every guided `acceptance-criteria.md` includes the supplied title and summary or derives criteria context from them.
- [x] Ensure every guided `risks.md` includes the supplied title and summary or derives risk context from them.
- [x] Ensure guided `feature` content remains feature-oriented and includes sections appropriate for a new capability.
- [x] Ensure guided `bugfix` content remains bugfix-oriented and includes sections appropriate for current behavior, expected behavior, impact, regression testing, and risk.
- [x] Ensure guided `docs` content remains documentation-oriented and includes sections appropriate for audience, source of truth, accuracy, validation, and documentation-only scope.
- [x] Ensure guided `refactor` content remains refactor-oriented and includes sections appropriate for boundaries, compatibility, behavior preservation, testing, and risk.
- [x] Ensure all generated guided `tasks.md` content uses unchecked tasks only.
- [x] Ensure all guided content is deterministic, local, human-readable, generic, and safe to commit.
- [x] Ensure guided content does not claim implementation, verification, review, or archiving has already happened.
- [x] Return clear adapter errors for unknown guided types and unknown required filenames.
- [x] Do not add runtime template discovery, custom template paths, remote templates, provider prompts, agent prompts, network access, source-control access, workflow access, config integration, terminal prompts, or external command execution.

## Phase 4: Generation Use Case Updates

- [x] Update the generation use case under `internal/core/usecase` to support guided generation.
- [x] Validate that required use case dependencies are present for blank, template, and guided modes.
- [x] Validate project root the same way blank and template generation do.
- [x] Validate change id the same way blank and template generation do.
- [x] Validate that guided generation has a non-empty guided type.
- [x] Validate that guided generation has a non-empty title.
- [x] Validate that guided generation has a non-empty summary.
- [x] Validate the selected guided type before target directory or file writes.
- [x] Reject unknown guided types with a clear error before target directory or file writes.
- [x] Reject unsafe change ids before filesystem writes.
- [x] Preserve existing OpenSpec project availability checks.
- [x] Preserve the existing clear error when `openspec/project.md` or `openspec/changes/` is missing.
- [x] Preserve the existing behavior that generation does not create `openspec/`, `openspec/project.md`, or `openspec/changes/`.
- [x] Build the change relative path as `openspec/changes/<change-id>`.
- [x] Create the target change directory when missing.
- [x] Continue without error when the target change directory already exists.
- [x] Treat partially existing change directories as recoverable.
- [x] Obtain required filenames from the shared domain required-file policy.
- [x] Obtain starter content from the guided content port for each selected guided type and required file.
- [x] Pass the validated title and summary to guided content generation.
- [x] Write each missing required file through write-if-absent behavior.
- [x] Skip existing files without overwriting content.
- [x] Record created files and skipped existing files in the structured result.
- [x] Return selected guided type in the structured result for CLI reporting.
- [x] Keep output printing out of the use case.
- [x] Keep providers, agents, source-control APIs, workflow connectors, network calls, external process execution, terminal prompts, and config access out of the use case.
- [x] Avoid adding unused public strategy registries, factories, chains, provider abstractions, agent abstractions, source-control abstractions, workflow abstractions, remote template abstractions, or marketplace abstractions.

## Phase 5: CLI Parsing and Reporting

- [x] Update `internal/adapters/cli` so `specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"` invokes guided generation.
- [x] Preserve `specharbor generate <change-id> --blank` parsing and output behavior.
- [x] Preserve `specharbor generate <change-id> --template <template-name>` parsing and output behavior.
- [x] Reject `--guided` without `--type` with a clear missing-type error.
- [x] Reject `--guided` without `--title` with a clear missing-title error.
- [x] Reject `--guided` without `--summary` with a clear missing-summary error.
- [x] Reject `--guided --type` without a following type value with a clear error.
- [x] Reject `--guided --title` without a following title value with a clear error.
- [x] Reject `--guided --summary` without a following summary value with a clear error.
- [x] Reject empty guided type, title, or summary values when representable by the parser.
- [x] Reject unknown guided types with a clear error.
- [x] Reject commands that provide both `--guided` and `--blank`.
- [x] Reject commands that provide both `--guided` and `--template`.
- [x] Preserve the existing error for commands that provide both `--blank` and `--template`.
- [x] Reject unsupported flags.
- [x] Reject extra positional arguments.
- [x] Reject duplicate `--guided`, `--type`, `--title`, and `--summary` flags where the current parser supports duplicate validation.
- [x] Preserve existing duplicate `--blank` and duplicate `--template` validation.
- [x] Preserve unsafe change id error behavior.
- [x] Obtain the current working directory as project root in the CLI adapter.
- [x] Construct the generation use case with the local filesystem adapter, blank content adapter, built-in template content adapter, and guided content adapter.
- [x] Print a concise deterministic success report for guided generation.
- [x] Include the selected guided type in the guided generation report.
- [x] Include the change id, relative path, directory status, created count, skipped existing count, created files, and skipped files in the report.
- [x] Include the title in the guided generation report.
- [x] Keep report formatting in the CLI adapter.
- [x] Avoid absolute local paths, debug output, provider details, agent details, source-control details, workflow details, network details, prompts, and validation summaries in the report.
- [x] Return argument and execution errors without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to existing process bootstrapping unless a minimal error-handling adjustment is strictly required.

## Phase 6: Tests

- [x] Add domain tests for guided generation mode behavior.
- [x] Add domain tests for supported guided type validation.
- [x] Add domain tests proving only `feature`, `bugfix`, `docs`, and `refactor` are valid guided types.
- [x] Add domain or use case tests for missing guided type errors.
- [x] Add domain or use case tests for missing guided title errors.
- [x] Add domain or use case tests for missing guided summary errors.
- [x] Add guided content adapter tests for every supported guided type and every required file.
- [x] Test that each generated guided file has non-empty useful Markdown content.
- [x] Test that guided output includes the supplied title.
- [x] Test that guided output includes the supplied summary.
- [x] Test that generated guided `tasks.md` files contain unchecked tasks only.
- [x] Test that unknown guided types return clear errors.
- [x] Test that unknown required filenames return clear errors.
- [x] Add use case tests for `--guided --type feature` behavior.
- [x] Add use case tests for `--guided --type bugfix` behavior.
- [x] Add use case tests for `--guided --type docs` behavior.
- [x] Add use case tests for `--guided --type refactor` behavior.
- [x] Test that guided generation creates the target change directory when missing.
- [x] Test that guided generation creates all required files when missing.
- [x] Test that guided generation creates only missing files in an existing change directory.
- [x] Test that existing files are skipped and their contents are preserved exactly.
- [x] Test that unknown guided types are rejected before target writes occur.
- [x] Test that missing guided inputs are rejected before target writes occur.
- [x] Test that unsafe change ids are rejected before filesystem writes and consistently with blank and template generation.
- [x] Test that missing OpenSpec project structure is rejected before target writes.
- [x] Test that blank generation still behaves exactly as before.
- [x] Test that template generation still behaves exactly as before.
- [x] Add CLI tests for successful guided `feature` generation.
- [x] Add CLI tests for successful guided `bugfix` generation.
- [x] Add CLI tests for successful guided `docs` generation.
- [x] Add CLI tests for successful guided `refactor` generation.
- [x] Add CLI tests for clear errors when `--guided` is missing `--type`.
- [x] Add CLI tests for clear errors when `--guided` is missing `--title`.
- [x] Add CLI tests for clear errors when `--guided` is missing `--summary`.
- [x] Add CLI tests for clear errors on unknown guided types.
- [x] Add CLI tests for clear errors when `--guided` and `--blank` are provided together.
- [x] Add CLI tests for clear errors when `--guided` and `--template` are provided together.
- [x] Add CLI tests for unsupported flags and extra positional arguments.
- [x] Add CLI tests for duplicate conflicting flags where supported by the current parser.
- [x] Add CLI tests or regression coverage proving `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `help`, `version`, and unknown command behavior are not changed.

## Phase 7: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor generate <change-id> --blank` remains unchanged.
- [x] Manually verify `specharbor generate <change-id> --template feature` remains unchanged.
- [x] Manually verify `specharbor generate <change-id> --template bugfix` remains unchanged.
- [x] Manually verify `specharbor generate <change-id> --template docs` remains unchanged.
- [x] Manually verify `specharbor generate <change-id> --template refactor` remains unchanged.
- [x] Manually verify `specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"` creates deterministic guided feature starter files.
- [x] Manually verify `specharbor generate <change-id> --guided --type bugfix --title "<title>" --summary "<summary>"` creates deterministic guided bugfix starter files.
- [x] Manually verify `specharbor generate <change-id> --guided --type docs --title "<title>" --summary "<summary>"` creates deterministic guided docs starter files.
- [x] Manually verify `specharbor generate <change-id> --guided --type refactor --title "<title>" --summary "<summary>"` creates deterministic guided refactor starter files.
- [x] Manually verify guided output includes the supplied title and summary.
- [x] Manually verify running guided generation again skips existing files and does not overwrite their contents.
- [x] Manually verify an existing partial change directory is completed by creating only missing required files.
- [x] Manually verify unknown guided types return a clear error.
- [x] Manually verify missing `--type` returns a clear error.
- [x] Manually verify missing `--title` returns a clear error.
- [x] Manually verify missing `--summary` returns a clear error.
- [x] Manually verify providing both `--guided` and `--blank` returns a clear error.
- [x] Manually verify providing both `--guided` and `--template` returns a clear error.
- [x] Manually verify unsupported flags and extra arguments return clear errors.
- [x] Manually verify unsafe change ids are rejected before filesystem writes.
- [x] Manually verify no interactive prompts are introduced.
- [x] Manually verify no provider APIs, local model APIs, external agents, source-control APIs, workflow tools, network APIs, external processes, or config integrations are used.
- [x] Inspect `git status --short`, `git diff --stat`, `git diff --name-only`, and `git diff`.
- [x] Confirm no README, docs outside this OpenSpec change, CI, `.github/workflows/ci.yml`, or `.specharbor/config.yml` changes are included.
- [x] Update this `tasks.md` by checking off only implementation tasks actually completed.
