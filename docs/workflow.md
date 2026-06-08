# Workflow

SpecHarbor is built around an OpenSpec change package. The goal is to make the task explicit before an agent or contributor starts changing code.

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

## Change Package

A change lives under `openspec/changes/<change-id>/`:

```text
openspec/changes/<change-id>/
  proposal.md
  design.md
  tasks.md
  acceptance-criteria.md
  risks.md
```

File responsibilities:

- `proposal.md`: states the problem, goal, scope, out-of-scope work, and success criteria.
- `design.md`: describes the technical approach and important tradeoffs.
- `tasks.md`: lists the implementation and verification work.
- `acceptance-criteria.md`: defines observable conditions for completion.
- `risks.md`: records known risks and mitigations.

The files are not meant to be ceremony. They constrain implementation scope, make review easier, and give coding agents concrete work to follow.

## From Idea to Change

Start with a change id that describes the work:

```bash
go run ./cmd/specharbor generate add-example-feature --blank
```

Fill in the generated files before meaningful implementation work begins. The blank generator creates structure only; it does not infer requirements.

## Validate Before Implementation

```bash
go run ./cmd/specharbor validate add-example-feature
```

Validation checks that the required OpenSpec files exist. It does not prove the design is correct.

## Generate a Role Prompt

```bash
go run ./cmd/specharbor prompt add-example-feature --role implementer
```

Role prompts point an agent at the repository rules, architecture spec, and active change. Agent-assisted workflows do not require provider API keys because SpecHarbor prints prompts for external coding agents to consume.

## Implement and Review

Implementation should stay inside the active change scope. For this repository, Go code should stay within the architecture boundaries described in `openspec/specs/architecture/spec.md`.

After implementation:

```bash
go run ./cmd/specharbor review add-example-feature
go test ./...
```

Inspect the diff before finalizing work. Update `tasks.md` only for work actually completed.

## Archive

Archive only after the change is complete:

```bash
go run ./cmd/specharbor archive add-example-feature
```

Archiving moves the active change into the dated archive area so completed work is not confused with active work.

## Dogfooding

SpecHarbor uses OpenSpec changes for its own development. Meaningful work in this repository should start with an OpenSpec package under `openspec/changes/`, and implementation should remain scoped to that package.
