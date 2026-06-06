# Proposal: Implement Generation Foundation

## Problem

SpecHarbor can initialize a project, render agent prompts, and validate OpenSpec change structure, but `specharbor generate` is still a placeholder.

Generation is a core product capability. The project must support blank, guided, template, AI-assisted, agent-assisted, and hybrid generation over time. Without a clear generation foundation, future modes are likely to mix CLI parsing, file creation policy, template behavior, AI provider calls, and workflow-agent concerns in the wrong layers.

The first concrete capability should be intentionally small: create a blank OpenSpec change package that users can fill in manually.

## Goal

Implement the first generation capability:

```text
specharbor generate <change-id> --blank
```

The command creates a new OpenSpec change directory:

```text
openspec/changes/<change-id>/
```

The blank change must contain these files:

- `proposal.md`
- `design.md`
- `tasks.md`
- `acceptance-criteria.md`
- `risks.md`

Blank generation must reuse the existing domain-level required OpenSpec change file policy when available, such as `domain.RequiredOpenSpecChangeFiles()`, instead of duplicating the required file list in generation-specific code. That policy must not live in the CLI adapter.

The implementation must establish a generation foundation that is extensible for future modes while staying focused on blank OpenSpec change generation.

## Scope

- Replace the `generate` placeholder with real behavior for `specharbor generate <change-id> --blank`.
- Validate that exactly one change id is provided.
- Validate that `--blank` is provided.
- Reject unsupported flags.
- Reject extra positional arguments.
- Reject duplicate `--blank` flags.
- Obtain the project root from the current working directory in the CLI adapter.
- Validate that the project root is available.
- Validate OpenSpec project availability by verifying that `openspec/project.md` and `openspec/changes/` exist through the generation filesystem port.
- Return a clear execution error telling the user to run `specharbor init` first when OpenSpec project structure is missing.
- Do not create `openspec/`, `openspec/project.md`, or `openspec/changes/` from generation.
- Reject unsafe change ids, including absolute-path input, dot segments, and traversal-like input.
- Create `openspec/changes/<change-id>/` when it does not exist.
- Reuse the existing domain-level required OpenSpec change file policy for the files blank generation creates.
- Create all missing required blank OpenSpec files with useful starter Markdown content.
- Continue without error when `openspec/changes/<change-id>/` already exists.
- Skip existing required files and never overwrite existing file contents.
- Return a structured generation result containing the generation mode, change id, change path, created files, skipped existing files, and whether the change directory was created or already existed.
- Print a human-readable CLI report from the structured result.
- Keep generation orchestration in `internal/core/usecase`.
- Keep generation concepts and result types in `internal/core/domain`.
- Add a small generation-specific filesystem port under `internal/core/ports`.
- Add a small default blank-content port under `internal/core/ports` if content loading is kept outside the use case.
- Perform concrete filesystem behavior through `internal/adapters/filesystem`.
- Keep default blank file content in `internal/adapters/templates` if template loading or embedded content is used.
- Add focused tests for domain behavior, use case orchestration, filesystem adapter compatibility, template/default content, CLI parsing, CLI reporting, and regressions for existing commands.

## Out of Scope

- Guided generation.
- Template generation beyond built-in blank starter content.
- AI-assisted generation.
- Agent-assisted generation.
- Hybrid generation.
- Interactive prompts.
- Custom user templates.
- Provider integrations.
- External-agent integrations.
- Workflow connector integrations.
- Network calls.
- External process execution.
- OpenSpec project initialization.
- Project scanning.
- Semantic generation from a user idea.
- Auto-fix behavior.
- Archive, review, scan, config, or validation behavior changes.
- Modifying generated OpenSpec changes after creation.
- Machine-readable output formats.
- Updating the living architecture spec.
- Adding an exported generation strategy registry, factory, or Chain of Responsibility framework before multiple generation modes are implemented.
- Adding unused provider abstractions, agent abstractions, generation strategy registries, factories, or chains for future modes.

## Success Criteria

- Running `specharbor generate <change-id> --blank` in an initialized SpecHarbor/OpenSpec project creates `openspec/changes/<change-id>/` when missing.
- The command reuses the existing domain-level required OpenSpec change file policy for `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- The command creates missing required files with useful starter Markdown content.
- Existing files are not overwritten and are reported as skipped existing files.
- Running the command against a partially existing change directory creates only missing required files and skips existing required files.
- The use case returns a structured result with created files and skipped existing files.
- The CLI prints a concise human-readable generation report.
- Missing change id, missing `--blank`, unsupported flags, duplicate `--blank`, and extra positional arguments return clear CLI errors.
- Unsafe change ids are rejected before filesystem writes occur.
- Missing OpenSpec project structure is rejected before target change creation.
- Missing OpenSpec project structure returns a clear execution error telling the user to run `specharbor init` first.
- Generation follows Hexagonal Architecture: CLI parses and reports, the use case orchestrates, domain models generation concepts and results, ports define filesystem and content dependencies, and adapters perform concrete filesystem or template work.
- The implementation does not call AI providers, local model APIs, external agents, workflow tools, network APIs, or external processes.
- The implementation represents blank generation mode in the domain without adding unused generation strategy registries, factories, chains, provider abstractions, or AI abstractions.
- `go test ./...` succeeds.
