# Proposal: Implement Agent Runner Foundation

## Summary

Add the first safe `--execute` foundation for agent-assisted OpenSpec spec authoring:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>" --execute
```

Dry-run remains the default when `--execute` is absent. For valid recognized agent targets, dry-run output formatting and deterministic prompt rendering must remain backward-compatible. Unknown agent targets are now rejected even in dry-run mode; this is an intentional validation tightening, not a formatting change for valid agents.

The concrete supported AI agent targets are both recognized and executable local runner targets:

```text
codex     -> Codex
claude    -> Claude Code
devin     -> Devin
cursor    -> Cursor
copilot   -> GitHub Copilot
gemini    -> Gemini CLI
roo       -> Roo Code
windsurf  -> Windsurf
aider     -> Aider
```

`generic` remains supported as a recognized dry-run target only. It has no deterministic local command mapping in this change because generic agents require config-driven command mapping, which belongs to a future config-driven runner/templates feature.

With `--execute`, SpecHarbor must render the same deterministic OpenSpec authoring prompt used by dry-run mode, resolve the selected concrete agent target to an explicitly supported local executable mapping, send the prompt to that already-resolved local command through a core-owned `AgentRunner` port, capture stdout, stderr, exit code, and execution status for started processes, and print the runner result.

This is a run-and-report foundation only. SpecHarbor must not parse agent output, apply agent output, write OpenSpec files from agent output, modify production code, run source-control commands, run workflow tools, auto-commit, auto-push, or auto-merge.

## Problem

Agent-assisted spec authoring currently stops at deterministic dry-run output. The command validates inputs, prints an authoring plan, and prints a copy-pasteable prompt, but `--execute` is explicitly unsupported.

SpecHarbor's agent-assisted product positioning must support OpenSpec/SDD workflows for users working with Codex, Claude Code, Devin, Cursor, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, Aider, and generic agents. Users who already have local agent commands installed need a narrow way to hand the existing deterministic OpenSpec authoring prompt to a local agent command without adding provider APIs, IDE automation, remote execution, file application behavior, or source-control automation.

The first implementation must preserve SpecHarbor's safety boundary: this feature helps author an OpenSpec change package, not production implementation.

## Goal

Support explicit local runner execution for:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type feature --title "<title>" --summary "<summary>" --execute
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type bugfix --title "<title>" --summary "<summary>" --execute
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type docs --title "<title>" --summary "<summary>" --execute
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type refactor --title "<title>" --summary "<summary>" --execute
```

For each supported execution command, SpecHarbor must:

- validate the same inputs as existing agent-assisted dry-run behavior;
- support exactly `feature`, `bugfix`, `docs`, and `refactor` as authoring types;
- validate the selected agent as a recognized agent target;
- reject unknown agent targets with a clear validation error in dry-run and execute mode;
- preserve dry-run output formatting and deterministic prompt rendering for valid recognized targets;
- render the same deterministic OpenSpec authoring prompt as dry-run mode;
- resolve the selected concrete agent target to a built-in local command mapping only when `--execute` is present;
- reject `generic` in execute mode with a clear error because no deterministic local command mapping exists for it in this change;
- run the selected command through a core-owned `AgentRunner` port;
- implement concrete local process execution only in an adapter package;
- execute the command directly with optional fixed args, never through shell interpolation;
- pass the prompt predictably through stdin;
- run from the project root unless the design documents a different explicit working directory;
- capture stdout only when the runner process starts;
- capture stderr only when the runner process starts;
- capture the actual exit code when the runner process starts;
- capture a clear execution status for started processes;
- classify a zero exit code as `success`;
- classify a non-zero exit code as a completed runner execution with status `non_zero_exit`;
- print the full run-and-report output before making the SpecHarbor CLI process exit non-zero for non-zero runner exits;
- return clear startup errors when the command cannot start;
- treat startup failure as an execution error, not as a normal runner result;
- report no exit code for startup failures because the process never started;
- print command output as output only;
- never parse stdout or stderr as files;
- never write OpenSpec files from command output;
- never apply patches or modify production code from command output.

## Scope

