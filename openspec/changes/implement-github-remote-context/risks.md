# Risks: Implement GitHub Remote Context

## Local Commands Becoming Networked

Risk: existing local/offline commands could accidentally call GitHub or require tokens.

Mitigation: add a separate `context github` path only. Keep `discover`, `index`, `retrieve`, `brief`, and `prompt` wired to local adapters. Add regression and architecture tests.

## GitHub Mutation Creep

Risk: a GitHub adapter could grow write methods for PRs, issues, comments, labels, refs, releases, workflow dispatch, or source-control automation.

Mitigation: define a read-only port with only repository metadata, ref resolution, directory listing, and file read methods. Add tests scanning ports/adapters for mutation-oriented methods and forbidden APIs.

## Token Leakage

Risk: `SPECHARBOR_GITHUB_TOKEN` could appear in errors, logs, reports, raw URLs, or test output.

Mitigation: keep token handling inside the adapter/CLI boundary, never include it in structured errors, redact authorization data, and add token-redaction tests.

## Overfetching Remote Repositories

Risk: remote traversal could fetch too much data, become slow, or expose unnecessary private content.

Mitigation: use an approved source set, path filters that narrow only, file count limits, size limits, total byte limits, directory depth limits, and tree-entry limits. Do not clone or fetch arbitrary source code.

## Sensitive Content Exposure

Risk: snippets could expose secrets from remote repositories.

Mitigation: reuse or align sensitive/generated skip policies, skip unsupported/binary files, bound snippets, avoid whole-file dumps, and test `.env`, key, credential, `.npmrc`, `.pypirc`, and `.netrc` skip cases.

## Remote Evidence Mistaken For Confirmed Context

Risk: users or future prompt code could treat remote snippets as confirmed project truth.

Mitigation: mark output as remote, include source attribution, do not write project briefs or local indexes, and do not inject remote results into prompts in this change.

## Nondeterministic Ranking

Risk: GitHub listing order, API pagination, or map iteration could produce unstable results.

Mitigation: sort candidates and results deterministically. Use local lexical scoring and explicit tie-breaking. Do not rely on GitHub search ranking.

## Rate Limits And Offline Failures

Risk: public unauthenticated use can hit GitHub rate limits or fail offline.

Mitigation: keep command explicit, support optional `SPECHARBOR_GITHUB_TOKEN`, map rate-limit/offline failures to safe messages, and keep tests free of live network dependencies.

## Scope Creep Into RAG Or Providers

Risk: remote context could invite embeddings, vector stores, generated answers, provider reranking, or LLM APIs.

Mitigation: keep this feature to bounded retrieval and local lexical ranking. Reject `--rag`, `--embed`, and `--provider`. Add architecture tests for forbidden terms/imports.

## CLI Business Logic Drift

Risk: CLI code could accumulate validation, source selection, scoring, or truncation rules.

Mitigation: keep those rules in domain/usecase. CLI parses, wires, formats, and maps exit codes only.

## Documentation Overclaiming

Risk: docs may imply broad GitHub integration, automatic prompt context, or source-control automation.

Mitigation: document the command as explicit, read-only, bounded remote retrieval. State non-goals clearly and keep release/npm/Homebrew/install/GoReleaser behavior out of scope.
