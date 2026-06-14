# Tasks: Implement Context RAG Provider

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm scope is limited to explicit context RAG provider support, tests, documentation, and this change's task updates.
- [x] Confirm local/offline commands remain local and do not read provider credentials or call providers.
- [x] Confirm release automation, npm, Homebrew, `install.sh`, GoReleaser, publishing files, archive, merge, tags, and source-control automation remain out of scope.

## Domain

- [x] Add context RAG domain models under `internal/core/domain`.
- [x] Model provider names with `openai` as the only supported provider for this change.
- [x] Model source kinds for `local` and `github`.
- [x] Model query validation, source selection, source attribution, line ranges, truncation state, statuses, and limits.
- [x] Model provider request, provider response, and provider error codes.
- [x] Add deterministic source ordering and source limit behavior.
- [x] Add bounded provider prompt/context construction.
- [x] Add answer truncation and report construction.
- [x] Add domain tests for validation, limits, source ordering, prompt construction, provider request content, and report statuses.

## Ports And Use Case

- [x] Add a small context RAG provider port under `internal/core/ports`.
- [x] Ensure the provider port exposes no embedding, vector store, file upload, tool execution, shell execution, agent execution, or source-control methods.
- [x] Add a `GenerateContextRAGAnswer` use case under `internal/core/usecase`.
- [x] Inject source retrievers and provider interfaces so tests can use fakes.
- [x] Reuse existing local retrieval behavior for default local sources.
- [x] Reuse existing GitHub remote retrieval behavior only when GitHub sources are explicitly selected.
- [x] Enforce max sources, max snippet chars, max total context chars, max answer chars, max provider response bytes, timeout, and rendered output limits.
- [x] Return safe statuses for missing sources, missing credentials, provider timeout, auth failure, rate limit, network failure, malformed response, and oversized response.
- [x] Ensure the use case never writes `.specharbor/context-index.json`, retrieval caches, answer files, project files, prompts, commits, or remote state.
- [x] Ensure use cases never import adapters, CLI packages, `net/http`, `os`, `os/exec`, provider SDKs, workflow tools, source-control tools, package managers, shells, prompt execution, or agent tools.

## OpenAI Adapter

- [x] Add a concrete OpenAI context RAG provider adapter outside core.
- [x] Use a small HTTP implementation and no heavy provider SDK dependency.
- [x] Use the OpenAI Responses API with `store: false` when supported.
- [x] Use `gpt-5.4-mini` by default and allow `SPECHARBOR_OPENAI_MODEL` override.
- [x] Read the API key only from the explicit CLI/provider wiring path.
- [x] Never print, persist, return, or include the API key in errors or reports.
- [x] Enforce request timeout and response body limits.
- [x] Parse bounded `output_text` content only.
- [x] Map HTTP status codes, timeouts, rate limits, auth failures, oversized responses, malformed responses, and incomplete responses to safe provider errors.
- [x] Add adapter tests with fake HTTP transports or `httptest`; do not require live OpenAI network or tokens.

## CLI

- [x] Register `specharbor context rag`.
- [x] Parse `--query`, `--provider`, repeatable `--from`, optional `--repo`, optional `--ref`, repeatable `--path`, optional `--max-sources`, and optional `--max-answer-chars`.
- [x] Default omitted `--from` to local.
- [x] Reject missing query, empty query, missing provider, unsupported provider, duplicate scalar flags, unsupported `--from` values, invalid numeric limits, unsupported flags, and positional arguments.
- [x] Require `--repo` when `--from github` is selected.
- [x] Reject `--repo`, `--ref`, and `--path` unless GitHub sources are selected.
- [x] Read `SPECHARBOR_OPENAI_API_KEY` and `SPECHARBOR_OPENAI_MODEL` only for `context rag --provider openai`.
- [x] Fail safely before provider construction when the OpenAI API key is missing.
- [x] Format source-attributed RAG reports with provider, model, status, query, answer, source count, source list, local/remote markers, repository/ref metadata, line ranges, and truncation markers.
- [x] Preserve existing `context discover`, `context index`, `context retrieve`, and `context github` behavior.

## Documentation

- [x] Update `README.md` with the explicit optional RAG command and safety note.
- [x] Update `docs/usage.md` with syntax, source selection, OpenAI env vars, default model, limits, output, errors, and non-goals.
- [x] Update `docs/workflow.md` to mention explicit RAG answer generation as optional inspection, not prompt injection or confirmed context.
- [x] Keep documentation explicit that RAG is not local discovery, indexing, retrieval, prompt injection, embeddings, vector storage, agent execution, command execution, GitHub mutation, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, or publishing behavior.

## Tests

- [x] Add domain tests for context RAG query and source validation.
- [x] Add domain tests for provider request construction and prompt/context bounds.
- [x] Add use-case tests for local source selection.
- [x] Add use-case tests for optional GitHub source selection.
- [x] Add use-case tests for insufficient and missing sources.
- [x] Add use-case tests for fake provider success.
- [x] Add use-case tests for missing provider, unsupported provider, and missing API key.
- [x] Add use-case or adapter tests proving token redaction.
- [x] Add provider tests for timeout, network failure, rate limit, auth failure, oversized response, and malformed response.
- [x] Add CLI tests for successful fake-provider output with source attribution.
- [x] Add CLI tests for missing-token behavior.
- [x] Add CLI tests for unsupported flags, positional arguments, duplicate flags, source selection validation, GitHub-only flags, and numeric limits.
- [x] Add regression tests proving no provider call from `context discover`, `context index`, `context retrieve`, `brief`, `prompt`, `validate`, `review`, or `scan`.
- [x] Add tests proving no writes to `.specharbor/context-index.json`.
- [x] Add tests proving no prompt injection by default.
- [x] Add tests proving no source-control automation, shell execution, or agent execution.
- [x] Add architecture tests proving no provider SDK or `net/http` imports in core.
- [x] Add architecture tests proving no release/npm/Homebrew/install.sh/GoReleaser changes.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-context-rag-provider`.
- [x] Run `go run ./cmd/specharbor`.
- [x] Run `go run ./cmd/specharbor version`.
- [x] Run `go run ./cmd/specharbor context discover`.
- [x] Run `go run ./cmd/specharbor context index`.
- [x] Run `go run ./cmd/specharbor context retrieve --query "architecture" || true`.
- [x] Run `go run ./cmd/specharbor context github --repo guferreira1/spec-harbor --query "architecture" --path README.md || true`.
- [x] Run `go run ./cmd/specharbor context rag --query "architecture" --provider openai` and confirm missing-token behavior is safe when no token is configured.
- [x] Run `go run ./cmd/specharbor scan`.
- [x] Run `go run ./cmd/specharbor prompt implement-context-rag-provider --role spec-author`.
- [x] Run `go run ./cmd/specharbor prompt implement-context-rag-provider --role implementer`.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff --stat`.
- [x] Inspect `git diff`.
