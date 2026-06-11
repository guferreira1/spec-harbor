# Design: Implement Config-Driven Templates

## Overview

Config-driven templates add an alias resolution step in front of existing built-in and custom template generation. The feature does not add a new template renderer, a path-copy system, or a remote registry. It adds a project-local mapping from alias names to already-supported template references.

Execution flow:

1. Parse `specharbor generate <change-id> --config-template <alias>`.
2. Validate the change id and config alias before template access.
3. Load `.specharbor/config.yml` through a core-owned config port.
4. Validate config version and template alias schema.
5. Resolve the alias to a built-in or custom template reference.
6. Validate the referenced built-in or custom template using the same rules as the direct mode.
7. Generate the OpenSpec change files with existing write-if-absent behavior.
8. Return a structured result that records the config alias and resolved source.

All behavior is deterministic and local. No network, provider, runner, source-control, workflow, or shell behavior is introduced.

## Config Schema Decision

Use a small versioned schema under `.specharbor/config.yml`:

```yaml
version: 1

templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature

    default-feature:
      source: builtin
      template: feature
```

Rules:

- `version` is required for config-driven template resolution and must be `1`.
- `templates` is optional for projects that do not use config-driven templates.
- `templates.aliases` is optional and defaults to an empty alias map.
- Alias map keys are validated as config template aliases.
- Alias values must be mappings with exactly the supported semantic fields for this version:
  - `source`, required;
  - `template`, required.
- Supported `source` values are exactly `builtin` and `custom`.
- `builtin` references an existing built-in template name.
- `custom` references an existing custom template name under `.specharbor/templates/<template-name>/`.
- The `template` value for `builtin` is validated with the existing built-in template name rules.
- The `template` value for `custom` is validated with the existing custom template name rules.

Rejected in this change:

```yaml
templates:
  aliases:
    remote-feature:
      source: remote
      url: https://example.invalid/templates/feature

    local-service:
      source: local
      path: .specharbor/templates/service
```

Unsupported source kinds, `url`, `remote`, `path`, or local path fields should fail with clear config validation errors when the config is loaded for template alias resolution. The implementation may choose strict YAML field validation for `templates.aliases` only, while preserving existing behavior for unrelated top-level config keys until a broader config strictness change is approved.

### Source Kind Decision

Support `builtin` and `custom` only.

Do not support `local` in this first version. `custom` already covers the safe project-local template root and has established validation, required-file rules, symlink/path handling, rendering, and write behavior. Arbitrary local paths would add path traversal, project-root containment, symlink, and output-path confusion risks without a current product need.

Remote sources are reserved for a future `implement-remote-templates` change.

## Alias Validation

Add a domain-owned `ConfigTemplateAlias` value object under `internal/core/domain`.

Rules:

- trim surrounding whitespace;
- non-empty;
- maximum length 128;
- single path segment;
- no `/`;
- no `\`;
- no absolute paths or drive-letter paths;
- no `.` or `..` as the whole value;
- no `..` sequence anywhere;
- no leading `.`;
- no leading `-`;
- allowed characters are ASCII letters, digits, `.`, `_`, and `-`.

Internal single dots are allowed, matching `ChangeID` and `CustomTemplateName` style. Examples:

- valid: `api-feature`, `feature.v1`, `service_template`, `Feature42`;
- invalid: ``, `../escape`, `nested/template`, `nested\template`, `.hidden`, `-flag`, `a..b`, `/absolute`, `C:\absolute`, `api feature`, `api:feature`.

Invalid aliases must produce config validation errors that name the alias field and value where possible, for example:

```text
invalid config template alias "nested/template": config template alias must be a single path segment
```

Alias validation happens before any referenced custom template directory or file is inspected.

## Domain Model

Add domain concepts:

- `ConfigTemplateAlias`
- `ConfigTemplateSourceKind`
- `ConfigTemplateReference`
- `ConfigTemplateAliasSet` or equivalent map-owning config model
- `ConfigTemplateResolution`

Recommended source kind constants:

```go
const (
    ConfigTemplateSourceBuiltin ConfigTemplateSourceKind = "builtin"
    ConfigTemplateSourceCustom  ConfigTemplateSourceKind = "custom"
)
```

`ConfigTemplateReference` holds:

- alias;
- source kind;
- referenced template name as the validated domain type where practical:
  - `domain.TemplateName` for `builtin`;
  - `domain.CustomTemplateName` for `custom`.

Constructors and accessors must defensively copy maps and slices where applicable. Domain constructors enforce source-specific validation so an unsupported or partially formed reference cannot be represented as a valid domain value.

The existing `TemplateSource` result concept may remain for direct generation. If result output needs a separate config-specific value, add a small config-template source field rather than changing the existing built-in/custom result constants in a way that breaks current tests.

## Config Loading and Validation

Use the existing local config path:

```text
.specharbor/config.yml
```

Design decisions:

- Filesystem reads stay in adapters.
- YAML decoding stays in the existing config adapter or a config-specific adapter helper.
- Core receives structured config through ports/use cases.
- Semantic validation for aliases and template references belongs in domain constructors.
- The use case owns orchestration and error context.
- Missing config when `--config-template` is used is an error:
  - `missing config file: .specharbor/config.yml`
- Invalid YAML is an error:
  - `invalid config YAML in .specharbor/config.yml: <details>`
- Missing version or unsupported version is an error for config-driven template resolution:
  - `unsupported config version 0 in .specharbor/config.yml: supported version is 1`
- Omitted `templates` or omitted `templates.aliases` means there are zero configured aliases.
- A requested alias absent from the alias map is an error:
  - `config template alias not found: <alias>`

### Default Init Config

The current embedded default config does not include a `version` field, while the existing `config show` path expects supported version `1`. This change should update generated default `.specharbor/config.yml` for new projects:

```yaml
version: 1

