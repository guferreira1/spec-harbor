# Risks: Improve SpecHarbor CI

## Workflow still drifts from project Go version

Hardcoding the Go version in `.github/workflows/ci.yml` can drift from `go.mod` and create confusing local-versus-CI differences.

Mitigation:

- Use `actions/setup-go@v5`.
- Prefer `go-version-file: go.mod`.
- If a literal version is necessary, use the version declared by `go.mod` and document why `go-version-file` was not used.

## Formatting failures are unclear

A formatting check that only exits non-zero without printing file paths would make CI failures harder to fix.

Mitigation:

- Use `find . -name "*.go" -print0 | xargs -0 gofmt -l` or an equivalent small shell script.
- Store or print the formatting check output before failing.
- Include a short message explaining that listed files are not gofmt-formatted.

## Formatting step mutates files in CI

Using `go fmt` or `gofmt -w` in CI could hide the real problem by modifying the checkout instead of failing the workflow.

Mitigation:

- Use `find . -name "*.go" -print0 | xargs -0 gofmt -l` or an equivalent `gofmt -l` based command for detection only.
- Do not use `gofmt -w` or `go fmt` as the CI formatting step.

## Cached test results hide failures

The default `go test ./...` command can use cached package test results, which weakens confidence in CI when investigating reliability problems.

Mitigation:

- Run `go test -count=1 ./...` in CI.
- Run the same uncached test command locally during implementation verification.

## Overcomplicated CI

Adding matrices, third-party linters, coverage services, Docker services, or release automation would increase maintenance cost and may introduce new failure modes unrelated to repository verification.

Mitigation:

- Keep the workflow to checkout, Go setup, formatting check, and tests.
- Add new CI capabilities only through separate OpenSpec changes with a concrete need.

## User-project CI behavior leaks into this change

SpecHarbor may later scan or reason about user project workflows, but this change is only about SpecHarbor's own repository CI.

Mitigation:

- Do not add scan behavior.
- Do not add workflow connector behavior.
- Do not define templates or generated CI for user projects.
- Keep the implementation limited to `.github/workflows/ci.yml`.

## Accidental product code changes

CI reliability work should not alter CLI behavior or application architecture.

Mitigation:

- Do not modify `cmd/`, `internal/core/`, `internal/adapters/`, or `internal/platform/`.
- If an unexpected CI blocker appears in product code, stop and document the blocker before expanding scope.

## Local verification is skipped

Because this is a workflow change, it is possible to update YAML without confirming that the referenced local commands pass.

Mitigation:

- Run `find . -name "*.go" -print0 | xargs -0 gofmt -l`.
- Run `go test -count=1 ./...`.
- Run `go test ./...` for the repository mandatory workflow.
- Leave any verification task unchecked if the command was not run.
