# Design: Implement Archive Foundation

## Overview

`specharbor archive <change-id>` moves a completed active OpenSpec change directory from `openspec/changes/` into a date-based archive directory under `openspec/archive/`.

This change establishes archive as a first-class use case without implementing completion checks, living spec updates, changelog generation, source-control integration, release notes, rollback, metadata files, or AI-assisted summaries. The implementation should be small, deterministic, and ready for future archive extensions without adding unused extension frameworks.

## CLI Contract

Supported command shape:

```text
specharbor archive <change-id>
```

Reject:

- `specharbor archive`
- `specharbor archive <change-id> extra`
- unsupported flags such as `--date`, `--force`, `--dry-run`, `--metadata`, `--summary`, `--github`, or `--gitlab`
- change ids that are parsed as flags, such as `-bad-id`

On argument errors, return an error from the CLI adapter so `cmd/specharbor/main.go` handles the existing error flow.

On success, print a human-readable report to stdout and return a zero exit code.

## Expected CLI Output

Output should be concise and deterministic. For a successful archive, the report must follow this shape:

```text
SpecHarbor change archived.
Change: implement-archive-foundation
Source: openspec/changes/implement-archive-foundation
Archive: openspec/archive/2026-06-06/implement-archive-foundation
Archive date: 2026-06-06
Moved: yes

Moved directory:
- openspec/changes/implement-archive-foundation -> openspec/archive/2026-06-06/implement-archive-foundation
```

Tests must verify that the report contains the operation status, change id, relative source path, relative archive path, archive date, `Moved: yes`, and moved directory source and destination.

Do not print absolute local paths, debug output, validation summaries, provider names, agent names, source-control details, or future metadata fields.

## Archive Date

The initial CLI command archives under the current local calendar date formatted as `YYYY-MM-DD`.

The archive use case should accept an archive date value as input so use case tests can be deterministic. The CLI adapter is responsible for deriving the current local date for `specharbor archive <change-id>` and passing the formatted value to the use case.

Archive date derivation in the CLI adapter must stay small and directly testable, for example as a small helper or a narrowly scoped formatting boundary. Do not introduce broad clock interfaces, global mutable clocks, or unused time abstractions for future date selection modes.

The use case must validate that the archive date is non-empty and conforms to `YYYY-MM-DD` before building any archive path. It does not need to validate calendar semantics beyond using a reliable date formatting/parsing approach from the standard library.

Do not add a user-facing `--date` flag in this change.

## Domain Model

Add archive domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- an archive result containing the change id, source path, archive path, archive date, and moved directory information;
- a small moved directory concept if it keeps result construction and reporting explicit;
- defensive copying for any slices if moved items are represented as a collection.

A possible result shape:

```text
ChangeID string
SourcePath string
ArchivePath string
ArchiveDate string
MovedDirectory ArchiveMovedDirectory
```

With moved directory information such as:

```text
SourcePath string
ArchivePath string
```

The first implementation moves one directory. Do not walk the source directory only to report individual files. Future changes can add moved file lists or archive metadata when those behaviors are implemented.

The domain package must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, or concrete filesystem packages.

## Ports

Add an archive-specific filesystem port under:

```text
internal/core/ports
```

Expected contract:

```text
DirectoryExists(root string, relativePath string) (bool, error)
FileExists(root string, relativePath string) (bool, error)
PathExists(root string, relativePath string) (bool, error)
CreateDirectory(root string, relativePath string) error
MoveDirectory(root string, sourceRelativePath string, destinationRelativePath string) error
```

Use a behavior-specific name such as `ArchiveFileSystem`. Do not reuse initialization, validation, or generation ports directly, even if the local filesystem adapter can satisfy overlapping methods.

The archive port should not mix archive behavior with validation, generation, prompt rendering, provider APIs, workflow dispatch, release generation, source-control integrations, or project initialization.

## Use Case

Add an archive use case under:

```text
internal/core/usecase
```

Expected input:

- project root;
- change id;
- archive date formatted as `YYYY-MM-DD`.

Expected behavior:

- validate that the use case dependency is present;
- trim and validate that project root is non-empty;
- trim and validate that change id is non-empty;
- trim and validate that archive date is non-empty and formatted as `YYYY-MM-DD`;
- reject unsafe change ids before performing filesystem writes or moves;
- build the source relative path as `openspec/changes/<change-id>`;
- build the archive date directory as `openspec/archive/<archive-date>`;
- build the archive relative path as `openspec/archive/<archive-date>/<change-id>`;
- check OpenSpec project availability by verifying that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory through the archive filesystem port;
- return a clear execution error for missing OpenSpec project structure telling the user to run `specharbor init` first;
- verify that the source change path `openspec/changes/<change-id>/` exists as a directory, not merely as any filesystem path;
- return a clear execution error when `openspec/changes/<change-id>/` is missing or exists but is not a directory;
- check whether `openspec/archive` exists before creating archive directories;
- return a clear execution error when `openspec/archive` exists but is not a directory;
- check whether `openspec/archive/<archive-date>` exists before moving;
- return a clear execution error when `openspec/archive/<archive-date>` exists but is not a directory;
- check whether the archive target path already exists as either a file or directory before moving;
- return a clear execution error when `openspec/archive/<archive-date>/<change-id>` already exists and do not move or overwrite it;
- create `openspec/archive/` when missing;
- create `openspec/archive/<archive-date>/` when missing;
- move the source change directory to the archive path through the archive filesystem port;
- return a structured archive result;
- never print, call `os`, access terminal IO, call provider APIs, call source-control APIs, run external tools, import adapters, or import workflow SDKs.

