# Tasks: Implement GoReleaser Release

## Phase 0: Baseline And Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-goreleaser-release/`.
- [x] Inspect `README.md`, `docs/usage.md`, and `docs/release.md`.
- [x] Inspect `internal/platform/version/` and confirm the four linker variables are still `Version`, `Commit`, `Date`, and `Dirty`.
- [x] Inspect `cmd/specharbor/` and confirm the binary entry point remains `./cmd/specharbor`.
- [x] Inspect `.github/workflows/` and preserve existing CI behavior.
- [x] Inspect `.gitignore` and confirm whether `/dist/` is already ignored.
- [x] Inspect existing release-versioning architecture tests and identify checks that must change from "no GoReleaser files" to "only approved GoReleaser files".
- [x] Confirm this change is limited to release config, release workflow, tests/static checks, documentation, and the OpenSpec task updates after implementation.

## Phase 1: GoReleaser Configuration

- [x] Add `.goreleaser.yaml` using GoReleaser v2 community features only.
- [x] Set `project_name` to `specharbor`.
- [x] Keep `dist` set to `dist`.
- [x] Configure a single build id for `specharbor`.
- [x] Set `main` to `./cmd/specharbor`.
- [x] Set `binary` to `specharbor`.
- [x] Set `CGO_ENABLED=0`.
- [x] Configure `goos` exactly for `linux`, `darwin`, and `windows`.
- [x] Configure `goarch` exactly for `amd64` and `arm64`.
- [x] Do not silently omit `windows/arm64`; stop and report if the target is not practical.
- [x] Configure ldflags for `Version={{ .Version }}`.
- [x] Configure ldflags for `Commit={{ .FullCommit }}`.
- [x] Configure ldflags for `Date={{ .Date }}`.
- [x] Configure ldflags for `Dirty={{ .IsGitDirty }}`.
- [x] Preserve runtime behavior that displays injected values as-is.
- [x] Do not add runtime version normalization.
- [x] Do not add runtime Git, `.git`, shell, filesystem, or network lookup for metadata.

## Phase 2: Archives And Checksums

- [x] Configure Linux archives as `.tar.gz`.
- [x] Configure macOS archives as `.tar.gz`.
- [x] Configure Windows archives as `.zip`.
- [x] Configure release archive names exactly as `specharbor_Linux_x86_64.tar.gz`, `specharbor_Linux_arm64.tar.gz`, `specharbor_Darwin_x86_64.tar.gz`, `specharbor_Darwin_arm64.tar.gz`, `specharbor_Windows_x86_64.zip`, and `specharbor_Windows_arm64.zip`.
- [x] Configure `checksums.txt`.
- [x] Configure SHA-256 checksums.
- [x] Do not configure nfpm packages.
- [x] Do not configure Homebrew formulas or taps.
- [x] Do not configure npm packages.
- [x] Do not configure Scoop, Winget, AUR, Nix, or other package-manager publishing sections.
- [x] Do not configure Docker images.
- [x] Do not configure Docker manifests or Ko publishing.
- [x] Do not configure SBOM generation.
- [x] Do not configure signing, cosign, notarization, or attestations.
- [x] Do not configure announcement integrations, external publishers, or custom publishers.
- [x] Do not configure any documented or detected GoReleaser Pro-only field, config section, or behavior.
- [x] Do not configure Windows package-manager manifests.
- [x] Do not configure Linux package-manager artifacts.

## Phase 3: GitHub Actions Release Workflow

