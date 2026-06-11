# Tasks: Implement Release Versioning

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-release-versioning/`.
- [x] Inspect `internal/platform/version/`.
- [x] Inspect `cmd/specharbor/`.
- [x] Inspect `internal/adapters/cli/`.
- [x] Inspect current version command tests.
- [x] Confirm this change is limited to version metadata, CLI version reporting, tests, and documentation.
- [x] Confirm release publishing, package manager configuration, tags, PRs, commits, and GoReleaser configuration are out of scope.
- [x] Confirm Git release tags use `vX.Y.Z` while injected binary version metadata uses plain `X.Y.Z`.
- [x] Confirm this change does not implement tag-to-binary-version conversion tooling.

## Phase 1: Version Metadata Package

- [x] Replace the hardcoded version constant with linker-injectable package variables in `internal/platform/version`.
- [x] Define defaults exactly as `Version=dev`, `Commit=unknown`, `Date=unknown`, and `Dirty=unknown`.
- [x] Add a small metadata value representation for version, commit, date, and dirty.
- [x] Add a function that returns current metadata from the linker-injected variables.
- [x] Apply empty-value fallback to the deterministic defaults.
- [x] Preserve non-empty injected version strings exactly as provided.
- [x] Do not add a leading `v` to the version string at runtime.
- [x] Do not remove a leading `v` from the version string at runtime.
- [x] Add a human-readable formatter that renders the approved multiline output.
- [x] Keep optional fields such as `builtBy`, `goVersion`, and `platform` out of scope.
- [x] Ensure runtime code does not mutate the linker-injected variables.

## Phase 2: CLI Version Command

- [x] Update `specharbor version` to print the multiline human-readable report.
- [x] Keep the command read-only.
- [x] Keep the command independent of the current working directory.
- [x] Reject unsupported flags passed to `specharbor version`.
- [x] Reject unexpected positional arguments passed to `specharbor version`.
- [x] Do not add `specharbor version --json`.
- [x] Do not add `specharbor version --short`.
- [x] Do not add top-level `specharbor --version`.
- [x] Preserve existing help and unknown-command behavior outside the intentional `version` output change.

## Phase 3: Runtime Safety and Architecture

- [x] Ensure `internal/platform/version` does not read `.git`.
- [x] Ensure `internal/platform/version` does not inspect Git tags.
- [x] Ensure `internal/platform/version` does not import Git libraries.
- [x] Ensure `internal/platform/version` does not import `os/exec`.
- [x] Ensure `internal/platform/version` does not perform network calls.
- [x] Ensure `specharbor version` does not write files.
- [x] Ensure `specharbor version` performs no Git commands.
- [x] Ensure `specharbor version` works without a Git repository.
- [x] Preserve core/adapters dependency boundaries.
- [x] Keep CLI code free of business rules beyond command parsing and output writing.

## Phase 4: Tests

- [x] Add platform tests for default version metadata values.
- [x] Add tests for version output with default metadata.
- [x] Add platform tests for empty-value fallback behavior.
- [x] Add platform tests for formatted version output.
- [x] Add platform tests for overridden metadata values through a constructor, helper, or linker-injection integration test.
- [x] Add tests for version output with an injected plain SemVer-like value such as `0.1.0`.
- [x] Add tests confirming version output preserves injected values as-is, including a value such as `v0.1.0`.
- [x] Add tests or static checks confirming runtime code does not normalize version strings.
- [x] Add tests or static import checks that guard against runtime Git, shell, network, and filesystem dependencies in the version package.
- [x] Add tests or static checks confirming runtime code does not inspect Git tags.
- [x] Update CLI tests for `specharbor version` multiline output.
- [x] Add CLI tests confirming `specharbor version` works without arguments.
- [x] Add CLI tests confirming unsupported version flags are rejected.
- [x] Add CLI tests confirming unexpected version arguments are rejected.
- [x] Add CLI tests or assertions confirming the version command remains read-only.
- [x] Add or preserve regression coverage for `init`.
- [x] Add or preserve regression coverage for `generate`.
- [x] Add or preserve regression coverage for `validate`.
- [x] Add or preserve regression coverage for `prompt`.
- [x] Add or preserve regression coverage for `review`.
- [x] Add or preserve regression coverage for `archive`.
- [x] Add or preserve regression coverage for `workflow`.
- [x] Add or preserve regression coverage for `config`.
- [x] Add or preserve regression coverage for unknown commands.

## Phase 5: Release Automation Exclusion Checks

- [x] Add architecture/static checks confirming no `.goreleaser.yaml` or `.goreleaser.yml` is introduced.
- [x] Add architecture/static checks confirming no `.github/workflows/release.yml`, `.github/workflows/release.yaml`, or release-specific GitHub Actions workflow is introduced.
- [x] Add architecture/static checks confirming no `install.sh` or `scripts/install.sh` is introduced.
- [x] Add architecture/static checks confirming no npm package files or directories are introduced, including `package.json`, `package-lock.json`, `npm/`, or `packages/npm/`.
- [x] Add architecture/static checks confirming no Homebrew formula or tap files are introduced, including `Formula/` or `homebrew/`.
- [x] Add architecture/static checks confirming no Linux package files or directories are introduced, including `nfpm.yaml`, `.nfpm.yaml`, `packaging/`, `debian/`, or `rpm/`.
- [x] Add architecture/static checks confirming no Windows package-manager files or directories are introduced, including `winget/` or `scoop/`.
- [x] Add architecture/static checks confirming no release publishing scripts are introduced.
- [x] Add architecture/static checks confirming no package-manager artifacts are introduced.
- [x] Add architecture/static checks confirming no generated release archives or checksums are introduced.

## Phase 6: Documentation

- [x] Update `README.md` to list `specharbor version` as the way to inspect installed build metadata.
- [x] Update `docs/usage.md` with version command behavior and example output.
- [x] Add `docs/release.md` documenting release metadata fields, default development values, and linker injection.
- [x] Explain that `dev` means no release version was injected.
- [x] Explain that `unknown` means the build did not provide that metadata field.
- [x] Document that Git release tags use `vX.Y.Z`, for example `v0.1.0`.
- [x] Document that binary version metadata uses plain `X.Y.Z`, for example `0.1.0`.
- [x] Document that `specharbor version` displays injected version strings as-is.
- [x] Document that future release tooling will convert tags such as `v0.1.0` into injected binary versions such as `0.1.0`.
- [x] Document that this change does not implement tag-to-binary-version conversion tooling.
- [x] Document the future `-ldflags -X` injection variables using package path `github.com/guferreira1/spec-harbor/internal/platform/version`.
- [x] State that GitHub Releases are not implemented by this change.
- [x] State that install scripts are not implemented by this change.
- [x] State that npm publishing is not implemented by this change.
- [x] State that `specharbor` is the desired future npm package name, subject to verification during publishing.
- [x] State that Homebrew publishing to `guferreira1/homebrew-tap` is not implemented by this change.
- [x] State that native Linux packages and Windows package manager support are future work.
- [x] Avoid documenting unimplemented install commands as available.

## Phase 7: Verification

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor version` and confirm default development output.
- [x] If practical, run a build or test command with `-ldflags -X` overrides and confirm injected output.
- [x] Run `go run ./cmd/specharbor validate implement-release-versioning`.
- [x] Inspect `git status --short` for unexpected release automation, package-manager, publishing, or generated release artifact files.
- [x] Inspect `git diff -- internal/platform/version internal/adapters/cli cmd/specharbor README.md docs/usage.md docs/release.md openspec/changes/implement-release-versioning/`.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
- [x] Leave any task unchecked if the work was not performed or verification could not run.

## Phase 8: Test Engineer Follow-up

- [x] Document plain `go install` without `-ldflags` fallback metadata behavior and expected output in user-facing docs.
- [x] Document that release metadata requires injected `-ldflags` values and runtime does not inspect Git tags, read `.git`, run Git, or normalize versions.
- [x] Confirm work is on `feat/implement-release-versioning`.
