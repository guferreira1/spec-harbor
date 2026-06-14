# Proposal: Implement Context RAG Provider

## Problem

SpecHarbor now has confirmed project brief maintenance, a metadata-only local repository context index, deterministic local context retrieval, and explicit read-only GitHub remote context retrieval. Users can inspect source-attributed snippets, but there is no explicit provider-backed path that asks a model to answer a question using only those bounded retrieval results.

Adding provider integration is useful, but it is also the highest-risk context feature. It must not turn local/offline commands into networked commands, require provider keys for agent-assisted workflows, silently inject model output into role prompts, store generated answers, persist vectors, or treat provider output as confirmed project context.

## Goal

Add an explicit optional RAG/provider command:

```text
specharbor context rag --query "<query>" --provider openai
```

The first version must generate a bounded, source-attributed answer from existing retrieval results only. It must be clear that this command is a provider integration path, not local context discovery, indexing, retrieval, prompt generation, source-control automation, agent execution, or a persistent vector database.

## Command Shape

Supported first-version forms:

```text
specharbor context rag --query "<query>" --provider openai
specharbor context rag --query "<query>" --provider openai --from local
specharbor context rag --query "<query>" --provider openai --from github --repo owner/name
specharbor context rag --query "<query>" --provider openai --from local --from github --repo owner/name
specharbor context rag --query "<query>" --provider openai --from github --repo owner/name --ref <ref>
specharbor context rag --query "<query>" --provider openai --from github --repo owner/name --path README.md
specharbor context rag --query "<query>" --provider openai --max-sources 8
specharbor context rag --query "<query>" --provider openai --max-answer-chars 4000
```

Decision:

- `context rag` is explicit and separate from `context retrieve` so local retrieval remains local/offline.
- `--provider` is required and only `openai` is supported in this change.
- `--query` is required and follows the existing 512-character query boundary.
- `--from` is repeatable and supports `local` and `github`.
- Omitting `--from` means `--from local`.
- `--from github` requires `--repo owner/name`.
- `--repo`, `--ref`, and `--path` are valid only when GitHub is selected.
- `--path` is repeatable and narrows only approved GitHub sources.
- `--max-sources` and `--max-answer-chars` may reduce output bounds, but must not raise hard safety caps.

This shape is safe and practical because it reuses existing local and GitHub retrieval capabilities instead of introducing embeddings, vector stores, or arbitrary crawling.

## Scope

- Add the explicit `specharbor context rag` command.
- Use deterministic local retrieval results by default from `.specharbor/context-index.json`.
- Optionally include GitHub remote retrieval results only when `--from github --repo owner/name` is explicit.
- Reuse existing local and GitHub retrieval use cases for source collection.
- Add domain models for RAG queries, sources, limits, provider requests, provider responses, provider status, and source attribution.
- Add a small provider port under `internal/core/ports`.
- Add a use case under `internal/core/usecase` that builds bounded source context, calls the provider port, and returns a structured report.
- Add an OpenAI provider adapter outside core using a small HTTP client and no heavy SDK dependency.
- Read `SPECHARBOR_OPENAI_API_KEY` only when the explicit RAG command is invoked.
- Support optional `SPECHARBOR_OPENAI_MODEL`; default to `gpt-5.4-mini`.
- Use the OpenAI Responses API through the adapter.
- Do not persist provider requests, provider responses, generated answers, embeddings, vectors, or remote context.
- Render source-attributed answer output.
- Add focused tests with fake providers and fake HTTP transports.
- Update `README.md`, `docs/usage.md`, and `docs/workflow.md` after implementation.
- Update this change's `tasks.md` only for completed work.

## Out Of Scope

- Changing `specharbor context discover`.
- Changing `specharbor context index`.
- Changing `specharbor context retrieve`.
- Changing `specharbor context github` except for shared parsing helpers if needed.
- Provider calls from `specharbor brief`.
- Provider calls from `specharbor prompt`.
- Provider calls from `specharbor validate`.
- Provider calls from `specharbor review`.
- Provider calls from `specharbor scan`.
- Automatic prompt injection.
- Automatic code changes.
- Automatic file writes from generated answers.
- Persistent vector databases.
- Embedding generation or embedding storage.
- Provider reranking.
- Local model support.
- Multiple providers beyond `openai`.
- Arbitrary remote crawling.
- Shell execution.
- Agent execution.
- Source-control automation.
- GitHub mutation.
- Release automation.
- npm changes.
- Homebrew changes.
- `install.sh` changes.
- GoReleaser changes.
- Publishing flows.
- Archiving this OpenSpec change before merge.

## Provider Policy

The first supported provider is `openai`.

Credential and model configuration:

```text
SPECHARBOR_OPENAI_API_KEY
SPECHARBOR_OPENAI_MODEL
```

Rules:

- `SPECHARBOR_OPENAI_API_KEY` is optional until `context rag --provider openai` executes.
- Missing `SPECHARBOR_OPENAI_API_KEY` returns a safe missing-credential error.
- The API key is never printed.
- The API key is never persisted.
- The API key is never included in reports.
- The API key is never included in errors.
- The API key is not read by local/offline commands.
- The API key is read only at the CLI/provider-adapter boundary.
- `SPECHARBOR_OPENAI_MODEL` is optional; when absent, use `gpt-5.4-mini`.
- The model name is not a secret and may be shown in reports.
- Core does not import OpenAI SDKs, `net/http`, or provider adapter packages.

## Safety Model

The RAG command must:

- require an explicit user invocation;
- collect only bounded retrieval results;
- keep source attribution with every source snippet;
- instruct the provider to answer only from supplied sources;
- instruct the provider to say when the sources are insufficient;
- treat the provider answer as generated output, not confirmed project context;
- print answer status and source list;
- avoid raw request dumps;
- avoid raw unsafe or oversized response dumps;
- write no files;
- mutate no remote systems;
- execute no shell, `git`, `gh`, package manager, workflow, prompt, or agent command.

## Compatibility

Existing commands remain unchanged and local/offline by default:

- `specharbor context discover`
- `specharbor context index`
- `specharbor context retrieve`
- `specharbor brief`
- `specharbor prompt`
- `specharbor validate`
- `specharbor review`
- `specharbor scan`

`specharbor context github` remains explicit and read-only. RAG may call it only through the new explicit `context rag --from github` path.

## Success Criteria

- `specharbor context rag --query "<query>" --provider openai` is implemented.
- The command defaults to local retrieval sources only.
- Optional GitHub sources are included only with explicit `--from github --repo owner/name`.
- Provider calls happen only from `context rag`.
- Missing provider, unsupported provider, and missing API key fail safely.
- Local retrieval dependency failures remain actionable and do not write `.specharbor/context-index.json`.
- GitHub dependency failures remain safe and do not mutate or persist remote data.
- Provider request context is bounded, source-attributed, and excludes secrets.
- Provider answers are bounded and source-attributed in output.
- Provider output is not persisted and is not treated as confirmed context.
- No prompt injection happens by default.
- Architecture boundaries are preserved.
- Tests require no real OpenAI network or token.
- OpenSpec validation and `go test ./...` pass.
