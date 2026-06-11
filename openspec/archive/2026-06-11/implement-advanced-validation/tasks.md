# Tasks: Implement Advanced Validation

## Phase 1 - Domain model

- [x] Add a `ChangeID` value object in `internal/core/domain` with constructor validation: non-empty, single path segment, no `.`/`..`/`..`-sequences, no leading `/`, `\`, `-`, or `.`, allowed characters `[A-Za-z0-9._-]`, max length 128.
- [x] Add `ValidationFindingSeverityWarning` to the severity model.
- [x] Add the new `ValidationFindingCode` constants from `design.md`: `file_empty`, `file_missing_heading`, `file_missing_body`, `placeholder_content`, `boilerplate_only_content`, `proposal_section_missing`, `design_section_missing`, `tasks_checkbox_missing`, `tasks_checkbox_malformed`, `tasks_phase_heading_missing`, `tasks_all_completed`, `acceptance_criteria_item_missing`, `risks_mitigation_missing`, `design_architecture_section_missing`, `tasks_documentation_task_missing`.
- [x] Change `NewValidationResult` so `Status` is `invalid` only when an error-severity finding exists, and add `ErrorCount()` and `WarningCount()` helpers.
- [x] Define the domain-owned canonical starter-marker set in `internal/core/domain`: a small, stable list of plain-string starter guidance lines plus the deterministic line normalization (trim, strip list/checkbox markers, collapse whitespace) described in `design.md`. Do not import `internal/adapters/templates` or any adapter/CLI package, do not read the filesystem, and do not embed adapter template details.
- [x] Implement per-file content rules as pure domain functions over file content: empty file, missing heading, missing body, placeholder markers (with line number in the message), boilerplate-only detection comparing normalized meaningful lines against the domain-owned starter-marker set.
- [x] Implement file-specific rules: proposal section check, design section check, tasks checkbox grammar (valid, malformed with line numbers, none found, no phase heading, all completed), acceptance-criteria item check (placeholder-only item text such as `N/A`, `...`, `?`, or standalone `TBD`/`TODO`/`FIXME` does not count as a criterion), risks mitigation check.
- [x] Implement cross-file rules: design architecture section trigger (mentions of `internal/core`, `internal/adapters`, `internal/platform`) and documentation task trigger (CLI behavior mentioned but no doc/readme task).
- [x] Implement the ordered rule chain that evaluates per-file and cross-file rules and applies the `file_empty` suppression behavior.

## Phase 2 - Domain tests

- [x] Test `ChangeID`: accepted kebab-case ids, accepted internal single dots (`change.v1`), rejected `.` and `..` as whole ids, rejected traversal (`..` sequences, `a/../b`), separators (`/`, `\`), absolute paths, leading `-`/`.`, unsafe characters, whitespace, over-length (max 128), and empty input with clear messages.
- [x] Test severity/status semantics: warnings-only result stays `valid`; any error makes it `invalid`; `ErrorCount()`/`WarningCount()` are correct.
- [x] Test finding codes are stable strings matching `design.md`.
- [x] Test placeholder detection: `TBD`/`TODO`/`FIXME` standalone words, `N/A`/`...`/`?` placeholder-only list items, `lorem ipsum` case-insensitive; and that checkbox syntax, ordinary prose containing `?`, and words containing `todo` as a substring (e.g. `mastodon`) are not flagged.
- [x] Test heading/section detection for proposal.md and design.md, including case-insensitivity and heading levels.
- [x] Test checkbox grammar: valid `- [ ]`/`- [x]`/`* [X]` lines, malformed `- []`, `-[ ]`, `- [y]`, `- [x]` without text (with line numbers), zero-checkbox files, phase-heading detection, all-completed detection.
- [x] Test acceptance-criteria item detection: bullet, asterisk, ordered, and checkbox items; heading-only files fail; files whose only items are placeholder-only (`N/A`, `...`, `?`, standalone `TBD`/`TODO`/`FIXME`) fail.
- [x] Test risks rules: body present, mitigation heading/line detection, heading-only files.
- [x] Test boilerplate-only detection uses the domain-owned starter markers, not adapter template files: blank/starter content (replicated as in-test fixtures, never imported from `internal/adapters/templates`) produces `boilerplate_only_content`; template-like unedited content (headings plus starter guidance only) produces the warning, not an error; content with a meaningful body beyond headings and known starter lines produces no `boilerplate_only_content`; empty content produces `file_empty` and suppresses other same-file content findings.
- [x] Test cross-file rules: architecture mention with and without an Architecture heading in design.md; CLI mention with and without a documentation task.

## Phase 3 - Ports and adapters

- [x] Add `ReadFile(root string, relativePath string) (string, error)` to `ports.ValidationFileSystem`.
- [x] Confirm `LocalFileSystem` satisfies the extended port without behavior changes.
- [x] Add adapter tests: validation-relevant reads resolve under the project root, missing files are reported distinctly from read errors, and no write method is invoked during validation flows.
- [x] Add a drift-guard test in the adapter layer (adapters may import domain): feed freshly generated blank starter content to the domain boilerplate rule and assert it is recognized as boilerplate-only, so a template wording change breaks this test and forces an intentional domain marker and spec update instead of silently changing validation behavior.

## Phase 4 - Use case orchestration

- [x] Replace the local `validateChangeID` helper in `ValidateChange` with the domain `ChangeID` value object; keep missing-id and invalid-id failures as command errors raised before any filesystem access.
- [x] Keep existing structural validation (project root, change directory, required files) and early returns unchanged.
- [x] Read each existing required file once via the port and evaluate the domain rule chain; collect findings per file with `RelativePath` set.
- [x] Evaluate cross-file rules after all available files are loaded.
- [x] Return filesystem read failures as errors, not findings.

## Phase 5 - Use case tests

- [x] Valid, fully authored change returns zero findings and `valid` status.
- [x] Missing required files return `required_file_missing` errors per file.
- [x] Empty files return `file_empty` errors, suppress that file's other content rules, and make the result `invalid`.
- [x] A freshly generated blank/starter change (required files present, starter headings, starter/placeholder text only) returns `boilerplate_only_content` and applicable `placeholder_content` warnings only — no `required_file_missing`, no `file_empty` — and keeps status `valid`.
- [x] Template-generated unedited content returns the boilerplate warning, not an error; files edited with meaningful content return no `boilerplate_only_content` while other failing rules still report.
- [x] Malformed and missing checkboxes return the corresponding errors with line numbers.
- [x] Warnings-only results report `valid` status; error results report `invalid`.
- [x] Unknown change id returns the `change_directory_missing` finding; unsafe and missing change ids return clear errors without filesystem calls.
- [x] Filesystem read errors surface as Go errors.

## Phase 6 - CLI

- [x] Update `printValidationReport` to the grouped format in `design.md`: `Errors:`/`Warnings:` counts on success, grouped finding sections, `- [<severity>] <code>: <message> (<path>)` lines with the path omitted when empty.
- [x] Exit `0` when no error findings exist (including warnings-only runs); return `ExitError{Code: 1}` when error findings exist.
- [x] Keep argument parsing behavior: one positional id, unsupported flags rejected, extra arguments rejected, missing id rejected.

## Phase 7 - CLI and regression tests

- [x] CLI test: success output with zero findings and exit code 0.
- [x] CLI test: warnings-only output shows the `Warnings:` section and exits 0.
- [x] CLI test: error output groups `Errors:` (and `Warnings:` when present), includes file paths and codes, and exits non-zero.
- [x] CLI test: unsupported flags, extra arguments, and missing change id keep clear errors.
- [x] Regression tests confirm `init`, `scan`, `generate --blank/--template/--guided/--agent-assisted`, `prompt`, `review`, `archive`, and `config show`/`config` behavior is unchanged.
- [x] Architecture tests confirm core imports no adapters (explicitly including `internal/adapters/templates`), CLI packages, `os`, network, or external SDKs; that domain performs no filesystem reads (template files included); and that validation flows perform no writes.

## Phase 8 - Documentation

- [x] Update `docs/usage.md` "Validate a Change": what `validate <change-id>` checks, the required files, finding kinds, severity behavior, exit codes, example valid/warning/invalid output, and using validation before implementation, before review, and before PR.
- [x] Update `README.md` validation mentions to reflect content-quality validation and severity behavior.
- [x] Document the intentional behavior changes from `proposal.md` (new errors for unusable packages, warnings for fresh blank changes, updated report layout).

## Phase 9 - Verification

- [x] Run `gofmt -w $(find . -name "*.go")`, then confirm `find . -name "*.go" -print0 | xargs -0 gofmt -l` reports no files.
- [x] Run `go test ./...` and ensure all tests pass.
- [x] Run `go run ./cmd/specharbor validate <change-id>` against a complete change, a freshly generated blank change, and a deliberately broken change; confirm output and exit codes match `design.md` (warnings-only runs exit `0`; error runs exit non-zero).
- [x] Update this `tasks.md` checkboxes only for work actually completed.
