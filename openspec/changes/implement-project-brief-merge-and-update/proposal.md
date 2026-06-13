# Proposal: Implement Project Brief Merge And Update

## Problem

SpecHarbor can create `.specharbor/project-brief.md` through `specharbor brief` and can discover local context through `specharbor context discover`. The current briefing flow deliberately refuses to merge, update, overwrite, or append when a project brief already exists. That preserved user-owned context for the foundation, but it leaves users without a safe way to reconcile confirmed context with new repository evidence.

Without an explicit update flow, users must hand-edit the brief or delete it and start over. That is risky because detected facts and suggested assumptions must not become confirmed project context unless the user explicitly confirms them.

## Goal

Add a safe, confirmation-first update flow for existing project briefs:

```text
specharbor brief --update
```

The update flow must detect an existing `.specharbor/project-brief.md`, parse known confirmed values conservatively, compare them with current local discovery output, show a reviewable preview of proposed changes, ask before writing, and preserve user-confirmed context by default.

## Command Shape

The command shape is:

```text
specharbor brief --update
```

This extends the existing `brief` command because the feature updates the same project-owned context artifact created by `specharbor brief`. Existing CLI conventions already place project context collection under `brief`, and the current command explicitly rejects flags, including `--update`; accepting that flag is a narrow compatibility evolution instead of introducing a parallel command family.

`specharbor brief` without `--update` remains the safe create-if-absent flow. It must still refuse to silently overwrite an existing brief.

## Scope

- Detect whether `.specharbor/project-brief.md` exists before writing.
- Keep `specharbor brief` without `--update` safe and compatible with current write-if-absent behavior.
- Add explicit update behavior only through `specharbor brief --update`.
- Require an interactive TTY for the update flow.
- Parse known project brief sections conservatively:
  - project type;
  - purpose;
  - target users;
  - stack;
  - architecture;
  - install command;
  - test command;
  - build command;
  - run command;
  - agent behavior;
  - detected context records;
  - assumption records.
- Reuse existing `context discover` logic as the source of current detected facts and suggested assumptions.
- Preserve these source categories and keep them clearly separated:
  - `user_confirmed_context`;
  - `detected_fact`;
  - `suggested_assumption`.
- Preserve existing user-confirmed values by default.
- Surface conflicts between existing confirmed values and current detected facts.
- Prefer existing confirmed context when conflicts exist.
- Let the user explicitly choose whether to keep, replace, or confirm a new value.
- Support these update decisions:
  - keep existing confirmed value;
  - replace with a newly entered user answer;
  - accept a detected fact as a confirmed value;
  - ignore a detected fact;
  - keep stale assumptions;
  - remove stale assumptions;
  - cancel the update.
- Allow suggested assumptions to become confirmed context only through explicit user selection and final confirmation.
- Never promote detected facts or suggested assumptions to `user_confirmed_context` by default.
- Identify potentially stale confirmed values or stale detected/assumption records without deleting them automatically.
- Show a clear reviewable preview before writing. A structured field-by-field summary is acceptable; a textual diff-like summary is also acceptable.
- Require final confirmation before writing.
- Write no file when the user cancels.
- Render the updated project brief as deterministic, human-readable Markdown.
- Use atomic or otherwise safe write behavior so a write failure preserves the existing brief.
- Update public documentation after implementation to describe the update flow and safety boundaries.
- Add focused tests for parsing, merge decisions, conflict handling, rendering, CLI interaction, cancellation, safe writes, and regressions.

## Out Of Scope

- Repository-wide indexing.
- Local retrieval.
- Snippet ranking.
- Embeddings.
- Vector databases.
- RAG.
- GitHub remote context.
- GitLab or other remote context.
- Provider APIs.
- Local model APIs.
- Automatic command verification.
- Automatic project command execution.
- Agent execution.
- Prompt execution.
- Source-control automation.
- Release automation.
- npm changes.
- Homebrew changes.
- `install.sh` changes.
- GoReleaser changes.
- Publishing flows.
- Archiving this OpenSpec change before merge.
- Changing final decision labels or agent role behavior.

## Compatibility

- `specharbor brief` without `--update` continues to require an interactive TTY and continues to write only when `.specharbor/project-brief.md` is absent.
- Existing cancellation behavior remains safe.
- Existing `.specharbor/project-brief.md` files are not overwritten silently.
- `specharbor context discover` remains local, offline, read-only, and unchanged except that its existing use case may feed update suggestions.
- Context-aware `specharbor prompt` continues to consume project brief context safely and must still keep confirmed context, detected facts, and suggested assumptions separate.
- Existing `specharbor scan` behavior remains unchanged.
- Existing final decision labels and role prompt behavior remain unchanged.

## Success Criteria

- `specharbor brief --update` is specified and implemented as an explicit update flow.
- Existing project briefs are detected and never overwritten silently.
- The update flow is confirmation-first and writes only after final confirmation.
- Cancelling at any prompt leaves `.specharbor/project-brief.md` unchanged.
- Existing user-confirmed values are preserved by default.
- Detected facts are shown as evidence or suggestions and are not confirmed without explicit user action.
- Suggested assumptions are shown as assumptions and are not confirmed without explicit user action.
- Conflicts are surfaced safely with existing confirmed context preferred by default.
- Stale values are surfaced but not removed automatically.
- The updated project brief is deterministic and human-readable.
- Safe write behavior preserves the original brief on write failure.
- Tests cover the required domain, use case, adapter, and CLI flows.
- No indexing, retrieval, embeddings, RAG, remote context, provider APIs, command execution, agent execution, source-control automation, release automation, npm, Homebrew, `install.sh`, GoReleaser, or publishing behavior is introduced.
- `go run ./cmd/specharbor validate implement-project-brief-merge-and-update` passes with zero errors.
