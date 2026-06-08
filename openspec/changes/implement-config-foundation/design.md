# Design: Implement Config Foundation

## Overview

`specharbor config show` reads the local SpecHarbor project configuration from `.specharbor/config.yml` and prints a deterministic human-readable report.

This change is read-only. It establishes the domain model, parser boundary, filesystem boundary, use case, and CLI report shape needed for future config commands without implementing mutation, global config, environment resolution, secrets, provider setup, cloud sync, or interactive configuration.

The command must support only the local project config at this path:

```text
.specharbor/config.yml
```

The CLI obtains the project root from the current working directory. The use case validates that the project root is available, verifies the config path exists as a file, reads the file, parses YAML through a parser port, validates the supported config version, and returns a structured result. The CLI formats that result.

## CLI Contract

Supported command shapes:

```text
specharbor config show
specharbor config
```

`specharbor config` is an alias for `specharbor config show`.

Reject:

- `specharbor config show extra`
- `specharbor config get`
- `specharbor config set`
- `specharbor config unset`
- `specharbor config list`
- any other unsupported subcommand
- any flag such as `--json`, `--yaml`, `--format`, `--path`, `--global`, `--set`, `--get`, `--env`, `--secrets`, or `--interactive`

Argument parsing rules:

- If no arguments follow `config`, run `show`.
- If the first argument is `show`, run `show` only when no extra argument follows.
- If any argument starts with `-`, return `unsupported flag: <arg>`.
- If the first non-flag argument is not `show`, return `unsupported config subcommand: <arg>`.
- If `show` receives an extra positional argument, return `unexpected argument: <arg>`.

On success, print the config report to stdout and return zero. On argument or execution errors, return an error from the CLI adapter so `cmd/specharbor/main.go` can use the existing error flow.

## Expected CLI Output

For this config:

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

The report must follow this shape:

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

Output rules:

- Always print the completion line, path line, version line, and all six config sections in this order: Defaults, Validation, Review, Archive, Scan, Output.
- Separate the version line and each section with a single blank line.
- Print the relative config path exactly as `.specharbor/config.yml`.
- Do not print absolute local paths.
- Do not print debug output, raw YAML, secrets, environment variables, provider names, agent SDK details, source-control details, workflow details, or future metadata fields.
- Do not add JSON, YAML, table, or machine-readable output formats in this change.

## Config Schema

Supported version:

```text
1
```

Version `1` has these supported fields:

| YAML path | Type | Report label |
| --- | --- | --- |
| `version` | integer | `Version:` |
| `defaults.agent_role` | string | `agent role` |
| `defaults.generation_mode` | string | `generation mode` |
| `validation.require_all_change_files` | boolean | `require all change files` |
| `review.require_completed_tasks` | boolean | `require completed tasks` |
| `archive.date_layout` | string | `date layout` |
| `scan.include_common_project_files` | boolean | `include common project files` |
| `output.format` | string | `format` |

The first foundation validates only the supported config version and YAML decoding. It does not validate whether `agent_role`, `generation_mode`, `date_layout`, or `output.format` are semantically allowed values. Those value validations can be added when the settings are consumed by future behavior.

Unknown YAML keys are ignored in this first version. This keeps the parser focused on the fields this command reports and avoids adding broad schema validation before there is behavior attached to every future setting.

If an optional section or field is omitted, the parsed domain value uses the Go zero value and the report remains deterministic. The `version` field is required by behavior because missing version parses as zero and is rejected as an unsupported config version.

The existing init template is not changed by this OpenSpec change. If a project contains an older unversioned config, `specharbor config show` should return a clear unsupported config version error until a future change updates init defaults or migration behavior.

## Domain Model

Add config domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- a supported local config version constant or helper;
- a local config value with version and supported sections;
- section values for defaults, validation, review, archive, scan, and output;
- a structured config result containing the config path and local config value;
- a small validation helper for supported config versions if it keeps use case code clear.

A possible result shape:

```text
Path   string
Config LocalConfig
```

A possible local config shape:

```text
Version    int
Defaults   ConfigDefaults
Validation ConfigValidation
Review     ConfigReview
Archive    ConfigArchive
Scan       ConfigScan
Output     ConfigOutput
```

With section values such as:

```text
ConfigDefaults:
  AgentRole      string
  GenerationMode string

ConfigValidation:
  RequireAllChangeFiles bool

ConfigReview:
  RequireCompletedTasks bool

ConfigArchive:
  DateLayout string

ConfigScan:
  IncludeCommonProjectFiles bool

ConfigOutput:
  Format string
```

The domain package must not import adapters, CLI packages, ports, `os`, terminal IO, YAML libraries, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, or external process packages.

Do not reuse provider, agent, workflow, source-control, scan, archive, review, validation, generation, or prompt domain types as config storage unless those values are actually consumed by behavior in this change. Strings are acceptable for this read-only foundation.

## Ports

Add config-specific ports under:

```text
internal/core/ports
```

Expected filesystem contract:

```text
DirectoryExists(root string, relativePath string) (bool, error)
FileExists(root string, relativePath string) (bool, error)
ReadFile(root string, relativePath string) (string, error)
```

Use a behavior-specific name such as `ConfigFileSystem`.

