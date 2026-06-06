# Tasks: Implement Prompt Command

## Phase 0: Baseline and Scope

- [x] Read `AGENTS.md`, `.specharbor/rules/global.md`, `.specharbor/rules/implementer.md`, `openspec/project.md`, `openspec/specs/architecture/spec.md`, and all files under `openspec/changes/implement-prompt-command/`.
- [x] Inspect the existing role templates under `agent-prompts/roles/`.
- [x] Inspect current CLI files before editing: `cmd/specharbor/main.go` and `internal/adapters/cli/cli.go`.
- [x] Inspect existing initialization use case and adapter patterns for naming, dependency direction, and tests.
- [x] Run `go test ./...` to establish the pre-change baseline.
- [x] Keep implementation limited to the `specharbor prompt` behavior described by this change.
- [x] Do not implement AI-provider calls, external-agent dispatch, prompt template editing, OpenSpec validation, or unrelated commands.

## Phase 1: Core Contracts and Values

- [x] Add a small prompt template loading port under `internal/core/ports`, or add an equivalently small behavior-specific prompt template port.
- [x] Add a small template rendering port under `internal/core/ports` if loading and rendering are separate responsibilities.
- [x] Keep prompt ports separate from initialization, AI provider, workflow dispatcher, spec generation, and validation contracts.
- [x] Add prompt input and result types under `internal/core/usecase` or `internal/core/domain`, keeping them specific to this command.
- [x] Represent the supported role identifiers: `spec-author`, `architecture-reviewer`, `implementer`, `test-engineer`, and `change-reviewer`.
- [x] Ensure core packages remain free of imports from `internal/adapters`, `cmd`, `os`, `filepath`, terminal IO, network APIs, provider SDKs, and external-agent tooling.

## Phase 2: Prompt Rendering Use Case

- [x] Add the prompt rendering use case under `internal/core/usecase`.
- [x] Make the use case accept project root, change id, and role as input.
- [x] Validate that project root is non-empty after trimming whitespace.
- [x] Validate that change id is non-empty after trimming whitespace.
- [x] Validate that role is non-empty after trimming whitespace.
- [x] Reject roles outside the five supported identifiers.
- [x] Load the selected role template through the prompt template port.
- [x] Render at least `change_id` and `task` placeholders.
- [x] Use the deterministic default `task` value `Follow the active OpenSpec change.` because the CLI does not accept a task argument in this change.
- [x] Return a structured result containing the rendered prompt.
- [x] Return errors instead of panicking.
- [x] Do not print from the use case.
- [x] Do not check for OpenSpec change directory existence in this change.

## Phase 3: Template Adapter

- [x] Add or extend a concrete prompt template adapter under `internal/adapters/templates`.
- [x] Keep prompt template handling separate from initialization template handling, even if both live under `internal/adapters/templates`.
- [x] Load role templates from `agent-prompts/roles/<role>.md.tmpl`, or embed those templates while preserving the same role contents.
- [x] Ensure the adapter supports every required role template.
- [x] Render `{{change_id}}` placeholders.
- [x] Render `{{task}}` placeholders to the default task instruction supplied by the use case.
- [x] Ensure successful rendered output does not contain raw `{{change_id}}` or `{{task}}`.
- [x] Return useful errors for missing templates and render failures.
- [x] Keep adapter code limited to concrete template loading and rendering mechanics.
- [x] Do not add provider API keys, network calls, external process execution, or agent dispatch.

## Phase 4: CLI Wiring

- [x] Replace the `prompt` placeholder in `internal/adapters/cli/cli.go`.
- [x] Parse `specharbor prompt <change-id> --role <role>`.
- [x] Reject missing change id.
- [x] Reject missing `--role`.
- [x] Reject `--role` without a value.
- [x] Reject unsupported roles.
- [x] Reject unsupported flags.
- [x] Reject extra positional arguments.
- [x] Obtain the current working directory for the project root.
- [x] Construct the prompt use case with concrete template adapters.
- [x] Print only the rendered prompt to stdout on success.
- [x] Return non-success errors to `cmd/specharbor/main.go` without panicking.
- [x] Keep `cmd/specharbor/main.go` limited to process bootstrapping unless minimal error handling changes are required.
- [x] Preserve existing `help`, `version`, `init`, and unknown command behavior.

## Phase 5: Tests

- [x] Add use case tests with fake ports for successful rendering.
- [x] Test successful rendering for `spec-author`.
- [x] Test successful rendering for `architecture-reviewer`.
- [x] Test successful rendering for `implementer`.
- [x] Test successful rendering for `test-engineer`.
- [x] Test successful rendering for `change-reviewer`.
- [x] Test that empty project root is rejected.
- [x] Test that empty change id is rejected.
- [x] Test that empty role is rejected.
- [x] Test that unsupported role is rejected.
- [x] Test that rendered output replaces `{{change_id}}`.
- [x] Test that rendered output does not leave raw `{{task}}`.
- [x] Add adapter tests verifying every required role template can be loaded and rendered.
- [x] Add adapter tests for missing template and render errors.
- [x] Add CLI tests for successful `specharbor prompt <change-id> --role <role>` output.
- [x] Add CLI tests for missing change id, missing role, missing role value, unsupported role, unsupported flag, and extra argument errors.
- [x] Add regression tests or preserve existing tests for `help`, `version`, `init`, and unknown command behavior.

## Phase 6: Verification and Task Updates

- [x] Run `gofmt` on changed Go files.
- [x] Run `go test ./...`.
- [x] Manually verify `specharbor prompt implement-prompt-command --role spec-author` prints the rendered spec author prompt.
- [x] Manually verify `specharbor prompt implement-prompt-command --role architecture-reviewer` prints the rendered architecture reviewer prompt.
- [x] Manually verify `specharbor prompt implement-prompt-command --role implementer` prints the rendered implementer prompt.
- [x] Manually verify `specharbor prompt implement-prompt-command --role test-engineer` prints the rendered test engineer prompt.
- [x] Manually verify `specharbor prompt implement-prompt-command --role change-reviewer` prints the rendered change reviewer prompt.
- [x] Manually verify successful output contains no raw `{{change_id}}` or `{{task}}`.
- [x] Manually verify the command does not call any AI provider or external agent.
- [x] Update this `tasks.md` by checking off only tasks completed during implementation.
