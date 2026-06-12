# Tasks: Implement Project Briefing Foundation

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/spec-author.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm the implementation scope is limited to `specharbor brief`, project brief domain/use case/ports/adapters, tests, docs, and this change's task updates.
- [x] Confirm RAG, embeddings, repository indexing, GitHub remote discovery, prompt injection, release, publishing, source-control automation, and existing brief update behavior remain out of scope.

## Domain And Rendering

- [x] Add project briefing domain types under `internal/core/domain`.
- [x] Model confirmed answers separately from detected context and assumptions.
- [x] Add deterministic question definitions with three to five options per question.
- [x] Ensure every question definition ends with `Other / custom`.
- [x] Add project brief command fields for install, test, build, and run.
- [x] Add preferred missing-context agent behavior to the brief model.
- [x] Add deterministic Markdown rendering for `.specharbor/project-brief.md`.
- [x] Ensure rendered Markdown includes the required `# Project Brief` structure and all required sections.
- [x] Ensure rendered Markdown labels user-provided answers, detected context, and assumptions separately.
- [x] Ensure assumptions and detected suggestions are never rendered as confirmed facts.

## Ports And Use Case

- [x] Add a small briefing filesystem port under `internal/core/ports`.
- [x] Add a project briefing use case under `internal/core/usecase`.
- [x] Validate that all required answers are present before rendering.
- [x] Create `.specharbor/` only after user confirmation.
- [x] Write `.specharbor/project-brief.md` with write-if-absent behavior.
- [x] Return a clear error when `.specharbor/project-brief.md` already exists.
- [x] Avoid merge, update, overwrite, append, and existing brief parsing behavior.
- [x] Keep terminal IO, concrete filesystem access, provider APIs, network APIs, source-control APIs, workflow tools, agent tools, and external command execution out of the use case.

## CLI

- [x] Register the new `brief` command in the CLI command registry.
- [x] Reject all `brief` positional arguments and flags.
- [x] Require an interactive TTY before prompting.
- [x] Print the initial project-context briefing message.
- [x] Prompt for project type, purpose, target users, stack, architecture, install command, test command, build command, run command, and missing-context agent behavior.
- [x] Support custom non-empty answers through `Other / custom` for every question.
- [x] Retry invalid choices and invalid custom answers up to three attempts.
- [x] Treat detected context as suggestions only until the user confirms an answer.
- [x] Print a deterministic pre-write summary with `.specharbor/project-brief.md`.
- [x] Require confirmation before invoking the write use case.
- [x] Treat `n`, `no`, empty input, and EOF as cancellation.
- [x] Ensure cancellation and retry exhaustion write no file.
- [x] Print a concise success report after the file is written.

## Documentation

- [x] Update `README.md` after implementation to list `specharbor brief` as available.
- [x] Update `docs/usage.md` with syntax, prompt behavior, confirmation, cancellation, existing-file behavior, and output file details.
- [x] Update `docs/workflow.md` if the implemented command should be documented as an optional preparation step.
- [x] Keep documentation clear that briefing is not RAG, indexing, prompt injection, provider integration, agent execution, or source-control automation.
- [x] Do not change release, npm, Homebrew, `install.sh`, publishing, or package-manager documentation for this feature.

## Tests

- [x] Add unit tests for project brief data/model construction.
- [x] Add unit tests for source category validation.
- [x] Add unit tests for Markdown rendering.
- [x] Add tests proving Markdown output is deterministic.
- [x] Add tests proving generated Markdown separates user-provided answers, detected context, and assumptions.
- [x] Add use case tests for successful project brief creation.
- [x] Add use case tests for existing file refusal.
- [x] Add CLI command tests for successful brief creation.
- [x] Add CLI command tests for custom answers.
- [x] Add CLI command tests for confirmation denied.
- [x] Add CLI command tests ensuring no file is written when cancelled.
- [x] Add CLI command tests for non-TTY behavior.
- [x] Add CLI command tests for invalid choices and retry exhaustion.
- [x] Add tests proving detected context hints remain suggestions until confirmed.
- [x] Add regression tests ensuring existing CLI commands still work.
- [x] Add architecture tests or extend existing boundary tests if needed.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-project-briefing-foundation`.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff -- openspec/changes/implement-project-briefing-foundation/`.