templates:
  aliases: {}
```

The implementation should preserve existing default settings below that version field. It should not create migration commands or rewrite existing user configs.

### Config Show Decision

`config show` does not need to become the primary validation surface for template aliases in this change. Generation with `--config-template` is the required validation path.

If implementation can update `config show` with low risk, it may add a read-only `Template aliases` section that lists alias names and source kinds. If this is done, it must be explicitly tested and documented. If not done, `config show` should continue to behave as currently documented except for any parser errors caused by invalid YAML shape.

Do not add `config get`, `config set`, or config mutation behavior.

## Command Surface Decision

Use an explicit flag:

```bash
specharbor generate <change-id> --config-template <alias>
```

Decision rationale:

- `--template` already means built-in templates.
- `--custom-template` already means project-local custom templates.
- `--config-template` clearly means a project config alias.
- The flag avoids shared namespace precedence and preserves existing command behavior.

Flag rules:

- `--config-template` requires a value; missing or flag-like value returns `config template alias is required`.
- Duplicate `--config-template` returns `config-template generation flag specified more than once`.
- `--config-template` is mutually exclusive with:
  - `--blank`;
  - `--template`;
  - `--custom-template`;
  - `--guided`;
  - `--agent-assisted`;
  - `--ai-assisted`.
- `--type`, `--agent`, `--from-file`, `--execute`, and `--overwrite` remain mode-specific and must not be accepted with `--config-template`.
- `--title` and `--summary` may be accepted with `--config-template` only when the resolved source is `custom`, matching optional custom-template rendering behavior. If accepted for all config-template invocations, built-in resolution must ignore them without changing built-in output. The safer first implementation may reject title/summary with config-template and document that aliases do not add variables beyond the resolved template mode. The implementation must choose one behavior and test it.

Preferred behavior for the first implementation: accept optional `--title` and `--summary` with `--config-template` and pass them through only when the alias resolves to `custom`. This keeps config aliases feature-equivalent to direct custom template generation without changing built-in templates.

## Precedence and Conflicts

Use disjoint namespaces by flag:

- `--template feature` always resolves built-in template `feature`.
- `--custom-template feature` always resolves `.specharbor/templates/feature/`.
- `--config-template feature` always resolves config alias `feature`.

Same names across sources are allowed. No precedence order exists because the selected flag names the namespace. A custom template named `feature` cannot shadow the built-in `feature`, and a config alias named `feature` cannot change direct `--template feature` behavior.

Future remote templates should follow the same explicit-source model unless a later OpenSpec change defines and tests a shared resolver.

## Source Resolution

### Builtin Source

For:

```yaml
templates:
  aliases:
    default-feature:
      source: builtin
      template: feature
