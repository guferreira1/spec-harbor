# Tasks: Implement Custom Templates

## Phase 1 - Domain model

- [x] Add a `CustomTemplateName` value object in `internal/core/domain` with constructor validation mirroring `ChangeID`: non-empty, single path segment, no `/` or `\`, no `.`/`..`/`..`-sequences, no leading `.` or `-`, allowed characters `[A-Za-z0-9._-]`, max length 128, and error messages that name the custom template concept.
- [x] Add the `TemplateSource` type with `BuiltInTemplateSource` (`built-in`) and `CustomTemplateSource` (`custom`) constants.
- [x] Add a `CustomTemplate` domain model keyed by the allowed filenames from `RequiredOpenSpecChangeFiles()`; construction fails for missing or empty (whitespace-only) files, and accessors copy contents defensively.
- [x] Add a pure rendering function that substitutes `{{change_id}}` always, and `{{title}}`/`{{summary}}` only when non-empty trimmed values are provided, leaving unresolved and unknown `{{...}}` tokens verbatim.
- [x] Extend `GenerationResult` with `TemplateSource`, `CustomTemplateName`, and `TemplatePath` fields plus a `NewCustomTemplateGenerationResult` constructor with defensive slice copying; set `BuiltInTemplateSource` in `NewTemplateGenerationResult` without changing its signature or existing fields.

## Phase 2 - Domain tests

- [x] Test `CustomTemplateName`: accepted kebab-case and dotted names, rejected empty/whitespace, `/` and `\` separators, `.` and `..` and embedded `..` sequences, leading `.` and `-`, unsafe characters, over-length (max 128), with clear messages.
- [x] Test the allowed template file list is exactly the five required OpenSpec change files.
- [x] Test `CustomTemplate` construction rejects missing files and empty/whitespace-only files, and that accessors return defensive copies (mutating a returned value does not affect the model).
- [x] Test rendering: each supported variable substitutes correctly, unresolved known variables remain verbatim when no value is provided, unknown tokens remain verbatim, substitution is deterministic, and no other content changes.
- [x] Test `NewCustomTemplateGenerationResult` field mapping, template source values, and defensive copying.

## Phase 3 - Ports and adapters

- [x] Add the `CustomTemplateFileSystem` port in `internal/core/ports` with `DirectoryExists`, `FileExists`, and `ReadFile`, documented as the read side for project-local custom templates.
- [x] Confirm `LocalFileSystem` satisfies the new port without implementation changes.
- [x] Add adapter tests: template files are read from `.specharbor/templates/<template-name>/` under a temp project root, reads resolve only under the provided root, missing files are reported distinctly from read errors, and no write method is invoked during template loading.

## Phase 4 - Use case orchestration

- [x] Extend `GenerateChange` (input fields and constructor wiring) with the custom-template path: validate the custom template name through the domain value object before any filesystem access, and keep all other mode validation unchanged.
- [x] Build the fixed relative template paths under `.specharbor/templates/<template-name>/` and check the directory through the read port; return the unknown-template error naming the expected path when absent.
- [x] Load all five required files through the read port, aggregating missing filenames into one clear error and failing on empty files via the domain model, before creating any directory or writing any file.
- [x] Render the loaded content with the domain rendering function using the change id and the optional title/summary inputs.
- [x] Reuse the existing OpenSpec project-structure requirement, change-directory creation, and `WriteFileIfAbsent` flow so writes happen only under `openspec/changes/<change-id>/` for the five known filenames, and return `NewCustomTemplateGenerationResult`.

## Phase 5 - Use case tests

- [x] Generate from a valid custom template writes the five rendered files and returns the custom result fields.
- [x] Invalid custom template names fail before any filesystem port call (assert via mock call recording).
- [x] A missing template directory returns the unknown-template error naming `.specharbor/templates/<name>`.
- [x] Missing required template files return one aggregated error listing every missing filename; empty template files return the empty-file error.
- [x] Any template loading failure produces zero writes and no change-directory creation.
- [x] Extra files in the template directory are never read or copied (only the five known relative paths are requested).
- [x] Variable substitution output is written; omitted title/summary leave tokens verbatim in written files.
- [x] Existing files under the change path are skipped, not overwritten, and reported as skipped.
- [x] Generated writes target only `openspec/changes/<change-id>/` relative paths.
- [x] Built-in `--template`, blank, and guided behavior through the use case is unchanged, including when a custom template shares a built-in name.
- [x] Missing OpenSpec project structure keeps the existing `specharbor init` guidance error.

## Phase 6 - CLI

- [x] Parse `--custom-template <template-name>` with value-required and duplicate-flag errors in the existing style.
- [x] Reject combinations: `--custom-template` with `--blank`, `--template`, `--guided`, or `--agent-assisted`, and with `--type`, `--agent`, or `--execute`; accept optional `--title`/`--summary` with `--custom-template` while keeping their existing guided/agent-assisted requirements unchanged.
- [x] Wire the custom-template path in `generateCommand` using `LocalFileSystem` for both the read port and the existing write port.
- [x] Print the custom template report from `design.md`: headline, change id, `Template: <name> (custom)`, `Template source: .specharbor/templates/<name>`, change path, directory status, created files, skipped existing files when present, and the safety note that only OpenSpec change files were written.
- [x] Keep built-in `--template` output and all exit-code mapping unchanged.

## Phase 7 - CLI and regression tests

- [x] CLI test: valid custom template usage prints the report including the custom source kind, source path, file lists, and safety note.
- [x] CLI test: invalid custom template name, missing template directory, missing template file, and empty template file produce clear errors and non-zero exit.
- [x] CLI test: duplicate `--custom-template`, missing flag value, forbidden flag combinations, unsupported flags, extra positional arguments, and missing change id produce clear errors.
- [x] CLI regression test: existing `--template <built-in>` parsing, output, and unknown-name error are byte-identical to current behavior, including when a same-named custom template directory exists.
- [x] Regression tests confirm blank, guided, and agent-assisted generation, `validate`, `prompt`, `review`, `archive`, and `config` behavior is unchanged.
- [x] Architecture tests confirm core imports no adapters, CLI packages, `os`, network, or provider/source-control/workflow/agent SDKs, and that the feature adds no external command execution and no writes outside `openspec/changes/<change-id>/`.

## Phase 8 - Documentation

- [x] Update `docs/generation-modes.md`: move custom templates from Planned to Implemented; document the `.specharbor/templates/<template-name>/` structure, the five required files, the command, variable substitution and leave-as-is behavior, the disjoint built-in vs custom flag model, skip-existing behavior, and the explicit non-goals (no remote templates, no config-driven registry, no marketplace, no template or script execution, no writes outside the change directory).
- [x] Update `docs/usage.md` with the `--custom-template` command, an example template directory, error behavior, and the recommendation to run `specharbor validate <change-id>` after generation.
- [x] Update `README.md` generation mentions to include project-local custom templates with the same safety boundaries.
- [x] Verify every documented command example against the implemented CLI behavior.

## Phase 9 - Verification

- [x] Run `gofmt -w $(find . -name "*.go")`, then confirm `find . -name "*.go" -print0 | xargs -0 gofmt -l` reports no files.
- [x] Run `go test ./...` and ensure all tests pass.
- [x] Manually exercise: generate from a valid custom template, re-run to confirm skip-existing behavior, trigger the invalid-name, unknown-template, missing-file, and empty-file errors, and run `go run ./cmd/specharbor validate <change-id>` on a generated change.
- [x] Update this `tasks.md` checkboxes only for work actually completed.
