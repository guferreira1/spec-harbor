# Design: Implement Context RAG Provider

## Overview

This change adds a narrow provider-backed answer generation path over existing bounded context retrieval results. The command is explicit:

```text
specharbor context rag --query "<query>" --provider openai
```

The design intentionally keeps retrieval and generation separate:

- local retrieval remains deterministic and offline;
- GitHub remote retrieval remains explicit and read-only;
- RAG answer generation happens only after the user invokes `context rag`;
- generated answers are never confirmed project context;
- provider requests and responses are not persisted.

Dependency direction remains:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

## Command Contract

Supported:

```text
specharbor context rag --query "architecture" --provider openai
specharbor context rag --query "architecture" --provider openai --from local
specharbor context rag --query "architecture" --provider openai --from github --repo guferreira1/spec-harbor
specharbor context rag --query "architecture" --provider openai --from local --from github --repo guferreira1/spec-harbor
specharbor context rag --query "architecture" --provider openai --from github --repo guferreira1/spec-harbor --ref main
specharbor context rag --query "architecture" --provider openai --from github --repo guferreira1/spec-harbor --path README.md
specharbor context rag --query "architecture" --provider openai --max-sources 6 --max-answer-chars 3000
```

Rejected:

```text
specharbor context rag
specharbor context rag --query "architecture"
specharbor context rag --provider openai
specharbor context rag "architecture" --provider openai
specharbor context rag --query "" --provider openai
specharbor context rag --query "architecture" --provider anthropic
specharbor context rag --query "architecture" --provider openai --from remote
specharbor context rag --query "architecture" --provider openai --from github
specharbor context rag --query "architecture" --provider openai --repo owner/name
specharbor context rag --query "architecture" --provider openai --ref main
specharbor context rag --query "architecture" --provider openai --path README.md
specharbor context rag --query "architecture" --provider openai --execute
specharbor context rag --query "architecture" --provider openai --agent codex
specharbor context rag --query "architecture" --provider openai --write
specharbor context rag --query "architecture" --provider openai --json
specharbor context rag --query "architecture" --provider openai --embed
```

CLI code owns parsing, environment reads, dependency wiring, report formatting, and exit-code mapping only. It must not own prompt construction, source normalization, provider request models, provider error mapping, or answer bounds.

## Context Sources

V1 supports both existing source families, but only through explicit rules:

- Local is the default when no `--from` is provided.
- Local uses the existing `RetrieveLocalContext` use case and requires a current `.specharbor/context-index.json`.
- GitHub is used only when `--from github --repo owner/name` is present.
- GitHub uses the existing `RetrieveGitHubRemoteContext` use case.
- `--ref` and `--path` are passed only to GitHub retrieval.
- Local and GitHub results are merged after retrieval into a single ordered source list.
- No new crawler, indexer, embedding search, or provider reranking is introduced.

Source selection:

1. run selected retrieval use cases;
2. reject or report retrieval dependency failures safely;
3. flatten retrieval results into RAG sources;
4. keep local/remote markers;
5. keep path, source category, evidence category, line range, repository/ref metadata when available, score, and bounded snippet or summary;
6. sort deterministically by source selection order, retrieval rank, source marker, repository, path, and line start;
7. enforce `MaxSources`, `MaxSnippetChars`, and `MaxTotalContextChars`.

The RAG command never writes `.specharbor/context-index.json`, never persists GitHub remote context, and never persists generated provider answers.

## Domain Model

Add `internal/core/domain/context_rag.go`.

Recommended concepts:

- `ContextRAGProviderName`
- `ContextRAGSourceKind`
- `ContextRAGSource`
- `ContextRAGQuery`
- `ContextRAGLimits`
- `ContextRAGProviderRequest`
- `ContextRAGProviderResponse`
- `ContextRAGProviderErrorCode`
- `ContextRAGReport`
- `ContextRAGStatus`

Provider names:

- supported: `openai`
- unsupported provider values fail before any provider or network call.

Source kinds:

- `local`
- `github`

Report statuses:

- `answered`
- `insufficient_sources`
- `missing_sources`
- `missing_credentials`
- `provider_failed`
- `provider_timeout`
- `provider_rate_limited`
- `provider_unauthorized`
- `provider_response_invalid`
- `provider_response_oversized`

Provider error model:

- missing credentials;
- unauthorized or forbidden;
- rate limit;
- timeout;
- network failure;
- malformed response;
- oversized response;
- content filtered or incomplete when identifiable;
- unsupported provider.

The domain owns:

- query validation;
- provider name validation;
- source kind validation;
- source normalization and deterministic ordering;
- limit validation;
- source text truncation;
- provider request construction;
- answer truncation;
- report assembly.

