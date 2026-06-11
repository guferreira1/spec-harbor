# Acceptance Criteria: Polish Release Documentation

## Documentation Accuracy

- README documents the current real install options accurately.
- `docs/install.md` documents GitHub Release/install.sh, npm, and Homebrew accurately.
- `docs/release.md` reflects `v0.1.0` and the current release process accurately.
- `docs/usage.md` remains consistent with install/version guidance, if updated.
- `packages/npm/specharbor/README.md` no longer contradicts published package state, if updated.
- GitHub Release `v0.1.0` is documented as available.
- Release assets and `checksums.txt` are documented accurately.
- `install.sh` is documented as working with the real release.
- npm package name is documented as unscoped `specharbor`.
- Homebrew command is documented as `brew install guferreira1/tap/specharbor`.
- `specharbor version` is the standard verification command.
- Expected public version output uses `0.1.0`, while GitHub Release/tag references use `v0.1.0`.
- Release binary metadata is described as including version, commit, date, and dirty fields.

## Channel Boundaries

- Unsupported/future channels are not presented as available.
- Future-only package work remains clearly future-only.
- `.deb/.rpm`, Scoop, Winget, signing, SBOM, Docker, and publishing automation are documented only as future work.
- Documentation does not claim package publishing automation exists.
- Documentation makes clear that npm publishing for `0.1.0` was manual.
- Documentation makes clear that the Homebrew tap is manually maintained for now.
- `go install`, if documented, is framed as a fallback or developer/source-build option rather than the primary release channel.

## Troubleshooting

- Troubleshooting covers common install failures.
- Troubleshooting includes `PATH` issues.
- Troubleshooting includes permission denied errors.
- Troubleshooting includes checksum mismatch handling.
- Troubleshooting includes unsupported platform/architecture errors.
- Troubleshooting includes npm `--ignore-scripts` behavior and first-run binary download issues.
- Troubleshooting includes Homebrew tap/install issues.
- Troubleshooting includes stale shell command cache after installing.
- Troubleshooting includes verifying version metadata.

## Validation

- `go test ./...` passes.
- `go run ./cmd/specharbor validate polish-release-documentation` passes with 0 errors.
- Documentation-relevant project checks pass, if any exist.
- Documentation diffs are reviewed for consistency and scope.

## Scope Control

- No production code is modified.
- No release automation is modified.
- No GoReleaser config is modified.
- No `install.sh` behavior is modified.
- No npm package metadata/runtime files are modified unless explicitly justified as npm README-only docs.
- No Homebrew tap files are modified.
- No GitHub Release assets are modified.
- No tags are created or moved.
- No npm package is published.
- No Homebrew formula or tap repository is modified.
- No Linux packages, Windows package-manager manifests, signing artifacts, SBOMs, Docker artifacts, or publishing automation are added.
- CLI/product behavior is unchanged.
