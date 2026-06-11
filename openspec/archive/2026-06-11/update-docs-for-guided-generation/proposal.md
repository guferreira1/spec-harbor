# Proposal: Update Docs For Guided Generation

## Summary

Update SpecHarbor public documentation to reflect that guided generation is now implemented.

The implemented guided command shape is:

```bash
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

The supported guided types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Existing blank and built-in template generation remain implemented and must stay documented:

```bash
go run ./cmd/specharbor generate <change-id> --blank
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

This change is documentation-only. It creates a follow-up documentation update package after `implement-guided-generation` was implemented and merged.

## Problem

The current documentation still describes guided generation as planned behavior. That is stale now that deterministic guided generation exists.

Stale documentation creates three problems:

- users may miss an implemented command that creates OpenSpec change packages from explicit title and summary inputs;
- users may incorrectly treat guided generation as part of the future roadmap instead of current behavior;
- users may confuse implemented non-interactive guided generation with future interactive, AI-assisted, agent-assisted, or hybrid generation concepts.

The documentation must be updated without changing CLI behavior, Go code, tests, CI, configuration, init templates, or unrelated OpenSpec changes.

## Goal

Update the README and relevant Markdown docs so they accurately describe:

- `go run ./cmd/specharbor generate <change-id> --blank`;
- `go run ./cmd/specharbor generate <change-id> --template <template-name>`;
- `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`;
- the supported guided types: `feature`, `bugfix`, `docs`, and `refactor`;
- guided generation as deterministic, local, and non-interactive;
- guided generation as driven by explicit CLI flags, with no prompts during command execution;
- the required OpenSpec files produced by blank, built-in template, and guided generation:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- guided generated content including the supplied title and summary;
- the no-overwrite policy for existing files;
- recoverability for partially existing change directories by creating only missing required files.

## Scope

- Update `README.md` if the command list, status sections, or planned roadmap wording still show guided generation as planned.
- Update `docs/usage.md` with command-level documentation for guided generation.
- Update `docs/generation-modes.md` so guided generation is listed as implemented, not planned.
- Keep `--blank` documented as implemented.
- Keep `--template <template-name>` documented as implemented.
- Keep examples copy-pasteable from the repository root using `go run ./cmd/specharbor ...`.
- Keep planned roadmap items clearly separated from implemented behavior.
- Update `openspec/changes/update-docs-for-guided-generation/tasks.md` during the future implementation step, checking off only completed work.

Allowed implementation files for the future documentation step:

- `README.md`
- Markdown files under `docs/`
- `openspec/changes/update-docs-for-guided-generation/tasks.md`

## Out of Scope

- Go code changes.
- Go test changes.
- CLI behavior changes.
- Changes to `.github/workflows/ci.yml` or other CI configuration.
- Changes to `.specharbor/config.yml`.
- Changes to init templates.
- Changes to unrelated OpenSpec changes.
- Documentation that claims AI-assisted generation is implemented.
- Documentation that claims agent-assisted generation is implemented.
- Documentation that claims custom templates are implemented.
- Documentation that claims remote templates are implemented.
- Documentation that claims config-driven templates are implemented.
- Documentation that claims hybrid generation is implemented.
- Documentation that claims interactive prompts are implemented.
- Documentation that describes AI-assisted generation, agent-assisted generation, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompts as current command behavior.

## Success Criteria

- Public docs describe blank generation as implemented.
- Public docs describe built-in template generation as implemented.
- Public docs describe guided generation as implemented with explicit CLI flags.
- Public docs list exactly the implemented guided types: `feature`, `bugfix`, `docs`, and `refactor`.
- Public docs explain that guided generation is deterministic, local, and non-interactive.
- Public docs explain that guided generation does not prompt during command execution.
- Public docs explain that guided generation creates `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- Public docs explain that guided generated content includes the supplied title and summary.
- Public docs explain that existing files are skipped and not overwritten.
- Public docs explain that partially existing change directories can be completed by creating only missing required files.
- Planned roadmap items remain clearly separated from implemented command behavior.
- The final diff is limited to `README.md`, Markdown files under `docs/`, and this change's `tasks.md`.
- `go test ./...` passes after implementation.
- `go run ./cmd/specharbor validate update-docs-for-guided-generation` passes after implementation.
