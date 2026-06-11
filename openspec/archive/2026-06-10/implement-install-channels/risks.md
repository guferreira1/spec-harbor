# Risks: Implement Install Channels

## Risks

### Dependency drift against implement-goreleaser-release

This change is authored in parallel with `implement-goreleaser-release`. The asset naming pattern, archive formats, checksum file name, and platform matrix referenced here are working assumptions. If the finalized release output differs, channel code written against these assumptions would download the wrong URLs or fail verification.

Mitigation: Phase 0 of `tasks.md` is a hard gate. Implementation must not start until the GoReleaser outputs are finalized, and every asset assumption must be reconciled against the merged spec and the first real release, updating this change first when they differ.

### npm package name availability

Maintainer checks (`npm view specharbor` returned 404; `npm search specharbor` returned no matches) suggest `specharbor` is available, but the name could be taken or blocked by npm policy before publishing.

Mitigation: re-verify availability at publish time. A scoped package is the documented fallback only if the unscoped name becomes unavailable. Publishing is manual and outside this change, so no automation breaks if the name changes.

### Postinstall download reliability

The chosen npm strategy downloads the binary at postinstall. Installs with `--ignore-scripts`, offline environments, registry mirrors, or corporate proxies can leave the package installed without a binary, or fail at install time.

Mitigation: the first-run fallback performs the same verified download when the binary is missing, and unsupported or offline situations produce a clear error pointing to manual install docs. Direct per-platform binary packages remain a documented future alternative if postinstall proves unreliable.

### curl-pipe-to-sh trust model

`curl ... | sh` executes whatever the default branch serves. Users cannot verify the script before execution unless they download it first, and a compromised branch would compromise installs.

Mitigation: the script lives in-repo and is reviewable; documentation encourages downloading and inspecting before running and offers the fully manual checksum-verified install as the trust-minimal path. The script itself downloads only archives and checksum files, never additional scripts, verifies every download, and never escalates privileges.

### Checksum file as the root of trust

All channels verify `sha256` against the release `checksums.txt`. The checksum file is served from the same GitHub Release as the assets, so it protects against corruption and tampering in transit, not against a fully compromised release.

Mitigation: accepted residual risk for this first version, mitigated by HTTPS-only transport. Signing (for example cosign) is explicitly documented as future work; this change must not pretend to provide stronger guarantees than checksums give.

### GitHub API rate limits and latest-release resolution

Unauthenticated latest-release resolution can hit GitHub rate limits in shared CI or NAT environments, breaking installs intermittently.

Mitigation: support explicit version pinning in install.sh, pin the release version in the npm package so it never queries "latest", and keep channel tests on local stubs so CI never depends on live GitHub endpoints.

### Homebrew tap lives in a separate repository

The formula cannot be fully delivered or tested from this repository, so the Homebrew channel can lag releases or drift from the spec here.

Mitigation: this change records the formula content and test expectations as explicit criteria, documents that delivery may happen via a separate repository/PR, notes that `brew audit`/`install`/`test` can run later in a GitHub Actions macOS runner, and treats formula bumps as a manual step until later automation is specified.

### Windows coverage gap in the first stage

install.sh targets POSIX systems and Homebrew targets macOS, leaving Windows with only npm and manual install at first.

Mitigation: documented interim paths (npm global install and manual `.zip` install) plus explicit future-only status for Scoop and Winget, so expectations are set rather than implied.

### Stale wrapper and formula versions

The npm package pins a release version and the Homebrew formula pins an asset and `sha256`. Each new release requires manual bumps, and forgotten bumps leave users installing old versions.

Mitigation: accepted for the first version; bump steps are documented as part of the release routine, and automation of version bumps is named as future work rather than silently omitted.

### Scope creep into release automation

Implementing install channels makes it tempting to add GoReleaser config, release workflows, publishing automation, signing, or Linux/Windows manifests in the same change.

Mitigation: proposal and tasks list these as hard out-of-scope items; verification inspects `git status --short` and the diff for release automation, packaging, and publishing artifacts; tests confirm no publishing, tagging, or source-control automation occurs.

### Documentation describing unavailable channels

Because this spec lands before releases exist, merged documentation could mislead users into running install commands that do not work yet.

Mitigation: documentation tasks require marking each channel as available or future-only, and the acceptance criteria forbid documenting a channel as available before its implementation stage and a published release exist.

## Trade-offs

- Postinstall download was chosen over packaged per-platform binaries to keep a single unscoped package with no org and low publishing complexity, at the cost of needing network at install time and an `--ignore-scripts` fallback path.
- A user-local, no-sudo default install target favors safety and predictability over the convenience of a globally available binary; PATH guidance compensates.
- Checksums without signing favor shipping a usable, verifiable first version now over a stronger but heavier supply-chain story later; signing remains explicit future work.
- Keeping the Homebrew formula in an external tap keeps this repository free of brew tooling at the cost of cross-repository coordination on each release.
