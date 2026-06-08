# Usage

Run examples from the repository root while developing SpecHarbor:

```bash
go run ./cmd/specharbor help
```

You can also build a local binary:

```bash
go build ./cmd/specharbor
```

That creates a local `specharbor` binary in the repository root.

## Implemented Commands

### Initialize a Project

```bash
go run ./cmd/specharbor init
```

`init` creates the OpenSpec and SpecHarbor project files that are missing in the current directory. Existing files are skipped.

### Scan a Project

```bash
go run ./cmd/specharbor scan
```

`scan` performs an informational local project scan. It is stack-agnostic: user projects do not need to be Go projects.

The report can include detected ecosystems, package managers, test command hints, CI signals, container or deployment signals, SpecHarbor/OpenSpec signals, and notes. The command does not require flags or arguments.

### Generate a Change

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

`generate <change-id> --blank` creates the expected OpenSpec change structure with blank/manual starter content.

Built-in template generation uses the same command with `--template <template-name>`:

```bash
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

For example:

```bash
go run ./cmd/specharbor generate add-example-feature --template feature
```

Supported built-in templates are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Both blank and built-in template generation create the same required OpenSpec change files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Built-in template content is deterministic, local, and generic starter content. It is safe to edit after generation, and it does not mean SpecHarbor inferred project-specific requirements.

Existing files are skipped and are not overwritten. If a change directory partially exists, running generation again recovers it by creating only the missing required files.

Copy-pasteable examples from the repository root:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-example-feature --template feature
go run ./cmd/specharbor generate fix-example-bug --template bugfix
go run ./cmd/specharbor generate update-example-docs --template docs
go run ./cmd/specharbor generate refactor-example-flow --template refactor
```

### Validate a Change

```bash
go run ./cmd/specharbor validate add-example-feature
```

`validate <change-id>` checks the change package under `openspec/changes/<change-id>/` for the required OpenSpec files.

### Generate a Role Prompt

```bash
go run ./cmd/specharbor prompt add-example-feature --role implementer
```

`prompt <change-id> --role <role>` prints an agent prompt for an existing OpenSpec change.

Supported roles:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

Use `--role` for prompt roles. Agent-target flags are not implemented.

### Review a Change

```bash
go run ./cmd/specharbor review add-example-feature
```

`review <change-id>` reviews the change package and task completion state. It exits non-zero when the review status is not approved.

### Archive a Change

```bash
go run ./cmd/specharbor archive add-example-feature
```

`archive <change-id>` moves a completed change from `openspec/changes/<change-id>/` to a dated archive path under `openspec/archive/<date>/<change-id>/`.

### Show Local Config

```bash
go run ./cmd/specharbor config show
go run ./cmd/specharbor config
```

`config show` and `config` are read-only. They read `.specharbor/config.yml` from the current project, support config version `1`, and print the local config report.

Current config behavior does not write config files and does not implement:

- `config get`
- `config set`
- `config unset`

## Normal Flow

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

A typical local sequence is:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor validate add-example-feature
go run ./cmd/specharbor prompt add-example-feature --role implementer
go run ./cmd/specharbor review add-example-feature
go run ./cmd/specharbor archive add-example-feature
```

Do not archive a change until the implementation is complete and reviewed.

## Planned Behavior

The following items are product direction, not implemented command behavior:

- guided generation;
- AI-assisted generation;
- agent-assisted generation;
- hybrid generation;
- custom, remote, and config-driven templates;
- interactive generation prompts;
- AI provider setup;
- provider API key management;
- config mutation commands;
- external workflow connectors.
