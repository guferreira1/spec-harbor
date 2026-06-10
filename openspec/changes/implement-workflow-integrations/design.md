# Design: Implement Workflow Integrations

## Overview

This change adds a read-only workflow guide to SpecHarbor:

```text
specharbor workflow
```

The command returns a structured, core-owned representation of the recommended OpenSpec/SDD workflow and formats it in the CLI adapter. The feature is intentionally advisory. It helps users see how existing commands fit together, but it does not execute commands, call agents, inspect source control, inspect CI, call provider APIs, modify files, or infer remote state.

## Command Contract

Supported:

```text
specharbor workflow
```

Rejected:

```text
specharbor workflow show
specharbor workflow status <change-id>
specharbor workflow next <change-id>
specharbor workflow --format json
specharbor workflow <anything>
```

Argument behavior:

- no arguments prints the recommended workflow and returns zero;
- any flag returns `unsupported flag: <flag>`;
- any positional argument returns `unexpected argument: <arg>`;
- no project root or change id is required because this command does not inspect local status.

No `workflow show` alias should be added in this change. The single command keeps the first surface small and avoids adding a subcommand hierarchy before there are multiple workflow views.

## Workflow Step Model

Add workflow domain concepts under:

```text
internal/core/domain
```

Expected core-owned concepts:

- `WorkflowStepID`
- `WorkflowStepMode`
- `WorkflowStep`
- `WorkflowCommandSuggestion`
- `Workflow`

Recommended step mode values:

- `agent-assisted`
- `manual`

Recommended workflow fields:

```text
Title string
Steps []WorkflowStep
```

Recommended workflow step fields:

```text
ID WorkflowStepID
DisplayName string
Description string
Order int
Mode WorkflowStepMode
Supported bool
AdvisoryOnly bool
Requires []WorkflowStepID
CommandSuggestions []WorkflowCommandSuggestion
SafetyNotes []string
```

Recommended command suggestion fields:

```text
Command string
Description string
```

The domain model should defensively copy step, requirement, command suggestion, and safety note slices when constructors or accessors expose slices.

The domain package must not import adapters, CLI packages, `os`, terminal IO, network APIs, provider SDKs, source-control SDKs, workflow SDKs, external-agent SDKs, or process execution packages.

## Recommended Steps

The recommended workflow must include exactly these steps in this order for the first implementation:

| Order | Step ID | Display Name | Mode | Supported | Advisory Only |
| --- | --- | --- | --- | --- | --- |
| 1 | `spec-author` | Spec Author Agent | `agent-assisted` | yes | no |
| 2 | `architecture-reviewer` | Architecture Reviewer Agent | `agent-assisted` | yes | no |
| 3 | `implementer` | Implementer Agent | `agent-assisted` | yes | no |
| 4 | `test-engineer` | Test Engineer Agent | `agent-assisted` | yes | no |
| 5 | `change-reviewer` | Change Reviewer Agent | `agent-assisted` | yes | no |
| 6 | `commit` | Commit | `manual` | no | yes |
| 7 | `pull-request` | Pull Request | `manual` | no | yes |
| 8 | `merge` | Merge | `manual` | no | yes |
| 9 | `archive` | Archive | `manual` | yes | no |

The first five step ids intentionally match the canonical prompt role ids from `domain.PromptRole`:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

Do not introduce new agent roles in this change.

## Step Purposes

Use concise descriptions:

- `spec-author`: Create or refine the OpenSpec change package.
- `architecture-reviewer`: Review the proposed scope and design against architecture boundaries.
- `implementer`: Apply the approved OpenSpec change.
- `test-engineer`: Add or run focused verification for the implemented change.
- `change-reviewer`: Review the final diff, task state, and verification evidence.
- `commit`: Commit the reviewed local changes manually.
- `pull-request`: Open a pull request manually in the user's source-control workflow.
- `merge`: Merge manually after the user's review and CI process passes.
- `archive`: Archive the completed OpenSpec change after the work is merged or otherwise accepted.

