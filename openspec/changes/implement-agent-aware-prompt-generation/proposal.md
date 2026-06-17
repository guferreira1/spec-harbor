# Proposal: Agent-Aware Prompt Generation

## Summary

Add optional agent-aware prompt generation to the existing `prompt` command by
separating the workflow role from the external coding assistant that will
receive the generated prompt.

The existing role-based flow remains the default:

```bash
go run ./cmd/specharbor prompt <change-id> --role <role>
```

The new form adds an optional target agent:

```bash
go run ./cmd/specharbor prompt <change-id> --role <role> --agent <agent>
```

When `--agent` is omitted, the target agent defaults to `generic`.

## Problem

SpecHarbor currently generates prompts by workflow role, such as:

```bash
go run ./cmd/specharbor prompt implement-config-foundation --role implementer
```

This correctly models the responsibility of the prompt recipient in the
OpenSpec workflow, but it does not let the user adapt the prompt text for the
external coding assistant they will paste it into. Users may use Codex, Claude
Code, Devin, Cursor, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, Aider, or a
generic assistant. These tools benefit from small differences in instruction
style, but that should not change the workflow role.

The missing distinction creates two risks:

- users may confuse a SpecHarbor workflow role with a target coding assistant;
- future prompt behavior could accidentally imply that SpecHarbor executes,
  configures, authenticates, or automates external agents.

## Goal

Support agent-aware prompt text while preserving the existing role-based prompt
workflow.

`role` means the workflow responsibility:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

`agent` means the external tool where the user will paste or use the prompt:

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

The generated prompt must remain deterministic, copy-pasteable, local/offline,
and clear that SpecHarbor prints prompt text only.

## Scope

- Add an optional `--agent <agent>` flag to `prompt <change-id>`.
- Preserve `prompt <change-id> --role <role>` behavior by defaulting omitted
  `--agent` to `generic`.
- Validate supported prompt target agents.
- Return clear command errors for unknown agents, missing `--agent` values,
  duplicate `--agent`, unsupported flags, and extra positional arguments.
- Keep role validation and role prompt behavior intact.
- Add a core domain concept for prompt target agents and keep supported values
  centralized and deterministic.
- Pass the selected agent through prompt orchestration in
  `internal/core/usecase`.
- Keep prompt template rendering behind a core-owned port/interface.
- Implement concrete agent-specific prompt text adaptations in
  `internal/adapters/templates`.
- Include a short agent-specific instruction/header where useful.
- Make generated prompts state both the selected workflow role and selected
  target agent.
- Make generated prompts clear that the user should paste or use the prompt in
  the selected external tool.
- Update public CLI documentation in the future implementation PR:
  `README.md` if command lists or examples need changing, `docs/usage.md`,
  `docs/agent-roles.md`, and `docs/workflow.md` if needed.
- Add focused tests for domain validation, use case behavior, template output,
  CLI parsing, backward compatibility, and documentation snippets.

## Out of Scope

This change must not introduce:

- execution of Codex, Claude Code, Devin, Cursor, GitHub Copilot, Gemini CLI,
  Roo Code, Windsurf, Aider, or any other external agent;
- an `AgentRunner`, runner registry, or command execution path for `prompt`;
- provider API integration;
- local model integration;
- credential management;
- API key instructions;
- network access;
- prompt execution;
- workflow automation;
- source-control automation;
- auto-commit, auto-push, auto-merge, or pull request automation;
- changes to `generate`, `validate`, `review`, `archive`, `config`, or CI;
- changes to `.github/workflows/ci.yml` or `.specharbor/config.yml`;
- modifications to generated OpenSpec change contents outside this change
  package;
- provider-specific SDK behavior or target-agent SDK integration.

## Success Criteria

- Existing `prompt <change-id> --role <role>` output remains compatible and
  continues to use the same role prompt behavior.
- `--agent` is optional and defaults to `generic`.
- `--agent <agent>` accepts only the documented supported target agents.
- Unknown, missing, duplicated, unsupported, or extra prompt arguments fail with
  clear errors before rendering.
- Generated prompt output includes the selected role and selected agent.
- Agent-specific guidance is lightweight and limited to prompt-text style.
- Prompt output does not claim or imply that SpecHarbor executes, configures, or
  authenticates the selected tool.
- No provider APIs, network access, credentials, source-control automation,
  workflow automation, or external process execution are introduced.
- README and docs are updated in the implementation PR to explain the role vs
  agent distinction, supported agents, default `generic` behavior, and
  non-execution boundary.
- Existing prompt role tests continue to pass, and new tests cover the agent
  flag and prompt text adaptations.
