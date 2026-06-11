# Design: Implement AI-Assisted Generation

## Overview

AI-assisted generation imports a local text file containing AI-authored OpenSpec content:

```text
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> [--overwrite]
```

The command is intentionally split from agent runner execution:

1. A user obtains AI-authored OpenSpec content through any external process.
2. The user saves only the strict SpecHarbor file blocks to a local file.
3. SpecHarbor reads that local file.
4. SpecHarbor parses deterministic file blocks.
5. SpecHarbor writes only approved OpenSpec files under `openspec/changes/<change-id>/`.
6. SpecHarbor validates the resulting change package.

SpecHarbor does not call a model, does not execute an agent, does not parse live runner output, and does not apply patches.

## Command Contract

Supported:

```text
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file>
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --overwrite
```

Rejected:

```text
specharbor generate <change-id> --ai-assisted
specharbor generate <change-id> --from-file <agent-output-file>
specharbor generate <change-id> --ai-assisted --execute
specharbor generate <change-id> --ai-assisted --agent codex
specharbor generate <change-id> --ai-assisted --type feature
specharbor generate <change-id> --ai-assisted --title "Title"
specharbor generate <change-id> --ai-assisted --summary "Summary"
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --blank
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --template feature
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --guided
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --agent-assisted
```

Rules:

- exactly one generation mode is allowed;
- `--from-file` is required with `--ai-assisted`;
- `--from-file` without `--ai-assisted` is rejected with a clear mode-specific error;
- `--overwrite` is accepted only with `--ai-assisted`;
- duplicate `--ai-assisted`, `--from-file`, or `--overwrite` flags are rejected;
- unsupported flags and extra positional arguments keep the existing CLI style;
- `--execute` remains tied only to `--agent-assisted` run-and-report behavior and is rejected with `--ai-assisted`.

## AI Output Format

Use strict line-oriented file blocks:

```text
---FILE: proposal.md---
# Proposal

...
---END FILE---
---FILE: design.md---
# Design

...
---END FILE---
---FILE: tasks.md---
# Tasks

## Phase 1

- [ ] Implement the approved work.
---END FILE---
---FILE: acceptance-criteria.md---
# Acceptance Criteria

- The approved behavior is observable.
---END FILE---
---FILE: risks.md---
# Risks

## Risks

- A concrete risk is identified.

## Mitigations

- A mitigation is defined.
---END FILE---
```

Parser rules:

- block start lines must exactly match `---FILE: <filename>---`;
- block end lines must exactly match `---END FILE---`;
- only the five required filenames from `domain.RequiredOpenSpecChangeFiles()` are allowed;
- filenames must be single file names, not paths;
- duplicate file blocks are rejected;
- unknown filenames are rejected;
- missing required blocks are rejected;
- empty or whitespace-only file content is rejected before writes;
- absolute paths are rejected;
- path traversal is rejected;
- nested paths are rejected;
- filename separators `/` and `\` are rejected;
- non-whitespace text outside file blocks is rejected;
- unclosed blocks are rejected;
- `---END FILE---` outside a block is rejected;
- parser output ordering follows `domain.RequiredOpenSpecChangeFiles()`, not source order;
- parser findings include stable codes, human-readable messages, filename when known, and line numbers where useful.

The parser must be deterministic and local. It must not execute content, interpret shell commands, interpret Markdown code blocks as commands, parse diffs, apply patches, fetch URLs, or expand paths from the AI output.

### Format Decision

Choose delimiter blocks instead of fenced Markdown blocks.

Rationale:

- OpenSpec Markdown can naturally contain fenced code examples; fenced `specharbor-file` blocks would make nested fences ambiguous.
- Exact delimiter lines are easy to parse with a small deterministic state machine.
- Reserving the exact `---END FILE---` line is a clear trade-off for the first version.

If generated file content needs the literal end marker line in the future, that requires a separate escaped-format change.

## Domain Model

Add domain concepts under `internal/core/domain`:

- AI-assisted generation mode remains `domain.AIAssistedMode`.
- Allowed generated filename policy based on `domain.RequiredOpenSpecChangeFiles()`.
- A parsed generated file block model with filename and content.
- An AI output parse result with parsed files and parse findings.
- Parse finding severity or status with stable codes such as:
  - `unknown_file_block`
  - `duplicate_file_block`
  - `missing_file_block`
  - `empty_file_content`
  - `path_traversal_file_name`
  - `absolute_file_name`
  - `nested_file_name`
  - `malformed_file_block`
  - `text_outside_file_block`
  - `unclosed_file_block`
- An AI-assisted generation result with change id, source path, target path, created directory status, generated files, skipped files, overwritten files, safety flags, and validation result.

Domain constructors and accessors must defensively copy slices and maps where applicable.

Domain code must not import adapters, CLI packages, `os`, `os/exec`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.

## Use Case Flow

Add a use case under `internal/core/usecase`, such as `GenerateAIAssistedChange`.

Flow:

1. Validate use case dependencies.
2. Trim and validate project root.
3. Parse `<change-id>` with `domain.NewChangeID`.
4. Validate `--from-file` is non-empty.
5. Require OpenSpec project structure: `openspec/project.md` and `openspec/changes/`.
6. Read the source AI output file through a port.
7. Parse the source content with the domain parser.
8. If parser findings contain errors, return a structured parse failure and perform no writes.
9. Resolve the target directory as `openspec/changes/<change-id>`.
10. If the target path exists and is not a directory, fail before writes.
11. If the target directory is missing, create it after parsing succeeds.
12. Preflight all target filenames using the allowed required file list.
13. If `--overwrite` is false, skip existing required files and write only missing required files.
14. If `--overwrite` is true, replace existing required files and write missing required files.
15. Never write a filename from AI output directly; write only the normalized required filenames in required-file order.
16. After all planned writes succeed, run existing validation for `<change-id>`.
17. Return a structured result containing write outcomes and validation result.

Invalid change ids must fail before source-file reads or target filesystem writes.

## Ports

Add small core-owned ports under `internal/core/ports`.

Expected source read contract:

```text
ReadSourceFile(path string) (string, error)
```

Expected target write contract:

```text
DirectoryExists(root string, relativePath string) (bool, error)
FileExists(root string, relativePath string) (bool, error)
PathExists(root string, relativePath string) (bool, error)
CreateDirectory(root string, relativePath string) error
WriteFileIfAbsent(root string, relativePath string, contents string) (bool, error)
WriteFile(root string, relativePath string, contents string) error
```

The exact interface names may differ, but the consuming use case must own the interfaces. The write side must only receive relative paths constructed by the use case from `openspec/changes/<change-id>/` plus known required filenames.

The source file path is user-provided and read-only. It may be absolute or relative, but it must be treated as a local filesystem path only. No adapter may fetch URLs or call remote services for `--from-file`.

## Write Behavior

Target path:

```text
openspec/changes/<change-id>/
```

Only these files may be written:

```text
proposal.md
design.md
tasks.md
acceptance-criteria.md
risks.md
```

Decisions:

- The command creates the change directory if it is missing.
- The command requires the OpenSpec project structure to exist.
- Existing files are skipped by default and reported under `Skipped existing`.
- Existing files are overwritten only with explicit `--overwrite` and reported under `Overwritten`.
- All parsing and target preflight checks run before any write.
- Parser errors, source read errors, unsafe change ids, target path errors, and unsupported flags write nothing.
- The first implementation does not require temporary files because the approved write set is small and temp filenames would expand the write surface.
- Runtime write failures are reported clearly. The implementation should avoid expected partial writes through preflight checks, but it must not delete or roll back user files automatically after a partial runtime failure.
- Validation runs only after the planned writes complete successfully.
- Validation warnings do not undo writes and do not make the command fail.
- Validation errors do not undo writes, but they are reported and make the command exit non-zero.

## Validation Integration

After successful writes, run the existing validation behavior for the same change id.

Expected output facts:

- validation status (`valid` or `invalid`);
- required file count;
- error count;
- warning count;
- grouped validation findings using the existing severity, code, message, and path format.

Exit behavior:

- Parser, source read, argument, or write errors exit non-zero.
- Malformed AI output exits non-zero and writes nothing.
- Validation warnings alone keep exit code `0`.
- Validation errors make the command exit non-zero after printing the generation and validation report.

The implementation must reuse the existing validation rule logic rather than duplicating validation checks inside the AI-assisted generation parser.

## Relationship With Agent Runner

This feature does not change agent runner safety boundaries.

The existing local runner may still produce stdout and stderr in run-and-report mode. SpecHarbor still does not parse or apply that live output.

Safe first model:

- A user may save AI-authored strict file blocks to a local file.
- The user then runs `specharbor generate <change-id> --ai-assisted --from-file <file>`.
- The command parses that file in a separate explicit generation mode.

Direct live application is out of scope:

```text
specharbor generate <change-id> --agent-assisted --execute --apply
```

If a future change adds live runner output application, it must define separate confirmation, provenance, validation, rollback, and test requirements.

## CLI Output

Successful parse/write output should include:

```text
SpecHarbor AI-assisted change generated.
Change: <change-id>
Source file: <agent-output-file>
Path: openspec/changes/<change-id>
Directory: created|existing
Overwrite: yes|no
Generated files: <n>
Skipped existing files: <n>
Overwritten files: <n>

