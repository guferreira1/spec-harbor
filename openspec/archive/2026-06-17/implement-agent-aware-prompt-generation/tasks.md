# Tasks: Agent-Aware Prompt Generation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`,
      `.specharbor/rules/spec-author.md`, `openspec/project.md`,
      `openspec/specs/architecture/spec.md`, and this change package before
      implementation.
- [x] Inspect current prompt command behavior, role validation, prompt use case,
      template adapter, and prompt tests before editing.
- [x] Record current examples of `prompt <change-id> --role <role>` output for
      at least one representative role so backward compatibility can be checked.
- [x] Confirm the implementation scope is limited to prompt generation,
      documentation, and tests; do not modify `generate`, `validate`, `review`,
      `archive`, `config`, CI, `.github/workflows/ci.yml`, or
      `.specharbor/config.yml`.

## Phase 1: Domain Concepts

- [x] Add a core domain concept for prompt target agents under
      `internal/core/domain`.
- [x] Define supported prompt agent ids exactly as `generic`, `codex`,
      `claude-code`, `devin`, `cursor`, `copilot`, `gemini`, `roo`,
      `windsurf`, and `aider`.
- [x] Provide deterministic validation for unknown prompt agent ids.
- [x] Provide a deterministic supported-agent list for error messages and docs
      tests.
- [x] Provide a default prompt agent value of `generic`.
- [x] Add display labels only if useful for prompt output, keeping the stable
      ids as the command contract.
- [x] Keep the domain package free of adapter, filesystem, terminal, provider,
      network, workflow, source-control, external-agent, and process execution
      imports.

## Phase 2: Ports and Prompt Rendering

- [x] Update the core-owned prompt rendering port/interface in
      `internal/core/ports` to receive the selected prompt target agent as part
      of a structured render request.
- [x] Keep the interface small and owned by the prompt use case.
- [x] Ensure core packages do not import concrete template adapters.
- [x] Preserve existing role prompt inputs and context-aware prompt data passed
      to rendering.
- [x] Add or update rendering tests with fake ports where useful instead of
      depending on concrete adapters from core tests.

## Phase 3: Agent-Aware Template Adaptation

- [x] Update concrete prompt templates/adaptations in
      `internal/adapters/templates` to include a short target-agent section.
- [x] Ensure generated prompt output includes the selected workflow role and
      selected target agent.
- [x] Ensure generated prompt output says the user should paste or use the
      prompt in the selected external tool.
- [x] Ensure generated prompt output states that SpecHarbor does not execute the
      selected agent.
- [x] Add lightweight guidance for each supported prompt agent:
      `generic`, `codex`, `claude-code`, `devin`, `cursor`, `copilot`,
      `gemini`, `roo`, `windsurf`, and `aider`.
- [x] Preserve the existing role-specific prompt body and read-first behavior.
- [x] Keep output deterministic, bounded, copy-pasteable, and free of
      credentials, provider setup, network assumptions, API key instructions,
      SDK behavior, and external command invocations.

## Phase 4: Prompt Use Case Updates

- [x] Update prompt orchestration in `internal/core/usecase` to accept role and
      agent as separate concepts.
- [x] Default omitted agent input to `generic`.
- [x] Validate the selected prompt agent before rendering.
- [x] Preserve all existing role validation and role-specific prompt behavior.
- [x] Preserve context-aware prompt behavior, including classified Project
      Context ordering and assumptions labeling.
- [x] Return structured errors from the use case without CLI-specific formatting.
- [x] Confirm the use case does not import adapters, `os`, terminal IO,
      provider SDKs, network APIs, workflow SDKs, source-control SDKs,
      external-agent SDKs, or process execution packages.

## Phase 5: CLI Parsing and Reporting

- [x] Update `internal/adapters/cli` prompt parsing to support
      `prompt <change-id> --role <role> --agent <agent>`.
- [x] Preserve `prompt <change-id> --role <role>` and default the request agent
      to `generic`.
- [x] Reject unknown prompt agents with a clear error that lists the supported
      prompt agent ids.
- [x] Reject `--agent` without a value with a clear error.
- [x] Reject duplicate `--agent` with a clear error.
- [x] Reject unsupported flags with a clear error.
- [x] Reject extra positional arguments with a clear error.
- [x] Preserve existing `--role` errors and unsupported-role behavior.
- [x] Keep CLI formatting and parsing in adapters; do not move business rules
      into the CLI layer.

## Phase 6: Documentation Updates

- [x] Update `docs/usage.md` with the new prompt command syntax, examples,
      supported prompt agents, default `generic`, and error behavior.
- [x] Update `docs/agent-roles.md` to explain that `role` is the workflow
      responsibility and `agent` is the external target tool style.
- [x] Update `README.md` if command lists, quickstart examples, or capability
      summaries need to mention `--agent`.
- [x] Update `docs/workflow.md` if workflow examples or explanations need to
      show or clarify `--agent`.
- [x] Ensure docs state that `--agent` does not execute the target tool and only
      adapts generated prompt text.
- [x] Ensure docs clearly list supported prompt agents exactly as `generic`,
      `codex`, `claude-code`, `devin`, `cursor`, `copilot`, `gemini`, `roo`,
      `windsurf`, and `aider`.
- [x] Ensure docs say the default prompt agent is `generic`.
- [x] Keep documentation free of credentials, provider setup instructions,
      network assumptions, external-agent execution steps, source-control
      automation, auto-commit, auto-push, and auto-merge instructions.

## Phase 7: Tests

- [x] Add domain tests for supported prompt agents, default `generic`,
      deterministic supported-agent ordering, and unknown agent rejection.
- [x] Add use case tests for omitted agent defaulting to `generic`, explicit
      agent selection, unknown agent errors, and preservation of existing role
      prompt behavior.
- [x] Add template adapter tests for role and agent inclusion, paste/use
      guidance, non-execution wording, deterministic output, and each supported
      agent guidance block.
- [x] Add CLI tests for valid `--agent`, omitted `--agent`, unknown agent,
      missing `--agent` value, duplicate `--agent`, unsupported flags, and
      extra positional arguments.
- [x] Keep all existing prompt role tests passing.
- [x] Add or update documentation guardrail tests if this repository already
      validates public command documentation snippets.
- [x] Confirm tests do not execute external agents, provider APIs, network APIs,
      source-control automation, workflow automation, auto-commit, auto-push, or
      auto-merge behavior.

## Phase 8: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate
      implement-agent-aware-prompt-generation`.
- [x] Manually verify
      `go run ./cmd/specharbor prompt <existing-change-id> --role implementer`
      still works without `--agent`.
- [x] Manually verify
      `go run ./cmd/specharbor prompt <existing-change-id> --role implementer --agent codex`
      prints role and agent information and does not claim execution.
- [x] Manually verify representative error cases for unknown, missing,
      duplicate, unsupported, and extra prompt arguments.
- [x] Update this `tasks.md` only for work actually completed by the
      implementation agent.
- [x] Confirm no production behavior outside prompt generation was changed.
