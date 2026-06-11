# Design: Implement Release Versioning

## Overview

This change makes version reporting a small platform concern with a thin CLI presentation layer.

The platform package owns build metadata defaults, linker-injected variables, empty-value fallback handling, and human-readable formatting. The CLI adapter owns command routing, argument validation, and writing the report to stdout.

No use case, domain model, provider adapter, workflow connector, or release automation is required for this first scope.

## Current State

The current version package is:

```go
package version

const Version = "0.1.0-dev"
```

The current CLI command prints:

```go
fmt.Fprintln(ctx.Output, version.Version)
```

Because `Version` is a constant, it cannot be overridden with Go linker `-X` flags. Release builds need package-level string variables instead.

## Metadata Model

First-scope metadata fields:

| Field | Type | Meaning |
| --- | --- | --- |
| `version` | string | Product version, usually a semantic version such as `0.1.0`. |
| `commit` | string | Source commit used to build the binary, usually a short or full Git SHA supplied by the build tool. |
| `date` | string | Build timestamp supplied by the build tool. Release automation should use UTC RFC3339, for example `2026-06-10T19:00:00Z`. |
| `dirty` | string | Working tree state supplied by the build tool: `true`, `false`, or `unknown`. |

Optional fields such as `builtBy`, `goVersion`, and `platform` are intentionally out of first scope.

The implementation should expose a small value object in `internal/platform/version`, for example:

```go
type Metadata struct {
	Version string
	Commit  string
	Date    string
	Dirty   string
}
```

The exact struct shape may vary, but the field meanings and rendered output must remain the same.

## Default Development Values

Default values must be deterministic:

| Field | Default |
| --- | --- |
| `version` | `dev` |
| `commit` | `unknown` |
| `date` | `unknown` |
| `dirty` | `unknown` |

These defaults apply to local development, plain `go run`, and plain `go install` builds that do not pass linker flags.

If an internal metadata helper or constructor receives an empty value, the report should fall back to the same deterministic defaults instead of printing an empty field.

This empty-value fallback is missing-metadata handling, not version string normalization. Non-empty injected version strings must be displayed exactly as provided. Runtime code must not add, strip, or otherwise transform a leading `v` on non-empty version strings.

## Version String Convention

The project version string convention is:

- Git release tags use a leading `v`, for example `v0.1.0` and `v0.2.0`.
- Binary version metadata uses plain SemVer without the leading `v`, for example `0.1.0` and `0.2.0`.
- `specharbor version` displays the injected `Version` value exactly as provided.
- Runtime code must not normalize the version string.
- Runtime code must not strip a leading `v`.
- Runtime code must not add a leading `v`.
- Runtime code must not inspect Git tags.
- Runtime code must not run Git commands.
- Future release tooling will be responsible for converting a Git tag such as `v0.1.0` into the injected binary version `0.1.0`.
- This change does not implement that release tooling.

If a maintainer manually injects `v0.1.0`, the binary may display `v0.1.0` because the command displays the injected value as-is. The documented project convention is still to inject plain SemVer such as `0.1.0`.

## Build-Time Injection

Release metadata must be injected at build time with Go linker flags.

For release builds, the injected `Version` value must follow the binary metadata convention: plain SemVer without a leading `v`.

The version package must use package-level string variables, not constants, for linker-injected values:

```go
var Version = "dev"
var Commit = "unknown"
var Date = "unknown"
var Dirty = "unknown"
```

These variables are global only because Go linker injection requires addressable package variables. Runtime application code must not mutate them.

Manual build example:

```bash
go build \
  -ldflags "\
-X github.com/guferreira1/spec-harbor/internal/platform/version.Version=0.1.0 \
-X github.com/guferreira1/spec-harbor/internal/platform/version.Commit=abc1234 \
-X github.com/guferreira1/spec-harbor/internal/platform/version.Date=2026-06-10T19:00:00Z \
-X github.com/guferreira1/spec-harbor/internal/platform/version.Dirty=false" \
  ./cmd/specharbor
```

Implementation must not gather metadata by:

- reading `.git`;
- inspecting Git tags;
- shelling out to `git`;
- executing shell commands;
- making network calls;
- reading release service APIs at runtime.

Any Git lookup needed by future release automation belongs in the build pipeline before the binary is produced, not inside the `specharbor` binary.

## CLI Output

First scope uses deterministic human-readable multiline output.

No JSON or short mode is added in this change.

The CLI must render the version string exactly as supplied by `internal/platform/version`. It must not add or remove a leading `v`.

Default local output:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Release output:

```text
SpecHarbor 0.1.0
commit: abc1234
date: 2026-06-10T19:00:00Z
dirty: false
```

The report should end with a trailing newline when printed by the CLI.

