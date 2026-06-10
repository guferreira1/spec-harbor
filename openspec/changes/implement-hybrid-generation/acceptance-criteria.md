# Acceptance Criteria: Implement Hybrid Generation

## Command Behavior

- `specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>"` generates from a built-in template using hybrid metadata behavior.
- `specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>"` generates from a project-local custom template using hybrid metadata behavior.
- `specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>"` generates from a config alias using hybrid metadata behavior.
- `--hybrid` requires exactly one of `--template`, `--custom-template`, or `--config-template`.
- `--hybrid` without a source selector fails clearly and writes nothing.
- `--hybrid` with multiple source selectors fails clearly and writes nothing.
- `--hybrid` cannot be combined with `--blank`, `--guided`, `--ai-assisted`, `--agent-assisted`, `--from-file`, `--overwrite`, `--agent`, or `--execute`.
- Unsupported flags, duplicate flags, extra positional arguments, missing change id, and unsafe change id return clear errors.

## Source Selection

- `--template <name>` with `--hybrid` resolves only built-in templates.
- `--custom-template <name>` with `--hybrid` resolves only `.specharbor/templates/<name>/`.
- `--config-template <alias>` with `--hybrid` resolves only `.specharbor/config.yml` aliases.
- Same names across built-in templates, custom templates, and config aliases are allowed because flags disambiguate the namespace.
- There is no source guessing, fallback, or shadowing.
- `--blank` is not a hybrid source.
- Config aliases may resolve to `builtin`, `custom`, or `remote`.
- Remote templates are reachable only through `--config-template <alias>`.
- No `--remote-template` flag is introduced.

## Metadata

- `--title` is required for hybrid generation.
- `--summary` is required for hybrid generation.
- Empty or whitespace-only title and summary values fail clearly before writes.
- `--type` is optional for hybrid generation.
- Supported type values are exactly `feature`, `bugfix`, `docs`, and `refactor`.
- Unsupported type values fail clearly before writes.
- Direct built-in `--template feature` without `--type` derives effective type `feature`.
- Direct built-in `--template bugfix` without `--type` derives effective type `bugfix`.
- Direct built-in `--template docs` without `--type` derives effective type `docs`.
- Direct built-in `--template refactor` without `--type` derives effective type `refactor`.
- A config alias resolving to a built-in template without `--type` derives effective type from the resolved built-in template name.
- If the selected or resolved source is a built-in template and `--type` is provided, the type must match the built-in template name.
- Built-in source with matching provided type succeeds.
- Built-in source and type mismatch fails clearly and writes nothing.
- Custom sources do not infer type when `--type` is omitted.
- Config custom aliases do not infer type when `--type` is omitted.
- Config remote aliases do not infer type when `--type` is omitted.
- Custom, config custom, and config remote sources use a provided `--type` value for rendering.

## Rendering

- Hybrid rendering replaces `{{change_id}}` with the validated change id.
- Hybrid rendering replaces `{{title}}` with the trimmed title.
- Hybrid rendering replaces `{{summary}}` with the trimmed summary.
- Hybrid rendering replaces `{{type}}` when a type value is provided.
- Hybrid rendering replaces `{{type}}` when a direct built-in source derives an effective type from the selected template.
- Hybrid rendering replaces `{{type}}` when a config built-in source derives an effective type from the resolved built-in template.
- Direct custom sources without `--type` leave `{{type}}` unresolved/verbatim.
- Config custom aliases without `--type` leave `{{type}}` unresolved/verbatim.
- Config remote aliases without `--type` leave `{{type}}` unresolved/verbatim.
- Unknown `{{...}}` tokens remain verbatim.
- Supported tokens without values remain verbatim.
- No conditionals, loops, functions, includes, template hooks, scripts, shell commands, remote includes, or executable template behavior are introduced.
- Built-in, custom, config custom, and config remote source contents are all treated as static Markdown and are never executed.

## AI Overlay

- Hybrid generation does not support `--from-file`.
- Hybrid generation does not call the strict AI-assisted parser.
- Hybrid generation does not import AI-authored file overlays.
- Hybrid generation does not parse or apply live runner output.
- Hybrid generation does not call provider APIs, LLM APIs, local model APIs, or remote AI services.
- Hybrid generation does not execute agents.

## Write Safety