Descriptions must avoid implying that SpecHarbor executes implementation, tests, source-control actions, PRs, merges, CI, or remote workflows.

## Step Dependencies

Represent predecessor requirements with stable step ids:

- `spec-author`: none
- `architecture-reviewer`: `spec-author`
- `implementer`: `architecture-reviewer`
- `test-engineer`: `implementer`
- `change-reviewer`: `test-engineer`
- `commit`: `change-reviewer`
- `pull-request`: `commit`
- `merge`: `pull-request`
- `archive`: `merge`

These requirements are guidance for ordering only. In this first change, SpecHarbor must not enforce the requirements or infer whether they are complete.

## Command Suggestions

Command suggestions are advisory text only. The workflow command must not execute any suggested command.

Recommended suggestions:

- `spec-author`
  - `specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"`
  - `specharbor prompt <change-id> --role spec-author`
- `architecture-reviewer`
  - `specharbor validate <change-id>`
  - `specharbor prompt <change-id> --role architecture-reviewer`
- `implementer`
  - `specharbor prompt <change-id> --role implementer`
- `test-engineer`
  - `specharbor prompt <change-id> --role test-engineer`
- `change-reviewer`
  - `specharbor review <change-id>`
  - `specharbor prompt <change-id> --role change-reviewer`
- `commit`
  - no SpecHarbor command
- `pull-request`
  - no SpecHarbor command
- `merge`
  - no SpecHarbor command
- `archive`
  - `specharbor archive <change-id>`

Suggestions should use `specharbor` rather than `go run ./cmd/specharbor` because CLI output should describe the installed command. Documentation can still show development examples with `go run ./cmd/specharbor` where existing docs use that style.

## Safety Notes

The workflow should include clear safety notes for manual source-control steps:

- Commit: SpecHarbor does not commit, stage files, modify branches, push, or sign commits.
- Pull Request: SpecHarbor does not create PRs, call GitHub/GitLab APIs, set reviewers, edit labels, or inspect remote branches.
- Merge: SpecHarbor does not merge, approve, bypass review, inspect CI, trigger CI, or update remote repositories.
- Archive: `specharbor workflow` does not archive automatically; `specharbor archive <change-id>` remains an explicit user command.

These notes may be represented on the individual steps and repeated in a final "Safety" section of the CLI output.

## Use Case

Add a workflow show use case under:

```text
internal/core/usecase
```

Expected behavior:

- return the recommended workflow as structured domain data;
- sort or construct steps in deterministic order;
- validate that step order is stable and unique;
- validate that dependencies refer to known step ids if validation helpers are added;
- avoid filesystem, terminal IO, network access, provider calls, source-control calls, workflow calls, external-agent calls, and external command execution;
- avoid CLI-specific formatting.

Because this first workflow command does not inspect local project status, it does not need a filesystem port.

Possible shape:

```text
type ShowWorkflow struct{}
type ShowWorkflowResult struct {
    Workflow domain.Workflow
}
```

The exact names may differ, but the use case should preserve the existing architecture pattern: core returns structured results, CLI formats text.

## CLI Output

Output must be deterministic, readable, and copy-paste friendly. Do not use color, terminal width detection, current dates, absolute local paths, Git status, CI status, provider names beyond existing roles, or environment-specific content.

Recommended output shape:

