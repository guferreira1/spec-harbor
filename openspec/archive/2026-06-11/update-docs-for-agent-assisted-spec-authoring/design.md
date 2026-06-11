# Design: Update Docs For Agent-Assisted Spec Authoring

## Overview

This is a documentation-only change. It updates public documentation after dry-run agent-assisted spec authoring has been implemented.

The documentation should present four implemented generation paths:

```bash
go run ./cmd/specharbor generate <change-id> --blank
go run ./cmd/specharbor generate <change-id> --template <template-name>
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

Agent-assisted spec authoring must be described as implemented dry-run behavior, not as planned generation. The docs must also make clear that this mode does not generate files. It prints a deterministic authoring plan and copy-pasteable prompt to stdout so a human can paste the prompt into an external coding-agent tool.

The docs must not present AI-assisted generation, agent execution, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompts as implemented behavior.

## Documentation Source Of Truth

The documentation update must be based on the implemented `implement-agent-assisted-spec-authoring` behavior and the existing generation modes:

- `--blank` creates the standard OpenSpec change files with blank/manual content.
- `--template <template-name>` creates the standard OpenSpec change files with built-in template starter content.
- `--guided --type <type> --title "<title>" --summary "<summary>"` creates the standard OpenSpec change files with deterministic guided starter content.
- `--agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"` runs dry-run agent-assisted spec authoring.
- The only supported agent-assisted authoring types are `feature`, `bugfix`, `docs`, and `refactor`.
- Agent-assisted spec authoring is dry-run only in this first version.
- Dry-run prints a deterministic authoring plan to stdout.
- Dry-run prints a deterministic, copy-pasteable authoring prompt to stdout.
- Dry-run writes no files.
- Dry-run writes no prompt file.
- Dry-run does not create or modify OpenSpec files.
- Dry-run does not create or modify production code.
- Dry-run does not execute external agents or local agent commands.
- Dry-run does not call provider APIs, local models, network APIs, source-control APIs, or workflow tools.
- `--execute` is explicitly unsupported in this first version and returns a clear error.
- The generated prompt is meant to help an external agent author or refine only the OpenSpec change package.
- Implementation remains a later step through the normal SpecHarbor workflow.

The standard required OpenSpec change files are:

- `proposal.md`
- `design.md`
- `tasks.md`
- `acceptance-criteria.md`
- `risks.md`

The docs must not infer implemented behavior from the architecture roadmap. `openspec/project.md` and the architecture spec mention broader generation strategies, but this documentation update must distinguish current command behavior from future product direction.

## Files To Update Later

The future implementation step may update only:

- `README.md`
- Markdown files under `docs/`
- `openspec/changes/update-docs-for-agent-assisted-spec-authoring/tasks.md`

Expected documentation touch points:

- `README.md`: update the command list and status sections if they still show agent-assisted generation only as planned.
- `docs/usage.md`: add command-level usage for dry-run agent-assisted spec authoring.
- `docs/generation-modes.md`: move agent-assisted spec authoring from planned behavior to implemented dry-run behavior.

Other Markdown files under `docs/` may be updated only if needed to keep command examples, workflow descriptions, or roadmap language consistent.

## README Guidance

The README should stay concise. If it lists implemented commands, it should include dry-run agent-assisted spec authoring alongside blank, built-in template, and guided generation.

