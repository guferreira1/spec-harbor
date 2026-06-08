# Proposal: Implement Config Foundation

## Problem

SpecHarbor initializes projects with a local `.specharbor/config.yml`, and the product direction depends on configuration for defaults, validation, review, archive, scan, and output behavior. The CLI currently has a `config` command placeholder, but there is no deterministic way to read and report the current local project configuration.

Configuration will eventually support richer commands and settings, but the first step should be a local, read-only foundation. That foundation must parse the project config through explicit boundaries without mixing filesystem access, YAML parsing, CLI formatting, or business concepts in the same layer.

## Goal

Implement the first config capability:

```text
specharbor config show
```

Also support this alias:

```text
specharbor config
```

The command reads `.specharbor/config.yml` from the current project root, parses YAML, validates that the config version is supported, returns a structured config result, and prints a deterministic human-readable report.

Supported config version:

```text
1
```

Expected config areas:

- defaults;
- validation;
- review;
- archive;
- scan;
- output.

Example supported config:

```yaml
version: 1

defaults:
  agent_role: implementer
  generation_mode: blank

validation:
  require_all_change_files: true

review:
  require_completed_tasks: true

archive:
  date_layout: "2006-01-02"

scan:
  include_common_project_files: true

output:
  format: text
```

Expected report shape for that config:

```text
SpecHarbor configuration loaded.
Path: .specharbor/config.yml
Version: 1

Defaults:
- agent role: implementer
- generation mode: blank

Validation:
- require all change files: true

Review:
- require completed tasks: true

Archive:
- date layout: 2006-01-02

Scan:
- include common project files: true

Output:
- format: text
```

## Scope

- Replace the `config` placeholder with read-only local config behavior.
- Support `specharbor config show`.
- Support `specharbor config` as an alias for `show`.
- Accept no flags for `config` or `config show` in this first version.
- Reject unsupported flags with clear argument errors.
- Reject unsupported config subcommands with clear argument errors.
- Reject extra positional arguments after `show` with clear argument errors.
- Obtain the project root from the current working directory in the CLI adapter.
- Validate that the project root is available before reading config.
- Verify that `.specharbor/config.yml` exists as a file.
- Read `.specharbor/config.yml` through a config-specific filesystem port.
- Parse YAML through a small config parser port when an external YAML library is used.
- Validate that the parsed config version is supported.
- Return clear errors for missing config, unreadable config, invalid YAML or YAML decoding failures, unsupported config version, unsupported flags, unsupported subcommands, and extra arguments.
- Return a structured config result containing the config path and parsed config values.
- Print a human-readable config report from the structured result.
- Keep config concepts and results in `internal/core/domain`.
- Keep config orchestration in `internal/core/usecase`.
- Add config-specific ports under `internal/core/ports`.
- Keep concrete filesystem behavior in `internal/adapters/filesystem`.
- Keep concrete YAML parsing in an adapter package such as `internal/adapters/config`.
- Keep CLI report formatting in `internal/adapters/cli`.
- Add focused tests for domain config concepts, parser behavior, use case orchestration, filesystem adapter compatibility, CLI parsing, CLI report formatting, and existing command regressions.

## Out of Scope

- `specharbor config set`.
- `specharbor config get`.
- `specharbor config unset`.
- Writing config files.
- Modifying config files.
- Updating `.specharbor/config.yml`.
- Updating the config template used by `specharbor init`.
- Changing init behavior.
- Global config outside the project.
- User-level or system-level configuration.
- Config discovery above the current project root.
- Environment variable resolution.
- Secret storage.
- Encryption.
- API key management.
- AI provider setup.
- GitHub or GitLab integration setup.
- Cloud sync.
- Interactive config wizard.
- Changing scan behavior.
- Changing generate, validate, prompt, review, or archive behavior.
- Changing CI behavior.
- Updating README or docs.
- Updating `.github/workflows/ci.yml`.
- Updating the living architecture spec unless implementation proves a concrete architecture change is required.
- Machine-readable output formats such as JSON or YAML.
- Output format selection flags.
- Validation of every future config field.
- Provider, agent, workflow, source-control, cloud, or secret-management abstractions.

## Success Criteria

- Running `specharbor config show` in a project with `.specharbor/config.yml` version `1` prints a deterministic config report and exits zero.
- Running `specharbor config` behaves the same as `specharbor config show`.
- The report includes the config path, version, defaults, validation, review, archive, scan, and output sections in the expected order.
- The use case returns a structured config result rather than formatted text.
- The CLI formats the human-readable report from the structured result.
- Missing project root, missing config file, unreadable config file, invalid YAML, unsupported config version, unsupported flags, unsupported subcommands, and extra arguments return clear errors.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, YAML libraries, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, or external process packages.
- The implementation does not write files, modify config, change init, change other commands, call AI providers, call external agents, call workflow tools, call source-control host APIs, access the network at runtime, run external processes, or require credentials.
- Existing `help`, `version`, `init`, `scan`, `generate`, `prompt`, `validate`, `review`, `archive`, and unknown command behavior is preserved.
- `go test ./...` succeeds.
