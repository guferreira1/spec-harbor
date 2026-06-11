# Tasks: Implement Install Channels

Implementation is gated: do not start Phase 1 or later until Phase 0 confirms the `implement-goreleaser-release` outputs are finalized.

## Phase 0: Dependency Gate and Baseline

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-install-channels/`.
- [x] Confirm `implement-goreleaser-release` is merged. No GitHub Release is published yet, so the finalized asset names were reconciled against the merged `.goreleaser.yaml` and `docs/release.md` as the authoritative source; channels resolve releases at run time and fail clearly when none exists.
- [x] Record the finalized release asset naming pattern, archive formats per OS/arch, checksum file name and format, and binary name inside each archive.
- [x] Reconcile every asset-name assumption in this change's `design.md` against the finalized release assets; update this change first if they differ.
- [x] Confirm the supported platform matrix in `design.md` matches the platforms actually published by the release pipeline.
- [x] Confirm this change makes no modifications to CLI runtime behavior, `cmd/`, or `internal/` outside the documented static boundary-test update in `internal/architecture/release_versioning_boundaries_test.go` (its earlier assertions forbade `install.sh` and `packages/npm`, which this approved change introduces).
- [x] Confirm no GoReleaser config, release workflow, release asset generation, npm publishing, Homebrew publishing, Linux package, Windows package-manager, signing, SBOM, Docker, auto-update, or telemetry work is performed in this change.
- [x] Stop and report blocked if the dependency gate cannot be confirmed.

## Phase 1: install.sh

- [x] Add `install.sh` at the repository root as a POSIX `sh` script with `set -eu` and a cleanup `trap` for its temporary directory.
- [x] Implement OS detection from `uname -s` supporting Linux and Darwin, failing clearly on other systems.
- [x] Implement architecture detection from `uname -m`, mapping `x86_64`/`amd64 -> x86_64` and `aarch64`/`arm64 -> arm64` (the finalized asset arch names), failing clearly on other architectures.
- [x] Implement version resolution: latest release by default, explicit version via `SPECHARBOR_VERSION` (and `--version` if argument parsing is added).
- [x] Validate explicit version strings against a strict `v?[0-9]+\.[0-9]+\.[0-9]+` pattern before using them in URLs.
- [x] Construct the release asset URL and checksum file URL deterministically from the finalized naming pattern, restricted to `https://github.com/guferreira1/spec-harbor/` endpoints.
- [x] Download the archive and checksum file over HTTPS only into a `mktemp -d` directory.
- [x] Verify the archive `sha256` against the checksum file with `sha256sum` or `shasum -a 256`; fail with instructions when neither tool exists; never skip verification.
- [x] Extract the binary and install it with `0755` permissions only after checksum verification succeeds.
- [x] Implement install target selection: `SPECHARBOR_INSTALL_DIR` override, defaulting to a user-local directory preferring `$HOME/.local/bin`.
- [x] Do not invoke `sudo` anywhere in the script; fail with user-local guidance when the target directory is not writable.
- [x] Warn with copy-pasteable guidance when the install directory is not on `PATH`.
- [x] Print the installed path and suggest `specharbor version` for verification.
- [x] Ensure failures leave no partial files in the install target and remove the temporary directory.
- [x] If a dry-run mode is implemented, make it print resolved OS, arch, version, asset URL, and install target without downloading or writing.
- [x] Ensure the script executes no downloaded content, uses no `eval` over remote or user-controlled data, performs no package-manager calls, requires no tokens, and writes only to its temporary directory and the install target.

## Phase 2: npm Wrapper Package

- [x] Create the npm wrapper package directory (for example `packages/npm/specharbor/`) without touching `internal/` or `cmd/`.
- [x] Define `package.json` with name `specharbor`, a `bin` entry exposing the `specharbor` command through a Node launcher, and a pinned SpecHarbor release version mapping to exactly one GitHub Release tag.
- [x] Re-verify npm name availability at publish time; document the scoped-package fallback only if `specharbor` becomes unavailable. (The publish-time re-verification requirement and scoped fallback are documented in `docs/install.md` and the package README; publishing remains a manual later step.)
- [x] Implement OS/arch detection from `process.platform` and `process.arch` mapped to the release asset matrix.
- [x] Implement the postinstall download: construct the asset URL from the pinned version, download archive and checksum file over HTTPS from `https://github.com/guferreira1/spec-harbor/releases/download/` only, verify `sha256` with Node `crypto`, extract the binary into a package-local or cache directory, and set executable permissions on POSIX systems.
- [x] Implement the first-run fallback: when the binary is missing (for example after `--ignore-scripts`), perform the same verified download before executing.
- [x] Implement the launcher to forward all arguments and stdio to the native binary with array-argument process APIs and propagate the child exit code; never build shell command strings.
- [x] Fail with a clear, documented error naming the platform and linking to manual install docs on unsupported platforms, at postinstall and at run time.
- [x] Ensure the wrapper requires no tokens, performs no telemetry, mutates no Git state, and writes only inside its package or cache directory.
- [x] Do not run `npm publish` or any registry publishing in this change.

