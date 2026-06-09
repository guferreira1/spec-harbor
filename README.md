# SpecHarbor

**SpecHarbor** is a Go CLI for OpenSpec-based AI coding-agent workflows.

It helps a team turn a loose idea into a scoped OpenSpec change, implementation tasks, an agent prompt, review checks, and an archive trail.

## Why

AI coding agents are useful when the task is explicit. They are risky when the work starts from a vague instruction, because they can:

- modify unrelated files;
- ignore project architecture;
- skip tests;
- invent requirements;
- create inconsistent implementations.

SpecHarbor keeps the work centered on a change package that both humans and agents can inspect:

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

SpecHarbor dogfoods this workflow for its own development. Meaningful changes to this repository should start from an OpenSpec change under `openspec/changes/`.

## Current Commands

Implemented commands on the current branch:

```bash
go run ./cmd/specharbor init
go run ./cmd/specharbor scan
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-feature --template feature
go run ./cmd/specharbor generate add-reports --guided --type feature --title "Add reports" --summary "Create report generation support"
go run ./cmd/specharbor validate add-example-feature
go run ./cmd/specharbor prompt add-example-feature --role implementer
go run ./cmd/specharbor review add-example-feature
go run ./cmd/specharbor archive add-example-feature
go run ./cmd/specharbor config show
go run ./cmd/specharbor config
```

`config` is a read-only alias for `config show`. It reads `.specharbor/config.yml` and prints a local config report when the project has a supported version `1` config file.

Built-in template generation supports exactly `feature`, `bugfix`, `docs`, and `refactor`. See [Usage](docs/usage.md) and [Generation modes](docs/generation-modes.md) for details.

Guided generation is deterministic, local, non-interactive, and supports exactly `feature`, `bugfix`, `docs`, and `refactor`. It uses explicit `--type`, `--title`, and `--summary` flags.

## Status

Implemented:

- OpenSpec project initialization.
- Stack-agnostic local project scanning.
- Blank OpenSpec change generation.
- Built-in OpenSpec change template generation for `feature`, `bugfix`, `docs`, and `refactor`.
- Guided OpenSpec change generation for `feature`, `bugfix`, `docs`, and `refactor`.
- Change validation, review, and archive.
- Role-based prompt generation with `--role`.
- Read-only local config display with `config show` and `config`.

In progress:

- Documentation updates are tracked through OpenSpec changes and should not describe unmerged behavior as available.

Planned:

- AI-assisted, agent-assisted, and hybrid spec generation.
- Future template capabilities such as custom, remote, and config-driven templates.
- Interactive generation prompts.
- Config mutation commands such as `config get`, `config set`, and `config unset`.
- AI-provider and workflow connector support.

## Local Development

Build the CLI from the repository root:

```bash
go build ./cmd/specharbor
```

Run tests:

```bash
go test ./...
```

## Docs

- [Usage](docs/usage.md)
- [Generation modes](docs/generation-modes.md)
- [Workflow](docs/workflow.md)
- [Agent roles](docs/agent-roles.md)
- [Contributing](docs/contributing.md)
- [Development](docs/development.md)

## License

MIT
