# Design: Implement Project Briefing Foundation

## Overview

`specharbor brief` is an interactive CLI workflow for collecting explicit project context from the user and writing it to `.specharbor/project-brief.md`.

The feature is intentionally narrow. It creates a human-readable project brief that future features can consume, but it does not change prompt generation, agent execution, repository indexing, RAG, provider integration, source-control behavior, release behavior, or publishing behavior.

The command should follow the existing Hexagonal Architecture boundary:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

The CLI adapter owns terminal prompting and user-facing output. The use case owns briefing orchestration and write policy. The domain owns the brief data model, supported question definitions, source classification, and deterministic Markdown rendering policy if no adapter-specific renderer is needed. Filesystem behavior remains behind a small core-owned port and a concrete filesystem adapter.

## Command Contract

Supported command:

```text
specharbor brief
```

Rejected in the first version:

```text
specharbor brief <anything>
specharbor brief --force
specharbor brief --update
specharbor brief --overwrite
specharbor brief --json
specharbor brief --from-scan
specharbor brief --github
specharbor brief --rag
```

The command must reject every positional argument and every flag with clear errors. It must not support a non-interactive mode in this first change.

When stdin is not a TTY, the command must fail before prompting and must write nothing. Use the same terminal boundary pattern as existing interactive generation: terminal detection and prompt rendering stay in `internal/adapters/cli`, not in core packages.

## Prompt Flow

The command should start with a short context line:

```text
SpecHarbor could not find enough confirmed project context.
```

Then it asks a deterministic sequence of questions. Each question must present three to five options. The final option must always be:

```text
Other / custom
```

The first version must ask at least these questions, in this order:

1. What type of project is this?
2. What is the primary purpose of this project?
3. Who are the target users?
4. What stack should agents assume only after confirmation?
5. What architecture should agents preserve?
6. What install command should agents use?
7. What test command should agents use?
8. What build command should agents use?
9. What run command should agents use?
10. What should agents do when context is missing?

Recommended option sets:

```text
Project type:
1. Backend API
2. Full-stack web application
3. CLI/tooling project
4. Library/package
5. Other / custom

Project purpose:
1. Customer-facing product
2. Internal operations tool
3. Developer productivity tool
4. Data or automation pipeline
5. Other / custom

Target users:
1. External customers
2. Internal business users
3. Developers or platform engineers
4. Operators or support teams
5. Other / custom

Stack:
1. Go
2. Node.js / TypeScript
3. Python
4. Multi-stack or monorepo
5. Other / custom

Architecture:
1. MVC / layered architecture
2. Clean Architecture / Hexagonal
3. DDD-oriented modules
4. Simple modular structure
5. Other / custom

Install command:
1. No install command
2. npm install
3. pnpm install
4. go mod download
5. Other / custom

Test command:
1. No test command
2. go test ./...
3. npm test
4. pytest
5. Other / custom

Build command:
1. No build command
2. go build ./...
3. npm run build
4. make build
5. Other / custom

Run command:
1. No run command
2. go run ./cmd/<name>
3. npm run dev
4. make run
5. Other / custom

Missing context behavior:
1. Ask before assuming
2. Suggest assumptions and ask for confirmation
3. Proceed with best-effort suggestions
4. Other / custom
```

The implementation may adjust default options if the final option remains `Other / custom`, each question has three to five options, and the options are deterministic. If detected local context is available, the CLI may add or reorder a suggestion only when the output remains deterministic and still satisfies the option count and final custom option rule.

Selecting `Other / custom` must prompt for a non-empty custom answer. Invalid menu choices and empty required custom answers should retry up to three attempts, then fail clearly and write nothing.

## Detected Context And Suggestions

This foundation may use existing local scan signals as optional suggestions, but detected context must never become a confirmed fact unless the user selects or confirms it.

Allowed detected context sources:

- existing `specharbor scan` style local, deterministic, shallow project signals;
- existing OpenSpec or SpecHarbor files if detected by the scan foundation;
- package manager and ecosystem hints already modeled by the scan foundation.

Detected command hints, such as `go test ./...` or `npm test`, are suggestions only. They are not verified commands and must be labeled as detected suggestions in the brief unless the user confirms them as an answer.

The command must not:

- recursively index the repository;
- read arbitrary source files for semantic analysis;
- parse dependency graphs;
- compute embeddings;
- call vector databases;
- call RAG providers;
- call GitHub, GitLab, or other remote APIs;
- run package managers or command probes;
- call AI providers or agent tools.

## Confirmation And Cancellation

Before any write, the CLI must print a deterministic summary:

```text
SpecHarbor will create:

.specharbor/project-brief.md

Confirm? [y/N]
```

The summary should include the target file and a concise statement that:

- stack, architecture, and commands come from confirmed answers only;
- detected context remains separate from user answers;
- assumptions are not confirmed facts;
- no repository indexing, RAG, provider API, agent execution, source-control automation, release, or publishing behavior will run.

Confirmation rules:

- trimmed `y` and `yes`, in any casing, proceed;
- trimmed `n` and `no`, in any casing, cancel;
- empty input cancels;
- EOF cancels;
- unsupported confirmation input retries up to three attempts;
- cancellation returns a clear `operation cancelled` style error and writes nothing;
- confirmation retry exhaustion fails and writes nothing.

## Project Brief Markdown

The generated file must be deterministic Markdown. It should use stable headings and stable ordering so tests can compare exact output.

Required structure:

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

## Assumptions
```

Each answer section should include a source label such as:

```text
Source: user-provided answer
```

The `Context sources` section should contain separate subsections for:

```text
### User-provided answers

