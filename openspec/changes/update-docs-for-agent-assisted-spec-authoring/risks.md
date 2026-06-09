# Risks: Update Docs For Agent-Assisted Spec Authoring

## Overstating Implemented Behavior

The architecture roadmap includes AI-assisted generation, hybrid generation, custom templates, remote templates, config-driven templates, interactive prompts, and future agent execution concepts. Documentation could accidentally present those future modes as implemented.

Mitigation:

- Keep implemented and planned sections separate.
- Describe only blank, built-in template, guided, and dry-run agent-assisted spec authoring as implemented generation behavior.
- List AI-assisted generation, hybrid generation, custom templates, remote templates, config-driven templates, and interactive prompts only as planned when mentioned.
- List future agent execution or non-dry-run agent-assisted behavior only as planned or deferred when mentioned.

## Dry-Run Safety Confusion

Users may assume agent-assisted spec authoring executes an agent, writes OpenSpec files, writes a prompt file, or changes production code.

Mitigation:

- State that this first version is dry-run only.
- State that it prints a deterministic authoring plan and copy-pasteable prompt to stdout.
- State that it writes no files.
- State that it writes no prompt file.
- State that it does not create or modify OpenSpec files.
- State that it does not create or modify production code.
- State that it does not execute agents.

## External Integration Confusion

Users may assume agent-assisted spec authoring calls provider APIs, local models, network APIs, source-control APIs, or workflow tools because it mentions external agents.

Mitigation:

- Explicitly document that dry-run does not call provider APIs.
- Explicitly document that dry-run does not call local models.
- Explicitly document that dry-run does not call network APIs.
- Explicitly document that dry-run does not call source-control APIs.
- Explicitly document that dry-run does not call workflow tools.
- Avoid provider setup, credential setup, source-control automation, and workflow automation instructions in current-behavior docs.

## Unsupported Execute Confusion

Users may expect `--execute` to run an agent because the flag exists as a deferred concept.

Mitigation:

- Document that `--execute` is currently unsupported.
- Document that `--execute` returns a clear error.
- Keep future execution behavior separated from implemented dry-run behavior.

## Agent-Assisted Type List Drift

Documentation could list agent-assisted authoring types that are not implemented or omit implemented types.

Mitigation:

- List exactly `feature`, `bugfix`, `docs`, and `refactor`.
- Avoid speculative wording such as "for example" when naming supported authoring types.
- Verify the final docs against the implemented command behavior.

## Existing Generation Mode Regression

While updating agent-assisted docs, existing blank, built-in template, or guided generation documentation could be removed or made ambiguous.

Mitigation:

- Keep `go run ./cmd/specharbor generate <change-id> --blank` documented as implemented.
- Keep `go run ./cmd/specharbor generate <change-id> --template <template-name>` documented as implemented.
- Keep `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"` documented as implemented.
- Review final docs for all four implemented generation paths.

## Workflow Boundary Confusion

The generated prompt is meant to help an external agent author or refine only the OpenSpec change package. Users may misunderstand it as an instruction to implement production code.

Mitigation:

- Explain that the prompt is for OpenSpec change-package authoring or refinement only.
- Explain that implementation remains a later step through the normal SpecHarbor workflow.
- Avoid language that describes agent-assisted spec authoring as code generation or implementation.

## Example Drift

Examples that use installed binaries, local absolute paths, unsupported flags, or non-root command shapes may be harder to copy and may drift from repository development usage.

Mitigation:

- Use `go run ./cmd/specharbor ...` examples.
- Keep examples runnable from the repository root.
- Include complete agent-assisted examples with `--agent-assisted`, `--agent`, `--type`, `--title`, and `--summary`.
- Avoid local absolute paths and environment-specific setup.

## Broad Documentation Diff

The documentation update could expand beyond stale agent-assisted wording and accidentally change unrelated docs.

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
