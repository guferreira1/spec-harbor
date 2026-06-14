# Remaining Context Initiative Plan

## Purpose

This document plans the remaining Context Initiative work. It is a planning artifact only. It does not create active OpenSpec changes, does not implement product code, and does not modify production behavior.

The completed foundation already provides:

- `specharbor brief` for confirmed project context in `.specharbor/project-brief.md`.
- `specharbor context discover` for bounded local/offline classified context.
- context-aware `specharbor prompt <change-id> --role <role>` output for the five supported roles.

The remaining features should extend that lifecycle without weakening the current distinction between user-confirmed context, detected facts, and suggested assumptions.

## Branch Strategy

Use one branch for this planning document:

```text
chore/plan-remaining-context-initiative
```

Use separate feature branches for actual implementation:

```text
feat/implement-project-brief-merge-and-update
feat/implement-repository-context-index
feat/implement-local-context-retrieval
feat/implement-github-remote-context
feat/implement-context-rag-provider
```

Use separate archive branches after each merged feature:

```text
chore/archive-implement-project-brief-merge-and-update
chore/archive-implement-repository-context-index
chore/archive-implement-local-context-retrieval
chore/archive-implement-github-remote-context
chore/archive-implement-context-rag-provider
```

A single branch is acceptable for this plan because it is one documentation artifact with no production behavior change. A single branch is not ideal for implementing all five features together because the features would compete for the same domain models, ports, use cases, CLI context commands, prompt integration points, docs, and tests. Combining them would make OpenSpec scope harder to enforce, review harder to reason about, failures harder to isolate, and archive history less useful.

Each implementation branch should contain exactly one active OpenSpec change and one feature's product edits. After that feature merges, archive it from a fresh branch based on updated `main` so archive housekeeping is separated from implementation review.

## Subagent Strategy

Use a main coordinator with planning subagents.

The Main Agent / Coordinator owns:

- sequencing across the five features;
- branch and worktree validation;
- OpenSpec scope control;
- architecture and dependency decisions;
- integration of subagent recommendations;
- final acceptance of each spec before implementation;
- final validation before PR;
- archive sequencing after merge.

Subagents may:

- propose OpenSpec spec guidance;
- identify risks and acceptance criteria;
- propose task breakdowns;
- inspect likely dependencies and overlap;
- recommend focused tests and documentation changes.

Subagents must not:

- implement production code simultaneously across the five features;
- edit shared production files without coordinator sequencing;
- create active OpenSpec changes unless the coordinator explicitly assigns one feature;
- change release, npm, Homebrew, `install.sh`, publishing, tag, merge, or source-control automation files;
- silently promote assumptions into facts.

Every subagent must follow OpenSpec/SDD discipline. The coordinator validates every subagent output before any implementation begins. Parallelism is for planning and review only; production edits should be sequenced by feature.

## Recommended Execution Order

Use this order:

1. `implement-project-brief-merge-and-update`
2. `implement-repository-context-index`
3. `implement-local-context-retrieval`
4. `implement-github-remote-context`
5. `implement-context-rag-provider`

Dependency reasoning:

- Project brief merge/update comes first because it stabilizes the confirmed context lifecycle. Later features need a safe way to reconcile stale or incomplete `.specharbor/project-brief.md` data without overwriting user intent.
- Repository context index comes second because retrieval and remote/RAG features need safe inventory metadata before they select, rank, or enrich context.
- Local context retrieval comes third because it can build on the index or bounded inventory while preserving local/offline behavior.
- GitHub remote context comes fourth because it should reuse the same context abstractions without becoming mandatory for local usage.
- RAG provider comes last because it depends on stable context records and retrieval boundaries, and it must keep local/offline fallback intact.

## Feature Brief: implement-project-brief-merge-and-update

Change ID: `implement-project-brief-merge-and-update`

Goal: Add a safe, confirmation-first way to merge or update `.specharbor/project-brief.md` without overwriting user-owned context or treating detected signals as confirmed facts.

Why it comes in this order: Existing `brief` intentionally refuses to merge, update, overwrite, or append. That was correct for the foundation, but the remaining context lifecycle needs a controlled update path before indexing, retrieval, remote, or RAG features rely on confirmed context.

Main scope:

