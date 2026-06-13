# Design: Implement Local Context Retrieval

## Overview

Local context retrieval adds a deterministic read-only use case that selects relevant bounded snippets from local sources already represented by `.specharbor/context-index.json`.

The command surface is:

```text
specharbor context retrieve --query "<query>"
```

Retrieval depends on the repository context index for source selection and freshness metadata. It does not duplicate context discovery traversal, does not make the index confirmed project context, and does not add RAG, embeddings, remote context, provider APIs, or command execution.

## Command Contract

Supported:

```text
specharbor context retrieve --query "architecture rules"
```

Rejected:

```text
specharbor context retrieve
specharbor context retrieve ""
specharbor context retrieve architecture
specharbor context retrieve --query
specharbor context retrieve --query ""
specharbor context retrieve --query "architecture" extra
specharbor context retrieve --query "architecture" --json
specharbor context retrieve --query "architecture" --github
specharbor context retrieve --query "architecture" --remote
specharbor context retrieve --query "architecture" --rag
specharbor context retrieve --query "architecture" --embed
specharbor context retrieve --query "architecture" --provider
specharbor context retrieve --query "architecture" --execute
specharbor context retrieve --query "architecture" --agent
specharbor context retrieve --query "architecture" --deep
```

The CLI owns parsing and report formatting only. It must not own query normalization, source eligibility, scoring, snippet limits, or safety policy.

## Index Dependency

The retrieval use case must require an existing `.specharbor/context-index.json`.

Execution order:

1. Validate and normalize the query.
2. Load the stored index.
3. Validate schema version, generation marker, paths, supported source categories, file types, classification hints, and metadata-only shape.
4. Rebuild the current metadata index in memory using the existing index use case or shared index builder boundary.
5. Compare stored and current indexes using the existing index comparison rules.
6. Fail if the stored index is missing, invalid, stale, unsupported, unreadable, or truncated.
7. Iterate only over current stored entries marked `supported_for_retrieval`.
8. Revalidate each entry path before reading.
9. Read bounded bytes through a retrieval filesystem port.
10. Extract candidate snippets and score them.
11. Return a bounded structured report.

No retrieval path writes `.specharbor/context-index.json` or any retrieval cache. Missing or stale index states should tell the user to run:

```text
specharbor context index --write
```

This design intentionally uses in-memory index rebuilding only for freshness checks, not as an implicit retrieval index.

## Domain Model

Add domain models under `internal/core/domain` for:

- `LocalContextRetrievalQuery`;
- normalized query terms and phrase;
- retrieval limits;
- retrieval snippet;
- retrieval result;
- retrieval score;
- retrieval report;
- retrieval index dependency status;
- output truncation state.

The domain owns:

- query validation and normalization;
- allowed query length and term limits;
- snippet and report limits;
- source category priority;
- classification hint priority;
- lexical scoring rules;
- deterministic tie-breaking;
- path validation alignment with repository context index paths;
- result ordering.

The domain must reuse existing `ContextSourceCategory`, repository context index entry metadata, classification hints, and safe path rules where practical.

## Use Case

Add a use case under `internal/core/usecase` that orchestrates retrieval.

Responsibilities:

- validate dependencies and project root;
- load and normalize the stored repository context index;
- build current metadata for freshness comparison;
- convert missing, invalid, stale, truncated, and unsupported index states into structured report/error states;
- select supported retrieval entries;
- enforce read limits before and during reads;
- ask the filesystem port for bounded local file contents;
- skip entries that become unsafe before read;
- extract candidate line windows;
- score and rank candidate snippets through domain rules;
- enforce result, snippet, and output bounds;
- return structured report data without CLI formatting.

The use case must not import adapters, CLI packages, `os`, network packages, provider SDKs, source-control SDKs, terminal IO, shell/process execution, or agent runner packages.

## Ports

Add or extend small ports under `internal/core/ports`.

The retrieval use case needs only:

- read stored index contents;
- check whether supported files still exist through the index freshness flow;
- read bounded file contents by safe relative path;
- optionally get safe file metadata if needed to enforce read limits.

Recommended interface shape:

```go
type LocalContextRetrievalFileSystem interface {
    ReadFileSafely(root string, relativePath string) (string, error)
    ReadFileBytes(root string, relativePath string, maxBytes int64) ([]byte, error)
}
```

