# Design: Agent-Aware Prompt Generation

## Overview

Prompt generation will accept two separate inputs:

- workflow role: what responsibility the receiving session should perform in
  the SpecHarbor workflow;
- target agent: which external coding assistant style the printed prompt should
  be adapted for.

SpecHarbor still prints deterministic prompt text to stdout. It does not run
agents, call provider APIs, configure credentials, use the network, or automate
source control.

## Architecture

This change follows the existing hexagonal boundaries:

- `internal/core/domain` owns the prompt target agent concept and validation.
- `internal/core/ports` owns the rendering contract consumed by the prompt use
  case.
- `internal/core/usecase` orchestrates role and agent validation, context-aware
  prompt data, and rendering through the port.
- `internal/adapters/templates` owns concrete prompt text adaptations.
- `internal/adapters/cli` owns command-line parsing, duplicate flag detection,
  missing value detection, unsupported flag reporting, and user-facing command
  errors.

No core package should import adapters or perform filesystem, terminal,
provider, network, workflow, source-control, external-agent, or process
execution behavior for this feature.

## Command Shape

Supported forms:

```bash
go run ./cmd/specharbor prompt <change-id> --role <role>
go run ./cmd/specharbor prompt <change-id> --role <role> --agent <agent>
```

Examples:

```bash
go run ./cmd/specharbor prompt implement-config-foundation --role implementer --agent codex
go run ./cmd/specharbor prompt implement-config-foundation --role architecture-reviewer --agent claude-code
go run ./cmd/specharbor prompt implement-config-foundation --role change-reviewer --agent devin
```

`--agent` is optional. When omitted, the effective agent is `generic`.

## Terminology

### Role

Roles are workflow responsibilities. The supported prompt roles remain:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

The existing role prompt bodies and responsibilities remain the source of truth
for what the receiving agent should do.

### Agent

Agents are target external tools that consume the printed prompt. The supported
prompt target agents are exactly:

- `generic`
- `codex`
- `claude-code`
- `devin`
- `cursor`
- `copilot`
- `gemini`
- `roo`
- `windsurf`
- `aider`

The prompt command should not reuse or change `generate --agent-assisted`
execution mappings. This change is scoped to `prompt` only.

## Domain Concepts

Add or extend core domain prompt concepts under `internal/core/domain` so the
selected target agent is represented as a validated value, not a free-form
string passed through adapters.

The domain concept should provide:

- stable ids for the supported target agents;
- validation for unknown target agents;
- a deterministic supported-agent list for error messages and docs tests;
- a default value of `generic`;
- a display label where useful for prompt output.

The domain layer must remain pure. It must not import adapters, `os`, terminal
IO, provider SDKs, network APIs, workflow SDKs, source-control SDKs,
external-agent SDKs, or process execution packages.

## Ports and Rendering

Prompt rendering must stay behind a core-owned port/interface in
`internal/core/ports`. The use case should pass a structured prompt render
request containing at least:

- change id;
- selected workflow role;
- selected target agent;
- role-specific prompt data;
- bounded project context when available.

The concrete renderer in `internal/adapters/templates` decides how to combine
the role prompt body with a small target-agent header or guidance block.

Core use cases must not import the template adapter or know concrete template
details.

## Agent-Aware Template Adaptation

The concrete template adaptation should be intentionally lightweight. It should
not fork the entire role prompt for every supported agent. Prefer a common role
prompt body plus a small deterministic header such as `## Target Agent` or
`## External Tool Guidance`.

The generated prompt must include:

- selected workflow role;
- selected target agent;
- instruction that SpecHarbor generated prompt text for the user to paste or
  use in the selected external tool;
- a clear non-execution boundary stating that SpecHarbor does not execute the
  target agent.

Suggested guidance per target:

- `generic`: neutral prompt suitable for any coding assistant.
- `codex`: mention local repository execution, CLI commands, and explicit
  verification steps.
- `claude-code`: emphasize careful file edits, reasoning over project context,
  and validation commands.
