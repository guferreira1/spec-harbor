# Design: Update Docs and README

## Overview

This change is a documentation polish feature. It should update the public-facing documentation to match SpecHarbor's current direction as a Go CLI for OpenSpec-based development workflows for AI coding agents.

The docs should present the verified core workflow commands as implemented and present scan/config capabilities according to their actual merge status at implementation time.

The documentation must be careful about status. The core workflow commands listed in this change may be documented as implemented after the implementation agent verifies them on the stabilized branch. Scan, config, and CI work must be documented according to actual merge status at implementation time.

The implementation must not change CLI behavior, Go code, tests, or GitHub Actions. It should only update Markdown documentation and this change's task status.

## Documentation File Plan

The implementation should use a small documentation set with clear responsibilities:

```text
README.md
docs/usage.md
docs/workflow.md
docs/agent-roles.md
docs/contributing.md
docs/development.md
```

These files are justified as follows:

- `README.md`: short public entry point, product definition, current workflow, quick examples, local build/test basics, status notes, and links to deeper docs.
- `docs/usage.md`: command-focused usage guide with copy-pasteable examples.
- `docs/workflow.md`: OpenSpec/SDD workflow guide from idea to archive.
- `docs/agent-roles.md`: update the existing role guide with current role prompt behavior.
- `docs/contributing.md`: contributor workflow, OpenSpec requirements, review expectations, and parallel branch hygiene.
- `docs/development.md`: local development commands, Go-specific repository CI notes, and verification expectations.

Existing docs must be reconciled:

- `docs/agent-roles.md` should be updated in place.
- `docs/getting-started.md` should not remain stale. Either fold its content into `docs/usage.md` and replace it with a short pointer, or update it as a concise quickstart if keeping it avoids broken expectations.
- `docs/generation-modes.md` should not claim unimplemented generation modes are complete. Either update it with implemented/planned status or fold the relevant content into `docs/workflow.md`.

Do not create additional documentation files beyond the planned set unless the implementer identifies a concrete need and documents the reason in `tasks.md` or the final implementation response.

## README Content

`README.md` should be concise and avoid duplicating the full docs.

It should include:

- one direct definition of SpecHarbor;
- a short explanation of why OpenSpec/SDD helps AI coding-agent workflows;
- a current command flow using implemented commands only;
- a brief current status section;
- local build and test commands;
- links to usage, workflow, agent roles, contributing, and development docs;
- a short note that scan/config are roadmap or implemented based on merge status;
- a short note that SpecHarbor is dogfooding OpenSpec changes for its own development.

Recommended command examples:

```bash
go run ./cmd/specharbor init
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor validate add-example-feature
go run ./cmd/specharbor prompt add-example-feature --role implementer
go run ./cmd/specharbor review add-example-feature
go run ./cmd/specharbor archive add-example-feature
```

If the binary build command is documented, use:

```bash
go build ./cmd/specharbor
```

The README should not include badges unless the implementation explicitly justifies them. It should not document package-manager installation, release downloads, or install scripts unless those capabilities exist.

## Usage Documentation

`docs/usage.md` should be the command reference for ordinary users.

It should cover:

- prerequisites at a high level;
- running from the repository root with `go run ./cmd/specharbor ...`;
- optional local binary build with `go build ./cmd/specharbor`;
- `specharbor init`;
- `specharbor generate <change-id> --blank`;
- `specharbor validate <change-id>`;
- `specharbor prompt <change-id> --role <role>`;
- `specharbor review <change-id>`;
- `specharbor archive <change-id>`;
- command ordering in the normal workflow;
- clear notes for unsupported, planned, or not-yet-merged commands.

Usage examples should be simple and copy-pasteable. They should avoid local absolute paths, secrets, API keys, provider-specific credentials, and commands that depend on unpublished releases.

If `specharbor scan` is merged and verified at documentation implementation time, include it as an implemented informational command. If it is not merged, mention scan only in a roadmap section.

If config behavior is not implemented and verified at documentation implementation time, do not include config as an implemented command. Mention config only as planned support for future provider, template, scanner, or workflow configuration.

## Workflow Documentation

`docs/workflow.md` should explain SpecHarbor's OpenSpec/SDD flow:

```text
Idea -> OpenSpec change -> validation -> role prompt -> implementation -> review -> archive
```

It should explain the expected change package:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

It should describe the role of each file without turning the docs into a long template. It should make clear that OpenSpec changes are meant to constrain implementation scope, reduce unrelated edits, and give coding agents executable tasks.

It should also explain dogfooding: SpecHarbor changes should be specified through OpenSpec before meaningful implementation work.

## Agent Role Documentation

`docs/agent-roles.md` should describe role-based work at a practical level:

- Spec Author creates or refines OpenSpec change files.
- Architecture Reviewer checks specs or diffs against architecture boundaries.
- Implementer applies an approved change.
- Test Engineer adds or updates focused tests.
- Change Reviewer reviews the final diff against the active change.

The docs should show how to generate a role prompt:

```bash
go run ./cmd/specharbor prompt add-example-feature --role implementer
```

Role names and examples must match the current prompt generator behavior on the stabilized branch. The docs should link to `.specharbor/rules/` as the source of detailed role rules instead of duplicating all instructions.

The docs may mention agent-assisted workflows at a high level. They must not require provider API keys for agent-assisted workflows and must not add detailed provider setup instructions in this change.

## Contributor Documentation

`docs/contributing.md` should explain how contributors and coding agents should work on SpecHarbor:

- read `AGENTS.md`, `.specharbor/rules/global.md`, the relevant role rule, `openspec/project.md`, and `openspec/specs/architecture/spec.md`;
- create or update an OpenSpec change before meaningful implementation;
- keep implementation limited to the active change;
- update the active change's `tasks.md` only for work actually completed;
- run `go test ./...` after implementation;
- keep core, ports, use cases, adapters, and CLI boundaries intact;
- avoid changing docs, code, tests, and CI in one broad change unless the active change explicitly calls for it.

Contributor docs should reference the architecture rules without copying the full living architecture spec.

## Development Documentation

`docs/development.md` should cover local development basics:

```bash
go test ./...
go test -count=1 ./...
go build ./cmd/specharbor
go run ./cmd/specharbor help
```

It should explain that SpecHarbor's own CI is for this Go CLI repository. If the CI change is merged, document that it checks formatting and uncached Go tests according to the actual workflow. If the CI change is not merged, document only the current verified workflow behavior.

It must also state that user project scanning is intended to remain stack-agnostic. The fact that SpecHarbor itself is a Go project must not make scan documentation imply that user projects are expected to be Go projects.

## Parallel Agent and Branch Hygiene

Documentation should include practical warnings for running multiple coding agents in parallel:

- use separate branches or worktrees for separate OpenSpec changes;
- keep one active change id per branch where possible;
- do not have multiple agents edit the same files unless coordination is explicit;
- check `git status` before starting and before finalizing work;
- review diffs before marking tasks complete;
- avoid claiming scan/config/CI features are complete until the relevant branch is merged into the documentation branch;
- reconcile docs after merging parallel feature branches so command lists remain accurate.

This section should be in `docs/contributing.md` and may be summarized from the README.

## Status and Roadmap Language

Documentation must use conservative status labels:

- Implemented: only commands verified in the current branch.
- In progress: active or recently worked changes that are not merged into the documentation branch.
- Planned: product direction that has no verified implementation yet.

Do not use phrases that imply unfinished features are available. Avoid "supports" for a feature unless the command or behavior is implemented and verified. Prefer "planned", "intended", or "roadmap" for scan/config/provider capabilities that are not complete.

## Formatting and Tone

Documentation should be:

- Markdown-only;
- direct and practical;
- light on marketing language;
- oriented toward developers and contributors;
- consistent in command syntax;
- careful with implemented vs planned status;
- free of secrets, API keys, local absolute paths, and provider-specific credential examples.

Examples should use stable change ids such as `add-example-feature`.

## Verification Approach

After the documentation implementation:

- inspect the final docs for stale claims and duplicate command lists;
- verify implemented command syntax against the current CLI code or tests;
- verify scan/config/CI status against the current branch;
- run `go test ./...`;
- run `specharbor validate update-docs-and-readme` if the local binary or `go run` command can execute it from the repository root;
- update this change's `tasks.md` by checking only tasks completed during the documentation implementation.
