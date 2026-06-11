# Risks: Implement GoReleaser Release

## Risks

- GoReleaser configuration drift could break releases if the config uses outdated or Pro-only fields.
- GoReleaser Community and Pro feature boundaries may change between versions, so static denylist checks alone can become stale.
- The previous release-versioning tests intentionally rejected `.goreleaser.yaml` and release workflows, so incomplete test updates could leave conflicting architecture expectations.
- A broad GitHub Actions trigger could publish releases from pull requests or normal branch pushes.
- Workflow permission creep could accidentally broaden release capabilities beyond GitHub Release asset publishing.
- Extra secrets could expand the release trust boundary beyond GitHub's repository-scoped token.
- GoReleaser template mistakes could inject `v0.1.0` instead of `0.1.0`, a short commit instead of the decided full SHA, or an unstable date format.
- Archive naming templates could produce unexpected names, duplicate names, or platform names that do not match documentation.
- `windows/arm64` could expose a toolchain or GoReleaser target issue during dry-run verification.
- Snapshot artifacts in `dist/` could be accidentally committed.
- Documentation could imply npm, Homebrew, install scripts, package managers, signing, SBOM, or Docker images are available before they are implemented.
- Future install-channel changes could accidentally sneak into this release foundation instead of going through their own review.
- Future GoReleaser Pro features, package-manager publishing, signing, SBOM, or Docker support could expand the release boundary without a dedicated OpenSpec change.
- Release automation could tempt runtime code to inspect Git tags or normalize versions, which would violate the previous versioning boundary.

## Mitigations

- Use `goreleaser check` and `goreleaser release --snapshot --clean` as required verification, pairing static checks with GoReleaser's own config validation.
- Keep `.goreleaser.yaml` minimal and limited to builds, archives, checksums, and default GitHub Release publishing.
- Add static tests that reject Pro-only and out-of-scope publishing sections, and update them if GoReleaser documents or detects new Pro-only fields.
- Update the old exclusion tests so they allow only the approved GoReleaser files and continue rejecting package-manager artifacts, install scripts, generated archives, and unrelated publishing files.
- Configure `.github/workflows/release.yml` with only `on.push.tags: v*`.
- Add workflow static tests proving there is no pull request trigger and no normal branch push trigger.
- Use only `secrets.GITHUB_TOKEN` and exact top-level `permissions: contents: write`.
- Add workflow static tests proving no `packages`, `id-token`, `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or unrelated write permissions are present.
- Require future install channels, GoReleaser Pro features, package-manager publishing, signing, SBOM, and Docker support to use separate OpenSpec changes.
- Inject `Version={{ .Version }}`, `Commit={{ .FullCommit }}`, `Date={{ .Date }}`, and `Dirty={{ .IsGitDirty }}`.
- Require manual verification of a generated snapshot binary with `specharbor version`.
- Preserve existing runtime tests that reject Git, `.git`, shell, network, filesystem lookup, and version normalization in runtime version code.
- Preserve or add `/dist/` in `.gitignore` and inspect `git status --short` after snapshot builds.
- Document unavailable distribution channels explicitly as future work.

## Open Questions

- None blocking. The spec decides to include `windows/arm64`; if implementation verification proves it is not practical with the selected GoReleaser and Go versions, the implementer must stop and report the blocker instead of silently dropping the asset.

## Trade-Offs

- Full commit SHA is longer than a short SHA in `specharbor version`, but it avoids ambiguity for release provenance.
- Archive filenames omit the version because the GitHub Release tag already scopes the assets. This keeps stable asset names but requires users to keep tag context when downloading files outside GitHub.
- The first release foundation uses GoReleaser defaults for release notes rather than custom changelog automation. This keeps the scope small and leaves richer release notes for a future change.
- Signing, SBOMs, package-manager publishing, and install scripts are deferred to keep the first public release path reviewable and reversible.
