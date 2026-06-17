# Acceptance Criteria: Agent-Aware Prompt Generation

## Backward Compatibility

- Existing `prompt <change-id> --role <role>` behavior remains supported.
- Existing role prompt responsibilities for `spec-author`,
  `architecture-reviewer`, `implementer`, `test-engineer`, and
  `change-reviewer` remain intact.
- All existing prompt role tests continue to pass.
- Omitting `--agent` defaults to the prompt target agent `generic`.

## Supported Roles and Agents

- Supported workflow roles are `spec-author`, `architecture-reviewer`,
  `implementer`, `test-engineer`, and `change-reviewer`.
- Supported prompt target agents are exactly `generic`, `codex`,
  `claude-code`, `devin`, `cursor`, `copilot`, `gemini`, `roo`, `windsurf`,
  and `aider`.
- The supported prompt target agents are documented clearly in user-facing docs.
- The `generic` target produces a neutral prompt suitable for any coding
  assistant.

## CLI Behavior

- `go run ./cmd/specharbor prompt <change-id> --role <role> --agent <agent>`
  accepts each supported prompt target agent.
- `--agent` is optional.
- An unknown agent returns a clear error and lists supported prompt target
  agents.
- `--agent` without a value returns a clear error.
- Duplicate `--agent` returns a clear error.
- Unsupported flags return clear errors.
- Extra positional arguments return clear errors.
- Existing `--role` validation and error behavior remain compatible.

## Generated Prompt Output

- Generated prompt output includes the selected workflow role.
- Generated prompt output includes the selected prompt target agent.
- Generated prompt output is deterministic and copy-pasteable.
- Generated prompt output makes clear the user should paste or use the prompt in
  the selected external tool.
- Generated prompt output does not claim that SpecHarbor executes the selected
  agent.
- Generated prompt output does not claim that SpecHarbor configures,
  authenticates, starts, supervises, or automates the selected agent.
- Agent-specific guidance remains lightweight and limited to prompt style.
- Prompt output does not include credentials, provider setup, network
  assumptions, API key instructions, SDK instructions, or external-agent command
  invocations.

## Architecture and Safety

- CLI parsing remains in `internal/adapters/cli`.
- Prompt orchestration remains in `internal/core/usecase`.
- Prompt domain concepts belong in `internal/core/domain`.
- Prompt template rendering remains behind a core-owned port/interface.
- Concrete prompt template/adaptation behavior belongs in
  `internal/adapters/templates`.
- Core packages do not import adapters, `os`, terminal IO, provider SDKs,
  network APIs, workflow SDKs, source-control SDKs, external-agent SDKs, or
  process execution packages.
- No provider APIs are introduced.
- No network access is introduced.
- No external process execution is introduced.
- No external agent execution is introduced.
- No credentials, API key management, or provider setup behavior is introduced.
- No source-control automation, workflow automation, auto-commit, auto-push, or
  auto-merge behavior is introduced.
- `generate`, `validate`, `review`, `archive`, `config`, CI,
  `.github/workflows/ci.yml`, and `.specharbor/config.yml` are not modified by
  the implementation.

## Documentation

- `docs/usage.md` documents the new prompt syntax and examples.
- `docs/agent-roles.md` explains that role and agent are different concepts.
- `README.md` is updated if command lists, quickstart examples, or capability
  summaries need to mention agent-aware prompt generation.
- `docs/workflow.md` is updated if workflow examples or explanations need to
  mention agent targeting.
- Documentation states that `--role` controls workflow responsibility.
- Documentation states that `--agent` controls target external tool style.
- Documentation states that `--agent` does not execute the target tool.
- Documentation states that `--agent` only adapts generated prompt text.
- Documentation states that the default prompt target agent is `generic`.
- Documentation lists supported prompt target agents exactly as `generic`,
  `codex`, `claude-code`, `devin`, `cursor`, `copilot`, `gemini`, `roo`,
  `windsurf`, and `aider`.

## Verification

- `go test ./...` passes.
- `go run ./cmd/specharbor validate implement-agent-aware-prompt-generation`
  reports the change package as valid.
- Manual or automated prompt checks verify omitted-agent defaulting, explicit
  agent selection, and all specified error cases.