- Define explicit update behavior for existing project briefs.
- Preserve known project brief sections and user-confirmed values.
- Offer detected facts and assumptions only as suggestions requiring confirmation.
- Add conflict handling between existing confirmed values and current repository evidence.
- Keep write behavior deterministic, reviewable, and safe.
- Update docs only after behavior exists.

Explicit out-of-scope boundaries:

- No repository-wide indexing.
- No retrieval, snippet ranking, embeddings, vector stores, or RAG.
- No GitHub or remote discovery.
- No provider APIs, local model APIs, or agent execution.
- No automatic command verification.
- No release, npm, Homebrew, `install.sh`, publishing, tag, merge, or source-control automation changes.

Expected architectural boundaries:

- Domain owns brief field models, confirmed/detected/assumption categories, merge decisions, and conflict records.
- Ports expose only the filesystem operations the update use case consumes.
- Use cases orchestrate parsing, proposed updates, confirmation input, rendering, and write policy.
- CLI owns prompts and user-facing formatting only.
- Concrete filesystem behavior remains in adapters.

Testing focus:

- Existing brief parsing and preservation.
- Update proposal rendering.
- Confirmation and cancellation.
- Conflict handling.
- Assumptions never becoming facts.
- No partial writes on failure.
- Regression coverage for existing `brief`, `context discover`, and `prompt`.

Documentation impact:

- Update `README.md`, `../usage.md`, and `../workflow.md` only if the implementation exposes a user-facing update command or flag.
- Document that update behavior is confirmation-first and not indexing, RAG, remote discovery, or command verification.

Main risks:

- Overwriting user-maintained context.
- Turning detected facts or assumptions into confirmed values without explicit user action.
- Hiding stale confirmed context instead of surfacing conflicts.
- Letting CLI code own merge rules.

Subagent responsibility:

- Produce spec guidance for update semantics, conflict cases, confirmation flow, acceptance criteria, and tests.
- Identify how existing brief/discovery/prompt behavior should remain stable.
- Do not write code.

Coordinator validation checklist:

- The spec has exactly five OpenSpec files.
- Scope is limited to brief merge/update.
- Existing write-if-absent behavior remains available or is intentionally evolved with explicit compatibility notes.
- No indexing, retrieval, remote context, embeddings, RAG, or release files appear in scope.
- Tests cover confirmation, conflict handling, stale data, and no silent assumption promotion.

## Feature Brief: implement-repository-context-index

Change ID: `implement-repository-context-index`

Goal: Add a safe repository context index that records bounded inventory metadata for supported context sources without performing retrieval, embeddings, RAG, or remote discovery.

Why it comes in this order: Once confirmed context can be updated safely, an index can provide stable local inventory metadata for later retrieval. The index should not select or rank snippets yet.

Main scope:

- Define the context index model and persistence/update behavior.
- Inventory supported local context sources, source categories, metadata, and freshness information.
- Reuse existing skip rules for sensitive files and heavy/generated directories.
- Keep index content bounded and deterministic.
- Provide validation or reporting that helps users understand what is indexed.

Explicit out-of-scope boundaries:

- No local retrieval or snippet ranking.
- No embeddings, vector stores, or RAG.
- No GitHub or remote source collection.
- No prompt behavior changes unless strictly limited to reading stable metadata.
- No command execution or dependency graph analysis.
- No release or publishing changes.

Expected architectural boundaries:

- Domain owns index records, source categories, freshness markers, skip policy references, and deterministic ordering.
- Ports expose bounded filesystem inventory and index persistence operations.
- Use cases build, read, validate, and report index state.
- Adapters implement safe filesystem traversal and persistence.
- CLI parses commands and formats reports without owning index rules.

Testing focus:

- Deterministic index generation.
- Safe path handling and symlink behavior.
- Sensitive/heavy folder skip policy.
- Freshness and stale index detection.
- Empty and ambiguous repository behavior.
- No retrieval or embedding behavior.

Documentation impact:

- Document command shape and index file location only after the spec chooses them.
- Explain that the index is metadata/inventory, not a semantic database, vector store, or RAG system.

Main risks:

- Indexing too much content.
- Storing raw secrets or large raw file bodies.
- Creating a persistence format that retrieval cannot safely reuse.
- Users mistaking inventory metadata for confirmed context.

Subagent responsibility:

