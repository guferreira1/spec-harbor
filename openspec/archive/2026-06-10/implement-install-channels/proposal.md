# Proposal: Implement Install Channels

## Problem

SpecHarbor is preparing for public distribution, but users currently have no supported install channel other than building from source with `go build` or `go install`.

The merged `implement-release-versioning` change provides the binary metadata foundation: release builds inject `Version`, `Commit`, `Date`, and `Dirty` through Go `-ldflags -X` variables under `github.com/guferreira1/spec-harbor/internal/platform/version`, and `specharbor version` reports them deterministically.

A parallel change, `implement-goreleaser-release`, is expected to define GitHub Release assets and checksums. Until that change finalizes release asset names and checksum format, install channels cannot be implemented safely, because every channel must download and verify exactly those assets.

Without defined install channels, users would invent ad-hoc install paths: unverified downloads, copy-pasted binaries without checksums, or unofficial packages under the project name. That is a distribution and supply-chain risk.

## Goal

Define safe, user-friendly install channels for SpecHarbor that consume official GitHub Release assets after GitHub Releases are available.

Target install channels:

- a shell install script (`install.sh`);
- an npm global package named `specharbor`;
- a Homebrew tap under `guferreira1/homebrew-tap` with a `specharbor` formula;
- future native Linux packages (`.deb` and `.rpm`);
- future Windows package managers (Scoop and Winget).

Desired user experience after implementation:

```bash
curl -sSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
```

```bash
npm install -g specharbor
```

```bash
brew install guferreira1/tap/specharbor
```

Manual GitHub Release install: download the archive, verify the checksum, place the binary on `PATH`.

Every channel must verify checksums for downloaded artifacts, use HTTPS only, and install a binary whose `specharbor version` output reflects the injected release metadata.

## Dependency

This change depends on `implement-goreleaser-release`.

- Implementation of this change must not start until the GoReleaser release asset names and checksum file format are finalized by `implement-goreleaser-release`.
- All install channels must consume official GitHub Release assets published for `guferreira1/spec-harbor`.
- No channel may build from source unless explicitly documented as a fallback. The only documented fallback is manual `go install github.com/guferreira1/spec-harbor/cmd/specharbor@<tag>`, which yields development fallback metadata (`dev`/`unknown`) and is documented as such.
- Asset naming patterns referenced in this change are working assumptions and must be reconciled against the finalized `implement-goreleaser-release` output before implementation.

## Scope

This is a planning and specification change for staged implementation. The first implementation stage, once unblocked, covers:

- An `install.sh` script at the repository root that detects OS and architecture, resolves the latest release or an explicit version, downloads the matching GitHub Release asset over HTTPS, verifies its checksum, and installs the binary to a user-local directory without requiring sudo by default.
- An npm wrapper package named `specharbor` that exposes the `specharbor` CLI command through `bin`, detects OS and architecture, downloads the matching GitHub Release asset at postinstall with required checksum verification, and falls back to a first-run download when postinstall scripts were skipped.
- A Homebrew formula `specharbor` in a personal tap repository `guferreira1/homebrew-tap` that installs from GitHub Release assets with `sha256` verification. The formula itself lives in a separate repository and may require a separate repository/PR; this repository documents the tap and install command.
- Documentation in `README.md` and a new `docs/install.md` covering all installation options, checksum verification, `specharbor version` verification, `PATH` troubleshooting, and what remains future-only.
- Tests for the install script, the npm wrapper, and the Homebrew formula expectations, plus regression and safety checks.

Linux packages and Windows package managers are defined as future channels only and are documented but not implemented.

## Out of Scope

- Implementing any code, script, package, or formula before `implement-goreleaser-release` finalizes asset names and checksums.
- GoReleaser configuration (`.goreleaser.yaml` or `.goreleaser.yml`).
- GitHub Release workflows or release-specific GitHub Actions workflows.
- Release asset generation, archives, or checksum files.
- Publishing the npm package to the npm registry.
- Creating or publishing the Homebrew tap repository content as part of this repository's diff.
- Native Linux packages, `nfpm.yaml`, `.nfpm.yaml`, `packaging/`, `debian/`, or `rpm/` files.
- Windows package managers, `scoop/`, or `winget/` manifests.
- Binary signing, cosign, or signature verification.
- SBOM generation.
- Docker images.
- Auto-update mechanisms.
- Telemetry of any kind.
- Source-control automation: no tags, no releases, no commits, no pushes, no PRs created by any channel or test.
- Changes to CLI runtime behavior, `internal/`, or `cmd/`.

## Success Criteria

- The change defines all five target install channels with explicit first-stage versus future-only status.
- The change states that implementation must not start until `implement-goreleaser-release` finalizes asset names and checksums.
- The change states that all channels consume official GitHub Release assets and that building from source is only a documented manual fallback.
- `install.sh` behavior is fully specified: OS/arch detection, version resolution, HTTPS download, mandatory checksum verification, user-local install target, no sudo by default, no arbitrary shell execution, no package-manager side effects, and cleanup on failure.
- The npm strategy is decided: a manual wrapper package named `specharbor`, postinstall download with checksum verification and first-run fallback, no GoReleaser Pro dependency, no npm org, scoped package only as a fallback if the name becomes unavailable.
- The Homebrew strategy is decided: personal tap `guferreira1/homebrew-tap`, formula `specharbor`, GitHub Release asset with `sha256`, no org, validation possible later in a GitHub Actions macOS runner, implementation possibly in a separate repository/PR.
- Linux packages and Windows package managers are documented as future-only with no manifest files in this change.
- The security model is explicit: HTTPS only, checksum verification for all downloads, no arbitrary command execution, no tokens required, no automatic publishing, no writes outside install target and cache directories, no Git mutation, no tags/releases/PRs.
- The documentation plan covers all install options, manual install, checksum verification, `specharbor version` verification, `PATH` troubleshooting, and future-only channels.
- `tasks.md` requires tests for install.sh, npm, Homebrew, and regression/safety, and stages implementation behind the GoReleaser dependency gate.
- `go run ./cmd/specharbor validate implement-install-channels` reports the change as valid.
