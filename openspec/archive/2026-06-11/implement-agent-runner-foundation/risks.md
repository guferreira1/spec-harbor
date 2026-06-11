# Risks: Implement Agent Runner Foundation

## Dry-Run Regression

Dry-run is currently the safe default. Adding an execute branch could accidentally change dry-run output for valid recognized targets, require runner dependencies, resolve executable commands, or execute commands when `--execute` is absent.

Mitigation:

- Keep dry-run as the default path.
- Preserve dry-run report formatting and deterministic prompt rendering for valid recognized targets.
- Document unknown-agent rejection as an intentional validation tightening.
- Do not require an `AgentRunner` for dry-run.
- Do not resolve executable commands for dry-run.
- Add regression tests for existing dry-run output with valid recognized targets.
- Add tests proving dry-run never calls a runner.

## Agent Coverage Conflation

SpecHarbor supports a broad ecosystem of AI agents for OpenSpec/SDD workflows. In this change, concrete targets such as Devin, Cursor, GitHub Copilot, Roo Code, and Windsurf are executable only as deterministic local command mappings. Users could mistake that for provider APIs, cloud execution, IDE automation, OAuth integration, marketplace integration, or remote execution.

Mitigation:

- Model recognized agent targets separately from resolved executable commands in core/domain.
- Resolve concrete executable targets only through explicit built-in mappings.
- Document that executable means local command execution with stdin/stdout/stderr/exit-code reporting only.
- Document that this feature does not add provider, cloud, IDE, OAuth, marketplace, or remote integrations.
- Keep output parse/apply behavior out of scope.
- Add docs and tests proving all concrete mappings are local-command mappings only.

## Generic Agent Ambiguity

`generic` is useful for dry-run prompt generation, but it has no deterministic local command mapping. Making it executable without configuration would either invent an unsafe default command or imply arbitrary command execution.

Mitigation:

- Keep `generic` as a recognized dry-run-only target.
- Reject `generic` in execute mode with a clear unmapped-target error.
- Document that generic execution belongs to a future config-driven runner/templates feature.
- Do not add `generic -> generic` in this change.
- Do not add generic arbitrary command flags.

## Unsafe Command Resolution

Treating `--agent` as an arbitrary executable command would allow shell-like behavior, surprising command execution, and unsafe local side effects.

Mitigation:

- Use a built-in executable mapping for concrete supported targets.
- Reject unknown agents with a clear error.
- Execute commands directly without a shell.
- Pass optional fixed args directly to the process execution API without shell interpolation.
- Do not add generic arbitrary command flags in this change.
- Keep dry-run recognized-agent behavior separate from execute command resolution.
- Ensure the local runner adapter receives only already-resolved executable commands.

## External Process Side Effects

Local agent CLIs are external programs and may have their own behavior, including file writes or network use. SpecHarbor cannot fully control an installed tool after launching it.

Mitigation:

- Require explicit `--execute`.
- Document that local agent command behavior is controlled by the installed local tool.
- Send an OpenSpec-authoring-only prompt that forbids production-code changes.
- Do not provide file application or patch application paths in SpecHarbor.
- Do not parse or write agent output.
- Report that SpecHarbor did not write files from output and did not modify production code from output.

## Non-Zero Exit Ambiguity

A non-zero local agent exit could be confused with a SpecHarbor startup failure or could hide useful stdout/stderr if handled as an ordinary error too early.

Mitigation:

- Treat started processes that exit non-zero as completed runner results with status `non_zero_exit`.
- Capture and print stdout, stderr, actual exit code, and status for non-zero exits.
- Avoid panics.
- Print the full run-and-report output before the SpecHarbor CLI process exits.
- Make the SpecHarbor CLI process exit non-zero after printing the report for deterministic shell and automation behavior.
- Cover this behavior with use case and CLI tests.

## Startup Failure Handling

Missing local executables, permission errors, or invalid working directories can prevent the command from starting. Representing these as normal runner results would blur the boundary between a failed launch and a launched process that returned an exit code.

Mitigation:

