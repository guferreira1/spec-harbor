# Risks: Polish Release Documentation

## Risks and Mitigations

### Docs could claim an unavailable install channel works

The documentation could say a channel is available without verifying that the public release, package, or tap is actually reachable.

Mitigation: require the implementation tasks to confirm GitHub Release `v0.1.0`, release assets, `checksums.txt`, npm package `specharbor@0.1.0`, and Homebrew tap `guferreira1/tap/specharbor` before updating user-facing wording.

### Docs could leave stale future-work wording for npm or Homebrew

Old pre-publication language could continue to say npm or Homebrew are planned even though they are now published and validated.

Mitigation: explicitly search README, install docs, release docs, usage docs, and npm README for stale npm/Homebrew future-work wording and update only the wording that conflicts with the validated current state.

### Docs could imply publishing automation exists while publishing is still manual

Users or maintainers could believe npm package publishing or Homebrew formula updates happen automatically when only the tag-triggered GitHub Release workflow exists.

Mitigation: document that GitHub Releases are created through the tag workflow, npm was published manually for `0.1.0`, Homebrew tap maintenance is manual for now, and package publishing automation remains future work.

### Docs could confuse `v0.1.0` tag with binary version `0.1.0`

Mixing tag and binary version forms could make users think the binary output is wrong or that the release tag is missing the leading `v`.

Mitigation: use `v0.1.0` only for Git tags and GitHub Release references, use `0.1.0` for `specharbor version` output, and include a short explanation in release/version documentation.

### Docs could document the wrong Homebrew tap command

The external repository is `guferreira1/homebrew-tap`, while the install command uses Homebrew's tap shorthand `guferreira1/tap/specharbor`. Mixing these forms could create a broken command.

Mitigation: document the install command exactly as `brew install guferreira1/tap/specharbor` and mention `guferreira1/homebrew-tap` only when describing the external repository boundary.

### Docs could document the wrong npm package name or a scoped fallback

Earlier planning mentioned scoped-package fallback only if the unscoped name became unavailable. Now the package is published as `specharbor`, so stale scoped examples would be misleading.

Mitigation: require docs to use unscoped `specharbor`, require package metadata inspection for context, and avoid scoped package examples unless a future change intentionally changes package naming.

### Troubleshooting could become too verbose or duplicated across files

Repeating full troubleshooting guidance in README, install docs, release docs, and npm README could create drift and conflicting advice.

Mitigation: keep detailed troubleshooting primarily in `docs/install.md`, use README as a concise summary with a link, and update package README only for npm-specific facts that users need there.

### Future distribution channel status could drift

Future-only channels such as `.deb/.rpm`, Scoop, Winget, signing, SBOM, Docker, and publishing automation may become stale if the status matrix is not maintained when future work lands.

Mitigation: make the channel status matrix explicit and easy to update, and require future distribution changes to update the matrix as part of their OpenSpec acceptance criteria.