```text
SpecHarbor recommended workflow.
Title: OpenSpec/SDD agent-driven workflow

Steps:
1. spec-author - Spec Author Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: none
   Purpose: Create or refine the OpenSpec change package.
   Commands:
   - specharbor generate <change-id> --guided --type feature --title "<title>" --summary "<summary>"
   - specharbor prompt <change-id> --role spec-author

2. architecture-reviewer - Architecture Reviewer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: spec-author
   Purpose: Review the proposed scope and design against architecture boundaries.
   Commands:
   - specharbor validate <change-id>
   - specharbor prompt <change-id> --role architecture-reviewer

3. implementer - Implementer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: architecture-reviewer
   Purpose: Apply the approved OpenSpec change.
   Commands:
   - specharbor prompt <change-id> --role implementer

4. test-engineer - Test Engineer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: implementer
   Purpose: Add or run focused verification for the implemented change.
   Commands:
   - specharbor prompt <change-id> --role test-engineer

5. change-reviewer - Change Reviewer Agent
   Mode: agent-assisted
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: test-engineer
   Purpose: Review the final diff, task state, and verification evidence.
   Commands:
   - specharbor review <change-id>
   - specharbor prompt <change-id> --role change-reviewer

6. commit - Commit
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Requires: change-reviewer
   Purpose: Commit the reviewed local changes manually.
   Commands:
   - none
   Safety:
   - SpecHarbor does not commit, stage files, modify branches, push, or sign commits.

7. pull-request - Pull Request
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Requires: commit
   Purpose: Open a pull request manually in your source-control workflow.
   Commands:
   - none
   Safety:
   - SpecHarbor does not create PRs, call source-control APIs, set reviewers, edit labels, or inspect remote branches.

8. merge - Merge
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Requires: pull-request
   Purpose: Merge manually after your review and CI process passes.
   Commands:
   - none
   Safety:
   - SpecHarbor does not merge, approve, inspect CI, trigger CI, or update remote repositories.

9. archive - Archive
   Mode: manual
   Supported by SpecHarbor: yes
   Advisory only: no
   Requires: merge
   Purpose: Archive the completed OpenSpec change after the work is accepted.
   Commands:
   - specharbor archive <change-id>
   Safety:
   - specharbor workflow does not archive automatically.

Notes:
- Suggestions are advisory and are not executed by this command.
- This command does not inspect Git, GitHub, GitLab, CI, provider APIs, agent CLIs, or remote workflow state.
```

Minor wording changes are acceptable during implementation, but the output must include the required facts and remain deterministic.

## Status Detection

Do not implement workflow status detection in this change.

Reasoning:

- Local OpenSpec files can prove some facts, such as required file existence, validation result, task checkbox state, and archived directory presence.
- Local OpenSpec files cannot prove that an architecture review happened, an implementer completed production code, a test engineer ran the correct test suite, a change reviewer approved the final diff, a commit exists for the intended work, a PR was opened, CI passed, or a merge occurred.
- Inspecting Git branches, remotes, CI, GitHub, GitLab, or agents is explicitly out of scope.
- A partial status command would likely look authoritative while reporting many important steps as unknown.

Future status behavior should be specified in a separate OpenSpec change. If added later, it must:

- use a new command such as `specharbor workflow status <change-id>`;
- define workflow status result types in `internal/core/domain`;
- orchestrate status calculation in `internal/core/usecase`;
- read local files only through a small read-only port;
- perform concrete filesystem reads in adapters;
- reject unsafe change ids before filesystem checks;
- never call GitHub, GitLab, CI, provider APIs, agent CLIs, source-control SDKs, workflow SDKs, network APIs, or external commands;
- clearly mark undetectable steps as unknown rather than inferred.

## Integration With Existing Commands

The workflow command connects conceptually with existing commands by listing them as suggestions. It must not change their behavior.

No behavior changes are allowed for:

- `init`
- `scan`
- `generate`
- `validate`
- `prompt`
- `review`
- `archive`
- `config`

The workflow output may mention that:

- `generate` helps create an OpenSpec change package;
- `validate` checks required OpenSpec change files;
- `prompt --role ...` generates role-specific prompts for existing canonical roles;
- `review` checks local change package and task checkbox completion;
- `archive` explicitly moves a completed change to the archive.

These statements must remain advisory text, not automatic execution.

## Architecture

Responsibilities:

- `internal/core/domain`: workflow ids, modes, steps, command suggestions, safety notes, and workflow result concepts.
- `internal/core/usecase`: constructing and returning the recommended workflow.
- `internal/core/ports`: no new port is required for the first read-only show command because it has no external dependency.
- `internal/adapters/cli`: command registration, argument parsing, current CLI output formatting, and error handling.
- `internal/adapters/filesystem`: no workflow changes required for the first show command.
- `internal/platform`: no workflow changes expected.

Forbidden in core:

- adapter imports;
- CLI imports;
- `os`;
- terminal IO;
- network APIs;
- provider SDKs;
- source-control SDKs;
- workflow SDKs;
- external-agent SDKs;
- external command execution.

Report formatting must stay outside core. Core should return data that can later support other presentation formats, but this change should not add JSON or machine-readable output.

## Documentation

Implementation must update public documentation in the same change:

- `README.md`
- `docs/usage.md`
- `docs/workflow.md`
- `docs/agent-roles.md` if role alignment needs clarification

Documentation must explain:

- what `specharbor workflow` does;
- the recommended workflow order;
- how workflow output relates to `generate`, `validate`, `prompt`, `review`, and `archive`;
- that command suggestions are advisory text;
- that `workflow` does not execute commands;
- that SpecHarbor does not commit, push, create PRs, merge, call GitHub/GitLab APIs, inspect CI, call provider APIs, call agent CLIs, or run external commands;
- examples of workflow output;
- that status and next-step detection are intentionally not implemented in this first workflow foundation.

## Testing

Tests should be focused and local.

Domain tests:

- workflow step ids are stable;
- workflow step ordering is deterministic;
- display names are correct;
- manual vs agent-assisted classification is correct;
- supported/advisory flags are correct;
- required predecessor dependencies are correct;
- command suggestion metadata is correct;
- role step ids align with canonical prompt roles.

Use case tests:

- workflow show returns the ordered recommended workflow;
- workflow includes all required steps;
- workflow includes advisory command suggestions;
- workflow includes safety notes for commit, pull request, and merge;
- workflow show does not require filesystem, provider, source-control, workflow, agent, network, or process collaborators;
- workflow show does not execute commands.

CLI tests:

- `workflow` parses correctly;
- `workflow` prints deterministic output;
- `workflow` returns zero on successful read-only output;
- unsupported flags are rejected clearly;
- extra arguments are rejected clearly;
- `workflow status <change-id>` is rejected as an unexpected argument in this change;
- output contains the workflow title, ordered steps, modes, supported/advisory indicators, command suggestions, safety notes, and global advisory notes.

Architecture tests:

- core does not import adapters;
- core does not import CLI packages;
- core does not perform terminal IO;
- workflow logic stays deterministic;
- no provider, network, source-control, workflow, or agent SDKs are introduced;
- workflow feature does not write files;
- workflow feature does not execute external commands.

Regression tests:

- `init` remains unchanged;
- `scan` remains unchanged;
- generation modes remain unchanged;
- `validate` remains unchanged;
- `prompt` remains unchanged;
- `review` remains unchanged;
- `archive` remains unchanged;
- `config` remains unchanged.

Documentation tests may be added if the existing architecture/documentation test pattern is extended.

## Safety Boundaries

The first workflow integration must be read-only and advisory.

SpecHarbor must not:

- commit;
- stage files;
- push;
- create pull requests;
- merge pull requests;
- approve pull requests;
- call GitHub APIs;
- call GitLab APIs;
- call provider APIs;
- call local model APIs;
- call agent CLIs;
- execute external commands;
- inspect remote branches;
- inspect CI;
- trigger CI;
- trigger workflows;
- modify OpenSpec files;
- modify production code;
- change `tasks.md` checkboxes automatically;
- archive automatically;
- add OAuth;
- store credentials;
- add workflow connector adapters.

Any future GitHub, source-control, CI, or workflow automation must be proposed in a separate OpenSpec change.
