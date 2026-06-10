# Tasks: Implement Workflow Integrations

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-workflow-integrations/`.
- [x] Inspect current command registry and CLI error flow in `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing domain, use case, port, filesystem adapter, template, and test patterns before editing.
- [x] Inspect current prompt role definitions and templates to confirm canonical role ids.
- [x] Inspect existing documentation in `README.md`, `docs/usage.md`, `docs/workflow.md`, `docs/agent-roles.md`, and `docs/generation-modes.md`.
- [x] Keep implementation limited to the read-only `specharbor workflow` command and workflow model described by this change.
- [x] Do not implement `workflow show`, `workflow status`, `workflow next`, workflow execution, GitHub/GitLab integrations, CI integrations, source-control automation, provider calls, agent CLI calls, external command execution, automatic archive, automatic `tasks.md` updates, OpenSpec writes, production code writes, OAuth, credential storage, or new agent roles.

## Phase 1: Domain Model

- [x] Add workflow domain types under `internal/core/domain`.
- [x] Represent stable workflow step ids for `spec-author`, `architecture-reviewer`, `implementer`, `test-engineer`, `change-reviewer`, `commit`, `pull-request`, `merge`, and `archive`.
- [x] Represent workflow step mode values for `agent-assisted` and `manual`.
- [x] Represent workflow command suggestion metadata.
- [x] Represent workflow steps with id, display name, description, order, mode, supported flag, advisory-only flag, required predecessor step ids, command suggestions, and safety notes.
- [x] Represent a workflow result containing the workflow title and ordered steps.
- [x] Ensure role step ids align with `domain.PromptRole` canonical values.
- [x] Ensure the recommended workflow includes exactly the nine steps in the required order.
- [x] Ensure workflow dependencies reference known stable step ids.
- [x] Defensively copy slices when constructors or accessors expose workflow steps, dependencies, command suggestions, or safety notes.
- [x] Keep domain code free of adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, concrete filesystem packages, and external process execution.
- [x] Do not add workflow status domain concepts in this change.

## Phase 2: Use Case

- [x] Add a workflow show use case under `internal/core/usecase`.
- [x] Return the recommended workflow as structured domain data.
- [x] Keep workflow ordering deterministic.
- [x] Keep workflow construction independent of filesystem, provider, source-control, workflow, agent, network, environment, time, and process execution dependencies.
- [x] Do not print from the use case.
- [x] Do not call `os`, terminal IO, provider SDKs, network APIs, source-control APIs, external agents, external processes, workflow tools, Git commands, GitHub APIs, GitLab APIs, or CI APIs from the use case.
- [x] Do not add a filesystem port for the first show command because it does not inspect local status.

## Phase 3: CLI Wiring and Output

- [x] Register the `workflow` command in the CLI command registry.
- [x] Parse `specharbor workflow` with no required arguments.
- [x] Reject unsupported flags clearly.
- [x] Reject extra positional arguments clearly.
- [x] Reject `workflow show`, `workflow status <change-id>`, and `workflow next <change-id>` as unexpected arguments in this change.
- [x] Invoke the workflow show use case from the CLI adapter.
- [x] Print a deterministic human-readable workflow report from the structured domain result.
- [x] Include the workflow title.
- [x] Include ordered steps with ids and display names.
- [x] Include each step purpose.
- [x] Include manual vs agent-assisted indication.
- [x] Include supported-by-SpecHarbor indication.
- [x] Include advisory-only indication.
- [x] Include required predecessor step ids or `none`.
- [x] Include related SpecHarbor command suggestions where applicable.
- [x] Print `none` for manual steps with no SpecHarbor command.
- [x] Include safety notes for commit, pull request, merge, and archive.
- [x] Include global notes stating that command suggestions are advisory and are not executed.
- [x] Ensure successful read-only output returns zero.
- [x] Keep CLI formatting outside core.
- [x] Keep the CLI adapter free of workflow ordering policy beyond formatting the structured result.
- [x] Preserve existing `help`, `version`, `init`, `scan`, `generate`, `prompt`, `validate`, `review`, `archive`, `config`, and unknown command behavior.