- Hybrid generation writes only under `openspec/changes/<change-id>/`.
- Hybrid generation writes only `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- Source resolution and metadata validation complete before creating the change directory.
- Invalid source selection, invalid metadata, invalid config, missing alias, unsupported source, unknown built-in template, missing custom template, missing custom template files, empty custom template files, invalid remote reference, checksum mismatch, network failure, unsafe archive content, and write preflight failures write nothing.
- Existing files are skipped and preserved.
- No overwrite flag is supported for hybrid generation.
- Template content, config values, and remote archive paths cannot control output paths.
- Hybrid generation writes no production code, docs, config, CI files, prompt files, archive files, source-control files, or arbitrary paths.
- Hybrid generation performs no source-control automation, workflow automation, auto-commit, auto-push, pull request creation, merge, or archive.

## Validation Integration

- Validation runs after successful hybrid writes or successful skip-only completion.
- Validation uses the existing validation use case and finding model.
- Validation status is printed.
- Validation required-file count is printed.
- Validation error and warning counts are printed.
- Validation findings are printed with severity, code, message, and path.
- Validation warnings alone keep the command exit code `0`.
- Validation errors make the command exit non-zero after the hybrid generation and validation report is printed.
- Validation never auto-fixes files.
- Validation failures do not undo generated files.

## Remote Templates

- Hybrid remote behavior uses only config aliases with `source: remote`.
- Direct built-in template generation remains network-free.
- Direct custom template generation remains network-free.
- Hybrid built-in template generation remains network-free.
- Hybrid custom template generation remains network-free.
- Hybrid config aliases resolving to `builtin` or `custom` remain network-free.
- Network access occurs only for the selected config alias when it resolves to `remote`.
- Existing HTTPS-only, checksum, no-redirect, no-credential, no-query, no-fragment, ZIP, archive path, symlink, executable-bit, duplicate, extra-file, missing-file, empty-file, and size-limit safeguards remain unchanged.
- Hybrid remote rendering never executes remote content.

## CLI Output

- Success output includes the change id.
- Success output identifies mode as `hybrid`.
- Success output includes selected source kind.
- Success output includes selected source name or alias.
- Config-source success output includes resolved source kind.
- Config-source success output includes resolved source name or safe remote facts.
- Output includes title and summary.
- Output includes `Type: <type>` when a provided or built-in-derived effective type exists.
- Output for direct built-in sources without `--type` reports the derived effective type.
- Output for config aliases resolving to built-in sources without `--type` reports the derived effective type.
- Output for custom, config custom, and config remote sources without `--type` omits the `Type:` line.
- Output does not require a separate `provided` or `derived` label for type. If implementation adds such a label, the exact labels must be specified in `design.md` and tested before acceptance.
- Output includes change path and change directory status.
- Output lists created files.
- Output lists skipped existing files when present.
- Output includes validation status, required file count, errors, warnings, and findings.
- Output includes safety notes stating that no provider APIs were called.
- Output includes safety notes stating that no LLM APIs were called.
- Output includes safety notes stating that no agent commands were executed.
- Output includes safety notes stating that no AI output file was imported.
- Output includes safety notes stating that no production code was modified.
- Output includes safety notes stating that no source-control commands were run.
- Output includes safety notes stating that no auto-commit, auto-push, pull request, merge, or archive was performed.
- Remote success output never displays credentials, query-token values, auth headers, cookies, or environment-derived secrets.

## Architecture

- Hybrid source selection and metadata concepts live in `internal/core/domain`.
- Hybrid result concepts live in `internal/core/domain`.
- Hybrid orchestration lives in `internal/core/usecase`.
- Filesystem reads and writes go through core-owned ports.
- Config parsing goes through core-owned ports and adapter-owned YAML parsing.
- Remote fetching and ZIP decoding go through core-owned ports and adapter-owned implementations.
- Built-in and default template content remain in template adapters.
- CLI parsing, dependency wiring, output formatting, and exit-code mapping live in `internal/adapters/cli`.
- Core packages do not import adapters.
- Core packages do not import CLI packages.
- Core packages do not import `os`.
- Core packages do not import `os/exec`.
- Core packages do not perform terminal IO.
- Core packages do not import concrete HTTP clients, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, shell execution, or script execution packages for hybrid behavior.
- Use cases depend on interfaces, not concrete adapters.
- CLI code contains no hybrid source resolution, rendering, validation, remote safety, or write-policy business rules.

## Documentation

- `README.md` documents hybrid generation after implementation.
- `docs/usage.md` documents command syntax, examples, required metadata, optional type behavior, built-in type derivation, validation behavior, exit codes, and safety notes.
- `docs/generation-modes.md` identifies hybrid generation as implemented after this feature lands.
- Generation/template docs present at implementation time are updated.
- Documentation explains how hybrid differs from direct built-in, custom, config-template, guided, AI-assisted, and agent-assisted modes.
- Documentation explains that built-in hybrid sources derive type from the template when `--type` is omitted.
- Documentation explains that custom, config custom, and config remote sources do not infer type when `--type` is omitted.
- Documentation explains that provided type must match built-in template sources.
- Documentation includes a direct built-in hybrid example without explicit type and explains that it derives `type=feature`.
- Documentation includes a built-in mismatch example and explains that it fails.
- Documentation includes a custom-source example and explains that `{{type}}` remains unresolved unless `--type` is provided.
- Documentation states that AI overlay and live runner output application are intentionally out of scope.
- Documentation states that remote templates are available only through config aliases and retain existing remote safeguards.

## Tests And Regression

- Domain tests cover hybrid source selection, exactly-one-source validation, invalid and missing source behavior, metadata validation, built-in omitted-type derivation, config built-in omitted-type derivation, built-in type consistency, non-built-in no-inference behavior, rendering behavior, hybrid result model, and defensive copying.
- Use case tests cover built-in, custom, config builtin, config custom, and config remote hybrid generation.
- Use case tests cover missing source, multiple source, invalid source, missing metadata, built-in omitted-type rendering, config built-in omitted-type rendering, built-in mismatch write prevention, non-built-in omitted-type preservation, non-built-in provided-type rendering, write boundaries, existing-file skip behavior, validation integration, no production/docs/config/CI/source-control/prompt/archive/arbitrary writes, no network access unless remote alias is selected, and unchanged remote safety delegation.
- CLI tests cover `--hybrid` parsing, accepted source flags, conflicts with non-hybrid modes, missing source, multiple sources, title/summary/type parsing, unsupported flags, extra args, built-in derived-type output, built-in mismatch error output, non-built-in omitted-type output, success output, error output, and validation output integration.
- Regression tests prove blank, built-in template, custom template, config-template, remote config-template, AI-assisted, agent-assisted, guided, validate, prompt, review, archive, config, workflow, scan, init, help, and version behavior remain unchanged.
- Architecture tests prove import boundaries and absence of provider, network, source-control, workflow, external-agent, shell/script execution, arbitrary output path, production-code write, and automation behavior.

## Verification

- `go test ./...` passes after implementation.
- `go run ./cmd/specharbor validate implement-hybrid-generation` passes for this OpenSpec change.
- Implementation updates `tasks.md` only for work actually completed.