```

Resolution rules:

- Alias validates as `ConfigTemplateAlias`.
- Source validates as `builtin`.
- Template validates with `ParseTemplateName`.
- Unknown built-in names return the existing clear unknown-template error with config context.
- Generation delegates to existing built-in template content and write behavior.

### Custom Source

For:

```yaml
templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature
```

Resolution rules:

- Alias validates as `ConfigTemplateAlias`.
- Source validates as `custom`.
- Template validates with `NewCustomTemplateName`.
- Missing custom template directory, missing required files, and empty files use the existing custom template generation errors with config alias context.
- Generation delegates to existing custom template rendering and write behavior.

### Local Path Source

Not included.

If config contains `source: local`, `path:`, absolute paths, traversal paths, or arbitrary local roots, generation must fail clearly. Do not silently ignore remote/path fields under a requested alias.

## Write Behavior

Config-driven generation inherits the write behavior of the resolved source:

- Output directory is always `openspec/changes/<change-id>/`.
- Output filenames are exactly:
  - `proposal.md`;
  - `design.md`;
  - `tasks.md`;
  - `acceptance-criteria.md`;
  - `risks.md`.
- Existing files are skipped, not overwritten.
- Alias resolution and source validation complete before change files are written.
- Invalid config, invalid alias, missing alias, unsupported source, unknown built-in template, missing custom template, missing custom template files, and empty custom template files produce zero writes where the current resolved-source behavior can support it.
- Template content cannot influence output paths.
- No production code, docs, config, CI, source-control, archive, or prompt files are written by this command.

Generated changes remain compatible with:

```bash
specharbor validate <change-id>
```

Generation does not need to auto-run validation, because direct built-in and custom template generation do not auto-run validation.

## CLI Output

Success output should include:

```text
SpecHarbor config template change generated.
Change: add-payment-flow
Config template: api-feature
Resolved source: custom
Resolved template: api-feature
Template source: .specharbor/templates/api-feature
Change path: openspec/changes/add-payment-flow
Change directory: created
Created files:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
Only OpenSpec change files under openspec/changes/add-payment-flow/ were written.
```

For built-in aliases, omit `Template source` or use a safe value such as `built-in`. The report must always list created files and skipped files when present.

Errors should clearly identify:

- missing config;
- invalid YAML;
- unsupported config version;
- invalid config shape;
- invalid alias;
- missing alias;
- unknown source type;
- unsupported remote/path fields;
- invalid referenced template;
- unknown built-in template;
- missing custom template;
- missing required custom template files;
- empty custom template files;
- generation flag conflicts.

## Architecture

Layer responsibilities:

- `internal/core/domain`
  - config template alias value object;
  - config template source kind;
  - config template reference model;
  - config template alias map model;
  - source-specific validation rules;
  - defensive copying.
- `internal/core/ports`
  - config reads/parsing port if existing ports need extension;
  - custom template reads through existing port;
  - OpenSpec writes through existing generation port.
- `internal/core/usecase`
  - config-template generation orchestration;
  - loading project config through ports;
  - resolving config aliases;
  - delegating to built-in/custom generation behavior;
  - returning structured generation results and errors.
- `internal/adapters/config`
  - YAML DTO decoding;
  - mapping YAML fields into domain config models;
  - invalid YAML and shape errors.
- `internal/adapters/filesystem`
  - local reads and writes;
  - existing safe path behavior.
- `internal/adapters/templates`
  - existing built-in and initialization templates;
  - update generated default config content only.
- `internal/adapters/cli`
  - argument parsing;
  - dependency wiring;
  - output formatting;
  - exit-code mapping.

Core must not import adapters, CLI packages, `os`, terminal IO, network APIs, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.

## Documentation Plan

Implementation tasks must update:

- `README.md`
- `docs/usage.md`
- `docs/generation-modes.md`
- any config/template-specific docs present at implementation time

Docs must explain:

- purpose of config-driven templates;
- `.specharbor/config.yml` schema;
- `version: 1` requirement;
- `templates.aliases` examples;
- `--config-template <alias>` command usage;
- source kinds `builtin` and `custom`;
- disjoint behavior versus `--template` and `--custom-template`;
- allowed and invalid alias names;
- write and conflict behavior;
- generated default config changes for new projects;
- validation expectations;
- no remote templates yet;
- no marketplace;
- no arbitrary local paths;
- no template script execution;
- no network/provider behavior;
- no production code writes or source-control automation.

## Testing Strategy

Testing should be layered:

- Domain tests cover alias validation, source kind validation, reference construction, unsupported sources, model copying, and config alias lookup behavior.
- Use case tests cover config loading, alias resolution, delegation to built-in/custom generation, error ordering, missing/invalid config, and write boundaries.
- Adapter tests cover YAML parsing, config file reads, default config content, custom template reads, path traversal rejection, symlink behavior where already supported, no network calls, and no external execution.
- CLI tests cover flag parsing, conflicts with every other generation mode, success output, and source-specific error output.
- Regression tests cover all existing generation modes and non-generation commands.
- Architecture tests confirm dependency boundaries and absence of forbidden SDK/process imports.

## Validation

Implementation verification should include:

```bash
go test ./...
go run ./cmd/specharbor validate implement-config-driven-templates
```

Manual verification should include generating from:

- a config alias pointing to a built-in template;
- a config alias pointing to a custom template;
- a missing alias;
- an invalid alias;
- an unsupported source kind;
- an unknown built-in template;
- a missing custom template.
