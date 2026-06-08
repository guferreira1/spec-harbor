# Tasks: Update Docs and README

## Phase 0: Baseline and Scope

- [ ] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/update-docs-and-readme/`.
- [ ] Inspect `git status --short` before editing and identify unrelated in-progress work that must not be modified.
- [ ] Confirm this change is documentation-only and does not require CLI behavior changes.
- [ ] Confirm no Go code, Go tests, or GitHub Actions workflow files should be modified.
- [ ] Inspect the current stabilized CLI command registry and tests before documenting implemented commands.
- [ ] Verify whether `specharbor scan` is merged and implemented on the documentation branch before listing it as implemented.
- [ ] Verify whether config behavior is merged and implemented on the documentation branch before listing it as implemented.
- [ ] Verify whether the CI improvement is merged before describing formatting or uncached test checks as current CI behavior.
- [ ] Keep scan/config/CI wording conservative when merge status is uncertain.
- [ ] Keep implementation limited to Markdown documentation and this change's task updates.
- [ ] Do not update the living architecture spec unless a concrete documentation need is identified and justified before editing.
- [ ] Do not add release automation, install scripts, package-manager distribution, badges, website files, docs-site tooling, AI provider setup instructions, or unrelated documentation.

## Phase 1: Documentation Inventory

- [ ] Inventory the existing public documentation files: `README.md` and all Markdown files under `docs/`.
- [ ] Identify stale command examples, especially examples that use prompt `--agent` instead of the current role-based command shape.
- [ ] Identify places that describe planned generation modes or commands as implemented.
- [ ] Identify duplicated command lists that could drift.
- [ ] Identify docs that mention scan, config, CI, providers, agents, or generation modes and mark whether each statement is implemented, in progress, or planned.
- [ ] Decide how to reconcile existing `docs/getting-started.md` with the planned `docs/usage.md` without leaving stale content.
- [ ] Decide how to reconcile existing `docs/generation-modes.md` with the planned `docs/workflow.md` without overstating unimplemented modes.
- [ ] Confirm that `docs/agent-roles.md` can be updated in place.
- [ ] Create only the documentation files justified by `design.md`.
- [ ] Avoid creating extra docs files beyond `README.md`, `docs/usage.md`, `docs/workflow.md`, `docs/agent-roles.md`, `docs/contributing.md`, and `docs/development.md` unless the reason is documented.

## Phase 2: README Update Plan

- [ ] Update `README.md` with a direct definition of SpecHarbor as a Go CLI for OpenSpec-based AI coding-agent workflows.
- [ ] Explain why SpecHarbor exists without marketing fluff.
- [ ] Describe the supported OpenSpec/SDD workflow at a high level.
- [ ] List implemented commands separately from planned or in-progress commands.
- [ ] Use the verified implemented command syntax for `init`.
- [ ] Use the verified implemented command syntax for `generate <change-id> --blank`.
- [ ] Use the verified implemented command syntax for `validate <change-id>`.
- [ ] Use the verified implemented command syntax for `prompt <change-id> --role <role>`.
- [ ] Use the verified implemented command syntax for `review <change-id>`.
- [ ] Use the verified implemented command syntax for `archive <change-id>`.
- [ ] Include simple repository-root examples using `go run ./cmd/specharbor ...`.
- [ ] Include local build guidance with `go build ./cmd/specharbor` if still valid.
- [ ] Include local test guidance with `go test ./...`.
- [ ] Link to usage, workflow, agent roles, contributing, and development docs.
- [ ] Mention scan/config only according to verified merge status.
- [ ] Make clear that SpecHarbor dogfoods OpenSpec changes for its own development.
- [ ] Do not add badges unless a concrete reason is documented.
- [ ] Keep the README concise and avoid duplicating every detail from deeper docs.

## Phase 3: Usage Documentation Plan

- [ ] Create or update `docs/usage.md` as the command-focused usage guide.
- [ ] Explain how to run commands from the repository root with `go run ./cmd/specharbor ...`.
- [ ] Explain how to build a local binary with `go build ./cmd/specharbor` if still valid.
- [ ] Document `specharbor init` with a simple example.
- [ ] Document `specharbor generate <change-id> --blank` with a simple example.
- [ ] Document the expected generated OpenSpec change file structure.
- [ ] Document `specharbor validate <change-id>` with a simple example.
- [ ] Document `specharbor prompt <change-id> --role <role>` with a simple example.
- [ ] Document `specharbor review <change-id>` with a simple example.
- [ ] Document `specharbor archive <change-id>` with a simple example.
- [ ] If `specharbor scan` is merged and verified, document it as an implemented informational command.
- [ ] If `specharbor scan` is not merged and verified, mention it only as roadmap or in progress.
- [ ] If config behavior is merged and verified, document only the verified config commands.
- [ ] If config behavior is not merged and verified, mention config only as planned roadmap support.
- [ ] Keep examples copy-pasteable and free of secrets, provider credentials, unpublished release commands, and local absolute paths.
- [ ] Avoid documenting AI provider setup beyond high-level roadmap notes.
- [ ] Ensure usage docs do not imply user projects must be Go projects.

## Phase 4: Contributor Workflow Documentation Plan

- [ ] Create or update `docs/contributing.md`.
- [ ] Explain that meaningful implementation work should start with an OpenSpec change.
- [ ] Document the required reading for contributors and coding agents.
- [ ] Explain the role of `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`.
- [ ] Explain that implementation must stay within the active change scope.
- [ ] Explain that `tasks.md` should be updated only for work actually completed.
- [ ] Explain that contributors should run `go test ./...` after implementation.
- [ ] Explain that Go code should remain inside the existing architecture boundaries.
- [ ] Explain that CLI code should not contain business rules.
- [ ] Explain that documentation, code, tests, and CI should not be mixed unless the active change explicitly requires it.
- [ ] Link to `openspec/project.md` and `openspec/specs/architecture/spec.md` rather than duplicating their full contents.
- [ ] Include guidance for reviewing diffs before finalizing an agent's work.

## Phase 5: Agent Workflow Documentation Plan

- [ ] Update `docs/agent-roles.md` in place.
- [ ] Verify role names against the current prompt generator behavior before documenting examples.
- [ ] Explain the Spec Author role.
- [ ] Explain the Architecture Reviewer role.
- [ ] Explain the Implementer role.
- [ ] Explain the Test Engineer role.
- [ ] Explain the Change Reviewer role.
- [ ] Show a role prompt example using `specharbor prompt <change-id> --role <role>`.
- [ ] Link to `.specharbor/rules/` as the source of detailed role instructions.
- [ ] Explain that agent-assisted workflows should not require provider API keys.
- [ ] Avoid copying every rule file into the docs.
- [ ] Avoid detailed AI provider setup instructions.
- [ ] Ensure role docs match the OpenSpec/SDD workflow described in `docs/workflow.md`.

## Phase 6: Parallel Agent / Branch Hygiene Notes

- [ ] Add branch hygiene guidance to `docs/contributing.md`.
- [ ] Warn contributors to use separate branches or worktrees for separate OpenSpec changes.
- [ ] Recommend one active change id per branch where possible.
- [ ] Warn against multiple agents editing the same files without explicit coordination.
- [ ] Tell contributors to check `git status --short` before starting work.
- [ ] Tell contributors to check `git status --short` and inspect diffs before finalizing work.
- [ ] Explain that unrelated dirty worktree changes must not be reverted or overwritten.
- [ ] Explain that docs should be reconciled after merging parallel feature branches.
- [ ] Explain that scan/config/CI should not be described as complete until the relevant branch is merged into the documentation branch.
- [ ] Keep this guidance practical and short enough that contributors will read it.

## Phase 7: Verification

- [ ] Review all changed documentation for stale command syntax.
- [ ] Review all changed documentation for claims that unfinished features are implemented.
- [ ] Review all changed documentation for duplicated command lists that could drift.
- [ ] Verify examples are simple, copy-pasteable, and intended to work from the repository root.
- [ ] Verify the docs clearly separate implemented, in-progress, and planned behavior.
- [ ] Verify the docs explain that SpecHarbor's own CI is Go-specific.
- [ ] Verify the docs explain that user project scanning must remain stack-agnostic.
- [ ] Verify no Go files were modified.
- [ ] Verify no Go tests were modified.
- [ ] Verify `.github/workflows/ci.yml` was not modified.
- [ ] Run `go test ./...`.
- [ ] Run `go run ./cmd/specharbor validate update-docs-and-readme` if the local command can execute from the repository root.
- [ ] Inspect the final diff for unrelated changes.
- [ ] If any verification command cannot be run, leave its task unchecked and record the reason in the final implementation response.

## Phase 8: Task Updates

- [ ] Update this `tasks.md` by checking off only tasks completed during documentation implementation.
- [ ] Leave tasks unchecked for work not performed.
- [ ] Record any justified documentation file-plan deviation in the final implementation response.
- [ ] Confirm the final implementation remains scoped to Markdown documentation and OpenSpec task updates.
