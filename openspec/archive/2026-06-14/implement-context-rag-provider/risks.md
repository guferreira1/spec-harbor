# Risks: Implement Context RAG Provider

## Risks

### Local Commands Becoming Networked

Existing local/offline commands could accidentally read provider credentials or call providers.

### Provider Secrets Leaking

`SPECHARBOR_OPENAI_API_KEY` could leak through errors, reports, request dumps, test failures, logs, or source context.

### Generated Answers Mistaken For Confirmed Context

Provider output can sound authoritative even when it is only generated from bounded snippets.

### Scope Creep Into Magic AI Behavior

The feature could drift into automatic prompt injection, automatic file writes, code changes, agent execution, source-control automation, or command execution.

### Overbroad Context Exposure

RAG context construction could accidentally include full files, sensitive files, generated files, unbounded remote content, raw index files, or raw provider responses.

### Provider Dependency Fragility

OpenAI API errors, rate limits, timeouts, invalid credentials, model changes, malformed responses, or oversized responses could make the command unreliable or unsafe.

### Architecture Boundary Drift

Provider HTTP details or credential handling could leak into core domain/usecase code, or CLI code could own prompt construction and source selection rules.

### GitHub Remote Coupling

Optional GitHub RAG sources could make local RAG depend on network availability or GitHub credentials.

### Hidden Persistence

Provider requests, provider responses, answers, remote context, embeddings, or vectors could be cached or written without explicit user approval.

### Documentation Overclaiming

Docs could imply that RAG verifies truth, updates project context, injects prompts, provides semantic indexing, or performs source-control automation.

## Mitigations

- Keep `context rag` as the only provider command path.
- Add regression tests proving `context discover`, `context index`, `context retrieve`, `brief`, `prompt`, `validate`, `review`, and `scan` do not construct or call providers.
- Read `SPECHARBOR_OPENAI_API_KEY` only after parsing confirms `context rag --provider openai`.
- Return a safe missing-credential error before provider construction when the token is absent.
- Never print, persist, log, or include the token in structured reports or errors.
- Build provider requests from bounded source snippets only.
- Enforce max sources, max snippet chars, max total context chars, max answer chars, max provider response bytes, provider timeout, and max rendered output.
- Include source attribution and local/remote markers in every RAG report.
- Label provider output as generated answer text, not confirmed project context.
- Do not write answers to `.specharbor/project-brief.md`, OpenSpec files, prompts, caches, indexes, or project files.
- Do not automatically inject RAG output into role prompts.
- Do not add embeddings, vector stores, persistent RAG indexes, provider reranking, file uploads, OpenAI hosted tools, conversations, shell execution, agent execution, or source-control automation.
- Keep provider ports small and generation-only.
- Keep OpenAI HTTP behavior in `internal/adapters/openai`.
- Add architecture tests for no `net/http`, provider SDK, adapter, CLI, or `os/exec` imports in core.
- Keep GitHub source use behind explicit `--from github --repo owner/name`; local remains the default.
- Use fake providers and fake HTTP transports for tests.
- Document missing-token behavior and live provider smoke as optional.
- Keep docs explicit that RAG is optional generated output over bounded sources, not confirmed project context, prompt injection, semantic indexing, release automation, or source-control automation.
