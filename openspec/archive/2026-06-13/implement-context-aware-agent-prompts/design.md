# Design: Implement Context-Aware Agent Prompts

## Overview

This change makes role prompt rendering consume project context that SpecHarbor already knows how to model:

- confirmed context from `.specharbor/project-brief.md`;
- detected facts from local context discovery;
- suggested assumptions from local context discovery.

The prompt output remains deterministic Markdown printed to stdout. SpecHarbor still does not call provider APIs, execute agents, run project commands, perform remote discovery, index repositories, or apply changes based on context.

The feature should preserve the existing Hexagonal Architecture direction:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

The CLI layer should parse `prompt` arguments, invoke the prompt use case, and print output only. Core/usecase should orchestrate context-aware prompt generation. Domain should own prompt-context models or reuse the existing context discovery models where appropriate. Concrete filesystem reads must stay behind adapter and port boundaries.

## Command Contract

The command remains:

```text
specharbor prompt <change-id> --role <role>
```

No new flags are required for the first version. Context-aware rendering should be the default behavior when context sources are present.

The command must continue to reject:

```text
specharbor prompt
specharbor prompt <change-id>
specharbor prompt <change-id> --role
specharbor prompt <change-id> --role unknown
specharbor prompt <change-id> --target codex
specharbor prompt <change-id> --role implementer extra
```

On success, stdout still contains only the rendered prompt. Do not add banners, debug output, absolute local paths, discovery reports, or command execution results around the prompt.

## Supported Roles

The current supported prompt roles remain the implementation target:

```text
spec-author
architecture-reviewer
implementer
test-engineer
change-reviewer
```

The first five recommended workflow steps keep their existing order:

```text
spec-author -> architecture-reviewer -> implementer -> test-engineer -> change-reviewer
```

Commit, Pull Request, Merge, and Archive remain workflow steps as currently defined. Pull Request is manual and unsupported by `specharbor prompt`; Archive is a supported command but not a role prompt. This change should not add PR Agent or Archive Housekeeping Agent roles. If another accepted change adds those prompt roles before implementation, context policy should be:

- PR Agent: minimal context only, focused on confirmed project identity, change summary, safety boundaries, and manual source-control constraints.
- Archive Housekeeping Agent: minimal context only, focused on confirmed project identity, active change path, archive safety, and no command execution beyond explicit user-requested archive behavior.

Full Project Context is required for Spec Author, Architecture Reviewer, Implementer, Test Engineer, and Change Reviewer role prompts.

## Project Context Section

When project context is available, generated prompts should include:

```text
## Project Context

Use the context below as guidance, but do not treat assumptions as facts.

### User-confirmed context

- Project type: CLI/tooling project
- Stack: Go
- Architecture: Hexagonal Architecture
- Test command: go test ./...

### Detected facts

- Language: Go
  Source: go.mod
  Confidence: high

- Agent rules: .specharbor/rules/
  Source: .specharbor/rules/
  Confidence: high

### Suggested assumptions

- Test command may be `go test ./...`
  Source: go.mod (go.mod convention)
  Confidence: medium

Rules:
- Prefer user-confirmed context over detected facts.
- Do not treat suggested assumptions as facts.
- Ask before making major architecture, persistence, or workflow decisions when context is missing or conflicting.
```

The exact wording can be adjusted, but the rendered structure must keep these concepts separate:

- `### User-confirmed context`
- `### Detected facts`
- `### Suggested assumptions`
- prompt-local safety rules for missing, ambiguous, or conflicting context

If there are no signals at all, the renderer may omit the populated Project Context section, but it must still include missing-context instructions telling the receiving agent not to invent project facts and to ask or label assumptions before large decisions.

## Context Source Rules

Prompt generation should consume classified context signals, not raw files.

Rules:

- `user_confirmed_context` must be preferred over `detected_fact`.
- Only known project brief sections parsed by the existing context discovery boundary may become `user_confirmed_context`; unknown or custom `.specharbor/project-brief.md` sections must not be rendered as confirmed context.
- `detected_fact` may be included only with source evidence and confidence.
- `suggested_assumption` may be included only under `Suggested assumptions`.
- `suggested_assumption` must never be rendered under `User-confirmed context` or `Detected facts`.
- `suggested_assumption` wording should use language such as `may be`, `assumption`, or `suggested`.
- Detected facts must not be silently converted into confirmed context.
- Suggested assumptions must not be silently converted into detected facts or confirmed context.
- If confirmed context conflicts with detected facts, prompts must prefer confirmed context and may include a concise conflict note.
- If context is missing or ambiguous, prompts must instruct agents to ask or explicitly label assumptions.
- Agents must be instructed not to invent stack, architecture, commands, persistence decisions, workflow decisions, or project direction.

