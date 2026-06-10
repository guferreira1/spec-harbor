# Design: Implement Advanced Validation

## Overview

Validation becomes a deterministic pipeline with three stages:

1. **Input validation** — the change id is parsed into a domain value object before any filesystem access. Invalid ids fail fast with a command error (not a finding).
2. **Structural validation** — existing behavior: project structure, change directory, and required files. Missing pieces produce error findings exactly as today.
3. **Content validation** — new behavior: every required file that exists is read once through the validation filesystem port, and pure domain rules are evaluated over the file contents. Cross-file rules (architecture section, documentation task) run after all files are loaded.

All rules are pure functions over strings. The same inputs always produce the same findings. Validation performs reads only.

## Architecture

Layer placement follows the existing dependency rule (`cmd -> adapters -> core/usecase -> core/ports + core/domain`):

- `internal/core/domain` — change-id value object, severity model, finding codes, `ValidationResult` status semantics, the canonical starter-marker set for boilerplate detection, and all content rules (placeholder detection, heading/section detection, checkbox parsing, criteria/risk rules, cross-file rules). No new imports beyond the standard library string/regex facilities already permitted in domain. Domain must not import `internal/adapters/templates` or any other adapter or CLI package, must not read the filesystem, and must not embed adapter template details.
- `internal/core/ports` — `ValidationFileSystem` gains one method:

  ```go
  type ValidationFileSystem interface {
      DirectoryExists(root string, relativePath string) (bool, error)
      FileExists(root string, relativePath string) (bool, error)
      ReadFile(root string, relativePath string) (string, error)
  }
  ```

  The interface stays small and consumer-owned. `LocalFileSystem` already implements `ReadFile`, so no adapter API change is required.
- `internal/core/usecase` — `ValidateChange` orchestrates the three stages, reads file contents only for files that exist, aggregates findings, and returns a structured `domain.ValidationResult`. No formatting, no `os`, no terminal IO.
- `internal/adapters/filesystem` — unchanged implementation; tests assert that paths resolve under the project root and that validation triggers no write methods.
- `internal/adapters/cli` — argument parsing (unchanged shape), report formatting with severity grouping, and exit-code mapping. No rule logic.

Report formatting stays outside core. Validation result types stay core-owned.

### Boundary contract

The implementation must preserve all of the following, and the architecture tests must assert them:

- Core/domain validation rules are pure functions over strings and in-memory content.
- Core/domain does not import adapters, `internal/adapters/templates`, generated template files, or CLI packages.
- Core/domain does not read template files or anything else from the filesystem; boilerplate detection uses only the domain-owned starter markers defined below.
- The use case reads OpenSpec files exclusively through the `ValidationFileSystem` port; adapters perform the actual filesystem reads.
- The CLI formats reports and maps exit codes; it contains no rule logic.
- No validation rule writes files.
- No validation rule calls the network, provider APIs, source control, workflow connectors, agents, or external commands.

## Domain Model

### Change ID value object

```go
// domain/change_id.go (new)
func NewChangeID(raw string) (ChangeID, error)
```

Rules (all produce a clear command error, before any filesystem access):