Missing OpenSpec project structure and a missing or non-directory source change path must stop archive execution before archive directory creation or movement. Archive parent paths that exist but are not directories, and existing archive destination paths, must stop archive execution before directory movement. Unsafe change ids must stop archive execution before any filesystem write or move.

If archive parent directories are created and the final move later fails, rollback is not required in this change. The source directory must remain under the active change path unless the move operation succeeds.

## Change ID Safety

The change id is used to build paths below both `openspec/changes/` and `openspec/archive/YYYY-MM-DD/`, so it must be constrained before any filesystem write or move.

Reject at least:

- empty or whitespace-only ids;
- `.` and `..`;
- ids containing `/`;
- ids containing `\`;
- ids containing `:`;
- ids with leading `-`;
- any value that would be interpreted as an absolute path or path traversal input.

The implementation may reuse or extract existing safe change id validation behavior if that keeps use cases consistent. Do not broaden existing validation or generation behavior as part of this change unless necessary for shared private helper extraction.

## Extensibility Direction

Prepare for future archive capabilities through structured archive inputs and results, but do not create unused public archive strategy registries, factories, chains, provider abstractions, source-control abstractions, workflow abstractions, release abstractions, or AI abstractions.

Acceptable first implementation shapes:

- a single archive use case with small private helpers;
- explicit archive result and moved directory domain types;
- a dedicated archive filesystem port;
- direct CLI wiring similar to `generate` and `validate`;
- local filesystem adapter methods needed by the archive port.

Avoid:

- completion validator chains;
- living spec mutation logic;
- changelog or release note generators;
- source-control host clients;
- external process execution;
- AI summary providers;
- archive rollback services;
- metadata writer abstractions;
- dry-run or force-mode plumbing.

Future changes can introduce additional archive collaborators once completion checks, spec updates, changelog generation, release notes, rollback, metadata, AI summaries, or external integrations are implemented.

## Filesystem Adapter

Use `internal/adapters/filesystem` as the concrete implementation of the archive filesystem port.

The local filesystem adapter must support:

- checking whether a directory exists;
- checking whether a file exists;
- checking whether any path exists;
- creating directories;
- moving a directory from one relative path under the project root to another relative path under the project root.

Directory checks must distinguish directories from files. Path existence checks must detect both files and directories so the use case can reject an existing archive destination path before moving.

The move operation should move the directory as a directory, not copy individual files for reporting. Since source and destination are under the same project root, using a standard directory rename operation is acceptable. A copy-and-delete fallback for cross-device moves is out of scope.

The move operation must not overwrite an existing destination path, whether that destination is a file or a directory. The use case performs the explicit destination check, and the adapter should remain defensive if the destination appears between the check and the move.

The adapter must not know:

- which OpenSpec project paths are required;
- how archive dates are chosen;
- how archive results are structured;
- how CLI reports are formatted;
- future completion, changelog, release, AI, source-control, or rollback policy.

## CLI Adapter

Update `internal/adapters/cli` so the `archive` command:

- parses exactly one change id argument;
- rejects missing change id;
- rejects unsupported flags;
- rejects extra positional arguments;
- obtains the current working directory as project root;
- derives the current local date formatted as `YYYY-MM-DD`;
- keeps date derivation small and testable without broad or unused clock abstractions;
- constructs the archive use case with the local filesystem adapter;
- invokes the use case with project root, change id, and archive date;
- prints a human-readable archive report from the structured result;
- returns argument and execution errors without panicking.

The CLI adapter may format human-readable output, but it must not contain project-structure policy, overwrite policy, source-directory policy, move policy, future archive completion policy, provider logic, source-control logic, or workflow logic.

`cmd/specharbor/main.go` should remain limited to existing process bootstrapping unless a minimal error-handling adjustment is strictly required.

## Testing Strategy

Add focused tests for:

- domain archive result behavior;
- defensive copies if moved directory or moved items are represented as slices;
- use case succeeds when OpenSpec project structure exists, the source change directory exists, archive directories are missing, and archive target is absent;
- use case creates `openspec/archive/` when missing;
- use case creates `openspec/archive/YYYY-MM-DD/` when missing;
- use case moves `openspec/changes/<change-id>/` to `openspec/archive/YYYY-MM-DD/<change-id>/`;
- use case returns change id, source path, archive path, archive date, and moved directory information;
- empty project root is rejected;
- empty change id is rejected;
- empty or malformed archive date is rejected;
- unsafe change ids are rejected before filesystem writes or moves;
- missing `openspec/project.md` is rejected before archive directory creation or movement;
- missing `openspec/changes/` is rejected before archive directory creation or movement;
- missing source change directory is rejected before archive directory creation or movement;
- source change path that exists as a file is rejected before archive directory creation or movement;
- `openspec/archive` that exists as a file is rejected before archive directory creation or movement;
- `openspec/archive/YYYY-MM-DD` that exists as a file is rejected before archive directory creation or movement;
- existing archive target path, as either a file or directory, is rejected before moving and does not overwrite existing archived content;
- filesystem check, create, and move errors are returned as errors;
- local filesystem adapter satisfies the archive filesystem port;
- local filesystem adapter moves directories and preserves nested contents;
- CLI report for successful archive contains status, change id, relative source path, relative archive path, archive date, `Moved: yes`, and moved directory line;
- CLI archive date derivation remains small enough to test directly without a broad clock abstraction;
- CLI rejects missing change id, unsupported flags, extra positional arguments, and unsafe change ids;
- existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, and unknown command behavior is preserved.

Use fake ports for use case tests. Use temporary directories for filesystem adapter and CLI integration-style tests.

## Validation

Run:

```text
go test ./...
```

Do not require network access, provider credentials, local model credentials, source-control credentials, external-agent tools, or external processes for this change.
