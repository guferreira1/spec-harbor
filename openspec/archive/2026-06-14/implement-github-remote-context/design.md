# Design: Implement GitHub Remote Context

## Overview

GitHub remote context adds an explicit remote retrieval path under the existing `context` command group. The implementation is read-only and bounded. It builds sibling remote-context models rather than reusing `.specharbor/context-index.json`, because remote results are transient evidence and must not be mixed into local generated state.

Dependency direction remains:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

## Command Contract

Supported:

```text
specharbor context github --repo owner/name --query "architecture"
specharbor context github --repo https://github.com/owner/name --query "architecture"
specharbor context github --repo owner/name --ref main --query "architecture"
specharbor context github --repo owner/name --query "architecture" --path docs
specharbor context github --repo owner/name --query "architecture" --path docs/architecture.md --path README.md
```

Rejected:

```text
specharbor context github
specharbor context github --repo owner/name
specharbor context github --query architecture
specharbor context github owner/name
specharbor context github --repo owner/name --query ""
specharbor context github --repo owner/name --query architecture extra
specharbor context github --repo owner/name --query architecture --json
specharbor context github --repo owner/name --query architecture --rag
specharbor context github --repo owner/name --query architecture --embed
specharbor context github --repo owner/name --query architecture --provider openai
specharbor context github --repo owner/name --query architecture --execute
specharbor context github --repo owner/name --query architecture --agent codex
specharbor context github --repo owner/name --query architecture --deep
```

CLI code parses flags, obtains the optional token from `SPECHARBOR_GITHUB_TOKEN`, wires the GitHub HTTP adapter and use case, formats the report, and maps structured failures to exit codes. CLI code must not own repository validation, source selection, scoring, snippet extraction, or token redaction rules.

## Domain Model

Add remote context domain concepts under `internal/core/domain`:

- `GitHubRepositoryLocator`
- `GitHubRemoteRef`
- `GitHubRemotePathFilter`
- `GitHubRemoteContextQuery`
- `GitHubRemoteContextLimits`
- `GitHubRemoteSource`
- `GitHubRemoteCandidate`
- `GitHubRemoteSnippet`
- `GitHubRemoteContextResult`
- `GitHubRemoteContextReport`
- `GitHubRemoteContextStatus`
- `GitHubRemoteContextErrorCode`

Domain responsibilities:

- validate repository locators, refs, query values, path filters, and limits;
- define the approved source set;
- define sensitive/generated skip policy reuse or alignment;
- define source categories and evidence categories;
- define lexical scoring weights;
- define deterministic ordering and tie-breaking;
- define output and snippet truncation rules;
- define safe redacted error messages.

Domain code must not import adapters, CLI packages, `os`, `net/http`, provider SDKs, source-control SDKs, shell/process packages, terminal IO, or workflow packages.

## Limits

Initial defaults:

- max repo locator chars: 200;
- max ref chars: 200;
- max query chars: 512;
- max query terms: 32;
- max path filters: 20;
- max path chars: 512;
- max files fetched: 50;
- max file size bytes: 128 KiB;
- max total file bytes: 1 MiB;
- max snippets per file: 2;
- max snippet chars: 600;
- max total results: 10;
- max rendered snippet/summary chars: 8,000;
- max tree entries scanned: 500;
- max directory depth: 4;
- max skipped records in report: 50;
- HTTP timeout: 10 seconds;
- max HTTP response bytes for metadata: 1 MiB;
- max HTTP response bytes for file content: 256 KiB.

Limits should be injectable in tests.

## Source Selection

The use case owns approved source traversal. It should build a deterministic source plan from fixed files and bounded directories:

- fixed files: `README.md`, `AGENTS.md`, `CONTRIBUTING.md`, `.specharbor/project-brief.md`, `openspec/project.md`, package/build/dependency manifests, Docker/compose, Makefile, and Taskfile files;
- bounded directories: `docs/` Markdown, `openspec/specs/` Markdown, `.specharbor/rules/` Markdown, and `.github/workflows/` YAML.

`--path` filters are applied after normalization and skip checks. They may select a fixed file, a supported directory prefix, or a supported file under a supported directory. They cannot introduce arbitrary sources.

Sensitive file checks must include `.env`, `.env.*`, `*.pem`, `*.key`, `id_rsa`, `id_ed25519`, `secrets.*`, `credentials.*`, `.npmrc`, `.pypirc`, and `.netrc`. Generated or heavy folder checks must include `.git`, `node_modules`, `dist`, `build`, `target`, `vendor`, `coverage`, `.tmp`, `.cache`, `.next`, `.nuxt`, `out`, `bin`, and `obj`.

## Ports

Add a small read-only GitHub remote context port under `internal/core/ports`.

Recommended shape:

```go
type GitHubRemoteContextReader interface {
    ResolveRepository(locator domain.GitHubRepositoryLocator) (domain.GitHubRemoteRepository, error)
    ResolveRef(locator domain.GitHubRepositoryLocator, ref domain.GitHubRemoteRef) (domain.GitHubRemoteResolvedRef, error)
    ListDirectory(locator domain.GitHubRepositoryLocator, resolved domain.GitHubRemoteResolvedRef, path string) ([]domain.GitHubRemoteEntry, error)
    ReadFile(locator domain.GitHubRepositoryLocator, resolved domain.GitHubRemoteResolvedRef, path string, maxBytes int64) (domain.GitHubRemoteFile, error)
}
```

