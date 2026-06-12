# Proposal: Implement Context Discovery

## Problem

SpecHarbor can run a shallow `scan` command and can collect explicit project context with `specharbor brief`, but it still does not have a deterministic local workflow that inspects common repository files and produces structured context signals for briefing and future context-aware workflows.

The current `brief` foundation intentionally records no detected context unless a future change adds shallow suggestions. That keeps confirmed context safe, but it means users must manually answer questions even when a repository already contains clear evidence in files such as `README.md`, `go.mod`, `package.json`, `openspec/project.md`, `.specharbor/rules/`, or `.github/workflows/`.

SpecHarbor needs a local/offline discovery foundation that can read common project files, separate facts from assumptions, prefer confirmed project brief data when available, and present suggestions without silently converting repository hints into confirmed project decisions.

## Goal

Add the first local project context discovery capability:

```text
specharbor context discover
```

The command should inspect a bounded set of common repository files and directories, then print a deterministic human-readable summary of detected project context signals.

The same discovery use case may also be used by:

```text
specharbor brief
```

When used by `brief`, discovered values are suggestions only. The user must still confirm answers before `.specharbor/project-brief.md` is written, and the existing write-if-absent behavior must be preserved.

## Scope

- Add a new `specharbor context discover` command.
- Keep existing `specharbor scan` behavior unchanged as the shallow presence report.
- Preserve existing `specharbor brief` behavior:
  - no positional arguments or flags;
  - interactive TTY required;
  - explicit confirmation before writing;
  - `.specharbor/project-brief.md` written only when absent;
  - no merge, update, overwrite, or append behavior.
- Add deterministic local discovery for these files and directories when present:
  - `AGENTS.md`
  - `README.md`
  - `CONTRIBUTING.md`
  - `docs/`
  - `openspec/project.md`
  - `openspec/specs/`
  - `.specharbor/rules/`
  - `.specharbor/project-brief.md`
  - `package.json`
  - `go.mod`
  - `pom.xml`
  - `build.gradle`
  - `Cargo.toml`
  - `pyproject.toml`
  - `requirements.txt`
  - `Dockerfile`
  - `docker-compose.yml`
  - `Makefile`
  - `Taskfile.yml`
  - `.github/workflows/`
- Produce structured context signals for:
  - project type;
  - purpose summary when clearly available;
  - stack, languages, and frameworks;
  - architecture hints;
  - package manager;
  - test command;
  - build command;
  - run command;
  - documentation sources;
  - agent instruction sources;
  - OpenSpec sources.
- Assign each context signal a classification:
  - `detected_fact`
  - `suggested_assumption`
  - `user_confirmed_context`
- Assign each signal a confidence level and source evidence.
- Prefer `.specharbor/project-brief.md` as `user_confirmed_context` when it exists.
- Treat explicit repository evidence as `detected_fact`.
- Treat conventional or incomplete inferences as `suggested_assumption`.
- Never treat suggested assumptions as facts.
- Never invent stack, architecture, commands, or project decisions without repository evidence.
- When discovery finds missing or ambiguous context, keep the existing interactive briefing flow instead of blocking abruptly.
- Keep discovery deterministic, local, offline, and read-only.
- Skip obvious sensitive files and heavy/generated folders.
- Keep discovery orchestration in `internal/core/usecase`.
- Keep context signal models, source categories, classifications, and confidence modeling in `internal/core/domain`.
- Put filesystem traversal and reading behind a small core-owned port implemented by an adapter.
- Keep CLI behavior limited to command parsing and report formatting.
- Add focused implementation tests for domain modeling, source/confidence classification, file-specific detectors, skip rules, deterministic ordering, brief precedence, CLI/usecase edge cases, and command regressions.
- Update public documentation in the implementation change after behavior exists.

## Out Of Scope

- Implementing code in this spec-authoring task.
- Modifying files outside `openspec/changes/implement-context-discovery/` in this spec-authoring task.
- Repository-wide indexing.
- Embeddings.
- Vector databases.
- RAG providers.
- GitHub remote discovery.
- GitLab remote discovery.
- Bitbucket remote discovery.
- Local retrieval or snippet ranking.
- Prompt injection into generated agent prompts.
- Changing generated role prompt behavior.
- Changing agent-assisted authoring prompt content.
- Merge, update, overwrite, or append behavior for existing `.specharbor/project-brief.md` files.
- Running package managers, build tools, tests, scripts, shell commands, agents, or workflow tools.
- Verifying that detected commands work.
- Calling provider APIs, local model APIs, network APIs, source-control APIs, or remote services.
- Adding provider abstractions, vector-store abstractions, workflow connector abstractions, or agent-runner abstractions for this local discovery step.
- Changing release, npm, Homebrew, `install.sh`, GoReleaser, publishing, package-manager, tagging, commit, push, pull request, merge, or archive flows.

## Success Criteria

- `specharbor context discover` is specified as a deterministic local/offline discovery command.
- The command reads only supported repository sources through a filesystem port/adapter boundary.
- The command prints a readable discovered context summary.
- Detected facts, suggested assumptions, and user-confirmed context are clearly separated.
- Every signal includes source evidence and a confidence level.
- `.specharbor/project-brief.md` data is preferred as user-confirmed context when present.
- Discovered suggestions can be offered in `specharbor brief`, but user confirmation is still required before writing confirmed context.
- Existing `specharbor brief` write-if-absent and cancellation behavior is preserved.
- Missing or ambiguous context does not abruptly block the briefing workflow.
- Sensitive files and heavy/generated folders are skipped.
- No external APIs, command execution, RAG, indexing, or remote discovery is introduced.
- Existing CLI commands continue to work.
- `go run ./cmd/specharbor validate implement-context-discovery` passes with zero errors.
