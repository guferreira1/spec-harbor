# Design: Implement Prompt Command

## Overview

`specharbor prompt` renders a role-specific Markdown prompt for a target OpenSpec change.

The command is part of the agent-assisted workflow. It prepares text that a user can pass to an external coding agent, but it must not call that agent or any AI provider.

## CLI Contract

Supported command shape:

```text
specharbor prompt <change-id> --role <role>
```

The implementation may also accept `--role <role>` before the change id if that falls naturally out of a simple parser, but the required documented form is the change id followed by `--role`.

Required roles:

```text
spec-author
architecture-reviewer
implementer
test-engineer
change-reviewer
```

On success, stdout must contain only the rendered prompt. Do not add banners, summaries, debug lines, or absolute local paths around the prompt.

On failure, the command should return an error for the process entrypoint to handle. Error messages should be concise and should identify the invalid or missing input.

## Prompt Data

The use case input should include:

- project root;
- change id;
- role.

The rendering data must include at least:

- `change_id`

The existing role templates also contain `{{task}}`. This change does not require a CLI task argument. The implementation should render `task` to a deterministic default instruction that tells the agent to read and follow the active change, for example:

```text
Read the active OpenSpec change and perform the role-specific work described there.
```

Leaving raw `{{task}}` text in successful output is not acceptable.

## Core Use Case

Add a prompt rendering use case under:

```text
internal/core/usecase
```

The use case should:

- validate that project root, change id, and role are non-empty after trimming whitespace;
- validate that the role is one of the supported role identifiers;
- request the role template through a small port;
- render placeholders through a small port or a composed prompt template port;
- return a structured result containing the rendered prompt;
- return errors instead of panicking.

The use case must not:

- import concrete adapters;
- call `os`, `filepath`, terminal IO, network APIs, provider SDKs, or external agent tooling;
- print user-facing output;
- contain CLI argument parsing.

## Ports

Add only the ports required by the prompt use case under:

```text
internal/core/ports
```

Expected small contracts:

- a role template repository or loader that can fetch a template for a role from a project root;
- a template renderer that can render named placeholders from string data.

If a single prompt template port is simpler and remains small, it may combine loading and rendering behind behavior-specific methods. Avoid broad interfaces that mix prompt rendering with spec generation, validation, AI providers, workflow dispatch, or initialization.

## Domain Values

Represent supported prompt roles as a small domain value or use-case-owned value list.

The role identifiers are stable CLI/API values:

```text
spec-author
architecture-reviewer
implementer
test-engineer
change-reviewer
```

Do not add provider, agent-target, workflow connector, or spec-generation abstractions unless they are directly needed for this command.

## Template Adapter

Add or extend an adapter under a clear package such as:

```text
internal/adapters/templates
```

The adapter should load role templates from:

```text
agent-prompts/roles/<role>.md.tmpl
```

For this change, repository-local template loading is sufficient because the command is defined around the existing project templates. If the implementation embeds the templates with `go:embed`, it must still preserve the same role template contents and behavior.

Rendering must support `{{change_id}}` and `{{task}}`. A standard library template engine is preferred over ad hoc string replacement if it fits the existing template syntax cleanly. If a lightweight explicit replacement implementation is chosen, tests must cover unknown placeholders and required placeholders.

Template loading errors should distinguish an unsupported role from a missing template file when practical.

## CLI Adapter

Update `internal/adapters/cli` so the `prompt` command:

- parses exactly one change id argument;
- parses `--role <role>`;
- obtains the current working directory as project root;
- constructs the prompt use case with concrete template adapters;
- prints only the rendered prompt to stdout on success;
- returns errors without panicking.

The CLI adapter must not know the template file naming policy beyond adapter construction. It may contain only argument parsing and output formatting.

`cmd/specharbor/main.go` should remain limited to process bootstrapping unless minimal error handling changes are required.

## Validation Rules

The command should reject:

- `specharbor prompt`
- `specharbor prompt <change-id>`
- `specharbor prompt --role implementer`
- `specharbor prompt <change-id> --role`
- `specharbor prompt <change-id> --role unknown`
- extra positional arguments
- unsupported flags

The command does not need to verify that `openspec/changes/<change-id>/` exists in this change. It renders a prompt for the provided id and leaves deeper OpenSpec validation to future validation behavior.

## Testing Strategy

Add focused tests for:

- use case renders each supported role;
- use case rejects empty change id;
- use case rejects empty role;
- use case rejects unsupported role;
- rendered output replaces `{{change_id}}`;
- rendered output does not leave raw `{{task}}`;
- template adapter loads every required role template;
- missing template errors are returned;
- CLI prints only the rendered prompt on success;
- CLI rejects missing change id, missing role, unsupported role, extra arguments, and unsupported flags;
- existing `help`, `version`, `init`, and unknown command behavior remains intact.

Use temporary directories or fake ports where they keep tests focused and deterministic.

## Validation

The Implementer Agent must run:

```text
gofmt
go test ./...
```