Domain code must not import adapters, CLI packages, `os`, `net/http`, provider SDKs, source-control SDKs, `os/exec`, workflow packages, terminal IO, or agent runner packages.

## Limits

Initial defaults:

- max query chars: `512`;
- max query terms reused from retrieval: `32`;
- default max sources: `8`;
- hard max sources: `20`;
- max snippet chars per source: `600`;
- max total context chars: `6000`;
- default max answer chars: `4000`;
- hard max answer chars: `8000`;
- max provider response bytes: `65536`;
- provider timeout: `20 seconds`;
- max rendered output chars: `12000`;
- max path filters: reuse GitHub limit `20`;
- max remote files: reuse GitHub remote context limit `50`;
- max local retrieval results reused: governed by local retrieval and `MaxSources`.

All limits must be injectable or settable in tests. CLI flags may reduce `MaxSources` and `MaxAnswerChars`; they must not exceed hard caps.

## Provider Port

Add `internal/core/ports/context_rag_provider.go`.

Recommended interface:

```go
type ContextRAGProvider interface {
    GenerateContextAnswer(request domain.ContextRAGProviderRequest) (domain.ContextRAGProviderResponse, error)
}
```

The request model includes:

- provider name;
- model name;
- user query;
- bounded rendered context;
- structured source metadata;
- max answer chars or provider output hint;
- timeout hint if needed;
- instructions to answer only from sources;
- instructions to say when sources are insufficient.

The response model includes:

- answer text;
- provider name;
- model name;
- status or finish reason when safe;
- output-truncated marker;
- safe usage metadata only if available and non-sensitive.

The provider port is generation-only. It must not expose embeddings, vector search, file upload, conversation persistence, tool execution, source-control methods, shell methods, or agent execution methods.

## Use Case

Add `internal/core/usecase/generate_context_rag_answer.go`.

Responsibilities:

1. validate dependencies and project root;
2. validate query, provider, source selections, and limits;
3. obtain local retrieval results when local is selected;
4. obtain GitHub remote retrieval results only when GitHub is explicitly selected;
5. convert selected retrieval results into normalized RAG sources;
6. fail safely if no sources are available;
7. build the bounded provider request;
8. call the provider port exactly once;
9. map provider errors to safe structured statuses;
10. bound and normalize the answer text;
11. return a structured report without CLI formatting.

The use case may depend on small interfaces for local and GitHub retrieval so tests can use fakes. It must not import concrete adapters, CLI packages, `os`, `net/http`, provider SDKs, source-control SDKs, shell/process packages, workflow packages, or agent runners.

## Prompt And Context Construction

The provider request must include a bounded instruction set and a bounded source block.

The instruction must include:

- answer only from the supplied sources;
- cite source ids inline or in a concise references section;
- say `The provided sources are insufficient to answer this question.` when the sources do not support an answer;
- do not infer unsupported project facts;
- do not treat generated answer text as confirmed project context;
- do not request or reveal hidden chain of thought;
- do not execute commands;
- do not propose file writes as completed work.

The source block must include for each source:

- source id;
- source kind: local or github;
- local/remote marker;
- repository and ref metadata when remote;
- path;
- source category or evidence category;
- line range when available;
- score when available;
- truncation marker when content was trimmed;
- bounded snippet or summary.

The source block must not include:

- raw full repository contents;
- unbounded files;
- skipped sensitive files;
- provider secrets;
- GitHub tokens;
- OpenAI API keys;
- shell output;
- generated provider responses from prior runs;
- hidden prompts.

## OpenAI Adapter

Add a concrete provider adapter outside core, preferably:

```text
internal/adapters/openai/
```

Implementation rules:

- use a small HTTP adapter rather than a heavy SDK;
- POST to the OpenAI Responses API;
- set `Authorization: Bearer <token>` only when the explicit RAG command provided a token;
- set an explicit `User-Agent`;
- use `gpt-5.4-mini` by default;
- allow `SPECHARBOR_OPENAI_MODEL` to override the model;
- set `store: false` when supported by the request shape;
- do not request hosted tools;
- do not upload files;
- do not create vector stores;
- do not create conversations;
- enforce HTTP timeout and response body size limits;
- parse only bounded text output from `output_text` content;
- map HTTP 401/403, 429, timeout, network, oversized response, malformed response, and incomplete response to safe errors;
- never include request bodies, response bodies, authorization headers, or tokens in errors.

Adapter tests must use fake transports or `httptest` and require no live OpenAI network or token.

## Credentials

Credential rules:

- CLI reads `SPECHARBOR_OPENAI_API_KEY` only after command parsing confirms `context rag --provider openai`.
- CLI reads `SPECHARBOR_OPENAI_MODEL` only after command parsing confirms `context rag --provider openai`.
- Missing API key returns a safe error before any provider HTTP request.
- Token values are never printed, persisted, or included in structured reports.
- Token values are never passed into source context.
- Local/offline commands do not read OpenAI environment variables.

