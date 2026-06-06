# Risks: Implement Archive Foundation

## Architecture leakage

Archive touches CLI parsing, project-root discovery, archive date selection, project-structure checks, directory creation, overwrite prevention, directory movement, and user-facing reporting. The main risk is placing archive rules directly in the CLI adapter or filesystem adapter.

Mitigation:

- Keep CLI responsibilities limited to argument parsing, current-working-directory lookup, current-date derivation, dependency construction, and report formatting.
- Keep archive orchestration in `internal/core/usecase`.
- Keep archive concepts and structured results in `internal/core/domain`.
- Use an archive-specific filesystem port from `internal/core/ports`.
- Keep concrete filesystem behavior in `internal/adapters/filesystem`.
- Keep OpenSpec path policy, overwrite policy, and result construction out of the filesystem adapter.

## Overbuilding future archive behavior

SpecHarbor must eventually support completion checks, living spec updates, changelogs, release notes, source-control integrations, AI-assisted summaries, rollback, and metadata. Building strategy registries, chains, provider frameworks, source-control clients, release generators, or rollback services for the first archive move would add unused surface area.

Mitigation:

- Implement only `specharbor archive <change-id>`.
- Return structured archive data that future features can extend.
- Avoid exported archive strategy, registry, factory, Chain of Responsibility, provider, source-control, workflow, release, AI, metadata, or rollback abstractions until concrete behavior needs them.
- Do not add unsupported CLI flags or plumbing for future modes.

## Underbuilding the foundation

A simple CLI function that calls a filesystem rename would satisfy the first movement behavior but would make future archive capabilities harder to add cleanly.

Mitigation:

- Introduce domain result concepts for archived changes.
- Add an archive-specific filesystem port.
- Keep archive execution in a use case with focused input validation and path construction.
- Return moved directory information as structured data instead of deriving behavior from printed output.
- Cover use case behavior with fake ports so orchestration remains independent of local filesystem behavior.

## Accidental overwrite

Users may already have an archive destination path at `openspec/archive/YYYY-MM-DD/<change-id>`. Overwriting it would destroy historical records, whether the existing destination is a file or a directory.

Mitigation:

- Check whether the destination path exists as either a file or directory before moving.
- Return a clear error if the archive destination path already exists.
- Do not add `--force` in this change.
- Make the local filesystem move defensive against existing destinations.
- Add tests proving existing archived content is preserved and the active change directory remains in place when the archive target exists.

## Archive parent path type safety

Archive parent paths may already exist as files due to user error or unrelated tooling. Treating those paths as directories would produce confusing errors or unsafe movement attempts.

Mitigation:

- Check `openspec/archive` before creating archive directories and return a clear execution error if it exists but is not a directory.
- Check `openspec/archive/YYYY-MM-DD` before moving and return a clear execution error if it exists but is not a directory.
- Keep parent path type checks in the use case through the archive filesystem port.
- Add tests for archive root and archive date paths that exist as files.

## Unsafe change id path handling

The change id is used to build both source and destination paths. Absolute paths, path separators, dot segments, drive prefixes, or traversal-like input could move files outside the intended OpenSpec directories.

Mitigation:

- Validate the change id before any filesystem write or move.
- Reject empty ids, `.`, `..`, `/`, `\`, `:`, leading `-`, absolute-path input, and traversal-like input.
- Build source and destination paths only as `openspec/changes/<change-id>` and `openspec/archive/<archive-date>/<change-id>`.
- Add tests that unsafe ids are rejected before archive directories are created or moves occur.

## Archive date ambiguity

The archive path includes a date. If date formatting is inconsistent, archive paths will be hard to predict and tests may become brittle.

Mitigation:

- Use the current local calendar date for the CLI command.
- Format the date as zero-padded `YYYY-MM-DD`.
- Pass the formatted archive date into the use case so use case tests can use deterministic dates.
- Keep CLI date derivation small and directly testable without adding broad or unused clock abstractions.
- Validate the archive date format in the use case.
- Do not add user-provided archive date flags in this change.

## Project initialization ambiguity

Archive depends on an existing OpenSpec project. If the command silently creates `openspec/project.md` or `openspec/changes/`, it would blur the boundary between `init` and `archive`.

Mitigation:

- Verify `openspec/project.md` and `openspec/changes/` through the archive filesystem port.
- Return a clear execution error telling the user to run `specharbor init` first when project structure is unavailable.
- Do not create missing project-level OpenSpec structure in the archive use case.
- Leave project initialization to `specharbor init`.

## Source change loss

A failed archive move could leave users unsure whether the active change still exists. A source path that exists as a file instead of a directory could also make movement behavior ambiguous. Rollback is out of scope, but the implementation should avoid destructive partial behavior before the final move.

Mitigation:

- Perform all validation and overwrite checks before moving.
- Verify that `openspec/changes/<change-id>/` exists as a directory, not merely as any filesystem path.
- Return a clear execution error when the source path is missing or is not a directory.
- Create only archive parent directories before the move.
- Move the directory as the final filesystem operation.
- Return filesystem move errors clearly.
- Do not delete the source directory manually.
- Keep archive rollback out of scope for this change and document any residual behavior in tests where practical.

## Report format churn

Human-readable reports can become fragile if they rely on decorative formatting or incidental wording.

Mitigation:

- Keep output concise and deterministic.
- Test for important content: operation status, change id, source path, archive path, archive date, `Moved: yes`, and moved directory line.
- Avoid banners, absolute local paths, debug output, validation summaries, provider output, source-control output, and unrelated metadata.

## Accidental behavior changes

Archive is adjacent to validation and generation because all three operate on OpenSpec change directories. Broad refactors could accidentally change existing commands.

Mitigation:

- Keep this change scoped to archive.
- Do not modify validation, generation, prompt, init, review, scan, or config behavior unless a minimal shared helper extraction is required.
- Preserve existing tests and add regression coverage for existing CLI commands.
- Run `go test ./...`.
