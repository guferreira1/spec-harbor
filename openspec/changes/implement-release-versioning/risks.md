# Risks: Implement Release Versioning

## Linker-injected variables conflict with the no-global-mutable-state preference

Go `-ldflags -X` requires package-level string variables. That creates global mutable values even though the project prefers avoiding global mutable state.

Mitigation:

- Limit package variables to `Version`, `Commit`, `Date`, and `Dirty`.
- Treat those variables as build metadata injection points only.
- Do not mutate them at runtime.
- Expose current metadata through a small function that returns a value.

## Plain source builds report development metadata

`go run` and plain `go install` builds will report `dev` and `unknown` values unless linker flags are supplied.

Mitigation:

- Make the fallback values deterministic and documented.
- Document that release artifacts should be built with linker-injected metadata.
- Keep the fallback output clear rather than trying to infer metadata at runtime.

## Dirty state may be unavailable in some build paths

Some future build tools or package managers may not know whether the source tree was dirty.

Mitigation:

- Model `dirty` as a string with allowed values `true`, `false`, or `unknown`.
- Use `unknown` when the build path cannot provide a reliable dirty value.
- Do not inspect the working tree at runtime.

## Runtime Git discovery could make installed binaries unreliable

Reading `.git`, inspecting Git tags, or running `git` from `specharbor version` would fail in installed binaries, package manager environments, source archives, and directories that are not Git repositories.

Mitigation:

- Require build-time injection through linker flags.
- Forbid runtime Git commands.
- Forbid runtime `.git` reads.
- Forbid runtime Git tag inspection.
- Add tests or static checks for version package imports and dependencies.

## Tag and binary version conventions could drift

The project convention is for Git release tags to use a leading `v`, such as `v0.1.0`, while binary version metadata is injected as plain SemVer without the leading `v`, such as `0.1.0`. Future tooling must consistently inject plain SemVer without a leading `v`.

Mitigation:

- Document the convention in design, acceptance criteria, tasks, and release documentation.
- Keep tag-to-binary-version conversion in future release tooling, not runtime code.
- Add tests confirming the runtime displays injected version strings as-is and does not add or remove a leading `v`.

## Manual version injection can bypass the convention

If a maintainer manually injects `v0.1.0`, the binary will display `v0.1.0` because runtime normalization is intentionally out of scope.

Mitigation:

- Treat this as acceptable for manual builds because `specharbor version` is defined to display injected metadata as-is.
- Standardize release injection through future automation that converts tags such as `v0.1.0` to binary versions such as `0.1.0`.
- Do not add runtime tag parsing, Git inspection, or prefix normalization to compensate for manual injection mistakes.

## Output changes may break callers that expect one line

The current command prints only the version string. Multiline output is more useful for releases, but it intentionally changes the command output.

Mitigation:

- Document the output change in this OpenSpec change.
- Update CLI tests intentionally.
- Do not add JSON or short output in this first scope.
- Consider a future `--short` flag only through a separate change if users need script-focused output.

## Release tooling scope could expand too far

Reliable version metadata is a prerequisite for distribution, but publishing automation is a separate concern.

Mitigation:

- Keep GoReleaser config, GitHub Release workflows, install scripts, npm publishing, Homebrew publishing, native packages, Windows package manager support, release publishing scripts, package-manager artifacts, and generated release archives or checksums out of this change.
- Document only the metadata contract those future tools should inject.
- Add release automation through separate OpenSpec changes.
- Add architecture/static checks that fail if this change introduces release automation or package-manager files.

## Documentation could overstate availability

Because the project is preparing for distribution, documentation might accidentally imply that npm, Homebrew, install scripts, or GitHub Releases are already available.

Mitigation:

- Document `specharbor version` as implemented behavior.
- Document linker injection as release-build preparation.
- Explicitly state that publishing and installation channels are future work.
- Avoid unimplemented install commands in current usage docs.

## npm package name availability can change before publishing

The maintainer checked `specharbor` with `npm view` and `npm search`, and no package was found at that time. The name can still become unavailable before a future publishing change.

Mitigation:

- Treat `specharbor` as the desired future npm package name.
- Re-verify name availability during the future npm publishing change.
- Keep this change independent of npm registry writes.

## Linker path drift could break future builds

If the module path or version package path changes, release build commands that inject metadata could silently stop working or inject the wrong package variables.

Mitigation:

- Document the current package path exactly: `github.com/guferreira1/spec-harbor/internal/platform/version`.
- Keep linker variable names stable.
- Add tests or build verification that exercise overridden metadata when practical.
