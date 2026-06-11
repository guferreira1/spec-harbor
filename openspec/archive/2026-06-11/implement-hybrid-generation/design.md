# Design: Implement Hybrid Generation

## Overview

Hybrid generation is a deterministic composition mode:

1. Select exactly one approved template source.
2. Collect explicit metadata from CLI flags.
3. Resolve and preflight the source with existing source-specific safety rules.
4. Apply a small metadata substitution pass.
5. Write only the five required OpenSpec files under `openspec/changes/<change-id>/`, skipping existing files.
6. Run existing validation and report the result.

This first version intentionally chooses the safer model: no AI overlay, no live runner output application, no provider APIs, no shell or script execution, no production code writes, and no source-control automation.

Hybrid differs from existing modes as follows:

- Direct `--template` selects built-in starter content and does not require metadata or run validation.
- Direct `--custom-template` selects project-local template files and supports optional title and summary substitution, but does not require metadata or run validation.
- Direct `--config-template` resolves a config alias and delegates to built-in, custom, or remote behavior, but does not require metadata or run validation.
- Direct `--guided` uses explicit type, title, and summary but does not use project templates.
- Direct `--ai-assisted` imports all five files from a strict local AI output file and supports explicit overwrite.
- Direct `--agent-assisted` prints or runs a prompt in report-only mode and never applies output.
- Hybrid uses a deterministic template source plus guided metadata in one command, then validates the resulting change.

## Command Contract

Supported:

```bash
specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --template feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --template feature --type feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --config-template <alias> --type feature --title "<title>" --summary "<summary>"
```

Rejected:

```bash
specharbor generate <change-id> --hybrid
specharbor generate <change-id> --hybrid --blank --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --template feature --custom-template api-feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --template feature --type bugfix --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --guided --type feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --ai-assisted --from-file output.txt --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --agent-assisted --agent codex --type feature --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --config-template api-feature --from-file output.txt --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --config-template api-feature --overwrite --title "<title>" --summary "<summary>"
specharbor generate <change-id> --hybrid --config-template api-feature --execute --title "<title>" --summary "<summary>"
```

Flag rules:

- `--hybrid` is required for hybrid behavior.
- `--hybrid` specified more than once fails.
- Exactly one of `--template`, `--custom-template`, or `--config-template` is required.
- The same source names can exist in multiple namespaces because the flag selects the namespace.
- `--blank` is not a hybrid source.
- `--title` and `--summary` are required and non-empty after trimming.
- `--type` is optional, but when present it must be one of `feature`, `bugfix`, `docs`, or `refactor`.
- `--agent`, `--from-file`, `--overwrite`, and `--execute` are rejected with `--hybrid`.
- Extra positional arguments and unsupported flags keep the existing clear error style.

## Source Selection

Hybrid source selection is explicit and has no fallback:

- `--template <name>` resolves only built-in templates.
- `--custom-template <name>` resolves only `.specharbor/templates/<name>/`.
- `--config-template <alias>` resolves only `.specharbor/config.yml` aliases.

There is no source guessing, namespace shadowing, or precedence ordering. A built-in template, custom template, and config alias may share the same name without changing one another. Invalid or missing source selectors fail before any write.

Config aliases may resolve to:

- `builtin`;
- `custom`;
- `remote`.

Remote is allowed only through `--config-template <alias>`. Hybrid adds no remote URL flag and does not bypass any existing remote template rule.

## Metadata Behavior

Hybrid input metadata is explicit:

- `--title` is required.
- `--summary` is required.
- `--type` is optional.
- `--type`, when provided, must be one of `feature`, `bugfix`, `docs`, or `refactor`.

Empty title or summary values fail before source resolution or writes. Type values are parsed with the existing supported type set before writes.

Hybrid also computes an effective type for rendering and reporting:

- Direct `--template feature` derives effective type `feature` when `--type` is omitted.
- Direct `--template bugfix` derives effective type `bugfix` when `--type` is omitted.
- Direct `--template docs` derives effective type `docs` when `--type` is omitted.
- Direct `--template refactor` derives effective type `refactor` when `--type` is omitted.
- A `--config-template <alias>` that resolves to a built-in template follows the same derivation from the resolved built-in template name when `--type` is omitted.
- Custom sources do not infer type.
- Config custom aliases do not infer type.
- Config remote aliases do not infer type.

Built-in type consistency is mandatory:

- `--hybrid --template feature --type feature` succeeds.
- `--hybrid --template feature --type bugfix` fails clearly and writes nothing.
- `--hybrid --config-template alias-to-feature --type feature` succeeds when the alias resolves to built-in template `feature`.
- `--hybrid --config-template alias-to-feature --type bugfix` fails clearly and writes nothing when the alias resolves to built-in template `feature`.

For custom, config custom, and config remote sources, a provided type is still validated against the supported type set and is used for rendering. If omitted, no type substitution occurs and `{{type}}` remains verbatim where present.

Metadata substitution uses a small deterministic renderer:

- `{{change_id}}` is always replaced.
- `{{title}}` is always replaced in hybrid because title is required.
- `{{summary}}` is always replaced in hybrid because summary is required.
- `{{type}}` is replaced when a type was provided.
- `{{type}}` is replaced when the effective type was derived from a direct built-in template source.
- `{{type}}` is replaced when the effective type was derived from a config alias resolving to a built-in template source.
- `{{type}}` remains unchanged for custom, config custom, and config remote sources when `--type` is omitted.
- Unknown `{{...}}` tokens remain unchanged.
- Unresolved supported tokens remain unchanged when their value is absent.
- No conditionals, loops, functions, includes, front matter, template hooks, script execution, shell execution, or remote includes are introduced.

Built-in templates are rendered through the same substitution pass, even if current built-in content contains no metadata tokens. Custom templates reuse the existing deterministic substitution behavior and extend it for hybrid type metadata. Remote template content, after checksum and archive validation, receives only this same string substitution pass in hybrid mode.

## AI Overlay Decision

Do not include AI overlay in the first hybrid version.

Reasons:

- Existing AI-assisted generation already provides a strict local from-file import path.
- Applying AI-authored content on top of template content introduces overwrite, partial overlay, provenance, and validation questions that deserve a separate design.
- The first hybrid feature can deliver useful value by composing deterministic template sources with metadata and validation.

Consequences:

- `--from-file` is rejected with `--hybrid`.
- No strict AI block parser is called from hybrid generation.
- No live runner output is parsed or applied.
- No provider or model APIs are called.

A future overlay change may define a separate explicit flag, reuse the strict parser, preflight all blocks before writes, and specify overwrite behavior. That future work is not part of this change.

## Write And Conflict Behavior

Hybrid generation writes only:

```text
openspec/changes/<change-id>/proposal.md
openspec/changes/<change-id>/design.md
openspec/changes/<change-id>/tasks.md
openspec/changes/<change-id>/acceptance-criteria.md
openspec/changes/<change-id>/risks.md
```

Rules:

- The change id is validated before filesystem writes.
- Source resolution and source validation complete before creating the change directory.
- Metadata validation completes before creating the change directory.
- Remote config aliases complete URL, checksum, fetch, checksum verification, archive parsing, and bundle validation before writes.
- Invalid source selection, invalid metadata, invalid config, missing alias, unsupported source, unknown built-in template, missing custom template, invalid remote reference, checksum mismatch, or unsafe archive content writes nothing and does not create the change directory.
- Existing files are skipped and preserved.
- No `--overwrite` is supported in the first hybrid version.
- Template content cannot influence output filenames or directories.
- No production code, docs, config, CI, source-control, prompt, archive, or arbitrary files are written by the command.

The expected implementation should use preflight to avoid normal partial writes. If an unexpected runtime write failure occurs after some writes, the command reports the failure clearly and does not attempt rollback or deletion of user files.

## Validation Integration

Hybrid generation runs validation after successful generation or successful skip-only completion.

Validation rules:

- Reuse existing `ValidateChange` behavior.
- Run validation only after all planned writes or skips complete successfully.
- Do not run validation after argument, source resolution, parse, fetch, checksum, archive, or write errors.
- Print validation status, required file count, error count, warning count, and findings.
- Validation warnings do not undo writes and keep exit code `0`.
- Validation errors do not undo writes, but they are reported and make the CLI exit non-zero after printing the hybrid generation report.
- Validation never auto-fixes files.

This intentionally differs from direct template generation. The explicit `--hybrid` mode promises a more complete authoring pass, and validation is part of that pass.

## Remote Template Interaction

Hybrid remote behavior is reachable only through:

```bash
specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>"
```

The alias must resolve to `source: remote` in `.specharbor/config.yml`.

Hybrid remote rules:

- No `--remote-template` flag is added.
- No ad hoc URL is accepted on the command line.
- HTTPS-only URL validation remains unchanged.
- Userinfo, query strings, fragments, redirects, auth headers, cookies, OAuth, environment token expansion, and git URLs remain rejected by the existing remote safety model.
- The configured `sha256` checksum is verified before archive parsing.
- ZIP archive safety rules remain unchanged.
- No persistent cache is introduced.
- Built-in, custom, config built-in, and config custom sources remain network-free.

