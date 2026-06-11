# Design: Polish Release Documentation

## Overview

This is a documentation-only change. It should consolidate public release documentation after the real `v0.1.0` release, npm publication, and external Homebrew tap validation. The implementation should edit public docs only and must not change CLI behavior, release assets, package metadata, release automation, installer behavior, or external tap files.

## 1. Public Install Documentation

Install documentation should have one clear hierarchy:

- `README.md` provides a concise install summary and links to `docs/install.md` for details.
- `docs/install.md` is the authoritative install guide, including manual GitHub Release install, `install.sh`, npm, Homebrew, checksum verification, version verification, channel status, and troubleshooting.
- `docs/release.md` explains release metadata, assets, and release process boundaries, then points users to `docs/install.md` for install steps.
- `docs/usage.md` should only be updated if its version or install guidance becomes inconsistent with the install and release docs.
- `packages/npm/specharbor/README.md` should be updated if present and stale, especially if it still says the package is unpublished.

The docs should include these clear install paths:

```bash
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
npm install -g specharbor
brew install guferreira1/tap/specharbor
```

Manual GitHub Release download should remain documented with checksum verification. If `go install` is documented, it must be presented as a fallback or developer/source-build option, not the main release channel unless existing project documentation already says otherwise. `go install` builds are expected to show development fallback metadata unless release metadata is injected manually.

## 2. Version Verification Documentation

Documentation should consistently tell users to verify installation with:

```bash
specharbor version
```

For the current public release, the expected version line is:

```text
SpecHarbor 0.1.0
```

Docs should explain the tag/version distinction:

- Git tags and GitHub Releases use `v0.1.0`.
- Release binary metadata displays plain `0.1.0`.

Release binaries include metadata fields for version, commit, date, and dirty state. Public examples should describe the shape without hardcoding volatile build dates:

```text
SpecHarbor 0.1.0
commit: <full commit sha>
date: <UTC RFC3339 build date>
dirty: false
```

The expected release commit for `v0.1.0` is `e6faff91feef07e5c1e47181243286268daf17b5`.

## 3. Release Asset Documentation

Release documentation should list or describe the supported public assets for each release:

```text
Linux amd64
Linux arm64
macOS amd64
macOS arm64
Windows amd64
Windows arm64
checksums.txt
```

Where filenames are useful, docs should use the GoReleaser asset names:

```text
specharbor_Linux_x86_64.tar.gz
specharbor_Linux_arm64.tar.gz
specharbor_Darwin_x86_64.tar.gz
specharbor_Darwin_arm64.tar.gz
specharbor_Windows_x86_64.zip
specharbor_Windows_arm64.zip
checksums.txt
```

Docs should describe checksum verification where appropriate:

- manual GitHub Release downloads should instruct users to verify the archive against `checksums.txt` before installing;
- `install.sh`, npm, and Homebrew should state that they verify checksums or use a pinned SHA-256 value as part of their channel behavior;
- checksum mismatch guidance should tell users not to install the artifact and to retry from the official release source.

## 4. Channel Status Matrix

Public docs should include one consistent distribution-channel status table or section. It may live primarily in `docs/install.md`, with README using a shorter summary.

Expected available channels:

| Channel | Status |
| --- | --- |
| GitHub Releases | Available for `v0.1.0` |
| `install.sh` | Available for Linux and macOS using real release assets |
| npm | Available as unscoped package `specharbor@0.1.0` |
| Homebrew | Available as `brew install guferreira1/tap/specharbor` |

Expected future-only channels:

| Channel | Status |
| --- | --- |
| `.deb` / `.rpm` | Future only |
| Scoop | Future only |
| Winget | Future only |
| signing | Future only |
| SBOM | Future only |
| Docker | Future only |
| publishing automation | Future only |

The docs must not present future-only channels as available. `go install` may appear as a fallback or developer option, but it should not be grouped as the main release distribution channel unless existing project documentation already requires that framing.

## 5. Troubleshooting Documentation

Troubleshooting should be useful but not scattered. Prefer a consolidated section in `docs/install.md`, with README pointing there instead of duplicating details.

Topics to include or refine:

- `PATH` issues after installing to `$HOME/.local/bin` or another user-local directory;
- permission denied errors and user-writable install directories;
- checksum mismatch and the requirement to abort installation;
- unsupported platform or architecture errors;
- npm installs using `--ignore-scripts` and the first-run binary download fallback;
- npm first-run binary download failures, including offline/proxy/corporate network cases;
- Homebrew tap or formula install issues, including stale taps and `brew update`/retap guidance where appropriate;
- stale shell command cache after installing or replacing the binary;
- verifying version metadata and understanding `0.1.0` versus `dev`/`unknown` output.

## 6. Consistency and Duplication Cleanup

Implementation should reduce contradictions and stale release wording. The review should specifically check for:

- removing or updating future-work wording for npm and Homebrew now that those channels are published and validated;
- avoiding duplicated conflicting install commands across README and docs;
- using `curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh` consistently for the install script one-liner;
- keeping `specharbor version` as the standard verification command;
- using `v0.1.0` for the GitHub Release tag and `0.1.0` for binary version output;
- documenting the npm package as unscoped `specharbor`, not a scoped package;
- documenting the Homebrew command exactly as `brew install guferreira1/tap/specharbor`;
- ensuring `packages/npm/specharbor/README.md`, if updated, no longer says `specharbor` is unpublished.

## 7. Boundary With Automation

Documentation may mention future automation, but must not claim that package publishing is automated.

Manual or validated state should be clear:

- GitHub Release assets are created through the tag-triggered release workflow;
- npm was published manually for `0.1.0`;
- Homebrew tap formula maintenance is manual for now in external repository `guferreira1/homebrew-tap`;
- package publishing automation remains future work.

Docs must not imply that npm publishing, Homebrew formula bumping, Linux package publishing, Windows package-manager manifests, signing, SBOM generation, Docker image publishing, or release promotion are automated unless a later OpenSpec change implements and validates that behavior.

## Implementation Boundary

No architecture changes are required. The implementation should be limited to public documentation files in the approved scope. It must not change production Go code, `install.sh` behavior, `.goreleaser.yaml`, `.github/workflows/release.yml`, npm package metadata/runtime files, or Homebrew tap files. If the npm package README is updated, that should be a README-only documentation change unless explicitly justified during review.
