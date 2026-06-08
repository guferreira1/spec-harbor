# Acceptance Criteria: Implement Config Foundation

- Running `specharbor config show` inside a project with `.specharbor/config.yml` version `1` returns a structured config result.
- Running `specharbor config` behaves the same as `specharbor config show`.
- The structured result contains the relative config path `.specharbor/config.yml`.
- The structured result contains version, defaults, validation, review, archive, scan, and output config values.
- The CLI obtains the project root from the current working directory.
- The project root is validated before reading config.
- An empty project root returns a clear execution error.
- An unavailable, unreadable, or non-directory project root returns a clear execution error.
- The use case verifies that `.specharbor/config.yml` exists as a file before reading.
- A missing `.specharbor/config.yml` returns a clear missing-config error.
- A `.specharbor/config.yml` path that exists but is not a file returns a clear missing-config or not-a-file error.
- The use case reads `.specharbor/config.yml` through a config-specific filesystem port.
- An unreadable `.specharbor/config.yml` returns a clear unreadable-config error.
- YAML parsing is performed behind a small config parser port when an external YAML library is used.
- The concrete YAML parser lives in an adapter package such as `internal/adapters/config`.
- Invalid YAML returns a clear invalid-YAML/config decoding error.
- YAML values that cannot decode into supported field types return a clear invalid-YAML/config decoding error.
- Unknown YAML keys are ignored in this first foundation.
- Omitted optional sections and fields parse as zero values.
- The `version` field is required by behavior; missing or zero version is rejected as unsupported.
- Supported config version is exactly `1`.
- Unsupported config versions, including negative and future versions, return a clear unsupported-version error.
- Version `1` supports `defaults.agent_role`.
- Version `1` supports `defaults.generation_mode`.
- Version `1` supports `validation.require_all_change_files`.
- Version `1` supports `review.require_completed_tasks`.
- Version `1` supports `archive.date_layout`.
- Version `1` supports `scan.include_common_project_files`.
- Version `1` supports `output.format`.
- The first implementation does not validate semantic values for agent roles, generation modes, date layouts, or output formats beyond YAML decoding and supported version.
- The CLI prints this report shape for the conceptual example:

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

- The CLI always prints the completion line, path line, version line, and all six config sections in order: Defaults, Validation, Review, Archive, Scan, Output.
- The version line and each section are separated by a single blank line.
- The `Path:` line prints `.specharbor/config.yml`, not an absolute local path.
- The report does not print raw YAML, debug output, secrets, environment variables, provider SDK details, source-control details, workflow details, or future metadata fields.
- `specharbor config show` exits zero when the report is printed successfully.
- `specharbor config` exits zero when the report is printed successfully.
- Running `specharbor config --json` or any other flag returns a clear unsupported flag error.
- Running `specharbor config show --json` or any other flag returns a clear unsupported flag error.
- Running `specharbor config show extra` returns a clear unexpected argument error.
- Running `specharbor config get`, `specharbor config set`, `specharbor config unset`, or any other unsupported config subcommand returns a clear unsupported subcommand error.
- Filesystem checks and reads are performed through a config-specific port owned by `internal/core/ports`.
- YAML parsing is performed through a parser port owned by `internal/core/ports`.
- Concrete filesystem behavior lives in `internal/adapters/filesystem`.
- Concrete YAML parsing lives in `internal/adapters/config` or another clearly named adapter package.
- Config orchestration lives in `internal/core/usecase`.
- Config concepts and the structured config result live in `internal/core/domain`.
- CLI argument parsing, current-working-directory lookup, dependency construction, and human-readable report formatting live in the CLI adapter.
- CLI report formatting does not live in the use case.
- Config path policy, supported-version validation, and structured result construction do not live in the filesystem adapter.
- YAML parsing does not live in the CLI adapter, use case, domain package, or filesystem adapter.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, YAML libraries, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, or external process packages.
- The domain package does not import ports.
- The implementation does not write config files.
- The implementation does not modify config files.
- The implementation does not update the config template used by init.
- The implementation does not change init behavior.
- The implementation does not implement `config set`, `config get`, `config unset`, config listing, config editing, or config validation commands.
- The implementation does not add global config, user config, system config, parent-directory discovery, environment variable resolution, secret storage, encryption, API key management, provider setup, source-control integration setup, workflow integration setup, cloud sync, or an interactive config wizard.
- The implementation does not change scan, generate, validate, prompt, review, or archive behavior.
- The implementation does not update README/docs.
- The implementation does not modify `.github/workflows/ci.yml`.
- The implementation does not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, source-control host APIs, network APIs, or external processes at runtime.
- The implementation does not require provider API keys, local model credentials, agent credentials, source-control credentials, workflow credentials, or cloud credentials.
- The implementation does not add unused exported config registries, factories, provider abstractions, workflow abstractions, source-control abstractions, AI abstractions, secret stores, cloud sync abstractions, or mutation services.
- Existing `help`, `version`, `init`, `scan`, `generate`, `prompt`, `validate`, `review`, `archive`, and unknown command behavior is preserved.
- Focused tests cover domain config concepts, supported-version validation, parser adapter behavior, use case orchestration, filesystem adapter compatibility, CLI parsing, CLI reporting, CLI exit behavior, and existing command regressions.
- Tests do not depend on the real SpecHarbor repository config file or user-local config.
- `go test ./...` succeeds.