- trimmed value must be non-empty ("change id is required");
- must be a single path segment: no `/`, no `\`;
- must not be `.` or `..` and must not contain a `..` sequence;
- must not start with `/`, `\`, `-`, or `.` (rejects absolute paths, flag look-alikes, and hidden/relative segments);
- allowed characters: ASCII letters, digits, `-`, `_`, `.`;
- maximum length 128 characters.

Internal single dots are allowed: `change.v1` is a valid id. `.` alone, `..` alone, any `..` sequence, and a leading `.` remain rejected. Every existing change id in this repository (kebab-case) is accepted. The existing `validateChangeID` function in the use case is replaced by this value object; other use cases that validate change ids may adopt it in a future change but are not modified here.

### Severity model

Two levels, decided explicitly:

- `error` — the change package is not usable downstream. Errors set `ValidationStatusInvalid` and make the command exit non-zero.
- `warning` — the package is usable but has quality gaps. Warnings are always reported but never change status or exit code.
- `info` — **not included.** No current rule needs a third level; adding one without consumers would be a placeholder abstraction.

`domain` gains `ValidationFindingSeverityWarning`. `NewValidationResult` changes: `Status` is `invalid` only when at least one finding has severity `error`; a findings list that contains only warnings keeps `Status` `valid`. `ValidationResult` gains `ErrorCount()` and `WarningCount()` helpers so the CLI does not re-derive severity semantics.

Severity is fixed per rule, not contextual:

- `file_empty` is an error.
- `boilerplate_only_content` is a warning.
- `placeholder_content` is a warning. It never escalates; when a placeholder leaves a required structural element absent (for example, an acceptance-criteria file whose only items are placeholder-only), the corresponding structural error rule fires independently (see `acceptance_criteria_item_missing`).
- `tasks_all_completed` is a warning.
- The architecture/documentation cross-file advisory rules (`design_architecture_section_missing`, `tasks_documentation_task_missing`) are warnings; this design defines no structural cross-file errors.

Exit codes follow status: exit `0` for a `valid` result, including warnings-only results; non-zero only for an `invalid` result containing at least one error.

### Rule catalog

Every rule has a stable snake_case `ValidationFindingCode`. Existing codes are unchanged. Each finding carries `Severity`, `Code`, `Message` (human-readable, includes line numbers where applicable), `RelativePath` (always set for file-scoped rules), and `Subject`.

Structural rules (existing, unchanged):

| Code | Severity | Trigger |
| --- | --- | --- |
| `project_root_unavailable` | error | `openspec/project.md` or `openspec/changes/` missing |
| `change_directory_missing` | error | `openspec/changes/<change-id>/` missing (unknown change id) |
| `required_file_missing` | error | one of the five required files missing |

Per-file content rules (run on each required file that exists):

| Code | Severity | Trigger |
| --- | --- | --- |
| `file_empty` | error | file content is empty or whitespace-only |
| `file_missing_heading` | error | file has no ATX heading line (`#` ... `######`) |
| `file_missing_body` | error | file contains only headings and blank lines |
| `placeholder_content` | warning | a non-heading line contains a placeholder marker: the uppercase tokens `TBD`, `TODO`, `FIXME` as standalone words, a list item whose entire text is `N/A`, `...`, or `?`, or the case-insensitive phrase `lorem ipsum`. Task checkbox syntax (`- [ ]`) is never a placeholder; ordinary prose is never flagged merely for containing `?`, and words containing `todo` as a substring (e.g. `mastodon`) are never flagged. One finding per file listing the first offending line number |
| `boilerplate_only_content` | warning | the file has at least one meaningful (non-heading, non-blank) line and every meaningful line, after normalization, matches the domain-owned canonical starter-marker set (the file contains only starter guidance and was never meaningfully edited) |

File-specific rules:

| Code | Severity | File | Trigger |
| --- | --- | --- | --- |
| `proposal_section_missing` | warning | proposal.md | no level-2+ heading whose text starts (case-insensitive) with `Problem`, `Goal`, or `Summary` |
| `design_section_missing` | warning | design.md | no level-2+ heading whose text starts (case-insensitive) with `Overview`, `Approach`, `Design`, `Architecture`, `Technical Decisions`, or `Decisions` |
| `tasks_checkbox_missing` | error | tasks.md | no valid checkbox task line in the whole file |
| `tasks_checkbox_malformed` | error | tasks.md | a line whose first non-whitespace characters are `- [` or `* [` does not match the valid checkbox grammar below; message includes the 1-based line number |
| `tasks_phase_heading_missing` | warning | tasks.md | valid checkbox tasks exist but the file has no level-2 heading (no phases) |
| `tasks_all_completed` | warning | tasks.md | at least one valid checkbox task exists and every one is checked — flagged deterministically from file content only, as a reminder to confirm implementation evidence before review/archive; no guessing about whether implementation happened |
| `acceptance_criteria_item_missing` | error | acceptance-criteria.md | no list item (`- `, `* `, or `1. ` style) or checkbox line with meaningful text after the marker; item text that is only a placeholder (`N/A`, `...`, `?`, or a standalone `TBD`, `TODO`, or `FIXME`) does not count as meaningful |
| `risks_mitigation_missing` | warning | risks.md | file has body content but no heading containing `Mitigation` (case-insensitive) and no non-heading line containing `mitigation` |

Valid checkbox grammar (deterministic): optional leading whitespace, `-` or `*`, one space, `[ ]` or `[x]`/`[X]`, one space, at least one non-whitespace character. Anything starting with the bullet-plus-bracket prefix that deviates (e.g. `- []`, `-[ ]`, `- [y]`, `- [x]` with no text) is malformed.

Cross-file rules (run after all files load; skipped for files that are missing):

| Code | Severity | Trigger |
| --- | --- | --- |
| `design_architecture_section_missing` | warning | any required file mentions `internal/core`, `internal/adapters`, or `internal/platform`, and design.md has no heading containing `Architecture` (case-insensitive) |
| `tasks_documentation_task_missing` | warning | proposal.md or design.md contains a line referencing the `specharbor` CLI command (public behavior change) and no tasks.md line case-insensitively contains `doc` or `readme` |

These two rules are the architecture-aware layer. They protect boundaries by demanding the spec talk about architecture and documentation impact when relevant; they deliberately do not duplicate the Go import-boundary tests in `internal/architecture`, and they never parse Go source.

