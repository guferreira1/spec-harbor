# Proposal: Add Role Rules

## Problem

Role prompts duplicate operational rules and make manual prompts longer.

## Goal

Centralize global and role-specific rules under `.specharbor/rules/`.

## Scope

- Add global rules.
- Add role-specific rules.
- Update role prompts to reference rules.
- Update project context to mention rules.

## Out of Scope

- CLI rendering.
- Go code changes.
- External integrations.

## Success Criteria

- Rules exist.
- Role prompts reference rules.
- Project context mentions rules.
