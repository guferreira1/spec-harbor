# SpecHarbor Project Context

## Purpose

SpecHarbor is an open source CLI for generating, validating, reviewing, and archiving OpenSpec-based development workflows for coding agents.

## Core idea

The project helps developers convert a loose idea into:

- a structured OpenSpec change;
- a technical design;
- implementation tasks;
- acceptance criteria;
- risk notes;
- agent-specific implementation prompts.

## Generation strategies

- Blank generation
- Guided generation
- Template generation
- AI-assisted generation
- Hybrid generation

## Technical stack

- Language: Go
- Interface: CLI
- Configuration: YAML
- Documentation: Markdown
- CI: GitHub Actions

## Architecture boundaries

- Domain concepts and value objects belong in `internal/core/domain`.
- Ports consumed by use cases belong in `internal/core/ports`.
- Application orchestration belongs in `internal/core/usecase`.
- Concrete delivery mechanisms, providers, scanners, prompts, validators, templates, workflow connectors, and config implementations belong in `internal/adapters`.
- Minimal cross-cutting technical helpers belong in `internal/platform`.

Current packages may evolve gradually through focused OpenSpec changes. Avoid placeholder packages or abstractions until concrete behavior needs them.
