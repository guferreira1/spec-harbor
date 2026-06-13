# Proposal: Implement Release Publishing Automation

## Summary

Convert the current tag-triggered release workflow from a GitHub-Release-only
pipeline into safe, validated publishing automation for every currently
supported distribution channel: GitHub Releases, `checksums.txt`, `install.sh`
(as a consumer), the npm package `specharbor`, and the Homebrew tap
`guferreira1/homebrew-tap`. A single pushed tag `vX.Y.Z` must publish all of
these consistently, with a version-consistency gate that prevents accidental
or mismatched publishes.

## Problem

Today the release workflow (`.github/workflows/release.yml`) runs only on
pushed `v*` tags, runs Go tests, and invokes GoReleaser. GoReleaser produces
the GitHub Release assets and `checksums.txt`, but the workflow does **not**:

- publish the npm package `specharbor@X.Y.Z`, and
- update the external Homebrew tap `guferreira1/homebrew-tap`.

The release documentation (`docs/release.md`, `docs/install.md`) explicitly
states that npm publishing and Homebrew tap updates are manual maintainer
steps and that "package publishing automation remains future work". As a
result, every release requires error-prone manual npm publishing and manual
formula edits, and there is no automated guard that the git tag, the GoReleaser
version, and the npm `package.json` version agree before anything is published.

A second, structural problem: the architecture guardrail test
`internal/architecture/release_versioning_boundaries_test.go` was written for
the previous scope and actively forbids publishing automation. It asserts that
`.goreleaser.yaml` contains no `brews` section and that `release.yml` contains
no npm/Homebrew/`id-token`/extra-secret behavior. Those guardrails encode the
old "publishing automation is future work" decision and must be updated to the
new scope so the feature can land while still forbidding genuinely
out-of-scope channels.

## Goal

A maintainer pushes a stable SemVer tag `vX.Y.Z` and the release workflow:

1. Validates that the tag, the GoReleaser-resolved version, and the npm
   `package.json` version all agree on `X.Y.Z`, failing before any publish on
   mismatch.
2. Builds and publishes the GitHub Release assets and `checksums.txt` through
   GoReleaser exactly as today (same asset names).
3. Publishes the npm package `specharbor@X.Y.Z` only after the GitHub Release
   assets exist, because the npm wrapper downloads those assets at install
   time.
4. Updates the Homebrew tap `guferreira1/homebrew-tap` to the new version,
   referencing the correct release URLs and SHA-256 values.
5. Keeps `install.sh` working unchanged against the same official release
   assets and checksums.

Publishing must be tag-only, must never run on pull requests or branches, must
use least-privilege permissions, and must never print secrets.

## Scope

- Add a `validate-release-inputs` gate that enforces stable SemVer tag format
  `vX.Y.Z` and tag/`package.json` version consistency, implemented as a small,
  testable POSIX shell script under `scripts/`.
- Extend `.github/workflows/release.yml` into ordered jobs:
  `validate-release-inputs` -> `goreleaser` (GitHub Release + checksums +
  Homebrew tap via GoReleaser `brews`) -> `npm-publish` -> `release-summary`.
- Add a GoReleaser `brews` block that pushes the formula to
  `guferreira1/homebrew-tap` using a dedicated tap token
  (`HOMEBREW_TAP_GITHUB_TOKEN`), preserving the existing builds, archives,
  asset names, and `checksums.txt`.
- Add an npm publish job using npm trusted publishing (OIDC) with provenance
  as the primary path and a documented `NPM_TOKEN` fallback. The job runs the
  npm package tests and `npm pack --dry-run` content validation before
  publishing, and publishes only from tags.
- Add packaged-contents validation so a missing `README.pt-BR.md` (linked from
  the npm `README.md`) or accidental inclusion of `native/`, `node_modules/`,
  or test fixtures fails the release.
- Update the architecture guardrail test
  `internal/architecture/release_versioning_boundaries_test.go` to the new
  scope: allow GoReleaser `brews` and the npm/Homebrew publishing jobs while
  still forbidding out-of-scope channels and behaviors.
- Update `docs/release.md` (and `docs/install.md` where channel status text
  changes) to document the automated state, required secrets / trusted
  publisher setup, a maintainer release checklist, failure handling, and
  dry-run/snapshot validation.

## Out of Scope

The following remain explicitly future work and must not be implemented:

- Linux native packages (`.deb`, `.rpm`, `.apk`).
- Windows package managers (Winget, Scoop, Chocolatey).
- Docker images or manifests.
- Signing, cosign, attestations, or SBOM generation.
- Provider APIs, AI/RAG/runtime features, and context-initiative work.
- Release-notes automation beyond GoReleaser's existing default behavior.
- Any change to runtime CLI behavior unrelated to release publishing.

## Success Criteria

- The release workflow defines an ordered, tag-only pipeline whose publishing
  jobs run only after version validation and tests pass.
- Version mismatch between the tag and `package.json` fails the release before
  any npm publish or Homebrew tap update.
- GoReleaser still emits the six approved archives and `checksums.txt` with
  unchanged names, keeping `install.sh`, the npm wrapper, and the Homebrew
  formula compatible.
- npm publishing and Homebrew tap updating are driven from CI with documented
  secrets and least-privilege permissions, and neither runs on pull requests
  or branches.
- `go test ./...`, `go run ./cmd/specharbor validate
  implement-release-publishing-automation`, `goreleaser check`,
  `goreleaser release --snapshot --clean`, npm tests, and `npm pack --dry-run`
  all pass locally without performing any real publish, tag push, or tap
  write.
- Documentation accurately reflects which channels are automated and which
  supply-chain features remain future work.
