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

- CLI routing belongs in `internal/cli`.
- Project scanning belongs in `internal/scanner`.
- Spec generation belongs in `internal/generator`.
- Agent prompt rendering belongs in `internal/prompt`.
- Spec validation belongs in `internal/validator`.
- Provider integration belongs in `internal/ai`.
- OpenSpec file operations belong in `internal/openspec`.
