# Acceptance Criteria: Release Publishing Automation

## Version consistency gate

- `scripts/validate-release-version.sh` accepts a `vX.Y.Z` tag whose `X.Y.Z`
  equals `packages/npm/specharbor/package.json` `version` and exits zero.
- The script exits non-zero for `0.2.0`, `v0.2`, `v0.2.0-beta.1`, and for any
  tag whose version does not equal the `package.json` version.
- `scripts/test-validate-release-version.sh` exercises the accept and reject
  cases and passes.

## Release workflow

- `.github/workflows/release.yml` triggers only on `push` tags matching `v*`
  and has no `pull_request`, `branches`, `schedule`, or `workflow_dispatch`
  trigger.
- The workflow defines ordered jobs `validate-release-inputs`, `goreleaser`,
  `npm-publish`, and `release-summary`, where `goreleaser` and `npm-publish`
  depend (directly or transitively) on `validate-release-inputs`, and
  `npm-publish` depends on `goreleaser`.
- Permissions are least privilege: top-level `contents: read`, `goreleaser`
  job `contents: write`, `npm-publish` job `contents: read` plus
  `id-token: write`.
- No secret value is echoed; tokens are referenced only as
  `${{ secrets.* }}` or GoReleaser `.Env` lookups.

## GitHub Release assets and install.sh

- `goreleaser check` passes and `goreleaser release --snapshot --clean`
  produces the six approved archives and `checksums.txt` with unchanged names.
- `install.sh` asset resolution (`specharbor_<Os>_<arch>.tar.gz` and
  `checksums.txt`) and the npm wrapper `assetName()` mapping still match the
  produced asset names.

## npm publishing

- The npm job publishes with `npm publish --provenance --access public` using
  OIDC trusted publishing, with a documented `NPM_TOKEN`/`NODE_AUTH_TOKEN`
  fallback, and runs only after version validation, the npm tests, and the
  packaged-contents check pass.
- `npm test` passes from `packages/npm/specharbor` with no skipped or disabled
  tests.
- `npm pack --dry-run` lists `bin/`, `lib/`, `scripts/`, `README.md`,
  `README.pt-BR.md`, and `package.json`, and omits `native/`, `node_modules/`,
  and test fixtures.
- `packages/npm/specharbor/package.json` has no lockfile, no runtime
  dependencies, and no publish lifecycle scripts.

## Homebrew tap

- `.goreleaser.yaml` is unchanged (no `brews`/cask block), and
  `goreleaser check` passes.
- `scripts/render-homebrew-formula.sh` renders a macOS formula for
  `guferreira1/homebrew-tap` that installs the prebuilt binary, includes a
  `specharbor version` `test do` block, and pins each download to its SHA-256
  from `checksums.txt`; `scripts/test-render-homebrew-formula.sh` passes.
- The `homebrew-publish` job (`needs: goreleaser`) updates the tap after
  archives and checksums exist, authenticates with `HOMEBREW_TAP_GITHUB_TOKEN`
  via `actions/checkout`, references the correct release assets and SHA-256
  values, and does not run on pull requests.

## Architecture guardrails and tests

- `internal/architecture/release_versioning_boundaries_test.go` is updated to
  the new scope and passes, still forbidding `nfpms`, `dockers`, `sboms`,
  `signs`, `cosign`, `scoops`, `winget`, `aurs`, `nix`, `chocolateys`,
  `goreleaser-pro`, `write-all`, and non-tag publish triggers.
- `go test ./...` and `go test -count=1 ./...` pass.
- `go run ./cmd/specharbor validate implement-release-publishing-automation`
  reports the change as valid.

## Documentation

- `docs/release.md` documents the automated GitHub Release, npm, and Homebrew
  publishing, the required secrets / trusted-publisher setup, a maintainer
  release checklist, failure handling and rollback notes, and dry-run/snapshot
  validation.
- Documentation still lists Linux native packages, Windows package managers,
  Docker, signing/cosign, attestations, and SBOMs as future work.

## Safety

- No real release, npm publish, tag push, or Homebrew tap write is performed
  by this change.
- No generated artifacts (`dist/`, archives, `checksums.txt`, `native/`,
  `node_modules/`) are staged or committed.