The exact shape may reuse the repository context index filesystem port if that keeps interfaces small and avoids duplication. Filesystem reads must still go through port and adapter boundaries.

## Filesystem Adapter

Implement concrete retrieval reads in `internal/adapters/filesystem`.

Adapter responsibilities:

- reuse or align with the existing safe relative path and symlink checks;
- reject path traversal, absolute paths, Windows drive paths, null-byte paths, and root escapes;
- reject symlink files and symlink path components;
- reject directories and non-regular files;
- read no more than the configured byte limit;
- avoid following symlinks between safety checks and reads;
- return errors without leaking sensitive file contents.

The adapter must not implement query normalization, ranking, source selection, or output formatting.

## Source Eligibility

The source set is the stored index entries after freshness validation.

Eligible entries:

- safe path;
- recognized source category;
- supported file type;
- not `.specharbor/context-index.json`;
- marked `supported_for_retrieval`;
- not skipped by sensitive/generated path policy;
- file size within retrieval read limits;
- current according to index freshness checks.

If a future valid index entry is not marked `supported_for_retrieval`, the report may include only a metadata summary when it scores through path/category matches. The first implementation may simply skip non-retrievable entries if no such entries exist.

## Query Normalization

Use deterministic normalization:

- trim whitespace;
- reject empty query;
- reject more than 512 characters;
- lowercase;
- split on non-letter and non-digit boundaries;
- discard empty tokens;
- deduplicate terms while preserving first occurrence order;
- keep at most 32 terms;
- retain a normalized phrase from the normalized terms in query order;
- use sorted term copies only where deterministic ordering is needed for output or tie-independent scoring;
- retain a display query formed from the trimmed original query.

The query string is data only. It must not be evaluated, executed, interpolated into shell commands, sent to providers, sent to remote systems, or used to expand paths.

## Snippet Extraction

Read file contents as text for supported indexed sources. Invalid UTF-8 should be handled deterministically by replacement or by treating the file as unreadable for snippets; do not panic.

Snippet extraction:

- split contents into lines;
- find lines matching query terms or phrase;
- create a window of up to 2 lines before and 2 lines after each best matching line;
- compute 1-based line ranges;
- collapse duplicate or overlapping windows per file;
- retain at most 2 snippets per file before global ranking;
- trim each snippet to 600 characters;
- include concise ellipsis markers when content is omitted;
- never switch into whole-file dump mode.

When no content line matches but path/category metadata matches strongly enough, the result may include a metadata-only summary instead of a snippet.

## Scoring

Scoring must be implemented in domain code with deterministic integer weights.

Initial scoring inputs:

- exact normalized phrase match in snippet text;
- exact normalized phrase match in path;
- unique normalized query term matches in snippet text;
- normalized query term matches in path;
- normalized query term matches in filename;
- source category priority;
- classification hint priority;
- heading-line match inside a snippet window.

Suggested category priority from highest to lowest:

1. project brief;
2. OpenSpec project;
3. OpenSpec specs;
4. SpecHarbor rules and AGENTS;
5. README and CONTRIBUTING;
6. docs;
7. package, dependency, and build manifests;
8. workflow, task runner, and container config.

Tie-breaking order:

1. higher score;
2. higher source category priority;
3. lower path lexicographically;
4. lower starting line number;
5. lower snippet text lexicographically.

Forbidden scoring inputs:

- embeddings;
- vector similarity;
- LLM reranking;
- provider reranking;
- remote search;
- command output.

## Report Model

The structured report should include:

- original display query;
- normalized query terms;
- index path;
- index status;
- result count;
- results;
- truncation flags;
- warnings or actionable messages for missing/stale/invalid/truncated indexes.

Each result should include:

- rank;
- path;
- source category;
- source evidence category;
- score;
- classification hints;
- line range when a snippet exists;
- bounded snippet text or metadata summary.

The CLI should render this report concisely and should exit non-zero for missing, stale, invalid, unreadable, unsupported, or truncated indexes and for query validation failures.

## Limits

Use these initial defaults:

- max query chars: 512;
- max query terms: 32;
- max source read bytes per file: 128 KiB;
- max total source read bytes: 1 MiB;
- max total results: 10;
- max snippets per file: 2;
- max snippet chars: 600;
- max context window lines before and after match: 2;
- max rendered snippet and summary chars: 8,000.

Domain defaults should be easy to test through constructor injection or equivalent use-case configuration.

## Index State Behavior

