# Design: Archive Completed OpenSpec Changes

## Overview

This change plans a narrow repository hygiene step before the first public release. It does not add product behavior. It uses the existing archive flow only after the Implementer Agent confirms the candidate changes are still active, completed, and already merged.

Current project documentation and code indicate that `archive <change-id>` moves a completed change from `openspec/changes/<change-id>/` to `openspec/archive/<date>/<change-id>/`. The current archive command accepts a single change id and rejects flags such as `--help`, so implementation must inspect current code and documentation again rather than depending on this observation.

## Archive Candidate Discovery

The Implementer Agent must discover active completed changes immediately before archive execution.

Discovery steps:

- Inspect `openspec/changes/`.
- Confirm `implement-release-versioning` exists before archiving it, if it is still active.
- Confirm `implement-goreleaser-release` exists before archiving it, if it is still active.
- Confirm `implement-install-channels` exists before archiving it, if it is still active.
- Confirm each present candidate has completed implementation evidence, including completed task checklists and relevant merged-work evidence.
- Confirm each present candidate was already merged before archiving it.
- Confirm no unrelated active change is selected for archive.
- Treat a missing expected candidate as a signal to inspect whether it was already archived or moved; do not recreate or remove it.

Presence under `openspec/changes/` is not enough. The archive action is allowed only for candidates that are active, completed, and already merged.

## Archive Command Behavior

Before executing the archive command, the Implementer Agent must inspect the current archive behavior in the repository.

The inspection must answer:

- What command interface is currently supported.
- What archive destination or history location is used.
- Whether the command creates the archive root or date directory.
- Whether the command updates live specs or only moves the change directory.
- Whether validation should run before archive, after archive, or both.
- Whether any relevant archive, list, workflow, config, or help command is available for verification.

The implementation must not assume behavior without checking code and documentation. If the command interface differs from the examples below, the implementation must follow the actual project behavior.

## Archive Execution

The likely execution plan is:

```bash
go run ./cmd/specharbor archive implement-release-versioning
go run ./cmd/specharbor archive implement-goreleaser-release
go run ./cmd/specharbor archive implement-install-channels
```

Execution rules:

- Run the archive flow only for candidates that are present, completed, and already merged.
- Run candidates one at a time so failures identify the exact change.
- Stop and inspect before continuing if any archive command fails.
- Do not run unsupported flags or force options.
- Do not remove directories manually.
- Do not modify live specs manually unless the inspected archive behavior explicitly requires and documents it.
- Do not perform source-control automation, release automation, package publishing, tag creation, or install-channel changes.

## Post-Archive Validation

After archive execution, verify:

- Target candidates archived during the implementation are removed from `openspec/changes/`.
- Archived records exist in the expected archive/history location.
- No unrelated active change was removed from `openspec/changes/`.
- Live specs were not corrupted.
- Repository diffs contain only the intended archive movement and the active `archive-completed-openspec-changes` task update.
- `go test ./...` passes.
- `go run ./cmd/specharbor validate archive-completed-openspec-changes` passes.
- Any relevant archive, list, workflow, config, or help command identified during inspection works as expected.

## Release Boundary

This cleanup is not the first public release.

The implementation must not:

- Create a release tag.
- Create a GitHub Release.
- Publish npm packages.
- Create or modify a Homebrew tap.
- Modify install-channel behavior.
- Modify GoReleaser configuration.
- Modify package-manager configuration.
- Modify release automation.
- Polish release docs unless explicitly required by the implementation of this cleanup.
- Change templates, validation rules, workflow behavior, generated outputs, or CLI behavior.
