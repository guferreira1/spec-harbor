# Design: Implement Generation Foundation

## Overview

`specharbor generate <change-id> --blank` creates the standard OpenSpec change file package for manual authoring.

This change establishes generation as a first-class use case without implementing guided, template, AI-assisted, agent-assisted, or hybrid generation. The first mode is blank generation only. It should be structured so future generation modes can add behavior without moving filesystem policy into the CLI or mixing provider and agent concerns into the core.

## CLI Contract

Supported command shape:

```text
specharbor generate <change-id> --blank
```

The implementation may accept `--blank` before or after the change id if that keeps parsing simple and deterministic, but exactly one change id and exactly one `--blank` flag must be present.

Reject:

- `specharbor generate`
- `specharbor generate <change-id>`
- `specharbor generate --blank`
- `specharbor generate <change-id> --blank extra`
- `specharbor generate <change-id> --blank --blank`
- unsupported flags such as `--guided`, `--template`, `--ai`, `--agent`, `--hybrid`, `--format`, or `--force`
- change ids that are parsed as flags, such as `-bad-id`

On argument errors, return an error from the CLI adapter so `cmd/specharbor/main.go` handles the existing error flow.

On success, print a human-readable report to stdout and return a zero exit code.

## Expected CLI Output

Output should be concise and deterministic. For a newly generated change, the success report must follow this shape:

```text
SpecHarbor blank change generated.
Change: implement-generation-foundation
Path: openspec/changes/implement-generation-foundation
Directory: created
Created files: 5
Skipped existing files: 0

Created:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
```

When running the same command again, the success report must follow this shape:

```text
SpecHarbor blank change generated.
Change: implement-generation-foundation
Path: openspec/changes/implement-generation-foundation
Directory: existing
Created files: 0
Skipped existing files: 5

Skipped existing:
- proposal.md
- design.md
- tasks.md
- acceptance-criteria.md
- risks.md
```

When the change directory or some files already exist, the report must keep existing file contents unchanged, identify skipped files, and list created files only when files were created.

Tests must verify that the report contains the operation status, change id, relative change path, directory status, created file count, skipped existing file count, and the relevant filenames.

Do not print absolute local paths, debug output, provider names, agent names, or validation summaries.

## Required Files

Blank generation creates the required OpenSpec change files directly under the target change directory:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

The source of truth for this list must be the existing domain-level required OpenSpec change file policy when available, currently `domain.RequiredOpenSpecChangeFiles()`. Generation-specific code must not duplicate a separate required-file list for iteration or policy decisions.

Blank starter content may be keyed by these filenames, but the use case must obtain the files to create from the shared domain policy. The required-file policy must not live in the CLI adapter, filesystem adapter, template/content adapter, or generation-specific registry.

## Starter Content

Each generated file must contain useful starter Markdown content, not an empty file.

Minimum useful content:

- `proposal.md`: title plus sections for problem, goal, scope, out of scope, and success criteria.
- `design.md`: title plus sections for overview, architecture, technical decisions, testing strategy, and validation.
- `tasks.md`: title plus unchecked implementation phases that remind the implementer to read required context, keep scope limited, add domain/usecase/port/adapter/CLI/tests as needed, run verification, and update tasks only when implementation work is complete.
- `acceptance-criteria.md`: title plus verifiable outcome bullets.
- `risks.md`: title plus risk and mitigation sections.

Starter content should be generic enough for any future OpenSpec change, but specific enough that a user is not starting from blank files.

Generated `tasks.md` must not mark implementation tasks as complete.

## Domain Model

Add generation domain concepts under:

```text
internal/core/domain
```

Expected concepts:

- a generation mode value for blank generation, represented in the domain;
- generation item kind or file result concepts if they keep reporting clear;
- a generation result containing change id, relative change path, mode, created files, skipped existing files, and change-directory status.
- reuse of the existing domain-level required OpenSpec change file policy for the required blank change files.

A possible result shape:

```text
ChangeID string
Mode GenerationMode
ChangePath string
ChangeDirectoryCreated bool
CreatedFiles []string
SkippedExistingFiles []string
```

The domain package must not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, or concrete template engines.

## Ports

Add a generation-specific filesystem port under:

```text
internal/core/ports
```

Expected contract:

```text
DirectoryExists(root string, relativePath string) (bool, error)
FileExists(root string, relativePath string) (bool, error)
CreateDirectory(root string, relativePath string) error
WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error)
```

Use a behavior-specific name such as `GenerationFileSystem`. Do not reuse the initialization or validation port directly, even if the local filesystem adapter already happens to satisfy the same methods.

If starter content is kept outside the use case, add a small generation content port such as:

```text
BlankChangeContentFor(relativePath string) (string, error)
```

or:

```text
ContentFor(relativePath string) (string, error)
```

The content port should not mix generation with prompt rendering, validation, provider APIs, workflow dispatch, or project initialization.

## Use Case

Add a generation use case under:

```text
internal/core/usecase
```

Expected input:

- project root;
- change id;
- generation mode.

For this change, the only supported mode is blank generation.

Expected behavior:

- validate that the use case dependencies are present;
- trim and validate that project root is non-empty;
- trim and validate that change id is non-empty;
- validate that mode is `blank`;
- reject unsupported generation modes;
- reject unsafe change ids before performing filesystem writes;
- build the target relative path as `openspec/changes/<change-id>`;
- check OpenSpec project availability by verifying that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory through the generation filesystem port;
- return a clear execution error for missing OpenSpec project structure telling the user to run `specharbor init` first;
- never create `openspec/`, `openspec/project.md`, or `openspec/changes/`;
- create the target change directory when missing;
- continue without error when the target change directory already exists;
- obtain required filenames from the shared domain-level required OpenSpec change file policy;
- obtain starter content for each required file;
- write each required file with an if-absent operation;
- record created files when the write succeeds;
- record skipped existing files when the file already exists;
- return a structured generation result;
- never print, call `os`, access terminal IO, call provider APIs, import adapters, run external tools, or import workflow SDKs.