Rules are evaluated as an ordered list in domain (the chain-of-responsibility shape the architecture spec prescribes for validation pipelines), so future rules are added by appending an entry, not by modifying the use case.

Suppression: when `file_empty` fires for a file, the remaining content rules for that file are skipped to avoid cascading noise. When `file_missing_heading` and `file_missing_body` both apply, both are reported (they are distinct fixes).

### Boilerplate source of truth

`internal/core/domain` owns a small, explicit, stable set of canonical starter/boilerplate marker lines used only by validation. This set — not the adapter template files — is the single runtime source of truth for `boilerplate_only_content`:

- The markers are plain strings (exact normalized lines) plus the deterministic normalization rules applied before comparison: trim surrounding whitespace, strip a leading list or checkbox marker (`- `, `* `, `1. `, `- [ ] `, `- [x] `), and collapse internal whitespace runs to one space.
- The set is intentionally small: it contains the starter guidance lines currently emitted by blank authoring (the `Describe...`/`List...`/`Identify...`/`Record...` guidance sentences and the starter task lines), expressed as domain constants. No similarity scoring, no heuristics.
- The rule compares every normalized meaningful (non-heading, non-blank) line of a file against this set; the file is boilerplate-only when all meaningful lines match.
- Domain does not import `internal/adapters/templates`, generated template files, or any adapter/CLI package, does not read the filesystem, and does not embed adapter template details to determine boilerplate.
- Adapter template files may share the same starter wording, but they are never read at validation time and are not the source of truth.
- If blank/template starter wording changes in the future, the domain marker set must be updated intentionally through domain tests and a spec change. A drift-guard test in the adapter layer (adapters may import domain) pins that freshly generated blank starter content is still recognized as boilerplate by the domain rule, so template edits cannot silently change validation behavior — the drift test fails and forces an explicit domain update.

### Validation behavior by authoring state

The boilerplate, placeholder, and structural rules together must produce this behavior:

| Authoring state | Expected findings | Status / exit |
| --- | --- | --- |
| Freshly generated blank/starter change (required files present, starter headings, placeholder/starter text only) | no `required_file_missing`, no `file_empty`; `boilerplate_only_content` warning per unedited file; `placeholder_content` warning where applicable | `valid`, exit `0` (warnings only) |
| Truly empty required file (no meaningful content) | `file_empty` error for that file; other content rules for the same file suppressed | `invalid`, exit non-zero |
| Template-generated, unedited (headings plus starter guidance only) | treated as boilerplate-only: `boilerplate_only_content` warning, not an error and not fully authored content | `valid`, exit `0` (absent other errors) |
| Edited (meaningful content beyond headings and known starter lines) | no `boilerplate_only_content` for that file; other findings still fire when their specific rules fail | per remaining findings |
| Fully authored, complete change | zero errors and zero warnings, unless a non-blocking cross-file advisory warning is intentionally triggered | `valid`, exit `0` |

Built-in templates whose guidance wording matches the domain marker set are detected as boilerplate when unedited; a template whose wording has drifted from the canonical markers may not be, which is the documented, intentional trade-off (see `risks.md`) resolved by updating the domain set deliberately.

## Use Case Flow

```text
Execute(input):
  parse ChangeID            -> error return on invalid/missing id (no filesystem access yet)
  validate project structure -> existing behavior, early return with findings
  validate change directory  -> existing behavior, early return with findings
  for each required file:
    FileExists -> missing  => required_file_missing finding
    exists     => ReadFile => per-file content rules (domain)
  cross-file rules (domain) over the loaded contents
  return NewValidationResult(...)
```

Filesystem errors (permission failures, read errors) remain Go errors, not findings, matching current behavior. The use case never calls any write method; the port exposes none.

## CLI Behavior

Argument parsing is unchanged: exactly one positional change id, any flag is rejected with `unsupported flag: <flag>`, extra positionals are rejected with `unexpected argument: <arg>`, and a missing id yields `change id is required`.

Output format (text, the only format in this change):

Valid, no findings:

```text
SpecHarbor change is valid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature
Required files: 5
Errors: 0
Warnings: 0
```

Valid with warnings (exit `0`):

```text
SpecHarbor change is valid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature
Required files: 5
Errors: 0
Warnings: 2

Warnings:
- [warning] placeholder_content: Placeholder marker "TBD" found (line 12) (openspec/changes/add-example-feature/design.md)
- [warning] risks_mitigation_missing: Risks are listed without mitigation notes. (openspec/changes/add-example-feature/risks.md)
```

Invalid (exit code `1` via the existing `ExitError`):

