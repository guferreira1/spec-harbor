# Design: Release Publishing Automation

## Overview

The release pipeline becomes a tag-only, multi-job GitHub Actions workflow that
validates version consistency, builds and publishes GitHub Release assets and
`checksums.txt` through GoReleaser, updates the external Homebrew tap through
GoReleaser's `brews` integration, and publishes the npm wrapper package after
the release assets exist. All publishing is gated behind a strict version check
and the existing test suites, uses least-privilege permissions, and never
prints secrets.

## Approach

### Release trigger and SemVer gate

The workflow trigger stays `on.push.tags: ["v*"]` because GitHub Actions tag
filters are globs, not regular expressions. Exactness is enforced by a separate
`validate-release-inputs` job that runs a deterministic POSIX shell script,
`scripts/validate-release-version.sh`, which:

- reads the tag from its first argument or `GITHUB_REF_NAME`;
- rejects anything that is not exactly `vX.Y.Z` (stable SemVer only — `0.2.0`,
  `v0.2`, and `v0.2.0-beta.1` all fail), because the repository currently ships
  stable releases only and the npm wrapper maps one package version to exactly
  one tag;
- strips the leading `v` to obtain `X.Y.Z`;
- reads `packages/npm/specharbor/package.json` `version`;
- fails unless the tag version equals the package version.

The npm wrapper already maps package version `X.Y.Z` to release tag `vX.Y.Z`
through `releaseTag()` in `packages/npm/specharbor/lib/platform.js`; asserting
`package.json` equals the tag's `X.Y.Z` therefore guarantees the wrapper will
download assets from the correct release. The script is also runnable locally
and is covered by a companion shell test so its logic is verifiable without a
real tag.

GoReleaser derives its version from the git tag (`{{ .Version }}` resolves to
`X.Y.Z`), so once the tag-vs-package check passes, the GoReleaser version is
consistent by construction; the existing version-injection ldflags are
unchanged.

### Job ordering

```
validate-release-inputs   (contents: read)
        |
        v
goreleaser                (contents: write)  -> GitHub Release assets + checksums.txt
        |
        +-------------------------+
        v                         v
npm-publish                 homebrew-publish
(contents: read,            (contents: read,
 id-token: write)            tap token)
        |                         |
        +-----------+-------------+
                    v
            release-summary       (contents: read)
```

`npm-publish` and `homebrew-publish` both declare `needs: goreleaser` because
they consume the GitHub Release assets and `checksums.txt`: the npm wrapper's
postinstall downloads the asset for the same version, and the Homebrew formula
pins each download to its SHA-256 from the release `checksums.txt`. Publishing
either before the release exists would produce a broken install, so both wait
for the GoReleaser job to finish.

### GitHub Release assets (unchanged)

The GoReleaser `builds`, `archives`, and `checksum` sections are kept
byte-compatible with the current configuration. The six approved archives and
`checksums.txt` keep their exact names:

```
specharbor_Linux_x86_64.tar.gz
specharbor_Linux_arm64.tar.gz
specharbor_Darwin_x86_64.tar.gz
specharbor_Darwin_arm64.tar.gz
specharbor_Windows_x86_64.zip
specharbor_Windows_arm64.zip
checksums.txt
```

Keeping these names preserves compatibility with `install.sh`
(`asset_name()`), the npm wrapper (`assetName()`), and the Homebrew formula.

### Homebrew tap via a dedicated formula job

GoReleaser's `brews` (formula) integration is **deprecated** in the pinned
GoReleaser v2 used by this repository, and `goreleaser check` (a required
validation gate) fails when it is present. Its non-deprecated replacement,
`homebrew_casks`, would convert the tap from a Homebrew **formula** to a
**cask**, changing the tap layout and the documented `test do` behavior and
breaking backward compatibility with the existing manually maintained formula.
The spec's preferred path is "GoReleaser Homebrew support *if compatible*"; it
is not compatible, so this change uses the explicitly permitted alternative: a
dedicated `homebrew-publish` job that updates the formula in the external tap.

`.goreleaser.yaml` therefore stays byte-compatible with `main` (no `brews`
block), keeping `goreleaser check` green and the GoReleaser guardrail tests
unchanged.

The formula is produced by a small, offline-testable POSIX script,
`scripts/render-homebrew-formula.sh`, which reads the release `checksums.txt`,
extracts the SHA-256 for the macOS arm64 and x86_64 archives, and emits a
standard macOS formula (`on_macos`/`on_arm`/`on_intel`) that installs the
prebuilt binary (`bin.install "specharbor"`) and keeps the `test do` block
running `specharbor version`. macOS-only coverage matches the documented
current formula, preserving backward compatibility for
`brew install guferreira1/tap/specharbor`.

The `homebrew-publish` job (`needs: goreleaser`) downloads `checksums.txt` from
the just-created GitHub Release with the repository `GITHUB_TOKEN`, renders the
formula, checks out `guferreira1/homebrew-tap` using a dedicated
`HOMEBREW_TAP_GITHUB_TOKEN` (the default `GITHUB_TOKEN` cannot write to a
separate repository), writes `Formula/specharbor.rb`, and commits and pushes
only when the formula changed. `actions/checkout` manages the tap credential so
the token is never placed on a command line or printed, and GitHub Actions
masks the secret in logs.

### npm publishing

npm publishing runs in its own job with `node` set up against
`https://registry.npmjs.org`. The primary, recommended path is npm **trusted
publishing via GitHub OIDC** with provenance:

