# Design: Implement Review Foundation

## Overview

`specharbor review <change-id>` reviews the deterministic readiness of an active OpenSpec change.

This first review foundation checks local OpenSpec structure and `tasks.md` checkbox completion state only. It does not inspect source code, Git diffs, source-control hosts, architecture boundaries, provider APIs, external agents, or workflow systems.

The implementation should be small enough to review easily while establishing the product shape needed for future review extensions: structured domain results, a use case orchestration boundary, a review-specific filesystem port, and CLI report formatting outside the core.

## CLI Contract

Supported command shape:

```text
specharbor review <change-id>
```

Reject:

- `specharbor review`
- `specharbor review <change-id> extra`
- unsupported flags such as `--format`, `--json`, `--ai`, `--github`, `--gitlab`, `--diff`, `--fix`, or `--agent`
- change ids parsed as flags, such as `-bad-id`
- unsafe change ids such as absolute paths, traversal-like input, dot segments, path separators, drive-prefix-like input, or leading-dash input

On argument errors, return an error from the CLI adapter so `cmd/specharbor/main.go` handles the existing error flow.

On review status `approved`, print the report to stdout and return zero.

On review status `needs-work` or `invalid`, print the report to stdout and then return a non-zero exit signal so `specharbor review <change-id>` can be used in CI. Non-approved review statuses are structured review results, not generic execution errors.

## Expected CLI Output

Approved output:

```text
SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: approved
Tasks: 12 total, 12 completed, 0 incomplete
Findings: 0
```

Needs-work output:

```text
SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: needs-work
Tasks: 12 total, 10 completed, 2 incomplete

Findings:
- [warning] incomplete_task: Task is not completed: Run go test ./...
- [warning] incomplete_task: Task is not completed: Update this tasks.md by checking off only completed tasks.
```

Invalid output:

```text
SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: invalid
Tasks: 0 total, 0 completed, 0 incomplete

Findings:
- [error] required_file_missing: Missing required file: risks.md
```

The CLI report must not print absolute local paths, debug output, provider names, agent names, source-control details, Git diff summaries, architecture scan results, validation command output, archive summaries, or future metadata fields.

## Review Status

Add review status values under `internal/core/domain`:

- `approved`
- `needs-work`
- `invalid`

Status calculation:

- `invalid` when project structure, change directory, required files, or readable task state are missing or invalid.
- `needs-work` when the structure is valid, `tasks.md` is readable, at least one task checkbox is detected, and one or more detected tasks are incomplete.
- `approved` when the structure is valid, `tasks.md` is readable, at least one task checkbox is detected, and every detected task is completed.

`tasks.md` with no detected Markdown task checkboxes is invalid for this review because the command cannot determine task completion state.

## Findings

Add review finding severity values under `internal/core/domain`:

- `error`
- `warning`

Initial finding codes:

- `project_root_unavailable`
- `change_directory_missing`
- `required_file_missing`
- `tasks_file_unreadable`
- `tasks_not_found`
- `incomplete_task`

Recommended severity mapping:

- `project_root_unavailable`: `error`
- `change_directory_missing`: `error`
- `required_file_missing`: `error`
- `tasks_file_unreadable`: `error`
- `tasks_not_found`: `error`
- `incomplete_task`: `warning`

Invalid structural findings should stop deeper review checks once the next checks would depend on missing state. For example, if the change directory is missing, do not also report every required file as missing. If required files are missing, do not read or parse `tasks.md`.

## Required Files

The review use case must use:

```text
domain.RequiredOpenSpecChangeFiles()
```

for required OpenSpec change file policy.

Do not duplicate the required file list in review-specific code for iteration or policy decisions. Do not put required file policy in the CLI adapter. Do not create a review-specific required-file list that can drift from generation and validation.

The first review requires every required file to exist directly under:

```text
openspec/changes/<change-id>/
```

Only file existence is required. Empty files and section quality are not semantically reviewed in this change.

## Task Parsing

Task parsing should be deterministic and line-based.

Recommended parser behavior:

- Accept the full `tasks.md` contents as a string.
- Split into lines.
- Trim leading whitespace before matching the task marker.
- Count a completed task when the trimmed line starts with `- [x]` or `- [X]`.
- Count an incomplete task when the trimmed line starts with `- [ ]`.
- Extract the task text by trimming the remaining text after the checkbox marker.
- Use the extracted task text in `incomplete_task` finding messages.
- Count nested task lines when leading whitespace is present and the trimmed line starts with a supported marker.
- Ignore lines that do not start with a supported marker after leading whitespace is removed.

