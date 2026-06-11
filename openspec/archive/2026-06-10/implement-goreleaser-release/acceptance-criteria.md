# Acceptance Criteria: Implement GoReleaser Release

## GoReleaser Configuration

- `.goreleaser.yaml` exists at the repository root.
- `.goreleaser.yaml` uses community GoReleaser features only.
- `.goreleaser.yaml` does not require GoReleaser Pro.
- `.goreleaser.yaml` uses only builds, archives, checksums, and default GitHub Release assets.
- `.goreleaser.yaml` does not include Pro-only config fields, config sections, or behavior.
- `.goreleaser.yaml` sets the project name to `specharbor`.
- `.goreleaser.yaml` builds from `./cmd/specharbor`.
- `.goreleaser.yaml` names the binary `specharbor`.
- `.goreleaser.yaml` outputs artifacts under `dist/`.
- `.goreleaser.yaml` configures `linux`, `darwin`, and `windows`.
- `.goreleaser.yaml` configures `amd64` and `arm64`.
- `.goreleaser.yaml` includes `windows/arm64` unless implementation is explicitly blocked and reported.
- `.goreleaser.yaml` does not configure `npm`, `nfpms`, `brews`, `homebrew_casks`, `scoops`, `winget`, `aurs`, `nix`, `dockers`, `docker_manifests`, `sboms`, `signs`, `kos`, or `announce`.
- `.goreleaser.yaml` does not configure package-manager publishing, Docker publishing, SBOM generation, signing/cosign, notarization, attestations, announcement integrations, external publishers, or custom publishers.
- `.goreleaser.yaml` is reviewable without external tokens, GoReleaser Pro licenses, or GoReleaser Pro keys.

## Version Injection

- GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Version`.
- GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Commit`.
- GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Date`.
- GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Dirty`.
- `Version` receives `{{ .Version }}`.
- A release tag `v0.1.0` injects binary version `0.1.0`.
- `Commit` receives `{{ .FullCommit }}`.
- `Date` receives `{{ .Date }}` in UTC RFC3339 format.
- `Dirty` receives `{{ .IsGitDirty }}`.
- A clean CI release build injects `Dirty=false`.
- Runtime displays injected values as-is.
- Runtime does not normalize versions.
- Runtime does not inspect Git, read `.git`, execute Git, execute shell commands, call the network, or write files for release metadata.

## Release Assets

- Linux amd64 release asset is `specharbor_Linux_x86_64.tar.gz`.
- Linux arm64 release asset is `specharbor_Linux_arm64.tar.gz`.
- Darwin amd64 release asset is `specharbor_Darwin_x86_64.tar.gz`.
- Darwin arm64 release asset is `specharbor_Darwin_arm64.tar.gz`.
- Windows amd64 release asset is `specharbor_Windows_x86_64.zip`.
- Windows arm64 release asset is `specharbor_Windows_arm64.zip`.
- `checksums.txt` is generated.
- Checksums use SHA-256.
- No deb, rpm, apk, Homebrew formula, npm package, winget manifest, scoop manifest, chocolatey package, Docker image, SBOM, signature, or attestation asset is generated.

## GitHub Actions Release Workflow

- `.github/workflows/release.yml` exists.
- The release workflow triggers only on pushed tags matching `v*`.
- The release workflow does not trigger on pull requests.
- The release workflow does not trigger on normal branch pushes.
- The release workflow defines explicit top-level permissions.
- The release workflow permissions are exactly/minimally scoped for the approved release behavior.
- The release workflow includes `contents: write`.
- The release workflow does not include unrelated write permissions.
- The release workflow does not include `packages` permissions.
- The release workflow does not include `id-token` permissions.
- The release workflow does not include `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or other unrelated permissions.
- The release workflow uses `actions/checkout` with full history.
- The release workflow uses `actions/setup-go` with `go-version-file: go.mod`.
- The release workflow runs Go tests before publishing.
- The release workflow uses `goreleaser/goreleaser-action`.
- The release workflow uses the community `goreleaser` distribution.
- The release workflow runs `release --clean`.
- The release workflow uses only `secrets.GITHUB_TOKEN`.
- The release workflow does not reference npm tokens, Homebrew tokens, package-manager tokens, Docker registry tokens, signing keys, GoReleaser Pro keys, or personal access tokens.
- The release workflow does not create tags.
- The release workflow does not push commits.

## Snapshot And Local Dry Run

- `goreleaser check` succeeds.
- `goreleaser check` succeeds without Pro-only config requirements.
- `goreleaser release --snapshot --clean` succeeds.
- Snapshot artifacts are written under `dist/`.
- `dist/` is gitignored.
- Snapshot artifacts are not committed.
- At least one generated snapshot binary runs successfully.
- The generated snapshot binary's `specharbor version` output shows GoReleaser-injected metadata rather than the default `dev` and `unknown` fallback values.

## Documentation

- `README.md` explains that release tags use `vX.Y.Z`.
- `README.md` explains that release binaries display plain `X.Y.Z`.
- `README.md` explains that GoReleaser produces GitHub Release assets.
- `docs/release.md` documents the tag-based release workflow.
- `docs/release.md` documents the exact archive names.
- `docs/release.md` documents `checksums.txt`.
- `docs/release.md` documents local GoReleaser dry-run commands.
- `docs/release.md` states that npm, Homebrew, install scripts, package managers, signing, SBOM, and Docker images are future work.
- `docs/usage.md` remains accurate for `specharbor version`.
- Documentation does not describe unimplemented install channels as available.

## Tests And Verification

- Static/config tests verify `.goreleaser.yaml` exists and targets the correct binary/package.
- Static/config tests verify GoReleaser ldflags target the correct version package variables.
- Static/config tests verify release assets are configured for Linux, macOS, and Windows.
- Static/config tests verify checksums are enabled.
- Static/config tests verify package-manager artifacts are not configured.
- Static/config tests verify `.goreleaser.yaml` uses GoReleaser Community/free features only.
- Static/config tests verify GoReleaser Pro-only fields/config sections are absent.
- Static/config tests verify package-manager, Docker publishing, SBOM, signing/cosign, announcement, external publisher, and custom publisher sections are absent.
- Static/workflow tests verify the release workflow triggers only on tags.
- Static/workflow tests verify the release workflow does not trigger on normal pushes or pull requests.
- Static/workflow tests verify explicit top-level workflow permissions are exactly/minimally scoped.
- Static/workflow tests verify `contents: write` is present.
- Static/workflow tests verify `packages`, `id-token`, `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, and unrelated write permissions are absent.
- Static/workflow tests verify the release workflow does not use npm, Homebrew, package-manager, Docker, signing, or registry secrets.
- Existing runtime safety tests continue to pass.
- `go test ./...` passes.
- `go test -count=1 ./...` passes.
- `go run ./cmd/specharbor validate implement-goreleaser-release` passes.
- `goreleaser check` passes.
- `goreleaser release --snapshot --clean` passes.
