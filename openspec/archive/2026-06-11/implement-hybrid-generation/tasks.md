# Tasks: Implement Hybrid Generation

## Phase 1 - Domain Model

- [x] Add hybrid source selection model concepts under `internal/core/domain` for built-in, custom, and config source selectors.
- [x] Add domain validation for exactly one hybrid source selector.
- [x] Add domain validation for missing hybrid source selectors.
- [x] Add domain validation for invalid source selector values through existing `TemplateName`, `CustomTemplateName`, and `ConfigTemplateAlias` value objects.
- [x] Add hybrid metadata model with required title, required summary, and optional type.
- [x] Validate title and summary as non-empty trimmed strings for hybrid generation.
- [x] Validate optional type against the supported `feature`, `bugfix`, `docs`, and `refactor` set.
- [x] Add effective type derivation for direct built-in hybrid sources so omitted type derives from selected template name.
- [x] Add effective type derivation for config aliases resolving to built-in hybrid sources so omitted type derives from resolved built-in template name.
- [x] Add built-in source and type consistency validation so provided type must match a selected or resolved built-in template.
- [x] Add non-built-in type rules so custom sources, config custom aliases, and config remote aliases do not infer omitted type.
- [x] Expose effective type in the hybrid metadata/result model for rendering and CLI output.
- [x] Expose provided-versus-derived type origin only if CLI output is updated to print that distinction.
- [x] Add deterministic hybrid metadata rendering for `{{change_id}}`, `{{title}}`, `{{summary}}`, and optional `{{type}}`.
- [x] Preserve unknown or unresolved `{{...}}` tokens verbatim.
- [x] Add hybrid result model concepts for change id, mode, selected source kind, selected source name or alias, resolved source kind, resolved source name, metadata, change path, directory status, created files, skipped files, validation result, and safety flags.
- [x] Defensively copy slices and maps in hybrid models and result accessors where applicable.

## Phase 2 - Domain Tests

- [x] Test valid hybrid source selection for exactly one built-in source.
- [x] Test valid hybrid source selection for exactly one custom source.
- [x] Test valid hybrid source selection for exactly one config source.
- [x] Test missing source selector fails clearly.
- [x] Test multiple source selectors fail clearly.
- [x] Test invalid built-in template source fails through existing template name validation.
- [x] Test invalid custom template source fails through existing custom template name validation.
- [x] Test invalid config alias source fails through existing config template alias validation.
- [x] Test title and summary are required, trimmed, and preserved in the metadata model.
- [x] Test optional type accepts only supported values and rejects unsupported values.
- [x] Test direct built-in omitted type derives effective type from selected template.
- [x] Test config built-in omitted type derives effective type from resolved built-in template.
- [x] Test built-in provided matching type succeeds.
- [x] Test built-in provided mismatched type fails clearly.
- [x] Test custom omitted type does not infer an effective type.
- [x] Test config custom omitted type does not infer an effective type.
- [x] Test config remote omitted type does not infer an effective type.
- [x] Test metadata rendering replaces `{{change_id}}`, `{{title}}`, `{{summary}}`, and `{{type}}` when provided or built-in-derived values exist.
- [x] Test non-built-in provided type renders `{{type}}`.
- [x] Test omitted type leaves `{{type}}` unchanged for custom, config custom, and config remote sources.
- [x] Test unknown tokens remain verbatim.
- [x] Test effective metadata model exposes whether type was provided or derived if CLI output labels require that distinction.
- [x] Test hybrid result field mapping.
- [x] Test defensive copying for result file lists and metadata collections where applicable.

## Phase 3 - Use Case Orchestration

