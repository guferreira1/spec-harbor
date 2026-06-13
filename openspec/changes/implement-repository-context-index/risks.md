# Risks: Implement Repository Context Index

## Index Mistaken For Confirmed Context

Users or future code could treat inventory metadata as user-confirmed project facts.

Mitigation: name and document the index as metadata-only inventory. Keep confirmed context, detected facts, suggested assumptions, and inventory metadata distinct. Do not change project brief or prompt precedence rules in this feature.

## Scope Creep Into Retrieval

An index can invite retrieval, ranking, snippets, embeddings, vector databases, or RAG behavior before the safety model is ready.

Mitigation: restrict this feature to bounded inventory metadata. Do not add query APIs, ranking scores, snippets, embeddings, vectors, semantic stores, provider calls, or prompt injection.

## Sensitive Data Exposure

Hashing or reporting files from supported directories could accidentally include secrets or reveal sensitive filenames.

Mitigation: reuse or align with the existing skip policy for `.env`, keys, credential files, secret files, generated folders, and symlinks. Store only metadata and concise skip reason codes. Never store raw file contents or snippets.

## Over-Indexing The Repository

A broad walk could index source code, generated assets, dependency folders, build outputs, or large files.

Mitigation: use an explicit supported-source catalog, bounded directory families, max file count, max individual file size, max total bytes, max skip records, and directory depth limits. Report truncation deterministically.

## Nondeterministic Output

Generated timestamps, filesystem traversal order, JSON map ordering, or local absolute paths could make the index noisy and hard to test.

Mitigation: avoid generated timestamps in the first schema, sort entries and skip records deterministically, use stable JSON structs, use relative paths only, and include a deterministic generation marker.

## Staleness False Confidence

Users could assume a stored index is current when files changed after the last write.

Mitigation: implement `--check` to rebuild current metadata and compare entry sets, hashes, sizes, modified markers, limits, and skip/truncation state. Make stale checks explicit and avoid executing commands.

## Partial Writes

Writing `.specharbor/context-index.json` could leave a corrupt or empty file if a write fails.

Mitigation: use safe replacement through the filesystem adapter, preferably temp-file-and-rename, and add tests proving the original index survives write failure.

## CLI Business Logic Drift

The CLI could accumulate source selection, skip policy, file classification, or staleness comparison rules.

Mitigation: keep those rules in domain/usecase code. The CLI should parse arguments, wire dependencies, and format structured reports only.

## Duplicate Context Discovery Logic

The index source catalog overlaps with `context discover`, creating a risk of diverging source categories and skip behavior.

Mitigation: align with existing `ContextSourceCategory` and skip policy where practical. Add index-specific logic only for metadata concerns such as file type, hash, and freshness.

## Generated File Committed Accidentally

The local index is useful for development but may create noisy commits and stale metadata in the repository.

Mitigation: treat `.specharbor/context-index.json` as generated local state and ignore it. Do not stage generated index files unless a future change intentionally adds fixtures.

## Documentation Overclaiming

Docs could describe the index as retrieval, RAG, semantic search, command verification, remote discovery, or provider integration.

Mitigation: document command shape, index path, schema summary, limits, and safety boundaries only. Explicitly state that retrieval, RAG, embeddings, vector stores, remote context, provider APIs, command execution, prompt execution, agent execution, and source-control automation remain out of scope.
