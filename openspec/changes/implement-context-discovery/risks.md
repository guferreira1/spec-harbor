# Risks: Implement Context Discovery

## Assumptions Becoming Facts

Conventional suggestions such as `go test ./...`, `mvn test`, or `npm test` can look authoritative even when no repository file explicitly declares them.

Mitigation: classify conventional values as `suggested_assumption`, require source evidence, render assumptions in a separate group, and prevent assumptions from becoming project brief answers unless the user confirms them in `specharbor brief`.

## Over-Reading The Repository

Context discovery could drift into repository-wide indexing, source-code semantic analysis, dependency graph analysis, local retrieval, snippet ranking, embeddings, vector databases, or RAG.

Mitigation: restrict the source catalog to explicitly listed files and directories, enforce skip rules, keep traversal bounded and deterministic, and leave indexing, retrieval, snippet ranking, embeddings, and RAG explicitly out of scope.

## Sensitive File Exposure

Discovery that walks documentation or configuration folders could accidentally read secrets, keys, credentials, environment files, or generated artifacts that should never appear in CLI output.

Mitigation: define a tested skip policy for sensitive filenames and heavy/generated directories, apply the policy before reads, avoid following symlinks, and never print raw file contents.

## Project Brief Precedence Confusion

If `.specharbor/project-brief.md` conflicts with repository files, users may be confused about which context SpecHarbor trusts.

Mitigation: classify parsed brief values as `user_confirmed_context`, render them first, state their source, keep conflicting repository evidence separately labeled, and avoid merge or update behavior in this change.

## Breaking Existing Brief Behavior

Adding suggestions to `specharbor brief` could accidentally change the current confirmation-first and write-if-absent behavior.

Mitigation: treat discovery as an optional suggestion source only, preserve the existing question and confirmation flow, keep user selection required for confirmed answers, and add regression tests for cancellation, retries, existing brief refusal, and successful brief creation.

## CLI Business Logic Drift

The CLI formatter could start deciding what counts as a fact, assumption, confirmed value, or high-confidence signal.

Mitigation: keep classifications, confidence, source categories, skip policy, and deterministic ordering in domain/usecase code. The CLI should only parse arguments, invoke the use case, and print structured results.

## Parser Fragility

Parsing common files such as `package.json`, `pyproject.toml`, `Makefile`, workflow YAML, or project brief Markdown could become brittle or require large dependencies.

Mitigation: use structured parsers already available in the codebase or standard library when practical, keep extraction shallow and conservative, avoid large framework dependencies, and emit ambiguity notes instead of guessing when parsing is unclear.

## Scope Creep Into Prompt Injection

Because discovery produces project context, it may be tempting to inject discovered signals into generated agent prompts immediately.

Mitigation: keep prompt injection and role prompt changes out of scope. Future context-aware prompts require a separate OpenSpec change with its own safety and precedence rules.

## Command Probing Expectations

Users may assume discovered commands were executed or verified.

Mitigation: never run package managers, tests, builds, scripts, shells, agents, or workflow tools during discovery. Label command sources and distinguish explicit command facts from conventional assumptions.

## Documentation Drift

Public docs could describe discovery as RAG, indexing, remote discovery, command verification, or automatic prompt enrichment.

Mitigation: update documentation only after implementation, describe the command as local/offline/read-only discovery, and explicitly document all out-of-scope automation and remote behavior.
