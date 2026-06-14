# Installing SpecHarbor

SpecHarbor is distributed through official GitHub Release assets built by
GoReleaser for `guferreira1/spec-harbor`. Every install channel downloads
those assets over HTTPS and verifies SHA-256 checksums before installing.

## Channel status

| Channel | Status |
| --- | --- |
| GitHub Releases | Prepared for `v0.2.0` after the tag is published |
| `install.sh` | Available for Linux and macOS using real release assets |
| npm | Prepared as unscoped package `specharbor@0.2.0` after the tag is published |
| Homebrew | Available as `brew install guferreira1/tap/specharbor` |
| `go install` from source | Fallback/developer option; prints development fallback metadata |
| package publishing automation | Automated on tag push for npm and Homebrew |
| Linux `.deb` / `.rpm` packages | Future only |
| Windows Scoop / Winget | Future only |
| signing | Future only |
| SBOM | Future only |
| Docker | Future only |

Binary install channels require a published GitHub Release with matching
assets and checksums. The next public release is prepared as `v0.2.0`.

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

Release assets cover Linux `amd64`, Linux `arm64`, macOS `amd64`, macOS
`arm64`, Windows `amd64`, and Windows `arm64`. `install.sh` supports Linux
and macOS. The npm wrapper supports Linux, macOS, and Windows on `x64` and
`arm64`. The current Homebrew formula is macOS-only.

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
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
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
SPECHARBOR_VERSION=v0.2.0 sh install.sh
sh install.sh --version v0.2.0

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
`specharbor`. Version bumps are published automatically when a `vX.Y.Z` tag is
pushed, after the GitHub Release assets exist; see
[Release metadata](release.md) for the workflow and required secrets.

Install it with:

```bash
npm install -g specharbor
specharbor version
```

You can also run the published package without a global install:

```bash
npx specharbor version
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
GitHub Actions macOS runners. The formula is updated automatically on each
`vX.Y.Z` tag release by the `homebrew-publish` job, which renders
`Formula/specharbor.rb` from the release `checksums.txt`; see
[Release metadata](release.md). No Homebrew formula files are committed to this
repository.

## Future-only channels

The following are planned but intentionally not implemented yet:

- Linux native packages (`.deb` and `.rpm`), likely via nfpm in a later
  change.
- Windows package managers: Scoop and Winget.
- Binary signing (for example cosign), SBOM generation, Docker images, and
  auto-update mechanisms.
- Package publishing automation for Linux packages, Windows package managers,
  signing, SBOMs, and Docker.

npm and Homebrew publishing are automated on tag releases; see
[Release metadata](release.md). Interim paths: Linux users have `install.sh`,
npm, and manual install; Windows users have npm and manual `.zip` install.

## Verifying an installation

```bash
specharbor version
```

A release binary prints injected release metadata:

```text
SpecHarbor 0.2.0
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

## Troubleshooting

### `specharbor: command not found`

User-local install directories such as `$HOME/.local/bin` may not be on your
`PATH`. Add the directory to your shell profile:

```bash
# bash: ~/.bashrc — zsh: ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"
```

Then restart the shell or `source` the profile. Verify with:

```bash
command -v specharbor
```

If a shell still resolves an old binary after replacing it, clear the command
cache and retry:

```bash
hash -r
specharbor version
```

### Permission denied

`install.sh` never invokes `sudo`. If the install directory is not writable,
choose a user-writable directory:

```bash
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh -s -- --install-dir "$HOME/bin"
```

For manual installs, create the target directory yourself and use a directory
already on `PATH`.

### Checksum mismatch

Do not install an archive whose checksum does not verify. Delete the partial
download, fetch the archive and `checksums.txt` again from the official
GitHub Release, and retry. If the mismatch persists, treat the artifact as
untrusted and open an issue with the release URL and checksum output.

### Unsupported platform or architecture

Release assets exist for Linux, macOS, and Windows on `amd64`/`x64` and
`arm64`. `install.sh` supports only Linux and macOS. For unsupported systems,
use a supported machine, try the npm package if Node supports your
OS/architecture pair, or build from source with Go.

### npm postinstall skipped or first-run download failed

`npm install --ignore-scripts -g specharbor` skips the postinstall download.
That is supported: `npx specharbor version` or `specharbor version` performs
the same checksum-verified download on first run. First-run download failures
usually indicate offline use, proxy restrictions, GitHub access restrictions,
or an unsupported platform. Re-run with network access to GitHub Releases, or
use a manual GitHub Release install.

### Homebrew tap/install issues

Use the tap shorthand exactly:

```bash
brew install guferreira1/tap/specharbor
```

If Homebrew cannot find or update the formula, refresh taps and retry:

```bash
brew update
brew untap guferreira1/tap
brew tap guferreira1/tap
brew install guferreira1/tap/specharbor
```

The external tap repository is `guferreira1/homebrew-tap`; no formula files
live in this repository.

### Version metadata looks unexpected

Verify the installed binary with:

```bash
specharbor version
```

Release binaries for `v0.2.0` print `SpecHarbor 0.2.0` and include commit,
date, and dirty metadata. `dev`/`unknown` output usually means the binary was
built from source without injected release metadata, such as with plain
`go install`.

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
- npm and Homebrew publishing are automated on `vX.Y.Z` tag releases through a
  tag-only workflow that validates version consistency and runs package tests
  before publishing; see [Release metadata](release.md).
