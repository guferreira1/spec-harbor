# specharbor (npm wrapper)

npm wrapper package for the [SpecHarbor](https://github.com/guferreira1/spec-harbor) CLI.

> **Status:** `specharbor@0.1.0` is published to the npm registry and maps to
> GitHub Release `v0.1.0`. Publishing version bumps remains an explicit
> manual maintainer step and is not automated.

Install globally:

```bash
npm install -g specharbor
specharbor version
```

Run without a global install:

```bash
npx specharbor version
```

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
`vX.Y.Z`. Package version `0.1.0` resolves to GitHub Release `v0.1.0`, which
is the single source of the downloaded binary. At publish time the package
version must match a published SpecHarbor release tag.

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

## Troubleshooting

- `npm install --ignore-scripts -g specharbor` skips postinstall by design.
  Run `specharbor version` or `npx specharbor version` later to trigger the
  same checksum-verified binary download.
- First-run download failures usually indicate offline use, proxy
  restrictions, GitHub Release access restrictions, or an unsupported
  platform/architecture pair. Retry with network access to GitHub Releases or
  use the manual install guide.
- If version output shows `SpecHarbor 0.1.0`, the wrapper resolved the
  `v0.1.0` release binary successfully.

See
[docs/install.md](https://github.com/guferreira1/spec-harbor/blob/main/docs/install.md)
for the full install guide.

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
