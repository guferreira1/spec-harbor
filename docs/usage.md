# Usage

Run examples from the repository root while developing SpecHarbor:

```bash
go run ./cmd/specharbor help
```

You can also build a local binary:

```bash
go build ./cmd/specharbor
```

That creates a local `specharbor` binary in the repository root.

## Implemented Commands

### Initialize a Project

```bash
go run ./cmd/specharbor init
```

`init` creates the OpenSpec and SpecHarbor project files that are missing in the current directory. Existing files are skipped.

### Scan a Project

```bash
go run ./cmd/specharbor scan
```

`scan` performs an informational local project scan. It is stack-agnostic: user projects do not need to be Go projects.

The report can include detected ecosystems, package managers, test command hints, CI signals, container or deployment signals, SpecHarbor/OpenSpec signals, and notes. The command does not require flags or arguments.

### Show the Recommended Workflow

```bash
go run ./cmd/specharbor workflow
```

The installed command form is `specharbor workflow`. It prints the recommended OpenSpec/SDD workflow as read-only advisory text:

```text
SpecHarbor recommended workflow.
Title: OpenSpec/SDD agent-driven workflow

Steps:
1. spec-author - Spec Author Agent
2. architecture-reviewer - Architecture Reviewer Agent
3. implementer - Implementer Agent
4. test-engineer - Test Engineer Agent
5. change-reviewer - Change Reviewer Agent
6. commit - Commit
7. pull-request - Pull Request
8. merge - Merge
9. archive - Archive
```

The full output includes each step id, display name, purpose, mode, supported/advisory indicators, predecessor step ids, command suggestions, and safety notes. The suggestions connect the workflow to existing commands:

- `generate` creates or starts an OpenSpec change package.
- `validate` checks required OpenSpec change files and their content quality.
- `prompt --role ...` prints prompts for Spec Author Agent, Architecture Reviewer Agent, Implementer Agent, Test Engineer Agent, and Change Reviewer Agent.
- `review` checks the local change package and task checkbox completion.
- `archive` explicitly moves an accepted change to the archive.

Command suggestions are advisory and `specharbor workflow` does not execute them. Commit, Pull Request, and Merge remain manual steps; SpecHarbor does not commit, does not create PRs, and does not merge. This command is read-only and does not call GitHub, GitLab, CI, provider APIs, agent CLIs, source-control automation, workflow execution, external commands, or remote automation.

### Generate a Change

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

`generate <change-id> --blank` creates the expected OpenSpec change structure with blank/manual starter content.

Built-in template generation uses the same command with `--template <template-name>` and deterministic built-in starter content:

```bash
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

For example:

```bash
go run ./cmd/specharbor generate add-example-feature --template feature
```

Supported built-in templates are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Guided generation uses explicit CLI flags:

```bash
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

Guided generation is deterministic, local, and non-interactive. It does not prompt during command execution; it uses the supplied `--type`, `--title`, and `--summary` values.

Supported guided types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Agent-assisted spec authoring uses explicit CLI flags:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

For example:

```bash
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support"
```

Supported agent-assisted authoring types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Dry-run remains the default. Without `--execute`, agent-assisted spec authoring prints a deterministic authoring plan to stdout and prints a deterministic, copy-pasteable prompt to stdout.

The generated prompt is meant to help an external agent author or refine only the OpenSpec change package. Implementation remains a later step through the normal SpecHarbor workflow.

Dry-run agent-assisted spec authoring:

- writes no files;
- writes no prompt file;
- does not create or modify OpenSpec files;
- does not create or modify production code;
- does not execute agents or local agent commands;
- does not require a runner;
- does not resolve executable command mappings;
- does not call provider APIs;
- does not call local models;
- does not call network APIs;
- does not call source-control APIs;
- does not call workflow tools.

Recognized agent targets for dry-run are:

- `codex` - Codex
- `claude` - Claude Code
- `devin` - Devin
- `cursor` - Cursor
- `copilot` - GitHub Copilot
- `gemini` - Gemini CLI
- `roo` - Roo Code
- `windsurf` - Windsurf
- `aider` - Aider
- `generic` - Generic Agent

Unknown dry-run agents are rejected as an intentional validation tightening.