Recommended command examples:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-feature --template feature
go run ./cmd/specharbor generate add-reports --guided --type feature --title "Add reports" --summary "Create report generation support"
go run ./cmd/specharbor generate draft-agent-docs --agent-assisted --agent codex --type docs --title "Update agent-assisted docs" --summary "Document dry-run agent-assisted spec authoring."
```

If the README has a status section, it should list dry-run agent-assisted spec authoring as implemented and remove generic agent-assisted generation from planned behavior unless the planned entry is narrowed to future execution or non-dry-run capabilities.

The README should not become the full generation reference. Detailed mode behavior belongs in `docs/usage.md` and `docs/generation-modes.md`.

## Usage Documentation Guidance

`docs/usage.md` should document:

- how to run dry-run agent-assisted spec authoring from the repository root;
- the command shape `generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`;
- the four supported agent-assisted authoring types: `feature`, `bugfix`, `docs`, and `refactor`;
- that agent-assisted spec authoring is dry-run only in this first version;
- that dry-run prints a deterministic authoring plan and copy-pasteable prompt to stdout;
- that dry-run writes no files and writes no prompt file;
- that dry-run does not create or modify OpenSpec files;
- that dry-run does not create or modify production code;
- that dry-run does not execute agents;
- that dry-run does not call provider APIs, local models, network APIs, source-control APIs, or workflow tools;
- that `--execute` is currently unsupported and returns a clear error;
- that the generated prompt is intended for an external agent to author or refine only the OpenSpec change package;
- that implementation remains a later step through the normal SpecHarbor workflow;
- continued support for `generate <change-id> --blank`;
- continued support for `generate <change-id> --template <template-name>`;
- continued support for `generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`.

Examples must use `go run ./cmd/specharbor ...`, not installed binary-only examples.

Recommended example set:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-example-feature --template feature
go run ./cmd/specharbor generate add-guided-feature --guided --type feature --title "Add guided feature" --summary "Create a guided OpenSpec change from explicit CLI inputs."
go run ./cmd/specharbor generate draft-agent-feature --agent-assisted --agent codex --type feature --title "Draft agent feature spec" --summary "Prepare an OpenSpec change package for a feature."
go run ./cmd/specharbor generate draft-agent-bugfix --agent-assisted --agent codex --type bugfix --title "Draft agent bugfix spec" --summary "Prepare an OpenSpec change package for a bugfix."
go run ./cmd/specharbor generate draft-agent-docs --agent-assisted --agent codex --type docs --title "Draft agent docs spec" --summary "Prepare an OpenSpec change package for documentation work."
go run ./cmd/specharbor generate draft-agent-refactor --agent-assisted --agent codex --type refactor --title "Draft agent refactor spec" --summary "Prepare an OpenSpec change package for a behavior-preserving refactor."
```

The usage docs must not claim AI-assisted generation, agent execution, custom template paths, remote templates, config-backed template selection, hybrid generation, or interactive prompts are implemented.

## Generation Modes Guidance

`docs/generation-modes.md` should separate implemented behavior from planned behavior.

Implemented:

- blank generation;
- built-in template generation for `feature`, `bugfix`, `docs`, and `refactor`;
- guided generation for `feature`, `bugfix`, `docs`, and `refactor`;
- dry-run agent-assisted spec authoring for `feature`, `bugfix`, `docs`, and `refactor`.

Planned or deferred:

- AI-assisted generation;
- agent execution or non-dry-run agent-assisted workflows;
- hybrid generation;
- custom templates;
- remote templates;
- config-driven templates;
- interactive prompts.

Agent-assisted spec authoring should no longer appear as a generic planned item. If planned agent-related behavior is mentioned, it must be clearly limited to deferred execution, file application, workflow automation, or other non-dry-run capabilities that are not implemented.

## Workflow Guidance

If `docs/workflow.md` or `docs/agent-roles.md` are updated for consistency, they should preserve the normal SpecHarbor sequence:

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

Dry-run agent-assisted spec authoring belongs before implementation. It helps author or refine an OpenSpec change package. It is not an implementation step and must not be documented as generating production code or executing agents.

## Verification Strategy

Because this is documentation-only, verification focuses on scope and accuracy:

- inspect the final diff and confirm only allowed Markdown files changed;
- confirm no Go code, Go tests, CI, config, or init templates changed;
- confirm command examples are copy-pasteable from the repository root;
- confirm blank, built-in template, guided, and dry-run agent-assisted modes are all documented as implemented;
- confirm the agent-assisted command syntax uses `--agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"`;
- confirm supported agent-assisted authoring types are exactly `feature`, `bugfix`, `docs`, and `refactor`;
- confirm dry-run boundaries are explicit;
- confirm `--execute` is documented as unsupported;
- confirm planned roadmap items are not mixed with implemented behavior;
- run `go test ./...`;
- run `go run ./cmd/specharbor validate update-docs-for-agent-assisted-spec-authoring`.

No `gofmt` step is required unless Go files are unexpectedly modified. Go files must not be modified by this change.