- `devin`: emphasize autonomous task boundaries, PR readiness, and reporting.
- `cursor`: emphasize editor-assisted implementation and file-scoped changes.
- `copilot`: emphasize applying the prompt inside the coding environment.
- `gemini`: emphasize repository context and command verification.
- `roo`: emphasize role-based task execution.
- `windsurf`: emphasize IDE agent workflow.
- `aider`: emphasize patch-oriented changes and explicit file scope.

These adaptations must not include credentials, provider setup, network
assumptions, API keys, SDK instructions, marketplace behavior, or external
agent command invocations.

## Prompt Use Case

Prompt orchestration remains in `internal/core/usecase`.

The use case should:

- accept a selected role and selected agent;
- default the agent to `generic` when omitted by the CLI request model;
- validate the role using the existing role rules;
- validate the agent using the new domain concept;
- preserve existing context-aware role prompt behavior;
- call the prompt rendering port with the structured role and agent request;
- return structured output or errors without CLI-specific formatting.

The use case must not call `os`, terminal IO, provider APIs, network APIs,
agent CLIs, workflow tools, source-control tools, or process execution
packages.

## CLI Parsing and Errors

CLI parsing remains in `internal/adapters/cli`.

The `prompt` command should accept:

```text
prompt <change-id> --role <role> [--agent <agent>]
```

Parsing requirements:

- `--agent` may appear at most once.
- `--agent` requires a value.
- Unknown agent values fail clearly.
- Unsupported flags fail clearly.
- Extra positional arguments fail clearly.
- Existing `--role` parsing and errors stay compatible.
- Missing role behavior stays compatible with the current command contract.

Error messages should include the bad value where safe and list supported
agents for unknown agent errors. They must not include credentials, environment
values, or local absolute paths beyond existing command conventions.

## Documentation

The implementation PR must update public CLI documentation because this changes
user-facing command behavior.

Required docs updates:

- `docs/usage.md`: command syntax, examples, supported agents, default
  `generic`, errors, and safety boundaries.
- `docs/agent-roles.md`: explain role vs agent and clarify that prompt output
  is pasted into external tools.
- `README.md`: update command lists or quickstart examples if needed.
- `docs/workflow.md`: update workflow command examples if the prompt examples
  or explanation need to mention agent targeting.

Docs must explain:

- role and agent are different concepts;
- `--role` controls workflow responsibility;
- `--agent` controls target external tool style;
- `--agent` does not execute the tool;
- `--agent` only adapts generated prompt text;
- supported agents are listed clearly;
- default agent is `generic`.

## Testing Strategy

Add focused tests at the smallest useful layer:

- domain tests for supported agent ids, default `generic`, display labels, and
  unknown agent rejection;
- use case tests proving omitted agent defaults to `generic`, selected agents
  are passed to rendering, unknown agents fail, and existing role behavior is
  preserved;
- template adapter tests proving generated output includes role, selected
  agent, paste/use guidance, the non-execution boundary, and deterministic
  lightweight guidance;
- CLI parsing tests for valid `--agent`, omitted `--agent`, unknown agents,
  missing values, duplicates, unsupported flags, and extra args;
- regression tests for all existing prompt roles;
- documentation snippet tests where the repository already uses docs
  guardrails.

Run `gofmt` on changed Go files and `go test ./...`.

## Decisions and Tradeoffs

- **Separate role and agent.** Role remains the workflow responsibility; agent
  is only target prompt style. This avoids overloading role ids and keeps the
  existing workflow model intact.
- **Default to `generic`.** Backward compatibility is preserved without making
  users choose a target tool for existing commands.
- **Lightweight headers instead of per-agent full templates.** A shared role
  body avoids duplicated instruction drift while still allowing target-specific
  prompt style.
- **No aliases in the first version.** Supported prompt agent ids are exactly
  documented. For example, `claude-code` is the prompt target id in this
  change; alternate spellings can be considered separately if needed.
- **No execution behavior.** This is prompt text adaptation only. Runner
  behavior, provider APIs, credentials, network access, and workflow automation
  remain out of scope.