- [x] Add `.github/workflows/release.yml`.
- [x] Trigger the release workflow only for pushed tags matching `v*`.
- [x] Do not add a pull request trigger.
- [x] Do not add a branch push trigger.
- [x] Do not add scheduled publishing.
- [x] Define explicit top-level workflow permissions.
- [x] Use exactly `permissions: contents: write` for the approved release behavior.
- [x] Do not add broad workflow permissions such as `write-all`.
- [x] Do not add `packages` permissions.
- [x] Do not add `id-token` permissions.
- [x] Do not add `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or other unrelated permissions.
- [x] Use `actions/checkout` with `fetch-depth: 0`.
- [x] Use `actions/setup-go` with `go-version-file: go.mod`.
- [x] Run Go tests before the GoReleaser release step.
- [x] Use `goreleaser/goreleaser-action`.
- [x] Use `distribution: goreleaser`, not `goreleaser-pro`.
- [x] Use a v2-compatible GoReleaser version.
- [x] Run `goreleaser release --clean`.
- [x] Set only `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}` for publishing.
- [x] Do not reference npm, Homebrew, package-manager, Docker, signing, or registry secrets.
- [x] Do not add commands that create tags.
- [x] Do not add commands that push commits.

## Phase 4: Tests And Static Checks

- [x] Add or update tests proving `.goreleaser.yaml` exists.
- [x] Add or update tests proving `.goreleaser.yaml` references `./cmd/specharbor` and binary `specharbor`.
- [x] Add or update tests proving GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Version`.
- [x] Add or update tests proving GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Commit`.
- [x] Add or update tests proving GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Date`.
- [x] Add or update tests proving GoReleaser ldflags target `github.com/guferreira1/spec-harbor/internal/platform/version.Dirty`.
- [x] Add or update tests proving the `Version`, `Commit`, `Date`, and `Dirty` templates are `{{ .Version }}`, `{{ .FullCommit }}`, `{{ .Date }}`, and `{{ .IsGitDirty }}`.
- [x] Add or update tests proving Linux, Darwin, and Windows targets are configured.
- [x] Add or update tests proving `amd64` and `arm64` targets are configured.
- [x] Add or update tests proving Linux/macOS archives use `tar.gz` and Windows archives use `zip`.
- [x] Add or update tests proving `checksums.txt` and SHA-256 checksums are configured.
- [x] Add or update tests proving npm, Homebrew, nfpm, Docker, signing, SBOM, package-manager, and custom publisher sections are absent.
- [x] Add or update tests proving `.goreleaser.yaml` uses only GoReleaser Community/free features.
- [x] Add or update tests proving GoReleaser Pro-only fields/config sections are absent.
- [x] Add or update tests proving npm, nfpm, Homebrew, Homebrew cask, Scoop, Winget, AUR, Nix, and other package-manager publishing sections are absent.
- [x] Add or update tests proving Docker publishing, Docker manifest, SBOM, signing/cosign, Ko, announce, external publisher, and custom publisher sections are absent.
- [x] Add or update tests proving `.goreleaser.yaml` is reviewable without external tokens, GoReleaser Pro licenses, or GoReleaser Pro keys.
- [x] Add or update tests proving `.github/workflows/release.yml` exists.
- [x] Add or update tests proving the release workflow triggers only on `push.tags` matching `v*`.
- [x] Add or update tests proving the release workflow has no pull request trigger.
- [x] Add or update tests proving the release workflow has no branch push trigger.
- [x] Add or update tests proving the release workflow has explicit top-level permissions.
- [x] Add or update tests proving workflow permissions are exactly/minimally scoped for approved release behavior.
- [x] Add or update tests proving `contents: write` is present.
- [x] Add or update tests proving unrelated write permissions are absent.
- [x] Add or update tests proving `packages` permissions are absent.
- [x] Add or update tests proving `id-token` permissions are absent.
- [x] Add or update tests proving `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, and other unrelated permissions are absent.
- [x] Add or update tests proving the release workflow uses only `secrets.GITHUB_TOKEN`.
- [x] Add or update tests proving the release workflow does not use GoReleaser Pro.
- [x] Add or update tests proving the release workflow does not reference npm, Homebrew, package-manager, Docker, signing, or registry secrets.
- [x] Preserve tests proving runtime version code does not read `.git`, inspect tags, run Git, execute shell commands, call the network, or normalize versions.
- [x] Preserve tests proving core packages do not import adapters for release behavior.
- [x] Update the earlier release-versioning exclusion checks so they allow only the approved GoReleaser config and release workflow while continuing to reject out-of-scope package and publishing files.

## Phase 5: Documentation

- [x] Update `README.md` to describe tag-based GoReleaser release assets as implemented release behavior.
- [x] Update `docs/release.md` to describe the GoReleaser release flow.
- [x] Update `docs/release.md` to list the exact release archive names and `checksums.txt`.
- [x] Update `docs/release.md` to explain that release tags use `vX.Y.Z`.
- [x] Update `docs/release.md` to explain that binary version metadata displays `X.Y.Z`.
- [x] Update `docs/release.md` to explain that GoReleaser injects `Version`, `Commit`, `Date`, and `Dirty`.
- [x] Update `docs/release.md` to explain local dry-run commands with `goreleaser check` and `goreleaser release --snapshot --clean`.
- [x] Update `docs/usage.md` only if the version command or release metadata description needs consistency updates.
- [x] Keep docs clear that npm, Homebrew, install scripts, Linux packages, Windows package managers, signing, SBOM, and Docker images are future work.
- [x] Avoid documenting unimplemented install commands or package-manager channels as available.

## Phase 6: Local Verification

- [x] Run `go test ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-goreleaser-release`.
- [x] Run `goreleaser check`.
- [x] Run `goreleaser release --snapshot --clean`.
- [x] Confirm `goreleaser check` passes without Pro-only config requirements.
- [x] Inspect `dist/` after the snapshot build.
- [x] Run at least one generated snapshot binary and verify `specharbor version` prints GoReleaser-injected metadata.
- [x] Confirm generated snapshot artifacts under `dist/` are not committed.
- [x] Run `git status --short`.
- [x] Inspect `git diff -- .goreleaser.yaml .github/workflows/release.yml README.md docs/release.md docs/usage.md internal/architecture openspec/changes/implement-goreleaser-release/`.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.

## Phase 7: Final Safety Review

- [x] Confirm the release workflow does not publish on pull requests.
- [x] Confirm the release workflow does not publish on normal branch pushes.
- [x] Confirm the release workflow does not require secrets beyond `GITHUB_TOKEN`.
- [x] Confirm no npm package files were added.
- [x] Confirm no Homebrew files were added.
- [x] Confirm no nfpm, deb, rpm, apk, winget, scoop, or chocolatey files were added.
- [x] Confirm no install script was added.
- [x] Confirm no Docker, SBOM, signing, cosign, or attestation configuration was added.
- [x] Confirm no GoReleaser Pro-only config fields, config sections, or behavior were added.
- [x] Confirm workflow permissions are exactly/minimally scoped to `contents: write`.
- [x] Confirm no `packages`, `id-token`, `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or unrelated write permissions were added.
- [x] Confirm no source-control push or tag automation was added.
- [x] Confirm no runtime release logic was added to core packages.
- [x] Confirm no runtime Git lookup was introduced.