Missing OpenSpec project structure should stop generation before any target directory or file writes occur. This behavior belongs to generation execution, not project initialization.

If the target change directory already exists, generation must still attempt to create missing required files inside it and skip existing files. Partially existing change directories are recoverable: the use case must continue without error, create only missing required files, skip existing required files, and never overwrite existing file contents.

Filesystem execution failures and content-loading failures are execution errors.

## Change ID Safety

The change id is used to build a relative path under `openspec/changes/`, so it must be constrained before any filesystem write.

Reject at least:

- empty or whitespace-only ids;
- `.` and `..`;
- ids containing `/`;
- ids containing `\`;
- ids containing `:`;
- ids with leading `-`;
- any value that would be interpreted as an absolute path or path traversal input.

The implementation may enforce a narrower safe format, such as lowercase letters, numbers, and hyphens, if tests and error messages make that clear. A narrower format should not be introduced unless it is intentional product behavior.

## Extensibility Direction

Prepare for future generation modes through the domain generation mode and structured result, but do not create unused exported strategy, registry, factory, or chain abstractions in this change. Blank generation mode must be represented in the domain, but implementation scope remains only:

```text
specharbor generate <change-id> --blank
```

Acceptable first implementation shapes:

- a single generation use case with a private blank-generation helper;
- a directly used private mode dispatch that currently accepts only blank;
- required-file access through the existing shared domain-level required OpenSpec change file policy;
- a small content adapter for blank starter Markdown.

Avoid:

- provider or model abstractions;
- agent target abstractions;
- workflow connector abstractions;
- a public generation strategy registry with one implementation;
- factories or Chain of Responsibility structures for a single generation mode;
- template engines or rendering layers beyond what starter content requires;
- semantic Markdown generation or parsing.

Future changes can extract generation strategies once guided, template, AI-assisted, agent-assisted, or hybrid modes need different collaborators.

## Filesystem Adapter

Use `internal/adapters/filesystem` as the concrete implementation of the generation filesystem port.

The existing local filesystem adapter already has behavior close to what generation needs. If compatibility work is required, keep it limited to filesystem operations and tests.

File writes must remain defensive against overwrites. Prefer exclusive-create behavior so an existing file is not replaced even if a previous existence check raced.

The filesystem adapter must not know:

- which OpenSpec files are required;
- what blank file contents should be;
- how generation results are structured;
- how CLI reports are formatted;
- future generation mode policy.

## Template or Content Adapter

If generated content is not hardcoded in the use case, add or extend an adapter under:

```text
internal/adapters/templates
```

The adapter may use embedded files or constants for the five blank starter documents.

It must return deterministic content for each filename returned by the shared required OpenSpec change file policy:

- `proposal.md`
- `design.md`
- `tasks.md`
- `acceptance-criteria.md`
- `risks.md`

It must return a clear error for unknown blank file paths.

Do not introduce custom template selection, user-provided template paths, provider prompts, agent prompts, or semantic rendering in this change.

## CLI Adapter

Update `internal/adapters/cli` so the `generate` command:

- parses exactly one change id argument;
- parses exactly one `--blank` flag;
- rejects missing change id;
- rejects missing `--blank`;
- rejects unsupported flags;
- rejects duplicate `--blank`;
- rejects extra positional arguments;
- obtains the current working directory as project root;
- constructs the generation use case with concrete filesystem and content adapters;
- invokes the use case with blank mode;
- prints a human-readable report from the structured result;
- returns argument and execution errors without panicking.

The CLI adapter may format human-readable output, but it must not contain required-file policy, starter content, project-structure policy, overwrite policy, or future generation strategy decisions.

`cmd/specharbor/main.go` should remain limited to existing process bootstrapping unless a minimal error-handling adjustment is strictly required.

## Testing Strategy

Add focused tests for:

- domain generation result behavior, including defensive copies if slices are exposed;
- use case returns created files for a new blank change;
- use case creates the target change directory when missing;
- use case fills in missing files when the target change directory already exists;
- use case skips existing files without overwriting their contents;
- use case obtains required filenames from the shared domain-level required OpenSpec change file policy;
- use case rejects empty project root;
- use case rejects empty change id;
- use case rejects unsafe change ids before filesystem writes;
- use case rejects unsupported generation modes;
- use case returns a clear execution error telling the user to run `specharbor init` first when `openspec/project.md` is missing;
- use case returns a clear execution error telling the user to run `specharbor init` first when `openspec/changes/` is missing;
- use case does not create `openspec/`, `openspec/project.md`, or `openspec/changes/`;
- use case returns filesystem execution errors;
- use case returns content-loading errors;
- filesystem adapter satisfies the generation filesystem port;
- content adapter returns useful starter content for every required blank file;
- content adapter returns an error for unknown paths;
- CLI prints a successful report for a newly generated blank change;
- CLI reports skipped existing files when run against a partially existing change;
- CLI rejects missing change id, missing `--blank`, duplicate `--blank`, unsupported flags, extra positional arguments, and unsafe change ids;
- existing `help`, `version`, `init`, `prompt`, `validate`, and unknown command behavior remains intact.

Use temporary directories and fake ports where they keep tests deterministic.

## Validation

The Implementer Agent must run:

```text
gofmt
go test ./...
```
