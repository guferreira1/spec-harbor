# Design: Implement Agent Runner Foundation

## Overview

This change turns `--execute` into a supported explicit path for agent-assisted OpenSpec spec authoring:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>" --execute
```

The existing dry-run path remains the default when `--execute` is absent. For valid recognized agent targets, dry-run report formatting and deterministic prompt rendering must remain backward-compatible. Unknown agent targets are now rejected even in dry-run mode; this is an intentional validation tightening.

The execute path is intentionally narrow:

1. Validate the same agent-assisted authoring inputs.
2. Validate the selected agent target.
3. Render the same deterministic OpenSpec authoring prompt.
4. Resolve the selected concrete agent target to a supported local executable mapping only when `--execute` is present.
5. Send the prompt to the local command through stdin.
6. Capture stdout, stderr, exit code, and execution status when the process starts.
7. Print the runner result.
8. Stop without parsing, applying, writing, committing, pushing, merging, or dispatching workflow actions.

This feature is still about OpenSpec spec authoring. It is not a production-code implementation mode.

## Command Contract

Dry-run remains:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

Execute becomes supported only for:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type feature --title "<title>" --summary "<summary>" --execute
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type bugfix --title "<title>" --summary "<summary>" --execute
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type docs --title "<title>" --summary "<summary>" --execute
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type refactor --title "<title>" --summary "<summary>" --execute
```

Supported authoring types remain exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

`--execute` must remain unsupported for non-agent-assisted generation modes and must continue to return a clear error when used with `--blank`, `--template`, or `--guided`.

## Agent Target Model

The domain model must define recognized agent targets separately from resolved executable commands.

Concrete recognized agent targets are domain-level display, prompt-rendering, and executable local runner targets:

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

`generic` remains a recognized dry-run target only:

```text
generic   -> Generic Agent
```

The `generic` exception exists because a useful generic runner needs user- or project-specific command configuration. Config-driven command mapping belongs to a future config-driven runner/templates feature. This change must not invent a `generic -> generic` executable mapping.

Dry-run validation must validate `--agent` against the recognized target set, including `generic`. Dry-run must not require the selected target to have an executable mapping. Dry-run must not resolve executable mappings and must not call an `AgentRunner`.

Unknown agent targets must return a clear error such as:

```text
unknown agent target: <agent-name>
```

The recognized target model belongs in `internal/core/domain`. It must not be owned by CLI code, adapters, provider SDKs, or workflow connectors.

## Executable Local Runner Mapping

Executable local runner targets are concrete recognized agent targets that SpecHarbor may run through the first local command runner foundation.

Use a small built-in mapping for supported local executable runner targets. Every concrete supported agent target is executable in this foundation:

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

The resolved executable target should be modeled as:

- command name;
- optional fixed args;
- display name;
- agent id.

The first implementation uses no fixed args for the listed mappings. If deterministic fixed args are added later, the adapter must execute the command directly with those args and must not use shell interpolation.

Do not add aliases, marketplace resolution, config-driven resolution, remote resolution, provider-backed resolution, shell-interpreted command resolution, or arbitrary command flags in this change.

Execute mode must reject unknown targets distinctly from known but unmapped targets:

```text
unknown agent target: <agent-name>
```

Execute mode must reject `generic` with a clear error because it is recognized for dry-run but has no deterministic executable mapping:

```text
agent target has no executable local runner mapping in this change: generic
```

Dry-run without `--execute` must not reject `generic` only because it is absent from the executable mapping.

Devin, Cursor, GitHub Copilot, Roo Code, and Windsurf are executable local runner targets in this feature only in the local-command sense. This does not add provider API integration, IDE automation, OAuth integration, cloud execution, marketplace integration, or remote execution for those tools.

## Dry-Run Contract

The statement "existing dry-run behavior/output remains unchanged" means:

