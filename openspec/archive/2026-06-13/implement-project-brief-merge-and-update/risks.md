# Risks: Implement Project Brief Merge And Update

## Overwriting User-Owned Context

Updating an existing project brief could accidentally replace values the user already confirmed.

Mitigation: make `specharbor brief --update` explicit, preserve existing confirmed values by default, show a preview before writing, require final confirmation, and keep normal `specharbor brief` write-if-absent behavior.

## Detected Facts Becoming Confirmed Facts

Detected repository evidence can look authoritative even though it is not user-confirmed context.

Mitigation: model `detected_fact` separately, require an explicit accept decision before using a detected fact as a confirmed value, and test that detected facts are never promoted by default.

## Suggested Assumptions Becoming Facts

Suggested assumptions are weaker than detected facts and must not be treated as confirmed project decisions.

Mitigation: keep `suggested_assumption` records separate, label assumptions clearly in previews and Markdown, and require explicit user selection plus final confirmation before any assumption can become confirmed context.

## Unsafe Conflict Resolution

Confirmed context may conflict with current detected facts, such as an existing stack value differing from current manifests.

Mitigation: surface conflicts with source evidence, prefer existing confirmed context by default, and let the user explicitly choose whether to update the confirmed value.

## Stale Data Removal

The update flow could delete stale confirmed values or assumptions too aggressively.

Mitigation: identify stale values as review items, keep confirmed values and assumptions by default, and remove only stale assumptions or ignored records when the user explicitly chooses that action.

## Partial Writes

A failure during update writing could truncate or partially rewrite `.specharbor/project-brief.md`.

Mitigation: add safe replacement behavior behind the filesystem port, prefer temp-file-and-rename semantics, avoid following symlinks, and add tests proving the original file remains intact on write failure.

## CLI Owning Merge Rules

Interactive prompts could grow business rules for merge decisions inside the CLI adapter.

Mitigation: keep CLI code limited to prompt rendering, choice collection, and preview formatting. Put merge decisions, conflict detection, stale detection, category preservation, and rendering in domain/usecase code.

## Scope Creep Into Context Intelligence

Brief update work could expand into repository indexing, retrieval, embeddings, RAG, remote context, provider APIs, or command verification.

Mitigation: reuse existing context discovery only, explicitly exclude indexing/retrieval/RAG/remote/provider/command execution behavior, and add or update architecture tests for forbidden dependencies and terminology where useful.

## Documentation Drift

Documentation could imply that update behavior verifies commands, performs repository intelligence, or changes prompt/agent execution behavior.

Mitigation: update docs only after behavior exists, document the confirmation-first safety model, and explicitly state that no indexing, RAG, provider calls, command execution, agent execution, source-control automation, release automation, or publishing behavior is added.

## Backward Compatibility Regressions

Changing `brief` argument parsing could break existing `specharbor brief`, `context discover`, context-aware `prompt`, or `scan` behavior.

Mitigation: preserve the existing create-if-absent path, add regression tests for the existing commands, and run the required CLI verification commands before PR.
