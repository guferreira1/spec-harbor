# Proposal: Implement Local Context Retrieval

## Problem

SpecHarbor can discover classified local project context and can build a metadata-only repository context index, but users still cannot ask for relevant local context from indexed sources.

Without a scoped retrieval feature, future prompt or RAG work would be tempted to read arbitrary repository files, duplicate discovery/index traversal rules, dump large file contents, or mix local retrieval with embeddings, provider APIs, remote context, or agent execution.

## Goal

Add deterministic local context retrieval over supported sources represented by `.specharbor/context-index.json`.

Local retrieval means deterministic offline lookup over supported local context sources. It returns bounded, source-attributed snippets or metadata summaries that help a user inspect relevant repository context.

Local retrieval is not:

- RAG answer generation;
- embedding search;
- vector storage;
- semantic search provider integration;
- direct provider API usage;
- remote context collection;
- confirmed project context by itself;
- command execution;
- prompt execution;
- agent execution.

## Command Shape

Add this command:

```text
specharbor context retrieve --query "<query>"
```

The `context retrieve` subcommand fits the existing `specharbor context discover` and `specharbor context index` command group. The explicit `--query` flag makes the query boundary clear, prevents accidental positional arguments from being treated as retrieval input, and matches the safety posture of rejecting ambiguous or unsupported flags.

Supported:

```text
specharbor context retrieve --query "hexagonal architecture"
```

Rejected:

```text
specharbor context retrieve
specharbor context retrieve ""
specharbor context retrieve "hexagonal architecture"
specharbor context retrieve --query ""
specharbor context retrieve --query "context" extra
specharbor context retrieve --query "context" --json
specharbor context retrieve --query "context" --github
specharbor context retrieve --query "context" --remote
specharbor context retrieve --query "context" --rag
specharbor context retrieve --query "context" --embed
specharbor context retrieve --query "context" --provider openai
specharbor context retrieve --query "context" --execute
specharbor context retrieve --query "context" --agent codex
specharbor context retrieve --query "context" --deep
```

`--json` remains out of scope for this change. Adding machine-readable output later should not change the local retrieval domain contract.

## Relationship To The Repository Context Index

Retrieval requires an existing valid, current, non-truncated `.specharbor/context-index.json`.

The command must:

- read the stored index from `.specharbor/context-index.json`;
- validate schema version, generation marker, required fields, safe relative paths, and metadata-only structure;
- rebuild the current metadata index in memory only to check freshness;
- compare the stored index with current metadata using the existing index comparison rules;
- read snippets only from stored entries that are still current and marked `supported_for_retrieval`;
- write nothing.

Missing, invalid, stale, unsupported-schema, unreadable, or truncated indexes must fail safely with an actionable message asking the user to run:

```text
specharbor context index --write
```

The command must not silently create, update, or persist an index. This keeps retrieval deterministic, prevents hidden generated state, and lets users decide when local inventory should be refreshed.

If a source file is missing after the index was written, the freshness check must classify the index as stale and stop before snippet reads.

## Scope

- Add deterministic local/offline retrieval command support.
- Validate and normalize explicit query input.
- Load and freshness-check `.specharbor/context-index.json`.
- Read only supported indexed local sources through safe filesystem ports.
- Extract bounded snippets from text-like sources.
- Return metadata summaries only when a future index entry is valid but not marked `supported_for_retrieval`.
- Apply deterministic lexical scoring and tie-breaking.
- Print concise source-attributed retrieval reports.
- Enforce result, snippet, read, and output limits.
- Preserve existing `brief`, `brief --update`, `context discover`, `context index`, `prompt`, `scan`, final decision labels, and agent prompt behavior.
- Add focused tests and user-facing docs after implementation exists.
- Update this change's `tasks.md` only for completed work.

## Supported Local Sources

Retrieval must use only source entries represented by the repository context index and must align with the previous index feature:

- `AGENTS.md`
- `README.md`
- `CONTRIBUTING.md`
- bounded Markdown files under `docs/`
- `openspec/project.md`
- bounded Markdown files under `openspec/specs/`
- bounded Markdown files under `.specharbor/rules/`
- `.specharbor/project-brief.md`
- `package.json`
- `go.mod`
- `pom.xml`
- `build.gradle`
- `build.gradle.kts`
- `Cargo.toml`
- `pyproject.toml`
- `requirements.txt`
- `Dockerfile`
- `docker-compose.yml`
- `docker-compose.yaml`
- `Makefile`
- `Taskfile.yml`
- `Taskfile.yaml`
- bounded YAML files under `.github/workflows/`

Package, dependency, build, container, task-runner, and workflow manifests are retrievable as bounded snippets when the index entry is marked `supported_for_retrieval`. They are not parsed for command execution and are not treated as confirmed context.

## Query Behavior

The retrieval query must be explicit through `--query`.

Rules:

