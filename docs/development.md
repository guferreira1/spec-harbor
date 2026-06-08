# Development

SpecHarbor is a Go CLI. The repository Go version is declared in `go.mod`:

```text
go 1.23
```

Use Go 1.23 or a compatible newer Go toolchain.

## Local Commands

Run tests:

```bash
go test ./...
```

Run uncached tests when you want the same test mode used by CI:

```bash
go test -count=1 ./...
```

Build the CLI:

```bash
go build ./cmd/specharbor
```

Run the CLI during development:

```bash
go run ./cmd/specharbor help
```

Check formatting:

```bash
find . -name '*.go' -print0 | xargs -0 gofmt -l
```

Format changed Go files with `gofmt` before finalizing code changes.

## CI

SpecHarbor's own CI is Go-specific because SpecHarbor is written in Go.

GitHub Actions runs on pull requests and pushes to `main`. The workflow:

- uses the repository Go module through `go.mod`;
- checks `gofmt`;
- runs `go test -count=1 ./...`.

This does not make user project scanning Go-specific. `specharbor scan` is intended to be stack-agnostic and should not assume user projects are Go projects.

## Documentation-Only Changes

For documentation-only OpenSpec changes, keep the diff to Markdown files and the active change's `tasks.md`. Do not modify Go code, Go tests, `.github/workflows/ci.yml`, CLI behavior, or init templates unless the active change explicitly requires it.
