# SpecHarbor

**SpecHarbor** is an open source CLI for generating, validating, and managing OpenSpec-based development workflows for AI coding agents.

It helps developers move from a loose feature idea to a structured specification, task list, acceptance criteria, and agent-ready implementation prompt.

## Why

AI coding agents are powerful, but without a clear specification they can easily:

- modify unrelated files;
- ignore project architecture;
- skip tests;
- invent requirements;
- create inconsistent implementations.

SpecHarbor turns the workflow into something more controlled:

```text
Idea -> OpenSpec change -> Tasks -> Agent prompt -> Implementation -> Review -> Archive
```

## Supported agent targets

SpecHarbor is designed to generate prompts for multiple AI coding agents, including:

- Codex
- Claude Code
- Cursor
- Devin
- GitHub Copilot
- Gemini CLI
- Roo Code
- Windsurf
- Aider
- Generic agents

## Generation modes

SpecHarbor supports multiple specification generation strategies:

- **Blank mode**: create a minimal OpenSpec change structure for teams that want full control.
- **Guided mode**: ask questions and generate a spec from the answers.
- **Template mode**: generate specs from built-in or custom templates.
- **AI-assisted mode**: generate specs using an AI provider.
- **Hybrid mode**: combine project scanning, user answers, templates, and AI.

## Planned commands

```bash
specharbor init
specharbor scan
specharbor generate "Add payment webhook with idempotency"
specharbor prompt add-payment-webhook --agent codex
specharbor validate add-payment-webhook
specharbor review add-payment-webhook
specharbor archive add-payment-webhook
specharbor config set ai.provider openai
```

## Current status

This repository contains the initial project structure and the first OpenSpec change for bootstrapping the CLI.

## License

MIT
