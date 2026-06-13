# Design: Implement Repository Context Index

## Overview

Repository context indexing adds a deterministic local inventory use case for supported context sources. It builds on the context discovery foundation but does not change discovery semantics and does not implement retrieval.

The command surface should be:

```text
specharbor context index
specharbor context index --write
specharbor context index --check
```

`context discover` continues to classify project context signals. `context index` records bounded file metadata that future local retrieval can consume later. The index is inventory metadata, not confirmed context and not ranked retrieval output.

## Command Contract

Supported:

```text
specharbor context index
specharbor context index --write
specharbor context index --check
```

Rejected:

```text
specharbor context index --write --check
specharbor context index --json
specharbor context index --path <path>
specharbor context index --deep
specharbor context index --github
specharbor context index --rag
specharbor context index extra
specharbor context update
```

Behavior:

- no flag: build current metadata in memory and print a concise report;
- `--write`: build current metadata and safely persist it at `.specharbor/context-index.json`;
- `--check`: load stored metadata, rebuild current metadata in memory, compare stable fields, and report freshness.

`--check` should exit with code `0` only when the stored index is current. Missing, invalid, unreadable, unsupported schema, or stale index states should print a clear report and exit non-zero. It must not write a new index.

## Index Path And Commit Policy

The index path is:

```text
.specharbor/context-index.json
```

The file is generated project-local state. It should be ignored by source control in this repository and should not be staged or committed by this feature. The implementation may add a focused `.gitignore` entry for this generated file.

The index writer must create `.specharbor/` when needed through the filesystem port and adapter. It must write atomically or with an equivalent safe replacement strategy so partial index files are not left behind on failure.

The generated index file must not be included as an index entry.

## Index Schema

Use deterministic JSON with stable field names and stable entry ordering. The first schema version is `1`.

Recommended shape:

```json
{
  "schema_version": 1,
  "generated": {
    "mode": "deterministic",
    "tool": "specharbor context index"
  },
  "project": {
    "root_marker": "openspec/project.md"
  },
  "limits": {
    "max_indexed_files": 500,
    "max_file_size_bytes": 262144,
    "max_total_file_bytes": 5242880,
    "max_skipped_records": 200
  },
  "entries": [],
  "skipped": [],
  "truncated": false
}
```

Avoid a generated timestamp in the first version so repeated writes are deterministic when repository files have not changed. Use the deterministic generation marker above instead.

## Entry Model

Each entry should include metadata only:

- `path`: safe project-relative path with `/` separators;
- `source_category`: existing or aligned context source category;
- `file_type`: markdown, json, yaml, xml, gradle, toml, text, dockerfile, makefile, go_module, or another small explicit type;
- `language_or_ecosystem`: Markdown, Go, Node.js, JVM, Rust, Python, Docker, GitHub Actions, OpenSpec, SpecHarbor, or empty when unknown;
- `size_bytes`: regular file size;
- `content_hash`: `sha256:<hex>`;
- `modified_time`: stable UTC RFC3339Nano timestamp or another documented filesystem freshness marker;
- `supported_for_retrieval`: `true` for text-like supported sources intended to be readable by a future retrieval feature, `false` for records retained only as inventory if any;
- `classification_hints`: zero or more of `user_confirmed_context`, `detected_fact`, `suggested_assumption`, or `inventory_metadata`;
- `source_evidence_category`: concise evidence category such as `readme`, `openspec_spec`, `package_manifest`, or `workflow_config`.

The index must not include raw file contents, text snippets, parsed secrets, command output, prompt text, embeddings, vectors, provider responses, remote API data, or absolute local paths.

## Skip And Limit Records

Skip records should be concise:

- `path`: safe project-relative path;
- `reason`: stable reason code such as `sensitive_file`, `generated_directory`, `symlink`, `unsupported_source`, `file_too_large`, `max_files_reached`, `max_total_bytes_reached`, or `unsafe_path`.

Skip records must not include file contents or secret values.

Use these initial limits:

- max indexed files: `500`;
- max individual file size for metadata hashing: `256 KiB`;
- max total indexed file bytes: `5 MiB`;
- max skip records in the persisted report: `200`;
- bounded directory depth for `docs/`, `openspec/specs/`, `.specharbor/rules/`, and `.github/workflows/`: `4`;
- max files per supported directory family before global limits: governed by the global max indexed files.

When limits are exceeded, produce stable skip records and set `truncated` to `true`.

## Supported Source Selection

Source selection should be explicit and bounded, not a repository-wide walk.

Fixed files:

- `AGENTS.md`
- `README.md`
- `CONTRIBUTING.md`
- `.specharbor/project-brief.md`
- `openspec/project.md`
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

Bounded directories:

- `docs/` with Markdown files;
- `openspec/specs/` with Markdown files;
- `.specharbor/rules/` with Markdown files;
- `.github/workflows/` with `.yml` and `.yaml` files.

The implementation should align source categories with `ContextSourceCategory` where practical. It may add index-specific categories only when the discovery categories are insufficient.

## Freshness And Staleness

Freshness is based on stored metadata, not command execution.

`--check` should:

1. read `.specharbor/context-index.json`;
2. validate schema version and required fields;
3. rebuild the current index in memory using the same limits and source selection;
4. compare stable serialized index content or compare stable entry, skip, limit, and truncation fields;
5. report current or stale.

Stale conditions include:

- stored index missing;
- unreadable or invalid JSON;
- unsupported schema version;
- changed entry set;
- changed file size;
- changed content hash;
- changed modified freshness marker;
- changed skip/truncation state;
- changed index limits or generation marker.

`--check` must not run tests, builds, package managers, scripts, shells, agents, prompt execution, provider APIs, local model APIs, source-control commands, GitHub APIs, GitLab APIs, CI, workflow APIs, or network calls.

## Report Behavior

Reports should be concise and deterministic. Suggested output information:

- index path;
- mode: report, write, or check;
- status: built, written, current, stale, missing, invalid, or truncated;
- schema version;
- indexed file count;
- skipped record count;
- total indexed bytes;
- selected limits;
- stale reason count when checking;
- first few stale or skip reason summaries, bounded and without raw contents.

Do not dump raw file contents.

## Architecture

Preserve the dependency rule:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

Domain responsibilities:

- index record models;
- source category and evidence models;
- file type and language/ecosystem hints;
- skip policy catalog aligned with context discovery;
- freshness/staleness models;
- deterministic ordering;
- metadata-only validation;
- index comparison rules.

Use-case responsibilities:

- validate dependencies and project root;
- orchestrate source inventory;
- apply limits and skip rules;
- build the index;
- read stored indexes for checks;
- compare current and stored metadata;
- produce structured report data;
- call persistence only for `--write`;
- avoid CLI formatting and concrete filesystem dependencies.

Port responsibilities:

- expose bounded filesystem operations for metadata, listing, hashing/reading bytes, safe writes, and stored index reads;
- keep interfaces small and owned by the index use cases.

Adapter responsibilities:

- implement filesystem traversal, `lstat` metadata, safe relative path handling, symlink rejection, hashing, directory creation, index JSON reads, and safe writes;
- reuse existing local filesystem safety helpers where practical.

CLI responsibilities:

- parse `context index` flags and reject unsupported inputs;
- wire concrete adapters and use cases;
- format reports only.

Forbidden responsibilities:

- CLI must not own source selection, skip rules, staleness rules, file classification, or index comparison logic.
- Core must not import adapters, CLI, `os`, terminal IO, network APIs, source-control SDKs, provider SDKs, workflow SDKs, agent runners, or process execution packages.

## Compatibility

The implementation must preserve:

- `specharbor brief`;
- `specharbor brief --update`;
- `specharbor context discover`;
- context-aware `specharbor prompt`;
- `specharbor scan`;
- existing final decision labels in role prompts;
- existing agent prompt behavior outside any documented Project Context already implemented.

`context discover` should continue to be read-only and should not depend on the index.

## Future Compatibility

Future local retrieval should be able to consume the index through domain/use-case boundaries. The schema should carry enough source category, evidence, hash, and support metadata to know which files may be considered later.

This change must not add retrieval queries, snippet selection, ranking scores, embeddings, vectors, semantic databases, RAG provider configuration, remote context, or prompt-injection behavior.

## Documentation Plan

Update public docs after implementation exists:

- `README.md`: add `specharbor context index`, `--write`, and `--check` to implemented command examples and describe the index as metadata-only.
- `docs/usage.md`: document command syntax, index path, generated-file policy, schema summary, supported sources, skip rules, limits, freshness checks, and safety boundaries.
- `docs/workflow.md`: mention the index as optional local inventory metadata for future context work, not a replacement for confirmed briefing or discovery.
- `docs/agent-roles.md` only if needed to clarify that prompt generation still consumes classified context, not raw index contents.

Documentation must not claim retrieval, RAG, embeddings, vector stores, remote context, provider APIs, command execution, prompt execution, agent execution, source-control automation, release automation, or publishing behavior.

## Testing Strategy

Add focused tests for:

- index model validation;
- metadata-only index entries;
- deterministic index generation;
- stable entry and skip ordering;
- supported source inventory;
- source category and file type inference;
- language/ecosystem hints;
- sensitive file skipping;
- heavy/generated folder skipping;
- symlink traversal blocking;
- path traversal, absolute path, Windows drive path, and null-byte rejection;
- max file count behavior;
- max file size behavior;
- max total bytes behavior;
- hash correctness;
- size and modified-time metadata correctness;
- stale index detection for changed hash, size, entry set, schema, limits, and missing file;
- safe write behavior preserving the original index on failure;
- generated index file not being indexed;
- no raw file content stored;
- empty repository behavior;
- ambiguous repository behavior;
- CLI report behavior for report, write, check-current, check-stale, missing, and invalid states;
- existing `specharbor brief` regression;
- existing `specharbor brief --update` regression where practical without requiring interactive manual input;
- existing `specharbor context discover` regression;
- existing context-aware prompt regression;
- existing `specharbor scan` regression;
- final decision label regression;
- OpenSpec validation.

Implementation verification should run:

```text
gofmt
go test ./...
go test -count=1 ./...
go run ./cmd/specharbor validate implement-repository-context-index
go run ./cmd/specharbor
go run ./cmd/specharbor version
go run ./cmd/specharbor context discover
go run ./cmd/specharbor prompt implement-repository-context-index --role spec-author
go run ./cmd/specharbor prompt implement-repository-context-index --role implementer
go run ./cmd/specharbor scan
go run ./cmd/specharbor context index
go run ./cmd/specharbor context index --write
go run ./cmd/specharbor context index --check
```

Generated `.specharbor/context-index.json` files created during manual verification should be removed before commit because the index is generated local state.