- Missing index: fail with status `missing_index` and tell the user to run `specharbor context index --write`.
- Invalid JSON: fail with status `invalid_index`.
- Unsupported schema: fail with status `invalid_index`.
- Invalid metadata shape: fail with status `invalid_index`.
- Unreadable index: fail with status `invalid_index`.
- Stale index: fail with status `stale_index` and include bounded stale reason summaries.
- Truncated index: fail with status `truncated_index`; retrieval requires a complete index so results are not silently incomplete.
- Source file missing after index write: freshness comparison marks the index stale and retrieval stops before snippets.

The command may print index status details, but must not print stored raw file content because the index stores none.

## Architecture

Preserve the dependency rule:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

Domain responsibilities:

- query model;
- retrieval result and snippet models;
- scoring and ranking;
- limits;
- safety policy references;
- deterministic tie-breaking.

Use-case responsibilities:

- index loading and freshness checking;
- orchestration of safe local reads;
- snippet extraction;
- ranking and report assembly;
- output bound enforcement.

Port responsibilities:

- small filesystem read contracts consumed by the use case.

Adapter responsibilities:

- concrete safe filesystem reads and stored index reads.

CLI responsibilities:

- parse `context retrieve --query`;
- reject unsupported flags and positional arguments;
- wire concrete adapters and use cases;
- format the structured report;
- return non-zero exit codes for validation and index dependency failures.

Forbidden responsibilities:

- CLI must not own retrieval rules, ranking weights, snippet bounds, source eligibility, or path safety policy.
- Core must not import adapters, CLI, provider SDKs, remote SDKs, source-control SDKs, workflow SDKs, agent runners, terminal IO, or process execution packages.

## Documentation Plan

Update public docs after implementation exists:

- `README.md`: add `specharbor context retrieve --query "<query>"` to implemented command examples and summarize local/offline retrieval.
- `docs/usage.md`: document command syntax, index dependency, query rules, source set, snippet/result limits, report shape, safety boundaries, and non-goals.
- `docs/workflow.md`: mention local retrieval as optional inspection over the local context index and not a replacement for confirmed briefing.
- `docs/agent-roles.md` only if needed to clarify that prompt generation does not automatically inject retrieval results in this change.

Documentation must not describe embeddings, vector databases, RAG answer generation, remote context, provider APIs, command execution, prompt execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, or publishing behavior as part of this feature.

## Testing Strategy

Add focused tests for:

- query validation;
- query normalization;
- missing query rejection;
- empty query rejection;
- bounded query length;
- max query terms;
- missing index behavior;
- stale index behavior;
- invalid index behavior;
- truncated index behavior;
- unsupported schema behavior;
- source file missing after index write;
- supported local source retrieval;
- package and build manifest snippets;
- unsupported source skipping;
- sensitive file skipping;
- symlink traversal blocking;
- unsafe path rejection;
- bounded source reads;
- max total source read bytes;
- snippet extraction limits;
- max snippets per file;
- max total results;
- max output size;
- no raw full-file dumps;
- no secret content leaks;
- deterministic scoring;
- deterministic tie-breaking;
- source attribution;
- line range metadata;
- metadata-only result behavior if implemented;
- no embeddings, vectors, RAG, provider calls, remote calls, project command execution, prompt execution, or agent execution;
- CLI report behavior;
- unsupported flag rejection;
- unsupported positional argument rejection;
- existing `specharbor brief` regression;
- existing `specharbor brief --update` regression;
- existing `specharbor context discover` regression;
- existing `specharbor context index` regression;
- existing context-aware `specharbor prompt` regression;
- existing `specharbor scan` regression;
- final decision label regression;
- OpenSpec validation.

Implementation verification should run:

```text
gofmt
go test ./...
go test -count=1 ./...
go run ./cmd/specharbor validate implement-local-context-retrieval
go run ./cmd/specharbor
go run ./cmd/specharbor version
go run ./cmd/specharbor context discover
go run ./cmd/specharbor context index
go run ./cmd/specharbor prompt implement-local-context-retrieval --role spec-author
go run ./cmd/specharbor prompt implement-local-context-retrieval --role implementer
go run ./cmd/specharbor scan
go run ./cmd/specharbor context index --write
go run ./cmd/specharbor context retrieve --query "architecture"
```

Generated `.specharbor/context-index.json` files created during manual verification should be removed before commit because the index is generated local state.
