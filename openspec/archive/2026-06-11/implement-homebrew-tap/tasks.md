# Tasks: Implement Homebrew Tap

## Preparation

- [x] Verify worktree and branch state before making implementation changes.
- [x] Confirm the `v0.1.0` GitHub Release exists for `guferreira1/spec-harbor`.
- [x] Confirm Darwin/macOS assets exist for `amd64`/`x86_64` and `arm64`.
- [x] Confirm SHA-256 values for the macOS assets from the published `checksums.txt`.
- [x] Confirm `specharbor@0.1.0` npm publication remains untouched by this work.
- [x] Inspect current Homebrew formula conventions for binary formulas and architecture-specific assets.

## External Tap Repository

- [x] Create or prepare the external repository `guferreira1/homebrew-tap`.
- [x] Add `Formula/specharbor.rb` in the external tap repository.
- [x] Configure the formula to use official `v0.1.0` GitHub Release assets from `guferreira1/spec-harbor`.
- [x] Configure formula SHA-256 values for macOS assets using the verified release checksums.
- [x] Map Darwin `arm64` to `specharbor_Darwin_arm64.tar.gz`.
- [x] Map Darwin `amd64`/`x86_64` to `specharbor_Darwin_x86_64.tar.gz`.
- [x] Install the `specharbor` binary into Homebrew's `bin`.
- [x] Preserve executable permissions for the installed binary.
- [x] Add a formula test invoking `specharbor version`.
- [x] Assert the formula test output contains `0.1.0` if the assertion is compatible with Homebrew conventions.

## Tap Validation

- [x] Add GitHub Actions validation in the tap repository.
- [x] Validate formula asset URLs.
- [x] Validate SHA-256 values.
- [x] Validate macOS `amd64`/`x86_64` and `arm64` asset mapping.
- [x] Run `brew audit --strict --online specharbor` if feasible.
- [x] Run `brew install ./Formula/specharbor.rb` on a macOS runner.
- [x] Run `specharbor version` after installation.
- [x] Run `brew test specharbor`.
- [x] Validate both macOS architecture paths as much as GitHub Actions runner availability allows.
- [x] Document any limitation if only one runner architecture is available (not applicable; CI covered Intel and arm64 runners).
- [x] Confirm `brew install guferreira1/tap/specharbor` works.
- [x] Confirm `specharbor version` reports `0.1.0`.

## Main Repository Boundaries

- [x] Confirm no changes to GoReleaser release workflow.
- [x] Confirm no changes to `.goreleaser.yaml`.
- [x] Confirm no npm publish or npm metadata changes.
- [x] Confirm no new GitHub Release or tag is created.
- [x] Confirm no `install.sh` behavior changes.
- [x] Confirm no Linux or Windows package manager support is added.
- [x] Confirm no signing, SBOM generation, or package publishing automation is added.
- [x] Update main repository installation docs only after tap validation.
- [x] If main repository docs are changed, update only the relevant documentation files.
- [x] Run relevant repository validation/tests if main repository docs are changed.
- [x] Verify no unintended release, npm, install-channel, or production-code modifications are present.

## Review and Handoff

- [x] Inspect git status and diffs in all touched repositories.
- [x] Prepare PR for the main repository OpenSpec/docs changes if needed (not applicable; no SpecHarbor PR was opened).
- [x] Prepare PR or commit for the external tap repository according to its workflow.
- [x] Record validation commands and results in the implementation handoff.
