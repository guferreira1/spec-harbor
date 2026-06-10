# Generation Modes

SpecHarbor currently implements blank generation, built-in template generation, guided generation, and dry-run agent-assisted spec authoring.

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

### Dry-Run Agent-Assisted Spec Authoring

```bash
go run ./cmd/specharbor generate <change-id> --agent-assisted --agent <agent-name> --type <type> --title "<title>" --summary "<summary>"
```

Implemented agent-assisted authoring types are exactly:

- `feature`
- `bugfix`
- `docs`
- `refactor`

Agent-assisted spec authoring is dry-run only in this first version. It prints a deterministic authoring plan and a deterministic, copy-pasteable prompt to stdout.

The generated prompt is meant to help an external agent author or refine only the OpenSpec change package. Implementation remains a later step through the normal SpecHarbor workflow.

Dry-run agent-assisted spec authoring:

- writes no files;
- writes no prompt file;
- does not create or modify OpenSpec files;
- does not create or modify production code;
- does not execute agents;
- does not call provider APIs;
- does not call local models;
- does not call network APIs;
- does not call source-control APIs;
- does not call workflow tools.

`--execute` is currently unsupported for agent-assisted spec authoring and returns a clear error.

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
- future agent execution and non-dry-run agent-assisted workflows;
- hybrid generation;
- custom templates;
- remote templates;
- config-driven templates;
- interactive prompts.

Detailed provider setup, agent setup, workflow automation, and non-dry-run agent-assisted behavior are not part of the current implemented generation command set.
