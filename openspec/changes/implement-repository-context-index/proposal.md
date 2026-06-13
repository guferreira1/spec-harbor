# Proposal: Implement Repository Context Index

## Problem

SpecHarbor can discover bounded local project context with `specharbor context discover`, and it can render that classified context into agent prompts. It still lacks a durable local inventory of supported context sources that future local retrieval work can consume safely.

Without an index, future retrieval work would need to re-walk supported sources or invent a storage model at the same time it implements selection and ranking. That would increase the risk of over-reading the repository, storing raw file contents, treating repository hints as confirmed context, or mixing retrieval, RAG, embeddings, and provider behavior into a foundational metadata change.

## Goal

Add a safe repository context index that records bounded inventory metadata for supported local context sources.

The index is:

- local and offline;
- deterministic enough for tests and review;
- metadata-only;
- bounded by file count, file size, and total indexed bytes;
- safe around sensitive files, generated folders, symlinks, path traversal, and absolute paths;
- useful for future local retrieval work through stable domain/use-case boundaries.

The index is not:

- retrieval;
- snippet ranking;
- RAG;
- a semantic database;
- a vector store;
- confirmed project context;
- a provider API integration;
- a command execution or command verification mechanism.

## Command Shape

Add a `context index` subcommand:

```text
specharbor context index
specharbor context index --write
specharbor context index --check
```

`specharbor context index` builds the current index in memory and prints a concise report without writing files. This mirrors `context discover` as a safe local inspection command.

`specharbor context index --write` builds the current index and safely writes it to the project-local index path. The explicit flag is required because the existing `context discover` command is read-only and users should opt into generated local state.

`specharbor context index --check` reads the stored index, rebuilds current metadata in memory, and reports whether the stored index is current or stale. The check must not execute project commands. It should exit non-zero when the stored index is missing, unreadable, invalid, or stale.

The flags are mutually exclusive. Unsupported flags and positional arguments are rejected.

## Index File Location

Persist the index at:

```text
.specharbor/context-index.json
```

This path keeps SpecHarbor-owned local metadata under the existing `.specharbor/` project directory while avoiding OpenSpec change packages and production source directories. The file is generated state and is not intended to be committed. The implementation should ignore this generated file, for example with a `.gitignore` entry, unless a future change explicitly introduces committed fixtures.

The generated index file itself must not be indexed.

## Scope

- Add the `specharbor context index` command family described above.
- Define repository context index domain models for schema version, generation marker, limits, entries, skip records, source categories, file type, language/ecosystem hints, freshness metadata, and report status.
- Build deterministic index entries for supported local context sources.
- Persist the index as stable JSON at `.specharbor/context-index.json` only when `--write` is provided.
- Read and compare the stored index when `--check` is provided.
- Record metadata only:
  - relative path;
  - source category;
  - file type;
  - language or ecosystem hint when safely detectable from path or manifest type;
  - file size;
  - SHA-256 content hash;
  - modified timestamp or equivalent filesystem freshness marker;
  - whether the source is supported for future retrieval;
  - classification hints;
  - source evidence category.
- Record concise skip metadata without raw file contents.
- Reuse or align with existing context discovery source categories and skip policies.
- Keep index generation deterministic with stable ordering and documented limits.
- Add focused tests for domain models, use cases, adapters, CLI behavior, safety rules, staleness, safe writes, and regressions.
- Update user-facing docs after implementation exists.
- Update this change's `tasks.md` only for completed implementation work.

## Supported Local Sources

The index should inventory these supported local sources when present:

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

The variant filenames are included because existing context discovery already recognizes them or their adjacent command families. The implementation should not add broad source-code indexing.

## Safety Rules

The index must skip obvious sensitive files:

- `.env`
- `.env.*`
- `*.pem`
- `*.key`
- `id_rsa`
- `id_ed25519`
- `secrets.*`
- `credentials.*`

The index must skip heavy or generated folders:

- `.git/`
- `node_modules/`
- `dist/`
- `build/`
- `target/`
- `vendor/`
- `coverage/`
- `.tmp/`
- `.cache/`
- `.next/`
- `.nuxt/`
- `out/`
- `bin/`
- `obj/`

The index must not follow symlinks. Absolute paths, path traversal, Windows drive paths, null-byte paths, and paths that escape the project root must be rejected before reading, hashing, or writing.

## Out Of Scope

- Local retrieval.
- Snippet ranking.
- Embeddings.
- Vector databases.
- RAG.
- Semantic indexing.
- GitHub remote context.
- GitLab remote context.
- Bitbucket remote context.
- Provider APIs.
- Local model APIs.
- Prompt execution.
- Agent execution.
- Automatic command verification.
- Automatic project command execution.
- Dependency graph analysis.
- Source-control automation.
- Release automation.
- npm changes.
- Homebrew changes.
- `install.sh` changes.
- GoReleaser changes.
- Publishing flows.
- Archiving this OpenSpec change before merge.

## Success Criteria

- `specharbor context index` builds a deterministic metadata-only report without writing.
- `specharbor context index --write` safely writes `.specharbor/context-index.json`.
- `specharbor context index --check` detects current, missing, invalid, and stale index states without executing commands.
- Index entries include stable relative paths, source categories, file type, language/ecosystem hints, size, hash, freshness metadata, future retrieval support, classification hints, and evidence category.
- The index stores no raw file contents, secrets, snippets, command output, provider output, or remote data.
- Sensitive files, generated folders, symlinks, unsafe paths, oversized files, excessive file counts, and excessive total bytes are handled safely and reported concisely.
- Existing `specharbor brief`, `specharbor brief --update`, `specharbor context discover`, `specharbor prompt`, `specharbor scan`, final decision labels, and agent prompt behavior remain compatible.
- Architecture boundaries are preserved.
- OpenSpec validation and `go test ./...` pass.
