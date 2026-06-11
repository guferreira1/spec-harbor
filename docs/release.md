# Release Metadata

SpecHarbor uses GoReleaser to build GitHub Release assets from pushed version tags.

## Current Public Release

The current public release is GitHub Release `v0.1.0`, built from commit
`e6faff91feef07e5c1e47181243286268daf17b5`. Release binaries display plain
version `0.1.0`.

The validated public distribution channels for this release are:

- GitHub Releases.
- `install.sh`.
- npm package `specharbor@0.1.0`.
- Homebrew tap install command `brew install guferreira1/tap/specharbor`.

The Homebrew formula lives in the external tap repository
`guferreira1/homebrew-tap`; the validated tap commit is
`a61783bcfa44f7eafdce72c70043b76e6f80df9c`.

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

This is expected behavior.

## Version Convention

Git release tags use `vX.Y.Z`, for example `v0.1.0`.

Release binary version metadata uses plain `X.Y.Z`, for example `0.1.0`. GoReleaser injects the plain version value, so a release built from tag `v0.1.0` displays:

```text
SpecHarbor 0.1.0
commit: <full commit sha>
date: <UTC RFC3339 build date>
dirty: false
```

Runtime displays the injected version string as-is and does not normalize it. If a manual build injects `v0.1.0`, `specharbor version` may display `v0.1.0`.

## Build-Time Injection

Release builds inject metadata through Go `-ldflags -X` variables in:

```text
github.com/guferreira1/spec-harbor/internal/platform/version
```

GoReleaser injects exactly:

- `Version={{ .Version }}`
- `Commit={{ .FullCommit }}`
- `Date={{ .Date }}`
- `Dirty={{ .IsGitDirty }}`

The runtime command does not inspect Git tags, read `.git`, run Git commands, execute shell commands, call the network, write files, or normalize versions.

## Release Workflow

Maintainers publish a release by pushing a tag that matches `v*`, such as:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions release workflow runs only for pushed tags matching `v*`. It does not run on normal branch pushes or pull requests. The workflow uses GoReleaser with the repository `GITHUB_TOKEN` and top-level `contents: write` permission to create or update the GitHub Release and upload assets.

The tag-based workflow creates GitHub Release assets and `checksums.txt`. It
does not publish npm packages, update the Homebrew tap, publish Linux native
packages, publish Windows package-manager manifests, sign binaries, generate
SBOMs, or publish Docker images.

## Release Assets

GoReleaser builds one `specharbor` binary from `./cmd/specharbor` for these archives:

- `specharbor_Linux_x86_64.tar.gz`
- `specharbor_Linux_arm64.tar.gz`
- `specharbor_Darwin_x86_64.tar.gz`
- `specharbor_Darwin_arm64.tar.gz`
- `specharbor_Windows_x86_64.zip`
- `specharbor_Windows_arm64.zip`

Linux and macOS assets use `.tar.gz`. Windows assets use `.zip`. GoReleaser also generates `checksums.txt` with SHA-256 checksums.

Installation options that consume these assets — manual download, `install.sh`, the npm wrapper package, and the Homebrew tap — are documented in [Install](install.md).

For `v0.1.0`, those assets cover:

- Linux amd64.
- Linux arm64.
- macOS amd64.
- macOS arm64.
- Windows amd64.
- Windows arm64.
- `checksums.txt`.

## Package Channels

npm package `specharbor@0.1.0` is published and maps package version `0.1.0`
to GitHub Release tag `v0.1.0`. The `0.1.0` npm publish was a manual
maintainer action.

Homebrew is available through:

```bash
brew install guferreira1/tap/specharbor
```

The formula is maintained manually for now in the external
`guferreira1/homebrew-tap` repository. Package publishing automation remains
future work.

## Local Snapshot Verification

Local snapshot releases are for verification only. They write generated artifacts under `dist/`, which is ignored by Git.

Use `goreleaser check` and `goreleaser release --snapshot --clean` before publishing release changes.

Run:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

Then run one generated snapshot binary:

```bash
./dist/specharbor_linux_amd64_v1/specharbor version
```

Snapshot versions may include GoReleaser snapshot metadata instead of a normal release version. They should still show injected `commit`, `date`, and `dirty` values instead of the default `unknown` fallback values.

## Future Work

`install.sh`, the npm wrapper package, and the external Homebrew tap are
available and documented in [Install](install.md). The release foundation does
not implement these future-only items:

- Native Linux packages such as `.deb`, `.rpm`, or `.apk`.
- Windows package-manager manifests such as Winget, Scoop, or Chocolatey.
- Signing, cosign, attestations, or SBOM generation.
- Docker images or Docker manifests.
- Package publishing automation for npm, Homebrew, Linux packages, Windows
  package managers, signing, SBOMs, or Docker.

Those publishing steps and supply-chain features require separate OpenSpec
changes or manual maintainer action.
