# SpecHarbor

**SpecHarbor** is a Go CLI for OpenSpec-based AI coding-agent workflows.

It helps teams convert a loose idea into a scoped change package that humans and agents can execute with guardrails.

## Português

Português: [docs/pt-BR/README.md](docs/pt-BR/README.md)

## What is SpecHarbor

SpecHarbor is a CLI that supports the OpenSpec/SDD workflow used by coding agents.
It does not replace implementation planning, implementation itself, or source-control workflows.
Instead, it prepares structure, context, prompts, and checks so work stays explicit and reviewable.

## Why it exists

AI agents perform better when instructions are explicit. Vague prompts often lead to:

- unrelated file edits
- architecture drift
- skipped tests
- invented requirements
- inconsistent implementations

SpecHarbor reduces this risk by centering work on explicit OpenSpec change packages and explicit workflow handoffs between planning, authoring, implementation, validation, review, and archiving.

## Who it is for

- developers using Codex, Claude Code, Cursor, Devin, GitHub Copilot, Gemini CLI, Roo Code, Windsurf, Aider, or generic agents;
- teams adopting OpenSpec/SDD and explicit agent handoffs;
- maintainers who want auditable AI-assisted development;
- solo developers who want safer agent workflows.

## Core workflow

Core flow:

```text
Idea -> OpenSpec change -> tasks.md -> Agent prompt -> Implementation -> Review -> Archive
```

Command flow for the same sequence:

```text
specharbor generate <change-id> ... -> specharbor validate <change-id> -> specharbor prompt <change-id> --role <role> -> specharbor review <change-id> -> specharbor archive <change-id>
```

`specharbor workflow` prints the advisory nine-step workflow used by SpecHarbor and clarifies where each role fits.

## Key concepts

- **OpenSpec change**
  Every concrete work item is an OpenSpec change package under `openspec/changes/<change-id>/` with five required files:
  `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, `risks.md`.

- **`tasks.md`**
  Implementation checklist used to track scope and completion. Validation only checks structure and content quality; it does not execute the implementation for you.

- **agent roles**
  Role prompts are generated for `spec-author`, `architecture-reviewer`, `implementer`, `test-engineer`, and `change-reviewer`.

- **project brief**
  `specharbor brief` stores user-confirmed context in `.specharbor/project-brief.md` with explicit confirmation.

- **context discovery**
  `specharbor context discover` reads bounded local sources and returns classified evidence:
  user-confirmed context, detected facts, suggested assumptions, sources, and confidence.

- **repository context index**
  `specharbor context index` produces metadata-only local inventory in `.specharbor/context-index.json` for supported context sources.
  It stores no raw file contents and is ignored by source control.

- **validation**
  `specharbor validate <change-id>` enforces required files, readable structure, and content quality signals.
  Errors fail the command; warnings do not fail by default.

- **review**
  `specharbor review <change-id>` checks task completion state and change usability before review and archive.

- **archive**
  `specharbor archive <change-id>` moves accepted changes into a dated archive directory under `openspec/archive/`.

## Quickstart

From the repository root:

```bash
specharbor init
specharbor generate add-login-feature --guided --type feature --title "Add login feature" --summary "Add a secure login flow"
specharbor validate add-login-feature
specharbor prompt add-login-feature --role implementer
specharbor review add-login-feature
specharbor archive add-login-feature
```

Optional context preparation:

```bash
specharbor context discover
specharbor context index --write
specharbor brief
```

Move to `docs/usage.md` for the full command reference.

## Installation

Supported channels:

- **install.sh**
  `curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh`
- **npm**
  `npm install -g specharbor`
- **Homebrew**
  `brew install guferreira1/tap/specharbor`
- **GitHub Releases**
  manual download and checksum verification
- **go install** (fallback/developer option)
  `go install github.com/guferreira1/spec-harbor/cmd/specharbor@latest`

Use `specharbor version` to confirm installation.
See [docs/install.md](docs/install.md) for platform support, checksum steps, and troubleshooting.

## Current capabilities

- OpenSpec project initialization (`specharbor init`)
- context discovery and repository context index (`context discover`, `context index`)
- context collection and update (`brief`, `brief --update`)
- change generation (blank, guided, templates, custom templates, config templates, hybrid, AI-assisted from-file, agent-assisted spec authoring)
- change validation (`validate`)
- role prompt generation (`prompt --role`)
- change review (`review`)
- change archive (`archive`)
- advisory workflow guidance (`workflow`)
- project brief and config report (`project brief`, `config`)
- release/install channels (GitHub Releases, install.sh, npm wrapper, Homebrew)

## Implementation details that matter

### Agent-assisted spec authoring

Dry-run remains the default for agent-assisted mode.

Supported command mappings are:

- `codex -> codex`
- `claude -> claude`
- `devin -> devin`
- `cursor -> cursor`
- `copilot -> copilot`
- `gemini -> gemini`
- `roo -> roo`
- `windsurf -> windsurf`
- `aider -> aider`

The generic fallback key is `generic`.
`--agent generic` is currently supported only in dry-run mode.

Execute mode runs agent commands in run-and-report mode.
In this mode, SpecHarbor does not parse or apply output, does not write files from output, does not modify production code, and does not auto-commit, auto-push, or auto-merge.

### AI-assisted from-file generation

AI-assisted mode uses strict local delimiter blocks with `--ai-assisted --from-file` and optional `--overwrite`.

Expected output files are `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
The parser is local, strict, and deterministic.
SpecHarbor validates symlink output targets before writing.
No provider APIs or remote AI services are called, and no source-control automation is run.

