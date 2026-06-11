# Acceptance Criteria: Implement Config-Driven Templates

## Config Schema

- `.specharbor/config.yml` supports `version: 1` with a `templates.aliases` map.
- Each alias entry supports exactly `source` and `template` for this change.
- Supported source values are exactly `builtin` and `custom`.
- Omitted `templates` or omitted `templates.aliases` is treated as an empty alias map.
- Unsupported `remote`, `local`, `url`, `path`, or arbitrary local path entries fail clearly when used for config-driven template resolution.
- New `init` output includes `version: 1` and an empty `templates.aliases` section without removing existing default config fields.

## Alias and Reference Validation

- Config template aliases are validated as safe single path segments before template filesystem access.
- Invalid aliases include empty values, separators, absolute/path-like values, traversal, `.` or `..`, embedded `..`, leading `.` or `-`, unsupported characters, and values longer than 128 characters.
- `builtin` references validate against the existing built-in template names.
- `custom` references validate against the existing custom template name rules.
- Unsupported source kinds and malformed alias entries return clear config validation errors.

## Command Behavior

- `specharbor generate <change-id> --config-template <alias>` resolves the alias from `.specharbor/config.yml`.
- A `builtin` alias generates the same OpenSpec files as the referenced built-in template.
- A `custom` alias generates the same rendered OpenSpec files as the referenced custom template.
- `--config-template` is mutually exclusive with `--blank`, `--template`, `--custom-template`, `--guided`, `--agent-assisted`, and `--ai-assisted`.
- Same names across built-in templates, custom templates, and config aliases are allowed because flags select disjoint namespaces.
- Missing config, unsupported config version, invalid config, missing alias, invalid alias, unknown source kind, unknown built-in template, missing custom template, missing required custom template files, and empty custom template files each produce clear errors.

## Write Safety

- Generated files are written only under `openspec/changes/<change-id>/`.
- Generated filenames are limited to `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- Existing files are skipped and never overwritten.
- Invalid config or invalid template references write nothing.
- Template content and config values cannot control output paths.
- The command writes no production code, docs, config, CI files, prompt files, archive files, source-control files, or arbitrary paths.

## CLI Output

- Success output includes the change id, config alias, resolved source kind, resolved template name, change path, generated files, skipped files when present, and a safety note that only OpenSpec change files were written.
- Custom-source success output includes the safe project-relative custom template source path.
- Built-in-source success output clearly identifies the source as built-in without exposing absolute local paths.
- Error output identifies whether the failure came from config loading, alias validation, alias lookup, source validation, built-in template validation, custom template loading, or write behavior.

## Documentation

- `README.md`, `docs/usage.md`, and `docs/generation-modes.md` document config-driven templates as implemented behavior.
- Documentation includes `.specharbor/config.yml` examples, source kinds, command usage, disjoint namespace behavior, validation expectations, write behavior, and safety boundaries.
- Documentation states that remote templates, marketplace templates, arbitrary local paths, template script execution, shell execution, network/provider behavior, production code writes, and source-control automation are not implemented.

## Tests and Regression

- Domain tests cover alias validation, source validation, reference validation, unsupported source behavior, and defensive copying.
- Use case tests cover built-in alias generation, custom alias generation, missing config, invalid config, missing alias, invalid alias, unsupported source, unknown built-in template, missing custom template, missing required custom files, write boundaries, and no production code writes.
- Adapter tests cover config reads from `.specharbor/config.yml`, invalid YAML, invalid template alias shape, safe template reads, path traversal rejection, no external command execution, and no network calls.
- CLI tests cover `--config-template` parsing, missing value, invalid alias, error output, success output, conflicts with every other generation mode, unsupported flags, and extra arguments.
- Regression tests prove blank, built-in template, custom template, guided, agent-assisted, AI-assisted, validate, prompt, review, archive, workflow, scan, and config behavior remain unchanged except for explicitly documented config-default or config-show updates.
- Architecture tests confirm core does not import adapters, CLI packages, `os`, provider/network/source-control/workflow/agent SDKs, shell execution, or script execution packages.

## Verification

- `go test ./...` passes.
- `go run ./cmd/specharbor validate implement-config-driven-templates` passes for this OpenSpec change.
- Implementation updates `tasks.md` only for work actually completed.