### Detected context
```

The user-provided subsection should list the fields answered during the prompt flow. The detected context subsection should list only local deterministic scan signals or `None recorded.` when none were available.

The `Assumptions` section must never present suggestions as facts. If no assumptions are explicitly recorded, render:

```text
None recorded.
```

If a future implementation records assumptions, each assumption must be explicitly labeled as an assumption and must not be mixed into confirmed answer sections.

## Domain Model

Add project briefing domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- `ProjectBrief`
- `ProjectBriefAnswer`
- `ProjectBriefCommandSet`
- `ProjectBriefContextSource`
- `ProjectBriefAssumption`
- `ProjectBriefQuestion`
- `ProjectBriefOption`
- `ProjectBriefAnswerSource`

The model should distinguish at least these source categories:

- user-provided answer;
- detected suggestion;
- assumption.

The renderer or model must make it impossible to accidentally treat a detected suggestion as a confirmed answer without the prompt flow recording a user confirmation.

The model should support future detected facts by keeping context sources as structured records rather than a single free-form string. That makes it possible to add richer discovery later without rewriting the briefing model.

Domain code must not import adapters, CLI packages, `os`, terminal IO, network APIs, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, concrete filesystem packages, or process execution packages.

## Ports

Add a small briefing filesystem port under:

```text
internal/core/ports
```

Expected operations:

```text
DirectoryExists(root string, relativePath string) (bool, error)
FileExists(root string, relativePath string) (bool, error)
CreateDirectory(root string, relativePath string) error
WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error)
```

The use case should own this port. The concrete implementation should reuse the local filesystem adapter where possible.

Do not add provider, source-control, workflow connector, vector store, RAG, repository indexing, agent runner, prompt injection, or command execution ports in this change.

## Use Case

Add a project briefing use case under:

```text
internal/core/usecase
```

Expected input:

- project root;
- confirmed answers;
- detected context snapshot, if any;
- assumptions, if any.

Expected behavior:

- validate required dependencies;
- validate project root is non-empty;
- validate every required answer is present;
- validate answer source classifications;
- build a deterministic `ProjectBrief`;
- render deterministic Markdown;
- ensure `.specharbor/` exists after confirmation;
- write `.specharbor/project-brief.md` with write-if-absent semantics;
- return a clear error when `.specharbor/project-brief.md` already exists;
- return a structured result with target path and write status;
- never merge, update, overwrite, append, or read an existing brief for modification;
- never call terminal IO, `os`, adapters, provider APIs, network APIs, source-control APIs, workflow tools, agent tools, command execution, or repository indexing directly.

The use case should not know how prompts are rendered. It receives confirmed answers from the CLI adapter.

## CLI Adapter

Update `internal/adapters/cli` to register `brief`.

The CLI adapter should:

- parse zero arguments;
- reject any flag or positional argument;
- obtain the current working directory as project root;
- create or reuse a terminal abstraction for prompt IO;
- fail clearly when no TTY is available;
- build deterministic question menus;
- collect custom answers when needed;
- optionally run existing shallow scan behavior before prompting to prepare suggestions;
- keep detected context labeled as suggestions until confirmed;
- print the pre-write summary;
- ask for confirmation;
- instantiate the briefing use case with filesystem dependencies after confirmation;
- print a concise success report with `.specharbor/project-brief.md` on success.

The CLI adapter may format prompt menus and summaries. It must not own project brief persistence rules, Markdown policy, detected-context source classification, or business rules about confirmed facts.

## Architecture

The feature must keep these boundaries:

- Domain contains the project brief model, answer source categories, question definitions, and deterministic render policy or render input model.
- Ports contain only the briefing filesystem operations needed by the use case.
- Use cases orchestrate validation, rendering, directory creation, and write-if-absent persistence.
- Adapters provide terminal prompting and concrete filesystem behavior.
- CLI parses arguments and formats prompt interaction only.

Core packages must not import CLI packages or adapters. CLI code must not contain business rules that decide whether detected suggestions are confirmed facts.

Future compatibility goals:

- The brief model can accept detected context records later without changing the Markdown contract.
- A future update flow can merge or update `.specharbor/project-brief.md` without replacing this first write-if-absent behavior.
- Future prompt generation can choose to read the brief in a separate OpenSpec change without coupling this change to prompt injection.

## Documentation Plan

The implementation change should update public documentation after code exists:

- `README.md`: add `specharbor brief` to implemented commands only after implementation.
- `docs/usage.md`: document command syntax, TTY requirement, questions, confirmation, cancellation, existing-file behavior, and output file.
- `docs/generation-modes.md` or a more appropriate docs page: mention that briefing is context collection, not generation, prompt injection, RAG, or agent execution.
- `docs/workflow.md`: mention project briefing as an optional preparation step only if the implemented command is available.

This spec-author task does not modify documentation outside this OpenSpec change.

## Testing Strategy

Add focused tests for:

- project brief domain construction;
- source category validation for user-provided answers, detected context, and assumptions;
- deterministic Markdown rendering;
- Markdown output includes all required headings;
- Markdown output separates user-provided answers, detected context, and assumptions;
- Markdown output never treats assumptions as confirmed facts;
- use case writes `.specharbor/project-brief.md` after valid confirmed input;
- use case creates `.specharbor/` when absent;
- use case refuses to overwrite an existing `.specharbor/project-brief.md`;
- CLI rejects flags and positional arguments;
- CLI requires a TTY;
- CLI prompt flow succeeds with default options;
- CLI prompt flow succeeds with custom answers;
- CLI retries invalid menu choices and invalid custom answers;
- CLI confirmation denial writes no file;
- CLI EOF cancellation writes no file;
- CLI confirmation retry exhaustion writes no file;
- detected scan hints are presented as suggestions only;
- generated `.specharbor/project-brief.md` is deterministic;
- existing CLI commands still work.

Verification after implementation:

```text
gofmt
go test ./...
go run ./cmd/specharbor validate implement-project-briefing-foundation
```
