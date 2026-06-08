# Generation Modes

SpecHarbor currently implements blank generation and built-in template generation.

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

Both implemented generation modes create the required OpenSpec change files:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

Built-in template generation writes deterministic, local, generic starter content for the selected template. The generated content is safe to edit and does not mean SpecHarbor inferred project-specific requirements.

Existing files are skipped and are not overwritten. Partially existing change directories are recoverable because generation creates only missing required files.

## Planned

The following items are product direction, not implemented command behavior:

- guided generation;
- AI-assisted generation;
- agent-assisted generation;
- hybrid generation;
- custom templates;
- remote templates;
- config-driven templates;
- interactive prompts.

Agent-assisted generation is intended to avoid requiring provider API keys, but detailed provider or agent setup is not part of the current implemented generation command set.
