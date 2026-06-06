# Proposal: Define Project Architecture

## Problem

SpecHarbor has a broad roadmap involving OpenSpec workflows, spec authoring, AI providers, coding agents, workflow connectors, validators, templates, and project scanners.

Without a formal architecture specification, future changes may introduce inconsistent package boundaries, duplicated logic, provider-specific coupling, large switch statements, or business rules inside adapters.

## Goal

Define the official architecture rules for SpecHarbor using Hexagonal Architecture, SOLID, Clean Code, and pragmatic design patterns.

## Scope

- Define architectural layers.
- Define dependency rules.
- Define SOLID rules.
- Define required design patterns.
- Define spec authoring strategies.
- Separate AI providers, agent targets, and workflow connectors.
- Define package structure.
- Define testing rules.
- Define rules for coding agents.

## Out of Scope

- Refactoring all packages immediately.
- Implementing all adapters.
- Implementing all providers.
- Implementing all prompt generators.
- Implementing all validators.

## Success Criteria

- The project has a living architecture spec.
- Future changes have clear architectural boundaries.
- Spec authoring supports provider-based and agent-assisted workflows.
- Adding new providers, agents, or authoring modes follows Open/Closed Principle.