- Return clear startup errors from the `AgentRunner` port.
- Produce no normal `AgentRunnerResult` for startup failures.
- Do not report an exit code when no process started.
- Include the resolved command and diagnostic context in CLI errors.
- Keep startup failure distinct from `non_zero_exit`.
- Cover missing executable and startup failure paths with adapter, use case, and CLI tests.

## Output Application Creep

Once stdout is captured, future work may be tempted to parse it into files or apply patches immediately. That would cross the first-version safety boundary.

Mitigation:

- Explicitly prohibit parsing stdout or stderr.
- Do not add write/apply ports.
- Do not write OpenSpec files from agent output.
- Do not apply patches.
- Do not modify production code from output.
- Add tests and architecture checks for absence of write/apply behavior.

## Architecture Leakage

Process execution belongs in adapters. Adding `os`, `os/exec`, command handling, or CLI process-exit behavior to core would violate the architecture contract.

Mitigation:

- Add a core-owned `AgentRunner` port.
- Implement process execution only in an adapter.
- Keep recognized targets and executable mappings in core/domain.
- Keep use cases dependent on interfaces.
- Keep CLI parsing, dependency wiring, current working directory lookup, report formatting, and process exit behavior in `internal/adapters/cli`.
- Keep core imports free of adapters, `os`, `os/exec`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, and external-agent SDKs.
- Update architecture boundary tests to allow only the narrow runner port and adapter.

## Provider, Cloud, IDE, and Credential Creep

Supporting broad agent names can be mistaken for supporting provider APIs, provider credentials, OAuth, local model APIs, IDE automation, marketplace integrations, remote execution, or cloud-agent integrations.

Mitigation:

- Treat executable agent names as local command targets only.
- Do not add provider SDKs or provider ports.
- Do not add IDE automation, remote execution, marketplace integrations, or workflow connector behavior for agent targets.
- Do not add OAuth or credential storage.
- Do not require provider API keys.
- Document that provider APIs, IDE automation, remote execution, OAuth, credentials, and marketplace integrations are not implemented by this feature.

## Documentation Drift

Public docs currently say `--execute` is unsupported. After implementation, stale docs would mislead users about CLI behavior, startup failures, non-zero exit handling, generic behavior, and safety boundaries.

Mitigation:

- Update README and docs in the implementation PR.
- Document the recognized agent target list.
- Document the executable local runner mapping.
- Document `generic` as dry-run-only.
- Document dry-run default behavior.
- Document unknown-agent validation tightening for dry-run.
- Document that dry-run can generate prompts for recognized agents.
- Document explicit run-and-report execute behavior.
- Document captured stdout, stderr, exit code, and status for started processes.
- Document startup failure errors with no runner result and no exit code.
- Document non-zero runner exits producing a full report followed by non-zero SpecHarbor CLI exit.
- Document non-application of output and no source-control/workflow automation.

## Existing Generation Regression

Changing generate parsing to support `--execute` could alter blank, template, guided, or dry-run agent-assisted generation.

Mitigation:

- Keep parser changes narrowly scoped.
- Preserve existing mode conflict rules.
- Add CLI regression tests for blank, template, guided, and dry-run agent-assisted modes.
- Manually verify existing generation examples before finalizing.

## Test Fragility Around Local Commands

Tests that depend on installed agent CLIs would be unreliable across developer machines and CI.

Mitigation:

- Test executable mappings without launching real agent CLIs.
- Test the local runner adapter with controlled helper commands or Go test helper process patterns.
- Do not require `codex`, `claude`, `devin`, `cursor`, `copilot`, `gemini`, `roo`, `windsurf`, or `aider` to be installed for automated tests.
- Keep executable mapping tests separate from real process execution tests.

## Overbuilding

Future versions may support custom command configuration, provider APIs, confirmation flows, file application, source-control integration, workflow dispatch, or hybrid generation. Adding those now would increase risk.

Mitigation:

- Implement only explicit local run-and-report behavior.
- Avoid config mutation, registries, marketplaces, provider integrations, source-control integrations, workflow integrations, and file application.
- Keep interfaces small and driven by the immediate behavior.
