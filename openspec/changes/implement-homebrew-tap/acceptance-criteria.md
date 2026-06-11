# Acceptance Criteria: Implement Homebrew Tap

## Tap and Formula

- The Homebrew tap repository exists at `guferreira1/homebrew-tap`.
- The formula exists at `Formula/specharbor.rb` in the external tap repository.
- The formula installs SpecHarbor from official GitHub Release assets under `guferreira1/spec-harbor`.
- The formula initially pins SpecHarbor version `0.1.0`.
- The formula uses verified SHA-256 values from the published release `checksums.txt`.
- The formula supports macOS `amd64`/`x86_64` and `arm64` assets.
- The formula installs the `specharbor` binary into `bin`.
- The installed command is available as `specharbor`.

## User-Visible Behavior

- `brew install guferreira1/tap/specharbor` succeeds.
- `specharbor version` reports `0.1.0`.
- `brew test specharbor` succeeds.
- Formula tests do not require network access after installation.

## Validation

- GitHub Actions validation in the external tap repository passes.
- Validation covers formula asset URLs, SHA-256 values, macOS architecture mapping, formula install behavior, `specharbor version`, and `brew test specharbor`.
- Any limitation in validating both Intel and Apple Silicon runners is documented.

## Main Repository Boundary

- Main repository docs mention Homebrew as available only after the external tap has been validated.
- No new SpecHarbor release tag is created.
- No GitHub Release is modified.
- No npm package is published or modified.
- No GoReleaser config is modified unless explicitly approved in a later change.
- No `install.sh` behavior is modified.
- No Linux package manager support is added in this change.
- No Windows package manager support is added in this change.
- No signing, SBOM generation, or package publishing automation is added in this change.
- No production behavior changes are introduced in the main SpecHarbor repository.