No full Markdown parser is required. Do not perform semantic Markdown review beyond this checkbox counting. Do not inspect code fences, headings, acceptance criteria, risk sections, implementation files, or Git diffs.

If no task checkboxes are detected, return an invalid review result with a `tasks_not_found` finding.

If incomplete task text is empty, the finding message may use a stable fallback such as `Task is not completed.`. For ordinary tasks, the message should follow:

```text
Task is not completed: <task text>
```

## Domain Model

Add review domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- review status values;
- review finding severity values;
- review finding code values;
- review finding with severity, code, message, relative path, and optional subject;
- task summary with total, completed, and incomplete counts;
- review result containing change id, checked path, status, required files, task summary, and findings.

A possible result shape:

```text
ChangeID string
CheckedPath string
Status ReviewStatus
RequiredFiles []string
Tasks ReviewTaskSummary
Findings []ReviewFinding
```

The result should defensively copy slices when exposing required files or findings if helper constructors are used.

The domain package must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, or external process packages.

## Ports

Add a review-specific filesystem port under:

```text
internal/core/ports
```

Expected contract:

```text
DirectoryExists(root string, relativePath string) (bool, error)
FileExists(root string, relativePath string) (bool, error)
ReadFile(root string, relativePath string) (string, error)
```

Use a behavior-specific name such as `ReviewFileSystem`.

Do not reuse initialization, validation, generation, archive, prompt rendering, AI provider, workflow dispatcher, source-control, or agent contracts directly, even if the local filesystem adapter can satisfy overlapping methods.

The review port should contain only operations needed by this change.

## Use Case

Add a review use case under:

```text
internal/core/usecase
```

Expected input:

- project root;
- change id.

Expected behavior:

- validate that the use case dependency is present;
- trim and validate that project root is non-empty;
- trim and validate that change id is non-empty;
- reject unsafe change ids before performing filesystem checks;
- build the checked relative path as `openspec/changes/<change-id>`;
- load required files from `domain.RequiredOpenSpecChangeFiles()`;
- check OpenSpec project availability by verifying `openspec/project.md` as a file and `openspec/changes/` as a directory through the review filesystem port;
- return an invalid review result with `project_root_unavailable` when OpenSpec project structure is unavailable;
- check that `openspec/changes/<change-id>/` exists as a directory;
- return an invalid review result with `change_directory_missing` when the change directory is missing or is not a directory;
- check every required OpenSpec change file using the domain required-file policy;
- return one `required_file_missing` finding per missing required file;
- stop before reading `tasks.md` when required files are missing;
- read `openspec/changes/<change-id>/tasks.md` through the review filesystem port;
- return an invalid review result with `tasks_file_unreadable` when `tasks.md` cannot be read even though required file checks passed;
- parse task checkboxes deterministically;
- return an invalid review result with `tasks_not_found` when no task checkboxes are detected;
- return a needs-work review result with one `incomplete_task` warning per incomplete task;
- return an approved review result when no findings remain and all detected tasks are completed;
- return errors for dependency failures and filesystem execution failures that prevent checks from completing, except unreadable `tasks.md`, which is represented as a structured invalid result;
- never print, call `os`, access terminal IO, call provider APIs, call source-control APIs, run external tools, import adapters, or import workflow SDKs.

The use case should return structured invalid or needs-work results for review findings. The CLI is responsible for converting non-approved statuses into a non-zero process exit.

## Change ID Safety

The change id is used to build a path below `openspec/changes/`, so it must be constrained before filesystem checks.

Reject at least:

- empty or whitespace-only ids;
- `.` and `..`;
- ids containing `/`;
- ids containing `\`;
- ids containing `:`;
- ids with leading `-`;
- any value that would be interpreted as an absolute path or traversal-like input.

The implementation may reuse or extract existing safe change id validation behavior if that keeps use cases consistent. Do not broaden validation, generation, prompt, or archive behavior as part of this change unless a small shared private helper extraction is necessary and covered by regression tests.

## Extensibility Direction

Prepare for future review capabilities through structured inputs, task summaries, findings, and statuses, but do not add unused extension frameworks.

Acceptable first implementation shapes:

- a single review use case with small private helper methods;
- explicit review result, task summary, and finding domain types;
- a dedicated review filesystem port;
- deterministic line-based task parsing in core;
- direct CLI wiring similar to `validate` and `archive`.

Avoid:

- Git diff analyzers;
- source-control clients;
- provider clients;
- external-agent dispatchers;
- workflow connectors;
- architecture scanner packages;
- exported review chains, registries, factories, providers, strategies, or plugin abstractions;
- dry-run, format, AI, agent, GitHub, GitLab, or auto-fix flag plumbing.

Future changes can introduce additional collaborators once concrete review behavior requires them.

## Filesystem Adapter

Use `internal/adapters/filesystem` as the concrete implementation of the review filesystem port.

The local filesystem adapter must support:

- checking whether a directory exists;
- checking whether a file exists;
- reading a file as text.

Directory checks must distinguish directories from files. File checks must distinguish files from directories.

The adapter must not know:

- which OpenSpec project paths are required;
- which change files are required;
- how review status is calculated;
- how task checkboxes are counted;
- how review findings are structured;
- how CLI reports are formatted;
- future Git, AI, agent, workflow, source-control, archive-readiness, or semantic-review policy.

## CLI Adapter

Update `internal/adapters/cli` so the `review` command:

- parses `specharbor review <change-id>`;
- rejects missing change id;
- rejects unsupported flags;
- rejects extra positional arguments;
- rejects leading-dash change ids through flag parsing;
- obtains the current working directory for the project root;
- constructs the review use case with the local filesystem adapter;
- invokes the use case with project root and change id;
- prints a human-readable review report from the structured result;
- returns zero only when the result status is `approved`;
- returns a non-zero exit after printing the report when the result status is `needs-work` or `invalid`;
- returns argument and execution errors without panicking.

The CLI adapter may format human-readable output, but it must not contain required-file policy, task parsing policy, status calculation, filesystem policy, provider logic, source-control logic, workflow logic, architecture scanning, or semantic review.

`cmd/specharbor/main.go` should remain limited to existing process bootstrapping unless a minimal error-handling adjustment is strictly required.

## Testing Strategy

Add focused tests for:

- domain review status behavior;
- domain review result construction;
- defensive copies if required files or findings are represented as slices;
- task summary behavior;
- task parser counts completed `- [x]` tasks;
- task parser counts completed `- [X]` tasks;
- task parser counts incomplete `- [ ]` tasks;
- task parser ignores non-task lines;
- task parser extracts incomplete task text;
- task parser handles leading whitespace before task markers;
- use case returns approved when project structure exists, required files exist, `tasks.md` is readable, and all detected tasks are completed;
- use case returns needs-work when one or more detected tasks are incomplete;
- use case returns one `incomplete_task` warning per incomplete task;
- use case returns invalid with `project_root_unavailable` when `openspec/project.md` or `openspec/changes/` is unavailable;
- use case returns invalid with `change_directory_missing` when the active change directory is missing or is not a directory;
- use case returns invalid with one `required_file_missing` finding per missing required file;
- use case uses `domain.RequiredOpenSpecChangeFiles()` for required file checks;
- use case returns invalid with `tasks_file_unreadable` when `tasks.md` read fails;
- use case returns invalid with `tasks_not_found` when no checkbox tasks are detected;
- use case rejects empty project root;
- use case rejects empty change id;
- use case rejects unsafe change ids before filesystem checks;
- filesystem check errors are returned as errors where no structured finding can be produced safely;
- local filesystem adapter satisfies the review filesystem port;
- local filesystem adapter reads `tasks.md` contents;
- local filesystem adapter distinguishes files from directories for review checks;
- CLI prints approved, needs-work, and invalid reports in the expected shape;
- CLI returns zero for approved results;
- CLI returns non-zero after printing needs-work and invalid reports;
- CLI rejects missing change id, unsupported flags, extra positional arguments, and unsafe change ids;
- existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, `archive`, and unknown command behavior is preserved.

Use fake ports for use case tests. Use temporary directories for filesystem adapter and CLI integration-style tests.

## Validation

Run:

```text
gofmt
go test ./...
```

Do not require network access, provider credentials, local model credentials, source-control credentials, external-agent tools, workflow credentials, or external processes for this change.
