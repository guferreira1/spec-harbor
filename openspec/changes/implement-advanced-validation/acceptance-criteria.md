# Acceptance Criteria: Implement Advanced Validation

## Command behavior

- [ ] `specharbor validate <change-id>` against a complete, well-authored change prints the valid report with `Errors: 0` and `Warnings: 0` and exits `0`.
- [ ] A change with quality warnings but no errors prints the valid report, lists findings under `Warnings:`, and exits `0`.
- [ ] A change with at least one error prints the invalid report, lists findings grouped under `Errors:` (and `Warnings:` when present), and exits non-zero.
- [ ] Every file-scoped finding line shows severity, stable rule code, human-readable message, and the relative file path.
- [ ] Output ordering is deterministic: identical inputs produce identical output across runs.
- [ ] Text output is the only format; no `--format` flag exists, and any flag is rejected with `unsupported flag: <flag>`.
- [ ] Extra positional arguments are rejected with `unexpected argument: <arg>`; a missing change id is rejected with `change id is required`.

## Change-id safety

- [ ] Change ids containing `/`, `\`, `..`, leading `/`, leading `-`, leading `.`, whitespace, or characters outside `[A-Za-z0-9._-]` are rejected with a clear error before any filesystem access, and the command exits non-zero.
- [ ] All existing change ids under `openspec/changes/` in this repository are accepted, and ids with internal single dots (e.g. `change.v1`) are accepted; `.` alone, `..` alone, and any `..` sequence are rejected; ids longer than 128 characters are rejected.
- [ ] An unknown but well-formed change id produces the `change_directory_missing` error finding naming the checked path.

## Structural and content rules

- [ ] Missing `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, or `risks.md` each produce a `required_file_missing` error naming the file.
- [ ] An empty or whitespace-only required file produces a `file_empty` error and no further content findings for that file.
- [ ] A required file with no markdown heading produces `file_missing_heading`; a file with headings but no body lines produces `file_missing_body`; both are errors.
- [ ] Placeholder markers (standalone `TBD`, `TODO`, `FIXME`; list items that are only `N/A`, `...`, or `?`; case-insensitive `lorem ipsum`) produce a `placeholder_content` warning naming the file and first offending line number; ordinary prose containing `?` and words containing `todo` as a substring are not flagged.
- [ ] A file whose meaningful (non-heading, non-blank) lines all match the domain-owned canonical starter-marker set produces a `boilerplate_only_content` warning — detection never reads or imports adapter template files.
- [ ] `proposal.md` without a Problem, Goal, or Summary section produces a `proposal_section_missing` warning; `design.md` without an Overview/Approach/Design/Architecture/Decisions section produces a `design_section_missing` warning.

## Authoring-state behavior

- [ ] A freshly generated blank/starter change (required files present, starter headings, starter/placeholder text only) produces no `required_file_missing` and no `file_empty` findings, produces `boilerplate_only_content` (and `placeholder_content` where applicable) warnings, keeps status `valid`, and exits `0`.
- [ ] A template-generated change that was never edited (headings plus starter guidance only) is treated as boilerplate-only: warning, not error, not fully authored content.
- [ ] A file edited with meaningful content beyond headings and known starter lines produces no `boilerplate_only_content` finding for that file, while other findings still fire when their specific structural rules fail.
- [ ] A complete, fully authored change produces zero errors and zero warnings, unless a non-blocking cross-file advisory warning is intentionally triggered, and validates as `valid`.

## Task rules

- [ ] `tasks.md` with no valid checkbox task produces a `tasks_checkbox_missing` error.
- [ ] Lines starting with checkbox-like syntax that violate the grammar (`- []`, `-[ ]`, `- [y]`, `- [x]` with no text) produce `tasks_checkbox_malformed` errors with 1-based line numbers.
- [ ] Valid unchecked (`- [ ] text`) and checked (`- [x] text`, `- [X] text`, `* [ ] text`) tasks are accepted without findings.
- [ ] `tasks.md` with checkbox tasks but no level-2 heading produces a `tasks_phase_heading_missing` warning.
- [ ] `tasks.md` where every checkbox task is checked produces a `tasks_all_completed` warning derived only from file content.

## Acceptance-criteria and risk rules

- [ ] `acceptance-criteria.md` without a single list/checkbox item containing meaningful text produces an `acceptance_criteria_item_missing` error; items whose entire text is a placeholder (`N/A`, `...`, `?`, standalone `TBD`/`TODO`/`FIXME`) do not count as criteria.
- [ ] `risks.md` with body content but no mitigation heading or mitigation-mentioning line produces a `risks_mitigation_missing` warning.

## Consistency rules

- [ ] When any required file mentions `internal/core`, `internal/adapters`, or `internal/platform` and `design.md` has no heading containing `Architecture`, a `design_architecture_section_missing` warning is produced.
- [ ] When `proposal.md` or `design.md` references the `specharbor` CLI command and no `tasks.md` line contains `doc` or `readme` (case-insensitive), a `tasks_documentation_task_missing` warning is produced.

## Safety and determinism

- [ ] Validation performs no writes: no files are created, modified, checked, unchecked, or deleted in any validation run.
- [ ] Validation makes no network calls and invokes no AI providers, agents, or external SDKs.
- [ ] All rules are deterministic functions of the change files' content.

## Architecture

- [ ] All rule logic lives in `internal/core/domain`; orchestration lives in `internal/core/usecase`; filesystem access lives in `internal/adapters/filesystem`; formatting and exit codes live in `internal/adapters/cli`.
- [ ] `ports.ValidationFileSystem` has exactly three methods (`DirectoryExists`, `FileExists`, `ReadFile`) and exposes no write operations.
- [ ] Existing architecture boundary tests pass; core gains no imports of adapters, CLI packages, `os`, HTTP clients, or external SDKs — explicitly including no import of `internal/adapters/templates` or generated template files.
- [ ] Boilerplate detection is proven by tests to use the domain-owned starter markers: domain reads no template files from the filesystem, and an adapter-layer drift-guard test pins that freshly generated blank starter content is recognized as boilerplate, so adapter/template changes cannot implicitly change validation behavior without a domain test update.

## Compatibility and regression

- [ ] `init`, `scan`, `generate --blank`, `generate --template`, `generate --guided`, `generate --agent-assisted`, `prompt`, `review`, `archive`, `config show`, and `config` behave exactly as before.
- [ ] All previously valid, fully authored changes in this repository still validate without errors.
- [ ] The only behavior changes are those listed in the proposal's Compatibility section.

## Tests and documentation

- [ ] Domain, use case, adapter, CLI, regression, and architecture tests described in `tasks.md` exist and pass via `go test ./...`.
- [ ] After `gofmt -w $(find . -name "*.go")`, the check `find . -name "*.go" -print0 | xargs -0 gofmt -l` reports no files.
- [ ] `docs/usage.md` and `README.md` document the checks, required files, finding kinds, severity behavior, exit codes, example output, and pre-implementation/pre-review/pre-PR usage.
