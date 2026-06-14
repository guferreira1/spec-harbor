# specharbor

`specharbor` is the npm distribution for the SpecHarbor CLI.

Português: [README.pt-BR.md](README.pt-BR.md)

This package installs and runs the official SpecHarbor binary for your platform. It is not just a command launcher: it gives you the same OpenSpec/SDD workflow entrypoint used in the repository:

- explicit change scaffolding
- local validation
- role prompts
- review and archive flows
- installable CLI workflow commands for AI-assisted development

## Install

```bash
npm install -g specharbor
specharbor version
```

You can also run without a global install:

```bash
npx specharbor version
```

## What SpecHarbor does

SpecHarbor is a Go CLI for OpenSpec-based AI coding-agent workflows. It helps you move from a scoped OpenSpec change to implementation, validation, review, and archive with explicit handoffs.

## Common commands

- `specharbor init`
- `specharbor workflow`
- `specharbor generate ...`
- `specharbor validate ...`
- `specharbor prompt ...`
- `specharbor review ...`
- `specharbor archive ...`
- `specharbor brief`
- `specharbor context discover`
- `specharbor context index`

## How the npm wrapper works

This package is a small Node launcher for the native binary.

- The package version `X.Y.Z` maps to exactly one GitHub Release tag `vX.Y.Z`.
- On install (or first run when scripts are skipped), it downloads the matching prebuilt asset from GitHub Releases.
- It verifies SHA-256 against `checksums.txt` before extracting.
- The extracted binary is stored in the package `native/` directory.
- All CLI arguments are forwarded as an array to the binary process with inherited stdio.
- The binary exit code is preserved.
- No shell command-string construction is used for forwarding.

## Supported platforms

| `process.platform` | `process.arch` | Release asset |
| --- | --- | --- |
| `linux` | `x64` | `specharbor_Linux_x86_64.tar.gz` |
| `linux` | `arm64` | `specharbor_Linux_arm64.tar.gz` |
| `darwin` | `x64` | `specharbor_Darwin_x86_64.tar.gz` |
| `darwin` | `arm64` | `specharbor_Darwin_arm64.tar.gz` |
| `win32` | `x64` | `specharbor_Windows_x86_64.zip` |
| `win32` | `arm64` | `specharbor_Windows_arm64.zip` |

Unsupported platforms fail with a clear, deterministic message.

## Version mapping

| npm version | GitHub Release tag |
| --- | --- |
| `0.2.0` | `v0.2.0` |

The mapping is one-to-one and fixed by package version.

## Troubleshooting

- **install scripts skipped**
  `npm install --ignore-scripts -g specharbor` skips postinstall by design.
  The first run (`specharbor version` or `npx specharbor version`) still performs the same checksum-verified download.

- **offline/proxy/GitHub access problems**
  First-run download failures can happen when outbound network access is restricted.
  Retry with normal network access, or use a manual install channel from [docs/install.md](https://github.com/guferreira1/spec-harbor/blob/main/docs/install.md).

- **unsupported platform or architecture**
  Only Linux/macOS/Windows on `x64` and `arm64` are supported.
  The error message points to manual install options.

- **checksum verification failure**
  The package refuses installation/boot with a mismatch. Retry the install command, then open an issue with the package version and error output.

- **checking version**
  Use:

```bash
specharbor version
```

Release binaries print injected metadata. Source builds without release metadata may print `dev`/`unknown`.

## Security model

- HTTPS-only release source URLs under
  `https://github.com/guferreira1/spec-harbor/releases/download/`
- Mandatory SHA-256 verification against release `checksums.txt`.
- No token-based access and no runtime authentication step.
- No shell command execution for argument forwarding.
- No writable paths outside the package directory except the package-native binary path (`native/`) and temporary runtime directories.

## Related documentation

- [Install and verify options](https://github.com/guferreira1/spec-harbor/blob/main/docs/install.md)
- [OpenSpec product docs](https://github.com/guferreira1/spec-harbor/blob/main/README.md)
- [Release metadata](https://github.com/guferreira1/spec-harbor/blob/main/docs/release.md)

## Tests

```bash
npm test
```

Tests use offline fixtures and do not publish artifacts.