## Phase 3: Homebrew Tap Specification

- [x] Document the tap repository `guferreira1/homebrew-tap`, the formula name `specharbor`, and the install command `brew install guferreira1/tap/specharbor`.
- [x] Specify the formula content: `url` pointing at an official pinned GitHub Release asset, mandatory `sha256` per asset, binary install (no source build), and a `test do` block running `specharbor version` asserting the expected version.
- [x] Document that formula delivery may happen through a separate repository/PR in the tap repository, and that formula bumps are manual until later automation is specified.
- [x] Document that `brew audit`, `brew install`, and `brew test` validation can run later in a GitHub Actions macOS runner inside the tap repository.
- [x] Do not commit Homebrew formula or tap files into this repository.

## Phase 4: Tests

- [x] Add install.sh tests covering OS/arch detection mapping, including `x86_64`/`amd64 -> x86_64` and `aarch64`/`arm64 -> arm64`.
- [x] Add install.sh tests covering asset URL construction for explicit versions and the latest-release resolution path.
- [x] Add install.sh tests covering checksum verification success and failure, with failure aborting and cleaning up.
- [x] Add install.sh tests covering install target selection: default user-local directory and explicit override.
- [x] Add install.sh tests or static checks confirming no `sudo` invocation exists in any code path.
- [x] Add install.sh tests for dry-run mode output if dry-run is implemented.
- [x] Add install.sh tests confirming failure paths write nothing to the install target and leave no partial files.
- [x] Run install.sh tests against local fixtures or a stub HTTP server, never against real GitHub Releases in CI.
- [x] Add npm tests confirming the package name is `specharbor` and `bin` exposes the `specharbor` command.
- [x] Add npm tests covering OS/arch selection for supported and unsupported `process.platform`/`process.arch` values.
- [x] Add npm tests covering checksum verification success and failure for the download path.
- [x] Add npm tests confirming the launcher forwards arguments and exit codes to the native binary.
- [x] Add npm tests confirming unsupported platforms produce the documented error.
- [x] Add npm tests or static checks confirming no shell-string command construction (no shell-injection surface).
- [x] Add npm tests confirming only the allowed `https://github.com/guferreira1/spec-harbor/` URL prefix is used and other hosts or schemes are rejected.
- [x] Record the Homebrew formula expectations as testable criteria for the tap repository: release asset URL, `sha256`, binary install, `test do` running `specharbor version`, and the documented tap path.
- [x] Add or preserve regression coverage confirming CLI runtime behavior is unchanged (`go test ./...` passes; the only `internal/` change is the documented static boundary-test update, no `cmd/` modifications).
- [x] Add checks confirming this change generates no release assets, archives, or checksum artifacts and adds no `.goreleaser.yaml`, `.goreleaser.yml`, or release workflow files.
- [x] Confirm tests perform no source-control automation and create no tags, releases, commits, or PRs.
- [x] Confirm tests publish no packages: no `npm publish`, no registry calls, no tap pushes.

## Phase 5: Documentation

- [x] Add `docs/install.md` covering all installation options and marking each as available or future-only.
- [x] Document manual GitHub Release install: download the archive for the OS/arch, verify the checksum against the release checksum file, extract, and place the binary on `PATH`.
- [x] Document `install.sh` usage including the curl one-liner, version pinning, install directory override, and the no-sudo default.
- [x] Document npm global install with `npm install -g specharbor`, postinstall download behavior, the `--ignore-scripts` first-run fallback, and unsupported-platform behavior.
- [x] Document Homebrew install with `brew install guferreira1/tap/specharbor` and the tap path `guferreira1/homebrew-tap`.
- [x] Document install verification with `specharbor version`, including expected release metadata versus `dev`/`unknown` source-build output.
- [x] Document checksum verification commands for manual installs (`sha256sum -c` and `shasum -a 256 -c` guidance).
- [x] Document troubleshooting for `PATH` issues with user-local install directories.
- [x] Document the manual `go install` fallback and its expected development fallback metadata.
- [x] Document as future-only: Linux `.deb`/`.rpm` packages, Scoop, Winget, signing, SBOM, Docker images, and auto-update.
- [x] Update `README.md` install section and docs list, and add a pointer from `docs/release.md`.
- [x] Do not document any channel as available before its implementation stage and a published release exist.

## Phase 6: Verification

- [x] Run `gofmt` on any changed Go files (only the static boundary test changed; it is gofmt-clean).
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate implement-install-channels`.
- [x] Inspect `git status --short` for unexpected release automation, package-publishing, Linux/Windows packaging, or generated artifact files.
- [x] Inspect `git diff` over `install.sh`, the npm package directory, `README.md`, `docs/`, and `openspec/changes/implement-install-channels/`.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
- [x] Leave any task unchecked if the work was not performed or verification could not run.
