# Acceptance Criteria: Implement Agent Runner Foundation

## Dry-Run Preservation

- Dry-run remains the default for `generate --agent-assisted` when `--execute` is absent.
- Dry-run never executes external commands.
- Dry-run never requires an `AgentRunner`.
- Dry-run never calls an `AgentRunner`.
- Dry-run does not resolve local executable commands.
- Dry-run validates agents against the recognized agent target model, not against the executable local runner mapping.
- Existing dry-run agent-assisted output formatting remains unchanged for valid recognized targets when `--execute` is absent.
- Existing dry-run authoring plan output remains deterministic for valid recognized targets.
- Existing dry-run copy-pasteable prompt output remains deterministic for valid recognized targets.
- Dry-run works for `codex`, `claude`, `devin`, `cursor`, `copilot`, `gemini`, `roo`, `windsurf`, `aider`, and `generic` where prompt rendering is supported.
- Dry-run rejects unknown agent targets with a clear validation error.
- Unknown-agent rejection in dry-run is documented as an intentional validation tightening.
- Dry-run writes no files.
- Dry-run writes no prompt file.
- Dry-run creates no OpenSpec files.
- Dry-run modifies no production code.

## Agent Target Model

- Concrete recognized agent targets include Codex, Claude Code, Devin, Cursor, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, and Aider.
- Concrete recognized agent target ids include exactly `codex`, `claude`, `devin`, `cursor`, `copilot`, `gemini`, `roo`, `windsurf`, and `aider`.
- `generic` remains recognized for dry-run only.
- `generic` is not executable in this change because no deterministic local command mapping exists for it.
- The `generic` exception is documented as future config-driven runner/templates scope.
- Recognized agent targets can be described, validated, and used for deterministic prompt rendering and dry-run reports.
- Concrete recognized agent targets can be resolved to executable local runner mappings in execute mode.
- The recognized target and executable local runner mapping model is owned by `internal/core/domain`.
- Unknown agent targets are rejected with a clear error.
- This feature does not implement provider, cloud, IDE, OAuth, credential, marketplace, or remote execution integrations for any agent target.

## Executable Mappings

- Execute mode resolves supported executable local runner targets through a built-in local command mapping.
- The executable mapping includes `codex -> codex`.
- The executable mapping includes `claude -> claude`.
- The executable mapping includes `devin -> devin`.
- The executable mapping includes `cursor -> cursor`.
- The executable mapping includes `copilot -> copilot`.
- The executable mapping includes `gemini -> gemini`.
- The executable mapping includes `roo -> roo`.
- The executable mapping includes `windsurf -> windsurf`.
- The executable mapping includes `aider -> aider`.
- Resolved executable targets include command name, optional fixed args, display name, and agent id.
- No `generic -> generic` executable mapping is introduced.
- Agent names are not treated as provider names.
- Agent names are not resolved through remote registries or marketplaces.
- Agent names are not shell-interpreted commands.
- The local runner adapter receives already-resolved executable commands.
- No generic arbitrary shell command execution is introduced.

## Execute Command

- `--execute` is required for runner execution.
- `--execute` is supported only with `generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`.
- `--execute` remains unsupported for `--blank`.
- `--execute` remains unsupported for `--template`.
- `--execute` remains unsupported for `--guided`.
- `--execute` remains unsupported when no generation mode is provided.
- Supported authoring types for execute are exactly `feature`, `bugfix`, `docs`, and `refactor`.
- Execute mode renders the same deterministic OpenSpec authoring prompt as dry-run mode.
- Execute mode validates the selected agent as a recognized agent target.
- Execute mode rejects unknown agent targets with a clear error.
- Execute mode rejects `generic` with a clear unmapped-target error.
- Execute mode accepts all mapped local executable targets.
- Execute mode sends the rendered prompt to the local runner through stdin.
- Execute mode uses the documented working directory.
- Execute mode runs only through the core-owned `AgentRunner` port.
- The use case does not run local commands directly.
- Concrete local command execution lives in an adapter package.

## Local Runner Result

- The local runner captures stdout only for started processes.
- The local runner captures stderr only for started processes.
- The local runner captures the actual exit code when the process starts.
- The local runner captures execution status when the process starts.
- Zero exit code is reported as a successful runner execution with status `success`.
- Non-zero exit code is reported as a completed runner execution with status `non_zero_exit`.
- Non-zero exit code does not panic.
- Non-zero exit reporting includes captured stdout.
- Non-zero exit reporting includes captured stderr.
- Non-zero exit reporting includes the actual exit code.
- Non-zero exit reporting includes status `non_zero_exit`.
- Startup failures return clear errors.
- Startup failures do not panic.
- Startup failures are SpecHarbor execution errors, not successful runner results.
- Startup failures produce no normal `AgentRunnerResult`.
- Startup failures report no exit code.
- Missing executable failures return clear errors including the resolved command and useful diagnostic context.
- Startup failure is distinct from a started process that exits non-zero.
- Runner output is displayed as output only.

## CLI Exit Behavior

- For zero runner exits, the CLI prints the run-and-report output and exits zero.
- For non-zero runner exits, the CLI prints the full run-and-report output first.
- For non-zero runner exits, the printed report includes stdout, stderr, exit code, and status `non_zero_exit`.
- After printing the report for a non-zero runner exit, the SpecHarbor CLI process exits non-zero.
- For startup failures, the CLI returns a clear error message.
- For startup failures, the CLI exits non-zero.
- For startup failures, the CLI does not print a normal runner result and does not report an exit code.

