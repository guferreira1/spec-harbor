# Tasks: Implement Config-Driven Templates

## Phase 1 - Domain model

- [x] Add a `ConfigTemplateAlias` value object in `internal/core/domain` with validation for non-empty values, maximum length 128, a single safe path segment, no `/` or `\`, no absolute paths, no `.` or `..`, no `..` sequence, no leading `.` or `-`, and allowed characters `[A-Za-z0-9._-]`.
- [x] Add a config template source kind model supporting exactly `builtin` and `custom`.
- [x] Add a config template reference model that validates `builtin` references with `TemplateName` rules and `custom` references with `CustomTemplateName` rules.
- [x] Add a domain model for configured template aliases that owns lookup behavior and defensively copies maps or slices where applicable.
- [x] Define clear domain errors for invalid alias names, unsupported source kinds, missing source, missing template, and unsupported remote/path fields.

## Phase 2 - Domain tests

- [x] Test valid aliases including kebab-case, snake_case, mixed alphanumeric names, and internal single dots.
- [x] Test invalid aliases: empty, whitespace-only, forward slash, backslash, absolute path, Windows drive path, path traversal, single dot, double dot, embedded `..`, leading dot, leading dash, spaces, colon, shell metacharacters, template delimiters, and over-length values.
- [x] Test source type validation accepts only `builtin` and `custom`.
- [x] Test `builtin` reference construction accepts supported built-in template names and rejects unknown names with clear errors.
- [x] Test `custom` reference construction accepts safe custom template names and rejects unsafe custom template names with clear errors.
- [x] Test unsupported source type behavior for `remote`, `local`, URL-like values, and empty values.
- [x] Test unsupported remote/path fields are rejected for requested aliases.
- [x] Test alias lookup returns a clear missing-alias error.
- [x] Test defensive copying for alias maps, references, and result slices where applicable.

## Phase 3 - Config parsing and defaults

- [x] Extend the local config domain model to include configured template aliases without changing unrelated config behavior.
- [x] Extend the YAML config parser to decode `version: 1` plus `templates.aliases`.
- [x] Preserve existing parser behavior for supported unrelated config fields unless this change explicitly tightens a template-specific shape.
- [x] Return clear errors for invalid YAML, invalid alias map shape, invalid alias entries, missing required `source`, missing required `template`, unsupported source kinds, and unsupported remote/path fields.
- [x] Treat omitted `templates` and omitted `templates.aliases` as an empty alias set.
- [x] Require supported config version `1` for config-driven template generation.
- [x] Update embedded default `.specharbor/config.yml` so new `init` output includes `version: 1` and `templates.aliases: {}` while preserving existing defaults.
- [x] Decide during implementation whether `config show` should list template aliases; if yes, add output, docs, and tests. If no, keep `config show` behavior unchanged except for parser compatibility.

## Phase 4 - Ports and adapters

- [x] Reuse or extend the existing core-owned config read/parser ports so the generation use case can load `.specharbor/config.yml`.
- [x] Keep filesystem reads in `internal/adapters/filesystem` and YAML decoding in the config adapter or existing platform location.
- [x] Ensure config file reads target `.specharbor/config.yml` under the project root.
- [x] Add adapter tests for missing config, unreadable config, invalid YAML, invalid alias YAML shape, omitted templates, and valid alias decoding.
- [x] Add tests that template file reads remain under `.specharbor/templates/<template-name>/` for `custom` references.
- [x] Add tests that path traversal and absolute/path-like references are rejected before template filesystem access.
- [x] Confirm no adapter path performs network calls or external command execution.

## Phase 5 - Use case orchestration

- [x] Add config-template generation input fields without changing direct built-in/custom generation behavior.
- [x] Validate change id and config alias before loading or accessing referenced custom template files.
- [x] Load `.specharbor/config.yml` only when `--config-template` mode is requested.
- [x] Resolve the requested alias from the parsed config.
- [x] Delegate `builtin` aliases to the existing built-in template generation behavior.
- [x] Delegate `custom` aliases to the existing custom template loading, rendering, and write behavior.
- [x] Pass optional `--title` and `--summary` through to custom-template rendering when the config alias resolves to `custom`.
- [x] Ignore optional title/summary for `builtin` aliases or reject them if implementation chooses the stricter command contract; document and test the chosen behavior.
- [x] Return a structured generation result that includes change id, config alias, resolved source kind, resolved template name, safe custom source path when applicable, created files, skipped files, and change path.
- [x] Ensure invalid config, missing alias, invalid alias, unsupported source, unknown built-in template, missing custom template, and missing required custom template files produce clear errors.
- [x] Ensure invalid alias fails before custom template filesystem access.
- [x] Ensure generated files are written only under `openspec/changes/<change-id>/`.
- [x] Ensure no production code, docs, config, CI, source-control, prompt, archive, or arbitrary output files are written by generation.

## Phase 6 - Use case tests

- [x] Generate from a config alias pointing to a built-in template.
- [x] Generate from a config alias pointing to a custom template.
- [x] Missing config returns a clear error when `--config-template` is used.
- [x] Invalid config returns a clear error and writes nothing.
- [x] Missing alias returns a clear error and writes nothing.
- [x] Invalid alias fails before filesystem/template access.
- [x] Unknown source type returns a clear error and writes nothing.
- [x] Unknown built-in template returns a clear error and writes nothing.
- [x] Missing custom template returns a clear error and writes nothing.
- [x] Missing required custom template files return a clear error and write nothing.
- [x] Empty custom template files keep existing custom template error behavior.
- [x] Same text across built-in template name, custom template name, and config alias remains disambiguated by flag.
- [x] Generated files are written only under `openspec/changes/<change-id>/`.
- [x] Existing files are skipped and not overwritten.
- [x] Existing built-in and custom generation behavior remains unchanged when direct flags are used.
- [x] Blank, guided, agent-assisted, and AI-assisted generation behavior remains unchanged.
- [x] No production code writes occur.

## Phase 7 - CLI

- [x] Parse `--config-template <alias>` with required value and duplicate-flag errors in the existing CLI style.
- [x] Reject `--config-template` with `--blank`.
- [x] Reject `--config-template` with `--template`.
- [x] Reject `--config-template` with `--custom-template`.
- [x] Reject `--config-template` with `--guided`.
- [x] Reject `--config-template` with `--agent-assisted`.
- [x] Reject `--config-template` with `--ai-assisted`.
- [x] Reject or clearly handle mode-specific flags with `--config-template`: `--type`, `--agent`, `--from-file`, `--execute`, and `--overwrite`.
- [x] Accept optional `--title` and `--summary` with `--config-template` only if the implementation follows the preferred pass-through behavior, with duplicate and missing-value errors.
- [x] Wire the config-template path through CLI dependency construction without adding business rules to CLI code.
- [x] Print success output with change id, config alias, resolved source kind, resolved template name, generated files, skipped files, safe custom source path when applicable, and the safety note that only OpenSpec change files were written.
- [x] Print clear errors for missing config, invalid config, missing alias, invalid alias, unsupported source, unknown built-in template, missing custom template, and missing required custom template files.

## Phase 8 - CLI and regression tests

- [x] Test `--config-template` parsing with required alias value.
- [x] Test invalid alias error output.
- [x] Test missing config error output.
- [x] Test missing alias error output.
- [x] Test source-specific error output for unsupported source, unknown built-in template, missing custom template, and missing custom template files.
- [x] Test conflicts with `--blank`, `--template`, `--custom-template`, `--guided`, `--agent-assisted`, and `--ai-assisted`.
- [x] Test unsupported flags and extra arguments.
- [x] Test success output includes config alias and resolved source kind.
- [x] Regression test blank generation unchanged.
- [x] Regression test built-in template generation unchanged.
- [x] Regression test custom template generation unchanged.
- [x] Regression test guided generation unchanged.
- [x] Regression test agent-assisted generation unchanged.
- [x] Regression test AI-assisted generation unchanged.
- [x] Regression test validate unchanged unless explicitly integrated.
- [x] Regression test prompt, review, archive, workflow, scan, init, and config behavior unchanged except for documented default config or config-show changes.

## Phase 9 - Documentation

- [x] Update `README.md` to list config-driven templates as implemented behavior and explain the safety boundary.
- [x] Update `docs/usage.md` with `--config-template <alias>` usage, schema examples, error behavior, and validation guidance.
- [x] Update `docs/generation-modes.md` to move config-driven templates from planned to implemented.
- [x] Update any config/template-specific docs present at implementation time.
- [x] Document `.specharbor/config.yml` schema with `version: 1` and `templates.aliases`.
- [x] Document supported source kinds `builtin` and `custom`.
- [x] Document that `local`, `remote`, `url`, and arbitrary `path` sources are not supported.
- [x] Document disjoint behavior versus `--template` and `--custom-template`.
- [x] Document alias validation expectations.
- [x] Document write/conflict behavior and that existing files are skipped.
- [x] Document that no remote templates, marketplace, template script execution, network/provider behavior, production code writes, source-control automation, or archive automation exist in this feature.
- [x] Verify documented command examples match implemented CLI behavior.

## Phase 10 - Architecture and safety verification

- [x] Add or update architecture tests confirming core does not import adapters.
- [x] Add or update architecture tests confirming core does not import CLI packages.
- [x] Add or update architecture tests confirming core does not import `os`.
- [x] Confirm no provider, network, source-control, workflow, agent SDK, shell, or script execution packages are introduced into core.
- [x] Confirm generation writes no arbitrary output paths.
- [x] Run `gofmt` on Go changes.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-config-driven-templates`.
- [x] Update this task list only for implementation work actually completed.
