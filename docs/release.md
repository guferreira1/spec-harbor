# Release Metadata

SpecHarbor currently provides a local release-versioning foundation for the CLI binary. It does not publish releases or packages.

## Check Version

Use:

```bash
specharbor version
```

Default development output:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Fields:

- `version`: product version metadata displayed on the first line.
- `commit`: source commit supplied by the build.
- `date`: build timestamp supplied by the build.
- `dirty`: working tree state supplied by the build.

`dev` means no release version was injected. `unknown` means the build did not provide that metadata field.

Plain `go install` without `-ldflags` uses the same development fallback metadata. An installed binary built that way is expected to print:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

This is expected behavior. To get release metadata, the binary must be built with injected `-ldflags` values.

## Version Convention

Git release tags use `vX.Y.Z`, for example `v0.1.0`.

Binary version metadata uses plain `X.Y.Z`, for example `0.1.0`.

Future release tooling should convert a tag such as `v0.1.0` into injected binary metadata such as `0.1.0`. Runtime displays the injected version string as-is and does not normalize it. If a manual build injects `v0.1.0`, `specharbor version` may display `v0.1.0`.

## Build-Time Injection

Release builds inject metadata through Go `-ldflags -X` variables in:

```text
github.com/guferreira1/spec-harbor/internal/platform/version
```

Supported metadata variables:

- `Version`
- `Commit`
- `Date`
- `Dirty`

Example:

```bash
go build \
  -ldflags "
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Version=0.1.0
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Commit=abc1234
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Date=2026-06-10T19:00:00Z
    -X github.com/guferreira1/spec-harbor/internal/platform/version.Dirty=false
  " \
  ./cmd/specharbor
```

Expected injected output:

```text
SpecHarbor 0.1.0
commit: abc1234
date: 2026-06-10T19:00:00Z
dirty: false
```

Future release automation will inject release metadata with these linker variables. The runtime command does not inspect Git tags, read `.git`, run Git commands, execute shell commands, call the network, write files, or normalize versions.

## Not Implemented Here

This metadata foundation does not implement release publishing or package distribution.

Future OpenSpec changes may add:

- GitHub Releases.
- install scripts.
- npm publishing for the desired future package name `specharbor`, subject to verification during publishing.
- Homebrew publishing to `guferreira1/homebrew-tap`.
- Native Linux packages.
- Windows package-manager support.

No GoReleaser configuration, GitHub Release workflow, install script, npm package, Homebrew tap, Linux package, Windows package, publishing script, release archive, or checksum artifact is implemented by this change.
