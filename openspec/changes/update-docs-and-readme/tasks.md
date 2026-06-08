# Tasks: Update Docs and README

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/update-docs-and-readme/`.
- [x] Inspect `git status --short` before editing and identify unrelated in-progress work that must not be modified.
- [x] Confirm this change is documentation-only and does not require CLI behavior changes.
- [x] Confirm no Go code, Go tests, or GitHub Actions workflow files should be modified.
- [x] Inspect the current stabilized CLI command registry and tests before documenting implemented commands.
- [x] Verify whether `specharbor scan` is merged and implemented on the documentation branch before listing it as implemented.
- [x] Verify whether config behavior is merged and implemented on the documentation branch before listing it as implemented.
- [x] Verify whether the CI improvement is merged before describing formatting or uncached test checks as current CI behavior.
- [x] Keep scan/config/CI wording conservative when merge status is uncertain.
- [x] Keep implementation limited to Markdown documentation and this change's task updates.
- [x] Do not update the living architecture spec unless a concrete documentation need is identified and justified before editing.
- [x] Do not add release automation, install scripts, package-manager distribution, badges, website files, docs-site tooling, AI provider setup instructions, or unrelated documentation.

## Phase 1: Documentation Inventory

- [x] Inventory the existing public documentation files: `README.md` and all Markdown files under `docs/`.
- [x] Identify stale command examples, especially examples that use prompt `--agent` instead of the current role-based command shape.
- [x] Identify places that describe planned generation modes or commands as implemented.
- [x] Identify duplicated command lists that could drift.
- [x] Identify docs that mention scan, config, CI, providers, agents, or generation modes and mark whether each statement is implemented, in progress, or planned.
- [x] Decide how to reconcile existing `docs/getting-started.md` with the planned `docs/usage.md` without leaving stale content.
- [x] Decide how to reconcile existing `docs/generation-modes.md` with the planned `docs/workflow.md` without overstating unimplemented modes.
- [x] Confirm that `docs/agent-roles.md` can be updated in place.
- [x] Create only the documentation files justified by `design.md`.
- [x] Avoid creating extra docs files beyond `README.md`, `docs/usage.md`, `docs/workflow.md`, `docs/agent-roles.md`, `docs/contributing.md`, and `docs/development.md` unless the reason is documented.

## Phase 2: README Update Plan

- [x] Update `README.md` with a direct definition of SpecHarbor as a Go CLI for OpenSpec-based AI coding-agent workflows.
- [x] Explain why SpecHarbor exists without marketing fluff.
- [x] Describe the supported OpenSpec/SDD workflow at a high level.
- [x] List implemented commands separately from planned or in-progress commands.
- [x] Use the verified implemented command syntax for `init`.
- [x] Use the verified implemented command syntax for `generate <change-id> --blank`.
- [x] Use the verified implemented command syntax for `validate <change-id>`.
- [x] Use the verified implemented command syntax for `prompt <change-id> --role <role>`.
- [x] Use the verified implemented command syntax for `review <change-id>`.
- [x] Use the verified implemented command syntax for `archive <change-id>`.
- [x] Include simple repository-root examples using `go run ./cmd/specharbor ...`.
- [x] Include local build guidance with `go build ./cmd/specharbor` if still valid.
- [x] Include local test guidance with `go test ./...`.
- [x] Link to usage, workflow, agent roles, contributing, and development docs.
- [x] Mention scan/config only according to verified merge status.
- [x] Make clear that SpecHarbor dogfoods OpenSpec changes for its own development.
- [x] Do not add badges unless a concrete reason is documented.
- [x] Keep the README concise and avoid duplicating every detail from deeper docs.

## Phase 3: Usage Documentation Plan

- [x] Create or update `docs/usage.md` as the command-focused usage guide.
- [x] Explain how to run commands from the repository root with `go run ./cmd/specharbor ...`.
- [x] Explain how to build a local binary with `go build ./cmd/specharbor` if still valid.
- [x] Document `specharbor init` with a simple example.
- [x] Document `specharbor generate <change-id> --blank` with a simple example.
- [x] Document the expected generated OpenSpec change file structure.
- [x] Document `specharbor validate <change-id>` with a simple example.
- [x] Document `specharbor prompt <change-id> --role <role>` with a simple example.
- [x] Document `specharbor review <change-id>` with a simple example.
- [x] Document `specharbor archive <change-id>` with a simple example.
- [x] If `specharbor scan` is merged and verified, document it as an implemented informational command.
- [ ] If `specharbor scan` is not merged and verified, mention it only as roadmap or in progress.
- [x] If config behavior is merged and verified, document only the verified config commands.
- [ ] If config behavior is not merged and verified, mention config only as planned roadmap support.
- [x] Keep examples copy-pasteable and free of secrets, provider credentials, unpublished release commands, and local absolute paths.
- [x] Avoid documenting AI provider setup beyond high-level roadmap notes.
- [x] Ensure usage docs do not imply user projects must be Go projects.

## Phase 4: Contributor Workflow Documentation Plan

- [x] Create or update `docs/contributing.md`.
- [x] Explain that meaningful implementation work should start with an OpenSpec change.
- [x] Document the required reading for contributors and coding agents.
- [x] Explain the role of `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [x] Explain that implementation must stay within the active change scope.
- [x] Explain that `tasks.md` should be updated only for work actually completed.
- [x] Explain that contributors should run `go test ./...` after implementation.
- [x] Explain that Go code should remain inside the existing architecture boundaries.
- [x] Explain that CLI code should not contain business rules.
- [x] Explain that documentation, code, tests, and CI should not be mixed unless the active change explicitly requires it.
- [x] Link to `openspec/project.md` and `openspec/specs/architecture/spec.md` rather than duplicating their full contents.
- [x] Include guidance for reviewing diffs before finalizing an agent's work.

