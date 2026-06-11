# Proposal: Implement GoReleaser Release

## Problem

SpecHarbor now has deterministic build metadata reporting through `specharbor version`, but it still has no release automation for public binary distribution.

Maintainers need a safe foundation that turns a version tag such as `v0.1.0` into GitHub Release assets without adding package-manager publishing, install scripts, signing, Docker images, or runtime Git discovery.

The previous release-versioning change intentionally kept GoReleaser and GitHub release workflows out of scope. This change should add only the minimum release automation needed to build native CLI archives from tags and upload them to GitHub Releases.

## Goal

Define and implement a minimal GoReleaser-based release foundation for the `specharbor` CLI.

The release flow should be:

1. A maintainer creates and pushes a Git tag such as `v0.1.0`.
2. GitHub Actions runs a release workflow only for tags matching `v*`.
3. GoReleaser builds native `specharbor` binaries for Linux, macOS, and Windows on `amd64` and `arm64`.
4. GoReleaser injects build metadata into `internal/platform/version`.
5. GoReleaser uploads archives and a checksum file to the GitHub Release for that tag.

The implementation must remain deterministic, testable, and narrow. It must not change runtime version behavior, core business logic, or any package-manager distribution channel.

## Scope

- Add a minimal `.goreleaser.yaml` using only free GoReleaser features.
- Add `.github/workflows/release.yml`.
- Configure the release workflow to trigger only on pushed tags matching `v*`.
- Configure the release workflow with explicit top-level `permissions: contents: write` and no unrelated permissions.
- Use the GitHub-provided `GITHUB_TOKEN` only.
- Build `specharbor` from `./cmd/specharbor`.
- Build targets:
  - `linux/amd64`
  - `linux/arm64`
  - `darwin/amd64`
  - `darwin/arm64`
  - `windows/amd64`
  - `windows/arm64`
- Inject linker metadata into:
  - `github.com/guferreira1/spec-harbor/internal/platform/version.Version`
  - `github.com/guferreira1/spec-harbor/internal/platform/version.Commit`
  - `github.com/guferreira1/spec-harbor/internal/platform/version.Date`
  - `github.com/guferreira1/spec-harbor/internal/platform/version.Dirty`
- Use plain SemVer such as `0.1.0` for injected binary `Version` when the Git tag is `v0.1.0`.
- Use full commit SHA for injected `Commit`.
- Use GoReleaser's UTC RFC3339 build date for injected `Date`.
- Use `true` or `false` for injected `Dirty`; clean CI tag builds should inject `false`.
- Produce release archives:
  - `specharbor_Linux_x86_64.tar.gz`
  - `specharbor_Linux_arm64.tar.gz`
  - `specharbor_Darwin_x86_64.tar.gz`
  - `specharbor_Darwin_arm64.tar.gz`
  - `specharbor_Windows_x86_64.zip`
  - `specharbor_Windows_arm64.zip`
- Produce `checksums.txt` using SHA-256 checksums.
- Keep generated artifacts under `dist/`.
- Ensure `dist/` is gitignored; preserve the existing `/dist/` ignore entry if present.
- Update tests and static checks for the new release boundary.
- Update `README.md` and `docs/release.md`; update `docs/usage.md` only if needed to keep release/version documentation consistent.

## Out of Scope

- npm publishing.
- npm package files.
- Homebrew formula or tap automation.
- `install.sh`.
- Linux packages through nfpm or any other package manager.
- Windows package-manager manifests.
- Docker images.
- SBOM generation.
- signing or cosign.
- GoReleaser Pro-only config fields or behavior.
- `packages`, `id-token`, `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or other unrelated workflow permissions.
- provenance or attestations.
- release notes automation beyond GoReleaser defaults.
- nightly releases.
- snapshots published from CI.
- creating tags automatically.
- pushing commits.
- opening pull requests.
- merging pull requests.
- source-control automation beyond GitHub Release creation from pushed tags.
- runtime release logic inside core packages.
- runtime Git lookup, `.git` reads, shell execution, or network calls for version metadata.

## Success Criteria

- A tag such as `v0.1.0` can drive GoReleaser through GitHub Actions.
- The release workflow runs only for pushed tags matching `v*`.
- The release workflow does not run for normal branch pushes or pull requests.
- The release workflow defines explicit top-level `permissions: contents: write` and grants no `packages`, `id-token`, `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or unrelated write permissions.
- GoReleaser builds the six approved OS/architecture assets.
- GoReleaser injects `Version=0.1.0`, `Commit=<full SHA>`, `Date=<UTC RFC3339 value>`, and `Dirty=false` for a clean `v0.1.0` CI release build.
- Release archives use `.tar.gz` for Linux and macOS and `.zip` for Windows.
- `checksums.txt` is generated and uploaded with the release assets.
- No package-manager artifacts, package-manager tokens, install scripts, Docker images, SBOMs, signing config, or source-control push/tag automation are added.
- `.goreleaser.yaml` uses GoReleaser Community/free features only and does not include Pro-only config sections or behavior.
- Runtime version behavior remains unchanged: injected values are displayed as-is and runtime code does not normalize versions or inspect Git.
- Local verification with `goreleaser check` and `goreleaser release --snapshot --clean` works.
- Snapshot artifacts in `dist/` are not committed.
