# Proposal: Add Agent Role Templates

## Problem

SpecHarbor needs reusable prompts for different AI-assisted development responsibilities. Without role-specific templates, one prompt can mix specification, implementation, testing, and review responsibilities.

## Goal

Add role-based prompt templates for common OpenSpec workflows.

## Scope

- Document the agent role model.
- Add prompt templates for spec authoring, architecture review, implementation, testing, and change review.
- Keep templates generic enough for Codex, Claude Code, Devin, Cursor, and other coding agents.

## Out of Scope

- Implementing prompt rendering logic.
- Implementing CLI commands for role prompt generation.
- Integrating with external agents or workflow tools.

## Success Criteria

- The repository documents the available agent roles.
- Each role has a reusable Markdown template.
- Templates reinforce OpenSpec, architecture, and scoped execution rules.
