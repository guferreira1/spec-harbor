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

## Install

SpecHarbor `v0.1.0` is available through GitHub Releases, `install.sh`,
npm, and Homebrew. See [Install](docs/install.md) for manual downloads,
checksum verification, platform support, and troubleshooting.

Quick install options:

```bash
curl -fsSL https://raw.githubusercontent.com/guferreira1/spec-harbor/main/install.sh | sh
npm install -g specharbor
brew install guferreira1/tap/specharbor
```

Verify any installed release binary with:

```bash
specharbor version
```

The current public release prints `SpecHarbor 0.1.0` on the first line.

Manual GitHub Release downloads are available from the
[releases page](https://github.com/guferreira1/spec-harbor/releases); verify
the archive against `checksums.txt` before installing. `go install` remains a
source-build fallback and prints development fallback metadata unless release
metadata is injected manually. Native Linux packages, Windows package managers,
signing, SBOMs, Docker images, and package publishing automation remain
future work.

## Current Commands

Implemented commands on the current branch:

```bash
go run ./cmd/specharbor version
go run ./cmd/specharbor init
go run ./cmd/specharbor scan
go run ./cmd/specharbor context discover
go run ./cmd/specharbor brief
go run ./cmd/specharbor brief --update
go run ./cmd/specharbor generate add-example-feature --interactive
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-feature --template feature
go run ./cmd/specharbor generate add-payment-flow --custom-template api-feature
go run ./cmd/specharbor generate add-payment-flow --config-template api-feature
go run ./cmd/specharbor generate add-login --hybrid --template feature --title "Add login" --summary "Add an OpenSpec change for login"
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

`specharbor version` prints deterministic build metadata:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

Plain `go install` without `-ldflags` uses the same development fallback metadata. An installed binary built that way is expected to print:

```text
SpecHarbor dev
commit: unknown
date: unknown
dirty: unknown
```

This is expected behavior. `dev` means no release version was injected.
`unknown` means the build did not provide that metadata field. Git release
tags use `vX.Y.Z`, such as `v0.1.0`, while release binaries display plain
`X.Y.Z`, such as `0.1.0`. Runtime displays the injected version string as-is
and does not inspect Git tags. See [Install](docs/install.md) for install
channels and [Release metadata](docs/release.md) for release assets,
checksums, and build metadata.

`config` is a read-only alias for `config show`. It reads `.specharbor/config.yml` and prints a local config report when the project has a supported version `1` config file.

`validate <change-id>` runs deterministic, local, read-only validation over the required OpenSpec change files, including content-quality rules with `error` and `warning` severities. Errors (missing or empty files, malformed task checkboxes, missing acceptance criteria) exit non-zero; warnings (placeholders, boilerplate-only starter content, missing recommended sections) never change the exit code, so a freshly generated blank change validates as valid with warnings. See [Usage](docs/usage.md) for the full rule list, exit codes, and example output.

The recommended workflow guide is available as `specharbor workflow`. It prints the nine-step advisory path: Spec Author Agent, Architecture Reviewer Agent, Implementer Agent, Test Engineer Agent, Change Reviewer Agent, Commit, Pull Request, Merge, and Archive. The command is read-only, does not execute command suggestions, does not commit, does not create PRs, does not merge, and does not call GitHub, GitLab, CI, provider APIs, agent CLIs, source-control automation, workflow execution, or remote automation.

Structured local context discovery is available as:

```bash
specharbor context discover
```

`context discover` inspects a bounded set of common repository files and directories, then prints user-confirmed context, detected facts, suggested assumptions, source evidence, and confidence levels. It is local, offline, read-only, skips sensitive files and generated folders, avoids symlink traversal, does not execute project commands, and does not perform indexing, RAG, retrieval, prompt injection, provider calls, remote discovery, or source-control automation. Existing `scan` behavior remains a shallow presence report, while `brief` remains the confirmation-first workflow for writing `.specharbor/project-brief.md`.

Interactive project briefing is available as:

```bash
specharbor brief
```

`brief` requires an interactive TTY, asks multiple-choice questions for explicit project context, and writes `.specharbor/project-brief.md` only after confirmation. It refuses to overwrite an existing project brief. The brief separates user-provided answers, detected context, and assumptions, and this command does not run repository indexing, RAG, provider APIs, agent execution, source-control automation, release, or publishing behavior.

Existing briefs can be reviewed and updated explicitly with:

```bash
specharbor brief --update
```

The update flow requires an interactive TTY and an existing `.specharbor/project-brief.md`. It parses known brief fields, reuses local `context discover` results as detected facts and suggested assumptions, shows conflicts and stale records, prints a reviewable preview, and writes only after final confirmation. Existing user-confirmed values are kept by default. Detected facts and suggested assumptions are never promoted into confirmed context unless the user explicitly selects them and confirms the final write. Cancellation leaves the existing brief unchanged.

Role prompt generation is context-aware for the supported `prompt --role` roles. When confirmed or discovered local context is available, generated prompts include a bounded `## Project Context` section with user-confirmed context, detected facts, and suggested assumptions kept separate. Confirmed project brief values take precedence over detected facts; assumptions stay labeled as assumptions and include source/confidence evidence when rendered.

Interactive generation is available as:

```bash
specharbor generate <change-id> --interactive
```

`<change-id>` remains required on the command line. Interactive mode requires a TTY, guides the user through one supported path (`blank`, built-in template, custom template, config template, or hybrid), validates answers with the same value objects and generation rules used by direct flags, prints a deterministic pre-write summary, and asks for confirmation before delegating to existing generation behavior. The summary includes `Validation: automatic no` for blank, built-in template, custom template, and config template paths, and `Validation: automatic yes` for hybrid paths. It also prints deterministic safety notes: writes stay limited to OpenSpec change files, production code is not modified, source-control and workflow commands are not run, provider/LLM/agent APIs are not called, and no auto-commit, auto-push, PR creation, merge, or archive is performed.

Interactive prompts do not offer direct guided generation, AI-assisted generation, agent-assisted generation, local agent execution, or raw remote URL entry in this version. Remote templates remain reachable only through existing config aliases and only after confirmation.

Built-in template generation supports exactly `feature`, `bugfix`, `docs`, and `refactor`. See [Usage](docs/usage.md) and [Generation modes](docs/generation-modes.md) for details.

Custom template generation renders reusable, project-local templates from `.specharbor/templates/<template-name>/` with `--custom-template`. A custom template is a plain directory containing the five required OpenSpec change files; generation performs minimal deterministic variable substitution (`{{change_id}}`, plus `{{title}}`/`{{summary}}` when provided, with unresolved tokens left verbatim) and writes only under `openspec/changes/<change-id>/`, skipping existing files. Built-in and custom templates resolve from disjoint sources, so a custom template never shadows a built-in one. See [Usage](docs/usage.md) and [Generation modes](docs/generation-modes.md) for details.

Config-driven template generation uses project-owned aliases from `.specharbor/config.yml`:

```bash
specharbor generate <change-id> --config-template <alias> [--title "<title>"] [--summary "<summary>"]
```

Aliases live under `version: 1` config as `templates.aliases.<alias>` and resolve to `source: builtin`, `source: custom`, or `source: remote`. Remote aliases are available only through `--config-template`; there is no `--remote-template` flag.

```yaml
version: 1

templates:
  aliases:
    service-feature:
      source: remote
      url: https://example.com/specharbor/templates/service-feature.zip
      checksum: sha256:<64-hex>
      format: zip
```

Remote templates require HTTPS URLs with no credentials, query strings, fragments, redirects, auth headers, cookies, OAuth, or environment token expansion. The required `sha256` checksum is verified over the downloaded ZIP bytes before archive parsing. ZIP bundles must contain exactly `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md` as non-empty root-level files. This first version has no persistent cache, marketplace discovery, git clone, provider APIs, script/shell execution, production code writes, source-control automation, auto-commit, PR, merge, or archive automation.

`--template`, `--custom-template`, and `--config-template` are disjoint namespaces; the same name can exist in all three without shadowing or fallback.

Hybrid generation composes exactly one deterministic template source with required guided metadata, writes only the five OpenSpec change files, skips existing files, and then runs validation:

```bash
specharbor generate <change-id> --hybrid --template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
specharbor generate <change-id> --hybrid --custom-template <name> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
specharbor generate <change-id> --hybrid --config-template <alias> --title "<title>" --summary "<summary>" [--type <feature|bugfix|docs|refactor>]
```

Hybrid requires exactly one of `--template`, `--custom-template`, or `--config-template`. Direct built-in sources and config aliases resolving to built-ins derive type from the built-in template when `--type` is omitted, and a provided type must match that built-in template. Custom, config custom, and config remote sources do not infer type; omitted `{{type}}` tokens remain verbatim unless `--type` is provided. Remote templates remain available only through config aliases and keep the HTTPS, checksum, ZIP, no-credential, no-redirect, no-cache, no-script, and OpenSpec-only write safeguards. Hybrid intentionally has no AI overlay, no `--blank`, no `--from-file`, no `--overwrite`, no `--agent`, no `--execute`, no provider or LLM API calls, no shell or script execution, no production code writes, and no auto-commit, auto-push, PR, merge, or archive automation.

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
- Deterministic local context discovery with classified facts, assumptions, confirmed context, confidence, and source evidence.
- Interactive project briefing that writes `.specharbor/project-brief.md` after confirmation.
- Confirmation-first project brief update with conflict/stale review and safe writes.
- Interactive generation prompts for blank, built-in template, custom template, config template, and hybrid paths.
- Blank OpenSpec change generation.
- Built-in OpenSpec change template generation for `feature`, `bugfix`, `docs`, and `refactor`.
- Project-local custom template generation from `.specharbor/templates/<template-name>/` with deterministic variable substitution and OpenSpec-only writes.
- Config-driven template aliases for built-in, project-local custom, and pinned HTTPS remote ZIP templates.
- Hybrid generation from one built-in, custom, or config-template source with required title and summary metadata, optional type metadata, validation integration, and OpenSpec-only writes.
- Guided OpenSpec change generation for `feature`, `bugfix`, `docs`, and `refactor`.
- AI-assisted from-file OpenSpec generation with strict local delimiter blocks, default skip behavior, explicit overwrite, validation integration, and no provider/API/runner/source-control automation.
- Dry-run agent-assisted spec authoring for `feature`, `bugfix`, `docs`, and `refactor`.
- Explicit agent-assisted local runner execution in run-and-report mode for supported concrete agent targets.
- Deterministic, local, read-only change validation with content-quality rules, error/warning severities, and grouped reports: errors (missing/empty/unusable files, malformed task checkboxes, missing acceptance criteria) exit non-zero, while quality warnings (placeholders, boilerplate-only starter content, missing recommended sections or mitigations) keep exit code `0`, so freshly generated blank changes stay valid.
- Change review and archive.
- Context-aware role-based prompt generation with `--role`.
- Read-only local config display with `config show` and `config`.
- Read-only advisory workflow guide with `workflow`.
- Deterministic version metadata reporting with `version`.
- Tag-based GoReleaser GitHub Release assets with SHA-256 checksums.
- Checksum-verified `install.sh` installer for Linux and macOS.
- Published npm wrapper package under `packages/npm/specharbor/`.
- External Homebrew tap formula in `guferreira1/homebrew-tap`.

In progress:

- Documentation updates are tracked through OpenSpec changes and should not describe unmerged behavior as available.

Planned:

- Native Linux packages (`.deb`/`.rpm`) and Windows package managers (Scoop, Winget).
- Future config-driven generic runner mappings and richer templates.
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
- [Install](docs/install.md)
- [Release metadata](docs/release.md)
- [Generation modes](docs/generation-modes.md)
- [Workflow](docs/workflow.md)
- [Agent roles](docs/agent-roles.md)
- [Contributing](docs/contributing.md)
- [Development](docs/development.md)

## License

MIT
