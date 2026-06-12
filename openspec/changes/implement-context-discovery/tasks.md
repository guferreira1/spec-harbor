# Tasks: Implement Context Discovery

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm implementation scope is limited to local context discovery, optional briefing suggestions, tests, docs, and this change's task updates.
- [x] Confirm repository-wide indexing, embeddings, RAG, remote discovery, prompt injection, brief merge/update behavior, release flows, publishing flows, source-control automation, provider APIs, agent execution, and command execution remain out of scope.

## Domain Model

- [x] Add context discovery domain types under `internal/core/domain`.
- [x] Model context signal kinds for project type, purpose, stack, frameworks, architecture, package manager, commands, documentation sources, agent instruction sources, OpenSpec sources, CLI entrypoints, containers, and workflows.
- [x] Model signal classifications for `detected_fact`, `suggested_assumption`, and `user_confirmed_context`.
- [x] Model confidence levels and validate that every signal has confidence.
- [x] Model source categories and source evidence for every signal.
- [x] Add deterministic ordering rules for grouped discovery output.
- [x] Add skip policy definitions for sensitive files.
- [x] Add skip policy definitions for heavy and generated folders.
- [x] Ensure assumptions cannot be constructed or rendered as detected facts.

## Discovery Use Case And Ports

- [x] Add a small context discovery filesystem port under `internal/core/ports`.
- [x] Reuse or extend the local filesystem adapter with safe read/list behavior required by discovery.
- [x] Add a context discovery use case under `internal/core/usecase`.
- [x] Validate project root and required dependencies.
- [x] Read only supported repository files and directories.
- [x] Apply skip rules before file reads or directory traversal.
- [x] Avoid following symlinks and preserve existing path safety behavior.
- [x] Build deterministic source snapshots for discovery rules.
- [x] Classify detected facts, suggested assumptions, and user-confirmed context.
- [x] Assign confidence and source evidence to every signal.
- [x] Prefer `.specharbor/project-brief.md` data as user-confirmed context when present.
- [x] Return structured results without CLI formatting.
- [x] Ensure the use case never calls external APIs, external commands, provider SDKs, workflow tools, source-control tools, or agent tools.

## Discovery Rules

- [x] Detect agent instruction sources from `AGENTS.md` and `.specharbor/rules/`.
- [x] Detect documentation sources from `README.md`, `CONTRIBUTING.md`, and bounded files under `docs/`.
- [x] Detect purpose summary only when clearly available from supported project context files.
- [x] Detect OpenSpec sources from `openspec/project.md` and `openspec/specs/`.
- [x] Detect confirmed context from `.specharbor/project-brief.md`.
- [x] Detect Go context from `go.mod`.
- [x] Detect Node context, package manager hints, and scripts from `package.json`.
- [x] Detect Java context from `pom.xml` and `build.gradle`.
- [x] Detect Rust context from `Cargo.toml`.
- [x] Detect Python context from `pyproject.toml` and `requirements.txt`.
- [x] Detect container signals from `Dockerfile` and `docker-compose.yml`.
- [x] Detect explicit test, build, and run commands from `Makefile` when present.
- [x] Detect explicit test, build, and run tasks from `Taskfile.yml` when present.
- [x] Detect GitHub Actions workflow sources from `.github/workflows/` without remote API calls.
- [x] Emit conventional command suggestions only as `suggested_assumption` when no explicit command fact exists.
- [x] Emit ambiguity notes instead of collapsing conflicting evidence into unsupported facts.

## CLI

- [x] Register `specharbor context discover`.
- [x] Reject `specharbor context` without `discover`.
- [x] Reject unsupported context subcommands.
- [x] Reject positional arguments and flags for `specharbor context discover`.
- [x] Obtain the project root from the current working directory.
- [x] Invoke the context discovery use case.
- [x] Print grouped user-confirmed context, detected facts, suggested assumptions, and notes.
- [x] Print `- none detected` for empty groups.
- [x] Keep CLI code free of discovery rules, source classification rules, confidence rules, and skip policy rules.

## Brief Integration

- [x] Run context discovery from `specharbor brief` only after TTY validation.
- [x] Present detected facts and assumptions as suggestions only.
- [x] Preserve the existing three-to-five option rule and `Other / custom` final option for every question.
- [x] Require user confirmation through selection or custom input before any suggestion becomes a project brief answer.
- [x] Pass detected context sources separately from confirmed answers when creating a new project brief.
- [x] Preserve existing cancellation, retry, confirmation, write-if-absent, and existing-brief refusal behavior.
- [x] Keep missing or ambiguous context on the existing interactive briefing path instead of blocking abruptly.

## Documentation

- [x] Update `README.md` after implementation to list `specharbor context discover`.
- [x] Update `docs/usage.md` with command syntax, output shape, supported sources, classifications, confidence levels, skip rules, and safety boundaries.
- [x] Update `docs/workflow.md` to mention context discovery as an optional preparation step.
- [x] Keep documentation clear that discovery is local/offline and not RAG, indexing, remote discovery, provider integration, prompt injection, agent execution, command execution, or brief update behavior.
- [x] Do not change release, npm, Homebrew, `install.sh`, GoReleaser, publishing, or package-manager documentation for this feature.

## Tester Follow-up Fixes

- [x] Add component-by-component symlink checks for context discovery file reads and directory traversal.
- [x] Add regression coverage for symlinked `openspec`, symlinked `docs`, supported symlink files, and normal supported files.
- [x] Detect explicit command evidence from `README.md`, `CONTRIBUTING.md`, and bounded Markdown files under `docs/`.
- [x] Detect explicit architecture evidence from bounded Markdown files under `docs/`.
- [x] Update stale `docs/usage.md` project brief wording for detected context and assumptions.

## Change Reviewer Follow-up Fixes

- [x] Make Markdown purpose summary detection conservative so generic README and CONTRIBUTING procedure text does not become `purpose_summary`.
- [x] Disable Markdown purpose summary extraction for `CONTRIBUTING.md` while preserving documentation and command detection.
- [x] Require explicit Markdown architecture evidence with bounded matching so vague SOLID/MVC prose does not become `architecture_hint`.
- [x] Add regression tests for vague Markdown false positives and explicit purpose and architecture evidence.

## Tests

- [x] Add unit tests for context signal modeling.
- [x] Add unit tests for classification validation.
- [x] Add unit tests for confidence and source classification.
- [x] Add unit tests for deterministic signal ordering.
- [x] Add unit tests for detecting Go project context from `go.mod`.
- [x] Add unit tests for detecting Node project context from `package.json`.
- [x] Add unit tests for detecting Java project context from `pom.xml` or `build.gradle`.
- [x] Add unit tests for detecting Python project context from `pyproject.toml` or `requirements.txt`.
- [x] Add unit tests for detecting Rust project context from `Cargo.toml`.
- [x] Add tests for command detection from package and build files.
- [x] Add tests for documentation-source detection.
- [x] Add tests for agent-instruction-source detection.
- [x] Add tests for OpenSpec-source detection.
- [x] Add tests for skipping secrets.
- [x] Add tests for skipping generated and heavy folders.
- [x] Add tests for deterministic discovery output.
- [x] Add tests for project brief precedence when `.specharbor/project-brief.md` exists.
- [x] Add CLI/usecase tests for missing context.
- [x] Add CLI/usecase tests for ambiguous context.
- [x] Add regression tests ensuring existing `specharbor brief` behavior still works.
- [x] Add regression tests ensuring existing CLI commands still work.
- [x] Add architecture boundary tests or extend existing tests if needed.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-context-discovery`.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff -- openspec/changes/implement-context-discovery/`.