## Phase 4: Documentation

- [x] Update `README.md` to list `workflow` as an implemented command once implemented.
- [x] Update `README.md` to explain that `workflow` is read-only and advisory.
- [x] Update `docs/usage.md` with `specharbor workflow` usage.
- [x] Update `docs/usage.md` with an example workflow output or abbreviated output shape.
- [x] Update `docs/workflow.md` to include the recommended nine-step workflow.
- [x] Update `docs/workflow.md` to explain how `workflow` relates to `generate`, `validate`, `prompt`, `review`, and `archive`.
- [x] Update `docs/workflow.md` to state that commit, pull request, and merge remain manual.
- [x] Update `docs/workflow.md` to state that no GitHub API, GitLab API, CI, provider API, agent CLI, source-control automation, workflow execution, or remote automation is introduced.
- [x] Update `docs/agent-roles.md` if needed to align workflow role names with supported prompt roles.
- [x] Do not create a separate documentation-only OpenSpec change.

## Phase 5: Tests

- [x] Add domain tests for workflow step ids.
- [x] Add domain tests for workflow step ordering.
- [x] Add domain tests for workflow step display names.
- [x] Add domain tests for manual vs agent-assisted step classification.
- [x] Add domain tests for supported and advisory-only step flags.
- [x] Add domain tests for required step dependencies.
- [x] Add domain tests for command suggestion metadata.
- [x] Add domain tests proving role step ids align with canonical prompt roles.
- [x] Add domain tests for defensive copies if workflow result slices are exposed.
- [x] Add use case tests proving workflow show returns the ordered recommended workflow.
- [x] Add use case tests proving workflow output data includes all required steps.
- [x] Add use case tests proving advisory command suggestions are present.
- [x] Add use case tests proving safety notes are present for manual source-control steps.
- [x] Add use case tests proving workflow show does not require filesystem, provider, source-control, workflow, agent, network, or process collaborators.
- [x] Add CLI tests proving `workflow` parses correctly.
- [x] Add CLI tests proving `workflow` prints deterministic output.
- [x] Add CLI tests proving `workflow` returns zero on successful read-only output.
- [x] Add CLI tests proving unsupported flags are rejected clearly.
- [x] Add CLI tests proving extra arguments are rejected clearly.
- [x] Add CLI tests proving `workflow status <change-id>` is not supported in this change.
- [x] Add CLI tests proving output contains the title, ordered steps, role names, purposes, modes, supported/advisory indicators, command suggestions, safety notes, and advisory caveats.
- [x] Add architecture tests proving core does not import adapters, CLI packages, `os`, terminal IO, provider SDKs, network APIs, source-control SDKs, workflow SDKs, external-agent SDKs, or external process packages for workflow behavior.
- [x] Add or update tests proving no provider, network, source-control, workflow, agent SDK, or external execution dependency is introduced.
- [x] Add regression tests proving `init` remains unchanged.
- [x] Add regression tests proving `scan` remains unchanged.
- [x] Add regression tests proving generation modes remain unchanged.
- [x] Add regression tests proving `validate` remains unchanged.
- [x] Add regression tests proving `prompt` remains unchanged.
- [x] Add regression tests proving `review` remains unchanged.
- [x] Add regression tests proving `archive` remains unchanged.
- [x] Add regression tests proving `config` remains unchanged.

## Phase 6: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor workflow` prints the recommended workflow and exits zero.
- [x] Manually verify unsupported flags return clear errors.
- [x] Manually verify extra arguments return clear errors.
- [x] Manually verify `workflow status <change-id>` is rejected as unsupported in this change.
- [x] Manually verify `specharbor workflow` does not write files, modify OpenSpec files, modify production code, change `tasks.md`, archive, commit, push, create PRs, merge, call provider APIs, call agent CLIs, call GitHub/GitLab APIs, inspect CI, execute external commands, or trigger workflows.
- [x] Run `git status --short`.
- [x] Run `git diff -- openspec/changes/implement-workflow-integrations/` and inspect the active change diff.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
