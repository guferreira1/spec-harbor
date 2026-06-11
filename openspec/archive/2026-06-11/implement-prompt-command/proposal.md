# Proposal: Implement Prompt Command

## Problem

`specharbor prompt` is currently a placeholder. SpecHarbor already has reusable role templates under `agent-prompts/roles/`, but users cannot render one of those prompts for a specific OpenSpec change from the CLI.

This leaves agent-assisted workflows manual even though the project already defines the supported roles and template placeholders.

## Goal

Implement `specharbor prompt` so it renders an agent role prompt for a given OpenSpec change and prints the result to stdout.

The primary command shape is:

```text
specharbor prompt <change-id> --role <role>
```

The command must support at least these roles:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

## Scope

- Replace the placeholder `prompt` command with real prompt rendering behavior.
- Accept one change id argument.
- Accept a required `--role` argument.
- Load role templates from `agent-prompts/roles/`.
- Render template placeholders such as `{{change_id}}`.
- Provide deterministic stdout output containing only the rendered prompt on success.
- Return clear errors for missing change id, missing role, unsupported role, missing template, and rendering failures.
- Keep prompt orchestration in a core use case.
- Keep template loading and rendering details in adapters.
- Add focused tests for use case, template adapter, and CLI behavior.

## Out of Scope

- Calling AI providers or local model APIs.
- Requiring provider API keys.
- Dispatching prompts to external coding agents.
- Integrating with Codex, Claude Code, Devin, Cursor, GitHub Copilot, Windsurf, Aider, Roo Code, or workflow tools.
- Creating or modifying OpenSpec change files.
- Validating full OpenSpec change content.
- Adding interactive prompts.
- Adding role template editing commands.
- Adding custom template directories or configuration-driven template lookup.
- Adding prompt output formats other than stdout Markdown.
- Implementing `scan`, `generate`, `validate`, `review`, `archive`, or `config` behavior.

## Success Criteria

- Running `specharbor prompt implement-prompt-command --role implementer` prints the rendered implementer prompt to stdout.
- The rendered output includes the requested change id wherever the template uses `{{change_id}}`.
- All five required roles can be rendered.
- The command does not call any AI provider or external agent.
- Prompt behavior follows Hexagonal Architecture: CLI parses and formats, use case orchestrates, adapters load and render templates.
- `go test ./...` succeeds.
