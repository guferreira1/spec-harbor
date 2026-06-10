# Design: Implement GoReleaser Release

## Overview

This change adds release automation at the repository boundary, not runtime release logic inside the CLI.

The implementation should add:

- `.goreleaser.yaml` for local and CI release assembly;
- `.github/workflows/release.yml` for tag-triggered GitHub Release publishing;
- focused tests or static checks that prove the release configuration is present, safe, and limited to the approved asset surface;
- documentation that explains the tag-based release flow and current distribution boundary.

No domain, use case, AI provider, workflow connector, or CLI business behavior is required for release publishing. The existing `internal/platform/version` package remains the metadata injection target.

## GoReleaser Configuration Decision

Add `.goreleaser.yaml`.

Configuration requirements:

- Use the community GoReleaser distribution only.
- Use config schema compatible with GoReleaser v2.
- Keep the file minimal and explicit.
- Set the project name to `specharbor`.
- Build only one binary named `specharbor`.
- Build from `./cmd/specharbor`.
- Keep `dist/` as the output directory.
- Use `CGO_ENABLED=0`.
- Use only the `go` builder, `archives`, `checksum`, and default GitHub release behavior needed to upload assets.
- Do not configure GoReleaser Pro-only fields, config sections, or behavior.
- Do not configure npm, Homebrew, nfpm, Docker, SBOM, signing, notarization, winget, scoop, chocolatey, snap, AUR, Nix, Krew, custom publishers, blob storage, or announcement integrations.
- Do not configure package-manager publishing, signing, SBOM, Docker publishing, announcement, or external publishing sections.
- Require `goreleaser check` as the authoritative validation that the config is accepted by the selected community GoReleaser distribution.
- Require `goreleaser release --snapshot --clean` as local proof that the community/free feature set can build the approved artifacts without external tokens.
- Keep the config reviewable without GoReleaser Pro licenses, GoReleaser Pro keys, package-manager tokens, signing keys, Docker registry credentials, or other external publishing credentials.

The approved GoReleaser feature surface is limited to:

- builds;
- archives;
- checksums;
- default GitHub Release asset upload.

Static checks must deny known out-of-scope or Pro-only publishing/config sections, including:

- `npm`;
- `nfpms`;
- `brews`;
- `homebrew_casks`;
- `scoops`;
- `winget`;
- `aurs`;
- `nix`;
- `dockers`;
- `docker_manifests`;
- `sboms`;
- `signs`;
- `kos`;
- `announce`.

If implementation discovers additional documented or detected GoReleaser Pro-only fields, package-manager sections, signing/SBOM/Docker publishing sections, or external publisher sections, those fields must also be denied by static checks or the implementer must stop and request a separate OpenSpec change.

The intended shape is:

```yaml
version: 2
project_name: specharbor

dist: dist

builds:
  - id: specharbor
    main: ./cmd/specharbor
    binary: specharbor
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - >-
        -s -w
        -X github.com/guferreira1/spec-harbor/internal/platform/version.Version={{ .Version }}
        -X github.com/guferreira1/spec-harbor/internal/platform/version.Commit={{ .FullCommit }}
        -X github.com/guferreira1/spec-harbor/internal/platform/version.Date={{ .Date }}
        -X github.com/guferreira1/spec-harbor/internal/platform/version.Dirty={{ .IsGitDirty }}

archives:
  - id: specharbor
    ids:
      - specharbor
    formats:
      - tar.gz
    format_overrides:
      - goos: windows
        formats:
          - zip
    name_template: >-
      {{ .ProjectName }}_{{ title .Os }}_{{ if eq .Arch "amd64" }}x86_64{{ else }}{{ .Arch }}{{ end }}

checksum:
  name_template: checksums.txt
  algorithm: sha256
```

The implementation may adjust YAML formatting if `goreleaser check` requires it, but it must preserve the decisions above.

## GitHub Actions Release Workflow Decision

Add `.github/workflows/release.yml`.

Workflow requirements:

- Trigger only on `push` tags matching `v*`.
- Do not trigger on normal branch pushes.
- Do not trigger on pull requests.
- Define top-level `permissions: contents: write` because GitHub Release creation/update and asset upload require repository contents write access.
- Do not grant any other workflow permissions unless a future implementation explicitly justifies them through a separate OpenSpec change.
- Use `actions/checkout` with `fetch-depth: 0` so GoReleaser can inspect tags and commits during CI.
- Use `actions/setup-go` with `go-version-file: go.mod`.
- Run Go tests before the release step.
- Use `goreleaser/goreleaser-action` with `distribution: goreleaser`, not `goreleaser-pro`.
- Use a v2-compatible GoReleaser version such as `~> v2`.
- Run `goreleaser release --clean`.
- Set only `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}` in the GoReleaser step.

The workflow must not include:

- `pull_request`;
- branch push triggers;
- scheduled triggers;
- manual publishing from arbitrary refs;
- broad permissions such as `write-all`;
- `packages` permissions;
- `id-token` permissions;
- `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or other unrelated permissions;
- npm tokens;
- Homebrew tokens;
- package registry tokens;
- GoReleaser Pro keys;
- signing keys;
- Docker registry credentials;
- commands that create tags;
- commands that push commits.

## Version Injection Decision

Use the existing package path:

```text
github.com/guferreira1/spec-harbor/internal/platform/version
```

Use the existing linker variables:

- `Version`
- `Commit`
- `Date`
- `Dirty`

Injection values:

| Variable | GoReleaser template | Decision |
| --- | --- | --- |
| `Version` | `{{ .Version }}` | Plain SemVer without leading `v` for normal `vX.Y.Z` release tags. |
| `Commit` | `{{ .FullCommit }}` | Full commit SHA for unambiguous provenance. |
| `Date` | `{{ .Date }}` | GoReleaser UTC RFC3339 build date. |
| `Dirty` | `{{ .IsGitDirty }}` | String-rendered boolean, `false` for clean CI release builds. |

For a clean release tag `v0.1.0`, the expected output from an uploaded release binary is:

```text
SpecHarbor 0.1.0
commit: <full commit sha>
date: <UTC RFC3339 build date>
dirty: false
```

Runtime behavior must not change. `specharbor version` displays injected values as-is. Runtime code must not strip or add `v`, inspect tags, read `.git`, call Git, execute a shell, call the network, or write files.

Snapshot builds may use GoReleaser's snapshot version value. That is acceptable for local verification as long as tag builds inject plain SemVer for `vX.Y.Z` tags.

## Release Asset Decision

Release assets must use these exact archive names for a normal release:

- `specharbor_Linux_x86_64.tar.gz`
- `specharbor_Linux_arm64.tar.gz`
- `specharbor_Darwin_x86_64.tar.gz`
- `specharbor_Darwin_arm64.tar.gz`
- `specharbor_Windows_x86_64.zip`
- `specharbor_Windows_arm64.zip`
- `checksums.txt`

Archive rules:

- Linux and macOS use `.tar.gz`.
- Windows uses `.zip`.
- `amd64` is named `x86_64` in archive filenames.
- `arm64` remains `arm64` in archive filenames.
- Archive contents may include GoReleaser's default README and license files plus the binary.
- No package-manager artifacts are produced.

Windows `arm64` is included because modern Go supports `windows/arm64`. If a concrete toolchain issue blocks this target, the implementer must stop and report it instead of silently dropping the target.

## Snapshot And Local Dry-Run Decision

Local verification must include:

```bash
goreleaser check
goreleaser release --snapshot --clean
```

Generated files must stay under `dist/`.

Snapshot artifacts must not be committed. The repository already ignores `/dist/`; the implementation must preserve that entry and add an ignore entry only if the target branch no longer has one.

Manual snapshot verification must inspect `dist/`, run at least one generated snapshot binary, and confirm `specharbor version` prints injected GoReleaser metadata rather than the default development fallback.

## CI Safety Boundaries

The release workflow may create or update GitHub Releases only as a result of pushed `v*` tags.

The workflow and GoReleaser config must not:

- publish from pull requests;
- publish from normal branch pushes;
- create Git tags;
- push commits;
- use tokens other than `secrets.GITHUB_TOKEN`;
- publish npm packages;
- publish Homebrew formulas;
- publish nfpm packages;
- publish Docker images;
- publish SBOMs;
- sign artifacts;
- create install scripts;
- upload package-manager manifests.

The implementation should add static checks that inspect `.goreleaser.yaml` and `.github/workflows/release.yml` for these boundaries.

## Documentation Plan

Update `README.md` and `docs/release.md`.

Update `docs/usage.md` only if needed to keep the version command documentation consistent.

Documentation must explain:

- release tags use `vX.Y.Z`;
- binary version metadata displays `X.Y.Z`;
- GoReleaser produces release builds;
- GitHub Release assets include OS/architecture archives and `checksums.txt`;
- `specharbor version` reports injected GoReleaser metadata for release binaries;
- GoReleaser release artifacts are the current distribution foundation;
- npm, Homebrew, install scripts, native Linux packages, Windows package managers, signing, SBOMs, and Docker images remain future work.

Do not document npm, Homebrew, `install.sh`, package managers, or Docker as available install channels.

## Testing Strategy

Add static/config tests that verify:

- `.goreleaser.yaml` exists.
- `.goreleaser.yaml` references `./cmd/specharbor`.
- `.goreleaser.yaml` builds binary `specharbor`.
- ldflags target the four required variables in `github.com/guferreira1/spec-harbor/internal/platform/version`.
- `Version` uses `{{ .Version }}`.
- `Commit` uses `{{ .FullCommit }}`.
- `Date` uses `{{ .Date }}`.
- `Dirty` uses `{{ .IsGitDirty }}`.
- Linux, Darwin, and Windows targets are configured.
- `amd64` and `arm64` targets are configured.
- Linux and macOS archives use `tar.gz`.
- Windows archives use `zip`.
- `checksums.txt` is configured with SHA-256.
- GoReleaser Community/free features are used only.
- GoReleaser Pro-only fields and config sections are absent.
- npm, nfpm, Homebrew, Homebrew cask, Scoop, Winget, AUR, Nix, Docker, Docker manifest, SBOM, signing/cosign, Ko, announce, package-manager, and external publisher sections are absent.
- `goreleaser check` passes without a Pro license, Pro-only behavior, package-manager credentials, signing keys, Docker credentials, or external tokens.

Add workflow tests or static checks that verify:

- `.github/workflows/release.yml` exists.
- The workflow triggers only on `push.tags` matching `v*`.
- The workflow has no `pull_request` trigger.
- The workflow has no branch push trigger.
- The workflow has explicit top-level permissions.
- The workflow permissions are exactly/minimally scoped for the approved release behavior.
- The workflow contains `contents: write`.
- The workflow does not contain unrelated write permissions.
- The workflow does not contain `packages` permissions.
- The workflow does not contain `id-token` permissions.
- The workflow does not contain `pull-requests`, `issues`, `deployments`, `security-events`, `actions`, `administration`, or other unrelated permissions.
- The workflow uses only `secrets.GITHUB_TOKEN`.
- The workflow uses the community GoReleaser distribution.
- The workflow does not use GoReleaser Pro.
- The workflow does not include commands that push commits or create tags.
- The workflow does not reference npm, Homebrew, package-manager, Docker, signing, or registry secrets.

Keep existing runtime release-versioning tests that prove no runtime Git lookup or version normalization is introduced.

Update the previous architecture/static test that intentionally forbade `.goreleaser.yaml` and release workflow files. Its new responsibility should be to allow the approved GoReleaser files while continuing to reject out-of-scope publishing and generated artifacts.

Repository verification must include:

```bash
go test ./...
go test -count=1 ./...
go run ./cmd/specharbor validate implement-goreleaser-release
goreleaser check
goreleaser release --snapshot --clean
```

Manual verification must include inspecting `dist/`, running a generated snapshot binary with `specharbor version`, and confirming generated artifacts are untracked and not committed.

## Architecture

Release automation belongs at the repository and CI boundary. It must not introduce new core use cases or runtime release services.

Architecture constraints:

- Core packages must not import adapters for release behavior.
- CLI packages must not contain release publishing rules.
- `internal/platform/version` remains a simple linker metadata target.
- Runtime code must not inspect Git tags, read `.git`, execute Git, call the shell, call the network, or write release files.
- GoReleaser, GitHub Actions, archive naming, and checksums stay outside runtime application code.
- This change does not require living architecture spec updates.