## Output

CLI output should be concise and source-attributed:

```text
Context RAG answer:
Provider: openai
Model: gpt-5.4-mini
Status: answered
Query: architecture
Sources: 3
Output truncated: no

Answer:
...

Sources:
1. local README.md lines 10-14
   Category: readme
   Truncated: no
2. github guferreira1/spec-harbor README.md lines 20-25
   Ref: main
   Remote: yes
   Category: readme
   Truncated: no
```

Output must include:

- answer;
- provider name;
- model name;
- status;
- source count;
- source list;
- path/source type;
- local/remote marker;
- repository/ref metadata for remote sources;
- line range when available;
- truncation markers;
- safe detail for missing credentials or provider errors.

Output must not include:

- API keys;
- raw provider request;
- raw provider response;
- Authorization headers;
- credential-bearing URLs;
- unbounded source snippets;
- hidden reasoning or chain-of-thought text.

## Safety Boundaries

This change explicitly forbids:

- provider calls from local/offline commands;
- provider calls during `prompt`;
- provider calls during `validate`;
- provider calls during `review`;
- provider calls during `scan`;
- automatic code changes;
- automatic project file writes from provider answers;
- automatic commits, pushes, PRs, merges, tags, releases, or archive work;
- GitHub mutation;
- shell execution;
- agent execution;
- vector database persistence;
- embedding generation or storage;
- unbounded RAG;
- using generated answers as confirmed project context;
- logging secrets.

## Architecture

Expected files:

```text
internal/core/domain/context_rag.go
internal/core/ports/context_rag_provider.go
internal/core/usecase/generate_context_rag_answer.go
internal/adapters/openai/http_context_rag_provider.go
internal/adapters/cli/context_discovery.go
internal/architecture/context_rag_provider_boundaries_test.go
```

Existing local and GitHub retrieval files may be updated only to expose reusable result conversion or parser wiring when necessary.

Forbidden dependencies:

- `internal/core/domain` must not import `net/http`, adapters, CLI, provider SDKs, `os`, or `os/exec`.
- `internal/core/ports` must not import `net/http`, provider SDKs, adapters, or CLI.
- `internal/core/usecase` must not import `net/http`, provider SDKs, adapters, CLI, `os`, or `os/exec`.
- Provider HTTP details stay inside `internal/adapters/openai`.
- GitHub HTTP details stay inside `internal/adapters/github`.

## Documentation

Update public docs after implementation exists:

- `README.md`: add a concise optional RAG command mention and safety note.
- `docs/usage.md`: document syntax, source selection, OpenAI env vars, default model, limits, output, errors, and non-goals.
- `docs/workflow.md`: mention RAG as optional explicit answer generation over retrieved context, not prompt injection or confirmed context.

Docs must not touch release automation, npm publishing, Homebrew, `install.sh`, GoReleaser, package release docs, or publishing workflows.

## Testing Strategy

Add focused tests for:

- command parsing;
- missing provider;
- unsupported provider;
- missing API key;
- token redaction;
- local source selection;
- GitHub source selection;
- `--repo`, `--ref`, and `--path` only being valid with GitHub sources;
- bounded prompt/context construction;
- insufficient or missing sources;
- provider success with a fake provider;
- provider timeout;
- provider network failure;
- provider rate limit;
- provider auth failure;
- provider oversized response;
- provider malformed response;
- answer output with source attribution;
- no provider call from local/offline commands;
- no writes to `.specharbor/context-index.json`;
- no prompt injection by default;
- no source-control automation;
- no shell/exec;
- no provider imports in core;
- no `net/http` in core;
- no release/npm/Homebrew/install.sh/GoReleaser changes.

Verification should run:

```text
gofmt
go test ./...
go test -count=1 ./...
go run ./cmd/specharbor validate implement-context-rag-provider
go run ./cmd/specharbor
go run ./cmd/specharbor version
go run ./cmd/specharbor context discover
go run ./cmd/specharbor context index
go run ./cmd/specharbor context retrieve --query "architecture"
go run ./cmd/specharbor context github --repo guferreira1/spec-harbor --query "architecture" --path README.md
go run ./cmd/specharbor context rag --query "architecture" --provider openai
go run ./cmd/specharbor scan
go run ./cmd/specharbor prompt implement-context-rag-provider --role spec-author
go run ./cmd/specharbor prompt implement-context-rag-provider --role implementer
```

The RAG command smoke check is expected to verify missing-token behavior unless a live token is intentionally provided. Live provider smoke is optional and must be reported separately.
