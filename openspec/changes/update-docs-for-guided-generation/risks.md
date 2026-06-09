# Risks: Update Docs For Guided Generation

## Overstating Implemented Behavior

The architecture roadmap includes AI-assisted, agent-assisted, hybrid, custom template, remote template, config-driven, and interactive generation concepts. Documentation could accidentally present those future modes as implemented.

Mitigation:

- Keep implemented and planned sections separate.
- Describe only `--blank`, built-in `--template <template-name>`, and non-interactive `--guided --type <type> --title "<title>" --summary "<summary>"` generation as implemented.
- List AI-assisted, agent-assisted, hybrid, custom template, remote template, config-driven, and interactive generation only as planned when mentioned.

## Interactive Prompt Confusion

The word "guided" can imply a terminal wizard. The implemented guided generation mode is guided by explicit CLI flags and does not prompt during command execution.

Mitigation:

- State that guided generation is non-interactive.
- State that guided generation uses explicit CLI flags.
- State that guided generation does not prompt during command execution.
- Keep interactive prompts clearly labeled as planned when they are mentioned.

## Guided Type List Drift

Documentation could list guided types that are not implemented or omit implemented guided types.

Mitigation:

- List exactly `feature`, `bugfix`, `docs`, and `refactor`.
- Avoid speculative wording such as "for example" when naming supported guided types.
- Verify the final docs against the implemented command behavior.

## Blank Or Template Documentation Regression

While updating guided generation docs, existing blank or built-in template generation documentation could be removed or made ambiguous.

Mitigation:

- Keep `go run ./cmd/specharbor generate <change-id> --blank` documented as implemented.
- Keep `go run ./cmd/specharbor generate <change-id> --template <template-name>` documented as implemented.
- Review final docs for all three implemented generation modes.

## Missing Overwrite Semantics

Users may run generation in a partially authored change directory. If docs omit overwrite behavior, users may not know whether existing work is safe.

Mitigation:

- Explicitly document that existing files are skipped and not overwritten.
- Explicitly document that partially existing change directories can be completed by creating only missing required files.

## Missing Guided Content Context

Users may not understand what guided generation does differently from blank or built-in template generation if docs do not mention title and summary usage.

Mitigation:

- Explain that guided generated content includes the supplied title and summary.
- Explain that guided output is deterministic, local starter content based on explicit CLI inputs.
- Avoid claiming SpecHarbor inferred additional project-specific requirements.

## Example Drift

Examples that use installed binaries, local absolute paths, unsupported flags, or non-root command shapes may be harder to copy and may drift from repository development usage.

Mitigation:

- Use `go run ./cmd/specharbor ...` examples.
- Keep examples runnable from the repository root.
- Include complete guided examples with `--guided`, `--type`, `--title`, and `--summary`.
- Avoid local absolute paths and environment-specific setup.

## Broad Documentation Diff

The documentation update could expand beyond stale guided-generation wording and accidentally change unrelated docs.

Mitigation:

- Target `README.md`, `docs/usage.md`, and `docs/generation-modes.md` first.
- Update other Markdown files under `docs/` only when needed for consistency.
- Inspect the final diff and confirm it is documentation-only.

## Accidental Behavior Or CI Changes

A documentation-only change could accidentally include Go code, tests, CLI behavior, CI, config, or init template changes.

Mitigation:

- Keep the allowed file list explicit in tasks.
- Inspect `git diff --name-only` before finalizing.
- Verify no Go code, Go tests, CLI behavior, CI configuration, `.specharbor/config.yml`, or init templates changed.

## Premature Task Completion

Because this package is being authored before the documentation implementation, tasks could be marked complete too early.

Mitigation:

- Leave implementation tasks unchecked in this spec package.
- During the future implementation step, check off only tasks actually completed.