- dry-run remains the default when `--execute` is absent;
- for valid recognized agent targets, dry-run output formatting remains backward-compatible;
- for valid recognized agent targets, deterministic prompt rendering remains backward-compatible;
- dry-run continues to generate copy-pasteable deterministic prompts;
- dry-run never requires an `AgentRunner`;
- dry-run never resolves executable command mappings;
- dry-run never executes external commands;
- dry-run supports `codex`, `claude`, `devin`, `cursor`, `copilot`, `gemini`, `roo`, `windsurf`, `aider`, and `generic`;
- unknown agent targets are rejected with a clear validation error, even in dry-run mode;
- unknown-agent rejection is an intentional validation tightening and does not require preserving prior output for invalid inputs;
- existing examples using recognized agents continue to produce the same dry-run style output.

## Prompt Handling

Prompt rendering continues to use the existing core-owned prompt renderer port:

```text
ports.AgentAssistedAuthoringPromptRenderer
```

The execute path must render the same deterministic prompt produced by dry-run mode. The runner receives that exact prompt content through stdin unless implementation discovers a compelling documented reason to use another deterministic channel.

The prompt must remain an OpenSpec authoring prompt. It must continue to instruct the agent to create or refine only:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

SpecHarbor must not write a prompt file.

## Domain Concepts

Add small domain types under:

```text
internal/core/domain
```

Expected concepts:

- recognized agent target with canonical id and display name;
- executable local runner target or executable command mapping;
- resolved executable command with command name, optional fixed args, display name, and agent id;
- agent run request;
- agent run result for started processes;
- execution status for started processes;
- exit code representation for started processes.

Execution statuses such as `success` and `non_zero_exit` apply only after the process starts. Startup failure is represented as an error path, not as an execution status in a normal `AgentRunnerResult`.

A practical started-process result shape should support CLI reporting without CLI formatting in core:

```text
RecognizedAgentTarget
AgentDisplayName
CommandName
FixedArgs
WorkingDirectory
Status
Stdout
Stderr
ExitCode
PromptSent
OutputParsed
FilesWrittenFromOutput
ProductionCodeModifiedBySpecHarbor
```

The exact field names may differ, but the domain result must make these facts expressible:

- the command was requested only because `--execute` was explicit;
- the selected recognized agent target is distinct from the resolved executable command;
- the prompt was sent to the runner;
- stdout was captured only for a started process;
- stderr was captured only for a started process;
- the actual exit code was captured for a started process;
- started-process execution status is either `success` or `non_zero_exit`;
- startup failure produced no normal runner result and no exit code;
- output was not parsed;
- files were not written from output;
- production code was not modified by SpecHarbor.

Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, and concrete process execution packages.

The local runner adapter must receive only already-resolved executable commands. It must not own recognized-target policy or decide whether an agent id is executable.

## Ports

Add a core-owned runner port under:

```text
internal/core/ports
```

Suggested interface:

```text
type AgentRunner interface {
    Run(request domain.AgentRunRequest) (domain.AgentRunResult, error)
}
```

The exact method and type names may differ, but the contract must:

- accept normalized domain input;
- include the executable command name;
- include optional fixed args;
- include the prompt content;
- include the working directory;
- return a normal runner result only when the process starts;
- capture stdout for started processes;
- capture stderr for started processes;
- capture the actual exit code for started processes;
- classify zero exit as `success`;
- classify non-zero exit as `non_zero_exit`;
- return a clear error for command startup failures;
- produce no normal `AgentRunnerResult` for startup failures;
- report no exit code for startup failures;
- keep startup failure distinct from a started process that exits non-zero;
- avoid panics;
- avoid file application, source-control automation, workflow automation, provider APIs, credentials, and terminal UI.

Do not add write/apply ports, source-control ports, workflow connector ports, provider ports, marketplace ports, remote registry ports, or confirmation-flow ports in this change.

## Use Case Orchestration

Update agent-assisted orchestration under:

```text
internal/core/usecase
```

The use case should continue to own the authoring workflow:

- validate required dependencies;
- validate project root;
- validate safe change id;
- validate agent name against the recognized target model;
- validate authoring type;
- validate title and summary;
- build the relative OpenSpec change path;
- obtain required files from `domain.RequiredOpenSpecChangeFiles()`;
- render the deterministic prompt through the prompt renderer port.

For dry-run:

- do not require an `AgentRunner`;
- do not resolve an executable command;
- do not execute a command;
- allow all recognized dry-run targets, including `generic`;
- reject unknown agents with a clear validation error;
- return the existing dry-run result shape and statuses for valid recognized targets;
- preserve current output formatting for valid recognized targets.

