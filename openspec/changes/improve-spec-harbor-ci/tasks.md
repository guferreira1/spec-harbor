# Tasks: Improve SpecHarbor CI

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/improve-spec-harbor-ci/`.
- [x] Inspect the current workflow at `.github/workflows/ci.yml`.
- [x] Inspect `go.mod` to confirm the Go version source for CI setup.
- [x] Confirm this change is limited to SpecHarbor repository CI and does not define, generate, or assume CI behavior for user projects.
- [x] Confirm no CLI, domain, use case, adapter, or platform behavior is required for this infrastructure-only change.
- [x] Keep implementation focused on `.github/workflows/ci.yml` and this change's task updates.
- [x] Do not modify `cmd/`, `internal/core/`, `internal/adapters/`, or `internal/platform/` unless a specific CI reliability blocker is found and documented before editing.

## Phase 1: Workflow Inspection

- [x] Review existing workflow triggers.
- [x] Verify the workflow runs on pull requests.
- [x] Verify the workflow runs on pushes to `main`.
- [x] Review the existing Go setup step.
- [x] Identify whether the workflow duplicates the Go version instead of reading from `go.mod`.
- [x] Review the existing test step.
- [x] Identify whether the existing test command can use cached Go test results.
- [x] Check whether the workflow already verifies Go formatting.
- [x] Note any existing workflow complexity that should be removed or avoided.

## Phase 2: CI Workflow Update

- [x] Update `.github/workflows/ci.yml`.
- [x] Keep the workflow name clear and stable.
- [x] Keep the pull request trigger.
- [x] Keep the push-to-`main` trigger.
- [x] Keep a single straightforward job unless a concrete reliability issue requires otherwise.
- [x] Keep the job on `ubuntu-latest`.
- [x] Keep the checkout step using `actions/checkout@v4`.
- [x] Configure Go setup using `actions/setup-go@v5`.
- [x] Prefer `go-version-file: go.mod` for Go version selection.
- [ ] If `go-version-file: go.mod` is not used, document the concrete reason and use the version declared by `go.mod`.
- [x] Avoid adding unrelated workflow steps, external services, Docker services, release automation, or coverage upload.

## Phase 3: Formatting Check

- [x] Add a dedicated formatting verification step before tests.
- [x] Use `find . -name "*.go" -print0 | xargs -0 gofmt -l` or an equivalent small shell script to detect unformatted Go files.
- [x] Ensure the formatting step checks all `.go` files in the SpecHarbor repository.
- [x] Ensure the formatting step does not rewrite files.
- [x] Ensure the formatting step exits successfully when no unformatted files are found.
- [x] Ensure the formatting step exits non-zero when any unformatted files are found.
- [x] Ensure the formatting failure output includes the list of unformatted file paths.
- [x] Keep the formatting script small enough to read directly in the workflow.

## Phase 4: Test Check

- [x] Replace the CI test command with `go test -count=1 ./...`.
- [x] Ensure the test step runs after the formatting check.
- [x] Ensure test failures surface directly through normal `go test` output.
- [x] Do not add test wrappers, custom scripts, or extra tools unless a specific reliability issue requires them.
- [x] Do not add linting, coverage upload, release, deployment, scan, or user-project workflow checks.

## Phase 5: Local Verification

- [x] Run `find . -name "*.go" -print0 | xargs -0 gofmt -l` locally and confirm it prints no files.
- [x] Run `go test -count=1 ./...` locally.
- [x] Run `go test ./...` to satisfy the repository-wide verification rule.
- [x] Inspect `git diff -- .github/workflows/ci.yml openspec/changes/improve-spec-harbor-ci/tasks.md` for unintended changes.
- [x] Confirm no files under `cmd/`, `internal/core/`, `internal/adapters/`, or `internal/platform/` were modified.
- [ ] If any verification command cannot be run, leave its task unchecked and record the reason in the final implementation response.

## Phase 6: Task Updates

- [x] Update this `tasks.md` by checking off only the tasks completed during implementation.
- [x] Leave tasks unchecked for any work not performed.
- [x] Record any justified deviation from the design in the implementation response.
- [x] Confirm the final change remains scoped to `.github/workflows/ci.yml` and OpenSpec task updates.