- Propose index schema options, safety constraints, acceptance criteria, and migration/staleness risks.
- Identify overlap with existing `context discover` sources.
- Do not write code.

Coordinator validation checklist:

- The spec explicitly says indexing does not implement retrieval.
- The index model preserves source evidence and classification boundaries.
- Skip policies are inherited or clearly extended.
- Persistence path and file format are justified.
- Tests prove deterministic and bounded behavior.

## Feature Brief: implement-local-context-retrieval

Change ID: `implement-local-context-retrieval`

Goal: Add deterministic local/offline context retrieval over the bounded index or inventory so SpecHarbor can select relevant local context without embeddings, RAG, provider APIs, or remote services.

Why it comes in this order: Retrieval needs a safe inventory first. It should prove the local abstraction before remote context or RAG providers add optional sources.

Main scope:

- Define local retrieval queries and result models.
- Retrieve bounded snippets or structured records from local supported sources.
- Use deterministic lexical, metadata, or rule-based ranking.
- Preserve source evidence, classification, and confidence.
- Make retrieval safe for future prompt/spec use without raw dumping.

Explicit out-of-scope boundaries:

- No embeddings.
- No vector databases.
- No RAG provider.
- No GitHub or remote context.
- No mandatory project brief update.
- No execution of project commands.
- No release or publishing changes.

Expected architectural boundaries:

- Domain owns retrieval query, result, rank, snippet, and source evidence models.
- Ports expose local indexed source access and bounded file reads.
- Use cases orchestrate retrieval and result limiting.
- Adapters perform concrete safe reads.
- CLI or prompt integration only formats returned results and must not own ranking rules.

Testing focus:

- Deterministic ranking and tie-breaks.
- Result limits and truncation.
- Safe relative paths.
- Skip policy enforcement.
- No assumption promotion.
- Behavior with missing or stale index.
- Regression coverage for discovery and prompts.

Documentation impact:

- Document retrieval behavior, limits, and local/offline constraints after implementation.
- Clarify that retrieval is not RAG and does not call providers.

Main risks:

- Retrieval ranking becoming hidden business logic in CLI code.
- Raw file content exposure.
- Confusing retrieved snippets with confirmed context.
- Premature RAG abstractions leaking into local retrieval.

Subagent responsibility:

- Propose retrieval use cases, result constraints, ranking rules, and test cases.
- Identify which index fields are required.
- Do not write code.

Coordinator validation checklist:

- The spec explicitly says retrieval does not implement embeddings or RAG.
- The retrieval model keeps confirmed, detected, and assumption classifications separate.
- Result limits and source evidence are specified.
- Tests cover deterministic ranking and safe truncation.
- No remote or provider dependency is required.

## Feature Brief: implement-github-remote-context

Change ID: `implement-github-remote-context`

Goal: Add optional GitHub remote context collection through source-control host abstractions while preserving local/offline usage when GitHub is unavailable or unconfigured.

Why it comes in this order: Remote context should reuse the same context source, indexing, and retrieval abstractions after they are stable. It should not define those abstractions from scratch.

Main scope:

- Define optional GitHub context sources, such as repository metadata, default branch files, issues, pull requests, discussions, or workflow metadata only as explicitly approved in the spec.
- Add a core-owned source-control context port.
- Implement a GitHub adapter behind that port.
- Keep credentials optional and explicit.
- Map remote context into existing context classifications and source evidence.
- Preserve local/offline fallback.

Explicit out-of-scope boundaries:

- No requirement for GitHub credentials for local workflows.
- No source-control automation such as commit, push, PR creation, merge, tagging, or release.
- No mandatory remote calls from existing local commands.
- No GitLab, Bitbucket, or generic forge implementation unless separately specified.
- No RAG provider or embeddings.
- No release or publishing changes.

Expected architectural boundaries:

- Domain owns remote context source records and classification mapping.
- Ports define small remote context read interfaces consumed by use cases.
- Use cases orchestrate optional remote collection and fallback.
- GitHub API details, authentication, pagination, rate limits, and response mapping stay in adapters.
- CLI/config wiring must not leak GitHub SDK types into core.

Testing focus:

- Fake port tests for use-case behavior.
- Adapter tests with mocked HTTP responses or local fixtures.
- Missing credential behavior.
- Rate limit and API error handling.
- Local fallback when remote is disabled or unavailable.
- No source-control automation side effects.