`specharbor version` accepts no arguments in this first scope. Unsupported flags and extra positional arguments should be rejected consistently with existing CLI error style:

- `specharbor version --json` returns an unsupported flag error.
- `specharbor version extra` returns an unexpected argument error.

This intentionally changes the current behavior where extra arguments after `version` are ignored.

Top-level `specharbor --version` remains out of scope and should keep the existing unknown-command behavior unless a separate change defines it.

## Distribution Mode Behavior

| Mode | Expected metadata behavior |
| --- | --- |
| Local development | Uses defaults: `dev`, `unknown`, `unknown`, `unknown`. |
| `go run ./cmd/specharbor version` | Uses defaults unless explicit `-ldflags` are supplied. |
| Plain `go install` from source | Uses defaults unless the user or wrapper supplies linker flags. |
| Release build | Injects all four fields with `-ldflags -X`; `Version` uses plain SemVer such as `0.1.0`. |
| Future release automation build | Converts a Git tag such as `v0.1.0` to `0.1.0` before injecting `Version`. Automation is not part of this change. |
| Future GoReleaser build | Injects the same four variables from GoReleaser templates. GoReleaser configuration is not part of this change. |
| Future npm package `specharbor` | Ships or downloads a binary already built with these variables injected. Publishing is not part of this change. |
| Future Homebrew tap `guferreira1/homebrew-tap` | Either downloads a prebuilt release artifact with metadata already injected or builds from source with the same `-ldflags`. Formula automation is not part of this change. |
| Future native Linux packages | Reuse release artifacts or build with the same linker variables in later changes. |
| Future Windows package managers | Reuse release artifacts or build with the same linker variables in later changes. |

The npm package name `specharbor` is the desired future package name based on maintainer checks showing `npm view specharbor` returned 404 and `npm search specharbor` returned no matches. This change does not reserve or publish that package.

## Documentation Plan

Update:

- `README.md`
- `docs/usage.md`
- a new `docs/release.md`

Documentation must explain:

- how to check the installed version with `specharbor version`;
- the multiline output fields;
- what `dev` and `unknown` mean;
- how release builds inject metadata with `-ldflags -X`;
- that Git tags use `vX.Y.Z` while injected binary versions use plain `X.Y.Z`;
- that `specharbor version` displays the injected version string as-is;
- that the first scope does not publish packages;
- that GitHub Releases, install scripts, npm, Homebrew, native Linux packages, and Windows package managers are future work.

Do not document unimplemented install commands or package-manager availability as current behavior.

## Testing Strategy

Platform tests should cover:

- default metadata values;
- empty metadata values falling back to defaults;
- formatted human-readable output;
- overridden metadata values through a testable constructor, helper, or linker-injection integration test;
- version output with an injected plain SemVer-like value such as `0.1.0`;
- preservation of an injected version value as-is, including a value such as `v0.1.0`;
- absence of runtime version string normalization such as adding or removing a leading `v`;
- absence of runtime Git, shell, network, and filesystem dependencies in the version package.

CLI tests should cover:

- `specharbor version` output;
- `specharbor version` with no arguments;
- unsupported version flags;
- unexpected version arguments;
- read-only behavior;
- existing help and unknown-command behavior.

Regression tests should confirm existing commands remain unaffected:

- `init`;
- `generate`;
- `validate`;
- `prompt`;
- `review`;
- `archive`;
- `workflow`;
- `config`;
- unknown commands.

Static and architecture checks should verify this change introduces none of the following:

- no `.goreleaser.yaml` or `.goreleaser.yml`;
- no `.github/workflows/release.yml` or `.github/workflows/release.yaml`;
- no release-specific GitHub Actions workflows;
- no `install.sh` or `scripts/install.sh`;
- no npm package files such as `package.json`, `package-lock.json`, `npm/`, or `packages/npm/`;
- no Homebrew formula or tap files such as `Formula/` or `homebrew/`;
- no Linux package files such as `nfpm.yaml`, `.nfpm.yaml`, `packaging/`, `debian/`, or `rpm/`;
- no Windows package-manager files such as `winget/` or `scoop/`;
- no release publishing scripts;
- no package-manager artifacts;
- no generated release archives or checksums.

Repository verification must include:

```text
go test ./...
```

## Architecture

`internal/platform/version` is an appropriate location because release metadata is cross-cutting technical information, not a domain concept.

Architecture constraints:

- Core packages must not import CLI or adapters for version reporting.
- The CLI adapter may import `internal/platform/version`.
- The version package must not import concrete adapters.
- The version package must not import Git libraries.
- The version package must not import `os/exec`.
- The version package must not perform network calls.
- The version package must not inspect Git tags.
- The version command must not write files.
- This change must not add release automation, package-manager, publishing, or generated release artifact files.

This change does not require living architecture spec updates.
