# Risks: Implement Agent-Assisted Spec Authoring

## Dry-Run Safety Regression

The central safety property is that `--agent-assisted` renders a prompt and report only. Accidentally running a local command, calling a provider, accessing the network, or writing files would violate the first-version product boundary.

Mitigation:

- Make the first version dry-run-only.
- Reject `--execute` as unsupported with a clear error.
- Add tests proving dry-run reports no external commands executed.
- Add tests proving dry-run writes no files.
- Add tests proving dry-run writes no prompt file.
- Print "no external command executed", "no files written", and "no prompt file written" in dry-run reports.

## Execute Scope Creep

`--execute` may appear to be the natural next step after prompt generation, but adding it now would require command configuration, process execution, output capture, and stronger safety controls.

Mitigation:

- Explicitly defer `--execute` to a future change.
- Return a clear unsupported flag/mode error when `--execute` is provided.
- Do not add `AgentRunner`, local command adapters, command output capture, execute use case paths, or execute tests except unsupported-error coverage.
- Ensure unsupported `--execute` runs nothing and writes nothing.

## Prompt Output Ambiguity

If dry-run references prompt output paths, users and implementers may assume SpecHarbor creates prompt files or sidecar artifacts.

Mitigation:

- Always print the deterministic authoring prompt to stdout.
- Do not support prompt-file output in this change.
- Report that no prompt file was written.
- Add tests proving the prompt is copy-pasteable from stdout and does not depend on a prompt file.

## Agent Scope Creep

Agent-assisted authoring could be confused with agent-assisted implementation. A prompt that asks the agent to implement code would break the workflow boundary.

Mitigation:

- Prompt the agent to create or refine only the OpenSpec change package.
- Explicitly prohibit production code changes.
- Explicitly prohibit unrelated file changes.
- Explicitly require implementation tasks to remain unchecked.
- Keep generated output expectations Markdown-only OpenSpec content.
- Add prompt tests for the production-code prohibition.

## File Application Risk

Parsing or applying agent output could overwrite authored specs, write malformed Markdown, or apply code changes hidden in agent output.

Mitigation:

- Never parse, apply, or write agent output in this change, including OpenSpec files.
- Do not add write/apply ports.
- Do not add confirmation flows.
- Do not write OpenSpec files from agent output.
- Do not run post-write validation because no files are written.

## Existing Generation Regression

Adding `--agent-assisted`, `--agent`, and unsupported `--execute` handling to `generate` could accidentally change existing blank, template, or guided parsing, output, validation, idempotency, or overwrite behavior.

Mitigation:

- Keep agent-assisted parsing additive.
- Preserve existing mode conflict behavior.
- Add regression tests for blank generation.
- Add regression tests for template generation.
- Add regression tests for guided generation.
- Add CLI tests for unrelated command behavior.

## Architecture Leakage

The feature touches CLI flags, prompt rendering, and reports. A common failure mode is putting business rules in the CLI adapter or introducing execution/write abstractions before they are needed.

Mitigation:

- Keep CLI responsibilities limited to parsing, dependency construction, current-working-directory lookup, and report formatting.
- Keep orchestration in `internal/core/usecase`.
- Keep domain concepts in `internal/core/domain`.
- Keep prompt rendering behind core-owned ports.
- Keep concrete prompt templates in adapters.
- Do not add `AgentRunner`, local command adapters, write/apply ports, workflow connector ports, or confirmation flows.
- Keep core free of `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, and process execution packages.

## Provider and Credential Creep

Supporting named agents may invite direct provider integrations, API keys, OAuth, credential storage, or provider-specific payloads before the architecture is ready.

Mitigation:

- Treat `--agent <agent-name>` as a displayable agent target for prompt generation, not as an AI provider.
- Do not add provider SDKs or provider API calls.
- Do not add OAuth or credential storage.
- Do not require provider API keys.
- Do not resolve agent names to local or remote commands in this change.
- Keep provider-specific integrations out of core.

## Prompt Drift

The prompt may omit important constraints, omit required files, or diverge from repository architecture rules over time.

Mitigation:

- Build required filenames from `domain.RequiredOpenSpecChangeFiles()`.
- Include architecture boundaries from `openspec/specs/architecture/spec.md`.
- Include scope boundaries in tests.
- Add prompt tests for required files and key safety instructions.
- Keep prompt template content deterministic and local.

## Documentation Scope Confusion

The prompt must prohibit README/docs changes for non-docs work while still allowing documentation-specific OpenSpec authoring when the requested type is `docs`.

Mitigation:

- Include type-aware prompt wording.
- For non-docs types, explicitly prohibit README/docs changes.
- For `docs`, allow documentation scope only as described by the title and summary while still prohibiting production code changes.
- Add tests for docs and non-docs prompt boundaries.

## Unsupported Agent Ambiguity

The product direction includes many agents and custom workflows. Rejecting unknown agent names during dry-run could block useful prompt generation, while treating names as executable commands would be unsafe.

Mitigation:

- Require a non-empty, safe agent name for dry-run.
- Do not treat unknown agent names as unknown providers.
- Do not treat agent names as executable commands.
- Defer command configuration and unsupported-agent execution behavior to a future `--execute` change.

## Overbuilding

Future SpecHarbor versions may support richer agent configuration, local execution, file application, confirmation flows, hybrid generation, provider APIs, or workflow integrations. Adding those abstractions now would increase risk and surface area.

Mitigation:

- Implement only the foundation required for deterministic stdout prompt generation and dry-run reporting.
- Avoid remote registries, marketplaces, provider integrations, source-control integrations, workflow integrations, config mutation, credential storage, command execution, and file application.
- Add ports only where current dry-run behavior consumes them.
- Defer file application, confirmation flows, local execution, and workflow automation.

## Validation Claims

Dry-run does not write files, so claiming validation success would be misleading.

Mitigation:

- In dry-run, instruct the agent to run or recommend `specharbor validate <change-id>` when files exist.
- Do not run validation automatically in dry-run.
- Do not print validation success for unwritten files.
- Defer post-write validation until a future change that actually writes files.