`--execute` is explicit and is supported only with `--agent-assisted`:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>" --execute
```

Execute mode sends the same deterministic OpenSpec authoring prompt through stdin to a supported local command. The runner working directory is the current project root.

Supported executable local command mappings are:

- `codex -> codex`
- `claude -> claude`
- `devin -> devin`
- `cursor -> cursor`
- `copilot -> copilot`
- `gemini -> gemini`
- `roo -> roo`
- `windsurf -> windsurf`
- `aider -> aider`

`generic` is recognized for dry-run only. `--execute --agent generic` fails because generic execution requires a future config-driven command mapping.

Execute mode is run-and-report only:

- missing local commands produce startup errors with no runner result and no exit code;
- started commands with non-zero exit codes produce a full report, then SpecHarbor exits non-zero;
- stdout, stderr, exit code, and execution status are captured for started processes;
- SpecHarbor does not parse or apply output;
- stdout and stderr are displayed only, not parsed or applied;
- SpecHarbor does not write OpenSpec files from runner output;
- SpecHarbor does not modify production code from runner output;
- SpecHarbor does not auto-commit, auto-push, or auto-merge.

Provider APIs, IDE automation, OAuth, credentials, marketplace integrations, remote execution, source-control automation, and workflow automation remain out of scope. Local agent command behavior is controlled by the installed local tool.

Blank, built-in template, and guided generation create the same required OpenSpec change files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Built-in template content is deterministic, local, and generic starter content. Guided content is deterministic, local starter content that includes the supplied title and summary.

Generated content is safe to edit after generation. Guided output does not mean SpecHarbor inferred project-specific requirements beyond the provided type, title, and summary.

Existing files are skipped and are not overwritten. If a change directory partially exists, running generation again recovers it by creating only the missing required files.

Copy-pasteable examples from the repository root:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor generate add-example-feature --template feature
go run ./cmd/specharbor generate fix-example-bug --template bugfix
go run ./cmd/specharbor generate update-example-docs --template docs
go run ./cmd/specharbor generate refactor-example-flow --template refactor
go run ./cmd/specharbor generate add-guided-feature --guided --type feature --title "Add guided feature" --summary "Create a guided OpenSpec change from explicit CLI inputs."
go run ./cmd/specharbor generate fix-guided-bug --guided --type bugfix --title "Fix guided bug" --summary "Describe the bugfix using deterministic guided starter content."
go run ./cmd/specharbor generate update-guided-docs --guided --type docs --title "Update guided docs" --summary "Document guided generation as implemented behavior."
go run ./cmd/specharbor generate refactor-guided-flow --guided --type refactor --title "Refactor guided flow" --summary "Describe a behavior-preserving refactor with explicit context."
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support"
go run ./cmd/specharbor generate add-reports --agent-assisted --agent codex --type feature --title "Add reports" --summary "Create report generation support" --execute
```

### Validate a Change

```bash
go run ./cmd/specharbor validate add-example-feature
```

`validate <change-id>` runs deterministic, local, read-only validation over the change package under `openspec/changes/<change-id>/`. It never writes files, never modifies the change, and never calls the network, AI providers, agents, or source-control tooling.

It checks the five required OpenSpec files:

- `proposal.md`
- `design.md`
- `tasks.md`
- `acceptance-criteria.md`
- `risks.md`

Every finding carries a severity, a stable snake_case rule code, a message (with line numbers where relevant), and the file path.

Error findings (the change package is not usable downstream):

- `project_root_unavailable` - the OpenSpec project structure is missing.
- `change_directory_missing` - the change id is well formed but unknown.
- `required_file_missing` - one of the five required files is missing.
- `file_empty` - a required file is empty or whitespace-only. Other content findings for that file are suppressed.
- `file_missing_heading` - a required file has no markdown heading.
- `file_missing_body` - a required file contains only headings.
- `tasks_checkbox_missing` - `tasks.md` has no valid checkbox task.
- `tasks_checkbox_malformed` - a checkbox-like line (for example `- []`, `-[ ]`, `- [y]`, or `- [x]` without text) breaks the `- [ ] text` grammar; reported with line numbers.
- `acceptance_criteria_item_missing` - `acceptance-criteria.md` has no list or checkbox item with meaningful text; placeholder-only items (`N/A`, `...`, `?`, `TBD`, `TODO`, `FIXME`) do not count.

Warning findings (quality gaps that never fail the command):

- `placeholder_content` - standalone `TBD`/`TODO`/`FIXME`, placeholder-only list items (`N/A`, `...`, `?`), or `lorem ipsum`.
- `boilerplate_only_content` - the file still contains only known starter/boilerplate guidance lines and was never meaningfully edited.
- `proposal_section_missing` - no Problem, Goal, or Summary section.
- `design_section_missing` - no Overview, Approach, Design, Architecture, Technical Decisions, or Decisions section.
- `tasks_phase_heading_missing` - checkbox tasks exist but no level-2 phase heading.
- `tasks_all_completed` - every checkbox task is checked; confirm implementation evidence before review.
- `risks_mitigation_missing` - risks are listed without mitigation notes.
- `design_architecture_section_missing` - change files mention `internal/core`, `internal/adapters`, or `internal/platform` but `design.md` has no Architecture section.
- `tasks_documentation_task_missing` - `proposal.md` or `design.md` references the `specharbor` CLI but `tasks.md` has no documentation task.

