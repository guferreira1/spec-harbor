# Workflow

SpecHarbor is built around an OpenSpec change package. The goal is to make the task explicit before an agent or contributor starts changing code.

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

## Recommended Workflow Command

Run the read-only workflow guide with:

```bash
go run ./cmd/specharbor workflow
```

The installed command form is `specharbor workflow`. It prints the recommended nine-step OpenSpec/SDD workflow:

1. Spec Author Agent
2. Architecture Reviewer Agent
3. Implementer Agent
4. Test Engineer Agent
5. Change Reviewer Agent
6. Commit
7. Pull Request
8. Merge
9. Archive

Abbreviated output shape:

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
   - specharbor generate <change-id> --guided ...
   - specharbor prompt <change-id> --role spec-author

6. commit - Commit
   Mode: manual
   Supported by SpecHarbor: no
   Advisory only: yes
   Commands:
   - none
```

The command suggestions are advisory. `specharbor workflow` does not execute commands, does not inspect local workflow status, and does not decide the next step. Status and next-step detection are intentionally deferred.

The workflow relates to existing commands as follows:

- `context discover` optionally inspects bounded local repository sources and reports classified context signals before briefing or authoring.
- `brief` optionally collects confirmed project context before authoring when repository context is missing or ambiguous.
- `generate` creates or starts the OpenSpec change package for the Spec Author Agent.
- `validate` checks required OpenSpec change files before review or implementation.
- `prompt --role ...` prints prompts for Spec Author Agent, Architecture Reviewer Agent, Implementer Agent, Test Engineer Agent, and Change Reviewer Agent.
- `review` checks local task checkbox completion and required change files.
- `archive` explicitly archives an accepted change.

`specharbor context discover` is read-only local discovery. It labels repository evidence as detected facts, conventional guesses as suggested assumptions, and existing project brief values as user-confirmed context. It does not execute commands, call provider APIs, perform remote discovery, build an index, or run RAG.

`specharbor brief` writes `.specharbor/project-brief.md` only after interactive confirmation. `specharbor brief --update` is the explicit maintenance path for an existing brief: it keeps confirmed values by default, shows detected facts and assumptions as review items, previews changes, and writes only after final confirmation. Briefing is explicit context collection and maintenance, not repository indexing, RAG, provider integration, agent execution, source-control automation, or remote automation. Role prompt generation can include confirmed brief values and discovered local signals in a bounded `## Project Context` section.

Commit, Pull Request, and Merge remain manual. SpecHarbor does not commit, does not push, does not create PRs, does not merge, does not call GitHub, does not call GitLab, does not inspect CI, does not call provider APIs, does not call agent CLIs, does not run source-control automation, does not run workflow execution, and does not perform remote automation.

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

Role prompts point an agent at the repository rules, architecture spec, active change, and any available classified Project Context. The context section keeps user-confirmed context, detected facts, and suggested assumptions separate; assumptions remain assumptions. Agent-assisted workflows do not require provider API keys because SpecHarbor prints prompts for external coding agents to consume, without executing commands or agents.

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
