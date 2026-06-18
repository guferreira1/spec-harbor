# Tasks: Release Publishing Automation

## 1. Version consistency gate

- [x] Add `scripts/validate-release-version.sh` (POSIX `sh`) that reads the tag
      from `$1` or `GITHUB_REF_NAME`, rejects anything that is not exactly
      `vX.Y.Z` stable SemVer, strips the leading `v`, reads
      `packages/npm/specharbor/package.json` `version`, and fails on any
      mismatch. It must not print tokens and must exit non-zero on failure.
- [x] Add `scripts/test-validate-release-version.sh` covering: valid `vX.Y.Z`
      matching `package.json`; rejected `0.2.0`, `v0.2`, and `v0.2.0-beta.1`;
      and a tag/`package.json` mismatch.
- [x] Make both scripts executable and confirm they run from the repository
      root with no network access.

## 2. GoReleaser configuration and Homebrew formula

- [x] Keep `.goreleaser.yaml` byte-compatible with `main` (no `brews`/cask
      block) so `builds`, `archives`, and `checksum` and the six approved
      archive names and `checksums.txt` are unchanged and `goreleaser check`
      stays green. GoReleaser `brews` is deprecated and `homebrew_casks` would
      break the existing formula tap.
- [x] Add `scripts/render-homebrew-formula.sh` (POSIX `sh`) that takes a plain
      `X.Y.Z` version and a `checksums.txt`, extracts the macOS arm64 and
      x86_64 SHA-256 values, and prints a macOS formula installing the prebuilt
      binary with a `test do` block running `specharbor version`.
- [x] Add `scripts/test-render-homebrew-formula.sh` covering the rendered URLs,
      pinned SHA-256 values, install/test blocks, and rejection of missing
      checksums or non-`X.Y.Z` versions.
- [x] Confirm `goreleaser check` passes and `goreleaser release --snapshot
      --clean` builds without attempting any tap write.

## 3. Release workflow

- [x] Restructure `.github/workflows/release.yml` into ordered jobs:
      `validate-release-inputs`, `goreleaser`, `npm-publish`,
      `homebrew-publish`, and `release-summary`, keeping the trigger
      `on.push.tags: ["v*"]`.
- [x] Set least-privilege permissions: top-level `contents: read`;
      `goreleaser` job `contents: write`; `npm-publish` job `contents: read`
      and `id-token: write`; `homebrew-publish` job `contents: read`.
- [x] Run `scripts/validate-release-version.sh` in `validate-release-inputs`
      and make the `goreleaser` job depend on it.
- [x] Run `go test ./...` and GoReleaser `release --clean` in the `goreleaser`
      job with `GITHUB_TOKEN`; do not echo secrets.
- [x] In `npm-publish` (`needs: goreleaser`): set up Node against
      `https://registry.npmjs.org`, ensure npm >= 11.5, re-validate the
      version, run the npm tests and packaged-contents validation, then
      `npm publish --provenance --access public` using OIDC trusted publishing
      with an `NPM_TOKEN`/`NODE_AUTH_TOKEN` fallback.
- [x] In `homebrew-publish` (`needs: goreleaser`): download the release
      `checksums.txt` with `GITHUB_TOKEN`, render the formula, check out
      `guferreira1/homebrew-tap` with `HOMEBREW_TAP_GITHUB_TOKEN`, and commit
      and push `Formula/specharbor.rb` only when it changed.
- [x] Add a Homebrew asset availability gate in `homebrew-publish` after
      downloading `checksums.txt` and before checking out the tap: require
      checksum entries for both macOS archives and download each archive with
      `gh release download` before rendering the formula.
- [x] Ensure no job publishes on pull requests or branches and that publishing
      runs only after validation and tests pass.

## 4. npm package validation

- [x] Add a packaged-contents assertion in
      `packages/npm/specharbor/test/package.test.js` that the `files` set and
      tarball include `bin`, `lib`, `scripts`, `README.md`, `README.pt-BR.md`,
      and `package.json`, and exclude `native`, `node_modules`, and test
      fixtures.
- [x] Confirm `npm test` and `npm pack --dry-run` pass from
      `packages/npm/specharbor` and that the dry-run lists `README.pt-BR.md`.
- [x] Do not add a lockfile, runtime dependencies, or publish lifecycle scripts
      to `package.json`.

## 5. Architecture guardrail tests

- [x] Update `internal/architecture/release_versioning_boundaries_test.go` so
      the workflow guardrails allow `release.yml` to contain the ordered
      publishing jobs, the npm-job `id-token: write` permission, the top-level
      `contents: read` default, and the publishing secrets, plus a new test
      asserting job ordering. Leave the GoReleaser-config guardrails unchanged
      because `.goreleaser.yaml` is unchanged.
- [x] Keep the test forbidding genuinely out-of-scope channels and behaviors:
      `dockers`, `sboms`, `signs`, `cosign`, `scoops`, `winget`,
      `chocolatey`, `nfpm`, `goreleaser-pro`, `write-all`, `pull_request`,
      `branches`, `schedule`, and `workflow_dispatch`.
- [x] Keep `TestReleaseVersioningDocumentationDescribesImplementedScopeOnly`
      passing by preserving the required documentation snippets.

## 6. Documentation

- [x] Update `docs/release.md` to describe automated GitHub Release assets,
      automated npm publishing, automated Homebrew tap updates, the required
      secrets and trusted-publisher setup, a maintainer release checklist,
      failure handling and rollback notes, and dry-run/snapshot validation.
- [x] Update `docs/install.md` only where channel status or Homebrew platform
      coverage text changes (no behavior change for users), and update the
      `README.md` distribution section if its "future work" wording needs to
      reflect automated publishing.
- [x] Keep documentation honest about what remains future work: Linux native
      packages, Windows package managers, Docker, signing/cosign,
      attestations, and SBOMs.

## 7. Verification

- [x] Run `gofmt` on changed Go files and `go test ./...` /
      `go test -count=1 ./...`.
- [x] Run `go run ./cmd/specharbor validate
      implement-release-publishing-automation`.
- [ ] Run `goreleaser check` and `goreleaser release --snapshot --clean`, then
      remove the generated `dist/` so nothing is staged.
- [x] Run `npm test` and `npm pack --dry-run` from `packages/npm/specharbor`.
- [x] Confirm no real publish, tag push, or tap write occurred.