Documentation impact:

- Document GitHub context as optional.
- Document credentials, failure modes, rate limits, and local fallback.
- Make clear that no PR, merge, push, release, or workflow automation is introduced.

Main risks:

- Making local context workflows depend on network or credentials.
- Pulling in too much remote data.
- Leaking private remote content into prompts without limits.
- Confusing remote issue/PR statements with confirmed project facts.

Subagent responsibility:

- Propose remote source boundaries, credential handling, fallback behavior, and tests.
- Identify which local abstractions should be reused.
- Do not write code.

Coordinator validation checklist:

- The spec states GitHub is optional and local/offline fallback remains.
- Remote data is classified and sourced, not confirmed by default.
- No source-control automation is included.
- Adapter responsibilities keep GitHub details out of core.
- Tests do not require live GitHub network access.

## Feature Brief: implement-context-rag-provider

Change ID: `implement-context-rag-provider`

Goal: Add optional RAG provider support after local retrieval and remote context boundaries are stable, while preserving deterministic local/offline fallback and agent-assisted workflows without provider API keys.

Why it comes in this order: RAG is the most dependency-heavy feature. It should build on stable confirmed context, indexing, local retrieval, and optional remote context rather than redefining those contracts.

Main scope:

- Define provider and retrieval augmentation boundaries.
- Support optional embedding or retrieval providers only through small core-owned ports.
- Keep local retrieval available when provider configuration is absent.
- Preserve classification, source evidence, and assumptions.
- Bound prompt/spec context size and raw content exposure.
- Add provider error handling and fallback behavior.

Explicit out-of-scope boundaries:

- No removal of local/offline context retrieval.
- No provider API keys required for agent-assisted workflows.
- No silent provider calls from unrelated commands.
- No source-control automation.
- No release or publishing changes.
- No treating generated RAG summaries as confirmed facts.

Expected architectural boundaries:

- Domain owns augmented retrieval result models, source evidence, confidence, and fallback status.
- Ports define small provider interfaces for embedding, vector search, or retrieval augmentation as needed.
- Use cases choose between local retrieval and optional provider-backed augmentation.
- Provider-specific authentication, payloads, errors, rate limits, retries, and response mapping stay in adapters.
- CLI/config only wires explicit provider configuration.

Testing focus:

- Fake provider tests for success, timeout, auth error, and rate-limit behavior.
- Local fallback when provider is unavailable or disabled.
- No provider calls without explicit configuration.
- Classification preservation through augmentation.
- Prompt/context size limits.
- Secret handling and config safety.

Documentation impact:

- Document RAG as optional augmentation.
- Document local/offline fallback and no-provider behavior.
- Document that RAG output is not user-confirmed context.
- Avoid claiming provider support beyond what is implemented.

Main risks:

- Making RAG mandatory.
- Requiring API keys for agent-assisted workflows.
- Letting provider output override confirmed context.
- Leaking secrets or raw private content.
- Adding broad placeholder abstractions before concrete behavior needs them.

Subagent responsibility:

- Propose provider boundary options, fallback rules, provider error cases, and tests.
- Confirm how local retrieval remains the baseline.
- Do not write code.

Coordinator validation checklist:

- The spec states local/offline fallback is mandatory.
- Provider calls require explicit configuration.
- Agent-assisted workflows do not require provider API keys.
- RAG output is classified and sourced, not confirmed.
- Tests cover provider failure and fallback.

## Conflict And Dependency Rules

To prevent overlap:

- `implement-project-brief-merge-and-update` must not implement indexing.
- `implement-repository-context-index` must not implement retrieval.
- `implement-local-context-retrieval` must not implement embeddings or RAG.
- `implement-github-remote-context` must not become mandatory for local usage.
- `implement-context-rag-provider` must not remove local/offline fallback.
- No feature should change release, npm, Homebrew, `install.sh`, GoReleaser, publishing, tag, merge, or source-control automation behavior without a separate release spec.
- No feature should silently promote assumptions into facts.
- Detected facts are not user-confirmed context unless the user explicitly confirms them through a specified flow.
- Remote context is not user-confirmed context by default.
- RAG summaries are not user-confirmed context by default.
- Prompt context must remain bounded and must not dump raw files.
- Shared files must be sequenced by the coordinator when multiple features need them.

