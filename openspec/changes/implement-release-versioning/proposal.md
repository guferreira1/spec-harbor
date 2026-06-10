# Proposal: Implement Release Versioning

## Problem

SpecHarbor is preparing for public distribution as a Go CLI, but its current version reporting is too small for release artifacts.

The current implementation exposes a hardcoded value from `internal/platform/version` and `specharbor version` prints only that value. That is enough for early local development, but it does not reliably identify which source revision produced an installed binary.

Public distribution needs a stable foundation before adding GitHub Releases, install scripts, npm distribution, Homebrew distribution, native Linux packages, or Windows package manager support.

## Goal

Define and implement a reliable release-versioning foundation for the `specharbor` binary.

After implementation, `specharbor version` must report deterministic build metadata:

- version;
- commit;
- date;
- dirty.

The first release-versioning scope must keep the implementation small, testable, and local. Release builds should inject metadata at build time with Go linker flags. Local development builds should have clear deterministic fallback values.

The change prepares SpecHarbor for later release automation, but it must not publish packages or configure release pipelines.

## Current Behavior

Current relevant files:

- `internal/platform/version/version.go`
- `cmd/specharbor/main.go`
- `internal/adapters/cli/cli.go`
- `internal/adapters/cli/cli_test.go`

Current behavior:

- `internal/platform/version` defines `Version` as a hardcoded constant.
- `specharbor version` prints only that version string.
- The current CLI test expects exactly `version.Version + "\n"`.
- The command currently does not parse or reject extra arguments passed after `version`.

## Scope

- Update the version metadata model in `internal/platform/version`.
- Replace the single hardcoded version string with build-time injectable metadata variables.
- Keep default local development values deterministic.
- Define the version string convention: Git tags use `vX.Y.Z`, while injected binary versions use plain `X.Y.Z`.
- Add a stable human-readable multiline version report.
- Update the `version` CLI command to print that report.
- Reject unsupported arguments and flags for `specharbor version`.
- Add tests for default values, overridden metadata where practical, formatting, and CLI output.
- Add regression tests to confirm unrelated commands keep their behavior.
- Update documentation to explain version reporting and release metadata injection.
- Keep all behavior local, read-only, and independent of Git at runtime.
- Keep release automation, package-manager publishing, and release artifact generation out of this change.

## Out of Scope

- Publishing npm packages.
- Creating GitHub Releases.
- Creating git tags.
- Pushing commits.
- Opening pull requests.
- Adding GoReleaser configuration.
- Adding GitHub release workflows.
- Adding release-specific GitHub Actions workflows.
- Adding an install script.
- Adding Homebrew formula or tap automation.
- Adding native Linux packages.
- Adding Windows package manager support.
- Adding npm wrapper implementation.
- Adding npm package files.
- Adding release publishing scripts.
- Adding package-manager artifacts.
- Adding generated release archives or checksums.
- Adding JSON version output.
- Adding `--short` version output.
- Adding top-level `specharbor --version`.
- Reading `.git` at runtime.
- Executing Git or shell commands at runtime.
- Performing network calls from version reporting.
- Writing files from `specharbor version`.

## Success Criteria

- `specharbor version` prints a deterministic multiline human-readable report.
- Local development builds report `dev`, `unknown`, `unknown`, and `unknown` for version, commit, date, and dirty when no build metadata is injected.
- Release builds can inject version metadata with Go `-ldflags -X` values.
- The documented convention is Git tags such as `v0.1.0` and injected binary versions such as `0.1.0`.
- The `specharbor version` command displays the injected version string exactly as provided.
- Runtime code does not add or remove a leading `v`, inspect Git tags, read `.git`, or run Git commands.
- The linker variable package path is `github.com/guferreira1/spec-harbor/internal/platform/version`.
- The first-scope linker variable names are `Version`, `Commit`, `Date`, and `Dirty`.
- Version reporting does not execute shell commands, call the network, or write files.
- `specharbor version` works from an installed binary and outside a Git repository.
- Existing commands other than `version` keep their behavior.
- Documentation explains how to check the installed version, what `dev` means, and how release metadata will be injected.
- Future distribution channels are documented as future work, not implemented behavior.
- No release automation files, package-manager files, publishing scripts, or generated release artifacts are introduced.