## Phase 5: Agent Workflow Documentation Plan

- [x] Update `docs/agent-roles.md` in place.
- [x] Verify role names against the current prompt generator behavior before documenting examples.
- [x] Explain the Spec Author role.
- [x] Explain the Architecture Reviewer role.
- [x] Explain the Implementer role.
- [x] Explain the Test Engineer role.
- [x] Explain the Change Reviewer role.
- [x] Show a role prompt example using `specharbor prompt <change-id> --role <role>`.
- [x] Link to `.specharbor/rules/` as the source of detailed role instructions.
- [x] Explain that agent-assisted workflows should not require provider API keys.
- [x] Avoid copying every rule file into the docs.
- [x] Avoid detailed AI provider setup instructions.
- [x] Ensure role docs match the OpenSpec/SDD workflow described in `docs/workflow.md`.

## Phase 6: Parallel Agent / Branch Hygiene Notes

- [x] Add branch hygiene guidance to `docs/contributing.md`.
- [x] Warn contributors to use separate branches or worktrees for separate OpenSpec changes.
- [x] Recommend one active change id per branch where possible.
- [x] Warn against multiple agents editing the same files without explicit coordination.
- [x] Tell contributors to check `git status --short` before starting work.
- [x] Tell contributors to check `git status --short` and inspect diffs before finalizing work.
- [x] Explain that unrelated dirty worktree changes must not be reverted or overwritten.
- [x] Explain that docs should be reconciled after merging parallel feature branches.
- [x] Explain that scan/config/CI should not be described as complete until the relevant branch is merged into the documentation branch.
- [x] Keep this guidance practical and short enough that contributors will read it.

## Phase 7: Verification

- [x] Review all changed documentation for stale command syntax.
- [x] Review all changed documentation for claims that unfinished features are implemented.
- [x] Review all changed documentation for duplicated command lists that could drift.
- [x] Verify examples are simple, copy-pasteable, and intended to work from the repository root.
- [x] Verify the docs clearly separate implemented, in-progress, and planned behavior.
- [x] Verify the docs explain that SpecHarbor's own CI is Go-specific.
- [x] Verify the docs explain that user project scanning must remain stack-agnostic.
- [x] Verify no Go files were modified.
- [x] Verify no Go tests were modified.
- [x] Verify `.github/workflows/ci.yml` was not modified.
- [x] Run `go test ./...`.
- [x] Run `go run ./cmd/specharbor validate update-docs-and-readme` if the local command can execute from the repository root.
- [x] Inspect the final diff for unrelated changes.
- [ ] If any verification command cannot be run, leave its task unchecked and record the reason in the final implementation response.

## Phase 8: Task Updates

- [x] Update this `tasks.md` by checking off only tasks completed during documentation implementation.
- [x] Leave tasks unchecked for work not performed.
- [ ] Record any justified documentation file-plan deviation in the final implementation response.
- [x] Confirm the final implementation remains scoped to Markdown documentation and OpenSpec task updates.