### Workflow command safety

`specharbor workflow` prints the advisory nine-step sequence:

- Spec Author Agent
- Architecture Reviewer Agent
- Implementer Agent
- Test Engineer Agent
- Change Reviewer Agent
- Commit
- Pull Request
- Merge
- Archive

It is read-only, advisory, and does not execute recommendations or commands.
It does not commit, create PRs, merge, run CI, or perform source-control automation.
It does not create PRs.
It does not merge.
It does not call GitHub, GitLab, CI, provider APIs, agent CLIs, source-control automation, workflow execution, or remote automation.

## Safety model

- local-first behavior for discovery, indexing, generation, validation, review, and prompts;
- explicit confirmation where applicable (for example, `brief` and `brief --update`);
- no auto-commit, auto-push, automatic PR creation, merge, or archive automation;
- no production code is modified during generation, discovery, validation, prompting, reviewing, or indexing;
- no provider API calls or local model calls in current workflows;
- deterministic path and symlink safety; unsafe and generated paths are rejected;
- no command execution from user prompts and no shell string construction in npm forwarding;
- context is separated into user-confirmed context, detected facts, and assumptions;
- release/install behavior is explicit and checksum-verified.

## Version and distribution scope

`specharbor version` shows version metadata from the release or local fallback build.
Source builds without injected release metadata commonly print `SpecHarbor dev`.

Current release conventions include:

- Git tags: `v0.1.0`
- Release binary metadata: `0.1.0`
- GitHub Release assets built with GoReleaser
- checksums in `checksums.txt` consumed by supported install channels

Supported distribution paths are:

- GitHub Releases
- install.sh (from official release assets)
- npm wrapper package
- Homebrew tap
- `go install` fallback

Native Linux package managers and broader package-manager automation are future work.

## Documentation

- [Install](docs/install.md)
- [Usage](docs/usage.md)
- [Workflow](docs/workflow.md)
- [Agent roles](docs/agent-roles.md)
- [Generation modes](docs/generation-modes.md)
- [Release metadata](docs/release.md)
- [Contributing](docs/contributing.md)

## Status and roadmap

Current release behavior is documented as implemented in this repository, including:

- OpenSpec initialization and structured change workflow
- context discovery and repository context indexing
- multiple generation modes
- validation and review
- npm wrapper and install channels
- archive flow

Planned features are documented in planning artifacts and not yet implemented:

- native Linux package formats (`.deb` / `.rpm`);
- Windows package-manager integration;
- config mutation commands (`config get`, `config set`, `config unset`);
- generic runner mapping for agent execution;
- broader AI-provider and workflow connector automation.

See [docs/planning/context-initiative-remaining-plan.md](docs/planning/context-initiative-remaining-plan.md) for current context initiative planning.

## Contributing

- Start meaningful changes with an OpenSpec change package under `openspec/changes/<change-id>/`.
- Keep changes scoped to the active change and track behavior in `tasks.md`.
- Prefer the OpenSpec/SDD workflow before implementation.
- Run repository tests and validation commands after docs or behavior edits.
- Preserve architecture boundaries:
  `internal/core/domain`, `internal/core/ports`, `internal/core/usecase`, and `internal/adapters`.

See [docs/contributing.md](docs/contributing.md).

## Local checks

```bash
go build ./cmd/specharbor
go test ./...
```

## License

MIT
