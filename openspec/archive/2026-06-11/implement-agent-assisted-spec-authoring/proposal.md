# Proposal: Implement Agent-Assisted Spec Authoring

## Summary

Add the first foundation for agent-assisted OpenSpec spec authoring:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

The first version is dry-run-only. SpecHarbor validates the request, builds a deterministic and copy-pasteable authoring prompt, prints that prompt to stdout with a dry-run/no-files-written/no-agent-executed report, and exits without running external commands or writing files.

The mode is limited to drafting or refining the OpenSpec change package:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

It must not implement production code, call provider APIs, require provider API keys, automate source control, automate workflows, write prompt files, parse agent output, apply agent output, or modify files.

## Problem

SpecHarbor supports controlled OpenSpec workflows and currently provides deterministic blank, built-in template, and guided generation. Those modes are local and useful, but they do not help users who already have a coding-agent tool available and want that tool to draft a richer OpenSpec package before implementation begins.

Many teams use Codex, Claude Code, Cursor, Devin, Windsurf, GitHub Copilot, Gemini CLI, Roo Code, Aider, or similar coding-agent tools, but they may not have direct provider API keys available to SpecHarbor. SpecHarbor needs a safe foundation that prepares an authoring prompt for those tools without executing them, without writing files, and without turning the agent loose on production code.

## Goal

Implement agent-assisted spec authoring as an additive generation mode for:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type bugfix --title "<title>" --summary "<summary>"
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type docs --title "<title>" --summary "<summary>"
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type refactor --title "<title>" --summary "<summary>"
```

For each dry-run command, SpecHarbor must:

- validate the change id, agent, type, title, summary, generation mode conflicts, unsupported flags, and extra positional arguments;
- support exactly `feature`, `bugfix`, `docs`, and `refactor` as agent-assisted authoring types;
- build a deterministic OpenSpec authoring prompt;
- include project context, change id, type, title, summary, required files, architecture boundaries, scope boundaries, and output expectations in the prompt;
- explicitly instruct the agent to create or refine only OpenSpec change files;
- explicitly instruct the agent not to implement production code or modify unrelated files;
- explicitly instruct the agent to leave implementation tasks unchecked;
- print the deterministic, copy-pasteable authoring prompt to stdout;
- print a deterministic dry-run report that states no files were written and no agent or external command was executed;
- write no files, including no prompt file;
- never parse, apply, or write agent output, including OpenSpec files;
- reject `--execute` as an explicitly unsupported flag/mode with a clear error.

`--execute`, local agent command execution, command output capture, file application, confirmation flows, post-write validation, and source-control/workflow automation are deferred to future changes.

## Scope

- Add `--agent-assisted` as a generation mode flag for `specharbor generate`.
- Require `--agent <agent-name>`, `--type <type>`, `--title "<title>"`, and `--summary "<summary>"` when `--agent-assisted` is used.
- Keep this first version dry-run-only.
- Reject `--execute` as unsupported with a clear error.
- Support exactly these authoring types:
  - `feature`
  - `bugfix`
  - `docs`
  - `refactor`
- Reuse the existing supported guided type validation concepts where possible.
- Reuse `domain.RequiredOpenSpecChangeFiles()` for the required OpenSpec file list.
- Keep agent-assisted domain concepts in `internal/core/domain`.
- Keep orchestration in `internal/core/usecase`.
- Keep prompt rendering behind a core-owned port/interface.
- Keep deterministic prompt template content in `internal/adapters/templates` or another justified adapter package.
- Keep CLI parsing and report formatting in `internal/adapters/cli`.
- Do not add `AgentRunner`, local command execution adapters, workflow connectors, write/apply ports, or confirmation flows in this change.
- Add focused tests for domain validation, prompt rendering, use case orchestration, unsupported `--execute` handling, CLI parsing/reporting, dry-run safety, mode conflicts, README/docs prompt boundaries, copy-pasteable prompt output, and existing generation mode regressions.

## Out of Scope

- Implementing production code.
- Allowing the agent to modify production code.
- Applying patches automatically.
- Writing OpenSpec files in dry-run.
- Writing a prompt file.
- Parsing, applying, or writing any agent output, including OpenSpec files.
- Adding write/apply ports.
- Adding confirmation flows.
- Running local agent commands.
- Adding `AgentRunner` or local command execution adapters.
- Supporting `--execute` beyond returning an unsupported error.
- Direct OpenAI, Anthropic, Google, GitHub Copilot, Devin, Cursor, or other provider APIs.
- OAuth.
- Credential management.
- Secret storage.
- Cloud execution.
- Remote execution.
- Marketplace behavior.
- Source-control automation.
- Workflow automation.
- Auto-commit.
- Auto-push.
- Auto-merge.
- Running tests automatically as part of dry-run.
- Interactive wizard behavior.
- Streaming provider output.
- Custom remote registries.
- Changing `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `generate --blank`, `generate --template`, or `generate --guided`.
- Updating README or docs outside this OpenSpec change.
- Changing CI.
- Modifying `.github/workflows/ci.yml`.
- Modifying `.specharbor/config.yml`.

## Success Criteria

- Dry-run is the default for `--agent-assisted`.
- Dry-run validates inputs and prints a deterministic authoring plan and copy-pasteable prompt to stdout.
- Dry-run does not execute local commands, external agents, provider APIs, local models, network APIs, source-control APIs, or workflow tools.
- Dry-run does not write OpenSpec files, production files, docs, README files, CI files, config files, or prompt files.
- Dry-run never parses, applies, or writes agent output, including OpenSpec files.
- Supported types are exactly `feature`, `bugfix`, `docs`, and `refactor`.
- Missing agent, type, title, or summary values return clear errors.
- Unknown type, unsafe change id, mode conflicts, unsupported flags, and extra positional arguments return clear errors.
- The generated prompt includes the change id, type, title, summary, required OpenSpec files, project context, architecture boundaries, scope boundaries, and explicit "do not implement code" instructions.
- For non-docs types, the generated prompt prohibits README and docs changes.
- For `docs`, the generated prompt permits documentation scope from the title and summary while still forbidding production code changes.
- `--execute` is deferred to a future change and returns a clear unsupported flag/mode error in this first version.
- No `AgentRunner`, local command adapter, write/apply port, confirmation flow, or execute use case path is introduced.
- Existing blank, template, and guided generation behavior remains unchanged.
- No README, docs outside this OpenSpec change, CI, `.github/workflows/ci.yml`, or `.specharbor/config.yml` changes are included.
- `go test ./...` succeeds after implementation.
