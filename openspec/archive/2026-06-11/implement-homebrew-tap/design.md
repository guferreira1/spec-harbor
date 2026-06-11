# Design: Implement Homebrew Tap

## Overview

Homebrew support is a distribution-channel change, not a product behavior change. The formula should install the already published SpecHarbor release binary from official GitHub Release assets and validate the installed command with `specharbor version`.

No application core, CLI behavior, GoReleaser configuration, npm package metadata, or `install.sh` behavior should change as part of creating the tap.

## 1. External Tap Repository Boundary

The Homebrew tap lives in a separate repository:

```text
guferreira1/homebrew-tap
```

The main SpecHarbor repository should only contain OpenSpec planning in this change. After the tap is implemented and validated, the main repository may receive documentation updates if needed.

The formula should be committed to the external tap repository as:

```text
Formula/specharbor.rb
```

The formula itself should not be committed directly to the main SpecHarbor repository unless it is only a documentation example or generated reference. Any production tap behavior belongs to `guferreira1/homebrew-tap`.

## 2. Formula Source and Assets

The formula should download SpecHarbor from official GitHub Release asset URLs under:

```text
https://github.com/guferreira1/spec-harbor/releases/download/v0.1.0/
```

The initial formula must pin version `0.1.0` and use the published `v0.1.0` release assets. It must not download unverified binaries and should not build from source unless a later change explicitly justifies a source-build formula.

Expected macOS asset mapping:

```text
Darwin arm64       -> specharbor_Darwin_arm64.tar.gz
Darwin amd64/x86_64 -> specharbor_Darwin_x86_64.tar.gz
```

The formula should prefer macOS/Darwin assets for Homebrew on macOS and support both Apple Silicon and Intel macOS. SHA-256 values must be copied from the validated published release `checksums.txt` for every referenced asset.

If the formula uses architecture conditionals, each conditional must point to the matching macOS archive and matching SHA-256. If Homebrew formula conventions favor a different structure, the implementer should follow current Homebrew guidance while preserving the verified-asset and architecture-mapping requirements.

## 3. Homebrew Formula Behavior

The formula should:

- install the `specharbor` binary into Homebrew's `bin`;
- preserve executable permissions so `specharbor` runs after install;
- expose the command as `specharbor`;
- avoid postinstall downloads because the formula should install the release asset directly;
- avoid network access in formula tests after installation;
- include a `test do` block that runs `specharbor version`;
- optionally assert that command output contains `0.1.0`.

The test should validate the installed binary, not source code. It should not call GitHub, npm, GoReleaser, or any package publishing workflow.

## 4. Validation Strategy Without Local macOS

The maintainer does not have a local Mac environment. Validation should run in GitHub Actions macOS runners in the external tap repository.

Expected tap-repository checks:

- `brew audit --strict --online specharbor`, if feasible for the tap and current Homebrew policy;
- `brew install ./Formula/specharbor.rb`;
- `specharbor version`;
- `brew test specharbor`;
- validate macOS architecture handling as much as GitHub Actions allows.

`brew install --build-from-source` should only be used if a later design switches to building from source. For this binary formula, validation should focus on installing the formula from the local formula file, running the binary, and running the formula test.

If GitHub Actions availability covers only one macOS architecture for the external tap, the implementation notes should document that limitation and explain what was validated for the other architecture, such as URL and SHA-256 checks against `checksums.txt`.

## 5. Documentation Update Boundary

Main repository documentation should be updated only after the external tap works.

Potential documentation files:

- `README.md`;
- `docs/install.md`;
- `docs/release.md`.

After validation, docs may include:

```bash
brew install guferreira1/tap/specharbor
```

Docs should clearly state Homebrew availability only after the external tap is validated. Before validation, the main repository must not claim that Homebrew installation is available.

## 6. Release and Package Boundary

This change does not create or modify release artifacts. It uses the already published `v0.1.0` assets.

Out of scope:

- new release tags;
- GitHub Release changes;
- GoReleaser config changes;
- npm changes;
- `install.sh` changes;
- Linux `.deb` or `.rpm` packages;
- Scoop or Winget support;
- signing;
- SBOM generation;
- package publishing automation.

Future releases will need either a manual formula update process or a later automation change. That automation is intentionally outside this change.

## Architecture Impact

No hexagonal architecture changes are required. The formula is external distribution metadata and should not introduce new `internal/core`, `internal/adapters`, `internal/platform`, or `cmd` dependencies.
