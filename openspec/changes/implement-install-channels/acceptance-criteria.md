# Acceptance Criteria: Implement Install Channels

## Dependency Gate

- [ ] Implementation work starts only after `implement-goreleaser-release` is merged and its release asset names, archive formats, checksum file name and format, and in-archive binary name are finalized.
- [ ] Every install channel consumes official GitHub Release assets published for `guferreira1/spec-harbor`.
- [ ] No channel builds from source, except the explicitly documented manual `go install` fallback, which is documented to produce development fallback metadata.
- [ ] Asset-name assumptions in this change are reconciled against the finalized release output before implementation, and this change is updated first when they differ.

## install.sh

- [ ] `curl -sSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh` installs the latest release on supported Linux and macOS platforms.
- [ ] The script detects OS and architecture, mapping `x86_64`/`amd64 -> x86_64` and `aarch64`/`arm64 -> arm64` (the finalized asset arch names), and fails clearly on unsupported platforms with a pointer to manual install docs.
- [ ] The script resolves the latest release by default and accepts an explicit version, validating the version string before using it in URLs.
- [ ] Downloads use HTTPS only and are restricted to `https://github.com/guferreira1/spec-harbor/` endpoints.
- [ ] The archive `sha256` is verified against the official checksum file before any file reaches the install target; missing checksum tooling causes failure, never skipped verification.
- [ ] The default install target is a user-local directory (preferring `$HOME/.local/bin`), overridable via `SPECHARBOR_INSTALL_DIR`.
- [ ] The script never invokes `sudo`; unwritable targets produce guidance instead of privilege escalation.
- [ ] A missing-`PATH` install directory produces a warning with copy-pasteable guidance, not a failure.
- [ ] Any failure leaves no partial binary in the install target and removes the temporary directory.
- [ ] The script executes no downloaded content, uses no `eval` over remote or user-controlled data, performs no package-manager side effects, and requires no tokens.
- [ ] If dry-run mode is implemented, it prints resolved OS, arch, version, asset URL, and install target without downloading or writing, and is covered by tests.

## npm Package

- [ ] The wrapper package is named `specharbor`, unscoped, with no npm org and no GoReleaser Pro dependency; a scoped package is the fallback only if the name becomes unavailable at publish time.
- [ ] `package.json` exposes the `specharbor` CLI command through `bin` and pins the wrapped release version to exactly one GitHub Release tag.
- [ ] The chosen first strategy is postinstall download with mandatory checksum verification, plus a first-run verified-download fallback when postinstall was skipped.
- [ ] OS/arch detection from `process.platform`/`process.arch` selects the correct release asset; unsupported platforms fail with a clear, documented error at postinstall and at run time.
- [ ] Downloads use HTTPS only, restricted to `https://github.com/guferreira1/spec-harbor/releases/download/` URLs, with `sha256` verified via Node `crypto` before the binary is used.
- [ ] The launcher forwards all arguments, stdio, and the exit code to the native binary using array-argument process APIs, with no shell-string command construction.
- [ ] The wrapper writes only inside its package or cache directory, requires no tokens, sends no telemetry, and mutates no Git state.
- [ ] No npm publishing occurs in this change or its tests.

## Homebrew Tap

- [ ] The documented tap is `guferreira1/homebrew-tap`, the formula is `specharbor`, and the documented install command is `brew install guferreira1/tap/specharbor`.
- [ ] The specified formula references an official pinned GitHub Release asset with a mandatory `sha256`, installs the prebuilt binary without building from source, and includes a `test do` block running `specharbor version`.
- [ ] It is documented that formula delivery may occur through a separate repository/PR and that `brew audit`/`brew install`/`brew test` validation can run later in a GitHub Actions macOS runner.
- [ ] No Homebrew formula or tap files are committed into this repository by this change.

## Linux and Windows Channels

- [ ] Linux `.deb` and `.rpm` packages are documented as future-only; no `nfpm.yaml`, `.nfpm.yaml`, `packaging/`, `debian/`, or `rpm/` files are introduced.
- [ ] Scoop and Winget are documented as future-only; no `scoop/` or `winget/` manifests are introduced.
- [ ] Documented interim paths exist for Linux (install.sh, npm, manual) and Windows (npm, manual `.zip` install).

## Security

- [ ] All channel downloads use HTTPS and verify `sha256` checksums against the official release checksum file.
- [ ] No channel executes downloaded scripts from untrusted URLs or runs arbitrary commands.
- [ ] No channel requires tokens or authenticated requests.
- [ ] No packages are published automatically; publishing is an explicit, manual, later step.
- [ ] No channel modifies SpecHarbor source code or writes outside its install target and cache/temporary directories.
- [ ] No channel, script, or test mutates the Git repository or creates tags, releases, commits, or PRs.

## Documentation

- [ ] `docs/install.md` exists and covers: all install options with available/future-only status, manual GitHub Release install, install.sh, npm global install, Homebrew tap install, `specharbor version` verification, checksum verification guidance, `PATH` troubleshooting, the `go install` fallback, and the future-only channel list.
- [ ] `README.md` links the install documentation, and `docs/release.md` points to it.
- [ ] No channel is documented as available before its implementation stage and a published release exist.

## Tests and Regression Safety

- [ ] install.sh tests cover OS/arch detection (including the `x86_64` asset arch mapping), asset URL construction, checksum verification success and failure, install target selection, absence of sudo, dry-run output if implemented, and clean failure with no partial files.
- [ ] npm tests cover package name/config, `bin` exposure, OS/arch selection, checksum verification, argument and exit-code forwarding, unsupported-platform errors, absence of shell injection, and rejection of non-allowlisted URLs.
- [ ] Homebrew expectations (asset URL, `sha256`, binary install, `test do` running `specharbor version`, documented tap path) are recorded as testable criteria for the tap repository.
- [ ] Channel tests run against local fixtures or stub servers; CI performs no real release downloads.
- [ ] `go test ./...` passes with no changes to CLI runtime behavior, `internal/`, or `cmd/`.
- [ ] This change generates no release assets and introduces no GoReleaser config or release workflow files.
- [ ] Tests perform no source-control automation and publish no packages.

## Validation

- [ ] `go run ./cmd/specharbor validate implement-install-channels` reports the change as valid.