For execute:

- require an `AgentRunner`;
- validate the agent name against the recognized target model;
- reject unknown agent targets with a clear error;
- resolve the recognized concrete agent target through the built-in executable mapping;
- reject `generic` with a clear unmapped-target error before runner invocation;
- render the same deterministic prompt as dry-run mode;
- build an agent run request with the prompt, command name, optional fixed args, and working directory;
- invoke the runner port exactly once;
- combine the authoring context and started-process run result into a structured result for CLI reporting;
- propagate startup failures as errors without producing a normal runner result;
- do not parse stdout or stderr;
- do not write files from output;
- do not apply patches;
- do not run validation against unwritten files;
- do not run source-control commands;
- do not run workflow tools;
- do not call provider APIs.

The use case must not import adapters, `os`, terminal IO, `os/exec`, provider SDKs, network APIs, source-control SDKs, workflow SDKs, or external-agent SDKs.

## Local Runner Adapter

Add a concrete local command runner under an adapter package such as:

```text
internal/adapters/agents
```

or:

```text
internal/adapters/agentrunner
```

The adapter may use Go process execution packages because concrete execution belongs outside core.

Adapter requirements:

- receive an already-resolved executable command from the use case;
- run the executable directly, not through a shell;
- pass optional fixed args directly to the process execution API;
- set the documented working directory, preferably the project root supplied by the CLI;
- send the prompt to stdin;
- capture stdout only when the process starts;
- capture stderr only when the process starts;
- capture the process exit code when the command starts;
- classify a zero exit code as `success`;
- classify a non-zero exit code as a completed run with status `non_zero_exit`;
- return startup failures as clear errors;
- include the resolved command and enough diagnostic context in startup failure errors;
- return no normal runner result for startup failures;
- report no exit code for startup failures;
- avoid panics for startup failures and non-zero exits;
- avoid source-control commands;
- avoid workflow tools;
- avoid provider SDKs;
- avoid credential storage;
- avoid parsing command output;
- avoid writing files from command output.

The adapter should not add shell quoting rules, command templates, mutable global state, remote configuration, or user credentials.

## Working Directory

Use the project root obtained by the CLI adapter as the runner working directory. This keeps relative paths in the existing deterministic prompt meaningful.

The CLI report and docs must state the working directory used for `--execute`. Do not print sensitive environment details or absolute paths unless existing error behavior requires them. If the implementation decides another working directory is safer, it must document that directory explicitly in code comments, tests, CLI output, and docs.

## CLI Parsing, Reporting, and Exit Behavior

CLI parsing remains in:

```text
internal/adapters/cli
```

Update parsing so:

- `--execute` is accepted only with `--agent-assisted`;
- `--execute` remains an error with `--blank`, `--template`, `--guided`, or no generation mode;
- dry-run remains the default when `--execute` is absent;
- dry-run rejects unknown agents with a clear validation error;
- dry-run does not require or construct execution results;
- execute mode rejects unknown agents with a clear validation error;
- execute mode rejects `generic` with a clear unmapped-target error;
- existing mode conflict, missing flag, duplicate flag, unsupported flag, and extra positional argument behavior remains intact.

CLI construction should inject:

- the existing prompt template adapter;
- the new local runner adapter for execute-capable agent-assisted use cases.

CLI reporting should print a deterministic runner report containing:

- operation name;
- dry-run or execute mode;
- change id;
- recognized agent target id;
- recognized agent display name;
- resolved command name;
- resolved fixed args when present;
- authoring type;
- title;
- summary;
- relative OpenSpec change path;
- working directory label;
- execution status for started processes;
- exit code for started processes;
- stdout section for started processes;
- stderr section for started processes;
- output-not-parsed status;
- files-not-written-from-output status;
- no-production-code-modified-by-SpecHarbor status;
- no auto-commit, auto-push, or auto-merge status.

For a started process that exits non-zero:

- the runner result status must be `non_zero_exit`;
- the report must include captured stdout;
- the report must include captured stderr;
- the report must include the actual exit code;
- the report must be printed before the SpecHarbor CLI process exits;
- after printing the report, the SpecHarbor CLI process must exit non-zero.

