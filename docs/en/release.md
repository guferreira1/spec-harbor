# Release Metadata

SpecHarbor uses GoReleaser to build GitHub Release assets from pushed version tags.

## Current Public Release

The current public release is `v0.2.0` after the `v0.2.0` tag is published.
The exact release commit is the commit referenced by tag `v0.2.0`. Release
binaries display plain version `0.2.0`.

The validated public distribution channels for this release are:

- GitHub Releases.
- `install.sh`.
- npm package `specharbor@0.2.0`.
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

Git release tags use `vX.Y.Z`, for example `v0.2.0`.

Release binary version metadata uses plain `X.Y.Z`, for example `0.2.0`. GoReleaser injects the plain version value, so a release built from tag `v0.2.0` displays:

```text
SpecHarbor 0.2.0
commit: <full commit sha>
date: <UTC RFC3339 build date>
dirty: false
```

Runtime displays the injected version string as-is and does not normalize it. If a manual build injects `v0.2.0`, `specharbor version` may display `v0.2.0`.

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
git tag v0.2.0
git push origin v0.2.0
```

The GitHub Actions release workflow runs only for pushed tags matching `v*`. It
does not run on normal branch pushes or pull requests, so a pull request can
never publish. The workflow is least privilege: the top-level permission is
`contents: read` and only the jobs that need more escalate.

The workflow runs ordered jobs:

1. `validate-release-inputs` — runs `scripts/validate-release-version.sh` to
   reject any tag that is not an exact stable SemVer tag `vX.Y.Z` and to fail
   if the tag's `X.Y.Z` does not equal the npm `package.json` version.
2. `goreleaser` (`contents: write`) — runs `go test ./...` and GoReleaser
   `release --clean` with the repository `GITHUB_TOKEN` to create the GitHub
   Release, upload assets, and publish `checksums.txt`.
3. `npm-publish` (`contents: read`, `id-token: write`, `needs: goreleaser`) —
   re-validates the version, runs the npm package tests and packaged-contents
   validation, then publishes `specharbor@X.Y.Z`.
4. `homebrew-publish` (`contents: read`, `needs: goreleaser`) — renders the
   formula from the release `checksums.txt` and updates the
   `guferreira1/homebrew-tap` repository.
5. `release-summary` — records a per-channel summary.

`npm-publish` and `homebrew-publish` wait for `goreleaser` because both consume
the published GitHub Release assets and `checksums.txt`. This tag-based workflow
now publishes GitHub Release assets, `checksums.txt`, the npm package, and the
Homebrew tap. It does not publish Linux native packages, publish Windows
package-manager manifests, sign binaries, generate SBOMs, or publish Docker
images.

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

For `v0.2.0`, those assets cover:

- Linux amd64.
- Linux arm64.
- macOS amd64.
- macOS arm64.
- Windows amd64.
- Windows arm64.
- `checksums.txt`.

## Package Channels

npm package `specharbor@X.Y.Z` is published automatically by the `npm-publish`
job and maps package version `X.Y.Z` to GitHub Release tag `vX.Y.Z`. Before
publishing, the job runs `scripts/validate-release-version.sh`, the npm package
tests (`npm test`), and a `npm pack --dry-run` contents check that requires
`bin/`, `lib/`, `scripts/`, `README.md`, `README.pt-BR.md`, and `package.json`
and rejects `native/`, `node_modules/`, and test fixtures. Publishing uses
`npm publish --provenance --access public`.

Homebrew is available through:

```bash
brew install guferreira1/tap/specharbor
```

The macOS formula in the external `guferreira1/homebrew-tap` repository is
updated automatically by the `homebrew-publish` job, which renders
`Formula/specharbor.rb` from the release `checksums.txt` with
`scripts/render-homebrew-formula.sh` and commits it to the tap. The formula
pins each download to its SHA-256 and keeps a `test do` block running
`specharbor version`.

### Required secrets and trusted-publisher setup

| Secret / setting | Purpose |
| --- | --- |
| `GITHUB_TOKEN` (built in) | Create the GitHub Release and read its assets. |
| npm trusted publisher (recommended) | Configure a trusted publisher for the `specharbor` package on npmjs.com pointing at `guferreira1/spec-harbor` and `.github/workflows/release.yml`. With `id-token: write` and npm >= 11.5, the `npm-publish` job publishes without a long-lived token. |
| `NPM_TOKEN` (fallback) | A granular npm automation token used as `NODE_AUTH_TOKEN` when a trusted publisher is not configured. |
| `HOMEBREW_TAP_GITHUB_TOKEN` | A token with write access to `guferreira1/homebrew-tap` so the `homebrew-publish` job can commit the formula. The default `GITHUB_TOKEN` cannot write to a separate repository. |

Tokens are read only from the `secrets` context and are never printed.

## Maintainer Release Checklist

1. Bump `packages/npm/specharbor/package.json` `version` to `X.Y.Z`.
2. Ensure the tag you will push is `vX.Y.Z` and matches the package version
   (locally: `sh scripts/validate-release-version.sh vX.Y.Z`).
3. Ensure the npm trusted publisher or `NPM_TOKEN` is configured.
4. Ensure `HOMEBREW_TAP_GITHUB_TOKEN` is configured.
5. Push the tag: `git tag vX.Y.Z && git push origin vX.Y.Z`.
6. Verify the GitHub Release assets and `checksums.txt`.
7. Verify the npm package: `npm view specharbor@X.Y.Z`.
8. Verify the Homebrew formula updated in `guferreira1/homebrew-tap`.
9. Verify `install.sh` resolves the new release.
10. Verify `specharbor version` prints `X.Y.Z`.

## Failure Handling and Rollback

- **GitHub Release succeeds but npm fails.** Re-run the `npm-publish` job; it is
  idempotent until the version exists on npm. npm forbids republishing the same
  version, so never attempt to overwrite — if the package version was already
  published, bump to a new `X.Y.Z`, retag, and release again.
- **npm succeeds but Homebrew fails.** Re-run the `homebrew-publish` job; it
  re-renders the formula from the existing release `checksums.txt` and pushes
  only if the formula changed, so re-runs are safe.
- **Version mismatch.** `validate-release-inputs` fails the release before any
  publish; fix `package.json` or the tag and release again.
- **Manual recovery.** The Homebrew formula can be regenerated locally with
  `scripts/render-homebrew-formula.sh` against a downloaded `checksums.txt` and
  committed to the tap by hand if CI is unavailable.

## Local Snapshot and Dry-Run Verification

Release workflow changes are reviewable without pushing a real tag. Pull
requests never publish because the workflow is tag-only, and every publishing
step has a local dry-run equivalent.

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

Validate the rest of the publishing path without publishing:

```bash
# Version consistency gate (and its tests).
sh scripts/validate-release-version.sh v0.2.0
sh scripts/test-validate-release-version.sh

# Homebrew formula rendering (and its tests).
sh scripts/test-render-homebrew-formula.sh

# npm package tests and contents.
cd packages/npm/specharbor
npm test
npm pack --dry-run
cd -
```

## Future Work

GitHub Releases, `checksums.txt`, `install.sh`, the npm wrapper package, and the
external Homebrew tap are automated and documented in [Install](install.md).
Publishing automation does not implement these future-only items:

- Native Linux packages such as `.deb`, `.rpm`, or `.apk`.
- Windows package-manager manifests such as Winget, Scoop, or Chocolatey.
- Signing, cosign, attestations, or SBOM generation.
- Docker images or Docker manifests.

Those channels and supply-chain features remain future work and require
separate OpenSpec changes or manual maintainer action.
