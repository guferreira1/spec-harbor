# Agent Roles

SpecHarbor uses role-based prompt templates so different coding-agent sessions can work from the same OpenSpec change without taking the same responsibilities.

Generate a role prompt from the repository root:

```bash
go run ./cmd/specharbor prompt implement-config-foundation --role implementer
```

Supported role names:

- `spec-author`
- `architecture-reviewer`
- `implementer`
- `test-engineer`
- `change-reviewer`

These five ids are also the agent-assisted step ids shown by `specharbor workflow`.

Generated role prompts may include a `## Project Context` section when confirmed or discovered local context is available. The section separates user-confirmed context, detected facts, and suggested assumptions. Confirmed context from `.specharbor/project-brief.md` takes precedence over detected facts; assumptions are labeled as assumptions and are not facts. The context section is bounded, includes source/confidence evidence for detected facts and assumptions, and does not execute commands, run agents, call provider APIs, perform RAG, index the repository, or use remote discovery.

## Spec Author Agent

Creates or refines OpenSpec change files. This role focuses on `proposal.md`, `design.md`, `tasks.md`, `acceptance-criteria.md`, and `risks.md`; it should not implement the change.

## Architecture Reviewer Agent

Reviews specs or diffs against the architecture contract. This role checks whether the proposed work respects boundaries such as domain, ports, use cases, adapters, and CLI responsibilities.

## Implementer Agent

Applies an approved OpenSpec change. This role should read the active change, keep edits inside the described scope, update `tasks.md` only for completed work, and run the requested verification.

## Test Engineer Agent

Adds or updates focused tests when the active change calls for test work. This role should keep tests aligned with the behavior being specified and avoid unrelated coverage churn.

## Change Reviewer Agent

Reviews the final diff against the active OpenSpec change. This role should prioritize scope drift, architecture violations, stale task status, missing verification, and claims that do not match the implementation.

## Recommended flow

```text
Spec Author -> Architecture Reviewer -> Implementer -> Test Engineer -> Change Reviewer
```

The broader recommended workflow continues manually through Commit, Pull Request, Merge, and Archive. `specharbor workflow` shows that full advisory sequence without executing commands or automating source-control steps.

Pull Request and Archive are workflow steps, not supported prompt roles in this version.

For small changes:

```text
Spec Author -> Implementer -> Change Reviewer
```

Detailed global and role-specific rules live in `.specharbor/rules/`. The docs should link to those rules instead of copying every instruction into this page.

Agent-assisted workflows do not require provider API keys. SpecHarbor generates prompts that can be pasted into an external coding-agent tool.
