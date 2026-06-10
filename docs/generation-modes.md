# Generation Modes

SpecHarbor currently implements blank generation, built-in template generation, project-local custom template generation, config-driven template aliases, guided generation, AI-assisted from-file generation, dry-run agent-assisted spec authoring, and explicit agent-assisted local runner execution in run-and-report mode.

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

### Custom Template Generation

```bash
go run ./cmd/specharbor generate <change-id> --custom-template <template-name>
```

Custom template generation renders reusable, project-local OpenSpec change templates. A custom template is a plain directory under the fixed root `.specharbor/templates/` containing all five required OpenSpec change files:

```text
.specharbor/templates/<template-name>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

All five files are required and must be non-empty. A missing template directory, missing required files, or an empty (whitespace-only) file fails with a clear error and produces zero writes — not even the change directory is created. Unknown extra files and subdirectories inside the template directory are ignored and never copied.

Custom template names are validated before any filesystem access: allowed characters are `[A-Za-z0-9._-]`, names must be a single path segment with no `/` or `\`, no `..` sequences, no leading `.` or `-`, and at most 128 characters.

Template content supports minimal deterministic variable substitution only:

- `{{change_id}}` is always replaced with the change id.
- `{{title}}` is replaced only when the optional `--title` flag is provided.
- `{{summary}}` is replaced only when the optional `--summary` flag is provided.
- Unresolved and unknown `{{...}}` tokens remain in the output verbatim; they are never an error.

```bash
go run ./cmd/specharbor generate <change-id> --custom-template <template-name> --title "<title>" --summary "<summary>"
```

There is no templating language: no conditionals, no loops, no functions, no includes, and no template-defined variables. Templates are static Markdown content and are never executed; no scripts, hooks, shell commands, or external processes run during generation.

Built-in and custom templates resolve from disjoint sources:

- `--template <name>` resolves only the four built-in templates; its behavior, content, output, and unknown-name errors are unchanged.
- `--custom-template <name>` resolves only `.specharbor/templates/<name>/`.
- `--config-template <alias>` resolves only aliases declared in `.specharbor/config.yml`.
- A custom template sharing a built-in name (for example `feature`) never shadows or overrides the built-in template.

`--custom-template` is mutually exclusive with `--blank`, `--template`, `--guided`, and `--agent-assisted`. The success report identifies the template as custom, shows the relative template source path, lists created and skipped files, and notes that only OpenSpec change files were written.

Generated changes work with `specharbor validate <change-id>` like any other change; generation does not run validation automatically, and validation findings depend on the template's content quality.

Safety boundaries: templates are local and project-scoped; there are no remote templates, no marketplace, no network or provider calls, no credentials, and no writes outside `openspec/changes/<change-id>/`. Existing files are skipped, never overwritten.

### Config-Driven Template Aliases

```bash
go run ./cmd/specharbor generate <change-id> --config-template <alias>
go run ./cmd/specharbor generate <change-id> --config-template <alias> --title "<title>" --summary "<summary>"
```

Config-driven templates let a project define stable aliases for existing built-in or custom templates. The alias map lives in `.specharbor/config.yml`:

```yaml
version: 1

templates:
  aliases:
    api-feature:
      source: custom
      template: api-feature

    default-feature:
      source: builtin
      template: feature
```

`version: 1` is required for config-driven generation. Omitted `templates` or omitted `templates.aliases` means there are no aliases. A missing config file, missing version, unsupported version, invalid YAML, invalid alias entry, or missing requested alias returns a clear error.

Supported source kinds are exactly:

- `builtin`, resolving the named built-in template only.
- `custom`, resolving `.specharbor/templates/<template-name>/` only.

Aliases are safe single path segments: non-empty, at most 128 characters, allowed characters `[A-Za-z0-9._-]`, no `/` or `\`, no absolute paths, no traversal or `..` sequence, no leading `.` or `-`. Invalid CLI aliases fail before template resolution, and invalid alias names in config fail during config validation.

Config-driven generation delegates to the resolved source behavior:

- A built-in alias produces the same files as direct `--template <name>`.
- A custom alias uses the same custom-template directory, required-file validation, and deterministic `{{change_id}}`, `{{title}}`, and `{{summary}}` substitution as direct `--custom-template <name>`.
- Optional `--title` and `--summary` pass through to the resolved generation behavior; built-in templates do not use those values.

Namespaces are disjoint by flag. `--template feature`, `--custom-template feature`, and `--config-template feature` are three different lookups. There is no shadowing, fallback, or source inference.

`--config-template` is mutually exclusive with `--blank`, `--template`, `--custom-template`, `--guided`, `--agent-assisted`, `--ai-assisted`, and `--execute`. It also does not accept `--type`, `--agent`, `--from-file`, or `--overwrite`.

Safety boundaries: config aliases support no `local`, `remote`, `url`, arbitrary `path`, marketplace templates, template scripts, shell execution, network/provider behavior, production code writes, source-control automation, or archive automation. Generated files stay under `openspec/changes/<change-id>/`, use the five required OpenSpec filenames, and skip existing files without overwriting.

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

### AI-Assisted From-File Generation

```bash
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file>
go run ./cmd/specharbor generate <change-id> --ai-assisted --from-file <agent-output-file> --overwrite
```

AI-assisted generation imports AI-authored OpenSpec content from a local file that the user explicitly supplies. The file must use strict delimiter blocks:

```text
---FILE: proposal.md---
...
---END FILE---
---FILE: design.md---
...
---END FILE---
---FILE: tasks.md---
...
---END FILE---
---FILE: acceptance-criteria.md---
...
---END FILE---
---FILE: risks.md---
...
---END FILE---
```

Allowed block filenames are exactly `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`. All five are required. Unknown filenames, duplicates, missing blocks, empty content, absolute paths, traversal, nested paths, malformed delimiters, fenced wrapper formats, patch/diff formats, and text outside blocks are rejected before writes.

The command parses the complete source before any target file write, creates `openspec/changes/<change-id>/` when needed, writes only the five approved filenames under that directory, and runs validation after successful writes or skips. Existing files are skipped by default; `--overwrite` is required to replace them. Existing symlink output targets are rejected instead of followed. Validation warnings keep exit code `0`; validation errors are reported and make the command exit non-zero.

AI-assisted from-file generation is not provider-backed generation and is not live runner output application. It does not call provider APIs, remote AI services, local model APIs, OAuth, credentials, shell commands, source-control automation, workflow automation, or agent runners. It does not modify production code, does not apply patches, and does not auto-commit, auto-push, create PRs, merge, or archive.

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

Blank, built-in template, custom template, guided, and AI-assisted generation create the required OpenSpec change files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Existing files are skipped and are not overwritten for blank, built-in template, custom template, and guided generation. AI-assisted generation also skips existing files by default, and replaces them only with explicit `--overwrite`. Partially existing change directories are recoverable because generation creates only missing required files.

## Planned

The following items are product direction, not implemented command behavior:

- hybrid generation;
- remote templates;
- config-driven generic runner commands;
- interactive prompts.

Detailed provider setup, IDE automation, marketplace integrations, remote execution, workflow automation, and file-application behavior are not part of the current implemented generation command set.
