# specharbor (npm wrapper)

npm wrapper package for the [SpecHarbor](https://github.com/guferreira1/spec-harbor) CLI.

> **Status:** this package is maintained in the SpecHarbor repository and has
> not been published to the npm registry yet. `npm install -g specharbor`
> works only after a maintainer publishes the package manually. Publishing is
> an explicit manual step and is never automated. See
> [docs/install.md](https://github.com/guferreira1/spec-harbor/blob/main/docs/install.md)
> for currently available installation options.

## What it does

- Exposes the `specharbor` CLI command through a small Node launcher.
- Downloads the official prebuilt `specharbor` binary from GitHub Releases at
  `postinstall` time, verifies its SHA-256 checksum against the release
  `checksums.txt`, and stores it inside the package's own `native/` directory.
- Falls back to the same checksum-verified download on first run when
  postinstall scripts were skipped (for example `npm install --ignore-scripts`).
- Forwards all CLI arguments and stdio to the native binary with
  array-argument process APIs and preserves the binary's exit code. No shell
  command strings are ever constructed.

## Version mapping

The npm package version `X.Y.Z` maps to exactly one GitHub Release tag
`vX.Y.Z`. The pinned release is the single source of the downloaded binary.
At publish time the package version must match a published SpecHarbor
release tag.

## Supported platforms

| `process.platform` | `process.arch` | Release asset |
| --- | --- | --- |
| `linux` | `x64` | `specharbor_Linux_x86_64.tar.gz` |
| `linux` | `arm64` | `specharbor_Linux_arm64.tar.gz` |
| `darwin` | `x64` | `specharbor_Darwin_x86_64.tar.gz` |
| `darwin` | `arm64` | `specharbor_Darwin_arm64.tar.gz` |
| `win32` | `x64` | `specharbor_Windows_x86_64.zip` |
| `win32` | `arm64` | `specharbor_Windows_arm64.zip` |

Unsupported platforms fail with a deterministic error naming the detected
platform and pointing to the manual installation documentation, both at
postinstall and at run time.

## Security model

- Downloads use HTTPS only and are restricted to
  `https://github.com/guferreira1/spec-harbor/releases/download/` URLs;
  redirects must stay on HTTPS GitHub hosts.
- SHA-256 checksum verification against the release `checksums.txt` is
  mandatory; verification failure aborts the install and removes partial
  files.
- No tokens, no authenticated requests, no telemetry, no Git mutation, and no
  writes outside the package's own directory.

## Tests

```bash
npm test
```

Tests run fully offline against in-memory fixtures and an injected transport;
they never download real release assets and never publish anything.
