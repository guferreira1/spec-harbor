# Risks: Agent-Aware Prompt Generation

## Role and agent confusion

- **Risk:** Users may confuse workflow roles with target coding assistants,
  especially because both are used when generating prompts.
- **Mitigation:** Keep command syntax explicit (`--role` and `--agent`), include
  selected role and agent in prompt output, and update docs to define each
  concept separately.

## Backward compatibility regression

- **Risk:** Adding `--agent` could change existing role prompt output or break
  users who already run `prompt <change-id> --role <role>`.
- **Mitigation:** Default omitted `--agent` to `generic`, keep the existing
  role prompt body intact, and add regression tests for existing role prompt
  behavior.

## Accidental implication of execution

- **Risk:** Agent-specific wording could imply that SpecHarbor starts or
  controls Codex, Claude Code, Devin, Cursor, GitHub Copilot, Gemini CLI, Roo
  Code, Windsurf, Aider, or another tool.
- **Mitigation:** Add explicit non-execution wording to generated prompts and
  docs. Keep this change limited to prompt text adaptation and tests.

## Scope creep into runners or provider APIs

- **Risk:** Target-agent support could expand into agent runners, provider API
  calls, local model integration, credentials, network behavior, or workflow
  automation.
- **Mitigation:** The proposal, design, tasks, and acceptance criteria all make
  execution, provider APIs, network access, credentials, source-control
  automation, workflow automation, auto-commit, auto-push, and auto-merge out
  of scope.

## Prompt template duplication

- **Risk:** Creating separate full templates per agent could duplicate role
  instructions and cause role behavior to drift over time.
- **Mitigation:** Use a shared role prompt body with a small deterministic
  target-agent header or guidance block.

## Unsupported or ambiguous agent identifiers

- **Risk:** Different tools have common aliases, such as `claude` for Claude
  Code, which could make command input ambiguous.
- **Mitigation:** Support exactly the documented ids for this change and return
  clear errors listing those ids. Additional aliases can be evaluated in a
  separate change if needed.

## Documentation drift

- **Risk:** Public docs may continue to say agent-target flags are not
  implemented or may omit the default `generic` behavior.
- **Mitigation:** Include README and docs updates in the implementation PR and
  add documentation guardrail tests where existing test patterns support them.

## Overfitted provider-specific instructions

- **Risk:** Agent-specific text could become too detailed and accidentally
  encode provider SDKs, credentials, network assumptions, or brittle product
  behavior.
- **Mitigation:** Keep adaptations lightweight, stable, and focused on prompt
  style only. Avoid SDK, credential, network, marketplace, and execution
  instructions.

## CLI parser regressions

- **Risk:** Adding `--agent` parsing could weaken existing errors for
  unsupported flags, extra args, duplicate flags, or missing flag values.
- **Mitigation:** Add focused CLI tests for valid and invalid prompt command
  forms, including unknown, missing, duplicate, unsupported, and extra
  arguments.