- [x] Add hybrid generation input fields without changing direct generation mode behavior.
- [x] Validate project root, change id, `--hybrid` source selection, and metadata before source access or writes.
- [x] Require exactly one source selector for hybrid generation.
- [x] Reject `--blank`, direct guided mode, AI-assisted mode, agent-assisted mode, `--from-file`, `--overwrite`, `--agent`, and `--execute` for hybrid generation before writes.
- [x] Resolve built-in template sources through the existing built-in template behavior.
- [x] Resolve custom template sources through the existing custom template loading and required-file validation behavior.
- [x] Resolve config built-in aliases through existing config-template alias behavior.
- [x] Resolve config custom aliases through existing config-template alias and custom-template behavior.
- [x] Resolve config remote aliases through existing config-template remote behavior, including HTTPS, checksum, ZIP, archive, and size safeguards.
- [x] Derive effective type from direct built-in template sources when `--type` is omitted.
- [x] Derive effective type from resolved built-in template sources when `--config-template` resolves to built-in and `--type` is omitted.
- [x] Reject built-in provided type mismatches before creating directories or writing files.
- [x] Do not infer effective type for direct custom, config custom, or config remote sources when `--type` is omitted.
- [x] Confirm missing source errors write nothing.
- [x] Confirm multiple source errors write nothing.
- [x] Confirm invalid source errors write nothing.
- [x] Apply hybrid metadata substitution after deterministic source content is loaded and before writes.
- [x] Preflight all source resolution, source validation, metadata validation, remote fetch, checksum verification, archive decoding, and rendered required-file availability before creating a change directory.
- [x] Write files only under `openspec/changes/<change-id>/`.
- [x] Write only `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Preserve existing files by skipping them.
- [x] Do not support overwrite in the first hybrid version.
- [x] Do not write production, docs, config, CI, source-control, prompt, archive, or arbitrary files from the hybrid command.
- [x] Run existing validation after successful hybrid writes or successful skip-only completion.
- [x] Return structured validation status, error count, warning count, and findings.
- [x] Keep validation warnings exit-zero and validation errors non-zero after report printing.
- [x] Do not auto-fix validation findings.
- [x] Ensure no network access happens unless the selected config alias resolves to `source: remote`.
- [x] Ensure remote safety remains delegated and unchanged.

## Phase 4 - Use Case Tests

- [x] Test hybrid generation from a built-in template.
- [x] Test hybrid generation from a custom template.
- [x] Test hybrid generation from a config builtin alias.
- [x] Test hybrid generation from a config custom alias.
- [x] Test hybrid generation from a config remote alias.
- [x] Test missing source selector returns a clear error and writes nothing.
- [x] Test multiple source selectors return a clear error and writes nothing.
- [x] Test invalid source selector returns a clear error and writes nothing.
- [x] Test missing title returns a clear error and writes nothing.
- [x] Test missing summary returns a clear error and writes nothing.
- [x] Test hybrid direct built-in without `--type` renders `{{type}}` with the selected built-in template name.
- [x] Test hybrid config built-in without `--type` renders `{{type}}` with the resolved built-in template name.
- [x] Test built-in source and type mismatch returns a clear error and writes nothing.
- [x] Test hybrid custom without `--type` leaves `{{type}}` unchanged.
- [x] Test hybrid config custom without `--type` leaves `{{type}}` unchanged.
- [x] Test hybrid config remote without `--type` leaves `{{type}}` unchanged.
- [x] Test hybrid custom with `--type` renders `{{type}}`.
- [x] Test hybrid config custom with `--type` renders `{{type}}`.
- [x] Test hybrid config remote with `--type` renders `{{type}}`.
- [x] Test generated content includes title and summary substitutions.
- [x] Test generated files are written only under `openspec/changes/<change-id>/`.
- [x] Test generated filenames are limited to the five required OpenSpec files.
- [x] Test existing files are skipped and preserved.
- [x] Test source resolution failures do not create the change directory.
- [x] Test validation integration for valid output.
- [x] Test validation warnings are reported and keep success status for CLI exit mapping.
- [x] Test validation errors are reported after writes and produce non-zero CLI exit mapping.
- [x] Test no production, docs, config, CI, source-control, prompt, archive, or arbitrary files are written.
- [x] Test no network access happens for built-in, custom, config builtin, and config custom sources.
- [x] Test remote network access happens only for a selected config remote alias.
- [x] Test remote checksum, archive, and safety errors write nothing.

## Phase 5 - CLI Parsing And Output

- [x] Parse `--hybrid` with duplicate-flag errors in the existing CLI style.
- [x] Accept `--template`, `--custom-template`, or `--config-template` as hybrid source selectors only when `--hybrid` is present.
- [x] Preserve direct non-hybrid behavior for `--template`, `--custom-template`, and `--config-template`.
- [x] Reject `--hybrid` without a source selector.
- [x] Reject `--hybrid` with multiple source selectors.
- [x] Reject `--hybrid` with `--blank`.
- [x] Reject `--hybrid` with `--guided`.
- [x] Reject `--hybrid` with `--ai-assisted`.
- [x] Reject `--hybrid` with `--agent-assisted`.
- [x] Reject `--hybrid` with `--from-file`.
- [x] Reject `--hybrid` with `--overwrite`.
- [x] Reject `--hybrid` with `--agent`.
- [x] Reject `--hybrid` with `--execute`.
- [x] Parse required `--title` and `--summary` for hybrid generation with missing-value and duplicate-flag errors.
- [x] Parse optional `--type` for hybrid generation with missing-value and duplicate-flag errors.
- [x] Reject unsupported flags and extra positional arguments clearly.
- [x] Print success output with change id, mode `hybrid`, selected source kind, selected source name or alias, resolved source kind, resolved source name, title, summary, effective type when present, change path, directory status, created files, skipped files, validation status, and safety notes.
- [x] Print `Type: <type>` for direct built-in hybrid sources without `--type` using the derived effective type.
- [x] Print `Type: <type>` for config built-in hybrid sources without `--type` using the resolved derived effective type.
- [x] Omit the `Type:` line for custom, config custom, and config remote hybrid sources when `--type` is omitted.
- [x] Print error output that distinguishes argument, source selection, source resolution, metadata, remote, write, and validation failures.
- [x] Print clear error output for built-in type mismatches.
- [x] Include remote host, format, and checksum algorithm for config remote aliases without printing credentials or secret-bearing URL components.
- [x] Map validation warnings-only results to exit `0`.
- [x] Map validation errors after successful generation to non-zero exit after printing the report.

## Phase 6 - CLI Tests

- [x] Test `--hybrid` parsing with each accepted source flag.
- [x] Test conflict with non-hybrid direct modes.
- [x] Test missing source selector.
- [x] Test multiple source flags.
- [x] Test missing title and missing summary.
- [x] Test title, summary, and type parsing.
- [x] Test unsupported flags.
- [x] Test extra arguments.
- [x] Test `--from-file`, `--overwrite`, `--agent`, and `--execute` are rejected with `--hybrid`.
- [x] Test success output for direct built-in without `--type` shows effective type.
- [x] Test success output for config built-in without `--type` shows effective type.
- [x] Test error output for built-in type mismatch is clear.
- [x] Test non-built-in omitted type output is deterministic and omits the `Type:` line.
- [x] Test success output includes required hybrid report facts.
- [x] Test error output for argument and source resolution failures.
- [x] Test validation output integration for valid, warning-only, and invalid validation results.
- [x] Test remote output uses sanitized remote facts.

## Phase 7 - Regression Tests

- [x] Confirm blank generation is unchanged.
- [x] Confirm built-in template generation is unchanged.
- [x] Confirm custom template generation is unchanged.
- [x] Confirm config-template builtin generation is unchanged.
- [x] Confirm config-template custom generation is unchanged.
- [x] Confirm config-template remote generation is unchanged.
- [x] Confirm AI-assisted generation is unchanged.
- [x] Confirm agent-assisted dry-run generation is unchanged.
- [x] Confirm agent-assisted run-and-report generation is unchanged.
- [x] Confirm guided generation is unchanged.
- [x] Confirm validate behavior is unchanged except for reuse by hybrid generation.
- [x] Confirm prompt behavior is unchanged.
- [x] Confirm review behavior is unchanged.
- [x] Confirm archive behavior is unchanged.
- [x] Confirm config behavior is unchanged.
- [x] Confirm workflow behavior is unchanged.
- [x] Confirm scan behavior is unchanged.
- [x] Confirm init behavior is unchanged.
- [x] Confirm help and version behavior are unchanged.

## Phase 8 - Architecture And Safety Tests

- [x] Add or update architecture tests proving core does not import adapters for hybrid behavior.
- [x] Add or update architecture tests proving core does not import CLI packages.
- [x] Add or update architecture tests proving core does not import `os`.
- [x] Add or update architecture tests proving core does not import `os/exec`.
- [x] Add or update architecture tests proving no provider SDKs, network SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, shell execution, or script execution packages are introduced into core.
- [x] Add or update architecture tests proving no arbitrary output path write abstraction is introduced.
- [x] Add or update architecture tests proving no production-code write path is introduced by hybrid generation.
- [x] Add or update architecture tests proving no provider, LLM, local model, source-control, workflow, or live agent output application behavior is introduced.
- [x] Add or update tests proving no shell or script execution occurs.
- [x] Add or update tests proving no auto-commit, auto-push, pull request creation, merge, or archive automation is introduced.

## Phase 9 - Documentation

- [x] Update `README.md` to list hybrid generation after implementation.
- [x] Update `README.md` with hybrid examples and safety boundaries.
- [x] Update `docs/usage.md` with command syntax, required metadata, optional type behavior, built-in type derivation, source flags, validation behavior, exit codes, and safety notes.
- [x] Update `docs/generation-modes.md` to move hybrid generation from planned to implemented after code lands.
- [x] Update any generation/template docs present at implementation time.
- [x] Document built-in, custom, config builtin, config custom, and config remote examples.
- [x] Document that built-in hybrid sources derive type from the template when `--type` is omitted.
- [x] Document that custom, config custom, and config remote sources do not infer type when `--type` is omitted.
- [x] Document that provided type must match built-in template sources.
- [x] Document direct built-in hybrid without explicit type using `specharbor generate add-login --hybrid --template feature --title "Add login" --summary "Add an OpenSpec change for login"` and explain that it derives `type=feature`.
- [x] Document built-in mismatch using `specharbor generate add-login --hybrid --template feature --type bugfix --title "Add login" --summary "Add an OpenSpec change for login"` and explain that it fails.
- [x] Document custom source without type and explain that `{{type}}` remains unresolved unless `--type` is provided.
- [x] Document that `--blank` is not part of hybrid generation.
- [x] Document that AI overlay is intentionally not included in the first version.
- [x] Document that `--from-file`, `--overwrite`, `--agent`, and `--execute` are rejected with `--hybrid`.
- [x] Document remote-template interaction through config aliases only.
- [x] Document that hybrid does not call provider APIs, LLM APIs, local model APIs, agents, source-control tools, workflow tools, shell commands, or scripts.
- [x] Document that hybrid writes no production code and performs no auto-commit, auto-push, pull request, merge, or archive automation.

## Phase 10 - Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-hybrid-generation`.
- [x] Manually verify hybrid built-in generation.
- [x] Manually verify hybrid custom template generation.
- [x] Manually verify hybrid config builtin alias generation.
- [x] Manually verify hybrid config custom alias generation.
- [x] Manually verify hybrid config remote alias generation with a controlled pinned ZIP.
- [x] Manually verify existing files are skipped and preserved.
- [x] Manually verify missing and multiple source selector errors write nothing.
- [x] Manually verify validation warnings and errors produce the documented exit behavior.
- [x] Run `git status --short`.
- [x] Inspect `git diff -- openspec/changes/implement-hybrid-generation/`.
- [x] Update this `tasks.md` only for implementation work actually completed.
