# Risks: Archive Completed OpenSpec Changes

## Accidentally Archiving An Active Or Incomplete Change

Archiving an active or incomplete change would remove planning context that another agent still needs.

Mitigation: inspect `openspec/changes/`, verify the candidate exists, verify completed task state, verify merged-work evidence, and archive only the three approved target candidates that are still active and completed.

## Archive Command Updating Live Specs Unexpectedly

The existing archive command may update more than the change directory if behavior changes before implementation.

Mitigation: inspect CLI code, use case behavior, tests, help output, and documentation immediately before execution. Confirm whether live specs are updated, then validate live specs and inspect diffs after archive.

## Losing Traceability Of Completed Work

If archived records are missing or written to an unexpected location, completed work may become harder to audit.

Mitigation: confirm the archive destination/history location before execution, verify each archived record exists after execution, and inspect the resulting diff.

## Preparing Release With Stale Active Changes

If completed release-preparation changes remain active, the repository can look less ready for public release and reviewers may confuse completed work with pending work.

Mitigation: inspect active changes before and after archive, verify the target candidates are removed when archived, and confirm unrelated active changes remain untouched.

## Confusing Cleanup With Public Release

This hygiene task could be mistaken for the release task itself.

Mitigation: keep release boundaries explicit. Do not create tags, GitHub Releases, npm packages, Homebrew tap changes, install-channel changes, GoReleaser changes, or package-manager changes.

## Archiving A Change That Has Not Been Merged

Completed local task checklists do not prove the work was merged.

Mitigation: require merged-work confirmation for each target candidate before archive. If merge evidence cannot be confirmed, stop and report the candidate instead of archiving it.

## Archived Output Location Differing From Assumptions

The archive destination may differ from `openspec/archive/<date>/<change-id>/` if the command changes or project configuration affects it.

Mitigation: inspect the current archive implementation and docs before execution, follow actual project behavior, and verify archived records in the observed archive/history location.
