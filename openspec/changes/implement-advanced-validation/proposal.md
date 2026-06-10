# Proposal: Implement Advanced Validation

## Summary

Improve `specharbor validate <change-id>` from presence-only checks to a deterministic, local, read-only validation pipeline that catches structural, workflow, and consistency issues in OpenSpec changes:

```text
specharbor validate <change-id>
```

The command keeps its current shape (one positional change id, no new required flags) and gains:

- content-quality rules for the five required OpenSpec change files;
- stricter and safer change-id validation;
- task checkbox and phase validation for `tasks.md`;
- acceptance-criteria and risk content validation;
- proposal/design consistency and architecture-aware warning rules;
- an explicit two-level severity model (`error`, `warning`);
- grouped, path-bearing CLI output with stable rule codes;
- exit code `0` when no errors exist (warnings alone never fail), non-zero when errors exist.

Validation remains deterministic, offline, and read-only. No AI, no network, no auto-fix, no file writes.

## Problem

The current `validate <change-id>` command only verifies that `openspec/project.md`, `openspec/changes/`, the change directory, and the five required files exist. A change whose files are empty, contain only template boilerplate, have malformed task checkboxes, or lack acceptance criteria and mitigations still validates as fully valid.

Because SpecHarbor's product flow is `Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive`, weak validation lets low-quality change packages flow into implementation prompts and reviews, where problems are far more expensive to catch. Agents and humans both need a fast, deterministic gate that says whether a change package is structurally ready before implementation, before review, and before opening a PR.

## Goal

`specharbor validate <change-id>` reports actionable findings for structural and content problems in an OpenSpec change, distinguishes blocking errors from non-blocking warnings, identifies the file and rule for every finding, and exits non-zero only when errors exist — while remaining deterministic, local, read-only, and compatible with the existing hexagonal architecture.

## Scope

- Keep the existing structural rules (project structure available, change directory exists, the five required files exist) unchanged in meaning.
- Add a domain-owned change-id value object enforcing a safe, single-path-segment format; reject path traversal, absolute paths, separators, and unsafe characters with clear command errors before any filesystem access.
- Add content-quality rules for each required file: not empty, has a markdown title heading, has body content beyond headings, expected section headings present, placeholder markers reported, boilerplate-only starter content reported.
- Define a small, stable, domain-owned set of canonical starter/boilerplate marker lines in `internal/core/domain` as the only runtime source of truth for boilerplate detection; the domain never imports adapter or template packages and never reads template files to detect boilerplate.
- Add `tasks.md` rules: at least one valid checkbox task, malformed checkbox lines reported with line numbers, phase headings expected, all-tasks-checked reported as a warning.
- Add `acceptance-criteria.md` rules: at least one concrete criterion item, placeholder-only criteria reported.
- Add `risks.md` rules: risk content present, mitigation content expected.
- Add architecture-aware warning rules: design must carry an architecture section when change files reference core/adapter/platform packages, and tasks must carry a documentation task when the change documents public CLI behavior. These rules read OpenSpec markdown only; no Go source parsing.
- Introduce a two-level severity model: `error` blocks validation success; `warning` is reported but never fails the command. No `info` level in this change.
- Give every rule a stable snake_case finding code, extending the existing `ValidationFindingCode` style.
- Change `ValidationResult` status semantics: `invalid` only when at least one error-severity finding exists.
- Improve CLI output: clear success message, findings grouped by severity, every finding showing severity, code, message, and relative file path when applicable.
- Extend the core-owned `ValidationFileSystem` port with read access so the use case can validate file content through the existing filesystem adapter; validation performs no writes.
- Keep all rule logic in `internal/core/domain`, orchestration in `internal/core/usecase`, filesystem access in `internal/adapters/filesystem`, and output formatting in `internal/adapters/cli`.
- Add domain, use case, adapter, CLI, regression, and architecture tests as described in `tasks.md`.
- Update `README.md` and `docs/usage.md` inside this change's implementation work because public CLI behavior changes.

## Out of Scope

- Implementing code in this spec-authoring task.
- JSON or other machine-readable output formats (including a `--format` flag); deferred to a future validation/reporting change so this change stays focused. Text output remains the only format.
- An `info` severity level.
- AI-generated or LLM-backed validation; provider API calls of any kind.
- Remote or networked validation.
- Auto-fix behavior, auto-creating missing files, auto-checking or unchecking tasks, or any write to OpenSpec files during validation.
- Parsing Go source code; architecture-aware rules read OpenSpec markdown only.
- Validating `openspec/specs/` living specs or archived changes.
- Config-driven enabling/disabling of individual rules (the existing `validation.require_all_change_files` config field keeps its current behavior).
- Changes to `init`, `scan`, `generate` modes, `prompt`, `review`, `archive`, or `config` behavior.
- GitHub Actions or other CI workflow changes.
- Source-control automation.
- Spell-checking, grammar checks, or any non-deterministic content scoring.

## Compatibility

Advanced validation must not break existing well-authored changes. Intentional, documented behavior changes are limited to:

- Changes with empty required files, files without any markdown heading, files with headings but no body, tasks files without any valid checkbox task, malformed checkbox lines, or acceptance-criteria files without a single criterion item now fail with errors. These packages were never usable downstream; failing them is the purpose of this change.
- A freshly generated `--blank` change now validates as valid with warnings (boilerplate/placeholder findings) instead of valid with zero findings. Template-generated changes that were never edited are treated the same way: boilerplate-only warnings, not errors. The exit code remains `0`, so the documented `generate -> validate` flow keeps working.
- The success report replaces the single `Findings:` count with separate `Errors:` and `Warnings:` counts, and invalid reports group findings under `Errors:` and `Warnings:` headings with file paths appended.

## Success Criteria

- A complete, well-authored change validates with no errors and exits `0`.
- Missing files, empty files, heading-only files, checkbox-free tasks files, malformed checkboxes, and criterion-free acceptance files each produce a distinct error finding with a stable code and file path.
- Placeholder markers, boilerplate-only starter content, missing recommended sections, missing phase headings, fully checked task lists, missing mitigations, missing architecture sections, and missing documentation tasks each produce a distinct warning finding.
- Warnings alone keep status `valid` and exit code `0`; any error makes status `invalid` and exit code non-zero.
- Unsafe change ids (`..`, separators, absolute paths, unsafe characters) are rejected with a clear error before touching the filesystem; missing and unknown change ids keep clear errors.
- Validation never writes, creates, or modifies any file.
- Core packages gain no imports of adapters, CLI, `os`, network, or external SDKs; all new rule logic is deterministic and locally executable.
- README and docs/usage.md describe the checks, severities, exit codes, and example output, updated within this change.
