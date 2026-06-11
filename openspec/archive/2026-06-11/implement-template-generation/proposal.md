# Proposal: Implement Template Generation

## Summary

Add deterministic built-in template generation for OpenSpec changes:

```text
specharbor generate <change-id> --template <template-name>
```

The first supported templates are:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Template generation creates the same required OpenSpec change files as blank generation, but writes useful starter content tailored to the selected template. It must remain local, deterministic, generic, safe to commit, and independent of AI providers, external tools, network APIs, source-control APIs, workflow connectors, and configuration.

## Problem

SpecHarbor currently implements only blank OpenSpec change generation:

```text
specharbor generate <change-id> --blank
```

Blank generation is useful when the author wants to write every section manually, but common change types often need similar structure. The documentation already identifies template generation as a planned generation mode, but no command behavior exists yet.

Without a small local template mode, users must repeatedly recreate common proposal, design, task, acceptance, and risk sections for normal feature work, bug fixes, documentation changes, and refactors. Implementing this incrementally also gives SpecHarbor a concrete second generation mode without introducing provider, agent, workflow, custom template, or remote registry complexity.

## Goal

Implement template-based OpenSpec change generation for selected built-in templates:

```text
specharbor generate <change-id> --template feature
specharbor generate <change-id> --template bugfix
specharbor generate <change-id> --template docs
specharbor generate <change-id> --template refactor
```

For each command, SpecHarbor must:

- create `openspec/changes/<change-id>/` when missing;
- create the required files currently created by `--blank`;
- populate missing files with deterministic starter content based on the selected built-in template;
- skip existing files without overwriting content;
- preserve the existing idempotency behavior from blank generation;
- keep unsafe change id validation consistent with existing blank generation;
- report created and skipped files in a clear, human-readable way.

## Scope

- Add a template generation mode for `specharbor generate <change-id> --template <template-name>`.
- Support only these built-in template names: `feature`, `bugfix`, `docs`, and `refactor`.
- Reuse the current required OpenSpec change files:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- Keep existing `--blank` behavior unchanged.
- Return a clear error for unknown template names.
- Return a clear error when `--template` is provided without a template name.
- Return a clear error when `--blank` and `--template` are both provided.
- Return a clear error for unsupported flags or extra arguments.
- Keep filesystem writes behind generation ports.
- Keep generation orchestration in `internal/core/usecase`.
- Keep generation domain concepts in `internal/core/domain`.
- Keep CLI parsing and reporting in `internal/adapters/cli`.
- Provide deterministic built-in template content through a concrete adapter, likely under `internal/adapters/templates`.
- Keep template content generic and safe to commit.
- Add focused tests for domain, ports, template content, use case orchestration, CLI parsing/reporting, idempotency, and regressions for unrelated commands.

## Out of Scope

- Guided generation.
- AI-assisted generation.
- Agent-assisted generation.
- Hybrid generation.
- Custom user templates.
- Remote template registry.
- Template discovery from the filesystem.
- Template marketplace.
- Template configuration through `.specharbor/config.yml`.
- Interactive prompts.
- Provider API integration.
- Source-control integration.
- Workflow integration.
- Runtime network access.
- External process execution.
- Changing `init`.
- Changing `scan`.
- Changing `validate`.
- Changing `prompt`.
- Changing `review`.
- Changing `archive`.
- Changing `config`.
- Changing CI.
- Updating README or documentation outside this OpenSpec change.
- Modifying `.github/workflows/ci.yml`.
- Modifying `.specharbor/config.yml`.

## Success Criteria

- `specharbor generate <change-id> --blank` continues to behave exactly as it does today.
- `specharbor generate <change-id> --template feature` creates or completes a change package with feature-oriented starter content.
- `specharbor generate <change-id> --template bugfix` creates or completes a change package with bugfix-oriented starter content.
- `specharbor generate <change-id> --template docs` creates or completes a change package with documentation-oriented starter content.
- `specharbor generate <change-id> --template refactor` creates or completes a change package with refactor-oriented starter content.
- Existing required files are skipped and never overwritten.
- Unknown template names, missing template names, mixed `--blank` and `--template`, unsupported flags, extra arguments, and unsafe change ids return clear errors.
- Template generation does not call AI providers, external tools, network APIs, source-control APIs, workflow connectors, or configuration systems.
- No unrelated commands, docs, README files, CI files, or config files change.
- `go test ./...` succeeds after implementation.
