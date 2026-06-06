# Acceptance Criteria: Implement Validation Foundation

- Running `specharbor validate <change-id>` against a change directory containing `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md` returns a structured validation result with status `valid`.
- A valid validation result contains the requested change id, the checked relative change path, the required file list, and no findings.
- The CLI prints a human-readable valid report for a complete change.
- The valid CLI report follows this output shape:

```text
SpecHarbor change is valid.
Change: implement-validation-foundation
Checked path: openspec/changes/implement-validation-foundation
Required files: 5
Findings: 0
```

- Running `specharbor validate <change-id>` against an existing change directory with missing required files returns a structured validation result with status `invalid`.
- Missing required files are represented as structured findings with a stable missing-file code.
- The CLI invalid report lists the missing required filenames.
- The invalid CLI report follows this output shape:

```text
SpecHarbor change is invalid.
Change: implement-validation-foundation
Checked path: openspec/changes/implement-validation-foundation

Findings:
- [error] required_file_missing: Missing required file: proposal.md
- [error] required_file_missing: Missing required file: risks.md
```

- Invalid validation results print the validation report and cause the CLI process to exit with a non-zero status code.
- Running `specharbor validate <change-id>` against a missing change directory returns a structured validation result with status `invalid`.
- The missing change directory report identifies `openspec/changes/<change-id>`.
- Running `specharbor validate` returns a clear argument error through the existing error flow.
- Unsupported flags and extra positional arguments are rejected.
- The command obtains the project root from the current working directory in the CLI adapter.
- OpenSpec project availability is verified through the filesystem port by checking that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory.
- Filesystem checks are performed through a port owned by `internal/core/ports`.
- The concrete filesystem implementation lives in `internal/adapters/filesystem`.
- Validation orchestration lives in `internal/core/usecase`.
- Validation status, result, and finding concepts live in `internal/core/domain`.
- Missing OpenSpec project structure, missing change directories, and missing required files are returned as structured invalid validation results, not use case execution errors.
- The CLI adapter handles argument parsing and human-readable report formatting only.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, or workflow SDKs.
- The implementation does not perform semantic Markdown validation.
- The implementation does not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, or external processes.
- The implementation does not require provider API keys or agent credentials.
- The implementation does not add generate, archive, review, scan, config, provider, or workflow integration behavior.
- Existing `help`, `version`, `init`, `prompt`, and unknown command behavior is preserved.
- Focused tests cover domain, use case, filesystem adapter compatibility, and CLI behavior.
- `go test ./...` succeeds.