Severity drives status and exit codes:

- Zero error findings: status `valid`, exit code `0`. Warnings alone never fail the command.
- One or more error findings: status `invalid`, exit code `1`.
- Missing or unsafe change ids (`..`, separators, absolute paths, leading `.` or `-`, characters outside `[A-Za-z0-9._-]`, more than 128 characters) are rejected with a clear command error before any filesystem access. Internal single dots such as `change.v1` are accepted.

Example valid output:

```text
SpecHarbor change is valid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature
Required files: 5
Errors: 0
Warnings: 0
```

Example valid output with warnings (exit code `0`):

```text
SpecHarbor change is valid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature
Required files: 5
Errors: 0
Warnings: 2

Warnings:
- [warning] placeholder_content: Placeholder marker "TBD" found (line 12) (openspec/changes/add-example-feature/design.md)
- [warning] risks_mitigation_missing: Risks are listed without mitigation notes. (openspec/changes/add-example-feature/risks.md)
```

Example invalid output (exit code `1`):

```text
SpecHarbor change is invalid.
Change: add-example-feature
Checked path: openspec/changes/add-example-feature

Errors:
- [error] required_file_missing: Missing required file: design.md (openspec/changes/add-example-feature/design.md)
- [error] tasks_checkbox_missing: No checkbox tasks found. (openspec/changes/add-example-feature/tasks.md)

Warnings:
- [warning] proposal_section_missing: No Problem, Goal, or Summary section found. (openspec/changes/add-example-feature/proposal.md)
```

Run validation before implementation (to confirm the change package is structurally ready), before review (to catch content gaps cheaply), and before opening a PR (to keep low-quality packages out of the shared workflow).

Intentional behavior changes from the earlier presence-only validation:

- Unusable change packages now fail: empty required files, files without a heading or body, tasks files without valid checkboxes, malformed checkbox lines, and acceptance-criteria files without a meaningful item produce errors and a non-zero exit.
- A freshly generated `--blank` or template change now validates as valid with `boilerplate_only_content` (and applicable `placeholder_content`) warnings instead of zero findings; the exit code stays `0`, so the documented `generate -> validate` flow keeps working.
- Change ids are validated more strictly and rejected before any filesystem access.
- The report replaces the single `Findings:` count with `Errors:` and `Warnings:` counts and groups findings by severity with file paths appended.

### Generate a Role Prompt

```bash
go run ./cmd/specharbor prompt add-example-feature --role implementer
```

`prompt <change-id> --role <role>` prints an agent prompt for an existing OpenSpec change.

Supported roles:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

Use `--role` for prompt roles. Agent-target flags are not implemented.

### Review a Change

```bash
go run ./cmd/specharbor review add-example-feature
```

`review <change-id>` reviews the change package and task completion state. It exits non-zero when the review status is not approved.

### Archive a Change

```bash
go run ./cmd/specharbor archive add-example-feature
```

`archive <change-id>` moves a completed change from `openspec/changes/<change-id>/` to a dated archive path under `openspec/archive/<date>/<change-id>/`.

### Show Local Config

```bash
go run ./cmd/specharbor config show
go run ./cmd/specharbor config
```

`config show` and `config` are read-only. They read `.specharbor/config.yml` from the current project, support config version `1`, and print the local config report.

Current config behavior does not write config files and does not implement:

- `config get`
- `config set`
- `config unset`

## Normal Flow

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

For the detailed recommended nine-step workflow, run:

```bash
go run ./cmd/specharbor workflow
```

A typical local sequence is:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
go run ./cmd/specharbor validate add-example-feature
go run ./cmd/specharbor prompt add-example-feature --role implementer
go run ./cmd/specharbor review add-example-feature
go run ./cmd/specharbor archive add-example-feature
```

Do not archive a change until the implementation is complete and reviewed.

## Planned Behavior

The following items are product direction, not implemented command behavior:

- AI-assisted generation;
- hybrid generation;
- custom, remote, and config-driven templates;
- config-driven generic runner commands;
- interactive generation prompts;
- AI provider setup;
- provider API key management;
- config mutation commands;
- external workflow connectors.
