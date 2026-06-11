# Proposal: Implement Homebrew Tap

## Problem

SpecHarbor has published and validated its first public release, `v0.1.0`, from commit `e6faff91feef07e5c1e47181243286268daf17b5`. The release assets cover Linux, Darwin/macOS, and Windows for `amd64`/`x86_64` and `arm64`, and `checksums.txt` has been validated. The npm package publication for `specharbor@0.1.0` has also been completed manually.

The next distribution channel is Homebrew. SpecHarbor already documents Homebrew as planned, but users cannot install through Homebrew until an external tap repository contains a validated formula.

## Goal

Add Homebrew installation support for SpecHarbor through a personal external Homebrew tap repository:

- tap repository: `guferreira1/homebrew-tap`;
- formula name: `specharbor`;
- desired install command: `brew install guferreira1/tap/specharbor`.

The formula must consume official GitHub Release assets from `guferreira1/spec-harbor`, pin version `0.1.0` initially, use verified SHA-256 values from the published release `checksums.txt`, install the `specharbor` binary, and include a basic Homebrew test that runs `specharbor version`.

## Scope

This change defines the plan for adding Homebrew installation support. The implementation phase should later create or update the external tap repository `guferreira1/homebrew-tap` and add `Formula/specharbor.rb` there.

The formula should install from official `v0.1.0` GitHub Release assets:

- `specharbor_Darwin_arm64.tar.gz` for Apple Silicon macOS;
- `specharbor_Darwin_x86_64.tar.gz` for Intel macOS.

Documentation updates in the main SpecHarbor repository should happen only after the external tap has been validated.

## Repository Boundary

The tap is expected to live outside the main SpecHarbor repository. The main repository should contain this OpenSpec planning change and, after validation, documentation updates if needed.

The formula itself should not be committed directly to the main SpecHarbor repository unless it is clearly used only as documentation or a generated reference.

## Out of Scope

This change does not:

- create a new SpecHarbor release;
- modify GoReleaser release behavior;
- modify npm publishing;
- modify `install.sh` behavior;
- add Linux packages;
- add Windows package managers;
- add signing;
- add SBOM generation;
- automate package publishing;
- create tags;
- create releases;
- publish packages;
- modify production code.

## Success Criteria

- A reviewable implementation plan exists for adding the external Homebrew tap.
- The tap repository, formula name, install command, asset source, version pin, checksum requirement, binary install behavior, and formula test are explicit.
- Validation and documentation boundaries prevent the main repository from claiming Homebrew support before the external tap works.
