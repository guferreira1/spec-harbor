# Agent Roles

SpecHarbor uses role-based prompt templates to keep AI-assisted development controlled.

## Roles

- Spec Author: creates or refines OpenSpec changes only.
- Architecture Reviewer: checks specs or diffs against architecture rules.
- Implementer: applies an approved OpenSpec change.
- Test Engineer: adds or updates focused tests.
- Change Reviewer: reviews the final diff against the active change.

## Recommended flow

```text
Spec Author -> Architecture Reviewer -> Implementer -> Test Engineer -> Change Reviewer
```

For small changes:

```text
Spec Author -> Implementer -> Change Reviewer
```
