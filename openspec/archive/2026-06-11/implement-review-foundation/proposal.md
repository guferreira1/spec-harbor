# Proposal: Implement Review Foundation

## Problem

SpecHarbor can initialize OpenSpec projects, generate blank change packages, validate required change structure, render role-specific prompts, and archive completed changes. It still cannot review an active OpenSpec change before implementation or archive.

Review is a core product capability for OpenSpec-based development workflows. Future review behavior may inspect implementation diffs, check architecture boundaries, call optional AI providers, dispatch external agents, integrate with source-control hosts, or enforce archive readiness. The first implementation should establish a deterministic review foundation without mixing those future concerns into the initial command.

## Goal

Implement the first review capability:

```text
specharbor review <change-id>
```

The command reviews the local active OpenSpec change directory:

```text
openspec/changes/<change-id>/
```

The first version reviews only:

- OpenSpec project structure;
- active change directory existence;
- required OpenSpec change files using `domain.RequiredOpenSpecChangeFiles()`;
- deterministic `tasks.md` checkbox completion state.

The implementation must return a structured review result and print a concise human-readable CLI report.

## Scope

- Replace the `review` placeholder with real behavior for `specharbor review <change-id>`.
- Validate that exactly one change id is provided.
- Reject unsupported flags.
- Reject extra positional arguments.
- Reject unsafe change ids, including absolute-path input, traversal-like input, dot segments, path separators, drive-prefix-like input, and leading-dash input.
- Obtain the project root from the current working directory in the CLI adapter.
- Validate that the project root is available.
- Verify OpenSpec project availability through a review-specific filesystem port by checking:
  - `openspec/project.md` exists as a file;
  - `openspec/changes/` exists as a directory.
- Verify that `openspec/changes/<change-id>/` exists as a directory.
- Verify that all required OpenSpec change files exist directly under the change directory using `domain.RequiredOpenSpecChangeFiles()`.
- Read `openspec/changes/<change-id>/tasks.md` through the review filesystem port.
- Parse `tasks.md` deterministically for Markdown task checkbox lines.
- Count completed tasks:
  - `- [x]`
  - `- [X]`
- Count incomplete tasks:
  - `- [ ]`
- Return review status `approved` when all required files exist, task checkboxes are detected, and all detected tasks are completed.
- Return review status `needs-work` when required files exist and one or more detected tasks are incomplete.
- Return review status `invalid` when the project structure, change directory, required files, or readable task state are missing or invalid.
- Return one `incomplete_task` warning finding for each detected incomplete task.
- Return clear error findings for invalid project structure, missing change directory, missing required files, unreadable `tasks.md`, and `tasks.md` files with no detected task checkboxes.
- Print a human-readable CLI report from the structured review result.
- Return a non-zero CLI exit code when the review status is not `approved` so the command can be used in CI.
- Keep review orchestration in `internal/core/usecase`.
- Keep review status, task summary, findings, and result concepts in `internal/core/domain`.
- Add a small review-specific filesystem port under `internal/core/ports`.
- Perform concrete filesystem behavior through `internal/adapters/filesystem`.
- Keep task parsing deterministic and local.
- Add focused tests for domain behavior, task parsing, use case orchestration, filesystem adapter compatibility, CLI parsing, CLI reporting, CLI exit behavior, and existing command regressions.

## Out of Scope

- Git diff analysis.
- GitHub or GitLab API calls.
- Semantic code review.
- Semantic Markdown review beyond deterministic task checkbox counting.
- AI-assisted review.
- Agent-assisted review.
- Architecture boundary scanning.
- Archive-readiness enforcement.
- Auto-fix behavior.
- Updating files.
- Changing validation behavior.
- Changing generation behavior.
- Changing archive behavior.
- Changing prompt behavior.
- Scan command implementation.
- Config command implementation.
- Release notes.
- Changelog generation.
- Machine-readable output formats.
- User-configurable review rules.
- Review result persistence.
- Provider API keys, local model credentials, source-control credentials, workflow credentials, or external-agent credentials.
- Public review registries, factories, chains, AI abstractions, provider abstractions, source-control abstractions, workflow abstractions, or agent abstractions that are not directly used by this first command.
- Updating the living architecture spec.

## Success Criteria

- Running `specharbor review <change-id>` against a complete OpenSpec change with all detected tasks checked prints an approved review report and exits zero.
- Running `specharbor review <change-id>` against a complete OpenSpec change with one or more unchecked tasks prints a needs-work review report and exits non-zero.
- Running `specharbor review <change-id>` when project structure, change directory, required files, or readable task state are missing prints an invalid review report and exits non-zero.
- Missing change id, unsupported flags, extra positional arguments, and unsafe change ids are rejected clearly.
- Required OpenSpec change file policy is reused from `domain.RequiredOpenSpecChangeFiles()` and is not duplicated in review-specific code or the CLI adapter.
- The use case returns a structured review result containing the change id, checked path, status, task counts, required files, and findings.
- The CLI report follows the expected approved, needs-work, and invalid output shapes.
- Review follows Hexagonal Architecture: CLI parses and reports, the use case orchestrates, domain models review concepts, ports define filesystem dependencies, and adapters perform concrete filesystem behavior.
- The implementation does not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, source-control host APIs, network APIs, or external processes.
- Existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, `archive`, and unknown command behavior is preserved.
- `go test ./...` succeeds.