- Preserve dry-run as the default when `--execute` is absent.
- Clarify that existing dry-run behavior/output remains unchanged for valid recognized agent targets and existing recognized-agent examples.
- Tighten dry-run validation so unknown agent targets are rejected with a clear error.
- Define concrete recognized and executable agent targets for deterministic prompt rendering, dry-run reports, and local runner execution: `codex`, `claude`, `devin`, `cursor`, `copilot`, `gemini`, `roo`, `windsurf`, and `aider`.
- Retain `generic` as a recognized dry-run-only target because it has no deterministic local command mapping in this change.
- Ensure dry-run never requires an `AgentRunner`, never resolves executable mappings, and never executes external commands.
- Change `--execute` from unsupported to supported only for agent-assisted spec authoring run-and-report mode.
- Add domain concepts for recognized agent target, executable local runner target, resolved executable command, optional fixed args, agent runner request, result, exit code, and execution status.
- Add a small core-owned `AgentRunner` port/interface under `internal/core/ports`.
- Add a concrete local command runner adapter under `internal/adapters`.
- Add built-in executable mappings:

```text
codex    -> codex
claude   -> claude
devin    -> devin
cursor   -> cursor
copilot  -> copilot
gemini   -> gemini
roo      -> roo
windsurf -> windsurf
aider    -> aider
```

- Model resolved executable targets as command name, optional fixed args, display name, and agent id so future mappings can add deterministic args without shell execution.
- Prefer stdin for passing the deterministic prompt to the runner.
- Keep agent-assisted orchestration in `internal/core/usecase`.
- Keep CLI parsing, dependency wiring, current working directory lookup, report formatting, and CLI exit behavior in `internal/adapters/cli`.
- Keep prompt rendering behind the existing core-owned prompt renderer port.
- Update architecture boundary tests that currently forbid any agent runner abstraction so they now allow the narrow runner port and adapter while continuing to forbid write/apply/workflow/source-control/provider behavior.
- Add focused tests for recognized targets, executable mappings, generic dry-run-only behavior, unknown targets, domain result status, runner port orchestration, local adapter behavior, CLI parsing/reporting, dry-run regression, non-zero exits, startup errors, safety boundaries, and unchanged blank/template/guided generation.
- Update public docs in the implementation PR because CLI behavior changes.

## Out of Scope

- OpenAI, Anthropic, Gemini, Ollama, local model, or other provider API integration.
- Provider API keys.
- OAuth.
- Credential management.
- Secret storage.
- Cloud execution.
- Remote execution.
- IDE automation.
- Marketplace behavior.
- Remote registry behavior.
- Provider, cloud, IDE, OAuth, credential, or marketplace integrations for Devin, Cursor, GitHub Copilot, Roo Code, Windsurf, generic agents, or any other agent.
- Source-control automation.
- Workflow automation.
- Parsing agent output into files.
- Applying patches.
- Writing OpenSpec files from agent output.
- Writing prompt files.
- Interactive confirmation flows for applying files.
- Auto-commit.
- Auto-push.
- Auto-merge.
- Modifying production code.
- Implementing AI-assisted generation.
- Adding config-driven generic runner commands.
- Changing `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `generate --blank`, `generate --template`, or `generate --guided`.
- Changing CI behavior.
- Modifying `.github/workflows/ci.yml`.
- Modifying `.specharbor/config.yml`.

## Success Criteria

- Dry-run remains the default and does not execute external commands.
- Dry-run never requires an `AgentRunner`.
- Dry-run never resolves executable command mappings.
- Recognized concrete agent targets include Codex, Claude Code, Devin, Cursor, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, and Aider.
- `generic` remains available only as a recognized dry-run target.
- Existing dry-run agent-assisted output remains unchanged for valid recognized agent targets when `--execute` is absent.
- Unknown agent targets are rejected with a clear validation error in dry-run and execute mode.
- `--execute` is required for runner execution.
- `--execute` supports all mapped concrete local runner targets.
- `--execute` rejects `generic` with a clear unmapped-target error.
- CLI output clearly distinguishes the recognized agent target from the resolved executable command.
- `--execute` uses only the core-owned `AgentRunner` port from the use case boundary.
- Concrete command execution lives only in an adapter.
- The local runner captures stdout, stderr, exit code, and execution status only for started processes.
- Startup failures return clear errors, produce no normal runner result, and report no exit code.
- Non-zero runner exits produce a completed runner result with status `non_zero_exit`.
- Non-zero runner exits report stdout, stderr, actual exit code, and status without panic.
- The CLI prints the full run-and-report output before exiting non-zero for non-zero runner exits.
- Agent output is displayed as output only.
- Agent output is not parsed, applied, written to OpenSpec files, or used to modify production code.
- No provider APIs, credentials, source-control automation, workflow automation, auto-commit, auto-push, or auto-merge are introduced.
- Existing blank, template, and guided generation behavior remains unchanged.
- README and docs are updated in the implementation PR to describe the recognized agent ecosystem, executable local runner mappings, generic dry-run-only behavior, run-and-report behavior, non-zero exit behavior, startup-failure behavior, and safety boundaries.
