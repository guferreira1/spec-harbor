# Risks: Implement Validation Foundation

## Validation failures treated like execution errors

Validation failures are expected user-facing outcomes. If missing files are returned as ordinary use case execution errors instead of structured invalid results, future validators will be harder to compose and the CLI will be less useful.

At the same time, `specharbor validate <change-id>` must still be useful in CI. The CLI should print the invalid validation report first, then return a non-zero process status for invalid results.

Mitigation:

- Return errors for invalid command input, dependency failures, and filesystem execution failures.
- Return structured invalid validation results for missing project structure, missing change directories, and missing required files.
- After formatting an invalid validation result, have the CLI signal a non-zero exit without replacing the validation report with a generic execution error.
- Cover valid and invalid results with use case tests.
- Cover invalid CLI exit behavior with CLI tests.

## Architecture leakage

Validation touches CLI parsing, filesystem access, path construction, product rules, and report formatting. The main risk is placing required-file policy or filesystem details directly in the CLI adapter.

Mitigation:

- Keep required-file policy and result construction out of the CLI.
- Keep orchestration in `internal/core/usecase`.
- Keep validation result concepts in `internal/core/domain`.
- Use a small filesystem port from `internal/core/ports`.
- Keep the concrete filesystem adapter in `internal/adapters/filesystem`.

## Overbuilding the validator system

The project needs an extensible validation foundation, but a full Chain of Responsibility framework, registry, factory, and multiple validator abstractions would be premature for one structural check.

Mitigation:

- Model results and findings so future validators can append findings.
- Use small helper methods or directly used internal steps in the first use case.
- Avoid exported validator chain abstractions until a later change adds multiple independent validators.

## Underbuilding the foundation

A minimal command that only prints missing files without structured domain results would satisfy the first CLI behavior but would not support future validators cleanly.

Mitigation:

- Introduce domain result and finding types now.
- Include stable status and finding code values.
- Keep CLI report formatting based on the structured result.

## Accidental semantic validation

Checking file contents, sections, task checkboxes, acceptance criteria quality, risks, or architecture boundaries could expand this change beyond its intended scope.

Mitigation:

- Validate file existence only.
- Treat empty files as present for this change.
- Defer all semantic validators to separate OpenSpec changes.

## Accidental AI or external-tool integration

Validation may later support optional AI-assisted checks, but this first foundation must be local and deterministic.

Mitigation:

- Do not add provider SDKs, API calls, external process execution, workflow connectors, or agent dispatch.
- Do not require provider API keys or agent credentials.
- Keep tests local and deterministic.

## Change id path handling

The change id is used to build a relative path under `openspec/changes/`. If path handling is careless, unusual input could produce confusing results or unsafe filesystem checks.

Mitigation:

- Keep path construction constrained to `openspec/changes/<change-id>`.
- Reject clearly invalid argument shapes in the CLI.
- If additional safety checks are introduced, keep them narrow and covered by tests.

## Report format churn

Human-readable reports can become hard to test if the implementation over-specifies decorative formatting.

Mitigation:

- Keep report output simple and deterministic.
- Test for important content: status, change id, checked path, and missing filenames.
- Follow the expected valid and invalid output shapes documented in the acceptance criteria.
- Avoid banners, debug output, absolute local paths, and unrelated command summaries.