Generated:
- proposal.md

Skipped existing:
- design.md

Overwritten:
- tasks.md

Validation:
Status: valid|invalid
Required files: 5
Errors: <n>
Warnings: <n>
...

Safety:
- Provider APIs called: no
- Remote AI services called: no
- Agent commands executed: no
- Production code modified: no
- Source-control commands run: no
- Auto-commit, auto-push, PR, merge, or archive: no
```

Parse failures should include:

- change id;
- source file;
- parse status;
- stable parse codes;
- messages;
- filenames and line numbers where available;
- statement that no files were written.

Output must be deterministic and must not include provider credentials, environment values, current timestamps, terminal colors, or nondeterministic ordering.

## Architecture

Layer responsibilities:

- `internal/core/domain` owns file block models, allowed generated file names, parse result concepts, generation result concepts, parse finding codes, and pure parser rules.
- `internal/core/usecase` owns orchestration, change id validation, source read sequencing, parsing, safe write decisions, validation invocation, and structured result assembly.
- `internal/core/ports` owns small read/write interfaces consumed by the use case.
- `internal/adapters/filesystem` owns local source-file reads and local target-file writes.
- `internal/adapters/cli` owns argument parsing, dependency wiring, output formatting, and exit-code mapping.

Core must not import:

- adapters;
- CLI packages;
- `os`;
- `os/exec`;
- terminal IO;
- provider SDKs;
- network APIs;
- source-control SDKs;
- workflow SDKs;
- external-agent SDKs.

No AI provider adapter, workflow connector adapter, source-control adapter, or live agent output applier is introduced in this change.

## Documentation

Implementation must update public docs in the same change:

- `README.md`;
- `docs/usage.md`;
- `docs/generation-modes.md`;
- any existing AI-assisted generation docs if present.

Docs must explain:

- what AI-assisted generation does;
- the exact `--ai-assisted --from-file` command;
- the strict file block format;
- the local from-file workflow;
- generated file scope;
- default skip behavior for existing files;
- `--overwrite` behavior;
- validation behavior and exit codes;
- examples;
- safety boundaries;
- no provider API integration;
- no remote AI service calls;
- no production code modification;
- no source-control automation.

## Testing Strategy

Domain tests:

- allowed generated file names;
- rejected unknown filenames;
- rejected duplicate file blocks;
- rejected path traversal;
- rejected absolute paths;
- rejected nested paths;
- rejected separators;
- parser success;
- parser malformed input;
- missing required file blocks;
- empty generated content behavior;
- parse finding codes and line numbers;
- generation result model;
- defensive copying.

Use case tests:

- valid AI output generates required OpenSpec files;
- malformed AI output writes nothing;
- unknown filenames write nothing;
- duplicate file blocks write nothing;
- missing required blocks write nothing;
- existing files are not overwritten by default;
- overwrite requires explicit flag;
- writes are limited to `openspec/changes/<change-id>/`;
- missing change directory is created only after successful parse;
- existing non-directory target path fails before writes;
- validation runs after successful generation;
- validation warnings are reported and keep success exit semantics;
- validation errors are reported and affect exit behavior;
- unsafe change id fails before source reads and filesystem writes.

Ports/adapters tests:

- source AI output file is read locally;
- missing source file is reported clearly;
- write only approved relative OpenSpec paths;
- prevent path traversal and absolute target paths;
- write errors are reported clearly;
- overwrite behavior is explicit.

CLI tests:

- command parsing;
- required `--from-file`;
- duplicate flags;
- unsupported flags;
- extra args;
- mode conflicts;
- success output;
- parse error output;
- skipped existing output;
- overwrite output;
- validation report integration;
- exit-code behavior.

Regression tests:

- existing blank generation remains unchanged;
- existing template generation remains unchanged;
- existing guided generation remains unchanged;
- existing agent-assisted dry-run remains unchanged;
- existing agent-assisted run-and-report remains unchanged;
- validate remains unchanged except being invoked by this new command;
- prompt, review, archive, config, scan, init, workflow, help, and version remain unchanged.

Architecture tests:

- core does not import adapters;
- core does not import CLI packages;
- core does not import `os` or `os/exec`;
- no provider, network, source-control, workflow, or external-agent SDKs are introduced in core;
- no provider API adapter is introduced by this change;
- no workflow connector adapter is introduced by this change;
- no source-control automation is introduced;
- no production code write path is introduced;
- no arbitrary path write path is introduced;
- no patch application abstraction is introduced.
