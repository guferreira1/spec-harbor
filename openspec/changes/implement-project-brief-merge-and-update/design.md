# Design: Implement Project Brief Merge And Update

## Overview

This change adds a controlled update path for `.specharbor/project-brief.md`. It builds on the existing briefing and context discovery foundations without changing their safety model.

The implementation must preserve the existing Hexagonal Architecture direction:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

The CLI owns terminal prompts and user-facing formatting. The domain owns brief fields, source categories, merge decisions, conflict records, stale records, and deterministic rendering rules. The use case orchestrates parsing, discovery input, proposal creation, applying user decisions, preview data, final-confirmation state, rendering, and safe write policy. Filesystem reads and writes stay behind ports and adapters.

## Command Contract

Supported commands:

```text
specharbor brief
specharbor brief --update
```

`specharbor brief` remains the current create-if-absent flow.

`specharbor brief --update` is the only update path. It must:

- reject positional arguments;
- reject unrelated flags such as `--force`, `--overwrite`, `--json`, `--github`, `--rag`, and `--from-scan`;
- require an interactive TTY;
- require an existing `.specharbor/project-brief.md`;
- fail clearly if the brief does not exist and tell the user to run `specharbor brief` first;
- never write before the final confirmation prompt succeeds.

## Existing Brief Detection

Before writing, SpecHarbor must check:

```text
.specharbor/project-brief.md
```

Behavior:

- `specharbor brief` with no `--update` preserves current write-if-absent behavior and refuses to overwrite an existing brief.
- `specharbor brief --update` requires the file to exist.
- A missing file in update mode fails before prompting for merge decisions.
- Symlink and unsafe path behavior must follow existing filesystem adapter safety rules.

## Source Categories

The updated brief must preserve and clearly separate these categories:

- `user_confirmed_context`: values explicitly confirmed by the user through `brief` or `brief --update`;
- `detected_fact`: explicit repository evidence returned by local context discovery;
- `suggested_assumption`: conventional or incomplete inference returned by local context discovery.

The implementation may keep existing display labels such as `user-provided answer`, `detected context`, and `assumption` if compatibility requires it, but the core model must still preserve the category semantics above.

Detected facts and suggested assumptions must never become `user_confirmed_context` unless the user explicitly selects that action and also confirms the final write.

## Brief Parsing

Add conservative parsing for the deterministic Markdown format currently rendered by project brief creation.

Supported parsing scope:

- known answer sections:
  - `## Project type`;
  - `## Purpose`;
  - `## Target users`;
  - `## Stack`;
  - `## Architecture`;
  - `## Agent behavior`;
  - command subsections under `## Commands`: `### Install`, `### Test`, `### Build`, `### Run`;
- known detected context records under `### Detected context`;
- known assumption records under `## Assumptions`.

Parsing rules:

- parse only known headings and known `Answer:` lines for confirmed values;
- ignore unknown custom sections instead of treating them as confirmed context;
- preserve unsupported or unknown content only if the renderer has a deliberate compatibility mechanism for it;
- return structured parse errors for malformed known sections that prevent safe updates;
- never infer missing confirmed values from detected context or assumptions.

If a known confirmed field cannot be parsed, the update flow should surface it as missing or unsafe and ask the user for an explicit value before writing, or stop with a clear error if safe rendering cannot be guaranteed.

## Discovery Reuse

The update flow must reuse the existing context discovery use case. It must not duplicate detector logic in CLI code or in a separate update-specific scanner.

Current discovery results provide:

- user-confirmed context parsed from the existing brief;
- detected facts from supported local sources;
- suggested assumptions from supported local sources.

The update proposal should use current detected facts and suggested assumptions as candidates for review. It must not execute commands, run package managers, index repositories, call providers, call agents, or query remote services.

## Merge Model

Add domain models that make update decisions explicit and testable. Names may vary, but the model should represent:

- `ProjectBriefField`;
- `ProjectBriefSourceCategory`;
- `ProjectBriefParsed`;
- `ProjectBriefUpdateProposal`;
- `ProjectBriefFieldProposal`;
- `ProjectBriefMergeDecision`;
- `ProjectBriefConflict`;
- `ProjectBriefStaleRecord`;
- `ProjectBriefUpdateResult`.

Required decision kinds:

- keep existing confirmed value;
- replace with a custom user answer;
- accept a detected fact as confirmed context;
- accept a suggested assumption as confirmed context only after explicit user selection;
- ignore a detected fact;
- keep stale assumptions;
- remove stale assumptions;
- cancel update.

Default decision:

- keep existing confirmed value;
- keep existing detected context records unless the user removes or ignores them;
- keep existing assumptions unless the user removes stale assumptions;
- do not promote detected facts or assumptions.

## Conflict Handling

A conflict exists when a current detected fact for a known field differs from the existing confirmed value for the same field.

Conflict behavior:

- show existing confirmed value first;
- show the detected fact as evidence with source and confidence;
- default to keeping the existing confirmed value;
- let the user explicitly choose to replace the confirmed value;
- never silently replace confirmed context with detected evidence.

Conflicts should be included in the preview so the user can review the final decision set before writing.

## Stale Data Handling

The update flow should identify stale data without automatic deletion.

Potential stale cases:

- an existing confirmed value has no matching current detected fact for the same field;
- an existing detected context record no longer appears in current discovery output;
- an existing assumption no longer appears in current suggested assumptions.

Behavior:

- show stale confirmed values as "possibly stale" but keep them by default;
- show stale detected context and stale assumptions separately;
- allow stale assumptions to be kept or removed;
- do not delete confirmed values automatically;
- do not delete detected context automatically unless the user explicitly ignores or removes a record.

## CLI Interaction

The CLI prompt layer should be deterministic and concise.

