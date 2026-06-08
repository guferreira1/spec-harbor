# Proposal: Update Docs For Template Generation

## Summary

Update SpecHarbor public documentation to reflect that built-in template generation is now implemented.

The implemented command shape is:

```bash
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

The supported built-in templates are:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Existing blank generation remains implemented and must stay documented:

```bash
go run ./cmd/specharbor generate <change-id> --blank
```

This change is documentation-only. It creates a follow-up documentation update package after `implement-template-generation` was implemented and merged.

## Problem

The current documentation still describes template generation as planned behavior. That is stale now that built-in template generation exists.

Stale documentation creates two problems:

- users may miss an implemented command that helps create common OpenSpec change packages;
- users may incorrectly treat template generation as part of the future roadmap instead of current behavior.

The documentation must be updated without changing CLI behavior, Go code, tests, CI, configuration, init templates, or unrelated OpenSpec changes.

## Goal

Update the README and relevant Markdown docs so they accurately describe:

- `go run ./cmd/specharbor generate <change-id> --blank`;
- `go run ./cmd/specharbor generate <change-id> --template <template-name>`;
- the implemented built-in templates: `feature`, `bugfix`, `docs`, and `refactor`;
- the required OpenSpec files produced by blank and template generation:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- deterministic, local, generic starter content for built-in templates;
- the no-overwrite policy for existing files;
- recoverability for partially existing change directories by creating only missing required files.

## Scope

- Update `README.md` if the current command list or status sections still show template generation as planned.
- Update `docs/usage.md` with command-focused documentation for `generate <change-id> --template <template-name>`.
- Update `docs/generation-modes.md` so built-in template generation is listed as implemented, not planned.
- Keep `--blank` documented as implemented.
- Keep examples copy-pasteable from the repository root using `go run ./cmd/specharbor ...`.
- Keep planned roadmap items clearly separated from implemented behavior.
- Update `openspec/changes/update-docs-for-template-generation/tasks.md` during the future implementation step, checking off only completed work.

## Out of Scope

- Go code changes.
- Go test changes.
- CLI behavior changes.
- Changes to `.github/workflows/ci.yml` or other CI configuration.
- Changes to `.specharbor/config.yml`.
- Changes to init templates.
- Changes to unrelated OpenSpec changes.
- Documentation that claims guided generation is implemented.
- Documentation that claims AI-assisted generation is implemented.
- Documentation that claims agent-assisted generation is implemented.
- Documentation that claims custom templates are implemented.
- Documentation that claims remote templates are implemented.
- Documentation that claims config-driven templates are implemented.
- Documentation that claims interactive prompts are implemented.

## Success Criteria

- Public docs describe built-in template generation as implemented.
- Public docs still describe blank generation as implemented.
- Public docs list exactly the implemented built-in templates: `feature`, `bugfix`, `docs`, and `refactor`.
- Public docs explain that template generation creates `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- Public docs explain that template starter content is deterministic, local, generic, and safe to edit.
- Public docs explain that existing files are skipped and not overwritten.
- Public docs explain that partially existing change directories can be completed by creating missing required files.
- Planned roadmap items remain clearly separated from implemented command behavior.
- The final diff is limited to `README.md`, Markdown files under `docs/`, and this change's `tasks.md`.
- `go test ./...` passes after implementation.
- `go run ./cmd/specharbor validate update-docs-for-template-generation` passes after implementation.
