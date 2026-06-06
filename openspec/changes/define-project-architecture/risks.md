# Risks: Define Project Architecture

## Overengineering

The architecture may look heavy for an early CLI.

Mitigation:

- Apply boundaries gradually.
- Avoid unnecessary abstractions until behavior exists.
- Keep ports small and use-case driven.

## Pattern abuse

Design patterns can add ceremony if used without purpose.

Mitigation:

- Use patterns only where they isolate variation or support extension.
- Do not create abstractions without a clear reason.

## Provider complexity

AI providers and coding agents have different capabilities and interfaces.

Mitigation:

- Separate AI Providers from Agent Targets and Workflow Connectors.
- Keep provider-specific details inside adapters.

## Enterprise restrictions

Some companies may not allow direct API keys.

Mitigation:

- Support agent-assisted authoring for Devin, Codex, Claude Code, Cursor, Copilot, and similar tools.