## OpenSpec/SDD Workflow

Each actual feature should follow this workflow:

```text
Spec Author -> Architecture Reviewer -> Implementer -> Tester -> Change Reviewer -> PR -> merge -> Archive Housekeeping
```

Each actual feature must still create exactly five OpenSpec files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Each actual implementation requires:

- a dedicated worktree;
- the expected branch;
- a clean starting tree;
- branch and status verification before work starts;
- explicit staging of intended files only;
- no `git add -A`;
- focused tests and validation before PR;
- a PR for the implementation branch;
- merge before archive;
- a separate archive branch after merge;
- archive validation after moving the change into `openspec/archive/<date>/<change-id>/`.

Implementation agents must read `AGENTS.md`, `.specharbor/rules/global.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and the active change before production edits.

## Subagent Prompt Skeletons

### Brief Merge/Update Planning Subagent

```text
You are the Brief Merge/Update Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Read the completed brief, discovery, and context-aware prompt archives. Plan the OpenSpec scope for implement-project-brief-merge-and-update. Define goals, out-of-scope boundaries, architecture responsibilities, risks, acceptance criteria, and tests. Preserve confirmation-first behavior and never promote detected facts or assumptions into confirmed context without explicit user confirmation.
```

### Repository Index Planning Subagent

```text
You are the Repository Index Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-repository-context-index. Focus on bounded inventory metadata, deterministic index behavior, safe paths, skip rules, persistence choices, stale index handling, and tests. Do not include retrieval, embeddings, RAG, GitHub remote context, provider calls, or release changes.
```

### Local Retrieval Planning Subagent

```text
You are the Local Retrieval Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-local-context-retrieval. Define local/offline retrieval models, ranking constraints, result limits, source evidence, and tests. Build on the repository index or bounded inventory. Do not include embeddings, vector databases, RAG providers, GitHub remote context, or provider APIs.
```

### GitHub Remote Context Planning Subagent

```text
You are the GitHub Remote Context Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-github-remote-context. Define optional remote context sources, source-control context ports, GitHub adapter responsibilities, credential handling, API failure behavior, rate-limit behavior, classification mapping, and local/offline fallback. Do not include source-control automation, PR creation, merge, push, release, embeddings, or RAG.
```

### RAG Provider Planning Subagent

```text
You are the RAG Provider Planning Subagent for SpecHarbor.

Produce planning/spec guidance only. Do not write code, do not create files, and do not edit production docs.

Plan implement-context-rag-provider. Define optional provider boundaries, fallback to local retrieval, classification preservation, provider configuration, error handling, and tests. Do not require provider API keys for agent-assisted workflows. Do not remove local/offline behavior or treat RAG output as confirmed context.
```

### Main Coordinator Review Agent

```text
You are the Main Coordinator Review Agent for SpecHarbor.

Review subagent planning/spec guidance only. Do not write production code.

Validate sequence, scope boundaries, architecture boundaries, OpenSpec file completeness, conflict rules, testing focus, documentation impact, branch/worktree requirements, and archive sequencing. Reject guidance that overlaps features, silently promotes assumptions into facts, makes remote/RAG mandatory, or touches release/publishing/source-control automation without a separate spec.
```

## Decision Gates

Do not begin implementation until all gates pass:

- This planning document has been reviewed.
- Feature order is accepted or deliberately changed with written rationale.
- The first OpenSpec change is selected.
- No overlapping active changes exist unless intentionally planned.
- Scope boundaries are approved for the selected feature.
- Main branch is clean before branching or creating a worktree.
- The selected implementation worktree is on the expected branch.
- The selected implementation worktree starts clean.
- Archive state is clean and no stale active change exists for the selected feature.
- The feature's five OpenSpec files are complete before implementation.
- Architecture review has no blocking findings.
- Test and validation commands are agreed before implementation.

## Recommended Next Implementation

The next concrete feature should be:

```text
implement-project-brief-merge-and-update
```

It should be next because the project already has confirmed context creation, local discovery, and context-aware prompts, but it still lacks a safe way to reconcile an existing brief with new or changed repository evidence. Stabilizing that confirmed-context lifecycle first prevents later index, retrieval, GitHub, and RAG work from building on stale or hard-to-update project brief data.
