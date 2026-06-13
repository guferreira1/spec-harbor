# Tasks: Implement Local Context Retrieval

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm implementation scope is limited to deterministic local context retrieval, tests, documentation, and this change's task updates.
- [x] Confirm embeddings, vector databases, semantic providers, RAG answer generation, remote context, provider APIs, network APIs, command execution, prompt execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, publishing files, archive, and merge behavior remain out of scope.

## Domain

- [x] Add local context retrieval domain models under `internal/core/domain`.
- [x] Model retrieval query validation and deterministic normalization.
- [x] Model retrieval limits for query length, term count, source reads, snippets, results, and output size.
- [x] Model snippets with path, source category, evidence category, classification hints, score, and line range metadata.
- [x] Model retrieval reports and index dependency statuses.
- [x] Add deterministic lexical scoring rules.
- [x] Add deterministic source category and classification hint priority.
- [x] Add deterministic result ordering and tie-breaking.
- [x] Reuse or align with repository context index safe path and skip policy definitions.
- [x] Add tests proving retrieval models reject unsafe or unsupported inputs.

## Ports And Use Cases

- [x] Add or reuse small filesystem ports for stored index reads and bounded local source reads.
- [x] Add a use case that executes local context retrieval from an existing current repository context index.
- [x] Validate dependencies and project root.
- [x] Load and normalize `.specharbor/context-index.json`.
- [x] Rebuild current index metadata in memory only for freshness checks.
- [x] Fail safely for missing, invalid, unsupported-schema, stale, unreadable, and truncated indexes.
- [x] Select only safe indexed entries marked `supported_for_retrieval`.
- [x] Read bounded local source bytes through port boundaries.
- [x] Skip sensitive files, generated/heavy folders, symlinks, unsafe paths, and unsupported entries before content reads.
- [x] Extract bounded line-window snippets with line range metadata.
- [x] Apply deterministic scoring and tie-breaking.
- [x] Enforce max snippets per file, max total results, max source read bytes, max total source read bytes, and max output size.
- [x] Return structured reports without CLI formatting.
- [x] Ensure use cases never call external APIs, external commands, provider SDKs, workflow tools, source-control tools, package managers, shells, prompt execution, or agent tools.

## Filesystem Adapter

- [x] Implement concrete bounded local source reads in `internal/adapters/filesystem`.
- [x] Reuse existing safe path, symlink, and no-follow read helpers where practical.
- [x] Reject path traversal, absolute paths, Windows drive paths, null-byte paths, and paths outside the project root.
- [x] Reject symlink files and intermediate symlink path components.
- [x] Reject directories and non-regular files.
- [x] Enforce per-file read limits before reading.
- [x] Add adapter tests for bounded reads, symlink paths, unsafe paths, generated folder paths, and sensitive paths.

## CLI

- [x] Register `specharbor context retrieve`.
- [x] Preserve `specharbor context discover`.
- [x] Preserve `specharbor context index`.
- [x] Parse exactly `context retrieve --query <query>`.
- [x] Reject missing query, empty query, positional query arguments, duplicate query flags, unsupported flags, and unsupported positional arguments.
- [x] Reject unsupported flags including `--github`, `--remote`, `--rag`, `--embed`, `--provider`, `--json`, `--execute`, `--agent`, and `--deep`.
- [x] Obtain the project root from the current working directory.
- [x] Invoke the local context retrieval use case.
- [x] Print concise reports with normalized query, index path/status, result count, rank, path, category/evidence, score, line ranges, classification hints, and bounded snippets or metadata summaries.
- [x] Return non-zero exit codes for query validation and index dependency failures.
- [x] Keep CLI code free of retrieval, ranking, source eligibility, snippet, and safety rules.

## Documentation

- [x] Update `README.md` after implementation to list `specharbor context retrieve --query "<query>"`.
- [x] Update `docs/usage.md` with command syntax, required current index behavior, query rules, source set, snippet/result limits, report shape, and safety boundaries.
- [x] Update `docs/workflow.md` to mention retrieval as optional local/offline inspection over the context index.
- [x] Update `docs/agent-roles.md` only if needed to preserve prompt-context clarity.
- [x] Keep documentation explicit that retrieval is not embeddings, vector databases, RAG answer generation, remote context, provider APIs, command execution, prompt execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, or publishing behavior.

## Tests

- [x] Add domain tests for query validation and normalization.
- [x] Add domain tests for missing, empty, oversized, and term-heavy queries.
- [x] Add domain tests for scoring and deterministic tie-breaking.
- [x] Add domain tests for snippet limits and line range metadata.
- [x] Add use case tests for missing index behavior.
- [x] Add use case tests for invalid index behavior.
- [x] Add use case tests for unsupported schema behavior.
- [x] Add use case tests for stale index behavior.
- [x] Add use case tests for truncated index behavior.
- [x] Add use case tests for source file missing after index write.
- [x] Add use case tests for supported local source retrieval.
- [x] Add use case tests for package and build manifest retrieval.
- [x] Add use case tests for unsupported source skipping.
- [x] Add use case tests for sensitive file skipping.
- [x] Add use case tests for symlink traversal blocking.
- [x] Add use case tests for unsafe path rejection.
- [x] Add use case tests for bounded per-file reads.
- [x] Add use case tests for bounded total source reads.
- [x] Add use case tests for max snippets per file.
- [x] Add use case tests for max total results.
- [x] Add use case tests for max output size behavior.
- [x] Add tests proving no full-file dumps.
- [x] Add tests proving no secret content leaks.
- [x] Add tests proving no embeddings, vectors, RAG, provider calls, remote calls, project command execution, prompt execution, or agent execution.
- [x] Add CLI tests for successful retrieval reports.
- [x] Add CLI tests for no-results reports.
- [x] Add CLI tests for empty query failure.
- [x] Add CLI tests for oversized query failure.
- [x] Add CLI tests for missing/stale/invalid/truncated index failures.
- [x] Add CLI tests for unsupported flags and positional arguments.
- [x] Add regression tests for existing `specharbor brief`.
- [x] Add regression tests for existing `specharbor brief --update` parsing or compatible behavior.
- [x] Add regression tests for existing `specharbor context discover`.
- [x] Add regression tests for existing `specharbor context index`.
- [x] Add regression tests for context-aware `specharbor prompt`.
- [x] Add regression tests for existing `specharbor scan`.
- [x] Add regression tests for existing final decision labels.
- [x] Add or update architecture boundary tests if needed.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-local-context-retrieval`.
- [x] Run `go run ./cmd/specharbor`.
- [x] Run `go run ./cmd/specharbor version`.
- [x] Run `go run ./cmd/specharbor context discover`.
- [x] Run `go run ./cmd/specharbor context index`.
- [x] Run `go run ./cmd/specharbor context index --write`.
- [x] Run `go run ./cmd/specharbor context retrieve --query "architecture"`.
- [x] Run `go run ./cmd/specharbor context retrieve --query ""` and confirm it fails safely.
- [x] Run a retrieval query that returns no results safely.
- [x] Run retrieval against missing/stale/invalid/truncated index behavior using temporary fixtures or controlled local state.
- [x] Run `go run ./cmd/specharbor prompt implement-local-context-retrieval --role spec-author`.
- [x] Run `go run ./cmd/specharbor prompt implement-local-context-retrieval --role implementer`.
- [x] Run `go run ./cmd/specharbor scan`.
- [x] Remove generated `.specharbor/context-index.json` before commit unless a test fixture deliberately owns it.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff --stat`.
- [x] Inspect `git diff`.
