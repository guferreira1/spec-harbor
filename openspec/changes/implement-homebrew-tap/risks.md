# Risks: Implement Homebrew Tap

## Risks

- The Homebrew formula may use the wrong GitHub Release asset URL.
- A SHA-256 value may not match the published release asset.
- The formula may work on one macOS architecture but fail on another.
- GitHub Actions macOS runners may not cover both Intel and Apple Silicon validation paths.
- Homebrew audit rules may reject the formula style or binary-formula structure.
- External tap repository changes may drift from the main repository plan.
- Documentation may claim Homebrew support before the external tap is validated.
- Future SpecHarbor releases will require a formula update process.
- Package publishing automation is intentionally out of scope, so formula updates may remain manual at first.

## Mitigations

- Copy asset URLs from the published `v0.1.0` GitHub Release and review the formula against the expected Darwin asset mapping.
- Copy SHA-256 values directly from the validated release `checksums.txt` and verify them in tap CI.
- Test the formula on GitHub Actions macOS runners and document any architecture coverage gap.
- Run `brew audit --strict --online specharbor` when feasible before declaring the tap ready.
- Keep the production formula in `guferreira1/homebrew-tap` and keep main repository changes limited to OpenSpec planning and post-validation docs.
- Update main repository docs only after `brew install guferreira1/tap/specharbor`, `specharbor version`, and `brew test specharbor` succeed.
- Record the manual release-to-formula update steps until a later OpenSpec change defines automation.
- Treat GoReleaser `brews` config, signing, SBOM generation, and package publishing automation as separate future changes.
