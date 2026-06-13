# Tasks: Implement Repository Context Index

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm implementation scope is limited to repository context indexing, tests, documentation, and this change's task updates.
- [x] Confirm retrieval, snippet ranking, embeddings, vector databases, RAG, remote context, provider APIs, command execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, publishing files, archive, and merge behavior remain out of scope.

## Domain

- [x] Add repository context index domain models under `internal/core/domain`.
- [x] Model schema version, deterministic generation marker, project root marker, limits, entries, skipped records, truncation state, and reports.
- [x] Model index source categories aligned with existing context discovery categories where practical.
- [x] Model file type and language/ecosystem hints.
- [x] Model metadata-only entry validation with safe relative paths.
- [x] Model freshness/staleness status and stale reason records.
- [x] Add deterministic ordering and comparison rules.
- [x] Reuse or align skip policy definitions for sensitive files and heavy/generated folders.
- [x] Add tests proving raw file contents and snippets are not stored.

## Ports And Use Cases

- [x] Add small repository context index ports under `internal/core/ports`.
- [x] Add a use case that builds the current metadata-only index.
- [x] Add a use case path that safely writes the index only for `--write`.
- [x] Add a use case path that reads and checks the stored index only for `--check`.
- [x] Validate dependencies and project root.
- [x] Read and hash only supported local sources.
- [x] Apply skip rules before reads or directory traversal.
- [x] Enforce max indexed files, max individual file size, max total file bytes, directory depth, and max skipped records.
- [x] Ensure the generated index file is not indexed.
- [x] Return structured reports without CLI formatting.
- [x] Ensure use cases never call external APIs, external commands, provider SDKs, workflow tools, source-control tools, package managers, shells, prompt execution, or agent tools.

## Filesystem Adapter

- [x] Implement bounded metadata reads, directory listing, symlink rejection, safe relative path handling, hashing, index reads, and safe writes in `internal/adapters/filesystem`.
- [x] Reuse existing local filesystem safety helpers where practical.
- [x] Reject path traversal, absolute paths, Windows drive paths, null-byte paths, and paths that escape the project root.
- [x] Avoid following symlinks for files and intermediate directories.
- [x] Ensure safe writes preserve the original index on failure.
- [x] Add adapter tests for symlink paths, unsafe paths, hashing, size metadata, modified-time metadata, generated folder skipping, and safe writes.

## CLI

- [x] Register `specharbor context index`.
- [x] Preserve `specharbor context discover`.
- [x] Parse `context index`, `context index --write`, and `context index --check`.
- [x] Reject unsupported context subcommands.
- [x] Reject unsupported flags and positional arguments for `context index`.
- [x] Reject using `--write` and `--check` together.
- [x] Obtain the project root from the current working directory.
- [x] Invoke the repository context index use case.
- [x] Print concise reports for report, write, current, stale, missing, invalid, and truncated states.
- [x] Keep CLI code free of source selection, skip policy, staleness, hash, and classification rules.

## Documentation

- [x] Update `README.md` after implementation to list `specharbor context index`, `--write`, and `--check`.
- [x] Update `docs/usage.md` with command syntax, index path, generated-file policy, schema summary, supported sources, skip rules, limits, stale checks, and safety boundaries.
- [x] Update `docs/workflow.md` to mention the index as optional local inventory metadata for future context work.
- [x] Update `docs/agent-roles.md` only if needed to preserve prompt-context clarity.
- [x] Add or update `.gitignore` only for the generated `.specharbor/context-index.json` file if it is not already ignored.
- [x] Keep documentation explicit that indexing is not retrieval, snippet ranking, embeddings, vector databases, RAG, remote context, provider APIs, command execution, prompt execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, or publishing behavior.

## Tests

- [x] Add domain tests for index model validation.
- [x] Add domain tests for deterministic generation and stable ordering.
- [x] Add domain tests for source categories, file type hints, language/ecosystem hints, and classification hints.
- [x] Add use case tests for supported source inventory.
- [x] Add tests for sensitive file skipping.
- [x] Add tests for heavy/generated folder skipping.
- [x] Add tests for symlink traversal blocking.
- [x] Add tests for path traversal, absolute path, Windows drive path, and null-byte rejection.
- [x] Add tests for max file count behavior.
- [x] Add tests for max individual file size behavior.
- [x] Add tests for max total file bytes behavior.
- [x] Add tests for hash, size, and modified-time metadata correctness.
- [x] Add tests for stale detection when entries, hashes, sizes, mtimes, schema, limits, or truncation state change.
- [x] Add tests for missing and invalid stored indexes.
- [x] Add tests proving safe write behavior preserves the original index on failure.
- [x] Add tests proving the generated index file is not indexed.
- [x] Add tests proving no raw file contents are stored.
- [x] Add tests for empty repository behavior.
- [x] Add tests for ambiguous repository behavior.
- [x] Add CLI tests for report, write, check-current, check-stale, missing, invalid, unsupported flags, and unsupported positional arguments.
- [x] Add regression tests for existing `specharbor brief`.
- [x] Add regression tests for existing `specharbor brief --update` parsing or compatible behavior.
- [x] Add regression tests for existing `specharbor context discover`.
- [x] Add regression tests for context-aware `specharbor prompt`.
- [x] Add regression tests for existing `specharbor scan`.
- [x] Add regression tests for existing final decision labels.
- [x] Add or update architecture boundary tests if needed.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-repository-context-index`.
- [x] Run `go run ./cmd/specharbor`.
- [x] Run `go run ./cmd/specharbor version`.
- [x] Run `go run ./cmd/specharbor context discover`.
- [x] Run `go run ./cmd/specharbor prompt implement-repository-context-index --role spec-author`.
- [x] Run `go run ./cmd/specharbor prompt implement-repository-context-index --role implementer`.
- [x] Run `go run ./cmd/specharbor scan`.
- [x] Run `go run ./cmd/specharbor context index`.
- [x] Run `go run ./cmd/specharbor context index --write`.
- [x] Run `go run ./cmd/specharbor context index --check`.
- [x] Remove generated `.specharbor/context-index.json` before commit unless a test fixture deliberately owns it.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff --stat`.
- [x] Inspect `git diff`.
