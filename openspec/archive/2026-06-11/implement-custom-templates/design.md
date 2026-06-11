# Design: Implement Custom Templates

## Overview

Custom template generation is a fourth file-writing generation path next to blank, built-in template, and guided generation. It follows the same shape: validate inputs in the domain, load deterministic starter content, and write the five required OpenSpec change files with write-if-absent semantics. The only new capability is that the starter content comes from project-local files under `.specharbor/templates/<template-name>/` instead of compiled-in constants, which is why this change adds one small read-side port and strict, domain-owned name and content validation.

The pipeline is strictly ordered so that no write happens before the template is fully validated:

1. **Input validation** — change id (existing rules) and custom template name (new `CustomTemplateName` value object) are validated before any filesystem access.
2. **Project validation** — the OpenSpec project structure must exist, exactly as for other generation modes.
3. **Template loading** — the template directory is checked, all five required files are read and checked for missing/empty content. Any failure here aborts with a clear error and zero writes.
4. **Rendering** — pure domain variable substitution over the loaded content.
5. **Writing** — the existing change-directory creation and write-if-absent flow, restricted to `openspec/changes/<change-id>/` and the five known filenames.

## Command Surface Decision

The new flag is `--custom-template <template-name>`:

```bash
specharbor generate <change-id> --custom-template <template-name>
specharbor generate <change-id> --custom-template <template-name> --title "<title>" --summary "<summary>"
```

Decision rationale, against the alternative of extending `--template` with custom lookup:

- **No precedence problem exists.** `--template` resolves only the closed built-in set; `--custom-template` resolves only `.specharbor/templates/`. The two namespaces are disjoint by construction, so a custom template named `feature` can exist without ever shadowing the built-in `feature` and without any precedence rule to document, test, or get wrong.
- **No behavior change for existing users.** `--template <unknown-name>` keeps failing with the existing unknown-template error even when `.specharbor/templates/<unknown-name>/` exists. A project-local directory can never silently change what an existing documented command does — important because template content enters agent prompts downstream.
- **The output is self-describing.** The flag name states the source, and the report states it again.

Flag interaction rules (extending the existing strict parser):

- `--custom-template` requires a value; a missing value fails with `custom template name is required`.
- `--custom-template` specified more than once fails, matching existing duplicate-flag errors.
- `--custom-template` is mutually exclusive with `--blank`, `--template`, `--guided`, and `--agent-assisted`; combinations fail with errors in the existing style (for example `custom-template and template generation flags cannot be used together`).
- `--title` and `--summary` are optionally accepted with `--custom-template` (at most once each, value required when present). Their existing requirements with `--guided` and `--agent-assisted` are unchanged, and they remain rejected with `--blank` and `--template`.
- `--type`, `--agent`, and `--execute` remain rejected with `--custom-template` via the existing flag-combination errors.
- Exactly one positional change id; extra positionals and unsupported flags keep the existing errors.

## Architecture

Layer placement follows the dependency rule (`cmd -> adapters -> core/usecase -> core/ports + core/domain`):

- `internal/core/domain` — `CustomTemplateName` value object, `TemplateSource` ("built-in"/"custom"), the custom template model holding the five file contents with defensive copying, the allowed-template-file list (reusing `RequiredOpenSpecChangeFiles()`), and the pure rendering function for variable substitution. No filesystem access, no new imports beyond the standard library string facilities already used in domain.
- `internal/core/ports` — one new small, consumer-owned read port:

  ```go
  // CustomTemplateFileSystem provides only the filesystem reads required to
  // load a project-local custom template.
  type CustomTemplateFileSystem interface {
      DirectoryExists(root string, relativePath string) (bool, error)
      FileExists(root string, relativePath string) (bool, error)
      ReadFile(root string, relativePath string) (string, error)
  }
  ```

  `LocalFileSystem` already implements all three methods, so no adapter API change is required. The existing write-side `GenerationFileSystem` port is unchanged and remains the only write path.
- `internal/core/usecase` — `GenerateChange` gains the custom-template branch: it validates the name through the domain value object, builds the template's relative paths under the fixed root, loads and checks the five files through the read port, renders content through the domain function, and reuses the existing change-directory and write-if-absent flow. No `os`, no formatting, no terminal IO.
- `internal/adapters/filesystem` — unchanged implementation; `LocalFileSystem` satisfies the new port as-is. Adapter tests pin that template reads resolve under the project root.
- `internal/adapters/templates` — unchanged; built-in template content stays a separate adapter, per the open/closed rule.
- `internal/adapters/cli` — flag parsing for `--custom-template`, the custom-template report format, and unchanged exit-code mapping. No template logic.

### Boundary contract

The implementation must preserve all of the following, and tests must assert them:

- Core/domain validates names and renders content as pure functions over strings; it never reads the filesystem.
- Core packages do not import adapters, CLI packages, `os`, network clients, or provider/source-control/workflow/agent SDKs.
- All template reads go through `ports.CustomTemplateFileSystem`; all writes go through the existing `ports.GenerationFileSystem`.
- No external command execution and no network calls anywhere in the feature.
- The only paths written are `openspec/changes/<change-id>/` and the five known filenames inside it.
- Template file content never influences output paths; filenames are a fixed allowlist.

