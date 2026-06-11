# Acceptance Criteria: Implement Generation Foundation

- Running `specharbor generate <change-id> --blank` in an initialized SpecHarbor/OpenSpec project returns a structured generation result for blank generation.
- The structured result contains the requested change id, generation mode, relative change path, created files, skipped existing files, and change-directory status.
- Blank generation mode is represented in the domain.
- The required files created by blank generation come from the existing domain-level required OpenSpec change file policy, currently `domain.RequiredOpenSpecChangeFiles()`.
- Generation-specific code does not duplicate a separate required OpenSpec change file list for iteration or policy decisions.
- The command creates `openspec/changes/<change-id>/` when it does not exist.
- The command creates `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md` directly under the target change directory.
- Generated files contain useful starter Markdown content and are not empty.
- Generated `tasks.md` contains unchecked tasks only.
- Existing files are not overwritten.
- Existing files are reported as skipped existing files.
- Missing required blank files are created even when the target change directory already exists.
- Partially existing change directories are recoverable: the command continues without error, creates only missing required files, skips existing files, and never overwrites existing file contents.
- The CLI prints a human-readable success report for blank generation.
- The CLI report includes the change id, relative change path, change-directory status, created file count, skipped existing file count, and relevant filenames.
- The CLI report does not print absolute local paths, debug output, provider details, agent details, or validation summaries.
- The CLI success report for a newly generated change follows this output shape:

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

- The CLI success report when running the same command again follows this output shape:

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

- Running `specharbor generate` returns a clear argument error.
- Running `specharbor generate <change-id>` without `--blank` returns a clear argument error.
- Running `specharbor generate --blank` without a change id returns a clear argument error.
- Duplicate `--blank` flags are rejected.
- Unsupported flags are rejected.
- Extra positional arguments are rejected.
- Unsafe change ids are rejected before filesystem writes occur.
- Unsafe change ids include absolute-path input, dot segments, path separators, traversal-like input, colon-containing path input, and leading-dash input.
- The command obtains the project root from the current working directory in the CLI adapter.
- OpenSpec project availability is verified through the generation filesystem port by checking that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory.
- Missing OpenSpec project structure is rejected before the target change directory or files are created.
- Missing `openspec/project.md` or `openspec/changes/` returns a clear execution error telling the user to run `specharbor init` first.
- `specharbor generate <change-id> --blank` does not initialize a project and does not create `openspec/`, `openspec/project.md`, or `openspec/changes/`.
- Filesystem checks and writes are performed through a generation-specific port owned by `internal/core/ports`.
- Concrete filesystem behavior lives in `internal/adapters/filesystem`.
- Generation orchestration lives in `internal/core/usecase`.
- Generation concepts and result types live in `internal/core/domain`.
- Blank starter content comes from a small content/template dependency and not from CLI report formatting.
- The CLI adapter handles argument parsing, current-working-directory lookup, dependency construction, and human-readable report formatting only.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, or external processes.
- The implementation does not perform guided generation, custom template generation, AI-assisted generation, agent-assisted generation, hybrid generation, semantic generation, archive, review, scan, config, validation changes, or auto-fix behavior.
- The implementation does not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, network APIs, or external processes.
- The implementation does not require provider API keys, local model credentials, or agent credentials.
- The implementation does not add unused generation strategy registries, factories, chains, provider abstractions, AI abstractions, agent abstractions, or workflow abstractions.
- Existing `help`, `version`, `init`, `prompt`, `validate`, and unknown command behavior is preserved.
- Focused tests cover domain, use case, filesystem adapter compatibility, starter content, CLI behavior, and existing command regressions.
- `go test ./...` succeeds.
