# Design: Update Docs For Guided Generation

## Overview

This is a documentation-only change. It updates public documentation after guided generation has been implemented.

The documentation should present three implemented generation paths:

```bash
go run ./cmd/specharbor generate <change-id> --blank
go run ./cmd/specharbor generate <change-id> --template <template-name>
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

Guided generation must be described as deterministic, local, and non-interactive. The mode is guided by explicit CLI flags and does not prompt during command execution.

The docs must not present AI-assisted generation, agent-assisted generation, custom templates, remote templates, config-driven templates, hybrid generation, or interactive prompts as implemented behavior.

## Documentation Source Of Truth

The documentation update must be based on the implemented `implement-guided-generation` behavior:

- `--blank` creates the standard OpenSpec change files with blank/manual content.
- `--template <template-name>` creates the standard OpenSpec change files with built-in template starter content.
- `--guided --type <type> --title "<title>" --summary "<summary>"` creates the standard OpenSpec change files with deterministic guided starter content.
- The only supported guided types are `feature`, `bugfix`, `docs`, and `refactor`.
- Guided generation is deterministic, local, and non-interactive.
- Guided generation uses explicit CLI flags and does not prompt during command execution.
- Guided generated content includes the supplied title and summary.
- The standard required files are:
  - `proposal.md`
  - `design.md`
  - `tasks.md`
  - `acceptance-criteria.md`
  - `risks.md`
- Existing files are skipped and never overwritten.
- Partially existing change directories are recoverable because generation creates only missing required files.

The docs must not infer implemented behavior from the architecture roadmap. `openspec/project.md` and the architecture spec mention broader generation strategies, but this documentation update must distinguish current command behavior from future product direction.

## Files To Update Later

The future implementation step may update only:

- `README.md`
- Markdown files under `docs/`
- `openspec/changes/update-docs-for-guided-generation/tasks.md`

Expected documentation touch points:

- `README.md`: update the command list and status sections if they still show guided generation as planned.
- `docs/usage.md`: add command-level usage for `generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`.
- `docs/generation-modes.md`: move guided generation from planned behavior to implemented behavior.

Other Markdown files under `docs/` may be updated only if needed to keep command examples and roadmap language consistent.

## README Guidance

The README should stay concise. It should include copy-pasteable examples from the repository root, such as:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-feature --template feature
go run ./cmd/specharbor generate add-guided-feature --guided --type feature --title "Add guided feature" --summary "Create a guided OpenSpec change from explicit CLI inputs."
```

If the README lists implemented commands, it should include guided generation alongside blank and built-in template generation. If it lists planned behavior, guided generation must be removed from that planned list.

The README should not become the full generation reference. Detailed mode behavior belongs in `docs/usage.md` and `docs/generation-modes.md`.

## Usage Documentation Guidance

`docs/usage.md` should document:

- how to run guided generation from the repository root;
- the command shape `generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"`;
- the four supported guided types: `feature`, `bugfix`, `docs`, and `refactor`;
- that guided generation is deterministic, local, and non-interactive;
- that guided generation uses explicit CLI flags and does not prompt during command execution;
- that guided generated content includes the supplied title and summary;
- the generated OpenSpec file structure;
- no-overwrite behavior for existing files;
- partial-directory recovery by creating only missing required files;
- continued support for `generate <change-id> --blank`;
- continued support for `generate <change-id> --template <template-name>`.

Examples must use `go run ./cmd/specharbor ...`, not installed binary-only examples.

Recommended example set:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-example-feature --template feature
go run ./cmd/specharbor generate add-guided-feature --guided --type feature --title "Add guided feature" --summary "Create a guided OpenSpec change from explicit CLI inputs."
go run ./cmd/specharbor generate fix-guided-bug --guided --type bugfix --title "Fix guided bug" --summary "Describe the bugfix using deterministic guided starter content."
go run ./cmd/specharbor generate update-guided-docs --guided --type docs --title "Update guided docs" --summary "Document guided generation as implemented behavior."
go run ./cmd/specharbor generate refactor-guided-flow --guided --type refactor --title "Refactor guided flow" --summary "Describe a behavior-preserving refactor with explicit context."
```

The usage docs must not claim AI-assisted generation, agent-assisted generation, custom template paths, remote templates, config-backed template selection, hybrid generation, or interactive prompts are implemented.

## Generation Modes Guidance

`docs/generation-modes.md` should separate implemented behavior from planned behavior.

Implemented:

- blank generation;
- built-in template generation for `feature`, `bugfix`, `docs`, and `refactor`;
- guided generation for `feature`, `bugfix`, `docs`, and `refactor`.

Planned:

- AI-assisted generation;
- agent-assisted generation;
- hybrid generation;
- custom templates;
- remote templates;
- config-driven templates;
- interactive prompts.

If future guided or template capabilities are mentioned, they must be clearly labeled as planned and separate from the implemented non-interactive guided mode.

## Verification Strategy

Because this is documentation-only, verification focuses on scope and accuracy:

- inspect the final diff and confirm only allowed Markdown files changed;
- confirm no Go code, Go tests, CI, config, or init templates changed;
- confirm command examples are copy-pasteable from the repository root;
- confirm blank, built-in template, and guided generation are all documented as implemented;
- confirm the guided command syntax uses `--guided --type <type> --title "<title>" --summary "<summary>"`;
- confirm supported guided types are exactly `feature`, `bugfix`, `docs`, and `refactor`;
- confirm planned roadmap items are not mixed with implemented behavior;
- run `go test ./...`;
- run `go run ./cmd/specharbor validate update-docs-for-guided-generation`.

No `gofmt` step is required unless Go files are unexpectedly modified. Go files must not be modified by this change.
