# Design: Implement Validation Foundation

## Overview

`specharbor validate <change-id>` validates the minimum local structure of an OpenSpec change.

This change creates the foundation for future validators by introducing structured validation result concepts and a use case dedicated to validation orchestration. The first concrete validation rule is required-file presence. It does not parse Markdown, inspect file contents, call AI, or dispatch work to external tools.

## CLI Contract

Supported command shape:

```text
specharbor validate <change-id>
```

The command must accept exactly one positional argument.

Reject:

- `specharbor validate`
- `specharbor validate <change-id> extra`
- unsupported flags such as `specharbor validate --format json`

On argument errors, return an error from the CLI adapter so `cmd/specharbor/main.go` handles the exit through the existing error flow.

On validation success, print a human-readable report to stdout and return a zero exit code.

On validation failure, print the invalid validation report to stdout and then cause the CLI process to exit with a non-zero status code so `specharbor validate <change-id>` can be used in CI. Validation failures are represented as invalid validation results from the use case, not as use case execution errors.

The report format does not need to be machine-readable. It should clearly include:

- the validation status: `valid` or `invalid`;
- the change id;
- the checked change path;
- the required files checked or required file count;
- missing directory or missing file findings when present.

## Expected CLI Output

Valid output shape:

```text
SpecHarbor change is valid.
Change: implement-validation-foundation
Checked path: openspec/changes/implement-validation-foundation
Required files: 5
Findings: 0
```

Invalid output shape:

```text
SpecHarbor change is invalid.
Change: implement-validation-foundation
Checked path: openspec/changes/implement-validation-foundation

Findings:
- [error] required_file_missing: Missing required file: proposal.md
- [error] required_file_missing: Missing required file: risks.md
```

The exact finding subjects may vary by scenario. Tests must verify that valid and invalid reports follow this shape and contain the status, change id, checked path, required file count for valid results, findings count for valid results, and missing filenames where applicable.

## Required Files

The first validation rule checks for these files directly under the target change directory:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

Only file existence is required. Empty files, incomplete sections, unchecked tasks, and Markdown semantics are not validated in this change.

## Domain Model

Add validation domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- validation status values such as `valid` and `invalid`;
- validation finding severity, starting with `error`;
- validation finding code, message, relative path, and optional subject;
- validation result containing the change id, checked path, status, required files, and findings.

Useful finding codes for this change:

- `project_root_unavailable`
- `change_directory_missing`
- `required_file_missing`

The result should make it easy for future validators to append findings without changing the CLI contract every time a new validation rule is added.

The domain package must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, external APIs, or workflow tools.

## Ports

Add a small validation-specific filesystem port under:

```text
internal/core/ports
```

Expected contract:

```text
DirectoryExists(root string, relativePath string) (bool, error)
FileExists(root string, relativePath string) (bool, error)
```

The name should describe validation usage, for example `ValidationFileSystem`. Do not reuse a broad initialization interface just because the current concrete adapter supports the same methods.

The use case must depend on this port, not on `internal/adapters/filesystem`.

## Use Case

Add a validation use case under:

```text
internal/core/usecase
```

Expected input:

- project root;
- change id.

Expected behavior:

- validate that the use case dependencies are present;
- trim and validate that project root is non-empty;
- trim and validate that change id is non-empty;
- build the target relative path as `openspec/changes/<change-id>`;
- check OpenSpec project availability by verifying that `openspec/project.md` and `openspec/changes/` exist through the filesystem port;
- check that the target change directory exists;
- check that all required files exist;
- return a valid result when no findings are present;
- return an invalid result with structured findings when validation checks fail;
- return errors for dependency failures, invalid input, and filesystem execution errors;
- return structured invalid results, not execution errors, for missing OpenSpec project structure, missing change directories, and missing required files;
- never print, call `os`, access terminal IO, call provider APIs, or import adapters.

OpenSpec project availability means the use case receives a non-empty project root and the filesystem port can verify both of these under that root:

- `openspec/project.md` exists as a file;
- `openspec/changes/` exists as a directory.

If the CLI cannot obtain the current working directory, that remains a CLI execution error.

If OpenSpec project availability cannot be verified because `openspec/project.md` or `openspec/changes/` is missing, return an invalid result with a `project_root_unavailable` finding and do not continue into change-directory or required-file checks.

If the change directory is missing, return an invalid result with a `change_directory_missing` finding and do not list every required file as missing.

If the change directory exists, check each required file and add one `required_file_missing` finding per missing file.

## Extensibility Direction

Prepare for future Chain of Responsibility validators through the result model and use case structure, but do not create an exported validator chain, registry, factory, or placeholder interface unless it is directly used by this change.

Acceptable first implementation shapes:

- small private helper methods in the use case, such as `validateProjectRoot`, `validateChangeDirectory`, and `validateRequiredFiles`;
- an internal ordered list of directly used validation steps if that keeps the code simpler.

Avoid:

- unused exported validator interfaces;
- a validator registry with only one implementation;
- strategy/factory code that is not needed by this first command;
- AI validation abstractions;
- semantic Markdown parsers.

Future changes can extract the helper checks into a Chain of Responsibility once there are multiple independent validators.

## Filesystem Adapter

Use `internal/adapters/filesystem` as the concrete implementation of the validation filesystem port.

The adapter may already expose the required methods. If so, no business logic should be added to the adapter beyond any small compatibility work needed for the new port and tests.

The adapter must not know:

- which OpenSpec files are required;
- how validation findings are structured;
- how CLI reports are formatted.

## CLI Adapter

Update `internal/adapters/cli` so the `validate` command:

- parses exactly one change id argument;
- rejects missing change id, unsupported flags, and extra positional arguments;
- obtains the current working directory as project root;
- constructs the validation use case with the local filesystem adapter;
- invokes the use case;
- prints a human-readable report from the structured result;
- returns a non-zero exit code after printing an invalid validation report;
- returns argument and execution errors without panicking.

The CLI adapter may format human-readable output, but it must not contain validation business rules such as the required file list.

`cmd/specharbor/main.go` should remain limited to existing process bootstrapping unless a minimal error handling adjustment is strictly required.

## Testing Strategy

Add focused tests for:

- domain result status behavior;
- domain finding creation or helper behavior if helpers are added;
- use case returns valid when the project root, change directory, and all required files exist;
- use case returns invalid when OpenSpec project availability cannot be verified;
- use case returns invalid when the change directory is missing;
- use case returns one missing-file finding per missing required file;
- use case rejects empty project root;
- use case rejects empty change id;
- use case returns filesystem errors instead of converting them into validation findings;
- filesystem adapter satisfies the validation filesystem port;
- CLI prints a valid report for a complete change;
- CLI prints an invalid report listing missing required files;
- CLI returns a non-zero exit code after printing an invalid validation report;
- CLI rejects missing change id, unsupported flags, and extra positional arguments;
- existing `help`, `version`, `init`, `prompt`, and unknown command behavior remains intact.

Use temporary directories and fake ports where they keep tests deterministic.

## Validation

The Implementer Agent must run:

```text
gofmt
go test ./...
```
