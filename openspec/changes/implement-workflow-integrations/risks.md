# Risks: Implement Workflow Integrations

## Overbuilding workflow automation

Workflow integrations are adjacent to CI, source control, GitHub, GitLab, agent tools, and remote execution. The main risk is turning a read-only workflow guide into automation before SpecHarbor has explicit requirements and safety controls for those integrations.

Mitigation:

- Implement only `specharbor workflow`.
- Keep output advisory.
- Do not add workflow connector ports or adapters.
- Do not call GitHub, GitLab, CI, provider APIs, agent CLIs, source-control SDKs, workflow SDKs, network APIs, or external commands.
- Leave all automation for future OpenSpec changes.

## Status ambiguity

Some workflow facts are locally detectable, such as required OpenSpec files and task checkboxes. Other important facts are not, such as architecture review completion, implementation quality, test execution, change review approval, commit existence, PR state, CI status, and merge status.

Mitigation:

- Do not implement `workflow status` or `workflow next` in this change.
- Keep the first command to a deterministic recommended workflow view.
- Document status detection as future work with strict local-only and unknown-state boundaries.
- Continue using existing `validate` and `review` commands for deterministic local checks.

## Advisory text interpreted as execution

Users may see command suggestions and assume SpecHarbor ran those commands or verified those stages.

Mitigation:

- Label command suggestions as advisory.
- Add global output notes stating that suggestions are not executed.
- Add safety notes for manual source-control steps.
- Document that commit, pull request, merge, CI, provider, and agent execution are not automated.

## Architecture leakage

Workflow step definitions, ordering, role alignment, and safety policy could be embedded directly in CLI printing code, making future workflow behavior harder to evolve.

Mitigation:

- Put workflow step definitions and result concepts in `internal/core/domain`.
- Put workflow orchestration in `internal/core/usecase`.
- Keep CLI code limited to parsing, use case invocation, output formatting, and errors.
- Do not put report formatting in core.
- Add architecture tests to guard core dependency boundaries.

## Role naming drift

Workflow role step ids must align with current prompt role ids. If the workflow invents alternate names, users may see commands that do not match `prompt --role`.

Mitigation:

- Use the existing canonical prompt role ids:
  - `spec-author`
  - `architecture-reviewer`
  - `implementer`
  - `test-engineer`
  - `change-reviewer`
- Add tests proving role step ids align with `domain.PromptRole`.
- Do not introduce new agent roles in this change.

## Command surface creep

Adding `workflow show`, `workflow status`, and `workflow next` at the same time would increase parsing, docs, tests, and user expectations before status semantics are settled.

Mitigation:

- Implement only `specharbor workflow`.
- Reject extra arguments clearly.
- Document that status and next-step detection are deferred.
- Add subcommands later only through separate OpenSpec changes.

## Output churn

Human-readable CLI output can become brittle if it relies on decorative formatting, terminal width, colors, current dates, or environment-specific content.

Mitigation:

- Keep output plain text.
- Avoid colors, terminal width detection, timestamps, absolute paths, and local status.
- Test the important output facts rather than incidental spacing where possible.
- Keep wording concise and deterministic.

## Regression in existing commands

Adding a new command to the registry and updating help/docs could accidentally affect existing commands or shared parsing assumptions.

Mitigation:

- Keep workflow parsing independent and small.
- Preserve the existing command registry pattern.
- Add regression tests for `init`, `scan`, generation modes, `validate`, `prompt`, `review`, `archive`, `config`, and unknown commands.
- Run `go test ./...`.

## Documentation mismatch

Because the workflow command is a public CLI behavior, incomplete docs could make the feature appear more automated than it is.

Mitigation:

- Update `README.md`, `docs/usage.md`, and `docs/workflow.md` in the same implementation change.
- Update `docs/agent-roles.md` if role alignment needs clarification.
- Include example output.
- Explicitly document advisory vs automated behavior and all safety boundaries.

## Future status design constraints

Deferring status avoids guessing now, but a future status command still needs clear boundaries or it may overreach into GitHub, CI, source control, or agents.

Mitigation:

- Document future status constraints in `design.md`.
- Require future status concepts to live in `internal/core/domain`.
- Require future status orchestration to live in `internal/core/usecase`.
- Require future file reads to go through small read-only ports.
- Require undetectable steps to be reported as unknown rather than inferred.

## Underbuilding the foundation

A plain static help string in the CLI could print the workflow quickly, but it would not create reusable core-owned workflow concepts for future status, docs, or alternate presentation.

Mitigation:

- Add a small domain model for workflow steps and suggestions.
- Add a small use case that returns structured workflow data.
- Avoid unused extension frameworks, registries, workflow dispatchers, or connector abstractions.
- Keep the foundation structured but intentionally narrow.
