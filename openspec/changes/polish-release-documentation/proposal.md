# Proposal: Polish Release Documentation

## Summary

This change updates SpecHarbor's public documentation after the first real public release and validated distribution channels are available. It is a documentation polish and consolidation change for the current public release, `v0.1.0`, not a release, packaging, automation, or product behavior change.

Completed milestones that this documentation should now reflect:

- `perform-first-public-release`;
- `validate-first-public-release-assets`;
- `publish-npm-package`;
- `implement-homebrew-tap`.

The current public release is `v0.1.0`, built from expected release commit `e6faff91feef07e5c1e47181243286268daf17b5`. The validated Homebrew tap commit is `a61783bcfa44f7eafdce72c70043b76e6f80df9c` in the external `guferreira1/homebrew-tap` repository.

## Problem

SpecHarbor's release and install documentation has evolved through staged release, install-channel, npm, and Homebrew work. Some wording was originally written before real public distribution was available, so the docs can now become stale or contradictory if they continue to describe npm and Homebrew as future work, imply package publishing automation exists, or repeat install commands inconsistently.

Users need a single, accurate public story for installing and verifying SpecHarbor after `v0.1.0`:

- which install channels are available now;
- which channels remain future-only;
- how to verify the installed binary;
- where checksum verification applies;
- how to troubleshoot common install failures.

## Goal

Make the public documentation accurate, consistent, and ready for users after the real `v0.1.0` release, npm package, and Homebrew tap are available.

Candidate documentation files are:

- `README.md`;
- `docs/install.md`;
- `docs/release.md`;
- `docs/usage.md`;
- `packages/npm/specharbor/README.md`, if present and relevant.

Documentation should accurately reflect that:

- GitHub Release `v0.1.0` is available;
- release assets were validated;
- `install.sh` works with the real release;
- npm package `specharbor@0.1.0` is published;
- Homebrew tap `guferreira1/tap/specharbor` is available through external tap repository `guferreira1/homebrew-tap`;
- `go install` is a fallback or developer option only if already supported or documented by the project;
- unsupported and future distribution channels remain clearly marked as future work.

## Scope

This change defines the plan for a later documentation implementation. The implementation should:

- audit the candidate documentation files for stale release, npm, Homebrew, and future-work wording;
- update public install guidance around GitHub Releases, `install.sh`, npm, Homebrew, and any existing `go install` fallback;
- standardize version verification around `specharbor version` and expected public version `0.1.0`;
- describe supported release assets and checksum verification accurately;
- add or refine a distribution-channel status matrix;
- add or refine troubleshooting for common installation failures;
- reduce duplicated or conflicting install commands across public docs;
- keep package publishing automation described as future work unless a later change implements it.

Known validated commands to document consistently where appropriate:

```bash
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
npm install -g specharbor
npx specharbor version
brew install guferreira1/tap/specharbor
```

The standard installed-binary verification command should be:

```bash
specharbor version
```

## Out of Scope

This change does not:

- create a new release;
- create or move tags;
- modify GitHub Release assets;
- publish npm;
- modify npm metadata;
- modify Homebrew formula or tap files;
- modify `install.sh` behavior;
- modify GoReleaser config;
- modify the release workflow;
- add Linux packages;
- add Windows package managers;
- add signing;
- add SBOM;
- add Docker images;
- add publishing automation;
- change CLI or product behavior;
- modify production code.

## Success Criteria

A reviewable implementation plan exists for polishing the release documentation, with clear boundaries around current channels, future-only channels, manual publishing state, troubleshooting, validation, and files in scope. The resulting documentation implementation should be able to pass OpenSpec validation and repository tests without changing production code, release automation, package metadata, `install.sh` behavior, GoReleaser config, or external Homebrew tap files.
