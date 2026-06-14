# Tasks: Implement GitHub Remote Context

## Planning

- [x] Re-read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and this active change before implementation.
- [x] Confirm scope is limited to explicit read-only GitHub remote context, tests, documentation, and this change's task updates.
- [x] Confirm local/offline commands remain local and do not call GitHub.

## Domain

- [x] Add GitHub repository locator validation and normalization.
- [x] Add optional ref validation.
- [x] Add remote query validation and deterministic normalization.
- [x] Add remote path filter validation.
- [x] Add remote context limits.
- [x] Add approved remote source set and source categories.
- [x] Add sensitive/generated skip behavior aligned with local context.
- [x] Add remote candidate, snippet, result, report, status, and error models.
- [x] Add deterministic scoring and tie-breaking.
- [x] Add domain tests for repository, ref, query, path, limits, skip rules, scoring, ranking, and output bounds.

## Ports And Use Case

- [x] Add a small read-only GitHub remote context port under `internal/core/ports`.
- [x] Ensure the port exposes no GitHub mutation methods.
- [x] Add a use case that resolves repository metadata, resolves refs, builds bounded source candidates, fetches approved files, extracts snippets, ranks results, and returns a structured report.
- [x] Enforce max files, max file size, max total bytes, max snippets, max results, max rendered output, max tree entries, and max directory depth.
- [x] Handle no-results, invalid input, network failure, timeout, unauthorized, forbidden, not found, rate limit, oversized, unsupported content, and invalid token cases safely.
- [x] Add use-case tests with fake GitHub readers for success, no results, validation failures, skip rules, limits, source attribution, error mapping, and deterministic ranking.

## GitHub Adapter

- [x] Add a concrete read-only GitHub HTTP adapter outside core.
- [x] Use HTTPS `api.github.com` only.
- [x] Read optional token only from `SPECHARBOR_GITHUB_TOKEN` at the adapter/CLI boundary.
- [x] Never print, persist, or return token contents.
- [x] Set bounded timeouts and response body limits.
- [x] Map GitHub status codes and transport failures to safe errors.
- [x] Add adapter tests with fake transports or `httptest`; do not require live GitHub or tokens.

## CLI

- [x] Register `specharbor context github`.
- [x] Parse `--repo`, `--query`, optional `--ref`, and repeatable `--path`.
- [x] Reject missing required flags, duplicate scalar flags, positional arguments, and unsupported flags.
- [x] Preserve `context discover`, `context index`, and `context retrieve` behavior.
- [x] Format source-attributed remote reports with repository, ref/default branch, resolved SHA, rank, score, path, category/evidence, line range, bounded snippet or summary, and `Remote: yes`.
- [x] Return non-zero exit codes for validation and remote dependency failures.

## Documentation

- [x] Update `README.md` concisely after implementation.
- [x] Update `docs/usage.md` with command syntax, token env var, network behavior, source set, limits, output, safe errors, and non-goals.
- [x] Update `docs/workflow.md` to mention optional explicit remote inspection.
- [x] Keep docs clear that local context commands remain offline and remote context is not RAG, embeddings, provider behavior, prompt execution, agent execution, GitHub mutation, or source-control automation.

## Tests And Architecture

- [x] Add CLI tests for successful reports, no-results reports, invalid repo/ref/query/path inputs, unsupported flags, duplicate flags, and missing required flags.
- [x] Add tests for token redaction and invalid-token behavior.
- [x] Add tests for bounded candidate fetching, sensitive/generated skip rules, oversized files, unsupported/binary files, source attribution, line ranges, output size limits, and deterministic tie-breaking.
- [x] Add regression tests for `context discover`, `context index`, `context retrieve`, `brief`, `brief --update`, `prompt`, and `scan` where practical.
- [x] Add architecture tests proving core has no HTTP/adapter/exec/provider/source-control imports and remote ports/adapters expose no mutation behavior.

## Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-github-remote-context`.
- [x] Run `go run ./cmd/specharbor`.
- [x] Run `go run ./cmd/specharbor version`.
- [x] Run `go run ./cmd/specharbor context discover`.
- [x] Run `go run ./cmd/specharbor context index`.
- [x] Run local retrieval regression when a current local index fixture is available.
- [x] Run or explicitly skip a live `context github` smoke check depending on network/rate-limit availability.
- [x] Inspect `git status --short --untracked-files=all`.
- [x] Inspect `git diff --stat`.
- [x] Inspect `git diff`.