Recommended flow for `specharbor brief --update`:

1. Check TTY and argument validity.
2. Load and parse the existing brief through the update use case.
3. Run local context discovery through the existing use case.
4. Print an update summary:
   - target file;
   - parsed confirmed fields;
   - number of detected fact candidates;
   - number of suggested assumption candidates;
   - conflict count;
   - stale record count;
   - safety boundaries.
5. Prompt field-by-field only when decisions are needed or useful:
   - keep current value;
   - accept a listed detected fact;
   - enter custom value;
   - keep current and ignore candidate;
   - cancel.
6. Prompt for stale assumptions:
   - keep;
   - remove;
   - cancel.
7. Show a reviewable preview of the proposed final brief.
8. Ask final confirmation:

```text
Write updated project brief? [y/N]:
```

Confirmation accepts trimmed `y` and `yes` in any casing. Empty input, EOF, `n`, and `no` cancel. Unsupported confirmation answers retry up to the existing interactive retry limit.

Any cancellation must return `operation cancelled` style behavior and leave the file unchanged.

The CLI may collect terminal responses, but the write use case must still receive an explicit final-confirmed input and must refuse to write when that input is false. This keeps confirmation-first behavior enforced in core orchestration rather than relying only on CLI control flow.

## Preview Behavior

Before writing, the user must see a reviewable summary.

The preview may be structured, for example:

```text
Project brief update preview:

Confirmed context:
- Stack: keep existing `Go`
- Test command: replace with detected fact `go test ./...` from README.md

Detected facts:
- ignored: Build command `go build ./...` from docs/usage.md

Assumptions:
- kept stale assumption: Run command may be `go run ./cmd/specharbor`

Conflicts:
- Architecture: existing `Clean Architecture / Hexagonal`; detected `Hexagonal Architecture` from openspec/specs/architecture/spec.md
```

The preview does not need to be a unified diff, but it must clearly show what will change and what will remain unchanged.

## Rendering

Rendering must remain deterministic and human-readable.

The updated file should preserve the existing major structure:

```text
# Project Brief

## Project type
## Purpose
## Target users
## Stack
## Architecture
## Commands
### Install
### Test
### Build
### Run
## Agent behavior
## Context sources
### User-provided answers
### Detected context
## Assumptions
```

If source category labels evolve, tests must prove that:

- confirmed values are rendered separately from detected facts;
- detected facts are rendered separately from assumptions;
- assumptions are labeled as assumptions;
- stable ordering is used for repeated detected fact and assumption records;
- repeated renders of the same update result are byte-for-byte identical.

## Safe Write Policy

The update use case must not use the create-only `WriteFileIfAbsent` operation for updates.

Add or reuse a small port method for safe replacement, such as:

```text
ReadFile(root string, relativePath string) (string, error)
WriteFileSafely(root string, relativePath string, contents string) error
```

The concrete filesystem adapter must guarantee that a failed update does not leave a partial or empty project brief. A temp-file-and-rename strategy is preferred. If another approach is used, tests must prove the original file remains intact on write failure.

The write path must avoid following symlinks and must keep existing safe relative path checks.

## Architecture Responsibilities

Domain responsibilities:

- brief field model;
- source category model;
- parsing result model;
- merge decision model;
- conflict and stale record models;
- deterministic render model;
- validation that confirmed fields are not built from unconfirmed sources unless explicit user confirmation was recorded.

Use case responsibilities:

- validate dependencies and project root;
- read the existing brief through a port;
- parse the brief using domain/core logic;
- obtain current discovery results through an injected context provider or existing discovery use case;
- build update proposals;
- apply CLI-collected decisions;
- produce preview/update result data;
- render final Markdown;
- require explicit final-confirmed input before invoking safe write;
- return a no-write or cancelled result when final confirmation is false.

Port responsibilities:

- expose only filesystem and context provider operations the use case consumes.

Adapter responsibilities:

- CLI parses `--update`, performs terminal interaction, formats previews, and passes explicit decisions to the use case;
- filesystem adapter implements safe reads/writes;
- context discovery adapter/use case remains the single source of detected facts and assumptions.

Forbidden responsibilities:

- CLI must not own merge rules.
- Domain/usecase must not import adapters, CLI packages, `os`, terminal IO, network APIs, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.
- This change must not add provider, vector store, repository index, remote context, command execution, workflow connector, source-control automation, release, npm, Homebrew, `install.sh`, GoReleaser, or publishing abstractions.

## Documentation

After implementation exists, update user-facing docs to cover:

- `specharbor brief --update`;
- confirmation-first behavior;
- preview before write;
- source category separation;
- conflict and stale handling;
- cancellation leaves the file unchanged;
- no indexing, retrieval, RAG, provider API, command execution, agent execution, source-control automation, or release/publishing behavior.

At minimum update `README.md`, `docs/usage.md`, and `docs/workflow.md` if they describe `brief` behavior.

## Testing Plan

Testing must include:

- project brief parsing;
- merge decision models;
- conflict detection;
- stale record detection;
- update rendering;
- deterministic Markdown output;
- no existing brief in update mode;
- normal `specharbor brief` regression;
- update cancelled before decisions;
- update cancelled at final confirmation;
- update confirmed;
- keep current value;
- replace with detected fact;
- replace with custom value;
- ignore detected fact;
- detected fact not promoted without confirmation;
- suggested assumption not promoted without confirmation;
- suggested assumption promoted only after explicit confirmation if supported as a field value;
- stale value surfaced but not removed automatically;
- stale assumption keep/remove choices;
- write failure preserves original file;
- CLI interactive update flow;
- `specharbor context discover` regression;
- context-aware `specharbor prompt` regression;
- `specharbor scan` regression;
- OpenSpec validation.
