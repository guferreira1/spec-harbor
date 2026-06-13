# Risks: Release Publishing Automation

## Accidental real publish during development

- **Risk:** Workflow or local testing could publish to npm, push a tag, or
  write to the Homebrew tap.
- **Mitigation:** Publishing is tag-only and gated behind validation and tests;
  local verification uses only `goreleaser check`, `goreleaser release
  --snapshot --clean`, `npm test`, and `npm pack --dry-run`, none of which
  publish. No tag is pushed and no token is configured during this change.

## Version mismatch between tag and package

- **Risk:** A tag `vX.Y.Z` could be pushed while `package.json` is a different
  version, producing an npm package that downloads the wrong release assets.
- **Mitigation:** `validate-release-inputs` fails the whole release before any
  publish when the tag and `package.json` versions disagree, and the npm job
  re-validates before publishing.

## Homebrew tap token and external repository access

- **Risk:** The default `GITHUB_TOKEN` cannot push to `guferreira1/homebrew-tap`,
  and a misscoped or leaked token could expose write access.
- **Mitigation:** Use a dedicated least-privilege `HOMEBREW_TAP_GITHUB_TOKEN`
  scoped to the tap repository, consumed by `actions/checkout` (which keeps the
  token off command lines) and never printed; GitHub Actions masks the secret
  in logs. The required external setup is documented in `docs/release.md`.

## GoReleaser Homebrew deprecation and formula compatibility

- **Risk:** GoReleaser `brews` is deprecated (failing `goreleaser check`) and
  the `homebrew_casks` replacement would convert the tap from a formula to a
  cask, breaking the existing manually maintained formula and the documented
  install command.
- **Mitigation:** Keep `.goreleaser.yaml` unchanged and render the formula in a
  dedicated job from a small, offline-tested script that preserves the existing
  macOS formula structure, so `brew install guferreira1/tap/specharbor` keeps
  working.

## npm trusted publishing configuration

- **Risk:** OIDC trusted publishing requires one-time npm-side configuration
  and a recent npm; without it, publishing fails.
- **Mitigation:** Document the trusted-publisher setup and a `NPM_TOKEN`
  fallback, ensure an up-to-date npm in CI, and require `id-token: write` only
  on the npm job. The package `specharbor` already exists on npm, so the
  trusted publisher can be configured against it.

## Duplicate npm version republish

- **Risk:** Re-running a release for an already-published version fails because
  npm forbids republishing the same version.
- **Mitigation:** Documented failure handling instructs maintainers to bump the
  version and tag rather than republish, and describes recovery when the GitHub
  Release succeeds but npm or Homebrew fails.

## Partial publish across channels

- **Risk:** GitHub Release could succeed while npm or Homebrew fails, leaving
  channels out of sync.
- **Mitigation:** Job ordering publishes npm only after assets exist and
  Homebrew within the GoReleaser run; `docs/release.md` documents per-channel
  rerun and manual-recovery steps so a partial failure can be completed safely.

## Asset-name or guardrail regressions

- **Risk:** Editing GoReleaser or the workflow could change asset names or
  weaken scope guardrails, breaking `install.sh`, the npm wrapper, or the
  formula, or silently enabling out-of-scope channels.
- **Mitigation:** Keep `builds`/`archives`/`checksum` byte-compatible, update
  the architecture guardrail test to allow only the approved new behavior while
  still forbidding out-of-scope channels, and verify with `go test ./...`,
  `goreleaser check`, and the snapshot build.

## Scope creep into future channels

- **Risk:** Adding npm and Homebrew automation invites adding Docker, signing,
  SBOM, or Linux/Windows packaging in the same change.
- **Mitigation:** The proposal and guardrail test explicitly keep those out of
  scope; the documentation continues to list them as future work and the tests
  fail if they are introduced.
