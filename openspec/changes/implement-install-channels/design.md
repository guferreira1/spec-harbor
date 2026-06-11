# Design: Implement Install Channels

## Overview

This change plans the distribution layer that sits between published GitHub Releases and end users. It does not touch the application core: no domain, port, use case, or adapter changes are required. The deliverables are an install script, an npm wrapper package, a Homebrew formula in an external tap repository, and documentation.

All channels share one principle: they consume official GitHub Release assets for `guferreira1/spec-harbor`, verify checksums before use, and never build from source except through the explicitly documented manual `go install` fallback.

## Dependency Gate

This change was blocked on `implement-goreleaser-release`, which is now merged.

Reconciliation result (against the merged `.goreleaser.yaml` and `docs/release.md`): the original working assumption was `specharbor_<version>_<os>_<arch>.tar.gz|.zip` with lowercase OS names and `amd64`. The finalized GoReleaser output differs: asset names contain no version, OS names are title-cased, and `amd64` is rendered as `x86_64`. The finalized assets per release tag `vX.Y.Z` are:

```text
specharbor_Linux_x86_64.tar.gz
specharbor_Linux_arm64.tar.gz
specharbor_Darwin_x86_64.tar.gz
specharbor_Darwin_arm64.tar.gz
specharbor_Windows_x86_64.zip
specharbor_Windows_arm64.zip
checksums.txt
```

- The checksum file is `checksums.txt` with SHA-256 lines in `sha256sum` format (`<hex>  <asset-name>`).
- Each archive contains the `specharbor` binary at the archive root (`specharbor.exe` on Windows).
- Release asset download URLs follow `https://github.com/guferreira1/spec-harbor/releases/download/vX.Y.Z/<asset>`.
- At implementation time no GitHub Release had been published yet; reconciliation used the merged `.goreleaser.yaml` as the authoritative source. Channels resolve releases at run time and fail clearly when no release exists.
- Versions in injected binary metadata use plain `X.Y.Z` while Git tags and GitHub Release tags use `vX.Y.Z`, matching the merged release-versioning convention. Channel code handles that mapping explicitly and does not normalize what the binary itself reports.

## Supported Platforms (first stage)

| OS | Arch (asset name) | install.sh | npm | Homebrew |
| --- | --- | --- | --- | --- |
| Linux | x86_64 | yes | yes | future (Linuxbrew optional) |
| Linux | arm64 | yes | yes | future (Linuxbrew optional) |
| macOS (Darwin) | x86_64 | yes | yes | yes (planned) |
| macOS (Darwin) | arm64 | yes | yes | yes (planned) |
| Windows | x86_64 | no (future) | yes | no |
| Windows | arm64 | no (future) | yes | no |

Unsupported OS/arch combinations must fail with a clear error naming the detected platform and pointing to the manual install documentation. This matrix matches the platforms published by the merged `implement-goreleaser-release` pipeline, which includes Windows arm64.

## Decision 1: install.sh

Decision: add a POSIX shell install script at the repository root as `install.sh`, served raw from the default branch:

```bash
curl -sSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
```

Required behavior:

- Run under `sh` (POSIX), not bash-only features, with `set -eu`.
- Detect OS via `uname -s` (Linux, Darwin) and architecture via `uname -m`, mapping `x86_64`/`amd64 -> x86_64` and `aarch64`/`arm64 -> arm64` (the finalized asset arch names).
- Resolve the version to install:
  - default: latest release, resolved through the GitHub releases redirect endpoint `https://github.com/guferreira1/spec-harbor/releases/latest` or the public releases API, without authentication;
  - explicit: `SPECHARBOR_VERSION=v0.1.0` environment variable (or `--version v0.1.0` argument if argument parsing is implemented), accepting the `vX.Y.Z` tag form.