Exact names may vary, but the port must be read-only. It must not expose mutation methods for commits, branches, PRs, issues, comments, labels, releases, tags, workflow runs, or repository writes.

## Use Case

Add a use case under `internal/core/usecase` that:

1. validates project-independent input;
2. normalizes repository, ref, query, and path filters through domain constructors;
3. resolves repository metadata and default branch when needed through the GitHub port;
4. resolves the requested/default ref to a commit SHA when GitHub provides it;
5. builds the approved source plan;
6. lists bounded supported directories through the port;
7. applies skip rules, path filters, file count limits, tree-entry limits, size limits, and total byte limits;
8. reads only eligible text-like files through the port;
9. extracts bounded snippets with 1-based line ranges;
10. scores and ranks results locally;
11. assembles a structured report with safe messages.

The use case must not import adapters, CLI packages, `os`, `net/http`, provider SDKs, source-control SDKs, shell/process packages, terminal IO, workflow packages, or agent runners.

## GitHub HTTP Adapter

Add a concrete adapter under `internal/adapters/github` or another clearly named package.

Responsibilities:

- use HTTPS requests to `https://api.github.com`;
- set an explicit `User-Agent`;
- attach `Authorization: Bearer <token>` only when `SPECHARBOR_GITHUB_TOKEN` is present;
- never log, print, persist, or return token contents;
- enforce request timeouts;
- enforce response body limits;
- decode only the REST responses needed for repository metadata, refs/commits, directory contents, and file contents;
- map rate limit, unauthorized, forbidden, not found, offline/network, invalid response, oversized response, and unsupported content errors to safe structured errors;
- perform no write requests.

Adapter tests should use fake HTTP transports or `httptest`; no real GitHub network is required.

## Candidate Fetching

Preferred implementation uses the GitHub Contents API for fixed files and bounded directory traversal. Directory traversal must be deterministic and stop when limits are hit. If the implementation uses tree APIs, it must enforce `max tree entries scanned` before selecting candidates and must still fetch only approved source paths.

Binary or unsupported files are skipped. Text decoding should be deterministic; invalid UTF-8 may be skipped as unsupported content or rendered with replacement if tests define that behavior.

The use case should stop candidate expansion before remote reads when `max files fetched`, `max tree entries scanned`, or path-filter constraints are exceeded. The resulting report should include bounded skip or truncation metadata without dumping candidate contents.

## Scoring And Ranking

Ranking is deterministic and local.

Allowed inputs:

- exact normalized phrase matches in snippet text;
- exact normalized phrase matches in path;
- unique query-token matches in snippet text;
- query-token matches in path or filename;
- heading-line matches;
- source category priority;
- classification or source hints where available.

Tie-breaking:

1. higher score;
2. higher source category priority;
3. lower path lexicographically;
4. lower starting line number;
5. lower snippet text lexicographically.

Forbidden inputs:

- embeddings;
- vector similarity;
- LLM or provider reranking;
- RAG answer generation;
- GitHub search ranking as final ordering.

## Error Behavior

Return safe user-facing errors for:

- invalid repo input;
- unsupported host;
- invalid ref input;
- invalid query;
- invalid path filter;
- unsupported flags;
- network unavailable;
- timeout;
- GitHub rate limit;
- unauthorized;
- forbidden;
- not found;
- invalid token;
- oversized metadata or file responses;
- too many candidates;
- unsupported file type;
- binary or invalid text content.

Errors must not include tokens, credentials, Authorization headers, credential-bearing URLs, raw API responses that may contain secrets, or large content bodies.

## Architecture Tests

Add architecture coverage proving:

- core/domain and core/usecase do not import `net/http`, adapters, CLI, `os/exec`, provider SDKs, or source-control SDKs;
- remote context ports are read-only and expose no mutation methods;
- local context commands do not import or construct the GitHub adapter;
- the GitHub adapter is the only production package in this feature that imports `net/http`;
- no shell, `git`, `gh`, RAG, embedding, vector, provider, prompt execution, or agent execution behavior is introduced.

## Documentation

Update:

- `README.md`: mention explicit GitHub remote context briefly.
- `docs/usage.md`: document syntax, token env var, source set, limits, output, errors, and safety boundaries.
- `docs/workflow.md`: mention GitHub remote context as optional inspection, not local context, confirmed context, prompt injection, or source-control automation.

Docs must state that network is used only by `specharbor context github`, local context commands remain offline, token use is optional and explicit, no GitHub writes happen, and no RAG/embeddings/provider behavior is implemented.

## Verification

Implementation verification should run:

```text
gofmt
go test ./...
go test -count=1 ./...
go run ./cmd/specharbor validate implement-github-remote-context
go run ./cmd/specharbor
go run ./cmd/specharbor version
go run ./cmd/specharbor context discover
go run ./cmd/specharbor context index
go run ./cmd/specharbor context retrieve --query "architecture"
go run ./cmd/specharbor context github --repo guferreira1/spec-harbor --query "architecture"
```

The live GitHub smoke check is optional if network is unavailable or rate-limited; unit and adapter tests must not require live network.