## Conflict Handling

Conflict handling should be deterministic and conservative.

Recommended behavior:

- Group confirmed context first.
- For each signal kind, treat confirmed values as authoritative for prompt guidance.
- Keep conflicting detected facts visible only when useful for review, with a note such as:

```text
Conflict note: confirmed Stack is Go from .specharbor/project-brief.md; detected Stack includes Node.js from package.json. Prefer the confirmed value unless the user updates the brief.
```

- Do not ask the renderer to decide which repository file is correct beyond the precedence rule.
- Do not modify `.specharbor/project-brief.md`.
- Do not update context discovery results.

## Read-First Behavior

Role prompts should continue to include the existing required read-first sources and should add context-related sources as paths, not raw contents.

Required read-first sources for relevant roles:

- `AGENTS.md`
- `.specharbor/rules/global.md`
- the role-specific rule file
- `README.md`
- `docs/`
- `openspec/project.md`
- `openspec/specs/`
- `openspec/changes/<change-id>/`

When `.specharbor/project-brief.md` exists or confirmed context is rendered, include:

- `.specharbor/project-brief.md`

Discovered source evidence may appear in the Project Context section as relative paths and short evidence labels. The prompt must not dump raw README, docs, OpenSpec, workflow, manifest, or project brief file contents.

## Domain Model

Prefer reusing the existing context discovery domain types:

- `ContextDiscoveryResult`
- `ContextSignal`
- `ContextSignalKind`
- `ContextSignalClassification`
- `ContextConfidence`
- `ContextSource`

If prompt rendering needs a narrower view, add a prompt-specific domain model under `internal/core/domain`, such as:

- `PromptProjectContext`
- `PromptContextSection`
- `PromptContextItem`
- `PromptContextConflict`
- `PromptContextRenderPolicy`

The prompt-specific model should be derived from existing context signals, defensively copy slices, and validate:

- non-empty labels and values;
- safe relative source paths;
- supported classifications;
- source evidence for detected facts and assumptions;
- assumptions are rendered only in assumption sections;
- confirmed context source remains distinct from detected and assumed context.

Domain code must not import adapters, CLI packages, `os`, terminal IO, network APIs, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, concrete filesystem packages, or process execution packages.

## Architecture

Context-aware prompt generation should keep responsibilities aligned with the existing architecture contract:

- Domain owns context classifications, prompt context value objects, ordering rules, precedence rules, and size-limit policy.
- Ports define the small context and template dependencies consumed by prompt use cases.
- Use cases orchestrate role prompt rendering, context loading, context selection, conflict handling, and template data assembly.
- Adapters implement concrete filesystem access, template loading, template rendering, and CLI output.
- CLI code parses arguments, wires concrete adapters, calls the use case, and prints the rendered prompt.

Core packages must not import adapters or CLI packages. Adapters may depend inward on core ports and domain models. The implementation should add only the abstractions needed for this behavior and should avoid placeholder provider, RAG, workflow connector, or agent-runner interfaces.

## Ports And Use Cases

The prompt use case should continue to own prompt orchestration. It may be extended to depend on a small context provider boundary.

Recommended contracts:

```text
type PromptContextProvider interface {
    DiscoverPromptContext(projectRoot string) (domain.ContextDiscoveryResult, error)
}
```

or an equivalent small core-owned port/use-case boundary.

Implementation options:

- wrap the existing context discovery use case behind a prompt context provider;
- extract shared context collection into a small core service consumed by both `context discover` and `prompt`;
- inject the existing context discovery filesystem port into prompt orchestration without duplicating detector rules.

The implementation must not duplicate local discovery rules in CLI formatting or role templates. It must not create a dependency cycle where context discovery depends on prompt generation.

The render prompt use case should:

- validate project root, change id, and role as today;
- load the role template through the existing prompt template repository;
- obtain classified context through the context boundary;
- derive a role-aware prompt context model;
- render the prompt with `change_id`, `task`, and the Project Context block or missing-context instructions;
- return structured errors;
- return only the rendered prompt in the result;
- never print output;
- never call `os` directly;
- never execute commands or call external services.

If context discovery fails due to a recoverable missing-source condition, prompt generation should still render the role prompt with missing-context instructions. Filesystem safety errors should be returned when they prevent reliable, safe reads.

## Template Rendering

Templates may keep using embedded Markdown templates, but they need a deterministic placeholder for context, for example:

