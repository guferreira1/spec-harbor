# Design: Define Project Architecture

## Overview

The architecture specification introduces Hexagonal Architecture as the main structural style for SpecHarbor.

The core contains domain, ports, and use cases. Adapters implement external details such as CLI, file system, AI providers, coding agents, templates, project scanners, validators, and workflow connectors.

## Core Abstractions

### SpecAuthor

Central abstraction for creating OpenSpec changes.

Implementations:

- BlankSpecAuthor
- GuidedSpecAuthor
- TemplateSpecAuthor
- AIProviderSpecAuthor
- AgentAssistedSpecAuthor
- HybridSpecAuthor

### AIProvider

Represents direct AI model provider integrations.

Examples:

- OpenAI
- Anthropic
- Gemini
- Ollama
- Azure OpenAI
- OpenRouter
- Mistral

### AgentTarget

Represents coding tools that consume prompts or specs.

Examples:

- Codex
- Claude Code
- Devin
- Cursor
- GitHub Copilot
- Windsurf
- Aider
- Roo Code

### WorkflowConnector

Represents external workflow integrations.

Examples:

- GitHub
- GitLab
- Jira
- Linear
- Devin
- Slack

## Patterns

- Strategy for authoring modes, providers, prompt generators, validators, and connectors.
- Registry/Factory for resolving implementations.
- Adapter for external systems.
- Chain of Responsibility for validation.
- Command for CLI execution.

## Dependency Rule

All dependencies must point inward.

```text
cmd -> adapters -> core/usecase -> core/ports + core/domain
```

## Migration

The current structure may evolve gradually into the target structure.

Future OpenSpec changes should move code toward:

```text
internal/core
internal/adapters
internal/platform
```
