# Tasks: Implement Config Foundation

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-config-foundation/`.
- [x] Inspect the current CLI command registry and error flow before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing domain, use case, port, filesystem adapter, parser/adapter, and test patterns before editing.
- [x] Inspect the existing scan, review, archive, and validation implementations for report formatting, dependency wiring, error handling, and filesystem port patterns.
- [x] Inspect the current `.specharbor/config.yml` template under `internal/adapters/templates/defaults/config.yml` and confirm this change does not modify init behavior or the template.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to read-only local config display through `specharbor config show` and the `specharbor config` alias.
- [x] Do not implement `config set`, `config get`, `config unset`, config writes, global config, environment resolution, secret storage, encryption, provider setup, source-control setup, cloud sync, or an interactive wizard.
- [x] Do not change init, scan, generate, validate, prompt, review, archive, README/docs, CI, or `.github/workflows/ci.yml`.
- [x] Do not update the living architecture spec unless implementation proves a concrete architecture change is required and the reason is documented.

## Phase 1: Domain Concepts

- [x] Add config domain concepts under `internal/core/domain`.
- [x] Add a supported local config version constant or helper for version `1`.
- [x] Represent the local config value with version and the supported sections: defaults, validation, review, archive, scan, and output.
- [x] Represent defaults values for `agent_role` and `generation_mode`.
- [x] Represent validation values for `require_all_change_files`.
- [x] Represent review values for `require_completed_tasks`.
- [x] Represent archive values for `date_layout`.
- [x] Represent scan values for `include_common_project_files`.
- [x] Represent output values for `format`.
- [x] Add a structured config result containing the relative config path and parsed local config value.
- [x] Add a small domain helper or method to validate that the config version is supported, if it keeps the use case simple.
- [x] Keep config domain values read-only data concepts; do not consume or change behavior for generation, validation, review, archive, scan, prompt, providers, agents, or workflows in this change.
- [x] Keep domain code free of adapters, CLI packages, ports, `os`, terminal IO, YAML libraries, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, and external process execution.

## Phase 2: Ports

- [x] Add a config-specific filesystem port under `internal/core/ports`.
- [x] Include only the filesystem operations needed by this change: root directory availability checks, config file existence checks, and config file reads.
- [x] Use a behavior-specific name such as `ConfigFileSystem`.
- [x] Add a small config parser port under `internal/core/ports`.
- [x] Use a behavior-specific name such as `ConfigParser`.
- [x] Make the parser port accept config file contents and return `domain.LocalConfig` plus an error.
- [x] Keep config ports separate from initialization, scan, validation, generation, prompt rendering, review, archive, AI provider, workflow dispatcher, source-control, and agent contracts.
- [x] Ensure the config use case depends on config ports instead of `internal/adapters/filesystem`, `internal/adapters/config`, or any YAML library.

## Phase 3: Config Schema

- [x] Define the version `1` YAML schema in the parser adapter DTOs and domain mapping.
- [x] Map `version` to the domain config version.
- [x] Map `defaults.agent_role` to the domain defaults agent role field.
- [x] Map `defaults.generation_mode` to the domain defaults generation mode field.
- [x] Map `validation.require_all_change_files` to the domain validation field.
- [x] Map `review.require_completed_tasks` to the domain review field.
- [x] Map `archive.date_layout` to the domain archive date layout field.
- [x] Map `scan.include_common_project_files` to the domain scan field.
- [x] Map `output.format` to the domain output format field.
- [x] Treat missing or zero `version` as an unsupported config version rather than silently defaulting it to `1`.
- [x] Keep optional sections and fields as zero values when omitted; do not add config default-merging behavior in this change.
- [x] Ignore unknown YAML keys in this first foundation.
- [x] Do not validate semantic values such as agent role names, generation modes, date layouts, or output formats beyond YAML decoding and supported config version.
- [x] Do not read environment variables, expand secrets, load global config, discover parent directories, or merge multiple config sources.

## Phase 4: Config Parser Adapter

- [x] Add a concrete YAML parser adapter under `internal/adapters/config`.
- [x] Keep YAML parser DTOs adapter-local and map them into `domain.LocalConfig`.
- [x] Use a small external YAML library only inside the parser adapter if no existing project dependency already provides YAML parsing.
- [x] Update `go.mod` and `go.sum` if a YAML dependency is introduced.
- [x] Return a clear parse/decode error for syntactically invalid YAML.
- [x] Return a clear parse/decode error for YAML values that cannot decode into the supported field types, such as a string where a boolean is expected.
- [x] Do not perform supported-version validation in the parser adapter unless it delegates to a domain helper without importing use cases.
- [x] Do not read files, write files, print reports, access terminal IO, call AI providers, call external agents, call workflow tools, call source-control APIs, access the network at runtime, run external processes, or install packages at runtime from the parser adapter.

## Phase 5: Config Use Case

- [x] Add a config show use case under `internal/core/usecase`.
- [x] Use a focused name such as `ShowConfig`.
- [x] Make the use case accept the project root as input.
- [x] Validate that the use case value is present.
- [x] Validate that the config filesystem dependency is present.
- [x] Validate that the config parser dependency is present.
- [x] Trim and validate that project root is non-empty.
- [x] Validate project root availability through the config filesystem port, using `DirectoryExists(projectRoot, ".")` or an equivalent config-specific port operation.
- [x] Return a clear execution error when the project root cannot be checked, is unavailable, or is not a directory.
- [x] Use `.specharbor/config.yml` as the only supported config path.
- [x] Verify that `.specharbor/config.yml` exists as a file through the config filesystem port.
- [x] Return a clear missing-config error when `.specharbor/config.yml` is missing or is not a file.
- [x] Read `.specharbor/config.yml` through the config filesystem port.
- [x] Return a clear unreadable-config error when reading fails.
- [x] Parse the config contents through the config parser port.
- [x] Return a clear invalid-YAML/config decoding error when parsing fails.
- [x] Validate that the parsed config version is exactly `1`.
- [x] Return a clear unsupported-version error for missing, zero, negative, or future versions.
- [x] Return a structured domain config result containing path `.specharbor/config.yml` and the parsed config.
- [x] Do not print from the use case.
- [x] Do not call `os`, terminal IO, YAML libraries, provider SDKs, network APIs, source-control APIs, external agents, external processes, or workflow tools from the use case.
- [x] Do not write or modify config files.
- [x] Keep the structure ready for future config commands without adding unused exported registries, factories, provider abstractions, workflow abstractions, source-control abstractions, AI abstractions, secret stores, cloud sync abstractions, or mutation services.

## Phase 6: Filesystem Adapter

- [x] Use `internal/adapters/filesystem` as the concrete implementation of the config filesystem port.
- [x] Ensure the local filesystem adapter can check that the project root is an available directory.
- [x] Ensure the local filesystem adapter can verify that `.specharbor/config.yml` exists as a file rather than a directory.
- [x] Ensure the local filesystem adapter can read `.specharbor/config.yml` contents.
- [x] Preserve existing filesystem adapter behavior for init, scan, generate, validate, prompt, review, and archive.
- [x] Add or update adapter tests only as needed to prove compatibility with the config filesystem port and config read behavior.
- [x] Ensure the filesystem adapter does not contain config path policy, config version policy, YAML parsing, config result construction, CLI report formatting, provider logic, source-control logic, workflow logic, environment resolution, secret handling, or config mutation policy.

## Phase 7: CLI Wiring and Reporting

- [x] Replace the `config` placeholder in `internal/adapters/cli/cli.go`.
- [x] Parse `specharbor config` as an alias for `specharbor config show`.
- [x] Parse `specharbor config show`.
- [x] Reject unsupported flags for both `config` and `config show`.
- [x] Reject unsupported subcommands such as `get`, `set`, `unset`, `list`, `edit`, or any other non-`show` positional command.
- [x] Reject extra positional arguments after `show`.
- [x] Obtain the current working directory for the project root in the CLI adapter.
- [x] Construct the config show use case with the local filesystem adapter and YAML parser adapter.
- [x] Invoke the use case with the project root.
- [x] Print a human-readable config report from the structured result.
- [x] Print the completion line, config path line, version line, and the six sections in order: Defaults, Validation, Review, Archive, Scan, Output.
- [x] Separate the version line and each section with a single blank line.
- [x] Print the relative config path exactly as `.specharbor/config.yml`.
- [x] Print string and boolean values using the labels in `design.md`.
- [x] Match the expected populated report shape from `proposal.md` and `design.md`.
- [x] Return zero when the report is printed successfully.
- [x] Return argument and execution errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping unless a minimal existing error-flow adjustment is strictly required.
- [x] Preserve existing `help`, `version`, `init`, `scan`, `generate`, `prompt`, `validate`, `review`, `archive`, and unknown command behavior.
- [x] Keep the CLI adapter free of YAML parsing, config version validation, config file read policy, config mutation policy, provider logic, source-control logic, workflow logic, environment resolution, and secret handling.

## Phase 8: Tests

- [x] Add domain tests proving the supported config version is `1`.
- [x] Add domain tests proving version validation accepts `1` and rejects missing, zero, negative, and future versions.
- [x] Add domain tests proving config result values preserve path, version, defaults, validation, review, archive, scan, and output values.
- [x] Add parser adapter tests proving the complete conceptual version `1` YAML example parses into the expected domain config.
- [x] Add parser adapter tests proving invalid YAML returns a parse/decode error.
- [x] Add parser adapter tests proving incompatible scalar types return a parse/decode error.
- [x] Add parser adapter tests proving unknown YAML keys are ignored.
- [x] Add parser adapter tests proving omitted optional sections and fields parse as zero values.
- [x] Add use case tests with fake config filesystem and parser dependencies for a successful version `1` config.
- [x] Add use case tests proving the returned result path is `.specharbor/config.yml`.
- [x] Add use case tests proving nil use case, missing filesystem dependency, and missing parser dependency are rejected.
- [x] Add use case tests proving an empty project root is rejected.
- [x] Add use case tests proving an unavailable or non-directory project root returns a clear execution error.
- [x] Add use case tests proving a missing `.specharbor/config.yml` returns a clear missing-config error.
- [x] Add use case tests proving a config path that exists but is not a file returns a clear missing-config or not-a-file error.
- [x] Add use case tests proving an unreadable config returns a clear unreadable-config error.
- [x] Add use case tests proving parser failure returns a clear invalid-YAML/config decoding error.
- [x] Add use case tests proving unsupported versions return a clear unsupported-version error.
- [x] Add use case tests proving the config file is not read when the project root is unavailable.
- [x] Add use case tests proving the parser is not called when the config file is missing or unreadable.
- [x] Add filesystem adapter tests proving the local adapter satisfies the config filesystem port.
- [x] Add filesystem adapter tests proving the local adapter can read `.specharbor/config.yml`.
- [x] Add filesystem adapter tests proving directories are not treated as config files.
- [x] Add CLI tests for `specharbor config show` output matching the expected report shape.
- [x] Add CLI tests for the `specharbor config` alias output matching `config show`.
- [x] Add CLI tests proving unsupported flags are rejected.
- [x] Add CLI tests proving unsupported subcommands are rejected.
- [x] Add CLI tests proving extra arguments after `show` are rejected.
- [x] Add CLI tests proving missing, unreadable, invalid-YAML, and unsupported-version config failures return errors.
- [x] Ensure CLI tests build temporary project directories and do not depend on the real SpecHarbor repository config file or user-local config.
- [x] Preserve or add regression coverage for `help`, `version`, `init`, `scan`, `generate`, `prompt`, `validate`, `review`, `archive`, and unknown command behavior.

## Phase 9: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor config show` prints the expected report and exits zero inside a project with a version `1` `.specharbor/config.yml`.
- [x] Manually verify `specharbor config` behaves the same as `specharbor config show`.
- [x] Manually verify missing config, unreadable config, invalid YAML, unsupported config version, unsupported flags, unsupported subcommands, and extra arguments return clear errors.
- [x] Manually verify the command does not write files, modify config, change init output, change other command behavior, call AI providers, call external agents, call workflow tools, call source-control APIs, access the network at runtime, run external processes, or require credentials.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
