# Risks: Refactor to Hexagonal Architecture

## Over-abstraction

Moving to Hexagonal Architecture too early may create interfaces that do not yet have enough behavior behind them.

Mitigation:

- Add ports only when use cases consume them.
- Keep interfaces small and behavior-specific.
- Prefer value-object moves before introducing new abstractions.

## Large rewrite

Moving all packages in one step can create avoidable merge conflicts, import churn, and difficult reviews.

Mitigation:

- Migrate incrementally by responsibility.
- Keep each future implementation task focused on one package group or use case.
- Preserve behavior before introducing new capabilities.

## Business rules leaking into adapters

CLI, scanner, provider, template, prompt, or workflow adapters may accidentally accumulate orchestration or business rules.

Mitigation:

- Keep adapters focused on translation, IO, and external integration.
- Put orchestration in `internal/core/usecase`.
- Test use cases through ports.

## Provider and agent coupling

AI provider integrations and coding agent prompt generation may become coupled if both are treated as the same kind of extension.

Mitigation:

- Keep direct model APIs in `internal/adapters/ai`.
- Keep coding agent prompt rendering in `internal/adapters/prompt`.
- Keep workflow execution integrations in `internal/adapters/workflow`.
- Preserve agent-assisted authoring without API key requirements.

## Platform package growth

`internal/platform` can become a dumping ground for unrelated utilities.

Mitigation:

- Use platform only for minimal technical helpers such as version, logging, and shared error helpers.
- Do not place business rules or use-case decisions in platform packages.
- Prefer domain or adapter ownership when behavior has a clear architectural home.

## Temporary duplicate concepts

During incremental migration, concepts such as generation modes, providers, agents, or project context may temporarily exist in both old and new packages.

Mitigation:

- Keep compatibility aliases short-lived.
- Complete one concept migration at a time.
- Remove old package exports once callers are migrated.
