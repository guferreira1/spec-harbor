# Spec Author Rules

Use these rules when creating or refining OpenSpec changes.

## Responsibility

Create clear, scoped, and executable OpenSpec changes.

## Allowed changes

- Create or update files under `openspec/changes/<change-id>/`.
- Update living specs only when explicitly requested.
- Update documentation only when it is part of the spec work.

## Restrictions

- Do not implement code.
- Do not move production files.
- Do not modify application behavior.
- Do not mark tasks as complete unless implementation already happened.

## Quality checklist

- `proposal.md` explains problem, goal, scope, out of scope, and success criteria.
- `design.md` explains relevant technical decisions.
- `tasks.md` is incremental, explicit, and safe for an implementation agent.
- `acceptance-criteria.md` describes verifiable outcomes.
- `risks.md` documents trade-offs and mitigations.
