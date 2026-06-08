# Risks: Update Docs For Template Generation

## Overstating Implemented Behavior

The architecture roadmap includes guided, AI-assisted, agent-assisted, hybrid, custom template, remote template, and config-driven generation concepts. Documentation could accidentally present those future modes as implemented.

Mitigation:

- Keep implemented and planned sections separate.
- Describe only `--blank` and built-in `--template <template-name>` generation as implemented.
- List custom, remote, guided, AI-assisted, agent-assisted, hybrid, config-driven, and interactive generation only as planned when mentioned.

## Template List Drift

Documentation could list templates that are not implemented or omit implemented templates.

Mitigation:

- List exactly `feature`, `bugfix`, `docs`, and `refactor`.
- Avoid speculative wording such as "for example" when naming implemented built-in templates.
- Verify the final docs against the implemented command behavior.

## Missing Overwrite Semantics

Users may run generation in a partially authored change directory. If docs omit overwrite behavior, users may not know whether existing work is safe.

Mitigation:

- Explicitly document that existing files are skipped and not overwritten.
- Explicitly document that partially existing change directories can be completed by creating only missing required files.

## Example Drift

Examples that use installed binaries, local absolute paths, or non-root command shapes may be harder to copy and may drift from repository development usage.

Mitigation:

- Use `go run ./cmd/specharbor ...` examples.
- Keep examples runnable from the repository root.
- Avoid local absolute paths and environment-specific setup.

## Broad Documentation Diff

The documentation update could expand beyond the stale template-generation wording and accidentally change unrelated docs.

Mitigation:

- Target `README.md`, `docs/usage.md`, and `docs/generation-modes.md` first.
- Update other Markdown files under `docs/` only when needed for consistency.
- Inspect the final diff and confirm it is documentation-only.

## Accidental Behavior Or CI Changes

A documentation-only change could accidentally include Go code, tests, CLI behavior, CI, config, or init template changes.

Mitigation:

- Keep the allowed file list explicit in tasks.
- Inspect `git diff --name-only` before finalizing.
- Verify no Go code, Go tests, CI configuration, `.specharbor/config.yml`, or init templates changed.

## Premature Task Completion

Because this package is being authored before the documentation implementation, tasks could be marked complete too early.

Mitigation:

- Leave implementation tasks unchecked in this spec package.
- During the future implementation step, check off only tasks actually completed.