## Domain Model

### CustomTemplateName value object

```go
// domain/custom_template_name.go (new)
func NewCustomTemplateName(raw string) (CustomTemplateName, error)
```

The rules mirror `ChangeID` exactly, because both are user-supplied single path segments joined under the project root:

- trimmed value must be non-empty (`custom template name is required`);
- single path segment: no `/`, no `\`;
- must not be `.` or `..` and must not contain a `..` sequence;
- must not start with `.` or `-`;
- allowed characters: ASCII letters, digits, `-`, `_`, `.`;
- maximum length 128 characters.

Decision: **mirror, do not reuse, `ChangeID`.** A custom template name is a different concept with different error messages, and conflating the two types would let a template name flow where a change id is expected. The implementation may share the unexported character predicate inside the `domain` package, but the exported value objects stay distinct. Every error message names the custom template concept (for example `custom template name must be a single path segment`) and is raised before any filesystem access.

### TemplateSource

```go
type TemplateSource string

const (
    BuiltInTemplateSource TemplateSource = "built-in"
    CustomTemplateSource  TemplateSource = "custom"
)
```

Both sources stay within `domain.TemplateMode`; the source distinguishes where content came from. This avoids growing the `GenerationMode` enum for what is the same authoring strategy with a different content origin.

### Custom template model

A `CustomTemplate` value holds the validated name and the contents of the five files, keyed by the allowed filenames from `RequiredOpenSpecChangeFiles()`. Constructors and accessors copy the underlying map/slices defensively, matching the `GenerationResult` style. The model rejects construction with a missing or empty (whitespace-only) file so an invalid template cannot exist as a domain value.

### Rendering rules

Rendering is one pure function over the loaded content:

- `{{change_id}}` is always replaced with the validated change id.
- `{{title}}` is replaced only when a non-empty trimmed title was provided.
- `{{summary}}` is replaced only when a non-empty trimmed summary was provided.
- Replacement is plain, deterministic string substitution (the `strings.NewReplacer` approach already used by guided generation). No conditionals, loops, functions, includes, or escaping syntax.
- **Unresolved variables remain in the output verbatim.** Decision: leaving tokens as-is is deterministic, lossless, and visible in the generated files where the author edits anyway; failing generation or stripping tokens would either block legitimate generic templates or silently destroy content. Unknown `{{...}}` tokens are never an error and are never touched.
- `{{type}}` and `{{template_name}}` are not supported in this first version; neither has a consumer need yet, and adding them later is additive.

## Template Discovery

- The custom template root is the fixed relative path `.specharbor/templates/` under the project root. The project root is resolved exactly as existing generation resolves it (current working directory in the CLI).
- A template name resolves to exactly one directory: `.specharbor/templates/<template-name>/`. There is no search path, no fallback, and no registry.
- If the directory does not exist (including when `.specharbor/` or `.specharbor/templates/` is absent), generation fails with a clear error naming the expected path: `unknown custom template: <name>. Expected directory: .specharbor/templates/<name>`.
- Path traversal is impossible by construction: the name is a validated single safe segment before any read, and all reads pass `(projectRoot, relativePath)` through the port to the existing adapter.
- No directory listing is required for this change; the use case checks and reads exactly the paths it constructs. Template enumeration (for example a list command) is deferred until a feature needs it.

## Template File Requirements

- All five files are required: `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, `risks.md`. A template missing any of them is invalid. The error aggregates every missing filename in one message (for example `custom template <name> is missing required files: design.md, risks.md`) so the user fixes the template in one pass.
- Decision: **all five required, none optional.** The downstream flow (validate, prompt, review, archive) assumes the complete package, and partial templates would push the gap onto every consumer. Optional files can be revisited if a real need appears.
- An empty or whitespace-only template file is a generation error (`custom template file <name>/design.md is empty`), not a warning. Decision: an empty source file deterministically produces a `file_empty` validation error in every generated change; failing fast at the source with the template path is strictly more actionable.
- Unknown extra files and subdirectories inside the template directory are **ignored**. Decision: ignoring keeps templates forward-compatible (a team can keep a `README.md` or notes beside the template files) and keeps this change free of directory-listing requirements; only the five known filenames are ever read or copied, so ignored files can never affect output.
- Template files may contain the supported variable tokens anywhere in their content; they may also contain none.

## Write Behavior

- Generation writes only `openspec/changes/<change-id>/` plus the five known filenames inside it, through the existing `GenerationFileSystem` port.
- The existing conflict behavior is preserved unchanged: `WriteFileIfAbsent` skips files that already exist and reports them as skipped; nothing is ever overwritten.
- Ordering guarantee: the template is fully loaded, validated, and rendered **before** the change directory is created or any file is written. An invalid template produces zero filesystem writes, including no empty change directory.
- Template content and filenames cannot influence output paths; the output path set is the fixed cross product of the change path and the allowed-file list.
- No production code, documentation, configuration, or CI files are written by the command.

