# Proposal: Implement Validation Foundation

## Problem

`specharbor validate` is currently not implemented. SpecHarbor can initialize a project and render agent prompts, but it cannot yet verify that an OpenSpec change has the minimum structure expected by the workflow.

Without a validation foundation, future checks for sections, task completion, acceptance criteria, risks, architecture boundaries, agent readiness, archive readiness, and optional AI-assisted review would have no consistent result model or orchestration point.

## Goal

Implement the first validation capability:

```text
specharbor validate <change-id>
```

The command validates the local OpenSpec change directory:

```text
openspec/changes/<change-id>/
```

The first version validates structure only. A change is valid when these files exist:

- `proposal.md`
- `design.md`
- `tasks.md`
- `acceptance-criteria.md`
- `risks.md`

The implementation must establish an extensible validation foundation while staying focused on this concrete file-presence check.

## Scope

- Replace the `validate` placeholder with real behavior for `specharbor validate <change-id>`.
- Accept exactly one change id argument.
- Return an error through the existing CLI error flow when the change id argument is missing or invalid.
- Obtain the project root from the current working directory in the CLI adapter.
- Validate OpenSpec project availability by verifying that `openspec/project.md` and `openspec/changes/` exist through the validation filesystem port.
- Validate that `openspec/changes/<change-id>/` exists.
- Validate that all required OpenSpec change files exist.
- Return a structured valid result when all required files exist.
- Return a structured invalid result when the project structure, change directory, or required files are missing.
- Print a human-readable CLI report for valid and invalid validation results.
- Return a non-zero CLI exit code after printing an invalid validation report so `specharbor validate <change-id>` can be used in CI.
- Keep validation orchestration in `internal/core/usecase`.
- Keep validation concepts and result types in `internal/core/domain`.
- Add small validation-specific ports under `internal/core/ports`.
- Perform filesystem checks through a port implemented by `internal/adapters/filesystem`.
- Add focused tests for domain behavior, use case orchestration, filesystem adapter compatibility, and CLI reporting.

## Out of Scope

- Semantic validation of Markdown content.
- Required section validation.
- Task checkbox validation.
- Acceptance criteria content validation.
- Risk note content validation.
- Architecture boundary validation.
- Agent-readiness validation.
- Archive-readiness validation.
- Optional AI-assisted validation.
- Calling AI providers, local model APIs, provider SDKs, external agents, workflow tools, or external processes.
- Requiring provider API keys or agent credentials.
- Generating, modifying, archiving, reviewing, scanning, or configuring OpenSpec changes.
- Auto-fixing missing files.
- Writing validation results to disk.
- Adding machine-readable output formats.
- Adding a public validator registry or full Chain of Responsibility framework before more validators exist.
- Updating the living architecture spec.

## Success Criteria

- Running `specharbor validate <change-id>` against a change containing all required files prints a valid validation report.
- Running `specharbor validate <change-id>` verifies OpenSpec project availability by checking `openspec/project.md` and `openspec/changes/` through the validation filesystem port.
- Running `specharbor validate <change-id>` against a change with missing required files prints an invalid validation report listing the missing files and exits non-zero.
- Running `specharbor validate <change-id>` against a missing change directory prints an invalid validation report identifying the missing directory and exits non-zero.
- Running `specharbor validate` returns a clear argument error through the existing error flow.
- The command does not call AI providers, local model APIs, external agents, workflow tools, or external processes.
- Validation follows Hexagonal Architecture: CLI parses and reports, the use case orchestrates, domain models validation results, and the filesystem adapter performs concrete filesystem checks.
- `go test ./...` succeeds.