- the job requests `id-token: write`;
- the publish command is `npm publish --provenance --access public`;
- no long-lived token is required when a trusted publisher is configured for
  the `specharbor` package on npmjs.com.

A documented fallback uses a granular automation token stored as the
`NPM_TOKEN` repository secret, consumed as `NODE_AUTH_TOKEN`. Trusted
publishing and provenance both require a recent npm (>= 11.5), so the job
ensures an up-to-date npm before publishing. The choice between OIDC and token
is a one-time maintainer configuration documented in `docs/release.md`; the
workflow defaults to OIDC and never prints either credential.

Before publishing, the job runs the package test suite and content validation
(see below). It publishes only on tag events, only after `validate-release-inputs`
and `goreleaser` succeed, and never on pull requests or branches.

### npm pre-publish validation and package contents

The npm wrapper package has zero runtime dependencies and intentionally ships
no lockfile (enforced by `packages/npm/specharbor/test/package.test.js` and
`scripts/test-install-channels-safety.sh`). The correct install strategy is
therefore **not** `npm ci`/`npm install`; the job runs the tests directly with
`npm test` (which is `node --test`) and validates packaging with
`npm pack --dry-run`.

Packaged-contents validation asserts the tarball includes `bin/`, `lib/`,
`scripts/`, `README.md`, `README.pt-BR.md`, and `package.json`, and excludes
`native/` (downloaded binaries), `node_modules/`, and test fixtures. Because
the npm `README.md` links to `README.pt-BR.md`, a missing `README.pt-BR.md`
must fail the release; this is enforced both by the `files` array in
`package.json` and by an added assertion in `test/package.test.js` plus the
`npm pack --dry-run` check in CI.

### PR validation and dry-run

The release workflow is tag-only and therefore cannot publish from a pull
request. Reviewers validate release changes locally and in normal CI without a
tag using:

- `goreleaser check` (config validity);
- `goreleaser release --snapshot --clean` (full build without publishing);
- `npm test` and `npm pack --dry-run` in `packages/npm/specharbor`;
- `scripts/validate-release-version.sh` against a sample tag.

Snapshot output is written under `dist/`, which is git-ignored, so dry-runs
never produce committed artifacts.

## Architecture

This change touches release/packaging delivery mechanics only; it does not add
or modify domain, ports, use cases, or adapters, and it preserves the inward
dependency rule. The one Go file changed is the architecture guardrail test
`internal/architecture/release_versioning_boundaries_test.go`, which is updated
to match the new approved scope:

- `.goreleaser.yaml` is unchanged, so the GoReleaser guardrail tests
  (`TestGoReleaserConfigBuildsApprovedArchivesAndChecksums`,
  `TestGoReleaserConfigRejectsOutOfScopePublishingSections`) keep passing
  unchanged, including the existing prohibition on `brews`, `homebrew_casks`,
  `nfpms`, `dockers`, `sboms`, `signs`, `cosign`, `scoops`, `winget`, `aurs`,
  `nix`, and `chocolateys`;
- the release workflow guardrail tests are updated so `release.yml` may contain
  the ordered publishing jobs, an `id-token: write` permission scoped to the
  npm job, and the publishing secrets (`HOMEBREW_TAP_GITHUB_TOKEN`, optional
  `NPM_TOKEN`), while remaining tag-only with no `pull_request`, `branches`,
  `schedule`, `workflow_dispatch`, or `write-all`, and no Docker/cosign/SBOM/
  signing or Linux/Windows-package behavior;
- a new test asserts job ordering: `goreleaser` depends on
  `validate-release-inputs`, and `npm-publish` and `homebrew-publish` depend on
  `goreleaser`.

Updating these workflow tests is required because the previous assertions
explicitly block the publishing automation this change implements. The
documentation guardrail test
(`TestReleaseVersioningDocumentationDescribesImplementedScopeOnly`) keeps its
required snippets satisfied: the docs still mention npm, Homebrew,
`checksums.txt`, the snapshot commands, SBOM, Docker, and "future work" for the
items that remain genuinely out of scope.

## Decisions and Tradeoffs

- **Homebrew through a dedicated formula job instead of GoReleaser `brews`.**
  `brews` is deprecated (fails `goreleaser check`) and `homebrew_casks` would
  break the existing formula-based tap. A separate job renders and commits the
  formula from the release `checksums.txt`, keeping `.goreleaser.yaml` and its
  guardrail tests unchanged and preserving formula compatibility. Tradeoff: the
  formula is templated in-repo rather than by GoReleaser; mitigated by an
  offline-testable renderer script and a dedicated least-privilege tap token.
- **OIDC trusted publishing as the primary npm path.** Avoids long-lived
  tokens and adds provenance. Tradeoff: requires one-time npm-side
  configuration; mitigated by documenting the setup and supporting an
  `NPM_TOKEN` fallback.
- **Stable SemVer only (`vX.Y.Z`).** Matches current repository behavior and
  the one-version-to-one-tag npm wrapper mapping; prereleases are rejected to
  avoid accidental partial publishes. Revisiting prereleases would be a
  separate change.
- **Keep asset names unchanged.** Avoids breaking `install.sh`, the npm
  wrapper, and the formula. No consumer updates are required.
- **No lockfile, no `npm ci`.** Honors the existing zero-dependency,
  no-lockfile convention; the workflow runs tests and packaging checks directly.
