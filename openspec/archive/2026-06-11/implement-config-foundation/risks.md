# Risks: Implement Config Foundation

## Architecture leakage

Config touches CLI parsing, current-working-directory lookup, filesystem checks, file reading, YAML parsing, version validation, structured result construction, and report formatting. The main risk is collapsing these responsibilities into the CLI adapter or parser adapter.

Mitigation:

- Keep CLI responsibilities limited to argument parsing, current-working-directory lookup, dependency construction, and report formatting.
- Keep config orchestration in `internal/core/usecase`.
- Keep config concepts and structured results in `internal/core/domain`.
- Use config-specific ports from `internal/core/ports`.
- Keep concrete filesystem behavior in `internal/adapters/filesystem`.
- Keep concrete YAML parsing in `internal/adapters/config` or another config adapter package.
- Keep YAML parsing, config path policy, and supported-version validation out of the CLI formatter.

## YAML dependency leakage

Go does not provide YAML parsing in the standard library, so the implementation will likely need an external YAML package. The risk is importing that package directly into the core or spreading YAML DTOs across layers.

Mitigation:

- Put the YAML library behind a small config parser port.
- Import the YAML library only from the concrete parser adapter.
- Keep parser DTOs adapter-local.
- Map parser DTOs into domain config values before returning to the use case.
- Add tests proving use case behavior with a fake parser, not the concrete YAML library.

## Unsupported old config shape

The current init template may not match the new version `1` conceptual schema. Changing init behavior is out of scope, so projects initialized with an older unversioned config may receive an unsupported-version error.

Mitigation:

- Document that missing or zero version is rejected as unsupported.
- Keep this change focused on the read-only versioned config foundation.
- Do not silently treat missing version as `1`.
- Leave init template updates or migration behavior for a future OpenSpec change.
- Test `config show` with explicit temporary version `1` config files instead of relying on the current template.

## Over-validating early config values

It is tempting to validate agent roles, generation modes, output formats, and archive date layouts immediately. That can create policy before the settings are consumed and may conflict with future behavior.

Mitigation:

- Validate only YAML decoding and supported config version in this first foundation.
- Store supported fields as read-only config values.
- Add semantic validation later when a command actually consumes a setting.
- Keep unknown YAML keys ignored for now.

## Under-validating required version

Treating missing version as an implicit default would make future migrations ambiguous and weaken the config contract.

Mitigation:

- Require `version` by behavior.
- Treat missing, zero, negative, and future versions as unsupported.
- Include exact tests for missing and unsupported versions.
- Report supported version `1` clearly in errors.

## Ambiguous filesystem errors

The command must distinguish a missing config from an unreadable config and from an unavailable project root. Mapping every filesystem issue to the same error would make the command hard to diagnose.

Mitigation:

- Validate project root availability before checking the config path.
- Check that `.specharbor/config.yml` exists as a file before reading.
- Return a missing-config or not-a-file error when the config path is absent or not a file.
- Return an unreadable-config error when the file exists but cannot be read.
- Preserve underlying errors through wrapping where useful.
- Cover each error path in use case tests.

## Report formatting drift

Future config fields could make the report noisy or unstable if formatting is generated from map iteration or raw YAML order.

Mitigation:

- Print fields in a fixed section and field order.
- Format from the structured domain result, not from raw YAML nodes.
- Avoid maps for report order.
- Add CLI tests asserting the expected report shape.

## Overbuilding future config commands

Config will eventually need get, set, unset, validation, defaults, global scopes, env resolution, secrets, and provider setup. Adding registries, mutation services, secret stores, or provider abstractions now would create unused surface area.

Mitigation:

- Implement only `config show` and the `config` alias.
- Add only a read-only use case, a filesystem port, a parser port, domain config values, and CLI formatting.
- Reject all unsupported subcommands and flags.
- Do not add unused exported registries, factories, provider abstractions, workflow abstractions, source-control abstractions, AI abstractions, secret stores, cloud sync abstractions, or mutation services.

## Accidental behavior changes

The config command lives in the same CLI registry and filesystem adapter used by existing commands. Broad refactors could change init, scan, generate, validate, prompt, review, or archive behavior.

Mitigation:

- Replace only the `config` placeholder.
- Reuse existing filesystem adapter methods where possible.
- Keep helper extraction narrow and behavior-preserving.
- Do not modify init templates, README/docs, CI, or unrelated OpenSpec changes.
- Preserve existing tests and add focused regression coverage for existing commands.
- Run `go test ./...`.

## Credential exposure

Configuration is adjacent to provider setup, API keys, secrets, and environment variables. Even a read-only command could accidentally print sensitive fields if it dumps raw YAML or old config sections.

Mitigation:

- Print only the supported version `1` fields listed in this change.
- Do not print raw YAML.
- Do not print unknown keys.
- Do not resolve or print environment variable values.
- Do not add provider, API key, secret, cloud, source-control, or workflow settings to the report in this change.

## Runtime external effects

Parser installation during development is different from runtime behavior. The config command itself must remain local and deterministic.

Mitigation:

- Allow dependency updates during implementation only through normal Go module management.
- Do not download, install, or execute anything at runtime.
- Do not access the network at runtime.
- Do not call AI providers, external agents, workflow tools, source-control APIs, or external processes.
- Keep tests local and deterministic.
