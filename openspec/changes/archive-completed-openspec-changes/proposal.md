# Proposal: Archive Completed OpenSpec Changes

## Problem

SpecHarbor is preparing for the first public release, but some completed and merged OpenSpec changes are still active under `openspec/changes/`.

Leaving completed release-preparation work in the active change area makes it harder to distinguish release-blocking work from historical implementation records. This cleanup should archive only completed OpenSpec changes whose implementation has already landed.

## Goal

Define a safe, reviewable plan to archive completed OpenSpec changes before the first public release.

This change is repository hygiene only. It prepares the repository for the release task by moving completed OpenSpec change records into the existing archive/history flow, after implementation-time verification confirms the candidates are still active, completed, and already merged.

The Implementer Agent must re-check the active changes at implementation time before archiving anything. The candidate list in this proposal is an expectation, not permission to archive blindly.

## Expected Archive Candidates

The expected archival candidates are:

- `implement-release-versioning`
- `implement-goreleaser-release`
- `implement-install-channels`

Each candidate must be confirmed under `openspec/changes/` before archive execution if it is still active. Each candidate must also be confirmed as completed and already merged before it is archived.

## Scope

- Inspect the active OpenSpec changes under `openspec/changes/`.
- Confirm the expected candidates are still active before attempting to archive them.
- Confirm each candidate is completed and already merged.
- Inspect the existing archive command behavior before running it.
- Run the existing archive flow for only the completed target candidates that remain active.
- Verify archived records exist in the expected archive/history location.
- Verify no unrelated active OpenSpec change was removed.
- Verify live specs and repository state remain valid and reviewable.

## Out of Scope

- Changing CLI behavior.
- Changing release behavior.
- Publishing packages.
- Creating release tags.
- Creating GitHub Releases.
- Modifying install channels.
- Modifying GoReleaser configuration or release automation.
- Modifying package-manager configuration.
- Altering templates.
- Altering validation rules.
- Altering workflow behavior.
- Altering generated outputs.
- Implementing first public release behavior.
- Creating or modifying a Homebrew tap.
- Publishing npm packages.
- Adding Linux package or Windows package-manager support.
- Adding release signing, SBOM generation, or package publishing automation.
- Modifying production code.

## Success Criteria

The completed target changes are no longer active under `openspec/changes/`, their archived records exist in the expected archive/history location, unrelated active changes remain active, live specs remain valid, and required verification commands pass.

The repository remains ready for the follow-up public release task without creating tags, publishing packages, modifying install channels, or changing release automation.
