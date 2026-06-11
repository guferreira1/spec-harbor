# Acceptance Criteria: Implement Review Foundation

- Running `specharbor review <change-id>` against a complete active OpenSpec change returns a structured review result.
- The structured result contains the requested change id, checked relative path, review status, required files, task counts, and findings.
- The checked path is `openspec/changes/<change-id>`.
- Review status values are `approved`, `needs-work`, and `invalid`.
- Review finding severity values are `error` and `warning`.
- Initial review finding codes are `project_root_unavailable`, `change_directory_missing`, `required_file_missing`, `tasks_file_unreadable`, `tasks_not_found`, and `incomplete_task`.
- Required OpenSpec change files are loaded from `domain.RequiredOpenSpecChangeFiles()`.
- Required file policy is not duplicated in review-specific code for iteration or policy decisions.
- Required file policy does not live in the CLI adapter.
- OpenSpec project availability is verified through the review filesystem port by checking that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory.
- Missing OpenSpec project structure returns status `invalid` with a `project_root_unavailable` finding.
- The active change path must exist as a directory.
- Missing or non-directory active change paths return status `invalid` with a `change_directory_missing` finding.
- Missing required change files return status `invalid` with one `required_file_missing` finding per missing file.
- Required-file failures stop before reading or parsing `tasks.md`.
- `tasks.md` is read through the review filesystem port.
- An unreadable `tasks.md` returns status `invalid` with a `tasks_file_unreadable` finding.
- Task parsing is deterministic and line-based.
- Completed tasks are counted for `- [x]` and `- [X]`.
- Incomplete tasks are counted for `- [ ]`.
- Non-task lines are ignored by the task parser.
- `tasks.md` with no detected task checkboxes returns status `invalid` with a `tasks_not_found` finding.
- A complete change with detected tasks all completed returns status `approved`.
- A complete change with one or more incomplete detected tasks returns status `needs-work`.
- Each incomplete task produces one `incomplete_task` warning finding.
- Incomplete task finding messages include the incomplete task text when task text is present.
- The CLI prints an approved report for approved reviews.
- The approved CLI report follows this output shape:

```text
SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: approved
Tasks: 12 total, 12 completed, 0 incomplete
Findings: 0
```

- The CLI prints a needs-work report for reviews with incomplete tasks.
- The needs-work CLI report follows this output shape:

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

- The CLI prints an invalid report for invalid review results.
- The invalid CLI report follows this output shape:

```text
SpecHarbor review completed.
Change: implement-review-foundation
Checked path: openspec/changes/implement-review-foundation
Status: invalid
Tasks: 0 total, 0 completed, 0 incomplete

Findings:
- [error] required_file_missing: Missing required file: risks.md
```

- Approved review results cause the CLI process to exit zero.
- Needs-work review results print the report and cause the CLI process to exit non-zero.
- Invalid review results print the report and cause the CLI process to exit non-zero.
- Running `specharbor review` returns a clear argument error.
- Unsupported flags are rejected.
- Extra positional arguments are rejected.
- Unsafe change ids are rejected before filesystem checks.
- Unsafe change ids include absolute-path input, dot segments, path separators, traversal-like input, colon-containing path input, and leading-dash input.
- The command obtains the project root from the current working directory in the CLI adapter.
- Filesystem checks and file reads are performed through a review-specific port owned by `internal/core/ports`.
- Concrete filesystem behavior lives in `internal/adapters/filesystem`.
- Review orchestration lives in `internal/core/usecase`.
- Review status, result, task summary, and finding concepts live in `internal/core/domain`.
- The CLI adapter handles argument parsing, current-working-directory lookup, dependency construction, human-readable report formatting, and exit signaling only.
- CLI report formatting does not live in the use case.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, concrete filesystem packages, or external process packages.
- The implementation does not inspect Git diffs.
- The implementation does not call Git commands.
- The implementation does not call GitHub or GitLab APIs.
- The implementation does not perform semantic code review.
- The implementation does not perform semantic Markdown review beyond deterministic task checkbox counting.
- The implementation does not perform AI-assisted or agent-assisted review.
- The implementation does not implement architecture boundary scanning.
- The implementation does not enforce archive readiness.
- The implementation does not update files or auto-fix findings.
- The implementation does not change validation, generation, archive, prompt, scan, or config behavior.
- The implementation does not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, source-control host APIs, network APIs, or external processes.
- The implementation does not require provider API keys, local model credentials, agent credentials, source-control credentials, or workflow credentials.
- The implementation does not add unused review strategy registries, factories, chains, provider abstractions, AI abstractions, agent abstractions, workflow abstractions, source-control abstractions, scanner abstractions, or auto-fix abstractions.
- Existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, `archive`, and unknown command behavior is preserved.
- Focused tests cover domain behavior, task parsing, use case orchestration, filesystem adapter compatibility, CLI behavior, CLI exit behavior, and existing command regressions.
- `go test ./...` succeeds.
