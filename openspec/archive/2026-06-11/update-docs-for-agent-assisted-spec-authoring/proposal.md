# Proposal: Update Docs For Agent-Assisted Spec Authoring

## Summary

Update SpecHarbor public documentation to reflect that dry-run agent-assisted spec authoring is now implemented.

The implemented command shape is:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

The supported agent-assisted authoring types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

This first version is dry-run only. It prints a deterministic authoring plan and a copy-pasteable prompt to stdout. It does not write files, write prompt files, create or modify OpenSpec files, create or modify production code, execute agents, call provider APIs, call local models, call network APIs, call source-control APIs, or call workflow tools.

Existing blank, built-in template, and guided generation remain implemented and must stay documented:

```bash
go run ./cmd/specharbor generate <change-id> --blank
go run ./cmd/specharbor generate <change-id> --template <template-name>
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

This change is documentation-only. It creates a follow-up documentation update package after `implement-agent-assisted-spec-authoring` was implemented and merged.

## Problem

The current documentation still lists agent-assisted generation as planned behavior. That is stale now that dry-run agent-assisted spec authoring exists.

Stale documentation creates several problems:

- users may miss an implemented command that helps external agents draft or refine OpenSpec change packages;
- users may incorrectly treat dry-run agent-assisted spec authoring as future behavior instead of current behavior;
- users may assume the implemented command executes an agent or writes files when it is intentionally dry-run only;
- users may confuse the implemented dry-run prompt output with future AI-assisted generation, agent execution, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompt behavior.

The documentation must be updated without changing CLI behavior, Go code, tests, CI, configuration, init templates, or unrelated OpenSpec changes.

## Goal

Update the README and relevant Markdown docs so they accurately describe:

- `go run ./cmd/specharbor generate <change-id> --blank`;
- `go run ./cmd/specharbor generate <change-id> --template <template-name>`;
- `go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`;
- `go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`;
- the supported agent-assisted authoring types: `feature`, `bugfix`, `docs`, and `refactor`;
- agent-assisted spec authoring as dry-run only in this first version;
- agent-assisted dry-run output as a deterministic authoring plan and copy-pasteable prompt printed to stdout;
- the generated prompt as intended to help an external agent author or refine only the OpenSpec change package;
- implementation as a later step through the normal SpecHarbor workflow;
- `--execute` as currently unsupported with a clear error.

The documentation must explicitly state that dry-run agent-assisted spec authoring does not:

- write files;
- write a prompt file;
- create or modify OpenSpec files;
- create or modify production code;
- execute agents;
- call provider APIs;
- call local models;
- call network APIs;
- call source-control APIs;
- call workflow tools.

## Scope

- Update `README.md` command lists, status sections, and roadmap wording if they still show agent-assisted spec authoring as planned.
- Update `docs/usage.md` with command-level documentation for dry-run agent-assisted spec authoring.
- Update `docs/generation-modes.md` so agent-assisted spec authoring is listed as implemented dry-run behavior, not planned behavior.
- Keep `--blank` documented as implemented.
- Keep `--template <template-name>` documented as implemented.
- Keep `--guided --type <type> --title "<title>" --summary "<summary>"` documented as implemented.
- Keep examples copy-pasteable from the repository root using `go run ./cmd/specharbor ...`.
- Keep planned roadmap items clearly separated from implemented behavior.
- Update `openspec/changes/update-docs-for-agent-assisted-spec-authoring/tasks.md` during the future implementation step, checking off only completed work.

Allowed implementation files for the future documentation step:

- `README.md`
- Markdown files under `docs/`
- `openspec/changes/update-docs-for-agent-assisted-spec-authoring/tasks.md`

## Out of Scope

- Go code changes.
- Go test changes.
- CLI behavior changes.
- Changes to `.github/workflows/ci.yml` or other CI configuration.
- Changes to `.specharbor/config.yml`.
- Changes to init templates.
- Changes to unrelated OpenSpec changes.
- Documentation that claims AI-assisted generation is implemented.
- Documentation that claims agent execution is implemented.
- Documentation that claims custom templates are implemented.
- Documentation that claims remote templates are implemented.
- Documentation that claims config-driven templates are implemented.
- Documentation that claims hybrid generation is implemented.
- Documentation that claims interactive prompts are implemented.
- Documentation for provider setup, provider API keys, local models, network APIs, source-control APIs, workflow connectors, or automated agent execution as current behavior.
- Documentation that suggests agent-assisted dry-run writes prompt files, OpenSpec files, docs, README files, production code, CI files, config files, or init templates.

## Success Criteria

- Public docs describe blank generation as implemented.
- Public docs describe built-in template generation as implemented.
- Public docs describe guided generation as implemented with explicit CLI flags.
- Public docs describe dry-run agent-assisted spec authoring as implemented.
- Public docs list exactly the implemented agent-assisted authoring types: `feature`, `bugfix`, `docs`, and `refactor`.
- Public docs explain that agent-assisted spec authoring is dry-run only in this first version.
- Public docs explain that dry-run prints a deterministic authoring plan and copy-pasteable prompt to stdout.
- Public docs explain that dry-run writes no files and writes no prompt file.
- Public docs explain that dry-run does not create or modify OpenSpec files.
- Public docs explain that dry-run does not create or modify production code.
- Public docs explain that dry-run does not execute agents.
- Public docs explain that dry-run does not call provider APIs, local models, network APIs, source-control APIs, or workflow tools.
- Public docs explain that `--execute` is currently unsupported and returns a clear error.
- Public docs explain that the generated prompt is meant to help an external agent author or refine only the OpenSpec change package.
- Public docs explain that implementation remains a later step through the normal SpecHarbor workflow.
- Planned roadmap items remain clearly separated from implemented command behavior.
- The final diff is limited to `README.md`, Markdown files under `docs/`, and this change's `tasks.md`.
- `go test ./...` passes after implementation.
- `go run ./cmd/specharbor validate update-docs-for-agent-assisted-spec-authoring` passes after implementation.