## Validation Integration

- Generated changes are standard OpenSpec change packages and are compatible with `specharbor validate <change-id>` with no validator changes.
- Generation does **not** auto-run validation. Decision: no existing generation mode validates after writing, and adding a side effect here would make one mode behave differently. Documentation recommends running `specharbor validate <change-id>` after generation, matching the documented flow for other modes.
- Validation findings on generated content (placeholder or boilerplate warnings, structural errors from a low-quality template) are the template author's signal to improve the template; this change adds no new validation rules.

## CLI Output

Custom template success report (matching the existing report style):

```text
SpecHarbor custom template change generated.
Change: add-payment-flow
Template: api-feature (custom)
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

- When some files already existed, a `Skipped existing files:` section lists them, matching existing generation reports, and `Change directory: existing` is shown when the directory was already present.
- The template source path is always shown for custom templates; it is safe to display because it is a project-relative path derived from the validated name, never an absolute path.
- Built-in `--template` output is unchanged; its existing headline (`SpecHarbor template change generated.`) plus the new custom headline keep the two sources unambiguous without altering existing output that users or scripts may rely on.
- Error cases produce single clear errors through the existing error path and non-zero exit: invalid name, unknown template directory, missing required template files, empty template files, and invalid flag combinations.

## Result Model

`GenerationResult` gains the custom-template fields with a new constructor, keeping existing constructors and fields untouched:

- `TemplateSource` (`built-in` for results from `NewTemplateGenerationResult`, `custom` for the new constructor);
- `CustomTemplateName` (string form of the validated name);
- `TemplatePath` (the relative source path `.specharbor/templates/<name>`).

`NewCustomTemplateGenerationResult(changeID, customTemplateName, templatePath, changePath, changeDirectoryCreated, createdFiles, skippedExistingFiles)` mirrors the existing constructors, including defensive slice copying.

## Technical Decisions

- **Separate `--custom-template` flag instead of `--template` lookup fallback.** Disjoint namespaces eliminate the precedence question, keep built-in behavior byte-identical, and prevent a project directory from silently changing a documented command. This is the safest API and the one this design commits to.
- **Fixed discovery path, no registry.** `.specharbor/templates/` already hosts project-local SpecHarbor assets (`.specharbor/rules/` exists today), and a fixed path needs no configuration, ordering, or merge semantics. A config-driven registry remains a future change (`implement-config-driven-templates`) and nothing here constrains it.
- **All five files required and non-empty.** Keeps every generated package complete for the downstream flow; fails at the source with actionable errors.
- **Minimal variable set with leave-as-is semantics.** Reuses the deterministic replacement mechanism guided generation already proved; avoids inventing a templating language; unresolved tokens stay visible and editable.
- **New read port instead of widening `GenerationFileSystem`.** The write-side port stays write-only-plus-existence, preserving the current guarantee that the generation writer interface exposes no general read. The read port is satisfied by the existing adapter with zero adapter changes.
- **Load-validate-render before any write.** Guarantees invalid templates leave no partial change directory behind.
- **`TemplateSource` instead of a new `GenerationMode`.** Custom templates are template authoring with a different content origin; modes stay aligned with the architecture spec's authoring strategies.

## Testing Strategy

- Domain: table-driven `CustomTemplateName` cases (accepted names including internal dots, rejected empty, separators, traversal, absolute-path-like input, leading `.`/`-`, unsafe characters, over-length); allowed-file list matches `RequiredOpenSpecChangeFiles()`; custom template model rejects missing/empty files and copies defensively; rendering substitutes `{{change_id}}`/`{{title}}`/`{{summary}}`, leaves unresolved and unknown tokens verbatim, and is pure.
- Use case: mocked ports covering the full matrix — successful generation, invalid name fails before any port call, unknown template directory, missing files aggregated, empty file error, extra files never read, render output written, zero writes on any template failure, skip-existing conflict behavior, writes restricted to the change path, built-in/blank/guided behavior unchanged, custom template sharing a built-in name resolved independently per flag.
- Adapters: `LocalFileSystem` satisfies the new port; template reads under a temp project root resolve only under the root; missing files and read errors are distinct; no writes occur during template loading.
- CLI: parsing for the new flag and all combination errors; success output including the custom source labeling and safety note; error outputs; regression for existing `--template` parsing and output.
- Regression: blank, built-in template, guided, agent-assisted, validate, prompt, review, archive, and config behavior unchanged.
- Architecture: existing import-boundary tests keep passing; core gains no adapter/CLI/`os`/network/SDK imports; no external process execution paths are added.

## Validation

- `gofmt -w $(find . -name "*.go")` applied, then `find . -name "*.go" -print0 | xargs -0 gofmt -l` reports no files.
- `go test ./...` passes.
- Manual checks: generate from a valid custom template, re-run to confirm skip behavior, and exercise the invalid-name, missing-template, missing-file, and empty-file errors; run `go run ./cmd/specharbor validate <change-id>` on a generated change.
