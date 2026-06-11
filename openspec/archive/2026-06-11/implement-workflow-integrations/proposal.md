# Proposal: Implement Workflow Integrations

## Problem

SpecHarbor already supports the core OpenSpec/SDD flow through separate commands:

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

The current CLI can initialize projects, scan local context, generate OpenSpec changes, validate required files, render role prompts, review task completion, archive completed changes, and show local config. Users still need to understand the recommended end-to-end workflow by reading documentation or inferring it from command names.

The next product step is to make the recommended workflow visible from the CLI without turning SpecHarbor into CI/CD automation, GitHub automation, source-control automation, remote execution, or an agent runner for implementation work.

## Goal

Add a safe first workflow integration foundation that lets users inspect the recommended SpecHarbor workflow:

```text
specharbor workflow
```

The command prints the ordered OpenSpec/SDD workflow:

1. Spec Author Agent
2. Architecture Reviewer Agent
3. Implementer Agent
4. Test Engineer Agent
5. Change Reviewer Agent
6. Commit
7. Pull Request
8. Merge
9. Archive

The first implementation must be read-only, deterministic, local, and advisory. It should guide users toward existing commands but must not execute those commands or infer external workflow state.

## Command Surface Decision

Implement only:

```text
specharbor workflow
```

Do not implement these subcommands in this change:

- `specharbor workflow show`
- `specharbor workflow status <change-id>`
- `specharbor workflow next <change-id>`

Rationale:

- `specharbor workflow` is the smallest useful command surface and matches the preferred first scope.
- `workflow show` can be added later if the workflow command grows additional views.
- `workflow status` and `workflow next` would require careful local status semantics. Some local facts are detectable, but agent review completion, implementation completion, tests, commits, PRs, merges, and remote workflow state cannot be determined reliably from local OpenSpec files alone.
- Existing `validate` and `review` commands already provide deterministic local checks for required OpenSpec files and task checkbox completion.

## Scope

- Add a workflow domain model under `internal/core/domain`.
- Define stable workflow step ids for:
  - `spec-author`
  - `architecture-reviewer`
  - `implementer`
  - `test-engineer`
  - `change-reviewer`
  - `commit`
  - `pull-request`
  - `merge`
  - `archive`
- Represent each workflow step with:
  - stable id;
  - display name;
  - short description or purpose;
  - deterministic order;
  - manual vs agent-assisted classification;
  - whether SpecHarbor currently supports the step;
  - whether the step is advisory only;
  - required predecessor step ids;
  - related SpecHarbor command suggestions, where applicable;
  - safety notes, especially for commit, pull request, and merge.
- Align agent-assisted workflow step ids with canonical prompt role ids:
  - `spec-author`
  - `architecture-reviewer`
  - `implementer`
  - `test-engineer`
  - `change-reviewer`
- Add a workflow show use case under `internal/core/usecase` that returns the ordered recommended workflow as structured core-owned results.
- Add CLI parsing for `specharbor workflow`.
- Print deterministic, readable, copy-paste friendly workflow output from `internal/adapters/cli`.
- Include advisory command suggestions for existing commands:
  - `specharbor generate <change-id> --guided ...`
  - `specharbor validate <change-id>`
  - `specharbor prompt <change-id> --role architecture-reviewer`
  - `specharbor prompt <change-id> --role implementer`
  - `specharbor prompt <change-id> --role test-engineer`
  - `specharbor prompt <change-id> --role change-reviewer`
  - `specharbor review <change-id>`
  - `specharbor archive <change-id>`
- State clearly in output that commit, pull request, and merge are manual source-control steps, not SpecHarbor automation.
- Reject unsupported flags and unexpected arguments clearly.
- Preserve existing behavior for `init`, `scan`, `generate`, `validate`, `prompt`, `review`, `archive`, and `config`.
- Add documentation updates in the same implementation change for:
  - `README.md`
  - `docs/usage.md`
  - `docs/workflow.md`
  - `docs/agent-roles.md` if role alignment needs clarification.
- Add focused tests for domain, use case, CLI output, safety boundaries, architecture rules, documentation expectations, and command regressions.

## Out of Scope

- `specharbor workflow show`.
- `specharbor workflow status <change-id>`.
- `specharbor workflow next <change-id>`.
- Workflow execution.
- Workflow dispatching.
- GitHub API integrations.
- GitLab API integrations.
- CI integrations.
- Remote automation.
- Source-control automation.
- OAuth.
- Credential storage.
- Provider API calls.
- Local model calls.
- Agent CLI calls.
- External command execution.
- Commit automation.
- Push automation.
- Pull request creation.
- Pull request merge automation.
- Branch inspection.
- Remote branch inspection.
- CI status inspection.
- Git diff inspection.
- Automatic archive.
- Automatic `tasks.md` checkbox updates.
- OpenSpec file writes.
- Production code writes.
- Documentation-only separate changes.
- New agent roles beyond the existing canonical prompt roles.
- Machine-readable output formats.
- User-configurable workflow definitions.
- Workflow connector ports or adapters for external tools.

## Status Detection Decision

Workflow status detection is deferred from this first change.

This change should not introduce a `workflow status` command or a workflow status domain model. A future status feature may deterministically inspect only local facts such as active change file existence, validation results, task checkbox state, and archive paths. If that future feature is implemented, workflow status concepts must live in `internal/core/domain`, orchestration must live in `internal/core/usecase`, file reads must go through small core ports, and filesystem access must stay in adapters.

This change must not infer implementation, tests, architecture review, change review, commit, pull request, merge, CI, or remote completion without deterministic local evidence.

## Success Criteria

- Running `specharbor workflow` prints the recommended SpecHarbor workflow and exits zero.
- The output includes a workflow title, ordered steps, stable step ids, display names, short purposes, related SpecHarbor command suggestions where applicable, manual vs agent-assisted classification, supported/advisory indicators, required predecessor information, and safety notes for commit, pull request, and merge.
- The workflow model uses canonical prompt role ids for agent-assisted role steps.
- The output is deterministic and does not depend on local repository status, GitHub, CI, network APIs, provider APIs, agent CLIs, external commands, environment variables, or current date.
- Unsupported flags and extra arguments are rejected clearly.
- No files are created, modified, archived, or deleted by `specharbor workflow`.
- Existing commands keep their current behavior.
- Documentation explains the command, the recommended workflow, how output relates to existing commands, what is advisory vs automated, and the explicit safety boundaries.
- Focused tests cover domain, use case, CLI, architecture, documentation, safety, and existing command regressions.
- `go test ./...` succeeds after implementation.
