# Installing SpecHarbor

SpecHarbor is distributed through official GitHub Release assets built by
GoReleaser for `guferreira1/spec-harbor`. Every install channel downloads
those assets over HTTPS and verifies SHA-256 checksums before installing.

## Channel status

| Channel | Status |
| --- | --- |
| Manual GitHub Release download | Available |
| `install.sh` (Linux, macOS) | Available; installs published releases |
| npm global package (`specharbor`) | Available as `specharbor@0.1.0` |
| Homebrew tap (`guferreira1/tap/specharbor`) | Available through `guferreira1/homebrew-tap` |
| Linux `.deb` / `.rpm` packages | Future only |
| Windows Scoop / Winget | Future only |
| `go install` from source | Available (development fallback metadata) |

Binary install channels require a published GitHub Release with matching
assets and checksums. The first published release is `v0.1.0`.

## Release assets

Each release tag `vX.Y.Z` publishes these assets:

```text
specharbor_Linux_x86_64.tar.gz
specharbor_Linux_arm64.tar.gz
specharbor_Darwin_x86_64.tar.gz
specharbor_Darwin_arm64.tar.gz
specharbor_Windows_x86_64.zip
specharbor_Windows_arm64.zip
checksums.txt
```

Archives contain the `specharbor` binary (`specharbor.exe` on Windows).
`checksums.txt` contains one SHA-256 line per asset. Asset download URLs
follow this pattern:

```text
https://github.com/guferreira1/spec-harbor/releases/download/vX.Y.Z/<asset>
```

## Manual install from GitHub Releases

