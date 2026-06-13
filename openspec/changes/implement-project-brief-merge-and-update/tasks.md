# Tasks: Implement Project Brief Merge And Update

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm implementation scope is limited to project brief merge/update, tests, documentation, and this change's task updates.
- [x] Confirm indexing, retrieval, embeddings, RAG, remote context, provider APIs, command execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, and publishing files remain out of scope.

## Domain

- [x] Add project brief parsing support for known deterministic brief sections.
- [x] Add source category modeling for `user_confirmed_context`, `detected_fact`, and `suggested_assumption`.
- [x] Add project brief field models for confirmed values, detected fact candidates, suggested assumption candidates, conflicts, and stale records.
- [x] Add explicit merge decision models for keep, custom replacement, accept detected fact, accept suggested assumption when explicitly confirmed, ignore detected fact, keep stale assumptions, remove stale assumptions, and cancel.
- [x] Add conflict detection that prefers existing confirmed values by default.
- [x] Add stale record detection without automatic deletion.
- [x] Add deterministic update rendering that keeps confirmed context, detected facts, and assumptions separate.
- [x] Add tests proving detected facts and assumptions are not promoted without explicit confirmation.

## Ports And Use Cases

- [x] Extend or add a project brief update filesystem port for safe reads and safe replacement writes.
- [x] Add a project brief update use case under `internal/core/usecase`.
- [x] Inject or reuse the existing context discovery use case as the source of detected facts and suggested assumptions.
- [x] Build update proposals from the parsed brief and current discovery result.
- [x] Apply explicit merge decisions in the use case rather than the CLI.
- [x] Return structured preview/update result data for CLI formatting.
- [x] Require an explicit final-confirmed input in the write use case before any update write can occur.
- [x] Ensure update mode fails clearly when no existing brief is present.
- [x] Ensure safe write behavior preserves the original brief on write failure.

## CLI

- [x] Update `brief` argument parsing to accept exactly `--update`.
- [x] Preserve existing `specharbor brief` create-if-absent behavior.
- [x] Require an interactive TTY for `specharbor brief --update`.
- [x] Add deterministic update prompts for conflicts, detected facts, custom replacements, stale assumptions, and cancellation.
- [x] Show a clear preview before writing.
- [x] Require final confirmation before writing.
- [x] Ensure cancellation at any update prompt leaves `.specharbor/project-brief.md` unchanged.
- [x] Print a concise success report after an update write.
- [x] Keep CLI code limited to terminal interaction, formatting, and passing explicit decisions to use cases.

## Documentation

- [x] Update `README.md` to list `specharbor brief --update` only after implementation exists.
- [x] Update `docs/usage.md` with syntax, interactive update behavior, preview, confirmation, cancellation, conflict/stale handling, and source category safety.
- [x] Update `docs/workflow.md` to describe update as an optional context maintenance step.
- [x] Keep documentation explicit that update behavior is not indexing, retrieval, RAG, remote discovery, provider integration, command execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, or publishing behavior.

## Tests

- [x] Add domain tests for project brief parsing.
- [x] Add domain tests for source category preservation.
- [x] Add domain tests for merge decisions.
- [x] Add domain tests for conflict detection.
- [x] Add domain tests for stale record detection.
- [x] Add domain tests for deterministic Markdown rendering.
- [x] Add use case tests for no existing brief in update mode.
- [x] Add use case tests for update cancellation/no write.
- [x] Add use case tests proving final-confirmed false writes nothing.
- [x] Add use case tests for update confirmation/write.
- [x] Add use case tests for keeping current values.
- [x] Add use case tests for replacing with detected facts.
- [x] Add use case tests for custom values.
- [x] Add use case tests for ignoring detected facts.
- [x] Add use case tests proving detected facts are not promoted without confirmation.
- [x] Add use case tests proving suggested assumptions are not promoted without confirmation.
- [x] Add use case tests proving stale values are surfaced but not removed automatically.
- [x] Add use case or adapter tests proving write failure preserves the original file.
- [x] Add CLI tests for the interactive update flow.
- [x] Add CLI tests for final confirmation cancellation.
- [x] Update existing CLI tests so `brief --update` is accepted and unrelated flags remain rejected.
- [x] Add regression tests for existing `specharbor brief` behavior.
- [x] Add regression tests for `specharbor context discover`.
- [x] Add regression tests for context-aware `specharbor prompt`.
- [x] Add regression tests for `specharbor scan`.
- [x] Add or update architecture tests to guard against CLI-owned merge rules and forbidden behavior.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-project-brief-merge-and-update`.
- [x] Run `go run ./cmd/specharbor`.
- [x] Run `go run ./cmd/specharbor version`.
- [x] Run `go run ./cmd/specharbor brief`.
- [x] Run `go run ./cmd/specharbor context discover`.
- [x] Run `go run ./cmd/specharbor prompt implement-project-brief-merge-and-update --role spec-author`.
- [x] Run `go run ./cmd/specharbor prompt implement-project-brief-merge-and-update --role implementer`.
- [x] Run `go run ./cmd/specharbor scan`.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff --stat`.
- [x] Inspect `git diff`.

## PR Review Follow-up

- [x] Harden project brief update reads so symlink parent directories and symlink targets are rejected before parsing.
- [x] Add regression tests proving unsafe project brief reads do not print or use outside project content.