- Trim surrounding whitespace.
- Reject an empty normalized query.
- Reject queries longer than 512 characters after trimming.
- Normalize deterministically by lowercasing and splitting terms on non-letter and non-digit boundaries.
- Keep at most 32 normalized query terms.
- Remove duplicate terms while preserving first occurrence order, and use sorted copies only where deterministic ordering is needed.
- Reject a query that has no usable terms after normalization.
- Never execute the query.
- Never send the query to a provider, local model, network API, remote source, shell, package manager, workflow tool, prompt runner, or agent.

## Snippet And Output Bounds

Retrieval may read source file contents only after index validation and freshness checks pass.

Initial limits:

- max source read bytes per file: 128 KiB;
- max total source read bytes: 1 MiB;
- max total results: 10;
- max snippets per file: 2;
- max snippet chars: 600;
- max context window: 2 lines before and 2 lines after the best matching line;
- max rendered output chars for result snippets and summaries: 8,000.

Snippet behavior:

- Return matching line windows, not whole-file dumps.
- Include 1-based line ranges when practical.
- Preserve source text only inside bounded snippets.
- Trim snippets to the configured character limit.
- Mark omitted leading or trailing content with concise ellipsis text.
- Never print secret-file contents.
- Never print generated/heavy-folder contents.
- Never read or print `.specharbor/context-index.json` as a retrieval source.

## Ranking And Tie-Breaking

Ranking must be deterministic and local.

Allowed scoring inputs:

- lexical token matches in snippet text;
- exact normalized phrase matches in snippet text;
- token and phrase matches in path and filename;
- source category priority;
- classification hints from the index;
- OpenSpec and project brief priority;
- heading-line matches when discovered inside a snippet window.

Initial scoring should use simple integer weights owned by domain code, for example:

- exact phrase match in snippet text;
- exact phrase match in path;
- unique query token matches in content;
- query token matches in path or filename;
- source category priority;
- user-confirmed, detected-fact, suggested-assumption, and inventory metadata hints.

Tie-breaking order must be:

1. higher score;
2. higher source category priority;
3. lower path lexicographically;
4. lower starting line number;
5. lower snippet text lexicographically.

Not allowed:

- embeddings;
- vector similarity;
- LLM reranking;
- provider reranking;
- remote search.

## Output Behavior

The CLI output should be concise and source-attributed.

Each result must include:

- rank;
- path;
- source category or evidence category;
- score;
- line range when a snippet is present;
- bounded snippet or metadata summary;
- classification hint when present.

The report must also include:

- normalized query;
- index path;
- index status;
- result count;
- a no-results message when no supported source matches.

The report must not expose skipped or sensitive file contents.

## Safety Rules

The retrieval flow must:

- not follow symlinks;
- reject unsafe paths;
- not traverse outside the project root;
- skip sensitive files;
- skip generated or heavy folders;
- read bounded bytes only;
- not store raw retrieval cache;
- not write retrieval results;
- not execute commands;
- not call network APIs;
- not call provider APIs;
- not call local model APIs;
- not call remote source APIs;
- not run prompts;
- not run agents;
- not automate source control.

Sensitive files and generated/heavy folders must reuse or align with the repository context index skip policy.

## Compatibility

The implementation must preserve:

- `specharbor brief`;
- `specharbor brief --update`;
- `specharbor context discover`;
- `specharbor context index`;
- context-aware `specharbor prompt`;
- `specharbor scan`;
- existing final decision labels;
- existing agent prompt behavior.

`context discover` must remain independent from the index. Existing prompt generation must not automatically inject retrieval results in this change.

## Future Compatibility

The local retrieval domain should expose stable query, snippet, result, score, attribution, and report models so a future RAG provider can consume retrieval results without changing the local retrieval contract.

This change must not implement RAG answer generation.

## Out Of Scope

- Embeddings.
- Vector databases.
- Semantic search provider integrations.
- RAG answer generation.
- LLM calls.
- Provider APIs.
- Local model APIs.
- GitHub remote context.
- GitLab remote context.
- Bitbucket remote context.
- Network APIs.
- Automatic command verification.
- Automatic project command execution.
- Prompt execution.
- Agent execution.
- Source-control automation.
- Release automation.
- npm changes.
- Homebrew changes.
- `install.sh` changes.
- GoReleaser changes.
- Publishing flows.
- Archiving this OpenSpec change before merge.

## Success Criteria

- `specharbor context retrieve --query "<query>"` returns deterministic local retrieval results from current indexed sources.
- Empty, missing, oversized, or unusable queries fail safely.
- Missing, stale, invalid, truncated, unreadable, or unsupported index files fail safely with actionable guidance.
- Source reads are bounded and happen only through safe filesystem ports and adapters.
- Results are source-attributed and include rank, score, category/evidence, classification hints, line ranges when practical, and bounded snippets or metadata summaries.
- Ranking and tie-breaking are deterministic.
- Sensitive files, generated folders, symlinks, unsafe paths, and paths outside the project root are not read.
- No raw full-file dumps, secret leaks, embeddings, vectors, RAG generation, provider calls, remote context, command execution, prompt execution, agent execution, or source-control automation are introduced.
- Existing supported commands and prompt labels remain compatible.
- OpenSpec validation and `go test ./...` pass.
