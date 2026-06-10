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
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature
go run ./cmd/specharbor generate add-reports --guided --type feature --title "Add reports" --summary "Create report generation support"
go run ./cmd/specharbor generate add-reports --ai-assisted --from-file agent-output.txt
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support"
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support" --execute
go run ./cmd/specharbor validate add-example-feature
go run ./cmd/specharbor prompt add-example-feature --role implementer
go run ./cmd/specharbor review add-example-feature
go run ./cmd/specharbor archive add-example-feature
go run ./cmd/specharbor config show
go run ./cmd/specharbor config
go run ./cmd/specharbor workflow
```

`config` is a read-only alias for `config show`. It reads `.specharbor/config.yml` and prints a local config report when the project has a supported version `1` config file.

`validate <change-id>` runs deterministic, local, read-only validation over the required OpenSpec change files, including content-quality rules with `error` and `warning` severities. Errors (missing or empty files, malformed task checkboxes, missing acceptance criteria) exit non-zero; warnings (placeholders, boilerplate-only starter content, missing recommended sections) never change the exit code, so a freshly generated blank change validates as valid with warnings. See [Usage](docs/usage.md) for the full rule list, exit codes, and example output.

The recommended workflow guide is available as `specharbor workflow`. It prints the nine-step advisory path: Spec Author Agent, Architecture Reviewer Agent, Implementer Agent, Test Engineer Agent, Change Reviewer Agent, Commit, Pull Request, Merge, and Archive. The command is read-only, does not execute command suggestions, does not commit, does not create PRs, does not merge, and does not call GitHub, GitLab, CI, provider APIs, agent CLIs, source-control automation, workflow execution, or remote automation.

Built-in template generation supports exactly `feature`, `bugfix`, `docs`, and `refactor`. See [Usage](docs/usage.md) and [Generation modes](docs/generation-modes.md) for details.

Custom template generation renders reusable, project-local templates from `.specharbor/templates/<template-name>/` with `--custom-template`. A custom template is a plain directory containing the five required OpenSpec change files; generation performs minimal deterministic variable substitution (`{{change_id}}`, plus `{{title}}`/`{{summary}}` when provided, with unresolved tokens left verbatim) and writes only under `openspec/changes/<change-id>/`, skipping existing files. Built-in and custom templates resolve from disjoint sources, so a custom template never shadows a built-in one. Templates are local and static: no remote templates, no config-driven registry, no template or script execution, and no network or provider calls. See [Usage](docs/usage.md) and [Generation modes](docs/generation-modes.md) for details.

Guided generation is deterministic, local, non-interactive, and supports exactly `feature`, `bugfix`, `docs`, and `refactor`. It uses explicit `--type`, `--title`, and `--summary` flags.

AI-assisted generation imports AI-authored OpenSpec Markdown from a local file only:

```bash
specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> [--overwrite]
```

The source file must contain strict `---FILE: <name>---` and `---END FILE---` blocks for exactly `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`. SpecHarbor parses the whole file before writing, rejects malformed or unsafe output, rejects symlink output targets, writes only under `openspec/changes/<change-id>/`, skips existing files by default, overwrites only with `--overwrite`, then runs validation. It does not call provider APIs or remote AI services, does not execute agents, does not modify production code, and does not run source-control automation.

Agent-assisted spec authoring supports exactly `feature`, `bugfix`, `docs`, and `refactor`. Dry-run remains the default and prints a deterministic authoring plan plus copy-pasteable prompt to stdout; it writes no files, does not execute agents, and does not resolve executable mappings. With explicit `--execute`, supported local agent commands run in run-and-report mode only. SpecHarbor captures stdout, stderr, exit code, and status, but does not parse or apply output, does not write files from output, does not modify production code from output, and does not auto-commit, auto-push, or auto-merge. See [Usage](docs/usage.md) and [Generation modes](docs/generation-modes.md) for details.

Recognized agent targets are Codex, Claude Code, Devin, Cursor, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, Aider, and `generic`. `generic` is dry-run-only until a future config-driven command mapping exists. Executable local mappings are `codex -> codex`, `claude -> claude`, `devin -> devin`, `cursor -> cursor`, `copilot -> copilot`, `gemini -> gemini`, `roo -> roo`, `windsurf -> windsurf`, and `aider -> aider`.

## Status

Implemented:

- OpenSpec project initialization.
- Stack-agnostic local project scanning.
- Blank OpenSpec change generation.
- Built-in OpenSpec change template generation for `feature`, `bugfix`, `docs`, and `refactor`.
- Project-local custom template generation from `.specharbor/templates/<template-name>/` with deterministic variable substitution and OpenSpec-only writes.
- Guided OpenSpec change generation for `feature`, `bugfix`, `docs`, and `refactor`.
- AI-assisted from-file OpenSpec generation with strict local delimiter blocks, default skip behavior, explicit overwrite, validation integration, and no provider/API/runner/source-control automation.
- Dry-run agent-assisted spec authoring for `feature`, `bugfix`, `docs`, and `refactor`.
- Explicit agent-assisted local runner execution in run-and-report mode for supported concrete agent targets.
- Deterministic, local, read-only change validation with content-quality rules, error/warning severities, and grouped reports: errors (missing/empty/unusable files, malformed task checkboxes, missing acceptance criteria) exit non-zero, while quality warnings (placeholders, boilerplate-only starter content, missing recommended sections or mitigations) keep exit code `0`, so freshly generated blank changes stay valid.
- Change review and archive.
- Role-based prompt generation with `--role`.
- Read-only local config display with `config show` and `config`.
- Read-only advisory workflow guide with `workflow`.

In progress:

- Documentation updates are tracked through OpenSpec changes and should not describe unmerged behavior as available.

Planned:

- Hybrid spec generation.
- Future config-driven generic runner mappings and richer templates.
- Future template capabilities such as remote and config-driven templates.
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
