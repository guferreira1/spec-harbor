# Generation Modes

SpecHarbor currently implements blank generation, built-in template generation, guided generation, dry-run agent-assisted spec authoring, and explicit agent-assisted local runner execution in run-and-report mode.

## Implemented

### Blank Generation

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

Blank generation creates the OpenSpec change file structure so the user can write the content manually.

### Built-In Template Generation

```bash
go run ./cmd/specharbor generate <change-id> --template <template-name>
```

Implemented built-in templates are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Built-in template generation writes deterministic, local, generic starter content for the selected template.

### Guided Generation

```bash
go run ./cmd/specharbor generate <change-id> --guided --type <type> --title "<title>" --summary "<summary>"
```

Implemented guided types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Guided generation writes deterministic, local starter content based on explicit CLI inputs. It is non-interactive and does not prompt during command execution.

Guided generated content includes the supplied title and summary. The generated content is safe to edit and does not mean SpecHarbor inferred project-specific requirements beyond the provided inputs.

### Agent-Assisted Spec Authoring

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

Implemented agent-assisted authoring types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Dry-run remains the default. Without `--execute`, agent-assisted spec authoring prints a deterministic authoring plan and a deterministic, copy-pasteable prompt to stdout.

The generated prompt is meant to help an external agent author or refine only the OpenSpec change package. Implementation remains a later step through the normal SpecHarbor workflow.

Dry-run agent-assisted spec authoring:

- writes no files;
- writes no prompt file;
- does not create or modify OpenSpec files;
- does not create or modify production code;
- does not execute agents;
- does not require a runner;
- does not resolve executable command mappings;
- does not call provider APIs;
- does not call local models;
- does not call network APIs;
- does not call source-control APIs;
- does not call workflow tools.

Recognized agent targets are:

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

Execute mode is explicit:

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>" --execute
```

Execute mode sends the same deterministic OpenSpec authoring prompt through stdin to a supported local command. The working directory is the current project root.

Executable local command mappings are:

- `codex -> codex`
- `claude -> claude`
- `devin -> devin`
- `cursor -> cursor`
- `copilot -> copilot`
- `gemini -> gemini`
- `roo -> roo`
- `windsurf -> windsurf`
- `aider -> aider`

`generic` is dry-run-only until a future config-driven runner/templates feature defines a deterministic command mapping. `--execute --agent generic` fails with a clear unmapped-target error.

Execute mode is still run-and-report only:

- missing local commands produce startup errors with no runner result and no exit code;
- started commands with non-zero exit codes produce a report and then SpecHarbor exits non-zero;
- stdout, stderr, exit code, and execution status are captured and reported for started processes;
- SpecHarbor does not parse or apply output;
- SpecHarbor does not write files from output;
- SpecHarbor does not modify production code from output;
- SpecHarbor does not auto-commit, auto-push, or auto-merge.

Provider APIs, IDE automation, OAuth, credentials, marketplace integrations, remote execution, source-control automation, and workflow automation remain out of scope. Local agent command behavior is controlled by the installed local tool.

Blank, built-in template, and guided generation create the required OpenSpec change files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Existing files are skipped and are not overwritten. Partially existing change directories are recoverable because generation creates only missing required files.

## Planned

The following items are product direction, not implemented command behavior:

- AI-assisted generation;
- hybrid generation;
- custom templates;
- remote templates;
- config-driven templates;
- config-driven generic runner commands;
- interactive prompts.

Detailed provider setup, IDE automation, marketplace integrations, remote execution, workflow automation, and file-application behavior are not part of the current implemented generation command set.