Expected parser contract:

```text
ParseLocalConfig(contents string) (domain.LocalConfig, error)
```

Use a small name such as `ConfigParser`.

The parser port exists so the use case does not import a YAML library. The parser adapter may import `internal/core/domain` to return domain config values.

Do not reuse initialization, scan, validation, generation, prompt, review, archive, AI provider, workflow dispatcher, source-control, or agent contracts directly, even if the local filesystem adapter can satisfy overlapping methods.

## Parser Adapter

Add a concrete YAML parser adapter under:

```text
internal/adapters/config
```

The adapter should:

- parse YAML into a small adapter-local DTO with YAML tags;
- map the DTO into `domain.LocalConfig`;
- return clear parse/decode errors for syntactically invalid YAML or YAML type mismatches;
- ignore unknown keys;
- avoid semantic business validation beyond YAML decoding;
- avoid reading files, writing files, terminal IO, network access, provider SDKs, external agents, workflow SDKs, source-control SDKs, or external process execution.

If a YAML dependency is introduced, keep it narrow and update `go.mod` and `go.sum` as part of the implementation. The use case and domain must remain independent of that dependency.

## Use Case

Add a config show use case under:

```text
internal/core/usecase
```

Recommended names:

- `ShowConfig`
- `ShowConfigInput`

Expected input:

- project root.

Expected behavior:

- validate that the use case value is present;
- validate that the config filesystem dependency is present;
- validate that the config parser dependency is present;
- trim and validate that project root is non-empty;
- validate project root availability through `DirectoryExists(projectRoot, ".")`;
- return a clear execution error if the project root cannot be checked, is unavailable, or is not a directory;
- use `.specharbor/config.yml` as the only supported config path;
- verify that `.specharbor/config.yml` exists as a file through the config filesystem port;
- return a clear missing-config error when the config path is missing or is not a file;
- read `.specharbor/config.yml` through the config filesystem port;
- return a clear unreadable-config error when reading fails;
- parse the file contents through the config parser port;
- return a clear invalid-YAML/config decoding error when parsing fails;
- validate that `config.Version == 1`;
- return a clear unsupported-version error for missing, zero, negative, or future versions;
- return `domain.ConfigResult` with path `.specharbor/config.yml` and the parsed config;
- never print, call `os`, access terminal IO, import adapters, import YAML libraries, call provider APIs, call source-control APIs, run external tools, write files, or import workflow SDKs.

The use case should not apply environment variables, merge global config, discover config in parent directories, resolve secrets, or mutate config.

## Filesystem Adapter

Use `internal/adapters/filesystem` as the concrete implementation of the config filesystem port.

The local filesystem adapter already contains overlapping operations used by other commands. The implementation should ensure it satisfies the config port without adding config-specific policy to the filesystem adapter.

The adapter must not know:

- `.specharbor/config.yml` is the config path;
- which config version is supported;
- how YAML is parsed;
- how config values are reported;
- future config mutation, provider setup, secret, environment, cloud, source-control, or workflow policy.

## CLI Adapter

Update `internal/adapters/cli` so the `config` command:

- replaces the existing not-implemented placeholder;
- parses `specharbor config` as the `show` alias;
- parses `specharbor config show`;
- rejects unsupported flags;
- rejects unsupported subcommands;
- rejects extra positional arguments after `show`;
- obtains the current working directory as the project root;
- constructs the config show use case with the local filesystem adapter and YAML parser adapter;
- invokes the use case with the project root;
- prints the human-readable config report from the structured result;
- returns argument and execution errors without panicking.

The CLI adapter may format the report, but it must not contain YAML parsing, config version validation, filesystem policy, provider logic, source-control logic, workflow logic, secret handling, or environment resolution.

`cmd/specharbor/main.go` should remain limited to the existing process bootstrap unless a minimal error-flow adjustment is strictly required.

## Testing Strategy

Add focused tests for:

- domain config result and section value construction;
- supported config version validation;
- parser adapter parsing the complete version `1` example;
- parser adapter rejecting invalid YAML;
- parser adapter returning a decoding error for incompatible scalar types such as a string where a boolean is expected;
- parser adapter ignoring unknown keys;
- use case success for a readable `.specharbor/config.yml` with version `1`;
- use case returning the relative path `.specharbor/config.yml`;
- use case rejecting nil dependencies;
- use case rejecting empty project root;
- use case returning a clear error when the project root is unavailable;
- use case returning a clear error when `.specharbor/config.yml` is missing or is not a file;
- use case returning a clear error when reading config fails;
- use case returning a clear error when YAML parsing fails;
- use case returning a clear error when version is missing, zero, negative, or not `1`;
- filesystem adapter compatibility with the config filesystem port;
- CLI success output for `specharbor config show`;
- CLI alias output for `specharbor config`;
- CLI rejection of unsupported flags;
- CLI rejection of unsupported subcommands;
- CLI rejection of extra arguments after `show`;
- existing `help`, `version`, `init`, `scan`, `generate`, `prompt`, `validate`, `review`, `archive`, and unknown command behavior.

CLI tests should use temporary project directories with controlled `.specharbor/config.yml` contents. They must not depend on the real SpecHarbor repository config file or on user-local configuration.
