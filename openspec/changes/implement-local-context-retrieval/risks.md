# Risks: Implement Local Context Retrieval

## Retrieval Mistaken For Confirmed Context

Users or future code could treat retrieved snippets as user-confirmed project truth.

Mitigation: label results with source attribution, source category, and classification hints. Preserve the distinction between user-confirmed context, detected facts, suggested assumptions, and inventory metadata. Do not change prompt precedence or project brief behavior in this feature.

## Scope Creep Into RAG Or Embeddings

Retrieval can invite RAG answer generation, embedding stores, vector search, semantic providers, LLM reranking, or provider APIs.

Mitigation: keep this feature deterministic, lexical, local, and offline. Explicitly reject `--rag`, `--embed`, `--provider`, `--remote`, and related flags. Do not add provider ports, vector indexes, embeddings, semantic reranking, or generated answers.

## Sensitive Data Exposure

Reading snippets from local files could expose secrets or credentials if safety rules are incomplete.

Mitigation: read only entries from the current repository context index, reuse or align sensitive-file and generated-directory skip policies, revalidate paths before reads, reject symlinks, enforce bounded snippets, and add tests for no secret content leaks.

## Raw Full-File Dumps

Snippet retrieval could accidentally print entire files, especially small files or files with dense matches.

Mitigation: extract line windows around matches, enforce max snippets per file, max snippet chars, max results, and max output size. Never implement a whole-file rendering mode.

## Stale Index False Confidence

Retrieval results could be based on an old index after files have changed or disappeared.

Mitigation: require a valid current index. Rebuild current metadata in memory and compare against the stored index before reading snippets. Treat missing files, changed hashes, changed sizes, changed modified markers, changed entry sets, changed skip state, changed limits, invalid JSON, unsupported schema, and truncation as dependency failures.

## Hidden Writes Or Generated State

Users could be surprised if retrieval silently writes or refreshes `.specharbor/context-index.json`.

Mitigation: retrieval writes nothing. Missing or stale index states instruct users to run `specharbor context index --write` explicitly.

## CLI Business Logic Drift

The CLI could accumulate query normalization, scoring, snippet, source selection, or safety rules.

Mitigation: keep those rules in domain and usecase layers. CLI parses arguments, wires dependencies, formats reports, and maps failures to exit codes only.

## Duplicate Index Or Discovery Logic

Retrieval could duplicate source traversal or context discovery rules and drift from the index feature.

Mitigation: consume repository context index entries and existing index comparison behavior. Do not add repository-wide source discovery. Reuse safe path and skip policy helpers where practical.

## Nondeterministic Ranking

Filesystem order, map iteration, or unstable scoring could make results noisy and hard to test.

Mitigation: sort query terms, entries, snippets, and ties deterministically. Use integer scoring and explicit tie-breaking by score, category priority, path, line start, and snippet text.

## Output Too Large

Many matching files could produce excessive terminal output.

Mitigation: enforce max total results, max snippets per file, max snippet chars, max source read bytes, max total source read bytes, and max rendered snippet/summary characters.

## Future RAG Contract Instability

If retrieval results are shaped only for the CLI, a future RAG provider might require breaking changes.

Mitigation: define stable domain report, result, snippet, score, attribution, classification hint, and line range models now. Future RAG may consume these results, but this change does not implement RAG generation.

## Documentation Overclaiming

Docs could make retrieval sound like semantic search, RAG, provider-backed search, remote context, command verification, or confirmed context.

Mitigation: document retrieval as deterministic local/offline lookup over the repository context index with bounded local source reads. State non-goals clearly in README, usage, and workflow docs.
