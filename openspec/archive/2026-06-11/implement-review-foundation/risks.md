# Risks: Implement Review Foundation

## Architecture leakage

Review touches CLI parsing, project-root discovery, filesystem access, required-file policy, task parsing, status calculation, CI exit behavior, and report formatting. The main risk is placing review rules directly in the CLI adapter or filesystem adapter.

Mitigation:

- Keep CLI responsibilities limited to argument parsing, current-working-directory lookup, dependency construction, report formatting, and exit signaling.
- Keep review orchestration in `internal/core/usecase`.
- Keep review status, findings, task summary, and structured results in `internal/core/domain`.
- Use a review-specific filesystem port from `internal/core/ports`.
- Keep concrete filesystem behavior in `internal/adapters/filesystem`.
- Keep required-file policy, task parsing, status calculation, and result construction out of the CLI and filesystem adapter.

## Required file policy drift

Validation and generation already use `domain.RequiredOpenSpecChangeFiles()`. Duplicating the file list for review could make commands disagree about the required OpenSpec change structure.

Mitigation:

- Use `domain.RequiredOpenSpecChangeFiles()` directly in the review use case.
- Do not define a review-specific required file list.
- Do not put required file policy in the CLI adapter.
- Add tests proving the review use case checks the shared required file policy.

## Status ambiguity for missing task checkboxes

An existing `tasks.md` with no Markdown task checkboxes is readable, but it does not provide task completion state. Treating it as approved would make the review command too weak for CI.

Mitigation:

- Define no detected task checkboxes as invalid task state for this first review.
- Return a `tasks_not_found` error finding.
- Document that semantic Markdown review is still out of scope.
- Test the no-task case explicitly.

## Non-approved reviews treated like execution errors

`needs-work` and `invalid` are expected review outcomes. If they are returned only as generic errors, users and CI logs lose the structured report.

Mitigation:

- Return structured review results for approved, needs-work, and invalid outcomes.
- Print the report before signaling non-zero exit for non-approved results.
- Reserve generic execution errors for argument errors, dependency failures, current-working-directory failures, and filesystem execution failures that prevent a structured result.
- Cover non-zero exit behavior with CLI tests.

## Overbuilding future review behavior

SpecHarbor will eventually need richer review capabilities, but adding provider frameworks, source-control clients, scanner chains, workflow dispatchers, or external-agent abstractions for this first local check would add unused surface area.

Mitigation:

- Implement only `specharbor review <change-id>`.
- Model structured results and findings that future features can extend.
- Avoid exported review registries, factories, chains, providers, agents, source-control clients, workflow connectors, scanners, and auto-fix abstractions until concrete behavior requires them.
- Do not add unsupported flags or hidden plumbing for future modes.

## Underbuilding the review foundation

A simple CLI function that checks `tasks.md` directly could satisfy the first output, but it would make future review behavior harder to add cleanly.

Mitigation:

- Introduce domain result concepts for review status, task summary, and findings.
- Add a review-specific filesystem port.
- Keep review execution in a use case with focused input validation and path construction.
- Format the CLI report from the structured result.
- Cover use case behavior with fake ports.

## Task parsing false positives

A line-based parser may count checkbox-like text in places a full Markdown parser would ignore, such as code examples. A full Markdown parser would reduce false positives but expand the scope.

Mitigation:

- Keep the first parser deterministic and intentionally simple.
- Match only supported task markers at the start of a line after leading whitespace is trimmed.
- Document that semantic Markdown review beyond checkbox counting is out of scope.
- Add parser tests for supported and ignored line shapes.
- Revisit full Markdown parsing in a later change if real usage requires it.

## Unsafe change id path handling

The change id is used to build a path under `openspec/changes/`. Absolute paths, path separators, dot segments, drive-prefix-like input, or traversal-like input could read unintended files if validation is too permissive.

Mitigation:

- Validate the change id before filesystem checks.
- Reject empty ids, `.`, `..`, `/`, `\`, `:`, leading `-`, absolute-path input, and traversal-like input.
- Build checked paths only as `openspec/changes/<change-id>`.
- Add tests that unsafe ids are rejected before filesystem checks.

## Filesystem error classification

Missing files are review findings, but low-level filesystem errors can also occur. Mapping every filesystem error to a finding could hide real execution problems.

Mitigation:

- Return structured findings for expected review states such as missing project structure, missing change directories, missing required files, unreadable `tasks.md`, no tasks, and incomplete tasks.
- Return execution errors for filesystem check failures that prevent reliable review decisions.
- Treat `tasks.md` read failure as `tasks_file_unreadable` because the required file exists but the task state cannot be reviewed.
- Cover both structured findings and execution errors in use case tests.

## Report format churn

Human-readable reports can become fragile if they rely on decorative formatting or incidental wording.

Mitigation:

- Keep output concise and deterministic.
- Test for important content: completion line, change id, checked path, status, task counts, findings count, and finding lines.
- Follow the approved, needs-work, and invalid output shapes in the design.
- Avoid banners, absolute local paths, debug output, validation summaries, provider output, source-control output, architecture scan output, and unrelated metadata.

## Accidental external integration

Review is naturally adjacent to AI, agents, source control, workflow systems, and semantic code review. Accidentally pulling those concerns into this first change would increase complexity and credentials risk.

Mitigation:

- Do not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, source-control host APIs, network APIs, Git commands, or external processes.
- Do not require provider API keys, local model credentials, agent credentials, source-control credentials, or workflow credentials.
- Keep tests local and deterministic.
- Document richer review modes as future OpenSpec changes.

## Accidental behavior changes

Review is adjacent to validation, generation, prompt rendering, and archive because all operate on OpenSpec change directories. Broad helper extraction or CLI refactoring could accidentally change existing commands.

Mitigation:

- Keep this change scoped to review.
- Do not modify validation, generation, archive, prompt, init, scan, or config behavior unless a minimal shared helper extraction is necessary.
- Preserve existing tests and add regression coverage for existing CLI commands.
- Run `go test ./...`.