For startup failures:

- the `AgentRunner` port returns a clear error;
- no normal `AgentRunnerResult` is produced;
- no exit code is reported;
- stdout and stderr are not reported as captured process output;
- the CLI error message includes the resolved command and enough context to diagnose the missing or failed executable;
- the CLI process exits non-zero;
- the CLI does not attempt to parse or apply any output.

## Documentation Updates

This feature changes public CLI behavior, so the implementation PR must update:

- `README.md` if command lists or examples need changing;
- `docs/usage.md`;
- `docs/generation-modes.md`;
- any other relevant Markdown under `docs/`.

Docs must explain:

- SpecHarbor supports multiple AI agent targets for OpenSpec/SDD workflows;
- dry-run is still default;
- dry-run generates copy-pasteable deterministic prompts;
- unknown dry-run agents are rejected as an intentional validation tightening;
- `--execute` is explicit;
- `--execute` runs a supported local command mapping;
- the supported executable mappings are:

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

- `generic` is dry-run-only until a future config-driven runner/templates feature defines a deterministic command mapping;
- `--execute` is still run-and-report only;
- missing local commands produce startup errors;
- started commands with non-zero exit codes produce a report and then SpecHarbor exits non-zero;
- SpecHarbor captures stdout, stderr, exit code, and status for started processes;
- SpecHarbor does not parse or apply output;
- SpecHarbor does not write files from output;
- SpecHarbor does not modify production code from output;
- SpecHarbor does not auto-commit, auto-push, or auto-merge;
- provider APIs, IDE automation, OAuth, credentials, marketplace integrations, remote execution, and workflow automation remain out of scope;
- local agent command behavior is controlled by the installed local tool.

## Testing Strategy

Add focused tests at these levels:

- domain tests for recognized agent targets, generic dry-run-only behavior, executable mapping, unknown target errors, run status, exit code representation, startup-failure error contract, and result immutability;
- domain tests proving the executable mappings are `codex -> codex`, `claude -> claude`, `devin -> devin`, `cursor -> cursor`, `copilot -> copilot`, `gemini -> gemini`, `roo -> roo`, `windsurf -> windsurf`, and `aider -> aider`;
- port-level compile coverage by using fake `AgentRunner` implementations in use case tests;
- use case tests for dry-run default behavior, dry-run prompt rendering for all recognized targets, dry-run unknown-agent rejection, dry-run not requiring or calling a runner, dry-run not resolving executable mappings, execute success for all mapped targets, execute non-zero exit, startup failure, missing runner dependency, `generic` execute rejection, unknown target rejection, no prompt parsing, and no output application;
- local adapter tests using small test helper commands or the Go test binary pattern to simulate stdout, stderr, zero exit, non-zero exit, startup failure, stdin prompt delivery, and no shell interpolation;
- CLI tests for parsing `--execute`, rejecting `--execute` outside agent-assisted mode, reporting recognized target and resolved command distinctly, reporting runner stdout/stderr/exit code/status, printing full reports before non-zero CLI exit, startup failure errors without exit codes, and preserving existing output when `--execute` is absent for valid recognized targets;
- documentation tests or review checks proving README/docs describe recognized targets, executable local runners, `generic`, startup failures, non-zero exits, and safety boundaries;
- architecture tests proving core imports remain clean, executable mappings remain in core/domain, the local runner adapter receives already-resolved commands, and no write/apply/source-control/workflow/provider abstractions are introduced;
- regression tests for `generate --blank`, `generate --template`, and `generate --guided`.

After implementation, run:

```bash
go test ./...
```

## Architecture Boundaries

The dependency direction must remain:

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

Allowed additions:

- a core-owned `AgentRunner` port;
- domain recognized agent target, executable mapping, resolved command, run request/result/status concepts;
- a concrete local command runner adapter;
- CLI dependency construction, report formatting, and process exit behavior.

Forbidden additions:

- core imports of adapters;
- core imports of `os`, `os/exec`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, or external-agent SDKs;
- provider API clients;
- IDE automation;
- remote execution;
- credential storage;
- write/apply ports;
- patch application;
- source-control automation;
- workflow automation;
- auto-commit;
- auto-push;
- auto-merge.