## Output Safety

- Agent stdout is not parsed as files.
- Agent stderr is not parsed as files.
- Agent output is not applied.
- Agent output is not converted into patches.
- Agent output is not written into OpenSpec files.
- SpecHarbor does not write `proposal.md` from agent output.
- SpecHarbor does not write `design.md` from agent output.
- SpecHarbor does not write `tasks.md` from agent output.
- SpecHarbor does not write `acceptance-criteria.md` from agent output.
- SpecHarbor does not write `risks.md` from agent output.
- SpecHarbor does not modify production code from agent output.
- SpecHarbor does not auto-commit.
- SpecHarbor does not auto-push.
- SpecHarbor does not auto-merge.
- SpecHarbor does not run source-control commands.
- SpecHarbor does not run workflow tools.

## Architecture

- Agent-assisted orchestration remains in `internal/core/usecase`.
- Domain concepts remain in `internal/core/domain`.
- The `AgentRunner` interface belongs in `internal/core/ports`.
- Concrete local command execution belongs in `internal/adapters`.
- CLI parsing remains in `internal/adapters/cli`.
- CLI dependency wiring remains in `internal/adapters/cli`.
- CLI current working directory lookup remains in `internal/adapters/cli`.
- CLI reporting remains in `internal/adapters/cli`.
- CLI process exit behavior remains in `internal/adapters/cli`.
- Prompt rendering continues through the existing core-owned prompt renderer port.
- The CLI report clearly distinguishes recognized agent name from resolved executable command in execute mode.
- Core packages do not import adapters.
- Core packages do not import `os`.
- Core packages do not import `os/exec`.
- Core packages do not perform terminal IO.
- Core packages do not import provider SDKs.
- Core packages do not import network APIs.
- Core packages do not import source-control SDKs.
- Core packages do not import workflow SDKs.
- Core packages do not import external-agent SDKs.
- Use cases depend on interfaces, not concrete adapters.
- CLI code does not contain authoring business rules.

## Exclusions

- No OpenAI provider API integration is introduced.
- No Anthropic provider API integration is introduced.
- No Gemini provider API integration is introduced.
- No local model API integration is introduced.
- No OAuth is introduced.
- No credential management is introduced.
- No secret storage is introduced.
- No cloud execution is introduced.
- No remote execution is introduced.
- No IDE automation is introduced.
- No marketplace behavior is introduced.
- No source-control automation is introduced.
- No workflow automation is introduced.
- No write/apply port for agent output is introduced.
- No confirmation flow for applying files is introduced.
- No AI-assisted generation implementation is introduced.
- No production-code implementation behavior is introduced.

## Regression Behavior

- Existing `specharbor generate <change-id> --blank` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template feature` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template bugfix` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template docs` behavior remains unchanged.
- Existing `specharbor generate <change-id> --template refactor` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type bugfix --title "<title>" --summary "<summary>"` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type docs --title "<title>" --summary "<summary>"` behavior remains unchanged.
- Existing `specharbor generate <change-id> --guided --type refactor --title "<title>" --summary "<summary>"` behavior remains unchanged.
- `init`, `scan`, `validate`, `prompt`, `review`, `archive`, `config`, `help`, `version`, and unknown command behavior remain unchanged.

## Documentation and Verification

- `README.md` is updated in the same implementation PR if command lists or examples need changing.
- `docs/usage.md` is updated in the same implementation PR.
- `docs/generation-modes.md` is updated in the same implementation PR.
- Other relevant Markdown under `docs/` is updated if needed.
- Docs state that SpecHarbor supports multiple AI agent targets for OpenSpec/SDD workflows.
- Docs state that dry-run is still default.
- Docs state that dry-run generates copy-pasteable deterministic prompts.
- Docs state that dry-run rejects unknown agents as an intentional validation tightening.
- Docs list concrete recognized agent targets: Codex, Claude Code, Devin, Cursor, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, and Aider.
- Docs state that `generic` is dry-run-only until future config-driven runner/templates support.
- Docs state that `--execute` is explicit.
- Docs state that `--execute` runs a supported local command mapping.
- Docs state the executable mapping: `codex -> codex`, `claude -> claude`, `devin -> devin`, `cursor -> cursor`, `copilot -> copilot`, `gemini -> gemini`, `roo -> roo`, `windsurf -> windsurf`, and `aider -> aider`.
- Docs state that `--execute` is still run-and-report only.
- Docs state that missing local commands produce startup errors.
- Docs state that startup failures produce no runner result and no exit code.
- Docs state that started commands with non-zero exit codes produce a report and then SpecHarbor exits non-zero.
- Docs state that stdout, stderr, exit code, and execution status are captured and reported for started processes.
- Docs state that SpecHarbor does not parse or apply agent output.
- Docs state that SpecHarbor does not write OpenSpec files from agent output.
- Docs state that SpecHarbor does not modify production code from agent output.
- Docs state that SpecHarbor does not auto-commit, auto-push, or auto-merge.
- Docs state that provider APIs, IDE automation, remote execution, OAuth, credentials, marketplace integrations, source-control automation, and workflow automation remain out of scope.
- Focused tests cover recognized targets, executable mappings, generic behavior, dry-run behavior, unknown agents, execute success for all mapped targets, non-zero exits, startup failures, local runner adapter behavior, CLI reporting and exit behavior, architecture boundaries, safety exclusions, and blank/template/guided regression behavior.
- `go test ./...` succeeds.
