# Design: Implement Init Command

## Overview

`specharbor init` initializes the current working directory as a SpecHarbor/OpenSpec project.

The behavior belongs to a core use case. The CLI adapter should only determine the current working directory, call the use case, and format the structured result into user-facing output.

## Required Structure

The initializer must ensure these directories exist:

```text
openspec/
openspec/specs/
openspec/changes/
.specharbor/
.specharbor/rules/
```

The initializer must ensure these files exist:

```text
openspec/project.md
.specharbor/config.yml
.specharbor/rules/global.md
.specharbor/rules/spec-author.md
.specharbor/rules/implementer.md
.specharbor/rules/architecture-reviewer.md
.specharbor/rules/test-engineer.md
.specharbor/rules/change-reviewer.md
```

## Initialization Semantics

The command initializes the current working directory.

Existing files must not be overwritten by default. If a required file already exists, the initializer must skip it and keep its contents unchanged.

Existing directories are acceptable. Missing directories must be created.

The project is considered already initialized when every required directory and file exists. In that case the use case should return an `already initialized` result, and the CLI should print a clear success message such as:

```text
SpecHarbor project is already initialized.
```

If some required items already exist and others are missing, the initializer should create only the missing items and return a normal initialized result that includes created and skipped items.

## Core Use Case

Add a project initialization use case under:

```text
internal/core/usecase
```

The use case should:

- accept the target project root as input;
- define the required relative directories and files;
- ask a filesystem port whether items exist;
- create missing directories;
- write missing files only when absent;
- obtain file contents through a template/default-content port;
- return a structured result with status, created items, and skipped existing items;
- return errors instead of panicking.

The use case must not:

- call `os`, `filepath`, terminal IO, or environment APIs directly;
- print user-facing messages;
- know about CLI formatting;
- depend on concrete adapters.

## Ports

Add only the ports required by the initialization use case under:

```text
internal/core/ports
```

Expected small contracts:

- a filesystem port that can check existence, create directories, and write a file only when absent;
- a default-content or template port that provides the generated `project.md`, `config.yml`, and rule file contents.

Keep port method names behavior-specific and avoid broad interfaces that mix unrelated SpecHarbor features.

## Domain Values

Use simple domain or use-case result types for initialization status and created/skipped items. Do not add unrelated project, spec, provider, agent, or workflow abstractions.

Recommended statuses:

- initialized
- already initialized

## Filesystem Adapter

Add a concrete local filesystem adapter under:

```text
internal/adapters/filesystem
```

The adapter may use `os` and `path/filepath`.

File writes must be defensive against accidental overwrite. Prefer an exclusive-create operation for file creation so an existing file is not replaced even if the use case checked existence earlier.

## Default Template Adapter

Add concrete default initialization templates under an adapter package, for example:

```text
internal/adapters/templates
```

Default content may be embedded with `go:embed` so the installed CLI can initialize projects without relying on repository-relative source files at runtime.

The generated `.specharbor/config.yml` should be based on the existing example configuration shape and must not contain a real secret or API key. It may reference an environment variable name for provider credentials.

The generated rule files should match the repository's current default role rules unless the implementation documents a narrower generated default.

## CLI Adapter

Update `internal/adapters/cli` so the `init` command:

- obtains the current working directory;
- constructs the initialization use case with concrete adapters;
- calls the use case;
- prints a concise success message;
- prints a clear already-initialized message when appropriate;
- returns errors to the process entrypoint for non-successful failures.

The CLI adapter must not contain business rules such as the required OpenSpec file list or overwrite policy.

`cmd/specharbor/main.go` should remain limited to process bootstrapping unless wiring changes require minimal updates.

## Output

Output should be concise and deterministic.

For a new or partial initialization, include enough information for the user to know what happened. A suitable format is:

```text
SpecHarbor project initialized.
Created: <n>
Skipped existing: <n>
```

For an already initialized project:

```text
SpecHarbor project is already initialized.
```

Do not print secrets, absolute local paths, or verbose debug output.

## Testing Strategy

Add focused tests for:

- use case creates missing directories and files;
- use case skips existing files without changing contents;
- use case returns already initialized status when all required items exist;
- filesystem adapter preserves existing files;
- template adapter returns all required defaults;
- CLI `init` prints the expected message for first run and second run;
- existing `help`, `version`, and unknown command behavior remains intact.

Use temporary directories for integration-style filesystem and CLI tests.

## Validation

The Implementer Agent must run:

```text
gofmt
go test ./...
```
