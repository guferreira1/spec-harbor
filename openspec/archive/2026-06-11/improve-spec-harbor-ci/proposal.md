# Proposal: Improve SpecHarbor CI

## Problem

The SpecHarbor repository GitHub Actions CI workflow has been unreliable and has failed repeatedly even when local verification passes.

The current workflow is small, but it does not fully match the repository's expected local verification behavior. In particular, it does not check Go formatting and it runs tests with the default Go test cache behavior.

Because SpecHarbor is a Go CLI used to support OpenSpec-based workflows for coding agents, its own repository CI should be deterministic, simple, and easy to compare with local verification commands.

## Goal

Update the SpecHarbor repository's own GitHub Actions CI workflow so it:

- runs on pull requests;
- runs on pushes to `main`;
- sets up Go from the version declared by the project, preferably `go.mod`;
- fails when any `.go` file in the SpecHarbor repository is not gofmt-formatted;
- prints the list of unformatted Go files when formatting fails;
- runs tests without relying on cached test results;
- keeps failure output direct and actionable;
- remains simple and maintainable.

The expected test command is:

```text
go test -count=1 ./...
```

## Scope

- Update `.github/workflows/ci.yml`.
- Keep the workflow focused on checkout, Go setup, formatting verification, and uncached tests.
- Use `actions/checkout@v4`.
- Use `actions/setup-go@v5`.
- Configure Go setup from `go.mod` when supported by the action.
- Add a repository-wide formatting check based on `find . -name "*.go" -print0 | xargs -0 gofmt -l` or an equivalent small shell script.
- Make the formatting check inspect all `.go` files in the SpecHarbor repository.
- Make the formatting check detect unformatted files without rewriting files.
- Make the formatting check fail when any unformatted files are found.
- Make the formatting check print the unformatted file list before failing.
- Make the formatting check exit zero when no unformatted files are found.
- Change the test step to run `go test -count=1 ./...`.
- Run local verification after the workflow update.
- Update this change's `tasks.md` as implementation work is completed.

## Out of Scope

- Changing CLI behavior.
- Changing domain, use case, adapter, platform, or command code.
- Adding new SpecHarbor product commands.
- Adding user-project CI detection.
- Defining or generating CI behavior for user projects.
- Adding scan or workflow connector behavior.
- Adding lint tools that require extra installation.
- Adding release automation.
- Adding coverage upload.
- Adding complex matrix builds.
- Adding deployment.
- Adding external services.
- Adding Docker-based CI.
- Changing OpenSpec workflow behavior.
- Updating the living architecture spec.

## Success Criteria

- Pull requests run the SpecHarbor CI workflow.
- Pushes to `main` run the SpecHarbor CI workflow.
- CI sets up Go using the version declared by the SpecHarbor project.
- CI fails when any `.go` file in the SpecHarbor repository is not gofmt-formatted.
- CI checks formatting without rewriting files.
- Formatting failures print the unformatted file paths.
- CI formatting checks exit zero when no unformatted files are found.
- CI runs `go test -count=1 ./...`.
- The workflow remains limited to straightforward repository verification steps.
- No `cmd/`, `internal/core/`, `internal/adapters/`, or `internal/platform/` files are changed for this infrastructure-only change.
