# Tasks: Polish Release Documentation

## Phase 0: Preparation

- [x] Verify worktree and branch state.
- [x] Inspect `README.md`.
- [x] Inspect `docs/install.md`, if present.
- [x] Inspect `docs/release.md`, if present.
- [x] Inspect `docs/usage.md`, if present.
- [x] Inspect `packages/npm/specharbor/README.md`, if present.
- [x] Inspect `packages/npm/specharbor/package.json`, if present, for package name and version context only.
- [x] Inspect `install.sh`, if present, without changing behavior.
- [x] Inspect `.goreleaser.yaml` without changing release config.
- [x] Inspect `.github/workflows/release.yml` without changing release automation.
- [x] Confirm GitHub Release `v0.1.0` exists.
- [x] Confirm the `v0.1.0` release commit is `e6faff91feef07e5c1e47181243286268daf17b5`.
- [x] Confirm release assets and checksums are documented accurately.
- [x] Confirm npm package `specharbor@0.1.0` is published.
- [x] Confirm Homebrew tap `guferreira1/tap/specharbor` is available.
- [x] Confirm the external tap repository is `guferreira1/homebrew-tap`.

## Phase 1: Documentation Updates

- [x] Update README install section.
- [x] Update install documentation.
- [x] Update release documentation.
- [x] Update usage documentation only if needed.
- [x] Update npm README only if needed.
- [x] Add or refine channel status matrix.
- [x] Add or refine troubleshooting section.
- [x] Remove stale "future work" wording for npm/Homebrew.
- [x] Preserve future-only wording for `.deb/.rpm`, Scoop/Winget, signing, SBOM, Docker, and publishing automation.
- [x] Ensure GitHub Release `v0.1.0` is documented as available.
- [x] Ensure `install.sh` is documented as working with the real release.
- [x] Ensure npm package name is documented as unscoped `specharbor`.
- [x] Ensure Homebrew command is documented exactly as `brew install guferreira1/tap/specharbor`.
- [x] Ensure `go install` is documented only as a fallback or developer option if retained.
- [x] Ensure `specharbor version` is the standard install verification command.
- [x] Ensure references to `v0.1.0` and `0.1.0` are correct for tag versus binary version contexts.
- [x] Verify docs do not claim npm/Homebrew automation.
- [x] Verify docs do not claim unsupported package managers are available.

## Phase 2: Troubleshooting and Consistency Checks

- [x] Verify troubleshooting covers `PATH` issues.
- [x] Verify troubleshooting covers permission denied errors.
- [x] Verify troubleshooting covers checksum mismatch handling.
- [x] Verify troubleshooting covers unsupported platform/architecture errors.
- [x] Verify troubleshooting covers npm `--ignore-scripts` behavior.
- [x] Verify troubleshooting covers npm first-run binary download issues.
- [x] Verify troubleshooting covers Homebrew tap/install issues.
- [x] Verify troubleshooting covers stale shell command cache after installing.
- [x] Verify troubleshooting covers version metadata verification.
- [x] Verify docs mention available channels correctly.
- [x] Verify docs keep future-only channels clearly future-only.
- [x] Verify install commands are correct.
- [x] Verify version verification command is consistent.
- [x] Verify troubleshooting content exists and is not misleading.

## Phase 3: Validation

- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate polish-release-documentation`.
- [x] Run documentation-relevant checks available in the project, if any.
- [x] Inspect git status and docs diffs.
- [x] Confirm no unintended code/release/package/tap changes.
- [x] Confirm no production code, release automation, `install.sh`, npm metadata, GoReleaser config, or Homebrew tap files changed.
- [x] Confirm no release assets, tags, packages, formula files, signing artifacts, SBOMs, Docker artifacts, Linux packages, or Windows package-manager manifests were created.