1. Download the archive for your OS/arch and `checksums.txt` from the
   [releases page](https://github.com/guferreira1/spec-harbor/releases).

2. Verify the checksum. On Linux:

   ```bash
   grep specharbor_Linux_x86_64.tar.gz checksums.txt | sha256sum -c -
   ```

   On macOS:

   ```bash
   grep specharbor_Darwin_arm64.tar.gz checksums.txt | shasum -a 256 -c -
   ```

   Do not install an archive whose checksum does not verify.

3. Extract and place the binary on your `PATH`:

   ```bash
   tar -xzf specharbor_Linux_x86_64.tar.gz specharbor
   mkdir -p "$HOME/.local/bin"
   install -m 0755 specharbor "$HOME/.local/bin/specharbor"
   ```

   On Windows, extract `specharbor.exe` from the `.zip` and place it in a
   directory on your `PATH`.

4. Verify the install:

   ```bash
   specharbor version
   ```

## install.sh (Linux and macOS)

The repository root contains a POSIX `sh` install script that automates the
manual flow: it detects OS and architecture, resolves the latest release (or
a pinned version), downloads the matching archive and `checksums.txt` over
HTTPS, verifies the SHA-256 checksum, and installs the binary to a user-local
directory.

```bash
curl -sSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
```

If you prefer to review the script before running it (recommended):

```bash
curl -sSLO https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh
less install.sh
sh install.sh
```

Options:

```bash
# Pin a version (environment variable or flag):
SPECHARBOR_VERSION=v0.1.0 sh install.sh
sh install.sh --version v0.1.0

# Override the install directory (default: $HOME/.local/bin):
SPECHARBOR_INSTALL_DIR="$HOME/bin" sh install.sh
sh install.sh --install-dir "$HOME/bin"

# Dry run: print resolved OS, arch, version, asset URL, and install target
# without downloading the archive or writing anything:
sh install.sh --dry-run
```

Behavior and guarantees:

- Supports Linux and macOS (Darwin) on `x86_64`/`amd64` and
  `aarch64`/`arm64`. Other platforms fail with a clear error pointing here.
- Version strings are strictly validated (`X.Y.Z` or `vX.Y.Z`) before being
  used in URLs.
- Downloads are HTTPS-only and restricted to
  `https://github.com/guferreira1/spec-harbor/releases/` URLs.
- SHA-256 verification against `checksums.txt` is mandatory. If no
  `sha256sum` or `shasum` tool is available, the script fails instead of
  skipping verification.
- Never invokes `sudo`. If the install directory is not writable, the script
  fails with guidance to pick a user-local directory.
- Never executes downloaded content and writes only to its temporary
  directory and the install target. Failures clean up temporary files and
  leave no partial binary in the install target.

Test the script offline with:

```bash
sh scripts/test-install-sh.sh
```

## npm global package

The npm wrapper package lives in this repository at
`packages/npm/specharbor/` and is published to the npm registry as
`specharbor@0.1.0`. Publishing version bumps remains a manual, explicit
maintainer step that happens outside this repository's automation.

Install it with:

```bash
npm install -g specharbor
specharbor version
```

How the wrapper works:

- The package version `X.Y.Z` pins exactly one release tag `vX.Y.Z`.
- At `postinstall`, the wrapper detects `process.platform`/`process.arch`,
  downloads the matching release asset and `checksums.txt` over HTTPS from
  `https://github.com/guferreira1/spec-harbor/releases/download/` only,
  verifies the SHA-256 checksum with Node's `crypto`, and extracts the binary
  into the package's own `native/` directory.
- Installs with `--ignore-scripts` skip postinstall; the launcher then
  performs the same checksum-verified download on first run.
- The launcher forwards arguments and stdio to the native binary with
  array-argument process APIs (no shell strings) and preserves the exit code.
- Unsupported platforms fail with a deterministic error naming the platform
  and pointing to this document, both at postinstall and at run time.

See [packages/npm/specharbor/README.md](../packages/npm/specharbor/README.md)
for details and `npm test` for its offline test suite.

## Homebrew tap

Homebrew support is available through the personal external tap repository
`guferreira1/homebrew-tap`, with formula name `specharbor`. Install it with:

```bash
brew install guferreira1/tap/specharbor
```

The formula satisfies these expectations in the tap repository:

- `url` points at an official pinned GitHub Release asset.
- A `sha256` value, copied from the release `checksums.txt`, is mandatory for
  every referenced asset.
- The formula installs the prebuilt binary; it does not build from source.
- The formula `test do` block runs `specharbor version` and asserts the
  output contains the expected version.

The tap validates `brew audit --strict --online specharbor`, formula install,
`specharbor version`, `brew test specharbor`, and the user install command on
GitHub Actions macOS runners. Formula version bumps are manual until later
automation is specified. No Homebrew formula files are committed to this
repository.

## Future-only channels

The following are planned but intentionally not implemented yet:

- Linux native packages (`.deb` and `.rpm`), likely via nfpm in a later
  change.
- Windows package managers: Scoop and Winget.
- Binary signing (for example cosign), SBOM generation, Docker images, and
  auto-update mechanisms.

Interim paths: Linux users have `install.sh`, npm, and manual install;
Windows users have npm and manual `.zip` install.

## Verifying an installation

```bash
specharbor version
```

A release binary prints injected release metadata:

```text
SpecHarbor 0.1.0
commit: <full commit sha>
date: <UTC RFC3339 build date>
dirty: false
```

A source build without injected metadata prints the development fallback:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

See [Release metadata](release.md) for the full version convention.

## PATH troubleshooting

User-local install directories such as `$HOME/.local/bin` may not be on your
`PATH`. If `specharbor: command not found` appears after a successful
install, add the directory to your shell profile:

```bash
# bash: ~/.bashrc — zsh: ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"
```

Then restart the shell or `source` the profile. Verify with:

```bash
command -v specharbor
```

## go install fallback

Building from source is the documented manual fallback and requires Go:

```bash
go install github.com/guferreira1/spec-harbor/cmd/specharbor@latest
```

Pin a tag with `@vX.Y.Z` once releases exist. Binaries built this way use the
development fallback metadata (`dev`/`unknown`) because no release metadata
is injected; that output is expected and documented in
[Release metadata](release.md).

## Security model

All install channels follow the same rules:

- HTTPS only; downloads are restricted to official
  `https://github.com/guferreira1/spec-harbor/releases/` URLs.
- SHA-256 checksum verification against the release `checksums.txt` is
  mandatory for every downloaded archive; verification failure aborts the
  install and removes partial files.
- Checksums protect against corruption and tampering in transit; they are
  served from the same release as the assets, so they do not protect against
  a fully compromised release. Signing is future work.
- No channel executes downloaded scripts, builds shell command strings from
  user input, requires tokens or authentication, sends telemetry, mutates Git
  state, or writes outside its install target and cache/temporary
  directories.
- Package publishing and formula version bumps are explicit maintainer steps
  unless a later OpenSpec change defines automation.