- Construct the asset URL deterministically from repository, tag, and the finalized asset naming pattern. No URL may come from user-controlled input other than the version string, which must be validated against a strict `v?[0-9]+\.[0-9]+\.[0-9]+` pattern before use.
- Download the archive and `checksums.txt` over HTTPS only, into a private temporary directory created with `mktemp -d`.
- Verify the archive's `sha256` against `checksums.txt` using `sha256sum` or `shasum -a 256`, whichever is available. Verification is mandatory: if no checksum tool is available, the script must fail with instructions, not skip verification.
- Extract only the expected binary from the archive and install it to the target directory with `0755` permissions.
- Install target selection:
  - explicit: `SPECHARBOR_INSTALL_DIR` environment variable (or `--install-dir` argument);
  - default: a user-local directory, preferring `$HOME/.local/bin`;
  - the script must not default to `/usr/local/bin` and must not invoke `sudo` itself; if the chosen directory is not writable, fail with a message suggesting a user-local directory, rather than escalating privileges.
- Warn (not fail) when the install directory is not on `PATH`, with copy-pasteable shell guidance.
- Print the installed path and suggest running `specharbor version` to verify.
- Failure behavior: any failure must leave no partial binary in the install target; the temporary directory must be cleaned with a `trap` on exit. Downloads, extraction, and verification all happen in the temporary directory; the final move into the install target happens only after checksum verification succeeds.
- Optional but recommended: a dry-run mode (`SPECHARBOR_DRY_RUN=1` or `--dry-run`) that prints the resolved OS, arch, version, asset URL, and install target without downloading or writing. If implemented, it must be tested.

Hard prohibitions:

- no execution of downloaded content as shell code (the script downloads archives and a checksum file only, never scripts);
- no `eval` over remote or user-controlled data;
- no unverified downloads;
- no package-manager side effects (no apt/dnf/brew/npm calls);
- no writes outside the temporary directory and the install target;
- no tokens, no authenticated requests;
- no telemetry.

## Decision 2: npm package

Decision: a manually maintained npm wrapper package named `specharbor` (unscoped), living in this repository under a dedicated directory such as `packages/npm/specharbor/`. Maintainer checks indicate the name appears available (`npm view specharbor` returned 404, `npm search specharbor` returned no matches); availability must be re-verified at publish time, and a scoped package is the fallback only if the name becomes unavailable. No npm org is required. No GoReleaser Pro features are used.

Binary acquisition strategy — trade-offs considered:

| Strategy | Pros | Cons |
| --- | --- | --- |
| Download at postinstall | Small package; one published package; standard pattern; install-time failure is visible immediately | Requires network at install; skipped under `--ignore-scripts`; postinstall scripts attract scrutiny |
| Download on first run | Works under `--ignore-scripts`; small package | Surprising network and write at first runtime use; breaks in offline/CI environments after a seemingly successful install |
| Package binaries directly (per-platform `optionalDependencies`) | No network after registry fetch; most robust | Requires publishing and maintaining several additional top-level unscoped package names without an org; highest publishing complexity for a first version |

Chosen first implementation: download at postinstall with checksum verification, plus a first-run fallback that performs the same verified download if the binary is missing (covering `--ignore-scripts` installs). Direct per-platform binary packages may be revisited in a future change if postinstall proves unreliable.

Required wrapper behavior:

