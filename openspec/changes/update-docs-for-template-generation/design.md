# Design: Update Docs For Template Generation

## Overview

This is a documentation-only change. It updates public documentation after built-in template generation has been implemented.

The documentation should present two implemented generation paths:

```bash
go run ./cmd/specharbor generate <change-id> --blank
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

Template generation must be described as a local deterministic starter-content feature, not as guided, AI-assisted, agent-assisted, custom, remote, or config-driven generation.

## Documentation Source Of Truth

The documentation update must be based on the implemented `implement-template-generation` behavior:

- `--blank` creates the standard OpenSpec change files with blank/manual content.
- `--template <template-name>` creates the standard OpenSpec change files with starter content tailored to a selected built-in template.
- The only implemented built-in templates are `feature`, `bugfix`, `docs`, and `refactor`.
- The standard required files are:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- Template content is deterministic, local, generic starter content.
- Existing files are skipped and never overwritten.
- Partially existing change directories are recoverable because generation creates only missing required files.

The docs must not infer additional behavior from the architecture roadmap. `openspec/project.md` and the architecture spec mention broader generation strategies, but this documentation update must distinguish current behavior from future product direction.

## Files To Update Later

The future implementation step may update only:

- `README.md`
- Markdown files under `docs/`
- `openspec/changes/update-docs-for-template-generation/tasks.md`

Expected documentation touch points:

- `README.md`: update the command list and status sections if they still show template generation as planned.
- `docs/usage.md`: add command-level usage for `generate <change-id> --template <template-name>`.
- `docs/generation-modes.md`: move built-in template generation from planned behavior to implemented behavior.

Other Markdown files under `docs/` may be updated only if needed to keep command examples and roadmap language consistent.

## README Guidance

The README should stay concise. It should include copy-pasteable examples from the repository root, such as:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-example-feature --template feature
```

If the README lists implemented commands, it should include template generation alongside blank generation. If it lists planned behavior, template generation must be removed from that planned list or narrowed to future template capabilities that are not implemented.

The README should not become the full generation reference. Detailed mode behavior belongs in `docs/usage.md` and `docs/generation-modes.md`.

## Usage Documentation Guidance

`docs/usage.md` should document:

- how to run template generation from the repository root;
- the command shape `generate <change-id> --template <template-name>`;
- the four supported built-in templates;
- the generated OpenSpec file structure;
- deterministic local generic starter content;
- no-overwrite behavior for existing files;
- partial-directory recovery by creating only missing required files;
- continued support for `generate <change-id> --blank`.

Examples must use `go run ./cmd/specharbor ...`, not installed binary-only examples.

Recommended example set:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-example-feature --template feature
go run ./cmd/specharbor generate fix-example-bug --template bugfix
go run ./cmd/specharbor generate update-example-docs --template docs
go run ./cmd/specharbor generate refactor-example-flow --template refactor
```

The usage docs must not claim custom template paths, remote templates, guided prompts, provider APIs, agent-assisted authoring, or config-backed template selection are implemented.

## Generation Modes Guidance

`docs/generation-modes.md` should separate implemented behavior from planned behavior.

Implemented:

- blank generation;
- built-in template generation for `feature`, `bugfix`, `docs`, and `refactor`.

Planned:

- guided generation;
- AI-assisted generation;
- agent-assisted generation;
- hybrid generation;
- custom templates;
- remote templates;
- config-driven templates;
- interactive prompts.

If future template capabilities are mentioned, they must be clearly labeled as planned and separate from the implemented built-in template mode.

## Verification Strategy

Because this is documentation-only, verification focuses on scope and accuracy:

- inspect the final diff and confirm only allowed Markdown files changed;
- confirm no Go code, Go tests, CI, config, or init templates changed;
- confirm command examples are copy-pasteable from the repository root;
- confirm planned roadmap items are not mixed with implemented behavior;
- run `go test ./...`;
- run `go run ./cmd/specharbor validate update-docs-for-template-generation`.

No `gofmt` step is required unless Go files are unexpectedly modified. Go files must not be modified by this change.
