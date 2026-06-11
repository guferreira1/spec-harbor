# Acceptance Criteria: Implement Archive Foundation

- Running `specharbor archive <change-id>` in an initialized SpecHarbor/OpenSpec project returns a structured archive result.
- The structured result contains the requested change id, relative source path, relative archive path, archive date, and moved directory information.
- The archive date uses the current local date formatted as `YYYY-MM-DD`.
- The CLI date derivation remains small and testable without introducing broad or unused clock abstractions.
- Use case tests can pass a deterministic archive date without relying on the current clock.
- The source path is `openspec/changes/<change-id>`.
- The source path must exist as a directory, not merely as any filesystem path.
- The archive path is `openspec/archive/YYYY-MM-DD/<change-id>`.
- The command creates `openspec/archive/` when it does not exist.
- If `openspec/archive` exists but is not a directory, the command returns a clear execution error.
- The command creates `openspec/archive/YYYY-MM-DD/` when it does not exist.
- If `openspec/archive/YYYY-MM-DD` exists but is not a directory, the command returns a clear execution error.
- The command moves the active change directory from `openspec/changes/<change-id>/` to `openspec/archive/YYYY-MM-DD/<change-id>/`.
- Nested files and directories inside the active change directory are preserved by the move.
- The source change directory no longer exists after a successful archive.
- Existing archive destination paths are not overwritten.
- If `openspec/archive/YYYY-MM-DD/<change-id>` already exists as a file or directory, the command returns a clear execution error and leaves the active change directory in place.
- The CLI prints a human-readable success report for archive.
- The CLI report includes the change id, relative source path, relative archive path, archive date, `Moved: yes`, and moved directory line.
- The CLI report does not print absolute local paths, debug output, validation summaries, provider details, agent details, source-control details, release details, or metadata fields.
- The CLI success report follows this output shape:

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

- Running `specharbor archive` returns a clear argument error.
- Unsupported flags are rejected.
- Extra positional arguments are rejected.
- User-provided archive date flags are rejected.
- Force, dry-run, metadata, summary, GitHub, and GitLab flags are rejected.
- Unsafe change ids are rejected before filesystem writes or moves occur.
- Unsafe change ids include absolute-path input, dot segments, path separators, traversal-like input, colon-containing path input, and leading-dash input.
- The command obtains the project root from the current working directory in the CLI adapter.
- The use case rejects an unavailable or empty project root.
- OpenSpec project availability is verified through the archive filesystem port by checking that `openspec/project.md` exists as a file and `openspec/changes/` exists as a directory.
- Missing OpenSpec project structure is rejected before archive directory creation or movement.
- Missing `openspec/project.md` or `openspec/changes/` returns a clear execution error telling the user to run `specharbor init` first.
- `specharbor archive <change-id>` does not initialize a project and does not create `openspec/`, `openspec/project.md`, or `openspec/changes/`.
- Missing `openspec/changes/<change-id>/` returns a clear execution error.
- `openspec/changes/<change-id>` that exists as a file returns a clear execution error.
- Missing source change directories and non-directory source paths are rejected before archive directory creation or movement.
- `openspec/archive` that exists as a file returns a clear execution error before archive directory creation or movement.
- `openspec/archive/YYYY-MM-DD` that exists as a file returns a clear execution error before archive directory creation or movement.
- Filesystem checks, archive directory creation, and directory movement are performed through an archive-specific port owned by `internal/core/ports`.
- Concrete filesystem behavior lives in `internal/adapters/filesystem`.
- Archive orchestration lives in `internal/core/usecase`.
- Archive concepts and result types live in `internal/core/domain`.
- The CLI adapter handles argument parsing, current-working-directory lookup, current-date derivation, dependency construction, and human-readable report formatting only.
- Core packages do not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, external-agent tooling, workflow SDKs, source-control SDKs, external processes, or concrete filesystem packages.
- The implementation does not perform living spec updates, changelog generation, completion validation, task completion checks, git merge checks, GitHub/GitLab integration, release note generation, AI-assisted summaries, rollback, metadata creation, review, scan, config, validation changes, generation changes, prompt changes, or auto-fix behavior.
- The implementation does not call AI providers, local model APIs, provider SDKs, external agents, workflow tools, source-control host APIs, network APIs, or external processes.
- The implementation does not require provider API keys, local model credentials, agent credentials, source-control credentials, or workflow credentials.
- The implementation does not add unused archive strategy registries, factories, chains, provider abstractions, AI abstractions, agent abstractions, workflow abstractions, source-control abstractions, release abstractions, or rollback abstractions.
- Existing `help`, `version`, `init`, `prompt`, `validate`, `generate`, and unknown command behavior is preserved.
- Focused tests cover domain, use case, filesystem adapter compatibility, CLI behavior, and existing command regressions.
- `go test ./...` succeeds.
