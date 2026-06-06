# Risks: Implement Generation Foundation

## Architecture leakage

Generation touches CLI parsing, project-root discovery, filesystem writes, overwrite policy, required-file policy, starter content, and user-facing reporting. The main risk is placing generation rules directly in the CLI adapter.

Mitigation:

- Keep CLI responsibilities limited to argument parsing, current-working-directory lookup, dependency construction, and report formatting.
- Keep generation orchestration in `internal/core/usecase`.
- Keep generation concepts and structured results in `internal/core/domain`.
- Reuse the existing domain-level required OpenSpec change file policy, currently `domain.RequiredOpenSpecChangeFiles()`, for the required file list.
- Do not duplicate required-file policy in generation-specific code.
- Use a generation-specific filesystem port from `internal/core/ports`.
- Keep concrete filesystem behavior in `internal/adapters/filesystem`.
- Keep starter content in a content/template adapter if content is not simple enough to remain private use-case data.

## Overbuilding future generation modes

SpecHarbor must eventually support guided, template, AI-assisted, agent-assisted, and hybrid generation. Building a public strategy registry, provider framework, agent framework, template system, or workflow dispatch layer for the first blank mode would add unused surface area.

Mitigation:

- Accept a generation mode in the use case input, but implement only blank mode.
- Represent blank generation mode in the domain.
- Return structured results that future modes can reuse.
- Avoid exported generation strategy, registry, factory, or Chain of Responsibility abstractions until another mode needs them.
- Do not add provider, AI, agent target, workflow connector, or external execution ports in this change.

## Underbuilding the foundation

A simple CLI function that creates five files would satisfy the first behavior but would make future modes harder to add cleanly.

Mitigation:

- Introduce domain result concepts for generated changes.
- Add a generation-specific filesystem port.
- Reuse the existing shared domain required-file policy and keep it out of the CLI adapter.
- Return created and skipped files as structured data instead of deriving them from printed output.
- Cover the use case with fake ports so orchestration remains independent of local filesystem behavior.

## Accidental overwrite

Users may run blank generation against a partially authored change. Overwriting `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, or `risks.md` would destroy work.

Mitigation:

- Use write-if-absent behavior for every generated file.
- Prefer exclusive file creation in the filesystem adapter.
- Record skipped existing files in the structured result.
- Treat partially existing change directories as recoverable by creating only missing required files.
- Add tests proving existing file contents are preserved.
- Do not add `--force` or overwrite behavior in this change.

## Unsafe change id path handling

The change id is used to build a path below `openspec/changes/`. Absolute paths, path separators, dot segments, drive prefixes, or traversal-like input could write outside the intended directory or produce confusing results.

Mitigation:

- Validate the change id before any filesystem write.
- Reject empty ids, `.`, `..`, `/`, `\`, `:`, leading `-`, absolute-path input, and traversal-like input.
- Build the target path only as `openspec/changes/<change-id>`.
- Add tests that unsafe ids are rejected before directory or file creation.

## Project initialization ambiguity

Generation depends on an existing OpenSpec project. If the command silently creates `openspec/project.md` or `openspec/changes/`, it would blur the boundary between `init` and `generate`.

Mitigation:

- Verify `openspec/project.md` and `openspec/changes/` through the generation filesystem port.
- Return a clear execution error telling the user to run `specharbor init` first when project structure is unavailable.
- Do not create missing project-level OpenSpec structure in the generation use case.
- Leave project initialization to `specharbor init`.

## Starter content becomes too prescriptive

Blank generation should help users start authoring without pretending to perform semantic generation. Starter files that look complete could mislead users or agents.

Mitigation:

- Use clearly skeletal Markdown headings and placeholder guidance.
- Keep generated tasks unchecked.
- Avoid generated claims that implementation has happened.
- Do not call AI providers or infer project-specific requirements.
- Keep starter content generic and deterministic.

## Report format churn

Human-readable reports can become fragile if they rely on decorative formatting or incidental wording.

Mitigation:

- Keep output concise and deterministic.
- Test for important content: operation status, change id, relative path, created count, skipped count, and filenames.
- Avoid banners, absolute local paths, debug output, provider output, and unrelated summaries.

## Accidental validation behavior changes

Generation and validation both depend on the same required OpenSpec change files. Duplicating the list in generation could cause drift, while a broad refactor could accidentally change validation behavior.

Mitigation:

- Keep this change scoped to generation.
- Reuse the existing domain-level required OpenSpec change file policy without changing validation semantics.
- Preserve validation tests and CLI behavior.
- Add regression tests for `specharbor validate`.
