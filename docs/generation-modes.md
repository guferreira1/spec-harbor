# Generation Modes

SpecHarbor currently implements blank OpenSpec change generation only.

## Implemented

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

Blank generation creates the OpenSpec change file structure so the user can write the content manually.

## Planned

The following generation modes are product direction, not implemented behavior:

- guided generation;
- template generation;
- AI-assisted generation;
- agent-assisted generation;
- hybrid generation.

Agent-assisted workflows are intended to avoid requiring provider API keys, but detailed provider or agent setup is not part of the current implemented command set.
