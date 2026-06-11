# Tasks: Archive Completed OpenSpec Changes

## Phase 0: Worktree And Baseline

- [x] Verify worktree and branch state.
- [x] Confirm the branch is `feat/archive-completed-openspec-changes`.
- [x] Inspect `git status --short` before implementation.
- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/archive-completed-openspec-changes/`.
- [x] Inspect current active changes under `openspec/changes/`.
- [x] Inspect the current archive/history location, if present.

## Phase 1: Candidate Confirmation

- [x] Confirm `implement-release-versioning` exists under `openspec/changes/` before archive, if still active.
- [x] Confirm `implement-goreleaser-release` exists under `openspec/changes/` before archive, if still active.
- [x] Confirm `implement-install-channels` exists under `openspec/changes/` before archive, if still active.
- [x] Confirm `implement-release-versioning` is completed and already merged before archive.
- [x] Confirm `implement-goreleaser-release` is completed and already merged before archive.
- [x] Confirm `implement-install-channels` is completed and already merged before archive.
- [x] Confirm no unrelated active change is selected for archive.

## Phase 2: Archive Behavior Inspection

- [x] Inspect existing archive command behavior before running it.
- [x] Inspect CLI help, workflow output, documentation, and code related to archive behavior.
- [x] Confirm archive destination/history location.
- [x] Confirm whether archive updates live specs.
- [x] Confirm whether validation should run before archive, after archive, or both.
- [x] Identify any relevant archive, list, workflow, config, or help command available for verification.
- [x] Verify no documentation outside the new OpenSpec change needs modification for this cleanup.

## Phase 3: Archive Execution

- [x] Run archive for `implement-release-versioning` if present and completed.
- [x] Run archive for `implement-goreleaser-release` if present and completed.
- [x] Run archive for `implement-install-channels` if present and completed.
- [x] Stop and inspect before continuing if any archive command fails.
- [x] Do not manually remove existing change directories.
- [x] Do not create tags, releases, packages, install-channel changes, or release automation changes.

## Phase 4: Post-Archive Verification

- [x] Verify archived changes were removed from active `openspec/changes/`.
- [x] Verify archived records exist in the expected archive/history location.
- [x] Verify no unrelated active change was removed.
- [x] Verify live specs were not corrupted.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate archive-completed-openspec-changes`.
- [x] Run any relevant archive, list, workflow, config, or help command if available.
- [x] Inspect `git status --short`.
- [x] Inspect relevant diffs.
- [x] Confirm repository is ready for `perform-first-public-release`.

## Phase 5: Task Update

- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
- [x] Leave any task unchecked if the work was not performed or verification could not run.