```text
SpecHarbor change is invalid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature

Errors:
- [error] required_file_missing: Missing required file: design.md (openspec/changes/add-example-feature/design.md)
- [error] tasks_checkbox_missing: No checkbox tasks found. (openspec/changes/add-example-feature/tasks.md)

Warnings:
- [warning] proposal_section_missing: No Problem, Goal, or Summary section found. (openspec/changes/add-example-feature/proposal.md)
```

Finding line format: `- [<severity>] <code>: <message> (<relative path>)`, with the path segment omitted when `RelativePath` is empty. Findings are grouped by severity (errors first) and keep rule-evaluation order within each group, so output is stable across runs.

Exit codes: `0` when `ErrorCount() == 0` (warnings alone never fail); `ExitError{Code: 1}` when errors exist; argument and execution errors keep the existing generic error path.

## Machine-Readable Output Decision

JSON output (`--format json`) is **explicitly out of scope**. Rationale: the finding model (severity, stable code, message, path) is designed to serialize cleanly, but a JSON mode needs a committed schema, schema tests, and CI consumer documentation — enough surface to warrant its own validation/reporting change. Keeping flags unchanged also preserves the strict "no flags" parser without carve-outs. The stable rule codes introduced here are the forward-compatible foundation that future JSON output will reuse.

## Technical Decisions

- **Two severities, no info.** Every drafted rule maps cleanly to "blocks downstream use" or "quality gap"; an unused third level is a placeholder abstraction.
- **Warnings for quality, errors for structure.** The documented `generate --blank` then `validate` flow must keep exiting `0`; boilerplate and placeholder findings are therefore warnings, while structurally unusable packages (empty files, no tasks, no criteria) are errors.
- **Change-id validation stays a command error, not a finding.** This preserves the current contract (invalid input is rejected before validation runs) and keeps path-safety enforcement ahead of all filesystem access.
- **Extend the existing port instead of adding a new one.** One added read method keeps the interface small (three methods) and avoids a parallel port for the same adapter.
- **Boilerplate markers are domain-owned, not derived from templates.** `internal/core/domain` defines the small canonical starter-marker set as plain string constants and compares normalized meaningful lines against it deterministically — no similarity scoring, no heuristics. Domain never imports adapter or template packages and never reads template files to determine boilerplate; templates may share the wording but are not the runtime source of truth, and future template wording changes require an intentional domain-marker update via domain tests and a spec change.
- **No Go source parsing.** Architecture-aware rules trigger on markdown mentions of internal package paths only; deep architecture enforcement stays in `internal/architecture` tests.
- **Rule chain lives in domain.** Strategy/chain pattern per the architecture spec; the use case iterates rules, it does not implement them.

## Testing Strategy

- Domain: table-driven tests per rule with valid and invalid content, change-id value object cases (accepted internal dots such as `change.v1`, rejected traversal, separators, absolute paths, unsafe characters, length), severity/status semantics, finding-code stability, checkbox grammar edge cases.
- Domain (boilerplate): tests prove detection uses the domain-owned starter markers exclusively — blank/starter content (replicated as test fixtures, not imported from adapters) produces `boilerplate_only_content` and keeps status `valid`; template-like unedited content produces the warning, not an error; content edited with a meaningful body does not produce `boilerplate_only_content`; empty content produces `file_empty` and suppresses the other same-file content findings.
- Domain (severity/exit semantics): warnings-only findings keep the result `valid`; any error finding makes it `invalid`.
- Use case: mocked `ValidationFileSystem` covering the full matrix — valid change, missing files, empty files, blank/starter content (warnings-only, `valid`), placeholder-only, malformed tasks, warnings-only exit semantics, unknown id, unsafe id, read errors.
- Adapters: `LocalFileSystem` read behavior under a temp project root; confirm validation path resolution stays under root and no writes occur. Drift guard: an adapter-layer test feeds freshly generated blank starter content to the domain boilerplate rule and asserts it is recognized, so adapter/template wording changes cannot implicitly change validation behavior without a deliberate domain test and marker update.
- CLI: success, warning, and error outputs; warnings-only runs exit `0`; error runs exit non-zero; flag/argument rejection; output includes paths and codes.
- Regression: existing tests for `init`, `scan`, `generate`, `prompt`, `review`, `archive`, `config` keep passing unchanged.
- Architecture: existing import-boundary tests keep passing; core gains no new imports, specifically no import of `internal/adapters/templates` or any adapter/CLI package, and domain performs no filesystem reads (template files included).

## Validation

- `gofmt -w $(find . -name "*.go")` applied, then `find . -name "*.go" -print0 | xargs -0 gofmt -l` reports no files.
- `go test ./...` passes.
- `go run ./cmd/specharbor validate <change-id>` exercised manually against a complete change, a blank-generated change, and a broken change.
