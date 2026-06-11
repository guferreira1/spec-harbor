# Design: Improve SpecHarbor CI

## Overview

This change updates only the SpecHarbor repository's own GitHub Actions workflow.

It does not add product behavior, scan behavior, user-project CI detection, workflow connector behavior, or OpenSpec generation behavior. The workflow is project infrastructure that verifies the repository in the same spirit as local development checks.

## Target Workflow

The target workflow remains a single GitHub Actions workflow at:

```text
.github/workflows/ci.yml
```

The workflow should keep the existing trigger intent:

```yaml
on:
  push:
    branches: [main]
  pull_request:
```

The workflow should keep one simple job on `ubuntu-latest` unless a specific CI issue proves that another runner or matrix is required.

Expected steps:

- checkout repository;
- set up Go;
- verify formatting;
- run tests.

## Go Version Management

The project declares its Go version in `go.mod`.

The workflow should use `actions/setup-go@v5` with the Go version resolved from `go.mod` when possible:

```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version-file: go.mod
```

This avoids duplicating the Go version in the workflow and reduces drift when the project updates `go.mod`.

If `go-version-file: go.mod` cannot be used for a concrete compatibility reason, the implementation may use the exact Go version declared in `go.mod`, but the reason must be documented in the implementation notes or task update.

## Formatting Check

CI should verify formatting for every `.go` file in the SpecHarbor repository without modifying files.

The formatting step should use `find` to select repository Go files and `gofmt -l` to report unformatted files. It must print the list before exiting non-zero.

Expected shell shape:

```sh
unformatted="$(find . -name '*.go' -print0 | xargs -0 gofmt -l)"
if [ -n "$unformatted" ]; then
  echo "Go files are not gofmt-formatted:"
  printf '%s\n' "$unformatted"
  exit 1
fi
```

This keeps the formatting failure readable in GitHub Actions logs and gives contributors the exact file paths to fix.

The implementation may use an equivalent small shell script if it still:

- checks all `.go` files in the SpecHarbor repository;
- does not rewrite files;
- prints unformatted file paths when formatting fails;
- exits non-zero when any unformatted file is found;
- exits zero when no unformatted files are found.

## Test Check

CI should run all Go tests without relying on cached test results.

The test step must use:

```text
go test -count=1 ./...
```

The `-count=1` flag disables Go's package test result cache for this run while keeping the command familiar to Go contributors.

## Simplicity Constraints

The workflow should avoid additional moving parts unless a concrete reliability issue requires them.

Do not add:

- third-party lint tools;
- coverage upload;
- release or deployment steps;
- Docker services;
- external service dependencies;
- custom CI scripts;
- matrix builds;
- user-project scanning or workflow detection.

## Architecture Impact

This is a project infrastructure change only.

No application architecture changes are expected. The implementation should not modify:

- `cmd/`;
- `internal/core/`;
- `internal/adapters/`;
- `internal/platform/`.

The living architecture spec does not need to change because the dependency rule, hexagonal boundaries, AI-provider rules, agent-target rules, and workflow-connector rules are not affected.

## Local Verification

After updating the workflow, the implementer should run local checks that mirror the workflow:

```text
find . -name "*.go" -print0 | xargs -0 gofmt -l
go test -count=1 ./...
```

The implementer should also run the repository-required test command:

```text
go test ./...
```

If any command cannot be run, the reason must be recorded in the implementation response and the corresponding task must remain unchecked.
