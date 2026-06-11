# Design: Implement Agent-Assisted Spec Authoring

## Overview

This change adds an agent-assisted authoring mode to `specharbor generate`:

```text
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

The first implementation is dry-run-only. It builds a deterministic, copy-pasteable OpenSpec authoring prompt, prints that prompt to stdout with a structured dry-run report, and exits without executing any external agent and without writing any files.

This mode is not an implementation mode. It prepares a prompt that a human can copy into a coding agent to draft or refine only the OpenSpec change package:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

The implementation must preserve the dependency direction:

```text
CLI adapter -> use case -> ports + domain
adapters -> ports
```

Core packages must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, or concrete process execution packages.

`--execute` is explicitly deferred to a future change. In this first version, providing `--execute` must return a clear unsupported flag/mode error and must not render a prompt, run a command, access the network, parse output, apply output, or write files.

## CLI Contract

Supported dry-run command shapes after this change:

```text
specharbor generate <change-id> --blank
specharbor generate <change-id> --template <template-name>
specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type bugfix --title "<title>" --summary "<summary>"
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type docs --title "<title>" --summary "<summary>"
specharbor generate <change-id> --agent-assisted --agent <agent-name> --type refactor --title "<title>" --summary "<summary>"
```

The CLI adapter must:

- parse exactly one change id;
- parse `--agent-assisted` as a generation mode flag;
- require `--agent`, `--type`, `--title`, and `--summary` when `--agent-assisted` is present;
- reject `--agent-assisted` without `--agent`;
- reject `--agent-assisted` without `--type`;
- reject `--agent-assisted` without `--title`;
- reject `--agent-assisted` without `--summary`;
- reject empty agent, type, title, or summary values when representable by the parser;
- reject unknown types with a clear error;
- reject `--agent-assisted` combined with `--blank`;
- reject `--agent-assisted` combined with `--template`;
- reject `--agent-assisted` combined with `--guided`;
- preserve the existing `--blank`, `--template`, and `--guided` parser behavior;
- reject unsupported flags;
- reject `--execute` for agent-assisted generation with a clear unsupported flag/mode error;
- reject extra positional arguments;
- reject duplicate conflicting flags where current parser patterns support duplicate validation;
- preserve existing unsafe change id validation behavior;
- obtain the project root from the current working directory;
- construct concrete dependencies and invoke the agent-assisted authoring use case;
- print a deterministic human-readable report from the structured result.

The CLI adapter may format output, but it must not contain authoring business rules, required-file policy, prompt body policy, execution policy, provider logic, source-control logic, workflow logic, or file application logic.

## Agent-Assisted Inputs

Agent-assisted input should be represented explicitly at the use case boundary. A new use case is preferred over expanding `GenerateChange` because agent-assisted prompt authoring is materially different from file generation.

A possible input shape is:

```text
ProjectRoot
ChangeID
AgentName
AuthoringType
Title
Summary
```

The exact field names may differ, but the use case boundary must not depend on CLI-specific argument structs.

Input validation rules:

- project root is required;
- change id is required and must remain a safe single path segment;
- agent name is required;
- authoring type is required;
- title is required;
- summary is required;
- authoring type must be exactly one of `feature`, `bugfix`, `docs`, or `refactor`;
- mode conflicts must be rejected by the CLI before invoking the use case;
- all required inputs should be trimmed for validation;
- normalized values should be used in the generated prompt.

Agent names should be non-empty and safe to display in reports. Do not require provider API keys or provider-specific credentials. Unknown agent names should not be treated as unknown AI providers. This first version must not resolve agent names to configured commands because local execution is deferred.

## Authoring Types

The supported authoring types intentionally match guided generation:

```text
feature
bugfix
docs
refactor
```

The implementation may reuse `domain.GuidedType` if that keeps validation simple, or introduce a distinct `domain.AgentAssistedType` if that avoids confusing guided starter content with agent-assisted prompt authoring. Either way, the supported values must remain exactly the same four values and must be validated before prompt rendering.

Do not add remote type discovery, provider-backed type discovery, config-backed type selection, marketplace concepts, or custom type registries.

## Required Files

The authoring prompt must list the same required OpenSpec change files as existing generation modes:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

The implementation must reuse `domain.RequiredOpenSpecChangeFiles()` instead of duplicating the required list in CLI or prompt-specific code.

Dry-run must not create the change directory, write these files, or write a prompt file. The prompt may instruct the external agent to draft Markdown for these paths after the user copies the prompt, but SpecHarbor must never parse, apply, or write those drafts in this change.

## Authoring Prompt

The authoring prompt must be deterministic, local, human-readable, and safe to print. It must include:

- project context;
- change id;
- authoring type;
- title;
- summary;
- the required OpenSpec file list from `domain.RequiredOpenSpecChangeFiles()`;
- explicit instruction to create or refine only OpenSpec files under `openspec/changes/<change-id>/`;
- explicit instruction not to implement production code;
- explicit instruction not to modify unrelated files;
- explicit instruction not to change README or docs unless the requested type and scope is documentation;
- explicit instruction to leave implementation tasks unchecked;
- explicit instruction to preserve architecture boundaries;
- explicit instruction that domain code belongs in `internal/core/domain`;
- explicit instruction that ports belong in `internal/core/ports`;
- explicit instruction that use cases belong in `internal/core/usecase`;
- explicit instruction that concrete implementations belong in `internal/adapters`;
- explicit instruction that core must not import adapters;
- explicit instruction that CLI must not contain business rules;
- explicit instruction to run or recommend `specharbor validate <change-id>` when available;
- clear output expectations for Markdown-only OpenSpec content.

The prompt must avoid asking the agent to run implementation, tests, source-control commands, workflow commands, provider setup, credential setup, commits, pushes, merges, deployment, or production code edits.

For non-documentation types, the prompt must explicitly say not to change README or docs. For `docs`, the prompt may allow documentation scope only when the requested title and summary define documentation work, but it must still prohibit production code changes.

The prompt must be suitable for direct copy-paste into the selected agent. It must not depend on a sidecar prompt file, hidden local state, prior terminal output, or generated artifacts that SpecHarbor writes to disk.

## Domain Model

Keep agent-assisted authoring concepts under:

```text
internal/core/domain
```

Expected domain additions:

- an implemented generation mode value for `agent-assisted`, or a clearly named authoring mode value if the implementation separates generation modes from authoring strategies;
- a value object for agent name or agent target if useful;
- supported type validation for `feature`, `bugfix`, `docs`, and `refactor`, preferably reusing existing guided type validation concepts where appropriate;
- an authoring prompt request/result shape.

The result shape should support CLI reporting without formatting in core. Useful fields include:

- change id;
- agent name;
- authoring type;
- title;
- summary;
- required files;
- dry-run status;
- generated prompt;
- execution plan items;
- clear status indicating that no files were written.

Domain code must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, external process execution packages, or concrete template engines.

## Ports

Prompt rendering should go through a core-owned port/interface. Existing prompt rendering patterns can be reused where appropriate, but agent-assisted authoring should not be mixed with role-prompt behavior if that makes the contract unclear.

Possible small ports:

```text
AgentAuthoringPromptTemplate
AgentAuthoringPromptRenderer
```

The exact names may differ, but the prompt port must:

- be consumed by the agent-assisted authoring use case;
- render deterministic prompt content from normalized domain input;
- include every required OpenSpec file;
- return clear errors for missing template content or rendering failures;
- avoid filesystem writes, provider APIs, source-control APIs, workflow APIs, network APIs, terminal IO, and external execution.

Do not add an `AgentRunner` port, local command execution port, workflow connector port, or write/apply port for agent-proposed files in this change. File application, confirmation flows, command output parsing, and post-write validation are deferred.

## Prompt Template Adapter

Provide deterministic authoring prompt template content through an adapter, likely under:

```text
internal/adapters/templates
```

Acceptable implementation choices:

- constants in Go files;
- small local render helpers;
- embedded local template files if they remain repository assets and are deterministic.

The adapter must not load templates from the network, remote registries, provider services, source-control systems, workflow tools, or mutable config.

The template adapter should keep the prompt generic across agents. Agent-specific wording should be minimal and based on the selected agent name only when needed for display. Do not introduce provider-specific SDK payloads or model parameters.

## Deferred Execution and File Application

This first version must not include a concrete local command runner adapter.

Deferred capabilities:

- `--execute` support beyond returning an unsupported error;
- `AgentRunner` or similarly named execution ports;
- local command adapters;
- command stdout/stderr/exit-code capture;
- parsing agent output;
- showing proposed file changes from agent output;
- writing OpenSpec files from agent output;
- writing prompt files;
- write/apply ports;
- confirmation flows;
- post-write validation;
- source-control or workflow automation.

The unsupported `--execute` path must return a clear error and must not run anything. It must not create an execution plan that implies a command could run in this change.

## Agent-Assisted Authoring Use Case

Add orchestration under:

```text
internal/core/usecase
```

Expected dry-run behavior:

- validate required dependencies;
- trim and validate project root;
- trim and validate change id;
- validate safe change id before rendering;
- trim and validate agent name;
- trim and validate authoring type;
- validate title;
- validate summary;
- reject unknown authoring types;
- build the relative change path as `openspec/changes/<change-id>`;
- obtain required filenames from `domain.RequiredOpenSpecChangeFiles()`;
- build a structured execution plan;
- render the deterministic authoring prompt through the prompt port;
- return a structured result with the plan and prompt;
- report that no external commands were executed;
- report that no files were written;
- report that no prompt file was written;
- never parse, apply, or write agent output.

The use case must not include an execute branch in this change. If an execution mode value reaches this boundary, it must return the same clear unsupported error without rendering a prompt, invoking any adapter, or writing files.

The use case must not print output, call `os`, perform terminal IO, call provider SDKs, call network APIs, run external processes directly, access source-control APIs, import adapters, import workflow tools, read config directly, or write files.

## CLI Reporting

Output should be concise, deterministic, and human-readable.

Dry-run success output should identify:

- operation status;
- dry-run status;
- change id;
- agent name;
- authoring type;
- title;
- relative change path;
- required OpenSpec files;
- execution plan;
- generated authoring prompt;
- no external command executed;
- no files written;
- no prompt file written;
- no agent output parsed or applied.

Do not print absolute local paths unless existing CLI conventions require them for an error. Do not print debug output, provider credentials, source-control details, workflow details, network details, or validation claims for files that were not written.

## Testing Strategy

Add focused tests for:

- agent-assisted mode representation;
- supported authoring type validation for `feature`, `bugfix`, `docs`, and `refactor`;
- missing agent validation;
- missing type validation;
- missing title validation;
- missing summary validation;
- unknown type validation;
- unsafe change id validation;
- unsupported `--execute` validation that returns a clear error and runs nothing;
- prompt rendering includes project context, change id, type, title, summary, required files, scope boundaries, architecture boundaries, and "do not implement code" instructions;
- prompt rendering is copy-pasteable and does not depend on a prompt file;
- prompt rendering for non-docs types prohibits README/docs changes;
- prompt rendering for docs type handles documentation scope without permitting production code changes;
- dry-run use case returns a plan and prompt;
- dry-run use case reports no external commands executed;
- dry-run use case reports no files written;
- dry-run use case reports no prompt file written;
- dry-run use case never parses or applies agent output;
- no `AgentRunner`, local command adapter, or write/apply port is introduced;
- CLI accepts successful dry-run command shapes for all four types;
- CLI prints the generated prompt to stdout;
- CLI rejects missing required flags;
- CLI rejects `--execute` as unsupported;
- CLI rejects unknown type;
- CLI rejects unsafe change id;
- CLI rejects mode conflicts with `--blank`, `--template`, and `--guided`;
- CLI rejects unsupported flags and extra positional arguments;
- CLI preserves existing blank, template, and guided generation behavior;
- unrelated commands retain existing behavior.

Use temporary directories and fake prompt renderers where they keep tests deterministic.

## Validation

The Implementer Agent must run:

```text
gofmt
go test ./...
```

Because this first version should not update README, docs outside this OpenSpec change, CI, or config, the final diff must confirm those files are unchanged.