```text
{{project_context}}
```

or a typed rendering approach that inserts the same block in a controlled location.

The Project Context block should appear after `Read first` and before `Task`, unless a role-specific template has a clearer existing section boundary. This keeps context visible without replacing the active change as the source of scoped work.

Rendering rules:

- stable section order;
- stable item ordering;
- stable labels from domain signal kinds;
- relative source paths only;
- source evidence and confidence shown for detected facts and assumptions;
- no raw file contents;
- no secrets or skipped files;
- no external command output;
- no absolute local paths;
- no nondeterministic timestamps;
- no terminal-width-dependent wrapping.

## Prompt Size Limits

Context rendering must be bounded. Exact constants can be chosen during implementation, but the design should include:

- a maximum number of user-confirmed context items;
- a maximum number of detected fact items;
- a maximum number of suggested assumption items;
- a maximum character length for individual values and evidence labels;
- a maximum total Project Context character budget;
- deterministic truncation notices when items are omitted.

Suggested first version:

- render all supported confirmed brief fields when present;
- render only the most relevant detected facts for role work, prioritizing stack, language, framework, architecture, commands, documentation sources, agent instruction sources, and OpenSpec sources;
- render only a small number of assumptions, prioritizing commands and missing-context behavior;
- include a notice such as `Additional context signals omitted by prompt size limit.` when truncation occurs.

Truncation must not change classification. An omitted assumption must not reappear as a fact in a summary.

## Role-Aware Context Selection

All five supported roles should receive the same classification boundaries but may prioritize signals differently:

- Spec Author: project type, purpose, target users, stack, architecture, agent behavior, documentation, OpenSpec sources.
- Architecture Reviewer: architecture, stack, OpenSpec sources, agent rules, conflicting detected facts.
- Implementer: stack, architecture, install/test/build/run commands, agent behavior, active change sources.
- Test Engineer: stack, test/build commands, package manager, workflow signals, detected assumptions about tests.
- Change Reviewer: confirmed context, architecture, commands, OpenSpec sources, agent rules, conflicts, and assumptions that could affect review claims.

Role-specific selection should be deterministic and should not remove required classification sections when selected items exist.

## Documentation Plan

The implementation change should update public docs after behavior exists:

- `README.md`: update current command description so `prompt` is described as context-aware.
- `docs/usage.md`: document Project Context output, precedence rules, classifications, assumptions, conflict behavior, missing-context instructions, and safety boundaries.
- `docs/workflow.md`: update the workflow prompt step description to mention optional brief/discovery context and that prompt generation still does not execute commands or agents.
- `docs/agent-roles.md`: document that role prompts may include Project Context while role responsibilities remain unchanged.

Documentation must be clear that this feature is not RAG, indexing, remote discovery, provider integration, prompt execution, agent execution, command execution, or source-control automation.

## Testing Strategy

Add focused tests for:

- prompt context domain/model construction;
- deterministic Project Context rendering;
- inclusion of `.specharbor/project-brief.md` confirmed context;
- confirmed context precedence over detected facts;
- detected facts rendered with source evidence and confidence;
- unknown project brief sections not rendered as confirmed context;
- suggested assumptions rendered only under `Suggested assumptions`;
- assumptions never rendered as facts;
- conflict handling between confirmed context and detected facts;
- missing context instructions;
- prompt size limits and deterministic truncation notices;
- each supported role prompt receiving context when available;
- role-specific read-first lists including `.specharbor/project-brief.md` when relevant;
- regression coverage for existing `specharbor prompt` output structure;
- regression coverage for existing final decision labels, if present in templates;
- regression coverage for existing agent workflow order;
- regression coverage for `specharbor brief`;
- regression coverage for `specharbor context discover`;
- regression coverage for `specharbor scan`;
- architecture boundary tests proving core does not import adapters and CLI does not own context rules;
- OpenSpec validation.

Implementation verification should run:

```text
gofmt
go test ./...
go run ./cmd/specharbor validate implement-context-aware-agent-prompts
```

## Safety Boundaries

Prompt generation must not:

- run package managers, tests, builds, run commands, scripts, shells, agents, or workflow tools;
- call provider APIs, local model APIs, network APIs, source-control APIs, GitHub, GitLab, Bitbucket, CI, or remote services;
- perform repository-wide indexing, embeddings, vector database operations, RAG, retrieval, or snippet ranking;
- modify `.specharbor/project-brief.md`;
- modify context discovery output;
- write prompt files;
- modify production source files;
- update task checkboxes;
- commit, push, open PRs, merge, tag, release, publish, or archive automatically.
