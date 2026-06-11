# Acceptance Criteria: Archive Completed OpenSpec Changes

## Criteria

- Completed target changes are no longer active under `openspec/changes/`.
- Completed target changes are present in the expected archive/history location.
- No unrelated active changes are removed.
- Live specs remain valid.
- `go test ./...` passes.
- `go run ./cmd/specharbor validate archive-completed-openspec-changes` passes.
- Repository state is clean and reviewable before the first public release.
- The repository is ready for `perform-first-public-release`.
- No release tag is created.
- No GitHub Release is created.
- No npm package is published.
- No Homebrew tap is created or modified.
- No release automation is modified.
- No install-channel behavior is modified.
- No GoReleaser configuration is modified.
- No package-manager configuration is modified.
- No production code behavior is modified.
- No templates, validation rules, workflow behavior, or generated outputs are altered.