- `package.json` declares `bin` mapping the `specharbor` command to a small Node launcher script; declares supported `os`/`cpu` where helpful; pins the wrapped SpecHarbor release version so each npm version maps to exactly one GitHub Release tag.
- Postinstall (and the first-run fallback) detects OS/arch from `process.platform` and `process.arch`, maps to the release asset matrix, constructs the asset URL from the pinned version and finalized naming pattern, downloads archive and checksum file over HTTPS, verifies `sha256` (using Node's `crypto`, no extra hashing dependency), extracts the binary into a package-local or cache directory, and sets executable permissions on POSIX systems.
- The launcher resolves the native binary path and forwards all CLI arguments and stdio to it via `child_process` `spawn`/`execFileSync`-style array-argument APIs, propagating the child exit code. It must never build a shell command string from arguments (no `exec` with string concatenation), so there is no shell-injection surface.
- Unsupported platform: fail with a clear error naming the platform and linking to manual install documentation, both at postinstall and at run time.
- URLs are restricted to `https://github.com/guferreira1/spec-harbor/releases/download/...` (and the latest-release resolution endpoint if ever needed); no other hosts, no HTTP, no redirect to non-HTTPS, no tokens.
- Writes are limited to the package's own directory or a dedicated cache directory; no writes to the user's project, no Git mutation, no telemetry.
- No automatic publishing: publishing to the npm registry is a separate, manual, explicitly later step outside this change.

## Decision 3: Homebrew tap

Decision: a personal tap, no org:

- tap repository: `guferreira1/homebrew-tap`;
- formula name: `specharbor`;
- user command: `brew install guferreira1/tap/specharbor`.

Formula requirements:

- `url` points at an official GitHub Release asset for a pinned version (macOS archives; Linux/Linuxbrew support optional and may be added later).
- `sha256` is mandatory for every referenced asset, copied from the release `checksums.txt`.
- The formula installs the prebuilt binary; it does not build from source.
- `test do` block runs `specharbor version` (via `shell_output` or equivalent) and asserts the output contains the expected version, validating that release metadata was injected.
- Formula validation (`brew audit`, `brew install`, `brew test`) can run later in a GitHub Actions macOS runner inside the tap repository; that automation is not part of this change.

Because the tap is a separate repository, the formula itself may be delivered through a separate repository/PR. Within this repository, the deliverables are: the documented tap path and install command, the formula content specification, and the testing expectations the tap must satisfy. Bumping the formula on each release is a manual step until later automation is specified.

## Decision 4: Linux packages

Decision: out of scope for the first install channels implementation.

- Future support for `.deb` and `.rpm` is documented as planned, likely via nfpm driven by the GoReleaser pipeline in a later change.
- No `nfpm.yaml`, `.nfpm.yaml`, `packaging/`, `debian/`, or `rpm/` files are added by this change unless explicitly approved later.
- Linux users are served first by `install.sh`, npm, and manual GitHub Release install.

## Decision 5: Windows package managers

Decision: out of scope for the first install channels implementation.

- Future support for Scoop and Winget is documented as planned.
- No `scoop/` or `winget/` manifests are added by this change unless explicitly approved later.
- Windows users are served first by the npm package and manual GitHub Release install (download `.zip`, verify checksum, place `specharbor.exe` on `PATH`).

## Security Model

All install channels must satisfy:

- HTTPS only for every download; no plain HTTP, no downgrade.
- Checksum (`sha256`) verification wherever a download occurs, against the official release checksum file; verification failure aborts the install and removes partial files.
- No execution of downloaded scripts from untrusted URLs; channels download archives and checksum files, never executable scripts. The only script users run is `install.sh` itself, fetched from this repository's default branch, and its content is reviewable in-repo.
- No arbitrary command execution; argument forwarding uses array-based process APIs, never shell string interpolation.
- No tokens or authentication required; only public, unauthenticated GitHub endpoints are used.
- No automatic publishing of packages; npm and Homebrew publishing are explicit, manual, later steps.
- No modification of SpecHarbor source code by any channel.
- No writes outside the install target and channel-local cache/temporary directories.
- No Git repository mutation; no tags, releases, branches, commits, or PRs created by any channel, script, or test.
- No telemetry, no usage reporting, no phoning home beyond the release downloads themselves.

## Documentation Plan

Add `docs/install.md` and update `README.md` (plus a pointer from `docs/release.md` and the README docs list). Documentation must explain:

- all installation options and which are available versus future-only;
- manual GitHub Release install: download the archive for the OS/arch, verify the checksum against `checksums.txt`, extract, place the binary on `PATH`;
- `install.sh` usage, including version pinning, install directory override, and the no-sudo default;
- npm global install (`npm install -g specharbor`), including postinstall download behavior, the `--ignore-scripts` first-run fallback, and unsupported-platform behavior;
- Homebrew tap install (`brew install guferreira1/tap/specharbor`);
- verifying an install with `specharbor version` and what release metadata should look like (versus `dev`/`unknown` for source builds);
- checksum verification commands for manual installs (`sha256sum -c` / `shasum -a 256 -c` guidance);
- troubleshooting `PATH` issues for user-local install directories;
- the manual `go install` fallback and its expected `dev`/`unknown` metadata;
- explicitly future-only: Linux `.deb`/`.rpm`, Scoop, Winget, signing, SBOM, Docker images, auto-update.

Until implementation lands and releases exist, documentation merged from this change must not present any channel as currently available; the channel docs land with the staged implementation.

## Staging Plan

- Stage 0 (gate): `implement-goreleaser-release` merged, first release published, asset names and checksum format confirmed.
- Stage 1: `install.sh` + `docs/install.md` + README updates + tests.
- Stage 2: npm wrapper package in-repo + tests (publishing remains a separate manual step).
- Stage 3: Homebrew formula specification; tap content delivered via the `guferreira1/homebrew-tap` repository, possibly through a separate PR.
- Future changes: Linux packages, Windows package managers, signing, automation of formula/package bumps.

## Testing Strategy

install.sh tests (shell-based test harness or Go-driven script tests, run against local fixtures and a stub HTTP server; never against real releases in CI):

- OS/arch detection mapping, including `x86_64`/`amd64 -> x86_64` and `aarch64`/`arm64 -> arm64`;
- asset URL construction for explicit versions and the latest-release resolution path;
- checksum verification success and failure (failure must abort and clean up);
- install target selection: default user-local directory and explicit override;
- no sudo invocation in any code path;
- dry-run mode output, if implemented;
- failure paths write nothing to the install target and remove temporary files.

npm wrapper tests:

- `package.json` name is `specharbor`, `bin` exposes the `specharbor` command, and the pinned release version maps to a single tag;
- OS/arch selection across supported and unsupported `process.platform`/`process.arch` values;
- checksum verification success and failure for the download path;
- launcher forwards arguments and exit codes to the native binary using array-argument process APIs;
- unsupported platform produces the documented error;
- no shell injection: argument forwarding never passes through a shell string;
- only the allowed `https://github.com/guferreira1/spec-harbor/...` URL prefix is used; tests reject any other host or scheme.

Homebrew expectations (validated in the tap repository, specified here):

- formula references an official GitHub Release asset URL;
- formula declares `sha256` for each asset;
- formula installs the binary;
- formula `test do` runs `specharbor version`;
- the documented tap path is `guferreira1/homebrew-tap` and the install command is `brew install guferreira1/tap/specharbor`.

Regression and safety checks in this repository:

- `go test ./...` passes; no changes to CLI runtime behavior, `internal/`, or `cmd/`. One documented exception: the static boundary test `internal/architecture/release_versioning_boundaries_test.go`, written for the earlier release-versioning change, asserted that `install.sh` and `packages/npm` must not exist and that `docs/release.md` lists install scripts as future work; those assertions were updated to reflect this approved change. No runtime code changed;
- no release asset generation occurs in this change: no `.goreleaser.yaml`/`.goreleaser.yml`, no release workflows, no archives, no generated checksum artifacts committed;
- no source-control automation: tests and scripts create no tags, releases, commits, or PRs;
- no package publishing occurs during tests: no `npm publish`, no registry calls, no tap pushes;
- network-dependent tests use local stubs/fixtures so CI never depends on real GitHub Releases or performs real downloads.

## Architecture

Install channels are distribution artifacts, not application code. They live outside `internal/` and `cmd/`:

- `install.sh` at the repository root;
- the npm wrapper under a dedicated directory such as `packages/npm/specharbor/`;
- the Homebrew formula in the external `guferreira1/homebrew-tap` repository.

Constraints:

- no imports or modifications of `internal/core`, `internal/adapters`, `internal/platform`, or `cmd/`;
- the binary remains the single source of version truth: channels must not patch, wrap, or reinterpret `specharbor version` output;
- hexagonal boundaries are unaffected; no new ports, use cases, or adapters are introduced;
- no living architecture spec update is required by this change.