Hybrid may substitute metadata into the already-verified remote file contents, but it must never execute remote content.

## CLI Output

Hybrid output must be deterministic and include:

```text
SpecHarbor hybrid change generated.
Change: <change-id>
Mode: hybrid
Source kind: built-in|custom|config
Source: <name-or-alias>
Resolved source: builtin|custom|remote
Resolved template: <template-name>
Title: <title>
Summary: <summary>
Type: <effective-type>
Change path: openspec/changes/<change-id>
Change directory: created|existing
Created files:
- proposal.md
Skipped existing files:
- design.md

Validation:
Status: valid|invalid
Required files: 5
Errors: <n>
Warnings: <n>

Safety:
- Provider APIs called: no
- LLM APIs called: no
- Agent commands executed: no
- AI output file imported: no
- Production code modified: no
- Source-control commands run: no
- Auto-commit, auto-push, PR, merge, or archive: no
```

For config sources, output must show both the selected alias and the resolved source kind. For remote aliases, output must use the existing sanitized remote facts, such as host, format, and checksum algorithm, without printing credentials or secret-bearing URL components.

Type output rules:

- Hybrid output must include `Type: <type>` when an effective type exists.
- For built-in sources, an effective type exists even when `--type` is omitted, because it is derived from the direct or resolved built-in template.
- For custom, config custom, and config remote sources, an effective type exists only when `--type` is provided.
- When no effective type exists, the output omits the `Type:` line.
- The approved output does not print a separate `provided` or `derived` label. If implementation later adds that distinction, the exact labels must be specified in this design and covered by CLI tests before implementation is considered complete.

## Architecture

Layer responsibilities:

- `internal/core/domain` owns hybrid generation request concepts, hybrid source selection, hybrid metadata, metadata validation helpers, source selection result concepts, hybrid generation result concepts, validation summary facts, and defensive copying.
- `internal/core/usecase` owns orchestration: validate input, resolve the selected source, delegate to existing source loading behavior, apply deterministic rendering, plan writes, invoke validation, and return structured results.
- `internal/core/ports` owns any small interfaces consumed by the hybrid use case. Prefer reusing existing generation, custom template, config, remote fetch, remote bundle, and validation ports before adding new ports.
- `internal/adapters/filesystem` owns concrete local filesystem reads and writes.
- `internal/adapters/config` owns YAML parsing.
- `internal/adapters/remote` owns HTTP fetch and ZIP decoding.
- `internal/adapters/templates` owns built-in and default template content.
- `internal/adapters/cli` owns argument parsing, dependency wiring, output formatting, and exit-code mapping.

Core packages must not import:

- adapters;
- CLI packages;
- `os`;
- `os/exec`;
- terminal IO packages;
- concrete HTTP clients;
- provider SDKs;
- source-control SDKs;
- workflow SDKs;
- external-agent SDKs;
- shell or script execution packages.

Use cases depend on interfaces. CLI code must not contain source resolution, metadata substitution, validation, remote safety, or write-policy business rules.

## Documentation

Implementation must update:

- `README.md`;
- `docs/usage.md`;
- `docs/generation-modes.md`;
- any generation/template docs present at implementation time.

Docs must explain:

- what hybrid generation means;
- command usage;
- supported source flags;
- required title and summary;
- optional type behavior, built-in type derivation, and built-in type matching;
- custom, config custom, and config remote sources do not infer type when `--type` is omitted;
- metadata substitution tokens;
- validation behavior and exit codes;
- relationship with built-in, custom, config, remote, guided, AI-assisted, and agent-assisted modes;
- safety boundaries;
- examples for each source kind;
- what is intentionally out of scope.

Required examples:

```bash
specharbor generate add-login \
  --hybrid \
  --template feature \
  --title "Add login" \
  --summary "Add an OpenSpec change for login"
```

Documentation must explain that this derives `type=feature`.

```bash
specharbor generate add-login \
  --hybrid \
  --template feature \
  --type bugfix \
  --title "Add login" \
  --summary "Add an OpenSpec change for login"
```

Documentation must explain that this fails because the provided type does not match the built-in template.

Documentation must also include a custom-source example explaining that `{{type}}` remains unresolved unless `--type` is provided.

## Testing Strategy

Testing must cover domain, use case, CLI, regression, architecture, and documentation behavior. The task list defines the required test cases in detail.
